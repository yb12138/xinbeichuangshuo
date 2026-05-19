import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  soulAmpScenario,
  soulAmpConfirmPrompt,
  SS_PLAYER_ID,
  SS_SOUL_AMP_SKILL_ID,
} from '../../../scenarios/soulSorcerer';

test.describe('soulSorcerer soulAmp protocol harness', () => {
  test('activate soulAmp with gem', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(soulAmpScenario({ gems: 1 }));

    // Click skill button (startup skill)
    await page.getByTestId('action-skill').click();
    await page.getByTestId(`skill-${SS_SOUL_AMP_SKILL_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: SS_SOUL_AMP_SKILL_ID,
    });

    // Server pushes confirm prompt (branch_select overlay)
    await protocolHarness.pushServerMessage(soulAmpConfirmPrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    // Click "发动" (branch-option-0)
    await page.getByTestId('decision-overlay').getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('activate soulAmp with multiple gems', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(soulAmpScenario({ gems: 2 }));

    await page.getByTestId('action-skill').click();
    await page.getByTestId(`skill-${SS_SOUL_AMP_SKILL_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: SS_SOUL_AMP_SKILL_ID,
    });

    await protocolHarness.pushServerMessage(soulAmpConfirmPrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('decision-overlay').getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('decline soulAmp activation', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(soulAmpScenario({ gems: 1 }));

    await page.getByTestId('action-skill').click();
    await page.getByTestId(`skill-${SS_SOUL_AMP_SKILL_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: SS_SOUL_AMP_SKILL_ID,
    });

    await protocolHarness.pushServerMessage(soulAmpConfirmPrompt());
    // Click "不发动" (branch-option-1)
    await page.getByTestId('decision-overlay').getByTestId('branch-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });

  test('soulAmp not available without gems', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(soulAmpScenario({ gems: 0 }));

    // Skill should not be available since gem < 1
    await expect(page.getByTestId(`skill-${SS_SOUL_AMP_SKILL_ID}`)).not.toBeVisible();
  });
});