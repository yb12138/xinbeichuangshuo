import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  BP_BLOOD_WAIL_SKILL_ID,
  ENEMY_PLAYER_ID,
  ENEMY_2_PLAYER_ID,
  bloodWailConfirmPrompt,
  bloodWailScenario,
  bloodWailTargetPrompt,
  bloodWailXPrompt,
} from '../../../scenarios/bloodPriestess';

test.describe('blood priestess blood wail protocol harness', () => {
  test('blood wail: confirm then target then X=1', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(bloodWailScenario());

    // Activate skill
    await page.getByTestId('action-skill').click();
    await page.getByTestId(`skill-${BP_BLOOD_WAIL_SKILL_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: BP_BLOOD_WAIL_SKILL_ID,
    });

    // Confirm skill activation (discard unique card)
    await protocolHarness.pushServerMessage(bloodWailConfirmPrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('prompt-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Target selection (click enemy player card)
    await protocolHarness.pushServerMessage(bloodWailTargetPrompt());
    await page.getByTestId(`player-area-${ENEMY_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // X value selection (X<3)
    await protocolHarness.pushServerMessage(bloodWailXPrompt());
    await page.getByRole('button', { name: 'X=1' }).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('blood wail: confirm then target then X=2', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(bloodWailScenario());

    await page.getByTestId('action-skill').click();
    await page.getByTestId(`skill-${BP_BLOOD_WAIL_SKILL_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: BP_BLOOD_WAIL_SKILL_ID,
    });

    await protocolHarness.pushServerMessage(bloodWailConfirmPrompt());
    await page.getByTestId('prompt-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    await protocolHarness.pushServerMessage(bloodWailTargetPrompt());
    await page.getByTestId(`player-area-${ENEMY_2_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });

    await protocolHarness.pushServerMessage(bloodWailXPrompt());
    await page.getByRole('button', { name: 'X=2' }).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });

  test('blood wail: skip confirm', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(bloodWailScenario());

    await page.getByTestId('action-skill').click();
    await page.getByTestId(`skill-${BP_BLOOD_WAIL_SKILL_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: BP_BLOOD_WAIL_SKILL_ID,
    });

    await protocolHarness.pushServerMessage(bloodWailConfirmPrompt());
    await page.getByTestId('prompt-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});