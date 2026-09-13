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
// fields travel in one mutation, and each is applied even if the other fails.
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
  const answer = (await response.json()) as GraphQLAnswer<T>
  if (answer.errors !== undefined && answer.errors.length > 0) {
    throw new Error(answer.errors[0].message)
  }
  if (answer.data === undefined || answer.data === null) {
    throw new Error(`the router answered ${response.status()} without data`)
  }
  return answer.data
}

// resetBoth puts the model and the document back to their shipped state.
export async function resetBoth(request: APIRequestContext): Promise<void> {
  await graphql<ResetAnswer>(request, RESET)
}
