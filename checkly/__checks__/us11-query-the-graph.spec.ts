// US-11, query the graph: one query for PIPE-R1's text, verdict and document
// number is answered in one response from all three services, the served
// schema carries a type from each, and the playground is there for a
// visitor to send the same query by hand. Nothing is edited.

import { expect, test } from '@playwright/test'
import { demoUrl } from './lib/demo'
import { graphql } from './lib/graphql'

type JoinAnswer = {
  requirement: { text: string, verdict: string, verdictReason: string, documentNumber: string }
}
type SchemaAnswer = { __schema: { types: { name: string }[] } }

const JOIN = '{ requirement(id: "PIPE-R1") { text verdict verdictReason documentNumber } }'
const SCHEMA = '{ __schema { types { name } } }'

test('US-11 query the graph', async ({ page, request }) => {
  await test.step('send the join query and expect one answer carrying the text, the verdict and the document number', async () => {
    const answer = await graphql<JoinAnswer>(request, JOIN)
    expect(answer.requirement).toEqual({
      text: 'The pipeline shall sustain the required query rate',
      verdict: 'FAIL',
      verdictReason: 'capacity 1200 against 1500, limited by parse',
      documentNumber: '1',
    })
  })

  await test.step('read the served schema and expect a type from each of the three subgraphs', async () => {
    const answer = await graphql<SchemaAnswer>(request, SCHEMA)
    const names = answer.__schema.types.map((type) => type.name)
    expect(names).toEqual(expect.arrayContaining(['Model', 'Verdict', 'Document', 'Node']))
  })

  await test.step('open the playground and expect it to load', async () => {
    await page.goto(demoUrl() + '/playground')
    await expect(page).toHaveTitle(/Playground/)
  })
})
