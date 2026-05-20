import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  SC_TALISMAN_THUNDER_SKILL_ID,
  ENEMY_PLAYER_ID,
  ENEMY_2_PLAYER_ID,
  talismanThunderScenario,
  manaCollapseConfirmPrompt,
  talismanThunderDiscardPrompt,
  talismanThunderTargetPrompt,
} from '../../../scenarios/spiritCaster';

async function selectHandCards(page: import('@playwright/test').Page, indices: number[]) {
  for (const index of indices) {
    const card = page.getByTestId(`hand-card-${index}`);
    await card.scrollIntoViewIfNeeded();
    await card.click();
  }
  await page.getByTestId('prompt-confirm-btn').click();
}

test.describe('spirit caster talisman thunder protocol harness', () => {
  test('talisman thunder: with crystal -> mana collapse yes -> discard thunder -> select 2 targets', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(talismanThunderScenario({ hasCrystal: true }));

    // Activate skill
  await page.getByTestId('action-magic').click();
  await page.getByTestId('action-skill').click();
    await page.getByTestId(`skill-${SC_TALISMAN_THUNDER_SKILL_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: SC_TALISMAN_THUNDER_SKILL_ID,
    });

    // Mana collapse confirm (has crystal) - branch_select overlay
    await protocolHarness.pushServerMessage(manaCollapseConfirmPrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('prompt-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Discard thunder card (card_picker from hand)
    await protocolHarness.pushServerMessage(talismanThunderDiscardPrompt());
    await selectHandCards(page, [0]);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      card_ids: ['card_1'],
    });

    // Select 2 targets (multi target_picker - click both player-areas + prompt-confirm-btn)
    await protocolHarness.pushServerMessage(talismanThunderTargetPrompt());
    await expect(page.getByTestId('decision-overlay')).not.toBeVisible({ timeout: 5000 });
    await page.getByTestId(`player-area-${ENEMY_PLAYER_ID}`).click();
    await page.getByTestId(`player-area-${ENEMY_2_PLAYER_ID}`).click();
    await page.getByTestId('prompt-confirm-btn').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0, 1],
    });
  });

  test('talisman thunder: with crystal -> mana collapse no -> discard thunder -> select 2 targets', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(talismanThunderScenario({ hasCrystal: true }));
  await page.getByTestId('action-magic').click();
  await page.getByTestId('action-skill').click();
    await page.getByTestId(`skill-${SC_TALISMAN_THUNDER_SKILL_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: SC_TALISMAN_THUNDER_SKILL_ID,
    });

    await protocolHarness.pushServerMessage(manaCollapseConfirmPrompt());
    await page.getByTestId('prompt-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });

    await protocolHarness.pushServerMessage(talismanThunderDiscardPrompt());
    await selectHandCards(page, [0]);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      card_ids: ['card_1'],
    });

    await protocolHarness.pushServerMessage(talismanThunderTargetPrompt());
    await expect(page.getByTestId('decision-overlay')).not.toBeVisible({ timeout: 5000 });
    await page.getByTestId(`player-area-${ENEMY_PLAYER_ID}`).click();
    await page.getByTestId(`player-area-${ENEMY_2_PLAYER_ID}`).click();
    await page.getByTestId('prompt-confirm-btn').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0, 1],
    });
  });

  test('talisman thunder: no crystal -> discard thunder -> select 2 targets', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(talismanThunderScenario({ hasCrystal: false }));
  await page.getByTestId('action-magic').click();
  await page.getByTestId('action-skill').click();
    await page.getByTestId(`skill-${SC_TALISMAN_THUNDER_SKILL_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: SC_TALISMAN_THUNDER_SKILL_ID,
    });

    // No mana collapse prompt when no crystal
    await protocolHarness.pushServerMessage(talismanThunderDiscardPrompt());
    await selectHandCards(page, [0]);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      card_ids: ['card_1'],
    });

    await protocolHarness.pushServerMessage(talismanThunderTargetPrompt());
    await expect(page.getByTestId('decision-overlay')).not.toBeVisible({ timeout: 5000 });
    await page.getByTestId(`player-area-${ENEMY_PLAYER_ID}`).click();
    await page.getByTestId(`player-area-${ENEMY_2_PLAYER_ID}`).click();
    await page.getByTestId('prompt-confirm-btn').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0, 1],
    });
  });
});