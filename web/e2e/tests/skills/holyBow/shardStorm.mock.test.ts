import { test } from '../../../fixtures/protocolHarness.fixture';
import {
  shardStormDiscardPrompt,
  shardStormMissHealPrompt,
  shardStormMissTargetPrompt,
  shardStormScenario,
  ALLY_PLAYER_ID,
} from '../../../scenarios/holyBow';

test.describe('holy bow shard storm protocol harness', () => {
  test('discard phase: pick same-element attack cards from hand', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(shardStormScenario());

    // 后端 hb_holy_shard_combo 直接渲染手牌多选，玩家从手牌区点两张同系攻击牌。
    await protocolHarness.pushServerMessage(shardStormDiscardPrompt());
    await page.getByTestId('hand-card-0').click();
    await page.getByTestId('hand-card-1').click();
    await page.getByTestId('prompt-confirm-btn').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      card_ids: ['card_1', 'card_2'],
    });
  });

  test('shard storm miss follow-up: choose X, then ally', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(shardStormScenario());

    // 1) 弃牌组合阶段
    await protocolHarness.pushServerMessage(shardStormDiscardPrompt());
    await page.getByTestId('hand-card-0').click();
    await page.getByTestId('hand-card-1').click();
    await page.getByTestId('prompt-confirm-btn').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      card_ids: ['card_1', 'card_2'],
    });

    // 2) miss_x：选 X=1（无 X=0 选项，option_index 0 对应 X=1）
    await protocolHarness.pushServerMessage(shardStormMissHealPrompt(2));
    await page.getByTestId('numeric-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // 3) ally 选择
    await protocolHarness.pushServerMessage(shardStormMissTargetPrompt());
    await page.getByTestId(`player-area-${ALLY_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('shard storm miss: skip miss branch', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(shardStormScenario());

    await protocolHarness.pushServerMessage(shardStormDiscardPrompt());
    await page.getByTestId('hand-card-0').click();
    await page.getByTestId('hand-card-1').click();
    await page.getByTestId('prompt-confirm-btn').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      card_ids: ['card_1', 'card_2'],
    });

    await protocolHarness.pushServerMessage(shardStormMissHealPrompt(2));
    await page.getByTestId('prompt-cancel-btn').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Cancel',
    });
  });
});
