import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  BP_SHARED_LIFE_SKILL_ID,
  ENEMY_PLAYER_ID,
  ENEMY_2_PLAYER_ID,
  sharedLifeConfirmPrompt,
  sharedLifeScenario,
  sharedLifeTargetPrompt,
} from '../../../scenarios/bloodPriestess';

test.describe('blood priestess shared life protocol harness', () => {
  test('shared life: confirm then target', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(sharedLifeScenario());

    // Activate skill
    await page.getByTestId('action-skill').click();
    await page.getByTestId(`skill-${BP_SHARED_LIFE_SKILL_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: BP_SHARED_LIFE_SKILL_ID,
    });

    // Confirm skill activation
    await protocolHarness.pushServerMessage(sharedLifeConfirmPrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('prompt-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Target selection (click enemy player card)
    await protocolHarness.pushServerMessage(sharedLifeTargetPrompt());
    await page.getByTestId(`player-area-${ENEMY_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('shared life: confirm then select second target', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(sharedLifeScenario());

    await page.getByTestId('action-skill').click();
    await page.getByTestId(`skill-${BP_SHARED_LIFE_SKILL_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: BP_SHARED_LIFE_SKILL_ID,
    });

    await protocolHarness.pushServerMessage(sharedLifeConfirmPrompt());
    await page.getByTestId('prompt-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    await protocolHarness.pushServerMessage(sharedLifeTargetPrompt());
    await page.getByTestId(`player-area-${ENEMY_2_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });

  test('shared life: skip confirm', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(sharedLifeScenario());

    await page.getByTestId('action-skill').click();
    await page.getByTestId(`skill-${BP_SHARED_LIFE_SKILL_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: BP_SHARED_LIFE_SKILL_ID,
    });

    await protocolHarness.pushServerMessage(sharedLifeConfirmPrompt());
    await page.getByTestId('prompt-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});