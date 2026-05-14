import type { Page } from '@playwright/test';
import { test } from '../../../fixtures/protocolHarness.fixture';
import {
  ADVENTURER_STEAL_SKY_CHANGE_DAY_ID,
  ENEMY_PLAYER_ID,
  ALLY_PLAYER_ID,
  stealSkyChangeDayScenario,
  stealSkyChangeDayBranchPrompt,
  stealSkyChangeDayEnergySourcePrompt,
  stealSkyChangeDayEnergyTargetPrompt,
  stealSkyChangeDayCardSwapTargetPrompt,
  stealSkyChangeDayMyCardPrompt,
} from '../../../scenarios/adventurer';

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

test.describe('adventurer steal sky change day protocol harness', () => {
  test('steal sky change day: energy transfer branch', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(stealSkyChangeDayScenario());

    await activatePanelSkill(page, ADVENTURER_STEAL_SKY_CHANGE_DAY_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: ADVENTURER_STEAL_SKY_CHANGE_DAY_ID,
    });

    // Server pushes branch selection
    await protocolHarness.pushServerMessage(stealSkyChangeDayBranchPrompt());

    // Select energy transfer option
    await clickOverlayOption(page, 'prompt-option-energy_transfer');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Server pushes energy source (enemy) selection
    await protocolHarness.pushServerMessage(stealSkyChangeDayEnergySourcePrompt());

    // Select enemy target
    await selectTarget(page, ENEMY_PLAYER_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Server pushes energy target (ally) selection
    await protocolHarness.pushServerMessage(stealSkyChangeDayEnergyTargetPrompt());

    // Select ally target
    await selectTarget(page, ALLY_PLAYER_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('steal sky change day: card swap branch', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(stealSkyChangeDayScenario());

    await activatePanelSkill(page, ADVENTURER_STEAL_SKY_CHANGE_DAY_ID);

    await protocolHarness.pushServerMessage(stealSkyChangeDayBranchPrompt());

    // Select card swap option
    await clickOverlayOption(page, 'prompt-option-card_swap');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });

    // Server pushes card swap target selection
    await protocolHarness.pushServerMessage(stealSkyChangeDayCardSwapTargetPrompt());

    // Select ally target
    await selectTarget(page, ALLY_PLAYER_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Server pushes my card selection
    await protocolHarness.pushServerMessage(stealSkyChangeDayMyCardPrompt());

    // Select card to give
    await clickOverlayOption(page, 'prompt-option-adv-attack-1');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });
});