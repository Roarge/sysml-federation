// The alert channels of the session. The two addresses come from the
// environment at deploy time and are never committed: a mailbox in
// CHECKLY_ALERT_EMAIL and a webhook URL in CHECKLY_WEBHOOK_URL. A channel is
// constructed only when its variable is set, so a session that sets neither
// has an empty list and the groups alert nobody.

import { EmailAlertChannel, WebhookAlertChannel } from 'checkly/constructs'
import type { AlertChannel } from 'checkly/constructs'

// The webhook body is a Handlebars template the service fills in when it
// sends. Each field is one of the variables the service exposes.
const WEBHOOK_BODY = `{
  "title": "{{ALERT_TITLE}}",
  "type": "{{ALERT_TYPE}}",
  "check": "{{CHECK_NAME}}",
  "startedAt": "{{STARTED_AT}}",
  "result": "{{RESULT_LINK}}"
}`

export const alertChannels: AlertChannel[] = []

const email = process.env.CHECKLY_ALERT_EMAIL
if (email !== undefined && email !== '') {
  alertChannels.push(new EmailAlertChannel('email', {
    address: email,
    sendFailure: true,
    sendRecovery: true,
    sendDegraded: false,
    sslExpiry: true,
  }))
}

const webhook = process.env.CHECKLY_WEBHOOK_URL
if (webhook !== undefined && webhook !== '') {
  alertChannels.push(new WebhookAlertChannel('webhook', {
    name: 'sysml-federation webhook',
    url: webhook,
    method: 'POST',
    template: WEBHOOK_BODY,
    sendFailure: true,
    sendRecovery: true,
    sendDegraded: false,
  }))
}
