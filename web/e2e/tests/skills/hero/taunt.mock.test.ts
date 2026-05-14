import { test } from '../../../fixtures/protocolHarness.fixture';
import {
  HERO_TAUNT_SKILL_ID,
  ENEMY_PLAYER_ID,
  ENEMY_2_PLAYER_ID,
  tauntScenario,
  tauntTargetPrompt,
} from '../../../scenarios/hero';

test.describe('hero taunt protocol harness', () => {
  test('activate taunt and select target', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(tauntScenario({ anger: 1 }));

    // Click skill button
    await page.getByTestId('action-skill').click();
    await page.getByTestId(`skill-${HERO_TAUNT_SKILL_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: HERO_TAUNT_SKILL_ID,
    });

    // Target selection: click on enemy player card
    await protocolHarness.pushServerMessage(tauntTargetPrompt());
    await page.getByTestId(`player-area-${ENEMY_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('activate taunt and select second enemy', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(tauntScenario({ anger: 2 }));

    await page.getByTestId('action-skill').click();
    await page.getByTestId(`skill-${HERO_TAUNT_SKILL_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: HERO_TAUNT_SKILL_ID,
    });

    // Target selection: click on second enemy
    await protocolHarness.pushServerMessage(tauntTargetPrompt());
    await page.getByTestId(`player-area-${ENEMY_2_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});