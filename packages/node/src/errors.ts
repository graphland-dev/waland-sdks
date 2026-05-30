import type { WalandApiErrorBody } from "./types.js";

export class WalandError extends Error {
  readonly statusCode: number;
  readonly error?: string;
  readonly body: WalandApiErrorBody;

  constructor(body: WalandApiErrorBody) {
    const message = formatApiMessage(body.message);
    super(message);
    this.name = "WalandError";
    this.statusCode = body.statusCode;
    this.error = body.error;
    this.body = body;
  }
}

export class WalandValidationError extends Error {
  constructor(message: string) {
    super(message);
    this.name = "WalandValidationError";
  }
}

function formatApiMessage(message: string | string[]): string {
  if (Array.isArray(message)) {
    return message.join("; ");
  }
  return message;
}
