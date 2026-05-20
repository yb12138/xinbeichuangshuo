import type { Page } from '@playwright/test';
import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  SAINTESS_HEALING_LIGHT_ID,
  SAINTESS_HEAL_SKILL_ID,
  SAINTESS_HOLY_HEAL_ID,
  healingLightScenario,
  healingLightMultiTargetPrompt,
  healSkillScenario,
  healSkillTargetPrompt,
  holyHealScenario,
  holyHealBranchPrompt,
} from '../../../scenarios/saintess';

async function activatePanelSkill(page: Page, skillId: string) {
  await page.getByTestId('action-magic').click();
  await page.getByTestId('action-skill').click();
  await page.getByTestId(`skill-${skillId}`).click();
}

test.describe('saintess healing light protocol harness', () => {
  test('healing light: activate skill', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(healingLightScenario());

    await activatePanelSkill(page, SAINTESS_HEALING_LIGHT_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: SAINTESS_HEALING_LIGHT_ID,
    });
  });

  test('healing light: select single target (multi_target with 1)', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(healingLightScenario());

    await activatePanelSkill(page, SAINTESS_HEALING_LIGHT_ID);

    // Server pushes multi target selection (multi_target=true)
    await protocolHarness.pushServerMessage(healingLightMultiTargetPrompt());

    // Click ally player area to toggle selection, then confirm
    await page.getByTestId('player-area-ally_1').click();
    await page.getByTestId('prompt-confirm-btn').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [2],
    });
  });

  test('healing light: select 2 targets', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(healingLightScenario());

    await activatePanelSkill(page, SAINTESS_HEALING_LIGHT_ID);

    await protocolHarness.pushServerMessage(healingLightMultiTargetPrompt());

    // Click 2 player areas to toggle, then confirm
    await page.getByTestId('player-area-saintess_player').click();
    await page.getByTestId('player-area-ally_1').click();
    await page.getByTestId('prompt-confirm-btn').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0, 2],
    });
  });

  test('healing light: select 3 targets', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(healingLightScenario());

    await activatePanelSkill(page, SAINTESS_HEALING_LIGHT_ID);

    await protocolHarness.pushServerMessage(healingLightMultiTargetPrompt());

    // Click all 3 player areas, then confirm
    await page.getByTestId('player-area-saintess_player').click();
    await page.getByTestId('player-area-enemy_1').click();
    await page.getByTestId('player-area-ally_1').click();
    await page.getByTestId('prompt-confirm-btn').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0, 1, 2],
    });
  });
});

test.describe('saintess heal skill protocol harness', () => {
  test('heal skill: activate and select target', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(healSkillScenario());

    await activatePanelSkill(page, SAINTESS_HEAL_SKILL_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: SAINTESS_HEAL_SKILL_ID,
    });

    // Server pushes target selection (target_picker, single target - click player area)
    await protocolHarness.pushServerMessage(healSkillTargetPrompt());

    // Click ally player area (auto-submits for single target)
    await page.getByTestId('player-area-ally_1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [2],
    });
  });
});

test.describe('saintess holy heal protocol harness', () => {
  test('holy heal: activate skill', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(holyHealScenario());

    await activatePanelSkill(page, SAINTESS_HOLY_HEAL_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: SAINTESS_HOLY_HEAL_ID,
    });
  });

  test('holy heal: select branch after distribute - attack action', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(holyHealScenario());

    await activatePanelSkill(page, SAINTESS_HOLY_HEAL_ID);

    // Server pushes branch selection (branch_select overlay)
    await protocolHarness.pushServerMessage(holyHealBranchPrompt());

    // Select attack action (branch-option-0)
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('decision-overlay').getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('holy heal: select branch - magic action', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(holyHealScenario());

    await activatePanelSkill(page, SAINTESS_HOLY_HEAL_ID);

    await protocolHarness.pushServerMessage(holyHealBranchPrompt());

    // Select magic action (branch-option-1)
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('decision-overlay').getByTestId('branch-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});