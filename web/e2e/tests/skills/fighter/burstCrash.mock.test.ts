import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  burstCrashConfirmPrompt,
  burstCrashScenario,
} from '../../../scenarios/fighter';

test.describe('fighter burst crash protocol harness', () => {
  test('activate burst crash with qi', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(burstCrashScenario({ qi: 2 }));

    await protocolHarness.pushServerMessage(burstCrashConfirmPrompt());
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

  test('decline burst crash activation', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(burstCrashScenario({ qi: 3 }));

    await protocolHarness.pushServerMessage(burstCrashConfirmPrompt());
    await expect(page.getByTestId('skill-branch-overlay')).toBeVisible();
    await page
      .getByTestId('skill-branch-overlay')
      .getByTestId('prompt-option-skip')
      .click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});
