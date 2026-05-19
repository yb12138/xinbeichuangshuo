import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  angelSoulResponsePrompt,
  angelSoulScenario,
} from '../../../scenarios/swordEmperor';

// 天使之魂为响应技能，后端通过统一的 choose_skill 入口下发；
// 没有独立的 _confirm choice_type（interrupt_prompt_framework.go）。
test.describe('sword emperor angel soul protocol harness', () => {
  test('angel soul: activate via choose_skill', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(angelSoulScenario());

    await protocolHarness.pushServerMessage(angelSoulResponsePrompt());
    await expect(page.getByTestId('skill-branch-overlay')).toBeVisible();
    await page.getByTestId('skill-branch-overlay').getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('angel soul: skip via choose_skill', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(angelSoulScenario());

    await protocolHarness.pushServerMessage(angelSoulResponsePrompt());
    await expect(page.getByTestId('skill-branch-overlay')).toBeVisible();
    await page.getByTestId('prompt-option-skip').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});
