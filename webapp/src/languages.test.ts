import {languageName} from './languages';
import {userSettings} from './user_settings';

describe('languageName', () => {
    it('finds a code in any case', () => {
        expect(languageName('pt-PT')).toBe('Português (Portugal)');
    });

    it('falls back to the base language', () => {
        expect(languageName('nl-BE')).toBe('Nederlands');
    });

    it('returns an unknown code as is', () => {
        expect(languageName('ja')).toBe('ja');
    });
});

describe('userSettings', () => {
    it('offers Off first and stores under the language name', () => {
        const section = userSettings('id', 'Translator').sections[0];
        const setting = 'settings' in section ? section.settings[0] : undefined;

        expect(setting?.name).toBe('language');
        expect(setting?.type).toBe('radio');
        expect(setting && 'options' in setting ? setting.options[0].value : '').toBe('off');
    });
});
