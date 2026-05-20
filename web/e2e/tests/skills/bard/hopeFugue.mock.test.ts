import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  ALLY_PLAYER_ID,
  BD_HOPE_FUGUE_SKILL_ID,
  hopeDrawConfirmPrompt,
  hopeFugueScenario,
  hopeModePrompt,
  hopePlaceTargetPrompt,
  hopeTransferDiscardPrompt,
  hopeTransferTargetPrompt,
} from '../../../scenarios/bard';

test.describe('bard hope fugue protocol harness', () => {
  test('place eternal movement on ally (branch 0)', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(hopeFugueScenario());


  await page.getByTestId('action-magic').click();
  await page.getByTestId('action-skill').click();
    await page.getByTestId(`skill-${BD_HOPE_FUGUE_SKILL_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: BD_HOPE_FUGUE_SKILL_ID,
    });

    // Step 1: draw confirm
    await protocolHarness.pushServerMessage(hopeDrawConfirmPrompt());
    await page.getByTestId('prompt-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Step 2: mode branch (3 branches: place / transfer+heal / transfer+inspire)
    await protocolHarness.pushServerMessage(hopeModePrompt(false));
    await page.getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Step 3: place target (click on player card)
    await protocolHarness.pushServerMessage(hopePlaceTargetPrompt());
    await page.getByTestId(`player-area-${ALLY_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('transfer eternal movement with heal (branch 1)', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(hopeFugueScenario({ hasEternalHolder: true }));
  await page.getByTestId('action-magic').click();
  await page.getByTestId('action-skill').click();
    await page.getByTestId(`skill-${BD_HOPE_FUGUE_SKILL_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: BD_HOPE_FUGUE_SKILL_ID,
    });

    // draw confirm → yes
    await protocolHarness.pushServerMessage(hopeDrawConfirmPrompt());
    await page.getByTestId('prompt-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // mode → branch 1: transfer + heal
    await protocolHarness.pushServerMessage(hopeModePrompt(true));
    await page.getByTestId('branch-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });

    // transfer target (click on player card)
    await protocolHarness.pushServerMessage(hopeTransferTargetPrompt());
    await page.getByTestId(`player-area-${ALLY_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // discard card
    await protocolHarness.pushServerMessage(hopeTransferDiscardPrompt());
    await page.getByTestId('hand-card-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      card_ids: ['fire-atk-1'],
    });
  });

  test('transfer eternal movement with inspiration (branch 2)', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(hopeFugueScenario({ hasEternalHolder: true }));
  await page.getByTestId('action-magic').click();
  await page.getByTestId('action-skill').click();
    await page.getByTestId(`skill-${BD_HOPE_FUGUE_SKILL_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: BD_HOPE_FUGUE_SKILL_ID,
    });

    // draw confirm → no (skip draw)
    await protocolHarness.pushServerMessage(hopeDrawConfirmPrompt());
    await page.getByTestId('prompt-option-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });

    // mode → branch 2: transfer + inspiration
    await protocolHarness.pushServerMessage(hopeModePrompt(true));
    await page.getByTestId('branch-option-2').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [2],
    });

    // transfer target (click on player card)
    await protocolHarness.pushServerMessage(hopeTransferTargetPrompt());
    await page.getByTestId(`player-area-${ALLY_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // discard card
    await protocolHarness.pushServerMessage(hopeTransferDiscardPrompt());
    await page.getByTestId('hand-card-1').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      card_ids: ['fire-atk-2'],
    });
  });

  test('cancel at draw confirm stage', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(hopeFugueScenario());
  await page.getByTestId('action-magic').click();
  await page.getByTestId('action-skill').click();
    await page.getByTestId(`skill-${BD_HOPE_FUGUE_SKILL_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: BD_HOPE_FUGUE_SKILL_ID,
    });

    await protocolHarness.pushServerMessage(hopeDrawConfirmPrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    await page.getByTestId('prompt-option-1').click(); // Click "否" to decline
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});
