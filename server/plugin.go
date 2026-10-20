package main

import (
	"sync"
	"sync/atomic"

	"github.com/gorilla/mux"
	"github.com/mattermost/mattermost/server/public/plugin"
	"github.com/mattermost/mattermost/server/public/pluginapi"
	"golang.org/x/sync/singleflight"
)

// Plugin implements the interface expected by the Mattermost server to communicate between the server and plugin processes.
type Plugin struct {
	plugin.MattermostPlugin

	// client is the Mattermost server API client.
	client *pluginapi.Client

	// cache stores translations in the plugin's KV store.
	cache translationCache

	// inflight joins concurrent requests for the same translation into one call.
	inflight singleflight.Group

	// translateSlots caps the parallel calls to the translation service.
	translateSlots chan struct{}

	// quotaPausedUntil holds the Unix time in nanoseconds until which the plugin makes
	// no calls, after the service reported its quota used up.
	quotaPausedUntil atomic.Int64

	// router is the HTTP router for handling API requests.
	router *mux.Router

	// configurationLock synchronizes access to the configuration.
	configurationLock sync.RWMutex

	// configuration is the active plugin configuration. Consult getConfiguration and
	// setConfiguration for usage.
	configuration *configuration
}

// OnActivate is invoked when the plugin is activated. If an error is returned, the plugin will be deactivated.
func (p *Plugin) OnActivate() error {
	p.client = pluginapi.NewClient(p.API, p.Driver)

	p.cache = kvCache{kv: &p.client.KV}

	p.translateSlots = make(chan struct{}, maxParallelTranslations)

	p.router = p.initRouter()

	return nil
}

// See https://developers.mattermost.com/extend/plugins/server/reference/
