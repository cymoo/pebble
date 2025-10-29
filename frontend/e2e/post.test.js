import { expect, test } from '@playwright/test'

import { createPost, deletePost } from './helper.js'

test('will jump to 404 page when post not found', async ({ page }) => {
  await page.goto('/p/0')
  await expect(page).toHaveURL('/404')
})

test('can create a post and delete it', async ({ page }) => {
  await createPost(page, 'wakeup neo')
  await deletePost(page, 'wakeup neo')
})
