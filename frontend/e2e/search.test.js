// @ts-check
import { expect, test } from '@playwright/test'

import { createPost, deletePost } from './helper.js'

test.describe('search', () => {
  let postContent

  test.beforeEach(async ({ page }) => {
    postContent = 'Wake up neo...the matrix has you...'
    await createPost(page, postContent)
  })

  test.afterEach(async ({ page }) => {
    await deletePost(page, postContent)
  })

  test('has search results', async ({ page }) => {
    await page.goto('/search')
    await page.fill('[type="search"]', 'neo matrix')
    await expect(page.locator('[data-testid="virtuoso-item-list"] > *')).not.toHaveCount(0)
  })

  test('has no search result', async ({ page }) => {
    await page.goto('/search')
    await page.fill('[type="search"]', 'rabbit')
    await expect(page.locator('[data-testid="virtuoso-item-list"] > *')).toHaveCount(0)
  })
})
