import type { Page } from '@playwright/test';
import { test } from '../../../fixtures/protocolHarness.fixture';
import {
  powerBlessingTriggerPrompt,
  swiftBlessingTriggerPrompt,
  manaTideScenario,
  manaTidePrompt,
} from '../../../scenarios/prayer';

async function clickOverlayOption(page: Page, selector: string) {
  const overlay = page.getByTestId('decision-overlay');
  const overlayVisible = await overlay.isVisible({ timeout: 1000 }).catch(() => false);
  if (overlayVisible) {
    await overlay.getByTestId(selector).click();
  } else {
    await page.getByTestId('prompt-dialog').getByTestId(selector).click();
  }
}

test.describe('prayer power blessing trigger protocol harness', () => {
  test('power blessing: ally triggers on hit', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(powerBlessingTriggerPrompt() as any);

    // Server pushes power blessing trigger prompt to ally when they hit
    await protocolHarness.pushServerMessage(powerBlessingTriggerPrompt());

    // Click confirm button
    await clickOverlayOption(page, 'prompt-option-confirm');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('power blessing: ally skips trigger', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(powerBlessingTriggerPrompt() as any);

    await protocolHarness.pushServerMessage(powerBlessingTriggerPrompt());

    // Click skip button
    await clickOverlayOption(page, 'prompt-option-skip');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});

test.describe('prayer swift blessing trigger protocol harness', () => {
  test('swift blessing: ally triggers on action end', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(swiftBlessingTriggerPrompt() as any);

    // Server pushes swift blessing trigger prompt to ally after action
    await protocolHarness.pushServerMessage(swiftBlessingTriggerPrompt());

    // Click confirm button
    await clickOverlayOption(page, 'prompt-option-confirm');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('swift blessing: ally skips trigger', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(swiftBlessingTriggerPrompt() as any);

    await protocolHarness.pushServerMessage(swiftBlessingTriggerPrompt());

    // Click skip button
    await clickOverlayOption(page, 'prompt-option-skip');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});

test.describe('prayer mana tide protocol harness', () => {
  test('mana tide: confirm extra magic action', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(manaTideScenario());

    // Server pushes mana tide prompt after magic action ends
    await protocolHarness.pushServerMessage(manaTidePrompt());

    // Click confirm button
    await clickOverlayOption(page, 'prompt-option-confirm');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('mana tide: skip', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(manaTideScenario());

    await protocolHarness.pushServerMessage(manaTidePrompt());

    // Click skip button
    await clickOverlayOption(page, 'prompt-option-skip');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});