# highlighter

CLI with NIP-84 implement

## Setup

Create your config file in `~/.config/nostr/highlighter.json` containing:

```json
{
    "relays": ["wss://relay.highlighter.com/", "wss://relay.damus.io/"],
    "npub": "npub14ge829c4pvgx24c35qts3sv82wc2xwcmgng93tzp6d52k9de2xgqq0y4jk"
}
```

Finally, set your environment variable:

```shell
export NOSTR=~/.config/nostr/highlighter.json`
```

Build the executable

```shell
make build
```

## Usage

```shell
> highlighter -content "Some highlighted content" -context "Why did you make this highlight" -url "https://go.dev/blog/strings"
```
