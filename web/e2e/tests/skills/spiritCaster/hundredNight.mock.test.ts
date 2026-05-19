import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  SC_HUNDRED_NIGHT_SKILL_ID,
  ENEMY_PLAYER_ID,
  ENEMY_2_PLAYER_ID,
  ENEMY_3_PLAYER_ID,
  SC_PLAYER_ID,
  hundredNightScenario,
  hundredNightConfirmPrompt,
  hundredNightManaCollapsePrompt,
  hundredNightRemoveYouliPrompt,
  hundredNightFireBranchPrompt,
  hundredNightSingleTargetPrompt,
  hundredNightAoeExemptTargetPrompt,
} from '../../../scenarios/spiritCaster';

async function pickFirstYouliCover(page: import('@playwright/test').Page) {
  await expect(page.getByTestId('decision-overlay')).not.toBeVisible({ timeout: 5000 });
  const cover = page.getByTestId('cover-card-0');
  if (!(await cover.isVisible().catch(() => false))) {
    await page.getByRole('button', { name: /扩展区/ }).click();
  }
  await page.getByTestId('cover-card-0').click();
}

test.describe('spirit caster hundred night protocol harness', () => {
  test('hundred night: confirm -> no crystal -> remove fire youli -> show -> select first exempt', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(hundredNightScenario({ hasCrystal: false }));

    // Hundred night response prompt (branch_select overlay)
    await protocolHarness.pushServerMessage(hundredNightConfirmPrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('prompt-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Remove youli (fire card) - field card_picker, auto-submit with card_ids
    await protocolHarness.pushServerMessage(hundredNightRemoveYouliPrompt());
    await pickFirstYouliCover(page);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      card_ids: ['sc-youli-0'],
    });

    // Fire branch: show for AOE (branch_select overlay)
    await protocolHarness.pushServerMessage(hundredNightFireBranchPrompt());
    await page.getByTestId('prompt-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Select first exempt target (target_picker - click player area)
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
    await pickFirstYouliCover(page);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      card_ids: ['sc-youli-0'],
    });

    // Fire branch: hide for single target (branch_select overlay)
    await protocolHarness.pushServerMessage(hundredNightFireBranchPrompt());
    await page.getByTestId('prompt-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });

    // Select 1 target for single damage (target_picker - click player area)
    await protocolHarness.pushServerMessage(hundredNightSingleTargetPrompt());
    await expect(page.getByTestId('decision-overlay')).not.toBeVisible({ timeout: 5000 });
    await page.getByTestId(`player-area-${ENEMY_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
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

    // Mana collapse confirm (has crystal) - branch_select overlay
    await protocolHarness.pushServerMessage(hundredNightManaCollapsePrompt());
    await page.getByTestId('prompt-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    await protocolHarness.pushServerMessage(hundredNightRemoveYouliPrompt());
    await pickFirstYouliCover(page);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      card_ids: ['sc-youli-0'],
    });

    // Fire branch: show for AOE (branch_select overlay)
    await protocolHarness.pushServerMessage(hundredNightFireBranchPrompt());
    await page.getByTestId('prompt-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // AOE exempt target (target_picker - click player area)
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
    await pickFirstYouliCover(page);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      card_ids: ['sc-youli-0'],
    });

    await protocolHarness.pushServerMessage(hundredNightFireBranchPrompt());
    await page.getByTestId('prompt-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });

    // Single target (target_picker - click player area)
    await protocolHarness.pushServerMessage(hundredNightSingleTargetPrompt());
    await expect(page.getByTestId('decision-overlay')).not.toBeVisible({ timeout: 5000 });
    await page.getByTestId(`player-area-${ENEMY_2_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [2],
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