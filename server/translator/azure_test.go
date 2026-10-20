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

func fakeAzure(t *testing.T, status int, body string) (*Azure, *http.Request) {
	t.Helper()
	var got http.Request
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = *r.Clone(context.Background())
		var texts []azureText
		require.NoError(t, json.NewDecoder(r.Body).Decode(&texts))
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)
	return NewAzure(server.URL, "test-key", "westeurope"), &got
}

func TestAzureTranslate(t *testing.T) {
	azure, req := fakeAzure(t, http.StatusOK, `[
		{"detectedLanguage": {"language": "nl", "score": 1.0}, "translations": [{"text": "Bom dia", "to": "pt-pt"}]},
		{"detectedLanguage": {"language": "nl", "score": 0.9}, "translations": [{"text": "Até amanhã", "to": "pt-pt"}]}
	]`)

	results, err := azure.Translate(context.Background(), []string{"Goedemorgen", "Tot morgen"}, "pt-pt")

	require.NoError(t, err)
	assert.Equal(t, []Result{
		{Text: "Bom dia", DetectedLanguage: "nl"},
		{Text: "Até amanhã", DetectedLanguage: "nl"},
	}, results)
	assert.Equal(t, "/translate", req.URL.Path)
	assert.Equal(t, "3.0", req.URL.Query().Get("api-version"))
	assert.Equal(t, "pt-pt", req.URL.Query().Get("to"))
	assert.Equal(t, "test-key", req.Header.Get("Ocp-Apim-Subscription-Key"))
	assert.Equal(t, "westeurope", req.Header.Get("Ocp-Apim-Subscription-Region"))
}

func TestAzureQuotaExceeded(t *testing.T) {
	azure, _ := fakeAzure(t, http.StatusForbidden, `{"error": {"code": 403001, "message": "The operation is not allowed because the subscription has exceeded its free quota."}}`)

	_, err := azure.Translate(context.Background(), []string{"Goedemorgen"}, "pt-pt")

	assert.ErrorIs(t, err, ErrQuotaExceeded)
}

func TestAzureOtherError(t *testing.T) {
	azure, _ := fakeAzure(t, http.StatusUnauthorized, `{"error": {"code": 401000, "message": "Access denied due to invalid subscription key."}}`)

	_, err := azure.Translate(context.Background(), []string{"Goedemorgen"}, "pt-pt")

	require.Error(t, err)
	assert.NotErrorIs(t, err, ErrQuotaExceeded)
	assert.Contains(t, err.Error(), "401000")
}

func TestAzureResultCountMismatch(t *testing.T) {
	azure, _ := fakeAzure(t, http.StatusOK, `[]`)

	_, err := azure.Translate(context.Background(), []string{"Goedemorgen"}, "pt-pt")

	assert.Error(t, err)
}
