package entries

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

var (
	reFrontMatter     = regexp.MustCompile("^---\n(?:\n|.)+---\n+")
	reInitialNewlines = regexp.MustCompile("^\n+")
	reInitialSentence = regexp.MustCompile("^(.*?)[\\.!\\?](?:\\s|$)")
)

// YAMLFrontMatter represents the normal YAML front matter at the start of an entry.
type YAMLFrontMatter struct {
	date  string   `yaml:"date"`
	title string   `yaml:"title"`
	tags  []string `yaml:"tags"`
}

// Parser represents an entry parser.
type Parser struct {
	path string

	originalContent string
	strippedContent string

	dateLayout string

	reBuiltinTag *regexp.Regexp
	reCustomTag  *regexp.Regexp
}

// NewParser returns a new parser.
func NewParser(path, content, dateLayout, builtinTagPrefix, customTagPrefix string) (Parser, error) {
	reBuiltinTag, err := regexp.Compile(builtinTagPrefix)
	if err != nil {
		return Parser{}, ErrEntryParseFailed{
			Path: path, Err: fmt.Errorf("could not build builtin tag regex: %w", err),
		}
	}

	reCustomTag, err := regexp.Compile(regexp.QuoteMeta(customTagPrefix) + "[\\w|-]+")
	if err != nil {
		return Parser{}, ErrEntryParseFailed{
			Path: path, Err: fmt.Errorf("could not build custom tag regex: %w", err),
		}
	}

	return Parser{
		path:            path,
		originalContent: content,
		dateLayout:      dateLayout,
		reBuiltinTag:    reBuiltinTag,
		reCustomTag:     reCustomTag,
	}, nil
}

// err creates a new error with the default values filled in.
func (p Parser) err(format string, a ...interface{}) error {
	return ErrEntryParseFailed{
		Path: p.path,
		Err:  fmt.Errorf(format, a...),
	}
}

// Parse the parsed Entry struct.
// It does this in 4 stages:
// 1. Parse the front matter and remove it from the entry's content.
// 2. Gets a title and date value from the entry content if they weren't specified in the front-matter.
// 3. Parse tags.
// 4. Parse links.
func (p Parser) Parse() (*Entry, error) {
	var entry *Entry

	frontMatter, err := p.extractFrontMatter()
	if err != nil {
		return nil, err
	}

	concrete, err := p.parseFrontMatterStruct(frontMatter)
	if err != nil {
		return nil, err
	}

	if concrete.title == "" {
		title, err := p.getFirstSentence()
		if err != nil {
			return nil, err
		}

		entry.Title = title
	} else {
		entry.Title = concrete.title
	}

	return nil, nil
}

// extractFrontMatter extracts the YAML front matter text from the entry and returns it, along with setting
// the .strippedContent value to the original content without the front matter included.
func (p Parser) extractFrontMatter() (string, error) {
	if !reFrontMatter.MatchString(p.originalContent) {
		// No front-matter in text.
		p.strippedContent = reInitialNewlines.ReplaceAllString(p.originalContent, "")
		return "", nil
	}

	lines := strings.Split(p.originalContent, "\n")

	startOffset := 4 // "---\n", the byte offset of where the YAML starts.
	endOffset := 4   // The byte offset of where the YAML ends.

	// Iterate through the lines, skipping the first one as that is the opening
	// endOffset is initialised to 4 since we skip the inital line
	for _, line := range lines[1:] {
		if line != "---" {
			endOffset += len(line)
		} else {
			break
		}
	}

	endOffset += len(lines) // To include newlines since they were trimmed off the end.

	if startOffset > endOffset {
		return "", p.err("could not find end offset of yaml front matter")
	}

	frontMatter := p.originalContent[startOffset:endOffset]
	p.strippedContent = strings.ReplaceAll(p.originalContent, frontMatter, "")

	return frontMatter, nil
}

// parseFrontMatterStruct takes the string of a YAML front matter and unmarshals it to a struct.
func (p Parser) parseFrontMatterStruct(frontMatter string) (YAMLFrontMatter, error) {
	config := YAMLFrontMatter{}
	err := yaml.Unmarshal([]byte(frontMatter), &config)
	if err != nil {
		return YAMLFrontMatter{}, p.err("couldn't unmarshal front matter: %w", err)
	}

	return config, nil
}

// parseFrontMatterMap takes the string of a YAML front matter and unmarshals it to a map[string]interface{}.
func (p Parser) parseFrontMatterMap(frontMatter string) (map[string]interface{}, error) {
	config := make(map[string]interface{})
	err := yaml.Unmarshal([]byte(frontMatter), &config)
	if err != nil {
		return nil, p.err("couldn't unmarshal front matter: %w", err)
	}

	return config, nil
}

// getFirstSentence returns the first sentence from the entry text. This is used to get an alternate title if
// no other is available.
func (p Parser) getFirstSentence() (string, error) {
	initialSentence := reInitialSentence.FindString(p.strippedContent)
	initialSentence = strings.Trim(initialSentence, ".!? ") // Remove ending punctuation

	if initialSentence == "" {
		return "", p.err("could not locate title as front matter or initial sentence")
	}

	return initialSentence, nil
}

// parseFrontMatter extracts the key-value pairs from a markdown file with front-matter.
// It returns a map[string]interface{}, followed by a string which is the content with the front-matter removed.
func parseFrontMatter(path, content string) (map[string]interface{}, string, error) {
	if !reFrontMatter.MatchString(content) {
		// No front-matter in text.
		newContent := reInitialNewlines.ReplaceAllString(content, "")
		return make(map[string]interface{}), newContent, nil
	}

	lines := strings.Split(content, "\n")

	startOffset := 4 // "---\n", the byte offset of where the YAML starts.

	endOffset := 4 // The byte offset of where the YAML ends.

	// Iterate through the lines, skipping the first one as that is the opening
	// endOffset is initialised to 4 since we skip the inital line
	for _, line := range lines[1:] {
		if line != "---" {
			endOffset += len(line)
		} else {
			break
		}
	}

	endOffset += len(lines) // To include newlines since they were trimmed off the end.

	if startOffset > endOffset {
		return nil, "", ErrEntryParseFailed{Path: path, Err: errors.New("could not find end offset of yaml front matter")}
	}

	yamlContent := content[startOffset:endOffset]
	frontMatter := make(map[string]interface{})

	err := yaml.Unmarshal([]byte(yamlContent), &frontMatter)
	if err != nil {
		return nil, "", ErrEntryParseFailed{Path: path, Err: fmt.Errorf("could not unmarshal yaml: %w", err)}
	}

	// Remove the front matter all newlines before actual contents start.
	newContents := reFrontMatter.ReplaceAllString(content, "")
	return frontMatter, newContents, nil
}

// parseTags returns a list of tags present in the document.
func parseTags(path, content, builtinPrefix, customPrefix string) ([]string, error) {
	reBuiltinPrefix, err := regexp.Compile(regexp.QuoteMeta(builtinPrefix) + "[\\w|-]+")
	if err != nil {
		return nil, ErrEntryParseFailed{Path: path, Err: fmt.Errorf("could not build builtin tag regex: %w", err)}
	}

	reCustomPrefix, err := regexp.Compile(regexp.QuoteMeta(customPrefix) + "[\\w|-]+")
	if err != nil {
		return nil, ErrEntryParseFailed{Path: path, Err: fmt.Errorf("could not build custom tag regex: %w", err)}
	}

	results := []string{}
	// fmt.Fprintln(os.Stdout, "builtin MATCHES", reBuiltinPrefix.FindAllString(content, -1))

	builtinMatches := reBuiltinPrefix.FindAllString(content, -1)
	if builtinMatches != nil {
		results = append(results, builtinMatches...)
	}

	customMatches := reCustomPrefix.FindAllString(content, -1)
	if customMatches != nil {
		results = append(results, customMatches...)
	}

	return results, nil
}

func parseExtendedMarkdown(path, content, dateFormat, builtinTagPrefix, customTagPrefix string) (*Entry, error) {
	frontMatter, newContents, err := parseFrontMatter(path, content)
	if err != nil {
		return nil, err
	}

	entry := &Entry{Contents: newContents}

	if frontMatter["title"] == nil {
		initialSentence := reInitialSentence.FindString(newContents)
		initialSentence = strings.Trim(initialSentence, ".!? ") // Remove ending punctuation

		if initialSentence == "" {
			return nil, ErrEntryParseFailed{Path: path, Err: fmt.Errorf("could not locate title as front matter or initial sentence")}
		}

		entry.Title = initialSentence
	} else {
		title, ok := frontMatter["title"].(string)
		if !ok {
			return nil, ErrEntryParseFailed{Path: path, Err: fmt.Errorf("could not convert 'title' key in front matter to string")}
		}

		entry.Title = title
	}

	if frontMatter["date"] == nil {
		// TODO: logic to get date from text
	} else {
		dateString, ok := frontMatter["date"].(string)
		if !ok {
			return nil, ErrEntryParseFailed{Path: path, Err: fmt.Errorf("could not convert 'date' key in front matter to string")}
		}

		date, err := time.Parse(dateFormat, dateString)
		if err != nil {
			return nil, ErrEntryParseFailed{Path: path, Err: fmt.Errorf("could not parse date field with format '%s': %w", dateFormat, err)}
		}

		entry.Date = date
	}

	if frontMatter["tags"] != nil {
		tags, ok := frontMatter["tags"].([]string)
		if !ok {
			return nil, ErrEntryParseFailed{Path: path, Err: fmt.Errorf("could not convert 'tag' key in front matter to []string")}
		}

		entry.Tags = tags
	}

	tags, err := parseTags(path, content, builtinTagPrefix, customTagPrefix)
	if err != nil {
		return nil, err
	}

	if tags != nil {
		entry.Tags = append(entry.Tags, tags...)
	}

	return entry, nil
}
