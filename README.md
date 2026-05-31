# Waland SDK

Official client libraries for the [Waland WhatsApp API](https://api.waland.dev).

## Packages

| Language | Package | Path |
|----------|---------|------|
| Node.js | `waland` | [`packages/node`](./packages/node) |
| PHP | `waland/waland` | [`packages/php`](./packages/php) |
| Go | `github.com/graphland-dev/waland-sdks/packages/go` | [`packages/go`](./packages/go) |
| Rust | `waland` | [`packages/rust`](./packages/rust) |

## Credentials

Open the [merchant console](https://console.waland.dev/merchant):

- **API key** — create or copy your organization API key (`waland_...`).
- **Session ID** — go to **WhatsApp accounts**, connect your number by scanning the **QR code**, then copy the session ID for that connected account.

## Node.js

```bash
npm install waland
```

```ts
import { WalandClient } from "waland";

const client = new WalandClient(
  process.env.WALAND_API_KEY!,
  process.env.WALAND_SESSION_ID!,
);

await client.sendMessage({
  chatId: "8801712345678@s.whatsapp.net",
  text: "Hello from Waland",
});
```

## PHP

```bash
composer require waland/waland
```

```php
use Waland\WalandClient;

$client = new WalandClient(
    getenv('WALAND_API_KEY'),
    getenv('WALAND_SESSION_ID'),
);

$client->sendMessage([
    'chatId' => '8801712345678@s.whatsapp.net',
    'text' => 'Hello from Waland',
]);
```

## Go

```bash
go get github.com/graphland-dev/waland-sdks/packages/go
```

```go
client, _ := waland.NewClient(
    os.Getenv("WALAND_API_KEY"),
    os.Getenv("WALAND_SESSION_ID"),
    nil,
)

client.SendMessage(context.Background(), waland.SendMessageParams{
    ChatID: "8801712345678@s.whatsapp.net",
    Text:   "Hello from Waland",
})
```

## Rust

```bash
cargo add waland
```

```rust
use waland::{SendMessageParams, WalandClient};

#[tokio::main]
async fn main() {
    let client = WalandClient::new(
        std::env::var("WALAND_API_KEY").unwrap(),
        std::env::var("WALAND_SESSION_ID").unwrap(),
        None,
    ).unwrap();

    let _ = client.send_message(SendMessageParams {
        chat_id: "8801712345678@s.whatsapp.net".to_string(),
        text: Some("Hello from Waland".to_string()),
        media_url: None,
        media_filename: None,
    }).await;
}
```

See [`packages/node/README.md`](./packages/node/README.md), [`packages/php/README.md`](./packages/php/README.md), [`packages/go/README.md`](./packages/go/README.md), and [`packages/rust/README.md`](./packages/rust/README.md) for full API docs.

### PHP development (monorepo)

Root [`composer.json`](./composer.json) is what Packagist indexes. [`packages/php/composer.json`](./packages/php/composer.json) is a copy for working in that directory only.

```bash
composer install   # repo root
composer test
```
