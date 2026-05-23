import type { Page } from '@playwright/test';
import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  magicBulletControlScenario,
  magicBulletDirectionPrompt,
  magicBulletFusionScenario,
  magicExplosionEnemyDiscardScenario,
  magicExplosionEnemy2DiscardScenario,
  magicExplosionEnemyDiscardPrompt,
  magicExplosionEnemy2DiscardPrompt,
} from '../../../scenarios/magicGirl';

async function selectHandCard(page: Page, index: number) {
  const card = page.getByTestId(`hand-card-${index}`);
  await card.scrollIntoViewIfNeeded();
  await card.click();
}

test.describe('magic girl magic bullet control protocol harness', () => {
  test('magic bullet control: forward direction', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(magicBulletControlScenario());

    // Server pushes direction choice after using magic bullet
    await protocolHarness.pushServerMessage(magicBulletDirectionPrompt());

    // Click forward option on the decision overlay
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('decision-overlay').getByTestId('prompt-option-forward').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('magic bullet control: reverse direction', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(magicBulletControlScenario());

    await protocolHarness.pushServerMessage(magicBulletDirectionPrompt());

    // Click reverse option on the decision overlay
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('decision-overlay').getByTestId('prompt-option-reverse').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});

test.describe('magic girl magic bullet fusion protocol harness', () => {
  test('magic bullet fusion: scenario setup', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(magicBulletFusionScenario());
    // Magic bullet fusion is a passive response skill
    // Testing scenario setup only
  });
});

test.describe('magic girl magic explosion enemy discard protocol harness', () => {
  test('magic explosion: enemy discards magic card', async ({ page, protocolHarness }) => {
    // Boot game from enemy perspective so enemy sees their own prompt
    await protocolHarness.bootGame(magicExplosionEnemyDiscardScenario());

    // Server pushes discard prompt to enemy (who is now "me")
    await protocolHarness.pushServerMessage(magicExplosionEnemyDiscardPrompt());

    // Enemy selects their hand card to discard
    await selectHandCard(page, 0);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      card_ids: ['en-card-1'],
    });
  });

  test('magic explosion: enemy refuses to discard', async ({ page, protocolHarness }) => {
    // Boot game from enemy E2 perspective
    await protocolHarness.bootGame(magicExplosionEnemy2DiscardScenario());

    // Server pushes discard prompt to enemy E2 (who is now "me")
    await protocolHarness.pushServerMessage(magicExplosionEnemy2DiscardPrompt());

    // Enemy skips discard (takes damage instead)
    // card_picker with min=0 allows empty selection via confirm
    await page.getByTestId('prompt-confirm-btn').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      card_ids: [],
    });
  });
});