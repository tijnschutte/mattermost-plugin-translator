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

// languageName returns the display name for a code such as "nl" or "pt-PT", or the code
// itself when the list does not have it.
export function languageName(code: string): string {
    const lower = code.toLowerCase();
    const exact = LANGUAGES.find((l) => l.code === lower);
    if (exact) {
        return exact.name;
    }
    const base = lower.split('-')[0];
    return LANGUAGES.find((l) => l.code === base)?.name ?? code;
}
