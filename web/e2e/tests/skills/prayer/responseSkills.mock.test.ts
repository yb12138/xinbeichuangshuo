import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  powerBlessingTriggerScenario,
  swiftBlessingTriggerScenario,
  powerBlessingTriggerPrompt,
  swiftBlessingTriggerPrompt,
  manaTideScenario,
  manaTidePrompt,
} from '../../../scenarios/prayer';

test.describe('prayer power blessing trigger protocol harness', () => {
  test('power blessing: ally triggers on hit', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(powerBlessingTriggerScenario());

    // Server pushes power blessing trigger prompt to ally when they hit
    await protocolHarness.pushServerMessage(powerBlessingTriggerPrompt());

    // branch_select overlay - confirm (branch-option-0)
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('decision-overlay').getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('power blessing: ally skips trigger', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(powerBlessingTriggerScenario());

    await protocolHarness.pushServerMessage(powerBlessingTriggerPrompt());

    // branch_select overlay - skip (branch-option-1)
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('decision-overlay').getByTestId('branch-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});

test.describe('prayer swift blessing trigger protocol harness', () => {
  test('swift blessing: ally triggers on action end', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(swiftBlessingTriggerScenario());

    // Server pushes swift blessing trigger prompt to ally after action
    await protocolHarness.pushServerMessage(swiftBlessingTriggerPrompt());

    // branch_select overlay - confirm (branch-option-0)
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('decision-overlay').getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('swift blessing: ally skips trigger', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(swiftBlessingTriggerScenario());

    await protocolHarness.pushServerMessage(swiftBlessingTriggerPrompt());

    // branch_select overlay - skip (branch-option-1)
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('decision-overlay').getByTestId('branch-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});

test.describe('prayer mana tide protocol harness', () => {
  test('mana tide: confirm extra magic action', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(manaTideScenario());

    // Server pushes mana tide prompt after magic action ends
    await protocolHarness.pushServerMessage(manaTidePrompt());

    // branch_select overlay - confirm (branch-option-0)
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('decision-overlay').getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('mana tide: skip', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(manaTideScenario());

    await protocolHarness.pushServerMessage(manaTidePrompt());

    // branch_select overlay - skip (branch-option-1)
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('decision-overlay').getByTestId('branch-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});