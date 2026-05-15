import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  heavenDriveConfirmPrompt,
  heavenDriveScenario,
} from '../../../scenarios/fighter';

test.describe('fighter heaven drive protocol harness', () => {
  test('activate heaven drive with crystal at turn start', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(heavenDriveScenario({ crystals: 1 }));

    await protocolHarness.pushServerMessage(heavenDriveConfirmPrompt());
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

  test('activate heaven drive with more crystals', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(heavenDriveScenario({ crystals: 2 }));

    await protocolHarness.pushServerMessage(heavenDriveConfirmPrompt());
    await page
      .getByTestId('skill-branch-overlay')
      .getByTestId('branch-option-0')
      .click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('decline heaven drive activation', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(heavenDriveScenario({ crystals: 1 }));

    await protocolHarness.pushServerMessage(heavenDriveConfirmPrompt());
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
