import { test } from '../../../fixtures/protocolHarness.fixture';
import {
  SC_TALISMAN_WIND_SKILL_ID,
  ENEMY_PLAYER_ID,
  ENEMY_2_PLAYER_ID,
  talismanWindScenario,
  talismanWindDiscardPrompt,
  talismanWindTargetPrompt,
} from '../../../scenarios/spiritCaster';

test.describe('spirit caster talisman wind protocol harness', () => {
  test('talisman wind: discard wind card -> select first target', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(talismanWindScenario());

    // Activate skill
    await page.getByTestId('action-skill').click();
    await page.getByTestId(`skill-${SC_TALISMAN_WIND_SKILL_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: SC_TALISMAN_WIND_SKILL_ID,
    });

    // Discard wind card
    await protocolHarness.pushServerMessage(talismanWindDiscardPrompt());
    await page.getByTestId('hand-card-1').click();
    await page.getByTestId('prompt-confirm-btn').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });

    // Select first target (choose_target sends target_ref)
    await protocolHarness.pushServerMessage(talismanWindTargetPrompt());
    await page.getByTestId(`player-area-${ENEMY_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      targets: [{ target_user_id: ENEMY_PLAYER_ID }],
    });
  });

  test('talisman wind: discard wind card -> select second target', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(talismanWindScenario());

    await page.getByTestId('action-skill').click();
    await page.getByTestId(`skill-${SC_TALISMAN_WIND_SKILL_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: SC_TALISMAN_WIND_SKILL_ID,
    });

    await protocolHarness.pushServerMessage(talismanWindDiscardPrompt());
    await page.getByTestId('hand-card-1').click();
    await page.getByTestId('prompt-confirm-btn').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });

    await protocolHarness.pushServerMessage(talismanWindTargetPrompt());
    await page.getByTestId(`player-area-${ENEMY_2_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      targets: [{ target_user_id: ENEMY_2_PLAYER_ID }],
    });
  });
});