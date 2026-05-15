import type { Page } from '@playwright/test';
import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  ENEMY_2_PLAYER_ID,
  ENEMY_PLAYER_ID,
  MB_ATTACK_CARD_ID,
  MB_CHARGE_SKILL_ID,
  MB_DEMON_EYE_SKILL_ID,
  MB_MAGIC_PIERCE_SKILL_ID,
  MB_MULTI_SHOT_SKILL_ID,
  MB_THUNDER_SCATTER_SKILL_ID,
  chargeDiscardPrompt,
  chargeDrawPrompt,
  chargePlaceCardsMultiSelectPrompt,
  chargeScenario,
  demonEyeChargeCardPrompt,
  demonEyeModePrompt,
  demonEyeScenario,
  demonEyeTargetPrompt,
  magicBowScenario,
  magicPierceChargePrompt,
  magicPierceHitBonusPrompt,
  magicPierceHitChargePrompt,
  magicPierceResponsePrompt,
  multiShotChargePrompt,
  multiShotResponsePrompt,
  multiShotTargetPrompt,
  thunderScatterBaseChargePrompt,
  thunderScatterExtraPrompt,
  thunderScatterScenario,
  thunderScatterTargetPrompt,
} from '../../../scenarios/magicBow';

async function selectHandCards(page: Page, indices: number[]) {
  for (const index of indices) {
    const card = page.getByTestId(`hand-card-${index}`);
    await card.scrollIntoViewIfNeeded();
    await card.click();
  }
  await page.getByTestId('prompt-confirm-btn').click();
}

async function activatePanelSkill(page: Page, skillId: string) {
  await page.getByTestId('action-skill').click();
  await page.getByTestId(`skill-${skillId}`).click();
}

test.describe('magic bow protocol harness', () => {
  test('magic pierce: attack -> response confirm -> remove fire charge -> accept hit bonus -> remove fire charge', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(magicBowScenario());

    await page.getByTestId('action-attack').click();
    await page.getByTestId('hand-card-0').click();
    await page.getByTestId(`player-area-${ENEMY_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Attack',
      targets: [{ target_user_id: ENEMY_PLAYER_ID }],
      used_card_uuids: [MB_ATTACK_CARD_ID],
    });

    await protocolHarness.pushServerMessage(magicPierceResponsePrompt());
    await expect(page.getByTestId('skill-branch-overlay')).toBeVisible();
    await page.getByTestId('skill-branch-overlay').getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    await protocolHarness.pushServerMessage(magicPierceChargePrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('decision-overlay').getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    await protocolHarness.pushServerMessage(magicPierceHitBonusPrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('decision-overlay').getByTestId('prompt-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    await protocolHarness.pushServerMessage(magicPierceHitChargePrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('decision-overlay').getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('magic pierce: hit bonus can be declined without entering charge selection', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(magicBowScenario());

    await protocolHarness.pushServerMessage(magicPierceHitBonusPrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('decision-overlay').getByTestId('prompt-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });

  test('thunder scatter: active skill first submits skill, then base charge, extra X, target', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(thunderScatterScenario());

    await activatePanelSkill(page, MB_THUNDER_SCATTER_SKILL_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: MB_THUNDER_SCATTER_SKILL_ID,
    });

    await protocolHarness.pushServerMessage(thunderScatterBaseChargePrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('decision-overlay').getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    await protocolHarness.pushServerMessage(thunderScatterExtraPrompt(2));
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('decision-overlay').getByTestId('numeric-option-2').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [2],
    });

    await protocolHarness.pushServerMessage(thunderScatterTargetPrompt(2));
    await expect(page.getByTestId('decision-overlay')).not.toBeVisible({ timeout: 5000 });
    await page.getByTestId(`player-area-${ENEMY_2_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });

  test('thunder scatter: extra X=0 submits numeric choice and does not require a target click', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(thunderScatterScenario());

    await protocolHarness.pushServerMessage(thunderScatterExtraPrompt(2));
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('decision-overlay').getByTestId('numeric-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('multi shot: response confirm -> remove wind charge -> select alternate target only', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(magicBowScenario());

    await protocolHarness.pushServerMessage(multiShotResponsePrompt());
    await expect(page.getByTestId('skill-branch-overlay')).toBeVisible();
    await page.getByTestId('skill-branch-overlay').getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    await protocolHarness.pushServerMessage(multiShotChargePrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('decision-overlay').getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    await protocolHarness.pushServerMessage(multiShotTargetPrompt());
    await expect(page.getByTestId('decision-overlay')).not.toBeVisible({ timeout: 5000 });
    await page.getByTestId(`player-area-${ENEMY_2_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('charge: discard to four -> draw X -> multi-select charge cards (new simplified flow)', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(chargeScenario());

    await activatePanelSkill(page, MB_CHARGE_SKILL_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: MB_CHARGE_SKILL_ID,
    });

    await protocolHarness.pushServerMessage(chargeDiscardPrompt());
    await selectHandCards(page, [4, 5]);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [4, 5],
    });

    await protocolHarness.pushServerMessage(chargeDrawPrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('decision-overlay').getByTestId('numeric-option-3').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [3],
    });

    // 新流程：直接进入盖牌多选（跳过数量选择步骤）
    // 玩家一次性选择 0~maxPlace 张手牌作为充能
    await protocolHarness.pushServerMessage(chargePlaceCardsMultiSelectPrompt(3));
    await selectHandCards(page, [0, 1]); // 选择2张作为充能
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0, 1],
    });
  });

  test('demon eye branch 1: select discard target then charge one hand card', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(demonEyeScenario());

    await activatePanelSkill(page, MB_DEMON_EYE_SKILL_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: MB_DEMON_EYE_SKILL_ID,
    });

    await protocolHarness.pushServerMessage(demonEyeModePrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('decision-overlay').getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    await protocolHarness.pushServerMessage(demonEyeTargetPrompt());
    await expect(page.getByTestId('decision-overlay')).not.toBeVisible({ timeout: 5000 });
    await page.getByTestId(`player-area-${ENEMY_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });

    await protocolHarness.pushServerMessage(demonEyeChargeCardPrompt());
    await selectHandCards(page, [0]);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('demon eye branch 2: draw branch proceeds directly to charge card selection', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(demonEyeScenario());

    await protocolHarness.pushServerMessage(demonEyeModePrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('decision-overlay').getByTestId('branch-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });

    await protocolHarness.pushServerMessage(demonEyeChargeCardPrompt());
    await selectHandCards(page, [0]);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });
});
