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

  test('branch 2: heal removal, target selection, discard, damage', async ({ page, protocolHarness }) => {
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
    await protocolHarness.pushServerMessage(lightBurstBranch2HealPrompt());
    await page.getByTestId('numeric-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // 后端 hb_light_burst_mode_b_targets 为每次单选 + finish 按钮迭代：
    // 这里 X=1 → 只允许 1 个目标，无需 finish。
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
      option_indexes: [0],
    });
  });
});