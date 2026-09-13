// Viewer wraps the model viewer at /viewer/ in the elements a check reads
// and the controls it drives. Every method but open and edit answers a
// Locator, so a check asserts against the live page.

import type { Locator, Page } from '@playwright/test'
import { demoUrl } from './demo'

export class Viewer {
  private constructor(private readonly page: Page) {}

  static async open(page: Page): Promise<Viewer> {
    await page.goto(demoUrl() + '/viewer/')
    return new Viewer(page)
  }

  // status is the line under the header: empty after a successful round
  // trip, a sentence in red after a refusal.
  status(): Locator {
    return this.page.locator('#status')
  }

  // version is the "version N" mark beside the model text heading.
  version(): Locator {
    return this.page.locator('#version')
  }

  // caption is the capacity and bottleneck line under the wiring sketch.
  caption(): Locator {
    return this.page.locator('#sketch .caption')
  }

  // control is one editable number field, by the key the app gives it:
  // "attribute|<part id>|<name>" or "limit|<requirement id>".
  control(key: string): Locator {
    return this.page.locator(`#inputs input[data-key="${key}"]`)
  }

  // requirement is one requirement block with its verdict, by id.
  requirement(id: string): Locator {
    return this.page.locator(`#requirements .req[data-id="${id}"]`)
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
