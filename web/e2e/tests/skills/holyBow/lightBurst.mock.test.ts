import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  HB_LIGHT_BURST_SKILL_ID,
  lightBurstBranch1TargetPrompt,
  lightBurstBranch2DiscardPrompt,
  lightBurstBranch2HealPrompt,
  lightBurstBranch2TargetPrompt,
  lightBurstBranchPrompt,
  lightBurstScenario,
  ALLY_PLAYER_ID,
  ENEMY_PLAYER_ID,
  ENEMY_2_PLAYER_ID,
} from '../../../scenarios/holyBow';

test.describe('holy bow light burst protocol harness', () => {
  test('branch 1: draw + heal removal + faith, ally gains heal', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(lightBurstScenario({ heal: 2 }));

    // Activate skill
    await page.getByTestId('action-skill').click();
    await page.getByTestId(`skill-${HB_LIGHT_BURST_SKILL_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: HB_LIGHT_BURST_SKILL_ID,
    });

    // Branch selection
    await protocolHarness.pushServerMessage(lightBurstBranchPrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Target selection (click ally player card)
    await protocolHarness.pushServerMessage(lightBurstBranch1TargetPrompt());
    await page.getByTestId(`player-area-${ALLY_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('branch 2: X=1 single target selection (no finish button needed)', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(lightBurstScenario({ heal: 2 }));

    // Activate skill
    await page.getByTestId('action-skill').click();
    await page.getByTestId(`skill-${HB_LIGHT_BURST_SKILL_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: HB_LIGHT_BURST_SKILL_ID,
    });

    // Branch selection
    await protocolHarness.pushServerMessage(lightBurstBranchPrompt());
    await page.getByTestId('branch-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });

    // Heal removal (X=1) - numeric mode
    await protocolHarness.pushServerMessage(lightBurstBranch2HealPrompt(2));
    await page.getByTestId('numeric-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // X=1 → 只允许 1 个目标，无需 finish 按钮迭代
    await protocolHarness.pushServerMessage(
      lightBurstBranch2TargetPrompt({ xValue: 1, selectedCount: 0, withFinish: false }),
    );
    await page.getByTestId(`player-area-${ENEMY_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Discard 1 card
    await protocolHarness.pushServerMessage(lightBurstBranch2DiscardPrompt(1));
    await page.getByTestId('hand-card-0').click();
    await page.getByTestId('prompt-confirm-btn').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      card_ids: ['card_1'],
    });
  });

  // 新增测试：X=2 多目标迭代选择流程
  test('branch 2: X=2 multi-target iteration', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(lightBurstScenario({ heal: 3 }));

    // Activate skill
    await page.getByTestId('action-skill').click();
    await page.getByTestId(`skill-${HB_LIGHT_BURST_SKILL_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: HB_LIGHT_BURST_SKILL_ID,
    });

    // Branch selection
    await protocolHarness.pushServerMessage(lightBurstBranchPrompt());
    await page.getByTestId('branch-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });

    // Heal removal (选择 X=2) - numeric mode
    await protocolHarness.pushServerMessage(lightBurstBranch2HealPrompt(3));
    await page.getByTestId('numeric-option-2').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });

    // 多目标选择：一次 target_picker 内选择两个目标后确认
    await protocolHarness.pushServerMessage(
      lightBurstBranch2TargetPrompt({ xValue: 2 }),
    );
    await page.getByTestId(`player-area-${ENEMY_PLAYER_ID}`).click();
    await page.getByTestId(`player-area-${ENEMY_2_PLAYER_ID}`).click();
    await page.getByTestId('prompt-confirm-btn').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0, 1],
    });

    // Discard 2 cards
    await protocolHarness.pushServerMessage(lightBurstBranch2DiscardPrompt(2));
    await page.getByTestId('hand-card-0').click();
    await page.getByTestId('hand-card-1').click();
    await page.getByTestId('prompt-confirm-btn').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      card_ids: ['card_1', 'card_2'],
    });
  });

  // 新增测试：X=2 选择1个目标后点击finish按钮完成选择
  test('branch 2: X=2 select 1 target then click finish button', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(lightBurstScenario({ heal: 3 }));

    // Activate skill
    await page.getByTestId('action-skill').click();
    await page.getByTestId(`skill-${HB_LIGHT_BURST_SKILL_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: HB_LIGHT_BURST_SKILL_ID,
    });

    // Branch selection
    await protocolHarness.pushServerMessage(lightBurstBranchPrompt());
    await page.getByTestId('branch-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });

    // Heal removal (选择 X=2) - numeric mode
    await protocolHarness.pushServerMessage(lightBurstBranch2HealPrompt(3));
    await page.getByTestId('numeric-option-2').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });

    // 选择 1 个目标后点击确认完成
    await protocolHarness.pushServerMessage(
      lightBurstBranch2TargetPrompt({ xValue: 2 }),
    );
    await page.getByTestId(`player-area-${ENEMY_PLAYER_ID}`).click();
    await page.getByTestId('prompt-confirm-btn').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Discard 2 cards
    await protocolHarness.pushServerMessage(lightBurstBranch2DiscardPrompt(2));
    await page.getByTestId('hand-card-0').click();
    await page.getByTestId('hand-card-1').click();
    await page.getByTestId('prompt-confirm-btn').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      card_ids: ['card_1', 'card_2'],
    });
  });
});
