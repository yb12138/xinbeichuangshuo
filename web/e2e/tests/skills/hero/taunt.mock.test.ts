import { expect, test } from '../../../fixtures/protocolHarness.fixture';
import {
  HERO_TAUNT_SKILL_ID,
  ENEMY_PLAYER_ID,
  ENEMY_2_PLAYER_ID,
  tauntScenario,
  tauntTargetPrompt,
} from '../../../scenarios/hero';

async function expectBoardAnchored(page: import('@playwright/test').Page) {
  const metrics = await page.evaluate(() => {
    const rect = (selector: string) => {
      const el = document.querySelector(selector);
      if (!el) return null;
      const r = el.getBoundingClientRect();
      return { top: r.top, bottom: r.bottom, height: r.height };
    };
    return {
      viewportHeight: window.innerHeight,
      bottomHud: rect('.bottom-hud'),
      stageMain: rect('.stage-main'),
    };
  });

  expect(metrics.bottomHud, 'bottom HUD should exist').not.toBeNull();
  expect(metrics.stageMain, 'center battle stage should exist').not.toBeNull();
  expect(metrics.bottomHud!.bottom).toBeGreaterThan(metrics.viewportHeight * 0.72);
  expect(metrics.stageMain!.height).toBeGreaterThan(metrics.viewportHeight * 0.32);
}

test.describe('hero taunt protocol harness', () => {
  test('activate taunt and select target', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(tauntScenario({ anger: 1 }));
    await expectBoardAnchored(page);


    await page.getByTestId('action-hub-trigger').click();
  await page.getByTestId('action-magic').click();
  await page.getByTestId('action-skill').click();
    await page.getByTestId(`skill-${HERO_TAUNT_SKILL_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: HERO_TAUNT_SKILL_ID,
    });

    // Target selection: click on enemy player card
    await protocolHarness.pushServerMessage(tauntTargetPrompt());
    await expectBoardAnchored(page);
    await page.getByTestId(`player-area-${ENEMY_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('activate taunt and select second enemy', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(tauntScenario({ anger: 2 }));
    await page.getByTestId('action-hub-trigger').click();
  await page.getByTestId('action-magic').click();
  await page.getByTestId('action-skill').click();
    await page.getByTestId(`skill-${HERO_TAUNT_SKILL_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: HERO_TAUNT_SKILL_ID,
    });

    // Target selection: click on second enemy
    await protocolHarness.pushServerMessage(tauntTargetPrompt());
    await page.getByTestId(`player-area-${ENEMY_2_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});
