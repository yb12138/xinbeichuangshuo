import type { Page } from '@playwright/test';
import { test } from '../../../fixtures/protocolHarness.fixture';
import {
  BD_DANCE_SKILL_ID,
  cocoonOverflowDiscardPrompt,
  danceDiscardPrompt,
  danceModePrompt,
  danceScenario,
} from '../../../scenarios/butterflyDancer';

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

test.describe('butterfly dancer dance protocol harness', () => {
  test('dance: select draw mode', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(danceScenario());

    // Activate dance skill
    await activatePanelSkill(page, BD_DANCE_SKILL_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: BD_DANCE_SKILL_ID,
    });

    // Server pushes dance mode prompt
    await protocolHarness.pushServerMessage(danceModePrompt(true));

    // Select "摸1张牌" (option index 0)
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

  test('dance: select discard mode then pick card', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(danceScenario());

    // Activate dance skill
    await activatePanelSkill(page, BD_DANCE_SKILL_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: BD_DANCE_SKILL_ID,
    });

    // Server pushes dance mode prompt
    await protocolHarness.pushServerMessage(danceModePrompt(true));

    // Select "弃1张牌" (option index 1)
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

    // Server pushes discard pick prompt (choose_cards with hand indices)
    await protocolHarness.pushServerMessage(danceDiscardPrompt());

    // Select the first hand card to discard
    await selectHandCards(page, [0]);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('dance: cocoon overflow discard (single cocoon)', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(danceScenario());

    // Push cocoon overflow discard prompt (choose_cards with cocoon labels)
    // max=1 means clicking cover card immediately submits
    await protocolHarness.pushServerMessage(cocoonOverflowDiscardPrompt(1));

    // Click the first cocoon cover card in expansion zone
    // For choose_cards mode with max <= 1, clicking cover card submits fieldIndex
    const coverCard = page.getByTestId('hand-card-0');
    await coverCard.scrollIntoViewIfNeeded();
    await coverCard.click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });
});
