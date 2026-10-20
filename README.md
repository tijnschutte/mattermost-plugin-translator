# Translator for Mattermost

Translator shows each post in the language that each user picks. A Portuguese user reads
the Dutch team's posts in Portuguese, and the Dutch team reads the Portuguese user's posts
in Dutch. The original post does not change, and each user sees only their own
translation.

The plugin works with one of two translation services:

- [Azure AI Translator](https://azure.microsoft.com/products/ai-services/ai-translator)
  gives the best quality. Its free F0 tier covers 2,000,000 characters per month, but
  Azure asks for a credit card to sign up.
- [LibreTranslate](https://github.com/LibreTranslate/LibreTranslate) runs on your own
  server, so no text leaves it, and it needs no account. Its quality is lower, and it
  often translates technical words literally.

## How it works

- Each user picks a language in **Settings** > **Plugin Preferences** > **Translator**.
  The default is **Off**.
- When the web app shows a post in another language, it asks the plugin for a
  translation and shows it under the post, with a "Translated from" label and a **Hide**
  button.
- The plugin translates a post only when a user with a language setting views it. It
  translates each post version once per language and keeps the result for 30 days.
- Code blocks and inline code stay untranslated.
- The plugin translates a post only for users who can read its channel.

## Install the plugin

To install the plugin, you need a Mattermost server, version 10.12.0 or later, with
plugin uploads turned on.

1. Download the `.tar.gz` file of the latest release from the
   [releases page](https://github.com/tijnschutte/mattermost-plugin-translator/releases).
2. In the System Console, go to **Plugins** > **Plugin Management**.
3. Under **Upload Plugin**, upload the file. Then click **Enable** next to **Translator**.

## Connect a translation service

In the System Console, go to **Plugins** > **Translator**, and pick a **Translation
service**. Then follow the steps for that service.

### Azure AI Translator

1. In the [Azure portal](https://portal.azure.com), create a **Translator** resource.
   Choose the **Free F0** pricing tier. Azure asks for a credit card, but an F0 resource
   does not charge it.
2. In the resource, open **Keys and Endpoint**. Copy **Key 1** and **Location/Region**.
3. In the plugin settings, paste the key into **Azure key**, and the region into
   **Azure region**. Click **Save**.

When the free quota is used up, Azure refuses more text until the next month. The plugin
then shows no translations, logs one warning, and tries again an hour later.

### LibreTranslate

1. Run LibreTranslate where the Mattermost server can reach it. To load only the
   languages you need, set `LT_LOAD_ONLY`, for example to `nl,pt,en`. With those three
   languages, it used about 1.3 GB of memory after a few translations.
2. In the plugin settings, enter its address in **LibreTranslate URL**, for example
   `http://libretranslate:5000`. Click **Save**.

The settings window offers Dutch, Portuguese, and English, the languages LibreTranslate
loads in production. To offer another language, load its model in LibreTranslate first,
then add it to `webapp/src/languages.ts`.

## Privacy

With Azure AI Translator, the plugin sends the text of a post to Azure when a user with a
language setting views it. Check that this is allowed under your organization's data
rules before you turn the plugin on. With LibreTranslate, the text goes only to the
server you run.

## Known limitations

- The Mattermost mobile app does not run web app plugins, so mobile users see no
  translations.
- The "Translated from" label and the **Hide** button are in English.
- Mentions, links, and Markdown other than code can change in translation.

## Development

For the commands, the local Mattermost, the tests, and the design rules, see
[`AGENTS.md`](AGENTS.md). For the design decisions, see [`docs/PLAN.md`](docs/PLAN.md).

## License

This project started from the
[Mattermost plugin starter template](https://github.com/mattermost/mattermost-plugin-starter-template).
For the license, see [`LICENSE`](LICENSE).
