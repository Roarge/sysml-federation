// The document operations: the six editorial mutations the document app
// sends, through the router, one step each, with the document each one
// answers read for the numbering and the placement it states. The model
// version is read before and after and is the same, since none of this
// touches the model. The moves are US-07's and the heading, the prose and
// the exclusion are US-08's. The demo is put back as it was found.

import { expect, test } from '@playwright/test'
import type { APIRequestContext } from '@playwright/test'
import { graphql, resetBoth } from './lib/graphql'

type Node = { id: string, kind: 'HEADING' | 'PROSE' | 'REQUIREMENT', number: string | null, text: string | null, children?: Node[] }
type Document = { version: number, nodes: Node[] }
type VersionAnswer = { model: { version: number } }
type DocumentAnswer = { document: Document }
type MutationAnswer = Record<string, Document>

const VERSION = '{ model { version } }'

// DOCUMENT selects the tree four levels deep, which is as deep as this walk
// nests anything: a heading, a requirement under it, and its children.
const DOCUMENT = `{
  version
  nodes {
    id kind number text
    children {
      id kind number text
      children {
        id kind number text
        children { id kind number text }
      }
    }
  }
}`

const READ_DOCUMENT = `{ document ${DOCUMENT} }`

// mutate sends one of the document's mutations and answers the document it
// returns. Every mutation of the document service answers the whole
// document, so one shape reads them all.
async function mutate(request: APIRequestContext, field: string, args: string): Promise<Document> {
  const answer = await graphql<MutationAnswer>(request, `mutation { ${field}(${args}) ${DOCUMENT} }`)
  return answer[field]
}

// flatten lists every node of the tree in document order.
function flatten(nodes: Node[]): Node[] {
  return nodes.flatMap((node) => [node, ...flatten(node.children ?? [])])
}

function find(document: Document, id: string): Node {
  const node = flatten(document.nodes).find((n) => n.id === id)
  if (node === undefined) {
    throw new Error(`the document carries no node ${id}`)
  }
  return node
}

function numberOf(document: Document, id: string): string | null {
  return find(document, id).number
}

// childIds lists the ids under one node, in order.
function childIds(document: Document, id: string): string[] {
  return (find(document, id).children ?? []).map((child) => child.id)
}

test('the six document operations renumber the document and leave the model alone', async ({ request }) => {
  try {
    const before = await test.step('read the model version and the shipped document', async () => {
      const version = (await graphql<VersionAnswer>(request, VERSION)).model.version
      const shipped = (await graphql<DocumentAnswer>(request, READ_DOCUMENT)).document
      expect(childIds(shipped, 'PIPE-R1')).toEqual(['PIPE-R1.1', 'PIPE-R1.2', 'PIPE-R1.3', 'PIPE-R1.4', 'PIPE-R1.5'])
      expect(numberOf(shipped, 'PIPE-R2')).toBe('2')
      return version
    })

    await test.step('moveNode: PIPE-R1.5 above PIPE-R1.1, numbered 1.1 with the others shifted to 1.2 to 1.5', async () => {
      const document = await mutate(request, 'moveNode', 'id: "PIPE-R1.5", parentId: "PIPE-R1", index: 0')
      expect(childIds(document, 'PIPE-R1')).toEqual(['PIPE-R1.5', 'PIPE-R1.1', 'PIPE-R1.2', 'PIPE-R1.3', 'PIPE-R1.4'])
      expect(numberOf(document, 'PIPE-R1.5')).toBe('1.1')
      expect(numberOf(document, 'PIPE-R1.4')).toBe('1.5')
      expect(numberOf(document, 'PIPE-R2')).toBe('2')
    })

    const headingId = await test.step('insertHeading: a heading above PIPE-R1 takes number 1, PIPE-R1 becomes 1.1 and its children 1.1.1 to 1.1.5', async () => {
      const document = await mutate(request, 'insertHeading', 'aboveId: "PIPE-R1", text: "Performance"')
      const heading = flatten(document.nodes).find((node) => node.kind === 'HEADING' && node.text === 'Performance')
      if (heading === undefined) {
        throw new Error('the heading was not inserted')
      }
      expect(heading.number).toBe('1')
      expect(childIds(document, heading.id)).toEqual(['PIPE-R1'])
      expect(numberOf(document, 'PIPE-R1')).toBe('1.1')
      expect(numberOf(document, 'PIPE-R1.5')).toBe('1.1.1')
      expect(numberOf(document, 'PIPE-R1.4')).toBe('1.1.5')
      return heading.id
    })

    const proseId = await test.step('addProse: a paragraph under the heading, after PIPE-R1, with no number', async () => {
      const document = await mutate(request, 'addProse', `parentId: "${headingId}", index: 1, text: "A paragraph on the budget."`)
      const children = childIds(document, headingId)
      expect(children).toHaveLength(2)
      expect(children[0]).toBe('PIPE-R1')
      const prose = find(document, children[1])
      expect(prose.kind).toBe('PROSE')
      expect(prose.text).toBe('A paragraph on the budget.')
      expect(prose.number).toBeNull()
      return prose.id
    })

    await test.step('editText: the paragraph reads its new text', async () => {
      const document = await mutate(request, 'editText', `id: "${proseId}", text: "The paragraph, edited."`)
      expect(find(document, proseId).text).toBe('The paragraph, edited.')
      expect(find(document, proseId).number).toBeNull()
    })

    await test.step('excludeRequirement: PIPE-R1.4 leaves the document and the children that remain close up to 1.1.1 to 1.1.4', async () => {
      const document = await mutate(request, 'excludeRequirement', 'requirementId: "PIPE-R1.4"')
      expect(flatten(document.nodes).map((node) => node.id)).not.toContain('PIPE-R1.4')
      expect(childIds(document, 'PIPE-R1')).toEqual(['PIPE-R1.5', 'PIPE-R1.1', 'PIPE-R1.2', 'PIPE-R1.3'])
      expect(numberOf(document, 'PIPE-R1.3')).toBe('1.1.4')
    })

    await test.step('includeRequirement: PIPE-R1.4 returns as the last child of PIPE-R1, numbered 1.1.5', async () => {
      const document = await mutate(request, 'includeRequirement', 'requirementId: "PIPE-R1.4"')
      expect(childIds(document, 'PIPE-R1')).toEqual(['PIPE-R1.5', 'PIPE-R1.1', 'PIPE-R1.2', 'PIPE-R1.3', 'PIPE-R1.4'])
      expect(numberOf(document, 'PIPE-R1.4')).toBe('1.1.5')
    })

    await test.step('read the model version again and expect the one kept', async () => {
      const after = (await graphql<VersionAnswer>(request, VERSION)).model.version
      expect(after).toBe(before)
    })
  } finally {
    await test.step('put the demo back as it was found', async () => {
      await resetBoth(request)
      const shipped = (await graphql<DocumentAnswer>(request, READ_DOCUMENT)).document
      expect(shipped.nodes.map((node) => node.id)).toEqual(['intro', 'PIPE-R1', 'PIPE-R2'])
      expect(childIds(shipped, 'PIPE-R1')).toEqual(['PIPE-R1.1', 'PIPE-R1.2', 'PIPE-R1.3', 'PIPE-R1.4', 'PIPE-R1.5'])
    })
  }
})
