import { test } from '../../../fixtures/protocolHarness.fixture';
import {
  newMoonShelterScenario,
  newMoonShelterTransformScenario,
} from '../../../scenarios/moonGoddess';

// ============================================================
// 新月庇护 (mg_new_moon_shelter) - 后端通过 response_skills 自动触发
// 爆牌转化流程：from_damage_draw=true时进入暗月形态，吸收爆牌为暗月
// ============================================================

test.describe('moon goddess new moon shelter protocol harness', () => {
  test('new moon shelter: triggered via response_skills', async ({ protocolHarness }) => {
    await protocolHarness.bootGame(newMoonShelterScenario());

    // 后端会设置 response_skills 触发确认弹框
    // 用户点击发动按钮
    await protocolHarness.expectSubmitAction({
      action_type: 'UseSkill',
      skill_id: 'mg_new_moon_shelter',
    });
  });

  test('new moon shelter: transform overflow cards to dark moon form', async ({ protocolHarness }) => {
    await protocolHarness.bootGame(newMoonShelterTransformScenario({ discarded_cards: 2 }));

    // 爆牌转化时，新月庇护会自动吸收爆牌为暗月并进入暗月形态
    // 后端设置 from_damage_draw=true, discarded_cards 有牌
    // Execute: Enter dark moon form + add cards as field cards with EffectMoonDarkMoon
    // DamageVal is set to 0 (no morale loss)
    await protocolHarness.expectSubmitAction({
      action_type: 'UseSkill',
      skill_id: 'mg_new_moon_shelter',
    });
  });
});