import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  SC_INCANTATION_SKILL_ID,
  incantationScenario,
  incantationConfirmPrompt,
  incantationCoverPrompt,
} from '../../../scenarios/spiritCaster';

test.describe('spirit caster incantation protocol harness', () => {
  test('incantation: confirm and cover a card as youli', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(incantationScenario());

    // Incantation response prompt appears after talisman skill
    await protocolHarness.pushServerMessage(incantationConfirmPrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('prompt-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Cover card as youli
    await protocolHarness.pushServerMessage(incantationCoverPrompt());
    await page.getByTestId('hand-card-0').click();
    await page.getByTestId('prompt-confirm-btn').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('incantation: confirm and cover different card', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(incantationScenario());

    await protocolHarness.pushServerMessage(incantationConfirmPrompt());
    await page.getByTestId('prompt-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    await protocolHarness.pushServerMessage(incantationCoverPrompt());
    await page.getByTestId('hand-card-2').click();
    await page.getByTestId('prompt-confirm-btn').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [2],
    });
  });

  test('incantation: skip activation', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(incantationScenario());

    await protocolHarness.pushServerMessage(incantationConfirmPrompt());
    await page.getByTestId('prompt-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});