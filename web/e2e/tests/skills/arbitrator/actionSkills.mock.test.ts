import type { Page } from '@playwright/test';
import { test } from '../../../fixtures/protocolHarness.fixture';
import {
  ARBITRATOR_PLAYER_ID,
  ENEMY_PLAYER_ID,
  ritualInterruptScenario,
  ritualInterruptPrompt,
  doomJudgmentScenario,
  doomJudgmentTargetPrompt,
  doomJudgmentForcePrompt,
  arbitrationRitualScenario,
  arbitrationRitualPrompt,
  judgmentBalanceScenario,
  judgmentBalanceBranchPrompt,
} from '../../../scenarios/arbitrator';

async function clickOverlayOption(page: Page, selector: string) {
  const overlay = page.getByTestId('decision-overlay');
  const overlayVisible = await overlay.isVisible({ timeout: 1000 }).catch(() => false);
  if (overlayVisible) {
    await overlay.getByTestId(selector).click();
  } else {
    await page.getByTestId('prompt-dialog').getByTestId(selector).click();
  }
}

async function selectTarget(page: Page, targetId: string) {
  await page.getByTestId(`player-area-${targetId}`).click();
}

test.describe('arbitrator ritual interrupt protocol harness', () => {
  test('ritual interrupt: confirm to exit judgment form', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(ritualInterruptScenario());

    // Server pushes ritual interrupt prompt
    await protocolHarness.pushServerMessage(ritualInterruptPrompt());

    // Click confirm button
    await clickOverlayOption(page, 'prompt-option-confirm');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('ritual interrupt: skip', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(ritualInterruptScenario());

    await protocolHarness.pushServerMessage(ritualInterruptPrompt());

    // Click skip button
    await clickOverlayOption(page, 'prompt-option-skip');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});

test.describe('arbitrator doom judgment protocol harness', () => {
  test('doom judgment: select target enemy', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(doomJudgmentScenario({ judgmentCount: 3 }));

    // Server pushes target selection prompt
    await protocolHarness.pushServerMessage(doomJudgmentTargetPrompt());

    // Select enemy as target
    await selectTarget(page, ENEMY_PLAYER_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('doom judgment: select self as target', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(doomJudgmentScenario({ judgmentCount: 3 }));

    await protocolHarness.pushServerMessage(doomJudgmentTargetPrompt());

    // Select self as target
    await selectTarget(page, ARBITRATOR_PLAYER_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [2],
    });
  });

  test('doom judgment: forced when judgment at max', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(doomJudgmentScenario({ judgmentCount: 5, force: true }));

    // Server pushes forced prompt (no skip option)
    await protocolHarness.pushServerMessage(doomJudgmentForcePrompt());

    // Must select a target
    await selectTarget(page, ENEMY_PLAYER_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });
});

test.describe('arbitrator arbitration ritual protocol harness', () => {
  test('arbitration ritual: confirm with gem', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(arbitrationRitualScenario({ gem: 1 }));

    // Server pushes arbitration ritual prompt
    await protocolHarness.pushServerMessage(arbitrationRitualPrompt());

    // Click confirm button
    await clickOverlayOption(page, 'prompt-option-confirm');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('arbitration ritual: skip', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(arbitrationRitualScenario({ gem: 1 }));

    await protocolHarness.pushServerMessage(arbitrationRitualPrompt());

    // Click skip button
    await clickOverlayOption(page, 'prompt-option-skip');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});

test.describe('arbitrator judgment balance protocol harness', () => {
  test('judgment balance: discard all hand', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(judgmentBalanceScenario());

    // Server pushes branch selection prompt
    await protocolHarness.pushServerMessage(judgmentBalanceBranchPrompt());

    // Click discard all option
    await clickOverlayOption(page, 'prompt-option-discard_all');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('judgment balance: fill hand to max', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(judgmentBalanceScenario());

    await protocolHarness.pushServerMessage(judgmentBalanceBranchPrompt());

    // Click fill hand option
    await clickOverlayOption(page, 'prompt-option-fill_hand');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});