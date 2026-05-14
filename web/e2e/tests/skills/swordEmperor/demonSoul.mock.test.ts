import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  demonSoulConfirmPrompt,
  demonSoulScenario,
} from '../../../scenarios/swordEmperor';

test.describe('sword emperor demon soul protocol harness', () => {
  test('demon soul: confirm before attack', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(demonSoulScenario());

    // Server pushes confirm prompt before attack
    await protocolHarness.pushServerMessage(demonSoulConfirmPrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('prompt-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('demon soul: skip confirm', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(demonSoulScenario());

    await protocolHarness.pushServerMessage(demonSoulConfirmPrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('prompt-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});