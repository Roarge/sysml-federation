// US-05, tighten the limit: PIPE-R1's limit goes to 1000 and the verdict
// passes, then to 2500 and it fails naming parse, with the edited number in
// the served text where the literal was. The viewer follows on a live
// update, so the check is skipped where the session carries no stream. The
// demo is put back as it was found.

import { expect, test } from '@playwright/test'
import { resetBoth } from './lib/graphql'
import { expectWithin, skipUnlessStreams } from './lib/sse'
import { Viewer } from './lib/viewer'

const SHIPPED_CAPTION = 'capacity 1200, bottleneck parse'

test('US-05 tighten the limit', async ({ page, request }) => {
  skipUnlessStreams(test)
  try {
    const viewer = await Viewer.open(page)
    const block = viewer.requirement('PIPE-R1')

    await test.step('open the viewer on the shipped state', async () => {
      await expect(viewer.caption()).toHaveText(SHIPPED_CAPTION)
      await expect(block.locator('.verdict')).toHaveText('FAIL capacity 1200 against 1500, limited by parse')
    })

    await test.step('lower the limit of PIPE-R1 from 1500 to 1000 and expect PASS', async () => {
      await viewer.edit('limit|PIPE-R1', '1000')
      await expectWithin(block.locator('.verdict'), 'PASS capacity 1200 against 1000, limited by parse')
      await expect(block).not.toHaveClass(/fail/)
    })

    await test.step('raise the limit of PIPE-R1 to 2500 and expect FAIL, the block red and the reason naming parse at 1200', async () => {
      await viewer.edit('limit|PIPE-R1', '2500')
      await expectWithin(block.locator('.verdict'), 'FAIL capacity 1200 against 2500, limited by parse')
      await expect(block).toHaveClass(/fail/)
    })

    await test.step('expect the edited number in the served text where the literal was', async () => {
      await expect(page.locator('#text')).toContainText('requiredRate = 2500')
      await expect(page.locator('#text')).not.toContainText('requiredRate = 1500')
    })
  } finally {
    await test.step('put the demo back as it was found', async () => {
      await resetBoth(request)
      const viewer = await Viewer.open(page)
      await expect(viewer.caption()).toHaveText(SHIPPED_CAPTION)
    })
  }
})
