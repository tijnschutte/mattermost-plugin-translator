package translator

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func fakeLibreTranslate(t *testing.T, status int, body string) (*LibreTranslate, *libreRequest) {
	t.Helper()
	var got libreRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/translate", r.URL.Path)
		require.NoError(t, json.NewDecoder(r.Body).Decode(&got))
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)
	return NewLibreTranslate(server.URL+"/", "secret"), &got
}

func TestLibreTranslateTranslate(t *testing.T) {
	lt, req := fakeLibreTranslate(t, http.StatusOK, `{
		"translatedText": ["Bom dia", "Até amanhã"],
		"detectedLanguage": [{"confidence": 90, "language": "nl"}, {"confidence": 90, "language": "nl"}]
	}`)

	results, err := lt.Translate(context.Background(), []string{"Goedemorgen", "Tot morgen"}, "pt-PT")

	require.NoError(t, err)
	assert.Equal(t, []Result{
		{Text: "Bom dia", DetectedLanguage: "nl"},
		{Text: "Até amanhã", DetectedLanguage: "nl"},
	}, results)
	assert.Equal(t, libreRequest{Q: []string{"Goedemorgen", "Tot morgen"}, Source: "auto", Target: "pt", Format: "text", APIKey: "secret"}, *req)
}

func TestLibreTranslateError(t *testing.T) {
	lt, _ := fakeLibreTranslate(t, http.StatusBadRequest, `{"error": "xx is not supported"}`)

	_, err := lt.Translate(context.Background(), []string{"Goedemorgen"}, "xx")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "xx is not supported")
}

func TestLibreTranslateResultCountMismatch(t *testing.T) {
	lt, _ := fakeLibreTranslate(t, http.StatusOK, `{"translatedText": []}`)

	_, err := lt.Translate(context.Background(), []string{"Goedemorgen"}, "pt")

	assert.Error(t, err)
}
