// The live-update helpers. A page refreshes on a server-sent event, and only
// a tunnel that carries a streamed response lets that event through. A
// session without one sets SSE_STREAMS=0, and the checks that wait for a
// second page to catch up are skipped rather than failed.

import { expect } from '@playwright/test'
import type {
  Locator,
  PlaywrightTestArgs,
  PlaywrightTestOptions,
  PlaywrightWorkerArgs,
  PlaywrightWorkerOptions,
  TestType,
} from '@playwright/test'

type Test = TestType<PlaywrightTestArgs & PlaywrightTestOptions, PlaywrightWorkerArgs & PlaywrightWorkerOptions>

export function streamsExpected(): boolean {
  return process.env.SSE_STREAMS !== '0'
}

// expectWithin waits for the text to appear, with a budget short enough that
// a page which never heard the event fails the check rather than the run.
export async function expectWithin(locator: Locator, text: string | RegExp, ms = 2000): Promise<void> {
  await expect(locator).toContainText(text, { timeout: ms })
}

export function skipUnlessStreams(test: Test): void {
  test.skip(!streamsExpected(), 'live updates need the named tunnel')
}
