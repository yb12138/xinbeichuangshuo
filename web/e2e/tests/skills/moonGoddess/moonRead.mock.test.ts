import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  ENEMY_PLAYER_ID,
  ENEMY_2_PLAYER_ID,
  moonReadConfirmPrompt,
  moonReadScenario,
  moonReadTargetPrompt,
} from '../../../scenarios/moonGoddess';

test.describe('moon goddess moon read protocol harness', () => {
  test('moon read: confirm then target', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(moonReadScenario({ heal: 2 }));

    // Server pushes confirm prompt after magic damage draws
    await protocolHarness.pushServerMessage(moonReadConfirmPrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('prompt-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Target selection (click enemy player card)
    await protocolHarness.pushServerMessage(moonReadTargetPrompt());
    await page.getByTestId(`player-area-${ENEMY_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('moon read: skip confirm', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(moonReadScenario({ heal: 2 }));

    await protocolHarness.pushServerMessage(moonReadConfirmPrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('prompt-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });

  test('moon read: no trigger when heal = 0', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(moonReadScenario({ heal: 0 }));

    // Should not trigger if no heal available
    await expect(page.getByTestId('decision-overlay')).not.toBeVisible();
  });

  test('moon read: select second enemy as target', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(moonReadScenario({ heal: 2 }));

    await protocolHarness.pushServerMessage(moonReadConfirmPrompt());
    await page.getByTestId('prompt-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Target selection (click enemy 2 player card)
    await protocolHarness.pushServerMessage(moonReadTargetPrompt());
    await page.getByTestId(`player-area-${ENEMY_2_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});