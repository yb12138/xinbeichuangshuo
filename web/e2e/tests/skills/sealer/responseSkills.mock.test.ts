import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  magicSurgeScenario,
  magicSurgePrompt,
  fiveElementsBindCancelPrompt,
  sealTriggerPrompt,
  SEALER_WATER_SEAL_ID,
  enemyPerspectiveScenario,
} from '../../../scenarios/sealer';

test.describe('sealer magic surge protocol harness', () => {
  test('magic surge: confirm extra attack action', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(magicSurgeScenario());

    // Server pushes magic surge prompt after spell action ends
    await protocolHarness.pushServerMessage(magicSurgePrompt());

    // Click confirm button (branch_select: prompt-option-0)
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('prompt-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('magic surge: skip', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(magicSurgeScenario());

    await protocolHarness.pushServerMessage(magicSurgePrompt());

    // Click skip button (branch_select: prompt-option-1)
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('prompt-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});

test.describe('sealer five elements bind cancel protocol harness', () => {
  test('five elements bind cancel: enemy draws cards', async ({ page, protocolHarness }) => {
    // Boot from enemy perspective since prompt targets enemy
    await protocolHarness.bootGame(enemyPerspectiveScenario());

    // Server pushes cancel prompt to enemy (X=0, draw 2 cards)
    await protocolHarness.pushServerMessage(fiveElementsBindCancelPrompt(0));

    // Click draw option (branch_select: prompt-option-0)
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('prompt-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('five elements bind cancel: enemy skips', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(enemyPerspectiveScenario());

    await protocolHarness.pushServerMessage(fiveElementsBindCancelPrompt(0));

    // Click skip option (branch_select: prompt-option-1)
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('prompt-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });

  test('five elements bind cancel: X=2 draws 4 cards', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(enemyPerspectiveScenario());

    // Server pushes cancel prompt to enemy (X=2, draw 4 cards)
    await protocolHarness.pushServerMessage(fiveElementsBindCancelPrompt(2));

    // Click draw option (branch_select: prompt-option-0)
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('prompt-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });
});

test.describe('sealer seal trigger protocol harness', () => {
  test('seal trigger: enemy confirms damage', async ({ page, protocolHarness }) => {
    // Boot from enemy perspective since prompt targets enemy
    await protocolHarness.bootGame(enemyPerspectiveScenario());

    // Server pushes seal trigger prompt to enemy
    await protocolHarness.pushServerMessage(sealTriggerPrompt(SEALER_WATER_SEAL_ID));

    // Click confirm button (branch_select: prompt-option-0)
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('prompt-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });
});