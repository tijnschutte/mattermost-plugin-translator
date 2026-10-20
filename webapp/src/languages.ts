// The target languages a user can pick. The codes are Azure AI Translator codes.
export type Language = {
    code: string;
    name: string;
};

export const LANGUAGE_OFF = 'off';

export const LANGUAGES: Language[] = [
    {code: 'nl', name: 'Nederlands'},
    {code: 'pt-pt', name: 'Português (Portugal)'},
    {code: 'pt', name: 'Português (Brasil)'},
    {code: 'en', name: 'English'},
    {code: 'de', name: 'Deutsch'},
    {code: 'fr', name: 'Français'},
    {code: 'es', name: 'Español'},
];

const displayNames = new Intl.DisplayNames(['en'], {type: 'language'});

// detectedLanguageName names the language a service detected, such as "nl" or "pt-PT".
// It names only the base language, because services detect "pt" without a region.
// An unknown code comes back as is.
export function detectedLanguageName(code: string): string {
    const base = code.toLowerCase().split('-')[0];
    try {
        return displayNames.of(base) ?? code;
    } catch {
        return code;
    }
}
