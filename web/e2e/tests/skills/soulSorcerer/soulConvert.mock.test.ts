import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  soulConvertScenario,
  soulConvertColorPrompt,
} from '../../../scenarios/soulSorcerer';

test.describe('soulSorcerer soulConvert protocol harness', () => {
  test('activate soulConvert yellow to blue', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(soulConvertScenario({ blue_soul: 2, yellow_soul: 2 }));

    // Server pushes convert prompt with 3 options (y2b, b2y, cancel)
    await protocolHarness.pushServerMessage(soulConvertColorPrompt({ can_y2b: true, can_b2y: true }));
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await expect(page.getByText('黄色灵魂转蓝色灵魂')).toBeVisible();
    await expect(page.getByText('蓝色灵魂转黄色灵魂')).toBeVisible();
    await expect(page.getByText('取消')).toBeVisible();
    await page.getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('activate soulConvert blue to yellow', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(soulConvertScenario({ blue_soul: 3, yellow_soul: 1 }));

    await protocolHarness.pushServerMessage(soulConvertColorPrompt({ can_y2b: false, can_b2y: true }));
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await expect(page.getByText('蓝色灵魂转黄色灵魂')).toBeVisible();
    await expect(page.getByText('取消')).toBeVisible();
    await page.getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('soulConvert with both options available', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(soulConvertScenario({ blue_soul: 3, yellow_soul: 3 }));

    await protocolHarness.pushServerMessage(soulConvertColorPrompt({ can_y2b: true, can_b2y: true }));
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await expect(page.getByText('蓝色灵魂转黄色灵魂')).toBeVisible();
    await page.getByTestId('branch-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });

  test('soulConvert cancel', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(soulConvertScenario({ blue_soul: 3, yellow_soul: 3 }));

    await protocolHarness.pushServerMessage(soulConvertColorPrompt({ can_y2b: true, can_b2y: true }));
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await expect(page.getByText('取消')).toBeVisible();
    // Cancel is the last option (index 2)
    await page.getByTestId('branch-option-2').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [2],
    });
  });

  test('shows named direction options instead of numeric choices', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(soulConvertScenario({ blue_soul: 2, yellow_soul: 2 }));

    await protocolHarness.pushServerMessage(soulConvertColorPrompt({ can_y2b: true, can_b2y: true }));
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await expect(page.getByText('黄色灵魂转蓝色灵魂')).toBeVisible();
    await expect(page.getByText('蓝色灵魂转黄色灵魂')).toBeVisible();
    await expect(page.getByText('取消')).toBeVisible();
    await expect(page.getByTestId('numeric-option-1')).toHaveCount(0);
    await expect(page.getByTestId('numeric-option-2')).toHaveCount(0);
  });
});
