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
    // skill_choice 的跳过按钮使用固定 skip id。
    await page
      .getByTestId('skill-branch-overlay')
      .getByTestId('prompt-option-skip')
      .click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [2],
    });
  });
});
