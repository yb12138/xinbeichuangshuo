import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  ENEMY_2_PLAYER_ID,
  swordQiSlashConfirmPrompt,
  swordQiSlashRemovePrompt,
  swordQiSlashScenario,
  swordQiSlashTargetPrompt,
} from '../../../scenarios/swordEmperor';

test.describe('sword emperor sword qi slash protocol harness', () => {
  test('sword qi slash: confirm then remove sword qi then target', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(swordQiSlashScenario({ sword_qi: 3 }));

    // Server pushes confirm prompt after attack hit
    await protocolHarness.pushServerMessage(swordQiSlashConfirmPrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('prompt-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Remove X sword qi (choose X=2)
    await protocolHarness.pushServerMessage(swordQiSlashRemovePrompt(3));
    await page.getByTestId('branch-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });

    // Target selection (click enemy 2 player card - not the attack target)
    await protocolHarness.pushServerMessage(swordQiSlashTargetPrompt(2));
    await page.getByTestId(`player-area-${ENEMY_2_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('sword qi slash: skip confirm', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(swordQiSlashScenario({ sword_qi: 3 }));

    await protocolHarness.pushServerMessage(swordQiSlashConfirmPrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('prompt-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });

  test('sword qi slash: remove 1 sword qi', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(swordQiSlashScenario({ sword_qi: 3 }));

    await protocolHarness.pushServerMessage(swordQiSlashConfirmPrompt());
    await page.getByTestId('prompt-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Choose X=1
    await protocolHarness.pushServerMessage(swordQiSlashRemovePrompt(3));
    await page.getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Target selection for 1 damage
    await protocolHarness.pushServerMessage(swordQiSlashTargetPrompt(1));
    await page.getByTestId(`player-area-${ENEMY_2_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('sword qi slash: remove 3 sword qi (max)', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(swordQiSlashScenario({ sword_qi: 3 }));

    await protocolHarness.pushServerMessage(swordQiSlashConfirmPrompt());
    await page.getByTestId('prompt-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Choose X=3 (max)
    await protocolHarness.pushServerMessage(swordQiSlashRemovePrompt(3));
    await page.getByTestId('branch-option-2').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [2],
    });

    // Target selection for 3 damage
    await protocolHarness.pushServerMessage(swordQiSlashTargetPrompt(3));
    await page.getByTestId(`player-area-${ENEMY_2_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('sword qi slash: no prompt when sword qi = 0', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(swordQiSlashScenario({ sword_qi: 0 }));

    // Should not trigger if no sword qi available
    await expect(page.getByTestId('decision-overlay')).not.toBeVisible();
  });
});