// US-10, change from the viewer, watch the document: with parse already at
// 1700 and both apps open, indexA is raised from 700 to 900 in the viewer
// and the document's rows follow within two seconds with no reload. The
// document follows on a live update, so the check is skipped where the
// session carries no stream. The demo is put back as it was found.

import { expect, test } from '@playwright/test'
import { Document } from './lib/document'
import { graphql, resetBoth } from './lib/graphql'
import { expectWithin, skipUnlessStreams } from './lib/sse'
import { Viewer } from './lib/viewer'

const SHIPPED_CAPTION = 'capacity 1200, bottleneck parse'

type SetAttributeAnswer = { setAttribute: { id: string } }

// PARSE_TO_1700 is the story's given: the bottleneck has already moved to
// the index pair before the two apps are opened.
const PARSE_TO_1700 = 'mutation { setAttribute(partId: "PIPE-S2", name: "throughput", value: 1700) { id } }'

test('US-10 change from the viewer, watch the document', async ({ context, page, request }) => {
  skipUnlessStreams(test)
  try {
    const viewerPage = await context.newPage()

    const { viewer, document } = await test.step('set parse to 1700 through the router, then open the viewer and, on a second page, the document', async () => {
      await graphql<SetAttributeAnswer>(request, PARSE_TO_1700)
      const viewer = await Viewer.open(viewerPage)
      const document = await Document.open(page)
      await expect(viewer.caption()).toHaveText('capacity 1400, bottleneck indexA, indexB')
      await expect(document.row('PIPE-R1').locator(':scope > .row .verdict')).toHaveText('FAIL capacity 1400 against 1500, limited by indexA, indexB')
      return { viewer, document }
    })
    const rowOf = (id: string) => document.row(id).locator(':scope > .row')

    await test.step('raise indexA from 700 to 900 in the viewer and expect the document to follow within two seconds', async () => {
      await viewer.edit('attribute|PIPE-S3|throughput', '900')
      await expectWithin(rowOf('PIPE-R1').locator('.verdict'), 'PASS capacity 1600 against 1500, limited by indexA, indexB')
      await expect(rowOf('PIPE-R1.3').locator('.verdict')).toHaveText('PASS throughput 900 against 750')
      await expect(rowOf('PIPE-R1.4').locator('.verdict')).toHaveText('FAIL throughput 700 against 750')
    })
  } finally {
    await test.step('put the demo back as it was found', async () => {
      await resetBoth(request)
      const viewer = await Viewer.open(page)
      await expect(viewer.caption()).toHaveText(SHIPPED_CAPTION)
    })
  }
})
