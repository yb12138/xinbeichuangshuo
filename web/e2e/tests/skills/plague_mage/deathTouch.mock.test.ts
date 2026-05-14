import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  ENEMY_PLAYER_ID,
  PLAGUE_DEATH_TOUCH_SKILL_ID,
  deathTouchCardsPrompt,
  deathTouchElementPrompt,
  deathTouchTargetPrompt,
  deathTouchXPrompt,
  plagueMageDeathTouchScenario,
} from '../../../scenarios/plagueMage';

test.describe('plague mage death touch protocol harness', () => {
  test('runs the full Death Touch prompt flow through the real UI and action adapter', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(plagueMageDeathTouchScenario());

    await page.getByTestId('action-skill').click();
    await page.getByTestId(`skill-${PLAGUE_DEATH_TOUCH_SKILL_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: PLAGUE_DEATH_TOUCH_SKILL_ID,
    });

    await protocolHarness.pushServerMessage(deathTouchElementPrompt());
    await page.getByTestId('hand-card-0').click();
    await page.getByTestId('prompt-confirm-btn').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    await protocolHarness.pushServerMessage(deathTouchXPrompt());
    await page.getByTestId('numeric-option-2').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });

    await protocolHarness.pushServerMessage(deathTouchCardsPrompt());
    await page.getByTestId('hand-card-0').click();
    await page.getByTestId('hand-card-1').click();
    await page.getByTestId('prompt-confirm-btn').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0, 1],
    });

    await protocolHarness.pushServerMessage(deathTouchTargetPrompt());
    await page.locator(`[data-player-anchor="${ENEMY_PLAYER_ID}"]`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      targets: [{ target_user_id: ENEMY_PLAYER_ID }],
    });
  });

  test('cancels the activation-cost prompt and keeps the UI responsive', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(plagueMageDeathTouchScenario());

    await page.getByTestId('action-skill').click();
    await page.getByTestId(`skill-${PLAGUE_DEATH_TOUCH_SKILL_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: PLAGUE_DEATH_TOUCH_SKILL_ID,
    });

    await protocolHarness.pushServerMessage(deathTouchElementPrompt());
    await expect(page.getByTestId('prompt-dialog')).toBeVisible();
    await page.getByTestId('prompt-cancel-btn').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Cancel',
    });

    await expect(page.getByText('未连接到服务器')).toHaveCount(0);
    await expect(page.getByText('连接错误')).toHaveCount(0);
  });

  test('hides the skill entry when the server publishes no available skills', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(plagueMageDeathTouchScenario({ availableSkills: [] }));

    await expect(page.getByTestId('action-skill')).toHaveCount(0);
    await expect(page.getByText('发动技能')).toHaveCount(0);
  });
});
