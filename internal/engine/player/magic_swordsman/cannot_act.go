package magic_swordsman

import (
	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// CannotActChecker 魔剑士无法行动判断。
// 暗影抗拒：手牌全是法术牌时，即使有法术牌也不能使用，可以宣告无法行动。
func CannotActCheckerFn(p *model.Player) (bool, string) {
	// 不在暗影形态时，不拦截，走默认判断
	if !InShadowForm(p) {
		return false, ""
	}

	// 无手牌时走默认判断
	if len(p.Hand) == 0 {
		return false, ""
	}

	// 检查手牌是否全是法术牌
	allMagic := true
	for _, c := range p.Hand {
		if c.Type != model.CardTypeMagic {
			allMagic = false
			break
		}
	}

	if allMagic {
		return true, "暗影抗拒：手牌全是法术牌，无法使用"
	}

	// 有攻击牌时走默认判断
	return false, ""
}

func init() {
	// 确保编译期检查接口
	_ = player.CannotActChecker(CannotActCheckerFn)
}
