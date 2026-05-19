import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  holySwordScenario,
  holySwordThirdAttackPrompt,
  holySwordDiscardPrompt,
  galeSkillScenario,
  windBladeScenario,
  windBladeShieldPrompt,
} from '../../../scenarios/windSwordSaint';

async function selectHandCards(page: import('@playwright/test').Page, indices: number[]) {
  for (const index of indices) {
    const card = page.getByTestId(`hand-card-${index}`);
    await card.scrollIntoViewIfNeeded();
    await card.click();
  }
  await page.getByTestId('prompt-confirm-btn').click();
}

test.describe('wind sword saint holy sword protocol harness', () => {
  test('holy sword: third attack triggers forced hit', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(holySwordScenario({ attackCount: 3 }));

    // Server pushes holy sword prompt after third attack
    await protocolHarness.pushServerMessage(holySwordThirdAttackPrompt());

    // Select X=2 (numeric: numeric-option-2 inside decision-overlay)
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('decision-overlay').getByTestId('numeric-option-2').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [2],
    });

    // Server pushes discard prompt
    await protocolHarness.pushServerMessage(holySwordDiscardPrompt(2));

    // Select 2 cards to discard (card_picker: hand cards + confirm)
    await selectHandCards(page, [0, 1]);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      card_ids: ['wss-wind-atk1', 'wss-wind-atk2'],
    });
  });

  test('holy sword: select X=1', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(holySwordScenario({ attackCount: 3 }));

    await protocolHarness.pushServerMessage(holySwordThirdAttackPrompt());

    // Select X=1 (numeric: numeric-option-1 inside decision-overlay)
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('decision-overlay').getByTestId('numeric-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });

    await protocolHarness.pushServerMessage(holySwordDiscardPrompt(1));

    await selectHandCards(page, [0]);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      card_ids: ['wss-wind-atk1'],
    });
  });

  test('holy sword: select X=3', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(holySwordScenario({ attackCount: 3 }));

    await protocolHarness.pushServerMessage(holySwordThirdAttackPrompt());

    // Select X=3 (numeric: numeric-option-3 inside decision-overlay)
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('decision-overlay').getByTestId('numeric-option-3').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [3],
    });

    await protocolHarness.pushServerMessage(holySwordDiscardPrompt(3));

    await selectHandCards(page, [0, 1, 2]);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      card_ids: ['wss-wind-atk1', 'wss-wind-atk2', 'wss-wind-atk3'],
    });
  });
});

test.describe('wind sword saint gale skill protocol harness', () => {
  test('gale skill: auto triggers extra attack action', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(galeSkillScenario());

    // Just verify scenario loads - 疾风技 is auto trigger
    await page.getByTestId('game-board').waitFor({ state: 'visible' });
  });
});

test.describe('wind sword saint wind blade protocol harness', () => {
  test('wind blade: shield target triggers bypass', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(windBladeScenario());

    // Server pushes wind blade prompt when target has shield
    await protocolHarness.pushServerMessage(windBladeShieldPrompt());

    // Click confirm (branch_select: branch-option-0 inside decision-overlay)
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('decision-overlay').getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });
});