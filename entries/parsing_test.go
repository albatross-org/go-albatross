package entries

import (
	"testing"

	. "github.com/stretchr/testify/assert"
)

var (
	testDateLayout       = "2006-01-02 15:04"
	testBuiltinTagPrefix = "@!"
	testCustomTagPrefix  = "@?"
)

func dummyEntryWithContent(content string) string {
	return `---
title: "Dummy Entry"
date: "2020-08-05 11:58"
---

` + content
}

func newTestParser(t *testing.T) Parser {
	parser, err := NewParser(testDateLayout, testBuiltinTagPrefix, testCustomTagPrefix)
	if err != nil {
		t.Fatalf("couldn't create new parser: %s", err)
	}

	return parser
}

func parseForTest(t *testing.T, p Parser, content string) *Entry {
	entry, err := p.Parse("test/entry", content)
	if err != nil {
		t.Fatalf("wasn't expecting parsing error: %s", err)
	}

	return entry
}

func TestExtractFrontMatterString(t *testing.T) {
	p := newTestParser(t)
	content := `---
title: "Dummy Entry"
date: "2020-08-05 11:58"
---

This is some content.`

	expectedFrontMatter := `title: "Dummy Entry"
date: "2020-08-05 11:58"`
	expectedStrippedContent := `This is some content.`

	frontMatter, strippedContent, err := p.extractFrontMatter("/test/entry", content)
	NoError(t, err, "extractFrontMatter shouldn't return an error")

	Equal(t, expectedFrontMatter, frontMatter)
	Equal(t, expectedStrippedContent, strippedContent)
}

func TestParseFrontMatterConcrete(t *testing.T) {
	p := newTestParser(t)
	content := `---
title: "Dummy Entry"
date: "2020-08-05 11:58"
---

This is some content.`

	entry := parseForTest(t, p, content)

	Equal(t, "Dummy Entry", entry.Title)
	Equal(t, "2020-08-05 11:58", entry.Date.Format(testDateLayout))
}

func TestParseFrontMatterMap(t *testing.T) {
	p := newTestParser(t)
	content := `---
title: "Dummy Entry"
date: "2020-08-05 11:58"

geolocation: "tailed.emote.icicle"
custom:
    weight: "9 stone 10 pounds"
---

This is some content.`

	entry := parseForTest(t, p, content)

	Equal(t, map[string]interface{}{
		"title":       "Dummy Entry",
		"date":        "2020-08-05 11:58",
		"geolocation": "tailed.emote.icicle",
		"custom": map[interface{}]interface{}{
			"weight": "9 stone 10 pounds",
		},
	}, entry.Metadata)
}

func TestParseInferMissing(t *testing.T) {
	p := newTestParser(t)
	content := `Uh oh, I'm an entry without a title.`
	entry := parseForTest(t, p, content)

	Equal(t, "Uh oh, I'm an entry without a title", entry.Title)
}

func TestParseTags(t *testing.T) {
	p := newTestParser(t)
	content := `---
title: "Dummy Entry"
date: "2020-08-05 11:58"
tags:
 - "@!tag-frontmatter-builtin"
 - "@?tag-frontmatter-custom"
---

This is some content. I'm a @!tag-inline-builtin. Now I'm a @?tag-inline-custom.`

	entry := parseForTest(t, p, content)
	Equal(
		t,
		[]string{
			"@!tag-frontmatter-builtin",
			"@?tag-frontmatter-custom",
			"@!tag-inline-builtin",
			"@?tag-inline-custom",
		},
		entry.Tags,
	)
}

func TestParseLinksTitleNoName(t *testing.T) {
	p := newTestParser(t)
	content := dummyEntryWithContent(
		"As mentioned in [[Pizza]], I'm very [[Hungry]].",
	)

	links := p.parseLinks("test/entry", content)

	if len(links) != 2 {
		t.Fatalf("expected 2 links to be matched, got=%d", len(links))
	}

	Equal(t, LinkTitleNoName, links[0].Type, "Type should be correct")
	Equal(t, LinkTitleNoName, links[1].Type, "Type should be correct")

	Equal(t, "Pizza", links[0].Title, "Titles should match")
	Equal(t, "Hungry", links[1].Title, "Titles should match")

	Equal(t, "[[Pizza]]", content[links[0].Loc[0]:links[0].Loc[1]])
	Equal(t, "[[Hungry]]", content[links[1].Loc[0]:links[1].Loc[1]])
}

func TestParseLinksTitleWithName(t *testing.T) {
	p := newTestParser(t)
	content := dummyEntryWithContent(
		"As mentioned in [[Pizza](name 1)], I'm very [[Hungry](name 2)].",
	)

	links := p.parseLinks("test/entry", content)

	if len(links) != 2 {
		t.Fatalf("expected 2 links to be matched, got=%d", len(links))
	}

	Equal(t, LinkTitleWithName, links[0].Type, "Type should be correct")
	Equal(t, LinkTitleWithName, links[1].Type, "Type should be correct")

	Equal(t, "Pizza", links[0].Title, "Titles should match")
	Equal(t, "Hungry", links[1].Title, "Titles should match")

	Equal(t, "name 1", links[0].Name, "Names should match")
	Equal(t, "name 2", links[1].Name, "Names should match")

	Equal(t, "[[Pizza](name 1)]", content[links[0].Loc[0]:links[0].Loc[1]])
	Equal(t, "[[Hungry](name 2)]", content[links[1].Loc[0]:links[1].Loc[1]])
}

func TestParseLinksPathNoName(t *testing.T) {
	p := newTestParser(t)
	content := dummyEntryWithContent(
		"As mentioned in {{food/pizza}}, I'm very {{moods/hungry}}.",
	)

	links := p.parseLinks("test/entry", content)

	if len(links) != 2 {
		t.Fatalf("expected 2 links to be matched, got=%d", len(links))
	}

	Equal(t, LinkPathNoName, links[0].Type, "Type should be correct")
	Equal(t, LinkPathNoName, links[1].Type, "Type should be correct")

	Equal(t, "food/pizza", links[0].Path, "Paths should match")
	Equal(t, "moods/hungry", links[1].Path, "Paths should match")

	Equal(t, "{{food/pizza}}", content[links[0].Loc[0]:links[0].Loc[1]])
	Equal(t, "{{moods/hungry}}", content[links[1].Loc[0]:links[1].Loc[1]])
}

func TestParseLinksPathWithName(t *testing.T) {
	p := newTestParser(t)
	content := dummyEntryWithContent(
		"As mentioned in {{food/pizza}(name 1)}, I'm very {{moods/hungry}(name 2)}.",
	)

	links := p.parseLinks("test/entry", content)

	if len(links) != 2 {
		t.Fatalf("expected 2 links to be matched, got=%d", len(links))
	}

	Equal(t, LinkPathWithName, links[0].Type, "Type should be correct")
	Equal(t, LinkPathWithName, links[1].Type, "Type should be correct")

	Equal(t, "food/pizza", links[0].Path, "Paths should match")
	Equal(t, "moods/hungry", links[1].Path, "Paths should match")

	Equal(t, "name 1", links[0].Name, "Names should match")
	Equal(t, "name 2", links[1].Name, "Names should match")

	Equal(t, "{{food/pizza}(name 1)}", content[links[0].Loc[0]:links[0].Loc[1]])
	Equal(t, "{{moods/hungry}(name 2)}", content[links[1].Loc[0]:links[1].Loc[1]])
}
