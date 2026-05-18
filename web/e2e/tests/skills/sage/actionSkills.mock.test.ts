import type { Page } from '@playwright/test';
import { test } from '../../../fixtures/protocolHarness.fixture';
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

    // Server pushes arcane cards prompt (after skill activation and gem consumption)
    await protocolHarness.pushServerMessage(arcaneCardsPrompt());

    // Select 2 different-element cards (indices 0=Fire, 1=Water)
    await selectHandCards(page, [0, 1]);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0, 1],
    });

    // Server pushes target prompt
    await protocolHarness.pushServerMessage(arcaneTargetPrompt());

    // Click enemy player area
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
      option_indexes: [2, 3],
    });

    // Target self
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

    // Step 1: select 3 different-element cards
    await protocolHarness.pushServerMessage(holyCardsPrompt());
    await selectHandCards(page, [0, 1, 2]);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0, 1, 2],
    });

    // Step 2: select first (and only) target on the board, then confirm in the dock.
    await protocolHarness.pushServerMessage(
      holyTargetsStepPrompt(1, 1, ['Enemy E1', 'Ally A1', 'E2E Sage']),
    );
    await page.getByTestId(`player-area-${ENEMY_PLAYER_ID}`).click();
    await page.getByTestId('prompt-confirm-btn').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('holy codex: select 4 cards, 2 targets', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(holyCodexScenario());

    // Step 1: select 4 different-element cards (X=4, maxTargets = 4-2 = 2)
    await protocolHarness.pushServerMessage(holyCardsPrompt());
    await selectHandCards(page, [0, 1, 2, 3]);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0, 1, 2, 3],
    });

    // Step 2: select two targets on the board, then confirm once in the dock.
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
