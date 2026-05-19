import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  ENEMY_PLAYER_ID,
  moonCycleBranchPrompt,
  moonCycleScenario,
  moonCycleTargetPrompt,
} from '../../../scenarios/moonGoddess';

test.describe('moon goddess moon cycle protocol harness', () => {
  test('moon cycle: branch 1 - remove dark moon, target gains heal', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(moonCycleScenario({ dark_moon_cards: 1, heal: 2 }));

    // Server pushes branch prompt at turn end (branch_select overlay with has_decline)
    await protocolHarness.pushServerMessage(moonCycleBranchPrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    // decline option (index 0) is filtered into cancel dock button;
    // branch1 is the first displayed inline button, prompt-option-branch1
    await page.getByTestId('prompt-option-branch1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });

    // Target selection (target_picker single target: click player-area, auto-submits via submitSelect)
    await protocolHarness.pushServerMessage(moonCycleTargetPrompt());
    await expect(page.getByTestId('decision-overlay')).not.toBeVisible({ timeout: 5000 });
    await page.getByTestId(`player-area-${ENEMY_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('moon cycle: branch 2 - remove heal, gain new moon', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(moonCycleScenario({ dark_moon_cards: 0, heal: 2 }));

    // Server pushes branch prompt at turn end
    await protocolHarness.pushServerMessage(moonCycleBranchPrompt({ branch1: false }));
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    // decline filtered into cancel dock; branch2 is the only displayed inline button
    await page.getByTestId('prompt-option-branch2').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });

    // Branch 2 has no further frontend interaction
  });

  test('moon cycle: decline branch selection', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(moonCycleScenario({ dark_moon_cards: 1, heal: 2 }));

    await protocolHarness.pushServerMessage(moonCycleBranchPrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    // decline is shown as cancel dock button (prompt-cancel-btn), sends Cancel action
    await page.getByTestId('prompt-cancel-btn').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Cancel',
    });
  });

  test('moon cycle: branch 1 unavailable when no dark moon', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(moonCycleScenario({ dark_moon_cards: 0, heal: 2 }));

    await protocolHarness.pushServerMessage(moonCycleBranchPrompt({ branch1: false }));
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    // decline filtered into cancel dock; branch2 is the only displayed inline button
    await page.getByTestId('prompt-option-branch2').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});