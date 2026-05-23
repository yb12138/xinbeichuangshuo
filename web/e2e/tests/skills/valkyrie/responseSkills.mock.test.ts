import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  ENEMY_PLAYER_ID,
  ALLY_PLAYER_ID,
  holyPursuitScenario,
  holyPursuitPrompt,
  peaceWalkerScenario,
  peaceWalkerPrompt,
  martialGodLightScenario,
  martialGodLightBranchPrompt,
  martialGodLightTargetPrompt,
  heroicSummonScenario,
  heroicSummonSkillPrompt,
  heroicSummonDiscardPrompt,
} from '../../../scenarios/valkyrie';

async function selectTarget(page: import('@playwright/test').Page, targetId: string) {
  await page.getByTestId(`player-area-${targetId}`).click();
}

// ============================================================
// Holy Pursuit (神圣追击) - 后端通过 RequireAction + skill_choice 触发
// ============================================================

test.describe('valkyrie holy pursuit protocol harness', () => {
  test('holy pursuit: confirm skill', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(holyPursuitScenario());

    // Server pushes skill_choice prompt
    await protocolHarness.pushServerMessage(holyPursuitPrompt());

    // Click confirm (skill_choice with 1 skill: prompt-option-{skillId})
    await page.getByTestId('prompt-option-valkyrie_holy_pursuit').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('holy pursuit: skip skill', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(holyPursuitScenario());

    await protocolHarness.pushServerMessage(holyPursuitPrompt());

    // Click skip (skill_choice: prompt-option-skip)
    await page.getByTestId('prompt-option-skip').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});

// ============================================================
// Peace Walker (和平行者) - 后端通过 RequireAction + skill_choice 触发
// ============================================================

test.describe('valkyrie peace walker protocol harness', () => {
  test('peace walker: confirm skill', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(peaceWalkerScenario());

    // Server pushes skill_choice prompt
    await protocolHarness.pushServerMessage(peaceWalkerPrompt());

    // Click confirm (skill_choice with 1 skill: prompt-option-{skillId})
    await page.getByTestId('prompt-option-valkyrie_peace_walker').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('peace walker: skip skill', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(peaceWalkerScenario());

    await protocolHarness.pushServerMessage(peaceWalkerPrompt());

    // Click skip (skill_choice: prompt-option-skip)
    await page.getByTestId('prompt-option-skip').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});

// ============================================================
// Martial God Light (军威神光) - 使用后端定义的 choice_type
// ============================================================

test.describe('valkyrie martial god light protocol harness', () => {
  test('martial god light: select draw 2 cards', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(martialGodLightScenario());

    // Server pushes martial god light branch prompt at turn start (with maxX=2)
    await protocolHarness.pushServerMessage(martialGodLightBranchPrompt(2));

    // Select draw option (branch_select: branch-option-0)
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('decision-overlay').getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('martial god light: select damage to enemy', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(martialGodLightScenario());

    // Server pushes martial god light branch prompt at turn start (with maxX=2)
    await protocolHarness.pushServerMessage(martialGodLightBranchPrompt(2));

    // Select damage option (branch_select: branch-option-1)
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('decision-overlay').getByTestId('branch-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });

    // Server pushes target selection
    await protocolHarness.pushServerMessage(martialGodLightTargetPrompt());

    // Select enemy target (target_picker: click player area)
    await expect(page.getByTestId('decision-overlay')).not.toBeVisible({ timeout: 5000 });
    await selectTarget(page, ENEMY_PLAYER_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});

// ============================================================
// Heroic Summon (英灵召唤) - 后端通过 RequireAction + skill_choice 触发
// 弃牌使用 choice_type: valkyrie_heroic_discard_card
// ============================================================

test.describe('valkyrie heroic summon protocol harness', () => {
  test('heroic summon: confirm skill', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(heroicSummonScenario());

    // Server pushes skill_choice prompt
    await protocolHarness.pushServerMessage(heroicSummonSkillPrompt());

    // Click confirm (skill_choice with 1 skill: prompt-option-{skillId})
    await page.getByTestId('prompt-option-valkyrie_heroic_summon').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('heroic summon: skip skill', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(heroicSummonScenario());

    await protocolHarness.pushServerMessage(heroicSummonSkillPrompt());

    // Click skip (skill_choice: prompt-option-skip)
    await page.getByTestId('prompt-option-skip').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });

  test('heroic summon: discard card prompt', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(heroicSummonScenario());

    // 弃牌阶段使用后端定义的 choice_type
    await protocolHarness.pushServerMessage(heroicSummonDiscardPrompt());

    // Select magic card from hand (card_picker: hand-card-2 auto-submits for min=1,max=1)
    await page.getByTestId('hand-card-2').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      card_ids: ['valk-magic-1'],
    });
  });
});