import type { Page } from '@playwright/test';
import { expect, test } from '../../../fixtures/protocolHarness.fixture';
import {
  HEROIC_SPIRIT_RAGE_SUPPRESS_ID,
  HEROIC_SPIRIT_SEAL_STRIKE_ID,
  HEROIC_SPIRIT_MAGIC_FUSION_ID,
  HEROIC_SPIRIT_DOUBLE_ECHO_ID,
  rageSuppressScenario,
  rageSuppressPrompt,
  rageSuppressSealSelectPrompt,
  sealStrikeScenario,
  sealStrikePrompt,
  sealStrikeDiscardPrompt,
  sealStrikeSealSelectPrompt,
  magicFusionScenario,
  magicFusionPrompt,
  magicFusionDiscardPrompt,
  magicFusionSealSelectPrompt,
  sealSuppressComboScenario,
  sealSuppressComboPrompt,
  doubleEchoScenario,
  doubleEchoPrompt,
} from '../../../scenarios/heroicSpirit';

async function clickOverlayOption(page: Page, selector: string) {
  const overlay = page.getByTestId('decision-overlay');
  const overlayVisible = await overlay.isVisible({ timeout: 1000 }).catch(() => false);
  if (overlayVisible) {
    await overlay.getByTestId(selector).click();
  } else {
    await page.getByTestId('prompt-dialog').getByTestId(selector).click();
  }
}

async function clickSkillChoiceOption(page: Page, index: number) {
  await page
    .getByTestId('skill-branch-overlay')
    .getByTestId(`branch-option-${index}`)
    .click();
}

async function clickSkillChoiceSkip(page: Page) {
  await page
    .getByTestId('skill-branch-overlay')
    .getByTestId('prompt-option-skip')
    .click();
}

test.describe('heroic spirit rage suppress protocol harness', () => {
  test('rage suppress: confirm forced hit on miss', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(rageSuppressScenario());

    // Server pushes rage suppress prompt on miss
    await protocolHarness.pushServerMessage(rageSuppressPrompt());

    // Click activate option
    await clickSkillChoiceOption(page, 0);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Server pushes seal selection
    await protocolHarness.pushServerMessage(rageSuppressSealSelectPrompt());

    // Select seal to flip
    const decisionRoot = page.locator('.overlay-panel-root--decision');
    await expect(decisionRoot).toBeVisible();
    await decisionRoot.getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('rage suppress: skip', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(rageSuppressScenario());

    await protocolHarness.pushServerMessage(rageSuppressPrompt());

    // Click skip button
    await clickSkillChoiceSkip(page);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});

test.describe('heroic spirit seal strike protocol harness', () => {
  test('seal strike: confirm on hit', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(sealStrikeScenario());

    // Server pushes seal strike prompt on hit
    await protocolHarness.pushServerMessage(sealStrikePrompt());

    // Click activate option
    await clickSkillChoiceOption(page, 0);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Server pushes discard selection (same element as attack)
    await protocolHarness.pushServerMessage(sealStrikeDiscardPrompt('Fire'));

    // Select fire card to discard
    await page.getByTestId('hand-card-0').click();
    await page.getByTestId('prompt-confirm-btn').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      card_ids: ['hs-fire-attack'],
    });

    // Server pushes seal selection
    await protocolHarness.pushServerMessage(sealStrikeSealSelectPrompt());

    // Select seal to flip
    await clickOverlayOption(page, 'branch-option-0');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('seal strike: skip', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(sealStrikeScenario());

    await protocolHarness.pushServerMessage(sealStrikePrompt());

    // Click skip button
    await clickSkillChoiceSkip(page);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});

test.describe('heroic spirit magic fusion protocol harness', () => {
  test('magic fusion: confirm on miss', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(magicFusionScenario());

    // Server pushes magic fusion prompt on miss
    await protocolHarness.pushServerMessage(magicFusionPrompt());

    // Click activate option
    await clickSkillChoiceOption(page, 0);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Server pushes discard selection (different element from attack)
    await protocolHarness.pushServerMessage(magicFusionDiscardPrompt('Fire'));

    // Select non-fire card to discard
    await page.getByTestId('hand-card-2').click();
    await page.getByTestId('prompt-confirm-btn').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      card_ids: ['hs-water-attack'],
    });

    // Server pushes seal selection
    await protocolHarness.pushServerMessage(magicFusionSealSelectPrompt());

    // Select seal to flip
    await clickOverlayOption(page, 'branch-option-0');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('magic fusion: skip', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(magicFusionScenario());

    await protocolHarness.pushServerMessage(magicFusionPrompt());

    // Click skip button
    await clickSkillChoiceSkip(page);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});

test.describe('heroic spirit seal suppress combo protocol harness', () => {
  test('seal suppress combo: select rage suppress (forced hit)', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(sealSuppressComboScenario());

    // Server pushes combo prompt on miss (mutual exclusive skills)
    await protocolHarness.pushServerMessage(sealSuppressComboPrompt());

    // Select rage suppress for forced hit
    await clickSkillChoiceOption(page, 0);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('seal suppress combo: select magic fusion (magic seal form +1 damage)', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(sealSuppressComboScenario());

    await protocolHarness.pushServerMessage(sealSuppressComboPrompt());

    // Select magic fusion for magic seal form bonus
    await clickSkillChoiceOption(page, 1);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});

test.describe('heroic spirit double echo protocol harness', () => {
  test('double echo: confirm after hit with crystal', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(doubleEchoScenario());

    // Server pushes double echo prompt after hit
    await protocolHarness.pushServerMessage(doubleEchoPrompt());

    // Click activate option
    await clickSkillChoiceOption(page, 0);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('double echo: skip', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(doubleEchoScenario());

    await protocolHarness.pushServerMessage(doubleEchoPrompt());

    // Click skip button
    await clickSkillChoiceSkip(page);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});
