import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  attackSkillChoicePrompt,
  attackSkillChoiceScenario,
} from '../../../scenarios/fighter';

test.describe('fighter attack skill choice protocol harness', () => {
  test('choose charge attack from choose_skill panel', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(attackSkillChoiceScenario({ qi: 2 }));

    // C6: 后端不再用响应技能预归一化做互斥，而是 choose_skill 单次三选一面板。
    await protocolHarness.pushServerMessage(attackSkillChoicePrompt());
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

  test('choose burst crash from choose_skill panel', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(attackSkillChoiceScenario({ qi: 3 }));

    await protocolHarness.pushServerMessage(attackSkillChoicePrompt());
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

  test('skip both skills via the skip option', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(attackSkillChoiceScenario({ qi: 2 }));

    await protocolHarness.pushServerMessage(attackSkillChoicePrompt());
    await expect(page.getByTestId('skill-branch-overlay')).toBeVisible();
    // skip 是 choose_skill 列表中的最后一个选项（index = 2）。
    await page
      .getByTestId('skill-branch-overlay')
      .getByTestId('branch-option-2')
      .click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [2],
    });
  });
});
