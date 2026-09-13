// The story checks, constructed from the manifest. Every entry of checks.json
// whose kind is BrowserCheck or MultiStepCheck becomes one check here, driven
// by the spec file the entry names: the twelve browser specs, the refusals
// walk and the two multistep walks. The manifest is the one list of what the
// session runs, so a spec is added by adding its entry, its file and its
// display name in the NAMES map below, and the parse throws without the
// name. No check is constructed by hand in this file. Entries of other kinds
// are constructed in their own files and ignored here.

import path from 'node:path'
import { BrowserCheck, Frequency, MultiStepCheck } from 'checkly/constructs'
import type { CheckGroupV2 } from 'checkly/constructs'
import { crossApp, document, router, session, viewer } from './groups.check'
import manifest from './checks.json'

// ManifestEntry is one row of checks.json. deployed is a boolean, or the word
// whenHostnameSet for a monitor that exists only when a hostname is set, and
// a check on no schedule carries no frequencyMinutes at all.
interface ManifestEntry {
  id: string
  kind: string
  file: string
  story: string
  frequencyMinutes?: number
  locations: string[]
  deployed: boolean | 'whenHostnameSet'
  mutating: boolean
  group: string
  target: string
}

const entries = manifest as ManifestEntry[]

// The groups by the name the manifest writes in its group field.
const groups: Record<string, CheckGroupV2> = { router, viewer, document, crossApp, session }

// The display name of each spec-driven check, by its id. A missing name is a
// fault in this file rather than a check with a bare id, so it throws.
const NAMES: Record<string, string> = {
  'refusals': 'refusals: the router refuses a bound limit and a negative value',
  'us01-launch': 'US-01: launch with one command',
  'us02-read-the-model': 'US-02: read the model',
  'us03-nothing-moves': 'US-03: raise a server that is not the bottleneck',
  'us04-bottleneck-moves': 'US-04: raise the bottleneck',
  'us05-tighten-limit': 'US-05: tighten the limit',
  'us06-read-document': 'US-06: read the document',
  'us07-reorder-and-nest': 'US-07: reorder and nest',
  'us08-shape-the-document': 'US-08: shape the document',
  'us09-edit-from-document': 'US-09: change a value from the document',
  'us10-viewer-to-document': 'US-10: change from the viewer, watch the document',
  'us11-query-the-graph': 'US-11: query the graph',
  'us12-reset': 'US-12: reset',
  'arithmetic-walk': 'the arithmetic walk: the four rows of the capacity table',
  'document-operations': 'the document operations: six edits that leave the model alone',
}

// frequencyOf maps the minutes the manifest writes to the schedule the
// service offers. The four values the manifest uses are the four accepted,
// and any other number is a fault in the manifest.
function frequencyOf(minutes: number | undefined): Frequency {
  switch (minutes) {
    case 10: return Frequency.EVERY_10M
    case 60: return Frequency.EVERY_1H
    case 360: return Frequency.EVERY_6H
    case 1440: return Frequency.EVERY_24H
    default: throw new Error(`no schedule for a check every ${minutes} minutes`)
  }
}

function nameOf(entry: ManifestEntry): string {
  const name = NAMES[entry.id]
  if (name === undefined) {
    throw new Error(`the manifest entry ${entry.id} has no display name`)
  }
  return name
}

function groupOf(entry: ManifestEntry): CheckGroupV2 {
  const group = groups[entry.group]
  if (group === undefined) {
    throw new Error(`the manifest entry ${entry.id} names the group ${entry.group}, and no such group exists`)
  }
  return group
}

// The options every spec-driven check shares. An entry with deployed false
// is testOnly: it runs in a test session and is never deployed.
function optionsOf(entry: ManifestEntry) {
  return {
    name: nameOf(entry),
    group: groupOf(entry),
    frequency: frequencyOf(entry.frequencyMinutes),
    testOnly: entry.deployed === false,
    tags: ['demo', entry.story.toLowerCase()],
    code: { entrypoint: path.join(__dirname, entry.file) },
  }
}

for (const entry of entries) {
  switch (entry.kind) {
    case 'BrowserCheck':
      new BrowserCheck(entry.id, optionsOf(entry))
      break
    case 'MultiStepCheck':
      new MultiStepCheck(entry.id, optionsOf(entry))
      break
    default:
      break
  }
}
