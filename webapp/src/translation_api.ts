export type Translation = {
    text: string;
    detected_language: string;
};

// fetchTranslation asks the server for a post in the user's language. It returns null
// when there is nothing to show: no language set, the user's own post, a post already
// in the user's language, or the service is not available.
export async function fetchTranslation(siteURL: string, pluginId: string, postId: string): Promise<Translation | null> {
    const response = await fetch(`${siteURL}/plugins/${pluginId}/api/v1/posts/${encodeURIComponent(postId)}/translation`, {
        credentials: 'same-origin',

        // Mattermost accepts the session cookie only with this header.
        headers: {'X-Requested-With': 'XMLHttpRequest'},
    });
    if (response.status !== 200) {
        return null;
    }
    const translation: Translation = await response.json();
    return translation.text ? translation : null;
}
