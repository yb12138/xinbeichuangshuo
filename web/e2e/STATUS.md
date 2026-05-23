# E2E 测试现状总结

## 当前问题

测试失败的根本原因：**前端需要真实WebSocket连接才能进入游戏界面**

```
大厅界面 → 创建房间 → WebSocket连接 → 游戏界面
```

状态注入（修改Pinia）只能在**已进入游戏界面**后生效。

## 解决方案

### 方案1: 后端测试API（推荐）

后端新增测试场景创建API，直接返回游戏状态：

```go
// POST /api/test/create_scenario
func CreateTestScenario(c *gin.Context) {
  // 创建房间、设置玩家、返回WebSocket URL
  room := createTestRoom(req.ScenarioId)
  c.JSON(200, gin.H{
    "room_id": room.ID,
    "ws_url": room.WebSocketURL(),
  })
}
```

前端测试：
```typescript
const response = await page.request.post('/api/test/create_scenario', {
  data: { scenario_id: 'plague_death_touch_basic' }
});
const wsUrl = response.json().ws_url;
await page.goto(wsUrl); // 直接进入游戏界面
```

### 方案2: 前端调试入口

前端支持 `?debug_game=1` 参数，直接渲染游戏界面（不验证WebSocket）：

```typescript
// App.vue
if (route.query.debug_game === '1') {
  // 跳过大厅，直接渲染GameBoard
  renderGameBoard();
}
```

测试：
```typescript
await page.goto('/?debug_game=1');
await stateInject.loadScenario('plague_death_touch_basic');
// 现在状态注入可以生效
```

### 方案3: Mock WebSocket

拦截WebSocket连接，返回mock消息：

```typescript
await page.route('**/ws/**', route => {
  route.connectMockWebSocket({
    onMessage: (msg) => handleMockMessage(msg),
  });
});
```

复杂度较高，需要模拟完整消息流。

## 当前文件结构

```
e2e/
├── playwright.config.ts
├── README.md              # 使用说明
├── INFRASTRUCTURE.md      # 详细设计文档
├── fixtures/
│   ├── index.ts           # 整合的fixture（testRoom, stateInject）
│   ├── game.fixture.ts    # （已废弃，移到index.ts）
│   └── mockServer.fixture.ts
├── helpers/
│   ├── promptHelpers.ts
│   ├── actionHelpers.ts
│   ├── gameStateHelpers.ts
│   └── index.ts
└── tests/
    ├── skills/
    │   ├── plague_mage/deathTouch.test.ts
    │   └── bard/*.test.ts
    └── fixtures/          # （错误位置，应该在上层）
```

## 建议下一步

1. **与后端协商** - 是否可以提供测试API
2. **前端添加调试入口** - 低成本快速验证方案
3. **完善场景库** - 预定义更多测试场景

## 已完成工作

- ✅ Playwright框架安装配置
- ✅ Helper抽象层（PromptHelper, ActionHelper, GameStateHelper）
- ✅ data-testid属性添加到关键组件
- ✅ 场景库设计（SCENARIO_LIBRARY）
- ✅ 测试文件编写（流程文档化）

## 待完成工作

- ⏳ 真实游戏环境接入
- ⏳ WebSocket Mock实现
- ⏳ 更多角色场景库
- ⏳ 测试验证实际运行