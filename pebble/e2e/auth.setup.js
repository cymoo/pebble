import { test as setup } from '@playwright/test'
import fs from 'fs'
import path from 'path';


const authFile = 'tests/.auth/user.json';

setup('authenticate', async ({ page }) => {

  if (!fs.existsSync(path.dirname(authFile))) {
    fs.mkdirSync(path.dirname(authFile))
  }

  await page.goto('/login')
  const password = page.locator('[type=password]')

  await password.fill(process.env.PASSWORD || '')
  await password.press('Enter')
  await page.waitForURL('/')

  await page.context().storageState({ path: authFile });
});
