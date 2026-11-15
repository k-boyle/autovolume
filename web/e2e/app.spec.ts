import { expect, test } from '@playwright/test'

test('signs in, completes setup and shows the library', async ({ page }) => {
  await page.goto('/')
  await expect(page).toHaveURL('/login')

  await page.getByRole('button', { name: 'Sign in with Plex' }).click()
  await expect(page.getByRole('heading', { name: 'Set up AutoVolume' })).toBeVisible({ timeout: 10_000 })

  await page.getByRole('button', { name: /Basement NAS/ }).click()
  await page.getByRole('button', { name: /BRAVIA 4K UR3/ }).click()
  await page.getByLabel('TV or soundbar IP address').fill('192.168.1.50')
  await page.getByRole('button', { name: 'Test and save' }).click()
  await page.getByRole('button', { name: 'Save and finish' }).click()

  await expect(page).toHaveURL('/')
  await expect(page.getByRole('tab', { name: 'Movies' })).toBeVisible()
  await expect(page.getByRole('button', { name: 'Aftersun' })).toBeVisible()
})
