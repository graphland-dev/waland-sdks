<?php

declare(strict_types=1);

namespace Waland;

use Exception;

final class WalandException extends Exception
{
    /** @var array{statusCode: int, message: string|string[], error?: string} */
    public readonly array $body;

    public readonly int $statusCode;

    public readonly ?string $error;

    /**
     * @param array{statusCode: int, message: string|string[], error?: string} $body
     */
    public function __construct(array $body)
    {
        $message = self::formatMessage($body['message']);
        parent::__construct($message);

        $this->body = $body;
        $this->statusCode = $body['statusCode'];
        $this->error = $body['error'] ?? null;
    }

    /** @param string|string[] $message */
    private static function formatMessage(string|array $message): string
    {
        if (is_array($message)) {
            return implode('; ', $message);
        }

        return $message;
    }
}
