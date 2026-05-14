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

test.describe('spirit caster talisman thunder protocol harness', () => {
  test('talisman thunder: with crystal -> mana collapse yes -> discard thunder -> select target 1', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(talismanThunderScenario({ hasCrystal: true }));

    // Activate skill
    await page.getByTestId('action-skill').click();
    await page.getByTestId(`skill-${SC_TALISMAN_THUNDER_SKILL_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: SC_TALISMAN_THUNDER_SKILL_ID,
    });

    // Mana collapse confirm (has crystal)
    await protocolHarness.pushServerMessage(manaCollapseConfirmPrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('prompt-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Discard thunder card
    await protocolHarness.pushServerMessage(talismanThunderDiscardPrompt());
    await page.getByTestId('hand-card-0').click();
    await page.getByTestId('prompt-confirm-btn').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Select first target (choose_target type sends target_ref)
    await protocolHarness.pushServerMessage(talismanThunderTargetPrompt());
    await page.getByTestId(`player-area-${ENEMY_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      targets: [{ target_user_id: ENEMY_PLAYER_ID }],
    });
  });

  test('talisman thunder: with crystal -> mana collapse no -> discard thunder -> select target', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(talismanThunderScenario({ hasCrystal: true }));

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
    await page.getByTestId('hand-card-0').click();
    await page.getByTestId('prompt-confirm-btn').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    await protocolHarness.pushServerMessage(talismanThunderTargetPrompt());
    await page.getByTestId(`player-area-${ENEMY_2_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      targets: [{ target_user_id: ENEMY_2_PLAYER_ID }],
    });
  });

  test('talisman thunder: no crystal -> discard thunder -> select target', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(talismanThunderScenario({ hasCrystal: false }));

    await page.getByTestId('action-skill').click();
    await page.getByTestId(`skill-${SC_TALISMAN_THUNDER_SKILL_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: SC_TALISMAN_THUNDER_SKILL_ID,
    });

    // No mana collapse prompt when no crystal
    await protocolHarness.pushServerMessage(talismanThunderDiscardPrompt());
    await page.getByTestId('hand-card-0').click();
    await page.getByTestId('prompt-confirm-btn').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    await protocolHarness.pushServerMessage(talismanThunderTargetPrompt());
    await page.getByTestId(`player-area-${ENEMY_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      targets: [{ target_user_id: ENEMY_PLAYER_ID }],
    });
  });
});