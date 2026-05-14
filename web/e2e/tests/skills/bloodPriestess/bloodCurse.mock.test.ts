import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  BP_BLOOD_CURSE_SKILL_ID,
  ENEMY_PLAYER_ID,
  bloodCurseConfirmPrompt,
  bloodCurseDiscardPrompt,
  bloodCurseScenario,
  bloodCurseTargetPrompt,
} from '../../../scenarios/bloodPriestess';

test.describe('blood priestess blood curse protocol harness', () => {
  test('blood curse: confirm then target then discard 3 cards', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(bloodCurseScenario());

    // Activate skill
    await page.getByTestId('action-skill').click();
    await page.getByTestId(`skill-${BP_BLOOD_CURSE_SKILL_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: BP_BLOOD_CURSE_SKILL_ID,
    });

    // Confirm skill activation
    await protocolHarness.pushServerMessage(bloodCurseConfirmPrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('prompt-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Target selection (click enemy player card)
    await protocolHarness.pushServerMessage(bloodCurseTargetPrompt());
    await page.getByTestId(`player-area-${ENEMY_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Discard 3 cards
    await protocolHarness.pushServerMessage(bloodCurseDiscardPrompt());
    await page.getByTestId('hand-card-0').click();
    await page.getByTestId('hand-card-1').click();
    await page.getByTestId('hand-card-2').click();
    await page.getByTestId('prompt-confirm-btn').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0, 1, 2],
    });
  });

  test('blood curse: skip confirm', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(bloodCurseScenario());

    await page.getByTestId('action-skill').click();
    await page.getByTestId(`skill-${BP_BLOOD_CURSE_SKILL_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: BP_BLOOD_CURSE_SKILL_ID,
    });

    await protocolHarness.pushServerMessage(bloodCurseConfirmPrompt());
    await page.getByTestId('prompt-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});