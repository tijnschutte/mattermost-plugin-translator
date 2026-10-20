package translator

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// upperTranslator "translates" by upper-casing, and records what it was sent.
type upperTranslator struct {
	sent [][]string
	err  error
}

func (u *upperTranslator) Translate(_ context.Context, texts []string, _ string) ([]Result, error) {
	u.sent = append(u.sent, texts)
	if u.err != nil {
		return nil, u.err
	}
	out := make([]Result, len(texts))
	for i, text := range texts {
		out[i] = Result{Text: strings.ToUpper(text), DetectedLanguage: "nl"}
		if strings.HasPrefix(text, "short") {
			out[i].DetectedLanguage = "en"
		}
	}
	return out, nil
}

func TestTranslateMessage(t *testing.T) {
	tests := []struct {
		name    string
		message string
		want    string
		sent    []string
	}{
		{
			name:    "plain text",
			message: "goedemorgen allemaal",
			want:    "GOEDEMORGEN ALLEMAAL",
			sent:    []string{"goedemorgen allemaal"},
		},
		{
			name:    "inline code stays",
			message: "run `make quick` eerst",
			want:    "RUN `make quick` EERST",
			sent:    []string{"run", "eerst"},
		},
		{
			name:    "fenced code stays, whitespace kept",
			message: "kijk hier:\n```go\nfmt.Println(\"hoi\")\n```\n  klaar",
			want:    "KIJK HIER:\n```go\nfmt.Println(\"hoi\")\n```\n  KLAAR",
			sent:    []string{"kijk hier:", "klaar"},
		},
		{
			name:    "unclosed fence is prose",
			message: "```nog niet klaar",
			want:    "```NOG NIET KLAAR",
			sent:    []string{"```nog niet klaar"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &upperTranslator{}

			got, detected, err := TranslateMessage(context.Background(), fake, tt.message, "pt-pt")

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, "nl", detected)
			assert.Equal(t, [][]string{tt.sent}, fake.sent)
		})
	}
}

func TestTranslateMessageOnlyCode(t *testing.T) {
	fake := &upperTranslator{}

	got, detected, err := TranslateMessage(context.Background(), fake, "```\nls -la\n```", "pt-pt")

	require.NoError(t, err)
	assert.Equal(t, "```\nls -la\n```", got)
	assert.Empty(t, detected)
	assert.Empty(t, fake.sent)
}

func TestTranslateMessageDetectsFromLongestPart(t *testing.T) {
	fake := &upperTranslator{}

	_, detected, err := TranslateMessage(context.Background(), fake, "short `x` dit is het langste deel van het bericht", "pt-pt")

	require.NoError(t, err)
	assert.Equal(t, "nl", detected)
}

func TestTranslateMessageError(t *testing.T) {
	fake := &upperTranslator{err: ErrQuotaExceeded}

	_, _, err := TranslateMessage(context.Background(), fake, "goedemorgen", "pt-pt")

	assert.True(t, errors.Is(err, ErrQuotaExceeded))
}

func TestSameLanguage(t *testing.T) {
	assert.True(t, SameLanguage("pt", "pt-PT"))
	assert.True(t, SameLanguage("NL", "nl"))
	assert.False(t, SameLanguage("nl", "pt-pt"))
}
