import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  soulGrantScenario,
  soulGrantTargetPrompt,
  SS_PLAYER_ID,
  ALLY_PLAYER_ID,
  SS_SOUL_GRANT_SKILL_ID,
} from '../../../scenarios/soulSorcerer';

test.describe('soulSorcerer soulGrant protocol harness', () => {
  test('activate soulGrant and target ally', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(soulGrantScenario({ blue_soul: 3 }));


    await page.getByTestId('action-hub-trigger').click();
  await page.getByTestId('action-magic').click();
  await page.getByTestId('action-skill').click();
    await page.getByTestId(`skill-${SS_SOUL_GRANT_SKILL_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: SS_SOUL_GRANT_SKILL_ID,
    });

    // Target selection (target_picker - click ally player area, auto-submit)
    await protocolHarness.pushServerMessage(soulGrantTargetPrompt());
    await expect(page.getByTestId('decision-overlay')).not.toBeVisible({ timeout: 5000 });
    await page.getByTestId(`player-area-${ALLY_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });

  test('activate soulGrant and target self', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(soulGrantScenario({ blue_soul: 4 }));
    await page.getByTestId('action-hub-trigger').click();
  await page.getByTestId('action-magic').click();
  await page.getByTestId('action-skill').click();
    await page.getByTestId(`skill-${SS_SOUL_GRANT_SKILL_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: SS_SOUL_GRANT_SKILL_ID,
    });

    await protocolHarness.pushServerMessage(soulGrantTargetPrompt());
    // Click self player area
    await page.getByTestId(`player-area-${SS_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('soulGrant not available with insufficient blue souls', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(soulGrantScenario({ blue_soul: 2 }));

    // Skill should not be available since blue soul < 3
    await expect(page.getByTestId(`skill-${SS_SOUL_GRANT_SKILL_ID}`)).not.toBeVisible();
  });
});