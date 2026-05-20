import type { Page } from '@playwright/test';
import { test } from '../../../fixtures/protocolHarness.fixture';
import {
  HOLY_LANCER_GLORY_ID,
  HOLY_LANCER_PUNISHMENT_ID,
  HOLY_LANCER_HOLY_LIGHT_HEAL_ID,
  gloryScenario,
  punishmentScenario,
  punishmentTargetPrompt,
  holyLightHealScenario,
} from '../../../scenarios/holyLancer';

async function activatePanelSkill(page: Page, skillId: string) {
  await page.getByTestId('action-magic').click();
  await page.getByTestId('action-skill').click();
  await page.getByTestId(`skill-${skillId}`).click();
}

async function selectTarget(page: Page, targetId: string) {
  await page.getByTestId(`player-area-${targetId}`).click();
}

test.describe('holy lancer glory protocol harness', () => {
  test('glory: activate skill', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(gloryScenario());

    await activatePanelSkill(page, HOLY_LANCER_GLORY_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: HOLY_LANCER_GLORY_ID,
    });
  });
});

test.describe('holy lancer punishment protocol harness', () => {
  test('punishment: activate and select target', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(punishmentScenario());

    await activatePanelSkill(page, HOLY_LANCER_PUNISHMENT_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: HOLY_LANCER_PUNISHMENT_ID,
    });

    // Server pushes target selection prompt
    await protocolHarness.pushServerMessage(punishmentTargetPrompt());

    // Select enemy target
    await selectTarget(page, 'enemy_1');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });
});

test.describe('holy lancer holy light heal protocol harness', () => {
  test('holy light heal: activate skill with gem', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(holyLightHealScenario());

    await activatePanelSkill(page, HOLY_LANCER_HOLY_LIGHT_HEAL_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: HOLY_LANCER_HOLY_LIGHT_HEAL_ID,
    });
  });
});