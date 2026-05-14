import type { Page } from '@playwright/test';
import { test } from '../../../fixtures/protocolHarness.fixture';
import {
  SAINTESS_PLAYER_ID,
  ENEMY_PLAYER_ID,
  ALLY_PLAYER_ID,
  frostPrayerScenario,
  frostPrayerTargetPrompt,
  mercyScenario,
  mercyPrompt,
} from '../../../scenarios/saintess';

async function clickOverlayOption(page: Page, selector: string) {
  const overlay = page.getByTestId('decision-overlay');
  const overlayVisible = await overlay.isVisible({ timeout: 1000 }).catch(() => false);
  if (overlayVisible) {
    await overlay.getByTestId(selector).click();
  } else {
    await page.getByTestId('prompt-dialog').getByTestId(selector).click();
  }
}

async function selectTarget(page: Page, targetId: string) {
  await page.getByTestId(`player-area-${targetId}`).click();
}

test.describe('saintess frost prayer protocol harness', () => {
  test('frost prayer: select target after water/light card', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(frostPrayerScenario());

    // Server pushes frost prayer prompt after using water/light card
    await protocolHarness.pushServerMessage(frostPrayerTargetPrompt());

    // Select ally target
    await selectTarget(page, ALLY_PLAYER_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [2],
    });
  });

  test('frost prayer: select self', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(frostPrayerScenario());

    await protocolHarness.pushServerMessage(frostPrayerTargetPrompt());

    // Select self
    await selectTarget(page, SAINTESS_PLAYER_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('frost prayer: select enemy', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(frostPrayerScenario());

    await protocolHarness.pushServerMessage(frostPrayerTargetPrompt());

    // Select enemy
    await selectTarget(page, ENEMY_PLAYER_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});

test.describe('saintess mercy protocol harness', () => {
  test('mercy: confirm with gem', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(mercyScenario());

    // Server pushes mercy prompt at startup phase
    await protocolHarness.pushServerMessage(mercyPrompt());

    // Click confirm button
    await clickOverlayOption(page, 'prompt-option-confirm');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('mercy: skip', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(mercyScenario());

    await protocolHarness.pushServerMessage(mercyPrompt());

    // Click skip button
    await clickOverlayOption(page, 'prompt-option-skip');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});