// A weekly maintenance window over every check tagged demo: five minutes
// early on a Sunday morning, repeating each week from the first one, during
// which a failed run alerts nobody. Maintenance windows are a paid feature,
// and the flag keeps the project deployable on the free tier: the window
// exists only when CHECKLY_MAINTENANCE is 1.

import { MaintenanceWindow } from 'checkly/constructs'

if (process.env.CHECKLY_MAINTENANCE === '1') {
  new MaintenanceWindow('weekly', {
    name: 'Weekly maintenance',
    tags: ['demo'],
    startsAt: new Date('2026-09-06T03:00:00Z'),
    endsAt: new Date('2026-09-06T03:05:00Z'),
    repeatInterval: 1,
    repeatUnit: 'WEEK',
  })
}
