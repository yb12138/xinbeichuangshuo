import type { Page } from '@playwright/test';
import { test } from '../../../fixtures/protocolHarness.fixture';
import {
  VALKYRIE_ORDER_MARK_ID,
  ENEMY_PLAYER_ID,
  ALLY_PLAYER_ID,
  orderMarkScenario,
} from '../../../scenarios/valkyrie';

async function activatePanelSkill(page: Page, skillId: string) {
  await page.getByTestId('action-hub-trigger').click();
  await page.getByTestId('action-magic').click();
  await page.getByTestId('action-skill').click();
  await page.getByTestId(`skill-${skillId}`).click();
}

async function selectTarget(page: Page, targetId: string) {
  await page.getByTestId(`player-area-${targetId}`).click();
}

test.describe('valkyrie order mark protocol harness', () => {
  test('order mark: activate skill', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(orderMarkScenario());

    await activatePanelSkill(page, VALKYRIE_ORDER_MARK_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: VALKYRIE_ORDER_MARK_ID,
    });
  });
});