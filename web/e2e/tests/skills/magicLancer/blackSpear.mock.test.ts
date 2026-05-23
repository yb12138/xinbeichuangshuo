import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  blackSpearScenario,
  blackSpearXPrompt,
} from '../../../scenarios/magicLancer';

test.describe('magic lancer black spear protocol harness', () => {
  test('select X value from numeric picker', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(blackSpearScenario());

    // Server pushes black spear X selection prompt (response skill triggered by hit check)
    await protocolHarness.pushServerMessage(blackSpearXPrompt(3));

    // Verify numeric overlay is visible
    await expect(page.getByTestId('decision-overlay')).toBeVisible();

    // Select X=2 from the numeric overlay (second option, array index 1)
    await page.getByTestId('decision-overlay').getByTestId('numeric-option-2').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });

  test('select X=1 minimum value', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(blackSpearScenario());

    await protocolHarness.pushServerMessage(blackSpearXPrompt(2));

    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('decision-overlay').getByTestId('numeric-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });
});
