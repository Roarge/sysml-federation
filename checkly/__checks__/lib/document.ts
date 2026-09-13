// Document wraps the requirements document at /document/ in the elements a
// check reads and the controls it drives. Every method but open and edit
// answers a Locator. The heading and prose buttons open a prompt dialog,
// which a check answers with page.on('dialog', d => d.accept(text)) before
// it clicks.

import type { Locator, Page } from '@playwright/test'
import { demoUrl } from './demo'

export class Document {
  private constructor(private readonly page: Page) {}

  static async open(page: Page): Promise<Document> {
    await page.goto(demoUrl() + '/document/')
    return new Document(page)
  }

  // versions is the "document version N, model version M" line.
  versions(): Locator {
    return this.page.locator('#versions')
  }

  // row is one node of the tree by its id. In the shipped tree a requirement
  // node carries its requirement's id.
  row(id: string): Locator {
    return this.page.locator(`#tree li.node[data-id="${id}"]`)
  }

  // control is one editable number field, by the key the app gives it:
  // "attribute|<part id>|<name>" or "limit|<requirement id>".
  control(key: string): Locator {
    return this.page.locator(`input[data-key="${key}"]`)
  }

  headingAbove(id: string): Locator {
    return this.page.locator(`button[data-act="heading"][data-id="${id}"]`)
  }

  addProse(id: string): Locator {
    return this.page.locator(`button[data-act="prose"][data-id="${id}"]`)
  }

  exclude(reqId: string): Locator {
    return this.page.locator(`button[data-act="exclude"][data-req="${reqId}"]`)
  }

  // restore is the button beside an excluded requirement in the tray.
  restore(): Locator {
    return this.page.locator('#excluded button[data-act="include"]')
  }

  tray(): Locator {
    return this.page.locator('#excluded')
  }

  // depthMarker is the sentence the deepest row shows in place of the
  // controls that would nest anything further.
  depthMarker(): Locator {
    return this.page.locator('#tree .deeper')
  }

  reset(): Locator {
    return this.page.locator('#reset')
  }

  // edit types a value into a control and tabs away, which is what fires
  // the app's change handler and sends the mutation.
  async edit(key: string, value: string): Promise<void> {
    const field = this.control(key)
    await field.fill(value)
    await field.press('Tab')
  }
}
