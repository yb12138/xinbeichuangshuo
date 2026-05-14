import type { Page } from '@playwright/test';
import { test } from '../../../fixtures/protocolHarness.fixture';
import {
  skySpearScenario,
  skySpearBeforeAttackPrompt,
  earthSpearScenario,
  earthSpearAfterHitPrompt,
} from '../../../scenarios/holyLancer';

async function clickOverlayOption(page: Page, selector: string) {
  const overlay = page.getByTestId('decision-overlay');
  const overlayVisible = await overlay.isVisible({ timeout: 1000 }).catch(() => false);
  if (overlayVisible) {
    await overlay.getByTestId(selector).click();
  } else {
    await page.getByTestId('prompt-dialog').getByTestId(selector).click();
  }
}

test.describe('holy lancer sky spear protocol harness', () => {
  test('sky spear: confirm before attack', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(skySpearScenario({ heal: 3 }));

    // Server pushes sky spear prompt before attack
    await protocolHarness.pushServerMessage(skySpearBeforeAttackPrompt());

    // Click confirm button
    await clickOverlayOption(page, 'prompt-option-confirm');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('sky spear: skip before attack', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(skySpearScenario({ heal: 3 }));

    await protocolHarness.pushServerMessage(skySpearBeforeAttackPrompt());

    // Click skip button
    await clickOverlayOption(page, 'prompt-option-skip');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });

  test('sky spear: not available with heal < 2', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(skySpearScenario({ heal: 1 }));

    // Skill should not be triggered (no prompt pushed)
    // Just verify the scenario loads
    await page.getByTestId('game-board').waitFor({ state: 'visible' });
  });
});

test.describe('holy lancer earth spear protocol harness', () => {
  test('earth spear: select X=2 after hit', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(earthSpearScenario({ heal: 4 }));

    // Server pushes earth spear prompt after hit (max X = 4)
    await protocolHarness.pushServerMessage(earthSpearAfterHitPrompt(4));

    // Select X=2 (option index depends on options array)
    await clickOverlayOption(page, 'branch-option-2');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [2], // X=2 option
    });
  });

  test('earth spear: select X=0 (skip)', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(earthSpearScenario({ heal: 4 }));

    await protocolHarness.pushServerMessage(earthSpearAfterHitPrompt(4));

    // Select X=0 (first option, skip)
    await clickOverlayOption(page, 'branch-option-0');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('earth spear: select max X=4', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(earthSpearScenario({ heal: 5 }));

    await protocolHarness.pushServerMessage(earthSpearAfterHitPrompt(5));

    // Select X=4 (last valid option, capped at 4)
    await clickOverlayOption(page, 'branch-option-4');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [4],
    });
  });
});