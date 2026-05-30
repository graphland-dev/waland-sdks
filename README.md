# Waland SDK

Official client libraries for the [Waland WhatsApp API](https://api.waland.dev).

## Packages

| Language | Package | Path |
|----------|---------|------|
| Node.js | `waland` | [`packages/node`](./packages/node) |
| PHP | `waland/waland` | [`packages/php`](./packages/php) |

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

See [`packages/node/README.md`](./packages/node/README.md) and [`packages/php/README.md`](./packages/php/README.md) for full API docs.
