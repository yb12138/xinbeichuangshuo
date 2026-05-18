import type { Page } from '@playwright/test';
import { test } from '../../../fixtures/protocolHarness.fixture';
import {
  ELF_ARCHER_ELF_RITUAL_ID,
  ALLY_PLAYER_ID,
  elfRitualScenario,
  elfRitualWithBlessingScenario,
} from '../../../scenarios/elfArcher';

async function activatePanelSkill(page: Page, skillId: string) {
  await page.getByTestId('action-skill').click();
  await page.getByTestId(`skill-${skillId}`).click();
}

// ============================================================
// Elf Ritual (精灵密仪) - 启动技能(大招)
// 后端通过 startup_skills 触发，目标通过 min_targets 处理
// ============================================================

test.describe('elf archer elf ritual protocol harness', () => {
  test('elf ritual: activate skill with gem', async ({ page, protocolHarness }) => {
    await protocolHarness.bootGame(elfRitualScenario());

    await activatePanelSkill(page, ELF_ARCHER_ELF_RITUAL_ID);
    await protocolHarness.expectSubmitAction({
      action_type: 'Skill',
      skill_id: ELF_ARCHER_ELF_RITUAL_ID,
    });

    // 祝福选择通过 min_targets 处理，后端不发送单独的 prompt
  });

  test('elf ritual: turn end heal ally', async ({ protocolHarness }) => {
    await protocolHarness.bootGame(elfRitualWithBlessingScenario());

    // 回合结束触发通过 response_skills 处理
    // 目标选择通过 min_targets 处理
    await protocolHarness.expectSubmitAction({
      action_type: 'UseSkill',
      skill_id: ELF_ARCHER_ELF_RITUAL_ID,
      targets: [{ target_user_id: ALLY_PLAYER_ID }],
    });
  });
});
