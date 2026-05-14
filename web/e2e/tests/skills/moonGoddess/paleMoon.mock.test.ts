import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  paleMoonConfirmPrompt,
  paleMoonScenario,
} from '../../../scenarios/moonGoddess';

test.describe('moon goddess pale moon protocol harness', () => {
  // Test the confirm prompt only - backend handles branch selection internally
  test('pale moon: confirm activation', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(paleMoonScenario({ petrify_tokens: 3, new_moon_tokens: 2 }));

    // Server pushes confirm prompt when trigger condition met
    await protocolHarness.pushServerMessage(paleMoonConfirmPrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('prompt-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('pale moon: skip confirm', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(paleMoonScenario({ petrify_tokens: 3, new_moon_tokens: 2 }));

    await protocolHarness.pushServerMessage(paleMoonConfirmPrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('prompt-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});