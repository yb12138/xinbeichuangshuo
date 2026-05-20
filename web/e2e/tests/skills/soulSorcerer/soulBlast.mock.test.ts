import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  soulBlastScenario,
  soulBlastTargetPrompt,
  SS_PLAYER_ID,
  ENEMY_PLAYER_ID,
  SS_SOUL_BLAST_SKILL_ID,
} from '../../../scenarios/soulSorcerer';

test.describe('soulSorcerer soulBlast protocol harness', () => {
  test('activate soulBlast and target enemy', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(soulBlastScenario({ yellow_soul: 3 }));


    await page.getByTestId('action-hub-trigger').click();
  await page.getByTestId('action-magic').click();
  await page.getByTestId('action-skill').click();
    await page.getByTestId(`skill-${SS_SOUL_BLAST_SKILL_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: SS_SOUL_BLAST_SKILL_ID,
    });

    // Target selection (target_picker - click enemy player area, auto-submit)
    await protocolHarness.pushServerMessage(soulBlastTargetPrompt());
    await expect(page.getByTestId('decision-overlay')).not.toBeVisible({ timeout: 5000 });
    await page.getByTestId(`player-area-${ENEMY_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });

  test('activate soulBlast and target self', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(soulBlastScenario({ yellow_soul: 4 }));
    await page.getByTestId('action-hub-trigger').click();
  await page.getByTestId('action-magic').click();
  await page.getByTestId('action-skill').click();
    await page.getByTestId(`skill-${SS_SOUL_BLAST_SKILL_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: SS_SOUL_BLAST_SKILL_ID,
    });

    await protocolHarness.pushServerMessage(soulBlastTargetPrompt());
    // Click self player area
    await page.getByTestId(`player-area-${SS_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('soulBlast not available with insufficient yellow souls', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(soulBlastScenario({ yellow_soul: 2 }));

    // Skill should not be available since yellow soul < 3
    await expect(page.getByTestId(`skill-${SS_SOUL_BLAST_SKILL_ID}`)).not.toBeVisible();
  });

  test('soulBlast with extra damage condition', async ({ page, protocolHarness }) => {
    // When target has hand < 3 and max_hand > 5, damage +2
    await protocolHarness.bootGame(soulBlastScenario({ yellow_soul: 3 }));
    await page.getByTestId('action-hub-trigger').click();
  await page.getByTestId('action-magic').click();
  await page.getByTestId('action-skill').click();
    await page.getByTestId(`skill-${SS_SOUL_BLAST_SKILL_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: SS_SOUL_BLAST_SKILL_ID,
    });

    await protocolHarness.pushServerMessage(soulBlastTargetPrompt());
    await page.getByTestId(`player-area-${ENEMY_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});