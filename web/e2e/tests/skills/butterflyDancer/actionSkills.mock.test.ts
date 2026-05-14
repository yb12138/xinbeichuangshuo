import type { Page } from '@playwright/test';
import { test } from '../../../fixtures/protocolHarness.fixture';
import {
  BD_CHRYSALIS_SKILL_ID,
  ENEMY_PLAYER_ID,
  chrysalisScenario,
  reverseBranch2CostPrompt,
  reverseBranch2PickPrompt,
  reverseModePrompt,
  reverseScenario,
  reverseTargetPrompt,
} from '../../../scenarios/butterflyDancer';

async function activatePanelSkill(page: Page, skillId: string) {
  await page.getByTestId('action-skill').click();
  await page.getByTestId(`skill-${skillId}`).click();
}

async function clickOverlayOption(page: Page, selector: string) {
  const overlay = page.getByTestId('decision-overlay');
  const overlayVisible = await overlay.isVisible({ timeout: 1000 }).catch(() => false);
  if (overlayVisible) {
    await overlay.getByTestId(selector).click();
  } else {
    await page.getByTestId('prompt-dialog').getByTestId(selector).click();
  }
}

test.describe('butterfly dancer chrysalis protocol harness', () => {
  test('activate chrysalis skill', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(chrysalisScenario());

    await activatePanelSkill(page, BD_CHRYSALIS_SKILL_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: BD_CHRYSALIS_SKILL_ID,
    });
  });
});

test.describe('butterfly dancer reverse butterfly protocol harness', () => {
  test('reverse: branch 1 select target', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(reverseScenario());

    // Server pushes reverse mode prompt
    await protocolHarness.pushServerMessage(reverseModePrompt(true));

    // Select branch ① (option index 0)
    await clickOverlayOption(page, 'branch-option-0');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Server pushes target prompt for branch ①
    await protocolHarness.pushServerMessage(reverseTargetPrompt());

    // Click enemy player area to select target
    await page.getByTestId(`player-area-${ENEMY_PLAYER_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('reverse: branch 2 remove cocoons', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(reverseScenario({ canBranch2: true }));

    // Server pushes reverse mode prompt
    await protocolHarness.pushServerMessage(reverseModePrompt(true));

    // Select branch ② (option index 1)
    await clickOverlayOption(page, 'branch-option-1');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });

    // Server pushes branch ② cost prompt
    await protocolHarness.pushServerMessage(reverseBranch2CostPrompt(true));

    // Select "移除2个茧" (option index 0)
    await clickOverlayOption(page, 'branch-option-0');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // Server pushes cocoon pick prompt (choose_cards with cocoon labels, max=2)
    await protocolHarness.pushServerMessage(reverseBranch2PickPrompt());

    // Click two cocoon cover cards to select them, then confirm
    await page.getByTestId('hand-card-0').scrollIntoViewIfNeeded();
    await page.getByTestId('hand-card-0').click();
    await page.getByTestId('hand-card-1').scrollIntoViewIfNeeded();
    await page.getByTestId('hand-card-1').click();
    // Click "确认选择" button in expansion zone
    await page.locator('.expansion-cocoon-confirm-btn').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0, 1],
    });
  });

  test('reverse: branch 2 self-damage cost', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(reverseScenario({ canBranch2: true }));

    // Server pushes reverse mode prompt
    await protocolHarness.pushServerMessage(reverseModePrompt(true));

    // Select branch ②
    await clickOverlayOption(page, 'branch-option-1');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });

    // Server pushes branch ② cost prompt without cocoon option
    await protocolHarness.pushServerMessage(reverseBranch2CostPrompt(false));

    // Only "对自己造成4点法术伤害" is available (option index 0 for the single option)
    // When canRemoveCocoon=false, the only option id is '1' at index 0 in the filtered options
    // But the decision overlay shows it as branch-option-0 (first in inlinePrimaryButtons)
    // Actually, option id '1' maps to prompt.options index 0 (only one option: {id:'1', label:'自伤...'})
    await clickOverlayOption(page, 'branch-option-0');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('reverse: branch 1 only (no branch 2 available)', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(reverseScenario({ canBranch2: false }));

    // Server pushes reverse mode prompt without branch ②
    await protocolHarness.pushServerMessage(reverseModePrompt(false));

    // Only branch ① is available
    await clickOverlayOption(page, 'branch-option-0');
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });
});
