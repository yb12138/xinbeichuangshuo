import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import {
  BSW_IAIJUTSU_STYLE_SKILL_ID,
  iaijutsuStyleStartupPrompt,
  iaijutsuStyleModePrompt,
  iaijutsuStyleDiscardPrompt,
  iaijutsuStyleScenario,
} from '../../../scenarios/beastSoul';

test.describe('beast samurai iaijutsu style protocol harness', () => {
  test('iaijutsu style: activate then choose draw', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(iaijutsuStyleScenario({ gems: 1 }));

    await page.getByTestId('action-skill').click();
    await page.getByTestId(`skill-${BSW_IAIJUTSU_STYLE_SKILL_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: BSW_IAIJUTSU_STYLE_SKILL_ID,
    });

    // 启动技能 PromptChooseSkill：仅 1 个启动技能时 overlay 仍含「发动 / 跳过」两按钮
    await protocolHarness.pushServerMessage(iaijutsuStyleStartupPrompt());
    await expect(page.getByTestId('skill-branch-overlay')).toBeVisible();
    await page.getByTestId('skill-branch-overlay').getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // 模式选择：摸 1 张 (option id "0")
    await protocolHarness.pushServerMessage(iaijutsuStyleModePrompt());
    await page.getByRole('button', { name: /摸牌/ }).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });
  });

  test('iaijutsu style: activate then choose discard (1 card)', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(iaijutsuStyleScenario({ gems: 1 }));

    await page.getByTestId('action-skill').click();
    await page.getByTestId(`skill-${BSW_IAIJUTSU_STYLE_SKILL_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: BSW_IAIJUTSU_STYLE_SKILL_ID,
    });

    await protocolHarness.pushServerMessage(iaijutsuStyleStartupPrompt());
    await page.getByTestId('skill-branch-overlay').getByTestId('branch-option-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [0],
    });

    // 模式选择：弃 1 张 (option id "1")
    await protocolHarness.pushServerMessage(iaijutsuStyleModePrompt());
    await page.getByRole('button', { name: /弃牌/ }).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });

    // 实际弃 1 张牌（min=max=1，与后端 buildDiscardPrompt 一致）
    await protocolHarness.pushServerMessage(iaijutsuStyleDiscardPrompt());
    await page.getByTestId('hand-card-0').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      card_ids: ['card_1'],
    });
  });

  test('iaijutsu style: skip startup prompt', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(iaijutsuStyleScenario({ gems: 1 }));

    await page.getByTestId('action-skill').click();
    await page.getByTestId(`skill-${BSW_IAIJUTSU_STYLE_SKILL_ID}`).click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: BSW_IAIJUTSU_STYLE_SKILL_ID,
    });

    await protocolHarness.pushServerMessage(iaijutsuStyleStartupPrompt());
    await page.getByTestId('skill-branch-overlay').getByTestId('prompt-option-skip').click();
    await protocolHarness.expectSubmitAction({
      action_type: 'Select',
      option_indexes: [1],
    });
  });
});
