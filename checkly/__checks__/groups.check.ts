// The five groups of the session. A group carries the locations, the
// concurrency and the retry policy of every check in it, so a check states
// only what is its own. Every group is tagged demo, and with its own name.
//
// The demo has one shared in-memory state. The viewer, document and cross-app
// groups edit it, and a retried or a parallel run would edit it twice, so each
// of the three runs from one location, with no retries and no parallel
// locations, and alerts on the first failed run. That is the whole of the
// guard: a group's concurrency governs the runs a trigger or the API starts,
// nothing serialises the scheduled runs of two checks within or across the
// three groups, and two mutating checks due at the same moment can overlap.

import { AlertEscalationBuilder, CheckGroupV2, RetryStrategyBuilder } from 'checkly/constructs'
import type { CheckGroupV2Props } from 'checkly/constructs'
import { alertChannels } from './alerts.check'
import { privateLocation } from './private-location.check'

// Where a group's checks run: the public locations it names, or, when a
// private location is configured, that location alone. A group with a
// private location names no public one, so nothing runs from outside the
// demo's own network. This is the Team-plan route, untested on the free tier.
type Where = Pick<CheckGroupV2Props, 'locations' | 'privateLocations'>

function where(locations: NonNullable<CheckGroupV2Props['locations']>): Where {
  if (privateLocation === undefined) {
    return { locations }
  }
  return { privateLocations: [privateLocation] }
}

// The router group reads and never writes, so it may run from three
// locations, retry a failure in the same region, and alert on the second
// failed run rather than the first.
export const router = new CheckGroupV2('router', {
  name: 'Router',
  activated: true,
  tags: ['demo', 'router'],
  ...where(['eu-central-1', 'eu-west-2', 'us-east-1']),
  concurrency: 3,
  runParallel: false,
  retryStrategy: RetryStrategyBuilder.linearStrategy({ baseBackoffSeconds: 30, maxRetries: 2, sameRegion: true }),
  alertEscalationPolicy: AlertEscalationBuilder.runBasedEscalation(2),
  alertChannels,
})

// The viewer group edits the model through the viewer: one location, no
// retries, no parallel locations.
export const viewer = new CheckGroupV2('viewer', {
  name: 'Viewer',
  activated: true,
  tags: ['demo', 'viewer'],
  ...where(['eu-central-1']),
  concurrency: 1,
  runParallel: false,
  retryStrategy: RetryStrategyBuilder.noRetries(),
  alertEscalationPolicy: AlertEscalationBuilder.runBasedEscalation(1),
  alertChannels,
})

// The document group edits the document and the model through the document
// app: one location, no retries, no parallel locations.
export const document = new CheckGroupV2('document', {
  name: 'Document',
  activated: true,
  tags: ['demo', 'document'],
  ...where(['eu-central-1']),
  concurrency: 1,
  runParallel: false,
  retryStrategy: RetryStrategyBuilder.noRetries(),
  alertEscalationPolicy: AlertEscalationBuilder.runBasedEscalation(1),
  alertChannels,
})

// The cross-app group edits in one app and reads in the other: one location,
// no retries, no parallel locations.
export const crossApp = new CheckGroupV2('crossApp', {
  name: 'Cross-app',
  activated: true,
  tags: ['demo', 'crossApp'],
  ...where(['eu-central-1']),
  concurrency: 1,
  runParallel: false,
  retryStrategy: RetryStrategyBuilder.noRetries(),
  alertEscalationPolicy: AlertEscalationBuilder.runBasedEscalation(1),
  alertChannels,
})

// The session group holds the monitors: the three pages, the heartbeat and,
// when a hostname is set, the certificate, the name and the port. A group's
// channels apply only under a group policy, so the session alerts on the
// first failed run.
export const session = new CheckGroupV2('session', {
  name: 'Session',
  activated: true,
  tags: ['demo', 'session'],
  ...where(['eu-central-1']),
  concurrency: 1,
  runParallel: false,
  retryStrategy: RetryStrategyBuilder.noRetries(),
  alertEscalationPolicy: AlertEscalationBuilder.runBasedEscalation(1),
  alertChannels,
})
