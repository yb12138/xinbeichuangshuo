import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  beastSoulAlertResponsePrompt,
  beastSoulAlertScenario,
  beastSoulAlertTargetPrompt,
  beastSoulAlertDiscardPrompt,
} from '../../../scenarios/beastSoul';

test.describe('beast samurai beast soul alert protocol harness', () => {
  test('beast soul alert: confirm then target then target discards', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(beastSoulAlertScenario({ beast_souls: 2 }));

    // 1) 响应技能 choose_skill 弹框
    await protocolHarness.pushServerMessage(beastSoulAlertResponsePrompt());
    await expect(page.getByTestId('skill-branch-overlay')).toBeVisible();
    await page.getByTestId('skill-branch-overlay').getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // 2) bs_alert_target：选择 1 名让其弃 1 张牌的角色（option id 为序号）
    await protocolHarness.pushServerMessage(beastSoulAlertTargetPrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('prompt-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // 3) 目标弃 1 张牌（PromptChooseCards 走目标玩家手牌；本地玩家不可见，
    // 仅校验该 prompt 能被协议下发，无需本地 UI 交互）
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
