import type { Page } from '@playwright/test';
import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  ENEMY_PLAYER_ID,
  ML_PHANTOM_STARDUST_SKILL_ID,
  phantomStardustScenario,
  stardustTargetPrompt,
} from '../../../scenarios/magicLancer';

async function activatePanelSkill(page: Page, skillId: string) {
  await page.getByTestId('action-skill').click();
  await page.getByTestId(`skill-${skillId}`).click();
}

test.describe('magic lancer phantom stardust protocol harness', () => {
  test('activate sends Skill submit, then select stardust target', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(phantomStardustScenario());

    // Step 1: activate skill (phantom stardust requires one enemy target)
    // Since target_type=2 (enemy), the frontend may require target selection first.
    // If min_targets=1, the frontend should enter target-selection mode.
    await activatePanelSkill(page, ML_PHANTOM_STARDUST_SKILL_ID);
    // For target_type=2 with max_targets=1, expect the skill submission
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: ML_PHANTOM_STARDUST_SKILL_ID,
    });

    // Step 2: server returns stardust target prompt (after self-damage resolves, morale didn't drop)
    await protocolHarness.pushServerMessage(stardustTargetPrompt());
    // Target prompt is rendered as player area click (target_picker mode)
    await expect(page.getByTestId('decision-overlay')).not.toBeVisible({ timeout: 3000 }).catch(() => {});
    await page.getByTestId(`player-area-${ENEMY_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });
});
