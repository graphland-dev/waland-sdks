export const DEFAULT_BASE_URL = "https://api.waland.dev";

export type SmsLogStatus = "pending" | "sent" | "failed";

export interface WalandClientOptions {
  /** API base URL. Defaults to `https://api.waland.dev`. */
  baseUrl?: string;
  /** Request timeout in milliseconds. Defaults to `30000`. */
  timeoutMs?: number;
}

export interface SendMessageParams {
  /** WhatsApp JID, e.g. `8801712345678@s.whatsapp.net` or `{groupId}@g.us`. */
  chatId: string;
  /** Message body, or caption when `mediaUrl` is set. */
  text?: string;
  /** Public HTTPS URL of media to send. */
  mediaUrl?: string;
  /** Optional filename override for downloaded media. */
  mediaFilename?: string;
}

export interface SendMessageResult {
  id: string;
  sessionId: string;
  organizationId: string;
  chatId: string;
  text: string | null;
  mediaUrl: string | null;
  status: SmsLogStatus;
  messageId: string | null;
  error: string | null;
  createdAt: string;
}

export interface CheckNumberParams {
  /** Phone number in international format, e.g. `8801712345678`. */
  number: string;
}

export interface CheckNumberResult {
  number?: string;
  chatId?: string;
  jid?: string;
  exists?: boolean;
  isWhatsApp?: boolean;
  onWhatsApp?: boolean;
  [key: string]: unknown;
}

export interface WalandApiErrorBody {
  statusCode: number;
  message: string | string[];
  error?: string;
}
