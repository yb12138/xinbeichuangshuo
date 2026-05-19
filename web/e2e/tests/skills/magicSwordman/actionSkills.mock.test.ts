import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  asuraComboScenario,
  asuraComboPrompt,
  asuraComboDiscardPrompt,
  underworldTremorScenario,
  underworldTremorPrompt,
} from '../../../scenarios/magicSwordman';

test.describe('magic swordman asura combo protocol harness', () => {
  test('asura combo: confirm after damage >=2', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(asuraComboScenario());

    // Server pushes asura combo prompt after attack ends
    await protocolHarness.pushServerMessage(asuraComboPrompt());

    // Click confirm button (branch_select: branch-option-0)
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('decision-overlay').getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Server pushes discard selection
    await protocolHarness.pushServerMessage(asuraComboDiscardPrompt());

    // Select fire card from hand (card_picker: hand-card-0 auto-submits for min=1,max=1)
    await page.getByTestId('hand-card-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      card_ids: ['ms-fire-attack-1'],
    });
  });

  test('asura combo: skip', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(asuraComboScenario());

    await protocolHarness.pushServerMessage(asuraComboPrompt());

    // Click skip button (branch_select: branch-option-1)
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('decision-overlay').getByTestId('branch-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});

test.describe('magic swordman underworld tremor protocol harness', () => {
  test('underworld tremor: confirm with gem', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(underworldTremorScenario());

    // Server pushes underworld tremor prompt before attack
    await protocolHarness.pushServerMessage(underworldTremorPrompt());

    // Click confirm button (branch_select: branch-option-0)
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('decision-overlay').getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('underworld tremor: skip', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(underworldTremorScenario());

    await protocolHarness.pushServerMessage(underworldTremorPrompt());

    // Click skip button (branch_select: branch-option-1)
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('decision-overlay').getByTestId('branch-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});