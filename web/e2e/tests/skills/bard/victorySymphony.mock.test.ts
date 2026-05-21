import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  victoryConfirmPrompt,
  victoryExtractStonePrompt,
  victorySymphonyScenario,
} from '../../../scenarios/bard';

test.describe('bard victory symphony protocol harness', () => {
  test('branch 0: extract stone (gem)', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(victorySymphonyScenario());

    // Victory auto-triggers at turn end: branch choice and cancel are in one prompt.
    await protocolHarness.pushServerMessage(victoryConfirmPrompt());
    await page.getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Extract stone: choose gem
    await protocolHarness.pushServerMessage(victoryExtractStonePrompt());
    await page.getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('branch 0: extract stone (crystal)', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(victorySymphonyScenario());

    await protocolHarness.pushServerMessage(victoryConfirmPrompt());
    await page.getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Choose crystal
    await protocolHarness.pushServerMessage(victoryExtractStonePrompt());
    await page.getByTestId('branch-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });

  test('branch 1: camp gem + heal (instant, no further prompt)', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(victorySymphonyScenario());

    await protocolHarness.pushServerMessage(victoryConfirmPrompt());
    // Branch 1 = camp gem + heal (resolved instantly by backend, no more prompts).
    await page.getByTestId('branch-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });

  test('decline to use at confirm stage', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(victorySymphonyScenario());

    await protocolHarness.pushServerMessage(victoryConfirmPrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await expect(page.getByTestId('prompt-cancel-btn')).toHaveCount(0);
    await page.getByTestId('branch-option-2').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [2],
    });
  });
});
