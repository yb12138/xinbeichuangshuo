import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  darkmoonSlashConfirmPrompt,
  darkmoonSlashScenario,
} from '../../../scenarios/moonGoddess';

test.describe('moon goddess darkmoon slash protocol harness', () => {
  // Test the confirm prompt only - backend handles X value selection internally
  test('darkmoon slash: confirm activation', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(darkmoonSlashScenario());

    // Server pushes confirm prompt when trigger condition met
    await protocolHarness.pushServerMessage(darkmoonSlashConfirmPrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('prompt-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('darkmoon slash: skip confirm', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(darkmoonSlashScenario());

    await protocolHarness.pushServerMessage(darkmoonSlashConfirmPrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('prompt-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});