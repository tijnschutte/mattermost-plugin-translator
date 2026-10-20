package main

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tijnschutte/mattermost-plugin-translator/server/translator"
)

type memoryCache struct {
	mu     sync.Mutex
	values map[string]cachedTranslation
}

func (c *memoryCache) get(key string) (*cachedTranslation, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if v, ok := c.values[key]; ok {
		return &v, nil
	}
	return nil, nil
}

func (c *memoryCache) set(key string, value cachedTranslation) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.values[key] = value
	return nil
}

// fakeTranslator answers with a fixed result and counts its calls. If release is set,
// each call waits for it to close.
type fakeTranslator struct {
	result  translator.Result
	err     error
	calls   atomic.Int32
	release chan struct{}
}

func (f *fakeTranslator) Translate(_ context.Context, texts []string, _ string) ([]translator.Result, error) {
	f.calls.Add(1)
	if f.release != nil {
		<-f.release
	}
	if f.err != nil {
		return nil, f.err
	}
	out := make([]translator.Result, len(texts))
	for i := range texts {
		out[i] = f.result
	}
	return out, nil
}

func newTranslationPlugin(t *testing.T, tr translator.Translator) (*Plugin, *memoryCache, *logRecorder) {
	t.Helper()
	p, api := newTestPlugin()
	cache := &memoryCache{values: map[string]cachedTranslation{}}
	p.cache = cache
	p.translateSlots = make(chan struct{}, maxParallelTranslations)
	p.setConfiguration(&configuration{translator: tr})
	return p, cache, api
}

var dutchPost = &model.Post{Id: "post1", EditAt: 0, Message: "Goedemorgen"}

func TestTranslatePostTranslatesAndCaches(t *testing.T) {
	fake := &fakeTranslator{result: translator.Result{Text: "Bom dia", DetectedLanguage: "nl"}}
	p, cache, _ := newTranslationPlugin(t, fake)

	got, err := p.translatePost(context.Background(), dutchPost, "pt-pt")
	require.NoError(t, err)
	again, err := p.translatePost(context.Background(), dutchPost, "pt-pt")
	require.NoError(t, err)

	want := cachedTranslation{Text: "Bom dia", DetectedLanguage: "nl"}
	assert.Equal(t, want, got)
	assert.Equal(t, want, again)
	assert.Equal(t, int32(1), fake.calls.Load())
	assert.Equal(t, want, cache.values[cacheKey("post1", 0, "pt-pt")])
}

func TestTranslatePostEditedPostIsTranslatedAgain(t *testing.T) {
	fake := &fakeTranslator{result: translator.Result{Text: "Bom dia", DetectedLanguage: "nl"}}
	p, _, _ := newTranslationPlugin(t, fake)
	edited := &model.Post{Id: dutchPost.Id, EditAt: 42, Message: dutchPost.Message}

	_, err := p.translatePost(context.Background(), dutchPost, "pt-pt")
	require.NoError(t, err)
	_, err = p.translatePost(context.Background(), edited, "pt-pt")
	require.NoError(t, err)

	assert.Equal(t, int32(2), fake.calls.Load())
}

func TestTranslatePostSameLanguageStoresEmptyResult(t *testing.T) {
	fake := &fakeTranslator{result: translator.Result{Text: "Bom dia", DetectedLanguage: "pt"}}
	p, _, _ := newTranslationPlugin(t, fake)

	got, err := p.translatePost(context.Background(), &model.Post{Id: "post2", Message: "Bom dia"}, "pt-pt")
	require.NoError(t, err)
	_, err = p.translatePost(context.Background(), &model.Post{Id: "post2", Message: "Bom dia"}, "pt-pt")
	require.NoError(t, err)

	assert.Equal(t, cachedTranslation{DetectedLanguage: "pt"}, got)
	assert.Equal(t, int32(1), fake.calls.Load())
}

func TestTranslatePostWithoutTranslator(t *testing.T) {
	p, _, _ := newTranslationPlugin(t, nil)

	_, err := p.translatePost(context.Background(), dutchPost, "pt-pt")

	assert.ErrorIs(t, err, errNoTranslator)
}

func TestTranslatePostQuotaPausesCalls(t *testing.T) {
	fake := &fakeTranslator{err: translator.ErrQuotaExceeded}
	p, cache, api := newTranslationPlugin(t, fake)

	_, err := p.translatePost(context.Background(), dutchPost, "pt-pt")
	assert.ErrorIs(t, err, translator.ErrQuotaExceeded)
	_, err = p.translatePost(context.Background(), &model.Post{Id: "post3", Message: "Hallo"}, "pt-pt")
	assert.ErrorIs(t, err, translator.ErrQuotaExceeded)

	assert.Equal(t, int32(1), fake.calls.Load())
	assert.Empty(t, cache.values)
	require.Len(t, api.entries, 1)
	assert.Equal(t, "warn", api.entries[0].level)
	assert.Greater(t, p.quotaPausedUntil.Load(), time.Now().Add(59*time.Minute).UnixNano())
}

func TestTranslatePostJoinsConcurrentRequests(t *testing.T) {
	fake := &fakeTranslator{
		result:  translator.Result{Text: "Bom dia", DetectedLanguage: "nl"},
		release: make(chan struct{}),
	}
	p, _, _ := newTranslationPlugin(t, fake)

	var wg sync.WaitGroup
	results := make([]cachedTranslation, 10)
	for i := range results {
		wg.Go(func() {
			got, err := p.translatePost(context.Background(), dutchPost, "pt-pt")
			assert.NoError(t, err)
			results[i] = got
		})
	}
	require.Eventually(t, func() bool { return fake.calls.Load() == 1 }, time.Second, time.Millisecond)
	close(fake.release)
	wg.Wait()

	assert.Equal(t, int32(1), fake.calls.Load())
	for _, got := range results {
		assert.Equal(t, "Bom dia", got.Text)
	}
}
