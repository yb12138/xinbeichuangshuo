import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  calmMindConfirmPrompt,
  calmMindScenario,
} from '../../../scenarios/hero';

test.describe('hero calm mind protocol harness', () => {
  test('activate calm mind with 4 wisdom', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(calmMindScenario({ wisdom: 4 }));

    // Server pushes calm mind confirm prompt (triggered on attack initiation)
    await protocolHarness.pushServerMessage(calmMindConfirmPrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    // Click "发动"
    await page.getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
    // Attack becomes uncounterable, no further prompts
  });

  test('decline calm mind activation', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(calmMindScenario({ wisdom: 5 }));

    await protocolHarness.pushServerMessage(calmMindConfirmPrompt());
    await page.getByTestId('branch-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});
