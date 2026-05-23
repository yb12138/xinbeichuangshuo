import type { Page } from '@playwright/test';
import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  SAINTESS_PLAYER_ID,
  ENEMY_PLAYER_ID,
  ALLY_PLAYER_ID,
  frostPrayerScenario,
  frostPrayerTargetPrompt,
  mercyScenario,
  mercyPrompt,
} from '../../../scenarios/saintess';

test.describe('saintess frost prayer protocol harness', () => {
  test('frost prayer: select target after water/light card', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(frostPrayerScenario());

    // Server pushes frost prayer prompt after using water/light card
    await protocolHarness.pushServerMessage(frostPrayerTargetPrompt());

    // Target_picker: click ally player area (auto-submits for single target)
    await page.getByTestId(`player-area-${ALLY_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [2],
    });
  });

  test('frost prayer: select self', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(frostPrayerScenario());

    await protocolHarness.pushServerMessage(frostPrayerTargetPrompt());

    // Click self player area
    await page.getByTestId(`player-area-${SAINTESS_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('frost prayer: select enemy', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(frostPrayerScenario());

    await protocolHarness.pushServerMessage(frostPrayerTargetPrompt());

    // Click enemy player area
    await page.getByTestId(`player-area-${ENEMY_PLAYER_ID}`).click();
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

    // branch_select overlay - confirm (branch-option-0)
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('decision-overlay').getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('mercy: skip', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(mercyScenario());

    await protocolHarness.pushServerMessage(mercyPrompt());

    // branch_select overlay - skip (branch-option-1)
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('decision-overlay').getByTestId('branch-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});