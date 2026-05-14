import { test } from '../../../fixtures/protocolHarness.fixture';
import {
  shardStormDiscardPrompt,
  shardStormMissHealPrompt,
  shardStormMissTargetPrompt,
  shardStormScenario,
  ALLY_PLAYER_ID,
} from '../../../scenarios/holyBow';

test.describe('holy bow shard storm protocol harness', () => {
  test('discard phase for shard storm', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(shardStormScenario());

    // Server pushes discard prompt directly (skill validation done by backend)
    await protocolHarness.pushServerMessage(shardStormDiscardPrompt());

    // Discard 2 same-element attack cards
    await page.getByTestId('hand-card-0').click();
    await page.getByTestId('hand-card-1').click();
    await page.getByTestId('prompt-confirm-btn').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0, 1],
    });
  });

  test('shard storm miss follow-up: heal removal and ally discard', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(shardStormScenario());

    // Discard phase
    await protocolHarness.pushServerMessage(shardStormDiscardPrompt());
    await page.getByTestId('hand-card-0').click();
    await page.getByTestId('hand-card-1').click();
    await page.getByTestId('prompt-confirm-btn').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0, 1],
    });

    // Attack miss: choose heal removal (numeric mode)
    await protocolHarness.pushServerMessage(shardStormMissHealPrompt());
    await page.getByTestId('numeric-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });

    // Target selection (click ally player card)
    await protocolHarness.pushServerMessage(shardStormMissTargetPrompt());
    await page.getByTestId(`player-area-${ALLY_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });
});