import { test } from '../../../fixtures/protocolHarness.fixture';
import type { Page } from '@playwright/test';
import {
  BP_BLOOD_WAIL_SKILL_ID,
  bloodWailScenario,
  bloodWailXPrompt,
} from '../../../scenarios/bloodPriestess';

async function activateSkill(page: Page, skillId: string) {
  await page.getByTestId('action-skill').click();
  await page.getByTestId(`skill-${skillId}`).click();
}

// ============================================================
// 血之悲鸣 (bp_blood_wail) - 后端通过 available_skills 触发
// X值选择使用 choice_type: bp_blood_wail_x
// ============================================================

test.describe('blood priestess blood wail protocol harness', () => {
  test('blood wail: activate skill', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(bloodWailScenario());

    await activateSkill(page, BP_BLOOD_WAIL_SKILL_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: BP_BLOOD_WAIL_SKILL_ID,
    });
  });

  test('blood wail: X value selection', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(bloodWailScenario());

    // Backend pushes X value prompt after skill activation
    await protocolHarness.pushServerMessage(bloodWailXPrompt());
    await page.getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });
});