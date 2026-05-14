import type { Page } from '@playwright/test';
import { test } from '../../../fixtures/protocolHarness.fixture';
import {
  stealthScenario,
  stealthConfirmPrompt,
  stealthDrawPrompt,
  waterShadowScenario,
  waterShadowBeforeDrawPrompt,
  waterShadowStealthExtraPrompt,
  waterShadowExtraCardPrompt,
} from '../../../scenarios/assassin';

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

test.describe('assassin stealth protocol harness', () => {
  test('stealth: confirm with gem', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(stealthScenario({ gem: 1 }));

    // Server pushes stealth confirm prompt
    await protocolHarness.pushServerMessage(stealthConfirmPrompt());

    // Click confirm button
    await clickOverlayOption(page, 'prompt-option-confirm');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0], // confirm is first option
    });
  });

  test('stealth: skip', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(stealthScenario({ gem: 1 }));

    await protocolHarness.pushServerMessage(stealthConfirmPrompt());

    // Click skip button
    await clickOverlayOption(page, 'prompt-option-skip');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1], // skip is second option
    });
  });

  test('stealth: draw after confirm', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(stealthScenario({ gem: 1 }));

    // First confirm stealth
    await protocolHarness.pushServerMessage(stealthConfirmPrompt());
    await clickOverlayOption(page, 'prompt-option-confirm');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Then choose to draw
    await protocolHarness.pushServerMessage(stealthDrawPrompt());
    await clickOverlayOption(page, 'prompt-option-draw');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('stealth: skip draw after confirm', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(stealthScenario({ gem: 1 }));

    // First confirm stealth
    await protocolHarness.pushServerMessage(stealthConfirmPrompt());
    await clickOverlayOption(page, 'prompt-option-confirm');

    // Then choose not to draw
    await protocolHarness.pushServerMessage(stealthDrawPrompt());
    await clickOverlayOption(page, 'prompt-option-skip');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});

test.describe('assassin water shadow protocol harness', () => {
  test('water shadow: discard water cards before draw', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(waterShadowScenario());

    // Server pushes water shadow prompt before draw
    await protocolHarness.pushServerMessage(waterShadowBeforeDrawPrompt());

    // Select 2 water cards (indices 0, 1)
    await selectHandCards(page, [0, 1]);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0, 1],
    });
  });

  test('water shadow: discard 0 cards', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(waterShadowScenario());

    await protocolHarness.pushServerMessage(waterShadowBeforeDrawPrompt());

    // Click confirm without selecting any cards (min: 0)
    await page.getByTestId('prompt-confirm-btn').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [],
    });
  });

  test('water shadow: in stealth extra discard magic card', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(waterShadowScenario({ inStealth: true }));

    // First push water shadow prompt
    await protocolHarness.pushServerMessage(waterShadowBeforeDrawPrompt());
    await selectHandCards(page, [0]);

    // Then push extra magic card prompt (in stealth)
    await protocolHarness.pushServerMessage(waterShadowStealthExtraPrompt());

    // Choose to discard magic card
    await clickOverlayOption(page, 'prompt-option-yes');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Then select magic card
    await protocolHarness.pushServerMessage(waterShadowExtraCardPrompt());
    await selectHandCards(page, [2]); // ice magic
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('water shadow: in stealth skip extra discard', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(waterShadowScenario({ inStealth: true }));

    await protocolHarness.pushServerMessage(waterShadowBeforeDrawPrompt());
    await selectHandCards(page, [0]);

    // Push extra magic card prompt
    await protocolHarness.pushServerMessage(waterShadowStealthExtraPrompt());

    // Choose not to discard
    await clickOverlayOption(page, 'prompt-option-no');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});