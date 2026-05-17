import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  soulMirrorScenario,
  soulMirrorTargetPrompt,
  SS_PLAYER_ID,
  ALLY_PLAYER_ID,
  SS_SOUL_MIRROR_SKILL_ID,
} from '../../../scenarios/soulSorcerer';

test.describe('soulSorcerer soulMirror protocol harness', () => {
  test('activate soulMirror and target ally', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(soulMirrorScenario({ yellow_soul: 2 }));

    // Click skill button
    await page.getByTestId('action-skill').click();
    await page.getByTestId(`skill-${SS_SOUL_MIRROR_SKILL_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: SS_SOUL_MIRROR_SKILL_ID,
    });

    // Target selection prompt
    await protocolHarness.pushServerMessage(soulMirrorTargetPrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    // Click on ally player card
    await page.getByTestId(`player-area-${ALLY_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });

  test('activate soulMirror and target self', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(soulMirrorScenario({ yellow_soul: 3 }));

    await page.getByTestId('action-skill').click();
    await page.getByTestId(`skill-${SS_SOUL_MIRROR_SKILL_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: SS_SOUL_MIRROR_SKILL_ID,
    });

    await protocolHarness.pushServerMessage(soulMirrorTargetPrompt());
    // Click on self player card
    await page.getByTestId(`player-area-${SS_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('soulMirror not available with insufficient yellow souls', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(soulMirrorScenario({ yellow_soul: 1 }));

    // Skill should not be available since yellow soul < 2
    await expect(page.getByTestId(`skill-${SS_SOUL_MIRROR_SKILL_ID}`)).not.toBeVisible();
  });
});
