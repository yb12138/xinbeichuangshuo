import type { Page } from '@playwright/test';
import { test } from '../../../fixtures/protocolHarness.fixture';
import {
  piercingShotScenario,
  piercingShotMissPrompt,
  piercingShotDiscardPrompt,
  preciseShotScenario,
  preciseShotConfirmPrompt,
} from '../../../scenarios/archer';

async function clickOverlayOption(page: Page, selector: string) {
  const overlay = page.getByTestId('decision-overlay');
  const overlayVisible = await overlay.isVisible({ timeout: 1000 }).catch(() => false);
  if (overlayVisible) {
    await overlay.getByTestId(selector).click();
  } else {
    await page.getByTestId('prompt-dialog').getByTestId(selector).click();
  }
}

async function selectHandCards(page: Page, indices: number[]) {
  for (const index of indices) {
    const card = page.getByTestId(`hand-card-${index}`);
    await card.scrollIntoViewIfNeeded();
    await card.click();
  }
  await page.getByTestId('prompt-confirm-btn').click();
}

test.describe('archer piercing shot protocol harness', () => {
  test('piercing shot: confirm after miss', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(piercingShotScenario());

    // Server pushes piercing shot prompt after attack misses
    await protocolHarness.pushServerMessage(piercingShotMissPrompt());

    // Click confirm button
    await clickOverlayOption(page, 'prompt-option-piercing_shot');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Server pushes discard prompt
    await protocolHarness.pushServerMessage(piercingShotDiscardPrompt());

    // Select magic card
    await selectHandCards(page, [2]);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      card_ids: ['archer-thunder-magic'],
    });
  });

  test('piercing shot: skip after miss', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(piercingShotScenario());

    await protocolHarness.pushServerMessage(piercingShotMissPrompt());

    // Click skip button
    await clickOverlayOption(page, 'prompt-option-skip');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});

test.describe('archer precise shot protocol harness', () => {
  test('precise shot: confirm forced hit', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(preciseShotScenario());

    // Server pushes precise shot prompt
    await protocolHarness.pushServerMessage(preciseShotConfirmPrompt());

    // Click confirm button
    await clickOverlayOption(page, 'prompt-option-precise_shot');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('precise shot: skip', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(preciseShotScenario());

    await protocolHarness.pushServerMessage(preciseShotConfirmPrompt());

    // Click skip button
    await clickOverlayOption(page, 'prompt-option-skip');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});
