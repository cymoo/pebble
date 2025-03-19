export async function createPost(page, content) {
  await page.goto('/')
  await page.fill('[role="textbox"]', content)
  await page.click('[type="submit"]')
}

export async function deletePost(page, content) {
  // soft delete
  await page.goto('/')
  await page
    .locator(`div > p:has-text("${content}")`)
    .locator('xpath=./../../header/button')
    .click()
  await page.click('.text-destructive')

  // hard delete
  await page.goto('/?deleted=true')
  await page
    .locator(`div > p:has-text("${content}")`)
    .locator('xpath=./../../header/button')
    .click()

  await page.click('.text-destructive')
  await page.click('.bg-destructive')
}
