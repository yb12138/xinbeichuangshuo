import type { Page } from '@playwright/test';
import { expect, test } from '../../../fixtures/protocolHarness.fixture';
import {
  BLOOD_SWORD_SPIRIT_BLOOD_DYE_ROSE_ID,
  BLOOD_SWORD_SPIRIT_SCATTERING_DANCE_ID,
  ENEMY_PLAYER_ID,
  ALLY_PLAYER_ID,
  bloodDyeRoseScenario,
  bloodDyeRoseTargetPrompt,
  scatteringDanceScenario,
  scatteringDanceBranchPrompt,
  scatteringDanceDamageTargetPrompt,
  scatteringDanceHealTargetPrompt,
  roseCourtyardFieldScenario,
} from '../../../scenarios/bloodSwordSpirit';
import { syncStateMessage } from '../../../scenarios/builders';

async function activatePanelSkill(page: Page, skillId: string) {
  await page.getByTestId('action-magic').click();
  await page.getByTestId('action-skill').click();
  await page.getByTestId(`skill-${skillId}`).click();
}

async function clickOverlayOption(page: Page, selector: string) {
  const overlay = page.getByTestId('decision-overlay');
  const overlayVisible = await overlay.isVisible({ timeout: 1000 }).catch(() => false);
  if (overlayVisible) {
    await overlay.getByTestId(selector).click();
  } else {
    await page.getByTestId('prompt-dialog').getByTestId(selector).click();
  }
}

async function selectTarget(page: Page, targetId: string) {
  await page.getByTestId(`player-area-${targetId}`).click();
}

test.describe('blood sword spirit blood dye rose protocol harness', () => {
  test('blood dye rose: activate with blood>=3', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(bloodDyeRoseScenario());

    await activatePanelSkill(page, BLOOD_SWORD_SPIRIT_BLOOD_DYE_ROSE_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: BLOOD_SWORD_SPIRIT_BLOOD_DYE_ROSE_ID,
    });
  });

  test('blood dye rose: select targets', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(bloodDyeRoseScenario());

    await activatePanelSkill(page, BLOOD_SWORD_SPIRIT_BLOOD_DYE_ROSE_ID);

    // Server pushes target selection
    await protocolHarness.pushServerMessage(bloodDyeRoseTargetPrompt());

    // Select enemy (remove heal) and ally (gain heal)
    await selectTarget(page, ENEMY_PLAYER_ID);
    await selectTarget(page, ALLY_PLAYER_ID);
    await page.getByTestId('prompt-confirm-btn').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0, 1],
    });
  });
});

test.describe('blood sword spirit scattering dance protocol harness', () => {
  test('rose courtyard field vfx persists until the field effect is removed', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(roseCourtyardFieldScenario({ active: true }));

    await expect(page.getByTestId('rose-courtyard-vfx')).toBeVisible();
    await expect(page.locator('.center-battle .rose-courtyard-vfx')).toHaveCount(0);
    const centralOverlaps = await page.evaluate(() => {
      const center = document.querySelector('.center-battle')?.getBoundingClientRect();
      if (!center) return ['missing-center-battle'];
      return Array.from(document.querySelectorAll('.rose-courtyard-vfx__edge, .rose-courtyard-vfx__corner, .rose-courtyard-vfx__petal'))
        .filter((el) => {
          const rect = el.getBoundingClientRect();
          if (rect.width <= 0 || rect.height <= 0) return false;
          return rect.right > center.left &&
            rect.left < center.right &&
            rect.bottom > center.top &&
            rect.top < center.bottom;
        })
        .map((el) => el.className);
    });
    expect(centralOverlaps).toEqual([]);

    await protocolHarness.pushServerMessage(syncStateMessage(roseCourtyardFieldScenario({ active: false }).initialState));

    await expect(page.getByTestId('rose-courtyard-vfx')).toHaveCount(0);
  });

  test('scattering dance: damage branch', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(scatteringDanceScenario());

    await activatePanelSkill(page, BLOOD_SWORD_SPIRIT_SCATTERING_DANCE_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: BLOOD_SWORD_SPIRIT_SCATTERING_DANCE_ID,
    });

    // Server pushes branch selection
    await protocolHarness.pushServerMessage(scatteringDanceBranchPrompt());

    // Select damage branch
    await clickOverlayOption(page, 'prompt-option-damage');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Server pushes target selection
    await protocolHarness.pushServerMessage(scatteringDanceDamageTargetPrompt());

    // Select enemy target
    await selectTarget(page, ENEMY_PLAYER_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('scattering dance: heal transfer branch', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(scatteringDanceScenario());

    await activatePanelSkill(page, BLOOD_SWORD_SPIRIT_SCATTERING_DANCE_ID);

    await protocolHarness.pushServerMessage(scatteringDanceBranchPrompt());

    // Select heal transfer branch
    await clickOverlayOption(page, 'prompt-option-heal_transfer');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });

    // Server pushes target selection
    await protocolHarness.pushServerMessage(scatteringDanceHealTargetPrompt());

    // Select ally target
    await selectTarget(page, ALLY_PLAYER_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });
});
