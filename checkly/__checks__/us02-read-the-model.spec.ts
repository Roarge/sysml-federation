// US-02, read the model: the text pane shows the model file, the sketch
// shows the wiring with its capacity and bottleneck, and PIPE-R1 shows its
// limit, its verdict and the reason. Nothing is edited.

import { expect, test } from '@playwright/test'
import { Viewer } from './lib/viewer'

const SHIPPED_CAPTION = 'capacity 1200, bottleneck parse'

test('US-02 read the model', async ({ page }) => {
  const viewer = await Viewer.open(page)

  await test.step('expect the text pane to show the model file', async () => {
    const text = page.locator('#text')
    await expect(text).toContainText('package')
    await expect(text).toContainText('PIPE-S2')
    await expect(text).toContainText('connect')
    await expect(text).toContainText('globalThroughput')
    expect(await page.locator('#text .tok-number').count()).toBeGreaterThan(0)
  })

  await test.step('expect the sketch to show the wiring with capacity 1200 and parse as the bottleneck', async () => {
    await expect(viewer.caption()).toHaveText(SHIPPED_CAPTION)
    await expect(page.locator('#sketch .box')).toHaveCount(5)
    await expect(page.locator('#sketch .box.bottleneck .label')).toHaveText('parse')
  })

  await test.step('expect PIPE-R1 to show its limit of 1500, a verdict of FAIL and a reason naming parse at 1200', async () => {
    const block = viewer.requirement('PIPE-R1')
    await expect(block.locator('.constraint')).toContainText('1500')
    await expect(block.locator('.verdict')).toHaveText('FAIL capacity 1200 against 1500, limited by parse')
    await expect(block).toHaveClass(/fail/)
  })
})
