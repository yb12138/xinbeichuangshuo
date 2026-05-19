import type { Page } from '@playwright/test';
import { test, expect } from '../../../fixtures/protocolHarness.fixture';
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

test.describe('sage wisdom codex protocol harness', () => {
  test('wisdom codex: discard 1 card after magic damage > 3', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(sageScenario());

    // Server pushes system discard prompt (card_picker from hand)
    await protocolHarness.pushServerMessage(wisdomCodexDiscardPrompt());

    // Select first hand card to discard
    await selectHandCards(page, [0]);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      card_ids: ['sg-fire-atk'],
    });
  });

  test('wisdom codex: discard different card', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(sageScenario());

    await protocolHarness.pushServerMessage(wisdomCodexDiscardPrompt());

    // Select third hand card
    await selectHandCards(page, [2]);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      card_ids: ['sg-thunder-atk'],
    });
  });
});

test.describe('sage magic rebound protocol harness', () => {
  test('magic rebound: skip (否)', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(magicReboundScenario());

    // Server pushes magic rebound confirm prompt (branch_select overlay)
    await protocolHarness.pushServerMessage(magicReboundConfirmPrompt());

    // Click "否" (branch-option-1)
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('decision-overlay').getByTestId('branch-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });

  test('magic rebound: full flow (confirm → element → cards → target)', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(magicReboundScenario());

    // Step 1: confirm activation (branch_select overlay)
    await protocolHarness.pushServerMessage(magicReboundConfirmPrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('decision-overlay').getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Step 2: select element (branch_select overlay)
    await protocolHarness.pushServerMessage(magicReboundElementPrompt(2));
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('decision-overlay').getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Step 3: select same-element cards (card_picker from hand)
    await protocolHarness.pushServerMessage(magicReboundCardsPrompt());
    await selectHandCards(page, [0, 1]);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      card_ids: ['sg-fire-atk1', 'sg-fire-atk2'],
    });

    // Step 4: select target (target_picker - click enemy player area)
    await protocolHarness.pushServerMessage(magicReboundTargetPrompt());
    await expect(page.getByTestId('decision-overlay')).not.toBeVisible({ timeout: 5000 });
    await page.getByTestId(`player-area-${ENEMY_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('magic rebound: select max fire cards', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(magicReboundScenario());

    // Confirm (branch_select)
    await protocolHarness.pushServerMessage(magicReboundConfirmPrompt());
    await page.getByTestId('decision-overlay').getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({ action_type: 'Select', option_indexes: [0] });

    // Element (branch_select)
    await protocolHarness.pushServerMessage(magicReboundElementPrompt(2));
    await page.getByTestId('decision-overlay').getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({ action_type: 'Select', option_indexes: [0] });

    // Select all 3 fire cards (card_picker from hand)
    await protocolHarness.pushServerMessage(magicReboundCardsPrompt());
    await selectHandCards(page, [0, 1, 2]);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      card_ids: ['sg-fire-atk1', 'sg-fire-atk2', 'sg-fire-magic'],
    });
  });
});