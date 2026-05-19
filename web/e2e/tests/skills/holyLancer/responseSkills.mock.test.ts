import type { Page } from '@playwright/test';
import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  skySpearScenario,
  skySpearBeforeAttackPrompt,
  earthSpearScenario,
  earthSpearAfterHitPrompt,
} from '../../../scenarios/holyLancer';

test.describe('holy lancer sky spear protocol harness', () => {
  test('sky spear: confirm before attack', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(skySpearScenario({ heal: 3 }));

    // Server pushes sky spear prompt before attack
    await protocolHarness.pushServerMessage(skySpearBeforeAttackPrompt());

    // Click confirm button (branch_select: prompt-option-confirm)
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('prompt-option-confirm').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('sky spear: skip before attack', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(skySpearScenario({ heal: 3 }));

    await protocolHarness.pushServerMessage(skySpearBeforeAttackPrompt());

    // Click skip button (branch_select: prompt-option-skip)
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('prompt-option-skip').click();
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

    // Select X=2 (numeric: numeric-option-2)
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('numeric-option-2').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [2],
    });
  });

  test('earth spear: select X=0 (skip)', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(earthSpearScenario({ heal: 4 }));

    await protocolHarness.pushServerMessage(earthSpearAfterHitPrompt(4));

    // Select X=0 (numeric: numeric-option with Chinese button_label '不发动')
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('numeric-option-不发动').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('earth spear: select max X=4', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(earthSpearScenario({ heal: 5 }));

    await protocolHarness.pushServerMessage(earthSpearAfterHitPrompt(5));

    // Select X=4 (numeric: numeric-option-4)
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('numeric-option-4').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [4],
    });
  });
});