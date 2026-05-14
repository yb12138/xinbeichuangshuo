import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  autoFillBranchPrompt,
  autoFillRewardPrompt,
  autoFillScenario,
} from '../../../scenarios/holyBow';

test.describe('holy bow auto fill protocol harness', () => {
  test('branch 1: consume crystal for faith', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(autoFillScenario({ crystals: 1, gems: 0 }));

    // Server pushes auto fill branch prompt at turn end
    await protocolHarness.pushServerMessage(autoFillBranchPrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    // Click "消耗水晶"
    await page.getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Reward choice: faith
    await protocolHarness.pushServerMessage(autoFillRewardPrompt());
    await page.getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('branch 1: consume crystal for heal', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(autoFillScenario({ crystals: 1, gems: 0 }));

    await protocolHarness.pushServerMessage(autoFillBranchPrompt());
    await page.getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Reward choice: heal
    await protocolHarness.pushServerMessage(autoFillRewardPrompt());
    await page.getByTestId('branch-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });

  test('branch 2: consume gem for crystal and faith', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(autoFillScenario({ crystals: 0, gems: 1 }));

    await protocolHarness.pushServerMessage(autoFillBranchPrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    // Click "消耗红宝石"
    await page.getByTestId('branch-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });

    // Reward choice: faith
    await protocolHarness.pushServerMessage(autoFillRewardPrompt());
    await page.getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('branch 2: consume gem for crystal and heal', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(autoFillScenario({ crystals: 0, gems: 1 }));

    await protocolHarness.pushServerMessage(autoFillBranchPrompt());
    await page.getByTestId('branch-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });

    // Reward choice: heal
    await protocolHarness.pushServerMessage(autoFillRewardPrompt());
    await page.getByTestId('branch-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});