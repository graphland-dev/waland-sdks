<?php

declare(strict_types=1);

namespace Waland;

final class WalandClient
{
    public const DEFAULT_BASE_URL = 'https://api.waland.dev';

    private readonly string $apiKey;

    private readonly string $sessionId;

    private readonly string $baseUrl;

    private readonly int $timeoutSeconds;

    private readonly HttpClientInterface $http;

    /**
     * @param array{
     *   baseUrl?: string,
     *   timeoutSeconds?: int,
     *   httpClient?: HttpClientInterface
     * } $options
     */
    public function __construct(
        string $apiKey,
        string $sessionId,
        array $options = [],
    ) {
        MessageValidator::assertNonEmpty($apiKey, 'apiKey');
        MessageValidator::assertNonEmpty($sessionId, 'sessionId');

        $this->apiKey = trim($apiKey);
        $this->sessionId = trim($sessionId);
        $this->baseUrl = rtrim($options['baseUrl'] ?? self::DEFAULT_BASE_URL, '/');
        $this->timeoutSeconds = $options['timeoutSeconds'] ?? 30;
        $this->http = $options['httpClient'] ?? new CurlHttpClient();
    }

    /**
     * @param array{
     *   chatId: string,
     *   text?: string|null,
     *   mediaUrl?: string|null,
     *   mediaFilename?: string|null
     * } $params
     *
     * @return array<string, mixed>
     */
    public function sendMessage(array $params): array
    {
        MessageValidator::validateSendMessage($params);

        $body = ['chatId' => trim($params['chatId'])];

        if (isset($params['text']) && trim((string) $params['text']) !== '') {
            $body['text'] = trim((string) $params['text']);
        }
        if (isset($params['mediaUrl']) && trim((string) $params['mediaUrl']) !== '') {
            $body['mediaUrl'] = trim((string) $params['mediaUrl']);
        }
        if (isset($params['mediaFilename']) && trim((string) $params['mediaFilename']) !== '') {
            $body['mediaFilename'] = trim((string) $params['mediaFilename']);
        }

        $url = $this->baseUrl
            . '/v1/sessions/'
            . rawurlencode($this->sessionId)
            . '/send';

        $response = $this->http->post(
            $url,
            [
                'Authorization' => 'Bearer ' . $this->apiKey,
                'Content-Type' => 'application/json',
                'Accept' => 'application/json',
            ],
            json_encode($body, JSON_THROW_ON_ERROR),
            $this->timeoutSeconds,
        );

        $payload = self::parseJsonBody($response->body, $response->statusCode);

        if ($response->statusCode < 200 || $response->statusCode >= 300) {
            throw new WalandException(self::normalizeErrorBody($response->statusCode, $payload));
        }

        return $payload;
    }

    /**
     * @param array{number: string}|string $params
     *
     * @return array<string, mixed>
     */
    public function checkNumber(array|string $params): array
    {
        if (is_string($params)) {
            $params = ['number' => $params];
        }

        MessageValidator::validateCheckNumber($params);

        $url = $this->baseUrl
            . '/v1/sessions/'
            . rawurlencode($this->sessionId)
            . '/check-number';

        $response = $this->http->post(
            $url,
            [
                'Authorization' => 'Bearer ' . $this->apiKey,
                'Content-Type' => 'application/json',
                'Accept' => 'application/json',
            ],
            json_encode(['number' => trim((string) $params['number'])], JSON_THROW_ON_ERROR),
            $this->timeoutSeconds,
        );

        $payload = self::parseJsonBody($response->body, $response->statusCode);

        if ($response->statusCode < 200 || $response->statusCode >= 300) {
            throw new WalandException(self::normalizeErrorBody($response->statusCode, $payload));
        }

        return $payload;
    }

    /**
     * @return array<string, mixed>
     */
    private static function parseJsonBody(string $body, int $statusCode): array
    {
        if ($body === '') {
            return [];
        }

        try {
            $decoded = json_decode($body, true, 512, JSON_THROW_ON_ERROR);
            if (!is_array($decoded)) {
                return [
                    'statusCode' => $statusCode,
                    'message' => $body,
                    'error' => 'Error',
                ];
            }

            return $decoded;
        } catch (\JsonException) {
            return [
                'statusCode' => $statusCode,
                'message' => $body,
                'error' => 'Error',
            ];
        }
    }

    /**
     * @param array<string, mixed> $payload
     *
     * @return array{statusCode: int, message: string|string[], error?: string}
     */
    private static function normalizeErrorBody(int $status, array $payload): array
    {
        if (isset($payload['statusCode'])) {
            return [
                'statusCode' => (int) ($payload['statusCode'] ?? $status),
                'message' => $payload['message'] ?? self::statusMessage($status),
                'error' => isset($payload['error']) ? (string) $payload['error'] : null,
            ];
        }

        $message = $payload['message'] ?? self::statusMessage($status);

        return [
            'statusCode' => $status,
            'message' => is_string($message) ? $message : self::statusMessage($status),
            'error' => 'Error',
        ];
    }

    private static function statusMessage(int $status): string
    {
        return "Request failed with status {$status}";
    }
}
