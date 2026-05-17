import type { Page } from '@playwright/test';
import { test } from '../../../fixtures/protocolHarness.fixture';
import {
  ELEMENTALIST_ELEMENT_IGNITE_ID,
  ELEMENTALIST_THUNDER_STRIKE_ID,
  ELEMENTALIST_FREEZE_ID,
  ELEMENTALIST_WIND_BLADE_ID,
  ELEMENTALIST_METEOR_ID,
  ELEMENTALIST_FIREBALL_ID,
  ELEMENTALIST_MOONLIGHT_ID,
  ENEMY_PLAYER_ID,
  ALLY_PLAYER_ID,
  elementIgniteScenario,
  elementIgniteTargetPrompt,
  thunderStrikeScenario,
  thunderStrikeTargetPrompt,
  freezeScenario,
  freezeDamageTargetPrompt,
  freezeHealTargetPrompt,
  windBladeScenario,
  meteorScenario,
  fireballScenario,
  moonlightScenario,
  moonlightTargetPrompt,
} from '../../../scenarios/elementalist';

async function activatePanelSkill(page: Page, skillId: string) {
  await page.getByTestId('action-skill').click();
  await page.getByTestId(`skill-${skillId}`).click();
}

async function selectTarget(page: Page, targetId: string) {
  await page.getByTestId(`player-area-${targetId}`).click();
}

test.describe('elementalist element ignite protocol harness', () => {
  test('element ignite: activate skill', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(elementIgniteScenario());

    await activatePanelSkill(page, ELEMENTALIST_ELEMENT_IGNITE_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: ELEMENTALIST_ELEMENT_IGNITE_ID,
    });
  });

  test('element ignite: select target', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(elementIgniteScenario());

    await activatePanelSkill(page, ELEMENTALIST_ELEMENT_IGNITE_ID);

    // Server pushes target selection
    await protocolHarness.pushServerMessage(elementIgniteTargetPrompt());

    // Select enemy target
    await selectTarget(page, ENEMY_PLAYER_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('element ignite: select ally target', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(elementIgniteScenario());

    await activatePanelSkill(page, ELEMENTALIST_ELEMENT_IGNITE_ID);

    await protocolHarness.pushServerMessage(elementIgniteTargetPrompt());

    // Select ally target
    await selectTarget(page, ALLY_PLAYER_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});

test.describe('elementalist thunder strike protocol harness', () => {
  test('thunder strike: activate and select target', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(thunderStrikeScenario());

    await activatePanelSkill(page, ELEMENTALIST_THUNDER_STRIKE_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: ELEMENTALIST_THUNDER_STRIKE_ID,
    });

    await protocolHarness.pushServerMessage(thunderStrikeTargetPrompt());

    await selectTarget(page, ENEMY_PLAYER_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });
});

test.describe('elementalist freeze protocol harness', () => {
  test('freeze: activate and select damage target', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(freezeScenario());

    await activatePanelSkill(page, ELEMENTALIST_FREEZE_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: ELEMENTALIST_FREEZE_ID,
    });

    await protocolHarness.pushServerMessage(freezeDamageTargetPrompt());

    // Select damage target (enemy)
    await selectTarget(page, ENEMY_PLAYER_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('freeze: select heal target after damage target', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(freezeScenario());

    await activatePanelSkill(page, ELEMENTALIST_FREEZE_ID);

    // First select damage target
    await protocolHarness.pushServerMessage(freezeDamageTargetPrompt());
    await selectTarget(page, ENEMY_PLAYER_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Then select heal target (can be self)
    await protocolHarness.pushServerMessage(freezeHealTargetPrompt());
    await selectTarget(page, ALLY_PLAYER_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [2],
    });
  });
});

test.describe('elementalist wind blade protocol harness', () => {
  test('wind blade: activate skill', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(windBladeScenario());

    await activatePanelSkill(page, ELEMENTALIST_WIND_BLADE_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: ELEMENTALIST_WIND_BLADE_ID,
    });
  });
});

test.describe('elementalist meteor protocol harness', () => {
  test('meteor: activate skill', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(meteorScenario());

    await activatePanelSkill(page, ELEMENTALIST_METEOR_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: ELEMENTALIST_METEOR_ID,
    });
  });
});

test.describe('elementalist fireball protocol harness', () => {
  test('fireball: activate skill', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(fireballScenario());

    await activatePanelSkill(page, ELEMENTALIST_FIREBALL_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: ELEMENTALIST_FIREBALL_ID,
    });
  });
});

test.describe('elementalist moonlight protocol harness', () => {
  test('moonlight: activate skill with gem', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(moonlightScenario());

    await activatePanelSkill(page, ELEMENTALIST_MOONLIGHT_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: ELEMENTALIST_MOONLIGHT_ID,
    });
  });

  test('moonlight: select target with energy X=3', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(moonlightScenario({ energy: 3 }));

    await activatePanelSkill(page, ELEMENTALIST_MOONLIGHT_ID);

    // Server pushes target selection (damage = 3+1 = 4)
    await protocolHarness.pushServerMessage(moonlightTargetPrompt(3));

    await selectTarget(page, ENEMY_PLAYER_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('moonlight: select target with energy X=0', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(moonlightScenario({ energy: 0 }));

    await activatePanelSkill(page, ELEMENTALIST_MOONLIGHT_ID);

    // Server pushes target selection (damage = 0+1 = 1)
    await protocolHarness.pushServerMessage(moonlightTargetPrompt(0));

    await selectTarget(page, ENEMY_PLAYER_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });
});