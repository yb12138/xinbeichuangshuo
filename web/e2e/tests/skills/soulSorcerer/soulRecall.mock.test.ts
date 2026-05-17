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

    // Click skill button
    await page.getByTestId('action-skill').click();
    await page.getByTestId(`skill-${SS_SOUL_RECALL_SKILL_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: SS_SOUL_RECALL_SKILL_ID,
    });

    // Server pushes magic card selection prompt
    await protocolHarness.pushServerMessage(soulRecallPickPrompt());
    await expect(page.getByTestId('decision-overlay')).toBeVisible();
    // Select first magic card
    await page.getByTestId('card-slot-3').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [3],
    });
  });

  test('activate soulRecall and discard multiple magic cards', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(soulRecallScenario({ with_magic: true }));

    await page.getByTestId('action-skill').click();
    await page.getByTestId(`skill-${SS_SOUL_RECALL_SKILL_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: SS_SOUL_RECALL_SKILL_ID,
    });

    await protocolHarness.pushServerMessage(soulRecallPickPrompt());
    // Select both magic cards (multi-select)
    await page.getByTestId('card-slot-3').click();
    await page.getByTestId('card-slot-4').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [3, 4],
    });
  });

  test('soulRecall not available without magic cards', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(soulRecallScenario({ with_magic: false }));

    // Skill should not be available since no magic cards in hand
    await expect(page.getByTestId(`skill-${SS_SOUL_RECALL_SKILL_ID}`)).not.toBeVisible();
  });
});
