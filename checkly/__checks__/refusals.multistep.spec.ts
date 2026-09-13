// The refusals walk: two edits the router refuses, with the model version
// read before and after to show that nothing moved. Each step is one request
// to the router. Nothing changes, so there is nothing to put back afterwards.

import { expect, test } from '@playwright/test'
import type { APIRequestContext } from '@playwright/test'
import { demoUrl } from './lib/demo'
import { graphql } from './lib/graphql'

type VersionAnswer = { model: { version: number } }

// A refusal as the router sends it: the service's sentence as the first
// error's message, and no data at all.
type RawAnswer = { data?: Record<string, unknown> | null, errors?: { message: string }[] }

const VERSION = '{ model { version } }'

// PIPE-R1.3's limit is bound to an expression of PIPE-R1's, so an edit has no
// literal to land on. PIPE-S1's throughput is a literal, and a negative
// number is refused before the source is touched.
const DERIVED_LIMIT = 'mutation { setLimit(requirementId: "PIPE-R1.3", value: 900) { id } }'
const NEGATIVE_THROUGHPUT = 'mutation { setAttribute(partId: "PIPE-S1", name: "throughput", value: -5) { id } }'

// The two sentences the record carries, unwrapped.
const NOT_A_LITERAL = 'the value is not a literal in the source: limit of requirement "PIPE-R1.3"'
const NOT_A_NUMBER = 'the value must be a finite, non-negative number: got -5'

// refusal posts a mutation the router is expected to refuse and answers the
// body as it came. The shape of a refusal is what the steps assert, and the
// shared helper reads that shape away by throwing on the first error.
async function refusal(request: APIRequestContext, query: string): Promise<RawAnswer> {
  const response = await request.post(demoUrl() + '/graphql', {
    headers: { 'Content-Type': 'application/json' },
    data: { query },
  })
  expect(response.status()).toBe(200)
  return (await response.json()) as RawAnswer
}

test('the router refuses a bound limit and a negative value, and the version stands', async ({ request }) => {
  const before = await test.step('read the model version', async () => {
    const answer = await graphql<VersionAnswer>(request, VERSION)
    expect(answer.model.version).toBeGreaterThan(0)
    return answer.model.version
  })

  await test.step('a derived limit is refused as not a literal in the source', async () => {
    const answer = await refusal(request, DERIVED_LIMIT)
    expect(answer.errors?.[0]?.message).toBe(NOT_A_LITERAL)
    expect(answer.data).toBeNull()
  })

  await test.step('a negative throughput is refused as not a finite, non-negative number', async () => {
    const answer = await refusal(request, NEGATIVE_THROUGHPUT)
    expect(answer.errors?.[0]?.message).toBe(NOT_A_NUMBER)
    expect(answer.data).toBeNull()
  })

  await test.step('the model version is unchanged', async () => {
    const answer = await graphql<VersionAnswer>(request, VERSION)
    expect(answer.model.version).toBe(before)
  })
})
