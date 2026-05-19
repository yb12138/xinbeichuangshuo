import { test } from '../../../fixtures/protocolHarness.fixture';
import {
  ENEMY_PLAYER_ID,
  descentCardsDirectPrompt,
  descentConcertoScenario,
  descentTargetPrompt,
} from '../../../scenarios/bard';

test.describe('bard descent concerto protocol harness', () => {
  test('full flow: direct card picker → target (magic card triggers damage)', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(descentConcertoScenario());

    // Descent auto-triggers at turn end — server now pushes card picker directly
    // No element selection step anymore
    await protocolHarness.pushServerMessage(descentCardsDirectPrompt(2, [0, 1, 2, 3]));

    // Select 2 Fire cards (indices 0 and 1)
    await page.getByTestId('hand-card-0').click();
    await page.getByTestId('hand-card-1').click();
    await page.getByTestId('prompt-confirm-btn').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      card_ids: ['fire-atk-1', 'fire-atk-2'],
    });

    // No target prompt follows because Fire cards are Attack type (not Magic)
  });

  test('magic card triggers additional target step', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(descentConcertoScenario());

    // Direct card picker: all candidate indices from elements with >= 2 cards
    await protocolHarness.pushServerMessage(descentCardsDirectPrompt(2, [0, 1, 2, 3]));

    // Card picker: select 2 Water cards (indices 2 and 3 in hand)
    await page.getByTestId('hand-card-2').click();
    await page.getByTestId('hand-card-3').click();
    await page.getByTestId('prompt-confirm-btn').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      card_ids: ['water-mgc-1', 'water-mgc-2'],
    });

    // Target: magic card triggers 1 magic damage to enemy (click on enemy player card)
    await protocolHarness.pushServerMessage(descentTargetPrompt());
    await page.getByTestId(`player-area-${ENEMY_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });
});
