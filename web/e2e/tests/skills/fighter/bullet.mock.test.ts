import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  bulletConfirmPrompt,
  bulletScenario,
} from '../../../scenarios/fighter';

test.describe('fighter bullet protocol harness', () => {
  test('activate bullet after magic action', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(bulletScenario());

    // Server pushes bullet confirm prompt (triggered after magic action)
    await protocolHarness.pushServerMessage(bulletConfirmPrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    // Click "发动"
    await page.getByTestId('prompt-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('decline bullet activation', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(bulletScenario());

    await protocolHarness.pushServerMessage(bulletConfirmPrompt());
    await page.getByTestId('prompt-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});