import type { Page } from '@playwright/test';
import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  ALLY_PLAYER_ID,
  ENEMY_PLAYER_ID,
  SAGE_ARCANE_CODEX_ID,
  SAGE_HOLY_CODEX_ID,
  SAGE_PLAYER_ID,
  arcaneCardsPrompt,
  arcaneCodexScenario,
  arcaneTargetPrompt,
  holyCardsPrompt,
  holyCodexScenario,
  holyTargetsStepPrompt,
} from '../../../scenarios/sage';

async function activatePanelSkill(page: Page, skillId: string) {
  await page.getByTestId('action-hub-trigger').click();
  await page.getByTestId('action-magic').click();
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

test.describe('sage arcane codex protocol harness', () => {
  test('activate arcane codex skill', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(arcaneCodexScenario());

    await activatePanelSkill(page, SAGE_ARCANE_CODEX_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: SAGE_ARCANE_CODEX_ID,
    });
  });

  test('arcane codex: select cards then target enemy', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(arcaneCodexScenario());

    // Server pushes arcane cards prompt (card_picker from hand)
    await protocolHarness.pushServerMessage(arcaneCardsPrompt());

    // Select 2 different-element cards (indices 0=Fire, 1=Water)
    await selectHandCards(page, [0, 1]);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      card_ids: ['sg-fire-atk', 'sg-water-magic'],
    });

    // Server pushes target prompt (target_picker)
    await protocolHarness.pushServerMessage(arcaneTargetPrompt());

    // Click enemy player area (auto-submits for single target)
    await page.getByTestId(`player-area-${ENEMY_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('arcane codex: select self as target', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(arcaneCodexScenario());

    // Select cards
    await protocolHarness.pushServerMessage(arcaneCardsPrompt());
    await selectHandCards(page, [2, 3]);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      card_ids: ['sg-thunder-atk', 'sg-earth-magic'],
    });

    // Target self (target_picker: click player area)
    await protocolHarness.pushServerMessage(arcaneTargetPrompt());
    await page.getByTestId(`player-area-${SAGE_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [2],
    });
  });
});

test.describe('sage holy codex protocol harness', () => {
  test('activate holy codex skill', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(holyCodexScenario());

    await activatePanelSkill(page, SAGE_HOLY_CODEX_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: SAGE_HOLY_CODEX_ID,
    });
  });

  test('holy codex: full flow (cards → manual target picker confirm)', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(holyCodexScenario());

    // Step 1: select 3 different-element cards (card_picker from hand)
    await protocolHarness.pushServerMessage(holyCardsPrompt());
    await selectHandCards(page, [0, 1, 2]);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      card_ids: ['sg-fire-atk', 'sg-water-magic', 'sg-thunder-atk'],
    });

    // Step 2: select target on the board with multi_target picker, then confirm.
    await protocolHarness.pushServerMessage(
      holyTargetsStepPrompt(1, 1, ['Enemy E1', 'Ally A1', 'E2E Sage']),
    );
    await expect(page.getByTestId('decision-overlay')).not.toBeVisible({ timeout: 5000 });
    await page.getByTestId(`player-area-${ENEMY_PLAYER_ID}`).click();
    await page.getByTestId('prompt-confirm-btn').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('holy codex: select 4 cards, 2 targets', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(holyCodexScenario());

    // Step 1: select 4 different-element cards (card_picker from hand)
    await protocolHarness.pushServerMessage(holyCardsPrompt());
    await selectHandCards(page, [0, 1, 2, 3]);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      card_ids: ['sg-fire-atk', 'sg-water-magic', 'sg-thunder-atk', 'sg-earth-magic'],
    });

    // Step 2: select two targets on the board (multi_target picker), then confirm once.
    await protocolHarness.pushServerMessage(
      holyTargetsStepPrompt(1, 2, ['Enemy E1', 'Ally A1', 'E2E Sage']),
    );
    await page.getByTestId(`player-area-${ALLY_PLAYER_ID}`).click();
    await page.getByTestId(`player-area-${SAGE_PLAYER_ID}`).click();
    await page.getByTestId('prompt-confirm-btn').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1, 2],
    });
  });
});