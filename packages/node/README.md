# waland

Official Node.js SDK for the [Waland API](https://api.waland.dev). Send WhatsApp messages with your organization API key and session.

Requires **Node.js 18+**

## Install

```bash
npm install waland
```

## API key and session ID

Open the [Waland merchant console](https://console.waland.dev/merchant) and set up both credentials:

### API key

Create or copy your organization **API key** (starts with `waland_`) from the console.

### Session ID

1. In the console, go to the **WhatsApp accounts** page.
2. Connect your WhatsApp account by scanning the **QR code** with the WhatsApp app on your phone.
3. After the account is connected, copy the **session ID** for that account.

Use that session ID when creating the client.

Keep your API key and session ID secret. Use environment variables in production (e.g. `WALAND_API_KEY`, `WALAND_SESSION_ID`).

## Quick start

```ts
import { WalandClient } from "waland";

const client = new WalandClient(
  process.env.WALAND_API_KEY!,
  process.env.WALAND_SESSION_ID!,
);

const result = await client.sendMessage({
  chatId: "8801712345678@s.whatsapp.net",
  text: "Hello from Waland",
});

console.log(result.status, result.messageId);
```

## Constructor

```ts
new WalandClient(apiKey, sessionId, options?)
```


| Argument            | Description                                                                                                                         |
| ------------------- | ----------------------------------------------------------------------------------------------------------------------------------- |
| `apiKey`            | Organization API key (`waland_...`) from the [merchant console](https://console.waland.dev/merchant)                                |
| `sessionId`         | Session ID from **WhatsApp accounts** in the [merchant console](https://console.waland.dev/merchant) (after connecting via QR code) |
| `options.baseUrl`   | API base URL (default: `https://api.waland.dev`)                                                                                    |
| `options.timeoutMs` | Request timeout in ms (default: `30000`)                                                                                            |


## `sendMessage(params)`


| Field           | Required | Description                                                              |
| --------------- | -------- | ------------------------------------------------------------------------ |
| `chatId`        | Yes      | WhatsApp JID: `{number}@s.whatsapp.net` (DM) or `{groupId}@g.us` (group) |
| `text`          | No*      | Message body or media caption                                            |
| `mediaUrl`      | No*      | Public HTTPS URL of media to send                                        |
| `mediaFilename` | No       | Filename override for downloaded media                                   |


 At least one of `text` or `mediaUrl` must be provided.

### Media example

```ts
await client.sendMessage({
  chatId: "8801712345678@s.whatsapp.net",
  text: "Check this out",
  mediaUrl: "https://example.com/photo.jpg",
  mediaFilename: "photo.jpg",
});
```

## `checkNumber(params)`

Checks whether a phone number is available on WhatsApp for the current session.

| Field    | Required | Description                                  |
| -------- | -------- | -------------------------------------------- |
| `number` | Yes      | Phone number in international format (digits) |

```ts
const result = await client.checkNumber({ number: "8801712345678" });
console.log(result.exists ?? result.isWhatsApp ?? result.onWhatsApp);
```

## Errors

- `WalandValidationError` — invalid arguments before the request is sent
- `WalandError` — API returned a non-success status (`statusCode`, `message`, `body`)

```ts
import { WalandClient, WalandError } from "waland";

try {
  await client.sendMessage({ chatId: "...", text: "Hi" });
} catch (error) {
  if (error instanceof WalandError) {
    console.error(error.statusCode, error.message);
  }
}
```

## Multiple sessions

Use one client per session (same API key is fine):

```ts
const support = new WalandClient(apiKey, supportSessionId);
const alerts = new WalandClient(apiKey, alertsSessionId);
```


