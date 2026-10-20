package main

import (
	"fmt"
	"time"

	"github.com/mattermost/mattermost/server/public/pluginapi"
)

// cacheTTL bounds how long a translation stays in the KV store.
const cacheTTL = 30 * 24 * time.Hour

// cachedTranslation is a stored translation. An empty Text means the post was already
// in the target language, or had nothing to translate.
type cachedTranslation struct {
	Text             string `json:"text"`
	DetectedLanguage string `json:"detected_language"`
}

// translationCache stores one translation per post version and target language.
type translationCache interface {
	get(key string) (*cachedTranslation, error)
	set(key string, value cachedTranslation) error
}

// cacheKey includes the post's EditAt, so an edited post gets a new key and the old
// entry expires on its own.
func cacheKey(postID string, editAt int64, language string) string {
	return fmt.Sprintf("tr:%s:%d:%s", postID, editAt, language)
}

type kvCache struct {
	kv *pluginapi.KVService
}

func (c kvCache) get(key string) (*cachedTranslation, error) {
	var value *cachedTranslation
	if err := c.kv.Get(key, &value); err != nil {
		return nil, fmt.Errorf("read translation %s: %w", key, err)
	}
	return value, nil
}

func (c kvCache) set(key string, value cachedTranslation) error {
	if _, err := c.kv.Set(key, value, pluginapi.SetExpiry(cacheTTL)); err != nil {
		return fmt.Errorf("store translation %s: %w", key, err)
	}
	return nil
}
