# E2E 测试基础设施

## 三个核心组件

| 组件 | 目标 | 实现方式 |
|------|------|---------|
| **Mock Server** | 模拟WebSocket通信 | 拦截WS消息或使用真实后端+测试API |
| **测试房间** | 自动创建房间、添加Bot玩家 | 直接注入前端状态或调用后端测试API |
| **状态注入** | 快速设置游戏状态 | 场景库 + Pinia状态注入 |

## 使用方式

### 标准进房入口

前端测试需要进入真实游戏房间时，统一通过 `web/package.json` 中的测试对局脚本进入：

```bash
cd web
npm run dev:battle
```

该命令会调用 `scripts/open-test-battle.mjs`，通过后端测试 API 创建一个已选定 6 个角色并自动开始的 3v3 测试对局，然后打开首个玩家的房间页面。后续前端手动测试、浏览器验证和需要真实房间态的 E2E 调试都默认使用这个入口。

前置条件：

```bash
STARCUP_TEST_MODE=1 go run ./cmd/server
cd web && npm run dev
```

```typescript
// 测试文件中使用
test('死亡之触完整流程', async ({ page, testRoom, stateInject }) => {
  // 1. 创建测试房间
  await testRoom.createRoom({
    myPlayerId: 'plague_player',
    players: [...],
  });

  // 2. 加载预定义场景
  await stateInject.loadScenario('plague_death_touch_basic');

  // 3. 执行测试
  await page.click('[data-testid="action-skill"]');
  // ...
});
```

## 后端支持（可选）

如果使用真实后端，需要提供测试API：

```
POST /api/test/create_scenario
{
  "scenario_id": "plague_death_touch_basic",
  "players": [{ "id": "p1", "role": "plague_mage", "heal": 3, ... }]
}
```

---

详细设计见 [INFRASTRUCTURE.md](./INFRASTRUCTURE.md)
