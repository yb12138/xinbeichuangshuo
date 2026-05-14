# E2E 测试基础设施

## 三个核心组件

| 组件 | 目标 | 实现方式 |
|------|------|---------|
| **Mock Server** | 模拟WebSocket通信 | 拦截WS消息或使用真实后端+测试API |
| **测试房间** | 自动创建房间、添加Bot玩家 | 直接注入前端状态或调用后端测试API |
| **状态注入** | 快速设置游戏状态 | 场景库 + Pinia状态注入 |

## 使用方式

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