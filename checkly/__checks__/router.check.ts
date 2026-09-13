// The seven API checks on the router: four queries through the composed
// graph and three of the four paths on the one port. Each is one request and
// its assertions, run hourly from the router group's three locations. Every
// URL is the demo's address as the session hands it in, plus a path.

import { ApiCheck, AssertionBuilder, Frequency } from 'checkly/constructs'
import type { HttpHeader } from 'checkly/constructs'
import { router } from './groups.check'

const DEMO = '{{{DEMO_URL}}}'

const JSON_BODY: HttpHeader = { key: 'Content-Type', value: 'application/json' }

// The response-time thresholds shared by the seven: slow above two seconds,
// failed above ten.
const timing = {
  frequency: Frequency.EVERY_1H,
  degradedResponseTime: 2000,
  maxResponseTime: 10000,
}

// graphqlBody is the JSON body of one query, the way the two apps send it.
function graphqlBody(query: string): string {
  return JSON.stringify({ query })
}

new ApiCheck('router-version', {
  name: 'router: the model version',
  group: router,
  ...timing,
  request: {
    method: 'POST',
    url: `${DEMO}/graphql`,
    headers: [JSON_BODY],
    bodyType: 'JSON',
    body: graphqlBody('{ model { version } }'),
    followRedirects: true,
    assertions: [
      AssertionBuilder.statusCode().equals(200),
      AssertionBuilder.jsonBody('$.data.model.version').greaterThan(0),
    ],
  },
})

new ApiCheck('router-join', {
  name: 'router: one requirement from three services',
  group: router,
  ...timing,
  request: {
    method: 'POST',
    url: `${DEMO}/graphql`,
    headers: [JSON_BODY],
    bodyType: 'JSON',
    body: graphqlBody('{ requirement(id: "PIPE-R1") { text verdict verdictReason documentNumber } }'),
    followRedirects: true,
    assertions: [
      AssertionBuilder.statusCode().equals(200),
      AssertionBuilder.jsonBody('$.data.requirement.text').equals('The pipeline shall sustain the required query rate'),
      AssertionBuilder.jsonBody('$.data.requirement.verdict').notEmpty(),
      AssertionBuilder.jsonBody('$.data.requirement.documentNumber').equals('1'),
    ],
  },
})

// The pipeline is the part with a bottleneck: a single server's capacity is
// its own throughput and its bottleneck is empty. The first bottleneck's id
// is what is asserted, so an empty list fails the check.
new ApiCheck('router-capacity-bottleneck', {
  name: 'router: the capacity and the bottleneck',
  group: router,
  ...timing,
  request: {
    method: 'POST',
    url: `${DEMO}/graphql`,
    headers: [JSON_BODY],
    bodyType: 'JSON',
    body: graphqlBody('{ part(id: "PIPE-P1") { capacity bottleneck { id } } }'),
    followRedirects: true,
    assertions: [
      AssertionBuilder.statusCode().equals(200),
      AssertionBuilder.jsonBody('$.data.part.capacity').isNotNull(),
      AssertionBuilder.jsonBody('$.data.part.bottleneck[0].id').isNotNull(),
    ],
  },
})

new ApiCheck('router-introspection', {
  name: 'router: types from each subgraph in one schema',
  group: router,
  ...timing,
  request: {
    method: 'POST',
    url: `${DEMO}/graphql`,
    headers: [JSON_BODY],
    bodyType: 'JSON',
    body: graphqlBody('{ __schema { types { name } } }'),
    followRedirects: true,
    assertions: [
      AssertionBuilder.statusCode().equals(200),
      AssertionBuilder.textBody().contains('Verdict'),
      AssertionBuilder.textBody().contains('Document'),
      // Node and Model are substrings of other type names, so the two are
      // asserted as the whole name field, the way the router writes it.
      AssertionBuilder.textBody().contains('"name":"Node"'),
      AssertionBuilder.textBody().contains('"name":"Model"'),
    ],
  },
})

new ApiCheck('router-playground', {
  name: 'router: the playground is served',
  group: router,
  ...timing,
  request: {
    method: 'GET',
    url: `${DEMO}/playground`,
    headers: [{ key: 'Accept', value: 'text/html' }],
    followRedirects: true,
    assertions: [
      AssertionBuilder.statusCode().equals(200),
      AssertionBuilder.textBody().contains('<html'),
    ],
  },
})

// The redirect itself is the answer, so it is not followed.
new ApiCheck('router-root-redirect', {
  name: 'router: the root redirects to the viewer',
  group: router,
  ...timing,
  request: {
    method: 'GET',
    url: `${DEMO}/`,
    followRedirects: false,
    assertions: [
      AssertionBuilder.statusCode().equals(302),
      AssertionBuilder.headers('location').equals('/viewer/'),
    ],
  },
})

// The router's own health path is not proxied, so a 404 is the pass.
new ApiCheck('router-health-not-proxied', {
  name: 'router: the health path is not proxied',
  group: router,
  ...timing,
  shouldFail: true,
  request: {
    method: 'GET',
    url: `${DEMO}/health/ready`,
    followRedirects: true,
    assertions: [
      AssertionBuilder.statusCode().equals(404),
    ],
  },
})
