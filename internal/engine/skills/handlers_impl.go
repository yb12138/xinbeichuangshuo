package skills

import (
	"fmt"
	"sort"
	"starcup-engine/internal/model"
	"strconv"
)

// --- Angel Handlers ---

type PoisonHandler struct{ BaseHandler }

func (h *PoisonHandler) CanUse(ctx *model.Context) bool {
	// 1. 时机：Buff结算阶段
	if ctx.Trigger != model.TriggerOnBuffPhase {
		return false
	}

	// 2. 检查是否有中毒牌
	hasPoison := false
	for _, fc := range ctx.User.Field {
		if fc.Mode == model.FieldEffect && fc.Effect == model.EffectPoison {
			hasPoison = true
			break
		}
	}
	return hasPoison
}

func (h *PoisonHandler) Execute(ctx *model.Context) error {
	user := ctx.User

	// 我们需要把 ctx.Game 转换为 *GameEngine 才能调用 resolveDamage
	// 或者你在 IGameEngine 接口里暴露了 ResolveDamage 方法
	// 这里假设通过接口或者类型断言调用

	// 1. 找到中毒的源头（可能有多个中毒，通常结算第一个或合并结算）
	var poisonCard *model.FieldCard
	for _, fc := range user.Field {
		if fc.Mode == model.FieldEffect && fc.Effect == model.EffectPoison {
			poisonCard = fc
			break
		}
	}

	if poisonCard == nil {
		return nil
	}

	ctx.Game.Log(fmt.Sprintf("[Buff] %s 中毒发作，正在结算伤害...", user.Name))

	damageCard := poisonCard.Card
	damageCard.Damage = 1

	// 2. 发起伤害结算
	// 参数: 来源ID (施加中毒的人), 目标ID (中毒者), 卡牌 (中毒牌), 伤害类型 ("Poison")
	// ResolveDamage 会触发 TriggerOnDamageTaken，从而允许圣盾、技能减免伤害
	err := ctx.Game.ResolveDamage(poisonCard.SourceID, user.ID, &damageCard, "Poison")
	if err != nil {
		return err
	}

	ctx.Game.RemoveFieldCard(user.ID, model.EffectPoison)
	ctx.Game.Log(fmt.Sprintf("[Buff] %s 的中毒效果已结束", user.Name))

	return nil
}

type WeaknessHandler struct{ BaseHandler }

func (h *WeaknessHandler) CanUse(ctx *model.Context) bool {
	// 1. 时机必须是 Buff 结算阶段
	if ctx.Trigger != model.TriggerOnBuffPhase {
		return false
	}
	// 2. 检查玩家场上是否有虚弱牌
	hasWeak := false
	for _, fc := range ctx.User.Field {
		if fc.Mode == model.FieldEffect && fc.Effect == model.EffectWeak {
			hasWeak = true
			break
		}
	}
	return hasWeak
}

func (h *WeaknessHandler) Execute(ctx *model.Context) error {
	user := ctx.User
	// game := ctx.Game.(model.IGameEngine) // Removed assertion, use interface directly

	// 抛出一个通用的选择中断
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice, // 【新增】使用选择类型
		PlayerID: user.ID,
		Context: map[string]interface{}{
			"choice_type": "weak", // 标记这是虚弱的选择
			// 可以在这里存更多上下文，比如虚弱牌的 ID
		},
	})

	ctx.Game.Log(fmt.Sprintf("[Buff] %s 触发虚弱判定，等待玩家选择...", user.Name))
	return nil
}

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

	// 3. 检查玩家场上是否真的有【圣盾】效果牌
	// (Dispatcher 遍历 Field 时会传入 User，这里做双重保险)
	// 烈风技：本次攻击无视圣盾
	if ctx.Target != nil && ctx.Target.TurnState.GaleSlashActive {
		return false
	}
	// 血腥咆哮：本次攻击无视圣盾
	if ctx.Target != nil && ctx.Target.Tokens != nil && ctx.Target.Tokens["berserker_blood_roar_ignore_shield"] > 0 {
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

func (h *AngelBondHandler) CanUse(ctx *model.Context) bool {
	// 场景 A: 移除基础效果
	if ctx.Trigger == model.TriggerOnBuffRemoved {
		if ctx.TriggerCtx == nil || ctx.User == nil {
			return false
		}
		// 只在“天使本人移除基础效果”时触发。
		if ctx.TriggerCtx.SourceID != ctx.User.ID {
			return false
		}
		return model.IsBasicEffect(ctx.TriggerCtx.BuffID)
	}

	// 场景 B: 使用圣盾 或 天使之墙
	if ctx.Trigger == model.TriggerOnCardUsed {
		if ctx.TriggerCtx == nil || ctx.TriggerCtx.Card == nil {
			return false
		}
		name := ctx.TriggerCtx.Card.Name
		// 【修正】兼容天使之墙
		return name == "圣盾" || name == "天使之墙"
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
	// 风之洁净：弃一张风系牌，移除场上任意一个基础效果
	if ctx.Target != nil {
		target := ctx.Target
		var effectsToRemove []string
		// 检查 FieldCards
		for _, fc := range target.Field {
			if fc.Mode == model.FieldEffect {
				if model.IsBasicEffect(string(fc.Effect)) {
					effectsToRemove = append(effectsToRemove, string(fc.Effect))
				}
			}
		}

		if len(effectsToRemove) == 0 {
			ctx.Game.Log(fmt.Sprintf("%s 的 [风之洁净] 发动，但 %s 没有基础效果可移除", ctx.User.Name, ctx.Target.Name))
			return nil
		}

		// 移除逻辑
		removeEffect := func(name string) {
			ctx.Game.RemoveFieldCardBy(target.ID, model.EffectType(name), ctx.User.ID)
			ctx.Game.Log(fmt.Sprintf("%s 的 [风之洁净] 发动，移除了 %s 的 %s", ctx.User.Name, ctx.Target.Name, name))
		}

		// 1. 优先使用 Args
		if len(ctx.Args) > 0 && ctx.Args[0] != "" {
			name := ctx.Args[0]
			for _, e := range effectsToRemove {
				if e == name {
					removeEffect(name)
					return nil
				}
			}
		}

		// 2. 默认移除第一个
		removeEffect(effectsToRemove[0])
	}
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
	effectLabel := func(effect model.EffectType) string {
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
	var options []map[string]interface{}
	for _, p := range ctx.Game.GetAllPlayers() {
		if p == nil {
			continue
		}
		for _, fc := range p.Field {
			if fc.Mode != model.FieldEffect || !model.IsBasicEffect(string(fc.Effect)) {
				continue
			}
			optID := fmt.Sprintf("%s|%s", p.ID, string(fc.Effect))
			options = append(options, map[string]interface{}{
				"id":        optID,
				"label":     fmt.Sprintf("%s：%s", p.Name, effectLabel(fc.Effect)),
				"target_id": p.ID,
				"effect":    string(fc.Effect),
			})
		}
	}
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
			"choice_type": "angel_song_pick",
			"user_id":     ctx.User.ID,
			"options":     options,
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
	angel := ctx.User
	loss := *ctx.TriggerCtx.DamageVal

	// 计算可以抵御多少 (1水晶抵御1点；红宝石可替代水晶)
	usable := ctx.Game.GetUsableCrystal(angel.ID)
	mitigate := loss
	if mitigate > usable {
		mitigate = usable
	}

	if mitigate <= 0 {
		return nil
	}
	if !ctx.Game.ConsumeCrystalCost(angel.ID, mitigate) {
		return fmt.Errorf("神之庇护结算失败：可用水晶不足")
	}

	// 减少士气损失
	*ctx.TriggerCtx.DamageVal -= mitigate

	ctx.Game.Log(fmt.Sprintf("%s 发动 [神之庇护]，消耗 %d 水晶（可由红宝石替代）抵御了 %d 点士气下降！", angel.Name, mitigate, mitigate))
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

// BerserkerFrenzyHandler removed - passive skills are handled directly in game logic

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
	// 规则：必须作为主动攻击打出 (非应战反弹)
	can := info.ActionType == "Attack" && info.CounterInitiator == ""
	return can
}

func (h *BerserkerTearHandler) Execute(ctx *model.Context) error {
	// 撕裂：作为主动攻击打出时发动，若攻击命中时②，消耗1宝石，使本次攻击伤害额外+2
	if ctx.TriggerCtx != nil && ctx.TriggerCtx.AttackInfo != nil {
		info := ctx.TriggerCtx.AttackInfo
		// 规则：必须作为主动攻击打出 (非应战反弹)
		if info.ActionType == "Attack" && info.CounterInitiator == "" {
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
				info.IsHitForced = true
				info.CanBeResponded = false
				if ctx.User.Tokens == nil {
					ctx.User.Tokens = map[string]int{}
				}
				// 血腥咆哮强制命中本次攻击同时无视圣盾（仅本次攻击）。
				ctx.User.Tokens["berserker_blood_roar_ignore_shield"] = 1
				ctx.Game.Log(fmt.Sprintf("%s 发动 [血腥咆哮]！目标治疗剂为2，强制命中且无视圣盾", ctx.User.Name))
			}
		}
	}
	return nil
}

type BloodBladeHandler struct{ BaseHandler }

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
	token := model.ActionContext{
		Source:      "法术激荡",
		MustElement: nil,      // 无属性限制
		MustType:    "Attack", // 必须是攻击行动
	}
	ctx.User.TurnState.PendingActions = append(ctx.User.TurnState.PendingActions, token)
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
	// 封印破碎：收回“目标角色”面前的一张基础效果牌到自己手里。
	if ctx == nil || ctx.User == nil || ctx.Target == nil {
		return fmt.Errorf("封印破碎缺少目标")
	}

	type effectOption struct {
		TargetID    string
		TargetName  string
		FieldIndex  int
		Effect      model.EffectType
		DisplayName string
	}
	effectLabel := func(effect model.EffectType) string {
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

	target := ctx.Target
	options := make([]effectOption, 0)
	for idx, fc := range target.Field {
		if fc == nil || fc.Mode != model.FieldEffect {
			continue
		}
		if !model.IsBasicEffect(string(fc.Effect)) {
			continue
		}
		options = append(options, effectOption{
			TargetID:    target.ID,
			TargetName:  target.Name,
			FieldIndex:  idx,
			Effect:      fc.Effect,
			DisplayName: effectLabel(fc.Effect),
		})
	}
	if len(options) == 0 {
		return fmt.Errorf("%s 面前没有可收回的基础效果", target.Name)
	}

	// 若目标身上有多张基础效果，弹窗让封印师选择具体哪一张。
	if len(options) > 1 {
		pickOptions := make([]map[string]interface{}, 0, len(options))
		for _, op := range options {
			pickOptions = append(pickOptions, map[string]interface{}{
				"target_id":    op.TargetID,
				"field_index":  op.FieldIndex,
				"effect":       string(op.Effect),
				"display_name": op.DisplayName,
				"label":        fmt.Sprintf("%s：%s", op.TargetName, op.DisplayName),
			})
		}
		ctx.Game.PushInterrupt(&model.Interrupt{
			Type:     model.InterruptChoice,
			PlayerID: ctx.User.ID,
			Context: map[string]interface{}{
				"choice_type": "seal_break_pick_effect",
				"user_id":     ctx.User.ID,
				"options":     pickOptions,
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
	ctx.Game.Log(fmt.Sprintf("%s 的 [封印破碎] 发动，将 %s 的 %s 收入手中", ctx.User.Name, options[0].TargetName, options[0].DisplayName))
	return nil
}

type FiveElementsBindHandler struct{ BaseHandler }

func (h *FiveElementsBindHandler) Execute(ctx *model.Context) error {
	// 五系束缚：［水晶］将五系束缚放置于目标对手面前
	// 放置逻辑由 UseSkill 的 PlaceCard 通用处理
	// 效果触发在目标回合开始时由 applyFiveElementsBindEffect 处理
	// 目标选择：摸(2+X)张牌(X=场上封印数，最多2)，或者放弃行动移除此牌
	if ctx.Target != nil {
		ctx.Game.Log(fmt.Sprintf("%s 对 %s 发动五系束缚", ctx.User.Name, ctx.Target.Name))
	}
	return nil
}

// SealLogic 是所有封印技能共用的核心逻辑
// 放置阶段由 UseSkill 的 PlaceCard 通用逻辑处理，触发阶段由此 Execute 处理
type SealLogic struct {
	TargetElement model.Element    // 该封印针对的属性
	EffectName    string           // 封印名称（用于日志）
	EffectType    model.EffectType // 对应的 Effect 枚举，用于移除
}

func (s *SealLogic) CanUse(ctx *model.Context) bool {
	// 1. 触发时机必须是"使用卡牌"或"展示卡牌"
	// 规则：当玩家展示/使用对应系的卡牌时，触发效果
	if ctx.Trigger != model.TriggerOnCardUsed && ctx.Trigger != model.TriggerOnCardRevealed {
		return false
	}

	// 2. 检查使用/展示的卡牌是否匹配封印属性
	if ctx.TriggerCtx == nil || ctx.TriggerCtx.Card == nil {
		return false
	}

	// 封印规则：只要是该系牌（攻击或法术）都触发
	return ctx.TriggerCtx.Card.Element == s.TargetElement
}

func (s *SealLogic) Execute(ctx *model.Context) error {
	// 放置阶段由 UseSkill 的 PlaceCard 处理，此处仅处理触发阶段
	// 规则：当玩家展示/使用对应系的卡牌时，触发效果
	if ctx.Trigger != model.TriggerOnCardUsed && ctx.Trigger != model.TriggerOnCardRevealed {
		return nil
	}

	user := ctx.User // 触发封印的人（被封印的玩家）

	ctx.Game.Log(fmt.Sprintf("[Seal] %s 使用了 %s 系牌，触发了 %s！",
		user.Name, s.TargetElement, s.EffectName))

	// 1. 找到封印牌以获取施放者 SourceID
	var sourceID string

	for _, fc := range user.Field {
		if fc.Mode == model.FieldEffect && fc.Effect == s.EffectType {
			sourceID = fc.SourceID
			break
		}
	}

	if sourceID == "" {
		sourceID = user.ID // 兜底
	}

	// 将伤害作为延迟效果推入队列
	ctx.Game.AddPendingDamage(model.PendingDamage{
		SourceID:           sourceID,
		TargetID:           user.ID,
		Damage:             3,
		DamageType:         "magic",
		EffectTypeToRemove: s.EffectType, // 伤害结算后需要移除的封印
	})

	ctx.Game.Log(fmt.Sprintf("[Seal] 封印伤害已推入延迟队列，Source: %s, Target: %s, Damage: 3, EffectToRemove: %s", sourceID, user.ID, s.EffectType))

	return nil
}

// 五系封印 Handler（放置由 PlaceCard 处理，触发由 SealLogic 处理）
type WaterSealHandler struct{ SealLogic }
type FireSealHandler struct{ SealLogic }
type EarthSealHandler struct{ SealLogic }
type WindSealHandler struct{ SealLogic }
type ThunderSealHandler struct{ SealLogic }

func NewWaterSealHandler() *WaterSealHandler {
	return &WaterSealHandler{SealLogic{
		TargetElement: model.ElementWater,
		EffectName:    "水之封印",
		EffectType:    model.EffectSealWater,
	}}
}

func NewFireSealHandler() *FireSealHandler {
	return &FireSealHandler{SealLogic{
		TargetElement: model.ElementFire,
		EffectName:    "火之封印",
		EffectType:    model.EffectSealFire,
	}}
}

func NewEarthSealHandler() *EarthSealHandler {
	return &EarthSealHandler{SealLogic{
		TargetElement: model.ElementEarth,
		EffectName:    "地之封印",
		EffectType:    model.EffectSealEarth,
	}}
}

func NewWindSealHandler() *WindSealHandler {
	return &WindSealHandler{SealLogic{
		TargetElement: model.ElementWind,
		EffectName:    "风之封印",
		EffectType:    model.EffectSealWind,
	}}
}

func NewThunderSealHandler() *ThunderSealHandler {
	return &ThunderSealHandler{SealLogic{
		TargetElement: model.ElementThunder,
		EffectName:    "雷之封印",
		EffectType:    model.EffectSealThunder,
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

	if len(ctx.User.Hand) == 0 {
		return false
	}

	hasWindAttackCard := false
	for _, card := range ctx.User.Hand {
		if card.Type == model.CardTypeAttack && card.Element == model.ElementWind {
			hasWindAttackCard = true
			break
		}
	}

	if !hasWindAttackCard {
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
	token := model.ActionContext{
		Source:      "风怒追击",
		MustElement: []model.Element{model.ElementWind}, // 必须是风系牌
		MustType:    "Attack",                           // 必须是攻击行动
	}
	ctx.User.TurnState.PendingActions = append(ctx.User.TurnState.PendingActions, token)
	ctx.Game.Log(fmt.Sprintf("%s 发动 [风怒追击]，获得一次额外的[风系]攻击行动机会", ctx.User.Name))
	return nil
}

type HolySwordHandler struct{ BaseHandler }

func (h *HolySwordHandler) CanUse(ctx *model.Context) bool {
	// 圣剑：仅在第3次攻击时可用
	return ctx.User != nil && ctx.User.TurnState.AttackCount+1 == 3
}

func (h *HolySwordHandler) Execute(ctx *model.Context) error {
	// 圣剑：强制命中对方无法抵挡
	ctx.Game.Log(fmt.Sprintf("%s 的 [圣剑] 发动，本回合第3次攻击强制命中，对方无法抵挡", ctx.User.Name))
	if ctx.TriggerCtx != nil && ctx.TriggerCtx.AttackInfo != nil {
		ctx.TriggerCtx.AttackInfo.IsHitForced = true
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

	// 构造行动令牌
	token := model.ActionContext{
		Source:      "剑影",
		MustElement: nil,      // 剑影不限制额外攻击的属性
		MustType:    "Attack", // 强制要求下一次行动必须是攻击
	}

	// 添加到待执行队列
	ctx.User.TurnState.PendingActions = append(ctx.User.TurnState.PendingActions, token)

	ctx.Game.Log(fmt.Sprintf("%s 发动 [剑影]，获得一次额外的攻击行动机会", ctx.User.Name))
	return nil
}

type GaleSkillHandler struct{ BaseHandler }

func (h *GaleSkillHandler) Execute(ctx *model.Context) error {
	// 疾风技：独有技，持有该卡牌并作为主动攻击打出时可触发响应，额外增加一次攻击行动。
	// 向行动队列添加一个无限制的攻击行动令牌
	token := model.ActionContext{
		Source:      "疾风技",
		MustElement: nil,      // 无属性限制
		MustType:    "Attack", // 必须是攻击行动
	}
	ctx.User.TurnState.PendingActions = append(ctx.User.TurnState.PendingActions, token)
	ctx.Game.Log(fmt.Sprintf("%s 发动 [疾风技]，额外获得1次攻击行动", ctx.User.Name))
	return nil
}

type GaleSlashHandler struct{ BaseHandler }

func (h *GaleSlashHandler) CanUse(ctx *model.Context) bool {
	// 烈风技：目标拥有圣盾时发动
	if ctx.Target == nil {
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
	// 烈风技：无视圣盾效果，被攻击目标无法应战
	ctx.Game.Log(fmt.Sprintf("%s 发动 [烈风技]，目标拥有圣盾，无视圣盾效果且目标无法应战", ctx.User.Name))
	// 设置标记，表示这次攻击发动了烈风技
	ctx.User.TurnState.GaleSlashActive = true
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
		token := model.ActionContext{
			Source:      "狙击",
			MustElement: nil,      // 无属性限制
			MustType:    "Attack", // 必须是攻击行动
		}
		ctx.User.TurnState.PendingActions = append(ctx.User.TurnState.PendingActions, token)
		ctx.Game.Log(fmt.Sprintf("%s 发动 [狙击]，额外获得1次攻击行动", ctx.User.Name))
	}
	return nil
}

type PreciseShotHandler struct{ BaseHandler }

func (h *PreciseShotHandler) Execute(ctx *model.Context) error {
	// 精准射击：此攻击强制命中，但本次攻击伤害-1
	ctx.Game.Log(fmt.Sprintf("%s 发动 [精准射击]，攻击强制命中但伤害-1", ctx.User.Name))
	if ctx.TriggerCtx != nil && ctx.TriggerCtx.AttackInfo != nil {
		ctx.TriggerCtx.AttackInfo.IsHitForced = true
		ctx.TriggerCtx.AttackInfo.CanBeResponded = false
	}
	// 设置标记，表示这次攻击强制命中
	ctx.User.TurnState.PreciseShotActive = true
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
	// 检查是否有水系牌
	return ctx.User.HasElement(model.ElementWater)
}

func (h *WaterShadowHandler) Execute(ctx *model.Context) error {
	// 水影：弃X张水系牌，潜行状态下可额外弃法术牌，避免爆牌

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

	// 检查潜行状态
	isStealthed := false
	for _, fc := range player.Field {
		if fc.Mode == model.FieldEffect && fc.Effect == model.EffectStealth {
			isStealthed = true
			break
		}
	}

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

	// 将弃牌信息存储在Selections中，供外部处理
	ctx.Selections["discardedCards"] = discardedCards

	// 记录日志
	ctx.Game.Log(fmt.Sprintf("%s 发动 [水影]，弃置了 %d 张水系牌", player.Name, waterCards))
	if magicCards > 0 {
		ctx.Game.Log(fmt.Sprintf("%s 额外弃置了 %d 张法术牌", player.Name, magicCards))
	}

	return nil
}

type StealthHandler struct{ BaseHandler }

func (h *StealthHandler) CanUse(ctx *model.Context) bool {
	// 检查是否有足够的宝石
	if ctx.User == nil {
		return false
	}
	return ctx.User.Gem >= 1 // 需要1个宝石
}

func (h *StealthHandler) Execute(ctx *model.Context) error {
	// 消耗宝石
	if ctx.User.Gem < 1 {
		return fmt.Errorf("宝石不足，无法发动潜行")
	}
	ctx.User.Gem -= 1

	// 摸1张牌
	ctx.Game.DrawCards(ctx.User.ID, 1)

	// 进入潜行状态（场上效果），供后续技能和UI判定
	if !ctx.User.HasFieldEffect(model.EffectStealth) {
		ctx.User.AddFieldCard(&model.FieldCard{
			Card: model.Card{
				ID:   fmt.Sprintf("effect-stealth-%s-%d", ctx.User.ID, len(ctx.User.Field)),
				Name: "潜行",
				Type: model.CardTypeMagic,
			},
			OwnerID:  ctx.User.ID,
			SourceID: ctx.User.ID,
			Mode:     model.FieldEffect,
			Effect:   model.EffectStealth,
			Trigger:  model.EffectTriggerManual,
			Duration: -1,
		})
	}

	ctx.Game.Log(fmt.Sprintf("%s 发动 [潜行]，消耗1宝石，摸1张牌并进入潜行状态", ctx.User.Name))
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
	// 圣疗：[水晶] 任意分配3点治疗给1~3名角色，额外+1攻击行动
	// 资源扣除由 UseSkill 统一处理，这里不重复扣费。

	targets := ctx.Targets
	if len(targets) == 0 && ctx.Target != nil {
		targets = []*model.Player{ctx.Target}
	}
	if len(targets) == 0 {
		return fmt.Errorf("需要指定目标")
	}

	// 检查是否传入了分配参数（Args）
	// Args 格式: ["目标1治疗点数", "目标2治疗点数", ...]
	if ctx.Args != nil && len(ctx.Args) > 0 {
		// 自定义分配模式
		totalHeal := 0
		for i, t := range targets {
			if i < len(ctx.Args) {
				healAmount, err := strconv.Atoi(ctx.Args[i])
				if err == nil && healAmount > 0 {
					ctx.Game.Heal(t.ID, healAmount)
					ctx.Game.Log(fmt.Sprintf("[Skill] %s 获得 %d 点治疗", t.Name, healAmount))
					totalHeal += healAmount
				}
			}
		}
		if totalHeal != 3 {
			ctx.Game.Log(fmt.Sprintf("[Warning] 圣疗治疗分配总计 %d 点，应为3点", totalHeal))
		}
	} else {
		// 默认分配逻辑：
		// 1个目标: +3
		// 2个目标: +2, +1
		// 3个目标: +1, +1, +1
		points := 3
		for i, t := range targets {
			if points <= 0 {
				break
			}
			healAmount := 1
			if len(targets) == 1 {
				healAmount = 3
			} else if len(targets) == 2 && i == 0 {
				healAmount = 2
			}

			ctx.Game.Heal(t.ID, healAmount)
			ctx.Game.Log(fmt.Sprintf("[Skill] %s 获得 %d 点治疗", t.Name, healAmount))
			points -= healAmount
		}
	}

	// 额外攻击行动
	token := model.ActionContext{
		Source:   "圣疗",
		MustType: "Attack",
	}
	ctx.User.TurnState.PendingActions = append(ctx.User.TurnState.PendingActions, token)
	ctx.Game.Log(fmt.Sprintf("%s 发动 [圣疗]，分配治疗并获得额外攻击行动", ctx.User.Name))
	return nil
}

type MercyHandler struct{ BaseHandler }

func (h *MercyHandler) Execute(ctx *model.Context) error {
	user := ctx.User
	game := ctx.Game

	// 怜悯：持续状态，宝石，水晶+1，手牌上限恒定为7
	// 消耗宝石
	user.Gem -= 1

	// 给己方阵营 +1 水晶
	camp := user.Camp
	game.ModifyCrystal(string(camp), 1)

	// 修改最大手牌 (需要引擎支持动态修改或 flag)
	user.MaxHand = 7
	// 设置Flag防止重置? 或 TurnStart 自动重置?
	// 描述 "恒定为7".

	game.Log(fmt.Sprintf("%s 的怜悯发动，获得1水晶，手牌上限恒定为7", user.Name))
	return nil
}

// --- Magical Girl Handlers ---

type MagicBulletControlHandler struct{ BaseHandler }

func (h *MagicBulletControlHandler) Execute(ctx *model.Context) error {
	// 魔弹掌控：主动使用魔弹时可以选择逆向传递
	// 设置标记，在 PerformMagic 逻辑中检查
	// 或者如果此时魔弹链已经启动，尝试修改方向?
	// TriggerOnAttackStart (Using Magic Bullet)
	// 假设引擎检查 TurnState.MagicBulletReverse
	// 这里我们需要在 PlayerTurnState 中加个字段? 或者 Context Flag?

	// 暂且打印日志，实际方向控制需要在 PerformMagic 中实现
	ctx.Game.Log(fmt.Sprintf("%s 发动 [魔弹掌控]，魔弹将逆向传递", ctx.User.Name))
	// TODO: Implement direction reversal in Engine
	return nil
}

type MagicBulletFusionHandler struct{ BaseHandler }

func (h *MagicBulletFusionHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil {
		return fmt.Errorf("上下文无效")
	}
	var fusionCard *model.Card
	if ctx.Selections != nil {
		if cards, ok := ctx.Selections["discardedCards"].([]model.Card); ok && len(cards) > 0 {
			c := cards[0]
			fusionCard = &c
		}
	}
	if fusionCard == nil {
		return fmt.Errorf("魔弹融合缺少弃牌信息")
	}
	// 视为发动魔弹，并沿用“魔弹掌控”方向选择流程。
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptMagicBulletDirection,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"source_id":   ctx.User.ID,
			"is_fusion":   true,
			"fusion_card": *fusionCard,
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [魔弹融合]，弃置 %s 并视为发动【魔弹】", ctx.User.Name, fusionCard.Name))
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
	// 魔爆冲击复杂逻辑：
	// 1. 弃一张地系牌（已在 UseSkill 弃牌检查中处理）
	// 2. 战绩区+1红宝石
	// 3. 选择两个目标对手，他们各需弃一张法术牌
	// 4. 未弃者受2点伤害，同时魔法少女弃一张牌

	// 弃牌已处理，战绩区+1宝石
	ctx.Game.ModifyGem(string(ctx.User.Camp), 1)
	ctx.Game.Log(fmt.Sprintf("%s 发动 [魔爆冲击]，弃地系牌后战绩区+1宝石", ctx.User.Name))

	// 获取两个目标
	targets := ctx.Targets
	if len(targets) == 0 && ctx.Target != nil {
		targets = []*model.Player{ctx.Target}
	}

	if len(targets) == 0 {
		// 如果没有选择目标，技能效果结束
		ctx.Game.Log("[Skill] 魔爆冲击：未选择目标")
		return nil
	}

	// 限制最多2个目标
	if len(targets) > 2 {
		targets = targets[:2]
	}

	// 推送魔爆冲击中断，让目标玩家选择弃法术牌
	targetIDs := make([]string, len(targets))
	for i, t := range targets {
		targetIDs[i] = t.ID
	}

	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptMagicBlast,
		PlayerID: targetIDs[0], // 第一个目标先响应
		Context: map[string]interface{}{
			"choice_type":    "magic_blast",
			"caster_id":      ctx.User.ID,
			"targets":        targetIDs,
			"current_target": 0,
			"failed_count":   0, // 未弃牌的目标数
		},
	})
	ctx.Game.Log(fmt.Sprintf("[Skill] %s 需要选择弃一张法术牌或受到2点伤害", targets[0].Name))

	return nil
}

type DestructionStormHandler struct{ BaseHandler }

func (h *DestructionStormHandler) Execute(ctx *model.Context) error {
	// 毁灭风暴：[宝石] 对任2名目标对手各造成2点法术伤害
	ctx.User.Gem -= 1

	targets := ctx.Targets
	if len(targets) == 0 && ctx.Target != nil {
		targets = []*model.Player{ctx.Target}
	}

	if len(targets) == 0 {
		return fmt.Errorf("需要指定目标")
	}

	for _, t := range targets {
		ctx.Game.InflictDamage(ctx.User.ID, t.ID, 2, "magic")
	}

	ctx.Game.Log(fmt.Sprintf("%s 发动 [毁灭风暴]，对 %d 名目标造成伤害", ctx.User.Name, len(targets)))
	return nil
}
