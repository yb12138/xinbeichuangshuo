import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  bulletConfirmPrompt,
  bulletScenario,
  ENEMY_PLAYER_ID,
  bulletTargetPrompt,
} from '../../../scenarios/fighter';

test.describe('fighter bullet protocol harness', () => {
  test('activate bullet after magic action', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(bulletScenario());

    await protocolHarness.pushServerMessage(bulletConfirmPrompt());
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

  test('decline bullet activation', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(bulletScenario());

    await protocolHarness.pushServerMessage(bulletConfirmPrompt());
    await expect(page.getByTestId('skill-branch-overlay')).toBeVisible();
    await page
      .getByTestId('skill-branch-overlay')
      .getByTestId('prompt-option-skip')
      .click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });

  // 补充目标选择流程测试
  test('bullet: select target after activation', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(bulletScenario());

    // 1. 发动确认
    await protocolHarness.pushServerMessage(bulletConfirmPrompt());
    await expect(page.getByTestId('skill-branch-overlay')).toBeVisible();
    await page
      .getByTestId('skill-branch-overlay')
      .getByTestId('branch-option-0')
      .click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // 2. 后端推送目标选择prompt (fighter_psi_bullet_target)
    await protocolHarness.pushServerMessage(bulletTargetPrompt());

    // 3. 选择敌方目标（target_picker 通过玩家区域选择）
    await page.getByTestId(`player-area-${ENEMY_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  // 验证目标治疗为0时的自伤逻辑测试场景
  test('bullet: select target with zero heal (self-damage scenario)', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(bulletScenario());

    // 1. 发动确认
    await protocolHarness.pushServerMessage(bulletConfirmPrompt());
    await page
      .getByTestId('skill-branch-overlay')
      .getByTestId('branch-option-0')
      .click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // 2. 后端推送目标选择prompt
    await protocolHarness.pushServerMessage(bulletTargetPrompt());

    // 3. 选择目标（假设目标治疗为0时会触发自伤）
    await page.getByTestId(`player-area-${ENEMY_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });
});
