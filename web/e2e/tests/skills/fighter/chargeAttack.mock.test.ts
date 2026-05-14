import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  chargeAttackConfirmPrompt,
  chargeAttackScenario,
} from '../../../scenarios/fighter';

test.describe('fighter charge attack protocol harness', () => {
  test('activate charge attack with qi not maxed', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(chargeAttackScenario({ qi: 1 }));

    // Server pushes charge attack confirm prompt (triggered on attack initiation)
    await protocolHarness.pushServerMessage(chargeAttackConfirmPrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    // Click "发动"
    await page.getByTestId('prompt-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('decline charge attack activation', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(chargeAttackScenario({ qi: 2 }));

    await protocolHarness.pushServerMessage(chargeAttackConfirmPrompt());
    await page.getByTestId('prompt-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});