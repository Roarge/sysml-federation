// US-12, reset: a value is edited in the viewer and the document's structure
// in the document, then the viewer's Reset control brings both apps back to
// the shipped state within two seconds. Each app follows on a live update,
// so the check is skipped where the session carries no stream. The reset is
// the story, and the demo is put back through the router as well, so that
// a failure part way through leaves it as it was found.

import { expect, test } from '@playwright/test'
import { Document } from './lib/document'
import { resetBoth } from './lib/graphql'
import { expectWithin, skipUnlessStreams } from './lib/sse'
import { Viewer } from './lib/viewer'

const SHIPPED_CAPTION = 'capacity 1200, bottleneck parse'
const SHIPPED_VERDICT = 'FAIL capacity 1200 against 1500, limited by parse'

test('US-12 reset', async ({ context, page, request }) => {
  skipUnlessStreams(test)
  try {
    const documentPage = await context.newPage()

    const { viewer, document } = await test.step('open the viewer and, on a second page, the document, both on the shipped state', async () => {
      const viewer = await Viewer.open(page)
      const document = await Document.open(documentPage)
      await expect(viewer.caption()).toHaveText(SHIPPED_CAPTION)
      await expect(document.row('PIPE-R1.4').locator(':scope > .row .number')).toHaveText('1.4')
      return { viewer, document }
    })
    const numberOf = (id: string) => document.row(id).locator(':scope > .row .number')

    await test.step('raise parse to 1700 in the viewer and exclude PIPE-R1.4 in the document', async () => {
      await viewer.edit('attribute|PIPE-S2|throughput', '1700')
      await expectWithin(viewer.caption(), 'capacity 1400, bottleneck indexA, indexB')
      await document.exclude('PIPE-R1.4').click()
      await expectWithin(document.tray(), 'PIPE-R1.4')
      await expect(document.row('PIPE-R1.4')).toHaveCount(0)
    })

    await test.step('press Reset in the viewer and expect both apps back at the shipped state within two seconds', async () => {
      await viewer.reset().click()
      await expectWithin(viewer.caption(), SHIPPED_CAPTION)
      await expect(viewer.requirement('PIPE-R1').locator('.verdict')).toHaveText(SHIPPED_VERDICT)
      await expectWithin(document.tray(), 'No requirement is excluded.')
      await expect(numberOf('PIPE-R1.4')).toHaveText('1.4')
      await expect(document.row('PIPE-R1').locator(':scope > .row .verdict')).toHaveText(SHIPPED_VERDICT)
    })
  } finally {
    await test.step('put the demo back as it was found', async () => {
      await resetBoth(request)
      const viewer = await Viewer.open(page)
      await expect(viewer.caption()).toHaveText(SHIPPED_CAPTION)
    })
  }
})
