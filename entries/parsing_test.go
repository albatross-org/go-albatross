package entries

import (
	"testing"

	. "github.com/stretchr/testify/assert"
)

func TestParseCorrectExtendedMarkdown(t *testing.T) {
	ts := []struct {
		name, contents string

		Title    string
		Date     string
		Contents string
		Metadata map[string]interface{}
		Tags     []string
		Links    []Link
	}{
		{
			"front-matter-required", `---
title: "Hello, Testing!"
date: "2020-08-05 11:58"
---

This is a test of the entry parsing functionality.`,
			"Hello, Testing!",
			"2020-08-05 11:58",
			"This is a test of the entry parsing functionality.",
			map[string]interface{}{"title": "Hello, Testing!", "date": "2020-08-05 11:58"},
			nil,
			[]Link{},
		},
		{
			"front-matter-additional", `---
title: "Hello, Testing!"
date: "2020-08-05 11:58"

geolocation: "tailed.emote.icicle"
custom:
    weight: "9 stone 10 pounds"
---

This is a test of the entry parsing functionality.`,
			"Hello, Testing!",
			"2020-08-05 11:58",
			"This is a test of the entry parsing functionality.",
			map[string]interface{}{
				"title": "Hello, Testing!", "date": "2020-08-05 11:58",
				"custom": map[string]interface{}{
					"weight": "9 stone 10 pounds",
				},
			},
			nil,
			[]Link{},
		},
		{
			"title-implied", `---
date: "2020-08-05 11:58"
---

This is the first sentence. This is another sentence.`,
			"This is the first sentence",
			"2020-08-05 11:58",
			"This is the first sentence. This is another sentence.",
			map[string]interface{}{"date": "2020-08-05 11:58"},
			nil,
			[]Link{},
		},
		{
			"tags-inline", `---
date: "2020-08-05 11:58"
title: "Inline Tag Test!"
---

Hey, I'm a @!builtin-tag. I'm a @?custom-tag. Tags don't @?need @!dashes though. @!but-you-can @?have-a-lot if you want.`,
			"Inline Tag Test!",
			"2020-08-05 11:58",
			"Hey, I'm a @!builtin-tag. I'm a @?custom-tag. Tags don't @?need @!dashes though. @!but-you-can @?have-a-lot if you want.",
			map[string]interface{}{"title": "Tag Test!", "date": "2020-08-05 11:58"},
			[]string{
				"@!builtin-tag",
				"@!dashes",
				"@!but-you-can",
				"@?custom-tag",
				"@?need",
				"@?have-a-lot",
			},
			[]Link{},
		},
		{
			"tags-front-matter", `---
date: "2020-08-05 11:58"
title: "Front-Matter Tag Test!"
tags: ["@!test", "@?test"]
---

Content.`,
			"Front-Matter Tag Test!",
			"2020-08-05 11:58",
			"Content.",
			map[string]interface{}{"title": "Tag Test!", "date": "2020-08-05 11:58"},
			[]string{
				"@!test",
				"@?test",
			},
			[]Link{},
		},
	}

	for _, tc := range ts {
		t.Run(tc.name, func(t *testing.T) {
			entry, err := parseExtendedMarkdown("test/path/entry.md", tc.contents, "2006-01-02 15:04", "@!", "@?")
			if err != nil {
				t.Fatalf("got error while parsing markdown: %v", err)
			}

			Equal(t, tc.Title, entry.Title)
			Equal(t, tc.Date, entry.Date.Format("2006-01-02 15:04"))
			Equal(t, tc.Contents, entry.Contents)
			Equal(t, tc.Tags, entry.Tags)
			// Equal(t, tc.Links, entry.OutboundLinks)
		})
	}
}
