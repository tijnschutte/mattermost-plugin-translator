package main

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mattermost/mattermost/server/public/model"

	"github.com/tijnschutte/mattermost-plugin-translator/server/translator"
)

// translationResponse is the body of GET /api/v1/posts/{postId}/translation. An empty
// Text means there is nothing to show: the post is already in the user's language.
type translationResponse struct {
	Text             string `json:"text"`
	DetectedLanguage string `json:"detected_language"`
}

// handleGetTranslation returns the post translated into the requesting user's language.
// It answers 204 when there is nothing to translate for this user.
func (p *Plugin) handleGetTranslation(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("Mattermost-User-ID")
	postID := mux.Vars(r)["postId"]
	log := requestLogger(r)

	if !model.IsValidId(postID) {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	post, appErr := p.API.GetPost(postID)
	if appErr != nil || post.DeleteAt != 0 {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}
	if !p.API.HasPermissionToChannel(userID, post.ChannelId, model.PermissionReadChannel) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// System messages and the user's own posts need no translation.
	if post.Type != "" || post.UserId == userID {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	language, err := p.targetLanguage(userID)
	if err != nil {
		log.WithError(err).Errorf("Failed to read the language preference")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if language == "" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	result, err := p.translatePost(r.Context(), post, language)
	switch {
	case errors.Is(err, errNoTranslator), errors.Is(err, translator.ErrQuotaExceeded):
		http.Error(w, "Translation is not available", http.StatusServiceUnavailable)
		return
	case err != nil:
		log.WithError(err).Errorf("Failed to translate a post")
		http.Error(w, "Translation failed", http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(translationResponse(result)); err != nil {
		log.WithError(err).Errorf("Failed to write response")
	}
}
