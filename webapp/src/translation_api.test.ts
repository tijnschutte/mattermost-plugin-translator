import {fetchTranslation} from './translation_api';

function mockFetch(status: number, body?: object) {
    const fn = jest.fn().mockResolvedValue({status, json: async () => body});
    global.fetch = fn as unknown as typeof fetch;
    return fn;
}

describe('fetchTranslation', () => {
    it('returns the translation and sends the cookie header', async () => {
        const fn = mockFetch(200, {text: 'Bom dia', detected_language: 'nl'});

        const result = await fetchTranslation('http://mm', 'plugin', 'post1');

        expect(result).toEqual({text: 'Bom dia', detected_language: 'nl'});
        expect(fn).toHaveBeenCalledWith('http://mm/plugins/plugin/api/v1/posts/post1/translation', expect.objectContaining({
            credentials: 'same-origin',
            headers: {'X-Requested-With': 'XMLHttpRequest'},
        }));
    });

    it('returns null for an empty translation', async () => {
        mockFetch(200, {text: '', detected_language: 'pt'});

        expect(await fetchTranslation('', 'plugin', 'post1')).toBeNull();
    });

    it.each([204, 403, 502, 503])('returns null for status %d', async (status) => {
        mockFetch(status);

        expect(await fetchTranslation('', 'plugin', 'post1')).toBeNull();
    });
});
