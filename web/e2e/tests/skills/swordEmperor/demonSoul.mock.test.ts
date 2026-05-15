import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  demonSoulResponsePrompt,
  demonSoulScenario,
} from '../../../scenarios/swordEmperor';

// 恶魔之魂为响应技能，后端通过统一的 choose_skill 入口下发；
// 没有独立的 _confirm choice_type（interrupt_prompt_framework.go）。
test.describe('sword emperor demon soul protocol harness', () => {
  test('demon soul: activate via choose_skill', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(demonSoulScenario());

    await protocolHarness.pushServerMessage(demonSoulResponsePrompt());
    await expect(page.getByTestId('skill-branch-overlay')).toBeVisible();
    await page.getByTestId('skill-branch-overlay').getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('demon soul: skip via choose_skill', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(demonSoulScenario());

    await protocolHarness.pushServerMessage(demonSoulResponsePrompt());
    await expect(page.getByTestId('skill-branch-overlay')).toBeVisible();
    await page.getByTestId('skill-branch-overlay').getByTestId('branch-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});
