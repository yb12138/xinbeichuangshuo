import { test } from '../../../fixtures/protocolHarness.fixture';
import {
  medusaEyeDarkMoonPrompt,
  medusaEyeScenario,
} from '../../../scenarios/moonGoddess';

// ============================================================
// 美杜莎之眼 (mg_medusa_eye) - 后端通过 response_skills 触发
// 闇月选择使用 choice_type: mg_medusa_darkmoon_pick
// 弃牌通过 system_discard_cards，目标通过 min_targets 处理
// ============================================================

test.describe('moon goddess medusa eye protocol harness', () => {
  test('medusa eye: dark moon card selection', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(medusaEyeScenario({ dark_moon_cards: 2 }));

    // 闇月选择使用后端定义的 choice_type
    await protocolHarness.pushServerMessage(medusaEyeDarkMoonPrompt());
    // Select the magic card (first option is 暗月法术)
    await page.getByTestId('hand-card-0').click();
    await page.getByTestId('prompt-confirm-btn').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('medusa eye: select attack card as dark moon', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(medusaEyeScenario({ dark_moon_cards: 2 }));

    // 闇月选择使用后端定义的 choice_type
    await protocolHarness.pushServerMessage(medusaEyeDarkMoonPrompt());
    // Select the attack card (second option is 火焰斩)
    await page.getByTestId('hand-card-1').click();
    await page.getByTestId('prompt-confirm-btn').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });

  test('medusa eye: triggered via response_skills', async ({ protocolHarness }) => {
    await protocolHarness.bootGame(medusaEyeScenario({ dark_moon_cards: 2 }));

    // 后端会设置 response_skills 触发确认弹框
    await protocolHarness.expectSubmitAction({
      action_type: 'UseSkill',
      skill_id: 'mg_medusa_eye',
    });
  });
});