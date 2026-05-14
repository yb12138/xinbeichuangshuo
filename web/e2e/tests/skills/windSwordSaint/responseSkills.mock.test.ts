import type { Page } from '@playwright/test';
import { test } from '../../../fixtures/protocolHarness.fixture';
import {
  windFuryScenario,
  windFuryPrompt,
  swordShadowScenario,
  swordShadowPrompt,
  wssComboScenario,
  wssComboPrompt,
} from '../../../scenarios/windSwordSaint';

async function clickOverlayOption(page: Page, selector: string) {
  const overlay = page.getByTestId('decision-overlay');
  const overlayVisible = await overlay.isVisible({ timeout: 1000 }).catch(() => false);
  if (overlayVisible) {
    await overlay.getByTestId(selector).click();
  } else {
    await page.getByTestId('prompt-dialog').getByTestId(selector).click();
  }
}

test.describe('wind sword saint wind fury protocol harness', () => {
  test('wind fury: confirm after attack action', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(windFuryScenario());

    // Server pushes wind fury prompt after attack action ends
    await protocolHarness.pushServerMessage(windFuryPrompt());

    // Click confirm button
    await clickOverlayOption(page, 'prompt-option-confirm');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('wind fury: skip', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(windFuryScenario());

    await protocolHarness.pushServerMessage(windFuryPrompt());

    // Click skip button
    await clickOverlayOption(page, 'prompt-option-skip');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});

test.describe('wind sword saint sword shadow protocol harness', () => {
  test('sword shadow: confirm with crystal', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(swordShadowScenario({ crystal: 1 }));

    // Server pushes sword shadow prompt after attack action ends
    await protocolHarness.pushServerMessage(swordShadowPrompt());

    // Click confirm button
    await clickOverlayOption(page, 'prompt-option-confirm');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('sword shadow: skip', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(swordShadowScenario({ crystal: 1 }));

    await protocolHarness.pushServerMessage(swordShadowPrompt());

    // Click skip button
    await clickOverlayOption(page, 'prompt-option-skip');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});

test.describe('wind sword saint combo protocol harness', () => {
  test('combo: select wind fury when both available', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(wssComboScenario({ crystal: 1 }));

    // Server pushes combo prompt when both skills available
    await protocolHarness.pushServerMessage(wssComboPrompt());

    // Select wind fury
    await clickOverlayOption(page, 'prompt-option-wind_fury');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('combo: select sword shadow when both available', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(wssComboScenario({ crystal: 1 }));

    await protocolHarness.pushServerMessage(wssComboPrompt());

    // Select sword shadow
    await clickOverlayOption(page, 'prompt-option-sword_shadow');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });

  test('combo: skip both', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(wssComboScenario({ crystal: 1 }));

    await protocolHarness.pushServerMessage(wssComboPrompt());

    // Select skip
    await clickOverlayOption(page, 'prompt-option-skip');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [2],
    });
  });
});