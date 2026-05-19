import type { Page } from '@playwright/test';
import { test } from '../../../fixtures/protocolHarness.fixture';
import {
  BW_BLAZING_CODEX_ID,
  BW_HEAVENFIRE_CLEAVE_ID,
  BW_PAIN_LINK_ID,
  blazingCodexScenario,
  heavenfireCleaveScenario,
  painLinkDiscardPrompt,
  painLinkScenario,
  witchWrathDrawPrompt,
  witchWrathScenario,
} from '../../../scenarios/blazeWitch';

async function activatePanelSkill(page: Page, skillId: string) {
  await page.getByTestId('action-skill').click();
  await page.getByTestId(`skill-${skillId}`).click();
}

async function selectHandCards(page: Page, indices: number[]) {
  for (const index of indices) {
    const card = page.getByTestId(`hand-card-${index}`);
    await card.scrollIntoViewIfNeeded();
    await card.click();
  }
  await page.getByTestId('prompt-confirm-btn').click();
}

test.describe('blaze witch blazing codex protocol harness', () => {
  test('activate blazing codex skill', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(blazingCodexScenario());

    await activatePanelSkill(page, BW_BLAZING_CODEX_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: BW_BLAZING_CODEX_ID,
    });
  });
});

test.describe('blaze witch heavenfire cleave protocol harness', () => {
  test('activate heavenfire cleave with rebirth available', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(heavenfireCleaveScenario({ tokens: { bw_rebirth: 1 } }));

    await activatePanelSkill(page, BW_HEAVENFIRE_CLEAVE_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: BW_HEAVENFIRE_CLEAVE_ID,
    });
  });
});

test.describe('blaze witch pain link protocol harness', () => {
  test('activate pain link skill', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(painLinkScenario());

    await activatePanelSkill(page, BW_PAIN_LINK_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: BW_PAIN_LINK_ID,
    });
  });

  test('pain link: discard down to 3 after damage', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(painLinkScenario());

    // Server pushes discard prompt after both damage hits resolve
    await protocolHarness.pushServerMessage(painLinkDiscardPrompt());

    // Need to discard 3 cards (6 hand → 3 remaining, so discard 3)
    await selectHandCards(page, [0, 1, 2]);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      card_ids: ['bw-fire-atk1', 'bw-fire-atk2', 'bw-fire-magic1'],
    });
  });
});

test.describe('blaze witch witch wrath protocol harness', () => {
  test('witch wrath: select draw 2 cards', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(witchWrathScenario());

    // Server pushes draw count prompt
    await protocolHarness.pushServerMessage(witchWrathDrawPrompt());

    // Select "摸2张" (ID "2", option index 2)
    const overlay = page.getByTestId('decision-overlay');
    const overlayVisible = await overlay.isVisible({ timeout: 1000 }).catch(() => false);
    if (overlayVisible) {
      await overlay.getByTestId('branch-option-2').click();
    } else {
      await page.getByTestId('prompt-dialog').getByTestId('branch-option-2').click();
    }
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [2],
    });
  });

  test('witch wrath: select draw 0 cards (skip draw)', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(witchWrathScenario());

    await protocolHarness.pushServerMessage(witchWrathDrawPrompt());

    // Select "摸0张" (ID "0", option index 0)
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

  test('witch wrath: select draw 1 card', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(witchWrathScenario());

    await protocolHarness.pushServerMessage(witchWrathDrawPrompt());

    // Select "摸1张" (ID "1", option index 1)
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
