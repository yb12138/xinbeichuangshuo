import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  darkmoonSlashScenario,
  darkmoonSlashXPrompt,
} from '../../../scenarios/moonGoddess';

// ============================================================
// 闇月斩 (mg_darkmoon_slash) - 后端通过 response_skills 自动触发
// X值选择使用 choice_type: mg_darkmoon_slash_x
// ============================================================

test.describe('moon goddess darkmoon slash protocol harness', () => {
  test('darkmoon slash: triggered via response_skills', async ({ protocolHarness }) => {
    await protocolHarness.bootGame(darkmoonSlashScenario({ dark_moon_cards: 2 }));

    // 后端会设置 response_skills 触发确认弹框
    await protocolHarness.expectSubmitAction({
      action_type: 'UseSkill',
      skill_id: 'mg_darkmoon_slash',
    });
  });

  test('darkmoon slash: X value selection flow', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(darkmoonSlashScenario({ dark_moon_cards: 3 }));

    // X值选择：mg_darkmoon_slash_x (maxX = 2 based on handler logic)
    await protocolHarness.pushServerMessage(darkmoonSlashXPrompt(2));
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    // 选择 X=1（第一项，index 0）
    await page.getByTestId('prompt-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('darkmoon slash: X=2 (max value)', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(darkmoonSlashScenario({ dark_moon_cards: 3 }));

    await protocolHarness.pushServerMessage(darkmoonSlashXPrompt(2));
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    // 选择 X=2（第二项，index 1）
    await page.getByTestId('prompt-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});