package main

import (
	"net/url"
	"reflect"
	"strings"

	"github.com/pkg/errors"

	"github.com/tijnschutte/mattermost-plugin-translator/server/translator"
)

// configuration captures the plugin's external configuration as exposed in the Mattermost server
// configuration, as well as values computed from the configuration. Any public fields will be
// deserialized from the Mattermost server configuration in OnConfigurationChange.
//
// As plugins are inherently concurrent (hooks being called asynchronously), and the plugin
// configuration can change at any time, access to the configuration must be synchronized. The
// strategy used in this plugin is to guard a pointer to the configuration, and clone the entire
// struct whenever it changes. You may replace this with whatever strategy you choose.
//
// If you add non-reference types to your configuration struct, be sure to rewrite Clone as a deep
// copy appropriate for your types.
type configuration struct {
	Service string

	AzureKey      string
	AzureRegion   string
	AzureEndpoint string

	LibreTranslateURL    string
	LibreTranslateAPIKey string

	// translator is built from the fields above. It is nil when the chosen service is
	// not set up.
	translator translator.Translator
}

const (
	serviceAzure          = "azure"
	serviceLibreTranslate = "libretranslate"
)

// service returns the chosen service. An empty value means Azure, the default.
func (c *configuration) service() string {
	if c.Service == "" {
		return serviceAzure
	}
	return c.Service
}

// validate rejects settings that can never work.
func (c *configuration) validate() error {
	switch c.service() {
	case serviceAzure:
		if c.AzureEndpoint == "" {
			return nil
		}
		u, err := url.Parse(c.AzureEndpoint)
		if err != nil || u.Scheme != "https" || u.Host == "" {
			return errors.Errorf("Azure endpoint %q must be an https URL", c.AzureEndpoint)
		}
	case serviceLibreTranslate:
		if c.LibreTranslateURL == "" {
			return nil
		}
		// A server in the same cluster or Docker network is often reached over plain http.
		u, err := url.Parse(c.LibreTranslateURL)
		if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" {
			return errors.Errorf("LibreTranslate URL %q must be an http or https URL", c.LibreTranslateURL)
		}
	default:
		return errors.Errorf("unknown translation service %q", c.Service)
	}
	return nil
}

// newTranslator returns the translator these settings describe, or nil when the chosen
// service is not set up.
func (c *configuration) newTranslator() translator.Translator {
	switch c.service() {
	case serviceAzure:
		key := strings.TrimSpace(c.AzureKey)
		if key == "" {
			return nil
		}
		return translator.NewAzure(strings.TrimRight(c.AzureEndpoint, "/"), key, strings.TrimSpace(c.AzureRegion))
	case serviceLibreTranslate:
		u := strings.TrimSpace(c.LibreTranslateURL)
		if u == "" {
			return nil
		}
		return translator.NewLibreTranslate(u, strings.TrimSpace(c.LibreTranslateAPIKey))
	}
	return nil
}

// Clone shallow copies the configuration. Your implementation may require a deep copy if
// your configuration has reference types.
func (c *configuration) Clone() *configuration {
	clone := *c
	return &clone
}

// getConfiguration retrieves the active configuration under lock, making it safe to use
// concurrently. The active configuration may change underneath the client of this method, but
// the struct returned by this API call is considered immutable.
func (p *Plugin) getConfiguration() *configuration {
	p.configurationLock.RLock()
	defer p.configurationLock.RUnlock()

	if p.configuration == nil {
		return &configuration{}
	}

	return p.configuration
}

// setConfiguration replaces the active configuration under lock.
//
// Do not call setConfiguration while holding the configurationLock, as sync.Mutex is not
// reentrant. In particular, avoid using the plugin API entirely, as this may in turn trigger a
// hook back into the plugin. If that hook attempts to acquire this lock, a deadlock may occur.
//
// This method panics if setConfiguration is called with the existing configuration. This almost
// certainly means that the configuration was modified without being cloned and may result in
// an unsafe access.
func (p *Plugin) setConfiguration(configuration *configuration) {
	p.configurationLock.Lock()
	defer p.configurationLock.Unlock()

	if configuration != nil && p.configuration == configuration {
		// Ignore assignment if the configuration struct is empty. Go will optimize the
		// allocation for same to point at the same memory address, breaking the check
		// above.
		if reflect.ValueOf(*configuration).NumField() == 0 {
			return
		}

		panic("setConfiguration called with the existing configuration")
	}

	p.configuration = configuration
}

// OnConfigurationChange is invoked when configuration changes may have been made.
func (p *Plugin) OnConfigurationChange() error {
	configuration := new(configuration)

	// Load the public configuration fields from the Mattermost server configuration.
	if err := p.API.LoadPluginConfiguration(configuration); err != nil {
		return errors.Wrap(err, "failed to load plugin configuration")
	}

	if err := configuration.validate(); err != nil {
		return errors.Wrap(err, "invalid plugin configuration")
	}
	configuration.translator = configuration.newTranslator()
	if configuration.translator == nil {
		p.API.LogWarn("The translation service is not set up, so the plugin translates nothing", "service", configuration.service())
	}

	p.setConfiguration(configuration)

	return nil
}
