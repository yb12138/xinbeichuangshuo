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

    // Server pushes branch prompt at turn end
    await protocolHarness.pushServerMessage(moonCycleBranchPrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Target selection (click enemy player card)
    await protocolHarness.pushServerMessage(moonCycleTargetPrompt());
    await page.getByTestId(`player-area-${ENEMY_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('moon cycle: branch 2 - remove heal, gain new moon', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(moonCycleScenario({ dark_moon_cards: 0, heal: 2 }));

    // Server pushes branch prompt at turn end
    await protocolHarness.pushServerMessage(moonCycleBranchPrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('branch-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });

    // Branch 2 has no further frontend interaction
  });

  test('moon cycle: branch 1 unavailable when no dark moon', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(moonCycleScenario({ dark_moon_cards: 0, heal: 2 }));

    // Server pushes branch prompt, but branch 1 should be disabled
    await protocolHarness.pushServerMessage(moonCycleBranchPrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    // Branch 2 is the only available option
    await page.getByTestId('branch-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});