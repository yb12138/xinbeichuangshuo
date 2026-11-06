package beast_samurai

import (
	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

const ZanshinCap = 4

func Zanshin(p *model.Player) int {
	return player.TokenValue(p, "bs_zanshin", ZanshinCap)
}

func AddZanshin(p *model.Player, delta int) int {
	return player.AddToken(p, "bs_zanshin", delta, ZanshinCap)
}

const BeastSoulCap = 2

// BeastSoul 读取兽魂数。兽魂可因"无视上限"的效果溢出到上限以上，
// 读取时不截断也不回写，避免溢出部分在第一次读取时就被吞掉。
func BeastSoul(p *model.Player) int {
	if p == nil {
		return 0
	}
	player.EnsurePlayerTokensMap(p)
	v := p.Tokens["bs_beast_soul"]
	if v < 0 {
		return 0
	}
	return v
}

// AddBeastSoul 增减兽魂。ignoreCap=false 时增加不会超过上限，
// 但已经溢出的部分保留（只是不再继续增加）；减少总是从实际值上扣。
func AddBeastSoul(p *model.Player, delta int, ignoreCap bool) int {
	if p == nil {
		return 0
	}
	current := BeastSoul(p)
	v := current + delta
	if v < 0 {
		v = 0
	}
	if !ignoreCap && delta > 0 && v > BeastSoulCap {
		v = BeastSoulCap
		if current > v {
			v = current
		}
	}
	p.Tokens["bs_beast_soul"] = v
	return v
}

// ClearAttackTokens zeros the reversal-pending state marker.
func ClearAttackTokens(p *model.Player) {
	player.EnsurePlayerTokensMap(p)
	if p.TurnState.SkillFlowState == nil {
		p.TurnState.SkillFlowState = map[string]int{}
	}
	p.TurnState.SkillFlowState["bs_reversal_pending_x"] = 0
	p.TurnState.SkillFlowState[reversalArmedKey] = 0
}

// reversalArmedKey 标记本次攻击已在攻击前发动逆反居合斩，命中时改为弃牌。
const reversalArmedKey = "bs_reversal_armed"
