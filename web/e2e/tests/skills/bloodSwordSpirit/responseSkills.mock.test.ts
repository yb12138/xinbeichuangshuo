import type { Page } from '@playwright/test';
import { test } from '../../../fixtures/protocolHarness.fixture';
import {
  redFlashScenario,
  redFlashPrompt,
  bloodBarrierScenario,
  bloodBarrierPrompt,
} from '../../../scenarios/bloodSwordSpirit';

async function clickOverlayOption(page: Page, selector: string) {
  const overlay = page.getByTestId('decision-overlay');
  const overlayVisible = await overlay.isVisible({ timeout: 1000 }).catch(() => false);
  if (overlayVisible) {
    await overlay.getByTestId(selector).click();
  } else {
    await page.getByTestId('prompt-dialog').getByTestId(selector).click();
  }
}

test.describe('blood sword spirit red flash protocol harness', () => {
  test('red flash: confirm self damage for +1 attack damage', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(redFlashScenario());

    // Server pushes red flash prompt before attack
    await protocolHarness.pushServerMessage(redFlashPrompt());

    // Click confirm button
    await clickOverlayOption(page, 'prompt-option-confirm');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('red flash: skip', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(redFlashScenario());

    await protocolHarness.pushServerMessage(redFlashPrompt());

    // Click skip button
    await clickOverlayOption(page, 'prompt-option-skip');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });

  test('red flash: can trigger multiple times', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(redFlashScenario({ bloodCount: 3 }));

    await protocolHarness.pushServerMessage(redFlashPrompt());

    // First trigger
    await clickOverlayOption(page, 'prompt-option-confirm');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Server pushes again for second trigger
    await protocolHarness.pushServerMessage(redFlashPrompt());

    // Second trigger
    await clickOverlayOption(page, 'prompt-option-confirm');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });
});

test.describe('blood sword spirit blood barrier protocol harness', () => {
  test('blood barrier: confirm to reduce magic damage', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(bloodBarrierScenario());

    // Server pushes blood barrier prompt when receiving magic damage
    await protocolHarness.pushServerMessage(bloodBarrierPrompt());

    // Click confirm button
    await clickOverlayOption(page, 'prompt-option-confirm');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('blood barrier: skip', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(bloodBarrierScenario());

    await protocolHarness.pushServerMessage(bloodBarrierPrompt());

    // Click skip button
    await clickOverlayOption(page, 'prompt-option-skip');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});