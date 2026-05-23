import type { Page } from '@playwright/test';
import { test } from '../../../fixtures/protocolHarness.fixture';
import {
  thunderStrikeExtraDiscardPrompt,
  freezeExtraDiscardPrompt,
  windBladeExtraDiscardPrompt,
  meteorExtraDiscardPrompt,
  fireballExtraDiscardPrompt,
  thunderStrikeScenario,
  freezeScenario,
  windBladeScenario,
  meteorScenario,
  fireballScenario,
} from '../../../scenarios/elementalist';

async function clickOverlayOption(page: Page, selector: string) {
  const overlay = page.getByTestId('decision-overlay');
  const overlayVisible = await overlay.isVisible({ timeout: 1000 }).catch(() => false);
  if (overlayVisible) {
    await overlay.getByTestId(selector).click();
  } else {
    await page.getByTestId('prompt-dialog').getByTestId(selector).click();
  }
}

test.describe('elementalist thunder strike extra discard protocol harness', () => {
  test('thunder strike: discard thunder card for +1 damage', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(thunderStrikeScenario());

    // Server pushes extra discard prompt
    await protocolHarness.pushServerMessage(thunderStrikeExtraDiscardPrompt());

    // Click yes to discard thunder card
    await clickOverlayOption(page, 'branch-option-0');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('thunder strike: skip extra discard', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(thunderStrikeScenario());

    await protocolHarness.pushServerMessage(thunderStrikeExtraDiscardPrompt());

    // Click no to skip
    await clickOverlayOption(page, 'branch-option-1');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});

test.describe('elementalist freeze extra discard protocol harness', () => {
  test('freeze: discard water card for +1 damage', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(freezeScenario());

    await protocolHarness.pushServerMessage(freezeExtraDiscardPrompt());

    await clickOverlayOption(page, 'branch-option-0');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('freeze: skip extra discard', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(freezeScenario());

    await protocolHarness.pushServerMessage(freezeExtraDiscardPrompt());

    await clickOverlayOption(page, 'branch-option-1');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});

test.describe('elementalist wind blade extra discard protocol harness', () => {
  test('wind blade: discard wind card for +1 damage', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(windBladeScenario());

    await protocolHarness.pushServerMessage(windBladeExtraDiscardPrompt());

    await clickOverlayOption(page, 'branch-option-0');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('wind blade: skip extra discard', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(windBladeScenario());

    await protocolHarness.pushServerMessage(windBladeExtraDiscardPrompt());

    await clickOverlayOption(page, 'branch-option-1');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});

test.describe('elementalist meteor extra discard protocol harness', () => {
  test('meteor: discard earth card for +1 damage', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(meteorScenario());

    await protocolHarness.pushServerMessage(meteorExtraDiscardPrompt());

    await clickOverlayOption(page, 'branch-option-0');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('meteor: skip extra discard', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(meteorScenario());

    await protocolHarness.pushServerMessage(meteorExtraDiscardPrompt());

    await clickOverlayOption(page, 'branch-option-1');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});

test.describe('elementalist fireball extra discard protocol harness', () => {
  test('fireball: discard fire card for +1 damage', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(fireballScenario());

    await protocolHarness.pushServerMessage(fireballExtraDiscardPrompt());

    await clickOverlayOption(page, 'branch-option-0');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('fireball: skip extra discard', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(fireballScenario());

    await protocolHarness.pushServerMessage(fireballExtraDiscardPrompt());

    await clickOverlayOption(page, 'branch-option-1');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});
