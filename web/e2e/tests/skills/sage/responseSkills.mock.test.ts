import type { Page } from '@playwright/test';
import { test } from '../../../fixtures/protocolHarness.fixture';
import {
  ENEMY_PLAYER_ID,
  magicReboundCardsPrompt,
  magicReboundConfirmPrompt,
  magicReboundElementPrompt,
  magicReboundScenario,
  magicReboundTargetPrompt,
  sageScenario,
  wisdomCodexDiscardPrompt,
} from '../../../scenarios/sage';

async function selectHandCards(page: Page, indices: number[]) {
  for (const index of indices) {
    const card = page.getByTestId(`hand-card-${index}`);
    await card.scrollIntoViewIfNeeded();
    await card.click();
  }
  await page.getByTestId('prompt-confirm-btn').click();
}

async function clickOverlayOption(page: Page, selector: string) {
  const overlay = page.getByTestId('decision-overlay');
  const overlayVisible = await overlay.isVisible({ timeout: 1000 }).catch(() => false);
  if (overlayVisible) {
    await overlay.getByTestId(selector).click();
  } else {
    await page.getByTestId('prompt-dialog').getByTestId(selector).click();
  }
}

test.describe('sage wisdom codex protocol harness', () => {
  test('wisdom codex: discard 1 card after magic damage > 3', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(sageScenario());

    // Server pushes system discard prompt (triggered by wisdom codex after magic damage > 3)
    await protocolHarness.pushServerMessage(wisdomCodexDiscardPrompt());

    // Select first hand card to discard
    await selectHandCards(page, [0]);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('wisdom codex: discard different card', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(sageScenario());

    await protocolHarness.pushServerMessage(wisdomCodexDiscardPrompt());

    // Select third hand card
    await selectHandCards(page, [2]);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [2],
    });
  });
});

test.describe('sage magic rebound protocol harness', () => {
  test('magic rebound: skip (否)', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(magicReboundScenario());

    // Server pushes magic rebound confirm prompt (triggered by magic damage == 1)
    await protocolHarness.pushServerMessage(magicReboundConfirmPrompt());

    // Click "否" (yes-no mode uses prompt-option-{id})
    await clickOverlayOption(page, 'prompt-option-1');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });

  test('magic rebound: full flow (confirm → element → cards → target)', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(magicReboundScenario());

    // Step 1: confirm activation
    await protocolHarness.pushServerMessage(magicReboundConfirmPrompt());
    await clickOverlayOption(page, 'prompt-option-0');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Step 2: select element (火系)
    await protocolHarness.pushServerMessage(magicReboundElementPrompt(2));
    await clickOverlayOption(page, 'prompt-option-0');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Step 3: select same-element cards (choose 2 fire cards)
    await protocolHarness.pushServerMessage(magicReboundCardsPrompt());
    await selectHandCards(page, [0, 1]);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0, 1],
    });

    // Step 4: select target (enemy)
    await protocolHarness.pushServerMessage(magicReboundTargetPrompt());
    await page.getByTestId(`player-area-${ENEMY_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('magic rebound: select max fire cards', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(magicReboundScenario());

    // Confirm
    await protocolHarness.pushServerMessage(magicReboundConfirmPrompt());
    await clickOverlayOption(page, 'prompt-option-0');
    await protocolHarness.expectSubmitAction({ action_type: 'Select', option_indexes: [0] });

    // Element
    await protocolHarness.pushServerMessage(magicReboundElementPrompt(2));
    await clickOverlayOption(page, 'prompt-option-0');
    await protocolHarness.expectSubmitAction({ action_type: 'Select', option_indexes: [0] });

    // Select all 3 fire cards
    await protocolHarness.pushServerMessage(magicReboundCardsPrompt());
    await selectHandCards(page, [0, 1, 2]);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0, 1, 2],
    });
  });
});
