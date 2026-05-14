import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  startSkillChoicePrompt,
  startSkillChoiceScenario,
} from '../../../scenarios/fighter';

test.describe('fighter start skill choice protocol harness', () => {
  test('choose hundred dragon from mutual exclusion', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(startSkillChoiceScenario({ qi: 3, crystals: 1 }));

    // Server pushes skill choice prompt (both hundred dragon and heaven drive can trigger)
    await protocolHarness.pushServerMessage(startSkillChoicePrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    // Click "发动【百式幻龙拳】"
    await page.getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('choose heaven drive from mutual exclusion', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(startSkillChoiceScenario({ qi: 4, crystals: 2 }));

    await protocolHarness.pushServerMessage(startSkillChoicePrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    // Click "发动【斗神天驱】"
    await page.getByTestId('branch-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });

  test('skip both start skills', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(startSkillChoiceScenario({ qi: 3, crystals: 1 }));

    await protocolHarness.pushServerMessage(startSkillChoicePrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    // Click "取消" button in overlay footer to skip
    await page.locator('.overlay-panel-cancel').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Cancel',
    });
  });
});