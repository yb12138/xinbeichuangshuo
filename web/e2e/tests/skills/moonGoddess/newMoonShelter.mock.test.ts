import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  MG_NEW_MOON_SHelter_SKILL_ID,
  newMoonShelterResponsePrompt,
  newMoonShelterScenario,
  newMoonShelterTransformScenario,
} from '../../../scenarios/moonGoddess';

// ============================================================
// 新月庇护 (mg_new_moon_shelter) - 后端通过 response_skills 自动触发
// skill_choice with 1 skill + skip → inline buttons in prompt-dialog
// ============================================================

test.describe('moon goddess new moon shelter protocol harness', () => {
  test('new moon shelter: triggered via response_skills', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(newMoonShelterScenario());

    await protocolHarness.pushServerMessage(newMoonShelterResponsePrompt());
    // skill_choice with 1 skill + skip → prompt-dialog inline buttons (NOT skill-branch-overlay)
    await expect(page.getByTestId('prompt-dialog')).toBeVisible();
    await page.getByTestId(`prompt-option-${MG_NEW_MOON_SHelter_SKILL_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('new moon shelter: transform overflow cards to dark moon form', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(newMoonShelterTransformScenario({ discarded_cards: 2 }));

    await protocolHarness.pushServerMessage(newMoonShelterResponsePrompt());
    await expect(page.getByTestId('prompt-dialog')).toBeVisible();
    await page.getByTestId(`prompt-option-${MG_NEW_MOON_SHelter_SKILL_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });
});