import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  BP_BACKFLOW_SKILL_ID,
  backflowConfirmPrompt,
  backflowDiscardPrompt,
  backflowScenario,
} from '../../../scenarios/bloodPriestess';

test.describe('blood priestess backflow protocol harness', () => {
  test('backflow: confirm then discard 2 cards', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(backflowScenario());

    // Activate skill
    await page.getByTestId('action-skill').click();
    await page.getByTestId(`skill-${BP_BACKFLOW_SKILL_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: BP_BACKFLOW_SKILL_ID,
    });

    // Confirm skill activation
    await protocolHarness.pushServerMessage(backflowConfirmPrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('prompt-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Discard 2 cards
    await protocolHarness.pushServerMessage(backflowDiscardPrompt());
    await page.getByTestId('hand-card-0').click();
    await page.getByTestId('hand-card-1').click();
    await page.getByTestId('prompt-confirm-btn').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0, 1],
    });
  });

  test('backflow: skip confirm', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(backflowScenario());

    await page.getByTestId('action-skill').click();
    await page.getByTestId(`skill-${BP_BACKFLOW_SKILL_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: BP_BACKFLOW_SKILL_ID,
    });

    await protocolHarness.pushServerMessage(backflowConfirmPrompt());
    await page.getByTestId('prompt-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});