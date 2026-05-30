# waland (PHP)

Official PHP SDK for the [Waland API](https://api.waland.dev). Send WhatsApp messages with your organization API key and session.

Requires **PHP 8.1+** with `ext-curl` and `ext-json`.

## Install

```bash
composer require waland/waland
```

## API key and session ID

Open the [Waland merchant console](https://console.waland.dev/merchant) and set up both credentials:

### API key

Create or copy your organization **API key** (starts with `waland_`) from the console.

### Session ID

1. In the console, go to the **WhatsApp accounts** page.
2. Connect your WhatsApp account by scanning the **QR code** with the WhatsApp app on your phone.
3. After the account is connected, copy the **session ID** for that account.

Keep your API key and session ID secret. Use environment variables in production.

## Quick start

```php
<?php

require 'vendor/autoload.php';

use Waland\WalandClient;

$client = new WalandClient(
    getenv('WALAND_API_KEY'),
    getenv('WALAND_SESSION_ID'),
);

$result = $client->sendMessage([
    'chatId' => '8801712345678@s.whatsapp.net',
    'text' => 'Hello from Waland',
]);

echo $result['status'], ' ', $result['messageId'];
```

## Constructor

```php
new WalandClient($apiKey, $sessionId, $options = [])
```

| Argument | Description |
|----------|-------------|
| `$apiKey` | Organization API key (`waland_...`) from the [merchant console](https://console.waland.dev/merchant) |
| `$sessionId` | Session ID from **WhatsApp accounts** in the [merchant console](https://console.waland.dev/merchant) (after connecting via QR code) |
| `$options['baseUrl']` | API base URL (default: `https://api.waland.dev`) |
| `$options['timeoutSeconds']` | Request timeout in seconds (default: `30`) |

## `sendMessage($params)`

| Field | Required | Description |
|-------|----------|-------------|
| `chatId` | Yes | WhatsApp JID: `{number}@s.whatsapp.net` (DM) or `{groupId}@g.us` (group) |
| `text` | No* | Message body or media caption |
| `mediaUrl` | No* | Public HTTPS URL of media to send |
| `mediaFilename` | No | Filename override for downloaded media |

\* At least one of `text` or `mediaUrl` must be provided.

### Media example

```php
$client->sendMessage([
    'chatId' => '8801712345678@s.whatsapp.net',
    'text' => 'Check this out',
    'mediaUrl' => 'https://example.com/photo.jpg',
    'mediaFilename' => 'photo.jpg',
]);
```

## Errors

- `Waland\WalandValidationException` — invalid arguments before the request is sent
- `Waland\WalandException` — API returned a non-success status (`statusCode`, `getMessage()`, `body`)

```php
use Waland\WalandClient;
use Waland\WalandException;

try {
    $client->sendMessage(['chatId' => '...', 'text' => 'Hi']);
} catch (WalandException $e) {
    echo $e->statusCode, ': ', $e->getMessage();
}
```

## Multiple sessions

Use one client per session (same API key is fine):

```php
$support = new WalandClient($apiKey, $supportSessionId);
$alerts = new WalandClient($apiKey, $alertsSessionId);
```

## Development

**Packagist** uses the repository root [`composer.json`](../../composer.json). A copy for this folder lives at [`composer.json`](./composer.json) (paths relative to `packages/php/`).

From the **repo root** (recommended):

```bash
composer install
composer test
```

Or only this package directory:

```bash
cd packages/php
composer install
composer test
```
