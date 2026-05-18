import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  moonReadScenario,
  moonReadTargetPrompt,
} from '../../../scenarios/moonGoddess';

// ============================================================
// 月渎 (mg_blasphemy) - 后端通过 response_skills 自动触发
// 目标选择通过 mg_blasphemy_target choice_type
// ============================================================

test.describe('moon goddess moon read protocol harness', () => {
  test('moon read: triggered via target choice prompt', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(moonReadScenario({ heal: 2 }));

    await protocolHarness.pushServerMessage(moonReadTargetPrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('decision-overlay').getByTestId('branch-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });

  test('moon read: target selection flow', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(moonReadScenario({ heal: 2 }));

    // 目标选择：mg_blasphemy_target
    await protocolHarness.pushServerMessage(moonReadTargetPrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    // 选择第一个敌方目标（index 1，index 0是跳过）
    await page.getByTestId('decision-overlay').getByTestId('branch-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });

  test('moon read: skip via target selection', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(moonReadScenario({ heal: 2 }));

    await protocolHarness.pushServerMessage(moonReadTargetPrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    // 选择跳过（index 0）
    await page.getByTestId('decision-overlay').getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('moon read: no trigger when heal = 0', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(moonReadScenario({ heal: 0 }));

    // Should not trigger if no heal available
    await expect(page.getByTestId('decision-overlay')).not.toBeVisible();
  });
});
