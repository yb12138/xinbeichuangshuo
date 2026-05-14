import type { Page } from '@playwright/test';
import { test } from '../../../fixtures/protocolHarness.fixture';
import {
  PRAYER_GLORY_BELIEF_ID,
  PRAYER_DARK_CURSE_ID,
  PRAYER_POWER_BLESSING_ID,
  PRAYER_SWIFT_BLESSING_ID,
  PRAYER_PRAY_ID,
  ENEMY_PLAYER_ID,
  ALLY_PLAYER_ID,
  gloryBeliefScenario,
  gloryBeliefDiscardPrompt,
  gloryBeliefTargetPrompt,
  darkCurseScenario,
  darkCurseDiscardPrompt,
  darkCurseTargetPrompt,
  powerBlessingScenario,
  powerBlessingTargetPrompt,
  swiftBlessingScenario,
  swiftBlessingTargetPrompt,
  prayScenario,
} from '../../../scenarios/prayer';

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

test.describe('prayer glory belief protocol harness', () => {
  test('glory belief: activate in prayer form', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(gloryBeliefScenario());

    await activatePanelSkill(page, PRAYER_GLORY_BELIEF_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: PRAYER_GLORY_BELIEF_ID,
    });
  });

  test('glory belief: discard and select target', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(gloryBeliefScenario());

    await activatePanelSkill(page, PRAYER_GLORY_BELIEF_ID);

    // Server pushes discard selection
    await protocolHarness.pushServerMessage(gloryBeliefDiscardPrompt());

    // Select magic card to discard
    await clickOverlayOption(page, 'prompt-option-prayer-magic-1');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Server pushes target selection
    await protocolHarness.pushServerMessage(gloryBeliefTargetPrompt());

    // Select ally target
    await selectTarget(page, ALLY_PLAYER_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });
});

test.describe('prayer dark curse protocol harness', () => {
  test('dark curse: activate in prayer form', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(darkCurseScenario());

    await activatePanelSkill(page, PRAYER_DARK_CURSE_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: PRAYER_DARK_CURSE_ID,
    });
  });

  test('dark curse: discard and select target', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(darkCurseScenario());

    await activatePanelSkill(page, PRAYER_DARK_CURSE_ID);

    // Server pushes discard selection
    await protocolHarness.pushServerMessage(darkCurseDiscardPrompt());

    // Select card to discard
    await clickOverlayOption(page, 'prompt-option-prayer-magic-1');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Server pushes target selection
    await protocolHarness.pushServerMessage(darkCurseTargetPrompt());

    // Select enemy target
    await selectTarget(page, ENEMY_PLAYER_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });
});

test.describe('prayer power blessing protocol harness', () => {
  test('power blessing: activate and select target', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(powerBlessingScenario());

    await activatePanelSkill(page, PRAYER_POWER_BLESSING_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: PRAYER_POWER_BLESSING_ID,
    });

    // Server pushes target selection
    await protocolHarness.pushServerMessage(powerBlessingTargetPrompt());

    // Select ally target
    await selectTarget(page, ALLY_PLAYER_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });
});

test.describe('prayer swift blessing protocol harness', () => {
  test('swift blessing: activate and select target', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(swiftBlessingScenario());

    await activatePanelSkill(page, PRAYER_SWIFT_BLESSING_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: PRAYER_SWIFT_BLESSING_ID,
    });

    // Server pushes target selection
    await protocolHarness.pushServerMessage(swiftBlessingTargetPrompt());

    // Select ally target
    await selectTarget(page, ALLY_PLAYER_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });
});

test.describe('prayer pray protocol harness', () => {
  test('pray: activate skill with gem', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(prayScenario());

    await activatePanelSkill(page, PRAYER_PRAY_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: PRAYER_PRAY_ID,
    });
  });
});