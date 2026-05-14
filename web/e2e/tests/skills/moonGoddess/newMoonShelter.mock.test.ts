import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  newMoonShelterConfirmPrompt,
  newMoonShelterScenario,
} from '../../../scenarios/moonGoddess';

test.describe('moon goddess new moon shelter protocol harness', () => {
  test('new moon shelter: confirm activation', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(newMoonShelterScenario());

    // Server pushes confirm prompt when trigger condition met
    await protocolHarness.pushServerMessage(newMoonShelterConfirmPrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('prompt-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('new moon shelter: skip activation', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(newMoonShelterScenario());

    await protocolHarness.pushServerMessage(newMoonShelterConfirmPrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('prompt-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});