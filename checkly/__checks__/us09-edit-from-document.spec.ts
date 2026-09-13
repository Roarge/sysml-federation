// US-09, change a value from the document: with the viewer open on a second
// page, the limit of PIPE-R1 and then the throughput of parse are corrected
// in the document's rows, and the viewer follows each within two seconds
// with no reload. The viewer follows on a live update, so the check is
// skipped where the session carries no stream. The demo is put back as it
// was found.

import { expect, test } from '@playwright/test'
import { Document } from './lib/document'
import { resetBoth } from './lib/graphql'
import { expectWithin, skipUnlessStreams } from './lib/sse'
import { Viewer } from './lib/viewer'

const SHIPPED_CAPTION = 'capacity 1200, bottleneck parse'

test('US-09 change a value from the document', async ({ context, page, request }) => {
  skipUnlessStreams(test)
  try {
    const viewerPage = await context.newPage()

    const { viewer, document } = await test.step('open the document and, on a second page, the viewer, both on the shipped state', async () => {
      const viewer = await Viewer.open(viewerPage)
      const document = await Document.open(page)
      await expect(viewer.caption()).toHaveText(SHIPPED_CAPTION)
      await expect(document.row('PIPE-R1').locator(':scope > .row .verdict')).toHaveText('FAIL capacity 1200 against 1500, limited by parse')
      return { viewer, document }
    })
    const block = viewer.requirement('PIPE-R1')
    const rowOf = (id: string) => document.row(id).locator(':scope > .row')

    await test.step('change the limit of PIPE-R1 to 1600 in its row and expect the viewer to show the new limit within two seconds', async () => {
      await document.edit('limit|PIPE-R1', '1600')
      await expectWithin(block.locator('.verdict'), 'FAIL capacity 1200 against 1600, limited by parse')
      await expect(block.locator('.constraint')).toContainText('1600')
      await expect(viewerPage.locator('#text')).toContainText('requiredRate = 1600')
    })

    await test.step('change the throughput of parse to 1700 in the row of PIPE-R1.2 and expect the viewer to follow within two seconds', async () => {
      await document.edit('attribute|PIPE-S2|throughput', '1700')
      // PIPE-R1.2's limit is derived from PIPE-R1's, so it too now reads 1600.
      await expectWithin(rowOf('PIPE-R1.2').locator('.verdict'), 'PASS throughput 1700 against 1600')
      await expect(rowOf('PIPE-R1').locator('.verdict')).toHaveText('FAIL capacity 1400 against 1600, limited by indexA, indexB')
      await expectWithin(viewer.caption(), 'capacity 1400, bottleneck indexA, indexB')
      await expect(viewerPage.locator('#text')).toContainText('throughput = 1700')
      await expect(viewerPage.locator('#sketch .box.bottleneck')).toHaveCount(2)
    })
  } finally {
    await test.step('put the demo back as it was found', async () => {
      await resetBoth(request)
      const viewer = await Viewer.open(page)
      await expect(viewer.caption()).toHaveText(SHIPPED_CAPTION)
    })
  }
})
