import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  soulRecallScenario,
  soulRecallPickPrompt,
  SS_PLAYER_ID,
  SS_SOUL_RECALL_SKILL_ID,
} from '../../../scenarios/soulSorcerer';

test.describe('soulSorcerer soulRecall protocol harness', () => {
  test('activate soulRecall and discard magic card', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(soulRecallScenario({ with_magic: true }));


  await page.getByTestId('action-magic').click();
  await page.getByTestId('action-skill').click();
    await page.getByTestId(`skill-${SS_SOUL_RECALL_SKILL_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: SS_SOUL_RECALL_SKILL_ID,
    });

    // Server pushes magic card selection prompt (card_picker from hand)
    await protocolHarness.pushServerMessage(soulRecallPickPrompt());
    // Select first magic card (hand index 3 = card_4)
    await page.getByTestId('hand-card-3').click();
    await page.getByTestId('prompt-confirm-btn').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      card_ids: ['card_4'],
    });
  });

  test('activate soulRecall and discard multiple magic cards', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(soulRecallScenario({ with_magic: true }));
  await page.getByTestId('action-magic').click();
  await page.getByTestId('action-skill').click();
    await page.getByTestId(`skill-${SS_SOUL_RECALL_SKILL_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: SS_SOUL_RECALL_SKILL_ID,
    });

    await protocolHarness.pushServerMessage(soulRecallPickPrompt());
    // Select both magic cards (multi-select)
    await page.getByTestId('hand-card-3').click();
    await page.getByTestId('hand-card-4').click();
    await page.getByTestId('prompt-confirm-btn').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      card_ids: ['card_4', 'card_5'],
    });
  });

  test('soulRecall not available without magic cards', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(soulRecallScenario({ with_magic: false }));

    // Skill should not be available since no magic cards in hand
    await expect(page.getByTestId(`skill-${SS_SOUL_RECALL_SKILL_ID}`)).not.toBeVisible();
  });
});