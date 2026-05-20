import type { Page } from '@playwright/test';
import { test } from '../../../fixtures/protocolHarness.fixture';
import {
  ADVENTURER_STEAL_SKY_CHANGE_DAY_ID,
  stealSkyChangeDayScenario,
  stealSkyChangeDayBranchPrompt,
} from '../../../scenarios/adventurer';

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

// For branch_select prompts with long labels, frontend renders testid as branch-option-${idx}
async function clickBranchOption(page: Page, branchIdx: number) {
  const overlay = page.getByTestId('decision-overlay');
  const overlayVisible = await overlay.isVisible({ timeout: 1000 }).catch(() => false);
  if (overlayVisible) {
    await overlay.getByTestId(`branch-option-${branchIdx}`).click();
  } else {
    await page.getByTestId('prompt-dialog').getByTestId(`branch-option-${branchIdx}`).click();
  }
}

async function selectTarget(page: Page, targetId: string) {
  await page.getByTestId(`player-area-${targetId}`).click();
}

test.describe('adventurer steal sky change day protocol harness', () => {
  test('steal sky change day: gem transfer branch', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(stealSkyChangeDayScenario());

    await activatePanelSkill(page, ADVENTURER_STEAL_SKY_CHANGE_DAY_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: ADVENTURER_STEAL_SKY_CHANGE_DAY_ID,
    });

    // Server pushes branch selection
    await protocolHarness.pushServerMessage(stealSkyChangeDayBranchPrompt());

    // Select gem transfer option (transfer enemy gem to ally)
    await clickBranchOption(page, 0);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('steal sky change day: crystal convert branch', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(stealSkyChangeDayScenario());

    await activatePanelSkill(page, ADVENTURER_STEAL_SKY_CHANGE_DAY_ID);

    await protocolHarness.pushServerMessage(stealSkyChangeDayBranchPrompt());

    // Select crystal convert option (convert all ally crystals to gems)
    await clickBranchOption(page, 1);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});