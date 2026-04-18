// gameflow: 魔枪士技能流。

package engine

import (
	"fmt"
	"starcup-engine/internal/engine/core/runtimeutil"

	"starcup-engine/internal/model"
)

// prepareMagicLancerFullnessStep 在"充盈"结算过程中推进到下一个需要选择的角色。
// 返回 true 表示所有角色都已处理完。
func (e *GameEngine) prepareMagicLancerFullnessStep(ctxData map[string]interface{}, user *model.Player) (bool, error) {
	if ctxData == nil || user == nil {
		return true, fmt.Errorf("充盈上下文无效")
	}
	orderIDs := runtimeutil.ParseStringSliceContextValue(ctxData["order_ids"])
	if len(orderIDs) == 0 {
		return true, nil
	}
	idx := runtimeutil.ToIntContextValue(ctxData["order_index"])
	if idx < 0 {
		idx = 0
	}
	for idx < len(orderIDs) {
		pid := orderIDs[idx]
		target := e.State.Players[pid]
		if target == nil {
			idx++
			continue
		}
		allowSkip := target.Camp == user.Camp
		candidates := allHandIndices(target)
		if len(candidates) == 0 {
			e.Log(fmt.Sprintf("%s 的 [充盈] 结算：%s 无手牌可弃，跳过", user.Name, target.Name))
			idx++
			continue
		}
		ctxData["order_index"] = idx
		ctxData["current_player_id"] = pid
		ctxData["allow_skip"] = allowSkip
		ctxData["candidates"] = candidates
		return false, nil
	}
	return true, nil
}
