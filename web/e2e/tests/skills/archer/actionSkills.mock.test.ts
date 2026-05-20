import type { Page } from '@playwright/test';
import { test } from '../../../fixtures/protocolHarness.fixture';
import {
  ARCHER_SNipe_ID,
  snipeScenario,
  snipeTargetPrompt,
  flashTrapScenario,
  flashTrapTargetPrompt,
} from '../../../scenarios/archer';

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

test.describe('archer snipe protocol harness', () => {
  test('snipe: activate skill', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(snipeScenario());

    await activatePanelSkill(page, ARCHER_SNipe_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: ARCHER_SNipe_ID,
    });
  });

  test('snipe: select enemy target', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(snipeScenario());

    await activatePanelSkill(page, ARCHER_SNipe_ID);

    // Server pushes target selection
    await protocolHarness.pushServerMessage(snipeTargetPrompt());

    // Select enemy target
    await selectTarget(page, 'enemy_1');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('snipe: select ally target', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(snipeScenario());

    await activatePanelSkill(page, ARCHER_SNipe_ID);

    await protocolHarness.pushServerMessage(snipeTargetPrompt());

    // Select ally target
    await selectTarget(page, 'ally_1');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});

test.describe('archer flash trap protocol harness', () => {
  test('flash trap: activate and select target', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(flashTrapScenario());

    // Server pushes flash trap target prompt
    await protocolHarness.pushServerMessage(flashTrapTargetPrompt());

    // Select enemy target
    await selectTarget(page, 'enemy_1');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('flash trap: select ally target', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(flashTrapScenario());

    await protocolHarness.pushServerMessage(flashTrapTargetPrompt());

    // Select ally target
    await selectTarget(page, 'ally_1');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});
