import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  soulLinkScenario,
  soulLinkTargetPrompt,
  SS_PLAYER_ID,
  ALLY_PLAYER_ID,
  ALLY_2_PLAYER_ID,
  SS_SOUL_LINK_SKILL_ID,
} from '../../../scenarios/soulSorcerer';

test.describe('soulSorcerer soulLink protocol harness', () => {
  test('activate soulLink and select first ally', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(soulLinkScenario({ blue_soul: 1, yellow_soul: 1 }));

    // Click skill button (startup skill)
    await page.getByTestId('action-hub-trigger').click();
  await page.getByTestId('action-magic').click();
  await page.getByTestId('action-skill').click();
    await page.getByTestId(`skill-${SS_SOUL_LINK_SKILL_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: SS_SOUL_LINK_SKILL_ID,
    });

    // Target selection (target_picker - click ally player area, auto-submit)
    await protocolHarness.pushServerMessage(soulLinkTargetPrompt());
    await expect(page.getByTestId('decision-overlay')).not.toBeVisible({ timeout: 5000 });
    await page.getByTestId(`player-area-${ALLY_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('activate soulLink and select second ally', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(soulLinkScenario({ blue_soul: 2, yellow_soul: 2 }));
    await page.getByTestId('action-hub-trigger').click();
  await page.getByTestId('action-magic').click();
  await page.getByTestId('action-skill').click();
    await page.getByTestId(`skill-${SS_SOUL_LINK_SKILL_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: SS_SOUL_LINK_SKILL_ID,
    });

    await protocolHarness.pushServerMessage(soulLinkTargetPrompt());
    // Click second ally
    await page.getByTestId(`player-area-${ALLY_2_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });

  test('soulLink not available with insufficient souls', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(soulLinkScenario({ blue_soul: 0, yellow_soul: 1 }));

    // Skill should not be available since blue_soul < 1
    await expect(page.getByTestId(`skill-${SS_SOUL_LINK_SKILL_ID}`)).not.toBeVisible();
  });

  test('soulLink not available without exclusive card', async ({ page, protocolHarness }) => {
    const scenario = soulLinkScenario({ blue_soul: 1, yellow_soul: 1 });
    // Modify to have no exclusive card
    scenario.initialState.players[0].exclusive_card_count = 0;
    await protocolHarness.bootGame(scenario);

    // Skill should not be available without exclusive card
    await expect(page.getByTestId(`skill-${SS_SOUL_LINK_SKILL_ID}`)).not.toBeVisible();
  });
});