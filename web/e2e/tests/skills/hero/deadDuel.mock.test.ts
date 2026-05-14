import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  deadDuelConfirmPrompt,
  deadDuelScenario,
} from '../../../scenarios/hero';

test.describe('hero dead duel protocol harness', () => {
  test('activate dead duel on magic damage', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(deadDuelScenario({ gems: 1 }));

    // Server pushes dead duel confirm prompt (triggered on magic damage received)
    await protocolHarness.pushServerMessage(deadDuelConfirmPrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    // Click "发动"
    await page.getByTestId('prompt-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('decline dead duel activation', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(deadDuelScenario({ gems: 2 }));

    await protocolHarness.pushServerMessage(deadDuelConfirmPrompt());
    await page.getByTestId('prompt-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});