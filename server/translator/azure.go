package translator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// DefaultAzureEndpoint is the global Azure AI Translator endpoint.
const DefaultAzureEndpoint = "https://api.cognitive.microsofttranslator.com"

// azureQuotaExceeded is the error code Azure returns when the F0 tier's monthly
// character quota is used up.
const azureQuotaExceeded = 403001

// Azure calls the Azure AI Translator Text API, version 3.0.
type Azure struct {
	endpoint string
	key      string
	region   string
	client   *http.Client
}

// NewAzure returns a Translator for the given endpoint, resource key, and resource region.
// An empty endpoint means DefaultAzureEndpoint.
func NewAzure(endpoint, key, region string) *Azure {
	if endpoint == "" {
		endpoint = DefaultAzureEndpoint
	}
	return &Azure{
		endpoint: endpoint,
		key:      key,
		region:   region,
		client:   &http.Client{Timeout: 10 * time.Second},
	}
}

type azureText struct {
	Text string `json:"Text"`
}

type azureResult struct {
	DetectedLanguage struct {
		Language string `json:"language"`
	} `json:"detectedLanguage"`
	Translations []struct {
		Text string `json:"text"`
	} `json:"translations"`
}

type azureError struct {
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func (a *Azure) Translate(ctx context.Context, texts []string, to string) ([]Result, error) {
	if len(texts) == 0 {
		return nil, nil
	}

	body := make([]azureText, len(texts))
	for i, text := range texts {
		body[i] = azureText{Text: text}
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("encode request: %w", err)
	}

	query := url.Values{"api-version": {"3.0"}, "to": {to}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.endpoint+"/translate?"+query.Encode(), bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Ocp-Apim-Subscription-Key", a.key)
	if a.region != "" {
		req.Header.Set("Ocp-Apim-Subscription-Region", a.region)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call Azure: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("read Azure response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var azErr azureError
		_ = json.Unmarshal(respBody, &azErr)
		if azErr.Error.Code == azureQuotaExceeded {
			return nil, ErrQuotaExceeded
		}
		return nil, fmt.Errorf("azure returned %d, code %d: %s", resp.StatusCode, azErr.Error.Code, azErr.Error.Message)
	}

	var results []azureResult
	if err := json.Unmarshal(respBody, &results); err != nil {
		return nil, fmt.Errorf("decode Azure response: %w", err)
	}
	if len(results) != len(texts) {
		return nil, fmt.Errorf("azure returned %d results for %d texts", len(results), len(texts))
	}

	out := make([]Result, len(results))
	for i, r := range results {
		if len(r.Translations) == 0 {
			return nil, fmt.Errorf("azure returned no translation for text %d", i)
		}
		out[i] = Result{Text: r.Translations[0].Text, DetectedLanguage: r.DetectedLanguage.Language}
	}
	return out, nil
}
