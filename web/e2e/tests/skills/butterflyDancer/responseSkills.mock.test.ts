import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  reverseScenario,
  pilgrimagePickPrompt,
  poisonPickPrompt,
  mirrorPairPrompt,
} from '../../../scenarios/butterflyDancer';

test.describe('butterfly dancer pilgrimage protocol harness', () => {
  test('pilgrimage: skip (不发动)', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(reverseScenario());

    // Server pushes pilgrimage pick prompt (triggered by damage before apply hook)
    await protocolHarness.pushServerMessage(pilgrimagePickPrompt());

    await expect(page.getByTestId('decision-overlay')).not.toBeVisible();
    await expect(page.getByRole('button', { name: '不发动' })).toBeVisible();
    await expect(page.getByText('请在扩展区点击对应的茧完成选择')).toBeVisible();
    await expect(page.getByText('移除茧[0]')).not.toBeVisible();
    await page.getByRole('button', { name: '不发动' }).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('pilgrimage: remove cocoon to block damage', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(reverseScenario());

    // Server pushes pilgrimage pick prompt
    await protocolHarness.pushServerMessage(pilgrimagePickPrompt());

    await expect(page.getByTestId('decision-overlay')).not.toBeVisible();
    await expect(page.getByText('请在扩展区点击对应的茧完成选择')).toBeVisible();
    await expect(page.getByText('移除茧[0]')).not.toBeVisible();
    await page.getByTestId('cover-card-0').click();
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

    await expect(page.getByTestId('decision-overlay')).not.toBeVisible();
    await expect(page.getByRole('button', { name: '不发动' })).toBeVisible();
    await expect(page.getByText('请在扩展区点击对应的茧完成选择')).toBeVisible();
    await expect(page.getByText('移除茧[0]')).not.toBeVisible();
    await page.getByRole('button', { name: '不发动' }).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('poison: remove cocoon to boost magic damage', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(reverseScenario());

    // Server pushes poison pick prompt
    await protocolHarness.pushServerMessage(poisonPickPrompt());

    await expect(page.getByTestId('decision-overlay')).not.toBeVisible();
    await expect(page.getByText('请在扩展区点击对应的茧完成选择')).toBeVisible();
    await expect(page.getByText('移除茧[0]')).not.toBeVisible();
    await page.getByTestId('cover-card-0').click();
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
    const overlay = page.getByTestId('decision-overlay');
    const overlayVisible = await overlay.isVisible({ timeout: 1000 }).catch(() => false);
    if (overlayVisible) {
      await overlay.getByTestId('branch-option-0').click();
    } else {
      await page.getByTestId('prompt-dialog').getByTestId('branch-option-0').click();
    }
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
    const overlay = page.getByTestId('decision-overlay');
    const overlayVisible = await overlay.isVisible({ timeout: 1000 }).catch(() => false);
    if (overlayVisible) {
      await overlay.getByTestId('branch-option-1').click();
    } else {
      await page.getByTestId('prompt-dialog').getByTestId('branch-option-1').click();
    }
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});
