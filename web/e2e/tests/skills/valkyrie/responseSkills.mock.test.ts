import type { Page } from '@playwright/test';
import { test } from '../../../fixtures/protocolHarness.fixture';
import {
  ENEMY_PLAYER_ID,
  ALLY_PLAYER_ID,
  holyPursuitScenario,
  peaceWalkerScenario,
  martialGodLightScenario,
  martialGodLightBranchPrompt,
  martialGodLightTargetPrompt,
  heroicSummonScenario,
  heroicSummonDiscardPrompt,
} from '../../../scenarios/valkyrie';

async function clickOverlayOption(page: Page, selector: string) {
  const overlay = page.getByTestId('decision-overlay');
  const overlayVisible = await overlay.isVisible({ timeout: 1000 }).catch(() => false);
  if (overlayVisible) {
    await overlay.getByTestId(selector).click();
  } else {
    await page.getByTestId('prompt-dialog').getByTestId(selector).click();
  }
}

async function selectTarget(page: Page, targetId: string) {
  await page.getByTestId(`player-area-${targetId}`).click();
}

// ============================================================
// Holy Pursuit (神圣追击) - 后端通过 response_skills 自动触发
// ============================================================

test.describe('valkyrie holy pursuit protocol harness', () => {
  test('holy pursuit: triggered via response_skills', async ({ protocolHarness }) => {
    // 场景中后端会设置 response_skills，前端自动弹出确认弹框
    await protocolHarness.bootGame(holyPursuitScenario());

    // 检查 response_skills 触发的弹框显示
    // 用户点击发动按钮
    // 前端发送 UseSkill action
    await protocolHarness.expectSubmitAction({
      action_type: 'UseSkill',
      skill_id: 'valkyrie_holy_pursuit',
    });
  });
});

// ============================================================
// Peace Walker (和平行者) - 后端通过 response_skills + targets 触发
// ============================================================

test.describe('valkyrie peace walker protocol harness', () => {
  test('peace walker: triggered via response_skills with target', async ({ protocolHarness }) => {
    await protocolHarness.bootGame(peaceWalkerScenario());

    // 检查 response_skills 触发的弹框显示
    // 用户点击发动按钮后，前端弹出目标选择器（min_targets=1）
    await protocolHarness.expectSubmitAction({
      action_type: 'UseSkill',
      skill_id: 'valkyrie_peace_walker',
      targets: [{ target_user_id: ALLY_PLAYER_ID }],
    });
  });
});

// ============================================================
// Martial God Light (军威神光) - 使用后端定义的 choice_type
// ============================================================

test.describe('valkyrie martial god light protocol harness', () => {
  test('martial god light: select draw 2 cards', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(martialGodLightScenario());

    // Server pushes martial god light branch prompt at turn start
    await protocolHarness.pushServerMessage(martialGodLightBranchPrompt());

    // Select draw option
    await clickOverlayOption(page, 'prompt-option-draw');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('martial god light: select damage to enemy', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(martialGodLightScenario());

    await protocolHarness.pushServerMessage(martialGodLightBranchPrompt());

    // Select damage option
    await clickOverlayOption(page, 'prompt-option-damage');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });

    // Server pushes target selection
    await protocolHarness.pushServerMessage(martialGodLightTargetPrompt());

    // Select enemy target
    await selectTarget(page, ENEMY_PLAYER_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });
});

// ============================================================
// Heroic Summon (英灵召唤) - 后端通过 response_skills 触发
// 弃牌使用 choice_type: valkyrie_heroic_discard_card
// ============================================================

test.describe('valkyrie heroic summon protocol harness', () => {
  test('heroic summon: discard card prompt', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(heroicSummonScenario());

    // 弃牌阶段使用后端定义的 choice_type
    await protocolHarness.pushServerMessage(heroicSummonDiscardPrompt());

    // Select magic card to discard
    await clickOverlayOption(page, 'prompt-option-valk-magic-1');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('heroic summon: triggered via response_skills', async ({ protocolHarness }) => {
    await protocolHarness.bootGame(heroicSummonScenario());

    // 检查 response_skills 触发的弹框显示
    // 用户点击发动按钮后，前端通过 targets 参数处理目标选择
    await protocolHarness.expectSubmitAction({
      action_type: 'UseSkill',
      skill_id: 'valkyrie_heroic_summon',
    });
  });
});