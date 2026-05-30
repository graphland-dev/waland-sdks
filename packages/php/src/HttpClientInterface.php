<?php

declare(strict_types=1);

namespace Waland;

interface HttpClientInterface
{
    /**
     * @param array<string, string> $headers
     */
    public function post(
        string $url,
        array $headers,
        string $body,
        int $timeoutSeconds,
    ): HttpResponse;
}
