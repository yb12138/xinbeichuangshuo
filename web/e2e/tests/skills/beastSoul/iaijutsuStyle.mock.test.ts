import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  BSW_IAIJUTSU_STYLE_SKILL_ID,
  iaijutsuStyleConfirmPrompt,
  iaijutsuStyleChoicePrompt,
  iaijutsuStyleDiscardPrompt,
  iaijutsuStyleScenario,
} from '../../../scenarios/beastSoul';

test.describe('beast soul warrior iaijutsu style protocol harness', () => {
  test('iaijutsu style: activate then choose draw', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(iaijutsuStyleScenario({ gems: 1 }));

    // Activate skill
    await page.getByTestId('action-skill').click();
    await page.getByTestId(`skill-${BSW_IAIJUTSU_STYLE_SKILL_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: BSW_IAIJUTSU_STYLE_SKILL_ID,
    });

    // Confirm skill activation
    await protocolHarness.pushServerMessage(iaijutsuStyleConfirmPrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('prompt-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Choose draw branch
    await protocolHarness.pushServerMessage(iaijutsuStyleChoicePrompt());
    await page.getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('iaijutsu style: activate then choose discard', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(iaijutsuStyleScenario({ gems: 1 }));

    // Activate skill
    await page.getByTestId('action-skill').click();
    await page.getByTestId(`skill-${BSW_IAIJUTSU_STYLE_SKILL_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: BSW_IAIJUTSU_STYLE_SKILL_ID,
    });

    // Confirm skill activation
    await protocolHarness.pushServerMessage(iaijutsuStyleConfirmPrompt());
    await page.getByTestId('prompt-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Choose discard branch
    await protocolHarness.pushServerMessage(iaijutsuStyleChoicePrompt());
    await page.getByTestId('branch-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });

    // Discard 2 cards
    await protocolHarness.pushServerMessage(iaijutsuStyleDiscardPrompt());
    await page.getByTestId('hand-card-0').click();
    await page.getByTestId('hand-card-1').click();
    await page.getByTestId('prompt-confirm-btn').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0, 1],
    });
  });

  test('iaijutsu style: skip confirm', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(iaijutsuStyleScenario({ gems: 1 }));

    // Activate skill
    await page.getByTestId('action-skill').click();
    await page.getByTestId(`skill-${BSW_IAIJUTSU_STYLE_SKILL_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: BSW_IAIJUTSU_STYLE_SKILL_ID,
    });

    // Skip skill activation
    await protocolHarness.pushServerMessage(iaijutsuStyleConfirmPrompt());
    await page.getByTestId('prompt-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});