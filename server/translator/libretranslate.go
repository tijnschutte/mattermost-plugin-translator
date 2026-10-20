package translator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// LibreTranslate calls a LibreTranslate server, usually one you host yourself.
type LibreTranslate struct {
	url    string
	apiKey string
	client *http.Client
}

// NewLibreTranslate returns a Translator for the LibreTranslate server at url. The API
// key is optional; servers you host yourself usually need none.
func NewLibreTranslate(url, apiKey string) *LibreTranslate {
	return &LibreTranslate{
		url:    strings.TrimRight(url, "/"),
		apiKey: apiKey,
		// The models run on CPU, so a long post can take several seconds.
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

type libreRequest struct {
	Q      []string `json:"q"`
	Source string   `json:"source"`
	Target string   `json:"target"`
	Format string   `json:"format"`
	APIKey string   `json:"api_key,omitempty"`
}

type libreResponse struct {
	TranslatedText   []string `json:"translatedText"`
	DetectedLanguage []struct {
		Language string `json:"language"`
	} `json:"detectedLanguage"`
}

type libreError struct {
	Error string `json:"error"`
}

// libreLanguage maps a target code to LibreTranslate's codes, which have no region:
// "pt-pt" becomes "pt".
func libreLanguage(code string) string {
	base, _, _ := strings.Cut(strings.ToLower(code), "-")
	return base
}

func (l *LibreTranslate) Translate(ctx context.Context, texts []string, to string) ([]Result, error) {
	if len(texts) == 0 {
		return nil, nil
	}

	payload, err := json.Marshal(libreRequest{Q: texts, Source: "auto", Target: libreLanguage(to), Format: "text", APIKey: l.apiKey})
	if err != nil {
		return nil, fmt.Errorf("encode request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, l.url+"/translate", bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := l.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call LibreTranslate: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("read LibreTranslate response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		var lErr libreError
		_ = json.Unmarshal(body, &lErr)
		return nil, fmt.Errorf("libretranslate returned %d: %s", resp.StatusCode, lErr.Error)
	}

	var result libreResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("decode LibreTranslate response: %w", err)
	}
	if len(result.TranslatedText) != len(texts) {
		return nil, fmt.Errorf("libretranslate returned %d results for %d texts", len(result.TranslatedText), len(texts))
	}

	// LibreTranslate detects one language for the whole batch and repeats it per text.
	detected := ""
	if len(result.DetectedLanguage) > 0 {
		detected = result.DetectedLanguage[0].Language
	}
	out := make([]Result, len(texts))
	for i, text := range result.TranslatedText {
		out[i] = Result{Text: text, DetectedLanguage: detected}
	}
	return out, nil
}
