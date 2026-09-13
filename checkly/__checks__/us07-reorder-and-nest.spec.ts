// US-07, reorder and nest: PIPE-R1.5 moves above PIPE-R1.1 and the numbers
// follow, PIPE-R2 moves under PIPE-R1 and keeps its relations, and the model
// is untouched. A pointer drag does not reach the vendored drag library in a
// headless run, so each move is the moveNode mutation the drop would send,
// through the router, and the document page is loaded afresh to read the
// result. The demo is put back as it was found.

import { expect, test } from '@playwright/test'
import { Document } from './lib/document'
import { graphql, resetBoth } from './lib/graphql'
import { Viewer } from './lib/viewer'

const SHIPPED_CAPTION = 'capacity 1200, bottleneck parse'

type MoveAnswer = { moveNode: { version: number } }

// MOVE_NODE is the mutation the document app sends from a drop: the node,
// the list it landed in and its index there.
const MOVE_NODE = `mutation MoveNode($id: ID!, $parentId: ID, $index: Int!) {
  moveNode(id: $id, parentId: $parentId, index: $index) { version }
}`

test('US-07 reorder and nest', async ({ page, request }) => {
  try {
    const before = await test.step('open the viewer and keep its version mark', async () => {
      const viewer = await Viewer.open(page)
      await expect(viewer.caption()).toHaveText(SHIPPED_CAPTION)
      await expect(viewer.version()).toContainText(/version \d+/)
      return (await viewer.version().textContent()) ?? ''
    })

    const document = await test.step('move PIPE-R1.5 above PIPE-R1.1, open the document and expect it numbered 1.1 with the others shifted in their former order', async () => {
      await graphql<MoveAnswer>(request, MOVE_NODE, { id: 'PIPE-R1.5', parentId: 'PIPE-R1', index: 0 })
      const document = await Document.open(page)
      const numberOf = (id: string) => document.row(id).locator(':scope > .row .number')
      await expect(numberOf('PIPE-R1.5')).toHaveText('1.1')
      await expect(numberOf('PIPE-R1.1')).toHaveText('1.2')
      await expect(numberOf('PIPE-R1.2')).toHaveText('1.3')
      await expect(numberOf('PIPE-R1.3')).toHaveText('1.4')
      await expect(numberOf('PIPE-R1.4')).toHaveText('1.5')
      await expect(numberOf('PIPE-R1')).toHaveText('1')
      await expect(numberOf('PIPE-R2')).toHaveText('2')
      return document
    })

    await test.step('move PIPE-R2 under PIPE-R1 and expect it numbered 1.6 with its relations still shown', async () => {
      await graphql<MoveAnswer>(request, MOVE_NODE, { id: 'PIPE-R2', parentId: 'PIPE-R1', index: 6 })
      await page.reload()
      const row = document.row('PIPE-R2').locator(':scope > .row')
      await expect(document.row('PIPE-R1').locator(':scope > ol.nodes > li.node[data-id="PIPE-R2"]')).toBeVisible()
      await expect(row.locator('.number')).toHaveText('1.6')
      await expect(row.locator('dt:text-is("satisfied by") + dd')).toHaveText('pipeline')
      await expect(row.locator('dt:text-is("verified by") + dd')).toHaveText('PIPE-VC1')
      await expect(row.locator('.verdict')).toHaveText('INCONCLUSIVE PIPE-VC1 is declared and no service runs it')
    })

    await test.step('open the viewer and expect the model version, the text and the sketch unchanged', async () => {
      const viewer = await Viewer.open(page)
      await expect(viewer.caption()).toHaveText(SHIPPED_CAPTION)
      await expect(viewer.version()).toHaveText(before)
      await expect(viewer.requirement('PIPE-R2')).toBeVisible()
    })
  } finally {
    await test.step('put the demo back as it was found', async () => {
      await resetBoth(request)
      const viewer = await Viewer.open(page)
      await expect(viewer.caption()).toHaveText(SHIPPED_CAPTION)
    })
  }
})
