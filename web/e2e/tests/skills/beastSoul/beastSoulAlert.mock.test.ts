import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  ENEMY_PLAYER_ID,
  beastSoulAlertConfirmPrompt,
  beastSoulAlertScenario,
  beastSoulAlertTargetPrompt,
  beastSoulAlertDiscardPrompt,
} from '../../../scenarios/beastSoul';

test.describe('beast soul warrior beast soul alert protocol harness', () => {
  test('beast soul alert: confirm then target then discard', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(beastSoulAlertScenario({ beast_souls: 2 }));

    // Server pushes confirm prompt when trigger condition met
    await protocolHarness.pushServerMessage(beastSoulAlertConfirmPrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('prompt-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Target selection (click enemy player card)
    await protocolHarness.pushServerMessage(beastSoulAlertTargetPrompt());
    await page.getByTestId(`player-area-${ENEMY_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Discard 1 card
    await protocolHarness.pushServerMessage(beastSoulAlertDiscardPrompt());
    await page.getByTestId('hand-card-0').click();
    await page.getByTestId('prompt-confirm-btn').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('beast soul alert: skip confirm', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(beastSoulAlertScenario({ beast_souls: 2 }));

    await protocolHarness.pushServerMessage(beastSoulAlertConfirmPrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('prompt-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});