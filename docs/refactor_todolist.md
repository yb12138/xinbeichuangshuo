# 重构执行清单

## 目标

依据 `docs/` 目录中的架构、协议与数据模型文档，对当前项目进行一次性协议切换下的前后端架构重构。

本次执行采用以下原则：

- 外部 WebSocket 协议一次性切换到 `Cmd/Data` 信封
- 游戏态下行统一收敛到 `SyncState / RequireAction / NotifyTimeline`
- `NotifyEvent` 仅作为迁移期兼容通道保留，后续逐步下线
- 房间管理命令改为 `RoomAction / RoomEvent`
- 先拆协议与路由边界，再逐步下沉到 store / session / timeline 子系统

## TodoList

### Phase 0: 基线确认

- [x] 读取并梳理 `docs/` 全量文档
- [x] 对照当前 `Go + Vue` 实现整理架构差距
- [x] 确认采用“一次性切换新协议”路线

### Phase 1: 协议骨架切换

- [x] 新增重构执行清单文档
- [x] 后端新增统一 `Cmd/Data` WS 信封
- [x] 后端新增 `SyncState / RequireAction / NotifyTimeline / NotifyEvent / RoomEvent / RoomAction / SubmitAction` DTO
- [x] 后端将 `Room.HandleMessage` 改为按 `Cmd` 路由
- [x] 后端将 gameplay 下行从旧 `event/event_type` 改为新命令
- [x] 后端引入时间线批次生成器，给现有 legacy 事件补 `NotifyTimeline`
- [x] 前端新增 `network/protocol.ts`
- [x] 前端新增 `network/messageRouter.ts`
- [x] 前端 `useWebSocket` 切到新信封与新命令
- [x] 前端将旧 `PlayerAction` 发送逻辑适配为新 `SubmitAction` 请求

### Phase 2: 后端职责拆分

- [x] 从 `internal/server/room.go` 抽离协议 DTO
- [x] 从 `internal/server/room.go` 抽离状态快照组装器
- [x] 从 `internal/server/room.go` 抽离时间线发送适配器
- [x] 从 `internal/server/room.go` 抽离房间消息分发器
- [x] 为房间层保留最小职责：会话管理、串行入口、广播协调

### Phase 3: 前端状态拆分

- [x] 新建 `session.store.ts`
- [x] 新建 `snapshot.store.ts`
- [x] 新建 `interrupt.store.ts`
- [x] 新建 `timeline.store.ts`
- [x] 新建 `ui.store.ts`
- [x] 新建交互派生层 `useBattleInteractionState`
- [x] 将现有 `gameStore.ts` 中的状态逐步迁移
- [x] 保持现有页面组件可运行

### Phase 4: 页面与交互迁移

- [x] 页面层统一经由 `messageRouter`
- [x] 提交流程统一经由 `useSubmitAction`
- [x] `PromptDialog` 改为读取 `interrupt.store`
- [x] `ActionTimeline` 改为读取 `timeline.store`
- [x] `GameBoard / ActionPanel / RoomLobby` 只依赖 store 和 composable
- [x] `GameBoard` 顶部 HUD / 房间座次 / 终局浮层优先读取 `session/snapshot/interrupt/ui`

### Phase 5: 后端架构继续下沉

- [x] 引入 session actor inbox，统一收口客户端动作/定时器/内部推进
- [x] 引入 snapshot assembler
- [x] 引入 interrupt adapter
- [x] 引入 timeline adapter
- [x] 从 `room.go` 抽离 game event dispatcher
- [x] 将 `room.go` 收敛为房间主循环与最小协调职责
- [x] 继续收缩 `gameStore.ts` 兼容门面

### Phase 6: 测试与收尾

- [x] 更新房间/重连相关测试以适配新协议
- [x] 增加后端协议适配测试
- [x] 增加前端 store 分层回归测试
- [x] 增加前端 messageRouter 契约测试
- [x] 增加前端 action adapter 测试
- [x] 跑通关键流程回归：开始游戏、攻击、响应、伤害、重连、托管
- [x] 将 gameplay 下行默认收敛到 `SyncState / RequireAction / NotifyTimeline`
- [x] 清理迁移期兼容字段，评估下线 `NotifyEvent`（运行时已移除，仅保留测试中的 unknown fallback 断言）
