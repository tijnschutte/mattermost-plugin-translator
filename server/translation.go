package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/mattermost/mattermost/server/public/model"

	"github.com/tijnschutte/mattermost-plugin-translator/server/translator"
)

const (
	// maxParallelTranslations caps the calls to the translation service at one time.
	// Opening a channel can ask for dozens of translations at once.
	maxParallelTranslations = 4

	// translateTimeout bounds one call to the translation service.
	translateTimeout = 15 * time.Second

	// quotaPause is how long the plugin stops calling the service after a quota error.
	quotaPause = time.Hour
)

// errNoTranslator means no translation service is configured.
var errNoTranslator = errors.New("no translation service is configured")

// translatePost returns the post's message in the target language. It reads the cache
// first. Concurrent requests for the same post and language share one service call.
func (p *Plugin) translatePost(ctx context.Context, post *model.Post, language string) (cachedTranslation, error) {
	key := cacheKey(post.Id, post.EditAt, language)

	cached, err := p.cache.get(key)
	if err != nil {
		return cachedTranslation{}, err
	}
	if cached != nil {
		return *cached, nil
	}

	value, err, _ := p.inflight.Do(key, func() (any, error) {
		// The first caller's request may end before the others, so do not tie the
		// shared call to its context.
		shared, cancel := context.WithTimeout(context.WithoutCancel(ctx), translateTimeout)
		defer cancel()
		return p.translateAndStore(shared, key, post.Message, language)
	})
	if err != nil {
		return cachedTranslation{}, err
	}
	return value.(cachedTranslation), nil
}

func (p *Plugin) translateAndStore(ctx context.Context, key, message, language string) (cachedTranslation, error) {
	tr := p.getConfiguration().translator
	if tr == nil {
		return cachedTranslation{}, errNoTranslator
	}
	if time.Now().UnixNano() < p.quotaPausedUntil.Load() {
		return cachedTranslation{}, translator.ErrQuotaExceeded
	}

	select {
	case p.translateSlots <- struct{}{}:
		defer func() { <-p.translateSlots }()
	case <-ctx.Done():
		return cachedTranslation{}, ctx.Err()
	}

	text, detected, err := translator.TranslateMessage(ctx, tr, message, language)
	if errors.Is(err, translator.ErrQuotaExceeded) {
		p.quotaPausedUntil.Store(time.Now().Add(quotaPause).UnixNano())
		p.API.LogWarn("The translation quota is used up. Translations stop for one hour.")
		return cachedTranslation{}, err
	}
	if err != nil {
		return cachedTranslation{}, fmt.Errorf("translate: %w", err)
	}

	// No detected language means there was no prose; the post needs no translation.
	result := cachedTranslation{Text: text, DetectedLanguage: detected}
	if detected == "" || translator.SameLanguage(detected, language) {
		result.Text = ""
	}
	if err := p.cache.set(key, result); err != nil {
		return cachedTranslation{}, err
	}
	return result, nil
}
