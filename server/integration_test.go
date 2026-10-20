package main

import (
	"context"
	"net/http"
	"testing"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/pluginapi/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTranslationEndpoint runs the real request path: server auth, then the plugin's
// ServeHTTP, against a real Mattermost. No translation service is configured, so a
// request that passes every check answers 503.
func TestTranslationEndpoint(t *testing.T) {
	th := testhelper.Setup(t)
	ctx := context.Background()

	colleague := th.User
	author := th.CreateUser()
	post := th.PostAs(author, th.Channel.Id, "Goedemorgen allemaal")
	_, _, err := th.AdminClient.AddChannelMember(ctx, th.Channel.Id, colleague.Id)
	require.NoError(t, err)

	private := th.CreateChannel(model.ChannelTypePrivate)
	secret := th.PostAs(author, private.Id, "Geheim")

	status := func(token, postID string) int {
		t.Helper()
		url := th.ServerURL + "/plugins/" + testhelper.PluginID() + "/api/v1/posts/" + postID + "/translation"
		req, err := http.NewRequest(http.MethodGet, url, nil)
		require.NoError(t, err)
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		_ = resp.Body.Close()
		return resp.StatusCode
	}
	setLanguage := func(value string) {
		t.Helper()
		_, err := th.Client.UpdatePreferences(ctx, colleague.Id, model.Preferences{{
			UserId:   colleague.Id,
			Category: preferenceCategory(testhelper.PluginID()),
			Name:     languagePreferenceName,
			Value:    value,
		}})
		require.NoError(t, err)
	}

	assert.Equal(t, http.StatusUnauthorized, status("", post.Id), "no login")
	assert.Equal(t, http.StatusBadRequest, status(th.Client.AuthToken, "not-an-id"), "bad post ID")
	assert.Equal(t, http.StatusNotFound, status(th.Client.AuthToken, model.NewId()), "unknown post")
	assert.Equal(t, http.StatusNoContent, status(th.Client.AuthToken, post.Id), "no language set")

	setLanguage(languageOff)
	assert.Equal(t, http.StatusNoContent, status(th.Client.AuthToken, post.Id), "language off")

	setLanguage("pt-pt")
	assert.Equal(t, http.StatusForbidden, status(th.Client.AuthToken, secret.Id), "private channel the user is not in")
	assert.Equal(t, http.StatusServiceUnavailable, status(th.Client.AuthToken, post.Id), "every check passes, no service")

	own := th.PostAs(colleague, th.Channel.Id, "Bom dia")
	assert.Equal(t, http.StatusNoContent, status(th.Client.AuthToken, own.Id), "own post")
}
