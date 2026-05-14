import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  ENEMY_PLAYER_ID,
  ENEMY_2_PLAYER_ID,
  medusaEyeDarkMoonPrompt,
  medusaEyeDiscardPrompt,
  medusaEyeScenario,
  medusaEyeTargetPrompt,
} from '../../../scenarios/moonGoddess';

test.describe('moon goddess medusa eye protocol harness', () => {
  test('medusa eye: dark moon is magic card, discard then target', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(medusaEyeScenario({ dark_moon_cards: 2 }));

    // Server pushes dark moon card choice (盖牌选择) - choose_card type renders as hand card picker
    await protocolHarness.pushServerMessage(medusaEyeDarkMoonPrompt());
    // Select the magic card (first option is 暗月法术)
    await page.getByTestId('hand-card-0').click();
    await page.getByTestId('prompt-confirm-btn').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Dark moon is magic card, discard 1 card
    await protocolHarness.pushServerMessage(medusaEyeDiscardPrompt());
    await page.getByTestId('hand-card-0').click();
    await page.getByTestId('prompt-confirm-btn').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Target selection (click enemy player card)
    await protocolHarness.pushServerMessage(medusaEyeTargetPrompt());
    await page.getByTestId(`player-area-${ENEMY_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('medusa eye: dark moon is attack card, no discard, just target', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(medusaEyeScenario({ dark_moon_cards: 2 }));

    // Server pushes dark moon card choice - choose_card type renders as hand card picker
    await protocolHarness.pushServerMessage(medusaEyeDarkMoonPrompt());
    // Select the attack card (second option is 火焰斩)
    await page.getByTestId('hand-card-1').click();
    await page.getByTestId('prompt-confirm-btn').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });

    // No discard prompt (attack card, not magic), directly target selection
    await protocolHarness.pushServerMessage(medusaEyeTargetPrompt());
    await page.getByTestId(`player-area-${ENEMY_2_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });

  test('medusa eye: select second enemy as target', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(medusaEyeScenario({ dark_moon_cards: 2 }));

    await protocolHarness.pushServerMessage(medusaEyeDarkMoonPrompt());
    await page.getByTestId('hand-card-0').click();
    await page.getByTestId('prompt-confirm-btn').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    await protocolHarness.pushServerMessage(medusaEyeDiscardPrompt());
    await page.getByTestId('hand-card-1').click();
    await page.getByTestId('prompt-confirm-btn').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });

    // Target selection (click enemy 2 player card)
    await protocolHarness.pushServerMessage(medusaEyeTargetPrompt());
    await page.getByTestId(`player-area-${ENEMY_2_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});