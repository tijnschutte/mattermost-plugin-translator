// Package translator turns text into another language. Code outside this package depends
// only on the Translator interface, so the service behind it can change.
package translator

import (
	"context"
	"errors"
	"strings"
)

// ErrQuotaExceeded means the service refuses more text until its quota resets.
var ErrQuotaExceeded = errors.New("translation quota exceeded")

// Result is one translated text and the language the service detected in the source.
type Result struct {
	Text             string
	DetectedLanguage string
}

// Translator translates a batch of texts into one target language. It returns one
// Result per text, in the same order.
type Translator interface {
	Translate(ctx context.Context, texts []string, to string) ([]Result, error)
}

// SameLanguage reports whether two language codes name the same base language, so
// "pt" and "pt-PT" match.
func SameLanguage(a, b string) bool {
	base := func(code string) string {
		code, _, _ = strings.Cut(strings.ToLower(code), "-")
		return code
	}
	return base(a) == base(b)
}
