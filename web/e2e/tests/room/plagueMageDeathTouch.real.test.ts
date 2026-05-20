import { expect, test } from '@playwright/test'
import { setupTestScenario } from '../../helpers/realTestSetup'

const PLAGUE_DEATH_TOUCH_SKILL_ID = 'plague_death_touch'

test('plague mage death touch: full skill flow against real engine', async ({ page }) => {
  const scenario = await setupTestScenario({
    human_player: { name: 'E2E Plague', camp: 'Red', char_role: 'plague_mage' },
    bot_players: [{ name: 'Bot Hero', camp: 'Blue', char_role: 'hero' }],
    setup: {
      first_turn_player: 'human',
      bots_paused: true,
      cheats: [
        { target: 'human', command: 'set', args: ['heal', '4'] },
        { target: 'human', command: 'discard', args: ['99'] },
        { target: 'human', command: 'card_element', args: ['fire', '4'] },
      ],
    },
  })

  // Connect to the room (reconnect by name).
  await page.goto(`/?room=${scenario.room_code}&name=E2E Plague`)
  await expect(page.getByTestId('game-board')).toBeVisible({ timeout: 15_000 })

  // Activate skill selection.
  await page.getByTestId('action-skill').click()
  await page.getByTestId(`skill-${PLAGUE_DEATH_TOUCH_SKILL_ID}`).click()
  await page.locator(`[data-player-anchor="${scenario.bot_player_ids[0]}"]`).click()

  // Element selection prompt: pick the first fire card.
  await expect(page.getByTestId('prompt-dialog')).toBeVisible({ timeout: 5_000 })
  await page.getByTestId('hand-card-0').click()
  await page.getByTestId('prompt-confirm-btn').click()

  // X value prompt: choose X=2.
  await expect(page.getByTestId('numeric-option-2')).toBeVisible({ timeout: 5_000 })
  await page.getByTestId('numeric-option-2').click()

  // Card selection prompt: pick 2 fire cards.
  await expect(page.getByTestId('prompt-dialog')).toBeVisible({ timeout: 5_000 })
  await page.getByTestId('hand-card-0').click()
  await page.getByTestId('hand-card-1').click()
  await page.getByTestId('prompt-confirm-btn').click()

  // Target selection prompt: click the enemy player.
  await page.locator(`[data-player-anchor="${scenario.bot_player_ids[0]}"]`).click()

  // Verify no errors and board still visible.
  await expect(page.getByText('连接错误')).toHaveCount(0)
  await expect(page.getByTestId('game-board')).toBeVisible()
})
