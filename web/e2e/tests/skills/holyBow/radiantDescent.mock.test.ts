import { test } from '../../../fixtures/protocolHarness.fixture';
import {
  HB_RADIANT_DESCENT_SKILL_ID,
  radiantDescentScenario,
} from '../../../scenarios/holyBow';

test.describe('holy bow radiant descent protocol harness', () => {
  test('activate radiant descent magic skill', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(radiantDescentScenario());

    // Activate skill directly
    await page.getByTestId('action-skill').click();
    await page.getByTestId(`skill-${HB_RADIANT_DESCENT_SKILL_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: HB_RADIANT_DESCENT_SKILL_ID,
    });
    // No further prompts - effect resolved by backend
  });
});