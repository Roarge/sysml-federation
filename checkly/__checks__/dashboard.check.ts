// The public dashboard: every check tagged demo, with its P95 and P99, on
// pages of fifteen that refresh and turn every minute. A dashboard is served
// under a subdomain of the service's own domain, named by its slug.

import { Dashboard } from 'checkly/constructs'

// accountSlug is the slug of one of the account's pages when none is given:
// the project's name, the first eight characters of the account id, and the
// page's own suffix. A slug is unique across all Checkly users, which is why
// the account id is folded in: two accounts deploying this project must not
// claim one address. Without an account id, as when the file loads with no
// credentials, the slug is a fixed local one.
export function accountSlug(suffix: string): string {
  const account = process.env.CHECKLY_ACCOUNT_ID
  if (account === undefined || account === '') {
    return 'sysml-federation-local' + suffix
  }
  return 'sysml-federation-' + account.slice(0, 8) + suffix
}

const given = process.env.CHECKLY_DASHBOARD_SLUG
const slug = given !== undefined && given !== '' ? given : accountSlug('')

new Dashboard('sysml-federation-dashboard', {
  customUrl: slug,
  header: 'sysml-federation',
  description: 'The demo, checked from outside',
  tags: ['demo'],
  showP95: true,
  showP99: true,
  refreshRate: 60,
  paginate: true,
  paginationRate: 60,
  checksPerPage: 15,
  hideTags: false,
  isPrivate: false,
})
