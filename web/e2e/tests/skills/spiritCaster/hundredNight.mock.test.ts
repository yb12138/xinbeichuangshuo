import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  SC_HUNDRED_NIGHT_SKILL_ID,
  ENEMY_PLAYER_ID,
  ENEMY_2_PLAYER_ID,
  ENEMY_3_PLAYER_ID,
  hundredNightScenario,
  hundredNightConfirmPrompt,
  hundredNightManaCollapsePrompt,
  hundredNightRemoveYouliPrompt,
  hundredNightFireBranchPrompt,
  hundredNightSingleTargetPrompt,
  hundredNightAoeExemptTargetPrompt,
} from '../../../scenarios/spiritCaster';

test.describe('spirit caster hundred night protocol harness', () => {
  test('hundred night: confirm -> no crystal -> remove fire youli -> show -> select first exempt', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(hundredNightScenario({ hasCrystal: false }));

    // Hundred night response prompt
    await protocolHarness.pushServerMessage(hundredNightConfirmPrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('prompt-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Remove youli (fire card) - rendered in overlay with branch-option testid
    await protocolHarness.pushServerMessage(hundredNightRemoveYouliPrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Fire branch: show for AOE
    await protocolHarness.pushServerMessage(hundredNightFireBranchPrompt());
    await page.getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Select first exempt target (option index 1 = ENEMY_PLAYER_ID)
    // Target options match player names → rendered as clickable player areas
    await protocolHarness.pushServerMessage(hundredNightAoeExemptTargetPrompt());
    await expect(page.getByTestId('decision-overlay')).not.toBeVisible({ timeout: 5000 });
    await page.getByTestId(`player-area-${ENEMY_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });

  test('hundred night: confirm -> no crystal -> remove fire youli -> hide -> select single target', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(hundredNightScenario({ hasCrystal: false }));

    await protocolHarness.pushServerMessage(hundredNightConfirmPrompt());
    await page.getByTestId('prompt-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    await protocolHarness.pushServerMessage(hundredNightRemoveYouliPrompt());
    await page.getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Fire branch: hide for single target
    await protocolHarness.pushServerMessage(hundredNightFireBranchPrompt());
    await page.getByTestId('branch-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });

    // Select 1 target for single damage (option index 0 = ENEMY_PLAYER_ID)
    // Target options match player names → rendered as clickable player areas
    await protocolHarness.pushServerMessage(hundredNightSingleTargetPrompt());
    await expect(page.getByTestId('decision-overlay')).not.toBeVisible({ timeout: 5000 });
    await page.getByTestId(`player-area-${ENEMY_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('hundred night: confirm -> with crystal -> mana collapse yes -> remove youli -> fire branch -> select exempt', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(hundredNightScenario({ hasCrystal: true }));

    await protocolHarness.pushServerMessage(hundredNightConfirmPrompt());
    await page.getByTestId('prompt-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Mana collapse confirm (has crystal)
    await protocolHarness.pushServerMessage(hundredNightManaCollapsePrompt());
    await page.getByTestId('prompt-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    await protocolHarness.pushServerMessage(hundredNightRemoveYouliPrompt());
    await page.getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Fire branch has long labels → text mode → branch-option-N
    await protocolHarness.pushServerMessage(hundredNightFireBranchPrompt());
    await page.getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // AOE exempt has 4 options matching player names → clickable player areas
    await protocolHarness.pushServerMessage(hundredNightAoeExemptTargetPrompt());
    await expect(page.getByTestId('decision-overlay')).not.toBeVisible({ timeout: 5000 });
    await page.getByTestId(`player-area-${ENEMY_3_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [3],
    });
  });

  test('hundred night: confirm -> with crystal -> mana collapse no -> remove youli -> select target', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(hundredNightScenario({ hasCrystal: true }));

    await protocolHarness.pushServerMessage(hundredNightConfirmPrompt());
    await page.getByTestId('prompt-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    await protocolHarness.pushServerMessage(hundredNightManaCollapsePrompt());
    await page.getByTestId('prompt-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });

    await protocolHarness.pushServerMessage(hundredNightRemoveYouliPrompt());
    await page.getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    await protocolHarness.pushServerMessage(hundredNightFireBranchPrompt());
    await page.getByTestId('branch-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });

    // Single target has 3 options matching player names → clickable player areas
    // option index 1 = ENEMY_2_PLAYER_ID
    await protocolHarness.pushServerMessage(hundredNightSingleTargetPrompt());
    await expect(page.getByTestId('decision-overlay')).not.toBeVisible({ timeout: 5000 });
    await page.getByTestId(`player-area-${ENEMY_2_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });

  test('hundred night: skip activation', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(hundredNightScenario({ hasCrystal: false }));

    await protocolHarness.pushServerMessage(hundredNightConfirmPrompt());
    await page.getByTestId('prompt-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});