import type { Page } from '@playwright/test';
import { test } from '../../../fixtures/protocolHarness.fixture';
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

    // Select two magic cards (indices 0 and 1)
    await selectHandCards(page, [0, 1]);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0, 1],
    });
  });

  test('select single thunder card', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(darkBarrierScenario());

    await protocolHarness.pushServerMessage(darkBarrierCardsPrompt());

    // Select only the thunder card (index 2)
    await selectHandCards(page, [2]);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [2],
    });
  });

  test('select all valid cards', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(darkBarrierScenario());

    await protocolHarness.pushServerMessage(darkBarrierCardsPrompt());

    // Select all valid cards (all magic OR all thunder; cascade constraint says magic first)
    await selectHandCards(page, [0, 1, 2]);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0, 1, 2],
    });
  });
});
