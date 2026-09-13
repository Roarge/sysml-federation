// The status page: one card, The demo, holding four services named for the
// groups whose checks they stand for. A page of this generation is built
// from components that point at it: the card is a group component and each
// service a service component nested under it. A component names no check,
// and nothing in this code links a check to a service: a service's status is
// set by incidents, opened by hand or, on a paid plan, by the automation
// rules at the end of this file, which match a failing check by its tags.
// The slug follows the dashboard's rule, with a -status suffix when none is
// given, and the page is constructed only when CHECKLY_STATUS_SLUG or
// CHECKLY_ACCOUNT_ID is set, for the reason the dashboard file gives.

import { StatusPageV3, StatusPageV3AutomationRule, StatusPageV3Component } from 'checkly/constructs'
import { pageSlug } from './dashboard.check'

const slug = pageSlug(process.env.CHECKLY_STATUS_SLUG, '-status')

// statusPage builds the page, its card and its services under one slug.
function statusPage(url: string): void {
  const page = new StatusPageV3('sysml-federation-status', {
    name: 'sysml-federation',
    url,
    defaultTheme: 'AUTO',
  })

  const card = new StatusPageV3Component('card-the-demo', {
    statusPage: page,
    type: 'GROUP',
    name: 'The demo',
    displayOrder: 1,
  })

  // The four services in the order they are shown, each with the tag of the
  // group whose checks it stands for.
  const services: Array<{ name: string, group: string }> = [
    { name: 'Viewer', group: 'viewer' },
    { name: 'Document', group: 'document' },
    { name: 'Router', group: 'router' },
    { name: 'Session', group: 'session' },
  ]

  services.forEach(({ name, group }, index) => {
    const component = new StatusPageV3Component('service-' + group, {
      statusPage: page,
      type: 'SERVICE',
      name,
      displayOrder: index + 1,
      parent: card,
    })

    // Incident automation is a paid feature, and the flag keeps the project
    // deployable on the free tier: a rule exists only when CHECKLY_INCIDENTS
    // is 1. A rule matches a failing check by tag, and a check carries its
    // group's tag through the group, so a rule tagged with the group's own
    // name opens an incident on its own service and no other, and closes it
    // when the check recovers.
    if (process.env.CHECKLY_INCIDENTS === '1') {
      new StatusPageV3AutomationRule('rule-' + name, {
        statusPage: page,
        name,
        tags: [group],
        firstUpdate: 'The demo is degraded',
        lastUpdate: 'The demo is back',
        components: [{ component, targetImpact: 'PARTIAL_OUTAGE' }],
      })
    }
  })
}

if (slug !== undefined) {
  statusPage(slug)
}
