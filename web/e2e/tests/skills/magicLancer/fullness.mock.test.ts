import type { Page } from '@playwright/test';
import { test } from '../../../fixtures/protocolHarness.fixture';
import {
  fullnessAllyDiscardPrompt,
  fullnessAllyDiscardScenario,
  fullnessCostCardPrompt,
  fullnessScenario,
  ML_FULLNESS_SKILL_ID,
} from '../../../scenarios/magicLancer';

async function activatePanelSkill(page: Page, skillId: string) {
  await page.getByTestId('action-skill').click();
  await page.getByTestId(`skill-${skillId}`).click();
}

test.describe('magic lancer fullness protocol harness', () => {
  test('activate fullness then select cost card', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(fullnessScenario());

    // Step 1: activate fullness skill
    await activatePanelSkill(page, ML_FULLNESS_SKILL_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: ML_FULLNESS_SKILL_ID,
    });

    // Step 2: server asks for cost card (select 1 magic or thunder card)
    await protocolHarness.pushServerMessage(fullnessCostCardPrompt());

    // The cost prompt is a hand-card picker. Click 圣光 (hand index 2) and confirm.
    await page.getByTestId('hand-card-2').click();
    await page.getByTestId('prompt-confirm-btn').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [2],
    });
  });

  test('ally discard step shows skip option and submits option indexes', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(fullnessAllyDiscardScenario());

    await protocolHarness.pushServerMessage(fullnessAllyDiscardPrompt());

    await page.getByTestId('prompt-cancel-btn').scrollIntoViewIfNeeded();
    await page.getByTestId('prompt-cancel-btn').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    await protocolHarness.pushServerMessage(fullnessAllyDiscardPrompt());

    await page.getByTestId('hand-card-0').click();
    await page.getByTestId('prompt-confirm-btn').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});
