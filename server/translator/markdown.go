package translator

import (
	"context"
	"regexp"
	"strings"
)

// codePattern matches fenced code blocks and inline code spans, which stay untranslated.
// An unclosed fence does not match, so its text is translated as prose.
var codePattern = regexp.MustCompile("(?s)```.*?```|~~~.*?~~~|`[^`\n]+`")

// TranslateMessage translates a Mattermost message and leaves its code untranslated.
// It returns the translated message and the language detected in the longest prose part.
// A message with no prose comes back unchanged, with no detected language.
func TranslateMessage(ctx context.Context, t Translator, message, to string) (string, string, error) {
	type part struct {
		text      string
		translate bool
	}
	var parts []part
	last := 0
	for _, loc := range codePattern.FindAllStringIndex(message, -1) {
		parts = append(parts, part{message[last:loc[0]], true}, part{message[loc[0]:loc[1]], false})
		last = loc[1]
	}
	parts = append(parts, part{message[last:], true})

	// The service may trim whitespace, so send only the core of each prose part and
	// put its leading and trailing whitespace back afterwards.
	var texts []string
	var indexes []int
	for i, p := range parts {
		if p.translate && strings.TrimSpace(p.text) != "" {
			texts = append(texts, strings.TrimSpace(p.text))
			indexes = append(indexes, i)
		}
	}
	if len(texts) == 0 {
		return message, "", nil
	}

	results, err := t.Translate(ctx, texts, to)
	if err != nil {
		return "", "", err
	}

	detected, longest := "", -1
	for j, i := range indexes {
		core := texts[j]
		before, after, _ := strings.Cut(parts[i].text, core)
		parts[i].text = before + results[j].Text + after
		if len(core) > longest {
			detected, longest = results[j].DetectedLanguage, len(core)
		}
	}

	var b strings.Builder
	for _, p := range parts {
		b.WriteString(p.text)
	}
	return b.String(), detected, nil
}
