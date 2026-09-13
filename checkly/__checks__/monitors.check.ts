// The monitors of the session group: the three pages every ten minutes, a
// heartbeat the session script pings, and, when DEMO_HOSTNAME is set, the
// public hostname's certificate, its name and its port. The hostname is the
// named tunnel's, handed in through the environment and never written here.

import {
  DnsAssertionBuilder,
  DnsMonitor,
  Frequency,
  HeartbeatMonitor,
  SslMonitor,
  TcpMonitor,
  UrlAssertionBuilder,
  UrlMonitor,
} from 'checkly/constructs'
import { session } from './groups.check'

const DEMO = '{{{DEMO_URL}}}'

new UrlMonitor('monitor-viewer', {
  name: 'monitor: the viewer answers',
  group: session,
  frequency: Frequency.EVERY_10M,
  request: {
    url: `${DEMO}/viewer/`,
    assertions: [UrlAssertionBuilder.statusCode().equals(200)],
  },
})

new UrlMonitor('monitor-document', {
  name: 'monitor: the document answers',
  group: session,
  frequency: Frequency.EVERY_10M,
  request: {
    url: `${DEMO}/document/`,
    assertions: [UrlAssertionBuilder.statusCode().equals(200)],
  },
})

// The root is followed through its redirect to the viewer.
new UrlMonitor('monitor-root', {
  name: 'monitor: the root answers',
  group: session,
  frequency: Frequency.EVERY_10M,
  request: {
    url: `${DEMO}/`,
    followRedirects: true,
    assertions: [UrlAssertionBuilder.statusCode().equals(200)],
  },
})

// The session script pings this monitor every ten minutes while it runs.
// Five minutes of grace, and then a missed ping is a failed session.
new HeartbeatMonitor('session-heartbeat', {
  name: 'session: the script is still running',
  group: session,
  period: 10,
  periodUnit: 'minutes',
  grace: 5,
  graceUnit: 'minutes',
})

const hostname = process.env.DEMO_HOSTNAME
if (hostname !== undefined && hostname !== '') {
  // The certificate is the edge's. Expiry alerts two weeks ahead.
  new SslMonitor('demo-certificate', {
    name: 'demo: the certificate is valid',
    group: session,
    frequency: Frequency.EVERY_1H,
    request: {
      hostname,
      sslConfig: { alertDaysBeforeExpiry: 14 },
    },
  })

  // The tunnel's CNAME is flattened at the edge, so an A query is what
  // answers, and any answer at all is the pass.
  new DnsMonitor('demo-dns', {
    name: 'demo: the name resolves',
    group: session,
    frequency: Frequency.EVERY_10M,
    request: {
      query: hostname,
      recordType: 'A',
      assertions: [DnsAssertionBuilder.textAnswer().notEmpty()],
    },
  })

  new TcpMonitor('demo-tcp', {
    name: 'demo: port 443 accepts a connection',
    group: session,
    frequency: Frequency.EVERY_10M,
    request: {
      hostname,
      port: 443,
      ipFamily: 'IPv4',
    },
  })
}
