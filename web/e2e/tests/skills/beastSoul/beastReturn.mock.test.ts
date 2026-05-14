import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  beastReturnConfirmPrompt,
  beastReturnRemovePrompt,
  beastReturnScenario,
} from '../../../scenarios/beastSoul';

test.describe('beast soul warrior beast return protocol harness', () => {
  test('beast return: confirm then remove beast souls', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(beastReturnScenario({ beast_souls: 3 }));

    // Server pushes confirm prompt as response skill
    await protocolHarness.pushServerMessage(beastReturnConfirmPrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('prompt-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Remove X beast souls (choose X=2)
    await protocolHarness.pushServerMessage(beastReturnRemovePrompt(3));
    await page.getByTestId('branch-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });

  test('beast return: skip confirm', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(beastReturnScenario({ beast_souls: 3 }));

    await protocolHarness.pushServerMessage(beastReturnConfirmPrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('prompt-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });

  test('beast return: remove 1 beast soul', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(beastReturnScenario({ beast_souls: 3 }));

    await protocolHarness.pushServerMessage(beastReturnConfirmPrompt());
    await page.getByTestId('prompt-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Choose X=1
    await protocolHarness.pushServerMessage(beastReturnRemovePrompt(3));
    await page.getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('beast return: remove 3 beast souls', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(beastReturnScenario({ beast_souls: 3 }));

    await protocolHarness.pushServerMessage(beastReturnConfirmPrompt());
    await page.getByTestId('prompt-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Choose X=3 (max)
    await protocolHarness.pushServerMessage(beastReturnRemovePrompt(3));
    await page.getByTestId('branch-option-2').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [2],
    });
  });
});