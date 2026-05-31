use reqwest::Url;
use serde::{Deserialize, Serialize};
use std::time::Duration;
use thiserror::Error;

pub const DEFAULT_BASE_URL: &str = "https://api.waland.dev";

#[derive(Debug, Clone, Copy, Deserialize, Serialize, PartialEq, Eq)]
#[serde(rename_all = "lowercase")]
pub enum SmsLogStatus {
    Pending,
    Sent,
    Failed,
}

#[derive(Debug, Clone)]
pub struct WalandClientOptions {
    pub base_url: Option<String>,
    pub timeout: Option<Duration>,
    pub http_client: Option<reqwest::Client>,
}

#[derive(Debug, Clone, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct SendMessageParams {
    pub chat_id: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub text: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub media_url: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub media_filename: Option<String>,
}

#[derive(Debug, Clone, Deserialize, PartialEq, Eq)]
#[serde(rename_all = "camelCase")]
pub struct SendMessageResult {
    pub id: String,
    pub session_id: String,
    pub organization_id: String,
    pub chat_id: String,
    pub text: Option<String>,
    pub media_url: Option<String>,
    pub status: SmsLogStatus,
    pub message_id: Option<String>,
    pub error: Option<String>,
    pub created_at: String,
}

#[derive(Debug, Clone, Deserialize, PartialEq, Eq)]
#[serde(untagged)]
pub enum ApiMessage {
    Single(String),
    Multiple(Vec<String>),
}

impl ApiMessage {
    fn fallback(status: u16) -> Self {
        Self::Single(format!("Request failed with status {status}"))
    }
}

impl std::fmt::Display for ApiMessage {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            Self::Single(message) => write!(f, "{message}"),
            Self::Multiple(messages) => write!(f, "{}", messages.join("; ")),
        }
    }
}

#[derive(Debug, Clone, Deserialize, PartialEq, Eq)]
#[serde(rename_all = "camelCase")]
pub struct WalandApiErrorBody {
    pub status_code: u16,
    pub message: ApiMessage,
    pub error: Option<String>,
}

#[derive(Debug, Clone, Error)]
#[error("{message}")]
pub struct WalandApiError {
    pub status_code: u16,
    pub message: String,
    pub error: Option<String>,
    pub body: WalandApiErrorBody,
}

#[derive(Debug, Clone, Error, PartialEq, Eq)]
#[error("{message}")]
pub struct WalandValidationError {
    pub message: String,
}

#[derive(Debug, Error)]
pub enum WalandError {
    #[error(transparent)]
    Validation(#[from] WalandValidationError),
    #[error(transparent)]
    Api(#[from] WalandApiError),
    #[error(transparent)]
    Transport(#[from] reqwest::Error),
    #[error(transparent)]
    Json(#[from] serde_json::Error),
}

#[derive(Debug, Clone)]
pub struct WalandClient {
    api_key: String,
    session_id: String,
    base_url: String,
    timeout: Duration,
    http: reqwest::Client,
}

impl WalandClient {
    pub fn new(
        api_key: impl Into<String>,
        session_id: impl Into<String>,
        options: Option<WalandClientOptions>,
    ) -> Result<Self, WalandValidationError> {
        let api_key = api_key.into();
        let session_id = session_id.into();

        assert_non_empty(&api_key, "apiKey")?;
        assert_non_empty(&session_id, "sessionId")?;

        let base_url = options
            .as_ref()
            .and_then(|opts| opts.base_url.as_deref())
            .map(trim_trailing_slash)
            .unwrap_or_else(|| DEFAULT_BASE_URL.to_string());

        let timeout = options
            .as_ref()
            .and_then(|opts| opts.timeout)
            .unwrap_or(Duration::from_secs(30));

        let http = if let Some(client) = options.and_then(|opts| opts.http_client) {
            client
        } else {
            reqwest::Client::builder()
                .timeout(timeout)
                .build()
                .map_err(|_| WalandValidationError {
                    message: "failed to initialize HTTP client".to_string(),
                })?
        };

        Ok(Self {
            api_key: api_key.trim().to_string(),
            session_id: session_id.trim().to_string(),
            base_url,
            timeout,
            http,
        })
    }

    pub async fn send_message(
        &self,
        params: SendMessageParams,
    ) -> Result<SendMessageResult, WalandError> {
        validate_send_message_params(&params)?;

        let body = SendMessageParams {
            chat_id: params.chat_id.trim().to_string(),
            text: normalize_optional(params.text),
            media_url: normalize_optional(params.media_url),
            media_filename: normalize_optional(params.media_filename),
        };

        let url = format!(
            "{}/v1/sessions/{}/send",
            self.base_url,
            urlencoding::encode(&self.session_id)
        );

        let response = self
            .http
            .post(url)
            .bearer_auth(&self.api_key)
            .header("Content-Type", "application/json")
            .header("Accept", "application/json")
            .json(&body)
            .send()
            .await
            .map_err(|error| {
                if error.is_timeout() {
                    WalandError::Api(WalandApiError {
                        status_code: 408,
                        message: format!("Request timed out after {}ms", self.timeout.as_millis()),
                        error: Some("Request Timeout".to_string()),
                        body: WalandApiErrorBody {
                            status_code: 408,
                            message: ApiMessage::Single(format!(
                                "Request timed out after {}ms",
                                self.timeout.as_millis()
                            )),
                            error: Some("Request Timeout".to_string()),
                        },
                    })
                } else {
                    WalandError::Transport(error)
                }
            })?;

        let status = response.status().as_u16();
        let payload: serde_json::Value = match response.text().await {
            Ok(text) if text.trim().is_empty() => serde_json::json!({}),
            Ok(text) => serde_json::from_str(&text).unwrap_or_else(|_| {
                serde_json::json!({
                    "statusCode": status,
                    "message": text,
                    "error": "Error"
                })
            }),
            Err(error) => return Err(WalandError::Transport(error)),
        };

        if !(200..300).contains(&status) {
            let body = normalize_error_body(status, payload);
            return Err(WalandError::Api(WalandApiError {
                status_code: body.status_code,
                message: body.message.to_string(),
                error: body.error.clone(),
                body,
            }));
        }

        serde_json::from_value(payload).map_err(WalandError::Json)
    }
}

fn trim_trailing_slash(value: &str) -> String {
    value.trim().trim_end_matches('/').to_string()
}

fn assert_non_empty(value: &str, field: &str) -> Result<(), WalandValidationError> {
    if value.trim().is_empty() {
        return Err(WalandValidationError {
            message: format!("{field} is required"),
        });
    }
    Ok(())
}

fn normalize_optional(value: Option<String>) -> Option<String> {
    value.and_then(|v| {
        let trimmed = v.trim().to_string();
        if trimmed.is_empty() {
            None
        } else {
            Some(trimmed)
        }
    })
}

fn validate_send_message_params(params: &SendMessageParams) -> Result<(), WalandValidationError> {
    assert_non_empty(&params.chat_id, "chatId")?;

    let chat_id = params.chat_id.trim();
    if !is_valid_chat_id(chat_id) {
        return Err(WalandValidationError {
            message:
                "chatId must be a WhatsApp JID, e.g. 8801712345678@s.whatsapp.net or {groupId}@g.us"
                    .to_string(),
        });
    }

    let text = params.text.as_deref().map(str::trim).unwrap_or_default();
    let media_url = params
        .media_url
        .as_deref()
        .map(str::trim)
        .unwrap_or_default();

    if text.is_empty() && media_url.is_empty() {
        return Err(WalandValidationError {
            message: "Either text or mediaUrl must be provided".to_string(),
        });
    }

    if !media_url.is_empty() {
        let parsed = Url::parse(media_url).map_err(|_| WalandValidationError {
            message: "mediaUrl must be a valid URL".to_string(),
        })?;

        if !parsed.scheme().starts_with("http") {
            return Err(WalandValidationError {
                message: "mediaUrl must include a protocol (http or https)".to_string(),
            });
        }
    }

    if params.media_filename.is_some()
        && params
            .media_filename
            .as_deref()
            .map(str::trim)
            .unwrap_or_default()
            .is_empty()
    {
        return Err(WalandValidationError {
            message: "mediaFilename cannot be empty".to_string(),
        });
    }

    Ok(())
}

fn is_valid_chat_id(chat_id: &str) -> bool {
    if let Some((local, domain)) = chat_id.split_once('@') {
        if local.is_empty() || local.contains(char::is_whitespace) {
            return false;
        }
        return domain == "s.whatsapp.net" || domain == "g.us";
    }
    false
}

fn normalize_error_body(status: u16, payload: serde_json::Value) -> WalandApiErrorBody {
    let status_code = payload
        .get("statusCode")
        .and_then(|v| v.as_u64())
        .map(|v| v as u16)
        .unwrap_or(status);

    let message = if let Some(value) = payload.get("message") {
        match value {
            serde_json::Value::String(message) => ApiMessage::Single(message.clone()),
            serde_json::Value::Array(messages) => ApiMessage::Multiple(
                messages
                    .iter()
                    .filter_map(|item| item.as_str().map(ToString::to_string))
                    .collect(),
            ),
            _ => ApiMessage::fallback(status),
        }
    } else {
        ApiMessage::fallback(status)
    };

    let error = payload
        .get("error")
        .and_then(|value| value.as_str())
        .map(ToString::to_string)
        .or_else(|| {
            if payload.get("statusCode").is_some() {
                None
            } else {
                Some("Error".to_string())
            }
        });

    WalandApiErrorBody {
        status_code,
        message,
        error,
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use httpmock::Method::POST;
    use httpmock::MockServer;

    const API_KEY: &str = "waland_test_key";
    const SESSION_ID: &str = "session-abc123";

    #[test]
    fn requires_api_key_and_session_id() {
        let api_key_error = WalandClient::new("", SESSION_ID, None).unwrap_err();
        assert_eq!(api_key_error.message, "apiKey is required");

        let session_error = WalandClient::new(API_KEY, "", None).unwrap_err();
        assert_eq!(session_error.message, "sessionId is required");
    }

    #[tokio::test]
    async fn sends_text_message_with_bearer_auth() {
        let server = MockServer::start();

        let mock = server.mock(|when, then| {
            when.method(POST)
                .path("/v1/sessions/session-abc123/send")
                .header_exists("authorization")
                .header("content-type", "application/json")
                .json_body(serde_json::json!({
                    "chatId": "8801712345678@s.whatsapp.net",
                    "text": "Hello"
                }));

            then.status(201).json_body(serde_json::json!({
                "id": "log-id",
                "sessionId": SESSION_ID,
                "organizationId": "org-id",
                "chatId": "8801712345678@s.whatsapp.net",
                "text": "Hello",
                "mediaUrl": null,
                "status": "sent",
                "messageId": "wa-msg-id",
                "error": null,
                "createdAt": "2026-05-24T10:00:00.000Z"
            }));
        });

        let client = WalandClient::new(
            API_KEY,
            SESSION_ID,
            Some(WalandClientOptions {
                base_url: Some(server.base_url()),
                timeout: None,
                http_client: None,
            }),
        )
        .unwrap();

        let result = client
            .send_message(SendMessageParams {
                chat_id: "8801712345678@s.whatsapp.net".to_string(),
                text: Some("Hello".to_string()),
                media_url: None,
                media_filename: None,
            })
            .await
            .unwrap();

        mock.assert();
        assert_eq!(result.status, SmsLogStatus::Sent);
        assert_eq!(result.message_id, Some("wa-msg-id".to_string()));
    }

    #[tokio::test]
    async fn throws_api_error_on_failure() {
        let server = MockServer::start();

        server.mock(|when, then| {
            when.method(POST).path("/v1/sessions/session-abc123/send");
            then.status(401).json_body(serde_json::json!({
                "statusCode": 401,
                "message": "Invalid or missing org API key",
                "error": "Unauthorized"
            }));
        });

        let client = WalandClient::new(
            API_KEY,
            SESSION_ID,
            Some(WalandClientOptions {
                base_url: Some(server.base_url()),
                timeout: None,
                http_client: None,
            }),
        )
        .unwrap();

        let error = client
            .send_message(SendMessageParams {
                chat_id: "8801712345678@s.whatsapp.net".to_string(),
                text: Some("Hi".to_string()),
                media_url: None,
                media_filename: None,
            })
            .await
            .unwrap_err();

        match error {
            WalandError::Api(api_error) => {
                assert_eq!(api_error.status_code, 401);
                assert_eq!(api_error.message, "Invalid or missing org API key");
            }
            _ => panic!("expected WalandError::Api"),
        }
    }

    #[tokio::test]
    async fn rejects_invalid_chat_id() {
        let client = WalandClient::new(API_KEY, SESSION_ID, None).unwrap();

        let error = client
            .send_message(SendMessageParams {
                chat_id: "not-a-jid".to_string(),
                text: Some("Hi".to_string()),
                media_url: None,
                media_filename: None,
            })
            .await
            .unwrap_err();

        match error {
            WalandError::Validation(validation_error) => {
                assert_eq!(
                    validation_error.message,
                    "chatId must be a WhatsApp JID, e.g. 8801712345678@s.whatsapp.net or {groupId}@g.us"
                );
            }
            _ => panic!("expected WalandError::Validation"),
        }
    }

    #[tokio::test]
    async fn rejects_missing_text_and_media() {
        let client = WalandClient::new(API_KEY, SESSION_ID, None).unwrap();

        let error = client
            .send_message(SendMessageParams {
                chat_id: "8801712345678@s.whatsapp.net".to_string(),
                text: None,
                media_url: None,
                media_filename: None,
            })
            .await
            .unwrap_err();

        match error {
            WalandError::Validation(validation_error) => {
                assert_eq!(
                    validation_error.message,
                    "Either text or mediaUrl must be provided"
                );
            }
            _ => panic!("expected WalandError::Validation"),
        }
    }
}
