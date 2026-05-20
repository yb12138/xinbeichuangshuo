import { test } from '../../../fixtures/protocolHarness.fixture';
import {
  BARD_PLAYER_ID,
  BD_DISSONANCE_CHORD_SKILL_ID,
  ENEMY_PLAYER_ID,
  dissonanceChordScenario,
  dissonanceDiscardStepPrompt,
  dissonanceModePrompt,
  dissonanceTargetPrompt,
  dissonanceXPrompt,
} from '../../../scenarios/bard';

test.describe('bard dissonance chord protocol harness', () => {
  test('branch 0: both draw cards (instant after target)', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(dissonanceChordScenario());


  await page.getByTestId('action-magic').click();
  await page.getByTestId('action-skill').click();
    await page.getByTestId(`skill-${BD_DISSONANCE_CHORD_SKILL_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: BD_DISSONANCE_CHORD_SKILL_ID,
    });

    // X value selection (max = inspiration = 3)
    await protocolHarness.pushServerMessage(dissonanceXPrompt(3));
    // Select X=3 (click numeric button with label "3")
    await page.getByTestId('numeric-option-3').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1], // X=3 is the 2nd option (index 1, IDs: 2,3)
    });

    // Branch select: draw (branch 0) or discard (branch 1)
    await protocolHarness.pushServerMessage(dissonanceModePrompt(3));
    await page.getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Target selection (click on enemy player card)
    await protocolHarness.pushServerMessage(dissonanceTargetPrompt());
    await page.getByTestId(`player-area-${ENEMY_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [2],
    });
  });

  test('branch 1: both discard cards (sequential per actor)', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(dissonanceChordScenario());
  await page.getByTestId('action-magic').click();
  await page.getByTestId('action-skill').click();
    await page.getByTestId(`skill-${BD_DISSONANCE_CHORD_SKILL_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: BD_DISSONANCE_CHORD_SKILL_ID,
    });

    // X=2
    await protocolHarness.pushServerMessage(dissonanceXPrompt(3));
    await page.getByTestId('numeric-option-2').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Branch 1 = discard
    await protocolHarness.pushServerMessage(dissonanceModePrompt(2));
    await page.getByTestId('branch-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });

    // Target (click on enemy player card)
    await protocolHarness.pushServerMessage(dissonanceTargetPrompt());
    await page.getByTestId(`player-area-${ENEMY_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [2],
    });

    // Discard step: bard discards 1 card (X=2, n=X-1=1)
    await protocolHarness.pushServerMessage(dissonanceDiscardStepPrompt(BARD_PLAYER_ID, 'E2E Bard', 1));
    await page.getByTestId('hand-card-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      card_ids: ['fire-atk-1'],
    });

    // Discard step: enemy discards 1 card (simulated - click inline button since enemy is bot)
    await protocolHarness.pushServerMessage(dissonanceDiscardStepPrompt(ENEMY_PLAYER_ID, 'Enemy Bot', 1));
    await page.getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });
});
