import type { Page } from '@playwright/test';
import { test } from '../../../fixtures/protocolHarness.fixture';
import {
  fraudScenario,
  fraudPickPrompt,
  fraudElementPrompt,
  adventurerParadiseScenario,
  adventurerParadisePrompt,
  adventurerParadiseAllyPickPrompt,
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

// For choose_cards prompts, cards render in hand area with hand-card-${index} testid
async function clickHandCard(page: Page, index: number) {
  await page.getByTestId(`hand-card-${index}`).click();
}

async function selectTarget(page: Page, targetId: string) {
  await page.getByTestId(`player-area-${targetId}`).click();
}

test.describe('adventurer fraud protocol harness', () => {
  // Backend flow: sequential card selection, then element if needed
  // 1. Push fraudPickPrompt(need_count=2)
  // 2. Select card 0 (light element) → expect option_indexes: [0]
  // 3. Push fraudPickPrompt(remaining=1) again
  // 4. Select card 1 (light element, same element) → expect option_indexes: [0, 1]
  // 5. If same element → done, if different → push fraudElementPrompt

  test('fraud: select 2 same element cards (no element prompt needed)', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(fraudScenario());

    // Initial prompt: need to select 2 cards
    await protocolHarness.pushServerMessage(fraudPickPrompt(2));

    // Select first card (index 0)
    await clickHandCard(page, 0);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Push next prompt: need 1 more card
    await protocolHarness.pushServerMessage(fraudPickPrompt(1));

    // Select second card (index 1, same light element as card 0)
    await clickHandCard(page, 1);
    // Backend detects same element → done, no element prompt
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0, 1],
    });
  });

  test('fraud: select 2 different element cards (element prompt needed)', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(fraudScenario());

    // Initial prompt
    await protocolHarness.pushServerMessage(fraudPickPrompt(2));

    // Select first card (index 0, light)
    await clickHandCard(page, 0);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Push next prompt
    await protocolHarness.pushServerMessage(fraudPickPrompt(1));

    // Select second card (index 2, fire - different element)
    await clickHandCard(page, 2);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0, 2],
    });

    // Backend pushes element prompt for different element cards
    await protocolHarness.pushServerMessage(fraudElementPrompt());

    // Select fire element (option.id="Fire", testid="prompt-option-Fire")
    await clickOverlayOption(page, 'prompt-option-Fire');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});

test.describe('adventurer paradise protocol harness', () => {
  // Backend flow: paradise check → ally pick (no remove target)
  test('adventurer paradise: activate and select ally', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(adventurerParadiseScenario());

    // Server pushes paradise check prompt
    await protocolHarness.pushServerMessage(adventurerParadisePrompt());

    // Click "yes" option (option.id="yes", testid="prompt-option-yes")
    await clickOverlayOption(page, 'prompt-option-yes');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Server pushes ally pick prompt
    await protocolHarness.pushServerMessage(adventurerParadiseAllyPickPrompt());

    // Select ally target
    await selectTarget(page, ALLY_PLAYER_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('adventurer paradise: skip', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(adventurerParadiseScenario());

    await protocolHarness.pushServerMessage(adventurerParadisePrompt());

    // Click "no" option (option.id="no", testid="prompt-option-no")
    await clickOverlayOption(page, 'prompt-option-no');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});