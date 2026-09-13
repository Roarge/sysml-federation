// The story suite: the twelve browser specs run as one Playwright project on
// the service's runner, daily, from the viewer group so that it runs one at
// a time beside the group's own checks. The configuration is the shipped
// playwright.config.ts beside this directory, whose testMatch selects the
// us*.spec.ts files and nothing else.

import { Frequency, PlaywrightCheck } from 'checkly/constructs'
import { viewer } from './groups.check'

new PlaywrightCheck('story-suite', {
  name: 'The story suite',
  playwrightConfigPath: '../playwright.config.ts',
  installCommand: 'npm ci',
  testCommand: 'npx playwright test',
  pwProjects: ['chromium'],
  group: viewer,
  frequency: Frequency.EVERY_24H,
  tags: ['demo', 'suite'],
})
