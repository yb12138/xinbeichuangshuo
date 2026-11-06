import { test, expect } from '../../../fixtures/protocolHarness.fixture';
import { reverseScenario, reverseBranch2PickPrompt, BD_PLAYER_ID } from '../../../scenarios/butterflyDancer';

const OUT = process.env.SHOT_DIR ?? '/tmp/xb-shots';

test('expansion shots', async ({ page, protocolHarness }) => {
  await page.setViewportSize({ width: Number(process.env.VW ?? 1600), height: Number(process.env.VH ?? 900) });
  const scenario = reverseScenario({ canBranch2: true });
  const me: any = (scenario.initialState as any).players.find((p: any) => p.id === BD_PLAYER_ID);
  me.exclusive_cards = [{ id: 'ex-1', name: '治愈之光', type: 'Magic', element: 'Light', exclusive_char1: 'butterfly_dancer', exclusive_skill1: '倒逆之蝶' }];
  me.field = [
    ...me.field,
    { card: { id: 'dm-1', name: '暗月牌', type: 'Attack', element: 'Dark' }, owner_id: BD_PLAYER_ID, source_id: BD_PLAYER_ID, mode: 'Cover', effect: 'MoonDarkMoon', field_hook: 'Manual', locked: false, duration: 0 },
    { card: { id: 'ss-1', name: '剑魂牌', type: 'Attack', element: 'Wind' }, owner_id: BD_PLAYER_ID, source_id: BD_PLAYER_ID, mode: 'Cover', effect: 'SwordSoul', field_hook: 'Manual', locked: false, duration: 0 },
  ];
  await protocolHarness.bootGame(scenario);
  await page.waitForTimeout(3500);
  await page.screenshot({ path: `${OUT}/1-collapsed.png` });
  const drawer = page.getByTestId('expansion-drawer');
  const chip = (name: RegExp) => page.getByRole('button', { name });
  // 悬停「茧」：只展开茧的牌
  await chip(/扩展区 茧/).hover();
  await page.waitForTimeout(500);
  await expect(drawer).toHaveAttribute('data-group', 'ButterflyCocoon');
  await expect(page.getByTestId('cover-card-0')).toBeVisible();
  await page.screenshot({ path: `${OUT}/2-expanded.png` });
  // 移到「暗月」：切换成暗月
  await chip(/扩展区 暗月/).hover();
  await expect(drawer).toHaveAttribute('data-group', 'MoonDarkMoon');
  await page.waitForTimeout(400);
  await page.screenshot({ path: `${OUT}/2b-darkmoon.png` });
  // 鼠标从图标移到展开的牌上，不应收起
  await page.getByTestId('cover-card-2').hover();
  await page.waitForTimeout(400);
  await expect(drawer).toBeVisible();
  // 移开自动收起
  await page.mouse.move(800, 300);
  await expect(drawer).toBeHidden();
  // 点击固定：移开后仍然展开
  await chip(/扩展区 专属牌/).click();
  await page.mouse.move(800, 300);
  await page.waitForTimeout(400);
  await expect(drawer).toHaveAttribute('data-group', 'Exclusive');
  await page.screenshot({ path: `${OUT}/2c-exclusive.png` });
  await drawer.getByRole('button', { name: '收起扩展区' }).click();
  await page.mouse.move(800, 300);
  await expect(drawer).toBeHidden();
  await protocolHarness.pushServerMessage(reverseBranch2PickPrompt());
  await page.waitForTimeout(600);
  await page.getByTestId('cover-card-0').click();
  await page.waitForTimeout(400);
  await page.screenshot({ path: `${OUT}/3-selecting.png` });
});

test('seat cover popover (others see backs)', async ({ page, protocolHarness }) => {
  await page.setViewportSize({ width: 1600, height: 900 });
  const scenario = reverseScenario({ canBranch2: true });
  // 他人视角：把敌人座位放上被服务端遮蔽过的茧（name=茧、无 id/type）
  const enemy: any = (scenario.initialState as any).players.find((p: any) => p.id !== BD_PLAYER_ID && p.camp === 'Blue');
  enemy.field = [0, 1, 2].map(() => ({
    card: { id: '', name: '茧', type: '', element: '', description: '盖牌（内容对他人不可见）' },
    owner_id: enemy.id, source_id: enemy.id, mode: 'Cover', effect: 'ButterflyCocoon', field_hook: 'Manual', locked: false, duration: 0,
  }));
  await protocolHarness.bootGame(scenario);
  await page.waitForTimeout(3000);
  const mark = page.locator(`[data-player-anchor="${enemy.id}"] .seat-mark--cover[data-mark-key="ButterflyCocoon"]`);
  await mark.hover();
  await expect(page.locator('.seat-cover-popover')).toBeVisible();
  await expect(page.locator('.seat-cover-popover .cover-card')).toHaveCount(3);
  await page.waitForTimeout(400);
  await page.screenshot({ path: `${OUT}/4-others-backs.png` });
  await page.mouse.move(800, 400);
  await expect(page.locator('.seat-cover-popover')).toBeHidden();
});
