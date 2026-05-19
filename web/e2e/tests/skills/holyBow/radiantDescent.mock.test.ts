import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  HB_RADIANT_DESCENT_SKILL_ID,
  radiantDescentScenario,
  radiantDescentCostPrompt,
} from '../../../scenarios/holyBow';

test.describe('holy bow radiant descent protocol harness', () => {
  // 场景A: 只有治疗>=2（信仰<2） → prompt只有一个选项"移除2点治疗"，直接确认发动
  test('scenario A: heal>=2, faith<2 - only heal option available', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(radiantDescentScenario({ heal: 2, faith: 1 }));

    // Activate skill directly
    await page.getByTestId('action-skill').click();
    await page.getByTestId(`skill-${HB_RADIANT_DESCENT_SKILL_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: HB_RADIANT_DESCENT_SKILL_ID,
    });

    // Backend sends hb_radiant_descent_cost prompt with only "heal" option
    await protocolHarness.pushServerMessage(radiantDescentCostPrompt(['heal']));
    await expect(page.getByTestId('decision-overlay')).toBeVisible();

    // 只有一个选项，选择"移除2点治疗"（branch-option-0）
    await page.getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  // 场景B: 只有信仰>=2（治疗<2） → prompt只有一个选项"移除2点信仰"，直接确认发动
  test('scenario B: faith>=2, heal<2 - only faith option available', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(radiantDescentScenario({ heal: 1, faith: 2 }));

    // Activate skill directly
    await page.getByTestId('action-skill').click();
    await page.getByTestId(`skill-${HB_RADIANT_DESCENT_SKILL_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: HB_RADIANT_DESCENT_SKILL_ID,
    });

    // Backend sends hb_radiant_descent_cost prompt with only "faith" option
    await protocolHarness.pushServerMessage(radiantDescentCostPrompt(['faith']));
    await expect(page.getByTestId('decision-overlay')).toBeVisible();

    // 只有一个选项，选择"移除2点信仰"（branch-option-0）
    await page.getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  // 场景C: 治疗>=2且信仰>=2 → prompt有两个选项，需要选择支付方式
  test('scenario C: heal>=2 and faith>=2 - both options available, user selects', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(radiantDescentScenario({ heal: 2, faith: 2 }));

    // Activate skill directly
    await page.getByTestId('action-skill').click();
    await page.getByTestId(`skill-${HB_RADIANT_DESCENT_SKILL_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: HB_RADIANT_DESCENT_SKILL_ID,
    });

    // Backend sends hb_radiant_descent_cost prompt with both options
    await protocolHarness.pushServerMessage(radiantDescentCostPrompt(['heal', 'faith']));
    await expect(page.getByTestId('decision-overlay')).toBeVisible();

    // 有两个选项，用户选择"移除2点信仰"
    await page.getByTestId('branch-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});
