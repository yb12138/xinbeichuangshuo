import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  medusaEyeDarkMoonPrompt,
  medusaEyeResponsePrompt,
  medusaEyeMagicDiscardPrompt,
  medusaEyeScenario,
} from '../../../scenarios/moonGoddess';

// ============================================================
// 美杜莎之眼 (mg_medusa_eye) - 后端通过 response_skills 触发
// 闇月选择使用 choice_type: mg_medusa_darkmoon_pick（扩展区盖牌点选）
// 法术闇月弃牌使用 choice_type: mg_medusa_magic_discard
// ============================================================

test.describe('moon goddess medusa eye protocol harness', () => {
  test('medusa eye: asks before opening dark moon picker', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(medusaEyeScenario({ dark_moon_cards: 2 }));

    await protocolHarness.pushServerMessage(medusaEyeResponsePrompt());
    await page.getByTestId('prompt-option-mg_medusa_eye').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('medusa eye: dark moon card selection via cover picker', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(medusaEyeScenario({ dark_moon_cards: 2 }));

    await protocolHarness.pushServerMessage(medusaEyeDarkMoonPrompt());
    await expect(page.getByTestId('decision-overlay')).not.toBeVisible();

    await page.getByTestId('cover-card-0').click();
    await expect(page.getByTestId('prompt-confirm-btn')).toBeVisible();
    await page.getByTestId('prompt-confirm-btn').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('medusa eye: select attack dark moon via cover picker', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(medusaEyeScenario({ dark_moon_cards: 2 }));

    await protocolHarness.pushServerMessage(medusaEyeDarkMoonPrompt());
    await expect(page.getByTestId('decision-overlay')).not.toBeVisible();

    await page.getByTestId('cover-card-1').click();
    await page.getByTestId('prompt-confirm-btn').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });

  test('medusa eye: magic dark moon triggers discard then damage', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(medusaEyeScenario({ dark_moon_cards: 2 }));

    await protocolHarness.pushServerMessage(medusaEyeDarkMoonPrompt());
    await expect(page.getByTestId('decision-overlay')).not.toBeVisible();
    await page.getByTestId('cover-card-0').click();
    await page.getByTestId('prompt-confirm-btn').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    await protocolHarness.pushServerMessage(medusaEyeMagicDiscardPrompt());
    await page.getByTestId('hand-card-0').click();
    await page.getByTestId('prompt-confirm-btn').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      card_ids: ['card_1'],
    });
  });

  test('medusa eye: triggered via field cover prompt', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(medusaEyeScenario({ dark_moon_cards: 2 }));

    await protocolHarness.pushServerMessage(medusaEyeDarkMoonPrompt());
    await expect(page.getByTestId('decision-overlay')).not.toBeVisible();
    await page.getByTestId('cover-card-0').click();
    await page.getByTestId('prompt-confirm-btn').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });
});
