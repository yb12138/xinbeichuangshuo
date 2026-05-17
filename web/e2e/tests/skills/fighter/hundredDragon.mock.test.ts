import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  hundredDragonConfirmPrompt,
  hundredDragonScenario,
  hundredDragonTargetPrompt,
} from '../../../scenarios/fighter';

test.describe('fighter hundred dragon protocol harness', () => {
  test('activate hundred dragon with 3+ qi at turn start', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(hundredDragonScenario({ qi: 3 }));

    await protocolHarness.pushServerMessage(hundredDragonConfirmPrompt());
    await expect(page.getByTestId('skill-branch-overlay')).toBeVisible();
    await page
      .getByTestId('skill-branch-overlay')
      .getByTestId('branch-option-0')
      .click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('activate hundred dragon with more qi', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(hundredDragonScenario({ qi: 5 }));

    await protocolHarness.pushServerMessage(hundredDragonConfirmPrompt());
    await page
      .getByTestId('skill-branch-overlay')
      .getByTestId('branch-option-0')
      .click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('decline hundred dragon activation', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(hundredDragonScenario({ qi: 4 }));

    await protocolHarness.pushServerMessage(hundredDragonConfirmPrompt());
    await page
      .getByTestId('skill-branch-overlay')
      .getByTestId('branch-option-1')
      .click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });

  // 补充目标锁定流程测试
  test('hundred dragon: select target after activation', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(hundredDragonScenario({ qi: 3 }));

    // 1. 发动确认
    await protocolHarness.pushServerMessage(hundredDragonConfirmPrompt());
    await expect(page.getByTestId('skill-branch-overlay')).toBeVisible();
    await page
      .getByTestId('skill-branch-overlay')
      .getByTestId('branch-option-0')
      .click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // 2. 后端推送目标锁定prompt (fighter_hundred_dragon_target)
    await protocolHarness.pushServerMessage(hundredDragonTargetPrompt());

    // 3. 选择锁定目标（confirm类型prompt使用prompt-option按钮）
    await page.getByTestId('prompt-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  // 验证进入百式幻龙拳形态后的目标锁定状态
  test('hundred dragon: target lock after form activation', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(hundredDragonScenario({ qi: 4 }));

    // 1. 发动确认
    await protocolHarness.pushServerMessage(hundredDragonConfirmPrompt());
    await page
      .getByTestId('skill-branch-overlay')
      .getByTestId('branch-option-0')
      .click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // 2. 后端推送目标锁定prompt
    await protocolHarness.pushServerMessage(hundredDragonTargetPrompt());

    // 3. 选择锁定目标（进入形态后锁定状态）
    await page.getByTestId('prompt-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });
});
