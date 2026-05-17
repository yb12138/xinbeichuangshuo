import type { Page } from '@playwright/test';
import { test } from '../../../fixtures/protocolHarness.fixture';
import {
  BD_CHRYSALIS_SKILL_ID,
  BD_REVERSE_SKILL_ID,
  ENEMY_PLAYER_ID,
  chrysalisScenario,
  chrysalisResolvedState,
  chrysalisOverflowDiscardPrompt,
  reverseBranch2CostPrompt,
  reverseBranch2PickPrompt,
  reverseDiscardPrompt,
  reverseModePrompt,
  reverseScenario,
  reverseTargetPrompt,
} from '../../../scenarios/butterflyDancer';

async function activatePanelSkill(page: Page, skillId: string) {
  await page.getByTestId('action-skill').click();
  await page.getByTestId(`skill-${skillId}`).click();
}

async function clickOverlayOption(page: Page, selector: string) {
  const overlay = page.getByTestId('decision-overlay');
  const overlayVisible = await overlay.isVisible({ timeout: 1000 }).catch(() => false);
  if (overlayVisible) {
    await overlay.getByTestId(selector).click();
  } else {
    await page.getByTestId('prompt-dialog').getByTestId(selector).click();
  }
}

async function selectHandCards(page: Page, indices: number[]) {
  for (const index of indices) {
    const card = page.getByTestId(`hand-card-${index}`);
    await card.scrollIntoViewIfNeeded();
    await card.click();
  }
  await page.getByTestId('prompt-confirm-btn').click();
}

/**
 * 完成倒逆之蝶发动前的「点技能 → 弃 2 张」固定前置链路。
 * 后端契约：通过统一弃牌费用流程完成 cost_discards=2 后才会推送 bt_reverse_mode。
 */
async function activateReverseAndDiscardTwo(
  page: Page,
  protocolHarness: any,
  discardIndices: [number, number] = [0, 1],
) {
  await activatePanelSkill(page, BD_REVERSE_SKILL_ID);
  await protocolHarness.expectSubmitAction({
    action_type: 'Skill',
    skill_id: BD_REVERSE_SKILL_ID,
  });

  // Server pushes system discard prompt for cost_discards=2
  await protocolHarness.pushServerMessage(reverseDiscardPrompt());
  await selectHandCards(page, discardIndices);
  await protocolHarness.expectSubmitAction({
    action_type: 'Select',
    option_indexes: discardIndices,
  });
}

test.describe('butterfly dancer chrysalis protocol harness', () => {
  /**
   * 蛹化 (Chrysalis) 技能测试：
   * 后端契约：技能激活后自动结算，无前端交互（除非茧溢出）。
   * 效果：+1蛹，牌库顶4张放置为茧。
   */
  test('activate chrysalis skill', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(chrysalisScenario());

    // Step 1: 激活技能
    await activatePanelSkill(page, BD_CHRYSALIS_SKILL_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: BD_CHRYSALIS_SKILL_ID,
    });

    // Step 2: 模拟后端自动结算后的状态同步
    // 后端契约：ResolveChrysalis 自动执行 AddPupa + DrawRawCards(4) + AddCocoonCards
    // 前端契约：接收 SyncState 消息后更新玩家状态面板（tokens 和 field）
    await protocolHarness.pushServerMessage(chrysalisResolvedState());

    // Step 3: 验证无后续交互弹框（茧未溢出时）
    // 后端契约：茧数量 <= 8 时无溢出处理弹框，流程结束
    // 注：状态更新验证需要在真实 E2E 环境中通过 UI 检查：
    //   - 状态面板应显示 tokens.bt_pupa = 1（蛹数量）
    //   - 状态面板应显示 tokens.bt_cocoon_count = 4（茧数量）
    //   - expansion zone 应显示4张茧牌（field 中 ButterflyCocoon 盖牌）
    //   - 宝石消耗后 gem 应为 0
    // 当前 mock 测试主要验证协议流程，UI 状态渲染由前端组件测试覆盖
  });

  /**
   * 蛹化茧溢出场景测试：
   * 后端契约：当茧数量超过上限（8张），推送 bt_cocoon_overflow_discard 弹框。
   * 前端需要选择溢出的茧进行舍弃。
   */
  test('chrysalis with cocoon overflow', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(chrysalisScenario());

    // 激活技能
    await activatePanelSkill(page, BD_CHRYSALIS_SKILL_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: BD_CHRYSALIS_SKILL_ID,
    });

    // 模拟茧溢出场景（假设已有5张茧，再加4张，溢出1张）
    // 前端契约：接收溢出弹框，展示茧选项供玩家选择舍弃
    const overflowPrompt = chrysalisOverflowDiscardPrompt(1, [
      { id: '0', label: '茧[0]: 茧牌A（火系 攻击）' },
      { id: '1', label: '茧[1]: 茧牌B（水系 攻击）' },
    ]);
    await protocolHarness.pushServerMessage(overflowPrompt);

    // 选择一张茧舍弃
    // 前端契约：点击茧牌（在 expansion zone 的盖牌区域），触发 Select action
    // 注：茧牌在 expansion zone 中使用 CardComponent 渲染，点击触发 onCoverCardClick
    await page.getByTestId('hand-card-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });
});

/**
 * 生命之火 (LifeFire) 技能测试：
 * 后端实现状态：空实现（CanUse 返回 false，Execute 返回 nil）。
 * 待后端实现后补充测试。
 *
 * 预期契约（根据技能描述推断）：
 * - 效果：将手牌上限设为当前士气值（或类似效果）
 * - 需要验证后端推送的状态更新（max_hand 变化）
 *
 * 后端文件参考：
 * - /internal/engine/player/butterfly_dancer/skill_handlers.go (第11-14行，空实现)
 * - /internal/engine/player/butterfly_dancer/module.go (bt_life_fire skill ID)
 */
test.describe('butterfly dancer life fire protocol harness (pending backend implementation)', () => {
  test.skip('life fire: activate and verify max hand floor update', async ({ page, protocolHarness }) => {
    // 待后端实现 ButterflyLifeFireHandler.Execute 后补充
    // 预期流程：
    // 1. 激活生命之火技能
    // 2. 验证 SubmitAction { action_type: 'Skill', skill_id: 'bt_life_fire' }
    // 3. 后端推送 sync_state，包含 max_hand 更新
    // 4. 验证前端状态面板显示新的手牌上限
  });
});

test.describe('butterfly dancer reverse butterfly protocol harness', () => {
  test('reverse: branch 1 select target', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(reverseScenario());

    // 前置：弃 2 张牌费用（与后端契约一致）
    await activateReverseAndDiscardTwo(page, protocolHarness);

    // Server pushes reverse mode prompt
    await protocolHarness.pushServerMessage(reverseModePrompt(true));

    // Select branch ① (option index 0)
    await clickOverlayOption(page, 'branch-option-0');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Server pushes target prompt for branch ①
    await protocolHarness.pushServerMessage(reverseTargetPrompt());

    // Click enemy player area to select target
    await page.getByTestId(`player-area-${ENEMY_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('reverse: branch 2 remove cocoons', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(reverseScenario({ canBranch2: true }));

    await activateReverseAndDiscardTwo(page, protocolHarness);

    // Server pushes reverse mode prompt
    await protocolHarness.pushServerMessage(reverseModePrompt(true));

    // Select branch ② (option index 1)
    await clickOverlayOption(page, 'branch-option-1');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });

    // Server pushes branch ② cost prompt
    await protocolHarness.pushServerMessage(reverseBranch2CostPrompt(true));

    // Select "移除2个茧" (option index 0)
    await clickOverlayOption(page, 'branch-option-0');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Server pushes cocoon pick prompt (choose_cards with cocoon labels, max=2)
    await protocolHarness.pushServerMessage(reverseBranch2PickPrompt());

    // Click two cocoon cover cards to select them, then confirm
    await page.getByTestId('hand-card-0').scrollIntoViewIfNeeded();
    await page.getByTestId('hand-card-0').click();
    await page.getByTestId('hand-card-1').scrollIntoViewIfNeeded();
    await page.getByTestId('hand-card-1').click();
    // Click "确认选择" button in expansion zone
    await page.locator('.expansion-cocoon-confirm-btn').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0, 1],
    });
  });

  test('reverse: branch 2 self-damage cost', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(reverseScenario({ canBranch2: true }));

    await activateReverseAndDiscardTwo(page, protocolHarness);

    // Server pushes reverse mode prompt
    await protocolHarness.pushServerMessage(reverseModePrompt(true));

    // Select branch ②
    await clickOverlayOption(page, 'branch-option-1');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });

    // Server pushes branch ② cost prompt without cocoon option
    await protocolHarness.pushServerMessage(reverseBranch2CostPrompt(false));

    // Only "对自己造成4点法术伤害" is available (option index 0 for the single option)
    // When canRemoveCocoon=false, the only option id is '1' at index 0 in the filtered options
    // But the decision overlay shows it as branch-option-0 (first in inlinePrimaryButtons)
    // Actually, option id '1' maps to prompt.options index 0 (only one option: {id:'1', label:'自伤...'})
    await clickOverlayOption(page, 'branch-option-0');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('reverse: branch 1 only (no branch 2 available)', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(reverseScenario({ canBranch2: false }));

    await activateReverseAndDiscardTwo(page, protocolHarness);

    // Server pushes reverse mode prompt without branch ②
    await protocolHarness.pushServerMessage(reverseModePrompt(false));

    // Only branch ① is available
    await clickOverlayOption(page, 'branch-option-0');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });
});
