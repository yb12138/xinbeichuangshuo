import type { Page } from '@playwright/test';
import { test } from '../../../fixtures/protocolHarness.fixture';
import {
  angelSongScenario,
  angelSongBeforeTurnPrompt,
  angelSongFieldSelectPrompt,
  godProtectionScenario,
  godProtectionPrompt,
} from '../../../scenarios/angel';

async function clickOverlayOption(page: Page, selector: string) {
  const overlay = page.getByTestId('decision-overlay');
  const overlayVisible = await overlay.isVisible({ timeout: 1000 }).catch(() => false);
  if (overlayVisible) {
    await overlay.getByTestId(selector).click();
  } else {
    await page.getByTestId('prompt-dialog').getByTestId(selector).click();
  }
}

test.describe('angel song protocol harness', () => {
  test('angel song: confirm before turn with crystal', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(angelSongScenario({ crystal: 1 }));

    // Server pushes angel song confirm before turn
    await protocolHarness.pushServerMessage(angelSongBeforeTurnPrompt());

    // Click confirm button
    await clickOverlayOption(page, 'prompt-option-confirm');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Server pushes field selection
    await protocolHarness.pushServerMessage(angelSongFieldSelectPrompt());

    // Select field effect
    await clickOverlayOption(page, 'prompt-option-enemy_shield');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('angel song: skip before turn', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(angelSongScenario({ crystal: 1 }));

    await protocolHarness.pushServerMessage(angelSongBeforeTurnPrompt());

    // Click skip button
    await clickOverlayOption(page, 'prompt-option-skip');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});

test.describe('angel god protection protocol harness', () => {
  test('god protection: select X=2 crystals', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(godProtectionScenario({ crystal: 3 }));

    // Server pushes protection prompt (max X = crystal count)
    await protocolHarness.pushServerMessage(godProtectionPrompt(3));

    // Select X=2
    await page.getByTestId('decision-overlay').getByTestId('numeric-option-2').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1], // X=2 is second option (X=1 is first)
    });
  });

  test('god protection: select X=1 crystal', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(godProtectionScenario({ crystal: 3 }));

    await protocolHarness.pushServerMessage(godProtectionPrompt(3));

    // Select X=1
    await page.getByTestId('decision-overlay').getByTestId('numeric-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('god protection: select max X', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(godProtectionScenario({ crystal: 5 }));

    await protocolHarness.pushServerMessage(godProtectionPrompt(5));

    // Select X=5 (max)
    await page.getByTestId('decision-overlay').getByTestId('numeric-option-5').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [4],
    });
  });
});
