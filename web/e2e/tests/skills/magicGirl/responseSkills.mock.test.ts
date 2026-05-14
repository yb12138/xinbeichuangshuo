import type { Page } from '@playwright/test';
import { test } from '../../../fixtures/protocolHarness.fixture';
import {
  magicBulletControlScenario,
  magicBulletDirectionPrompt,
  magicBulletFusionScenario,
  magicExplosionEnemyDiscardPrompt,
  ENEMY_PLAYER_ID,
  ENEMY2_PLAYER_ID,
} from '../../../scenarios/magicGirl';

async function clickOverlayOption(page: Page, selector: string) {
  const overlay = page.getByTestId('decision-overlay');
  const overlayVisible = await overlay.isVisible({ timeout: 1000 }).catch(() => false);
  if (overlayVisible) {
    await overlay.getByTestId(selector).click();
  } else {
    await page.getByTestId('prompt-dialog').getByTestId(selector).click();
  }
}

test.describe('magic girl magic bullet control protocol harness', () => {
  test('magic bullet control: forward direction', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(magicBulletControlScenario());

    // Server pushes direction choice after using magic bullet
    await protocolHarness.pushServerMessage(magicBulletDirectionPrompt());

    // Click forward option
    await clickOverlayOption(page, 'prompt-option-forward');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('magic bullet control: reverse direction', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(magicBulletControlScenario());

    await protocolHarness.pushServerMessage(magicBulletDirectionPrompt());

    // Click reverse option
    await clickOverlayOption(page, 'prompt-option-reverse');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});

test.describe('magic girl magic bullet fusion protocol harness', () => {
  test('magic bullet fusion: scenario setup', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(magicBulletFusionScenario());
    // Magic bullet fusion is a passive response skill
    // Testing scenario setup only
  });
});

test.describe('magic girl magic explosion enemy discard protocol harness', () => {
  test('magic explosion: enemy discards magic card', async ({ protocolHarness }) => {
    await protocolHarness.bootGame(magicBulletControlScenario());

    // Server pushes discard prompt to enemy
    await protocolHarness.pushServerMessage(magicExplosionEnemyDiscardPrompt(ENEMY_PLAYER_ID));

    // Enemy selects card to discard
    await clickOverlayOption(page, 'prompt-option-0');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('magic explosion: enemy refuses to discard', async ({ protocolHarness }) => {
    await protocolHarness.bootGame(magicBulletControlScenario());

    await protocolHarness.pushServerMessage(magicExplosionEnemyDiscardPrompt(ENEMY2_PLAYER_ID));

    // Enemy skips discard (takes damage instead)
    // min=0 allows skipping
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [],
    });
  });
});