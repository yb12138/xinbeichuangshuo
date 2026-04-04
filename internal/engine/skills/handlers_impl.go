package skills

import (
	"fmt"
	"sort"
	"starcup-engine/internal/model"
	"strings"
)

func hasAssassinStealthForm(p *model.Player) bool {
	return p != nil && p.Form == model.FormAssassinStealth
}

// --- Angel Handlers ---

type HolyShieldHandler struct{}

func (h *HolyShieldHandler) CanUse(ctx *model.Context) bool {
	// 1. 触发时机必须是受伤时
	if ctx.Trigger != model.TriggerOnDamageTaken {
		return false
	}

	// 2. 必须有伤害值上下文，且伤害值 > 0
	// 如果伤害已经被其他技能减为0了，圣盾就不需要触发了（省一个盾）
	if ctx.TriggerCtx == nil || ctx.TriggerCtx.DamageVal == nil || *ctx.TriggerCtx.DamageVal <= 0 {
		return false
	}

	// [重构] 3. 攻击伤害圣盾已经在 processPendingAttackHit 开头检查过了
	// 这里只处理魔弹伤害（法术伤害但卡牌名是"魔弹"）
	if ctx.Flags["IsMagicDamage"] {
		// 法术伤害：检查是否是魔弹
		if ctx.TriggerCtx.Card == nil || strings.TrimSpace(ctx.TriggerCtx.Card.Name) != "魔弹" {
			return false
		}
	} else {
		// 攻击伤害：这里不再处理（已在前面检查过）
		return false
	}

	// 4. 检查玩家场上是否真的有【圣盾】效果牌
	// 列风技：本次攻击无视圣盾
	if ctx.Flags["ignore_shield"] {
		return false
	}

	hasShield := false
	for _, fc := range ctx.User.Field {
		if fc.Mode == model.FieldEffect && fc.Effect == model.EffectShield {
			hasShield = true
			break
		}
	}
	return hasShield
}

// Execute 执行抵消逻辑
func (h *HolyShieldHandler) Execute(ctx *model.Context) error {
	// 1. 抵消伤害：直接修改指针指向的值
	originalDamage := *ctx.TriggerCtx.DamageVal
	*ctx.TriggerCtx.DamageVal = 0

	if ctx.Selections != nil {
		ctx.Selections["holy_shield_triggered"] = true
	}
	ctx.Game.Log(fmt.Sprintf("[Shield] %s 的【圣盾】自动触发，抵消了 %d 点伤害！", ctx.User.Name, originalDamage))
	ctx.Game.NotifyActionStep(fmt.Sprintf("%s 的【圣盾】触发，抵消了 %d 点伤害", ctx.User.Name, originalDamage))

	// 2. 移除圣盾状态（移除一张牌）
	// 我们需要精确移除一张圣盾牌
	newField := make([]*model.FieldCard, 0)
	removed := false

	for _, fc := range ctx.User.Field {
		// 找到第一张圣盾并移除
		if !removed && fc.Mode == model.FieldEffect && fc.Effect == model.EffectShield {
			removed = true
			// 调用引擎接口将牌放入弃牌堆
			ctx.Game.DiscardCard(fc)
			continue
		}
		newField = append(newField, fc)
	}

	// 更新玩家场上牌
	ctx.User.Field = newField

	return nil
}

type AngelBondHandler struct{ BaseHandler }

type basicEffectOption struct {
	TargetID    string
	TargetName  string
	FieldIndex  int
	Effect      model.EffectType
	DisplayName string
	Label       string
}

func basicEffectLabel(effect model.EffectType) string {
	switch effect {
	case model.EffectShield:
		return "圣盾"
	case model.EffectWeak:
		return "虚弱"
	case model.EffectPoison:
		return "中毒"
	case model.EffectSealFire:
		return "火之封印"
	case model.EffectSealWater:
		return "水之封印"
	case model.EffectSealEarth:
		return "地之封印"
	case model.EffectSealWind:
		return "风之封印"
	case model.EffectSealThunder:
		return "雷之封印"
	case model.EffectPowerBlessing:
		return "威力赐福"
	case model.EffectSwiftBlessing:
		return "迅捷赐福"
	default:
		return string(effect)
	}
}

func collectBasicEffectOptions(players ...*model.Player) []basicEffectOption {
	options := make([]basicEffectOption, 0)
	for _, player := range players {
		if player == nil {
			continue
		}
		for idx, fc := range player.Field {
			if fc == nil || fc.Mode != model.FieldEffect || !model.IsBasicEffect(string(fc.Effect)) {
				continue
			}
			displayName := basicEffectLabel(fc.Effect)
			options = append(options, basicEffectOption{
				TargetID:    player.ID,
				TargetName:  player.Name,
				FieldIndex:  idx,
				Effect:      fc.Effect,
				DisplayName: displayName,
				Label:       fmt.Sprintf("%s：%s", player.Name, displayName),
			})
		}
	}
	return options
}

func encodeBasicEffectOptions(options []basicEffectOption) []map[string]interface{} {
	encoded := make([]map[string]interface{}, 0, len(options))
	for _, option := range options {
		encoded = append(encoded, map[string]interface{}{
			"id":           fmt.Sprintf("%s|%d|%s", option.TargetID, option.FieldIndex, option.Effect),
			"target_id":    option.TargetID,
			"field_index":  option.FieldIndex,
			"effect":       string(option.Effect),
			"display_name": option.DisplayName,
			"label":        option.Label,
		})
	}
	return encoded
}

func (h *AngelBondHandler) CanUse(ctx *model.Context) bool {
	if ctx.TriggerCtx == nil || ctx.User == nil {
		return false
	}

	// 场景 A: 自己主动移除基础效果
	if ctx.Trigger == model.TriggerOnBuffRemoved {
		if ctx.TriggerCtx.SourceID != ctx.User.ID {
			return false
		}
		return model.IsBasicEffect(ctx.TriggerCtx.BuffID)
	}

	// 场景 B: 自己放置了圣盾效果（包括将独有牌当作圣盾使用的情况）
	if ctx.Trigger == model.TriggerOnBuffAdded {
		return ctx.TriggerCtx.SourceID == ctx.User.ID && ctx.TriggerCtx.BuffID == string(model.EffectShield)
	}
	return false
}

func (h *AngelBondHandler) Execute(ctx *model.Context) error {
	// 改为弹窗选择任意目标角色 +1 治疗
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return nil
	}
	targetIDs := make([]string, 0)
	for _, p := range ctx.Game.GetAllPlayers() {
		if p == nil {
			continue
		}
		targetIDs = append(targetIDs, p.ID)
	}
	if len(targetIDs) == 0 {
		return nil
	}
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type": "angel_bond_heal_target",
			"user_id":     ctx.User.ID,
			"target_ids":  targetIDs,
			// 规则：天使羁绊是响应链中的插入选择，需携带当前恢复点以便结算后回到原流程。
			"resume_phase": ctx.Selections["current_resume_point"],
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 的 [天使羁绊] 触发：请选择1名角色获得+1治疗", ctx.User.Name))
	return nil
}

type AngelBlessingHandler struct{ BaseHandler }

func (h *AngelBlessingHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil {
		return false
	}
	for _, card := range ctx.User.Hand {
		if card.Element == model.ElementWater {
			return true
		}
	}
	return false
}

func (h *AngelBlessingHandler) Execute(ctx *model.Context) error {
	// 天使祝福：弃1张水系牌，指定1名玩家给你2张牌，或指定2名玩家各给你1张牌。
	// 规则：如果指定的目标手牌为空或者不够数量，能弃几张弃几张。
	targets := ctx.Targets
	if len(targets) == 0 && ctx.Target != nil {
		targets = []*model.Player{ctx.Target}
	}

	if len(targets) == 0 {
		return fmt.Errorf("天使祝福需要指定目标")
	}

	receiverID := ctx.User.ID

	if len(targets) == 1 {
		// 模式 1: 1 名目标，给 2 张牌（如果手牌不足，能给几张给几张）
		target := targets[0]
		giveCount := 2
		if len(target.Hand) < giveCount {
			giveCount = len(target.Hand)
		}
		if giveCount > 0 {
			ctx.Game.PushInterrupt(&model.Interrupt{
				Type:     model.InterruptGiveCards,
				PlayerID: target.ID,
				Context: map[string]interface{}{
					"give_count":  giveCount,
					"receiver_id": receiverID,
				},
			})
			ctx.Game.Log(fmt.Sprintf("%s 发动天使祝福，%s 需选择 %d 张牌交给 %s", ctx.User.Name, target.Name, giveCount, ctx.User.Name))
		} else {
			ctx.Game.Log(fmt.Sprintf("%s 发动天使祝福，但 %s 没有手牌可交", ctx.User.Name, target.Name))
		}
	} else if len(targets) == 2 {
		// 模式 2: 2 名目标，各给 1 张牌（先推第二个进队列，再推第一个；手牌不足则跳过）
		for i := len(targets) - 1; i >= 0; i-- {
			t := targets[i]
			if len(t.Hand) >= 1 {
				ctx.Game.PushInterrupt(&model.Interrupt{
					Type:     model.InterruptGiveCards,
					PlayerID: t.ID,
					Context: map[string]interface{}{
						"give_count":  1,
						"receiver_id": receiverID,
					},
				})
			} else {
				ctx.Game.Log(fmt.Sprintf("%s 没有手牌可交给 %s", t.Name, ctx.User.Name))
			}
		}
		ctx.Game.Log(fmt.Sprintf("%s 发动天使祝福，%s 和 %s 需各选择 1 张牌交给 %s",
			ctx.User.Name, targets[0].Name, targets[1].Name, ctx.User.Name))
	} else {
		return fmt.Errorf("天使祝福只能指定 1 名或 2 名目标")
	}

	return nil
}

type AngelCleanseHandler struct{ BaseHandler }

func (h *AngelCleanseHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil {
		return false
	}
	for _, card := range ctx.User.Hand {
		if card.Element == model.ElementWind {
			return true
		}
	}
	return false
}

func (h *AngelCleanseHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Target == nil || ctx.Game == nil {
		return fmt.Errorf("风之洁净上下文无效")
	}
	options := collectBasicEffectOptions(ctx.Target)
	if len(options) == 0 {
		return fmt.Errorf("%s 面前没有可移除的基础效果", ctx.Target.Name)
	}
	if len(options) > 1 {
		ctx.Game.PushInterrupt(&model.Interrupt{
			Type:     model.InterruptChoice,
			PlayerID: ctx.User.ID,
			Context: map[string]interface{}{
				"choice_type":  "basic_effect_pick",
				"user_id":      ctx.User.ID,
				"skill_name":   "风之洁净",
				"operation":    "remove",
				"resume_phase": model.TurnStageActionExecution,
				"prompt":       "【风之洁净】请选择要移除的基础效果：",
				"options":      encodeBasicEffectOptions(options),
			},
		})
		ctx.Game.Log(fmt.Sprintf("%s 发动 [风之洁净]，请选择要移除的基础效果", ctx.User.Name))
		return nil
	}

	selected := options[0]
	if !ctx.Game.RemoveFieldCardBy(selected.TargetID, selected.Effect, ctx.User.ID) {
		return fmt.Errorf("%s 面前的基础效果已不存在", ctx.Target.Name)
	}
	ctx.Game.Log(fmt.Sprintf("%s 的 [风之洁净] 发动，移除了 %s", ctx.User.Name, selected.Label))
	return nil
}

type AngelSongHandler struct{ BaseHandler }

func (h *AngelSongHandler) CanUse(ctx *model.Context) bool {
	if !canPayCrystalLike(ctx, 1) { // 需要1个水晶，可由红宝石替代
		return false
	}
	if ctx == nil || ctx.Game == nil {
		return false
	}
	for _, p := range ctx.Game.GetAllPlayers() {
		if p == nil {
			continue
		}
		for _, fc := range p.Field {
			if fc.Mode == model.FieldEffect && model.IsBasicEffect(string(fc.Effect)) {
				return true
			}
		}
	}
	return false
}

func (h *AngelSongHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("天使之歌上下文无效")
	}
	options := collectBasicEffectOptions(ctx.Game.GetAllPlayers()...)
	if len(options) == 0 {
		return fmt.Errorf("发动天使之歌失败：场上没有可移除的基础效果")
	}
	if !spendCrystalLike(ctx, 1) {
		return fmt.Errorf("发动天使之歌失败：水晶不足（红宝石可替代）")
	}
	ctx.Game.Log(fmt.Sprintf("%s 消耗 1 水晶（可由红宝石替代）发动 [天使之歌]，请选择要移除的基础效果", ctx.User.Name))
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type":   "basic_effect_pick",
			"user_id":       ctx.User.ID,
			"skill_name":    "天使之歌",
			"operation":     "remove",
			"resume_phase":  model.TurnStageActionStart,
			"waiting_phase": model.TurnStageActionStart,
			"prompt":        "【天使之歌】请选择要移除的基础效果：",
			"options":       encodeBasicEffectOptions(options),
		},
	})
	return nil
}

// 辅助逻辑，Cleanse 和 Song 共用
func (h *BaseHandler) removeBasicEffectLogic(ctx *model.Context) error {
	if ctx.Target == nil {
		return nil
	}

	// 1. 寻找该目标身上所有的基础效果
	var basicBuffs []model.EffectType
	for _, fc := range ctx.Target.Field {
		if fc.Mode == model.FieldEffect && model.IsBasicEffect(string(fc.Effect)) {
			basicBuffs = append(basicBuffs, fc.Effect)
		}
	}

	if len(basicBuffs) == 0 {
		return nil
	}

	// 2. 决定移除哪一个
	// 如果前端传来了指定的 Buff 名称 (ctx.Args[0])
	targetBuff := basicBuffs[0]
	if len(ctx.Args) > 0 {
		requested := model.EffectType(ctx.Args[0])
		for _, b := range basicBuffs {
			if b == requested {
				targetBuff = requested
				break
			}
		}
	}

	// 3. 执行移除
	// removeFieldCard 内部应该会触发 TriggerOnBuffRemoved，从而连锁触发 天使羁绊
	ctx.Game.RemoveFieldCardBy(ctx.Target.ID, targetBuff, ctx.User.ID)
	return nil
}

type GodProtectionHandler struct{ BaseHandler }

func (h *GodProtectionHandler) CanUse(ctx *model.Context) bool {
	// 1. 必须是因为法术伤害
	if !ctx.Flags["IsMagicDamage"] {
		return false
	}
	// 2. 必须可支付至少1点水晶（可由红宝石替代）
	if !canPayCrystalLike(ctx, 1) {
		return false
	}
	// 3. 士气损失必须大于0
	if ctx.TriggerCtx.DamageVal == nil || *ctx.TriggerCtx.DamageVal <= 0 {
		return false
	}
	return true
}

func (h *GodProtectionHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil || ctx.TriggerCtx == nil || ctx.TriggerCtx.DamageVal == nil {
		return nil
	}
	angel := ctx.User
	loss := *ctx.TriggerCtx.DamageVal
	usable := ctx.Game.GetUsableCrystal(angel.ID)
	maxX := loss
	if maxX > usable {
		maxX = usable
	}
	if maxX <= 0 {
		return nil
	}
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: angel.ID,
		Context: map[string]interface{}{
			"choice_type": "god_protection_x",
			"user_id":     angel.ID,
			"max_x":       maxX,
			"user_ctx":    ctx,
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 触发 [神之庇护]：请选择要抵御的士气下降值（最多%d）", angel.Name, maxX))
	return nil
}

type AngelWallHandler struct{ BaseHandler }

func (h *AngelWallHandler) Execute(ctx *model.Context) error {
	// 天使之墙：PlaceCard逻辑已经在UseSkill中处理（放置FieldCard圣盾）
	// 这里只需要记录日志
	targetName := ctx.Target.Name

	if ctx.User.ID == ctx.Target.ID {
		ctx.Game.Log(fmt.Sprintf("%s 发动 [天使之墙]，自己获得圣盾保护", ctx.User.Name))
	} else {
		ctx.Game.Log(fmt.Sprintf("%s 发动 [天使之墙]，给 %s 提供圣盾保护", ctx.User.Name, targetName))
	}
	return nil
}

// --- Berserker Handlers ---

type BerserkerFrenzyHandler struct{ BaseHandler }

func (h *BerserkerFrenzyHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.TriggerCtx == nil || ctx.TriggerCtx.DamageVal == nil {
		return false
	}
	info := ctx.TriggerCtx.AttackInfo
	if info == nil || info.ActionType != "Attack" {
		return false
	}
	return ctx.Trigger == model.TriggerModifyDamage || ctx.Trigger == model.TriggerOnAttackHit
}

func (h *BerserkerFrenzyHandler) Execute(ctx *model.Context) error {
	bonus := 0
	switch ctx.Trigger {
	case model.TriggerModifyDamage:
		bonus = 1
	case model.TriggerOnAttackHit:
		if len(ctx.User.Hand) > 3 {
			bonus = 1
		}
	default:
		return nil
	}
	if bonus <= 0 {
		return nil
	}
	*ctx.TriggerCtx.DamageVal += bonus
	if ctx.Trigger == model.TriggerModifyDamage {
		ctx.Game.NotifyActionStep(fmt.Sprintf("%s 的被动技【狂化】生效：本次攻击伤害+1", model.GetPlayerDisplayName(ctx.User)))
		ctx.Game.Log(fmt.Sprintf("[Passive] %s 的【狂化】基础效果生效：伤害 +1", ctx.User.Name))
	} else {
		ctx.Game.NotifyActionStep(fmt.Sprintf("攻击命中，%s发动被动技狂化，当前其手牌数%d，伤害额外+1", model.GetPlayerDisplayName(ctx.User), len(ctx.User.Hand)))
		ctx.Game.Log(fmt.Sprintf("[Passive] %s 的【狂化】命中分支生效：手牌 %d，伤害再 +1", ctx.User.Name, len(ctx.User.Hand)))
	}
	return nil
}

type BerserkerTearHandler struct{ BaseHandler }

func (h *BerserkerTearHandler) CanUse(ctx *model.Context) bool {
	if ctx.TriggerCtx == nil || ctx.TriggerCtx.AttackInfo == nil {
		return false
	}
	// 2. [新增] 资源检查：必须至少有 1 颗宝石
	if ctx.User.Gem < 1 {
		return false
	}
	info := ctx.TriggerCtx.AttackInfo
	return info.ActionType == "Attack"
}

func (h *BerserkerTearHandler) Execute(ctx *model.Context) error {
	// 撕裂：攻击命中时发动，覆盖主动攻击与应战攻击。
	if ctx.TriggerCtx != nil && ctx.TriggerCtx.AttackInfo != nil {
		info := ctx.TriggerCtx.AttackInfo
		if info.ActionType == "Attack" {
			if ctx.TriggerCtx.DamageVal != nil {
				ctx.User.Gem -= 1
				*ctx.TriggerCtx.DamageVal += 2
				ctx.Game.NotifyActionStep(fmt.Sprintf("%s花费宝石发动撕裂，此次伤害再额外+2点", model.GetPlayerDisplayName(ctx.User)))
				ctx.Game.Log(fmt.Sprintf("%s 发动 [撕裂]，伤害 +2", ctx.User.Name))
			}
		}
	}
	return nil
}

type BloodRoarHandler struct{ BaseHandler }

func (h *BloodRoarHandler) Execute(ctx *model.Context) error {
	// 血腥咆哮：作为主动攻击打出时发动，若攻击的目标拥有的［治疗］为2，则本次攻击强制命中
	if ctx.TriggerCtx != nil && ctx.TriggerCtx.AttackInfo != nil {
		info := ctx.TriggerCtx.AttackInfo
		// 规则：必须作为主动攻击打出 (非应战反弹)
		if info.ActionType == "Attack" && info.CounterInitiator == "" {
			target := ctx.Target
			if target != nil && target.Heal == 2 {
				info.SetInterceptTag(model.CombatInterceptForceHit)
				info.SetInterceptTag(model.CombatInterceptIgnoreHolyShield)
				ctx.Game.Log(fmt.Sprintf("%s 发动 [血腥咆哮]！目标治疗剂为2，强制命中且无视圣盾", ctx.User.Name))
			}
		}
	}
	return nil
}

type BloodBladeHandler struct{ BaseHandler }

func (h *BloodBladeHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.Target == nil || ctx.Trigger != model.TriggerOnAttackHit || ctx.TriggerCtx == nil || ctx.TriggerCtx.DamageVal == nil {
		return false
	}
	info := ctx.TriggerCtx.AttackInfo
	if info == nil || info.ActionType != "Attack" || info.CounterInitiator != "" || ctx.TriggerCtx.Card == nil {
		return false
	}
	if ctx.User.Character == nil {
		return false
	}
	if !ctx.TriggerCtx.Card.MatchExclusive(ctx.User.Character.ID, "血影狂刀") {
		return false
	}
	handCount := len(ctx.Target.Hand)
	return handCount == 2 || handCount == 3
}

func (h *BloodBladeHandler) Execute(ctx *model.Context) error {
	// 血影狂刀：作为主动攻击打出时发动，根据对手手牌数额外伤害
	if ctx.TriggerCtx != nil && ctx.TriggerCtx.DamageVal != nil {
		target := ctx.Target
		if target != nil {
			extraDamage := 0
			handCount := len(target.Hand)

			if handCount == 2 {
				extraDamage = 2
			} else if handCount == 3 {
				extraDamage = 1
			}

			if extraDamage > 0 {
				*ctx.TriggerCtx.DamageVal += extraDamage
				ctx.Game.Log(fmt.Sprintf("%s 发动 [血影狂刀]！对手手牌%d张，伤害 +%d", ctx.User.Name, handCount, extraDamage))
			}
		}
	}
	return nil
}

// --- Sealer Handlers ---
type MagicSurgeHandler struct{ BaseHandler }

func (h *MagicSurgeHandler) CanUse(ctx *model.Context) bool {
	if ctx.TriggerCtx == nil {
		return false
	}
	// 【修正】只要是法术行动（含法术牌和主动技能），都满足条件
	return ctx.TriggerCtx.ActionType == model.ActionMagic
}

func (h *MagicSurgeHandler) Execute(ctx *model.Context) error {
	// 法术激荡：（［法术行动］结束时发动）额外+1［攻击行动］
	// 向行动队列添加一个无限制的攻击行动令牌
	model.AppendAttackAction(ctx.User, "法术激荡")
	ctx.Game.Log(fmt.Sprintf("%s 发动 [法术激荡]，额外获得1次攻击行动", ctx.User.Name))
	return nil
}

type SealBreakHandler struct{ BaseHandler }

// createBuffCard 根据buff名称创建对应的牌
func createBuffCard(buffName string) *model.Card {
	switch buffName {
	case "Shield":
		return &model.Card{
			ID:          "shield_card",
			Name:        "圣盾",
			Type:        model.CardTypeMagic,
			Element:     model.ElementLight,
			Damage:      0,
			Description: "抵挡一次伤害",
		}
	case "Weak":
		// 虚弱没有对应的牌，创建一个虚拟牌
		return &model.Card{
			ID:          "weak_card",
			Name:        "虚弱",
			Type:        model.CardTypeMagic,
			Element:     model.ElementDark,
			Damage:      0,
			Description: "虚弱状态牌",
		}
	case "Poison":
		// 中毒没有对应的牌，创建一个虚拟牌
		return &model.Card{
			ID:          "poison_card",
			Name:        "中毒",
			Type:        model.CardTypeMagic,
			Element:     model.ElementDark,
			Damage:      0,
			Description: "中毒状态牌",
		}
	default:
		// 默认创建一个通用状态牌
		return &model.Card{
			ID:          "buff_card_" + buffName,
			Name:        buffName,
			Type:        model.CardTypeMagic,
			Element:     model.ElementDark,
			Damage:      0,
			Description: "状态牌",
		}
	}
}

func (h *SealBreakHandler) Execute(ctx *model.Context) error {
	// 封印破碎：收回场上任意一张基础效果牌到自己手里。
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("封印破碎上下文无效")
	}
	var options []basicEffectOption
	if ctx.Target != nil {
		options = collectBasicEffectOptions(ctx.Target)
	} else {
		options = collectBasicEffectOptions(ctx.Game.GetAllPlayers()...)
	}
	if len(options) == 0 {
		return fmt.Errorf("场上没有可收回的基础效果")
	}

	// 若场上有多张基础效果，弹窗让封印师选择具体哪一张。
	if len(options) > 1 {
		ctx.Game.PushInterrupt(&model.Interrupt{
			Type:     model.InterruptChoice,
			PlayerID: ctx.User.ID,
			Context: map[string]interface{}{
				"choice_type":  "basic_effect_pick",
				"user_id":      ctx.User.ID,
				"skill_name":   "封印破碎",
				"operation":    "take",
				"resume_phase": model.TurnStageActionExecution,
				"prompt":       "【封印破碎】请选择要收回的基础效果：",
				"options":      encodeBasicEffectOptions(options),
			},
		})
		ctx.Game.Log(fmt.Sprintf("%s 发动 [封印破碎]，请选择要收回的基础效果", ctx.User.Name))
		return nil
	}

	takenCard, err := ctx.Game.TakeFieldCard(options[0].TargetID, options[0].FieldIndex, ctx.User.ID)
	if err != nil {
		return err
	}
	ctx.User.Hand = append(ctx.User.Hand, takenCard)
	ctx.Game.Log(fmt.Sprintf("%s 的 [封印破碎] 发动，将 %s 收入手中", ctx.User.Name, options[0].Label))
	return nil
}

type FiveElementsBindHandler struct{ BaseHandler }

func (h *FiveElementsBindHandler) CanUse(ctx *model.Context) bool {
	return ctx != nil && ctx.Trigger == model.TriggerNone && ctx.User != nil && ctx.Target != nil
}

func (h *FiveElementsBindHandler) Execute(ctx *model.Context) error {
	if !h.CanUse(ctx) {
		return nil
	}
	ctx.Game.Log(fmt.Sprintf("%s 对 %s 发动五系束缚", ctx.User.Name, ctx.Target.Name))
	return nil
}

// ==========================================
// 五系封印 Handler 设计说明
// ==========================================
// 五系封印的完整流程分为三个阶段：
//
// 阶段 1：放置封印（由技能使用流程处理）
//   - 封印师使用技能（水之封印等）
//   - UseSkill → consumeSkillInputs → placeSkillFieldCard
//   - 场牌被放置到目标玩家面前，Meta中记录绑定元素
//
// 阶段 2：触发封印（由SkillDispatcher处理）
//   - 目标玩家打出/展示对应元素牌
//   - 触发 TriggerOnCardUsed / TriggerOnCardRevealed
//   - collectTriggeredSkills 遍历Field，找到匹配的封印
//   - SealLogic.CanUse → canResolveElementalSealStatus
//
// 阶段 3：结算伤害（由processPendingDamages处理）
//   - SealLogic.Execute → executeElementalSealStatus
//   - 添加PendingDamage，标记EffectTypeToRemove
//   - 伤害结算后移除封印（game.go:5387）
// ==========================================

// SealLogic 五系封印的通用Handler逻辑
// 仅保留放置后的入口映射；
// 实际触发规则统一交给 field status resolver，避免继续耦合主流程。
type SealLogic struct {
	EffectType model.EffectType // 对应的 Effect 枚举，用于移除
}

func (s *SealLogic) CanUse(ctx *model.Context) bool {
	return canResolveFieldStatus(ctx, s.EffectType)
}

func (s *SealLogic) Execute(ctx *model.Context) error {
	if ctx != nil && ctx.Trigger == model.TriggerNone {
		return nil
	}
	return executeFieldStatus(ctx, s.EffectType)
}

// 五系封印具体Handler（放置由 PlaceCard 处理，触发由通用状态 resolver 处理）
type WaterSealHandler struct{ SealLogic }
type FireSealHandler struct{ SealLogic }
type EarthSealHandler struct{ SealLogic }
type WindSealHandler struct{ SealLogic }
type ThunderSealHandler struct{ SealLogic }

func NewWaterSealHandler() *WaterSealHandler {
	return &WaterSealHandler{SealLogic{
		EffectType: model.EffectSealWater,
	}}
}

func NewFireSealHandler() *FireSealHandler {
	return &FireSealHandler{SealLogic{
		EffectType: model.EffectSealFire,
	}}
}

func NewEarthSealHandler() *EarthSealHandler {
	return &EarthSealHandler{SealLogic{
		EffectType: model.EffectSealEarth,
	}}
}

func NewWindSealHandler() *WindSealHandler {
	return &WindSealHandler{SealLogic{
		EffectType: model.EffectSealWind,
	}}
}

func NewThunderSealHandler() *ThunderSealHandler {
	return &ThunderSealHandler{SealLogic{
		EffectType: model.EffectSealThunder,
	}}
}

// --- Blade Master Handlers ---
type WindFuryHandler struct{ BaseHandler }

func (h *WindFuryHandler) CanUse(ctx *model.Context) bool {
	// 1. 基础检查
	if ctx.TriggerCtx == nil {
		return false
	}

	// 2. 检查触发时机
	if ctx.Trigger != model.TriggerOnPhaseEnd {
		return false
	}

	// 3. 【关键】检查刚才结束的行动是不是攻击
	// 这就是我们在 Drive 里填入的 ActionType
	if ctx.TriggerCtx.ActionType != model.ActionAttack {
		return false
	}
	// “攻击行动”仅指主动攻击；应战攻击结束不触发该技能。
	if ctx.TriggerCtx.AttackInfo != nil && ctx.TriggerCtx.AttackInfo.CounterInitiator != "" {
		return false
	}

	// 4. 检查是否已经发动过 (回合限定)
	if ctx.User.TurnState.UsedSkillCounts["wind_fury"] > 0 {
		return false
	}

	return true
}

func (h *WindFuryHandler) Execute(ctx *model.Context) error {
	// 风怒追击：响应技，回合限定一回合只能触发一次，在发攻击行动结束后，可以额外再发动一次攻击行动，其使用的攻击牌必须是风系。
	// 向行动队列添加一个限制为风系攻击的行动令牌
	model.AppendAttackAction(ctx.User, "风怒追击", model.ElementWind)
	ctx.Game.Log(fmt.Sprintf("%s 发动 [风怒追击]，获得一次额外的[风系]攻击行动机会", ctx.User.Name))
	return nil
}

type HolySwordHandler struct{ BaseHandler }

func (h *HolySwordHandler) CanUse(ctx *model.Context) bool {
	// 圣剑：仅在第3次主动攻击宣告时生效。
	if ctx == nil || ctx.User == nil || ctx.Trigger != model.TriggerOnAttackStart || ctx.TriggerCtx == nil || ctx.TriggerCtx.AttackInfo == nil {
		return false
	}
	if ctx.TriggerCtx.AttackInfo.ActionType != string(model.ActionAttack) || ctx.TriggerCtx.AttackInfo.CounterInitiator != "" {
		return false
	}
	return ctx.User.TurnState.AttackCount+1 == 3
}

func (h *HolySwordHandler) Execute(ctx *model.Context) error {
	// 圣剑：强制命中对方无法抵挡
	ctx.Game.Log(fmt.Sprintf("%s 的 [圣剑] 发动，本回合第3次攻击强制命中，对方无法抵挡", ctx.User.Name))
	if ctx.TriggerCtx != nil && ctx.TriggerCtx.AttackInfo != nil {
		ctx.TriggerCtx.AttackInfo.SetInterceptTag(model.CombatInterceptForceHit)
	}
	return nil
}

type SwordShadowHandler struct{ BaseHandler }

func (h *SwordShadowHandler) CanUse(ctx *model.Context) bool {
	// 1. 基础防御性检查
	if ctx.TriggerCtx == nil {
		return false
	}

	// 2. 核心校验：刚才结束的行动必须是“攻击行动”
	// 注意：我们在 Engine 的 PerformMagic 和 PerformAttack 结束时
	// 都在 EventContext 里传入了 ActionType
	if ctx.TriggerCtx.ActionType != model.ActionAttack {
		return false
	}
	// “攻击行动”仅指主动攻击；应战攻击结束不触发该技能。
	if ctx.TriggerCtx.AttackInfo != nil && ctx.TriggerCtx.AttackInfo.CounterInitiator != "" {
		return false
	}

	// 3. 确认是 PhaseEnd 触发
	if ctx.Trigger != model.TriggerOnPhaseEnd {
		return false
	}

	// 4. 【规则】必须可支付1点蓝水晶（可由红宝石替代）
	if !canPayCrystalLike(ctx, 1) {
		return false
	}

	// 5. 检查是否已经发动过（回合限定）
	if ctx.User.TurnState.UsedSkillCounts["sword_shadow"] > 0 {
		return false
	}

	return true
}

func (h *SwordShadowHandler) Execute(ctx *model.Context) error {
	// 剑影：回合限定，攻击结束后消耗1蓝水晶，额外增加一次攻击行动。

	if !spendCrystalLike(ctx, 1) {
		return fmt.Errorf("发动剑影失败：水晶不足（红宝石可替代）")
	}
	ctx.Game.Log(fmt.Sprintf("%s 消耗1蓝水晶（可由红宝石替代）发动 [剑影]", ctx.User.Name))

	model.AppendAttackAction(ctx.User, "剑影")
	ctx.Game.Log(fmt.Sprintf("%s 发动 [剑影]，获得一次额外的攻击行动机会", ctx.User.Name))
	return nil
}

type GaleSkillHandler struct{ BaseHandler }

func (h *GaleSkillHandler) Execute(ctx *model.Context) error {
	// 疾风技：独有技，持有该卡牌并作为主动攻击打出时可触发响应，额外增加一次攻击行动。
	// 向行动队列添加一个无限制的攻击行动令牌
	model.AppendAttackAction(ctx.User, "疾风技")
	ctx.Game.Log(fmt.Sprintf("%s 发动 [疾风技]，额外获得1次攻击行动", ctx.User.Name))
	return nil
}

type GaleSlashHandler struct{ BaseHandler }

func (h *GaleSlashHandler) CanUse(ctx *model.Context) bool {
	// 列风技：目标拥有圣盾时发动
	if ctx == nil || ctx.Trigger != model.TriggerOnAttackStart || ctx.Target == nil || ctx.TriggerCtx == nil || ctx.TriggerCtx.AttackInfo == nil {
		return false
	}
	if ctx.TriggerCtx.AttackInfo.ActionType != string(model.ActionAttack) || ctx.TriggerCtx.AttackInfo.CounterInitiator != "" {
		return false
	}
	hasShield := false
	for _, fc := range ctx.Target.Field {
		if fc.Mode == model.FieldEffect && fc.Effect == model.EffectShield {
			hasShield = true
			break
		}
	}
	return hasShield
}

func (h *GaleSlashHandler) Execute(ctx *model.Context) error {
	// 列风技：无视圣盾效果，被攻击目标无法应战
	ctx.Game.Log(fmt.Sprintf("%s 发动 [列风技]，目标拥有圣盾，无视圣盾效果且目标无法应战", ctx.User.Name))
	if ctx.TriggerCtx != nil && ctx.TriggerCtx.AttackInfo != nil {
		ctx.TriggerCtx.AttackInfo.SetInterceptTag(model.CombatInterceptUnrespondable)
		ctx.TriggerCtx.AttackInfo.SetInterceptTag(model.CombatInterceptIgnoreHolyShield)
	}
	return nil
}

// --- Archer Handlers ---
// internal/engine/skills/handlers_impl.go

func (h *PiercingShotHandler) CanUse(ctx *model.Context) bool {
	// 仅主动攻击未命中可触发；应战攻击未命中不触发。
	if ctx.TriggerCtx == nil || ctx.TriggerCtx.AttackInfo == nil {
		return false
	}
	if ctx.TriggerCtx.AttackInfo.CounterInitiator != "" {
		return false
	}
	// 必须有法术牌才能发动
	for _, card := range ctx.User.Hand {
		if card.Type == model.CardTypeMagic {
			return true
		}
	}
	return false
}

type PiercingShotHandler struct{ BaseHandler }

func (h *PiercingShotHandler) Execute(ctx *model.Context) error {
	// 贯穿射击：弃1张法术牌后，对原目标造成2点法术伤害。
	discardRaw, hasDiscard := ctx.Selections["discard_indices"]
	if !hasDiscard {
		return fmt.Errorf("贯穿射击缺少弃牌选择")
	}
	indices, ok := discardRaw.([]int)
	if !ok || len(indices) != 1 {
		return fmt.Errorf("贯穿射击需要且仅需弃置1张法术牌")
	}
	idx := indices[0]
	if idx < 0 || idx >= len(ctx.User.Hand) {
		return fmt.Errorf("贯穿射击弃牌索引无效: %d", idx)
	}
	card := ctx.User.Hand[idx]
	if card.Type != model.CardTypeMagic {
		return fmt.Errorf("贯穿射击必须弃置法术牌")
	}
	// 文档口径为“弃1张法术牌[展示]”，因此需要走公开展示通知，
	// 以便前端可见并驱动“打出/展示”类被动结算（如元素封印）。
	ctx.Game.NotifyCardRevealed(ctx.User.ID, []model.Card{card}, "discard")
	ctx.User.Hand = append(ctx.User.Hand[:idx], ctx.User.Hand[idx+1:]...)
	ctx.Selections["discardedCards"] = []model.Card{card}

	if ctx.Target != nil {
		ctx.Game.InflictDamage(ctx.User.ID, ctx.Target.ID, 2, "magic")
		ctx.Game.Log(fmt.Sprintf("%s 发动 [贯穿射击]，对 %s 造成2点法术伤害",
			ctx.User.Name, ctx.Target.Name))
	}
	return nil
}

type LightningArrowHandler struct{ BaseHandler }

func (h *LightningArrowHandler) CanUse(ctx *model.Context) bool {
	// 闪电箭：仅在雷系攻击时可用
	if ctx.TriggerCtx != nil && ctx.TriggerCtx.Card != nil {
		return ctx.TriggerCtx.Card.Element == model.ElementThunder
	}
	return false
}

func (h *LightningArrowHandler) Execute(ctx *model.Context) error {
	// 闪电箭：你的雷系攻击对手无法应战
	if ctx.TriggerCtx != nil && ctx.TriggerCtx.AttackInfo != nil {
		// 设置无法应战标志 (CanUse 已验证 ElementThunder)
		ctx.TriggerCtx.AttackInfo.CanBeResponded = false
		ctx.Game.Log(fmt.Sprintf("%s 发动 [闪电箭]，雷系攻击不可被应战", ctx.User.Name))
	}
	return nil
}

type SnipeHandler struct{ BaseHandler }

func (h *SnipeHandler) Execute(ctx *model.Context) error {
	// 狙击：目标角色手牌补到5张[强制]，额外+1攻击行动
	// 规则：若其手牌数大于5则无事发生。若玩家手牌上限小于5，会触发爆牌。
	if ctx.Target != nil {
		currentHand := len(ctx.Target.Hand)
		if currentHand < 5 {
			needCards := 5 - currentHand
			// 强制补牌到5张（不检查手牌上限，让后续的手牌检查逻辑处理爆牌）
			ctx.Game.DrawCards(ctx.Target.ID, needCards)
			ctx.Game.Log(fmt.Sprintf("%s 的 [狙击] 发动，%s 手牌补到5张", ctx.User.Name, ctx.Target.Name))
		} else {
			// 手牌数已经>=5，无事发生
			ctx.Game.Log(fmt.Sprintf("%s 的 [狙击] 发动，但 %s 手牌已有%d张，无事发生", ctx.User.Name, ctx.Target.Name, currentHand))
		}

		// 向行动队列添加一个无限制的攻击行动令牌
		model.AppendAttackAction(ctx.User, "狙击")
		ctx.Game.Log(fmt.Sprintf("%s 发动 [狙击]，额外获得1次攻击行动", ctx.User.Name))
	}
	return nil
}

type PreciseShotHandler struct{ BaseHandler }

func (h *PreciseShotHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.User.Character == nil || ctx.TriggerCtx == nil || ctx.TriggerCtx.AttackInfo == nil || ctx.TriggerCtx.Card == nil {
		return false
	}
	info := ctx.TriggerCtx.AttackInfo
	if info.ActionType != string(model.ActionAttack) || info.CounterInitiator != "" {
		return false
	}
	return ctx.TriggerCtx.Card.MatchExclusive(ctx.User.Character.ID, "精准射击")
}

func (h *PreciseShotHandler) Execute(ctx *model.Context) error {
	// 精准射击：攻击宣告时强制命中，伤害结算时 -1。
	if ctx == nil || ctx.TriggerCtx == nil {
		return nil
	}
	switch ctx.Trigger {
	case model.TriggerOnAttackStart:
		ctx.Game.Log(fmt.Sprintf("%s 发动 [精准射击]，攻击强制命中但伤害-1", ctx.User.Name))
		if ctx.TriggerCtx.AttackInfo != nil {
			ctx.TriggerCtx.AttackInfo.IsHitForced = true
		}
	case model.TriggerModifyDamage:
		if ctx.TriggerCtx.DamageVal != nil {
			*ctx.TriggerCtx.DamageVal -= 1
		}
	default:
		return nil
	}
	return nil
}

type FlashTrapHandler struct{ BaseHandler }

func (h *FlashTrapHandler) Execute(ctx *model.Context) error {
	// 闪光陷阱：对目标造成2点法术伤害
	if ctx.Target != nil {
		ctx.Game.InflictDamage(ctx.User.ID, ctx.Target.ID, 2, "法术")
	}

	// 主动技能使用后，结束当前回合
	ctx.Game.Log(fmt.Sprintf("%s 使用技能后回合结束", ctx.User.Name))

	// 这里需要想办法调用NextTurn，但IGameEngine接口没有NextTurn方法
	// 或者在UseSkill中处理回合结束逻辑

	return nil
}

// --- Assassin Handlers ---
type BacklashHandler struct{ BaseHandler }

func (h *BacklashHandler) CanUse(ctx *model.Context) bool {
	// 仅在“承受攻击伤害”时触发：法术/中毒等非攻击伤害不触发。
	if ctx == nil || ctx.Trigger != model.TriggerOnDamageTaken || ctx.TriggerCtx == nil {
		return false
	}
	if ctx.TriggerCtx.DamageVal == nil || *ctx.TriggerCtx.DamageVal <= 0 {
		return false
	}
	if ctx.Flags["IsMagicDamage"] {
		return false
	}
	if ctx.TriggerCtx.SourceID == "" || ctx.User == nil || ctx.TriggerCtx.SourceID == ctx.User.ID {
		return false
	}
	return true
}

func (h *BacklashHandler) Execute(ctx *model.Context) error {
	// 反噬：强制让攻击者摸1张牌（非伤害，不可被治疗等抵挡）。
	attackerID := ctx.TriggerCtx.SourceID
	attackerName := attackerID
	for _, p := range ctx.Game.GetAllPlayers() {
		if p.ID == attackerID {
			attackerName = model.GetPlayerDisplayName(p)
			break
		}
	}
	ctx.Game.NotifyActionStep(fmt.Sprintf("%s发动被动技反噬，%s强制摸1张牌", model.GetPlayerDisplayName(ctx.User), attackerName))
	ctx.Game.DrawCards(attackerID, 1)
	ctx.Game.Log(fmt.Sprintf("%s 的 [反噬] 发动，%s 强制摸1张牌", ctx.User.Name, attackerID))
	return nil
}

type WaterShadowHandler struct{ BaseHandler }

func (h *WaterShadowHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.TriggerCtx == nil {
		return false
	}
	if ctx.Trigger != model.TriggerBeforeDraw {
		return false
	}
	if ctx.TriggerCtx.TargetID != "" && ctx.TriggerCtx.TargetID != ctx.User.ID {
		return false
	}
	if ctx.TriggerCtx.DrawCount == nil || *ctx.TriggerCtx.DrawCount <= 0 {
		return false
	}
	if ctx.TriggerCtx.ActionType == model.ActionBuy ||
		ctx.TriggerCtx.ActionType == model.ActionSynthesize ||
		ctx.TriggerCtx.ActionType == model.ActionExtract {
		return false
	}
	return ctx.User.HasElement(model.ElementWater)
}

func (h *WaterShadowHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil || ctx.TriggerCtx == nil {
		return fmt.Errorf("水影上下文无效")
	}
	if ctx.Trigger != model.TriggerBeforeDraw {
		return fmt.Errorf("水影只能在摸牌前发动")
	}

	// 获取玩家的弃牌选择
	selection, exists := ctx.Selections["discard_indices"]
	if !exists {
		return fmt.Errorf("没有弃牌选择")
	}

	discardIndices, ok := selection.([]int)
	if !ok {
		return fmt.Errorf("弃牌选择格式错误")
	}

	if len(discardIndices) == 0 {
		return fmt.Errorf("至少需要弃1张牌")
	}

	// 验证牌索引
	player := ctx.User
	usedIndices := make(map[int]bool)
	waterCards := 0
	magicCards := 0

	for _, idx := range discardIndices {
		if idx < 0 || idx >= len(player.Hand) {
			return fmt.Errorf("牌索引越界: %d", idx)
		}
		if usedIndices[idx] {
			return fmt.Errorf("不能重复选择同一张牌: %d", idx)
		}
		usedIndices[idx] = true

		// 统计牌类型
		if player.Hand[idx].Element == model.ElementWater {
			waterCards++
		} else if player.Hand[idx].Type == model.CardTypeMagic {
			magicCards++
		} else {
			return fmt.Errorf("选择的牌既不是水系牌也不是法术牌: %s", player.Hand[idx].Name)
		}
	}

	isStealthed := hasAssassinStealthForm(player)

	// 验证规则
	if waterCards == 0 {
		return fmt.Errorf("至少需要弃1张水系牌")
	}

	if !isStealthed && magicCards > 0 {
		return fmt.Errorf("不在潜行状态下不能弃法术牌")
	}

	if isStealthed && magicCards > 1 {
		return fmt.Errorf("潜行状态下最多只能弃1张法术牌")
	}

	// 执行弃牌
	sort.Sort(sort.Reverse(sort.IntSlice(discardIndices)))

	discardedCards := make([]model.Card, 0, len(discardIndices))
	for _, idx := range discardIndices {
		discardedCards = append(discardedCards, player.Hand[idx])
		player.Hand = append(player.Hand[:idx], player.Hand[idx+1:]...)
	}

	ctx.Game.NotifyCardRevealed(player.ID, discardedCards, "discard")

	// 将弃牌信息存储在Selections中，供外部处理
	ctx.Selections["discardedCards"] = discardedCards
	ctx.Flags["cancelDraw"] = true
	*ctx.TriggerCtx.DrawCount = 0

	// 记录日志
	ctx.Game.Log(fmt.Sprintf("%s 发动 [水影]，展示并弃置了 %d 张水系牌，本次摸牌改为弃牌", player.Name, waterCards))
	if magicCards > 0 {
		ctx.Game.Log(fmt.Sprintf("%s 处于[潜行]，额外展示并弃置了 %d 张法术牌", player.Name, magicCards))
	}

	return nil
}

type StealthHandler struct{ BaseHandler }

func (h *StealthHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil {
		return false
	}
	if ctx.Trigger != model.TriggerOnTurnStart {
		return false
	}
	if ctx.User.Gem < 1 {
		return false
	}
	return !hasAssassinStealthForm(ctx.User)
}

func (h *StealthHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("潜行上下文无效")
	}
	// 消耗宝石
	if ctx.User.Gem < 1 {
		return fmt.Errorf("宝石不足，无法发动潜行")
	}
	if hasAssassinStealthForm(ctx.User) {
		return fmt.Errorf("已处于潜行状态")
	}
	ctx.User.Gem -= 1

	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type":   "assassin_stealth_draw",
			"user_id":       ctx.User.ID,
			"waiting_phase": model.TurnStageActionStart,
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [潜行]，消耗1宝石，等待选择是否摸1张牌后进入潜行状态", ctx.User.Name))
	return nil
}

// --- Saintess Handlers ---

type FrostPrayerHandler struct{ BaseHandler }

func (h *FrostPrayerHandler) CanUse(ctx *model.Context) bool {
	// 触发时机：使用卡牌 或 展示卡牌
	if ctx.Trigger != model.TriggerOnCardUsed && ctx.Trigger != model.TriggerOnCardRevealed {
		return false
	}
	if ctx.TriggerCtx == nil || ctx.TriggerCtx.Card == nil {
		return false
	}
	card := ctx.TriggerCtx.Card
	// 条件：水系牌 或 圣光
	return card.Element == model.ElementWater || card.Name == "圣光"
}

func (h *FrostPrayerHandler) Execute(ctx *model.Context) error {
	// 冰霜祷言：触发后由圣女选择任意目标 +1 治疗
	options := make([]model.PromptOption, 0, len(ctx.Game.GetAllPlayers()))
	for _, p := range ctx.Game.GetAllPlayers() {
		if p == nil {
			continue
		}
		options = append(options, model.PromptOption{
			ID:    p.ID,
			Label: p.Name,
		})
	}
	if len(options) == 0 {
		return nil
	}
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type": "frost_prayer_target",
			"user_id":     ctx.User.ID,
			"target_ids": func() []string {
				ids := make([]string, 0, len(options))
				for _, opt := range options {
					ids = append(ids, opt.ID)
				}
				return ids
			}(),
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 的 [冰霜祷言] 触发，等待选择治疗目标", ctx.User.Name))
	return nil
}

type HealingLightHandler struct{ BaseHandler }

func (h *HealingLightHandler) Execute(ctx *model.Context) error {
	// 治愈之光：指定最多3名角色各+1治疗
	// ctx.Targets 包含选中的目标
	targets := ctx.Targets
	if len(targets) == 0 && ctx.Target != nil {
		targets = []*model.Player{ctx.Target}
	}

	if len(targets) == 0 {
		return fmt.Errorf("需要指定目标")
	}

	for _, t := range targets {
		ctx.Game.Heal(t.ID, 1)
	}
	ctx.Game.Log(fmt.Sprintf("%s 发动 [治愈之光]，%d 名角色各 +1 治疗", ctx.User.Name, len(targets)))
	return nil
}

type HealHandler struct{ BaseHandler }

func (h *HealHandler) Execute(ctx *model.Context) error {
	// 治疗术：目标角色+2治疗
	if ctx.Target == nil {
		return fmt.Errorf("需要指定目标")
	}
	ctx.Game.Heal(ctx.Target.ID, 2)
	ctx.Game.Log(fmt.Sprintf("%s 发动 [治疗术]，%s 获得 +2 治疗", ctx.User.Name, ctx.Target.Name))
	return nil
}

type SaintHealHandler struct{ BaseHandler }

func (h *SaintHealHandler) Execute(ctx *model.Context) error {
	// 圣疗：[水晶] 任意分配3点治疗给1~3名角色，再选择额外+1攻击行动或法术行动。
	targets := ctx.Targets
	if len(targets) == 0 && ctx.Target != nil {
		targets = []*model.Player{ctx.Target}
	}
	if len(targets) == 0 || len(targets) > 3 {
		return fmt.Errorf("圣疗需要指定1-3名目标")
	}

	targetIDs := make([]string, 0, len(targets))
	for _, target := range targets {
		if target == nil {
			continue
		}
		targetIDs = append(targetIDs, target.ID)
	}
	if len(targetIDs) == 0 {
		return fmt.Errorf("圣疗缺少有效目标")
	}

	data := map[string]interface{}{
		"targets": targetIDs,
	}
	if len(targetIDs) == 2 {
		data["stage"] = "allocate_heal"
	} else {
		data["stage"] = "choose_extra_action"
		allocations := map[string]int{}
		if len(targetIDs) == 1 {
			allocations[targetIDs[0]] = 3
		} else {
			for _, targetID := range targetIDs {
				allocations[targetID] = 1
			}
		}
		data["allocations"] = allocations
	}

	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptSaintHeal,
		PlayerID: ctx.User.ID,
		Context:  data,
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [圣疗]，等待分配治疗并选择额外行动类型", ctx.User.Name))
	return nil
}

type MercyHandler struct{ BaseHandler }

func (h *MercyHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil {
		return false
	}
	if ctx.Trigger != model.TriggerOnTurnStart {
		return false
	}
	if ctx.User.Gem < 1 {
		return false
	}
	return !ctx.User.HasFieldEffect(model.EffectMercy)
}

func (h *MercyHandler) Execute(ctx *model.Context) error {
	user := ctx.User
	game := ctx.Game

	if user == nil || game == nil {
		return fmt.Errorf("怜悯上下文无效")
	}
	if user.Gem < 1 {
		return fmt.Errorf("宝石不足，无法发动怜悯")
	}
	if user.HasFieldEffect(model.EffectMercy) {
		return fmt.Errorf("已处于怜悯状态")
	}

	// 怜悯：进入持续状态，横置并使手牌上限恒定为7，同时自己+1水晶。
	user.Gem -= 1
	user.Crystal += 1

	user.AddFieldCard(&model.FieldCard{
		Card: model.Card{
			ID:      fmt.Sprintf("effect-mercy-%s-%d", user.ID, len(user.Field)),
			Name:    "怜悯",
			Type:    model.CardTypeMagic,
			Element: model.ElementLight,
		},
		OwnerID:  user.ID,
		SourceID: user.ID,
		Mode:     model.FieldEffect,
		Effect:   model.EffectMercy,
	})

	game.Log(fmt.Sprintf("%s 发动 [怜悯]：横置并获得1水晶，手牌上限恒定为7", user.Name))
	return nil
}

// --- Magical Girl Handlers ---

type MagicBulletControlHandler struct{ BaseHandler }

func (h *MagicBulletControlHandler) Execute(ctx *model.Context) error {
	// 魔弹掌控由 magic.go/game.go 中的魔弹中断链路直接处理；
	// 这里保留 handler 仅用于维持技能注册表完整。
	return nil
}

type MagicBulletFusionHandler struct{ BaseHandler }

func (h *MagicBulletFusionHandler) Execute(ctx *model.Context) error {
	// 魔弹融合由 PerformMagic 触发的确认中断统一处理；
	// 这里保留 handler 仅用于维持技能注册表完整。
	return nil
}

type MagicBlastHandler struct{ BaseHandler }

func (h *MagicBlastHandler) CanUse(ctx *model.Context) bool {
	// 需要有法术牌可弃才能发动
	for _, card := range ctx.User.Hand {
		if card.Type == model.CardTypeMagic {
			return true
		}
	}
	return false
}

func (h *MagicBlastHandler) Execute(ctx *model.Context) error {
	// 弃牌代价已在 UseSkill 中处理；这里从“我方战绩区 +1 宝石”开始。
	ctx.Game.ModifyGem(string(ctx.User.Camp), 1)
	ctx.Game.Log(fmt.Sprintf("%s 发动 [魔爆冲击]，我方战绩区+1宝石", ctx.User.Name))

	targets := ctx.Targets
	if len(targets) == 0 && ctx.Target != nil {
		targets = []*model.Player{ctx.Target}
	}
	if len(targets) != 2 {
		return fmt.Errorf("魔爆冲击需要且只能指定2名敌方目标")
	}

	targetIDs := make([]string, len(targets))
	for i, t := range targets {
		targetIDs[i] = t.ID
	}

	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptMagicBlast,
		PlayerID: targetIDs[0], // 第一个目标先响应
		Context: map[string]interface{}{
			"choice_type":    "magic_blast",
			"stage":          "target_discard",
			"caster_id":      ctx.User.ID,
			"targets":        targetIDs,
			"current_target": 0,
		},
	})
	ctx.Game.Log(fmt.Sprintf("[Skill] %s 需要选择弃一张法术牌或受到2点伤害", targets[0].Name))

	return nil
}

type DestructionStormHandler struct{ BaseHandler }

func (h *DestructionStormHandler) Execute(ctx *model.Context) error {
	// 毁灭风暴：[宝石] 对任2名目标对手各造成2点法术伤害
	targets := ctx.Targets
	if len(targets) == 0 && ctx.Target != nil {
		targets = []*model.Player{ctx.Target}
	}
	if len(targets) != 2 {
		return fmt.Errorf("毁灭风暴需要且只能指定2名敌方目标")
	}

	for _, t := range targets {
		ctx.Game.InflictDamage(ctx.User.ID, t.ID, 2, "magic")
	}

	ctx.Game.Log(fmt.Sprintf("%s 发动 [毁灭风暴]，对 %d 名目标造成伤害", ctx.User.Name, len(targets)))
	return nil
}
