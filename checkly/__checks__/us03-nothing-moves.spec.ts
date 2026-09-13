// US-03, raise a server that is not the bottleneck: ingest goes from 2000 to
// 3000 and nothing moves. The version still grows by one, because the edit
// was accepted, and that is what shows the page has redrawn before the
// caption is read. The demo is put back as it was found.

import { expect, test } from '@playwright/test'
import { resetBoth } from './lib/graphql'
import { expectWithin } from './lib/sse'
import { Viewer } from './lib/viewer'

const SHIPPED_CAPTION = 'capacity 1200, bottleneck parse'

// versionNumber reads the number out of the viewer's "version N" mark.
async function versionNumber(viewer: Viewer): Promise<number> {
  const text = await viewer.version().textContent()
  const match = /version (\d+)/.exec(text ?? '')
  if (match === null) {
    throw new Error(`the viewer shows no version: ${JSON.stringify(text)}`)
  }
  return Number(match[1])
}

test('US-03 raise a server that is not the bottleneck', async ({ page, request }) => {
  try {
    const viewer = await Viewer.open(page)

    const before = await test.step('open the viewer on the shipped state and keep its version', async () => {
      await expect(viewer.caption()).toHaveText(SHIPPED_CAPTION)
      return versionNumber(viewer)
    })

    await test.step('raise ingest from 2000 to 3000 and expect the version to grow by one', async () => {
      await viewer.edit('attribute|PIPE-S1|throughput', '3000')
      await expectWithin(viewer.version(), `version ${before + 1}`)
    })

    await test.step('expect the capacity, the bottleneck and the verdict of PIPE-R1 unchanged', async () => {
      await expect(viewer.caption()).toHaveText(SHIPPED_CAPTION)
      await expect(viewer.requirement('PIPE-R1').locator('.verdict')).toHaveText('FAIL capacity 1200 against 1500, limited by parse')
    })

    await test.step('expect PIPE-R1.1 to pass with the new value', async () => {
      await expect(viewer.requirement('PIPE-R1.1').locator('.verdict')).toHaveText('PASS throughput 3000 against 1500')
    })
  } finally {
    await test.step('put the demo back as it was found', async () => {
      await resetBoth(request)
      const viewer = await Viewer.open(page)
      await expect(viewer.caption()).toHaveText(SHIPPED_CAPTION)
    })
  }
})
