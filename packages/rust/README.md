# waland (Rust)

Official Rust SDK for the [Waland API](https://api.waland.dev). Send WhatsApp messages with your organization API key and session.

## Install

```bash
cargo add waland
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

```rust
use waland::{SendMessageParams, WalandClient};

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let client = WalandClient::new(
        std::env::var("WALAND_API_KEY")?,
        std::env::var("WALAND_SESSION_ID")?,
        None,
    )?;

    let result = client
        .send_message(SendMessageParams {
            chat_id: "8801712345678@s.whatsapp.net".to_string(),
            text: Some("Hello from Waland".to_string()),
            media_url: None,
            media_filename: None,
        })
        .await?;

    println!("{:?} {:?}", result.status, result.message_id);
    Ok(())
}
```

## Constructor

```rust
WalandClient::new(api_key, session_id, options)
```

| Argument | Description |
|----------|-------------|
| `api_key` | Organization API key (`waland_...`) from the [merchant console](https://console.waland.dev/merchant) |
| `session_id` | Session ID from **WhatsApp accounts** in the [merchant console](https://console.waland.dev/merchant) (after connecting via QR code) |
| `options.base_url` | API base URL (default: `https://api.waland.dev`) |
| `options.timeout` | Request timeout (default: `30s`) |
| `options.http_client` | Custom `reqwest::Client` instance |

## `send_message(params)`

| Field | Required | Description |
|-------|----------|-------------|
| `chat_id` | Yes | WhatsApp JID: `{number}@s.whatsapp.net` (DM) or `{groupId}@g.us` (group) |
| `text` | No* | Message body or media caption |
| `media_url` | No* | Public HTTPS URL of media to send |
| `media_filename` | No | Filename override for downloaded media |

\* At least one of `text` or `media_url` must be provided.

## Errors

- `WalandValidationError` — invalid arguments before the request is sent
- `WalandApiError` — API returned a non-success status (`status_code`, `message`, `body`)
- `WalandError::Transport` — network/transport-level request failures

## Multiple sessions

Use one client per session (same API key is fine):

```rust
let support = WalandClient::new(api_key.clone(), support_session_id, None)?;
let alerts = WalandClient::new(api_key, alerts_session_id, None)?;
```
