<?php

declare(strict_types=1);

namespace Waland;

final class CurlHttpClient implements HttpClientInterface
{
    public function post(
        string $url,
        array $headers,
        string $body,
        int $timeoutSeconds,
    ): HttpResponse {
        $ch = curl_init($url);
        if ($ch === false) {
            throw new WalandException([
                'statusCode' => 500,
                'message' => 'Failed to initialize HTTP client',
                'error' => 'Internal Error',
            ]);
        }

        $headerLines = [];
        foreach ($headers as $name => $value) {
            $headerLines[] = "{$name}: {$value}";
        }

        try {
            curl_setopt_array($ch, [
                CURLOPT_POST => true,
                CURLOPT_RETURNTRANSFER => true,
                CURLOPT_HTTPHEADER => $headerLines,
                CURLOPT_POSTFIELDS => $body,
                CURLOPT_TIMEOUT => $timeoutSeconds,
            ]);

            $responseBody = curl_exec($ch);

            if ($responseBody === false) {
                $errno = curl_errno($ch);
                if ($errno === CURLE_OPERATION_TIMEDOUT) {
                    throw new WalandException([
                        'statusCode' => 408,
                        'message' => "Request timed out after {$timeoutSeconds}s",
                        'error' => 'Request Timeout',
                    ]);
                }

                throw new WalandException([
                    'statusCode' => 500,
                    'message' => curl_error($ch) ?: 'Request failed',
                    'error' => 'Internal Error',
                ]);
            }

            return new HttpResponse(
                (int) curl_getinfo($ch, CURLINFO_RESPONSE_CODE),
                $responseBody,
            );
        } finally {
            curl_close($ch);
        }
    }
}
