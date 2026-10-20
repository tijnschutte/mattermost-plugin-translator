package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestConfigurationNewTranslator(t *testing.T) {
	assert.Nil(t, (&configuration{AzureKey: "  "}).newTranslator())
	assert.NotNil(t, (&configuration{AzureKey: "key", AzureRegion: "westeurope"}).newTranslator())
}
