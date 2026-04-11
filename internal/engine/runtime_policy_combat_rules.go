// gameflow: 运行时策略装配：combat rules。

package engine

import (
	"fmt"

	"starcup-engine/internal/model"
)

func combatInteractionOnmyojiBindingInterruptHook(e *GameEngine, req *model.CombatRequest) bool {
	return e != nil && req != nil && e.tryStartOnmyojiBindingInterrupt(req)
}

func combatInteractionOnmyojiBindingCounterHook(e *GameEngine, req *model.CombatRequest) bool {
	return e != nil && req != nil && e.executeOnmyojiBindingCounter(req)
}

func combatInteractionOnmyojiYinYangInterruptHook(e *GameEngine, req *model.CombatRequest) bool {
	return e != nil && req != nil && e.tryStartOnmyojiYinYangInterrupt(req)
}

func combatInteractionDarkElementResponsePolicyHook(e *GameEngine, req *model.CombatRequest) bool {
	if e == nil || req == nil || req.Card == nil || req.Card.Element != model.ElementDark {
		return false
	}
	target := e.State.Players[req.TargetID]
	if target == nil || e.canUseShadowRejectResponseMagic(target) {
		return false
	}
	req.SetInterceptTag(model.CombatInterceptUnrespondable)
	return false
}

func combatDefendMagicLancerDarkBindPolicy(e *GameEngine, player *model.Player, req *model.CombatRequest) error {
	if e.isMagicLancer(player) {
		return fmt.Errorf("魔枪受[黑暗束缚]影响，不能使用法术牌防御")
	}
	return nil
}

func combatCounterShadowRejectMagicBulletPolicy(e *GameEngine, player *model.Player, req *model.CombatRequest, card model.Card) (bool, model.Card, error) {
	if !e.canUseShadowRejectResponseMagic(player) || card.Type != model.CardTypeMagic || card.Name != "魔弹" {
		return false, card, nil
	}
	e.Log(fmt.Sprintf("[Combat] %s 触发[暗影抗拒]：非自己行动阶段使用【魔弹】应战", player.Name))
	return true, card, nil
}

func combatCounterOnmyojiFactionElementPolicy(e *GameEngine, player *model.Player, req *model.CombatRequest, counterCard model.Card) (bool, bool) {
	if req.Card == nil {
		return false, false
	}
	if !e.isOnmyoji(player) || !onmyojiCanUseFactionCounter(req.Card) {
		return false, false
	}
	if counterCard.Faction == "" || counterCard.Faction != req.Card.Faction {
		return false, false
	}
	return true, true
}

func combatCounterOnmyojiFactionResolvePolicy(e *GameEngine, player *model.Player, req *model.CombatRequest, counterCard *model.Card, useFaction bool) {
	if !useFaction {
		return
	}
	e.applyOnmyojiFactionCounterBonuses(player, counterCard)
}

func magicMissileDefendMagicLancerDarkBindPolicy(e *GameEngine, player *model.Player, chain *model.MagicBulletChain) error {
	if e.isMagicLancer(player) {
		return fmt.Errorf("魔枪受[黑暗束缚]影响，不能使用法术牌防御")
	}
	return nil
}

func magicMissileCounterMagicLancerDarkBindPolicy(e *GameEngine, player *model.Player, chain *model.MagicBulletChain, card model.Card) error {
	if e.isMagicLancer(player) {
		return fmt.Errorf("魔枪受[黑暗束缚]影响，不能使用法术牌")
	}
	return nil
}
