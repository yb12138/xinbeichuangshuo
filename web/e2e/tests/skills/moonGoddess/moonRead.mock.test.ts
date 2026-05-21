import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  moonReadScenario,
  moonReadConfirmPrompt,
  moonReadTargetPrompt,
  MG_MOON_READ_SKILL_ID,
  ENEMY_PLAYER_ID,
  ENEMY_2_PLAYER_ID,
} from '../../../scenarios/moonGoddess';

// ============================================================
// 月渎 (mg_blasphemy) - 后端通过 response_skills 自动触发
// 确认：skill_choice (1 skill + skip) → skill-branch-overlay
// 目标选择：branch_select，目标分支在前，不发动在最后
// ============================================================

test.describe('moon goddess moon read protocol harness', () => {
  test('moon read: confirm then select enemy target', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(moonReadScenario({ heal: 2 }));

    await protocolHarness.pushServerMessage(moonReadConfirmPrompt());
    await expect(page.getByTestId('skill-branch-overlay')).toBeVisible();
    await page.getByTestId(`prompt-option-${MG_MOON_READ_SKILL_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Target selection via branch_select
    await protocolHarness.pushServerMessage(moonReadTargetPrompt());
    await page.locator('.overlay-panel-root--decision').getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('moon read: select second enemy target', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(moonReadScenario({ heal: 2 }));

    await protocolHarness.pushServerMessage(moonReadConfirmPrompt());
    await page.getByTestId(`prompt-option-${MG_MOON_READ_SKILL_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    await protocolHarness.pushServerMessage(moonReadTargetPrompt());
    await page.locator('.overlay-panel-root--decision').getByTestId('branch-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });

  test('moon read: skip activation', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(moonReadScenario({ heal: 2 }));

    await protocolHarness.pushServerMessage(moonReadConfirmPrompt());
    await expect(page.getByTestId('skill-branch-overlay')).toBeVisible();
    await page.getByTestId('prompt-option-skip').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });

  test('moon read: no trigger when heal = 0', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(moonReadScenario({ heal: 0 }));

    // Should not trigger if no heal available
    await expect(page.getByTestId('skill-branch-overlay')).not.toBeVisible();
  });
});
