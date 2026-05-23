import type { Page } from '@playwright/test';
import { test } from '../../../fixtures/protocolHarness.fixture';
import {
  scarletCovenantScenario,
  scarletCovenantPrompt,
  slaughterFeastScenario,
  slaughterFeastPrompt,
  modestyScenario,
  modestyExtraActionPrompt,
} from '../../../scenarios/redLotusKnight';

async function clickOverlayOption(page: Page, selector: string) {
  const overlay = page.getByTestId('decision-overlay');
  const overlayVisible = await overlay.isVisible({ timeout: 1000 }).catch(() => false);
  if (overlayVisible) {
    await overlay.getByTestId(selector).click();
  } else {
    await page.getByTestId('prompt-dialog').getByTestId(selector).click();
  }
}

test.describe('red lotus knight scarlet covenant protocol harness', () => {
  test('scarlet covenant: confirm +1 damage and heal after', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(scarletCovenantScenario());

    // Server pushes scarlet covenant prompt during attack
    await protocolHarness.pushServerMessage(scarletCovenantPrompt());

    // Click confirm button
    await clickOverlayOption(page, 'prompt-option-confirm');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('scarlet covenant: skip', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(scarletCovenantScenario());

    await protocolHarness.pushServerMessage(scarletCovenantPrompt());

    // Click skip button
    await clickOverlayOption(page, 'prompt-option-skip');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});

test.describe('red lotus knight slaughter feast protocol harness', () => {
  test('slaughter feast: confirm after hit with blood mark', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(slaughterFeastScenario());

    // Server pushes slaughter feast prompt after hit
    await protocolHarness.pushServerMessage(slaughterFeastPrompt());

    // Click confirm button
    await clickOverlayOption(page, 'prompt-option-confirm');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('slaughter feast: skip', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(slaughterFeastScenario());

    await protocolHarness.pushServerMessage(slaughterFeastPrompt());

    // Click skip button
    await clickOverlayOption(page, 'prompt-option-skip');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});

test.describe('red lotus knight modesty protocol harness', () => {
  test('modesty: extra action panel shows attack and magic', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(modestyScenario());

    await protocolHarness.pushServerMessage(modestyExtraActionPrompt());

    await page.getByTestId('decision-overlay').waitFor({ state: 'hidden' });
    await page.getByTestId('action-attack').waitFor({ state: 'visible' });
    await page.getByTestId('action-magic').waitFor({ state: 'visible' });
    await page.getByTestId('action-special').waitFor({ state: 'detached' });
  });
});
