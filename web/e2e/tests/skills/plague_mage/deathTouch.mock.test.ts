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

async function selectHandCards(page: import('@playwright/test').Page, indices: number[]) {
  for (const index of indices) {
    const card = page.getByTestId(`hand-card-${index}`);
    await card.scrollIntoViewIfNeeded();
    await card.click();
  }
  await page.getByTestId('prompt-confirm-btn').click();
}

test.describe('plague mage death touch protocol harness', () => {
  test('runs the full Death Touch prompt flow through the real UI and action adapter', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(plagueMageDeathTouchScenario());
  await page.getByTestId('action-magic').click();
  await page.getByTestId('action-skill').click();
    await page.getByTestId(`skill-${PLAGUE_DEATH_TOUCH_SKILL_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: PLAGUE_DEATH_TOUCH_SKILL_ID,
    });

    // Element selection (card_picker: plague_death_touch_element, needs prompt-confirm-btn)
    await protocolHarness.pushServerMessage(deathTouchElementPrompt());
    await expect(page.getByTestId('prompt-dialog')).toBeVisible();
    await page.getByTestId('hand-card-0').click();
    await page.getByTestId('prompt-confirm-btn').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // X value selection (numeric: decision-overlay with numeric-option buttons)
    await protocolHarness.pushServerMessage(deathTouchXPrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('numeric-option-2').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });

    // Card selection (card_picker from hand: click hand cards + prompt-confirm-btn)
    await protocolHarness.pushServerMessage(deathTouchCardsPrompt());
    await page.getByTestId('hand-card-0').click();
    await page.getByTestId('hand-card-1').click();
    await page.getByTestId('prompt-confirm-btn').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      card_ids: ['fire-attack-1', 'fire-attack-2'],
    });

    // Target selection (target_picker single target: click player-area, auto-submits via submitSelect)
    await protocolHarness.pushServerMessage(deathTouchTargetPrompt());
    await expect(page.getByTestId('decision-overlay')).not.toBeVisible({ timeout: 5000 });
    await page.getByTestId(`player-area-${ENEMY_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('cancels the activation-cost prompt and keeps the UI responsive', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(plagueMageDeathTouchScenario());
  await page.getByTestId('action-magic').click();
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