# E2E 测试基础设施设计

## 概述

要让技能E2E测试实际运行，需要三个核心基础设施：
1. **Mock Server** - 模拟后端WebSocket通信
2. **测试房间** - 自动创建房间并填充玩家
3. **状态注入** - 快速设置特定游戏状态

## 1. Mock Server

### 设计目标
- 拦截WebSocket消息，模拟后端响应
- 提供确定性游戏状态
- 支持场景录制/回放

### 方案选择

#### 方案A: WebSocket Mock（推荐）
在前端拦截WebSocket，模拟后端消息流：

```typescript
// e2e/fixtures/mockServer.fixture.ts
import { test as base, Page } from '@playwright/test';

type MockServerConfig = {
  scenarios: Record<string, ScenarioDefinition>;
};

type ScenarioDefinition = {
  initialState: GameStateSnapshot;
  messageFlow: MockMessage[];
};

type MockMessage = {
  type: 'state_update' | 'interrupt' | 'error';
  delay?: number;
  payload: any;
};

export const test = base.extend<{
  mockServer: MockServerHelper;
}>({
  mockServer: async ({ page }, use) => {
    const helper = new MockServerHelper(page);
    await use(helper);
  },
});

class MockServerHelper {
  constructor(private page: Page) {}

  /**
   * 拦截WebSocket，注入mock响应
   */
  async interceptWebSocket() {
    await this.page.route('**/ws/**', async (route) => {
      // 模拟WebSocket连接
      const mockWs = await route.connectMockWebSocket();

      // 监听前端发送的消息
      mockWs.onMessage((msg) => {
        const data = JSON.parse(msg);

        // 根据消息类型返回mock响应
        const response = this.handleClientMessage(data);
        if (response) {
          mockWs.send(JSON.stringify(response));
        }
      });
    });
  }

  /**
   * 设置特定测试场景
   */
  async loadScenario(scenarioId: string) {
    await this.page.evaluate((id) => {
      (window as any).__mockScenario = id;
    }, scenarioId);
  }

  /**
   * 注入游戏状态
   */
  async injectState(state: GameStateSnapshot) {
    await this.page.evaluate((s) => {
      // 直接写入前端store
      const store = (window as any).__PINIA__;
      if (store) {
        store.state.value.snapshot = s;
      }
    }, state);
  }

  /**
   * 触发特定中断/prompt
   */
  async triggerPrompt(prompt: PromptDefinition) {
    await this.page.evaluate((p) => {
      const store = (window as any).__PINIA__;
      if (store?.interrupt) {
        store.interrupt.currentPrompt = p;
      }
    }, prompt);
  }

  private handleClientMessage(data: any): MockMessage | null {
    // 处理前端发送的各种消息类型
    switch (data.type) {
      case 'Select':
        return this.handleSelect(data);
      case 'Confirm':
        return this.handleConfirm(data);
      case 'Skill':
        return this.handleSkill(data);
      default:
        return null;
    }
  }

  private handleSelect(data: any): MockMessage {
    // 返回选择后的状态更新
    return {
      type: 'state_update',
      payload: { /* ... */ }
    };
  }
}
```

#### 方案B: 使用真实后端 + 测试API
连接真实后端，但使用测试专用API快速设置状态：

```typescript
// 后端需要提供测试API
POST /api/test/create_scenario
{
  "scenario_id": "plague_death_touch_basic",
  "players": [
    {"id": "p1", "role": "plague_mage", "heal": 3, "hand": [...]}
  ]
}
```

### Mock Server 核心消息类型

| 消息方向 | 类型 | 说明 |
|---------|------|------|
| 前端→Mock | `Select` | 选择选项/目标 |
| 前端→Mock | `Confirm` | 确认操作 |
| 前端→Mock | `Skill` | 发动技能 |
| 前端→Mock | `Cancel` | 取消操作 |
| Mock→前端 | `state_update` | 游戏状态更新 |
| Mock→前端 | `interrupt` | 弹框/prompt |
| Mock→前端 | `error` | 错误消息 |

## 2. 测试房间

### 设计目标
- 自动创建房间
- 自动添加Bot玩家填充位置
- 支持不同配置（3v3、角色分配等）

### 实现方案

```typescript
// e2e/fixtures/testRoom.fixture.ts
import { test as base, Page } from '@playwright/test';

export const test = base.extend<{
  testRoom: TestRoomHelper;
}>({
  testRoom: async ({ page, mockServer }, use) => {
    const helper = new TestRoomHelper(page, mockServer);
    await use(helper);
  },
});

class TestRoomHelper {
  constructor(private page: Page, private mockServer: MockServerHelper) {}

  /**
   * 创建测试房间并设置玩家
   */
  async createRoom(config: RoomConfig) {
    // 方案1: 如果有真实后端
    if (config.useRealBackend) {
      await this.createRealRoom(config);
    } else {
      // 方案2: 使用mock
      await this.createMockRoom(config);
    }
  }

  /**
   * Mock房间 - 直接设置前端状态
   */
  private async createMockRoom(config: RoomConfig) {
    await this.page.goto('/');

    // 设置房间状态
    await this.page.evaluate((cfg) => {
      const stores = (window as any).__PINIA__;
      if (!stores) return;

      // 设置session
      stores.session.myPlayerId = cfg.myPlayerId;
      stores.session.myCamp = cfg.myCamp;

      // 设置房间信息
      stores.room.roomId = 'test-room-001';
      stores.room.players = cfg.players.map(p => ({
        id: p.id,
        name: p.name || `Bot-${p.id}`,
        camp: p.camp,
        isBot: p.isBot ?? true,
      }));

      // 如果需要立即开始游戏
      if (cfg.startGame) {
        stores.snapshot.currentPlayer = cfg.myPlayerId;
        stores.snapshot.phase = 'main_action';
      }
    }, config);
  }

  /**
   * 真实房间 - 通过API创建
   */
  private async createRealRoom(config: RoomConfig) {
    // 调用后端测试API创建房间
    const response = await this.page.request.post('/api/test/create_room', {
      data: {
        room_id: 'test-room-001',
        players: config.players,
        auto_start: config.startGame,
      }
    });

    const roomData = await response.json();

    // 前端连接房间
    await this.page.goto(`/room/${roomData.room_id}`);
  }
}

type RoomConfig = {
  myPlayerId: string;
  myCamp: 'Red' | 'Blue';
  players: PlayerConfig[];
  startGame?: boolean;
  useRealBackend?: boolean;
};

type PlayerConfig = {
  id: string;
  name?: string;
  camp: string;
  role?: string;
  isBot?: boolean;
};
```

### 房间配置示例

```typescript
// 3v3 标准配置
const standardRoom: RoomConfig = {
  myPlayerId: 'p1',
  myCamp: 'Red',
  players: [
    { id: 'p1', camp: 'Red', isBot: false },
    { id: 'p2', camp: 'Red', isBot: true },
    { id: 'p3', camp: 'Red', isBot: true },
    { id: 'p4', camp: 'Blue', isBot: true },
    { id: 'p5', camp: 'Blue', isBot: true },
    { id: 'p6', camp: 'Blue', isBot: true },
  ],
  startGame: true,
};

// 单角色测试配置
const plagueMageTestRoom: RoomConfig = {
  myPlayerId: 'plague_player',
  myCamp: 'Red',
  players: [
    { id: 'plague_player', camp: 'Red', role: 'plague_mage', isBot: false },
    { id: 'enemy_1', camp: 'Blue', role: 'hero', isBot: true },
    { id: 'enemy_2', camp: 'Blue', role: 'fighter', isBot: true },
  ],
  startGame: true,
};
```

## 3. 状态注入

### 设计目标
- 快速设置特定游戏状态（不经过完整游戏流程）
- 设置技能发动所需的前置条件
- 支持场景库管理

### 实现方案

```typescript
// e2e/fixtures/stateInject.fixture.ts
import { test as base, Page } from '@playwright/test';

export const test = base.extend<{
  stateInject: StateInjectHelper;
}>({
  stateInject: async ({ page }, use) => {
    const helper = new StateInjectHelper(page);
    await use(helper);
  },
});

class StateInjectHelper {
  constructor(private page: Page) {}

  /**
   * 注入完整游戏状态
   */
  async injectGameState(state: GameStateSnapshot) {
    await this.page.evaluate((s) => {
      const stores = (window as any).__PINIA__;
      if (!stores) return;

      // 更新snapshot store
      stores.snapshot.players = s.players;
      stores.snapshot.currentPlayer = s.currentPlayerId;
      stores.snapshot.phase = s.phase;

      // 设置每个玩家的状态
      for (const player of s.players) {
        stores.snapshot.players[player.id] = {
          id: player.id,
          name: player.name,
          camp: player.camp,
          role: player.role,
          hp: player.hp,
          heal: player.heal,
          hand: player.hand,
          deck: { count: player.deckSize },
          field: player.field || [],
          tokens: player.tokens || {},
        };
      }
    }, state);
  }

  /**
   * 加载预定义测试场景
   */
  async loadScenario(scenarioId: string) {
    const scenario = SCENARIO_LIBRARY[scenarioId];
    if (!scenario) {
      throw new Error(`Unknown scenario: ${scenarioId}`);
    }

    await this.injectGameState(scenario.initialState);

    // 如果场景有初始prompt
    if (scenario.initialPrompt) {
      await this.injectPrompt(scenario.initialPrompt);
    }
  }

  /**
   * 注入特定prompt/中断
   */
  async injectPrompt(prompt: PromptDefinition) {
    await this.page.evaluate((p) => {
      const stores = (window as any).__PINIA__;
      if (!stores?.interrupt) return;

      stores.interrupt.currentPrompt = {
        type: p.type,
        player_id: p.playerId,
        message: p.message,
        options: p.options,
        choice_type: p.choiceType,
        min: p.min,
        max: p.max,
        ...p.extra,
      };
    }, prompt);
  }

  /**
   * 设置玩家资源
   */
  async setPlayerResources(playerId: string, resources: PlayerResources) {
    await this.page.evaluate(({ pid, r }) => {
      const stores = (window as any).__PINIA__;
      const player = stores.snapshot.players[pid];
      if (!player) return;

      if (r.heal !== undefined) player.heal = r.heal;
      if (r.gem !== undefined) player.gem = r.gem;
      if (r.crystal !== undefined) player.crystal = r.crystal;
      if (r.tokens) player.tokens = { ...player.tokens, ...r.tokens };
      if (r.hand) player.hand = r.hand;
      if (r.hp !== undefined) player.hp = r.hp;
    }, { pid: playerId, r: resources });
  }

  /**
   * 设置回合状态
   */
  async setTurnState(playerId: string, turnState: TurnStateConfig) {
    await this.page.evaluate(({ pid, ts }) => {
      const stores = (window as any).__PINIA__;
      stores.snapshot.currentPlayer = pid;
      stores.snapshot.phase = ts.phase || 'main_action';

      // 清除回合状态计数器
      if (stores.snapshot.players[pid]) {
        stores.snapshot.players[pid].turnState = {
          usedSkillCounts: {},
          performedAction: ts.performedAction || false,
        };
      }
    }, { pid: playerId, ts: turnState });
  }

  /**
   * 触发技能发动流程
   */
  async triggerSkillFlow(skillId: string, config: SkillFlowConfig) {
    // 设置技能发动所需的前置条件
    await this.setPlayerResources(config.playerId, config.resources);

    // 设置回合状态
    await this.setTurnState(config.playerId, {
      phase: 'main_action',
      performedAction: false,
    });

    // 确保技能在可用列表中
    await this.page.evaluate(({ skill, pid }) => {
      const stores = (window as any).__PINIA__;
      const player = stores.snapshot.players[pid];
      if (!player) return;

      player.availableSkills = [{
        id: skill,
        title: SKILL_NAMES[skill] || skill,
        can_use: true,
        cost_discards: 0,
      }];
    }, { skill: skillId, pid: config.playerId });
  }
}

// 状态类型定义
type GameStateSnapshot = {
  currentPlayerId: string;
  phase: string;
  players: Record<string, PlayerSnapshot>;
  redMorale?: number;
  blueMorale?: number;
};

type PlayerSnapshot = {
  id: string;
  name: string;
  camp: string;
  role: string;
  hp: number;
  heal: number;
  gem?: number;
  crystal?: number;
  hand: CardSnapshot[];
  deckSize: number;
  field?: FieldCard[];
  tokens?: Record<string, number>;
};

type CardSnapshot = {
  id: string;
  name: string;
  type: 'Attack' | 'Magic';
  element: string;
  faction?: string;
};

type PlayerResources = {
  heal?: number;
  gem?: number;
  crystal?: number;
  hp?: number;
  tokens?: Record<string, number>;
  hand?: CardSnapshot[];
};

type TurnStateConfig = {
  phase?: string;
  performedAction?: boolean;
};

type SkillFlowConfig = {
  playerId: string;
  resources: PlayerResources;
  targetId?: string;
};
```

### 场景库定义

```typescript
// e2e/scenarios/index.ts

/**
 * 预定义测试场景库
 * 每个场景包含初始状态和可选的初始prompt
 */
export const SCENARIO_LIBRARY: Record<string, TestScenario> = {
  // ========== 瘟疫法师 - 死亡之触 ==========

  'plague_death_touch_basic': {
    description: '死亡之触基本场景：3点治疗、2张火系牌',
    initialState: {
      currentPlayerId: 'plague_player',
      phase: 'main_action',
      players: {
        'plague_player': {
          id: 'plague_player',
          name: '瘟疫法师',
          camp: 'Red',
          role: 'plague_mage',
          hp: 10,
          heal: 3,  // X可选2或3
          hand: [
            { id: 'card1', name: '火焰斩', type: 'Attack', element: 'Fire' },
            { id: 'card2', name: '火焰斩', type: 'Attack', element: 'Fire' },
            { id: 'card3', name: '圣光', type: 'Magic', element: 'Light' },
          ],
          deckSize: 15,
          tokens: {},
        },
        'enemy_1': {
          id: 'enemy_1',
          name: '敌方1',
          camp: 'Blue',
          role: 'hero',
          hp: 10,
          heal: 0,
          hand: [],
          deckSize: 20,
        },
      },
    },
  },

  'plague_death_touch_high': {
    description: '死亡之触高X/Y场景：5点治疗、4张火系牌',
    initialState: {
      currentPlayerId: 'plague_player',
      phase: 'main_action',
      players: {
        'plague_player': {
          id: 'plague_player',
          name: '瘟疫法师',
          camp: 'Red',
          role: 'plague_mage',
          hp: 10,
          heal: 5,
          hand: [
            { id: 'card1', name: '火焰斩', type: 'Attack', element: 'Fire' },
            { id: 'card2', name: '火焰斩', type: 'Attack', element: 'Fire' },
            { id: 'card3', name: '火焰斩', type: 'Attack', element: 'Fire' },
            { id: 'card4', name: '火焰斩', type: 'Attack', element: 'Fire' },
            { id: 'card5', name: '圣光', type: 'Magic', element: 'Light' },
          ],
          deckSize: 10,
        },
        'enemy_1': {
          id: 'enemy_1',
          name: '敌方1',
          camp: 'Blue',
          role: 'hero',
          hp: 10,
          heal: 0,
          hand: [],
          deckSize: 20,
        },
      },
    },
  },

  'plague_death_touch_insufficient_heal': {
    description: '死亡之触：治疗不足场景',
    initialState: {
      currentPlayerId: 'plague_player',
      phase: 'main_action',
      players: {
        'plague_player': {
          id: 'plague_player',
          name: '瘟疫法师',
          camp: 'Red',
          role: 'plague_mage',
          hp: 10,
          heal: 1,  // 不满足>=2条件
          hand: [
            { id: 'card1', name: '火焰斩', type: 'Attack', element: 'Fire' },
            { id: 'card2', name: '火焰斩', type: 'Attack', element: 'Fire' },
          ],
          deckSize: 20,
        },
      },
    },
  },

  'plague_death_touch_immortal': {
    description: '死亡之触：不朽抑制验证场景',
    initialState: {
      currentPlayerId: 'plague_player',
      phase: 'main_action',
      players: {
        'plague_player': {
          id: 'plague_player',
          name: '瘟疫法师',
          camp: 'Red',
          role: 'plague_mage',
          hp: 10,
          heal: 2,
          hand: [
            { id: 'card1', name: '火焰斩', type: 'Attack', element: 'Fire' },
            { id: 'card2', name: '火焰斩', type: 'Attack', element: 'Fire' },
          ],
          deckSize: 15,
          // 不朽激活状态（需要后端配合设置）
          tokens: { 'plague_immortal_active': 1 },
        },
        'enemy_1': {
          id: 'enemy_1',
          name: '敌方1',
          camp: 'Blue',
          role: 'hero',
          hp: 10,
          heal: 0,
          hand: [],
          deckSize: 20,
        },
      },
    },
  },

  // ========== 吟游诗人 - 激昂狂想曲 ==========

  'bard_rousing_rhapsody': {
    description: '激昂狂想曲：永恒乐章持有者回合开始',
    initialState: {
      currentPlayerId: 'eternal_holder',
      phase: 'turn_start',
      players: {
        'bard_player': {
          id: 'bard_player',
          name: '吟游诗人',
          camp: 'Red',
          role: 'bard',
          hp: 10,
          heal: 0,
          hand: [],
          deckSize: 15,
        },
        'eternal_holder': {
          id: 'eternal_holder',
          name: '永恒乐章持有者',
          camp: 'Red',
          role: 'hero',
          hp: 10,
          heal: 0,
          hand: [
            { id: 'eternal', name: '永恒乐章', type: 'Magic', element: 'Dark' },
          ],
          deckSize: 15,
        },
        'enemy_1': {
          id: 'enemy_1',
          name: '敌方1',
          camp: 'Blue',
          role: 'fighter',
          hp: 10,
          heal: 0,
          hand: [],
          deckSize: 20,
        },
        'enemy_2': {
          id: 'enemy_2',
          name: '敌方2',
          camp: 'Blue',
          role: 'assassin',
          hp: 10,
          heal: 0,
          hand: [],
          deckSize: 20,
        },
      },
    },
    initialPrompt: {
      type: 'confirm',
      playerId: 'eternal_holder',
      message: '持有永恒乐章，回合开始时满足【激昂狂想曲】的发动条件，询问是否发动',
      choiceType: 'bd_rousing_confirm',
      options: [
        { id: 'confirm', label: '发动' },
        { id: 'cancel', label: '不发动' },
      ],
    },
  },

  // 更多场景...
};

type TestScenario = {
  description: string;
  initialState: GameStateSnapshot;
  initialPrompt?: PromptDefinition;
};

type PromptDefinition = {
  type: string;
  playerId: string;
  message: string;
  choiceType: string;
  options: PromptOption[];
  min?: number;
  max?: number;
  extra?: Record<string, any>;
};

type PromptOption = {
  id: string;
  label: string;
};

// 技能名称映射
const SKILL_NAMES: Record<string, string> = {
  'plague_death_touch': '死亡之触',
  'plague_immortal': '不朽',
  'bd_rousing_rhapsody': '激昂狂想曲',
  'bd_victory_symphony': '胜利交响诗',
  // 更多...
};
```

## 整合使用示例

```typescript
// e2e/tests/skills/plague_mage/deathTouch.test.ts
import { test, expect } from '../../fixtures/testRoom.fixture';
import { StateInjectHelper } from '../../fixtures/stateInject.fixture';

test.describe('Plague Mage - 死亡之触', () => {

  test.use({
    // 自动创建mock房间
    testRoom: {
      myPlayerId: 'plague_player',
      players: [...],
      startGame: true,
    },
  });

  test('完整流程', async ({ page, stateInject }) => {
    // 加载预定义场景
    await stateInject.loadScenario('plague_death_touch_basic');

    // 或者手动设置状态
    await stateInject.triggerSkillFlow('plague_death_touch', {
      playerId: 'plague_player',
      resources: {
        heal: 3,
        hand: [
          { id: 'c1', name: '火焰斩', type: 'Attack', element: 'Fire' },
          { id: 'c2', name: '火焰斩', type: 'Attack', element: 'Fire' },
        ],
      },
    });

    // 现在可以执行实际测试
    await page.click('[data-testid="action-skill"]');
    await page.waitForSelector('[data-testid="skill-select-panel"]');
    await page.click('[data-testid="skill-plague_death_touch"]');

    // 验证元素选择prompt
    await page.waitForSelector('[data-testid="decision-overlay"]');
    expect(await page.locator('.overlay-panel-header h2').textContent())
      .toContain('死亡之触');

    // 选择火系
    await page.click('[data-testid="branch-option-0"]');

    // ...继续测试流程
  });
});
```

## 后端支持需求

如果选择使用真实后端 + 测试API，需要后端提供：

```go
// internal/api/test_api.go (建议新增)

// CreateTestScenario 创建测试场景
// POST /api/test/create_scenario
func CreateTestScenario(c *gin.Context) {
  var req CreateScenarioRequest
  if err := c.BindJSON(&req); err != nil {
    c.JSON(400, gin.H{"error": err.Error()})
    return
  }

  // 创建房间
  room := game.NewTestRoom(req.RoomId)

  // 设置玩家和状态
  for _, player := range req.Players {
    p := room.AddTestPlayer(player.Id, player.Role, player.Camp)
    p.Heal = player.Heal
    p.Hp = player.Hp
    p.Hand = player.Hand
    // ...
  }

  // 设置回合状态
  room.SetCurrentPlayer(req.CurrentPlayerId)
  room.SetPhase(req.Phase)

  // 如果需要触发prompt
  if req.TriggerPrompt != nil {
    room.PushInterrupt(&model.Interrupt{
      Type: model.InterruptChoice,
      PlayerID: req.TriggerPrompt.PlayerId,
      Context: req.TriggerPrompt.Context,
    })
  }

  c.JSON(200, gin.H{
    "room_id": room.ID,
    "ws_url": fmt.Sprintf("ws://localhost:8080/ws/%s", room.ID),
  })
}

type CreateScenarioRequest struct {
  RoomId          string
  Players         []TestPlayerConfig
  CurrentPlayerId string
  Phase           string
  TriggerPrompt   *TestPromptConfig
}

type TestPlayerConfig struct {
  Id     string
  Role   string
  Camp   string
  Hp     int
  Heal   int
  Hand   []CardConfig
  Tokens map[string]int
}
```

## 文件结构最终版

```
e2e/
├── playwright.config.ts
├── README.md
├── fixtures/
│   ├── testRoom.fixture.ts     # 测试房间创建
│   ├── stateInject.fixture.ts  # 状态注入
│   └── mockServer.fixture.ts   # WebSocket mock
├── helpers/
│   ├── promptHelpers.ts
│   ├── actionHelpers.ts
│   └── gameStateHelpers.ts
│   └── index.ts
├── scenarios/
│   ├── index.ts                # 场景库主入口
│   ├── plague_mage.ts          # 瘟疫法师场景
│   ├── bard.ts                  # 吟游诗人场景
│   └── ...                      # 其他角色场景
└── tests/
    ├── skills/
    │   ├── plague_mage/
    │   │   └── deathTouch.test.ts
    │   ├── bard/
    │   │   ├── rousingRhapsody.test.ts
    │   │   └── ...
    │   └── ...
    └── integration/
        └── fullGameFlow.test.ts  # 完整游戏流程测试
```