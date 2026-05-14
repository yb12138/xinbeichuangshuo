import type { Page } from '@playwright/test';
import { test } from '../../../fixtures/protocolHarness.fixture';
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

async function selectTarget(page: Page, targetId: string) {
  await page.getByTestId(`player-area-${targetId}`).click();
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

  test('healing light: select single target', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(healingLightScenario());

    await activatePanelSkill(page, SAINTESS_HEALING_LIGHT_ID);

    // Server pushes multi target selection
    await protocolHarness.pushServerMessage(healingLightMultiTargetPrompt());

    // Select single target
    await selectTarget(page, 'ally_1');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [2], // ally is third option
    });
  });

  test('healing light: select 2 targets', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(healingLightScenario());

    await activatePanelSkill(page, SAINTESS_HEALING_LIGHT_ID);

    await protocolHarness.pushServerMessage(healingLightMultiTargetPrompt());

    // Select 2 targets
    await selectTarget(page, 'saintess_player');
    await selectTarget(page, 'ally_1');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0, 2],
    });
  });

  test('healing light: select 3 targets', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(healingLightScenario());

    await activatePanelSkill(page, SAINTESS_HEALING_LIGHT_ID);

    await protocolHarness.pushServerMessage(healingLightMultiTargetPrompt());

    // Select all 3 targets
    await selectTarget(page, 'saintess_player');
    await selectTarget(page, 'enemy_1');
    await selectTarget(page, 'ally_1');
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

    // Server pushes target selection
    await protocolHarness.pushServerMessage(healSkillTargetPrompt());

    // Select ally target
    await selectTarget(page, 'ally_1');
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

  test('holy heal: select branch after distribute', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(holyHealScenario());

    await activatePanelSkill(page, SAINTESS_HOLY_HEAL_ID);

    // Server pushes branch selection after distribution
    await protocolHarness.pushServerMessage(holyHealBranchPrompt());

    // Select attack action
    await clickOverlayOption(page, 'prompt-option-attack');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('holy heal: select magic action', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(holyHealScenario());

    await activatePanelSkill(page, SAINTESS_HOLY_HEAL_ID);

    await protocolHarness.pushServerMessage(holyHealBranchPrompt());

    // Select magic action
    await clickOverlayOption(page, 'prompt-option-magic');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});