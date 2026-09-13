// US-06, read the document: the sections carry the document's own numbers,
// each row carries what the model knows about its requirement, and the
// requirement whose case nothing runs is INCONCLUSIVE. Nothing is edited.

import { expect, test } from '@playwright/test'
import type { Locator } from '@playwright/test'
import { Document } from './lib/document'

// field is the value beside one label of a row's definition list.
function field(row: Locator, label: string): Locator {
  return row.locator(`dt:text-is("${label}") + dd`)
}

test('US-06 read the document', async ({ page }) => {
  const document = await Document.open(page)
  // rowOf is the node's own row, without the rows of its children.
  const rowOf = (id: string) => document.row(id).locator(':scope > .row')
  const numberOf = (id: string) => rowOf(id).locator('.number')

  await test.step('expect the versions line and the sections numbered by the document', async () => {
    await expect(document.versions()).toContainText(/document version \d+, model version \d+/)
    await expect(numberOf('PIPE-R1')).toHaveText('1')
    await expect(numberOf('PIPE-R1.1')).toHaveText('1.1')
    await expect(numberOf('PIPE-R1.2')).toHaveText('1.2')
    await expect(numberOf('PIPE-R1.3')).toHaveText('1.3')
    await expect(numberOf('PIPE-R1.4')).toHaveText('1.4')
    await expect(numberOf('PIPE-R1.5')).toHaveText('1.5')
    await expect(numberOf('PIPE-R2')).toHaveText('2')
  })

  await test.step('expect the row of PIPE-R1.4 to carry its short name, text, limit, relations, current value and verdict', async () => {
    const row = rowOf('PIPE-R1.4')
    await expect(row.locator('.head')).toHaveText('PIPE-R1.4 indexBThroughput')
    await expect(row.locator('.reqtext')).toHaveText('indexB shall sustain its allocated rate')
    await expect(field(row, 'limit')).toHaveText('capacity >= 750')
    await expect(field(row, 'derived from')).toHaveText('PIPE-R1')
    await expect(field(row, 'satisfied by')).toHaveText('indexB')
    await expect(field(row, 'current value')).toHaveText('capacity 700')
    await expect(field(row, 'indexB throughput').locator('input')).toHaveValue('700')
    await expect(row.locator('.verdict')).toHaveText('FAIL throughput 700 against 750')
  })

  await test.step('expect PIPE-R2 INCONCLUSIVE because its case is declared and no service runs it', async () => {
    const row = rowOf('PIPE-R2')
    await expect(field(row, 'verified by')).toHaveText('PIPE-VC1')
    await expect(row.locator('.verdict')).toHaveText('INCONCLUSIVE PIPE-VC1 is declared and no service runs it')
  })
})
