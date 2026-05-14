import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  darkmoonSlashScenario,
} from '../../../scenarios/moonGoddess';

// ============================================================
// 闇月斩 (mg_darkmoon_slash) - 后端通过 response_skills 自动触发
// ============================================================

test.describe('moon goddess darkmoon slash protocol harness', () => {
  test('darkmoon slash: triggered via response_skills', async ({ protocolHarness }) => {
    await protocolHarness.bootGame(darkmoonSlashScenario());

    // 后端会设置 response_skills 触发确认弹框
    await protocolHarness.expectSubmitAction({
      action_type: 'UseSkill',
      skill_id: 'mg_darkmoon_slash',
    });
  });
});