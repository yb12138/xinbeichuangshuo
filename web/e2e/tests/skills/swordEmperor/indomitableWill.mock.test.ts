import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  indomitableWillResponsePrompt,
  indomitableWillScenario,
} from '../../../scenarios/swordEmperor';

// 后端实际链路：choose_skill (response) → 确认发动后自动执行（摸1张、剑气+1、追加攻击行动）
// 无后续选择步骤。
test.describe('sword emperor indomitable will protocol harness', () => {
  test('indomitable will: response choose_skill and activate', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(indomitableWillScenario({ sword_qi: 2, crystal: 1 }));

    // 1) 触发响应技能 choose_skill 弹框
    await protocolHarness.pushServerMessage(indomitableWillResponsePrompt());
    await expect(page.getByTestId('skill-branch-overlay')).toBeVisible();
    await page.getByTestId('skill-branch-overlay').getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('indomitable will: skip via choose_skill', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(indomitableWillScenario({ sword_qi: 2, crystal: 1 }));

    await protocolHarness.pushServerMessage(indomitableWillResponsePrompt());
    await expect(page.getByTestId('skill-branch-overlay')).toBeVisible();
    // "跳过" is choose_skill skip → prompt-option-skip
    await page.getByTestId('skill-branch-overlay').getByTestId('prompt-option-skip').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });

  test('indomitable will: no prompt when no crystal available', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(indomitableWillScenario({ sword_qi: 2, crystal: 0, gem: 0 }));

    // Should not trigger if no crystal/gem available to pay
    await expect(page.getByTestId('decision-overlay')).not.toBeVisible();
    await expect(page.getByTestId('skill-branch-overlay')).not.toBeVisible();
  });
});
