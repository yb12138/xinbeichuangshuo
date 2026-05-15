import { test } from '../../../fixtures/protocolHarness.fixture';
import type { Page } from '@playwright/test';
import {
  BP_SHARED_LIFE_SKILL_ID,
  ENEMY_PLAYER_ID,
  ENEMY_2_PLAYER_ID,
  sharedLifeScenario,
  sharedLifeTargetPrompt,
} from '../../../scenarios/bloodPriestess';

async function activateSkill(page: Page, skillId: string) {
  await page.getByTestId('action-skill').click();
  await page.getByTestId(`skill-${skillId}`).click();
}

async function selectTarget(page: Page, targetId: string) {
  await page.getByTestId(`player-area-${targetId}`).click();
}

// ============================================================
// 同生共死 (bp_shared_life) - 后端通过 available_skills 触发
// 目标选择使用 choice_type: bp_shared_life_target
// ============================================================

test.describe('blood priestess shared life protocol harness', () => {
  test('shared life: activate skill', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(sharedLifeScenario());

    await activateSkill(page, BP_SHARED_LIFE_SKILL_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: BP_SHARED_LIFE_SKILL_ID,
    });
  });

  test('shared life: target selection', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(sharedLifeScenario());

    await protocolHarness.pushServerMessage(sharedLifeTargetPrompt());
    await selectTarget(page, ENEMY_PLAYER_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('shared life: select second target', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(sharedLifeScenario());

    await protocolHarness.pushServerMessage(sharedLifeTargetPrompt());
    await selectTarget(page, ENEMY_2_PLAYER_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});