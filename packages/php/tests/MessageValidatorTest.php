<?php

declare(strict_types=1);

namespace Waland\Tests;

use PHPUnit\Framework\TestCase;
use Waland\MessageValidator;
use Waland\WalandValidationException;

final class MessageValidatorTest extends TestCase
{
    public function testAssertNonEmpty(): void
    {
        $this->expectException(WalandValidationException::class);
        MessageValidator::assertNonEmpty('   ', 'apiKey');
    }
}
