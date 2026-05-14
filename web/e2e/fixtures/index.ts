import { test as base, Page } from '@playwright/test';

/**
 * E2E测试基础Fixture
 *
 * 提供三个核心能力：
 * 1. testRoom - 自动创建测试房间
 * 2. stateInject - 状态注入（加载预定义场景）
 * 3. mockWs - WebSocket mock（可选）
 */

// ========== 类型定义 ==========

export interface GameStateSnapshot {
  currentPlayerId: string;
  phase: 'turn_start' | 'main_action' | 'turn_end' | 'response';
  players: Record<string, PlayerSnapshot>;
  redMorale?: number;
  blueMorale?: number;
}

export interface PlayerSnapshot {
  id: string;
  name: string;
  camp: 'Red' | 'Blue';
  role: string;
  hp: number;
  heal: number;
  gem?: number;
  crystal?: number;
  hand: CardSnapshot[];
  deckSize: number;
  field?: FieldCardSnapshot[];
  tokens?: Record<string, number>;
  availableSkills?: SkillSnapshot[];
}

export interface CardSnapshot {
  id: string;
  name: string;
  type: 'Attack' | 'Magic';
  element: string;
  faction?: string;
}

export interface FieldCardSnapshot {
  id: string;
  mode: 'Effect' | 'Equipment';
  effect?: string;
}

export interface SkillSnapshot {
  id: string;
  title: string;
  can_use: boolean;
  cost_discards?: number;
}

export interface PromptDefinition {
  type: string;
  playerId: string;
  message: string;
  choiceType: string;
  options: PromptOption[];
  min?: number;
  max?: number;
  presentation?: { kind: string };
  extra?: Record<string, any>;
}

export interface PromptOption {
  id: string;
  label: string;
}

export interface RoomConfig {
  myPlayerId: string;
  myCamp: 'Red' | 'Blue';
  players: PlayerConfig[];
  startGame?: boolean;
}

export interface PlayerConfig {
  id: string;
  name?: string;
  camp: 'Red' | 'Blue';
  role?: string;
  isBot?: boolean;
}

export interface TestScenario {
  description: string;
  initialState: GameStateSnapshot;
  initialPrompt?: PromptDefinition;
}

// ========== Helper类 ==========

class TestRoomHelper {
  constructor(private page: Page) {}

  /**
   * 创建测试房间
   */
  async createRoom(config: RoomConfig): Promise<void> {
    await this.page.goto('/');
    await this.page.waitForLoadState('networkidle');

    // 注入房间和session状态
    await this.page.evaluate((cfg: RoomConfig) => {
      const win = window as any;

      // 确保Pinia已初始化
      if (!win.__PINIA__) {
        console.warn('Pinia not initialized');
        return;
      }

      const stores = win.__PINIA__.state.value;

      // 设置session
      if (stores.session) {
        stores.session.myPlayerId = cfg.myPlayerId;
        stores.session.myCamp = cfg.myCamp;
      }

      // 设置房间玩家列表
      if (stores.room) {
        stores.room.roomId = 'test-room-' + Date.now();
        stores.room.players = cfg.players.map(p => ({
          id: p.id,
          name: p.name || (p.isBot ? `Bot-${p.id}` : p.id),
          camp: p.camp,
          isBot: p.isBot ?? true,
          connected: true,
        }));
      }
    }, config);
  }
}

class StateInjectHelper {
  constructor(private page: Page) {}

  /**
   * 注入完整游戏状态
   */
  async injectGameState(state: GameStateSnapshot): Promise<void> {
    await this.page.evaluate((s: GameStateSnapshot) => {
      const win = window as any;
      if (!win.__PINIA__) return;

      const stores = win.__PINIA__.state.value;

      // 设置snapshot
      if (stores.snapshot) {
        stores.snapshot.currentPlayer = s.currentPlayerId;
        stores.snapshot.phase = s.phase;

        // 构建players对象
        const playersObj: Record<string, any> = {};
        for (const [id, player] of Object.entries(s.players)) {
          playersObj[id] = {
            id: player.id,
            name: player.name,
            camp: player.camp,
            role: player.role,
            hp: player.hp,
            heal: player.heal,
            gem: player.gem || 0,
            crystal: player.crystal || 0,
            hand: player.hand,
            deck: { count: player.deckSize },
            field: player.field || [],
            tokens: player.tokens || {},
            availableSkills: player.availableSkills || [],
          };
        }
        stores.snapshot.players = playersObj;

        // 设置士气
        if (s.redMorale) stores.snapshot.redMorale = s.redMorale;
        if (s.blueMorale) stores.snapshot.blueMorale = s.blueMorale;
      }
    }, state);
  }

  /**
   * 加载预定义场景
   */
  async loadScenario(scenarioId: string): Promise<void> {
    const scenario = SCENARIO_LIBRARY[scenarioId];
    if (!scenario) {
      throw new Error(`Unknown scenario: ${scenarioId}`);
    }

    await this.injectGameState(scenario.initialState);

    if (scenario.initialPrompt) {
      await this.injectPrompt(scenario.initialPrompt);
    }
  }

  /**
   * 注入prompt/中断
   */
  async injectPrompt(prompt: PromptDefinition): Promise<void> {
    await this.page.evaluate((p: PromptDefinition) => {
      const win = window as any;
      if (!win.__PINIA__) return;

      const stores = win.__PINIA__.state.value;
      if (stores.interrupt) {
        stores.interrupt.currentPrompt = {
          type: p.type,
          player_id: p.playerId,
          message: p.message,
          choice_type: p.choiceType,
          options: p.options,
          min: p.min || 1,
          max: p.max || 1,
          presentation: p.presentation,
          ...p.extra,
        };
      }
    }, prompt);
  }

  /**
   * 设置玩家资源
   */
  async setPlayerResources(playerId: string, resources: Partial<PlayerSnapshot>): Promise<void> {
    await this.page.evaluate(({ pid, r }: { pid: string; r: Partial<PlayerSnapshot> }) => {
      const win = window as any;
      if (!win.__PINIA__) return;

      const stores = win.__PINIA__.state.value;
      const player = stores.snapshot?.players?.[pid];
      if (!player) return;

      if (r.heal !== undefined) player.heal = r.heal;
      if (r.gem !== undefined) player.gem = r.gem;
      if (r.crystal !== undefined) player.crystal = r.crystal;
      if (r.hp !== undefined) player.hp = r.hp;
      if (r.hand !== undefined) player.hand = r.hand;
      if (r.tokens !== undefined) player.tokens = r.tokens;
      if (r.availableSkills !== undefined) player.availableSkills = r.availableSkills;
    }, { pid: playerId, r: resources });
  }

  /**
   * 设置当前回合
   */
  async setCurrentTurn(playerId: string, phase?: string): Promise<void> {
    await this.page.evaluate(({ pid, ph }: { pid: string; ph?: string }) => {
      const win = window as any;
      if (!win.__PINIA__) return;

      const stores = win.__PINIA__.state.value;
      if (stores.snapshot) {
        stores.snapshot.currentPlayer = pid;
        if (ph) stores.snapshot.phase = ph;
      }
    }, { pid: playerId, ph: phase });
  }

  /**
   * 清除prompt
   */
  async clearPrompt(): Promise<void> {
    await this.page.evaluate(() => {
      const win = window as any;
      if (!win.__PINIA__) return;

      const stores = win.__PINIA__.state.value;
      if (stores.interrupt) {
        stores.interrupt.currentPrompt = null;
      }
    });
  }

  /**
   * 等待状态稳定
   */
  async waitForState(timeout = 1000): Promise<void> {
    await this.page.waitForTimeout(timeout);
  }
}

// ========== 场景库 ==========

const SCENARIO_LIBRARY: Record<string, TestScenario> = {
  // 瘟疫法师 - 死亡之触
  'plague_death_touch_basic': {
    description: '死亡之触：3点治疗、2张火系牌',
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
          heal: 3,
          hand: [
            { id: 'card1', name: '火焰斩', type: 'Attack', element: 'Fire' },
            { id: 'card2', name: '火焰斩', type: 'Attack', element: 'Fire' },
            { id: 'card3', name: '圣光', type: 'Magic', element: 'Light' },
          ],
          deckSize: 15,
          tokens: {},
          availableSkills: [
            { id: 'plague_death_touch', title: '死亡之触', can_use: true },
            { id: 'plague_immortal', title: '不朽', can_use: true },
          ],
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
    description: '死亡之触：5点治疗、4张火系牌',
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
          ],
          deckSize: 10,
          availableSkills: [
            { id: 'plague_death_touch', title: '死亡之触', can_use: true },
          ],
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
    description: '死亡之触：治疗不足（1点）',
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
          heal: 1,
          hand: [
            { id: 'card1', name: '火焰斩', type: 'Attack', element: 'Fire' },
            { id: 'card2', name: '火焰斩', type: 'Attack', element: 'Fire' },
          ],
          deckSize: 20,
          availableSkills: [
            { id: 'plague_death_touch', title: '死亡之触', can_use: false },
          ],
        },
      },
    },
  },

  // 吟游诗人 - 激昂狂想曲
  'bard_rousing_confirm': {
    description: '激昂狂想曲：回合开始确认弹框',
    initialState: {
      currentPlayerId: 'holder',
      phase: 'turn_start',
      players: {
        'bard': {
          id: 'bard',
          name: '吟游诗人',
          camp: 'Red',
          role: 'bard',
          hp: 10,
          heal: 0,
          hand: [],
          deckSize: 15,
        },
        'holder': {
          id: 'holder',
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
      playerId: 'holder',
      message: '持有永恒乐章，回合开始时满足【激昂狂想曲】的发动条件，询问是否发动',
      choiceType: 'bd_rousing_confirm',
      options: [
        { id: 'confirm', label: '发动' },
        { id: 'cancel', label: '不发动' },
      ],
      presentation: { kind: 'response' },
    },
  },

  'bard_victory_confirm': {
    description: '胜利交响诗：回合结束确认弹框',
    initialState: {
      currentPlayerId: 'holder',
      phase: 'turn_end',
      players: {
        'bard': {
          id: 'bard',
          name: '吟游诗人',
          camp: 'Red',
          role: 'bard',
          hp: 10,
          heal: 0,
          hand: [],
          deckSize: 15,
        },
        'holder': {
          id: 'holder',
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
      },
    },
    initialPrompt: {
      type: 'confirm',
      playerId: 'holder',
      message: '持有永恒乐章，回合结束时满足【胜利交响诗】的发动条件，询问是否发动',
      choiceType: 'bd_victory_confirm',
      options: [
        { id: 'confirm', label: '发动' },
        { id: 'cancel', label: '不发动' },
      ],
      presentation: { kind: 'response' },
    },
  },
};

// ========== Fixture导出 ==========

export const test = base.extend<{
  testRoom: TestRoomHelper;
  stateInject: StateInjectHelper;
}>({
  testRoom: async ({ page }, use) => {
    const helper = new TestRoomHelper(page);
    await use(helper);
  },

  stateInject: async ({ page }, use) => {
    const helper = new StateInjectHelper(page);
    await use(helper);
  },
});

export { expect } from '@playwright/test';
export { TestRoomHelper, StateInjectHelper };
export { SCENARIO_LIBRARY };