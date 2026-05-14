import type { Page } from '@playwright/test';
import { test } from '../../../fixtures/protocolHarness.fixture';
import {
  elementalShotScenario,
  elementalShotDiscardPrompt,
  animalCompanionScenario,
  animalCompanionPrompt,
  petEnhanceScenario,
  petEnhanceBranchPrompt,
} from '../../../scenarios/elfArcher';

async function clickOverlayOption(page: Page, selector: string) {
  const overlay = page.getByTestId('decision-overlay');
  const overlayVisible = await overlay.isVisible({ timeout: 1000 }).catch(() => false);
  if (overlayVisible) {
    await overlay.getByTestId(selector).click();
  } else {
    await page.getByTestId('prompt-dialog').getByTestId(selector).click();
  }
}

// ============================================================
// Elemental Shot (元素射击) - 后端通过 response_skills 触发
// 弃牌使用 choice_type: elf_archer_elemental_shot_pick
// ============================================================

test.describe('elf archer elemental shot protocol harness', () => {
  test('elemental shot: fire element (+1 damage)', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(elementalShotScenario());

    // 弃牌阶段使用后端定义的 choice_type
    await protocolHarness.pushServerMessage(elementalShotDiscardPrompt());

    // Select fire magic for +1 damage
    await clickOverlayOption(page, 'prompt-option-elf-fire-magic');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('elemental shot: thunder element (forced hit)', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(elementalShotScenario());

    await protocolHarness.pushServerMessage(elementalShotDiscardPrompt());

    // Select thunder magic for forced hit
    await clickOverlayOption(page, 'prompt-option-elf-thunder-magic');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [4],
    });
  });

  test('elemental shot: triggered via response_skills', async ({ protocolHarness }) => {
    await protocolHarness.bootGame(elementalShotScenario());

    // 检查 response_skills 触发的弹框显示
    // 用户点击发动按钮后，前端弹出弃牌选择
    await protocolHarness.expectSubmitAction({
      action_type: 'UseSkill',
      skill_id: 'elf_archer_elemental_shot',
    });
  });
});

// ============================================================
// Animal Companion (动物伙伴) - 后端通过 response_skills 触发
// 使用 choice_type: elf_animal_companion_confirm
// ============================================================

test.describe('elf archer animal companion protocol harness', () => {
  test('animal companion: confirm prompt', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(animalCompanionScenario());

    // 确认弹框使用后端定义的 choice_type
    await protocolHarness.pushServerMessage(animalCompanionPrompt());

    // Click confirm button
    await clickOverlayOption(page, 'prompt-option-confirm');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('animal companion: skip', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(animalCompanionScenario());

    await protocolHarness.pushServerMessage(animalCompanionPrompt());

    // Click skip button
    await clickOverlayOption(page, 'prompt-option-skip');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});

// ============================================================
// Pet Enhance (宠物强化) - 后端通过 response_skills 触发
// 分支选择使用 choice_type: elf_pet_empower_confirm
// ============================================================

test.describe('elf archer pet enhance protocol harness', () => {
  test('pet enhance: draw +1 upgrade', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(petEnhanceScenario());

    // 分支选择使用后端定义的 choice_type
    await protocolHarness.pushServerMessage(petEnhanceBranchPrompt());

    // Select draw +1 option
    await clickOverlayOption(page, 'prompt-option-draw_plus');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('pet enhance: discard -1 upgrade', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(petEnhanceScenario());

    await protocolHarness.pushServerMessage(petEnhanceBranchPrompt());

    // Select discard -1 option
    await clickOverlayOption(page, 'prompt-option-discard_minus');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });

  test('pet enhance: target discard upgrade', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(petEnhanceScenario());

    await protocolHarness.pushServerMessage(petEnhanceBranchPrompt());

    // Select target discard option
    await clickOverlayOption(page, 'prompt-option-target_discard');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [2],
    });

    // 目标选择通过 min_targets 处理，后端不发送单独的 prompt
  });

  test('pet enhance: triggered via response_skills', async ({ protocolHarness }) => {
    await protocolHarness.bootGame(petEnhanceScenario());

    // 检查 response_skills 触发的弹框显示
    await protocolHarness.expectSubmitAction({
      action_type: 'UseSkill',
      skill_id: 'elf_archer_pet_enhance',
    });
  });
});