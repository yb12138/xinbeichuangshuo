import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  angelSoulConfirmPrompt,
  angelSoulScenario,
} from '../../../scenarios/swordEmperor';

test.describe('sword emperor angel soul protocol harness', () => {
  test('angel soul: confirm before attack', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(angelSoulScenario());

    // Server pushes confirm prompt before attack
    await protocolHarness.pushServerMessage(angelSoulConfirmPrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('prompt-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('angel soul: skip confirm', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(angelSoulScenario());

    await protocolHarness.pushServerMessage(angelSoulConfirmPrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('prompt-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});