import type { Page } from '@playwright/test';
import { test } from '../../../fixtures/protocolHarness.fixture';
import {
  fraudScenario,
  fraudDiscardCountPrompt,
  fraudDiscard2Prompt,
  fraudElementSelectPrompt,
  fraudDiscard3Prompt,
  adventurerParadiseScenario,
  adventurerParadisePrompt,
  adventurerParadiseTransferTargetPrompt,
  adventurerParadiseRemoveTargetPrompt,
  ENEMY_PLAYER_ID,
  ALLY_PLAYER_ID,
} from '../../../scenarios/adventurer';

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

test.describe('adventurer fraud protocol harness', () => {
  test('fraud: discard 2 cards and select element', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(fraudScenario());

    // Server pushes fraud discard count prompt during purchase
    await protocolHarness.pushServerMessage(fraudDiscardCountPrompt());

    // Select discard 2 option
    await clickOverlayOption(page, 'prompt-option-2');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Server pushes discard 2 selection
    await protocolHarness.pushServerMessage(fraudDiscard2Prompt());

    // Select 2 cards to discard
    await clickOverlayOption(page, 'prompt-option-adv-attack-1');
    await clickOverlayOption(page, 'prompt-option-adv-attack-2');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0, 1],
    });

    // Server pushes element selection
    await protocolHarness.pushServerMessage(fraudElementSelectPrompt());

    // Select fire element
    await clickOverlayOption(page, 'prompt-option-Fire');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('fraud: discard 3 cards for dark element', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(fraudScenario());

    await protocolHarness.pushServerMessage(fraudDiscardCountPrompt());

    // Select discard 3 option (auto dark element)
    await clickOverlayOption(page, 'prompt-option-3');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });

    // Server pushes discard 3 selection
    await protocolHarness.pushServerMessage(fraudDiscard3Prompt());

    // Select 3 cards to discard
    await clickOverlayOption(page, 'prompt-option-adv-attack-1');
    await clickOverlayOption(page, 'prompt-option-adv-attack-2');
    await clickOverlayOption(page, 'prompt-option-adv-magic-1');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0, 1, 2],
    });
  });
});

test.describe('adventurer adventurer paradise protocol harness', () => {
  test('adventurer paradise: transfer energy to ally and remove enemy energy', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(adventurerParadiseScenario());

    // Server pushes adventurer paradise prompt during refine
    await protocolHarness.pushServerMessage(adventurerParadisePrompt());

    // Click confirm button
    await clickOverlayOption(page, 'prompt-option-confirm');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Server pushes transfer target selection
    await protocolHarness.pushServerMessage(adventurerParadiseTransferTargetPrompt());

    // Select ally target
    await selectTarget(page, ALLY_PLAYER_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Server pushes remove target selection
    await protocolHarness.pushServerMessage(adventurerParadiseRemoveTargetPrompt());

    // Select enemy target
    await selectTarget(page, ENEMY_PLAYER_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('adventurer paradise: skip', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(adventurerParadiseScenario());

    await protocolHarness.pushServerMessage(adventurerParadisePrompt());

    // Click skip button
    await clickOverlayOption(page, 'prompt-option-skip');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});