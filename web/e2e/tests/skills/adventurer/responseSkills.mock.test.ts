import type { Page } from '@playwright/test';
import { test } from '../../../fixtures/protocolHarness.fixture';
import {
  fraudScenario,
  fraudPickPrompt,
  fraudElementPrompt,
  fraudTargetPrompt,
  adventurerParadiseScenario,
  adventurerParadisePrompt,
  adventurerParadiseAllyPickPrompt,
  ALLY_PLAYER_ID,
  ENEMY_PLAYER_ID,
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

async function confirmCardSelection(page: Page) {
  await page.getByTestId('card-picker-prompt').getByTestId('prompt-confirm-btn').click();
}

async function selectTarget(page: Page, targetId: string) {
  await page.getByTestId(`player-area-${targetId}`).click();
}

test.describe('adventurer fraud protocol harness', () => {
  // Backend flow: multi-select same-element cards, then element if needed, then target.
  // 1. Push fraudPickPrompt(min=2,max=3)
  // 2. Select 2 or 3 same-element cards in hand
  // 3. Click confirm once → expect one Select action with all card_ids
  // 4. Click enemy avatar on the target prompt → expect one Select action for the target

  test('fraud: select 2 same element cards once, choose attack element, then target', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(fraudScenario());

    await protocolHarness.pushServerMessage(fraudPickPrompt());

    await clickHandCard(page, 0);
    await clickHandCard(page, 1);
    await confirmCardSelection(page);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      card_ids: ['adv-attack-1', 'adv-attack-2'],
    });

    await protocolHarness.pushServerMessage(fraudElementPrompt());

    await clickOverlayOption(page, 'prompt-option-Fire');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });

    await protocolHarness.pushServerMessage(fraudTargetPrompt());

    await selectTarget(page, ENEMY_PLAYER_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('fraud: select 3 same element cards once for dark attack, then target', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(fraudScenario());

    await protocolHarness.pushServerMessage(fraudPickPrompt());

    await clickHandCard(page, 0);
    await clickHandCard(page, 1);
    await clickHandCard(page, 2);
    await confirmCardSelection(page);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      card_ids: ['adv-attack-1', 'adv-attack-2', 'adv-light-magic'],
    });

    await protocolHarness.pushServerMessage(fraudTargetPrompt());

    await selectTarget(page, ENEMY_PLAYER_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
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
