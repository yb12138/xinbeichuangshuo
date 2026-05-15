import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  ENEMY_PLAYER_ID,
  ENEMY_2_PLAYER_ID,
  rousingDiscardCardsPrompt,
  rousingModePrompt,
  rousingRhapsodyScenario,
  rousingTargetsPrompt,
} from '../../../scenarios/bard';

test.describe('bard rousing rhapsody protocol harness', () => {
  test('branch 0: damage 2 enemy targets', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(rousingRhapsodyScenario());

    // Rousing auto-triggers at turn start — server pushes mode selection (3 branches)
    await protocolHarness.pushServerMessage(rousingModePrompt());
    // Select branch 0 = damage 2 enemies
    await page.getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Target step 1/2 (click on enemy player card)
    await protocolHarness.pushServerMessage(rousingTargetsPrompt(1));
    await page.getByTestId(`player-area-${ENEMY_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Target step 2/2 (click remaining enemy)
    await protocolHarness.pushServerMessage(rousingTargetsPrompt(2));
    await page.getByTestId(`player-area-${ENEMY_2_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('branch 1: discard 2 cards', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(rousingRhapsodyScenario());

    // Mode: branch 1 = discard 2 cards
    await protocolHarness.pushServerMessage(rousingModePrompt());
    await page.getByTestId('branch-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });

    // Discard step - select 2 hand cards
    await protocolHarness.pushServerMessage(rousingDiscardCardsPrompt(2));
    await page.getByTestId('hand-card-0').click();
    await page.getByTestId('hand-card-1').click();
    await page.getByTestId('prompt-confirm-btn').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0, 1],
    });
  });

  test('branch 2: skip (do not activate)', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(rousingRhapsodyScenario());

    await protocolHarness.pushServerMessage(rousingModePrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    // Select 跳过 (branch index 2)
    await page.getByTestId('branch-option-2').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [2],
    });
  });
});
