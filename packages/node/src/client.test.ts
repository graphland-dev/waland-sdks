import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { WalandClient } from "./client.js";
import { WalandError, WalandValidationError } from "./errors.js";

const API_KEY = "waland_test_key";
const SESSION_ID = "session-abc123";

const successBody = {
  id: "log-id",
  sessionId: SESSION_ID,
  organizationId: "org-id",
  chatId: "8801712345678@s.whatsapp.net",
  text: "Hello",
  mediaUrl: null,
  status: "sent" as const,
  messageId: "wa-msg-id",
  error: null,
  createdAt: "2026-05-24T10:00:00.000Z",
};

const checkNumberBody = {
  number: "8801712345678",
  chatId: "8801712345678@s.whatsapp.net",
  jid: "8801712345678@s.whatsapp.net",
  exists: true,
};

describe("WalandClient", () => {
  beforeEach(() => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(
        new Response(JSON.stringify(successBody), {
          status: 201,
          headers: { "Content-Type": "application/json" },
        }),
      ),
    );
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("requires apiKey and sessionId", () => {
    expect(() => new WalandClient("", SESSION_ID)).toThrow(
      WalandValidationError,
    );
    expect(() => new WalandClient(API_KEY, "")).toThrow(WalandValidationError);
  });

  it("sends a text message with bearer auth", async () => {
    const client = new WalandClient(API_KEY, SESSION_ID);

    const result = await client.sendMessage({
      chatId: "8801712345678@s.whatsapp.net",
      text: "Hello",
    });

    expect(result).toEqual(successBody);
    expect(fetch).toHaveBeenCalledOnce();
    expect(fetch).toHaveBeenCalledWith(
      "https://api.waland.dev/v1/sessions/session-abc123/send",
      expect.objectContaining({
        method: "POST",
        headers: expect.objectContaining({
          Authorization: `Bearer ${API_KEY}`,
          "Content-Type": "application/json",
        }),
        body: JSON.stringify({
          chatId: "8801712345678@s.whatsapp.net",
          text: "Hello",
        }),
      }),
    );
  });

  it("sends media with optional filename", async () => {
    const client = new WalandClient(API_KEY, SESSION_ID);

    await client.sendMessage({
      chatId: "8801712345678@s.whatsapp.net",
      text: "Caption",
      mediaUrl: "https://example.com/photo.jpg",
      mediaFilename: "photo.jpg",
    });

    expect(fetch).toHaveBeenCalledWith(
      expect.any(String),
      expect.objectContaining({
        body: JSON.stringify({
          chatId: "8801712345678@s.whatsapp.net",
          text: "Caption",
          mediaUrl: "https://example.com/photo.jpg",
          mediaFilename: "photo.jpg",
        }),
      }),
    );
  });

  it("rejects invalid chatId", async () => {
    const client = new WalandClient(API_KEY, SESSION_ID);

    await expect(
      client.sendMessage({ chatId: "not-a-jid", text: "Hi" }),
    ).rejects.toThrow(WalandValidationError);
  });

  it("rejects when neither text nor mediaUrl is provided", async () => {
    const client = new WalandClient(API_KEY, SESSION_ID);

    await expect(
      client.sendMessage({ chatId: "8801712345678@s.whatsapp.net" }),
    ).rejects.toThrow(WalandValidationError);
  });

  it("throws WalandError on API errors", async () => {
    vi.mocked(fetch).mockResolvedValueOnce(
      new Response(
        JSON.stringify({
          statusCode: 401,
          message: "Invalid or missing org API key",
          error: "Unauthorized",
        }),
        { status: 401, headers: { "Content-Type": "application/json" } },
      ),
    );

    const client = new WalandClient(API_KEY, SESSION_ID);

    await expect(
      client.sendMessage({
        chatId: "8801712345678@s.whatsapp.net",
        text: "Hi",
      }),
    ).rejects.toMatchObject({
      name: "WalandError",
      statusCode: 401,
      message: "Invalid or missing org API key",
    } satisfies Partial<WalandError>);
  });

  it("checks a number", async () => {
    vi.mocked(fetch).mockResolvedValueOnce(
      new Response(JSON.stringify(checkNumberBody), {
        status: 200,
        headers: { "Content-Type": "application/json" },
      }),
    );

    const client = new WalandClient(API_KEY, SESSION_ID);
    const result = await client.checkNumber({ number: "8801712345678" });

    expect(result).toEqual(checkNumberBody);
    expect(fetch).toHaveBeenCalledWith(
      "https://api.waland.dev/v1/sessions/session-abc123/check-number",
      expect.objectContaining({
        method: "POST",
        body: JSON.stringify({ number: "8801712345678" }),
      }),
    );
  });

  it("rejects empty number for checkNumber", async () => {
    const client = new WalandClient(API_KEY, SESSION_ID);

    await expect(client.checkNumber({ number: "  " })).rejects.toThrow(
      WalandValidationError,
    );
  });
});
