package engine

import (
	"fmt"

	"starcup-engine/internal/model"
)

type pendingDamageAttackHitHook func(e *GameEngine, pd *model.PendingDamage, attacker *model.Player, victim *model.Player)

// applyTimingOnHitCheckPendingDamageAttackHitRules 在命中判定时处理攻击伤害命中规则。
func (e *GameEngine) applyTimingOnHitCheckPendingDamageAttackHitRules(pd *model.PendingDamage, attacker *model.Player, victim *model.Player) {
	pendingDamageBerserkerBloodRoarHook(e, pd, attacker, victim)
}

func pendingDamageBerserkerBloodRoarHook(e *GameEngine, pd *model.PendingDamage, attacker *model.Player, victim *model.Player) {
	if e == nil || pd == nil || attacker == nil || victim == nil || pd.IsCounter || !isCharacter(attacker, "berserker") || victim.Heal != 2 {
		return
	}
	pd.SetInterceptTag(model.CombatInterceptForceHit)
	pd.SetInterceptTag(model.CombatInterceptIgnoreHolyShield)
	pd.IgnoreHeal = true
	e.Log(fmt.Sprintf("%s 发动 [血腥咆哮]！目标治疗剂为2，强制命中且无视圣盾", attacker.Name))
}
