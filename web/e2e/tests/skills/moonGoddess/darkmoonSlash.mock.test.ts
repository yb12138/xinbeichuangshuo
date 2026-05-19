import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  MG_DARKMOON_SLASH_SKILL_ID,
  darkmoonSlashScenario,
  darkmoonSlashResponsePrompt,
  darkmoonSlashXPrompt,
} from '../../../scenarios/moonGoddess';

// ============================================================
// 闇月斩 (mg_darkmoon_slash) - 后端通过 choose_skill 响应窗口触发
// skill_choice with 1 skill + skip → inline buttons in prompt-dialog
// X值选择使用 choice_type: mg_darkmoon_slash_x (numeric)
// ============================================================

test.describe('moon goddess darkmoon slash protocol harness', () => {
  test('darkmoon slash: triggered via response_skills', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(darkmoonSlashScenario({ dark_moon_cards: 2 }));

    // skill_choice with 1 skill + skip → prompt-dialog inline buttons (NOT skill-branch-overlay)
    await protocolHarness.pushServerMessage(darkmoonSlashResponsePrompt());
    await expect(page.getByTestId('prompt-dialog')).toBeVisible();
    await page.getByTestId(`prompt-option-${MG_DARKMOON_SLASH_SKILL_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('darkmoon slash: X value selection flow', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(darkmoonSlashScenario({ dark_moon_cards: 3 }));

    // X值选择：mg_darkmoon_slash_x (maxX = 2 based on handler logic)
    await protocolHarness.pushServerMessage(darkmoonSlashXPrompt(2));
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    // 选择 X=1（第一项，index 0）- numeric picker uses numeric-option-{buttonLabel}
    await page.getByTestId('numeric-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('darkmoon slash: X=2 (max value)', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(darkmoonSlashScenario({ dark_moon_cards: 3 }));

    await protocolHarness.pushServerMessage(darkmoonSlashXPrompt(2));
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    // 选择 X=2（第二项，index 1）- numeric picker uses numeric-option-{buttonLabel}
    await page.getByTestId('numeric-option-2').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});