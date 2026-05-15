import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  hundredDragonConfirmPrompt,
  hundredDragonScenario,
} from '../../../scenarios/fighter';

test.describe('fighter hundred dragon protocol harness', () => {
  test('activate hundred dragon with 3+ qi at turn start', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(hundredDragonScenario({ qi: 3 }));

    await protocolHarness.pushServerMessage(hundredDragonConfirmPrompt());
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

  test('activate hundred dragon with more qi', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(hundredDragonScenario({ qi: 5 }));

    await protocolHarness.pushServerMessage(hundredDragonConfirmPrompt());
    await page
      .getByTestId('skill-branch-overlay')
      .getByTestId('branch-option-0')
      .click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('decline hundred dragon activation', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(hundredDragonScenario({ qi: 4 }));

    await protocolHarness.pushServerMessage(hundredDragonConfirmPrompt());
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
