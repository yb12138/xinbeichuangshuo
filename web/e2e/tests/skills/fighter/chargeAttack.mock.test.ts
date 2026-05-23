import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  chargeAttackConfirmPrompt,
  chargeAttackScenario,
} from '../../../scenarios/fighter';

test.describe('fighter charge attack protocol harness', () => {
  test('activate charge attack with qi not maxed', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(chargeAttackScenario({ qi: 1 }));

    // 后端实际下发的是 choose_skill 入口（buildResponseSkillPrompt），不是独立 confirm。
    await protocolHarness.pushServerMessage(chargeAttackConfirmPrompt());
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

  test('decline charge attack activation', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(chargeAttackScenario({ qi: 2 }));

    await protocolHarness.pushServerMessage(chargeAttackConfirmPrompt());
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
});
