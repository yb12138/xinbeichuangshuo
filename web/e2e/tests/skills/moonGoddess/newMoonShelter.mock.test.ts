import { test } from '../../../fixtures/protocolHarness.fixture';
import {
  newMoonShelterScenario,
} from '../../../scenarios/moonGoddess';

// ============================================================
// 新月庇护 (mg_new_moon_shelter) - 后端通过 response_skills 自动触发
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
});