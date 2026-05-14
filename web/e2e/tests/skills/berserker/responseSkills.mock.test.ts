import type { Page } from '@playwright/test';
import { test } from '../../../fixtures/protocolHarness.fixture';
import {
  tearScenario,
  tearHitPrompt,
  bloodyRoarScenario,
  bloodyRoarTargetHeal2Prompt,
  bloodShadowScenario,
  bloodShadowHand2Prompt,
  bloodShadowHand3Prompt,
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

test.describe('berserker bloody roar protocol harness', () => {
  test('bloody roar: target heal=2 triggers forced hit', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(bloodyRoarScenario());

    // Server pushes bloody roar prompt when target has 2 heal
    await protocolHarness.pushServerMessage(bloodyRoarTargetHeal2Prompt());

    // Click confirm
    await clickOverlayOption(page, 'prompt-option-confirm');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });
});

test.describe('berserker blood shadow blade protocol harness', () => {
  test('blood shadow: enemy hand=2 triggers +2 damage', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(bloodShadowScenario({ enemyHandCount: 2 }));

    // Server pushes prompt after hit when enemy hand is 2
    await protocolHarness.pushServerMessage(bloodShadowHand2Prompt());

    // Click confirm
    await clickOverlayOption(page, 'prompt-option-confirm');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('blood shadow: enemy hand=3 triggers +1 damage', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(bloodShadowScenario({ enemyHandCount: 3 }));

    // Server pushes prompt after hit when enemy hand is 3
    await protocolHarness.pushServerMessage(bloodShadowHand3Prompt());

    // Click confirm
    await clickOverlayOption(page, 'prompt-option-confirm');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });
});