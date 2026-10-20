import type {PluginConfiguration} from 'types/mattermost-webapp';

import {LANGUAGE_OFF, LANGUAGES} from './languages';

// LANGUAGE_SETTING is the preference name. The server reads the same name.
export const LANGUAGE_SETTING = 'language';

export function userSettings(pluginId: string, uiName: string): PluginConfiguration {
    return {
        id: pluginId,
        uiName,
        sections: [{
            title: 'Translate posts',
            settings: [{
                name: LANGUAGE_SETTING,
                type: 'radio',
                title: 'Translate posts into',
                helpText: 'Under each post in another language, show a translation. Only you see it. The original post does not change.',
                default: LANGUAGE_OFF,
                options: [
                    {value: LANGUAGE_OFF, text: 'Off'},
                    ...LANGUAGES.map((l) => ({value: l.code, text: l.name})),
                ],
            }],
        }],
    };
}
