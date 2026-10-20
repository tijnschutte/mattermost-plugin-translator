import {detectedLanguageName} from './languages';
import {userSettings} from './user_settings';

describe('detectedLanguageName', () => {
    it('names the base language only', () => {
        expect(detectedLanguageName('pt')).toBe('Portuguese');
        expect(detectedLanguageName('pt-PT')).toBe('Portuguese');
        expect(detectedLanguageName('nl')).toBe('Dutch');
    });

    it('returns an invalid code as is', () => {
        expect(detectedLanguageName('123')).toBe('123');
    });
});

describe('userSettings', () => {
    it('offers Off and the three languages, stored under the language name', () => {
        const section = userSettings('id', 'Translator').sections[0];
        const setting = 'settings' in section ? section.settings[0] : undefined;

        expect(setting?.name).toBe('language');
        expect(setting?.type).toBe('radio');
        expect(setting && 'options' in setting ? setting.options.map((o) => o.value) : []).toEqual(['off', 'nl', 'pt-pt', 'en']);
    });
});
