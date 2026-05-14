import { test } from '../../../fixtures/protocolHarness.fixture';
import {
  ENEMY_PLAYER_ID,
  descentCardsPrompt,
  descentConcertoScenario,
  descentElementPrompt,
  descentTargetPrompt,
} from '../../../scenarios/bard';

test.describe('bard descent concerto protocol harness', () => {
  test('full flow: element → cards → target (magic card triggers damage)', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(descentConcertoScenario());

    // Descent auto-triggers at turn end — server pushes element prompt
    await protocolHarness.pushServerMessage(descentElementPrompt());
    // Select Fire element
    await page.getByTestId('prompt-option-Fire').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Select 2 Fire cards
    await protocolHarness.pushServerMessage(descentCardsPrompt('Fire', 2));
    await page.getByTestId('hand-card-0').click();
    await page.getByTestId('hand-card-1').click();
    await page.getByTestId('prompt-confirm-btn').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0, 1],
    });

    // No target prompt follows because Fire cards are Attack type (not Magic)
  });

  test('magic card triggers additional target step', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(descentConcertoScenario());

    // Element prompt
    await protocolHarness.pushServerMessage(descentElementPrompt());
    // Select Water element (index 1) — both Water cards are Magic type
    await page.getByTestId('prompt-option-Water').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });

    // Card picker: select 2 Water cards (indices 2 and 3 in hand)
    await protocolHarness.pushServerMessage(descentCardsPrompt('Water', 2));
    await page.getByTestId('hand-card-2').click();
    await page.getByTestId('hand-card-3').click();
    await page.getByTestId('prompt-confirm-btn').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [2, 3],
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
