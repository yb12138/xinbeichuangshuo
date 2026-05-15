import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  moonReadScenario,
} from '../../../scenarios/moonGoddess';

// ============================================================
// 月渎 (mg_blasphemy) - 后端通过 response_skills 自动触发
// 目标通过 min_targets 处理
// ============================================================

test.describe('moon goddess moon read protocol harness', () => {
  test('moon read: triggered via response_skills', async ({ protocolHarness }) => {
    await protocolHarness.bootGame(moonReadScenario({ heal: 2 }));

    // 后端会设置 response_skills 触发确认弹框
    await protocolHarness.expectSubmitAction({
      action_type: 'UseSkill',
      skill_id: 'mg_blasphemy',
    });
  });

  test('moon read: no trigger when heal = 0', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(moonReadScenario({ heal: 0 }));

    // Should not trigger if no heal available
    await expect(page.getByTestId('decision-overlay')).not.toBeVisible();
  });
});