// US-08, shape the document: a heading takes a number and pushes PIPE-R1
// under it, a paragraph of prose appears with no number, an excluded
// requirement leaves the document for the tray and comes back on restore,
// and the model is untouched throughout. The heading and prose buttons open
// a prompt, which the page's dialog handler answers before the click. The
// demo is put back as it was found.

import { expect, test } from '@playwright/test'
import { Document } from './lib/document'
import { resetBoth } from './lib/graphql'
import { Viewer } from './lib/viewer'

const SHIPPED_CAPTION = 'capacity 1200, bottleneck parse'
const HEADING = 'Performance'
const PROSE = 'A paragraph on the budget.'

test('US-08 shape the document', async ({ page, request }) => {
  try {
    const before = await test.step('open the viewer and keep its version mark', async () => {
      const viewer = await Viewer.open(page)
      await expect(viewer.caption()).toHaveText(SHIPPED_CAPTION)
      await expect(viewer.version()).toContainText(/version \d+/)
      return (await viewer.version().textContent()) ?? ''
    })

    const heading = page.locator('#tree li.node.kind-heading')
    const numberOf = (id: string) => page.locator(`#tree li.node[data-id="${id}"] > .row .number`)

    const document = await test.step('open the document, insert a heading above PIPE-R1 and expect it numbered 1, PIPE-R1 at 1.1 and its derived requirements 1.1.1 to 1.1.5', async () => {
      const document = await Document.open(page)
      await expect(numberOf('PIPE-R1')).toHaveText('1')
      page.once('dialog', (dialog) => { dialog.accept(HEADING).catch(() => {}) })
      await document.headingAbove('PIPE-R1').click()
      await expect(heading).toHaveCount(1)
      await expect(heading.locator(':scope > .row .number')).toHaveText('1')
      await expect(heading.locator(':scope > .row h2.text')).toHaveText(HEADING)
      await expect(numberOf('PIPE-R1')).toHaveText('1.1')
      await expect(numberOf('PIPE-R1.1')).toHaveText('1.1.1')
      await expect(numberOf('PIPE-R1.2')).toHaveText('1.1.2')
      await expect(numberOf('PIPE-R1.3')).toHaveText('1.1.3')
      await expect(numberOf('PIPE-R1.4')).toHaveText('1.1.4')
      await expect(numberOf('PIPE-R1.5')).toHaveText('1.1.5')
      return document
    })

    await test.step('add a paragraph of prose under the heading and expect it in place with no number', async () => {
      const headingId = await heading.getAttribute('data-id')
      expect(headingId).not.toBeNull()
      page.once('dialog', (dialog) => { dialog.accept(PROSE).catch(() => {}) })
      await document.addProse(headingId ?? '').click()
      const prose = heading.locator(':scope > ol.nodes > li.node.kind-prose')
      await expect(prose).toHaveCount(1)
      await expect(prose.locator('.text')).toHaveText(PROSE)
      await expect(prose.locator(':scope > .row .number')).toHaveText('')
    })

    await test.step('exclude PIPE-R1.4 and expect it in the tray, gone from the document, and PIPE-R1.5 renumbered', async () => {
      // The tray is drawn from the model's own requirement list, so PIPE-R1.4
      // listed there is the model still listing it.
      await document.exclude('PIPE-R1.4').click()
      await expect(document.tray()).toContainText('PIPE-R1.4')
      await expect(document.row('PIPE-R1.4')).toHaveCount(0)
      await expect(numberOf('PIPE-R1.5')).toHaveText('1.1.4')
    })

    await test.step('restore PIPE-R1.4 and expect it back as the last child of PIPE-R1, numbered 1.1.5', async () => {
      await document.restore().click()
      await expect(document.row('PIPE-R1.4')).toHaveCount(1)
      await expect(numberOf('PIPE-R1.4')).toHaveText('1.1.5')
      await expect(document.row('PIPE-R1').locator(':scope > ol.nodes > li.node').last()).toHaveAttribute('data-id', 'PIPE-R1.4')
      await expect(document.tray()).toHaveText('No requirement is excluded.')
    })

    await test.step('open the viewer and expect the model version unchanged and PIPE-R1.4 still listed', async () => {
      const viewer = await Viewer.open(page)
      await expect(viewer.caption()).toHaveText(SHIPPED_CAPTION)
      await expect(viewer.version()).toHaveText(before)
      await expect(viewer.requirement('PIPE-R1.4')).toBeVisible()
    })
  } finally {
    await test.step('put the demo back as it was found', async () => {
      await resetBoth(request)
      const viewer = await Viewer.open(page)
      await expect(viewer.caption()).toHaveText(SHIPPED_CAPTION)
    })
  }
})
