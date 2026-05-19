import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  SEALER_SEAL_BREAK_ID,
  SEALER_FIVE_ELEMENTS_BIND_ID,
  SEALER_WATER_SEAL_ID,
  SEALER_FIRE_SEAL_ID,
  ENEMY_PLAYER_ID,
  sealBreakScenario,
  sealBreakFieldSelectPrompt,
  fiveElementsBindScenario,
  fiveElementsBindTargetPrompt,
  waterSealScenario,
  waterSealTargetPrompt,
  fireSealScenario,
  fireSealTargetPrompt,
} from '../../../scenarios/sealer';

async function activatePanelSkill(page: import('@playwright/test').Page, skillId: string) {
  await page.getByTestId('action-skill').click();
  await page.getByTestId(`skill-${skillId}`).click();
}

async function selectTarget(page: import('@playwright/test').Page, targetId: string) {
  await page.getByTestId(`player-area-${targetId}`).click();
}

test.describe('sealer seal break protocol harness', () => {
  test('seal break: activate skill', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(sealBreakScenario());

    await activatePanelSkill(page, SEALER_SEAL_BREAK_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: SEALER_SEAL_BREAK_ID,
    });
  });

  test('seal break: select field effect to take', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(sealBreakScenario());

    await activatePanelSkill(page, SEALER_SEAL_BREAK_ID);

    // Server pushes field effect selection
    await protocolHarness.pushServerMessage(sealBreakFieldSelectPrompt());

    // Select enemy's shield (branch_select: branch-option-0)
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('decision-overlay').getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });
});

test.describe('sealer five elements bind protocol harness', () => {
  test('five elements bind: activate and select target', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(fiveElementsBindScenario());

    await activatePanelSkill(page, SEALER_FIVE_ELEMENTS_BIND_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: SEALER_FIVE_ELEMENTS_BIND_ID,
    });

    // Server pushes target selection
    await protocolHarness.pushServerMessage(fiveElementsBindTargetPrompt());

    // Select enemy target (target_picker: click player area)
    await expect(page.getByTestId('decision-overlay')).not.toBeVisible({ timeout: 5000 });
    await selectTarget(page, ENEMY_PLAYER_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });
});

test.describe('sealer elemental seal protocol harness', () => {
  test('water seal: activate and select target', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(waterSealScenario());

    await activatePanelSkill(page, SEALER_WATER_SEAL_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: SEALER_WATER_SEAL_ID,
    });

    // Server pushes target selection
    await protocolHarness.pushServerMessage(waterSealTargetPrompt());

    await selectTarget(page, ENEMY_PLAYER_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('fire seal: activate and select target', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(fireSealScenario());

    await activatePanelSkill(page, SEALER_FIRE_SEAL_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: SEALER_FIRE_SEAL_ID,
    });

    // Server pushes target selection
    await protocolHarness.pushServerMessage(fireSealTargetPrompt());

    await selectTarget(page, ENEMY_PLAYER_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });
});