import type { Page } from '@playwright/test';
import { test } from '../../../fixtures/protocolHarness.fixture';
import {
  darkReleaseScenario,
  ML_DARK_RELEASE_SKILL_ID,
} from '../../../scenarios/magicLancer';

async function activatePanelSkill(page: Page, skillId: string) {
  await page.getByTestId('action-magic').click();
  await page.getByTestId('action-skill').click();
  await page.getByTestId(`skill-${skillId}`).click();
}

test.describe('magic lancer dark release protocol harness', () => {
  test('activate skill sends SubmitAction', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(darkReleaseScenario());

    await activatePanelSkill(page, ML_DARK_RELEASE_SKILL_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: ML_DARK_RELEASE_SKILL_ID,
    });
  });
});
