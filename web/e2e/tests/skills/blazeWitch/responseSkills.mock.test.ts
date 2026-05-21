import type { Page } from '@playwright/test';
import { test } from '../../../fixtures/protocolHarness.fixture';
import {
  ALLY_PLAYER_ID,
  ENEMY_PLAYER_ID,
  manaInversionCardsPrompt,
  manaInversionScenario,
  manaInversionTargetPrompt,
  substituteDollCardPrompt,
  substituteDollScenario,
  substituteDollTargetPrompt,
} from '../../../scenarios/blazeWitch';

async function selectHandCards(page: Page, indices: number[]) {
  for (const index of indices) {
    const card = page.getByTestId(`hand-card-${index}`);
    await card.scrollIntoViewIfNeeded();
    await card.click();
  }
  await page.getByTestId('prompt-confirm-btn').click();
}

test.describe('blaze witch substitute doll protocol harness', () => {
  test('substitute doll: select magic card then ally target', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(substituteDollScenario());

    // Step 1: push magic card selection prompt (confirm type, magic hand indices)
    await protocolHarness.pushServerMessage(substituteDollCardPrompt());

    // Click the first magic card (火球)
    await selectHandCards(page, [2]);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      card_ids: ['bw-fire-magic1'],
    });

    // Step 2: push ally target prompt
    await protocolHarness.pushServerMessage(substituteDollTargetPrompt());

    // Click ally player area
    await page.getByTestId(`player-area-${ALLY_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('substitute doll: select different magic card', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(substituteDollScenario());

    await protocolHarness.pushServerMessage(substituteDollCardPrompt());

    // Click the third magic card option (雷击)
    await selectHandCards(page, [4]);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      card_ids: ['bw-thunder-magic'],
    });
  });
});

test.describe('blaze witch mana inversion protocol harness', () => {
  test('mana inversion: pick 2 magic cards as X then target', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(manaInversionScenario());

    // Step 1: push card selection (choose_cards, X is selected card count)
    await protocolHarness.pushServerMessage(manaInversionCardsPrompt());

    // Select 2 magic cards
    await selectHandCards(page, [2, 3]);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      card_ids: ['bw-fire-magic1', 'bw-fire-magic2'],
    });

    // Step 2: push target prompt
    await protocolHarness.pushServerMessage(manaInversionTargetPrompt());

    // Click enemy player area
    await page.getByTestId(`player-area-${ENEMY_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('mana inversion: pick 3 magic cards as X', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(manaInversionScenario());

    // Select 3 magic cards
    await protocolHarness.pushServerMessage(manaInversionCardsPrompt());
    await selectHandCards(page, [2, 3, 4]);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      card_ids: ['bw-fire-magic1', 'bw-fire-magic2', 'bw-thunder-magic'],
    });
  });
});
