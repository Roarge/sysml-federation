// graphql posts one operation to the demo's router and answers its data, the
// way the two shipped apps do. A refusal is reported with the sentence the
// router carries in its first error.

import type { APIRequestContext } from '@playwright/test'
import { demoUrl } from './demo'

type GraphQLError = { message: string }
type GraphQLAnswer<T> = { data?: T | null, errors?: GraphQLError[] }

type ResetAnswer = {
  resetModel: { version: number }
  resetDocument: { version: number }
}

// RESET is the document both apps send from their Reset button: the two root
// fields travel in one mutation, the way the apps send them.
const RESET = 'mutation Reset { resetModel { version } resetDocument { version } }'

export async function graphql<T>(
  request: APIRequestContext,
  query: string,
  variables?: Record<string, string | number | boolean>,
): Promise<T> {
  const response = await request.post(demoUrl() + '/graphql', {
    headers: { 'Content-Type': 'application/json' },
    data: { query, variables },
  })
  const answer = parse<T>(response.status(), await response.text())
  if (answer.errors !== undefined && answer.errors.length > 0) {
    throw new Error(answer.errors[0].message)
  }
  if (answer.data === undefined || answer.data === null) {
    throw new Error(`the router answered ${response.status()} without data`)
  }
  return answer.data
}

// parse reads a body as the object a GraphQL answer always is, and throws
// naming the status and the body's first line when it is not one: a tunnel
// answering a 502 page, or an empty body, is reported as what it is rather
// than as a syntax error.
function parse<T>(status: number, body: string): GraphQLAnswer<T> {
  let parsed: unknown
  try {
    parsed = JSON.parse(body)
  } catch {
    parsed = undefined
  }
  if (typeof parsed !== 'object' || parsed === null) {
    const line = body.split('\n', 1)[0].trim()
    throw new Error(`the router answered ${status} with a body that is not a JSON object: ${line}`)
  }
  return parsed as GraphQLAnswer<T>
}

// resetBoth puts the model and the document back to their shipped state.
export async function resetBoth(request: APIRequestContext): Promise<void> {
  await graphql<ResetAnswer>(request, RESET)
}
