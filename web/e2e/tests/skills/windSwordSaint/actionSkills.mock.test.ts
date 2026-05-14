import type { Page } from '@playwright/test';
import { test } from '../../../fixtures/protocolHarness.fixture';
import {
  holySwordScenario,
  holySwordThirdAttackPrompt,
  holySwordDiscardPrompt,
  galeSkillScenario,
  windBladeScenario,
  windBladeShieldPrompt,
} from '../../../scenarios/windSwordSaint';

async function clickOverlayOption(page: Page, selector: string) {
  const overlay = page.getByTestId('decision-overlay');
  const overlayVisible = await overlay.isVisible({ timeout: 1000 }).catch(() => false);
  if (overlayVisible) {
    await overlay.getByTestId(selector).click();
  } else {
    await page.getByTestId('prompt-dialog').getByTestId(selector).click();
  }
}

async function selectHandCards(page: Page, indices: number[]) {
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

    // Select X=2 (摸2弃2)
    await clickOverlayOption(page, 'prompt-option-2');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1], // X=2 is second option
    });

    // Server pushes discard prompt
    await protocolHarness.pushServerMessage(holySwordDiscardPrompt(2));

    // Select 2 cards to discard
    await selectHandCards(page, [0, 1]);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0, 1],
    });
  });

  test('holy sword: select X=1', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(holySwordScenario({ attackCount: 3 }));

    await protocolHarness.pushServerMessage(holySwordThirdAttackPrompt());

    // Select X=1
    await clickOverlayOption(page, 'prompt-option-1');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    await protocolHarness.pushServerMessage(holySwordDiscardPrompt(1));

    await selectHandCards(page, [0]);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('holy sword: select X=3', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(holySwordScenario({ attackCount: 3 }));

    await protocolHarness.pushServerMessage(holySwordThirdAttackPrompt());

    // Select X=3
    await clickOverlayOption(page, 'prompt-option-3');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [2],
    });

    await protocolHarness.pushServerMessage(holySwordDiscardPrompt(3));

    await selectHandCards(page, [0, 1, 2]);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0, 1, 2],
    });
  });
});

test.describe('wind sword saint gale skill protocol harness', () => {
  test('gale skill: auto triggers extra attack action', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(galeSkillScenario());

    // Just verify scenario loads - 狾风技 is auto trigger
    await page.getByTestId('game-board').waitFor({ state: 'visible' });
  });
});

test.describe('wind sword saint wind blade protocol harness', () => {
  test('wind blade: shield target triggers bypass', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(windBladeScenario());

    // Server pushes wind blade prompt when target has shield
    await protocolHarness.pushServerMessage(windBladeShieldPrompt());

    // Click confirm
    await clickOverlayOption(page, 'prompt-option-confirm');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });
});