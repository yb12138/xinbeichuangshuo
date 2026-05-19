import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  SC_INCANTATION_SKILL_ID,
  incantationScenario,
  incantationConfirmPrompt,
  incantationCoverPrompt,
} from '../../../scenarios/spiritCaster';

async function selectHandCards(page: import('@playwright/test').Page, indices: number[]) {
  for (const index of indices) {
    const card = page.getByTestId(`hand-card-${index}`);
    await card.scrollIntoViewIfNeeded();
    await card.click();
  }
  await page.getByTestId('prompt-confirm-btn').click();
}

test.describe('spirit caster incantation protocol harness', () => {
  test('incantation: confirm and cover a card as youli', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(incantationScenario());

    // Incantation response prompt (branch_select overlay)
    await protocolHarness.pushServerMessage(incantationConfirmPrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('prompt-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Cover card as youli (card_picker from hand)
    await protocolHarness.pushServerMessage(incantationCoverPrompt());
    await selectHandCards(page, [0]);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      card_ids: ['card_1'],
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
    await selectHandCards(page, [2]);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      card_ids: ['card_3'],
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