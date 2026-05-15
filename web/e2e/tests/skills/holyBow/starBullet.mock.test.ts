import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  HB_STAR_BULLET_SKILL_ID,
  starBulletResponsePrompt,
  starBulletCostPrompt,
  starBulletTargetPrompt,
  starBulletScenario,
  ALLY_PLAYER_ID,
} from '../../../scenarios/holyBow';

test.describe('holy bow meteor bullet (流星圣弹) protocol harness', () => {
  test('activate meteor bullet: confirm via response skill, then cost, then ally target', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(starBulletScenario());

    // 1) 响应技能 PromptChooseSkill 入口
    await protocolHarness.pushServerMessage(starBulletResponsePrompt());
    await expect(page.getByTestId('skill-branch-overlay')).toBeVisible();
    await page
      .getByTestId('skill-branch-overlay')
      .getByTestId('branch-option-0')
      .click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // 2) hb_meteor_bullet_cost：选择移除 1 治疗
    await protocolHarness.pushServerMessage(starBulletCostPrompt());
    await page.getByTestId('prompt-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // 3) hb_meteor_bullet_target：选择获得治疗的队友
    await protocolHarness.pushServerMessage(starBulletTargetPrompt());
    await page.getByTestId(`player-area-${ALLY_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('decline meteor bullet activation via skip', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(starBulletScenario());

    await protocolHarness.pushServerMessage(starBulletResponsePrompt());
    await expect(page.getByTestId('skill-branch-overlay')).toBeVisible();
    await page
      .getByTestId('skill-branch-overlay')
      .getByTestId('branch-option-1')
      .click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});

// 保持 HB_STAR_BULLET_SKILL_ID 引用，避免 lint 误报。
void HB_STAR_BULLET_SKILL_ID;
