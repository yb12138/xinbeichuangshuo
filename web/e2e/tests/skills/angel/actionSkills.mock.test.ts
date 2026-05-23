import type { Page } from '@playwright/test';
import { test } from '../../../fixtures/protocolHarness.fixture';
import {
  ANGEL_BLESSING_ID,
  ANGEL_WIND_CLEANSE_ID,
  angelBlessingScenario,
  angelBlessingBranchPrompt,
  angelBlessingSingleTargetPrompt,
  angelBlessingDualTargetPrompt,
  windCleanseScenario,
  windCleanseFieldSelectPrompt,
} from '../../../scenarios/angel';

async function activatePanelSkill(page: Page, skillId: string) {
  await page.getByTestId('action-magic').click();
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

test.describe('angel blessing protocol harness', () => {
  test('angel blessing: activate skill', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(angelBlessingScenario());

    await activatePanelSkill(page, ANGEL_BLESSING_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: ANGEL_BLESSING_ID,
    });
  });

  test('angel blessing: branch 1 - single target gives 2 cards', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(angelBlessingScenario());

    // Server pushes branch selection
    await protocolHarness.pushServerMessage(angelBlessingBranchPrompt());

    // Select branch 1
    await clickOverlayOption(page, 'prompt-option-branch1');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Server pushes single target selection
    await protocolHarness.pushServerMessage(angelBlessingSingleTargetPrompt());

    // Select enemy as target
    await selectTarget(page, 'enemy_1');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('angel blessing: branch 2 - dual targets each give 1 card', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(angelBlessingScenario());

    // Server pushes branch selection
    await protocolHarness.pushServerMessage(angelBlessingBranchPrompt());

    // Select branch 2
    await clickOverlayOption(page, 'prompt-option-branch2');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });

    // Server pushes dual target selection (need to select 2)
    await protocolHarness.pushServerMessage(angelBlessingDualTargetPrompt());

    // Select both targets (order matters)
    await selectTarget(page, 'enemy_1');
    await selectTarget(page, 'ally_1');
    await page.getByTestId('prompt-confirm-btn').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0, 1],
    });
  });
});

test.describe('angel wind cleanse protocol harness', () => {
  test('wind cleanse: activate and select field effect', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(windCleanseScenario());

    await activatePanelSkill(page, ANGEL_WIND_CLEANSE_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: ANGEL_WIND_CLEANSE_ID,
    });

    // Server pushes field effect selection
    await protocolHarness.pushServerMessage(windCleanseFieldSelectPrompt());

    // Select enemy shield to remove
    await clickOverlayOption(page, 'prompt-option-enemy_shield');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('wind cleanse: select ally buff', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(windCleanseScenario());

    await activatePanelSkill(page, ANGEL_WIND_CLEANSE_ID);

    await protocolHarness.pushServerMessage(windCleanseFieldSelectPrompt());

    // Select ally buff to remove
    await clickOverlayOption(page, 'prompt-option-ally_buff');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});
