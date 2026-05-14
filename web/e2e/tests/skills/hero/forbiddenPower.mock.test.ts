import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  forbiddenPowerConfirmPrompt,
  forbiddenPowerScenario,
} from '../../../scenarios/hero';

test.describe('hero forbidden power protocol harness', () => {
  test('activate forbidden power after attack hit', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(forbiddenPowerScenario({ crystals: 1 }));

    // Server pushes forbidden power confirm prompt (triggered after attack hit/miss)
    await protocolHarness.pushServerMessage(forbiddenPowerConfirmPrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    // Click "发动"
    await page.getByTestId('prompt-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
    // Backend handles the skill effect, frontend has no further interaction
  });

  test('activate forbidden power after attack miss', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(forbiddenPowerScenario({ crystals: 2 }));

    await protocolHarness.pushServerMessage(forbiddenPowerConfirmPrompt());
    await page.getByTestId('prompt-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('decline forbidden power activation', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(forbiddenPowerScenario({ crystals: 1 }));

    await protocolHarness.pushServerMessage(forbiddenPowerConfirmPrompt());
    await page.getByTestId('prompt-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});