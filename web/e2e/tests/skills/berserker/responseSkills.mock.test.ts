import type { Page } from '@playwright/test';
import { test } from '../../../fixtures/protocolHarness.fixture';
import {
  tearScenario,
  tearHitPrompt,
} from '../../../scenarios/berserker';

async function clickOverlayOption(page: Page, selector: string) {
  const overlay = page.getByTestId('decision-overlay');
  const overlayVisible = await overlay.isVisible({ timeout: 1000 }).catch(() => false);
  if (overlayVisible) {
    await overlay.getByTestId(selector).click();
  } else {
    await page.getByTestId('prompt-dialog').getByTestId(selector).click();
  }
}

test.describe('berserker tear protocol harness', () => {
  test('tear: confirm on hit with gem', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(tearScenario({ gem: 1 }));

    // Server pushes tear response prompt after attack hits
    await protocolHarness.pushServerMessage(tearHitPrompt());

    // Click confirm button
    await clickOverlayOption(page, 'prompt-option-confirm');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0], // confirm is first option
    });
  });

  test('tear: skip on hit', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(tearScenario({ gem: 1 }));

    await protocolHarness.pushServerMessage(tearHitPrompt());

    // Click skip button
    await clickOverlayOption(page, 'prompt-option-skip');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1], // skip is second option
    });
  });
});

// Note: bloodyRoar and bloodShadow skills are auto-triggered (no popup prompts)
// The backend automatically executes these skills when conditions are met