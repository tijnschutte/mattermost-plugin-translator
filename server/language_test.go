package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeLanguage(t *testing.T) {
	tests := []struct{ in, want string }{
		{"pt-pt", "pt-pt"},
		{" PT-PT ", "pt-pt"},
		{"nl", "nl"},
		{"off", ""},
		{"", ""},
		{"pt_PT", ""},
		{"dutch", ""},
		{"nl&to=en", ""},
		{"zh-Hans", "zh-hans"},
		{"x", ""},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, normalizeLanguage(tt.in), "input %q", tt.in)
	}
}

func TestCacheKey(t *testing.T) {
	assert.Equal(t, "tr:abc:0:pt-pt", cacheKey("abc", 0, "pt-pt"))
	assert.NotEqual(t, cacheKey("abc", 0, "pt-pt"), cacheKey("abc", 1, "pt-pt"))
	// Mattermost limits KV keys to 150 characters. A post ID has 26.
	assert.LessOrEqual(t, len(cacheKey("abcdefghijklmnopqrstuvwxyz", 1<<62, "zh-hans")), 150)
}

func TestPreferenceCategory(t *testing.T) {
	assert.Equal(t, "pp_nl.tijnschutte.translator", preferenceCategory("nl.tijnschutte.translator"))
	assert.Equal(t, "pp_com.example.a-very-long-plugi", preferenceCategory("com.example.a-very-long-plugin-id"))
}
