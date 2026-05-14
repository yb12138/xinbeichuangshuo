import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  hundredDragonConfirmPrompt,
  hundredDragonScenario,
} from '../../../scenarios/fighter';

test.describe('fighter hundred dragon protocol harness', () => {
  test('activate hundred dragon with 3+ qi at turn start', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(hundredDragonScenario({ qi: 3 }));

    // Server pushes hundred dragon confirm prompt (triggered at turn start phase)
    await protocolHarness.pushServerMessage(hundredDragonConfirmPrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    // Click "发动"
    await page.getByTestId('prompt-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('activate hundred dragon with more qi', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(hundredDragonScenario({ qi: 5 }));

    await protocolHarness.pushServerMessage(hundredDragonConfirmPrompt());
    await page.getByTestId('prompt-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('decline hundred dragon activation', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(hundredDragonScenario({ qi: 4 }));

    await protocolHarness.pushServerMessage(hundredDragonConfirmPrompt());
    await page.getByTestId('prompt-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});