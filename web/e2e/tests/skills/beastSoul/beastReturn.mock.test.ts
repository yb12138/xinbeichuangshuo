import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  beastReturnResponsePrompt,
  beastReturnXPrompt,
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

    // 后端 X 范围为 0..3（含「X=0 不移除兽魂」），option_indexes 与 id 对齐
    await protocolHarness.pushServerMessage(beastReturnXPrompt(3));
    await page.getByTestId('prompt-option-2').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [2],
    });
  });

  test('beast return: skip via response skill prompt', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(beastReturnScenario({ beast_souls: 3 }));

    await protocolHarness.pushServerMessage(beastReturnResponsePrompt());
    await expect(page.getByTestId('skill-branch-overlay')).toBeVisible();
    await page.getByTestId('skill-branch-overlay').getByTestId('branch-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });

  test('beast return: pick X=0 (不移除兽魂)', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(beastReturnScenario({ beast_souls: 3 }));

    await protocolHarness.pushServerMessage(beastReturnResponsePrompt());
    await page.getByTestId('skill-branch-overlay').getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    await protocolHarness.pushServerMessage(beastReturnXPrompt(3));
    await page.getByTestId('prompt-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
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
    await page.getByTestId('prompt-option-3').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [3],
    });
  });
});
