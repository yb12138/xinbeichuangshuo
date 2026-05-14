import type { Page } from '@playwright/test';
import { test } from '../../../fixtures/protocolHarness.fixture';
import {
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

    // The cost card prompt is confirm type rendered as inline or overlay buttons.
    // Click the first valid option (圣光, index 2 in hand).
    const overlay = page.getByTestId('decision-overlay');
    const overlayVisible = await overlay.isVisible({ timeout: 1000 }).catch(() => false);
    if (overlayVisible) {
      await overlay.getByTestId('prompt-option-2').click();
    } else {
      // Inline button in prompt dialog
      await page.getByTestId('prompt-dialog').getByTestId('prompt-option-2').click();
    }
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [2],
    });
  });
});
