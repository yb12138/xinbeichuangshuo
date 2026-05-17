import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  soulConvertScenario,
  soulConvertColorPrompt,
  SS_PLAYER_ID,
  SS_SOUL_CONVERT_SKILL_ID,
} from '../../../scenarios/soulSorcerer';

test.describe('soulSorcerer soulConvert protocol harness', () => {
  test('activate soulConvert yellow to blue', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(soulConvertScenario({ blue_soul: 2, yellow_soul: 2 }));

    // Server pushes convert prompt (triggered on attack declaration)
    await protocolHarness.pushServerMessage(soulConvertColorPrompt({ can_y2b: true, can_b2y: true }));
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    // Click "黄魂 -> 蓝魂"
    await page.getByTestId('prompt-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('activate soulConvert blue to yellow', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(soulConvertScenario({ blue_soul: 3, yellow_soul: 1 }));

    await protocolHarness.pushServerMessage(soulConvertColorPrompt({ can_y2b: false, can_b2y: true }));
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    // Click "蓝魂 -> 黄魂"
    await page.getByTestId('prompt-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('soulConvert with both options available', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(soulConvertScenario({ blue_soul: 3, yellow_soul: 3 }));

    await protocolHarness.pushServerMessage(soulConvertColorPrompt({ can_y2b: true, can_b2y: true }));
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    // Click second option "蓝魂 -> 黄魂"
    await page.getByTestId('prompt-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });

  test('decline soulConvert', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(soulConvertScenario({ blue_soul: 2, yellow_soul: 2 }));

    await protocolHarness.pushServerMessage(soulConvertColorPrompt({ can_y2b: true, can_b2y: true }));
    // Close/cancel the prompt (decline skill)
    await page.getByTestId('prompt-cancel').click();
  });
});
