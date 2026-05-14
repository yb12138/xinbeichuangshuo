import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  paleMoonBranchPrompt,
  paleMoonXPrompt,
  paleMoonScenario,
} from '../../../scenarios/moonGoddess';

// ============================================================
// 苍白之月 (mg_pale_moon) - 后端通过 response_skills 触发
// 分支选择使用 choice_type: mg_pale_moon_mode
// X值选择使用 choice_type: mg_pale_moon_x
// ============================================================

test.describe('moon goddess pale moon protocol harness', () => {
  test('pale moon: branch selection', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(paleMoonScenario({ petrify_tokens: 3, new_moon_tokens: 2 }));

    // 分支选择使用后端定义的 choice_type
    await protocolHarness.pushServerMessage(paleMoonBranchPrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('pale moon: branch 2 - select X value', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(paleMoonScenario({ petrify_tokens: 3, new_moon_tokens: 2 }));

    // 分支选择
    await protocolHarness.pushServerMessage(paleMoonBranchPrompt());
    await page.getByTestId('branch-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });

    // X值选择使用后端定义的 choice_type
    await protocolHarness.pushServerMessage(paleMoonXPrompt(2));
    await page.getByTestId('branch-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });

  test('pale moon: triggered via response_skills', async ({ protocolHarness }) => {
    await protocolHarness.bootGame(paleMoonScenario({ petrify_tokens: 3, new_moon_tokens: 2 }));

    // 后端会设置 response_skills 触发确认弹框
    await protocolHarness.expectSubmitAction({
      action_type: 'UseSkill',
      skill_id: 'mg_pale_moon',
    });
  });
});