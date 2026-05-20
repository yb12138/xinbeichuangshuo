import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  reversalIaijutsuResponsePrompt,
  reversalIaijutsuXPrompt,
  reversalIaijutsuScenario,
  reversalIaijutsuTargetDiscardPrompt,
} from '../../../scenarios/beastSoul';

test.describe('beast samurai reversal iaijutsu protocol harness', () => {
  test('reversal iaijutsu: confirm then remove X then target discards X+2 cards', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(reversalIaijutsuScenario({ beast_souls: 3 }));

    // 响应技能 choose_skill 入口
    await protocolHarness.pushServerMessage(reversalIaijutsuResponsePrompt());
    await expect(page.getByTestId('skill-branch-overlay')).toBeVisible();
    await page.getByTestId('skill-branch-overlay').getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // 选 X=2（显示值为 2，option index = 1）→ 目标将弃置 X+2=4 张
    await protocolHarness.pushServerMessage(reversalIaijutsuXPrompt(3));
    await page.getByTestId('numeric-option-2').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });

    // 后端直接以攻击目标为弃牌对象，无需额外目标选择步。
    // 弃牌 PromptChooseCards 投递给目标玩家（本地玩家不渲染）。
    await protocolHarness.pushServerMessage(reversalIaijutsuTargetDiscardPrompt(4));
  });

  test('reversal iaijutsu: skip via response skill prompt', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(reversalIaijutsuScenario({ beast_souls: 3 }));

    await protocolHarness.pushServerMessage(reversalIaijutsuResponsePrompt());
    await expect(page.getByTestId('skill-branch-overlay')).toBeVisible();
    await page.getByTestId('skill-branch-overlay').getByTestId('prompt-option-skip').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });

  test('reversal iaijutsu: pick X=1 (target discards 3 cards)', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(reversalIaijutsuScenario({ beast_souls: 3 }));

    await protocolHarness.pushServerMessage(reversalIaijutsuResponsePrompt());
    await page.getByTestId('skill-branch-overlay').getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    await protocolHarness.pushServerMessage(reversalIaijutsuXPrompt(3));
    await page.getByTestId('numeric-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    await protocolHarness.pushServerMessage(reversalIaijutsuTargetDiscardPrompt(3));
  });

  test('reversal iaijutsu: pick X=max (target discards X+2=5 cards)', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(reversalIaijutsuScenario({ beast_souls: 3 }));

    await protocolHarness.pushServerMessage(reversalIaijutsuResponsePrompt());
    await page.getByTestId('skill-branch-overlay').getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    await protocolHarness.pushServerMessage(reversalIaijutsuXPrompt(3));
    await page.getByTestId('numeric-option-3').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [2],
    });

    await protocolHarness.pushServerMessage(reversalIaijutsuTargetDiscardPrompt(5));
  });
});
