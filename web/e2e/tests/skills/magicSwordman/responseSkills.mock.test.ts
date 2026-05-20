import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  MAGIC_SWORDMAN_SHADOW_GATHER_ID,
  MAGIC_SWORDMAN_SHADOW_METEOR_ID,
  ENEMY_PLAYER_ID,
  shadowGatherScenario,
  shadowMeteorScenario,
  shadowMeteorDiscardPrompt,
  shadowMeteorTargetPrompt,
} from '../../../scenarios/magicSwordman';

async function activatePanelSkill(page: Page, skillId: string) {
  await page.getByTestId('action-hub-trigger').click();
  await page.getByTestId('action-magic').click();
  await page.getByTestId('action-skill').click();
  await page.getByTestId(`skill-${skillId}`).click();
}

import type { Page } from '@playwright/test';

async function selectTarget(page: Page, targetId: string) {
  await page.getByTestId(`player-area-${targetId}`).click();
}

test.describe('magic swordman shadow gather protocol harness', () => {
  test('shadow gather: activate skill', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(shadowGatherScenario());

    await activatePanelSkill(page, MAGIC_SWORDMAN_SHADOW_GATHER_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: MAGIC_SWORDMAN_SHADOW_GATHER_ID,
    });
  });
});

test.describe('magic swordman shadow meteor protocol harness', () => {
  test('shadow meteor: activate in shadow form', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(shadowMeteorScenario());

    await activatePanelSkill(page, MAGIC_SWORDMAN_SHADOW_METEOR_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: MAGIC_SWORDMAN_SHADOW_METEOR_ID,
    });
  });

  test('shadow meteor: discard and select target', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(shadowMeteorScenario());

    await activatePanelSkill(page, MAGIC_SWORDMAN_SHADOW_METEOR_ID);

    // Server pushes discard selection
    await protocolHarness.pushServerMessage(shadowMeteorDiscardPrompt());

    // Select magic card from hand (card_picker: hand-card-2 auto-submits for min=1,max=1)
    await page.getByTestId('hand-card-2').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      card_ids: ['ms-fire-magic'],
    });

    // Server pushes target selection
    await protocolHarness.pushServerMessage(shadowMeteorTargetPrompt());

    // Select enemy target (target_picker: click player area, no overlay)
    await expect(page.getByTestId('decision-overlay')).not.toBeVisible({ timeout: 5000 });
    await selectTarget(page, ENEMY_PLAYER_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });
});