import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  paleMoonBranchPrompt,
  paleMoonXPrompt,
  paleMoonTargetPrompt,
  paleMoonDiscardPrompt,
  paleMoonScenario,
  ENEMY_PLAYER_ID,
} from '../../../scenarios/moonGoddess';

// ============================================================
// 苍白之月 (mg_pale_moon) - 后端通过 response_skills 触发
// 分支选择使用 choice_type: mg_pale_moon_mode
// X值选择使用 choice_type: mg_pale_moon_x
// 目标选择使用 choice_type: mg_pale_moon_target
// 弃牌使用 choice_type: mg_pale_moon_discard
// ============================================================

test.describe('moon goddess pale moon protocol harness', () => {
  // ============================================================
  // 分支①完整流程：移除3石化 → 不可应战 → 追加攻击 → 追加回合
  // ============================================================
  test('pale moon: branch 1 - remove 3 petrify for extra attack and turn', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(paleMoonScenario({ petrify_tokens: 3, new_moon_tokens: 0 }));

    // 分支选择（只有分支1可选）
    await protocolHarness.pushServerMessage(paleMoonBranchPrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // 分支①执行后：移除3石化，下次主动攻击不可应战，追加攻击行动，获得额外回合
    // 无后续选择，自动完成
  });

  // ============================================================
  // 分支②完整流程：X选择 → 目标选择 → 弃牌 → 造成伤害
  // ============================================================
  test('pale moon: branch 2 - complete flow', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(paleMoonScenario({ petrify_tokens: 0, new_moon_tokens: 2 }));

    // 分支选择（只有分支2可选）
    await protocolHarness.pushServerMessage(paleMoonBranchPrompt());
    await page.getByTestId('branch-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });

    // X值选择
    await protocolHarness.pushServerMessage(paleMoonXPrompt(2));
    await page.getByTestId('prompt-option-0').click(); // X=1
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // 目标选择
    await protocolHarness.pushServerMessage(paleMoonTargetPrompt());
    await page.getByTestId(`player-area-${ENEMY_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // 弃牌
    await protocolHarness.pushServerMessage(paleMoonDiscardPrompt());
    await page.getByTestId('hand-card-0').click();
    await page.getByTestId('prompt-confirm-btn').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // 执行后：移除X新月，石化+1，弃牌，造成(X+1)点法术伤害
  });

  test('pale moon: branch selection with both branches available', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(paleMoonScenario({ petrify_tokens: 3, new_moon_tokens: 2 }));

    // 分支选择使用后端定义的 choice_type
    await protocolHarness.pushServerMessage(paleMoonBranchPrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('pale moon: branch 2 - select X=2 value', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(paleMoonScenario({ petrify_tokens: 0, new_moon_tokens: 2 }));

    // 分支选择
    await protocolHarness.pushServerMessage(paleMoonBranchPrompt());
    await page.getByTestId('branch-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });

    // X=2 选择（第二项，index 1）
    await protocolHarness.pushServerMessage(paleMoonXPrompt(2));
    await page.getByTestId('prompt-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });

  test('pale moon: triggered via response_skills', async ({ protocolHarness }) => {
    await protocolHarness.bootGame(paleMoonScenario({ petrify_tokens: 3, new_moon_tokens: 2 }));

    // 后端会设置 response_skills 触发确认弹框
    await protocolHarness.expectSubmitAction({
      action_type: 'UseSkill',
      skill_id: 'mg_pale_moon',
    });
  });
});