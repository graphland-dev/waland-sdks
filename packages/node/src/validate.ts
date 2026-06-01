import { WalandValidationError } from "./errors.js";
import type { CheckNumberParams, SendMessageParams } from "./types.js";

const CHAT_ID_PATTERN = /^[^@\s]+@(s\.whatsapp\.net|g\.us)$/;

export function assertNonEmpty(value: string, field: string): void {
  if (!value?.trim()) {
    throw new WalandValidationError(`${field} is required`);
  }
}

export function validateSendMessageParams(params: SendMessageParams): void {
  assertNonEmpty(params.chatId, "chatId");

  if (!CHAT_ID_PATTERN.test(params.chatId.trim())) {
    throw new WalandValidationError(
      "chatId must be a WhatsApp JID, e.g. 8801712345678@s.whatsapp.net or {groupId}@g.us",
    );
  }

  export function validateCheckNumberParams(params: CheckNumberParams): void {
    assertNonEmpty(params.number, "number");
  }

  const text = params.text?.trim();
  const mediaUrl = params.mediaUrl?.trim();

  if (!text && !mediaUrl) {
    throw new WalandValidationError("Either text or mediaUrl must be provided");
  }

  if (mediaUrl) {
    let parsed: URL;
    try {
      parsed = new URL(mediaUrl);
    } catch {
      throw new WalandValidationError("mediaUrl must be a valid URL");
    }
    if (!parsed.protocol.startsWith("http")) {
      throw new WalandValidationError(
        "mediaUrl must include a protocol (http or https)",
      );
    }
  }

  if (params.mediaFilename !== undefined && !params.mediaFilename.trim()) {
    throw new WalandValidationError("mediaFilename cannot be empty");
  }
}
