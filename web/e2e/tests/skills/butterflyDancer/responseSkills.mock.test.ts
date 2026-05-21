import type { Page } from '@playwright/test';
import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  reverseScenario,
  pilgrimageConfirmPrompt,
  pilgrimagePickPrompt,
  poisonPickPrompt,
  mirrorPairPrompt,
} from '../../../scenarios/butterflyDancer';

async function clickOverlayOption(page: Page, selector: string) {
  const overlay = page.getByTestId('decision-overlay');
  const overlayVisible = await overlay.isVisible({ timeout: 1000 }).catch(() => false);
  if (overlayVisible) {
    await overlay.getByTestId(selector).click();
  } else {
    await page.getByTestId('prompt-dialog').getByTestId(selector).click();
  }
}

test.describe('butterfly dancer pilgrimage protocol harness', () => {
  test('pilgrimage: decline confirm prompt via cancel control', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(reverseScenario());

    await protocolHarness.pushServerMessage(pilgrimageConfirmPrompt());

    const overlay = page.getByTestId('decision-overlay');
    await expect(overlay).toBeVisible();
    await expect(overlay.getByTestId('prompt-option-0')).toHaveText('发动');
    await expect(overlay.getByTestId('prompt-option-1')).toHaveCount(0);
    await expect(overlay.getByTestId('prompt-cancel-btn')).toHaveText('取消');
    await expect(overlay.getByText('不发动')).toHaveCount(0);

    await page.getByTestId('decision-overlay').getByTestId('prompt-cancel-btn').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Cancel',
    });
  });

  test('pilgrimage: skip (不发动)', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(reverseScenario());

    // Server pushes pilgrimage pick prompt (triggered by damage before apply hook)
    await protocolHarness.pushServerMessage(pilgrimagePickPrompt());

    await clickOverlayOption(page, 'branch-option-0');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('pilgrimage: remove cocoon to block damage', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(reverseScenario());

    // Server pushes pilgrimage pick prompt
    await protocolHarness.pushServerMessage(pilgrimagePickPrompt());

    await clickOverlayOption(page, 'branch-option-1');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});

test.describe('butterfly dancer poison protocol harness', () => {
  test('poison: skip (不发动)', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(reverseScenario());

    // Server pushes poison pick prompt (triggered by magic damage)
    await protocolHarness.pushServerMessage(poisonPickPrompt());

    await clickOverlayOption(page, 'branch-option-0');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('poison: remove cocoon to boost magic damage', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(reverseScenario());

    // Server pushes poison pick prompt
    await protocolHarness.pushServerMessage(poisonPickPrompt());

    await clickOverlayOption(page, 'branch-option-1');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});

test.describe('butterfly dancer mirror protocol harness', () => {
  test('mirror: skip (不发动)', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(reverseScenario());

    // Server pushes mirror pair prompt (triggered by 2+ magic damage)
    await protocolHarness.pushServerMessage(mirrorPairPrompt());

    // Click "不发动" button
    await clickOverlayOption(page, 'branch-option-0');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('mirror: remove cocoon pair to nullify damage', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(reverseScenario());

    // Server pushes mirror pair prompt (labels use "茧：" not "茧[N]", so options render as buttons)
    await protocolHarness.pushServerMessage(mirrorPairPrompt());

    // Click the pair option button (option index 1 = "移除并展示：...")
    await clickOverlayOption(page, 'branch-option-1');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});
