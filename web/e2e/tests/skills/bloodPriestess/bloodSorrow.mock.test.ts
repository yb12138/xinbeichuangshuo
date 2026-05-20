import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  BP_BLOOD_SORROW_SKILL_ID,
  ENEMY_PLAYER_ID,
  bloodSorrowBranchPrompt,
  bloodSorrowScenario,
  bloodSorrowTargetPrompt,
} from '../../../scenarios/bloodPriestess';

test.describe('blood priestess blood sorrow protocol harness', () => {
  // 后端 buildBloodSorrowModePrompt: option_indexes[0]=移除，[1]=转移
  test('blood sorrow: choose transfer then target', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(bloodSorrowScenario());

    // Activate skill
  await page.getByTestId('action-magic').click();
  await page.getByTestId('action-skill').click();
    await page.getByTestId(`skill-${BP_BLOOD_SORROW_SKILL_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: BP_BLOOD_SORROW_SKILL_ID,
    });

    // Branch selection: 转移 = option_indexes[1]
    await protocolHarness.pushServerMessage(bloodSorrowBranchPrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('branch-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });

    // Target selection (click enemy player card)
    await protocolHarness.pushServerMessage(bloodSorrowTargetPrompt());
    await page.getByTestId(`player-area-${ENEMY_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  // 产品裁定（C1 选项 1）：「移除」分支由后端直接将同生共死从持有方移除，
  // 不再下发目标选择 prompt。本用例只验证 branch select 提交移除即流程结束。
  test('blood sorrow: choose remove (no target prompt)', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(bloodSorrowScenario());
  await page.getByTestId('action-magic').click();
  await page.getByTestId('action-skill').click();
    await page.getByTestId(`skill-${BP_BLOOD_SORROW_SKILL_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: BP_BLOOD_SORROW_SKILL_ID,
    });

    // Branch selection: 移除 = option_indexes[0]
    await protocolHarness.pushServerMessage(bloodSorrowBranchPrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // 当前 harness 里不会继续下发下一步 prompt，校验提交成功即可。
  });

  test('blood sorrow: skip', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(bloodSorrowScenario());
  await page.getByTestId('action-magic').click();
  await page.getByTestId('action-skill').click();
    await page.getByTestId(`skill-${BP_BLOOD_SORROW_SKILL_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: BP_BLOOD_SORROW_SKILL_ID,
    });

    await protocolHarness.pushServerMessage(bloodSorrowBranchPrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    // Skip is rendered as cancel button in overlay footer
    await page.getByTestId('prompt-cancel-btn').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Cancel',
    });
  });
});
