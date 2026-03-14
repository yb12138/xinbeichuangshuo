package testutils

import (
	"starcup-engine/internal/model"
	"testing"
)

// TestObserver 是一个能够打印日志的观察者
type TestObserver struct {
	T *testing.T
}

// NewTestObserver 创建一个新的测试观察者
func NewTestObserver(t *testing.T) *TestObserver {
	return &TestObserver{T: t}
}

// OnGameEvent 接收引擎事件并打印详细信息
func (o *TestObserver) OnGameEvent(event model.GameEvent) {
	switch event.Type {
	case model.EventLog:
		// 引擎内部的日志（包含 [Action], [Damage], [Skill], [Debug] Phase 等）
		o.T.Logf("  📄 %s", event.Message)
	case model.EventAskInput:
		// 到了需要玩家操作的时候
		if prompt, ok := event.Data.(*model.Prompt); ok {
			o.T.Logf("  ❓ [请求输入] 玩家: %s, 内容: %s", prompt.PlayerID, prompt.Message)
			for _, opt := range prompt.Options {
				o.T.Logf("      选项 [%s]: %s", opt.ID, opt.Label)
			}
		} else {
			o.T.Logf("  ❓ [请求输入] %s", event.Message)
		}
	case model.EventError:
		o.T.Logf("  ❌ [错误] %s", event.Message)
	case model.EventGameEnd:
		o.T.Logf("  🏁 [游戏结束] %s", event.Message)
	default:
		o.T.Logf("  ℹ️ [%s] %s", event.Type, event.Message)
	}
}
