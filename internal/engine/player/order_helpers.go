// gameflow: 逆序遍历共享 helper。

package player

import (
	"starcup-engine/internal/model"
)

// ReverseOrderOption 配置逆序遍历的过滤策略。
type ReverseOrderOption struct {
	IncludeSelf bool       // 是否包含 sourceID 自身（默认 false = 排除）
	EnemyOnly   bool       // 是否仅保留敌方阵营
	SourceCamp  model.Camp // EnemyOnly=true 时必须设置
}

// ReversePlayerIDsFromOrder 从座位顺序中按逆序返回玩家 ID。
// playerOrder 为座位顺序（[]string），sourceID 为起始玩家。
func ReversePlayerIDsFromOrder(playerOrder []string, sourceID string, opt ReverseOrderOption) []string {
	if len(playerOrder) == 0 {
		return nil
	}
	start := -1
	for i, pid := range playerOrder {
		if pid == sourceID {
			start = i
			break
		}
	}
	if start < 0 {
		return nil
	}
	n := len(playerOrder)
	stepStart := 1
	if opt.IncludeSelf {
		stepStart = 0
	}
	ids := make([]string, 0, n-1)
	for step := stepStart; step < n; step++ {
		idx := (start - step + n) % n
		ids = append(ids, playerOrder[idx])
	}
	return ids
}

// ReversePlayerIDsFromRuntime 基于 ChoiceRuntime 的便捷包装，
// 先获取 GetPlayerOrder() 再调用 ReversePlayerIDsFromOrder。
// 如果 EnemyOnly=true，会额外通过 GetPlayers() 过滤敌方。
func ReversePlayerIDsFromRuntime(rt ChoiceRuntime, sourceID string, opt ReverseOrderOption) []string {
	ids := ReversePlayerIDsFromOrder(rt.GetPlayerOrder(), sourceID, opt)
	if !opt.EnemyOnly {
		return ids
	}
	filtered := make([]string, 0, len(ids))
	for _, pid := range ids {
		if p := rt.GetPlayers()[pid]; p != nil && p.Camp != opt.SourceCamp {
			filtered = append(filtered, pid)
		}
	}
	return filtered
}

// ReversePlayersFromSlice 从 []*model.Player 切片中按逆序返回玩家对象。
// 用于不经过 ChoiceRuntime 的场景（如 GetAllPlayers()）。
func ReversePlayersFromSlice(players []*model.Player, sourceID string) []*model.Player {
	if len(players) == 0 {
		return nil
	}
	start := -1
	for i, p := range players {
		if p != nil && p.ID == sourceID {
			start = i
			break
		}
	}
	if start < 0 {
		return players
	}
	n := len(players)
	out := make([]*model.Player, 0, n-1)
	for step := 1; step < n; step++ {
		idx := (start - step + n) % n
		if players[idx] != nil {
			out = append(out, players[idx])
		}
	}
	return out
}
