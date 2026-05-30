import { WalandError } from "./errors.js";
import { validateSendMessageParams, assertNonEmpty } from "./validate.js";
import {
  DEFAULT_BASE_URL,
  type SendMessageParams,
  type SendMessageResult,
  type WalandApiErrorBody,
  type WalandClientOptions,
} from "./types.js";

export class WalandClient {
  private readonly apiKey: string;
  private readonly sessionId: string;
  private readonly baseUrl: string;
  private readonly timeoutMs: number;

  constructor(
    apiKey: string,
    sessionId: string,
    options: WalandClientOptions = {},
  ) {
    assertNonEmpty(apiKey, "apiKey");
    assertNonEmpty(sessionId, "sessionId");

    this.apiKey = apiKey.trim();
    this.sessionId = sessionId.trim();
    this.baseUrl = (options.baseUrl ?? DEFAULT_BASE_URL).replace(/\/$/, "");
    this.timeoutMs = options.timeoutMs ?? 30_000;
  }

  async sendMessage(params: SendMessageParams): Promise<SendMessageResult> {
    validateSendMessageParams(params);

    const url = `${this.baseUrl}/v1/sessions/${encodeURIComponent(this.sessionId)}/send`;
    const body: Record<string, string> = {
      chatId: params.chatId.trim(),
    };

    const text = params.text?.trim();
    if (text) {
      body.text = text;
    }
    if (params.mediaUrl?.trim()) {
      body.mediaUrl = params.mediaUrl.trim();
    }
    if (params.mediaFilename?.trim()) {
      body.mediaFilename = params.mediaFilename.trim();
    }

    const controller = new AbortController();
    const timeout = setTimeout(() => controller.abort(), this.timeoutMs);

    let response: Response;
    try {
      response = await fetch(url, {
        method: "POST",
        headers: {
          Authorization: `Bearer ${this.apiKey}`,
          "Content-Type": "application/json",
          Accept: "application/json",
        },
        body: JSON.stringify(body),
        signal: controller.signal,
      });
    } catch (error) {
      if (error instanceof Error && error.name === "AbortError") {
        throw new WalandError({
          statusCode: 408,
          message: `Request timed out after ${this.timeoutMs}ms`,
          error: "Request Timeout",
        });
      }
      throw error;
    } finally {
      clearTimeout(timeout);
    }

    const payload = await parseJsonBody(response);

    if (!response.ok) {
      throw new WalandError(normalizeErrorBody(response.status, payload));
    }

    return payload as SendMessageResult;
  }
}

async function parseJsonBody(response: Response): Promise<unknown> {
  const text = await response.text();
  if (!text) {
    return {};
  }
  try {
    return JSON.parse(text) as unknown;
  } catch {
    return {
      statusCode: response.status,
      message: text,
      error: response.statusText || "Error",
    };
  }
}

function normalizeErrorBody(
  status: number,
  payload: unknown,
): WalandApiErrorBody {
  if (payload && typeof payload === "object" && "statusCode" in payload) {
    const body = payload as WalandApiErrorBody;
    return {
      statusCode: body.statusCode ?? status,
      message: body.message ?? responseStatusMessage(status),
      error: body.error,
    };
  }

  return {
    statusCode: status,
    message:
      typeof payload === "object" &&
      payload !== null &&
      "message" in payload &&
      typeof (payload as { message: unknown }).message === "string"
        ? (payload as { message: string }).message
        : responseStatusMessage(status),
    error: "Error",
  };
}

function responseStatusMessage(status: number): string {
  return `Request failed with status ${status}`;
}
