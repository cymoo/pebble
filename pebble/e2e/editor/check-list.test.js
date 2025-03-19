import { expect, test } from '@playwright/test'

test.describe('check list', () => {
  test('should toggle check list', async ({ page }) => {
    await page.goto('/')

    const editor = page.getByRole('textbox')
    const toggleBtn = page.getByRole('button', { name: 'check-list' })
    const todoItem = editor.locator('.check-list')

    await editor.click()
    await editor.fill('todo 1')

    // change text to check-list
    await toggleBtn.click()
    await expect(todoItem).toHaveCount(1)
    await expect(todoItem.locator('input')).not.toBeChecked()

    // mark as completed
    await todoItem.locator('input').check()
    await expect(todoItem.locator('input')).toBeChecked()

    // change check-list to text
    await toggleBtn.click()
    await expect(todoItem).toHaveCount(0)
  })
})
