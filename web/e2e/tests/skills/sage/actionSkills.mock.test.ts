import type { Page } from '@playwright/test';
import { test } from '../../../fixtures/protocolHarness.fixture';
import {
  ENEMY_PLAYER_ID,
  SAGE_ARCANE_CODEX_ID,
  SAGE_HOLY_CODEX_ID,
  SAGE_PLAYER_ID,
  arcaneCardsPrompt,
  arcaneCodexScenario,
  arcaneTargetPrompt,
  holyCardsPrompt,
  holyCodexScenario,
  holyTargetCountPrompt,
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

async function clickOverlayOption(page: Page, selector: string) {
  const overlay = page.getByTestId('decision-overlay');
  const overlayVisible = await overlay.isVisible({ timeout: 1000 }).catch(() => false);
  if (overlayVisible) {
    await overlay.getByTestId(selector).click();
  } else {
    await page.getByTestId('prompt-dialog').getByTestId(selector).click();
  }
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

  test('holy codex: full flow (cards → target count → sequential targets)', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(holyCodexScenario());

    // Step 1: select 3 different-element cards
    await protocolHarness.pushServerMessage(holyCardsPrompt());
    await selectHandCards(page, [0, 1, 2]);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0, 1, 2],
    });

    // Step 2: select target count (X-2=1, so max 1 target)
    // Actually with 3 cards selected, maxTargetCount = 3-2 = 1, only 1 option
    await protocolHarness.pushServerMessage(holyTargetCountPrompt(1));

    // Only one option: "选择1名角色" (branch-option-0)
    await clickOverlayOption(page, 'branch-option-0');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Step 3: select first (and only) target
    await protocolHarness.pushServerMessage(
      holyTargetsStepPrompt(1, 1, ['Enemy E1', 'Ally A1', 'E2E Sage']),
    );
    // Select enemy (branch-option-0)
    await clickOverlayOption(page, 'branch-option-0');
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

    // Step 2: select target count = 2 (max 2 options: "选择1名角色", "选择2名角色")
    await protocolHarness.pushServerMessage(holyTargetCountPrompt(2));

    // Select "选择2名角色" (branch-option-1)
    await clickOverlayOption(page, 'branch-option-1');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });

    // Step 3: select first target (3 remaining options)
    await protocolHarness.pushServerMessage(
      holyTargetsStepPrompt(1, 2, ['Enemy E1', 'Ally A1', 'E2E Sage']),
    );
    // Select Ally (branch-option-1)
    await clickOverlayOption(page, 'branch-option-1');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });

    // Step 4: select second target (2 remaining)
    await protocolHarness.pushServerMessage(
      holyTargetsStepPrompt(2, 2, ['Enemy E1', 'E2E Sage']),
    );
    // Select self (E2E Sage) (branch-option-1)
    await clickOverlayOption(page, 'branch-option-1');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});
