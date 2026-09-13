import { defineConfig } from 'checkly'
import { Frequency, RetryStrategyBuilder } from 'checkly/constructs'

export default defineConfig({
  projectName: 'sysml-federation',
  logicalId: 'sysml-federation',
  repoUrl: 'https://github.com/Roarge/sysml-federation',
  checks: {
    runtimeId: '2026.04',
    locations: ['eu-central-1'],
    tags: ['demo'],
    frequency: Frequency.EVERY_1H,
    checkMatch: '**/__checks__/**/*.check.ts',
    ignoreDirectoriesMatch: ['node_modules', 'test-results', 'playwright-report'],
    // The suite is declared in __checks__/suite.check.ts. Naming the Playwright
    // configuration here would construct a second suite outside every group.
    retryStrategy: RetryStrategyBuilder.noRetries(),
  },
  cli: { runLocation: 'eu-central-1', reporters: ['list'] },
})
