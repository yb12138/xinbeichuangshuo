import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  victoryConfirmPrompt,
  victoryExtractStonePrompt,
  victoryModePrompt,
  victorySymphonyScenario,
} from '../../../scenarios/bard';

test.describe('bard victory symphony protocol harness', () => {
  test('branch 0: extract stone (gem)', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(victorySymphonyScenario());

    // Victory auto-triggers at turn end — server pushes confirm
    await protocolHarness.pushServerMessage(victoryConfirmPrompt());
    // 发动
    await page.getByTestId('prompt-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Mode: branch 0 = extract star stone
    await protocolHarness.pushServerMessage(victoryModePrompt());
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
    await page.getByTestId('prompt-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    await protocolHarness.pushServerMessage(victoryModePrompt());
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
    await page.getByTestId('prompt-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Branch 1 = camp gem + heal (resolved instantly by backend, no more prompts)
    await protocolHarness.pushServerMessage(victoryModePrompt());
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
    // 不发动 (option index 1)
    await page.getByTestId('prompt-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});
