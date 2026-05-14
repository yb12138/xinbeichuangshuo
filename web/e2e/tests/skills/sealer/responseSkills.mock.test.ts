import type { Page } from '@playwright/test';
import { test } from '../../../fixtures/protocolHarness.fixture';
import {
  magicSurgeScenario,
  magicSurgePrompt,
  fiveElementsBindCancelPrompt,
  sealTriggerPrompt,
  SEALER_WATER_SEAL_ID,
} from '../../../scenarios/sealer';

async function clickOverlayOption(page: Page, selector: string) {
  const overlay = page.getByTestId('decision-overlay');
  const overlayVisible = await overlay.isVisible({ timeout: 1000 }).catch(() => false);
  if (overlayVisible) {
    await overlay.getByTestId(selector).click();
  } else {
    await page.getByTestId('prompt-dialog').getByTestId(selector).click();
  }
}

test.describe('sealer magic surge protocol harness', () => {
  test('magic surge: confirm extra attack action', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(magicSurgeScenario());

    // Server pushes magic surge prompt after spell action ends
    await protocolHarness.pushServerMessage(magicSurgePrompt());

    // Click confirm button
    await clickOverlayOption(page, 'prompt-option-confirm');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('magic surge: skip', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(magicSurgeScenario());

    await protocolHarness.pushServerMessage(magicSurgePrompt());

    // Click skip button
    await clickOverlayOption(page, 'prompt-option-skip');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});

test.describe('sealer five elements bind cancel protocol harness', () => {
  test('five elements bind cancel: enemy draws cards', async ({ protocolHarness }) => {
    await protocolHarness.bootGame(magicSurgeScenario());

    // Server pushes cancel prompt to enemy (X=0, draw 2 cards)
    await protocolHarness.pushServerMessage(fiveElementsBindCancelPrompt(0));

    // Click draw option
    await clickOverlayOption(page, 'prompt-option-draw');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('five elements bind cancel: enemy skips', async ({ protocolHarness }) => {
    await protocolHarness.bootGame(magicSurgeScenario());

    await protocolHarness.pushServerMessage(fiveElementsBindCancelPrompt(0));

    // Click skip option
    await clickOverlayOption(page, 'prompt-option-skip');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });

  test('five elements bind cancel: X=2 draws 4 cards', async ({ protocolHarness }) => {
    await protocolHarness.bootGame(magicSurgeScenario());

    // Server pushes cancel prompt to enemy (X=2, draw 4 cards)
    await protocolHarness.pushServerMessage(fiveElementsBindCancelPrompt(2));

    await clickOverlayOption(page, 'prompt-option-draw');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });
});

test.describe('sealer seal trigger protocol harness', () => {
  test('seal trigger: enemy confirms damage', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(magicSurgeScenario());

    // Server pushes seal trigger prompt to enemy
    await protocolHarness.pushServerMessage(sealTriggerPrompt(SEALER_WATER_SEAL_ID));

    // Click confirm button
    await clickOverlayOption(page, 'prompt-option-confirm');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });
});