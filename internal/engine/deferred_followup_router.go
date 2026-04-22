// gameflow: DeferredFollowups 延迟后续任务执行顺序。

package engine

import (
	"fmt"

	engineplayer "starcup-engine/internal/engine/player"
	bloodpriestesspkg "starcup-engine/internal/engine/player/blood_priestess"
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
	registerDeferredFollowupHandlers(handlers, buildBloodPriestessDeferredFollowupHandlers())
	registerDeferredFollowupHandlers(handlers, buildAssassinDeferredFollowupHandlers())
	mountPlayerDeferredFollowupSpecs(handlers)
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

// ---- 血祭司延迟后续处理 ----

func buildBloodPriestessDeferredFollowupHandlers() map[string]deferredFollowupHandler {
	return map[string]deferredFollowupHandler{
		"blood_priestess_shared_life_place": {
			label:   "BloodPriestess",
			resolve: (*GameEngine).resolveBloodPriestessSharedLifePlaceFollowup,
		},
		"blood_priestess_blood_sorrow_apply": {
			label:   "BloodPriestess",
			resolve: (*GameEngine).resolveBloodPriestessBloodSorrowFollowup,
		},
		"blood_priestess_curse_discard": {
			label:   "BloodPriestess",
			resolve: (*GameEngine).resolveBloodPriestessCurseDiscardFollowup,
		},
	}
}

func (e *GameEngine) resolveBloodPriestessSharedLifePlaceFollowup(f model.DeferredFollowup) error {
	user := e.State.Players[f.UserID]
	if user == nil {
		return fmt.Errorf("执行者不存在: %s", f.UserID)
	}
	if !engineplayer.IsCharacter(user, "blood_priestess") {
		return fmt.Errorf("仅血之巫女可执行同生共死后续")
	}
	if len(f.TargetIDs) != 1 {
		return fmt.Errorf("同生共死后续目标数量错误: %d", len(f.TargetIDs))
	}
	target := e.State.Players[f.TargetIDs[0]]
	if target == nil {
		return fmt.Errorf("同生共死目标不存在: %s", f.TargetIDs[0])
	}

	var card model.Card
	if f.Data != nil {
		if v, ok := f.Data["card"]; ok {
			switch c := v.(type) {
			case model.Card:
				card = c
			case *model.Card:
				if c != nil {
					card = *c
				}
			}
		}
	}
	if card.ID == "" || card.Name == "" {
		return fmt.Errorf("同生共死后续缺少原始专属卡")
	}

	rt := newRoleChoiceRuntime(e)
	if err := bloodpriestesspkg.PlaceSharedLife(rt, user, target, card); err != nil {
		user.RestoreExclusiveCard(card)
		return err
	}
	e.Log(fmt.Sprintf("%s 的 [同生共死] 生效：放置于 %s 面前", user.Name, target.Name))

	e.checkHandLimit(user, nil)
	if target.ID != user.ID {
		e.checkHandLimit(target, nil)
	}
	return nil
}

func (e *GameEngine) resolveBloodPriestessBloodSorrowFollowup(f model.DeferredFollowup) error {
	user := e.State.Players[f.UserID]
	if user == nil {
		return fmt.Errorf("执行者不存在: %s", f.UserID)
	}
	mode := ""
	if f.Data != nil {
		if raw, ok := f.Data["mode"].(string); ok {
			mode = raw
		}
	}
	if mode == "" {
		return fmt.Errorf("血之哀伤后续模式缺失")
	}
	checked := map[string]bool{}
	checkCap := func(player *model.Player) {
		if player == nil || checked[player.ID] {
			return
		}
		checked[player.ID] = true
		e.checkHandLimit(player, nil)
	}

	rt := newRoleChoiceRuntime(e)
	switch mode {
	case "remove":
		if !bloodpriestesspkg.RemoveSharedLife(rt, user, true) {
			return fmt.Errorf("当前没有可移除的同生共死")
		}
		e.Log(fmt.Sprintf("%s 的 [血之哀伤] 后续结算：移除【同生共死】", user.Name))
		checkCap(user)
		return nil
	case "transfer":
		targetID := ""
		if f.Data != nil {
			if raw, ok := f.Data["target_id"].(string); ok {
				targetID = raw
			}
		}
		target := e.State.Players[targetID]
		if target == nil {
			return fmt.Errorf("转移目标不存在: %s", targetID)
		}
		holder, card, ok := bloodpriestesspkg.DetachSharedLife(rt, user)
		if !ok {
			return fmt.Errorf("当前没有可转移的同生共死")
		}
		if err := e.attachExclusiveEffectCard(user, target, model.EffectBloodSharedLife, card); err != nil {
			if holder != nil {
				_ = e.attachExclusiveEffectCard(user, holder, model.EffectBloodSharedLife, card)
			}
			return err
		}
		e.Log(fmt.Sprintf("%s 的 [血之哀伤] 后续结算：将【同生共死】转移至 %s", user.Name, target.Name))
		checkCap(user)
		checkCap(holder)
		checkCap(target)
		return nil
	default:
		return fmt.Errorf("未知的血之哀伤后续模式: %s", mode)
	}
}

func (e *GameEngine) resolveBloodPriestessCurseDiscardFollowup(f model.DeferredFollowup) error {
	user := e.State.Players[f.UserID]
	if user == nil {
		return fmt.Errorf("执行者不存在: %s", f.UserID)
	}
	discardNeed := 3
	if len(user.Hand) < discardNeed {
		discardNeed = len(user.Hand)
	}
	if discardNeed <= 0 {
		e.Log(fmt.Sprintf("%s 的 [血之诅咒] 后续：无需弃牌", user.Name))
		return nil
	}
	e.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: user.ID,
		Context: map[string]interface{}{
			"choice_type":   "bp_curse_discard",
			"user_id":       user.ID,
			"discard_count": discardNeed,
		},
	})
	e.Log(fmt.Sprintf("%s 的 [血之诅咒] 后续：伤害结算完成，请弃置%d张牌", user.Name, discardNeed))
	return nil
}
