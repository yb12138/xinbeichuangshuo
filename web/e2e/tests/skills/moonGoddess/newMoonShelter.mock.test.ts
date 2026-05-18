import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  newMoonShelterResponsePrompt,
  newMoonShelterScenario,
  newMoonShelterTransformScenario,
} from '../../../scenarios/moonGoddess';

// ============================================================
// 新月庇护 (mg_new_moon_shelter) - 后端通过 response_skills 自动触发
// 爆牌转化流程：from_damage_draw=true时进入暗月形态，吸收爆牌为暗月
// ============================================================

test.describe('moon goddess new moon shelter protocol harness', () => {
  test('new moon shelter: triggered via response_skills', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(newMoonShelterScenario());

    await protocolHarness.pushServerMessage(newMoonShelterResponsePrompt());
    await expect(page.getByTestId('skill-branch-overlay')).toBeVisible();
    await page.getByTestId('skill-branch-overlay').getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('new moon shelter: transform overflow cards to dark moon form', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(newMoonShelterTransformScenario({ discarded_cards: 2 }));

    await protocolHarness.pushServerMessage(newMoonShelterResponsePrompt());
    await expect(page.getByTestId('skill-branch-overlay')).toBeVisible();
    await page.getByTestId('skill-branch-overlay').getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });
});
