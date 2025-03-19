// @ts-check
import { expect, test } from '@playwright/test'

test.describe('login', () => {
  // Clear storage state
  let context
  test.beforeEach(async ({ browser }) => {
    context = await browser.newContext({ storageState: undefined })
  })

  test.afterEach(async () => {
    await context.close()
  })

  test('can navigate to login page when not login', async () => {
    const page = await context.newPage()
    await page.goto('/')
    await page.waitForURL('/login')
  })

  test('can login', async () => {
    const page = await context.newPage()
    await page.goto('/login')
    const password = page.locator('[type=password]')

    await password.fill(process.env.PASSWORD || '')
    await password.press('Enter')

    await page.waitForURL('/')
  })

  test('can not login when password is wrong', async () => {
    const page = await context.newPage()
    await page.goto('/login')
    const password = page.locator('[type=password]')

    await password.fill('foo')
    await password.press('Enter')

    await expect(page.locator('.text-destructive')).toBeVisible()
    await expect(page).toHaveTitle(/login/i)
  })
})
