package entries

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

// Regex hell. I'm sorry.

var (
	// reFrontMatter matches the front matter of an entry.
	reFrontMatter = regexp.MustCompile(`^---\n(?:\n|.)+---\n+`)

	// reInitialNewlines matches to any newlines at the beginning of a string.
	reInitialNewlines = regexp.MustCompile(`^\n+`)

	// reInitialSentence matches to the first sentence in a string.
	reInitialSentence = regexp.MustCompile(`^(.*?)[\.!\?\n](?:\s|$)`)

	// reLinkTitleNoName matches to links which specify only the other entry's title, e.g. "[[Pizza]]" or "[[Ice Cream]]"
	// Group 1 is the title of the entry that is being linked to.
	reLinkTitleNoName = regexp.MustCompile(`\[\[([^\]]+)\]\]`)

	// reLinkTitleWithName matches to links which specify the other entry's title and a name for the link, e.g. "[[Pizza](Shown title)]"
	// Group 1 is the title of the entry that is being linked to.
	// Group 2 is the name of the link.
	reLinkTitleWithName = regexp.MustCompile(`\[\[([^\]]+)\]\(([^\)]+)\)\]`)

	// reLinkPathNoName matches to links which specify only the other entry's path, e.g. "{{food/pizza}}" or "{{food/ice-cream}}"
	// Group 1 is the path of the entry being linked to.
	reLinkPathNoName = regexp.MustCompile(`{{([^\}]+)}}`)

	// reLinkPathWithName matches to links which specify the other entry's path and a name for the link, e.g. "{{food/ice-cream}(Ice Cream)}"
	// Group 1 is the path of the entry being linked to.
	// Group 2 is the name of the link.
	reLinkPathWithName = regexp.MustCompile(`{{([^}]+)}\(([^)]+)\)}`)
)

// YAMLFrontMatter represents the normal YAML front matter at the start of an entry.
type YAMLFrontMatter struct {
	Date  string   `yaml:"date"`
	Title string   `yaml:"title"`
	Tags  []string `yaml:"tags"`
}

// Parser represents an entry parser.
type Parser struct {
	dateLayout string

	reBuiltinTag *regexp.Regexp
	reCustomTag  *regexp.Regexp
}

// NewParser returns a new parser.
func NewParser(dateLayout, builtinTagPrefix, customTagPrefix string) (Parser, error) {
	reBuiltinTag, err := regexp.Compile(regexp.QuoteMeta(builtinTagPrefix) + "[\\w|-]+")
	if err != nil {
		return Parser{}, fmt.Errorf("could not build custom tag regex: %w", err)
	}

	reCustomTag, err := regexp.Compile(regexp.QuoteMeta(customTagPrefix) + "[\\w|-]+")
	if err != nil {
		return Parser{}, fmt.Errorf("could not build custom tag regex: %w", err)
	}

	return Parser{
		dateLayout:   dateLayout,
		reBuiltinTag: reBuiltinTag,
		reCustomTag:  reCustomTag,
	}, nil
}

// err creates a new ErrEntryParseFailed with the default values filled in.
func (p Parser) err(path string, format string, a ...interface{}) error {
	return ErrEntryParseFailed{
		Path: path,
		Err:  fmt.Errorf(format, a...),
	}
}

// Parse the content of an `entry.md` file into an Entry struct.
// It does this in 4 stages:
// 1. Parse the front matter and remove it from the entry's content.
// 2. Gets a title and date value from the entry content if they weren't specified in the front-matter.
// 3. Parse tags.
// 4. Parse links.
func (p Parser) Parse(path, content string) (*Entry, error) {
	var entry = &Entry{}

	// Extract the front matter text from the file and return the entry's content without the front matter present
	frontMatter, strippedContent, err := p.extractFrontMatter(path, content)
	if err != nil {
		return nil, err
	}

	// Attempt to parse the front matter into a YAMLFrontMatter struct. This is because we know the types of the Title,
	// Tags and Date keys.
	concrete, err := p.parseFrontMatterConcrete(path, frontMatter)
	if err != nil {
		return nil, err
	}

	// If we have no title specified in the front matter (concrete.Title), use the first sentence as the title of the entry.
	if concrete.Title == "" {
		title, err := p.getFirstSentence(path, strippedContent)
		if err != nil {
			return nil, err
		}

		entry.Title = title
	} else {
		entry.Title = concrete.Title
	}

	// If the date key in the front matter wasn't empty, we use that value to set the date of the entry. Later, in the
	// NewEntryFromFile function, we will set the date to the modification time of the file if no date was specified
	// explicitly.
	if concrete.Date != "" {
		d, err := time.Parse(p.dateLayout, concrete.Date)
		if err != nil {
			return nil, p.err(path, "couldn't parse date '%s' with layout '%s': %w", concrete.Date, p.dateLayout, err)
		}

		entry.Date = d
	}

	// Now we've extracted the "concrete" front matter, i.e. the key values we know the types of, we then parse the front
	// matter also into a map[string]interface{} -- this means that additional front matter information specified by the user
	// can be accessed.
	// A side effect is that we end up storing the title, date and tags for the entry twice; once in entry.Title and also in
	// entry.Metadata["title"]. However, entry.Title is a string whereas entry.Metadata["title"] cannot be used without a
	// type assertion as it's an interface{}.
	mapFrontMatter, err := p.parseFrontMatterMap(path, frontMatter)
	if err != nil {
		return nil, err
	}

	entry.Metadata = mapFrontMatter
	entry.Contents = strippedContent
	entry.OriginalContents = content

	// Here we deal with tags. We don't want duplicates so we initialise a new map which stores the tags present in the entry.
	// Setting the same tag twice will only result in one map entry so it acts like a set.
	tagMap := make(map[string]bool)
	for _, tag := range concrete.Tags {
		tagMap[tag] = true
	}

	tags, err := p.parseTags(path, strippedContent)
	if err != nil {
		return nil, err
	}

	for _, tag := range tags {
		tagMap[tag] = true
	}

	// Now we put the tags in the map into the entry.Tags field of the struct.
	for tag := range tagMap {
		entry.Tags = append(entry.Tags, tag)
	}

	entry.OutboundLinks = p.parseLinks(path, strippedContent)
	for i := range entry.OutboundLinks {
		entry.OutboundLinks[i].Parent = entry
	}

	return entry, nil
}

// extractFrontMatter extracts the YAML front matter text from the entry and returns it, along with setting
// the .strippedContent value to the original content without the front matter included.
func (p Parser) extractFrontMatter(path, content string) (frontMatter string, strippedContent string, err error) {
	if !reFrontMatter.MatchString(content) {
		// No front-matter in text.
		strippedContent = reInitialNewlines.ReplaceAllString(content, "")
		return "", strippedContent, nil
	}

	startOffset := strings.Index(content, "---") + 4
	endOffset := strings.Index(content[startOffset:], "---") + startOffset

	if startOffset > endOffset {
		return "", "", p.err(path, "could not find end offset of yaml front matter")
	}

	frontMatter = content[startOffset:endOffset]
	frontMatter = strings.Trim(frontMatter, "\n")

	strippedContent = content[endOffset+4:]
	strippedContent = strings.TrimLeft(strippedContent, "\n")

	return frontMatter, strippedContent, nil
}

// parseFrontMatterConcrete takes the string of a YAML front matter and unmarshals it to a struct.
func (p Parser) parseFrontMatterConcrete(path, frontMatter string) (YAMLFrontMatter, error) {
	config := YAMLFrontMatter{}
	err := yaml.Unmarshal([]byte(frontMatter), &config)
	if err != nil {
		return YAMLFrontMatter{}, p.err(path, "couldn't unmarshal front matter: %w", err)
	}

	return config, nil
}

// parseFrontMatterMap takes the string of a YAML front matter and unmarshals it to a map[string]interface{}.
func (p Parser) parseFrontMatterMap(path, frontMatter string) (map[string]interface{}, error) {
	config := make(map[string]interface{})
	err := yaml.Unmarshal([]byte(frontMatter), &config)
	if err != nil {
		return nil, p.err(path, "couldn't unmarshal front matter: %w", err)
	}

	return config, nil
}

// getFirstSentence returns the first sentence from the entry text. This is used to get an alternate title if
// no other is available.
// This function will remove an ending full stop, but no other pieces of punctuation. This is a stylistic choice, for example:
//   "A Day At A Restaurant." => "A Day At A Restaurant", "A Day At A Restaurant!!" => "A Day At A Restaurant!!"
// It will also remove trailing newlines.
func (p Parser) getFirstSentence(path, strippedContent string) (string, error) {
	initialSentence := reInitialSentence.FindString(strippedContent)

	initialSentence = strings.Trim(initialSentence, "\n")
	initialSentence = strings.Trim(initialSentence, ".")

	if initialSentence == "" {
		return "", p.err(path, "could not locate title as front matter or initial sentence")
	}

	return initialSentence, nil
}

// parseTags returns all the tags in the text. The prefixes are included.
func (p Parser) parseTags(path, strippedContent string) ([]string, error) {
	results := []string{}

	builtinMatches := p.reBuiltinTag.FindAllString(strippedContent, -1)
	if builtinMatches != nil {
		results = append(results, builtinMatches...)
	}

	customMatches := p.reCustomTag.FindAllString(strippedContent, -1)
	if customMatches != nil {
		results = append(results, customMatches...)
	}

	return results, nil
}

// parseLinks returns all the links present in the text.
func (p Parser) parseLinks(path, strippedContent string) []Link {
	var links []Link

	titleNoNameLinks := p.parseLinksTitleNoName(path, strippedContent)
	if titleNoNameLinks != nil {
		links = append(links, titleNoNameLinks...)
	}

	titleWithNameLinks := p.parseLinksTitleWithName(path, strippedContent)
	if titleWithNameLinks != nil {
		links = append(links, titleWithNameLinks...)
	}

	pathNoNameLinks := p.parseLinksPathNoName(path, strippedContent)
	if pathNoNameLinks != nil {
		links = append(links, pathNoNameLinks...)
	}

	pathWithNameLinks := p.parseLinksPathWithName(path, strippedContent)
	if pathWithNameLinks != nil {
		links = append(links, pathWithNameLinks...)
	}

	return links
}

func (p Parser) parseLinksTitleNoName(path, strippedContent string) []Link {
	var links []Link
	matches := reLinkTitleNoName.FindAllSubmatchIndex([]byte(strippedContent), -1)

	// match is an []int
	// [0] and [1] are the positions of the whole match.
	// [2] and [3] are the positions of the title.
	for _, match := range matches {
		title := strippedContent[match[2]:match[3]]
		links = append(links, Link{
			Title: title,
			Loc:   match[:2],
			Type:  LinkTitleNoName,
		})
	}

	return links
}

func (p Parser) parseLinksTitleWithName(path, strippedContent string) []Link {
	var links []Link
	matches := reLinkTitleWithName.FindAllSubmatchIndex([]byte(strippedContent), -1)

	// match is an []int
	// [0] and [1] are the positions of the whole match.
	// [2] and [3] are the positions of the title.
	// [4] and [5] are the positions of the name of the link
	for _, match := range matches {
		title := strippedContent[match[2]:match[3]]
		name := strippedContent[match[4]:match[5]]
		links = append(links, Link{
			Title: title,
			Name:  name,
			Loc:   match[:2],
			Type:  LinkTitleWithName,
		})
	}

	return links
}

func (p Parser) parseLinksPathNoName(path, strippedContent string) []Link {
	var links []Link
	matches := reLinkPathNoName.FindAllSubmatchIndex([]byte(strippedContent), -1)

	// match is an []int
	// [0] and [1] are the positions of the whole match.
	// [2] and [3] are the positions of the path.
	for _, match := range matches {
		path := strippedContent[match[2]:match[3]]
		links = append(links, Link{
			Path: path,
			Loc:  match[:2],
			Type: LinkPathNoName,
		})
	}

	return links
}

func (p Parser) parseLinksPathWithName(path, strippedContent string) []Link {
	var links []Link
	matches := reLinkPathWithName.FindAllSubmatchIndex([]byte(strippedContent), -1)

	// match is an []int
	// [0] and [1] are the positions of the whole match.
	// [2] and [3] are the positions of the path.
	// [4] and [5] are the positions of the name of the link
	for _, match := range matches {
		path := strippedContent[match[2]:match[3]]
		name := strippedContent[match[4]:match[5]]
		links = append(links, Link{
			Path: path,
			Name: name,
			Loc:  match[:2],
			Type: LinkPathWithName,
		})
	}

	return links
}
