import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  BP_BLOOD_SORROW_SKILL_ID,
  ENEMY_PLAYER_ID,
  bloodSorrowBranchPrompt,
  bloodSorrowScenario,
  bloodSorrowTargetPrompt,
} from '../../../scenarios/bloodPriestess';

test.describe('blood priestess blood sorrow protocol harness', () => {
  test('blood sorrow: choose transfer then target', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(bloodSorrowScenario());

    // Activate skill
    await page.getByTestId('action-skill').click();
    await page.getByTestId(`skill-${BP_BLOOD_SORROW_SKILL_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: BP_BLOOD_SORROW_SKILL_ID,
    });

    // Branch selection
    await protocolHarness.pushServerMessage(bloodSorrowBranchPrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Target selection (click enemy player card)
    await protocolHarness.pushServerMessage(bloodSorrowTargetPrompt());
    await page.getByTestId(`player-area-${ENEMY_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('blood sorrow: choose remove then target', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(bloodSorrowScenario());

    await page.getByTestId('action-skill').click();
    await page.getByTestId(`skill-${BP_BLOOD_SORROW_SKILL_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: BP_BLOOD_SORROW_SKILL_ID,
    });

    await protocolHarness.pushServerMessage(bloodSorrowBranchPrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('branch-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });

    await protocolHarness.pushServerMessage(bloodSorrowTargetPrompt());
    await page.getByTestId(`player-area-${ENEMY_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('blood sorrow: skip', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(bloodSorrowScenario());

    await page.getByTestId('action-skill').click();
    await page.getByTestId(`skill-${BP_BLOOD_SORROW_SKILL_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: BP_BLOOD_SORROW_SKILL_ID,
    });

    await protocolHarness.pushServerMessage(bloodSorrowBranchPrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    // Skip is rendered as cancel button in overlay footer
    await page.getByRole('button', { name: '取消' }).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Cancel',
    });
  });
});