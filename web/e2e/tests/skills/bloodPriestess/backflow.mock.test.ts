import { test } from '../../../fixtures/protocolHarness.fixture';
import type { Page } from '@playwright/test';
import {
  BP_BACKFLOW_SKILL_ID,
  backflowScenario,
} from '../../../scenarios/bloodPriestess';

async function activateSkill(page: Page, skillId: string) {
  await page.getByTestId('action-hub-trigger').click();
  await page.getByTestId('action-magic').click();
  await page.getByTestId('action-skill').click();
  await page.getByTestId(`skill-${skillId}`).click();
}

// ============================================================
// 逆流 (bp_backflow) - 后端通过 available_skills 触发
// 弃牌通过 cost_discards 自动处理
// ============================================================

test.describe('blood priestess backflow protocol harness', () => {
  test('backflow: activate skill with cost_discards', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(backflowScenario());

    // Click skill button - frontend shows discard picker (cost_discards=2)
    await activateSkill(page, BP_BACKFLOW_SKILL_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: BP_BACKFLOW_SKILL_ID,
    });
  });
});