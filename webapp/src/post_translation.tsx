import manifest from 'manifest';
import React, {useEffect, useState} from 'react';
import {useSelector} from 'react-redux';

import type {GlobalState} from '@mattermost/types/store';

import {LANGUAGE_OFF, detectedLanguageName} from './languages';
import type {Translation} from './translation_api';
import {fetchTranslation} from './translation_api';
import {LANGUAGE_SETTING} from './user_settings';

// The web app stores plugin user settings under "pp_" plus the plugin ID, cut to 32
// characters. The server mirrors this in preferenceCategory.
const PREFERENCE_KEY = `${`pp_${manifest.id}`.slice(0, 32)}--${LANGUAGE_SETTING}`;

type Props = {
    postId: string;
};

type PostUtils = {
    formatText: (text: string, options?: object) => string;
    messageHtmlToComponent: (html: string, options?: object) => React.ReactNode;
};

// renderMessage formats the translation as Markdown, the same way Mattermost formats posts.
function renderMessage(text: string): React.ReactNode {
    const utils = (window as unknown as {PostUtils?: PostUtils}).PostUtils;
    if (!utils) {
        return <span style={{whiteSpace: 'pre-wrap'}}>{text}</span>;
    }
    return utils.messageHtmlToComponent(utils.formatText(text, {}), {});
}

// PostTranslation shows the post translated into the user's language, under the post.
export default function PostTranslation({postId}: Props) {
    const post = useSelector((state: GlobalState) => state.entities.posts.posts[postId]);
    const currentUserId = useSelector((state: GlobalState) => state.entities.users.currentUserId);
    const language = useSelector((state: GlobalState) => state.entities.preferences.myPreferences[PREFERENCE_KEY]?.value ?? LANGUAGE_OFF);
    const siteURL = useSelector((state: GlobalState) => state.entities.general.config.SiteURL ?? '');

    const [translation, setTranslation] = useState<Translation | null>(null);
    const [hidden, setHidden] = useState(false);

    // Ask only when the server can have something to show for this user.
    const wanted = Boolean(post) && language !== LANGUAGE_OFF && post.user_id !== currentUserId && !post.type && post.message.trim() !== '';

    useEffect(() => {
        setTranslation(null);
        if (!wanted) {
            return undefined;
        }
        let cancelled = false;
        fetchTranslation(siteURL, manifest.id, postId).then(
            (result) => !cancelled && setTranslation(result),
            () => !cancelled && setTranslation(null),
        );
        return () => {
            cancelled = true;
        };
    }, [wanted, siteURL, postId, post?.edit_at, language]);

    if (!translation) {
        return null;
    }

    const label = `Translated from ${detectedLanguageName(translation.detected_language)}`;
    return (
        <div
            className='translator-post-translation'
            style={{marginTop: 4, paddingLeft: 8, borderLeft: '3px solid rgba(var(--center-channel-color-rgb), 0.16)'}}
        >
            <div style={{fontSize: 12, color: 'rgba(var(--center-channel-color-rgb), 0.64)'}}>
                {label}
                {' · '}
                <button
                    type='button'
                    className='style--none color--link'
                    onClick={() => setHidden(!hidden)}
                >
                    {hidden ? 'Show' : 'Hide'}
                </button>
            </div>
            {!hidden && <div className='post-message__text'>{renderMessage(translation.text)}</div>}
        </div>
    );
}
