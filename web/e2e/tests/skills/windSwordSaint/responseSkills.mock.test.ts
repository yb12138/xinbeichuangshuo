import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  windFuryScenario,
  windFuryPrompt,
  swordShadowScenario,
  swordShadowPrompt,
  wssComboScenario,
  wssComboPrompt,
} from '../../../scenarios/windSwordSaint';

test.describe('wind sword saint wind fury protocol harness', () => {
  test('wind fury: confirm after attack action', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(windFuryScenario());

    // Server pushes wind fury prompt after attack action ends
    await protocolHarness.pushServerMessage(windFuryPrompt());

    // Click confirm (skill_choice with 1 skill: prompt-option-{skillId})
    await page.getByTestId('prompt-option-wind_fury').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('wind fury: skip', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(windFuryScenario());

    await protocolHarness.pushServerMessage(windFuryPrompt());

    // Click skip (skill_choice: prompt-option-skip)
    await page.getByTestId('prompt-option-skip').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});

test.describe('wind sword saint sword shadow protocol harness', () => {
  test('sword shadow: confirm with crystal', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(swordShadowScenario({ crystal: 1 }));

    // Server pushes sword shadow prompt after attack action ends
    await protocolHarness.pushServerMessage(swordShadowPrompt());

    // Click confirm (skill_choice with 1 skill: prompt-option-{skillId})
    await page.getByTestId('prompt-option-sword_shadow').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('sword shadow: skip', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(swordShadowScenario({ crystal: 1 }));

    await protocolHarness.pushServerMessage(swordShadowPrompt());

    // Click skip (skill_choice: prompt-option-skip)
    await page.getByTestId('prompt-option-skip').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});

test.describe('wind sword saint combo protocol harness', () => {
  test('combo: select wind fury when both available', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(wssComboScenario({ crystal: 1 }));

    // Server pushes combo prompt when both skills available
    await protocolHarness.pushServerMessage(wssComboPrompt());

    // Select wind fury (skill_choice with 2+ skills: skill-branch-overlay + branch-option-0)
    await expect(page.getByTestId('skill-branch-overlay')).toBeVisible();
    await page.getByTestId('skill-branch-overlay').getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('combo: select sword shadow when both available', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(wssComboScenario({ crystal: 1 }));

    await protocolHarness.pushServerMessage(wssComboPrompt());

    // Select sword shadow (skill_choice with 2+ skills: skill-branch-overlay + branch-option-1)
    await expect(page.getByTestId('skill-branch-overlay')).toBeVisible();
    await page.getByTestId('skill-branch-overlay').getByTestId('branch-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });

  test('combo: skip both', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(wssComboScenario({ crystal: 1 }));

    await protocolHarness.pushServerMessage(wssComboPrompt());

    // Select skip (skill_choice: prompt-option-skip inside skill-branch-overlay)
    await expect(page.getByTestId('skill-branch-overlay')).toBeVisible();
    await page.getByTestId('skill-branch-overlay').getByTestId('prompt-option-skip').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [2],
    });
  });
});