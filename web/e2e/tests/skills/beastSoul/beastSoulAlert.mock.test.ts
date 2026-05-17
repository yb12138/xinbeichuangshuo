import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  beastSoulAlertResponsePrompt,
  beastSoulAlertScenario,
  beastSoulAlertDiscardPrompt,
} from '../../../scenarios/beastSoul';

test.describe('beast samurai beast soul alert protocol harness', () => {
  test('beast soul alert: confirm then tapped player discards', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(beastSoulAlertScenario({ beast_souls: 2 }));

    // 1) 响应技能 choose_skill 弹框
    await protocolHarness.pushServerMessage(beastSoulAlertResponsePrompt());
    await expect(page.getByTestId('skill-branch-overlay')).toBeVisible();
    await page.getByTestId('skill-branch-overlay').getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // 2) 触发横置的角色弃 1 张牌（无兽灵武士另选目标步骤）
    await protocolHarness.pushServerMessage(beastSoulAlertDiscardPrompt());
  });

  test('beast soul alert: skip via response skill prompt', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(beastSoulAlertScenario({ beast_souls: 2 }));

    await protocolHarness.pushServerMessage(beastSoulAlertResponsePrompt());
    await expect(page.getByTestId('skill-branch-overlay')).toBeVisible();
    await page.getByTestId('skill-branch-overlay').getByTestId('branch-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});
