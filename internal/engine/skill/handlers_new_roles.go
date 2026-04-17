// gameflow: 新角色批次集中 handler（与 data 配置 ID 对应）。

package skills

import (
	"fmt"
	"sort"
	"starcup-engine/internal/model"
	"strings"
)

func getToken(p *model.Player, key string) int {
	if p == nil {
		return 0
	}
	if p.Tokens == nil {
		p.Tokens = map[string]int{}
	}
	return p.Tokens[key]
}

func setToken(p *model.Player, key string, v int) {
	if p == nil {
		return
	}
	if p.Tokens == nil {
		p.Tokens = map[string]int{}
	}
	p.Tokens[key] = v
}

func getSkillFlow(p *model.Player, key string) int {
	if p == nil || p.TurnState.SkillFlowState == nil {
		return 0
	}
	return p.TurnState.SkillFlowState[key]
}

func setSkillFlow(p *model.Player, key string, v int) {
	if p == nil {
		return
	}
	if p.TurnState.SkillFlowState == nil {
		p.TurnState.SkillFlowState = make(map[string]int)
	}
	p.TurnState.SkillFlowState[key] = v
}

func addToken(p *model.Player, key string, delta int, minV int, maxV int) int {
	cur := getToken(p, key)
	cur += delta
	if cur < minV {
		cur = minV
	}
	if maxV >= minV && cur > maxV {
		cur = maxV
	}
	setToken(p, key, cur)
	return cur
}

func hasForm(p *model.Player, form string) bool {
	return p != nil && p.Form == form
}

func enterForm(p *model.Player, form string) {
	if p == nil {
		return
	}
	p.Orientation = model.OrientationTapped
	p.Form = form
}

func leaveForm(p *model.Player, form string) {
	if p == nil {
		return
	}
	if form != "" && p.Form != form {
		return
	}
	p.Orientation = model.OrientationNormal
	p.Form = ""
}

func playerEnergyCap(p *model.Player) int {
	if p == nil {
		return 3
	}
	cap := 3
	if p.Character != nil && p.Character.ID == "sage" {
		cap++
	}
	return cap
}

func addAttackAction(p *model.Player, source string) {
	model.AppendAttackAction(p, source)
}

func addMagicAction(p *model.Player, source string) {
	model.AppendMagicAction(p, source)
}

func hasElementCard(p *model.Player, element model.Element) bool {
	for _, c := range p.Hand {
		if c.Element == element {
			return true
		}
	}
	return false
}

// --- 9. 女武神 ---

type ValkyrieDivinePursuitHandler struct{ BaseHandler }

func (h *ValkyrieDivinePursuitHandler) CanUse(ctx *model.Context) bool {
	if ctx.Timing != model.TimingOnActionEnd || ctx.EventCtx == nil {
		return false
	}
	if ctx.EventCtx.ActionType != model.ActionAttack && ctx.EventCtx.ActionType != model.ActionMagic {
		return false
	}
	// 攻击行动仅指主动攻击；应战攻击结束不触发神圣追击。
	if ctx.EventCtx.ActionType == model.ActionAttack &&
		ctx.EventCtx.AttackInfo != nil &&
		ctx.EventCtx.AttackInfo.CounterInitiator != "" {
		return false
	}
	return ctx.User.Heal > 0
}

func (h *ValkyrieDivinePursuitHandler) Execute(ctx *model.Context) error {
	ctx.User.Heal--
	addAttackAction(ctx.User, "神圣追击")
	ctx.Game.Log(fmt.Sprintf("%s 发动 [神圣追击]，移除1点治疗并获得额外攻击行动", ctx.User.Name))
	return nil
}

type ValkyrieOrderSealHandler struct{ BaseHandler }

func (h *ValkyrieOrderSealHandler) Execute(ctx *model.Context) error {
	ctx.Game.DrawCards(ctx.User.ID, 2)
	ctx.Game.Heal(ctx.User.ID, 1)
	ctx.User.Crystal++
	ctx.Game.Log(fmt.Sprintf("%s 发动 [秩序之印]，摸2并获得1治疗+1蓝水晶", ctx.User.Name))
	return nil
}

type ValkyriePeaceWalkerHandler struct{ BaseHandler }

func (h *ValkyriePeaceWalkerHandler) CanUse(ctx *model.Context) bool {
	if ctx.EventCtx != nil && ctx.EventCtx.AttackInfo != nil && ctx.EventCtx.AttackInfo.CounterInitiator != "" {
		return false
	}
	return hasForm(ctx.User, model.FormValkyrieHeroic)
}

func (h *ValkyriePeaceWalkerHandler) Execute(ctx *model.Context) error {
	if !hasForm(ctx.User, model.FormValkyrieHeroic) {
		return nil
	}
	leaveForm(ctx.User, model.FormValkyrieHeroic)
	ctx.Game.Log(fmt.Sprintf("%s 的 [和平行者] 触发，脱离英灵形态", ctx.User.Name))
	return nil
}

type ValkyrieMilitaryGloryHandler struct{ BaseHandler }

func (h *ValkyrieMilitaryGloryHandler) CanUse(ctx *model.Context) bool {
	return ctx != nil && ctx.Timing == model.TimingOnTurnStart && hasForm(ctx.User, model.FormValkyrieHeroic)
}

func (h *ValkyrieMilitaryGloryHandler) Execute(ctx *model.Context) error {
	camp := string(ctx.User.Camp)
	energy := ctx.Game.GetCampCrystals(camp) + ctx.Game.GetCampGems(camp)
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type": "valkyrie_military_glory_mode",
			"user_id":     ctx.User.ID,
			"camp":        camp,
			"max_x":       minInt(2, energy),
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 的 [军威神光] 触发，等待选择效果", ctx.User.Name))
	return nil
}

type ValkyrieHeroicSummonHandler struct{ BaseHandler }

func (h *ValkyrieHeroicSummonHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.EventCtx == nil || ctx.EventCtx.AttackInfo == nil {
		return false
	}
	if ctx.Timing != model.TimingOnHitCheck {
		return false
	}
	info := ctx.EventCtx.AttackInfo
	if info.ActionType != string(model.ActionAttack) || !info.IsHit {
		return false
	}
	return canPayCrystalLike(ctx, 1)
}

func (h *ValkyrieHeroicSummonHandler) Execute(ctx *model.Context) error {
	if !canPayCrystalLike(ctx, 1) {
		return nil
	}
	if !spendCrystalLike(ctx, 1) {
		return fmt.Errorf("英灵召唤发动失败：水晶不足（红宝石可替代）")
	}
	if ctx.EventCtx != nil && ctx.EventCtx.DamageVal != nil {
		*ctx.EventCtx.DamageVal += 1
	}
	hasMagic := false
	magicIndices := make([]int, 0)
	for i, c := range ctx.User.Hand {
		if c.Type == model.CardTypeMagic {
			hasMagic = true
			magicIndices = append(magicIndices, i)
		}
	}
	if hasMagic {
		ctx.Game.PushInterrupt(&model.Interrupt{
			Type:     model.InterruptChoice,
			PlayerID: ctx.User.ID,
			Context: map[string]interface{}{
				"choice_type":   "valkyrie_heroic_discard_card",
				"user_id":       ctx.User.ID,
				"user_ctx":      ctx,
				"magic_indices": magicIndices,
			},
		})
	}
	// 仅在自己的行动回合内，英灵召唤才会令女武神进入英灵形态；
	// 应战命中依然可以发动该技能，但不会入形态。
	if ctx.User.IsActive {
		enterForm(ctx.User, model.FormValkyrieHeroic)
		ctx.Game.Log(fmt.Sprintf("%s 发动 [英灵召唤]，伤害+1并进入英灵形态", ctx.User.Name))
		return nil
	}
	ctx.Game.Log(fmt.Sprintf("%s 发动 [英灵召唤]，伤害+1", ctx.User.Name))
	return nil
}

// --- 10. 元素师 ---

type ElementalistAbsorbHandler struct{ BaseHandler }

func (h *ElementalistAbsorbHandler) CanUse(ctx *model.Context) bool {
	if ctx.Timing != model.TimingOnDamageTaken || ctx.EventCtx == nil {
		return false
	}
	if !ctx.Flags["IsMagicDamage"] {
		return false
	}
	// 检查伤害类型是否包含 "no_absorb"（元素吸收专属判断逻辑）
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
	return getToken(ctx.User, "element") < 3
}

func (h *ElementalistAbsorbHandler) Execute(ctx *model.Context) error {
	v := addToken(ctx.User, "element", 1, 0, 3)
	ctx.Game.Log(fmt.Sprintf("%s 的 [元素吸收] 触发，元素=%d", ctx.User.Name, v))
	return nil
}

type ElementalistIgniteHandler struct{ BaseHandler }

func (h *ElementalistIgniteHandler) CanUse(ctx *model.Context) bool {
	return getToken(ctx.User, "element") >= 3
}

func (h *ElementalistIgniteHandler) Execute(ctx *model.Context) error {
	if ctx.Target == nil {
		return fmt.Errorf("元素点燃需要目标")
	}
	if getToken(ctx.User, "element") < 3 {
		return fmt.Errorf("元素不足，至少需要3点元素")
	}
	addToken(ctx.User, "element", -3, 0, 3)
	ctx.Game.InflictDamage(ctx.User.ID, ctx.Target.ID, 2, "magic_no_absorb")
	addMagicAction(ctx.User, "元素点燃")
	ctx.Game.Log(fmt.Sprintf("%s 发动 [元素点燃]，对 %s 造成2点法术伤害并获得额外法术行动", ctx.User.Name, ctx.Target.Name))
	return nil
}

type ElementalistThunderStrikeHandler struct{ BaseHandler }

type ElementalistFreezeHandler struct{ BaseHandler }

type ElementalistWindBladeHandler struct{ BaseHandler }

type ElementalistMeteorHandler struct{ BaseHandler }

type ElementalistFireballHandler struct{ BaseHandler }

type ElementalistMoonlightHandler struct{ BaseHandler }

func elementalistMatchingElementCardIndices(user *model.Player, element model.Element) []int {
	if user == nil {
		return nil
	}
	out := make([]int, 0, len(user.Hand))
	for i, card := range user.Hand {
		if card.Element == element {
			out = append(out, i)
		}
	}
	return out
}

func (h *ElementalistThunderStrikeHandler) Execute(ctx *model.Context) error {
	if ctx.Target == nil {
		return fmt.Errorf("雷击需要目标")
	}
	if !hasElementCard(ctx.User, model.ElementThunder) {
		ctx.Game.InflictDamage(ctx.User.ID, ctx.Target.ID, 1, model.MagicAttack)
		ctx.Game.ModifyGem(string(ctx.User.Camp), 1)
		ctx.Game.Log(fmt.Sprintf("%s 发动 [雷击]，造成1点法术伤害并为阵营+1宝石", ctx.User.Name))
		return nil
	}
	matching := elementalistMatchingElementCardIndices(ctx.User, model.ElementThunder)
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
	if !hasElementCard(ctx.User, model.ElementWater) {
		ctx.Game.InflictDamage(ctx.User.ID, dmgTarget.ID, 1, model.MagicAttack)
		ctx.Game.Heal(healTarget.ID, 1)
		ctx.Game.Log(fmt.Sprintf("%s 发动 [冰冻]，对 %s 造成1点法术伤害并治疗 %s 1点", ctx.User.Name, dmgTarget.Name, healTarget.Name))
		return nil
	}
	matching := elementalistMatchingElementCardIndices(ctx.User, model.ElementWater)
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

func (h *ElementalistWindBladeHandler) Execute(ctx *model.Context) error {
	if ctx.Target == nil {
		return fmt.Errorf("风刃需要目标")
	}
	if !hasElementCard(ctx.User, model.ElementWind) {
		ctx.Game.InflictDamage(ctx.User.ID, ctx.Target.ID, 1, model.MagicAttack)
		addAttackAction(ctx.User, "风刃")
		ctx.Game.Log(fmt.Sprintf("%s 发动 [风刃]，造成1点法术伤害并获得额外攻击行动", ctx.User.Name))
		return nil
	}
	matching := elementalistMatchingElementCardIndices(ctx.User, model.ElementWind)
	if len(matching) == 0 {
		ctx.Game.InflictDamage(ctx.User.ID, ctx.Target.ID, 1, model.MagicAttack)
		addAttackAction(ctx.User, "风刃")
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

func (h *ElementalistMeteorHandler) Execute(ctx *model.Context) error {
	if ctx.Target == nil {
		return fmt.Errorf("陨石需要目标")
	}
	if !hasElementCard(ctx.User, model.ElementEarth) {
		ctx.Game.InflictDamage(ctx.User.ID, ctx.Target.ID, 1, model.MagicAttack)
		addMagicAction(ctx.User, "陨石")
		ctx.Game.Log(fmt.Sprintf("%s 发动 [陨石]，造成1点法术伤害并获得额外法术行动", ctx.User.Name))
		return nil
	}
	matching := elementalistMatchingElementCardIndices(ctx.User, model.ElementEarth)
	if len(matching) == 0 {
		ctx.Game.InflictDamage(ctx.User.ID, ctx.Target.ID, 1, model.MagicAttack)
		addMagicAction(ctx.User, "陨石")
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

func (h *ElementalistFireballHandler) Execute(ctx *model.Context) error {
	if ctx.Target == nil {
		return fmt.Errorf("火球需要目标")
	}
	if !hasElementCard(ctx.User, model.ElementFire) {
		ctx.Game.InflictDamage(ctx.User.ID, ctx.Target.ID, 2, model.MagicAttack)
		ctx.Game.Log(fmt.Sprintf("%s 发动 [火球]，造成2点法术伤害", ctx.User.Name))
		return nil
	}
	matching := elementalistMatchingElementCardIndices(ctx.User, model.ElementFire)
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

// --- 仲裁者 ---

type ArbiterLawHandler struct{ BaseHandler }

type ArbiterJudgmentTideHandler struct{ BaseHandler }

type ArbiterRitualHandler struct{ BaseHandler }

type ArbiterRitualBreakHandler struct{ BaseHandler }

type ArbiterDoomsdayHandler struct{ BaseHandler }

type ArbiterBalanceHandler struct{ BaseHandler }

func (h *ArbiterLawHandler) CanUse(ctx *model.Context) bool {
	return ctx != nil && ctx.User != nil && getToken(ctx.User, "arbiter_law_inited") == 0
}

func (h *ArbiterLawHandler) Execute(ctx *model.Context) error {
	ctx.User.Crystal += 2
	setToken(ctx.User, "arbiter_law_inited", 1)
	ctx.Game.Log(fmt.Sprintf("%s 的 [仲裁法则] 生效，获得2个蓝水晶", ctx.User.Name))
	return nil
}

func (h *ArbiterJudgmentTideHandler) Execute(ctx *model.Context) error {
	if ctx.EventCtx != nil && ctx.EventCtx.DamageVal != nil && *ctx.EventCtx.DamageVal <= 0 {
		return nil
	}
	v := addToken(ctx.User, "judgment", 1, 0, 4)
	ctx.Game.Log(fmt.Sprintf("%s 的 [审判浪潮] 触发，审判=%d", ctx.User.Name, v))
	return nil
}

func (h *ArbiterRitualHandler) CanUse(ctx *model.Context) bool {
	return !hasForm(ctx.User, model.FormArbiterJudgment) && ctx.User.Gem > 0
}

func (h *ArbiterRitualHandler) Execute(ctx *model.Context) error {
	if ctx.User.Gem <= 0 {
		return nil
	}
	ctx.User.Gem--
	enterForm(ctx.User, model.FormArbiterJudgment)
	ctx.User.MaxHand = 5
	ctx.Game.Log(fmt.Sprintf("%s 发动 [仲裁仪式]，进入审判形态，手牌上限恒定为5", ctx.User.Name))
	return nil
}

func (h *ArbiterRitualBreakHandler) CanUse(ctx *model.Context) bool {
	return hasForm(ctx.User, model.FormArbiterJudgment)
}

func (h *ArbiterRitualBreakHandler) Execute(ctx *model.Context) error {
	leaveForm(ctx.User, model.FormArbiterJudgment)
	if ctx.User.Character != nil && ctx.User.Character.MaxHand > 0 {
		ctx.User.MaxHand = ctx.User.Character.MaxHand
	} else {
		ctx.User.MaxHand = 6
	}
	ctx.Game.ModifyGem(string(ctx.User.Camp), 1)
	ctx.Game.Log(fmt.Sprintf("%s 发动 [仪式中断]，转正脱离审判形态并为阵营+1宝石", ctx.User.Name))
	return nil
}

func (h *ArbiterDoomsdayHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil {
		return false
	}
	return getToken(ctx.User, "judgment") > 0
}

func (h *ArbiterDoomsdayHandler) Execute(ctx *model.Context) error {
	if ctx.Target == nil {
		return fmt.Errorf("末日审判需要目标")
	}
	dmg := getToken(ctx.User, "judgment")
	setToken(ctx.User, "judgment", 0)
	if dmg > 0 {
		ctx.Game.InflictDamage(ctx.User.ID, ctx.Target.ID, dmg, model.MagicAttack)
	}
	ctx.Game.Log(fmt.Sprintf("%s 发动 [末日审判]，对 %s 造成%d点法术伤害", ctx.User.Name, ctx.Target.Name, dmg))
	return nil
}

func (h *ArbiterBalanceHandler) CanUse(ctx *model.Context) bool {
	return canPayCrystalLike(ctx, 1)
}

func (h *ArbiterBalanceHandler) Execute(ctx *model.Context) error {
	// 资源扣除由 UseSkill 统一处理，这里不重复扣费。
	v := addToken(ctx.User, "judgment", 1, 0, 4)
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type": "arbiter_balance_mode",
			"user_id":     ctx.User.ID,
			"judgment":    v,
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [判决天平]，审判=%d，等待选择分支", ctx.User.Name, v))
	return nil
}

// --- 冒险家 ---

type AdventurerFraudHandler struct{ BaseHandler }

type AdventurerLuckyFortuneHandler struct{ BaseHandler }

type AdventurerUndergroundLawHandler struct{ BaseHandler }

type AdventurerParadiseHandler struct{ BaseHandler }

type AdventurerStealSkyHandler struct{ BaseHandler }

func (h *AdventurerFraudHandler) CanUse(ctx *model.Context) bool {
	counts := map[model.Element]int{}
	for _, c := range ctx.User.Hand {
		counts[c.Element]++
	}
	for ele, n := range counts {
		// 弃2同系仅要求有同系牌；攻击系别在后续弹窗中单独选择（不含光/暗）
		if ele != "" && n >= 2 {
			return true
		}
		if n >= 3 {
			return true
		}
	}
	return false
}

func (h *AdventurerFraudHandler) Execute(ctx *model.Context) error {
	counts := map[model.Element]int{}
	for _, c := range ctx.User.Hand {
		counts[c.Element]++
	}
	canPick := false
	for ele, n := range counts {
		if ele == "" {
			continue
		}
		if n >= 2 {
			canPick = true
			break
		}
	}
	if !canPick {
		return nil
	}
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type": "adventurer_fraud_pick",
			"user_id":     ctx.User.ID,
			"user_ctx":    ctx,
			"fraud_target_id": func() string {
				if ctx.Target != nil {
					return ctx.Target.ID
				}
				return ""
			}(),
			"fraud_from_skill": true,
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [欺诈]，请先选择同系手牌", ctx.User.Name))
	return nil
}

func (h *AdventurerLuckyFortuneHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.Timing != model.TimingOnAttackDeclared {
		return false
	}
	if ctx.EventCtx == nil || ctx.EventCtx.Card == nil {
		return false
	}
	card := ctx.EventCtx.Card
	// 强运仅在“欺诈转化出的攻击”开始时自动触发。
	return card.ID == "fraud_virtual_attack" || card.Name == "欺诈"
}

func (h *AdventurerLuckyFortuneHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return nil
	}
	ctx.User.Crystal++
	ctx.Game.Log(fmt.Sprintf("%s 的 [强运] 触发，获得1蓝水晶", ctx.User.Name))
	return nil
}

func (h *AdventurerUndergroundLawHandler) CanUse(ctx *model.Context) bool {
	return ctx.EventCtx != nil && ctx.EventCtx.ActionType == model.ActionBuy
}

func (h *AdventurerUndergroundLawHandler) Execute(ctx *model.Context) error {
	ctx.Game.ModifyGem(string(ctx.User.Camp), 2)
	ctx.Game.Log(fmt.Sprintf("%s 的 [地下法则] 触发，战绩区+2红宝石", ctx.User.Name))
	return nil
}

func (h *AdventurerParadiseHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.Game == nil || ctx.EventCtx == nil {
		return false
	}
	if ctx.EventCtx.ActionType != model.ActionExtract {
		return false
	}
	all := ctx.Game.GetAllPlayers()
	for _, p := range all {
		if p != nil && p.Camp == ctx.User.Camp && p.ID != ctx.User.ID {
			return true
		}
	}
	return false
}

func (h *AdventurerParadiseHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return nil
	}
	transferGem := getSkillFlow(ctx.User, "adventurer_extract_last_gem")
	transferCrystal := getSkillFlow(ctx.User, "adventurer_extract_last_crystal")
	transferTotal := transferGem + transferCrystal
	if transferTotal <= 0 {
		setSkillFlow(ctx.User, "adventurer_extract_requires_paradise", 0)
		ctx.Game.Log(fmt.Sprintf("%s 的 [冒险者天堂] 未检测到本次提炼结果，效果取消", ctx.User.Name))
		return nil
	}

	all := ctx.Game.GetAllPlayers()
	var allyIDs []string
	for _, p := range all {
		if p == nil {
			continue
		}
		if p.Camp != ctx.User.Camp || p.ID == ctx.User.ID {
			continue
		}
		room := playerEnergyCap(p) - (p.Gem + p.Crystal)
		if room >= transferTotal {
			allyIDs = append(allyIDs, p.ID)
		}
	}
	if len(allyIDs) == 0 {
		setSkillFlow(ctx.User, "adventurer_extract_requires_paradise", 0)
		ctx.Game.Log(fmt.Sprintf("%s 的 [冒险者天堂] 无法发动：没有可完整承接%d点提炼能量的队友", ctx.User.Name, transferTotal))
		return nil
	}
	forceTransfer := getSkillFlow(ctx.User, "adventurer_extract_requires_paradise") > 0
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type":      "adventurer_paradise_target",
			"user_id":          ctx.User.ID,
			"ally_ids":         allyIDs,
			"transfer_gem":     transferGem,
			"transfer_crystal": transferCrystal,
			"transfer_total":   transferTotal,
			"from_pending":     forceTransfer,
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 的 [冒险者天堂] 触发，等待选择接收%d点提炼能量的队友", ctx.User.Name, transferTotal))
	return nil
}

func (h *AdventurerStealSkyHandler) CanUse(ctx *model.Context) bool {
	return canPayCrystalLike(ctx, 1)
}

func (h *AdventurerStealSkyHandler) Execute(ctx *model.Context) error {
	enemy := model.BlueCamp
	if ctx.User.Camp == model.BlueCamp {
		enemy = model.RedCamp
	}
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type": "adventurer_steal_sky_mode",
			"user_id":     ctx.User.ID,
			"enemy_camp":  string(enemy),
			"self_camp":   string(ctx.User.Camp),
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [偷天换日]，等待选择效果", ctx.User.Name))
	return nil
}

// --- 圣枪骑士 ---

type HolyLancerRevelationHandler struct{ BaseHandler }

type HolyLancerRadianceHandler struct{ BaseHandler }

type HolyLancerPunishmentHandler struct{ BaseHandler }

type HolyLancerHolyStrikeHandler struct{ BaseHandler }

type HolyLancerSkySpearHandler struct{ BaseHandler }

type HolyLancerEarthSpearHandler struct{ BaseHandler }

type HolyLancerPrayerHandler struct{ BaseHandler }

func (h *HolyLancerRevelationHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.Game == nil || ctx.User == nil {
		return nil
	}
	ctx.Game.RefreshPlayerDerivedState(ctx.User.ID)
	return nil
}

func (h *HolyLancerRadianceHandler) Execute(ctx *model.Context) error {
	for _, p := range ctx.Game.GetAllPlayers() {
		ctx.Game.Heal(p.ID, 1)
	}
	addAttackAction(ctx.User, "辉耀")
	ctx.Game.Log(fmt.Sprintf("%s 发动 [辉耀]，全场+1治疗并获得额外攻击行动", ctx.User.Name))
	return nil
}

func (h *HolyLancerPunishmentHandler) Execute(ctx *model.Context) error {
	if ctx.Target == nil {
		return fmt.Errorf("惩戒需要目标")
	}
	if ctx.Target.ID == ctx.User.ID {
		return fmt.Errorf("惩戒目标必须是其他角色")
	}
	if ctx.Target.Heal <= 0 {
		return fmt.Errorf("惩戒目标没有治疗，无法发动")
	}
	ctx.Target.Heal--
	if ctx.User.Heal < ctx.User.MaxHeal {
		ctx.User.Heal++
	}
	addAttackAction(ctx.User, "惩戒")
	ctx.Game.Log(fmt.Sprintf("%s 发动 [惩戒]，从 %s 转移1点治疗并获得额外攻击行动", ctx.User.Name, ctx.Target.Name))
	return nil
}

func (h *HolyLancerHolyStrikeHandler) CanUse(ctx *model.Context) bool {
	// 与地枪互斥：
	// 若当前“主动攻击命中”下地枪可发动，则先进入地枪响应窗口；
	// 仅当玩家不发动地枪（跳过响应）时，再由引擎补触发圣击治疗。
	if ctx == nil || ctx.User == nil || ctx.Timing != model.TimingOnHitCheck || ctx.EventCtx == nil || ctx.EventCtx.AttackInfo == nil {
		return false
	}
	info := ctx.EventCtx.AttackInfo
	if info.ActionType != string(model.ActionAttack) || !info.IsHit {
		return false
	}
	if info.CounterInitiator == "" && ctx.User.Heal > 0 {
		return false
	}
	return ctx.User.TurnState.UsedSkillCounts["holy_lancer_block_sacred_strike"] == 0
}

func (h *HolyLancerHolyStrikeHandler) Execute(ctx *model.Context) error {
	ctx.Game.Heal(ctx.User.ID, 1)
	return nil
}

func (h *HolyLancerSkySpearHandler) CanUse(ctx *model.Context) bool {
	if ctx.User.Heal < 2 {
		return false
	}
	if ctx.User.TurnState.UsedSkillCounts["holy_lancer_prayer"] > 0 {
		return false
	}
	if ctx.EventCtx == nil || ctx.EventCtx.AttackInfo == nil {
		return false
	}
	return ctx.EventCtx.AttackInfo.CounterInitiator == ""
}

func (h *HolyLancerSkySpearHandler) Execute(ctx *model.Context) error {
	ctx.User.Heal -= 2
	if ctx.EventCtx != nil && ctx.EventCtx.AttackInfo != nil {
		ctx.EventCtx.AttackInfo.CanBeResponded = false
	}
	// 通过令牌持久化“本次攻击无法应战”，避免攻击开始响应后的二次进入覆盖状态。
	ctx.User.TurnState.UsedSkillCounts["holy_lancer_sky_spear_no_counter"] = 1
	ctx.User.TurnState.UsedSkillCounts["holy_lancer_block_sacred_strike"] = 1
	ctx.Game.Log(fmt.Sprintf("%s 发动 [天枪]，移除2治疗，本次攻击不可应战", ctx.User.Name))
	return nil
}

func (h *HolyLancerEarthSpearHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.EventCtx == nil {
		return false
	}
	if ctx.Timing != model.TimingOnHitCheck {
		return false
	}
	if ctx.User.Heal <= 0 || ctx.EventCtx.DamageVal == nil {
		return false
	}
	// 地枪仅可在主动攻击命中后发动。
	if ctx.EventCtx.AttackInfo == nil {
		return false
	}
	if ctx.EventCtx.AttackInfo.ActionType != string(model.ActionAttack) {
		return false
	}
	if !ctx.EventCtx.AttackInfo.IsHit {
		return false
	}
	if ctx.EventCtx.AttackInfo.CounterInitiator != "" {
		return false
	}
	return true
}

func (h *HolyLancerEarthSpearHandler) Execute(ctx *model.Context) error {
	if ctx.EventCtx == nil || ctx.EventCtx.DamageVal == nil {
		return nil
	}
	x := ctx.User.Heal
	if x > 4 {
		x = 4
	}
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type": "holy_lancer_earth_spear_x",
			"user_id":     ctx.User.ID,
			"max_x":       x,
			"user_ctx":    ctx,
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [地枪]，等待选择X值", ctx.User.Name))
	return nil
}

func (h *HolyLancerPrayerHandler) CanUse(ctx *model.Context) bool {
	return ctx.User.Gem > 0
}

func (h *HolyLancerPrayerHandler) Execute(ctx *model.Context) error {
	ctx.User.Heal += 2
	if ctx.User.Heal > 5 {
		ctx.User.Heal = 5
	}
	ctx.User.TurnState.UsedSkillCounts["holy_lancer_prayer"] = 1
	addAttackAction(ctx.User, "圣光祈愈")
	ctx.Game.Log(fmt.Sprintf("%s 发动 [圣光祈愈]，治疗+2（上限5）并获得额外攻击行动", ctx.User.Name))
	return nil
}

// --- 14. 精灵射手 ---

type ElfElementalShotHandler struct{ BaseHandler }

type ElfAnimalCompanionHandler struct{ BaseHandler }

type ElfRitualHandler struct{ BaseHandler }

type ElfPetEmpowerHandler struct{ BaseHandler }

func (h *ElfElementalShotHandler) CanUse(ctx *model.Context) bool {
	if ctx.Timing != model.TimingOnAttackDeclared || ctx.EventCtx == nil || ctx.EventCtx.Card == nil {
		return false
	}
	if ctx.EventCtx.Card.Element == model.ElementDark {
		return false
	}
	if ctx.EventCtx.AttackInfo != nil && ctx.EventCtx.AttackInfo.CounterInitiator != "" {
		return false
	}
	hasMagic := false
	for _, c := range ctx.User.Hand {
		if c.Type == model.CardTypeMagic {
			hasMagic = true
			break
		}
	}
	return hasMagic || countElfBlessings(ctx.User) > 0
}

func (h *ElfElementalShotHandler) Execute(ctx *model.Context) error {
	if ctx.EventCtx == nil || ctx.EventCtx.Card == nil {
		return nil
	}
	hasMagic := false
	for _, c := range ctx.User.Hand {
		if c.Type == model.CardTypeMagic {
			hasMagic = true
			break
		}
	}
	hasBlessing := countElfBlessings(ctx.User) > 0
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type":       "elf_elemental_shot_cost",
			"user_id":           ctx.User.ID,
			"attack_element":    string(ctx.EventCtx.Card.Element),
			"can_discard_magic": hasMagic,
			"can_remove_bless":  hasBlessing,
			"user_ctx":          ctx,
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 可发动 [元素射击]，等待选择消耗方式", ctx.User.Name))
	return nil
}

func resolveElfForcedDrawDiscard(game model.IGameEngine, target *model.Player, prompt string, excludeBlessings bool) error {
	if game == nil || target == nil {
		return fmt.Errorf("动物伙伴结算目标无效")
	}
	game.DrawCards(target.ID, 1)
	if len(target.Hand) > game.GetMaxHand(target) {
		game.Log(fmt.Sprintf("[精灵射手] %s 摸牌后触发爆牌，本次弃1由爆牌弃牌结算承担", target.Name))
		return nil
	}
	game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: target.ID,
		Context: map[string]interface{}{
			"choice_type":       "system_discard_cards",
			"discard_subflow":   true,
			"discard_count":     1,
			"stay_in_turn":      true,
			"prompt":            prompt,
			"exclude_blessings": excludeBlessings,
		},
	})
	return nil
}

func (h *ElfAnimalCompanionHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.Game == nil || ctx.Target == nil || ctx.EventCtx == nil || ctx.EventCtx.DamageVal == nil {
		return false
	}
	if ctx.Timing != model.TimingOnDamageTaken || *ctx.EventCtx.DamageVal <= 0 {
		return false
	}
	if ctx.EventCtx.SourceID != ctx.User.ID || ctx.EventCtx.TargetID == "" || ctx.EventCtx.TargetID == ctx.User.ID {
		return false
	}
	if ctx.EventCtx.Card == nil || ctx.EventCtx.Card.Type != model.CardTypeAttack {
		return false
	}
	if ctx.EventCtx.AttackInfo == nil || ctx.EventCtx.AttackInfo.CounterInitiator != "" {
		return false
	}
	return ctx.User.IsActive
}

func (h *ElfAnimalCompanionHandler) Execute(ctx *model.Context) error {
	return resolveElfForcedDrawDiscard(ctx.Game, ctx.User, "【动物伙伴】请选择弃置1张牌：", true)
}

func (h *ElfRitualHandler) CanUse(ctx *model.Context) bool {
	return ctx.User.Gem > 0 && !hasForm(ctx.User, model.FormElfArcherRitual)
}

func (h *ElfRitualHandler) Execute(ctx *model.Context) error {
	if ctx.User.Gem <= 0 {
		return fmt.Errorf("精灵密仪需要至少1个红宝石")
	}
	ctx.User.Gem--
	enterForm(ctx.User, model.FormElfArcherRitual)
	before := len(ctx.User.Hand)
	ctx.Game.DrawCardsWithOptions(ctx.User.ID, 3, model.DrawOptions{
		PreventOverflow: true,
		Reason:          "elf_ritual",
	})

	if len(ctx.User.Hand)-before < 3 {
		return fmt.Errorf("精灵密仪抽取祝福数量不足")
	}
	cards := append([]model.Card{}, ctx.User.Hand[before:before+3]...)
	ctx.User.Hand = append(ctx.User.Hand[:before], ctx.User.Hand[before+3:]...)
	markElfBlessingCards(ctx.User, cards)
	ctx.Game.Log(fmt.Sprintf("%s 发动 [精灵密仪]，进入精灵祝福形态并获得3张祝福", ctx.User.Name))
	return nil
}

func (h *ElfPetEmpowerHandler) CanUse(ctx *model.Context) bool {
	if !canPayCrystalLike(ctx, 1) {
		return false
	}
	if ctx == nil || ctx.User == nil || ctx.Target == nil || ctx.EventCtx == nil || ctx.EventCtx.DamageVal == nil {
		return false
	}
	if ctx.Timing != model.TimingOnDamageTaken || *ctx.EventCtx.DamageVal <= 0 {
		return false
	}
	if ctx.EventCtx.SourceID != ctx.User.ID || ctx.EventCtx.TargetID != ctx.Target.ID {
		return false
	}
	if ctx.EventCtx.Card == nil || ctx.EventCtx.Card.Type != model.CardTypeAttack {
		return false
	}
	if ctx.EventCtx.AttackInfo == nil || ctx.EventCtx.AttackInfo.CounterInitiator != "" {
		return false
	}
	return ctx.User.IsActive && ctx.Target.Camp != ctx.User.Camp
}

func (h *ElfPetEmpowerHandler) Execute(ctx *model.Context) error {
	if !canPayCrystalLike(ctx, 1) {
		return fmt.Errorf("宠物强化需要至少1个蓝水晶")
	}
	if ctx == nil || ctx.Target == nil {
		return fmt.Errorf("宠物强化缺少受伤目标")
	}
	if !spendCrystalLike(ctx, 1) {
		return fmt.Errorf("宠物强化结算失败：水晶不足（红宝石可替代）")
	}
	ctx.Game.Log(fmt.Sprintf("%s 发动 [宠物强化]，动物伙伴效果改为目标摸1弃1", ctx.User.Name))
	return resolveElfForcedDrawDiscard(
		ctx.Game,
		ctx.Target,
		fmt.Sprintf("【宠物强化】%s 请弃置1张牌：", ctx.Target.Name),
		ctx.Target.Character != nil && ctx.Target.Character.ID == "elf_archer",
	)
}

// --- 15. 瘟疫法师 ---

type PlagueImmortalHandler struct{ BaseHandler }

type PlagueBlasphemyHandler struct{ BaseHandler }

type PlagueOutbreakHandler struct{ BaseHandler }

type PlagueDeathTouchHandler struct{ BaseHandler }

type PlagueToxicNovaHandler struct{ BaseHandler }

func (h *PlagueImmortalHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.Timing != model.TimingOnActionEnd || ctx.EventCtx == nil {
		return false
	}
	if ctx.EventCtx.ActionType != model.ActionMagic {
		return false
	}
	return ctx.User.IsActive
}

func (h *PlagueImmortalHandler) Execute(ctx *model.Context) error {
	if ctx.User.TurnState.UsedSkillCounts["plague_block_immortal"] > 0 {
		ctx.User.TurnState.UsedSkillCounts["plague_block_immortal"] = 0
		ctx.Game.Log(fmt.Sprintf("%s 的 [不朽] 本次被技能效果抑制", ctx.User.Name))
		return nil
	}
	ctx.Game.Heal(ctx.User.ID, 1)
	ctx.Game.Log(fmt.Sprintf("%s 的 [不朽] 触发，+1治疗", ctx.User.Name))
	return nil
}

func (h *PlagueBlasphemyHandler) Execute(ctx *model.Context) error { return nil }

func (h *PlagueOutbreakHandler) CanUse(ctx *model.Context) bool {
	return hasElementCard(ctx.User, model.ElementEarth)
}

func (h *PlagueOutbreakHandler) Execute(ctx *model.Context) error {
	ordered := reverseOrderPlayers(ctx.Game.GetAllPlayers(), ctx.User.ID)
	for _, p := range ordered {
		if p.ID == ctx.User.ID {
			continue
		}
		ctx.Game.AddPendingDamage(model.PendingDamage{
			SourceID:      ctx.User.ID,
			TargetID:      p.ID,
			Damage:        1,
			DamageType:    model.MagicAttack,
			SourceSkillID: "plague_outbreak",
		})
	}
	ctx.Game.Log(fmt.Sprintf("%s 发动 [瘟疫]，按逆序对其余角色各造成1点法术伤害", ctx.User.Name))
	return nil
}

func (h *PlagueDeathTouchHandler) CanUse(ctx *model.Context) bool {
	if ctx.User.Heal < 2 {
		return false
	}
	counts := map[model.Element]int{}
	for _, c := range ctx.User.Hand {
		if c.Element != "" {
			counts[c.Element]++
		}
	}
	for _, n := range counts {
		if n >= 2 {
			return true
		}
	}
	return false
}

func (h *PlagueDeathTouchHandler) Execute(ctx *model.Context) error {
	if ctx.Target == nil {
		return fmt.Errorf("死亡之触需要1名敌方目标")
	}
	if ctx.User.Heal < 2 {
		return fmt.Errorf("死亡之触需要至少2点治疗")
	}
	counts := map[model.Element]int{}
	for _, c := range ctx.User.Hand {
		if c.Element != "" {
			counts[c.Element]++
		}
	}
	var elements []string
	for _, ele := range []model.Element{
		model.ElementEarth, model.ElementWater, model.ElementFire,
		model.ElementWind, model.ElementThunder, model.ElementLight, model.ElementDark,
	} {
		if counts[ele] >= 2 {
			elements = append(elements, string(ele))
		}
	}
	if len(elements) == 0 {
		return fmt.Errorf("死亡之触需要至少2张同系牌")
	}
	// 该技能不触发不朽：先设置抑制标记，覆盖 UseSkill 的阶段结束触发。
	ctx.User.TurnState.UsedSkillCounts["plague_block_immortal"] = 1
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type":      "plague_death_touch_element",
			"user_id":          ctx.User.ID,
			"target_id":        ctx.Target.ID,
			"elements":         elements,
			"max_heal":         ctx.User.Heal,
			"element_counts":   counts,
			"selected_indices": []int{},
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [死亡之触]，等待选择X/Y与目标", ctx.User.Name))
	return nil
}

func (h *PlagueToxicNovaHandler) CanUse(ctx *model.Context) bool {
	return ctx.User.Gem > 0
}

func (h *PlagueToxicNovaHandler) Execute(ctx *model.Context) error {
	ordered := reverseOrderPlayers(ctx.Game.GetAllPlayers(), ctx.User.ID)
	for _, p := range ordered {
		if p.ID == ctx.User.ID {
			continue
		}
		ctx.Game.AddPendingDamage(model.PendingDamage{
			SourceID:   ctx.User.ID,
			TargetID:   p.ID,
			Damage:     2,
			DamageType: model.MagicAttack,
		})
	}
	ctx.Game.Heal(ctx.User.ID, 1)
	ctx.Game.Log(fmt.Sprintf("%s 发动 [剧毒新星]，对其余角色各造成2点法术伤害", ctx.User.Name))
	return nil
}

// --- 16. 魔剑士 ---

type MagicSwordsmanAsuraComboHandler struct{ BaseHandler }

type MagicSwordsmanShadowGatherHandler struct{ BaseHandler }

type MagicSwordsmanShadowPowerHandler struct{ BaseHandler }

type MagicSwordsmanShadowRejectHandler struct{ BaseHandler }

type MagicSwordsmanShadowMeteorHandler struct{ BaseHandler }

type MagicSwordsmanYellowSpringHandler struct{ BaseHandler }

func (h *MagicSwordsmanAsuraComboHandler) CanUse(ctx *model.Context) bool {
	if ctx.Timing != model.TimingOnActionEnd || ctx.EventCtx == nil {
		return false
	}
	if ctx.EventCtx.ActionType != model.ActionAttack {
		return false
	}
	// 修罗连斩响应“攻击行动结束”，不应在应战攻击结束后触发。
	if ctx.EventCtx.AttackInfo != nil && ctx.EventCtx.AttackInfo.CounterInitiator != "" {
		return false
	}
	return true
}

func (h *MagicSwordsmanAsuraComboHandler) Execute(ctx *model.Context) error {
	model.AppendAttackAction(ctx.User, "修罗连斩", model.ElementFire)
	ctx.Game.Log(fmt.Sprintf("%s 发动 [修罗连斩]，获得额外火系攻击行动", ctx.User.Name))
	return nil
}

func (h *MagicSwordsmanShadowGatherHandler) CanUse(ctx *model.Context) bool {
	return !hasForm(ctx.User, model.FormMagicSwordsmanShadow)
}

func (h *MagicSwordsmanShadowGatherHandler) Execute(ctx *model.Context) error {
	enterForm(ctx.User, model.FormMagicSwordsmanShadow)
	ctx.Game.AddPendingDamage(model.PendingDamage{
		SourceID:   ctx.User.ID,
		TargetID:   ctx.User.ID,
		Damage:     1,
		DamageType: model.MagicAttack,
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [暗影凝聚]，进入暗影形态并承受1点法术伤害", ctx.User.Name))
	return nil
}

func (h *MagicSwordsmanShadowPowerHandler) Execute(ctx *model.Context) error { return nil }

func (h *MagicSwordsmanShadowRejectHandler) Execute(ctx *model.Context) error { return nil }

func (h *MagicSwordsmanShadowMeteorHandler) CanUse(ctx *model.Context) bool {
	if !hasForm(ctx.User, model.FormMagicSwordsmanShadow) {
		return false
	}
	count := 0
	for _, c := range ctx.User.Hand {
		if c.Type == model.CardTypeMagic {
			count++
		}
	}
	return count >= 2
}

func (h *MagicSwordsmanShadowMeteorHandler) Execute(ctx *model.Context) error {
	if !hasForm(ctx.User, model.FormMagicSwordsmanShadow) {
		return fmt.Errorf("暗影流星仅可在暗影形态下发动")
	}
	if ctx.Target == nil {
		return fmt.Errorf("暗影流星需要1名敌方目标")
	}
	ctx.Game.AddPendingDamage(model.PendingDamage{
		SourceID:   ctx.User.ID,
		TargetID:   ctx.Target.ID,
		Damage:     2,
		DamageType: model.MagicAttack,
	})
	camp := string(ctx.User.Camp)
	total := ctx.Game.GetCampGems(camp) + ctx.Game.GetCampCrystals(camp)
	if total >= 2 {
		ctx.Game.PushInterrupt(&model.Interrupt{
			Type:     model.InterruptChoice,
			PlayerID: ctx.User.ID,
			Context: map[string]interface{}{
				"choice_type": "ms_shadow_meteor_release_confirm",
				"user_id":     ctx.User.ID,
				"camp":        camp,
			},
		})
	}
	ctx.Game.Log(fmt.Sprintf("%s 发动 [暗影流星]，对 %s 造成2点法术伤害", ctx.User.Name, model.GetPlayerDisplayName(ctx.Target)))
	return nil
}

func (h *MagicSwordsmanYellowSpringHandler) CanUse(ctx *model.Context) bool {
	if ctx.Timing != model.TimingOnAttackDeclared || ctx.EventCtx == nil || ctx.EventCtx.AttackInfo == nil {
		return false
	}
	if ctx.EventCtx.AttackInfo.CounterInitiator != "" {
		return false
	}
	return ctx.User.Gem > 0
}

func (h *MagicSwordsmanYellowSpringHandler) Execute(ctx *model.Context) error {
	if ctx.User.Gem <= 0 {
		return fmt.Errorf("黄泉震颤需要至少1个红宝石")
	}
	ctx.User.Gem--
	if ctx.EventCtx != nil && ctx.EventCtx.AttackInfo != nil {
		ctx.EventCtx.AttackInfo.CanBeResponded = false
	}
	ctx.User.TurnState.UsedSkillCounts["ms_yellow_spring_pending"] = 1
	ctx.Game.Log(fmt.Sprintf("%s 发动 [黄泉震颤]，本次攻击不可应战", ctx.User.Name))
	return nil
}

// --- 17. 血色剑灵 ---

type CrimsonBloodThornsHandler struct{ BaseHandler }

type CrimsonFlashHandler struct{ BaseHandler }

type CrimsonBloodRoseHandler struct{ BaseHandler }

type CrimsonBloodBarrierHandler struct{ BaseHandler }

type CrimsonRoseCourtyardHandler struct{ BaseHandler }

type CrimsonDanceHandler struct{ BaseHandler }

func (h *CrimsonBloodThornsHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.Timing != model.TimingOnHitCheck || ctx.EventCtx == nil || ctx.EventCtx.AttackInfo == nil {
		return false
	}
	info := ctx.EventCtx.AttackInfo
	return info.ActionType == string(model.ActionAttack) && info.IsHit
}

func (h *CrimsonBloodThornsHandler) Execute(ctx *model.Context) error {
	cur := addBlood(ctx.User, 1)
	ctx.Game.Log(fmt.Sprintf("%s 的 [血色荆棘] 触发，鲜血=%d", ctx.User.Name, cur))
	return nil
}

func (h *CrimsonFlashHandler) CanUse(ctx *model.Context) bool {
	if ctx.Timing != model.TimingOnActionEnd || ctx.EventCtx == nil {
		return false
	}
	if ctx.EventCtx.ActionType != model.ActionAttack {
		return false
	}
	// 赤色一闪只响应主动攻击行动结束。
	if ctx.EventCtx.AttackInfo != nil && ctx.EventCtx.AttackInfo.CounterInitiator != "" {
		return false
	}
	return getToken(ctx.User, "css_blood") > 0
}

func (h *CrimsonFlashHandler) Execute(ctx *model.Context) error {
	if getToken(ctx.User, "css_blood") <= 0 {
		return nil
	}
	addBlood(ctx.User, -1)
	ctx.Game.AddPendingDamage(model.PendingDamage{
		SourceID:   ctx.User.ID,
		TargetID:   ctx.User.ID,
		Damage:     2,
		DamageType: model.MagicAttack,
	})
	addAttackAction(ctx.User, "赤色一闪")
	ctx.Game.Log(fmt.Sprintf("%s 发动 [赤色一闪]，移除1鲜血并获得额外攻击行动", ctx.User.Name))
	return nil
}

func (h *CrimsonBloodRoseHandler) CanUse(ctx *model.Context) bool {
	return getToken(ctx.User, "css_blood") >= 2
}

func (h *CrimsonBloodRoseHandler) Execute(ctx *model.Context) error {
	if len(ctx.Targets) != 2 {
		return fmt.Errorf("血染蔷薇需要恰好2名目标")
	}
	if getToken(ctx.User, "css_blood") < 2 {
		return fmt.Errorf("鲜血不足")
	}
	healReduceTarget := ctx.Targets[0]
	healGainTarget := ctx.Targets[1]
	if healReduceTarget == nil || healGainTarget == nil {
		return fmt.Errorf("血染蔷薇目标无效")
	}
	if healGainTarget.Camp != ctx.User.Camp {
		return fmt.Errorf("血染蔷薇的第2个目标必须是我方角色")
	}
	addBlood(ctx.User, -2)
	if healReduceTarget.Heal > 0 {
		loss := 2
		if healReduceTarget.Heal < loss {
			loss = healReduceTarget.Heal
		}
		healReduceTarget.Heal -= loss
	}
	if ctx.Game.GetCampCrystals(string(ctx.User.Camp)) > 0 {
		ctx.Game.ModifyCrystal(string(ctx.User.Camp), -1)
		ctx.Game.ModifyGem(string(ctx.User.Camp), 1)
	}
	ctx.Game.Heal(healGainTarget.ID, 1)
	hasRoseCourtyard := false
	for _, fc := range ctx.User.Field {
		if fc != nil && fc.Mode == model.FieldEffect && fc.Effect == model.EffectRoseCourtyard {
			hasRoseCourtyard = true
			break
		}
	}
	if hasRoseCourtyard {
		for _, p := range ctx.Game.GetAllPlayers() {
			ctx.Game.AddPendingDamage(model.PendingDamage{
				SourceID:   ctx.User.ID,
				TargetID:   p.ID,
				Damage:     1,
				DamageType: model.MagicAttack,
			})
		}
	}
	ctx.Game.Log(fmt.Sprintf("%s 发动 [血染蔷薇]：%s -2治疗，阵营1水晶转1宝石，%s +1治疗", ctx.User.Name, healReduceTarget.Name, healGainTarget.Name))
	return nil
}

func (h *CrimsonBloodBarrierHandler) CanUse(ctx *model.Context) bool {
	if ctx.Timing != model.TimingOnDamageTaken || ctx.EventCtx == nil {
		return false
	}
	if !ctx.Flags["IsMagicDamage"] {
		return false
	}
	if getToken(ctx.User, "css_blood_barrier_lock") > 0 {
		return false
	}
	return getToken(ctx.User, "css_blood") > 0
}

func (h *CrimsonBloodBarrierHandler) Execute(ctx *model.Context) error {
	if getToken(ctx.User, "css_blood") <= 0 {
		return nil
	}
	setToken(ctx.User, "css_blood_barrier_lock", 1)
	addBlood(ctx.User, -1)
	if ctx.EventCtx != nil && ctx.EventCtx.DamageVal != nil && *ctx.EventCtx.DamageVal > 0 {
		*ctx.EventCtx.DamageVal--
	}
	sourceID := ""
	if ctx.EventCtx != nil {
		sourceID = strings.TrimSpace(ctx.EventCtx.SourceID)
	}
	if sourceID != "" && sourceID != ctx.User.ID {
		ctx.Game.AddPendingDamage(model.PendingDamage{
			SourceID:   ctx.User.ID,
			TargetID:   sourceID,
			Damage:     1,
			DamageType: model.MagicAttack,
		})
	}
	ctx.Game.Log(fmt.Sprintf("%s 发动 [血气屏障]，本次法术伤害-1，并对伤害来源造成1点法术伤害", ctx.User.Name))
	return nil
}

func (h *CrimsonRoseCourtyardHandler) Execute(ctx *model.Context) error { return nil }

func (h *CrimsonDanceHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.User.Character == nil {
		return false
	}
	if !(canPayCrystalLike(ctx, 1) || ctx.User.Gem > 0) {
		return false
	}
	return ctx.User.HasExclusiveCard(ctx.User.Character.ID, "血蔷薇庭院")
}

func (h *CrimsonDanceHandler) Execute(ctx *model.Context) error {
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type": "css_dance_mode",
			"user_id":     ctx.User.ID,
			"can_crystal": canPayCrystalLike(ctx, 1),
			"can_gem":     ctx.User.Gem > 0,
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [散华轮舞]，等待选择模式", ctx.User.Name))
	return nil
}

func reverseOrderPlayers(players []*model.Player, sourceID string) []*model.Player {
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

func countElfBlessings(p *model.Player) int {
	if p == nil {
		return 0
	}
	return len(elfBlessingCoverCards(p))
}

func markElfBlessingCards(p *model.Player, cards []model.Card) {
	if p == nil || len(cards) == 0 {
		return
	}
	existsBless := map[string]bool{}
	for _, c := range elfBlessingCards(p) {
		if c.ID != "" {
			existsBless[c.ID] = true
		}
	}
	for _, c := range cards {
		if c.ID == "" {
			continue
		}
		if existsBless[c.ID] {
			continue
		}
		p.AddFieldCard(&model.FieldCard{
			Card:     c,
			OwnerID:  p.ID,
			SourceID: p.ID,
			Mode:     model.FieldCover,
			Effect:   model.EffectElfBlessing,
			Hook:     model.FieldHookManual,
		})
		existsBless[c.ID] = true
	}
	blessings := elfBlessingCards(p)
	p.Blessings = blessings
	blessingIDs := map[string]bool{}
	for _, c := range blessings {
		if c.ID != "" {
			blessingIDs[c.ID] = true
		}
	}
	newZone := make([]string, 0, len(p.CharaZone)+len(blessings))
	zoneHas := map[string]bool{}
	for _, z := range p.CharaZone {
		if !strings.HasPrefix(z, "elf_blessing:") {
			newZone = append(newZone, z)
			zoneHas[z] = true
			continue
		}
		cardID := strings.TrimPrefix(z, "elf_blessing:")
		if blessingIDs[cardID] {
			newZone = append(newZone, z)
			zoneHas[z] = true
		}
	}
	for _, c := range blessings {
		if c.ID == "" {
			continue
		}
		key := "elf_blessing:" + c.ID
		if zoneHas[key] {
			continue
		}
		newZone = append(newZone, key)
	}
	p.CharaZone = newZone
}

func elfBlessingCoverCards(p *model.Player) []*model.FieldCard {
	if p == nil {
		return nil
	}
	return p.GetCoverCardsByEffect(model.EffectElfBlessing)
}

func elfBlessingCards(p *model.Player) []model.Card {
	covers := elfBlessingCoverCards(p)
	if len(covers) == 0 {
		return nil
	}
	out := make([]model.Card, 0, len(covers))
	for _, fc := range covers {
		if fc == nil {
			continue
		}
		out = append(out, fc.Card)
	}
	return out
}

func addBlood(p *model.Player, delta int) int {
	if p == nil {
		return 0
	}
	if p.Tokens == nil {
		p.Tokens = map[string]int{}
	}
	capV := p.Tokens["css_blood_cap"]
	if capV <= 0 {
		capV = 3
	}
	cur := p.Tokens["css_blood"] + delta
	if cur < 0 {
		cur = 0
	}
	if cur > capV {
		cur = capV
	}
	p.Tokens["css_blood"] = cur
	return cur
}

// 工具：为了避免不同阶段删除手牌导致索引错乱，批量删除时统一按降序处理。
func removeHandByIndices(p *model.Player, indices []int) []model.Card {
	sort.Sort(sort.Reverse(sort.IntSlice(indices)))
	var out []model.Card
	for _, i := range indices {
		if i < 0 || i >= len(p.Hand) {
			continue
		}
		out = append(out, p.Hand[i])
		p.Hand = append(p.Hand[:i], p.Hand[i+1:]...)
	}
	return out
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
