import { test } from '../../../fixtures/protocolHarness.fixture';
import {
  ENEMY_PLAYER_ID,
  witherConfirmPrompt,
  witherTargetPrompt,
  witherScenario,
} from '../../../scenarios/butterflyDancer';

test.describe('butterfly dancer wither protocol harness', () => {
  test('wither: confirm activate then select target', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(witherScenario());

    // Server pushes wither confirm prompt (triggered by cocoon removal)
    await protocolHarness.pushServerMessage(witherConfirmPrompt());

    // Click "发动凋零" (option index 0)
    const overlay = page.getByTestId('decision-overlay');
    const overlayVisible = await overlay.isVisible({ timeout: 1000 }).catch(() => false);
    if (overlayVisible) {
      await overlay.getByTestId('branch-option-0').click();
    } else {
      await page.getByTestId('prompt-dialog').getByTestId('branch-option-0').click();
    }
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Server pushes wither target prompt
    await protocolHarness.pushServerMessage(witherTargetPrompt());

    // Click enemy player area to select target
    await page.getByTestId(`player-area-${ENEMY_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('wither: skip activation', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(witherScenario());

    // Server pushes wither confirm prompt
    await protocolHarness.pushServerMessage(witherConfirmPrompt());

    // Click "不发动" (option index 1)
    const overlay = page.getByTestId('decision-overlay');
    const overlayVisible = await overlay.isVisible({ timeout: 1000 }).catch(() => false);
    if (overlayVisible) {
      await overlay.getByTestId('branch-option-1').click();
    } else {
      await page.getByTestId('prompt-dialog').getByTestId('branch-option-1').click();
    }
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});
