<?php

declare(strict_types=1);

namespace Waland\Tests;

use Waland\HttpClientInterface;
use Waland\HttpResponse;

final class FakeHttpClient implements HttpClientInterface
{
    /** @var list<HttpResponse> */
    private array $responses;

    /** @var list<array{url: string, headers: array<string, string>, body: string}> */
    public array $requests = [];

    /**
     * @param list<HttpResponse> $responses
     */
    public function __construct(array $responses)
    {
        $this->responses = $responses;
    }

    public function post(
        string $url,
        array $headers,
        string $body,
        int $timeoutSeconds,
    ): HttpResponse {
        $this->requests[] = [
            'url' => $url,
            'headers' => $headers,
            'body' => $body,
        ];

        if ($this->responses === []) {
            return new HttpResponse(500, '');
        }

        return array_shift($this->responses);
    }
}
