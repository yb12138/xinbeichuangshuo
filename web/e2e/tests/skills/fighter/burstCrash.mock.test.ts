import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  burstCrashConfirmPrompt,
  burstCrashScenario,
} from '../../../scenarios/fighter';

test.describe('fighter burst crash protocol harness', () => {
  test('activate burst crash with qi', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(burstCrashScenario({ qi: 2 }));

    // Server pushes burst crash confirm prompt (triggered on attack initiation)
    await protocolHarness.pushServerMessage(burstCrashConfirmPrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    // Click "发动"
    await page.getByTestId('prompt-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('decline burst crash activation', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(burstCrashScenario({ qi: 3 }));

    await protocolHarness.pushServerMessage(burstCrashConfirmPrompt());
    await page.getByTestId('prompt-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});