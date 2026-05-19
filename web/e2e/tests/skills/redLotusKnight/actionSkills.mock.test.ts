import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  RED_LOTUS_KNIGHT_BLOODY_PRAYER_ID,
  RED_LOTUS_KNIGHT_SCARLET_CROSS_ID,
  ENEMY_PLAYER_ID,
  bloodyPrayerScenario,
  bloodyPrayerXPrompt,
  bloodyPrayerTargetPrompt,
  scarletCrossScenario,
  scarletCrossDiscardPrompt,
  scarletCrossTargetPrompt,
} from '../../../scenarios/redLotusKnight';

async function activatePanelSkill(page: import('@playwright/test').Page, skillId: string) {
  await page.getByTestId('action-skill').click();
  await page.getByTestId(`skill-${skillId}`).click();
}

async function selectTarget(page: import('@playwright/test').Page, targetId: string) {
  await page.getByTestId(`player-area-${targetId}`).click();
}

test.describe('red lotus knight bloody prayer protocol harness', () => {
  test('bloody prayer: activate skill', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(bloodyPrayerScenario());

    await activatePanelSkill(page, RED_LOTUS_KNIGHT_BLOODY_PRAYER_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: RED_LOTUS_KNIGHT_BLOODY_PRAYER_ID,
    });
  });

  test('bloody prayer: select X=1', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(bloodyPrayerScenario());

    await activatePanelSkill(page, RED_LOTUS_KNIGHT_BLOODY_PRAYER_ID);

    // Server pushes X selection
    await protocolHarness.pushServerMessage(bloodyPrayerXPrompt());

    // Select X=1 (numeric: numeric-option-1)
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('decision-overlay').getByTestId('numeric-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Server pushes target selection
    await protocolHarness.pushServerMessage(bloodyPrayerTargetPrompt());

    // Select enemy target (target_picker: click player area)
    await expect(page.getByTestId('decision-overlay')).not.toBeVisible({ timeout: 5000 });
    await selectTarget(page, ENEMY_PLAYER_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('bloody prayer: select X=2', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(bloodyPrayerScenario({ heal: 2 }));

    await activatePanelSkill(page, RED_LOTUS_KNIGHT_BLOODY_PRAYER_ID);

    // Server pushes X selection
    await protocolHarness.pushServerMessage(bloodyPrayerXPrompt());

    // Select X=2 (numeric: numeric-option-2)
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('decision-overlay').getByTestId('numeric-option-2').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });

    // Server pushes target selection
    await protocolHarness.pushServerMessage(bloodyPrayerTargetPrompt());

    // Select enemy target (target_picker: click player area)
    await expect(page.getByTestId('decision-overlay')).not.toBeVisible({ timeout: 5000 });
    await selectTarget(page, ENEMY_PLAYER_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });
});

test.describe('red lotus knight scarlet cross protocol harness', () => {
  test('scarlet cross: activate with crystal', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(scarletCrossScenario());

    await activatePanelSkill(page, RED_LOTUS_KNIGHT_SCARLET_CROSS_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: RED_LOTUS_KNIGHT_SCARLET_CROSS_ID,
    });
  });

  test('scarlet cross: discard and select target', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(scarletCrossScenario());

    await activatePanelSkill(page, RED_LOTUS_KNIGHT_SCARLET_CROSS_ID);

    // Server pushes discard selection
    await protocolHarness.pushServerMessage(scarletCrossDiscardPrompt());

    // Select magic card from hand (card_picker: hand-card-2 auto-submits for min=1,max=1)
    await page.getByTestId('hand-card-2').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      card_ids: ['rlk-magic-1'],
    });

    // Server pushes target selection
    await protocolHarness.pushServerMessage(scarletCrossTargetPrompt());

    // Select enemy target (target_picker: click player area)
    await expect(page.getByTestId('decision-overlay')).not.toBeVisible({ timeout: 5000 });
    await selectTarget(page, ENEMY_PLAYER_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });
});