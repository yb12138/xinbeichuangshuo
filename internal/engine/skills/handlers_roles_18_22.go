package skills

import (
	"fmt"
	"starcup-engine/internal/model"
)

// --- 18. 祈祷师 ---

type PrayerEnterFormHandler struct{ BaseHandler }

type PrayerRuneGainHandler struct{ BaseHandler }

type PrayerRadiantFaithHandler struct{ BaseHandler }

type PrayerDarkCurseHandler struct{ BaseHandler }

type PrayerPowerBlessingHandler struct{ BaseHandler }

type PrayerSwiftBlessingHandler struct{ BaseHandler }

type PrayerManaTideHandler struct{ BaseHandler }

func (h *PrayerEnterFormHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil {
		return false
	}
	return ctx.User.Gem > 0 && getToken(ctx.User, "prayer_form") == 0
}

func (h *PrayerEnterFormHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil {
		return fmt.Errorf("上下文无效")
	}
	if ctx.User.Gem <= 0 {
		return fmt.Errorf("祈祷需要至少1个红宝石")
	}
	if getToken(ctx.User, "prayer_form") > 0 {
		return nil
	}
	ctx.User.Gem--
	setToken(ctx.User, "prayer_form", 1)
	ctx.Game.Log(fmt.Sprintf("%s 发动 [祈祷]，进入祈祷形态", ctx.User.Name))
	return nil
}

func (h *PrayerRuneGainHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.TriggerCtx == nil {
		return false
	}
	if getToken(ctx.User, "prayer_form") == 0 {
		return false
	}
	if ctx.Trigger != model.TriggerOnAttackStart {
		return false
	}
	// 仅主动攻击
	if ctx.TriggerCtx.AttackInfo != nil && ctx.TriggerCtx.AttackInfo.CounterInitiator != "" {
		return false
	}
	return true
}

func (h *PrayerRuneGainHandler) Execute(ctx *model.Context) error {
	v := addToken(ctx.User, "prayer_rune", 2, 0, 3)
	ctx.Game.Log(fmt.Sprintf("%s 的 [祈祷符文] 触发，祈祷符文=%d", ctx.User.Name, v))
	return nil
}

func (h *PrayerRadiantFaithHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil {
		return false
	}
	return getToken(ctx.User, "prayer_form") > 0 && getToken(ctx.User, "prayer_rune") > 0
}

func (h *PrayerRadiantFaithHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("上下文无效")
	}
	if getToken(ctx.User, "prayer_form") == 0 {
		return fmt.Errorf("不在祈祷形态，无法发动光辉信仰")
	}
	if getToken(ctx.User, "prayer_rune") <= 0 {
		return fmt.Errorf("祈祷符文不足")
	}
	addToken(ctx.User, "prayer_rune", -1, 0, 3)
	ctx.Game.ModifyGem(string(ctx.User.Camp), 1)
	target := ctx.Target
	if target == nil || target.Camp != ctx.User.Camp {
		target = firstAllySelf(ctx.Game.GetAllPlayers(), ctx.User.Camp)
	}
	if target != nil {
		ctx.Game.Heal(target.ID, 1)
		ctx.Game.Log(fmt.Sprintf("%s 发动 [光辉信仰]，移除1祈祷符文，战绩区+1红宝石，并治疗 %s 1点", ctx.User.Name, target.Name))
	} else {
		ctx.Game.Log(fmt.Sprintf("%s 发动 [光辉信仰]，移除1祈祷符文，战绩区+1红宝石", ctx.User.Name))
	}
	return nil
}

func (h *PrayerDarkCurseHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil {
		return false
	}
	return getToken(ctx.User, "prayer_form") > 0 && getToken(ctx.User, "prayer_rune") > 0
}

func (h *PrayerDarkCurseHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Target == nil || ctx.Game == nil {
		return fmt.Errorf("黑暗诅咒需要目标")
	}
	if getToken(ctx.User, "prayer_form") == 0 {
		return fmt.Errorf("不在祈祷形态，无法发动黑暗诅咒")
	}
	if getToken(ctx.User, "prayer_rune") <= 0 {
		return fmt.Errorf("祈祷符文不足")
	}
	addToken(ctx.User, "prayer_rune", -1, 0, 3)
	// 先结算对方，再结算自己
	ctx.Game.AddPendingDamage(model.PendingDamage{
		SourceID:   ctx.User.ID,
		TargetID:   ctx.Target.ID,
		Damage:     2,
		DamageType: "magic",
		Stage:      0,
	})
	ctx.Game.AddPendingDamage(model.PendingDamage{
		SourceID:   ctx.User.ID,
		TargetID:   ctx.User.ID,
		Damage:     2,
		DamageType: "magic",
		Stage:      0,
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [黑暗诅咒]，先对 %s 再对自己各造成2点法术伤害", ctx.User.Name, ctx.Target.Name))
	return nil
}

func (h *PrayerPowerBlessingHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Target == nil {
		return fmt.Errorf("威力赐福需要目标")
	}
	ctx.Game.Log(fmt.Sprintf("%s 发动 [威力赐福]，在 %s 面前放置威力赐福", ctx.User.Name, ctx.Target.Name))
	return nil
}

func (h *PrayerSwiftBlessingHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Target == nil {
		return fmt.Errorf("迅捷赐福需要目标")
	}
	ctx.Game.Log(fmt.Sprintf("%s 发动 [迅捷赐福]，在 %s 面前放置迅捷赐福", ctx.User.Name, ctx.Target.Name))
	return nil
}

func (h *PrayerManaTideHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.TriggerCtx == nil {
		return false
	}
	if ctx.Trigger != model.TriggerOnPhaseEnd {
		return false
	}
	if ctx.TriggerCtx.ActionType != model.ActionMagic {
		return false
	}
	return canPayCrystalLike(ctx, 1)
}

func (h *PrayerManaTideHandler) Execute(ctx *model.Context) error {
	if !spendCrystalLike(ctx, 1) {
		return fmt.Errorf("法力潮汐需要1蓝水晶（红宝石可替代）")
	}
	addMagicAction(ctx.User, "法力潮汐")
	ctx.Game.Log(fmt.Sprintf("%s 发动 [法力潮汐]，额外获得1次法术行动", ctx.User.Name))
	return nil
}

// --- 19. 红莲骑士 ---

type CrimsonKnightCrimsonPactHandler struct{ BaseHandler }

type CrimsonKnightCrimsonFaithHandler struct{ BaseHandler }

type CrimsonKnightBloodyPrayerHandler struct{ BaseHandler }

type CrimsonKnightKillingFeastHandler struct{ BaseHandler }

type CrimsonKnightHotBloodHandler struct{ BaseHandler }

type CrimsonKnightCalmMindHandler struct{ BaseHandler }

type CrimsonKnightCrimsonCrossHandler struct{ BaseHandler }

func (h *CrimsonKnightCrimsonPactHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.TriggerCtx == nil {
		return false
	}
	if ctx.Trigger != model.TriggerOnAttackStart {
		return false
	}
	if ctx.TriggerCtx.AttackInfo != nil && ctx.TriggerCtx.AttackInfo.CounterInitiator != "" {
		return false
	}
	return true
}

func (h *CrimsonKnightCrimsonPactHandler) Execute(ctx *model.Context) error {
	ctx.Game.Heal(ctx.User.ID, 1)
	ctx.Game.Log(fmt.Sprintf("%s 的 [腥红圣约] 触发，+1治疗", ctx.User.Name))
	return nil
}

func (h *CrimsonKnightCrimsonFaithHandler) Execute(ctx *model.Context) error { return nil }

func (h *CrimsonKnightBloodyPrayerHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil {
		return false
	}
	if ctx.User.Heal <= 0 {
		return false
	}
	for _, p := range ctx.Game.GetAllPlayers() {
		if p != nil && p.Camp == ctx.User.Camp && p.ID != ctx.User.ID {
			return true
		}
	}
	return false
}

func (h *CrimsonKnightBloodyPrayerHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("上下文无效")
	}
	if ctx.User.Heal <= 0 {
		return fmt.Errorf("血腥祷言需要至少1点治疗")
	}
	var allyIDs []string
	for _, p := range ctx.Game.GetAllPlayers() {
		if p != nil && p.Camp == ctx.User.Camp && p.ID != ctx.User.ID {
			allyIDs = append(allyIDs, p.ID)
		}
	}
	if len(allyIDs) == 0 {
		return fmt.Errorf("没有可分配治疗的队友")
	}
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type": "crk_bloody_prayer_x",
			"user_id":     ctx.User.ID,
			"max_x":       ctx.User.Heal,
			"ally_ids":    allyIDs,
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [血腥祷言]，请选择X与治疗队友", ctx.User.Name))
	return nil
}

func (h *CrimsonKnightKillingFeastHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.TriggerCtx == nil {
		return false
	}
	if ctx.Trigger != model.TriggerOnAttackHit {
		return false
	}
	if ctx.TriggerCtx.AttackInfo != nil && ctx.TriggerCtx.AttackInfo.CounterInitiator != "" {
		return false
	}
	return getToken(ctx.User, "crk_blood_mark") > 0
}

func (h *CrimsonKnightKillingFeastHandler) Execute(ctx *model.Context) error {
	if getToken(ctx.User, "crk_blood_mark") <= 0 {
		return nil
	}
	addToken(ctx.User, "crk_blood_mark", -1, 0, 3)
	// 先提升本次命中伤害，再追加自伤到 PendingDamageQueue。
	// 否则 append 触发底层扩容时，DamageVal 可能指向旧切片元素导致加伤丢失。
	if ctx.TriggerCtx != nil && ctx.TriggerCtx.DamageVal != nil {
		*ctx.TriggerCtx.DamageVal += 2
	}
	// 规则：先结算本技能自伤，再结算本次攻击命中伤害。
	ctx.Game.AddPendingDamageFront(model.PendingDamage{
		SourceID:              ctx.User.ID,
		TargetID:              ctx.User.ID,
		Damage:                4,
		DamageType:            "magic",
		AllowCrimsonFaithHeal: true,
		Stage:                 0,
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [杀戮盛宴]，移除1血印并对自己造成4伤害，本次攻击伤害+2", ctx.User.Name))
	return nil
}

func (h *CrimsonKnightHotBloodHandler) Execute(ctx *model.Context) error { return nil }

func (h *CrimsonKnightCalmMindHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.TriggerCtx == nil {
		return false
	}
	if ctx.Trigger != model.TriggerOnPhaseEnd {
		return false
	}
	if getToken(ctx.User, "crk_hot_form") == 0 {
		return false
	}
	if ctx.TriggerCtx.ActionType != model.ActionAttack && ctx.TriggerCtx.ActionType != model.ActionMagic {
		return false
	}
	return canPayCrystalLike(ctx, 1)
}

func (h *CrimsonKnightCalmMindHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("戒骄戒躁上下文无效")
	}
	if !spendCrystalLike(ctx, 1) {
		return fmt.Errorf("戒骄戒躁需要1蓝水晶（红宝石可替代）")
	}
	setToken(ctx.User, "crk_hot_form", 0)
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type": "crk_calm_mind_action",
			"user_id":     ctx.User.ID,
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [戒骄戒躁]，脱离热血沸腾形态并选择额外行动类型", ctx.User.Name))
	return nil
}

func (h *CrimsonKnightCrimsonCrossHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.Target == nil {
		return false
	}
	if getToken(ctx.User, "crk_blood_mark") <= 0 {
		return false
	}
	magicCount := 0
	for _, c := range ctx.User.Hand {
		if c.Type == model.CardTypeMagic {
			magicCount++
		}
	}
	return magicCount >= 2
}

func (h *CrimsonKnightCrimsonCrossHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Target == nil || ctx.Game == nil {
		return fmt.Errorf("腥红十字需要目标")
	}
	if getToken(ctx.User, "crk_blood_mark") <= 0 {
		return fmt.Errorf("血印不足")
	}
	addToken(ctx.User, "crk_blood_mark", -1, 0, 3)
	ctx.Game.AddPendingDamage(model.PendingDamage{
		SourceID:              ctx.User.ID,
		TargetID:              ctx.User.ID,
		Damage:                4,
		DamageType:            "magic",
		AllowCrimsonFaithHeal: true,
		Stage:                 0,
	})
	ctx.Game.AddPendingDamage(model.PendingDamage{
		SourceID:   ctx.User.ID,
		TargetID:   ctx.Target.ID,
		Damage:     3,
		DamageType: "magic",
		Stage:      0,
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [腥红十字]，对自己造成4点法术伤害，并对 %s 造成3点法术伤害", ctx.User.Name, ctx.Target.Name))
	return nil
}

// --- 20. 英灵人形 ---

type HomunculusBattlePatternHandler struct{ BaseHandler }

type HomunculusRageSuppressHandler struct{ BaseHandler }

type HomunculusRuneSmashHandler struct{ BaseHandler }

type HomunculusGlyphFusionHandler struct{ BaseHandler }

type HomunculusRuneReforgeHandler struct{ BaseHandler }

type HomunculusDualEchoHandler struct{ BaseHandler }

func (h *HomunculusBattlePatternHandler) Execute(ctx *model.Context) error { return nil }

func (h *HomunculusRageSuppressHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.TriggerCtx == nil {
		return false
	}
	if ctx.Trigger != model.TriggerOnAttackMiss {
		return false
	}
	if ctx.TriggerCtx.AttackInfo != nil && ctx.TriggerCtx.AttackInfo.CounterInitiator != "" {
		return false
	}
	return getToken(ctx.User, "hom_war_rune") > 0
}

func (h *HomunculusRageSuppressHandler) Execute(ctx *model.Context) error {
	if getToken(ctx.User, "hom_war_rune") <= 0 {
		return nil
	}
	addToken(ctx.User, "hom_war_rune", -1, 0, 99)
	addToken(ctx.User, "hom_magic_rune", 1, 0, 99)
	ctx.Game.Log(fmt.Sprintf("%s 发动 [怒火压制]，翻转1战纹为魔纹", ctx.User.Name))
	return nil
}

func (h *HomunculusRuneSmashHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.TriggerCtx == nil {
		return false
	}
	if ctx.Trigger != model.TriggerOnAttackHit {
		return false
	}
	if ctx.TriggerCtx.AttackInfo != nil && ctx.TriggerCtx.AttackInfo.CounterInitiator != "" {
		return false
	}
	if getToken(ctx.User, "hom_war_rune") <= 0 {
		return false
	}
	if ctx.TriggerCtx.Card == nil {
		return false
	}
	ele := ctx.TriggerCtx.Card.Element
	sameCnt := 0
	for _, c := range ctx.User.Hand {
		if c.Element == ele {
			sameCnt++
		}
	}
	return sameCnt > 0
}

func (h *HomunculusRuneSmashHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil || ctx.TriggerCtx == nil || ctx.TriggerCtx.Card == nil {
		return fmt.Errorf("战纹碎击上下文无效")
	}
	if getToken(ctx.User, "hom_war_rune") <= 0 {
		return fmt.Errorf("战纹不足")
	}
	attackEle := ctx.TriggerCtx.Card.Element
	var candidates []int
	for i, c := range ctx.User.Hand {
		if c.Element == attackEle {
			candidates = append(candidates, i)
		}
	}
	if len(candidates) == 0 {
		return fmt.Errorf("没有可弃置的同系牌")
	}
	maxY := 0
	if getToken(ctx.User, "hom_burst_form") > 0 {
		warRunes := getToken(ctx.User, "hom_war_rune")
		if warRunes > 1 {
			maxY = warRunes - 1
		}
	}
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type":       "hom_rune_smash_x",
			"user_id":           ctx.User.ID,
			"user_ctx":          ctx,
			"attack_element":    string(attackEle),
			"max_x":             len(candidates),
			"candidate_indices": candidates,
			"max_y":             maxY,
			"selected_indices":  []int{},
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [战纹碎击]，请选择X、弃牌与Y", ctx.User.Name))
	return nil
}

func (h *HomunculusGlyphFusionHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.TriggerCtx == nil {
		return false
	}
	if ctx.Trigger != model.TriggerOnAttackMiss {
		return false
	}
	if ctx.TriggerCtx.AttackInfo != nil && ctx.TriggerCtx.AttackInfo.CounterInitiator != "" {
		return false
	}
	if getToken(ctx.User, "hom_magic_rune") <= 0 {
		return false
	}
	attackEle := model.Element("")
	if ctx.TriggerCtx.Card != nil {
		attackEle = ctx.TriggerCtx.Card.Element
	}
	uniqueElements := map[model.Element]bool{}
	for _, c := range ctx.User.Hand {
		if c.Element != attackEle {
			uniqueElements[c.Element] = true
		}
	}
	return len(uniqueElements) >= 2
}

func (h *HomunculusGlyphFusionHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil || ctx.TriggerCtx == nil {
		return fmt.Errorf("魔纹融合上下文无效")
	}
	if getToken(ctx.User, "hom_magic_rune") <= 0 {
		return fmt.Errorf("魔纹不足")
	}
	attackEle := model.Element("")
	if ctx.TriggerCtx.Card != nil {
		attackEle = ctx.TriggerCtx.Card.Element
	}
	var candidates []int
	for i, c := range ctx.User.Hand {
		if c.Element != attackEle {
			candidates = append(candidates, i)
		}
	}
	uniqueElements := map[model.Element]bool{}
	for _, idx := range candidates {
		uniqueElements[ctx.User.Hand[idx].Element] = true
	}
	if len(uniqueElements) < 2 {
		return fmt.Errorf("异系牌不足2张")
	}
	maxX := len(uniqueElements)
	maxY := 0
	if getToken(ctx.User, "hom_burst_form") > 0 {
		magicRunes := getToken(ctx.User, "hom_magic_rune")
		if magicRunes > 1 {
			maxY = magicRunes - 1
		}
	}
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type":       "hom_glyph_fusion_x",
			"user_id":           ctx.User.ID,
			"user_ctx":          ctx,
			"attack_element":    string(attackEle),
			"max_x":             maxX,
			"candidate_indices": candidates,
			"max_y":             maxY,
			"selected_indices":  []int{},
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [魔纹融合]，请选择X、弃牌与Y", ctx.User.Name))
	return nil
}

func (h *HomunculusRuneReforgeHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil {
		return false
	}
	return ctx.User.Gem > 0 && getToken(ctx.User, "hom_burst_form") == 0
}

func (h *HomunculusRuneReforgeHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("符文改造上下文无效")
	}
	if ctx.User.Gem <= 0 {
		return fmt.Errorf("符文改造需要红宝石")
	}
	ctx.User.Gem--
	setToken(ctx.User, "hom_burst_form", 1)
	ctx.Game.DrawCards(ctx.User.ID, 1)
	totalRunes := getToken(ctx.User, "hom_war_rune") + getToken(ctx.User, "hom_magic_rune")
	if totalRunes <= 0 {
		totalRunes = 3
	}
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type": "hom_rune_reforge_distribution",
			"user_id":     ctx.User.ID,
			"total_runes": totalRunes,
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [符文改造]，进入蓄势迸发形态并摸1张牌，请调整战纹/魔纹分配", ctx.User.Name))
	return nil
}

func (h *HomunculusDualEchoHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.TriggerCtx == nil {
		return false
	}
	if ctx.Trigger != model.TriggerOnDamageTaken {
		return false
	}
	if ctx.TriggerCtx.SourceID != ctx.User.ID {
		return false
	}
	if ctx.TriggerCtx.DamageVal == nil || *ctx.TriggerCtx.DamageVal <= 0 {
		return false
	}
	return canPayCrystalLike(ctx, 1)
}

func (h *HomunculusDualEchoHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil || ctx.TriggerCtx == nil || ctx.TriggerCtx.DamageVal == nil {
		return fmt.Errorf("双重回响上下文无效")
	}
	if !canPayCrystalLike(ctx, 1) {
		return fmt.Errorf("双重回响需要1蓝水晶（红宝石可替代）")
	}
	damage := *ctx.TriggerCtx.DamageVal
	if damage > 3 {
		damage = 3
	}
	if damage <= 0 {
		return nil
	}
	var targetIDs []string
	for _, p := range ctx.Game.GetAllPlayers() {
		if p == nil || p.ID == ctx.User.ID || p.ID == ctx.TriggerCtx.TargetID {
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
			"choice_type": "hom_dual_echo_target",
			"user_id":     ctx.User.ID,
			"target_ids":  targetIDs,
			"damage":      damage,
			// 成本在最终选定目标后再扣除，便于在目标弹框中取消本次响应。
			"cost_pending": 1,
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [双重回响]，请选择追加伤害目标", ctx.User.Name))
	return nil
}

// --- 21. 神官 ---

type PriestDivineRevelationHandler struct{ BaseHandler }

type PriestDivineBlessHandler struct{ BaseHandler }

type PriestWaterPowerHandler struct{ BaseHandler }

type PriestGuardianHandler struct{ BaseHandler }

type PriestDivineContractHandler struct{ BaseHandler }

type PriestDivineDomainHandler struct{ BaseHandler }

func (h *PriestDivineRevelationHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.TriggerCtx == nil {
		return false
	}
	if ctx.Trigger != model.TriggerOnPhaseEnd {
		return false
	}
	return ctx.TriggerCtx.ActionType == model.ActionBuy ||
		ctx.TriggerCtx.ActionType == model.ActionSynthesize ||
		ctx.TriggerCtx.ActionType == model.ActionExtract
}

func (h *PriestDivineRevelationHandler) Execute(ctx *model.Context) error {
	ctx.Game.Heal(ctx.User.ID, 1)
	ctx.Game.Log(fmt.Sprintf("%s 的 [神圣启示] 触发，+1治疗", ctx.User.Name))
	return nil
}

func (h *PriestDivineBlessHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil {
		return false
	}
	cnt := 0
	for _, c := range ctx.User.Hand {
		if c.Type == model.CardTypeMagic {
			cnt++
		}
	}
	return cnt >= 2
}

func (h *PriestDivineBlessHandler) Execute(ctx *model.Context) error {
	ctx.Game.Heal(ctx.User.ID, 2)
	ctx.Game.Log(fmt.Sprintf("%s 发动 [神圣祈福]，恢复2点治疗", ctx.User.Name))
	return nil
}

func (h *PriestWaterPowerHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil {
		return false
	}
	return hasElementCard(ctx.User, model.ElementWater)
}

func (h *PriestWaterPowerHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("上下文无效")
	}
	target := ctx.Target
	if target == nil || target.Camp != ctx.User.Camp {
		return fmt.Errorf("水之神力需要指定队友")
	}

	discardedAny, _ := ctx.Selections["discardedCards"]
	discarded, _ := discardedAny.([]model.Card)
	if len(discarded) == 0 || discarded[0].Element != model.ElementWater {
		return fmt.Errorf("水之神力需要先弃置1张水系牌")
	}
	ctx.Game.Log(fmt.Sprintf("%s 为 [水之神力] 弃置了 %s", ctx.User.Name, discarded[0].Name))

	// 若弃完水系牌后仍有手牌，则第二张弃牌视为“交给队友的牌”。
	if len(discarded) >= 2 {
		give := discarded[1]
		target.Hand = append(target.Hand, give)
		ctx.Game.Log(fmt.Sprintf("%s 将 %s 交给了 %s", ctx.User.Name, give.Name, target.Name))
	}
	ctx.Game.Heal(ctx.User.ID, 1)
	ctx.Game.Heal(target.ID, 1)
	ctx.Game.Log(fmt.Sprintf("%s 发动 [水之神力]，与 %s 各+1治疗", ctx.User.Name, target.Name))
	return nil
}

func (h *PriestGuardianHandler) Execute(ctx *model.Context) error { return nil }

func (h *PriestDivineContractHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil {
		return false
	}
	return ctx.User.Heal > 0 && canPayCrystalLike(ctx, 1)
}

func (h *PriestDivineContractHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Target == nil || ctx.Game == nil {
		return fmt.Errorf("神圣契约需要目标")
	}
	if ctx.Target.Camp != ctx.User.Camp {
		return fmt.Errorf("神圣契约目标必须是队友")
	}
	if !spendCrystalLike(ctx, 1) {
		return fmt.Errorf("神圣契约需要1蓝水晶（红宝石可替代）")
	}
	if ctx.User.Heal <= 0 {
		return fmt.Errorf("神圣契约需要可转移治疗")
	}
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type": "priest_divine_contract_x",
			"user_id":     ctx.User.ID,
			"target_id":   ctx.Target.ID,
			"max_x":       ctx.User.Heal,
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [神圣契约]，请选择转移治疗值X（目标：%s）", ctx.User.Name, ctx.Target.Name))
	return nil
}

func (h *PriestDivineDomainHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil {
		return false
	}
	return canPayCrystalLike(ctx, 1)
}

func (h *PriestDivineDomainHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("神圣领域上下文无效")
	}
	allyIDs := []string{}
	allTargetIDs := []string{}
	for _, p := range ctx.Game.GetAllPlayers() {
		if p == nil {
			continue
		}
		allTargetIDs = append(allTargetIDs, p.ID)
		if p.Camp == ctx.User.Camp && p.ID != ctx.User.ID {
			allyIDs = append(allyIDs, p.ID)
		}
	}
	modeOptions := []string{}
	if ctx.User.Heal > 0 {
		modeOptions = append(modeOptions, "damage")
	}
	if len(allyIDs) > 0 {
		modeOptions = append(modeOptions, "heal")
	}
	if len(modeOptions) == 0 {
		return fmt.Errorf("神圣领域当前无可用分支（伤害分支需至少1点治疗，治疗分支需至少1名队友）")
	}
	if !spendCrystalLike(ctx, 1) {
		return fmt.Errorf("神圣领域需要1蓝水晶（红宝石可替代）")
	}
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type":     "priest_divine_domain_mode",
			"user_id":         ctx.User.ID,
			"mode_options":    modeOptions,
			"all_target_ids":  allTargetIDs,
			"ally_target_ids": allyIDs,
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [神圣领域]，等待选择分支", ctx.User.Name))
	return nil
}

// --- 22. 阴阳师 ---

type OnmyojiShikigamiDescendHandler struct{ BaseHandler }

type OnmyojiYinYangShiftHandler struct{ BaseHandler }

type OnmyojiShikigamiShiftHandler struct{ BaseHandler }

type OnmyojiDarkRitualHandler struct{ BaseHandler }

type OnmyojiBindingHandler struct{ BaseHandler }

type OnmyojiLifeBarrierHandler struct{ BaseHandler }

func (h *OnmyojiShikigamiDescendHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil {
		return false
	}
	if getToken(ctx.User, "onmyoji_form") > 0 {
		return false
	}
	if len(ctx.User.Hand) < 2 {
		return false
	}
	factionCount := map[string]int{}
	for _, c := range ctx.User.Hand {
		if c.Faction == "" {
			continue
		}
		factionCount[c.Faction]++
		if factionCount[c.Faction] >= 2 {
			return true
		}
	}
	return false
}

func (h *OnmyojiShikigamiDescendHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("上下文无效")
	}
	if getToken(ctx.User, "onmyoji_form") > 0 {
		return fmt.Errorf("已处于式神形态")
	}
	setToken(ctx.User, "onmyoji_form", 1)
	addToken(ctx.User, "onmyoji_ghost_fire", 1, 0, 3)
	addAttackAction(ctx.User, "式神降临")
	ctx.Game.Log(fmt.Sprintf("%s 发动 [式神降临]，弃2张同命格手牌后进入式神形态并+1鬼火，获得额外攻击行动", ctx.User.Name))
	return nil
}

func (h *OnmyojiYinYangShiftHandler) CanUse(ctx *model.Context) bool { return false }

func (h *OnmyojiYinYangShiftHandler) Execute(ctx *model.Context) error { return nil }

func (h *OnmyojiShikigamiShiftHandler) CanUse(ctx *model.Context) bool { return false }

func (h *OnmyojiShikigamiShiftHandler) Execute(ctx *model.Context) error { return nil }

func (h *OnmyojiDarkRitualHandler) Execute(ctx *model.Context) error { return nil }

func (h *OnmyojiBindingHandler) CanUse(ctx *model.Context) bool { return false }

func (h *OnmyojiBindingHandler) Execute(ctx *model.Context) error { return nil }

func (h *OnmyojiLifeBarrierHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return false
	}
	if !canPayCrystalLike(ctx, 1) {
		return false
	}
	for _, p := range ctx.Game.GetAllPlayers() {
		if p != nil && p.Camp == ctx.User.Camp && p.ID != ctx.User.ID {
			return true
		}
	}
	return false
}

func (h *OnmyojiLifeBarrierHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("生命结界上下文无效")
	}
	gf := addToken(ctx.User, "onmyoji_ghost_fire", 1, 0, 3)

	// 分支①可选队友（不含自己）
	var supportTargetIDs []string
	// 分支②可选队友（需有手牌可弃）
	var releaseTargetIDs []string
	for _, p := range ctx.Game.GetAllPlayers() {
		if p == nil || p.Camp != ctx.User.Camp || p.ID == ctx.User.ID {
			continue
		}
		supportTargetIDs = append(supportTargetIDs, p.ID)
		if len(p.Hand) > 0 {
			releaseTargetIDs = append(releaseTargetIDs, p.ID)
		}
	}
	if len(supportTargetIDs) == 0 {
		return fmt.Errorf("生命结界没有可选队友目标")
	}

	// 分支②：式神形态 + 手牌中存在“2张同命格”组合 + 有队友可弃牌
	var releaseCombos []string
	if getToken(ctx.User, "onmyoji_form") > 0 && len(releaseTargetIDs) > 0 {
		for i := 0; i < len(ctx.User.Hand); i++ {
			if ctx.User.Hand[i].Faction == "" {
				continue
			}
			for j := i + 1; j < len(ctx.User.Hand); j++ {
				if ctx.User.Hand[i].Faction == ctx.User.Hand[j].Faction {
					releaseCombos = append(releaseCombos, fmt.Sprintf("%d,%d", i, j))
				}
			}
		}
	}

	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type":         "onmyoji_life_barrier_mode",
			"user_id":             ctx.User.ID,
			"ghost_fire":          gf,
			"support_target_ids":  supportTargetIDs,
			"release_target_ids":  releaseTargetIDs,
			"release_card_combos": releaseCombos,
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [生命结界]，鬼火+1（当前%d），请选择分支效果", ctx.User.Name, gf))
	return nil
}

// --- 23. 苍炎魔女 ---

type BlazeWitchRebirthClockHandler struct{ BaseHandler }

type BlazeWitchBlazingCodexHandler struct{ BaseHandler }

type BlazeWitchHeavenfireCleaveHandler struct{ BaseHandler }

type BlazeWitchWitchWrathHandler struct{ BaseHandler }

type BlazeWitchSubstituteDollHandler struct{ BaseHandler }

type BlazeWitchPainLinkHandler struct{ BaseHandler }

type BlazeWitchManaInversionHandler struct{ BaseHandler }

func (h *BlazeWitchRebirthClockHandler) CanUse(ctx *model.Context) bool { return false }

func (h *BlazeWitchRebirthClockHandler) Execute(ctx *model.Context) error { return nil }

func (h *BlazeWitchBlazingCodexHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Target == nil || ctx.Game == nil {
		return fmt.Errorf("苍炎法典需要目标")
	}
	ctx.Game.AddPendingDamage(model.PendingDamage{
		SourceID:   ctx.User.ID,
		TargetID:   ctx.Target.ID,
		Damage:     2,
		DamageType: "magic",
		Stage:      0,
	})
	ctx.Game.AddPendingDamage(model.PendingDamage{
		SourceID:   ctx.User.ID,
		TargetID:   ctx.User.ID,
		Damage:     2,
		DamageType: "magic",
		Stage:      0,
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [苍炎法典]，先对 %s 后对自己各造成2点法术伤害", ctx.User.Name, ctx.Target.Name))
	return nil
}

func (h *BlazeWitchHeavenfireCleaveHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil {
		return false
	}
	if getToken(ctx.User, "bw_flame_form") > 0 {
		return true
	}
	return getToken(ctx.User, "bw_rebirth") > 0
}

func (h *BlazeWitchHeavenfireCleaveHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Target == nil || ctx.Game == nil {
		return fmt.Errorf("天火断空需要目标")
	}
	damage := 3
	userCampMorale := ctx.Game.GetCampMorale(string(ctx.User.Camp))
	targetCampMorale := ctx.Game.GetCampMorale(string(ctx.Target.Camp))
	if userCampMorale < targetCampMorale {
		damage++
	}
	if getToken(ctx.User, "bw_flame_form") == 0 {
		if getToken(ctx.User, "bw_rebirth") <= 0 {
			return fmt.Errorf("天火断空需要至少1点重生")
		}
		addToken(ctx.User, "bw_rebirth", -1, 0, 4)
	}
	ctx.Game.AddPendingDamage(model.PendingDamage{
		SourceID:   ctx.User.ID,
		TargetID:   ctx.Target.ID,
		Damage:     damage,
		DamageType: "magic",
		Stage:      0,
	})
	ctx.Game.AddPendingDamage(model.PendingDamage{
		SourceID:   ctx.User.ID,
		TargetID:   ctx.User.ID,
		Damage:     damage,
		DamageType: "magic",
		Stage:      0,
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [天火断空]，先对 %s 后对自己各造成%d点法术伤害", ctx.User.Name, ctx.Target.Name, damage))
	return nil
}

func (h *BlazeWitchWitchWrathHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil {
		return false
	}
	return len(ctx.User.Hand) < 4
}

func (h *BlazeWitchWitchWrathHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("魔女之怒上下文无效")
	}
	setToken(ctx.User, "bw_flame_form", 1)
	setToken(ctx.User, "bw_flame_release_pending", 1)
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type": "bw_witch_wrath_draw",
			"user_id":     ctx.User.ID,
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [魔女之怒]，进入烈焰形态并选择摸牌数量", ctx.User.Name))
	return nil
}

func (h *BlazeWitchSubstituteDollHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.Game == nil || ctx.TriggerCtx == nil {
		return false
	}
	if ctx.Trigger != model.TriggerOnDamageTaken {
		return false
	}
	if getToken(ctx.User, "bw_substitute_lock") > 0 {
		return false
	}
	if ctx.Flags["IsMagicDamage"] {
		return false
	}
	if ctx.TriggerCtx.DamageVal == nil || *ctx.TriggerCtx.DamageVal <= 0 {
		return false
	}
	magicCount := 0
	for _, c := range ctx.User.Hand {
		if c.Type == model.CardTypeMagic {
			magicCount++
		}
	}
	if magicCount <= 0 {
		return false
	}
	for _, p := range ctx.Game.GetAllPlayers() {
		if p != nil && p.Camp == ctx.User.Camp && p.ID != ctx.User.ID {
			return true
		}
	}
	return false
}

func (h *BlazeWitchSubstituteDollHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("替身玩偶上下文无效")
	}
	setToken(ctx.User, "bw_substitute_lock", 1)
	var magicIndices []int
	for i, c := range ctx.User.Hand {
		if c.Type == model.CardTypeMagic {
			magicIndices = append(magicIndices, i)
		}
	}
	var allyIDs []string
	for _, p := range ctx.Game.GetAllPlayers() {
		if p != nil && p.Camp == ctx.User.Camp && p.ID != ctx.User.ID {
			allyIDs = append(allyIDs, p.ID)
		}
	}
	if len(magicIndices) == 0 || len(allyIDs) == 0 {
		setToken(ctx.User, "bw_substitute_lock", 0)
		return fmt.Errorf("替身玩偶缺少可用牌或队友")
	}
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type":   "bw_substitute_doll_card",
			"user_id":       ctx.User.ID,
			"magic_indices": magicIndices,
			"ally_ids":      allyIDs,
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [替身玩偶]，请选择要弃置的法术牌", ctx.User.Name))
	return nil
}

func (h *BlazeWitchPainLinkHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil {
		return false
	}
	return canPayCrystalLike(ctx, 1)
}

func (h *BlazeWitchPainLinkHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Target == nil || ctx.Game == nil {
		return fmt.Errorf("痛苦链接需要目标")
	}
	ctx.Game.AddPendingDamage(model.PendingDamage{
		SourceID:   ctx.User.ID,
		TargetID:   ctx.Target.ID,
		Damage:     1,
		DamageType: "magic",
		Stage:      0,
	})
	ctx.Game.AddPendingDamage(model.PendingDamage{
		SourceID:   ctx.User.ID,
		TargetID:   ctx.User.ID,
		Damage:     1,
		DamageType: "magic",
		Stage:      0,
	})
	setToken(ctx.User, "bw_pain_link_pending_discard", 1)
	setToken(ctx.User, "bw_pain_link_pending_hits", 2)
	ctx.Game.Log(fmt.Sprintf("%s 发动 [痛苦链接]，先对 %s 后对自己各造成1点法术伤害", ctx.User.Name, ctx.Target.Name))
	return nil
}

func (h *BlazeWitchManaInversionHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.Game == nil || ctx.TriggerCtx == nil {
		return false
	}
	if ctx.Trigger != model.TriggerOnDamageTaken {
		return false
	}
	if getToken(ctx.User, "bw_mana_inversion_lock") > 0 {
		return false
	}
	if !ctx.Flags["IsMagicDamage"] {
		return false
	}
	if ctx.TriggerCtx.DamageVal == nil || *ctx.TriggerCtx.DamageVal <= 0 {
		return false
	}
	magicCount := 0
	for _, c := range ctx.User.Hand {
		if c.Type == model.CardTypeMagic {
			magicCount++
		}
	}
	if magicCount < 2 {
		return false
	}
	if !canPayCrystalLike(ctx, 1) {
		return false
	}
	for _, p := range ctx.Game.GetAllPlayers() {
		if p != nil && p.Camp != ctx.User.Camp {
			return true
		}
	}
	return false
}

func (h *BlazeWitchManaInversionHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("魔能反转上下文无效")
	}
	if !spendCrystalLike(ctx, 1) {
		return fmt.Errorf("魔能反转需要1蓝水晶（红宝石可替代）")
	}
	setToken(ctx.User, "bw_mana_inversion_lock", 1)
	magicCount := 0
	for _, c := range ctx.User.Hand {
		if c.Type == model.CardTypeMagic {
			magicCount++
		}
	}
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type": "bw_mana_inversion_x",
			"user_id":     ctx.User.ID,
			"max_x":       magicCount,
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [魔能反转]，请选择弃牌数量X", ctx.User.Name))
	return nil
}

// --- 24. 贤者 ---

type SageWisdomCodexHandler struct{ BaseHandler }

type SageMagicReboundHandler struct{ BaseHandler }

type SageArcaneCodexHandler struct{ BaseHandler }

type SageHolyCodexHandler struct{ BaseHandler }

func (h *SageWisdomCodexHandler) CanUse(ctx *model.Context) bool { return false }

func (h *SageWisdomCodexHandler) Execute(ctx *model.Context) error { return nil }

func (h *SageMagicReboundHandler) CanUse(ctx *model.Context) bool { return false }

func (h *SageMagicReboundHandler) Execute(ctx *model.Context) error { return nil }

func sageDistinctElements(user *model.Player) map[model.Element]int {
	out := map[model.Element]int{}
	if user == nil {
		return out
	}
	for _, c := range user.Hand {
		if c.Element == "" {
			continue
		}
		out[c.Element]++
	}
	return out
}

func (h *SageArcaneCodexHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil {
		return false
	}
	return ctx.User.Gem > 0 && len(sageDistinctElements(ctx.User)) >= 2
}

func (h *SageArcaneCodexHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("魔道法典上下文无效")
	}
	distinct := sageDistinctElements(ctx.User)
	maxX := len(distinct)
	if maxX < 2 {
		return fmt.Errorf("魔道法典需要至少2种不同元素手牌")
	}
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type": "sage_arcane_x",
			"user_id":     ctx.User.ID,
			"max_x":       maxX,
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [魔道法典]，请选择X并弃置异系牌", ctx.User.Name))
	return nil
}

func (h *SageHolyCodexHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil {
		return false
	}
	return ctx.User.Gem > 0 && len(sageDistinctElements(ctx.User)) >= 3
}

func (h *SageHolyCodexHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("圣洁法典上下文无效")
	}
	distinct := sageDistinctElements(ctx.User)
	maxX := len(distinct)
	if maxX < 3 {
		return fmt.Errorf("圣洁法典需要至少3种不同元素手牌")
	}
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type": "sage_holy_x",
			"user_id":     ctx.User.ID,
			"max_x":       maxX,
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [圣洁法典]，请选择X并弃置异系牌", ctx.User.Name))
	return nil
}
