import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  starBulletConfirmPrompt,
  starBulletScenario,
} from '../../../scenarios/holyBow';

test.describe('holy bow star bullet protocol harness', () => {
  test('activate star bullet on attack', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(starBulletScenario());

    // Server pushes star bullet confirm prompt (triggered on attack initiation)
    await protocolHarness.pushServerMessage(starBulletConfirmPrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    // Click "发动"
    await page.getByTestId('prompt-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('decline star bullet activation', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(starBulletScenario());

    await protocolHarness.pushServerMessage(starBulletConfirmPrompt());
    await page.getByTestId('prompt-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});