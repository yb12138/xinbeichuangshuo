import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  roarConfirmPrompt,
  roarDrawPrompt,
  roarScenario,
} from '../../../scenarios/hero';

test.describe('hero roar protocol harness', () => {
  test('activate roar and draw card', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(roarScenario({ anger: 1 }));

    // Server pushes roar confirm prompt (triggered on attack initiation)
    await protocolHarness.pushServerMessage(roarConfirmPrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    // Click "发动"
    await page.getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Draw prompt: ask if want to draw 1 card
    await protocolHarness.pushServerMessage(roarDrawPrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    // Click "摸1张牌" (text mode uses branch-option)
    await page.getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('activate roar but skip draw', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(roarScenario({ anger: 2 }));

    await protocolHarness.pushServerMessage(roarConfirmPrompt());
    await page.getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Draw prompt: skip draw (text mode uses branch-option)
    await protocolHarness.pushServerMessage(roarDrawPrompt());
    await page.getByTestId('branch-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });

  test('decline roar activation', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(roarScenario({ anger: 1 }));

    await protocolHarness.pushServerMessage(roarConfirmPrompt());
    await page.getByTestId('branch-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
    // No draw prompt follows when declined
  });
});
