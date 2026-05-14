import type { Page } from '@playwright/test';
import { test } from '../../../fixtures/protocolHarness.fixture';
import {
  ALLY_PLAYER_ID,
  ENEMY_PLAYER_ID,
  manaInversionCardsPrompt,
  manaInversionScenario,
  manaInversionTargetPrompt,
  manaInversionXPrompt,
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

    // Click the first magic card option (火球, id='2', option index 0)
    const overlay = page.getByTestId('decision-overlay');
    const overlayVisible = await overlay.isVisible({ timeout: 1000 }).catch(() => false);
    if (overlayVisible) {
      await overlay.getByTestId('branch-option-0').click();
    } else {
      await page.getByTestId('prompt-dialog').getByTestId('branch-option-0').click();
    }
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
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

    // Click the third magic card option (雷击, id='4', option index 2)
    const overlay = page.getByTestId('decision-overlay');
    const overlayVisible = await overlay.isVisible({ timeout: 1000 }).catch(() => false);
    if (overlayVisible) {
      await overlay.getByTestId('branch-option-2').click();
    } else {
      await page.getByTestId('prompt-dialog').getByTestId('branch-option-2').click();
    }
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [2],
    });
  });
});

test.describe('blaze witch mana inversion protocol harness', () => {
  test('mana inversion: select X=2 then pick cards then target', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(manaInversionScenario());

    // Step 1: push X selection (numeric overlay)
    await protocolHarness.pushServerMessage(manaInversionXPrompt(4));

    // Select X=2 from numeric overlay
    await page.getByTestId('decision-overlay').getByTestId('numeric-option-2').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Step 2: push card selection (choose_cards, 2 magic cards)
    await protocolHarness.pushServerMessage(manaInversionCardsPrompt(2));

    // Select 2 magic cards
    await selectHandCards(page, [2, 3]);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [2, 3],
    });

    // Step 3: push target prompt
    await protocolHarness.pushServerMessage(manaInversionTargetPrompt());

    // Click enemy player area
    await page.getByTestId(`player-area-${ENEMY_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('mana inversion: select X=3 with different cards', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(manaInversionScenario());

    // Select X=3
    await protocolHarness.pushServerMessage(manaInversionXPrompt(4));
    await page.getByTestId('decision-overlay').getByTestId('numeric-option-3').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });

    // Select 3 magic cards
    await protocolHarness.pushServerMessage(manaInversionCardsPrompt(3));
    await selectHandCards(page, [2, 3, 4]);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [2, 3, 4],
    });
  });
});
