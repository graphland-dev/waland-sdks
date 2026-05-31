# waland (Go)

Official Go SDK for the [Waland API](https://api.waland.dev). Send WhatsApp messages with your organization API key and session.

Requires **Go 1.22+**.

## Install

```bash
go get github.com/graphland-dev/waland-sdks/packages/go
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

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    waland "github.com/graphland-dev/waland-sdks/packages/go"
)

func main() {
    client, err := waland.NewClient(
        os.Getenv("WALAND_API_KEY"),
        os.Getenv("WALAND_SESSION_ID"),
        nil,
    )
    if err != nil {
        log.Fatal(err)
    }

    result, err := client.SendMessage(context.Background(), waland.SendMessageParams{
        ChatID: "8801712345678@s.whatsapp.net",
        Text:   "Hello from Waland",
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(result.Status, result.MessageID)
}
```

## Constructor

```go
waland.NewClient(apiKey, sessionID, options)
```

| Argument | Description |
|----------|-------------|
| `apiKey` | Organization API key (`waland_...`) from the [merchant console](https://console.waland.dev/merchant) |
| `sessionID` | Session ID from **WhatsApp accounts** in the [merchant console](https://console.waland.dev/merchant) (after connecting via QR code) |
| `options.BaseURL` | API base URL (default: `https://api.waland.dev`) |
| `options.Timeout` | Request timeout (default: `30s`) |

## `SendMessage(ctx, params)`

| Field | Required | Description |
|-------|----------|-------------|
| `ChatID` | Yes | WhatsApp JID: `{number}@s.whatsapp.net` (DM) or `{groupId}@g.us` (group) |
| `Text` | No* | Message body or media caption |
| `MediaURL` | No* | Public HTTPS URL of media to send |
| `MediaFilename` | No | Filename override for downloaded media |

\* At least one of `Text` or `MediaURL` must be provided.

## Errors

- `*waland.ValidationError` — invalid arguments before the request is sent
- `*waland.APIError` — API returned a non-success status (`StatusCode`, `Message`, `Body`)

## Multiple sessions

Use one client per session (same API key is fine):

```go
support, _ := waland.NewClient(apiKey, supportSessionID, nil)
alerts, _ := waland.NewClient(apiKey, alertsSessionID, nil)
```
