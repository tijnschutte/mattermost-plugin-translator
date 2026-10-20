// The target languages a user can pick. The codes are Azure AI Translator codes.
export type Language = {
    code: string;
    name: string;
};

export const LANGUAGE_OFF = 'off';

// Only the languages the production LibreTranslate loads (LT_LOAD_ONLY=nl,pt,en in
// infra-lab). A language missing there gives the user no translations and no error.
export const LANGUAGES: Language[] = [
    {code: 'nl', name: 'Nederlands'},
    {code: 'pt-pt', name: 'Português'},
    {code: 'en', name: 'English'},
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
