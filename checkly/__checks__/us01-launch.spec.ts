// US-01, launch with one command: both apps render from the one published
// port, and every request either page makes goes to that port. Nothing is
// edited, so there is nothing to put back.

import { expect, test } from '@playwright/test'
import { demoUrl } from './lib/demo'
import { Document } from './lib/document'
import { Viewer } from './lib/viewer'

const SHIPPED_CAPTION = 'capacity 1200, bottleneck parse'

test('US-01 launch with one command', async ({ page }) => {
  // Every request the two pages make, kept to show that none leaves the
  // demo's own address.
  const requests: string[] = []
  page.on('request', (request) => { requests.push(request.url()) })

  await test.step('open the viewer and expect it rendered', async () => {
    const viewer = await Viewer.open(page)
    await expect(viewer.status()).toBeAttached()
    await expect(viewer.version()).toContainText(/version \d+/)
    await expect(viewer.caption()).toHaveText(SHIPPED_CAPTION)
  })

  await test.step('open the document and expect the tree with PIPE-R1 rendered', async () => {
    const document = await Document.open(page)
    await expect(document.versions()).toContainText(/document version \d+, model version \d+/)
    await expect(document.row('PIPE-R1')).toBeVisible()
    await expect(document.row('PIPE-R1').locator(':scope > .row .number')).toHaveText('1')
  })

  await test.step('expect every request to have gone to the one published port', async () => {
    const elsewhere = requests.filter((url) => !url.startsWith(demoUrl() + '/'))
    expect(elsewhere).toEqual([])
  })
})
