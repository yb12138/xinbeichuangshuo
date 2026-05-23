import { expect, test } from '@playwright/test';

test.describe('real room smoke', () => {
  test('creates a room, fills a 1v1 lineup, and enters the game board', async ({ page }) => {
    await page.goto('/');

    await page.getByPlaceholder('输入你的名字').fill('E2E Host');
    await page.getByTestId('create-room-button').click();

    await expect(page.getByTestId('room-code')).toHaveText(/[A-Z0-9]{4}/, { timeout: 10_000 });

    await page.getByTestId('join-camp-red').click();
    await page.getByTestId('role-card-plague_mage').click();

    await page.getByTestId('add-bot-button').click();
    await page.getByTestId('bot-camp-blue-p2').click();
    await page.getByTestId('bot-role-p2').selectOption('hero');

    const gameBoard = page.getByTestId('game-board');
    await page.getByTestId('start-game-button').click({ timeout: 1_000 }).catch(() => {});
    await expect(gameBoard).toBeVisible({ timeout: 20_000 });

    await expect(page.getByText('连接错误')).toHaveCount(0);
    await expect(page.getByText('游戏尚未开始')).toHaveCount(0);
  });
});
