import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  oneStrikeResponsePrompt,
  oneStrikeScenario,
} from '../../../scenarios/beastSoul';

// 兽灵武士「攻击行动结束」目前唯一需要前端弹框的技能是 一击无念。
// - 武者残心 (bs_warrior_zanshin) 后端 ResponseSilent，自动结算无 UI
// - "不屈意志" 属于剑帝技能，不在兽灵武士技能表内
// 多技能互斥面板由通用 PromptChooseSkill 自动渲染：当多个响应技能同时被触发时，
// skill_ids 包含多项，前端在同一 overlay 渲染所有按钮，因此无需角色专属 mock。
test.describe('beast samurai one strike (attack action end)', () => {
  test('one strike: confirm via response skill prompt', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(oneStrikeScenario({ zanshin: 4 }));

    await protocolHarness.pushServerMessage(oneStrikeResponsePrompt());
    await expect(page.getByTestId('skill-branch-overlay')).toBeVisible();
    await page.getByTestId('skill-branch-overlay').getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('one strike: skip via response skill prompt', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(oneStrikeScenario({ zanshin: 4 }));

    await protocolHarness.pushServerMessage(oneStrikeResponsePrompt());
    await expect(page.getByTestId('skill-branch-overlay')).toBeVisible();
    await page.getByTestId('skill-branch-overlay').getByTestId('prompt-option-skip').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});
