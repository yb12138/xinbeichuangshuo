import { test } from '../../../fixtures/protocolHarness.fixture';
import type { Page } from '@playwright/test';
import {
  BP_BLOOD_CURSE_SKILL_ID,
  bloodCurseScenario,
  bloodCurseDiscardPrompt,
} from '../../../scenarios/bloodPriestess';

async function activateSkill(page: Page, skillId: string) {
  await page.getByTestId('action-hub-trigger').click();
  await page.getByTestId('action-magic').click();
  await page.getByTestId('action-skill').click();
  await page.getByTestId(`skill-${skillId}`).click();
}

// ============================================================
// 血之诅咒 (bp_blood_curse) - 后端通过 available_skills 触发
// 弃牌使用 choice_type: bp_curse_discard
// ============================================================

test.describe('blood priestess blood curse protocol harness', () => {
  test('blood curse: activate skill', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(bloodCurseScenario());

    await activateSkill(page, BP_BLOOD_CURSE_SKILL_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: BP_BLOOD_CURSE_SKILL_ID,
    });
  });

  test('blood curse: discard 3 cards', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(bloodCurseScenario());

    await protocolHarness.pushServerMessage(bloodCurseDiscardPrompt());
    await page.getByTestId('hand-card-0').click();
    await page.getByTestId('hand-card-1').click();
    await page.getByTestId('hand-card-2').click();
    await page.getByTestId('prompt-confirm-btn').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      card_ids: ['card_1', 'card_2', 'card_3'],
    });
  });
});
