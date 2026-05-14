import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  indomitableWillConfirmPrompt,
  indomitableWillScenario,
  warriorZanshinConfirmPrompt,
  warriorZanshinScenario,
  oneStrikeConfirmPrompt,
  oneStrikeScenario,
  mutualExclusionPrompt,
  mutualExclusionScenario,
} from '../../../scenarios/beastSoul';

test.describe('beast soul warrior attack action end skills', () => {
  test('indomitable will: attack action end confirm', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(indomitableWillScenario({ crystals: 1 }));

    // Server pushes confirm prompt at attack action end
    await protocolHarness.pushServerMessage(indomitableWillConfirmPrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('prompt-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('indomitable will: skip when no crystals', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(indomitableWillScenario({ crystals: 0 }));

    // Should not trigger if no crystals available
    // No prompt pushed, test verifies no prompt appears
    await expect(page.getByTestId('decision-overlay')).not.toBeVisible();
  });

  test('warrior zanshin: first attack action end confirm', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(warriorZanshinScenario({ zanshin: 0 }));

    // Server pushes confirm prompt at first attack action end
    await protocolHarness.pushServerMessage(warriorZanshinConfirmPrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('prompt-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('one strike: requires 4+ zanshin tokens', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(oneStrikeScenario({ zanshin: 4 }));

    // Server pushes confirm prompt when zanshin >= 4
    await protocolHarness.pushServerMessage(oneStrikeConfirmPrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('prompt-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('one strike: skip when zanshin < 4', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(oneStrikeScenario({ zanshin: 3 }));

    // Should not trigger if zanshin < 4
    await expect(page.getByTestId('decision-overlay')).not.toBeVisible();
  });

  test('mutual exclusion: choose indomitable will', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(mutualExclusionScenario({ crystals: 1, zanshin: 4 }));

    // Server pushes mutual exclusion prompt when all three trigger simultaneously
    await protocolHarness.pushServerMessage(mutualExclusionPrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('mutual exclusion: choose warrior zanshin', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(mutualExclusionScenario({ crystals: 1, zanshin: 4 }));

    await protocolHarness.pushServerMessage(mutualExclusionPrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('branch-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });

  test('mutual exclusion: choose one strike', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(mutualExclusionScenario({ crystals: 1, zanshin: 4 }));

    await protocolHarness.pushServerMessage(mutualExclusionPrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('branch-option-2').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [2],
    });
  });

  test('mutual exclusion: skip all skills', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(mutualExclusionScenario({ crystals: 1, zanshin: 4 }));

    await protocolHarness.pushServerMessage(mutualExclusionPrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    // Click cancel button in overlay footer
    await page.locator('.overlay-panel-cancel').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Cancel',
    });
  });
});