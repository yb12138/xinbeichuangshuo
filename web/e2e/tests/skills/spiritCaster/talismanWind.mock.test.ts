import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  SC_TALISMAN_WIND_SKILL_ID,
  ENEMY_PLAYER_ID,
  ENEMY_2_PLAYER_ID,
  talismanWindScenario,
  talismanWindDiscardPrompt,
  talismanWindTargetPrompt,
} from '../../../scenarios/spiritCaster';

async function selectHandCards(page: import('@playwright/test').Page, indices: number[]) {
  for (const index of indices) {
    const card = page.getByTestId(`hand-card-${index}`);
    await card.scrollIntoViewIfNeeded();
    await card.click();
  }
  await page.getByTestId('prompt-confirm-btn').click();
}

test.describe('spirit caster talisman wind protocol harness', () => {
  test('talisman wind: discard wind card -> select 2 targets', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(talismanWindScenario());


    await page.getByTestId('action-hub-trigger').click();
  await page.getByTestId('action-magic').click();
  await page.getByTestId('action-skill').click();
    await page.getByTestId(`skill-${SC_TALISMAN_WIND_SKILL_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: SC_TALISMAN_WIND_SKILL_ID,
    });

    // Discard wind card (card_picker from hand)
    await protocolHarness.pushServerMessage(talismanWindDiscardPrompt());
    await selectHandCards(page, [1]);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      card_ids: ['card_2'],
    });

    // Select 2 targets (multi target_picker - click both player-areas + prompt-confirm-btn)
    await protocolHarness.pushServerMessage(talismanWindTargetPrompt());
    await expect(page.getByTestId('decision-overlay')).not.toBeVisible({ timeout: 5000 });
    await page.getByTestId(`player-area-${ENEMY_PLAYER_ID}`).click();
    await page.getByTestId(`player-area-${ENEMY_2_PLAYER_ID}`).click();
    await page.getByTestId('prompt-confirm-btn').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0, 1],
    });
  });

  test('talisman wind: discard wind card -> select 2 targets (reverse order)', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(talismanWindScenario());
    await page.getByTestId('action-hub-trigger').click();
  await page.getByTestId('action-magic').click();
  await page.getByTestId('action-skill').click();
    await page.getByTestId(`skill-${SC_TALISMAN_WIND_SKILL_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: SC_TALISMAN_WIND_SKILL_ID,
    });

    await protocolHarness.pushServerMessage(talismanWindDiscardPrompt());
    await selectHandCards(page, [1]);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      card_ids: ['card_2'],
    });

    // Select 2 targets in reverse click order (option_indexes follow click order)
    await protocolHarness.pushServerMessage(talismanWindTargetPrompt());
    await expect(page.getByTestId('decision-overlay')).not.toBeVisible({ timeout: 5000 });
    await page.getByTestId(`player-area-${ENEMY_2_PLAYER_ID}`).click();
    await page.getByTestId(`player-area-${ENEMY_PLAYER_ID}`).click();
    await page.getByTestId('prompt-confirm-btn').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1, 0],
    });
  });
});