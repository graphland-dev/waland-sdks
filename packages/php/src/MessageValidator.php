<?php

declare(strict_types=1);

namespace Waland;

final class MessageValidator
{
    private const CHAT_ID_PATTERN = '/^[^@\s]+@(s\.whatsapp\.net|g\.us)$/';

    public static function assertNonEmpty(string $value, string $field): void
    {
        if (trim($value) === '') {
            throw new WalandValidationException("{$field} is required");
        }
    }

    /**
     * @param array{
     *   chatId: string,
     *   text?: string|null,
     *   mediaUrl?: string|null,
     *   mediaFilename?: string|null
     * } $params
     */
    public static function validateSendMessage(array $params): void
    {
        if (!isset($params['chatId'])) {
            throw new WalandValidationException('chatId is required');
        }

        self::assertNonEmpty($params['chatId'], 'chatId');

        $chatId = trim($params['chatId']);
        if (!preg_match(self::CHAT_ID_PATTERN, $chatId)) {
            throw new WalandValidationException(
                'chatId must be a WhatsApp JID, e.g. 8801712345678@s.whatsapp.net or {groupId}@g.us',
            );
        }

        $text = isset($params['text']) ? trim((string) $params['text']) : '';
        $mediaUrl = isset($params['mediaUrl']) ? trim((string) $params['mediaUrl']) : '';

        if ($text === '' && $mediaUrl === '') {
            throw new WalandValidationException('Either text or mediaUrl must be provided');
        }

        if ($mediaUrl !== '') {
            if (filter_var($mediaUrl, FILTER_VALIDATE_URL) === false) {
                throw new WalandValidationException('mediaUrl must be a valid URL');
            }

            $scheme = parse_url($mediaUrl, PHP_URL_SCHEME);
            if (!is_string($scheme) || !str_starts_with($scheme, 'http')) {
                throw new WalandValidationException(
                    'mediaUrl must include a protocol (http or https)',
                );
            }

            /**
             * @param array{number: string} $params
             */
            public static function validateCheckNumber(array $params): void
            {
                if (!isset($params['number'])) {
                    throw new WalandValidationException('number is required');
                }

                self::assertNonEmpty((string) $params['number'], 'number');
            }
        }

        if (array_key_exists('mediaFilename', $params) && trim((string) $params['mediaFilename']) === '') {
            throw new WalandValidationException('mediaFilename cannot be empty');
        }
    }
}
