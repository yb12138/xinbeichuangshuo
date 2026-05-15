import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  ENEMY_2_PLAYER_ID,
  swordQiSlashResponsePrompt,
  swordQiSlashScenario,
  swordQiSlashTargetPrompt,
  swordQiSlashXPrompt,
} from '../../../scenarios/swordEmperor';

// 后端实际链路：choose_skill (response) → se_sword_qi_slash_x (选 X) → se_sword_qi_slash_target (选目标)
// 不再有独立的 _confirm / _remove 中间步。
test.describe('sword emperor sword qi slash protocol harness', () => {
  test('sword qi slash: response choose_skill then X=2 then target', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(swordQiSlashScenario({ sword_qi: 3 }));

    // 1) 触发响应技能 choose_skill 弹框
    await protocolHarness.pushServerMessage(swordQiSlashResponsePrompt());
    await expect(page.getByTestId('skill-branch-overlay')).toBeVisible();
    await page.getByTestId('skill-branch-overlay').getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // 2) 选择 X 值（X=2 → option_indexes[1]，因为 X 从 1 开始）
    await protocolHarness.pushServerMessage(swordQiSlashXPrompt(3));
    await page.getByTestId('prompt-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });

    // 3) 选择目标（点击敌方 2 的玩家区，前端按 label 标记映射到 option 0）
    await protocolHarness.pushServerMessage(swordQiSlashTargetPrompt(2));
    await page.getByTestId(`player-area-${ENEMY_2_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('sword qi slash: skip via choose_skill', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(swordQiSlashScenario({ sword_qi: 3 }));

    await protocolHarness.pushServerMessage(swordQiSlashResponsePrompt());
    await expect(page.getByTestId('skill-branch-overlay')).toBeVisible();
    // "跳过" 是 choose_skill 的第二个选项 → option_indexes[1]
    await page.getByTestId('skill-branch-overlay').getByTestId('branch-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });

  test('sword qi slash: remove X=1 then target', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(swordQiSlashScenario({ sword_qi: 3 }));

    await protocolHarness.pushServerMessage(swordQiSlashResponsePrompt());
    await page.getByTestId('skill-branch-overlay').getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // X=1 → option_indexes[0]
    await protocolHarness.pushServerMessage(swordQiSlashXPrompt(3));
    await page.getByTestId('prompt-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    await protocolHarness.pushServerMessage(swordQiSlashTargetPrompt(1));
    await page.getByTestId(`player-area-${ENEMY_2_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('sword qi slash: remove X=3 (max) then target', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(swordQiSlashScenario({ sword_qi: 3 }));

    await protocolHarness.pushServerMessage(swordQiSlashResponsePrompt());
    await page.getByTestId('skill-branch-overlay').getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // X=3 → option_indexes[2]
    await protocolHarness.pushServerMessage(swordQiSlashXPrompt(3));
    await page.getByTestId('prompt-option-2').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [2],
    });

    await protocolHarness.pushServerMessage(swordQiSlashTargetPrompt(3));
    await page.getByTestId(`player-area-${ENEMY_2_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('sword qi slash: no prompt when sword qi = 0', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(swordQiSlashScenario({ sword_qi: 0 }));

    // Should not trigger if no sword qi available
    await expect(page.getByTestId('decision-overlay')).not.toBeVisible();
    await expect(page.getByTestId('skill-branch-overlay')).not.toBeVisible();
  });
});
