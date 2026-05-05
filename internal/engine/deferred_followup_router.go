// gameflow: DeferredFollowups 延迟后续任务执行顺序。

package engine

import (
	"fmt"

	"starcup-engine/internal/model"
)

type deferredFollowupResolver func(*GameEngine, model.DeferredFollowup) error

type deferredFollowupHandler struct {
	label   string
	resolve deferredFollowupResolver
}

var deferredFollowupHandlers = buildDeferredFollowupHandlers()

func buildDeferredFollowupHandlers() map[string]deferredFollowupHandler {
	handlers := map[string]deferredFollowupHandler{}
	registerDeferredFollowupHandlers(handlers, buildPostActionEndDeferredFollowupHandlers())
	registerDeferredFollowupHandlers(handlers, buildSkillEffectResumeFollowupHandler())
	// mountPlayerDeferredFollowupSpecs(handlers) — 已删除，角色 FollowupSpecs 改为 FlowContinuationHandlers
	return handlers
}

func registerDeferredFollowupHandlers(dst map[string]deferredFollowupHandler, src map[string]deferredFollowupHandler) {
	for followupType, handler := range src {
		if followupType == "" || handler.resolve == nil {
			continue
		}
		dst[followupType] = handler
	}
}

func (e *GameEngine) enqueueDeferredFollowup(f model.DeferredFollowup) {
	e.State.DeferredFollowups = append(e.State.DeferredFollowups, f)
}

func (e *GameEngine) processDeferredFollowups() bool {
	if len(e.State.DeferredFollowups) == 0 {
		return false
	}
	f := e.State.DeferredFollowups[0]
	e.State.DeferredFollowups = e.State.DeferredFollowups[1:]
	if handled, label, err := e.resolveDeferredFollowup(f); handled {
		if err != nil {
			e.Log(fmt.Sprintf("[%s] 延迟后续结算失败: %v", label, err))
		}
		return true
	}
	e.Log(fmt.Sprintf("[Warn] 未知的延迟后续类型: %s", f.Type))
	return true
}

func (e *GameEngine) resolveDeferredFollowup(f model.DeferredFollowup) (bool, string, error) {
	handler, ok := deferredFollowupHandlers[f.Type]
	if !ok || handler.resolve == nil {
		return false, "", nil
	}
	return true, handler.label, handler.resolve(e, f)
}
