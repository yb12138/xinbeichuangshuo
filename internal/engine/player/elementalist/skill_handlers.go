// gameflow: 元素师角色技能处理器。

package elementalist

import (
	"fmt"
	"strings"

	"starcup-engine/internal/model"
)

// --- Element Token Helpers ---

const elementCap = 3

func elementValue(p *model.Player) int {
	if p == nil || p.Tokens == nil {
		return 0
	}
	return p.Tokens["element"]
}

func addElementToken(p *model.Player, delta int) int {
	if p == nil {
		return 0
	}
	if p.Tokens == nil {
		p.Tokens = map[string]int{}
	}
	v := p.Tokens["element"] + delta
	if v < 0 {
		v = 0
	}
	if v > elementCap {
		v = elementCap
	}
	p.Tokens["element"] = v
	return v
}

func matchingElementCardIndices(p *model.Player, element model.Element) []int {
	if p == nil {
		return nil
	}
	out := make([]int, 0, len(p.Hand))
	for i, card := range p.Hand {
		if card.Element == element {
			out = append(out, i)
		}
	}
	return out
}

// --- Handlers ---

// ElementalistAbsorbHandler 元素吸收：受到法术伤害时获得1元素。
type ElementalistAbsorbHandler struct{}

func (h *ElementalistAbsorbHandler) CanUse(ctx *model.Context) bool {
	if ctx.Timing != model.TimingOnDamageTaken || ctx.EventCtx == nil {
		return false
	}
	if !ctx.Flags["IsMagicDamage"] {
		return false
	}
	noAbsorb := false
	if damageType, ok := ctx.Selections["damage_type"].(model.DamageType); ok {
		noAbsorb = strings.Contains(strings.ToLower(string(damageType)), "no_absorb")
	}
	if noAbsorb {
		return false
	}
	if ctx.EventCtx.SourceID != ctx.User.ID {
		return false
	}
	if ctx.EventCtx.Card != nil && ctx.EventCtx.Card.Name == "元素点燃" {
		return false
	}
	return elementValue(ctx.User) < elementCap
}

func (h *ElementalistAbsorbHandler) Execute(ctx *model.Context) error {
	v := addElementToken(ctx.User, 1)
	ctx.Game.Log(fmt.Sprintf("%s 的 [元素吸收] 触发，元素=%d", ctx.User.Name, v))
	return nil
}

// ElementalistIgniteHandler 元素点燃：消耗3元素，造成2点法术伤害并获得额外法术行动。
type ElementalistIgniteHandler struct{}

func (h *ElementalistIgniteHandler) CanUse(ctx *model.Context) bool {
	return elementValue(ctx.User) >= 3
}

func (h *ElementalistIgniteHandler) Execute(ctx *model.Context) error {
	if ctx.Target == nil {
		return fmt.Errorf("元素点燃需要目标")
	}
	if elementValue(ctx.User) < 3 {
		return fmt.Errorf("元素不足，至少需要3点元素")
	}
	addElementToken(ctx.User, -3)
	ctx.Game.InflictDamage(ctx.User.ID, ctx.Target.ID, 2, "magic_no_absorb")
	model.AppendMagicAction(ctx.User, "元素点燃")
	ctx.Game.Log(fmt.Sprintf("%s 发动 [元素点燃]，对 %s 造成2点法术伤害并获得额外法术行动", ctx.User.Name, ctx.Target.Name))
	return nil
}

// ElementalistThunderStrikeHandler 雷击。
type ElementalistThunderStrikeHandler struct{}

func (h *ElementalistThunderStrikeHandler) CanUse(ctx *model.Context) bool { return true }

func (h *ElementalistThunderStrikeHandler) Execute(ctx *model.Context) error {
	if ctx.Target == nil {
		return fmt.Errorf("雷击需要目标")
	}
	if !ctx.User.HasElement(model.ElementThunder) {
		ctx.Game.InflictDamage(ctx.User.ID, ctx.Target.ID, 1, model.MagicAttack)
		ctx.Game.ModifyGem(string(ctx.User.Camp), 1)
		ctx.Game.Log(fmt.Sprintf("%s 发动 [雷击]，造成1点法术伤害并为阵营+1宝石", ctx.User.Name))
		return nil
	}
	matching := matchingElementCardIndices(ctx.User, model.ElementThunder)
	if len(matching) == 0 {
		ctx.Game.InflictDamage(ctx.User.ID, ctx.Target.ID, 1, model.MagicAttack)
		ctx.Game.ModifyGem(string(ctx.User.Camp), 1)
		ctx.Game.Log(fmt.Sprintf("%s 发动 [雷击]，造成1点法术伤害并为阵营+1宝石", ctx.User.Name))
		return nil
	}
	data := map[string]interface{}{
		"choice_type":        "elementalist_bonus_card",
		"user_id":            ctx.User.ID,
		"damage_target_id":   ctx.Target.ID,
		"base_damage":        1,
		"bonus_element":      string(model.ElementThunder),
		"matching_indices":   matching,
		"camp_gem_bonus":     1,
		"grant_attack":       false,
		"grant_magic":        false,
		"skill_display_name": "雷击",
	}
	ctx.Game.PushInterrupt(&model.Interrupt{Type: model.InterruptChoice, PlayerID: ctx.User.ID, Context: data})
	return nil
}

// ElementalistFreezeHandler 冰冻。
type ElementalistFreezeHandler struct{}

func (h *ElementalistFreezeHandler) CanUse(ctx *model.Context) bool { return true }

func (h *ElementalistFreezeHandler) Execute(ctx *model.Context) error {
	if len(ctx.Targets) == 0 && ctx.Target == nil {
		return fmt.Errorf("冰冻需要至少1个目标")
	}
	var dmgTarget *model.Player
	var healTarget *model.Player
	if len(ctx.Targets) >= 1 {
		dmgTarget = ctx.Targets[0]
	}
	if len(ctx.Targets) >= 2 {
		healTarget = ctx.Targets[1]
	}
	if dmgTarget == nil {
		dmgTarget = ctx.Target
	}
	if healTarget == nil {
		healTarget = ctx.User
	}
	if !ctx.User.HasElement(model.ElementWater) {
		ctx.Game.InflictDamage(ctx.User.ID, dmgTarget.ID, 1, model.MagicAttack)
		ctx.Game.Heal(healTarget.ID, 1)
		ctx.Game.Log(fmt.Sprintf("%s 发动 [冰冻]，对 %s 造成1点法术伤害并治疗 %s 1点", ctx.User.Name, dmgTarget.Name, healTarget.Name))
		return nil
	}
	matching := matchingElementCardIndices(ctx.User, model.ElementWater)
	if len(matching) == 0 {
		ctx.Game.InflictDamage(ctx.User.ID, dmgTarget.ID, 1, model.MagicAttack)
		ctx.Game.Heal(healTarget.ID, 1)
		ctx.Game.Log(fmt.Sprintf("%s 发动 [冰冻]，对 %s 造成1点法术伤害并治疗 %s 1点", ctx.User.Name, dmgTarget.Name, healTarget.Name))
		return nil
	}
	data := map[string]interface{}{
		"choice_type":        "elementalist_bonus_card",
		"user_id":            ctx.User.ID,
		"damage_target_id":   dmgTarget.ID,
		"heal_target_id":     healTarget.ID,
		"base_damage":        1,
		"bonus_element":      string(model.ElementWater),
		"matching_indices":   matching,
		"camp_gem_bonus":     0,
		"grant_attack":       false,
		"grant_magic":        false,
		"skill_display_name": "冰冻",
	}
	ctx.Game.PushInterrupt(&model.Interrupt{Type: model.InterruptChoice, PlayerID: ctx.User.ID, Context: data})
	return nil
}

// ElementalistWindBladeHandler 风刃。
type ElementalistWindBladeHandler struct{}

func (h *ElementalistWindBladeHandler) CanUse(ctx *model.Context) bool { return true }

func (h *ElementalistWindBladeHandler) Execute(ctx *model.Context) error {
	if ctx.Target == nil {
		return fmt.Errorf("风刃需要目标")
	}
	if !ctx.User.HasElement(model.ElementWind) {
		ctx.Game.InflictDamage(ctx.User.ID, ctx.Target.ID, 1, model.MagicAttack)
		model.AppendAttackAction(ctx.User, "风刃")
		ctx.Game.Log(fmt.Sprintf("%s 发动 [风刃]，造成1点法术伤害并获得额外攻击行动", ctx.User.Name))
		return nil
	}
	matching := matchingElementCardIndices(ctx.User, model.ElementWind)
	if len(matching) == 0 {
		ctx.Game.InflictDamage(ctx.User.ID, ctx.Target.ID, 1, model.MagicAttack)
		model.AppendAttackAction(ctx.User, "风刃")
		ctx.Game.Log(fmt.Sprintf("%s 发动 [风刃]，造成1点法术伤害并获得额外攻击行动", ctx.User.Name))
		return nil
	}
	data := map[string]interface{}{
		"choice_type":        "elementalist_bonus_card",
		"user_id":            ctx.User.ID,
		"damage_target_id":   ctx.Target.ID,
		"base_damage":        1,
		"bonus_element":      string(model.ElementWind),
		"matching_indices":   matching,
		"camp_gem_bonus":     0,
		"grant_attack":       true,
		"grant_magic":        false,
		"skill_display_name": "风刃",
	}
	ctx.Game.PushInterrupt(&model.Interrupt{Type: model.InterruptChoice, PlayerID: ctx.User.ID, Context: data})
	return nil
}

// ElementalistMeteorHandler 陨石。
type ElementalistMeteorHandler struct{}

func (h *ElementalistMeteorHandler) CanUse(ctx *model.Context) bool { return true }

func (h *ElementalistMeteorHandler) Execute(ctx *model.Context) error {
	if ctx.Target == nil {
		return fmt.Errorf("陨石需要目标")
	}
	if !ctx.User.HasElement(model.ElementEarth) {
		ctx.Game.InflictDamage(ctx.User.ID, ctx.Target.ID, 1, model.MagicAttack)
		model.AppendMagicAction(ctx.User, "陨石")
		ctx.Game.Log(fmt.Sprintf("%s 发动 [陨石]，造成1点法术伤害并获得额外法术行动", ctx.User.Name))
		return nil
	}
	matching := matchingElementCardIndices(ctx.User, model.ElementEarth)
	if len(matching) == 0 {
		ctx.Game.InflictDamage(ctx.User.ID, ctx.Target.ID, 1, model.MagicAttack)
		model.AppendMagicAction(ctx.User, "陨石")
		ctx.Game.Log(fmt.Sprintf("%s 发动 [陨石]，造成1点法术伤害并获得额外法术行动", ctx.User.Name))
		return nil
	}
	data := map[string]interface{}{
		"choice_type":        "elementalist_bonus_card",
		"user_id":            ctx.User.ID,
		"damage_target_id":   ctx.Target.ID,
		"base_damage":        1,
		"bonus_element":      string(model.ElementEarth),
		"matching_indices":   matching,
		"camp_gem_bonus":     0,
		"grant_attack":       false,
		"grant_magic":        true,
		"skill_display_name": "陨石",
	}
	ctx.Game.PushInterrupt(&model.Interrupt{Type: model.InterruptChoice, PlayerID: ctx.User.ID, Context: data})
	return nil
}

// ElementalistFireballHandler 火球。
type ElementalistFireballHandler struct{}

func (h *ElementalistFireballHandler) CanUse(ctx *model.Context) bool { return true }

func (h *ElementalistFireballHandler) Execute(ctx *model.Context) error {
	if ctx.Target == nil {
		return fmt.Errorf("火球需要目标")
	}
	if !ctx.User.HasElement(model.ElementFire) {
		ctx.Game.InflictDamage(ctx.User.ID, ctx.Target.ID, 2, model.MagicAttack)
		ctx.Game.Log(fmt.Sprintf("%s 发动 [火球]，造成2点法术伤害", ctx.User.Name))
		return nil
	}
	matching := matchingElementCardIndices(ctx.User, model.ElementFire)
	if len(matching) == 0 {
		ctx.Game.InflictDamage(ctx.User.ID, ctx.Target.ID, 2, model.MagicAttack)
		ctx.Game.Log(fmt.Sprintf("%s 发动 [火球]，造成2点法术伤害", ctx.User.Name))
		return nil
	}
	data := map[string]interface{}{
		"choice_type":        "elementalist_bonus_card",
		"user_id":            ctx.User.ID,
		"damage_target_id":   ctx.Target.ID,
		"base_damage":        2,
		"bonus_element":      string(model.ElementFire),
		"matching_indices":   matching,
		"camp_gem_bonus":     0,
		"grant_attack":       false,
		"grant_magic":        false,
		"skill_display_name": "火球",
	}
	ctx.Game.PushInterrupt(&model.Interrupt{Type: model.InterruptChoice, PlayerID: ctx.User.ID, Context: data})
	return nil
}

// ElementalistMoonlightHandler 月光。
type ElementalistMoonlightHandler struct{}

func (h *ElementalistMoonlightHandler) CanUse(ctx *model.Context) bool {
	return ctx.User.Gem > 0
}

func (h *ElementalistMoonlightHandler) Execute(ctx *model.Context) error {
	if ctx.Target == nil {
		return fmt.Errorf("月光需要目标")
	}
	x := ctx.User.Gem + ctx.User.Crystal
	dmg := x + 1
	ctx.Game.InflictDamage(ctx.User.ID, ctx.Target.ID, dmg, model.MagicAttack)
	ctx.Game.Log(fmt.Sprintf("%s 发动 [月光]，造成%d点法术伤害", ctx.User.Name, dmg))
	return nil
}
