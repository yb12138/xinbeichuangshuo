// gameflow: 指示物（Token）读写基础设施。

package player

import "starcup-engine/internal/model"

// EnsurePlayerTokensMap 确保 player.Tokens map 已初始化。
func EnsurePlayerTokensMap(p *model.Player) {
	if p != nil && p.Tokens == nil {
		p.Tokens = map[string]int{}
	}
}

// TokenValue 读取并规范化玩家 token 值：
// 小于 0 归零；cap >= 0 时按上限裁剪；并回写到玩家状态。
func TokenValue(p *model.Player, key string, cap int) int {
	if p == nil {
		return 0
	}
	EnsurePlayerTokensMap(p)
	v := p.Tokens[key]
	if v < 0 {
		v = 0
	}
	if cap >= 0 && v > cap {
		v = cap
	}
	p.Tokens[key] = v
	return v
}

// AddToken 在规范化基础上增减 token，并应用统一上限规则。
func AddToken(p *model.Player, key string, delta int, cap int) int {
	return AddTokenIgnoreCap(p, key, delta, cap, false)
}

// AddTokenIgnoreCap 允许按场景跳过上限裁剪（仅保留非负约束）。
func AddTokenIgnoreCap(p *model.Player, key string, delta int, cap int, ignoreCap bool) int {
	if p == nil {
		return 0
	}
	EnsurePlayerTokensMap(p)
	baseCap := cap
	if ignoreCap {
		baseCap = -1
	}
	v := TokenValue(p, key, baseCap) + delta
	if v < 0 {
		v = 0
	}
	if !ignoreCap && cap >= 0 && v > cap {
		v = cap
	}
	p.Tokens[key] = v
	return v
}
