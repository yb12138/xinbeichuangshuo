import { expect, test } from '@playwright/test'
import { setupTestScenario } from '../../../helpers/realTestSetup'

const MB_CHARGE_SKILL_ID = 'mb_charge'

test('magic bow charge: hand <= 4 should show draw X prompt', async ({ page }) => {
  const scenario = await setupTestScenario({
    human_player: { name: 'E2E MagicBow', camp: 'Red', char_role: 'magic_bow' },
    bot_players: [{ name: 'Bot Enemy', camp: 'Blue', char_role: 'hero' }],
    setup: {
      first_turn_player: 'human',
      bots_paused: true,
      cheats: [
        // Set 1 crystal (required for 充能 skill cost)
        { target: 'human', command: 'set', args: ['crystal', '1'] },
        // Discard all starting cards first
        { target: 'human', command: 'discard', args: ['99'] },
        // Add 4 fire cards (hand size = 4, so no discard prompt should appear)
        { target: 'human', command: 'card_element', args: ['fire', '4'] },
      ],
    },
  })

  // Connect to the room
  await page.goto(`/?room=${scenario.room_code}&name=E2E MagicBow`)
  await expect(page.getByTestId('game-board')).toBeVisible({ timeout: 15_000 })

  // Activate 充能 skill (startup skill, should be in skill menu)
  await page.getByTestId('action-skill').click()
  await page.getByTestId(`skill-${MB_CHARGE_SKILL_ID}`).click()

  // Verify the draw X prompt is displayed (not blocked by any error)
  // This is the key fix: prompt should appear immediately after skill activation
  await expect(page.getByTestId('prompt-dialog')).toBeVisible({ timeout: 5_000 })

  // Check that the prompt contains the expected message
  const promptText = await page.getByTestId('prompt-dialog').textContent()
  expect(promptText).toContain('摸牌数量')

  // No connection errors should appear
  await expect(page.getByText('连接错误')).toHaveCount(0)
})