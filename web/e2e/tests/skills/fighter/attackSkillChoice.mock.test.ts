import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  attackSkillChoicePrompt,
  attackSkillChoiceScenario,
} from '../../../scenarios/fighter';

test.describe('fighter attack skill choice protocol harness', () => {
  test('choose charge attack from mutual exclusion', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(attackSkillChoiceScenario({ qi: 2 }));

    // Server pushes skill choice prompt (both charge attack and burst crash can trigger)
    await protocolHarness.pushServerMessage(attackSkillChoicePrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    // Click "发动【蓄力一击】"
    await page.getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('choose burst crash from mutual exclusion', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(attackSkillChoiceScenario({ qi: 3 }));

    await protocolHarness.pushServerMessage(attackSkillChoicePrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    // Click "发动【气绝崩击】"
    await page.getByTestId('branch-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });

  test('skip both skills', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(attackSkillChoiceScenario({ qi: 2 }));

    await protocolHarness.pushServerMessage(attackSkillChoicePrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    // Click "取消" button in overlay footer to skip
    await page.locator('.overlay-panel-cancel').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Cancel',
    });
  });
});