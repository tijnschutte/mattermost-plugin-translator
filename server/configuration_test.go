package main

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tijnschutte/mattermost-plugin-translator/server/translator"
)

func TestConfigurationValidate(t *testing.T) {
	tests := []struct {
		endpoint string
		valid    bool
	}{
		{"", true},
		{"https://api.cognitive.microsofttranslator.com", true},
		{"https://api-eur.cognitive.microsofttranslator.com/", true},
		{"http://api.cognitive.microsofttranslator.com", false},
		{"api.cognitive.microsofttranslator.com", false},
		{"https://", false},
	}
	for _, tt := range tests {
		t.Run(tt.endpoint, func(t *testing.T) {
			err := (&configuration{AzureEndpoint: tt.endpoint}).validate()
			assert.Equal(t, tt.valid, err == nil, "error: %v", err)
		})
	}
}

func TestConfigurationValidateLibreTranslate(t *testing.T) {
	tests := []struct {
		url   string
		valid bool
	}{
		{"", true},
		{"http://libretranslate:5000", true},
		{"https://translate.example.com", true},
		{"ftp://libretranslate", false},
		{"libretranslate:5000", false},
	}
	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			err := (&configuration{Service: serviceLibreTranslate, LibreTranslateURL: tt.url}).validate()
			assert.Equal(t, tt.valid, err == nil, "error: %v", err)
		})
	}
	assert.Error(t, (&configuration{Service: "deepl"}).validate())
}

func TestConfigurationNewTranslator(t *testing.T) {
	assert.Nil(t, (&configuration{AzureKey: "  "}).newTranslator())
	assert.IsType(t, &translator.Azure{}, (&configuration{AzureKey: "key", AzureRegion: "westeurope"}).newTranslator())

	// Each service uses only its own settings.
	assert.Nil(t, (&configuration{Service: serviceLibreTranslate, AzureKey: "key"}).newTranslator())
	assert.IsType(t, &translator.LibreTranslate{}, (&configuration{Service: serviceLibreTranslate, LibreTranslateURL: "http://libretranslate:5000"}).newTranslator())
	assert.Nil(t, (&configuration{Service: serviceAzure, LibreTranslateURL: "http://libretranslate:5000"}).newTranslator())
}
