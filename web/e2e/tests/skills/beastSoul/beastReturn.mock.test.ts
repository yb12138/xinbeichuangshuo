import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  beastReturnResponsePrompt,
  beastReturnXPrompt,
  beastReturnSelfDiscardPrompt,
  beastReturnSourceDiscardPrompt,
  beastReturnScenario,
} from '../../../scenarios/beastSoul';

test.describe('beast samurai beast return protocol harness', () => {
  test('beast return: confirm then remove X beast souls (X=2)', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(beastReturnScenario({ beast_souls: 3 }));

    await protocolHarness.pushServerMessage(beastReturnResponsePrompt());
    await expect(page.getByTestId('skill-branch-overlay')).toBeVisible();
    await page.getByTestId('skill-branch-overlay').getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // 后端 X 范围为 1..3；numeric_base=1，所以 X=2 对应 option index=1。
    await protocolHarness.pushServerMessage(beastReturnXPrompt(3));
    await page.getByTestId('numeric-option-2').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });

  test('beast return: skip via response skill prompt', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(beastReturnScenario({ beast_souls: 3 }));

    await protocolHarness.pushServerMessage(beastReturnResponsePrompt());
    await expect(page.getByTestId('skill-branch-overlay')).toBeVisible();
    await page.getByTestId('skill-branch-overlay').getByTestId('prompt-option-skip').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });

  test('beast return: pick X=max', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(beastReturnScenario({ beast_souls: 3 }));

    await protocolHarness.pushServerMessage(beastReturnResponsePrompt());
    await page.getByTestId('skill-branch-overlay').getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    await protocolHarness.pushServerMessage(beastReturnXPrompt(3));
    await page.getByTestId('numeric-option-3').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [2],
    });
  });

  // ============================================================
  // 完整交互流程测试：包含后续弃牌阶段
  // 参考后端 choices.go handleBeastReturnX / handleBeastReturnSelfDiscard / handleBeastReturnSourceDiscard
  // ============================================================

  test('beast return: full flow with X=2 - self discard 2 cards then source discard 1 card', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(beastReturnScenario({ beast_souls: 3 }));

    // Step 1: 响应技能确认弹框 - 选择发动兽返
    await protocolHarness.pushServerMessage(beastReturnResponsePrompt());
    await expect(page.getByTestId('skill-branch-overlay')).toBeVisible();
    await page.getByTestId('skill-branch-overlay').getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Step 2: 选择 X=2（移除2点兽魂，需弃2张牌）
    await protocolHarness.pushServerMessage(beastReturnXPrompt(3));
    await page.getByTestId('numeric-option-2').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });

    // Step 3: 自己弃 X=2 张牌 (bs_beast_return_self_discard)
    // 选牌类提示统一在手牌区完成，点击手牌卡片完成选择
    await protocolHarness.pushServerMessage(beastReturnSelfDiscardPrompt(2));
    // 选择第1张和第2张手牌（索引0和1）
    await page.getByTestId('hand-card-0').click();
    await page.getByTestId('hand-card-1').click();
    // 点击确认按钮提交选择
    await page.getByTestId('prompt-confirm-btn').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      card_ids: ['card_1', 'card_2'],
    });

    // Step 4: 伤害来源弃1张牌 (bs_beast_return_source_discard)
    // 弃牌 PromptChooseCards 投递给目标玩家（enemy，本地玩家不渲染）
    await protocolHarness.pushServerMessage(beastReturnSourceDiscardPrompt());
  });

  test('beast return: full flow with X=1 - self discard 1 card then source discard 1 card', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(beastReturnScenario({ beast_souls: 2 }));

    // Step 1: 响应技能确认弹框 - 选择发动兽返
    await protocolHarness.pushServerMessage(beastReturnResponsePrompt());
    await expect(page.getByTestId('skill-branch-overlay')).toBeVisible();
    await page.getByTestId('skill-branch-overlay').getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Step 2: 选择 X=1（移除1点兽魂，需弃1张牌）
    await protocolHarness.pushServerMessage(beastReturnXPrompt(2));
    await page.getByTestId('numeric-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Step 3: 自己弃 X=1 张牌 (bs_beast_return_self_discard)
    // 选牌类提示统一在手牌区完成，点击手牌卡片完成选择
    await protocolHarness.pushServerMessage(beastReturnSelfDiscardPrompt(1));
    await page.getByTestId('hand-card-0').click();
    await page.getByTestId('prompt-confirm-btn').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      card_ids: ['card_1'],
    });

    // Step 4: 伤害来源弃1张牌 (bs_beast_return_source_discard)
    // 弃牌 PromptChooseCards 投递给目标玩家（enemy，本地玩家不渲染）
    await protocolHarness.pushServerMessage(beastReturnSourceDiscardPrompt());
  });

  test('beast return: X=3 discard 3 cards from hand', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(beastReturnScenario({ beast_souls: 4 }));

    // Step 1: 响应技能确认弹框 - 选择发动兽返
    await protocolHarness.pushServerMessage(beastReturnResponsePrompt());
    await expect(page.getByTestId('skill-branch-overlay')).toBeVisible();
    await page.getByTestId('skill-branch-overlay').getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Step 2: 选择 X=3
    await protocolHarness.pushServerMessage(beastReturnXPrompt(4));
    await page.getByTestId('numeric-option-3').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [2],
    });

    // Step 3: 自己弃 X=3 张牌（玩家有4张手牌）
    // 选牌类提示统一在手牌区完成，点击手牌卡片完成选择
    await protocolHarness.pushServerMessage(beastReturnSelfDiscardPrompt(3));
    await page.getByTestId('hand-card-0').click();
    await page.getByTestId('hand-card-1').click();
    await page.getByTestId('hand-card-2').click();
    // 点击确认按钮提交选择
    await page.getByTestId('prompt-confirm-btn').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      card_ids: ['card_1', 'card_2', 'card_3'],
    });

    // Step 4: 伤害来源弃1张牌
    // 弃牌 PromptChooseCards 投递给目标玩家（enemy，本地玩家不渲染）
    await protocolHarness.pushServerMessage(beastReturnSourceDiscardPrompt());
  });
});
