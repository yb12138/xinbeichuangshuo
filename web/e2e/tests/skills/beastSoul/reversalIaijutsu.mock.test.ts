import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  ENEMY_PLAYER_ID,
  reversalIaijutsuConfirmPrompt,
  reversalIaijutsuRemovePrompt,
  reversalIaijutsuScenario,
  reversalIaijutsuTargetPrompt,
} from '../../../scenarios/beastSoul';

test.describe('beast soul warrior reversal iaijutsu protocol harness', () => {
  test('reversal iaijutsu: confirm then remove beast souls then target', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(reversalIaijutsuScenario({ beast_souls: 3 }));

    // Server pushes confirm prompt as response skill
    await protocolHarness.pushServerMessage(reversalIaijutsuConfirmPrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('prompt-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Remove X beast souls (choose X=2)
    await protocolHarness.pushServerMessage(reversalIaijutsuRemovePrompt(3));
    await page.getByTestId('branch-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });

    // Target selection (click enemy player card, enemy discards X+2=4 cards)
    await protocolHarness.pushServerMessage(reversalIaijutsuTargetPrompt(2));
    await page.getByTestId(`player-area-${ENEMY_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('reversal iaijutsu: skip confirm', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(reversalIaijutsuScenario({ beast_souls: 3 }));

    await protocolHarness.pushServerMessage(reversalIaijutsuConfirmPrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('prompt-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });

  test('reversal iaijutsu: remove 1 beast soul, target discards 3 cards', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(reversalIaijutsuScenario({ beast_souls: 3 }));

    await protocolHarness.pushServerMessage(reversalIaijutsuConfirmPrompt());
    await page.getByTestId('prompt-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Choose X=1
    await protocolHarness.pushServerMessage(reversalIaijutsuRemovePrompt(3));
    await page.getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Target discards X+2=3 cards
    await protocolHarness.pushServerMessage(reversalIaijutsuTargetPrompt(1));
    await page.getByTestId(`player-area-${ENEMY_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('reversal iaijutsu: remove 3 beast souls, target discards 5 cards', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(reversalIaijutsuScenario({ beast_souls: 3 }));

    await protocolHarness.pushServerMessage(reversalIaijutsuConfirmPrompt());
    await page.getByTestId('prompt-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Choose X=3 (max)
    await protocolHarness.pushServerMessage(reversalIaijutsuRemovePrompt(3));
    await page.getByTestId('branch-option-2').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [2],
    });

    // Target discards X+2=5 cards
    await protocolHarness.pushServerMessage(reversalIaijutsuTargetPrompt(3));
    await page.getByTestId(`player-area-${ENEMY_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });
});