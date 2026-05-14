import type { Page } from '@playwright/test';
import { test } from '../../../fixtures/protocolHarness.fixture';
import {
  asuraComboScenario,
  asuraComboPrompt,
  asuraComboDiscardPrompt,
  underworldTremorScenario,
  underworldTremorPrompt,
} from '../../../scenarios/magicSwordman';

async function clickOverlayOption(page: Page, selector: string) {
  const overlay = page.getByTestId('decision-overlay');
  const overlayVisible = await overlay.isVisible({ timeout: 1000 }).catch(() => false);
  if (overlayVisible) {
    await overlay.getByTestId(selector).click();
  } else {
    await page.getByTestId('prompt-dialog').getByTestId(selector).click();
  }
}

test.describe('magic swordman asura combo protocol harness', () => {
  test('asura combo: confirm after damage >=2', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(asuraComboScenario());

    // Server pushes asura combo prompt after attack ends
    await protocolHarness.pushServerMessage(asuraComboPrompt());

    // Click confirm button
    await clickOverlayOption(page, 'prompt-option-confirm');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Server pushes discard selection
    await protocolHarness.pushServerMessage(asuraComboDiscardPrompt());

    // Select fire card to discard
    await clickOverlayOption(page, 'prompt-option-ms-fire-attack-1');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('asura combo: skip', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(asuraComboScenario());

    await protocolHarness.pushServerMessage(asuraComboPrompt());

    // Click skip button
    await clickOverlayOption(page, 'prompt-option-skip');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});

test.describe('magic swordman underworld tremor protocol harness', () => {
  test('underworld tremor: confirm with gem', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(underworldTremorScenario());

    // Server pushes underworld tremor prompt before attack
    await protocolHarness.pushServerMessage(underworldTremorPrompt());

    // Click confirm button
    await clickOverlayOption(page, 'prompt-option-confirm');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('underworld tremor: skip', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(underworldTremorScenario());

    await protocolHarness.pushServerMessage(underworldTremorPrompt());

    // Click skip button
    await clickOverlayOption(page, 'prompt-option-skip');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});