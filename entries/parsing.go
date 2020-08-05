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
