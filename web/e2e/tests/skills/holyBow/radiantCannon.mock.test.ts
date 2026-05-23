import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  HB_RADIANT_CANNON_SKILL_ID,
  radiantCannonMoralePrompt,
  radiantCannonScenario,
} from '../../../scenarios/holyBow';

test.describe('holy bow radiant cannon protocol harness', () => {
  test('activate radiant cannon and choose red morale', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(radiantCannonScenario());


  await page.getByTestId('action-magic').click();
  await page.getByTestId('action-skill').click();
    await page.getByTestId(`skill-${HB_RADIANT_CANNON_SKILL_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: HB_RADIANT_CANNON_SKILL_ID,
    });

    // Morale choice
    await protocolHarness.pushServerMessage(radiantCannonMoralePrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    // Click "红方士气"
    await page.getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('activate radiant cannon and choose blue morale', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(radiantCannonScenario());


  await page.getByTestId('action-magic').click();
  await page.getByTestId('action-skill').click();
    await page.getByTestId(`skill-${HB_RADIANT_CANNON_SKILL_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: HB_RADIANT_CANNON_SKILL_ID,
    });

    // Morale choice
    await protocolHarness.pushServerMessage(radiantCannonMoralePrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    // Click "蓝方士气"
    await page.getByTestId('branch-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});