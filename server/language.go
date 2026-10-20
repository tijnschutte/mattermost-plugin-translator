package main

import (
	"regexp"
	"strings"
)

// languagePreferenceName is the name the web app's user setting stores the target
// language under. The category comes from preferenceCategory.
const languagePreferenceName = "language"

// languageOff is the setting value for "no translations".
const languageOff = "off"

// languageCodePattern accepts codes such as "nl", "pt", and "pt-pt".
var languageCodePattern = regexp.MustCompile(`^[a-z]{2,3}(-[a-z0-9]{2,8})?$`)

// normalizeLanguage returns the lower-case language code, or "" for no translation.
func normalizeLanguage(value string) string {
	code := strings.ToLower(strings.TrimSpace(value))
	if code == languageOff || !languageCodePattern.MatchString(code) {
		return ""
	}
	return code
}

// targetLanguage returns the language the user wants posts translated into, or "" when
// the user has not chosen one.
//
// It reads all of the user's preferences, because GetPreferenceForUser answers 400 both
// for a missing preference and for a database error.
func (p *Plugin) targetLanguage(userID string) (string, error) {
	prefs, appErr := p.API.GetPreferencesForUser(userID)
	if appErr != nil {
		return "", appErr
	}
	category := preferenceCategory(manifest.Id)
	for _, pref := range prefs {
		if pref.Category == category && pref.Name == languagePreferenceName {
			return normalizeLanguage(pref.Value), nil
		}
	}
	return "", nil
}

// preferenceCategory mirrors getPluginPreferenceKey in the Mattermost web app, which
// stores a plugin's user settings under "pp_" plus the plugin ID, cut to 32 characters.
func preferenceCategory(pluginID string) string {
	category := "pp_" + pluginID
	if len(category) > 32 {
		category = category[:32]
	}
	return category
}
