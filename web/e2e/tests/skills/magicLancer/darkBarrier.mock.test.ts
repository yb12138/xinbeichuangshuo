import type { Page } from '@playwright/test';
import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  darkBarrierCardsPrompt,
  darkBarrierScenario,
} from '../../../scenarios/magicLancer';

async function selectHandCards(page: Page, indices: number[]) {
  for (const index of indices) {
    const card = page.getByTestId(`hand-card-${index}`);
    await card.scrollIntoViewIfNeeded();
    await card.click();
  }
  await page.getByTestId('prompt-confirm-btn').click();
}

test.describe('magic lancer dark barrier protocol harness', () => {
  test('select magic cards only (cascade constraint)', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(darkBarrierScenario());

    // Server pushes dark barrier prompt (response skill triggered by damage)
    await protocolHarness.pushServerMessage(darkBarrierCardsPrompt());

    // Wait for prompt-dialog to be visible (card_picker from hand)
    await expect(page.getByTestId('prompt-dialog')).toBeVisible();

    // Select two magic cards (indices 0 and 1)
    await selectHandCards(page, [0, 1]);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      card_ids: ['ml-magic-1', 'ml-magic-2'],
    });
  });

  test('select single thunder card', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(darkBarrierScenario());

    await protocolHarness.pushServerMessage(darkBarrierCardsPrompt());

    await expect(page.getByTestId('prompt-dialog')).toBeVisible();

    // Select only the thunder card (index 2)
    await selectHandCards(page, [2]);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      card_ids: ['ml-thunder-1'],
    });
  });

  test('prevents mixing magic and thunder cards in one dark barrier discard', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(darkBarrierScenario());

    await protocolHarness.pushServerMessage(darkBarrierCardsPrompt());

    await expect(page.getByTestId('prompt-dialog')).toBeVisible();

    // Dark Barrier allows either all magic cards or all thunder cards. After picking
    // magic cards, the thunder card must remain unavailable for this submission.
    await page.getByTestId('hand-card-0').click();
    await page.getByTestId('hand-card-1').click();
    await expect(page.getByTestId('hand-card-2')).not.toHaveClass(/selectable/);
    await page.getByTestId('hand-card-2').click();
    await expect(page.getByTestId('hand-card-2')).not.toHaveClass(/selected/);
    await page.getByTestId('prompt-confirm-btn').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      card_ids: ['ml-magic-1', 'ml-magic-2'],
    });
  });
});
