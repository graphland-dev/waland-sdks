<?php

declare(strict_types=1);

namespace Waland\Tests;

use PHPUnit\Framework\TestCase;
use Waland\HttpResponse;
use Waland\WalandClient;
use Waland\WalandException;
use Waland\WalandValidationException;

final class WalandClientTest extends TestCase
{
    private const API_KEY = 'waland_test_key';

    private const SESSION_ID = 'session-abc123';

    public function testConstructorRequiresCredentials(): void
    {
        $this->expectException(WalandValidationException::class);
        new WalandClient('', self::SESSION_ID);
    }

    public function testSendTextMessage(): void
    {
        $success = json_encode([
            'id' => 'log-id',
            'sessionId' => self::SESSION_ID,
            'organizationId' => 'org-id',
            'chatId' => '8801712345678@s.whatsapp.net',
            'text' => 'Hello',
            'mediaUrl' => null,
            'status' => 'sent',
            'messageId' => 'wa-msg-id',
            'error' => null,
            'createdAt' => '2026-05-24T10:00:00.000Z',
        ], JSON_THROW_ON_ERROR);

        $http = new FakeHttpClient([new HttpResponse(201, $success)]);
        $client = new WalandClient(self::API_KEY, self::SESSION_ID, [
            'httpClient' => $http,
        ]);

        $result = $client->sendMessage([
            'chatId' => '8801712345678@s.whatsapp.net',
            'text' => 'Hello',
        ]);

        $this->assertSame('sent', $result['status']);
        $this->assertSame('wa-msg-id', $result['messageId']);
        $this->assertCount(1, $http->requests);
        $this->assertSame(
            'https://api.waland.dev/v1/sessions/session-abc123/send',
            $http->requests[0]['url'],
        );
        $this->assertSame('Bearer ' . self::API_KEY, $http->requests[0]['headers']['Authorization']);
        $this->assertStringContainsString('"text":"Hello"', $http->requests[0]['body']);
    }

    public function testSendMediaMessage(): void
    {
        $http = new FakeHttpClient([new HttpResponse(201, '{}')]);
        $client = new WalandClient(self::API_KEY, self::SESSION_ID, [
            'httpClient' => $http,
        ]);

        $client->sendMessage([
            'chatId' => '8801712345678@s.whatsapp.net',
            'text' => 'Caption',
            'mediaUrl' => 'https://example.com/photo.jpg',
            'mediaFilename' => 'photo.jpg',
        ]);

        $this->assertStringContainsString(
            '"mediaUrl":"https:\/\/example.com\/photo.jpg"',
            $http->requests[0]['body'],
        );
        $this->assertStringContainsString('"mediaFilename":"photo.jpg"', $http->requests[0]['body']);
    }

    public function testRejectsInvalidChatId(): void
    {
        $client = new WalandClient(self::API_KEY, self::SESSION_ID);

        $this->expectException(WalandValidationException::class);
        $client->sendMessage([
            'chatId' => 'not-a-jid',
            'text' => 'Hi',
        ]);
    }

    public function testRejectsMissingTextAndMedia(): void
    {
        $client = new WalandClient(self::API_KEY, self::SESSION_ID);

        $this->expectException(WalandValidationException::class);
        $client->sendMessage([
            'chatId' => '8801712345678@s.whatsapp.net',
        ]);
    }

    public function testThrowsWalandExceptionOnApiError(): void
    {
        $error = json_encode([
            'statusCode' => 401,
            'message' => 'Invalid or missing org API key',
            'error' => 'Unauthorized',
        ], JSON_THROW_ON_ERROR);

        $http = new FakeHttpClient([new HttpResponse(401, $error)]);
        $client = new WalandClient(self::API_KEY, self::SESSION_ID, [
            'httpClient' => $http,
        ]);

        try {
            $client->sendMessage([
                'chatId' => '8801712345678@s.whatsapp.net',
                'text' => 'Hi',
            ]);
            $this->fail('Expected WalandException');
        } catch (WalandException $e) {
            $this->assertSame(401, $e->statusCode);
            $this->assertSame('Invalid or missing org API key', $e->getMessage());
        }
    }

    public function testCheckNumber(): void
    {
        $success = json_encode([
            'number' => '8801712345678',
            'chatId' => '8801712345678@s.whatsapp.net',
            'jid' => '8801712345678@s.whatsapp.net',
            'exists' => true,
        ], JSON_THROW_ON_ERROR);

        $http = new FakeHttpClient([new HttpResponse(200, $success)]);
        $client = new WalandClient(self::API_KEY, self::SESSION_ID, [
            'httpClient' => $http,
        ]);

        $result = $client->checkNumber(['number' => '8801712345678']);

        $this->assertTrue($result['exists']);
        $this->assertSame(
            'https://api.waland.dev/v1/sessions/session-abc123/check-number',
            $http->requests[0]['url'],
        );
        $this->assertSame('{"number":"8801712345678"}', $http->requests[0]['body']);
    }

    public function testCheckNumberRequiresNumber(): void
    {
        $client = new WalandClient(self::API_KEY, self::SESSION_ID);

        $this->expectException(WalandValidationException::class);
        $client->checkNumber(['number' => '   ']);
    }
}
