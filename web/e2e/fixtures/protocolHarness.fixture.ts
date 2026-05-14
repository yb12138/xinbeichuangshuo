import { expect, test as base, type Page } from '@playwright/test';
import type { ClientActionRequest, WsMessage } from '../../src/network/protocol';
import {
  roomEvent,
  syncStateMessage,
  type ProtocolHarnessScenario,
} from '../scenarios/builders';

interface RecordedClientMessage {
  raw: string;
  message: WsMessage;
  timestamp: number;
}

type ClientMessagePredicate = (message: WsMessage, record: RecordedClientMessage) => boolean;

type SubmitActionExpectation = Partial<Pick<
  ClientActionRequest,
  'action_type' | 'skill_id' | 'option_indexes' | 'targets' | 'used_card_uuids' | 'target_ref'
>>;

export interface ProtocolHarness {
  bootGame: (scenario: ProtocolHarnessScenario) => Promise<void>;
  pushServerMessage: (message: WsMessage) => Promise<void>;
  waitForClientMessage: (
    predicate: ClientMessagePredicate,
    options?: { timeout?: number }
  ) => Promise<WsMessage>;
  expectSubmitAction: (
    expected: SubmitActionExpectation,
    options?: { timeout?: number }
  ) => Promise<ClientActionRequest>;
}

async function installFakeWebSocket(page: Page) {
  await page.addInitScript(() => {
    type Recorded = {
      raw: string;
      message: unknown;
      timestamp: number;
    };

    type FakeSocket = WebSocket & {
      __receiveServerMessage: (message: unknown) => void;
    };

    const sockets: FakeSocket[] = [];
    const sentMessages: Recorded[] = [];

    function isGameSocket(socket: FakeSocket) {
      return /\/ws(?:\?|$)/.test(socket.url);
    }

    class FakeWebSocket extends EventTarget {
      static readonly CONNECTING = 0;
      static readonly OPEN = 1;
      static readonly CLOSING = 2;
      static readonly CLOSED = 3;

      readonly CONNECTING = 0;
      readonly OPEN = 1;
      readonly CLOSING = 2;
      readonly CLOSED = 3;

      readonly url: string;
      readonly protocol = '';
      readonly extensions = '';
      bufferedAmount = 0;
      binaryType: BinaryType = 'blob';
      readyState = FakeWebSocket.CONNECTING;

      onopen: ((event: Event) => void) | null = null;
      onmessage: ((event: MessageEvent) => void) | null = null;
      onerror: ((event: Event) => void) | null = null;
      onclose: ((event: CloseEvent) => void) | null = null;

      constructor(url: string | URL) {
        super();
        this.url = String(url);
        sockets.push(this as unknown as FakeSocket);

        window.setTimeout(() => {
          if (this.readyState !== FakeWebSocket.CONNECTING) return;
          this.readyState = FakeWebSocket.OPEN;
          const event = new Event('open');
          this.onopen?.(event);
          this.dispatchEvent(event);
        }, 0);
      }

      send(data: string | ArrayBufferLike | Blob | ArrayBufferView) {
        if (this.readyState !== FakeWebSocket.OPEN) {
          throw new DOMException('Fake WebSocket is not open', 'InvalidStateError');
        }
        const raw = typeof data === 'string' ? data : String(data);
        let message: unknown = raw;
        try {
          message = JSON.parse(raw);
        } catch {
          // Keep the raw payload for debugging non-JSON sends.
        }
        sentMessages.push({
          raw,
          message,
          timestamp: Date.now(),
        });
      }

      close(code = 1000, reason = '') {
        if (this.readyState === FakeWebSocket.CLOSED || this.readyState === FakeWebSocket.CLOSING) return;
        this.readyState = FakeWebSocket.CLOSING;
        window.setTimeout(() => {
          this.readyState = FakeWebSocket.CLOSED;
          const event = new CloseEvent('close', {
            code,
            reason,
            wasClean: true,
          });
          this.onclose?.(event);
          this.dispatchEvent(event);
        }, 0);
      }

      __receiveServerMessage(message: unknown) {
        if (this.readyState !== FakeWebSocket.OPEN) {
          throw new Error('Cannot push a server message before the fake socket is open');
        }
        const event = new MessageEvent('message', {
          data: typeof message === 'string' ? message : JSON.stringify(message),
        });
        this.onmessage?.(event);
        this.dispatchEvent(event);
      }
    }

    Object.defineProperty(window, 'WebSocket', {
      configurable: true,
      writable: true,
      value: FakeWebSocket,
    });

    (window as any).__E2E_WS__ = {
      sentMessages: () => sentMessages.map((entry) => ({ ...entry })),
      clearSentMessages: () => {
        sentMessages.splice(0, sentMessages.length);
      },
      hasOpenSocket: () => sockets.some((socket) => isGameSocket(socket) && socket.readyState === FakeWebSocket.OPEN),
      pushServerMessage: (message: unknown) => {
        const socket = [...sockets]
          .reverse()
          .find((item) => isGameSocket(item) && item.readyState === FakeWebSocket.OPEN);
        if (!socket) {
          throw new Error('No open fake WebSocket found');
        }
        socket.__receiveServerMessage(message);
      },
    };
  });
}

function valuesMatch(actual: unknown, expected: unknown): boolean {
  if (expected === undefined) return true;
  return JSON.stringify(actual) === JSON.stringify(expected);
}

function submitActionMatches(actual: ClientActionRequest, expected: SubmitActionExpectation): boolean {
  return (
    valuesMatch(actual.action_type, expected.action_type) &&
    valuesMatch(actual.skill_id, expected.skill_id) &&
    valuesMatch(actual.option_indexes, expected.option_indexes) &&
    valuesMatch(actual.targets, expected.targets) &&
    valuesMatch(actual.used_card_uuids, expected.used_card_uuids) &&
    valuesMatch(actual.target_ref, expected.target_ref)
  );
}

async function getRecordedMessages(page: Page): Promise<RecordedClientMessage[]> {
  return page.evaluate(() => {
    const harness = (window as any).__E2E_WS__;
    return harness?.sentMessages?.() ?? [];
  });
}

export const test = base.extend<{
  protocolHarness: ProtocolHarness;
}>({
  protocolHarness: async ({ page }, use) => {
    let installed = false;

    const ensureInstalled = async () => {
      if (installed) return;
      await installFakeWebSocket(page);
      installed = true;
    };

    const pushServerMessage = async (message: WsMessage) => {
      await page.evaluate((serverMessage) => {
        (window as any).__E2E_WS__.pushServerMessage(serverMessage);
      }, message);
    };

    const waitForClientMessage = async (
      predicate: ClientMessagePredicate,
      options: { timeout?: number } = {}
    ): Promise<WsMessage> => {
      let matched: WsMessage | null = null;
      await expect.poll(async () => {
        const records = await getRecordedMessages(page);
        const record = records.find((item) => predicate(item.message, item));
        matched = record?.message ?? null;
        return !!matched;
      }, {
        timeout: options.timeout ?? 5_000,
        message: 'waiting for a matching fake WebSocket client message',
      }).toBe(true);

      if (!matched) {
        throw new Error('Expected a matching client message, but none was captured');
      }
      return matched;
    };

    const expectSubmitAction = async (
      expected: SubmitActionExpectation,
      options: { timeout?: number } = {}
    ): Promise<ClientActionRequest> => {
      let matched: ClientActionRequest | null = null;
      await expect.poll(async () => {
        const records = await getRecordedMessages(page);
        const submitMessages = records
          .map((record) => record.message)
          .filter((message) => message.Cmd === 'SubmitAction');
        const latest = submitMessages.at(-1);
        if (!latest) return false;
        const data = latest.Data as ClientActionRequest;
        if (!submitActionMatches(data, expected)) return false;
        matched = data;
        return true;
      }, {
        timeout: options.timeout ?? 5_000,
        message: `waiting for latest SubmitAction to match ${JSON.stringify(expected)}`,
      }).toBe(true);

      if (!matched) {
        throw new Error('Expected a matching SubmitAction, but none was captured');
      }
      return matched;
    };

    const bootGame = async (scenario: ProtocolHarnessScenario) => {
      await ensureInstalled();
      await page.goto('/');
      await page.getByPlaceholder('输入你的名字').fill(scenario.myPlayerName);
      await page.getByTestId('create-room-button').click();

      await page.waitForFunction(() => (window as any).__E2E_WS__?.hasOpenSocket?.(), null, {
        timeout: 5_000,
      });

      const me = scenario.players.find((player) => player.id === scenario.myPlayerId);
      await pushServerMessage(roomEvent({
        action: 'assigned',
        room_code: scenario.roomCode,
        player_id: scenario.myPlayerId,
        player_name: scenario.myPlayerName,
        camp: me?.camp ?? '',
        char_role: me?.char_role ?? '',
        reconnect_token: 'mock-token',
        characters: scenario.characters,
      }));
      await pushServerMessage(roomEvent({
        action: 'player_list',
        room_code: scenario.roomCode,
        players: scenario.players,
        characters: scenario.characters,
      }));
      await pushServerMessage(roomEvent({
        action: 'started',
        room_code: scenario.roomCode,
        characters: scenario.characters,
      }));
      await pushServerMessage(syncStateMessage(scenario.initialState));

      await expect(page.getByTestId('game-board')).toBeVisible({ timeout: 10_000 });
    };

    await use({
      bootGame,
      pushServerMessage,
      waitForClientMessage,
      expectSubmitAction,
    });
  },
});

export { expect };
