import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  startSkillChoicePrompt,
  startSkillChoiceScenario,
} from '../../../scenarios/fighter';

test.describe('fighter start skill choice protocol harness', () => {
  test('choose hundred dragon from choose_skill panel', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(startSkillChoiceScenario({ qi: 3, crystals: 1 }));

    await protocolHarness.pushServerMessage(startSkillChoicePrompt());
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

  test('choose heaven drive from choose_skill panel', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(startSkillChoiceScenario({ qi: 4, crystals: 2 }));

    await protocolHarness.pushServerMessage(startSkillChoicePrompt());
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

  test('skip both start skills via the skip option', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(startSkillChoiceScenario({ qi: 3, crystals: 1 }));

    await protocolHarness.pushServerMessage(startSkillChoicePrompt());
    await expect(page.getByTestId('skill-branch-overlay')).toBeVisible();
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
