import type { Page } from '@playwright/test';
import { test } from '../../../fixtures/protocolHarness.fixture';
import {
  MAGIC_GIRL_MAGIC_EXPLOSION_ID,
  MAGIC_GIRL_DESTRUCTION_STORM_ID,
  ENEMY_PLAYER_ID,
  ENEMY2_PLAYER_ID,
  magicExplosionScenario,
  magicExplosionTargetPrompt,
  destructionStormScenario,
  destructionStormTargetPrompt,
} from '../../../scenarios/magicGirl';

async function activatePanelSkill(page: Page, skillId: string) {
  await page.getByTestId('action-skill').click();
  await page.getByTestId(`skill-${skillId}`).click();
}

async function selectTarget(page: Page, targetId: string) {
  await page.getByTestId(`player-area-${targetId}`).click();
}

test.describe('magic girl magic explosion protocol harness', () => {
  test('magic explosion: activate skill', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(magicExplosionScenario());

    await activatePanelSkill(page, MAGIC_GIRL_MAGIC_EXPLOSION_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: MAGIC_GIRL_MAGIC_EXPLOSION_ID,
    });
  });

  test('magic explosion: select 2 targets', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(magicExplosionScenario());

    await activatePanelSkill(page, MAGIC_GIRL_MAGIC_EXPLOSION_ID);

    // Server pushes target selection
    await protocolHarness.pushServerMessage(magicExplosionTargetPrompt());

    // Select 2 enemy targets
    await selectTarget(page, ENEMY_PLAYER_ID);
    await selectTarget(page, ENEMY2_PLAYER_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0, 1],
    });
  });
});

test.describe('magic girl destruction storm protocol harness', () => {
  test('destruction storm: activate skill', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(destructionStormScenario());

    await activatePanelSkill(page, MAGIC_GIRL_DESTRUCTION_STORM_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: MAGIC_GIRL_DESTRUCTION_STORM_ID,
    });
  });

  test('destruction storm: select 2 targets for damage', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(destructionStormScenario());

    await activatePanelSkill(page, MAGIC_GIRL_DESTRUCTION_STORM_ID);

    // Server pushes target selection
    await protocolHarness.pushServerMessage(destructionStormTargetPrompt());

    // Select 2 enemy targets
    await selectTarget(page, ENEMY_PLAYER_ID);
    await selectTarget(page, ENEMY2_PLAYER_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0, 1],
    });
  });
});