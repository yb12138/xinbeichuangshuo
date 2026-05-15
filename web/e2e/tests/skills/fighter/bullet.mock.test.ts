import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  bulletConfirmPrompt,
  bulletScenario,
} from '../../../scenarios/fighter';

test.describe('fighter bullet protocol harness', () => {
  test('activate bullet after magic action', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(bulletScenario());

    await protocolHarness.pushServerMessage(bulletConfirmPrompt());
    await expect(page.getByTestId('skill-branch-overlay')).toBeVisible();
    await page
      .getByTestId('skill-branch-overlay')
      .getByTestId('branch-option-0')
      .click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('decline bullet activation', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(bulletScenario());

    await protocolHarness.pushServerMessage(bulletConfirmPrompt());
    await expect(page.getByTestId('skill-branch-overlay')).toBeVisible();
    await page
      .getByTestId('skill-branch-overlay')
      .getByTestId('branch-option-1')
      .click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});
