import type { Page } from '@playwright/test';
import { test, expect } from '../../../fixtures/protocolHarness.fixture';
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
  await page.getByTestId('action-hub-trigger').click();
  await page.getByTestId('action-magic').click();
  await page.getByTestId('action-skill').click();
  await page.getByTestId(`skill-${skillId}`).click();
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

    // Server pushes discard selection (card_picker from hand)
    await protocolHarness.pushServerMessage(gloryBeliefDiscardPrompt());

    // Select magic card to discard (hand-card-0 = prayer-magic-1)
    await page.getByTestId('hand-card-0').click();
    await page.getByTestId('prompt-confirm-btn').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      card_ids: ['prayer-magic-1'],
    });

    // Server pushes target selection (target_picker)
    await protocolHarness.pushServerMessage(gloryBeliefTargetPrompt());

    // Click ally player area (auto-submits for single target)
    await page.getByTestId(`player-area-${ALLY_PLAYER_ID}`).click();
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

    // Server pushes discard selection (card_picker from hand)
    await protocolHarness.pushServerMessage(darkCurseDiscardPrompt());

    // Select card to discard (hand-card-0 = prayer-magic-1)
    await page.getByTestId('hand-card-0').click();
    await page.getByTestId('prompt-confirm-btn').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      card_ids: ['prayer-magic-1'],
    });

    // Server pushes target selection (target_picker)
    await protocolHarness.pushServerMessage(darkCurseTargetPrompt());

    // Click enemy player area (auto-submits for single target)
    await page.getByTestId(`player-area-${ENEMY_PLAYER_ID}`).click();
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

    // Server pushes target selection (target_picker)
    await protocolHarness.pushServerMessage(powerBlessingTargetPrompt());

    // Click ally player area (auto-submits for single target)
    await page.getByTestId(`player-area-${ALLY_PLAYER_ID}`).click();
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

    // Server pushes target selection (target_picker)
    await protocolHarness.pushServerMessage(swiftBlessingTargetPrompt());

    // Click ally player area (auto-submits for single target)
    await page.getByTestId(`player-area-${ALLY_PLAYER_ID}`).click();
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