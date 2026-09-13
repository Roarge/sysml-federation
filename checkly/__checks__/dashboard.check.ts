// The public dashboard: every check tagged demo, with its P95 and P99, on
// pages of fifteen that refresh and turn every minute. A dashboard is served
// under a subdomain of the service's own domain, named by its slug, and is
// constructed only when CHECKLY_DASHBOARD_SLUG or CHECKLY_ACCOUNT_ID is set.

import { Dashboard } from 'checkly/constructs'

// pageSlug is the slug of one of the account's pages: the one given for it,
// or the project's name, the first eight characters of the account id, and
// the page's own suffix. A slug is unique across all Checkly users, which is
// why the account id is folded in: two accounts deploying this project must
// not claim one address. With neither a slug nor an account id there is no
// slug, and the page is not constructed. A fixed fallback was dropped because
// a login stored by checkly login puts the account id nowhere in the
// environment, and a fixed slug deployed from such a session could claim,
// or fail on, an address that a real deploy from another account holds.
export function pageSlug(given: string | undefined, suffix: string): string | undefined {
  if (given !== undefined && given !== '') {
    return given
  }
  const account = process.env.CHECKLY_ACCOUNT_ID
  if (account === undefined || account === '') {
    return undefined
  }
  return 'sysml-federation-' + account.slice(0, 8) + suffix
}

const slug = pageSlug(process.env.CHECKLY_DASHBOARD_SLUG, '')

if (slug !== undefined) {
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
}
