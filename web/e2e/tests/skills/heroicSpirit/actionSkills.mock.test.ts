import type { Page } from '@playwright/test';
import { test } from '../../../fixtures/protocolHarness.fixture';
import {
  HEROIC_SPIRIT_RUNE_MODIFICATION_ID,
  ALLY_PLAYER_ID,
  runeModificationScenario,
  runeModificationSealAdjustPrompt,
  doubleEchoScenario,
  doubleEchoTargetPrompt,
} from '../../../scenarios/heroicSpirit';

async function activatePanelSkill(page: Page, skillId: string) {
  await page.getByTestId('action-hub-trigger').click();
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

test.describe('heroic spirit rune modification protocol harness', () => {
  test('rune modification: activate skill with gem', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(runeModificationScenario());

    await activatePanelSkill(page, HEROIC_SPIRIT_RUNE_MODIFICATION_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: HEROIC_SPIRIT_RUNE_MODIFICATION_ID,
    });

    // Server pushes seal adjust selection
    await protocolHarness.pushServerMessage(runeModificationSealAdjustPrompt());

    // Adjust seals (flip one seal)
    await clickOverlayOption(page, 'branch-option-0');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('rune modification: choose another seal', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(runeModificationScenario());

    await activatePanelSkill(page, HEROIC_SPIRIT_RUNE_MODIFICATION_ID);

    await protocolHarness.pushServerMessage(runeModificationSealAdjustPrompt());

    // Choose the second seal option
    await clickOverlayOption(page, 'branch-option-1');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});

test.describe('heroic spirit double echo protocol harness', () => {
  test('double echo: select another target after hit', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(doubleEchoScenario());

    // Server pushes double echo target selection after hit
    await protocolHarness.pushServerMessage(doubleEchoTargetPrompt());

    // Select another target (ally) for equal damage
    await selectTarget(page, ALLY_PLAYER_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });
});
