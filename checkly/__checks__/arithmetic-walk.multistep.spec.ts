// The arithmetic walk: the four rows of the capacity table, read through the
// router one after another. As shipped, then ingest at 3000, then parse at
// 1700, then indexA at 900, each row asserting the capacity, the bottleneck
// and every verdict the table lists. The edits are the story's own, US-04,
// and the last row is the state US-10 is watched from. The demo is put back
// as it was found.

import { expect, test } from '@playwright/test'
import type { APIRequestContext } from '@playwright/test'
import { graphql, resetBoth } from './lib/graphql'

type Verdict = 'PASS' | 'FAIL' | 'INCONCLUSIVE' | 'ERROR'
type RowAnswer = {
  model: { requirements: { id: string, verdict: Verdict }[] }
  part: { capacity: number | null, bottleneck: { name: string }[] }
}
type SetAttributeAnswer = { setAttribute: { id: string } }

// ROW reads what one row of the table states: the capacity and the
// bottleneck of the pipeline, and the verdict of every requirement.
const ROW = '{ model { requirements { id verdict } } part(id: "PIPE-P1") { capacity bottleneck { name } } }'

const SET_ATTRIBUTE = `mutation SetAttribute($partId: ID!, $name: String!, $value: Float!) {
  setAttribute(partId: $partId, name: $name, value: $value) { id }
}`

// The requirements in the order the table lists them: the pipeline's, then
// the derived ones on ingest, parse, indexA, indexB and serve.
const DERIVED = ['PIPE-R1.1', 'PIPE-R1.2', 'PIPE-R1.3', 'PIPE-R1.4', 'PIPE-R1.5']

// A row of the table: the capacity, the servers in the cut, PIPE-R1's
// verdict and the five derived verdicts in server order.
type Row = { capacity: number, bottleneck: string[], pipeline: Verdict, derived: Verdict[] }

async function expectRow(request: APIRequestContext, want: Row): Promise<void> {
  const answer = await graphql<RowAnswer>(request, ROW)
  const verdictOf = new Map(answer.model.requirements.map((r) => [r.id, r.verdict]))
  expect(answer.part.capacity).toBe(want.capacity)
  expect(answer.part.bottleneck.map((b) => b.name)).toEqual(want.bottleneck)
  expect(verdictOf.get('PIPE-R1')).toBe(want.pipeline)
  expect(DERIVED.map((id) => verdictOf.get(id))).toEqual(want.derived)
}

async function raise(request: APIRequestContext, partId: string, value: number): Promise<void> {
  await graphql<SetAttributeAnswer>(request, SET_ATTRIBUTE, { partId, name: 'throughput', value })
}

test('the four rows of the arithmetic table hold through the router', async ({ request }) => {
  try {
    await test.step('as shipped: capacity 1200, cut parse, PIPE-R1 FAIL, derived PASS FAIL FAIL FAIL PASS', async () => {
      await expectRow(request, { capacity: 1200, bottleneck: ['parse'], pipeline: 'FAIL', derived: ['PASS', 'FAIL', 'FAIL', 'FAIL', 'PASS'] })
    })

    await test.step('ingest at 3000: nothing moves, the same number and the same cut', async () => {
      await raise(request, 'PIPE-S1', 3000)
      await expectRow(request, { capacity: 1200, bottleneck: ['parse'], pipeline: 'FAIL', derived: ['PASS', 'FAIL', 'FAIL', 'FAIL', 'PASS'] })
    })

    await test.step('parse at 1700: capacity 1400, the cut moves to indexA and indexB, PIPE-R1 FAIL, derived PASS PASS FAIL FAIL PASS', async () => {
      await raise(request, 'PIPE-S2', 1700)
      await expectRow(request, { capacity: 1400, bottleneck: ['indexA', 'indexB'], pipeline: 'FAIL', derived: ['PASS', 'PASS', 'FAIL', 'FAIL', 'PASS'] })
    })

    await test.step('indexA at 900: capacity 1600, PIPE-R1 PASS, derived PASS PASS PASS FAIL PASS', async () => {
      await raise(request, 'PIPE-S3', 900)
      await expectRow(request, { capacity: 1600, bottleneck: ['indexA', 'indexB'], pipeline: 'PASS', derived: ['PASS', 'PASS', 'PASS', 'FAIL', 'PASS'] })
    })
  } finally {
    await test.step('put the demo back as it was found', async () => {
      await resetBoth(request)
      await expectRow(request, { capacity: 1200, bottleneck: ['parse'], pipeline: 'FAIL', derived: ['PASS', 'FAIL', 'FAIL', 'FAIL', 'PASS'] })
    })
  }
})
