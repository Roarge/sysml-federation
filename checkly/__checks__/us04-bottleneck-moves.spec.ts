// US-04, raise the bottleneck: parse goes from 1200 to 1700 and the
// bottleneck moves to the index pair, then indexA goes from 700 to 900 and
// PIPE-R1 passes. The viewer follows on a live update, so the check is
// skipped where the session carries no stream. The demo is put back as it
// was found.

import { expect, test } from '@playwright/test'
import { resetBoth } from './lib/graphql'
import { expectWithin, skipUnlessStreams } from './lib/sse'
import { Viewer } from './lib/viewer'

const SHIPPED_CAPTION = 'capacity 1200, bottleneck parse'

test('US-04 raise the bottleneck', async ({ page, request }) => {
  skipUnlessStreams(test)
  try {
    const viewer = await Viewer.open(page)
    const block = viewer.requirement('PIPE-R1')

    await test.step('open the viewer on the shipped state', async () => {
      await expect(viewer.caption()).toHaveText(SHIPPED_CAPTION)
      await expect(block).toHaveClass(/fail/)
    })

    await test.step('raise parse from 1200 to 1700 and expect capacity 1400 with the bottleneck at indexA and indexB, PIPE-R1 still FAIL', async () => {
      await viewer.edit('attribute|PIPE-S2|throughput', '1700')
      await expectWithin(viewer.caption(), 'capacity 1400, bottleneck indexA, indexB')
      await expect(block.locator('.verdict')).toHaveText('FAIL capacity 1400 against 1500, limited by indexA, indexB')
      await expect(block).toHaveClass(/fail/)
    })

    await test.step('raise indexA from 700 to 900 and expect capacity 1600, PIPE-R1 PASS at 1600 against 1500 and the block no longer red', async () => {
      await viewer.edit('attribute|PIPE-S3|throughput', '900')
      await expectWithin(viewer.caption(), 'capacity 1600, bottleneck indexA, indexB')
      await expect(block.locator('.verdict')).toHaveText('PASS capacity 1600 against 1500, limited by indexA, indexB')
      await expect(block).not.toHaveClass(/fail/)
    })
  } finally {
    await test.step('put the demo back as it was found', async () => {
      await resetBoth(request)
      const viewer = await Viewer.open(page)
      await expect(viewer.caption()).toHaveText(SHIPPED_CAPTION)
    })
  }
})
