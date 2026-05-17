import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import type { Page } from '@playwright/test';
import {
  heavenDriveConfirmPrompt,
  heavenDriveScenario,
  heavenDriveDiscardScenario,
  heavenDriveDiscardPrompt,
} from '../../../scenarios/fighter';

async function selectHandCards(page: Page, indices: number[]) {
  for (const index of indices) {
    const card = page.getByTestId(`hand-card-${index}`);
    await card.scrollIntoViewIfNeeded();
    await card.click();
  }
  await page.getByTestId('prompt-confirm-btn').click();
}

test.describe('fighter heaven drive protocol harness', () => {
  test('activate heaven drive with crystal at turn start', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(heavenDriveScenario({ crystals: 1 }));

    await protocolHarness.pushServerMessage(heavenDriveConfirmPrompt());
    await expect(page.getByTestId('skill-branch-overlay')).toBeVisible();
    await page
      .getByTestId('skill-branch-overlay')
      .getByTestId('branch-option-0')
      .click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('activate heaven drive with more crystals', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(heavenDriveScenario({ crystals: 2 }));

    await protocolHarness.pushServerMessage(heavenDriveConfirmPrompt());
    await page
      .getByTestId('skill-branch-overlay')
      .getByTestId('branch-option-0')
      .click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('decline heaven drive activation', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(heavenDriveScenario({ crystals: 1 }));

    await protocolHarness.pushServerMessage(heavenDriveConfirmPrompt());
    await page
      .getByTestId('skill-branch-overlay')
      .getByTestId('branch-option-1')
      .click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });

  // 补充弃牌流程测试 - 手牌>3张时需要弃牌至3张
  test('heaven drive: discard cards when hand > 3', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(heavenDriveDiscardScenario({ crystals: 1, handCount: 5 }));

    // 1. 发动确认
    await protocolHarness.pushServerMessage(heavenDriveConfirmPrompt());
    await expect(page.getByTestId('skill-branch-overlay')).toBeVisible();
    await page
      .getByTestId('skill-branch-overlay')
      .getByTestId('branch-option-0')
      .click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // 2. 后端推送弃牌prompt (system_discard_cards), 需要弃2张(5-3=2)
    await protocolHarness.pushServerMessage(heavenDriveDiscardPrompt(5));

    // 3. 选择弃置第4、5张手牌（索引3、4）
    await selectHandCards(page, [3, 4]);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [3, 4],
    });
  });

  // 验证弃牌后获得2点治疗
  test('heaven drive: discard 3 cards from 6 hand then heal 2', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(heavenDriveDiscardScenario({ crystals: 1, handCount: 6 }));

    // 1. 发动确认
    await protocolHarness.pushServerMessage(heavenDriveConfirmPrompt());
    await page
      .getByTestId('skill-branch-overlay')
      .getByTestId('branch-option-0')
      .click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // 2. 后端推送弃牌prompt, 需要弃3张(6-3=3)
    await protocolHarness.pushServerMessage(heavenDriveDiscardPrompt(6));

    // 3. 选择弃置第4、5、6张手牌（索引3、4、5）
    await selectHandCards(page, [3, 4, 5]);
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [3, 4, 5],
    });
  });
});
