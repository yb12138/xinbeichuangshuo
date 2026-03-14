package engine

import (
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"starcup-engine/internal/data"
	"starcup-engine/internal/engine/skills"
	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
	"strconv"
	"strings"
	"time"
)

type GameEngine struct {
	State      *model.GameState
	dispatcher *SkillDispatcher
	observer   model.GameObserver // [新增] 持有观察者
	// 记录“当前回合内各阵营已对哪些敌方角色造成过法术伤害”。
	turnMagicDamageTargets map[model.Camp]map[string]bool
	actionSummary          *actionSummary
	actionSummaryTurn      int
	suppressSealOnDiscard  bool
}

func NewGameEngine(observer model.GameObserver) *GameEngine {
	skills.InitHandlers()
	engine := &GameEngine{
		State:    model.NewGameState(),
		observer: observer,
		turnMagicDamageTargets: map[model.Camp]map[string]bool{
			model.RedCamp:  {},
			model.BlueCamp: {},
		},
		actionSummaryTurn: 0,
	}
	engine.dispatcher = NewSkillDispatcher(engine)
	return engine
}

// AddPlayer 添加玩家
func (e *GameEngine) AddPlayer(id, name, role string, camp model.Camp) error {
	if len(e.State.Players) >= 6 {
		return errors.New("游戏人数已满 (6人)")
	}
	if _, exists := e.State.Players[id]; exists {
		return errors.New("玩家ID已存在")
	}

	player := &model.Player{
		ID:             id,
		Name:           name,
		Camp:           camp,
		Role:           role,
		Hand:           make([]model.Card, 0),
		Blessings:      make([]model.Card, 0),
		ExclusiveCards: make([]model.Card, 0),
		MaxHand:        6, // 初始手牌上限
		Heal:           0,
		MaxHeal:        2,
		IsActive:       false,
		Tokens:         map[string]int{},
		TurnState:      model.NewPlayerTurnState(),
	}

	// 查找并绑定角色数据
	characters := data.GetCharacters()
	for _, c := range characters {
		if c.ID == role || c.Name == role {
			// Make a copy or pointer? Pointer is fine as Character data is static
			// But we need to be careful if we modify it (we shouldn't)
			// Actually model.Player has *Character
			charCopy := c // Copy struct
			player.Character = &charCopy
			player.MaxHand = c.MaxHand
			break
		}
	}
	if player.Character == nil {
		// Fallback or warning
		e.Log(fmt.Sprintf("Warning: Character not found for role %s", role))
	}
	e.applyRoleDefaults(player)

	e.State.Players[id] = player
	e.State.PlayerOrder = append(e.State.PlayerOrder, id)
	return nil
}

// applyRoleDefaults 初始化角色的基础指示物/上限等（与 AddPlayer 保持一致）
func (e *GameEngine) applyRoleDefaults(player *model.Player) {
	if player == nil || player.Character == nil {
		return
	}
	switch player.Character.ID {
	case "plague_mage":
		// 圣渎：治疗上限初始+3（默认2 -> 5）
		player.MaxHeal = 5
	case "crimson_sword_spirit":
		// 鲜血默认上限3，由散华轮舞分支2临时提高到4。
		player.Tokens["css_blood_cap"] = 3
		player.Tokens["css_blood"] = 0
	case "prayer_master":
		player.Tokens["prayer_form"] = 0
		player.Tokens["prayer_rune"] = 0
	case "crimson_knight":
		// 腥红信仰：治疗上限初始+2。
		player.MaxHeal = 4
		player.Tokens["crk_blood_mark"] = 0
		player.Tokens["crk_hot_form"] = 0
	case "war_homunculus":
		// 战纹掌控：开局3战纹（当前实现为指示物）。
		player.Tokens["hom_war_rune"] = 3
		player.Tokens["hom_magic_rune"] = 0
		player.Tokens["hom_burst_form"] = 0
	case "priest":
		// 圣使守护：治疗上限+4。
		player.MaxHeal = 6
	case "onmyoji":
		player.Tokens["onmyoji_form"] = 0
		player.Tokens["onmyoji_ghost_fire"] = 0
	case "blaze_witch":
		player.Tokens["bw_rebirth"] = 0
		player.Tokens["bw_flame_form"] = 0
		player.Tokens["bw_flame_release_pending"] = 0
	case "magic_lancer":
		player.Tokens["ml_phantom_form"] = 0
		player.Tokens["ml_stardust_pending"] = 0
		player.Tokens["ml_stardust_wait_discard"] = 0
		player.Tokens["ml_stardust_morale_before"] = 0
	case "spirit_caster":
		player.Tokens["sc_power_count"] = 0
	case "bard":
		player.Tokens["bd_inspiration"] = 0
		player.Tokens["bd_prisoner_form"] = 0
		player.Tokens["bd_descent_used_turn"] = 0
	case "hero":
		player.Crystal += 2
		player.Tokens["hero_anger"] = 0
		player.Tokens["hero_wisdom"] = 0
		player.Tokens["hero_exhaustion_form"] = 0
		player.Tokens["hero_exhaustion_release_pending"] = 0
		player.Tokens["hero_roar_active"] = 0
		player.Tokens["hero_roar_damage_pending"] = 0
		player.Tokens["hero_calm_end_crystal_pending"] = 0
		player.Tokens["hero_dead_duel_pending"] = 0
	case "fighter":
		player.Tokens["fighter_qi"] = 0
		player.Tokens["fighter_hundred_dragon_form"] = 0
		player.Tokens["fighter_hundred_dragon_target_order"] = 0
		player.Tokens["fighter_attack_start_skill_lock"] = 0
		player.Tokens["fighter_charge_pending"] = 0
		player.Tokens["fighter_charge_damage_pending"] = 0
		player.Tokens["fighter_qiburst_force_no_counter"] = 0
	case "holy_bow":
		player.Crystal += 2
		player.MaxHeal += 1
		player.Tokens["hb_cannon"] = 1
		player.Tokens["hb_faith"] = 0
		player.Tokens["hb_form"] = 0
		player.Tokens["hb_special_used_turn"] = 0
		player.Tokens["hb_auto_fill_done_turn"] = 0
		player.Tokens["hb_shard_miss_pending"] = 0
	case "soul_sorcerer":
		player.Tokens["ss_blue_soul"] = 0
		player.Tokens["ss_yellow_soul"] = 0
		player.Tokens["ss_link_active"] = 0
	case "moon_goddess":
		player.Tokens["mg_dark_form"] = 0
		player.Tokens["mg_new_moon"] = 0
		player.Tokens["mg_petrify"] = 0
		player.Tokens["mg_dark_moon_count"] = 0
		player.Tokens["mg_blasphemy_used_turn"] = 0
		player.Tokens["mg_blasphemy_pending"] = 0
		player.Tokens["mg_next_attack_no_counter"] = 0
		player.Tokens["mg_extra_turn_pending"] = 0
	case "blood_priestess":
		player.Tokens["bp_bleed_form"] = 0
		player.Tokens["bp_shared_life_active"] = 0
		player.Tokens["bp_bleed_tick_done_turn"] = 0
	case "butterfly_dancer":
		player.Tokens["bt_pupa"] = 0
		player.Tokens["bt_cocoon_count"] = 0
		player.Tokens["bt_wither_active"] = 0
		player.Tokens["bt_wither_pending"] = 0
	}
}

func ensureExclusiveStarterCard(player *model.Player, skillTitle string, buildCard func() model.Card) bool {
	if player == nil || player.Character == nil || skillTitle == "" || buildCard == nil {
		return false
	}
	charName := player.Character.Name
	for _, c := range player.ExclusiveCards {
		if c.MatchExclusive(charName, skillTitle) {
			return false
		}
	}
	// 兼容旧状态：若该专属卡误在手牌区，迁移到专属卡区。
	for i, c := range player.Hand {
		if !c.MatchExclusive(charName, skillTitle) {
			continue
		}
		player.Hand = append(player.Hand[:i], player.Hand[i+1:]...)
		player.ExclusiveCards = append(player.ExclusiveCards, c)
		return true
	}
	player.ExclusiveCards = append(player.ExclusiveCards, buildCard())
	return true
}

func makeStarterFiveElementsBindCard(player *model.Player) model.Card {
	return model.Card{
		ID:              fmt.Sprintf("starter-%s-five_elements_bind", player.ID),
		Name:            "五系束缚",
		Type:            model.CardTypeMagic,
		Element:         model.ElementLight,
		Faction:         player.Character.Faction,
		Damage:          0,
		Description:     "封印师开局自带专属技能卡",
		ExclusiveChar1:  player.Character.Name,
		ExclusiveSkill1: "五系束缚",
	}
}

func makeStarterRoseCourtyardCard(player *model.Player) model.Card {
	return model.Card{
		ID:              fmt.Sprintf("starter-%s-css_rose_courtyard", player.ID),
		Name:            "血蔷薇庭院",
		Type:            model.CardTypeMagic,
		Element:         model.ElementDark,
		Faction:         player.Character.Faction,
		Damage:          0,
		Description:     "血色剑灵开局自带专属技能卡",
		ExclusiveChar1:  player.Character.Name,
		ExclusiveSkill1: "血蔷薇庭院",
	}
}

func makeStarterHeroTauntCard(player *model.Player) model.Card {
	return model.Card{
		ID:              fmt.Sprintf("starter-%s-hero_taunt", player.ID),
		Name:            "挑衅",
		Type:            model.CardTypeMagic,
		Element:         model.ElementFire,
		Faction:         player.Character.Faction,
		Damage:          0,
		Description:     "勇者开局自带专属技能卡",
		ExclusiveChar1:  player.Character.Name,
		ExclusiveSkill1: "挑衅",
	}
}

func makeStarterSoulLinkCard(player *model.Player) model.Card {
	return model.Card{
		ID:              fmt.Sprintf("starter-%s-soul_link", player.ID),
		Name:            "灵魂链接",
		Type:            model.CardTypeMagic,
		Element:         model.ElementDark,
		Faction:         player.Character.Faction,
		Damage:          0,
		Description:     "灵魂术士开局自带专属技能卡",
		ExclusiveChar1:  player.Character.Name,
		ExclusiveSkill1: "灵魂链接",
	}
}

func makeStarterBloodSharedLifeCard(player *model.Player) model.Card {
	return model.Card{
		ID:              fmt.Sprintf("starter-%s-bp_shared_life", player.ID),
		Name:            "同生共死",
		Type:            model.CardTypeMagic,
		Element:         model.ElementDark,
		Faction:         player.Character.Faction,
		Damage:          0,
		Description:     "血之巫女开局自带专属技能卡",
		ExclusiveChar1:  player.Character.Name,
		ExclusiveSkill1: "同生共死",
	}
}

func (e *GameEngine) returnRoseCourtyardToExclusive(player *model.Player) bool {
	if player == nil {
		return false
	}
	returned := false
	filtered := make([]*model.FieldCard, 0, len(player.Field))
	for _, fc := range player.Field {
		if fc != nil && fc.Mode == model.FieldEffect && fc.Effect == model.EffectRoseCourtyard && fc.SourceID == player.ID {
			player.RestoreExclusiveCard(fc.Card)
			returned = true
			continue
		}
		filtered = append(filtered, fc)
	}
	player.Field = filtered
	// 兜底：旧状态可能只有 active token 没有场上牌，避免专属卡永久丢失。
	if !returned && player.Character != nil && !player.HasExclusiveCard(player.Character.Name, "血蔷薇庭院") {
		player.RestoreExclusiveCard(makeStarterRoseCourtyardCard(player))
		returned = true
	}
	return returned
}

// ensureStarterRoleCards 为特定角色补充开局自带专属技能卡（置于专属卡区，不占手牌）。
func (e *GameEngine) ensureStarterRoleCards(player *model.Player) {
	if player == nil || player.Character == nil {
		return
	}
	switch player.Character.ID {
	case "sealer":
		if ensureExclusiveStarterCard(player, "五系束缚", func() model.Card {
			return makeStarterFiveElementsBindCard(player)
		}) {
			e.Log(fmt.Sprintf("[Setup] %s 获得开局专属技能卡【五系束缚】（专属卡区）", player.Name))
		}
	case "crimson_sword_spirit":
		if ensureExclusiveStarterCard(player, "血蔷薇庭院", func() model.Card {
			return makeStarterRoseCourtyardCard(player)
		}) {
			e.Log(fmt.Sprintf("[Setup] %s 获得开局专属技能卡【血蔷薇庭院】（专属卡区）", player.Name))
		}
	case "hero":
		if ensureExclusiveStarterCard(player, "挑衅", func() model.Card {
			return makeStarterHeroTauntCard(player)
		}) {
			e.Log(fmt.Sprintf("[Setup] %s 获得开局专属技能卡【挑衅】（专属卡区）", player.Name))
		}
	case "soul_sorcerer":
		if ensureExclusiveStarterCard(player, "灵魂链接", func() model.Card {
			return makeStarterSoulLinkCard(player)
		}) {
			e.Log(fmt.Sprintf("[Setup] %s 获得开局专属技能卡【灵魂链接】（专属卡区）", player.Name))
		}
	case "blood_priestess":
		if ensureExclusiveStarterCard(player, "同生共死", func() model.Card {
			return makeStarterBloodSharedLifeCard(player)
		}) {
			e.Log(fmt.Sprintf("[Setup] %s 获得开局专属技能卡【同生共死】（专属卡区）", player.Name))
		}
	}
}

// buildMagicMissilePrompt 构建魔弹响应提示
func (e *GameEngine) buildMagicMissilePrompt() *model.Prompt {
	chain := e.State.MagicBulletChain
	if chain == nil {
		return nil
	}

	playerID := chain.TargetID
	player := e.State.Players[playerID]
	if player == nil {
		return nil
	}

	damage := chain.CurrentDamage
	hasShield := player.HasFieldEffect(model.EffectShield)
	takeLabel := "承受伤害"
	if hasShield {
		takeLabel = "承受（将触发圣盾）"
	}
	effectHints := []string{}
	if hasShield {
		effectHints = append(effectHints, "你身上有【圣盾】：可先应战/防御；若本次选择承受伤害，将自动消耗圣盾抵挡魔弹。")
	}

	return &model.Prompt{
		Type:       model.PromptConfirm,
		PlayerID:   playerID,
		AttackerID: chain.SourcePlayerID,
		Message:    fmt.Sprintf("你成为了【魔弹】的目标，当前伤害为 %d，请选择应对：", damage),
		Options: []model.PromptOption{
			{ID: "take", Label: takeLabel},
			{ID: "counter", Label: "打出【魔弹】传递"},
			{ID: "defend", Label: "使用【圣光】抵挡"},
		},
		EffectHints: effectHints,
		Min:         1,
		Max:         1,
	}
}

// StartGame 开始游戏
func (e *GameEngine) StartGame() error {
	if len(e.State.Players) < 2 {
		return errors.New("玩家人数不足")
	}

	// 1. 初始化牌库
	e.State.Deck = rules.InitDeck()
	e.State.Deck = rules.Shuffle(e.State.Deck)

	// 2. 发初始手牌 (每人4张)
	for _, pid := range e.State.PlayerOrder {
		player := e.State.Players[pid]
		cards, newDeck, _ := rules.DrawCards(e.State.Deck, e.State.DiscardPile, 4)
		player.Hand = append(player.Hand, cards...)
		e.State.Deck = newDeck
		e.ensureStarterRoleCards(player)
	}

	// 3. 随机决定先手
	rand.Seed(time.Now().UnixNano())
	startIndex := rand.Intn(len(e.State.PlayerOrder))

	// 为了让 NextTurn 切换到 startIndex，我们将 CurrentTurn 设为前一个位置
	// 注意：CurrentTurn 是索引
	count := len(e.State.PlayerOrder)
	e.State.CurrentTurn = (startIndex - 1 + count) % count

	firstPlayerID := e.State.PlayerOrder[startIndex]
	e.Log(fmt.Sprintf("[Game] 游戏开始! 首发玩家: %s (%s)",
		e.State.Players[firstPlayerID].Name,
		e.State.Players[firstPlayerID].Camp))

	e.State.CurrentTurn = startIndex

	player := e.State.Players[firstPlayerID]
	player.IsActive = true
	player.TurnState = model.NewPlayerTurnState()
	e.actionSummaryTurn = 1

	e.State.Phase = model.PhaseBuffResolve
	e.resetTurnMagicDamageTracker()

	// 进入第一回合
	e.Drive()

	return nil
}

// triggerFieldEffects 触发场上效果牌
func (e *GameEngine) triggerFieldEffects(p *model.Player, trigger model.EffectTrigger, ctx *model.Context) {
	var remain []*model.FieldCard

	for _, fc := range p.Field {
		if fc.Mode != model.FieldEffect || fc.Trigger != trigger {
			remain = append(remain, fc)
			continue
		}

		// 触发效果
		switch fc.Effect {
		case model.EffectPoison:
			e.applyPoisonEffect(p, fc.SourceID, ctx)
		case model.EffectShield:
			e.applyShieldEffect(p, ctx)
		case model.EffectWeak:
			e.applyWeakEffect(p, ctx)
		case model.EffectSealFire:
			e.applySealEffect(p, ctx, "Fire")
		case model.EffectSealWater:
			e.applySealEffect(p, ctx, "Water")
		case model.EffectSealEarth:
			e.applySealEffect(p, ctx, "Earth")
		case model.EffectSealWind:
			e.applySealEffect(p, ctx, "Wind")
		case model.EffectSealThunder:
			e.applySealEffect(p, ctx, "Thunder")
		case model.EffectFiveElementsBind:
			e.applyFiveElementsBindEffect(p, ctx)
		}

		// 触发后进弃牌堆
		e.State.DiscardPile = append(e.State.DiscardPile, fc.Card)
		e.Log(fmt.Sprintf(
			"[Field] %s 面前的【%s】触发效果并被弃置",
			p.Name, fc.Card.Name,
		))
	}

	p.Field = remain
}

// applyPoisonEffect 应用中毒效果
func (e *GameEngine) applyPoisonEffect(p *model.Player, sourceID string, ctx *model.Context) {
	// 中毒：造成1点法术伤害
	allowCrimsonFaithHeal := sourceID != "" && sourceID == p.ID
	e.AddPendingDamage(model.PendingDamage{
		SourceID:              sourceID,
		TargetID:              p.ID,
		Damage:                1,
		DamageType:            "poison",
		AllowCrimsonFaithHeal: allowCrimsonFaithHeal,
		Stage:                 0,
	})
	e.Log(fmt.Sprintf("[Effect] %s 受到中毒伤害", p.Name))
}

// applyShieldEffect 应用圣盾效果
func (e *GameEngine) applyShieldEffect(p *model.Player, ctx *model.Context) {
	if ctx != nil {
		ctx.Flags["shielded"] = true
	}
	e.Log(fmt.Sprintf("[Effect] %s 的圣盾生效", p.Name))
}

// applyWeakEffect 应用虚弱效果
func (e *GameEngine) applyWeakEffect(p *model.Player, ctx *model.Context) {
	if ctx != nil {
		ctx.Flags["weakened"] = true
	}
	e.Log(fmt.Sprintf("[Effect] %s 陷入虚弱状态", p.Name))
}

// applySealEffect 应用封印效果
func (e *GameEngine) applySealEffect(p *model.Player, ctx *model.Context, element string) {
	if ctx != nil {
		ctx.Flags["sealed_"+element] = true
	}
	e.Log(fmt.Sprintf("[Effect] %s 受到%s封印限制", p.Name, element))
}

// triggerSealDamageForCardUse 响应阶段使用同系牌时触发封印伤害（不依赖 TriggerOnCardUsed）。
func (e *GameEngine) triggerSealDamageForCardUse(p *model.Player, card *model.Card) {
	if p == nil || card == nil {
		return
	}
	var effectType model.EffectType
	var effectName string
	switch card.Element {
	case model.ElementWater:
		effectType = model.EffectSealWater
		effectName = "水之封印"
	case model.ElementFire:
		effectType = model.EffectSealFire
		effectName = "火之封印"
	case model.ElementEarth:
		effectType = model.EffectSealEarth
		effectName = "地之封印"
	case model.ElementWind:
		effectType = model.EffectSealWind
		effectName = "风之封印"
	case model.ElementThunder:
		effectType = model.EffectSealThunder
		effectName = "雷之封印"
	default:
		return
	}

	var sourceID string
	for _, fc := range p.Field {
		if fc.Mode == model.FieldEffect && fc.Effect == effectType {
			sourceID = fc.SourceID
			break
		}
	}
	if sourceID == "" {
		return
	}
	e.Log(fmt.Sprintf("[Seal] %s 使用了 %s 系牌，触发了 %s！", p.Name, card.Element, effectName))
	e.AddPendingDamage(model.PendingDamage{
		SourceID:           sourceID,
		TargetID:           p.ID,
		Damage:             3,
		DamageType:         "magic",
		EffectTypeToRemove: effectType,
	})
}

// applyFiveElementsBindEffect 应用五系束缚效果
// 规则：回合开始时，玩家选择：1.摸(2+X)张牌(X=场上封印数,最多2) 2.放弃行动移除此牌
func (e *GameEngine) applyFiveElementsBindEffect(p *model.Player, ctx *model.Context) {
	// 计算场上封印数量
	sealCount := 0
	for _, player := range e.State.Players {
		for _, fc := range player.Field {
			if fc.Mode == model.FieldEffect {
				switch fc.Effect {
				case model.EffectSealWater, model.EffectSealFire, model.EffectSealEarth, model.EffectSealWind, model.EffectSealThunder:
					sealCount++
				}
			}
		}
	}
	// X 最多为 2
	x := sealCount
	if x > 2 {
		x = 2
	}
	drawCount := 2 + x

	// 设置中断，让玩家选择
	e.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: p.ID,
		Context: map[string]interface{}{
			"choice_type": "five_elements_bind",
			"draw_count":  drawCount,
			"player_id":   p.ID,
		},
	})
	e.Log(fmt.Sprintf("[Effect] %s 受到五系束缚影响，需要选择：摸%d张牌 或 放弃行动", p.Name, drawCount))
}

// UseSkill 使用技能
func (e *GameEngine) UseSkill(playerID, skillID string, targetIDs []string, discardIndices []int) error {
	player := e.State.Players[playerID]
	if player == nil {
		return fmt.Errorf("player not found")
	}

	if !player.IsActive {
		return fmt.Errorf("not your turn")
	}

	if player.Character == nil {
		return fmt.Errorf("no character assigned")
	}

	// 查找技能定义
	var skillDef *model.SkillDefinition
	for i := range player.Character.Skills {
		if player.Character.Skills[i].ID == skillID {
			skillDef = &player.Character.Skills[i]
			break
		}
	}

	if skillDef == nil {
		return fmt.Errorf("skill %s not found for character %s", skillID, player.Character.Name)
	}

	// 神官-神圣领域：弃牌数量为“2 或当前全部手牌（当手牌<2）”。
	requiredDiscards := skillDef.CostDiscards
	if skillID == "priest_divine_domain" && requiredDiscards > len(player.Hand) {
		requiredDiscards = len(player.Hand)
	}
	// 神官-水之神力：若弃完水系牌后无牌可交，则本次仅需弃1张水系牌。
	if skillID == "priest_water_power" && requiredDiscards > 0 && requiredDiscards > len(player.Hand) {
		requiredDiscards = len(player.Hand)
	}

	// 【修正】移除自动选择逻辑，改为交互式提示
	if requiredDiscards > 0 && len(discardIndices) == 0 {
		// 如果技能需要弃牌但玩家没选，创建一个交互式中断
		e.State.PendingInterrupt = &model.Interrupt{
			Type:     model.InterruptDiscard,
			PlayerID: playerID,
			SkillIDs: []string{skillID},
			Context: map[string]interface{}{
				"discard_count": requiredDiscards,
				"skill_id":      skillID,
				"target_ids":    targetIDs, // 传递多目标 ID
			},
		}
		e.State.Phase = model.PhaseDiscardSelection
		e.Log(fmt.Sprintf("%s 请选择用于发动 [%s] 的卡牌", player.Name, skillDef.Title))
		return nil // 流程暂停，等待交互输入
	}

	// 禁止重复索引
	seen := map[int]bool{}
	for _, idx := range discardIndices {
		if seen[idx] {
			return fmt.Errorf("不能重复选择同一张牌")
		}
		seen[idx] = true
	}

	// ===== 校验弃牌数量 =====
	if requiredDiscards > 0 {
		if len(discardIndices) != requiredDiscards {
			return fmt.Errorf(
				"技能需要弃 %d 张牌，你选择了 %d 张",
				requiredDiscards, len(discardIndices),
			)
		}
	}

	// ===== 校验弃牌内容（元素 / 类型 / 命格 / 独有）=====
	var discardedCards []model.Card
	for _, idx := range discardIndices {
		if idx < 0 || idx >= len(player.Hand) {
			return fmt.Errorf("弃牌索引越界: %d", idx)
		}

		card := player.Hand[idx]

		// 元素校验
		effectiveElement := card.Element
		if skillDef.DiscardElement != "" {
			effectiveElement = e.blazeWitchAttackElement(player, card)
		}
		if skillDef.DiscardElement != "" && effectiveElement != skillDef.DiscardElement {
			return fmt.Errorf("弃牌 %s 不符合元素要求", card.Name)
		}
		// 魔弹融合：仅允许弃置火系或地系牌。
		if skillID == "magic_bullet_fusion" &&
			card.Element != model.ElementFire &&
			card.Element != model.ElementEarth {
			return fmt.Errorf("魔弹融合只能弃置火系或地系牌")
		}

		// 类型校验
		if skillDef.DiscardType != "" && card.Type != skillDef.DiscardType {
			return fmt.Errorf("弃牌 %s 不符合卡牌类型要求", card.Name)
		}

		// 命格校验
		if skillDef.DiscardFate != "" && card.Faction != skillDef.DiscardFate {
			return fmt.Errorf("弃牌 %s 不符合命格要求", card.Name)
		}

		// 独有技校验
		if skillDef.RequireExclusive {
			if !card.MatchExclusive(player.Character.Name, skillDef.Title) {
				return fmt.Errorf("弃牌 %s 不是该技能对应的独有牌", card.Name)
			}
		}

		discardedCards = append(discardedCards, card)
	}

	// 阴阳师：式神降临必须弃置2张“命格相同”的手牌。
	if skillID == "onmyoji_shikigami_descend" {
		if len(discardedCards) != 2 {
			return fmt.Errorf("式神降临需要弃置2张手牌")
		}
		f1 := strings.TrimSpace(discardedCards[0].Faction)
		f2 := strings.TrimSpace(discardedCards[1].Faction)
		if f1 == "" || f2 == "" || f1 != f2 {
			return fmt.Errorf("式神降临需要弃置2张命格相同的手牌")
		}
	}
	// 神官：水之神力必须“先弃1张水系牌”，若手牌足够则第二张用于交给队友。
	if skillID == "priest_water_power" {
		if len(discardedCards) <= 0 {
			return fmt.Errorf("水之神力需要弃置1张水系牌")
		}
		if discardedCards[0].Element != model.ElementWater {
			return fmt.Errorf("水之神力第一张必须弃置水系牌")
		}
	}

	// 专属技能卡支持“专属卡区”：当技能要求独有牌但不要求手牌弃置时，自动消耗专属卡区中的对应卡牌。
	// 兼容旧测试：若外部仍显式提交了弃牌索引，则沿用上面的手牌弃置路径。
	var consumedExclusiveCard *model.Card
	if skillDef.RequireExclusive && skillDef.CostDiscards <= 0 && len(discardedCards) == 0 {
		if player.Character == nil || player.Character.Name == "" {
			return fmt.Errorf("角色信息缺失，无法校验独有牌")
		}
		card, ok := player.ConsumeExclusiveCard(player.Character.Name, skillDef.Title)
		if !ok {
			return fmt.Errorf("未找到技能 [%s] 对应的专属技能卡", skillDef.Title)
		}
		consumedExclusiveCard = &card
	}

	// 检查技能类型和当前阶段
	switch skillDef.Type {
	case model.SkillTypeStartup:
		if e.State.Phase != model.PhaseStartup {
			return fmt.Errorf("startup skills can only be used during trigger phase")
		}
	case model.SkillTypeAction:
		if e.State.Phase != model.PhaseActionSelection {
			return fmt.Errorf("action skills can only be used during action phase")
		}
	case model.SkillTypeResponse:
		// Response skills are handled through events
		return fmt.Errorf("response skills are triggered automatically")
	case model.SkillTypePassive:
		return fmt.Errorf("passive skills are triggered automatically")
	}

	// 检查额外行动类型限制
	if player.TurnState.CurrentExtraAction == "Attack" {
		// 在额外攻击行动中，不能使用主动技能
		return fmt.Errorf("当前是额外攻击行动，不能使用技能，只能发起攻击")
	}
	// 额外法术行动允许发动主动技能（视为法术行动）。

	// 检查回合限制
	if model.ContainsSkillTag(skillDef.Tags, model.TagTurnLimit) {
		if player.TurnState.UsedSkillCounts[skillID] > 0 {
			return fmt.Errorf("skill %s can only be used once per turn", skillID)
		}
	}

	// ===== 盖牌消耗 =====
	if skillDef.CostCoverCards > 0 {
		coverCards, err := player.ConsumeCoverCards(skillDef.CostCoverCards)
		if err != nil {
			return fmt.Errorf("盖牌消耗失败: %v", err)
		}
		e.State.DiscardPile = append(e.State.DiscardPile, coverCards...)
		e.Log(fmt.Sprintf("%s 消耗了 %d 张盖牌作为技能消耗", player.Name, skillDef.CostCoverCards))
	}

	// 检查资源消耗：
	// - 宝石消耗必须由宝石支付
	// - 水晶消耗可由红宝石替代
	if !canPaySkillEnergyCost(player, skillDef.CostGem, skillDef.CostCrystal) {
		return fmt.Errorf(
			"资源不足: 需要 宝石%d/水晶%d，当前 宝石%d/水晶%d（红宝石可替代水晶）",
			skillDef.CostGem, skillDef.CostCrystal, player.Gem, player.Crystal,
		)
	}

	// 验证目标
	var target *model.Player          // 单目标兼容
	var actualTargets []*model.Player // 实际解析出的目标列表

	if skillDef.TargetType != model.TargetNone {
		if len(targetIDs) == 0 {
			return fmt.Errorf("skill requires target(s)")
		}

		for _, id := range targetIDs {
			p := e.State.Players[id]
			if p == nil {
				return fmt.Errorf("target player %s not found", id)
			}
			actualTargets = append(actualTargets, p)
		}

		if len(actualTargets) == 0 {
			return fmt.Errorf("no valid targets found")
		}

		// 校验目标数量
		if skillDef.MaxTargets > 0 && len(actualTargets) > skillDef.MaxTargets {
			return fmt.Errorf("技能最多只能指定 %d 个目标，你指定了 %d 个", skillDef.MaxTargets, len(actualTargets))
		}
		if skillDef.MinTargets > 0 && len(actualTargets) < skillDef.MinTargets {
			return fmt.Errorf("技能最少需要指定 %d 个目标，你指定了 %d 个", skillDef.MinTargets, len(actualTargets))
		}

		// 单目标兼容性：如果只有一个目标，将其赋值给 target
		if len(actualTargets) == 1 {
			target = actualTargets[0]
		}

		// 检查目标类型 (对所有目标进行校验)
		for _, t := range actualTargets {
			switch skillDef.TargetType {
			case model.TargetSelf:
				if t.ID != playerID {
					return fmt.Errorf("skill can only target self")
				}
			case model.TargetEnemy:
				if t.Camp == player.Camp {
					return fmt.Errorf("skill can only target enemies")
				}
			case model.TargetAlly:
				if t.Camp != player.Camp {
					return fmt.Errorf("skill can only target allies")
				}
			case model.TargetAllySelf:
				if t.Camp != player.Camp {
					return fmt.Errorf("skill can only target allies or self")
				}
			case model.TargetAny:
				// Can target any player
			default:
				// For TargetSpecific and other cases, additional validation might be needed
			}
		}
	}

	// ===== 执行弃牌（索引从大到小）=====
	if skillDef.PlaceCard && len(actualTargets) > 0 {
		fcTarget := actualTargets[0]
		// 基础效果不可叠加：同一角色面前同名基础效果最多一张（中毒/虚弱/圣盾/赐福等）。
		if skillDef.PlaceMode == model.FieldEffect && model.IsBasicEffect(string(skillDef.PlaceEffect)) {
			for _, fc := range fcTarget.Field {
				if fc == nil || fc.Mode != model.FieldEffect {
					continue
				}
				if fc.Effect == skillDef.PlaceEffect {
					return fmt.Errorf("%s 面前已有同种基础效果，不可重复放置", fcTarget.Name)
				}
			}
		}
	}
	if skillID == "seal_break" {
		if len(actualTargets) == 0 || actualTargets[0] == nil {
			return fmt.Errorf("封印破碎需要指定目标")
		}
		hasBasic := false
		for _, fc := range actualTargets[0].Field {
			if fc == nil || fc.Mode != model.FieldEffect {
				continue
			}
			if model.IsBasicEffect(string(fc.Effect)) {
				hasBasic = true
				break
			}
		}
		if !hasBasic {
			return fmt.Errorf("%s 面前没有可收回的基础效果", actualTargets[0].Name)
		}
	}

	// ===== 执行弃牌（索引从大到小）=====
	e.NotifyCardRevealed(playerID, discardedCards, "discard")
	sort.Sort(sort.Reverse(sort.IntSlice(discardIndices)))
	for _, idx := range discardIndices {
		player.Hand = append(player.Hand[:idx], player.Hand[idx+1:]...)
	}
	_ = e.maybeAutoReleaseBloodPriestessByHand(player, "手牌<3强制脱离流血形态")
	if skillID == "priest_water_power" && len(discardedCards) >= 2 {
		// 第二张弃牌将由技能效果转交给队友，不进入弃牌堆。
		e.State.DiscardPile = append(e.State.DiscardPile, discardedCards[0])
	} else {
		e.State.DiscardPile = append(e.State.DiscardPile, discardedCards...)
	}

	// ===== 场上牌放置 =====
	if skillDef.PlaceCard {
		var placedCard model.Card
		placedCardReady := false
		if len(discardedCards) > 0 {
			// 常规放置：使用第一张弃掉的牌。
			placedCard = discardedCards[0]
			placedCardReady = true
		} else if consumedExclusiveCard != nil {
			// 专属卡区放置：使用自动消耗的专属技能卡。
			placedCard = *consumedExclusiveCard
			placedCardReady = true
		}
		if !placedCardReady {
			return fmt.Errorf("需要专属技能卡才能放置场上牌")
		}

		// 放置目标由 actualTargets 决定
		if len(actualTargets) == 0 {
			return fmt.Errorf("放置场上牌需要指定目标")
		}

		// 目前只支持 PlaceCard = true for single target or first target in multi-target
		fcTarget := actualTargets[0]

		fc := &model.FieldCard{
			Card:     placedCard,
			OwnerID:  fcTarget.ID,
			SourceID: player.ID,
			Mode:     skillDef.PlaceMode,
			Effect:   skillDef.PlaceEffect,
			Trigger:  skillDef.PlaceTrigger,
		}

		fcTarget.AddFieldCard(fc)

		e.Log(fmt.Sprintf(
			"[Skill] %s 在 %s 面前放置了场上牌: %s (效果: %s, 触发: %s)",
			player.Name, fcTarget.Name, placedCard.Name,
			fc.Effect, fc.Trigger,
		))
	}
	// 若自动消耗了专属卡但本次不是放置型技能，则将该卡正常进入弃牌堆。
	if consumedExclusiveCard != nil && !skillDef.PlaceCard {
		e.State.DiscardPile = append(e.State.DiscardPile, *consumedExclusiveCard)
	}

	// ===== 资源扣除必须放在弃牌校验之后 =====
	if !consumeSkillEnergyCost(player, skillDef.CostGem, skillDef.CostCrystal) {
		return fmt.Errorf(
			"资源扣除失败: 需要 宝石%d/水晶%d，当前 宝石%d/水晶%d",
			skillDef.CostGem, skillDef.CostCrystal, player.Gem, player.Crystal,
		)
	}

	// 更新回合状态
	player.TurnState.UsedSkillCounts[skillID]++

	// 灵符师灵符技能：先处理“展示触发封印”，后续伤害/弃牌通过 DeferredFollowups 串行结算。
	if skillID == "sc_talisman_thunder" || skillID == "sc_talisman_wind" {
		resolvedTargetIDs := make([]string, 0, len(actualTargets))
		for _, t := range actualTargets {
			if t != nil {
				resolvedTargetIDs = append(resolvedTargetIDs, t.ID)
			}
		}
		if err := e.beginSpiritCasterTalisman(player, skillID, resolvedTargetIDs, discardedCards); err != nil {
			return err
		}
		e.recordSkillUsage(player.ID, skillDef.Title, skillDef.Type)
		e.Log(fmt.Sprintf("[Skill] %s 使用了技能: %s (%s)", player.Name, skillDef.Title, skillDef.Description))

		// 与其他主动技能一致：视为本回合完成一次法术行动。
		player.TurnState.HasActed = true
		phaseEventCtx := &model.EventContext{
			Type:       model.EventPhaseEnd,
			SourceID:   player.ID,
			ActionType: model.ActionMagic,
		}
		phaseCtx := e.buildContext(player, nil, model.TriggerOnPhaseEnd, phaseEventCtx)
		e.dispatcher.OnTrigger(model.TriggerOnPhaseEnd, phaseCtx)
		return nil
	}

	// 执行技能逻辑
	handler := skills.GetHandler(skillID)
	if handler == nil {
		return fmt.Errorf("skill handler not found for %s", skillID)
	}

	ctx := e.buildContext(player, target, model.TriggerNone, nil) // target 兼容单目标
	ctx.Targets = actualTargets                                   // 填充多目标
	if ctx.Selections == nil {
		ctx.Selections = map[string]interface{}{}
	}
	ctx.Selections["discardedCards"] = discardedCards
	// ctx.Args 参数不再需要，因为弃牌索引已通过 discardIndices 传递

	err := handler.Execute(ctx)
	if err != nil {
		return fmt.Errorf("skill execution failed: %v", err)
	}

	e.recordSkillUsage(player.ID, skillDef.Title, skillDef.Type)
	e.Log(fmt.Sprintf("[Skill] %s 使用了技能: %s (%s)", player.Name, skillDef.Title, skillDef.Description))

	// 主动技能使用后，结束当前回合
	if skillDef.Type == model.SkillTypeAction && skillID != "adventurer_fraud" {
		// 1. 标记行动类型为 ActionMagic (法术/技能)
		// 这样系统知道玩家本回合进行的是"技能"而不是"普通攻击"
		player.TurnState.HasActed = true
		phaseEventCtx := &model.EventContext{
			Type:       model.EventPhaseEnd,
			SourceID:   player.ID,
			ActionType: model.ActionMagic,
		}
		phaseCtx := e.buildContext(player, nil, model.TriggerOnPhaseEnd, phaseEventCtx)
		e.dispatcher.OnTrigger(model.TriggerOnPhaseEnd, phaseCtx)
	}

	return nil
}

// Drive 状态机驱动函数，自动在阶段间转换或等待用户输入
func (e *GameEngine) Drive() {
	const maxIterations = 100
	iterations := 0
	for {
		// [调试] 打印状态流转，方便排查死循环
		e.Log(fmt.Sprintf("[Debug] Drive Loop: %d, Phase: %s", iterations, e.State.Phase))

		iterations++
		if iterations > maxIterations {
			e.Log(fmt.Sprintf("[System] 严重错误：状态机死循环检测 (最后状态: %s)", e.State.Phase))
			// 紧急制动：强制进入等待状态，避免崩溃
			return
		}
		// 如果有待处理的中断，不自动推进
		if e.State.PendingInterrupt != nil {
			return
		}
		// 仅在没有待处理延迟伤害时推进“延迟后续”。
		// 这样可保证诸如“封印伤害先结算，再继续技能后续”的严格顺序。
		if e.State.Phase != model.PhasePendingDamageResolution &&
			len(e.State.PendingDamageQueue) == 0 &&
			len(e.State.DeferredFollowups) > 0 {
			e.processDeferredFollowups()
			if e.State.PendingInterrupt != nil {
				return
			}
			continue
		}

		// 行动汇总：当系统回到可继续行动的空闲状态时输出汇总信息
		e.finalizeActionSummaryIfIdle()

		currentPid := e.State.PlayerOrder[e.State.CurrentTurn]
		player := e.State.Players[currentPid]
		// 血之巫女：手牌<3时立即强制脱离流血形态（跨玩家事件也要及时生效）。
		for _, p := range e.State.Players {
			_ = e.maybeAutoReleaseBloodPriestessByHand(p, "手牌<3强制脱离流血形态")
		}

		switch e.State.Phase {
		case model.PhaseBuffResolve:
			// 蝶舞者：凋零的“对方士气最低为1”持续到其下个回合开始前。
			if e.isButterflyDancer(player) {
				if player.Tokens == nil {
					player.Tokens = map[string]int{}
				}
				if player.Tokens["bt_wither_active"] > 0 {
					player.Tokens["bt_wither_active"] = 0
					e.Log(fmt.Sprintf("%s 的 [凋零] 效果到期：对方士气下限保护已解除", player.Name))
				}
			}
			// 血之巫女：流血形态回合开始先自损1点法术伤害（先于中毒/虚弱）。
			if e.isBloodPriestess(player) {
				if player.Tokens == nil {
					player.Tokens = map[string]int{}
				}
				if player.Tokens["bp_bleed_form"] > 0 && player.Tokens["bp_bleed_tick_done_turn"] <= 0 {
					player.Tokens["bp_bleed_tick_done_turn"] = 1
					e.Log(fmt.Sprintf("%s 的 [流血] 生效：回合开始对自己造成1点法术伤害", player.Name))
					e.AddPendingDamage(model.PendingDamage{
						SourceID:   player.ID,
						TargetID:   player.ID,
						Damage:     1,
						DamageType: "magic",
						Stage:      0,
					})
					e.State.Phase = model.PhasePendingDamageResolution
					e.State.ReturnPhase = model.PhaseBuffResolve
					continue
				}
			}
			// 回合开始触发类场上效果（如五系束缚）
			fieldCtx := e.buildContext(player, nil, model.TriggerOnBuffPhase, nil)
			e.triggerFieldEffects(player, model.EffectTriggerOnTurnStart, fieldCtx)
			if e.State.PendingInterrupt != nil {
				return
			}
			// 构建上下文
			skillCtx := e.buildContext(player, nil, model.TriggerOnBuffPhase, nil)

			// 触发！Dispatcher 会去遍历 Field，找到 Trigger == OnBuffPhase 的卡（中毒/虚弱）
			// 并执行对应的 Handler
			e.dispatcher.OnTrigger(model.TriggerOnBuffPhase, skillCtx)

			// 处理可能产生的中断（例如虚弱需要玩家选择弃牌还是跳过）
			if e.State.PendingInterrupt != nil {
				return // 等待用户输入
			}

			if e.State.Phase == model.PhaseTurnEnd {
				return
			}

			// 处理完 BuffResolve 后，检查是否有延迟伤害需要结算
			if len(e.State.PendingDamageQueue) > 0 {
				e.State.Phase = model.PhasePendingDamageResolution
				e.State.ReturnPhase = model.PhaseStartup
			} else {
				// 正常进入启动阶段
				e.State.Phase = model.PhaseStartup
			}

		case model.PhasePendingDamageResolution:
			// 延迟伤害结算阶段
			if e.processPendingDamages() {
				return // 有中断，暂停
			}

			// 队列处理完毕，进入下一阶段
			if e.State.ReturnPhase != "" {
				e.State.Phase = e.State.ReturnPhase
				e.State.ReturnPhase = ""
			} else {
				// 默认回退到 Startup (针对旧代码未设置 ReturnPhase 的兜底，或者根据之前逻辑)
				// 但为了安全，最好检查上下文。这里暂时默认为 Startup，
				// 因为之前是硬编码为 Startup 的。
				e.State.Phase = model.PhaseStartup
			}

		case model.PhaseStartup:
			// 2. 启动技能阶段
			if player.Tokens == nil {
				player.Tokens = map[string]int{}
			}
			// 魔剑士暗影形态在其“下一次”回合开始时转正。
			// 注意：本回合刚在启动阶段确认了暗影凝聚后，会再次回到 Startup。
			// 此时 HasUsedTriggerSkill=true，不能立即转正，否则同回合暗影之力无法生效。
			if e.isMagicSwordsman(player) &&
				!player.TurnState.HasUsedTriggerSkill &&
				player.Tokens["ms_shadow_form"] > 0 &&
				player.Tokens["ms_shadow_release_pending"] > 0 {
				player.Tokens["ms_shadow_form"] = 0
				player.Tokens["ms_shadow_release_pending"] = 0
				e.Log(fmt.Sprintf("%s 脱离暗影形态并转正", player.Name))
			}
			// 苍炎魔女烈焰形态在“下一次”行动阶段开始前转正。
			// 与暗影形态同理：同回合刚发动启动技后会再次进入 Startup，此时 HasUsedTriggerSkill=true，不能立刻转正。
			if e.isBlazeWitch(player) &&
				!player.TurnState.HasUsedTriggerSkill &&
				player.Tokens["bw_flame_form"] > 0 &&
				player.Tokens["bw_flame_release_pending"] > 0 {
				player.Tokens["bw_flame_form"] = 0
				player.Tokens["bw_flame_release_pending"] = 0
				e.Log(fmt.Sprintf("%s 脱离烈焰形态并转正", player.Name))
			}
			// 每回合开始重置“跳过行动免强制末日审判”和“本回合已强制末日审判”标记
			player.Tokens["arbiter_skip_forced_doomsday"] = 0
			player.Tokens["arbiter_forced_doomsday_done_turn"] = 0
			// 圣弓：每回合开始重置“特殊行动已使用/自动填充已处理”标记。
			player.Tokens["hb_special_used_turn"] = 0
			player.Tokens["hb_auto_fill_done_turn"] = 0
			// 吟游诗人：若当前回合角色持有永恒乐章，可触发激昂狂想曲。
			if e.maybeTriggerBardRousingAtTurnStart(player) {
				return
			}
			// 仲裁者处于审判形态时，每回合开始自动+1审判（上限4）。
			// 仅在数值实际变化时写日志，避免在满层状态反复刷屏。
			if player.Tokens["arbiter_form"] > 0 {
				before := player.Tokens["judgment"]
				if before < 4 {
					player.Tokens["judgment"] = before + 1
					e.Log(fmt.Sprintf("%s 处于审判形态，回合开始审判+1（当前%d）", player.Name, player.Tokens["judgment"]))
				}
			}

			// 检查是否有可用的启动技能
			eventCtx := &model.EventContext{
				Type:     model.EventTurnStart,
				SourceID: currentPid,
			}
			skillCtx := e.buildContext(player, nil, model.TriggerOnTurnStart, eventCtx)
			// 触发启动技能检查（这会设置 PendingInterrupt 如果有可用技能）
			e.dispatcher.OnTrigger(model.TriggerOnTurnStart, skillCtx)

			if e.State.PendingInterrupt != nil && e.State.PendingInterrupt.Type == model.InterruptStartupSkill {
				// 有启动技能可用，等待用户输入
				prompt := e.buildStartupSkillPrompt()
				e.Notify(model.EventAskInput, "请选择是否发动启动技能", prompt)
				return
			}

			// 没有启动技能，继续到 ActionSelection
			e.State.Phase = model.PhaseActionSelection

		case model.PhaseActionSelection:
			// 3. 行动选择阶段
			if player.Tokens == nil {
				player.Tokens = map[string]int{}
			}
			// 勇者：精疲力竭在“下个行动阶段开始”时结束并自伤3。
			if e.isHero(player) &&
				player.Tokens["hero_exhaustion_form"] > 0 &&
				player.Tokens["hero_exhaustion_release_pending"] > 0 &&
				!player.TurnState.HasActed &&
				player.TurnState.CurrentExtraAction == "" {
				player.Tokens["hero_exhaustion_form"] = 0
				player.Tokens["hero_exhaustion_release_pending"] = 0
				e.Log(fmt.Sprintf("%s 的 [精疲力竭] 结束，转正并对自己造成3点法术伤害", player.Name))
				e.InflictDamage(player.ID, player.ID, 3, "magic")
				return
			}

			tauntSourceID := ""
			tauntSrcName := ""
			if tauntCard := getHeroTauntCard(player); tauntCard != nil {
				src := e.State.Players[tauntCard.SourceID]
				// 仅对“敌方来源的挑衅”生效；非法残留直接移除。
				if src == nil || src.Camp == player.Camp {
					e.RemoveFieldCard(player.ID, model.EffectHeroTaunt)
				} else {
					tauntSourceID = src.ID
					tauntSrcName = model.GetPlayerDisplayName(src)
					hasAttackCard := false
					for idx := 0; idx < playableCardCount(player); idx++ {
						c, _, _, ok := getPlayableCardByIndex(player, idx)
						if ok && c.Type == model.CardTypeAttack {
							hasAttackCard = true
							break
						}
					}
					if !hasAttackCard {
						e.Log(fmt.Sprintf("[Taunt] %s 受到【挑衅】约束但无攻击牌，跳过本次行动阶段", player.Name))
						e.RemoveFieldCard(player.ID, model.EffectHeroTaunt)
						e.State.Phase = model.PhaseTurnEnd
						continue
					}
				}
			}
			hasHeroTaunt := tauntSourceID != ""

			judgment := player.Tokens["judgment"]
			if judgment >= 4 &&
				player.Tokens["arbiter_skip_forced_doomsday"] == 0 &&
				player.Tokens["arbiter_forced_doomsday_done_turn"] == 0 &&
				!hasHeroTaunt {
				e.PushInterrupt(&model.Interrupt{
					Type:     model.InterruptChoice,
					PlayerID: currentPid,
					Context: map[string]interface{}{
						"choice_type": "arbiter_forced_doomsday_target",
						"user_id":     currentPid,
						"target_ids":  append([]string{}, e.State.PlayerOrder...),
					},
				})
				return
			}

			var validOptions []model.PromptOption
			var specialOptions []model.PromptOption
			currentExtraAction := player.TurnState.CurrentExtraAction
			isRestrictedExtraAction := currentExtraAction == "Attack" || currentExtraAction == "Magic"
			canMagicAction := e.canCastMagicInAction(player)
			canMagicSkillAction := e.hasUsableActionSkillForExtraMagic(player)
			hasRestrictedExtraActionCard := true
			if isRestrictedExtraAction {
				hasRestrictedExtraActionCard = e.checkExtraActionCards(player, currentExtraAction, player.TurnState.CurrentExtraElement)
			}
			hasFighterHundredDragon := e.isFighter(player) && player.Tokens != nil && player.Tokens["fighter_hundred_dragon_form"] > 0

			// 行动类型选项：
			// - 额外攻击行动：只能选攻击
			// - 额外法术行动：只能选法术
			// - 常规行动：攻击/法术都可选
			switch currentExtraAction {
			case "Attack":
				if hasRestrictedExtraActionCard {
					validOptions = append(validOptions, model.PromptOption{ID: "attack", Label: "攻击"})
				}
			case "Magic":
				// 额外法术行动：允许“法术牌”或“主动技能(视为法术行动)”。
				if hasRestrictedExtraActionCard {
					validOptions = append(validOptions, model.PromptOption{ID: "magic", Label: "法术"})
				}
			default:
				if hasFighterHundredDragon {
					validOptions = append(validOptions, model.PromptOption{ID: "attack", Label: "攻击（百式幻龙拳）"})
				} else if hasHeroTaunt {
					validOptions = append(validOptions, model.PromptOption{ID: "attack", Label: "攻击（受挑衅约束）"})
				} else {
					validOptions = append(validOptions, model.PromptOption{ID: "attack", Label: "攻击"})
					// 常规阶段：即便当前形态不能“打出法术牌”，只要存在可用主动技能，
					// 仍应保留“法术行动”入口（例如魔剑士【暗影流星】）。
					if canMagicAction || canMagicSkillAction {
						validOptions = append(validOptions, model.PromptOption{ID: "magic", Label: "法术"})
					}
				}
			}

			if !hasHeroTaunt && !hasFighterHundredDragon && !isRestrictedExtraAction && !e.State.HasPerformedStartup {
				// 未执行启动技能时，按条件过滤特殊行动
				maxHand := e.GetMaxHand(player)
				canBuyOrSynth := len(player.Hand)+3 <= maxHand

				if canBuyOrSynth {
					specialOptions = append(specialOptions, model.PromptOption{ID: "buy", Label: "购买"})
				}

				var totalStones int
				if player.Camp == model.RedCamp {
					totalStones = e.State.RedGems + e.State.RedCrystals
				} else {
					totalStones = e.State.BlueGems + e.State.BlueCrystals
				}
				if canBuyOrSynth && totalStones >= 3 {
					specialOptions = append(specialOptions, model.PromptOption{ID: "synthesize", Label: "合成"})
				}

				currentEnergy := player.Gem + player.Crystal
				if totalStones > 0 && currentEnergy < 3 {
					specialOptions = append(specialOptions, model.PromptOption{ID: "extract", Label: "提炼"})
				}

				if len(specialOptions) > 0 {
					validOptions = append(validOptions, model.PromptOption{ID: "special", Label: "特殊"})
				}
			}

			// 常规行动下："无法行动"表示展示手牌并重摸；
			// 额外行动受限下：当无合法动作时，允许主动宣告跳过本次额外行动。
			if !isRestrictedExtraAction {
				hasAttackCard := false
				hasMagicCard := false
				for idx := 0; idx < playableCardCount(player); idx++ {
					c, _, _, ok := getPlayableCardByIndex(player, idx)
					if !ok {
						continue
					}
					if c.Type == model.CardTypeAttack {
						hasAttackCard = true
					}
					if c.Type == model.CardTypeMagic && canMagicAction {
						hasMagicCard = true
					}
				}
				canNormalAction := hasAttackCard || (!hasFighterHundredDragon && (hasMagicCard || canMagicSkillAction))
				// 仅当无法执行一般行动（无攻击牌也无法术牌）时提供"无法行动"
				if !canNormalAction {
					validOptions = append(validOptions, model.PromptOption{ID: "cannot_act", Label: "无法行动（展示手牌）"})
				}
			} else if !hasRestrictedExtraActionCard {
				validOptions = append(validOptions, model.PromptOption{ID: "cannot_act", Label: "跳过额外行动"})
			}

			promptMessage := fmt.Sprint("请选择行动类型")
			if currentExtraAction == "Attack" {
				promptMessage = fmt.Sprint("当前为额外攻击行动，仅可执行攻击。请选择行动类型")
			} else if currentExtraAction == "Magic" {
				promptMessage = fmt.Sprint("当前为额外法术行动，仅可执行法术。请选择行动类型")
			} else if hasFighterHundredDragon {
				promptMessage = fmt.Sprint("你处于【百式幻龙拳】状态：本行动阶段仅可执行攻击。")
			} else if hasHeroTaunt {
				promptMessage = fmt.Sprintf("你受到【挑衅】影响：本次行动阶段必须且只能主动攻击 %s。", tauntSrcName)
			}
			if isRestrictedExtraAction && !hasRestrictedExtraActionCard {
				promptMessage = fmt.Sprint("当前为额外行动阶段，但你没有满足约束的可执行动作。可选择跳过本次额外行动。")
			}

			prompt := &model.Prompt{
				Type:           model.PromptConfirm,
				PlayerID:       currentPid,
				Message:        promptMessage,
				Options:        validOptions,
				SpecialOptions: specialOptions,
				UIMode:         model.PromptUIModeActionHub,
			}
			e.Notify(model.EventAskInput, "请选择行动类型", prompt)
			return

		case model.PhaseDiscardSelection:
			// 弃牌阶段应当伴随 PendingInterrupt(Discard)。
			// 若中断已被消费但阶段未恢复，修复到可继续推进的阶段，避免空转。
			if e.State.PendingInterrupt == nil {
				e.Log("[Warn] PhaseDiscardSelection: 无待处理中断，执行阶段修复")
				if len(e.State.PendingDamageQueue) > 0 {
					e.State.Phase = model.PhasePendingDamageResolution
					e.State.ReturnPhase = model.PhaseExtraAction
				} else if len(e.State.ActionQueue) > 0 {
					e.State.Phase = model.PhaseBeforeAction
				} else if len(e.State.CombatStack) > 0 {
					e.State.Phase = model.PhaseCombatInteraction
				} else {
					e.State.Phase = model.PhaseTurnEnd
				}
				continue
			}
			return

		case model.PhaseBeforeAction:
			// 4. 行动前阶段
			// 从队列中获取当前行动（不弹出，因为后续阶段可能需要使用）
			if len(e.State.ActionQueue) == 0 {
				e.Log("[Warn] PhaseBeforeAction: 行动队列为空，执行阶段修复")
				if len(e.State.PendingDamageQueue) > 0 {
					e.State.Phase = model.PhasePendingDamageResolution
					e.State.ReturnPhase = model.PhaseExtraAction
				} else {
					e.State.Phase = model.PhaseExtraAction
				}
				continue
			}

			currentAction := e.State.ActionQueue[0] // 只读取，不弹出
			if !queuedActionUsesVirtualCard(currentAction.SourceSkill) {
				if !e.repairQueuedActionCard(player, &e.State.ActionQueue[0]) {
					e.Log("[Warn] PhaseBeforeAction: 无法修复队列中的卡牌索引，丢弃该行动")
					e.State.ActionQueue = e.State.ActionQueue[1:]
					if len(e.State.PendingDamageQueue) > 0 {
						e.State.Phase = model.PhasePendingDamageResolution
						e.State.ReturnPhase = model.PhaseExtraAction
					} else if len(e.State.ActionQueue) > 0 {
						e.State.Phase = model.PhaseBeforeAction
					} else {
						e.State.Phase = model.PhaseExtraAction
					}
					continue
				}
				currentAction = e.State.ActionQueue[0]
			}

			// 获取目标（从 HandleAction 传入的 TargetID，需要存储）
			// 注意：这里我们需要从某个地方获取目标ID，可能需要修改 QueuedAction 结构
			// 暂时假设目标已经在某个地方存储了，或者从 ActionStack 中获取

			// 根据行动类型触发相应事件
			if currentAction.Type == model.ActionAttack {
				// 触发攻击开始事件
				targetID := currentAction.TargetID

				if targetID == "" {
					e.Log("[Error] 攻击行动缺少目标")
					return
				}

				target := e.State.Players[targetID]
				if target == nil {
					e.Log("[Error] 目标玩家不存在")
					return
				}

				// [新增] 先触发 TriggerOnCardUsed (封印等通用卡牌触发)
				if !e.State.ActionQueue[0].HasTriggeredCardUsed {
					// 技能转化攻击（如欺诈/多重射击）不消耗攻击牌，不触发 CardUsed。
					if queuedActionUsesVirtualCard(currentAction.SourceSkill) {
						e.State.ActionQueue[0].HasTriggeredCardUsed = true
					} else {
						// 1. 获取使用的卡牌 (用于事件触发)
						cardIdx := currentAction.CardIndex
						cardUsed, _, _, ok := getPlayableCardByIndex(player, cardIdx)
						if !ok {
							e.Log("[Warn] PhaseBeforeAction: 卡牌索引失效，丢弃该行动")
							e.State.ActionQueue = e.State.ActionQueue[1:]
							e.State.Phase = model.PhaseExtraAction
							continue
						}
						// 此时还未消耗，获取副本
						cardUsed = e.applyBlazeWitchAttackCardTransform(player, cardUsed)

						// 2. 触发 TriggerOnCardUsed
						cardCtx := &model.EventContext{
							Type:     model.EventCardUsed,
							Card:     &cardUsed,
							SourceID: currentPid,
							TargetID: targetID,
						}
						skillCtxUsed := e.buildContext(player, nil, model.TriggerOnCardUsed, cardCtx)
						e.dispatcher.OnTrigger(model.TriggerOnCardUsed, skillCtxUsed)

						// 标记已触发
						e.State.ActionQueue[0].HasTriggeredCardUsed = true

						// 3. 处理可能产生的延迟伤害 (即封印伤害)
						if e.processPendingDamages() {
							return // 有中断 (如伤害导致爆牌)，暂停 Drive
						}

						// 4. 处理可能产生的其他中断
						if e.State.PendingInterrupt != nil {
							return
						}
					}
				}

				if e.isMagicBow(player) && player.TurnState.UsedSkillCounts != nil {
					for i, pid := range e.State.PlayerOrder {
						if pid == targetID {
							player.TurnState.UsedSkillCounts["mb_last_attack_target_order"] = i + 1
							break
						}
					}
				}

				eventCtx := &model.EventContext{
					Type:     model.EventAttack,
					SourceID: currentPid,
					TargetID: targetID,
					Card:     currentAction.Card,
					AttackInfo: &model.AttackEventInfo{
						IsHit:            false,
						CanBeResponded:   true,
						ActionType:       string(model.ActionAttack),
						CounterInitiator: "",
					},
				}

				// 仅在本条攻击尚未触发过 AttackStart 时触发（确认响应技能后会再次进入此处，不再重复触发）
				var attackStartCtx *model.Context
				if !e.State.ActionQueue[0].HasTriggeredAttackStart {
					if player.Tokens == nil {
						player.Tokens = map[string]int{}
					}
					// 每次“新攻击”开始前重置圣枪骑士本次攻击临时标记。
					player.Tokens["holy_lancer_block_sacred_strike"] = 0
					player.Tokens["holy_lancer_sky_spear_no_counter"] = 0
					// 每次攻击开始前重置“单次攻击生效”的临时标记，避免串到后续攻击
					player.TurnState.GaleSlashActive = false
					player.TurnState.PreciseShotActive = false
					if player.Tokens == nil {
						player.Tokens = map[string]int{}
					}
					player.Tokens["berserker_blood_roar_ignore_shield"] = 0
					player.Tokens["assassin_stealth_attack_bonus"] = 0

					// 新一轮攻击开始前清理“命中后单次生效”标记。
					player.Tokens["ms_yellow_spring_pending"] = 0
					// 格斗家：每次新攻击开始前重置“同次攻击前置技能互斥锁”。
					player.Tokens["fighter_attack_start_skill_lock"] = 0
					e.State.ActionQueue[0].HasTriggeredAttackStart = true
					attackStartCtx = e.buildContext(player, target, model.TriggerOnAttackStart, eventCtx)
					player.TurnState.LastActionType = string(model.ActionAttack)
					e.dispatcher.OnTrigger(model.TriggerOnAttackStart, attackStartCtx)
					if e.State.PendingInterrupt != nil {
						return
					}
					if e.maybeTriggerMoonGoddessMedusa(player, target, currentAction.SourceSkill, currentAction.Card, attackStartCtx) {
						return
					}
				}

				// 无中断或已确认响应后：初始化战斗
				if e.isHero(player) && player.Tokens != nil && player.Tokens["hero_calm_force_no_counter"] > 0 {
					if eventCtx.AttackInfo != nil {
						eventCtx.AttackInfo.CanBeResponded = false
					}
					player.Tokens["hero_calm_force_no_counter"] = 0
				}
				if e.isFighter(player) && player.Tokens != nil && player.Tokens["fighter_qiburst_force_no_counter"] > 0 {
					if eventCtx.AttackInfo != nil {
						eventCtx.AttackInfo.CanBeResponded = false
					}
					player.Tokens["fighter_qiburst_force_no_counter"] = 0
				}
				if e.isMoonGoddess(player) && player.Tokens != nil && player.Tokens["mg_next_attack_no_counter"] > 0 {
					if eventCtx.AttackInfo != nil {
						eventCtx.AttackInfo.CanBeResponded = false
					}
					player.Tokens["mg_next_attack_no_counter"]--
					if player.Tokens["mg_next_attack_no_counter"] < 0 {
						player.Tokens["mg_next_attack_no_counter"] = 0
					}
				}
				// 暗杀者：潜行状态下的主动攻击无法应战。
				if isCharacter(player, "assassin") && player.HasFieldEffect(model.EffectStealth) {
					if eventCtx.AttackInfo != nil {
						eventCtx.AttackInfo.CanBeResponded = false
					}
					if player.Tokens == nil {
						player.Tokens = map[string]int{}
					}
					player.Tokens["assassin_stealth_attack_bonus"] = player.Gem + player.Crystal
					e.Log(fmt.Sprintf("[Skill] %s 处于[潜行]：本次主动攻击无法应战", player.Name))
				}
				// 圣枪骑士：天枪在攻击开始响应中生效，恢复主流程后需继续保持“无法应战”。
				if e.isHolyLancer(player) && player.Tokens != nil && player.Tokens["holy_lancer_sky_spear_no_counter"] > 0 {
					if eventCtx.AttackInfo != nil {
						eventCtx.AttackInfo.CanBeResponded = false
					}
					player.Tokens["holy_lancer_sky_spear_no_counter"] = 0
				}
				// 魔剑士：黄泉震颤生效后，本次攻击视为暗灭且无法应战。
				if e.isMagicSwordsman(player) && player.Tokens != nil && player.Tokens["ms_yellow_spring_pending"] > 0 {
					if eventCtx.AttackInfo != nil {
						eventCtx.AttackInfo.CanBeResponded = false
					}
					if currentAction.Card != nil {
						currentAction.Card.Element = model.ElementDark
					}
				}
				// 精灵射手：元素射击·雷之矢使本次攻击无法应战。
				if e.isElfArcher(player) && player.Tokens != nil && player.Tokens["elf_elemental_shot_thunder_pending"] > 0 {
					if eventCtx.AttackInfo != nil {
						eventCtx.AttackInfo.CanBeResponded = false
					}
				}
				// 暗灭：规则上不可应战（仅可承受或防御）。
				if eventCtx.AttackInfo != nil && currentAction.Card != nil && currentAction.Card.Element == model.ElementDark {
					eventCtx.AttackInfo.CanBeResponded = false
				}
				isForcedHit := false
				if eventCtx.AttackInfo != nil && eventCtx.AttackInfo.IsHitForced {
					isForcedHit = true
				}
				// 独有技兜底校验：标记存在但当前出牌不匹配时，强制失效。
				if player.TurnState.PreciseShotActive && !cardMatchesExclusiveSkill(player, currentAction.Card, "精准射击") {
					player.TurnState.PreciseShotActive = false
				}
				if player.TurnState.GaleSlashActive && !cardMatchesExclusiveSkill(player, currentAction.Card, "烈风技") {
					player.TurnState.GaleSlashActive = false
				}
				if player.TurnState.PreciseShotActive || player.TurnState.GaleSlashActive {
					isForcedHit = true
				}

				// 消耗卡牌（从手牌中移除）
				card := *currentAction.Card
				if !queuedActionUsesVirtualCard(currentAction.SourceSkill) {
					cardIdx := currentAction.CardIndex
					usedCard, err := consumePlayableCardByIndex(player, cardIdx)
					if err != nil {
						e.Log("[Warn] PhaseBeforeAction: 卡牌索引失效，丢弃该行动")
						e.State.Phase = model.PhaseExtraAction
						continue
					}
					card = usedCard
					_ = e.maybeAutoReleaseBloodPriestessByHand(player, "手牌<3强制脱离流血形态")
					card = e.applyBlazeWitchAttackCardTransform(player, card)
					if e.isMagicSwordsman(player) && player.Tokens != nil && player.Tokens["ms_yellow_spring_pending"] > 0 {
						card.Element = model.ElementDark
					}
					e.NotifyCardRevealed(currentPid, []model.Card{card}, "attack")
					e.State.DiscardPile = append(e.State.DiscardPile, card)
				}

				// 记录攻击行动次数
				player.TurnState.AttackCount += 1

				// 从队列中弹出行动（因为即将执行）
				e.State.ActionQueue = e.State.ActionQueue[1:]

				// 初始化战斗（使用实际卡牌，而不是队列中的指针）
				e.initCombat(currentPid, targetID, &card, isForcedHit, eventCtx.AttackInfo.CanBeResponded)
				break

			} else if currentAction.Type == model.ActionMagic {
				// 触发卡牌使用事件
				targetID := currentAction.TargetID
				if targetID == "" && len(currentAction.TargetIDs) > 0 {
					targetID = currentAction.TargetIDs[0]
				}

				if !e.State.ActionQueue[0].HasTriggeredCardUsed {
					cardCtx := &model.EventContext{
						Type:     model.EventCardUsed,
						Card:     currentAction.Card,
						SourceID: currentPid,
						TargetID: targetID,
					}

					skillCtx := e.buildContext(player, nil, model.TriggerOnCardUsed, cardCtx)

					// 触发卡牌使用事件
					e.dispatcher.OnTrigger(model.TriggerOnCardUsed, skillCtx)
					e.State.ActionQueue[0].HasTriggeredCardUsed = true

					// 如果触发了中断，等待用户输入
					if e.State.PendingInterrupt != nil {
						return
					}

					// 处理可能产生的延迟伤害（如封印），确保优先结算
					if e.processPendingDamages() {
						return
					}
					if e.State.PendingInterrupt != nil {
						return
					}
				}

				// 从队列中弹出行动
				e.State.ActionQueue = e.State.ActionQueue[1:]

				player.TurnState.LastActionType = string(model.ActionMagic)

				// 没有中断，执行法术逻辑
				// targetID 已经在上面计算过，包含了 TargetIDs[0] 的回退逻辑
				if err := e.PerformMagic(currentPid, targetID, currentAction.CardIndex); err != nil {
					e.Log(fmt.Sprintf("[Error] 法术执行失败: %v", err))
				}

				// 【新增检查】
				// 如果 PerformMagic 导致了中断 (比如触发了减伤技能)，
				// Phase 会被 ResolveDamage 改为 PhaseDamageResolution 或其他响应阶段。
				// 此时我们应该 break，让 Drive 处理中断，而不是强制跳到 ExtraAction
				if e.State.PendingInterrupt != nil {
					break
				}

				// 法术执行完毕，进入回合结束阶段
				if len(e.State.PendingDamageQueue) > 0 {
					e.State.Phase = model.PhasePendingDamageResolution
					e.State.ReturnPhase = model.PhaseTurnEnd
				} else {
					e.State.Phase = model.PhaseTurnEnd
				}
				break
			}

		case model.PhaseCombatInteraction:
			// 6. 战斗交互阶段（等待响应）
			if len(e.State.CombatStack) == 0 {
				e.Log("[Error] PhaseCombatInteraction: 战斗栈为空")
				return
			}

			// 查看栈顶战斗请求
			idx := len(e.State.CombatStack) - 1
			combatReq := &e.State.CombatStack[idx]
			target := e.State.Players[combatReq.TargetID]

			if target == nil {
				e.Log("[Error] PhaseCombatInteraction: 目标玩家不存在")
				return
			}

			// 阴阳师式神咒束：在响应阶段开始前，先检查是否触发“代应战”。
			if e.tryStartOnmyojiBindingInterrupt(combatReq) {
				return
			}
			// 若已完成式神咒束选择，自动执行“视为应战”流程。
			if e.executeOnmyojiBindingCounter(combatReq) {
				return
			}
			// 阴阳师阴阳转换：若目标阴阳师存在“同命格应战”机会，先询问是否发动。
			if e.tryStartOnmyojiYinYangInterrupt(combatReq) {
				return
			}

			// 如果强制命中，直接结算伤害
			if combatReq.IsForcedHit {
				e.Log(fmt.Sprintf("[Combat] 攻击强制命中！跳过响应阶段，直接结算..."))
				e.resolveCombatDamage(*combatReq)
				if atk := e.State.Players[combatReq.AttackerID]; atk != nil && atk.Tokens != nil {
					atk.Tokens["elf_elemental_shot_thunder_pending"] = 0
				}
				e.clearCombatStack()
				// 圣剑第3次攻击：强制命中分支也需要进入摸X弃X选择
				if e.triggerHolySwordDrawIfNeeded(e.State.Players[combatReq.AttackerID]) {
					return
				}
				e.State.Phase = model.PhaseExtraAction
				break
			}

			// 圣盾改为“承受伤害(take)时”再触发，先给玩家应战/防御的选择机会。
			shieldFallbackReady := e.hasUsableShieldForCombat(target, *combatReq)

			// 应战反弹目标：攻击方的队友（不含攻击者本人）
			var counterTargets []string
			attacker := e.State.Players[combatReq.AttackerID]
			if attacker != nil {
				for pid, p := range e.State.Players {
					if p.Camp == attacker.Camp && pid != combatReq.AttackerID {
						counterTargets = append(counterTargets, pid)
					}
				}
			}
			attackerRole := combatReq.AttackerID
			if attacker != nil {
				attackerRole = attacker.Name
			}

			// 通知目标玩家选择响应方式（无圣盾时正常选项）
			var options []model.PromptOption
			// 暗灭规则兜底：无论来源如何，暗灭攻击均不可应战。
			if combatReq.Card != nil && combatReq.Card.Element == model.ElementDark {
				combatReq.CanBeResponded = false
			}
			takeLabel := "承受伤害"
			if shieldFallbackReady {
				takeLabel = "承受（将触发圣盾）"
			}
			if combatReq.CanBeResponded {
				options = []model.PromptOption{
					{ID: "take", Label: takeLabel},
					{ID: "defend", Label: "防御"},
				}
				if len(counterTargets) > 0 {
					options = append(options, model.PromptOption{ID: "counter", Label: "应战"})
				}
			} else {
				options = []model.PromptOption{
					{ID: "take", Label: takeLabel},
					{ID: "defend", Label: "防御"},
				}
			}
			hints := e.buildCombatEffectHints(*combatReq, attacker)
			if shieldFallbackReady {
				hints = append(hints, "你身上有【圣盾】：若本次选择承受伤害，将优先消耗圣盾并抵挡本次攻击。")
			}

			prompt := &model.Prompt{
				Type:             model.PromptConfirm,
				PlayerID:         combatReq.TargetID,
				AttackerID:       combatReq.AttackerID,
				CounterTargetIDs: counterTargets,
				AttackElement:    string(combatReq.Card.Element), // 应战须同系或暗灭
				EffectHints:      hints,
				Message: fmt.Sprintf("%s 需要响应来自 %s 的攻击 (%s)",
					target.Name,
					attackerRole,
					combatReq.Card.Name),
				Options: options,
			}

			e.Notify(model.EventAskInput, "请选择响应方式", prompt)
			return // 等待用户输入

		case model.PhaseExtraAction:
			// 若上一次行动在 OnPhaseEnd 触发中断后返回，这里补做“行动结束追加效果”
			// （如迅捷赐福），避免被前置中断吞掉。
			if player.TurnState.LastActionType == "" && player.Tokens != nil && player.Tokens["post_action_end_effect_pending"] > 0 {
				actionType := model.ActionAttack
				if player.Tokens["post_action_end_effect_magic"] > 0 {
					actionType = model.ActionMagic
				}
				player.Tokens["post_action_end_effect_pending"] = 0
				player.Tokens["post_action_end_effect_magic"] = 0
				if e.handlePostActionEndEffects(player, actionType) {
					return
				}
			}

			if player.TurnState.LastActionType != "" {
				lastActionType := model.ActionType(player.TurnState.LastActionType)
				specialPhaseEndDispatched := false
				if player.Tokens != nil &&
					player.Tokens["special_phase_end_dispatched"] > 0 &&
					(lastActionType == model.ActionBuy || lastActionType == model.ActionSynthesize || lastActionType == model.ActionExtract) {
					specialPhaseEndDispatched = true
					player.Tokens["special_phase_end_dispatched"] = 0
				}
				eventCtx := &model.EventContext{
					Type:       model.EventPhaseEnd,
					SourceID:   currentPid,
					ActionType: lastActionType, // 告诉技能，刚才结束的是 Attack
				}
				if eventCtx.ActionType == model.ActionAttack {
					eventCtx.AttackInfo = &model.AttackEventInfo{
						ActionType:       string(model.ActionAttack),
						CounterInitiator: "",
					}
				}

				skillCtx := e.buildContext(player, nil, model.TriggerOnPhaseEnd, eventCtx)

				// 清除记录，防止死循环触发（非常重要！）
				player.TurnState.LastActionType = ""

				// 广播事件！
				// 此时 WindFuryHandler.CanUse 会被调用
				// 如果 CanUse 返回 true，Dispatcher 会根据 ResponseOptional 推送中断给用户
				// 特殊行动(Buy/Synthesize/Extract)在 ActionSelection 已完成过一次 OnPhaseEnd，
				// 这里跳过重复触发，避免被动结算两次。
				if !specialPhaseEndDispatched {
					e.dispatcher.OnTrigger(model.TriggerOnPhaseEnd, skillCtx)
				}

				// 如果触发了技能（产生了中断，比如用户需要确认是否发动风怒），直接 return 等待用户
				if e.State.PendingInterrupt != nil {
					if player.Tokens == nil {
						player.Tokens = map[string]int{}
					}
					player.Tokens["post_action_end_effect_pending"] = 1
					if lastActionType == model.ActionMagic {
						player.Tokens["post_action_end_effect_magic"] = 1
					} else {
						player.Tokens["post_action_end_effect_magic"] = 0
					}
					// 【重要】恢复 LastActionType，因为中断回来后还要处理 PhaseExtraAction
					// 但为了避免重复触发 EventPhaseEnd，我们需要一个标志位，或者让中断处理完直接进队列检查
					// 简单做法：中断回来后，Phase 依然是 ExtraAction，但 LastActionType 已被清空，所以不会二次触发
					return
				}
				// 行动结束后场上赐福结算（如迅捷赐福）
				if e.handlePostActionEndEffects(player, lastActionType) {
					return
				}
			}
			// 8. 额外行动阶段（处理队列）
			if len(e.State.ActionQueue) > 0 {
				// 弹出队列第一个行动
				queuedAction := e.State.ActionQueue[0]
				e.State.ActionQueue = e.State.ActionQueue[1:]

				// 设置当前额外行动约束
				player.TurnState.CurrentExtraAction = string(queuedAction.Type)
				if queuedAction.Element != "" {
					// 【修改点】将单个 Element 包装成切片
					player.TurnState.CurrentExtraElement = []model.Element{queuedAction.Element}
				} else {
					// 如果没有限制，置为 nil (或空切片)
					player.TurnState.CurrentExtraElement = nil
				}

				// 设置阶段为 BeforeAction
				e.State.Phase = model.PhaseBeforeAction
			} else {
				// 队列为空，进入回合结束
				e.State.Phase = model.PhaseTurnEnd
			}

		case model.PhaseTurnEnd:
			// 9. 回合结束阶段
			// 精灵密仪：回合结束时若无祝福则可转正并对任意角色造成2点法术伤害。
			if e.isElfArcher(player) && player.Tokens != nil && player.Tokens["elf_ritual_form"] > 0 {
				syncElfBlessings(player)
				if countElfBlessings(player) == 0 && player.Tokens["elf_ritual_release_waiting"] == 0 {
					player.Tokens["elf_ritual_release_waiting"] = 1
					e.PushInterrupt(&model.Interrupt{
						Type:     model.InterruptChoice,
						PlayerID: player.ID,
						Context: map[string]interface{}{
							"choice_type": "elf_ritual_release_target",
							"user_id":     player.ID,
							"target_ids":  append([]string{}, e.State.PlayerOrder...),
						},
					})
					return
				}
			}
			// 月之女神：回合结束时先触发【月之轮回】（在永恒乐章之前）。
			if e.maybeTriggerMoonGoddessMoonCycleAtTurnEnd(player) {
				return
			}
			// 吟游诗人：若当前回合角色持有永恒乐章，可触发胜利交响诗。
			if e.maybeTriggerBardVictoryAtTurnEnd(player) {
				return
			}
			// 血蔷薇庭院：血色剑灵回合结束时移回专属卡区（状态失效）。
			if e.isCrimsonSwordSpirit(player) && player.Tokens != nil && player.Tokens["css_rose_courtyard_active"] > 0 {
				player.Tokens["css_rose_courtyard_active"] = 0
				player.Tokens["css_blood_cap"] = 3
				if player.Tokens["css_blood"] > 3 {
					player.Tokens["css_blood"] = 3
				}
				if e.returnRoseCourtyardToExclusive(player) {
					e.Log(fmt.Sprintf("%s 的 [血蔷薇庭院] 回合结束移回专属卡区", player.Name))
				} else {
					e.Log(fmt.Sprintf("%s 的 [血蔷薇庭院] 回合结束失效", player.Name))
				}
			}
			// 红莲骑士：回合结束时重置热血沸腾形态并+2治疗。
			e.resolveCrimsonKnightHotFormTurnEnd(player)
			// 英灵人形：符文改造形态在回合结束时自动转正，并按新手牌上限检查弃牌。
			if e.isWarHomunculus(player) && player.Tokens != nil && player.Tokens["hom_burst_form"] > 0 {
				player.Tokens["hom_burst_form"] = 0
				e.Log(fmt.Sprintf("%s 的 [符文改造] 效果结束，脱离蓄势迸发形态", player.Name))
				e.checkHandLimit(player, nil)
				if e.State.PendingInterrupt != nil {
					return
				}
			}
			// 格斗家：百式幻龙拳持续至回合结束，随后转正。
			if e.isFighter(player) && player.Tokens != nil && player.Tokens["fighter_hundred_dragon_form"] > 0 {
				player.Tokens["fighter_hundred_dragon_form"] = 0
				player.Tokens["fighter_hundred_dragon_target_order"] = 0
				e.Log(fmt.Sprintf("%s 的 [百式幻龙拳] 回合结束，效果取消并转正", player.Name))
			}
			// 阴阳师：回合结束时若鬼火达到上限，触发黑暗祭礼。
			if e.isOnmyoji(player) && player.Tokens != nil && player.Tokens["onmyoji_ghost_fire"] >= 3 {
				var targetIDs []string
				for _, pid := range e.State.PlayerOrder {
					p := e.State.Players[pid]
					if p != nil {
						targetIDs = append(targetIDs, p.ID)
					}
				}
				if len(targetIDs) > 0 {
					e.PushInterrupt(&model.Interrupt{
						Type:     model.InterruptChoice,
						PlayerID: player.ID,
						Context: map[string]interface{}{
							"choice_type": "onmyoji_dark_ritual_target",
							"user_id":     player.ID,
							"target_ids":  targetIDs,
							"ghost_fire":  player.Tokens["onmyoji_ghost_fire"],
						},
					})
					e.Log(fmt.Sprintf("%s 的 [黑暗祭礼] 触发，等待选择2点法术伤害目标", player.Name))
					return
				}
			}
			// 检查是否有待执行的行动令牌 (处理额外行动)
			// 将PendingActions逻辑迁移至此
			if len(player.TurnState.PendingActions) > 0 {
				// 取出第一个行动令牌
				currentAction := player.TurnState.PendingActions[0]
				player.TurnState.PendingActions = player.TurnState.PendingActions[1:]

				// 重置行动状态，允许再次行动
				player.TurnState.HasActed = false

				// 设置 TurnState 中的约束，然后调用 Drive (进入 ActionSelection)
				player.TurnState.CurrentExtraAction = currentAction.MustType
				player.TurnState.CurrentExtraElement = currentAction.MustElement

				e.State.Phase = model.PhaseActionSelection

				// 显示行动约束信息
				constraintInfo := e.buildConstraintInfo(currentAction.MustType, currentAction.MustElement)
				e.Log(fmt.Sprintf("[Turn] %s %s 额外行动开始 (剩余 %d 次额外行动)%s",
					player.Name, currentAction.Source, len(player.TurnState.PendingActions)+1, constraintInfo))

				e.Drive() // 继续 Drive 以进入 PhaseActionSelection 并发送提示
				return
			}

			// 圣弓：回合结束时若未执行特殊行动，可触发自动填充（每回合最多一次）。
			if e.isHolyBow(player) && player.Tokens != nil && player.Tokens["hb_auto_fill_done_turn"] <= 0 {
				if player.Tokens["hb_special_used_turn"] <= 0 {
					var resourceModes []string
					if e.CanPayCrystalCost(player.ID, 1) {
						resourceModes = append(resourceModes, "crystal")
					}
					if player.Gem > 0 {
						resourceModes = append(resourceModes, "gem")
					}
					if len(resourceModes) > 0 {
						player.Tokens["hb_auto_fill_done_turn"] = 1
						e.PushInterrupt(&model.Interrupt{
							Type:     model.InterruptChoice,
							PlayerID: player.ID,
							Context: map[string]interface{}{
								"choice_type":    "hb_auto_fill_resource",
								"user_id":        player.ID,
								"resource_modes": resourceModes,
							},
						})
						e.Log(fmt.Sprintf("%s 的 [自动填充] 触发：请选择消耗资源与增益", player.Name))
						return
					}
				}
				player.Tokens["hb_auto_fill_done_turn"] = 1
			}

			// 圣光祈愈“本回合已用”标记：仅在真正结束当前回合时清理。
			if player.Tokens != nil {
				player.Tokens["holy_lancer_prayer_used_turn"] = 0
			}

			e.NextTurn()

		case model.PhaseActionExecution:
			// 行动执行阶段通常用于“行动中弹出的中断”（如魔弹融合/圣疗等）。
			// 当中断被消费后，如果没有显式阶段回切，这里负责把流程接回主状态机，
			// 避免停在 ActionExecution 导致 Drive 直接返回而卡局。
			if len(e.State.PendingDamageQueue) > 0 {
				e.State.Phase = model.PhasePendingDamageResolution
				if e.State.ReturnPhase == "" {
					e.State.ReturnPhase = model.PhaseExtraAction
				}
				continue
			}
			if len(e.State.CombatStack) > 0 {
				e.State.Phase = model.PhaseCombatInteraction
				continue
			}
			if len(e.State.ActionQueue) > 0 {
				e.State.Phase = model.PhaseBeforeAction
				continue
			}
			e.State.Phase = model.PhaseExtraAction
			continue

		default:
			// 其他阶段（如 PhaseAction, PhaseResponse 等）由 HandleAction 处理
			// Drive 不处理这些阶段
			return
		}
	}
}

// resolveCrimsonKnightHotFormTurnEnd 统一处理红莲骑士“热血沸腾”在回合结束时的退形态逻辑。
// 返回 true 表示本次确实触发了退形态与治疗。
func (e *GameEngine) resolveCrimsonKnightHotFormTurnEnd(player *model.Player) bool {
	if player == nil || !e.isCrimsonKnight(player) {
		return false
	}
	if player.Tokens == nil || player.Tokens["crk_hot_form"] <= 0 {
		return false
	}
	player.Tokens["crk_hot_form"] = 0
	e.Heal(player.ID, 2)
	e.Log(fmt.Sprintf("%s 回合结束脱离 [热血沸腾形态]，获得2点治疗", player.Name))
	return true
}

// NextTurn 结束当前回合并开始下一回合
func (e *GameEngine) NextTurn() {
	// Guard against turn progression during interrupt phases
	if e.State.PendingInterrupt != nil {
		return // Silently prevent turn progression during interrupts
	}
	if e.actionSummaryTurn <= 0 {
		e.actionSummaryTurn = 1
	} else {
		e.actionSummaryTurn++
	}

	currentPid := e.State.PlayerOrder[e.State.CurrentTurn]
	player := e.State.Players[currentPid]

	// 兜底：若有路径直接调用 NextTurn 而跳过 PhaseTurnEnd，
	// 仍需保证红莲骑士“热血沸腾”在回合结束时正确退形态并+2治疗。
	e.resolveCrimsonKnightHotFormTurnEnd(player)

	extraTurn := false
	if e.isMoonGoddess(player) && player.Tokens != nil && player.Tokens["mg_extra_turn_pending"] > 0 {
		player.Tokens["mg_extra_turn_pending"]--
		if player.Tokens["mg_extra_turn_pending"] < 0 {
			player.Tokens["mg_extra_turn_pending"] = 0
		}
		extraTurn = true
	}

	// 真正的回合结束：清理当前玩家状态
	player.IsActive = false
	e.Log(fmt.Sprintf("[Turn] %s 结束回合", player.Name))

	// 切换到下一个玩家（若有额外回合则保持当前玩家）
	nextPid := currentPid
	if !extraTurn {
		e.State.CurrentTurn = (e.State.CurrentTurn + 1) % len(e.State.PlayerOrder)
		nextPid = e.State.PlayerOrder[e.State.CurrentTurn]
	} else {
		e.Log(fmt.Sprintf("%s 的 [苍白之月] 生效：立即获得额外回合", player.Name))
	}
	nextPlayer := e.State.Players[nextPid]

	// 初始化新回合玩家状态
	nextPlayer.IsActive = true
	nextPlayer.TurnState = model.NewPlayerTurnState()
	for _, p := range e.State.Players {
		if p == nil || p.Tokens == nil {
			continue
		}
		p.Tokens["mb_magic_pierce_pending"] = 0
		p.Tokens["fighter_attack_start_skill_lock"] = 0
		p.Tokens["fighter_charge_pending"] = 0
		p.Tokens["fighter_charge_damage_pending"] = 0
		p.Tokens["fighter_qiburst_force_no_counter"] = 0
		// 吟游诗人：回合级触发标记按“当前回合”重置。
		p.Tokens["bd_descent_used_turn"] = 0
		p.Tokens["hb_shard_miss_pending"] = 0
		p.Tokens["hb_auto_fill_done_turn"] = 0
		p.Tokens["mg_blasphemy_used_turn"] = 0
		p.Tokens["mg_blasphemy_pending"] = 0
		p.Tokens["bp_bleed_tick_done_turn"] = 0
		p.Tokens["bt_wither_pending"] = 0
	}
	e.resetTurnMagicDamageTracker()
	// 重置 Engine 状态
	e.State.HasPerformedStartup = false
	e.State.ActionQueue = []model.QueuedAction{}
	e.State.CombatStack = []model.CombatRequest{}

	e.Log(fmt.Sprintf("[Turn] %s 回合开始 (Hand:%d Gem:%d Cry:%d)",
		nextPlayer.Name, len(nextPlayer.Hand), nextPlayer.Gem, nextPlayer.Crystal))

	// 设置阶段为第1步：Buff结算
	e.State.Phase = model.PhaseBuffResolve

}

// handleBuy 购买行动：摸3牌，战绩区+1宝石+1水晶（规则：战绩区上限5，满则不加）
func (e *GameEngine) handleBuy(p *model.Player) error {
	maxHand := e.GetMaxHand(p)
	if len(p.Hand)+3 > maxHand {
		return fmt.Errorf("购买后手牌将超过上限(%d+3>%d)，无法购买", len(p.Hand), maxHand)
	}

	e.drawForAction(p, 3)
	// 战绩区+1宝石+1水晶（阵营资源，非个人能量）
	// 规则：战绩区已有4个星石时，可选择添加宝石或水晶
	const maxStones = 5
	var stones int
	if p.Camp == model.RedCamp {
		stones = e.State.RedGems + e.State.RedCrystals
	} else {
		stones = e.State.BlueGems + e.State.BlueCrystals
	}
	if stones >= maxStones {
		e.Log(fmt.Sprintf("[Action] %s 购买：摸3牌，战绩区已满不加星石", p.Name))
		return nil
	}
	if stones == 4 {
		// 战绩区4个星石，玩家选择添加宝石或水晶
		e.PushInterrupt(&model.Interrupt{
			Type:     model.InterruptChoice,
			PlayerID: p.ID,
			Context: map[string]interface{}{
				"choice_type": "buy_resource",
				"camp":        string(p.Camp),
			},
		})
		e.Log(fmt.Sprintf("[Action] %s 购买：摸3牌，战绩区4星石，请选择添加宝石或水晶", p.Name))
		return nil
	}
	// 0~3 个星石：添加 1 宝石 + 1 水晶
	if p.Camp == model.RedCamp {
		if e.State.RedGems+e.State.RedCrystals < maxStones {
			e.State.RedGems++
		}
		if e.State.RedGems+e.State.RedCrystals < maxStones {
			e.State.RedCrystals++
		}
	} else {
		if e.State.BlueGems+e.State.BlueCrystals < maxStones {
			e.State.BlueGems++
		}
		if e.State.BlueGems+e.State.BlueCrystals < maxStones {
			e.State.BlueCrystals++
		}
	}
	e.Log(fmt.Sprintf("[Action] %s 购买：摸3牌，战绩区+1宝石+1水晶", p.Name))
	return nil
}

// handleSynthesize 合成行动
func (e *GameEngine) handleSynthesize(p *model.Player) error {
	maxHand := e.GetMaxHand(p)
	if len(p.Hand)+3 > maxHand {
		return fmt.Errorf("合成后手牌将超过上限(%d+3>%d)，无法合成", len(p.Hand), maxHand)
	}

	// 合成消耗战绩区 3 星石（非个人能量）
	var totalStones int
	if p.Camp == model.RedCamp {
		totalStones = e.State.RedGems + e.State.RedCrystals
	} else {
		totalStones = e.State.BlueGems + e.State.BlueCrystals
	}
	if totalStones < 3 {
		return errors.New("战绩区星石不足3个，无法合成")
	}
	e.drawForAction(p, 3)
	// 从战绩区扣除 3 星石（优先扣宝石）
	cost := 3
	if p.Camp == model.RedCamp {
		if e.State.RedGems >= cost {
			e.State.RedGems -= cost
		} else {
			cost -= e.State.RedGems
			e.State.RedGems = 0
			e.State.RedCrystals -= cost
		}
	} else {
		if e.State.BlueGems >= cost {
			e.State.BlueGems -= cost
		} else {
			cost -= e.State.BlueGems
			e.State.BlueGems = 0
			e.State.BlueCrystals -= cost
		}
	}
	// 合成星杯：星杯+1，对方士气-1
	if p.Camp == model.RedCamp {
		e.State.RedCups++
		if e.State.RedCups > 5 {
			e.State.RedCups = 5
		}
		e.Log(fmt.Sprintf("[Action] %s 合成星杯！红方星杯+1，蓝方士气-1", p.Name))
		e.State.BlueMorale--
	} else {
		e.State.BlueCups++
		if e.State.BlueCups > 5 {
			e.State.BlueCups = 5
		}
		e.Log(fmt.Sprintf("[Action] %s 合成星杯！蓝方星杯+1，红方士气-1", p.Name))
		e.State.RedMorale--
	}
	e.checkGameEnd()
	return nil
}

// handleExtract 提取行动：展示战绩区所有星石，让玩家选择 1-2 个提炼到能量区
func (e *GameEngine) handleExtract(p *model.Player) error {
	e.clearAdventurerExtractState(p)

	currentEnergy := p.Gem + p.Crystal
	maxEnergy := e.getPlayerEnergyCap(p)

	var availableGems, availableCrystals int
	if p.Camp == model.RedCamp {
		availableGems = e.State.RedGems
		availableCrystals = e.State.RedCrystals
	} else {
		availableGems = e.State.BlueGems
		availableCrystals = e.State.BlueCrystals
	}

	totalAvailable := availableGems + availableCrystals
	if totalAvailable == 0 {
		return errors.New("阵营资源池中没有可提取的资源")
	}

	energyRoom := maxEnergy - currentEnergy
	allowParadise := e.playerHasSkill(p, "adventurer_paradise")
	maxAllyRoom := 0
	if allowParadise {
		maxAllyRoom = e.maxAllyEnergyRoom(p)
	}
	maxRecipientRoom := energyRoom
	if maxAllyRoom > maxRecipientRoom {
		maxRecipientRoom = maxAllyRoom
	}
	if maxRecipientRoom <= 0 {
		return errors.New("能量已达上限，且没有可承接提炼能量的队友")
	}

	// 构建战绩区所有星石列表（逐个展示，便于玩家选择）
	var opts []interface{}
	for i := 0; i < availableGems; i++ {
		opts = append(opts, map[string]interface{}{"type": "gem"})
	}
	for i := 0; i < availableCrystals; i++ {
		opts = append(opts, map[string]interface{}{"type": "crystal"})
	}

	maxSelect := 2
	if maxRecipientRoom < maxSelect {
		maxSelect = maxRecipientRoom
	}
	if totalAvailable < maxSelect {
		maxSelect = totalAvailable
	}

	e.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: p.ID,
		Context: map[string]interface{}{
			"choice_type":            "extract",
			"extract_options":        opts,
			"extract_min":            1,
			"extract_max":            maxSelect,
			"extract_self_room":      energyRoom,
			"extract_max_ally_room":  maxAllyRoom,
			"extract_allow_paradise": allowParadise,
		},
	})
	if allowParadise && maxAllyRoom > 0 {
		e.Log(fmt.Sprintf("[Action] %s 提炼：战绩区有 %d 红宝石 %d 蓝水晶，请选择 1-%d 个提炼（可通过冒险者天堂转移给队友）", p.Name, availableGems, availableCrystals, maxSelect))
	} else {
		e.Log(fmt.Sprintf("[Action] %s 提炼：战绩区有 %d 红宝石 %d 蓝水晶，请选择 1-%d 个提炼", p.Name, availableGems, availableCrystals, maxSelect))
	}
	return nil
}

// handleExtractChoiceResponse 处理提炼多选响应
func (e *GameEngine) handleExtractChoiceResponse(act model.PlayerAction) error {
	if e.State.PendingInterrupt == nil {
		return fmt.Errorf("没有待处理的中断")
	}
	data, ok := e.State.PendingInterrupt.Context.(map[string]interface{})
	if !ok {
		return fmt.Errorf("中断上下文格式错误")
	}
	player := e.State.Players[act.PlayerID]
	if player == nil {
		return fmt.Errorf("玩家不存在")
	}
	optsRaw, _ := data["extract_options"]
	optsIfaces, ok := optsRaw.([]interface{})
	if !ok || len(act.Selections) == 0 {
		return fmt.Errorf("请选择要提炼的星石")
	}
	minSel, _ := data["extract_min"].(int)
	maxSel, _ := data["extract_max"].(int)
	if minSel < 1 {
		minSel = 1
	}
	if maxSel < 1 {
		maxSel = 2
	}
	if len(act.Selections) < minSel || len(act.Selections) > maxSel {
		return fmt.Errorf("请选择 %d-%d 个星石提炼", minSel, maxSel)
	}

	extractedGems := 0
	extractedCrystals := 0
	seen := make(map[int]bool)
	for _, sel := range act.Selections {
		idx := sel
		if idx < 0 || idx >= len(optsIfaces) || seen[idx] {
			return fmt.Errorf("无效的提炼选择")
		}
		seen[idx] = true
		om, _ := optsIfaces[idx].(map[string]interface{})
		if om == nil {
			return fmt.Errorf("提炼选项格式错误")
		}
		typ, _ := om["type"].(string)
		if typ == "gem" {
			extractedGems++
		} else if typ == "crystal" {
			extractedCrystals++
		}
	}

	selfRoom := toIntContextValue(data["extract_self_room"])
	if selfRoom < 0 {
		selfRoom = 0
	}
	maxAllyRoom := toIntContextValue(data["extract_max_ally_room"])
	allowParadise, _ := data["extract_allow_paradise"].(bool)
	totalExtracted := extractedGems + extractedCrystals
	requiresParadise := totalExtracted > selfRoom
	if requiresParadise && (!allowParadise || maxAllyRoom < totalExtracted) {
		return fmt.Errorf("本次提炼超出自身能量上限，且没有可承接的队友")
	}

	if player.Camp == model.RedCamp {
		if extractedGems > e.State.RedGems || extractedCrystals > e.State.RedCrystals {
			return fmt.Errorf("战绩区星石不足")
		}
		e.State.RedGems -= extractedGems
		e.State.RedCrystals -= extractedCrystals
	} else {
		if extractedGems > e.State.BlueGems || extractedCrystals > e.State.BlueCrystals {
			return fmt.Errorf("战绩区星石不足")
		}
		e.State.BlueGems -= extractedGems
		e.State.BlueCrystals -= extractedCrystals
	}
	e.recordAdventurerExtractResult(player, extractedGems, extractedCrystals, requiresParadise)
	if requiresParadise {
		e.Log(fmt.Sprintf("[Action] %s 提炼：获得 %d 宝石 %d 水晶，等待通过冒险者天堂分配给队友",
			player.Name, extractedGems, extractedCrystals))
	} else {
		player.Gem += extractedGems
		player.Crystal += extractedCrystals
		e.Log(fmt.Sprintf("[Action] %s 提炼：从战绩区获得 %d 宝石 %d 水晶（当前能量: %d）",
			player.Name, extractedGems, extractedCrystals, player.Gem+player.Crystal))
	}

	e.PopInterrupt()
	if e.State.PendingInterrupt == nil {
		e.State.Phase = model.PhaseTurnEnd
	}
	return nil
}

func (e *GameEngine) playerHasSkill(p *model.Player, skillID string) bool {
	if p == nil || p.Character == nil {
		return false
	}
	for _, s := range p.Character.Skills {
		if s.ID == skillID {
			return true
		}
	}
	return false
}

func (e *GameEngine) getPlayerEnergyCap(player *model.Player) int {
	if player == nil {
		return 3
	}
	cap := 3
	if e.isSage(player) {
		cap++
	}
	return cap
}

func (e *GameEngine) maxAllyEnergyRoom(p *model.Player) int {
	if p == nil {
		return 0
	}
	maxRoom := 0
	for _, ally := range e.State.Players {
		if ally == nil || ally.Camp != p.Camp || ally.ID == p.ID {
			continue
		}
		maxEnergy := e.getPlayerEnergyCap(ally)
		room := maxEnergy - (ally.Gem + ally.Crystal)
		if room > maxRoom {
			maxRoom = room
		}
	}
	return maxRoom
}

func (e *GameEngine) clearAdventurerExtractState(p *model.Player) {
	if p == nil {
		return
	}
	if p.Tokens == nil {
		p.Tokens = map[string]int{}
	}
	p.Tokens["adventurer_extract_last_gem"] = 0
	p.Tokens["adventurer_extract_last_crystal"] = 0
	p.Tokens["adventurer_extract_requires_paradise"] = 0
}

func (e *GameEngine) recordAdventurerExtractResult(p *model.Player, gem, crystal int, requiresParadise bool) {
	if p == nil {
		return
	}
	if p.Tokens == nil {
		p.Tokens = map[string]int{}
	}
	p.Tokens["adventurer_extract_last_gem"] = gem
	p.Tokens["adventurer_extract_last_crystal"] = crystal
	if requiresParadise {
		p.Tokens["adventurer_extract_requires_paradise"] = 1
	} else {
		p.Tokens["adventurer_extract_requires_paradise"] = 0
	}
}

func toIntContextValue(v interface{}) int {
	if i, ok := v.(int); ok {
		return i
	}
	if f, ok := v.(float64); ok {
		return int(f)
	}
	return 0
}

func toBoolContextValue(v interface{}) bool {
	if b, ok := v.(bool); ok {
		return b
	}
	return false
}

func isMagicLikeDamageType(damageType string) bool {
	return !strings.EqualFold(damageType, "Attack")
}

func (e *GameEngine) resetTurnMagicDamageTracker() {
	if e.turnMagicDamageTargets == nil {
		e.turnMagicDamageTargets = map[model.Camp]map[string]bool{}
	}
	e.turnMagicDamageTargets[model.RedCamp] = map[string]bool{}
	e.turnMagicDamageTargets[model.BlueCamp] = map[string]bool{}
}

func buildElementCardIndexMap(player *model.Player) map[model.Element][]int {
	out := map[model.Element][]int{}
	if player == nil {
		return out
	}
	for i, c := range player.Hand {
		if c.Element == "" {
			continue
		}
		out[c.Element] = append(out[c.Element], i)
	}
	return out
}

func maxSameElementCount(player *model.Player) int {
	maxCount := 0
	for _, idxs := range buildElementCardIndexMap(player) {
		if len(idxs) > maxCount {
			maxCount = len(idxs)
		}
	}
	return maxCount
}

func distinctElementCount(player *model.Player) int {
	return len(buildElementCardIndexMap(player))
}

func elementOrderForPrompt() []model.Element {
	return []model.Element{
		model.ElementEarth,
		model.ElementWater,
		model.ElementFire,
		model.ElementWind,
		model.ElementThunder,
		model.ElementLight,
		model.ElementDark,
	}
}

func availableElementsByMinCount(player *model.Player, minCount int) []string {
	if minCount <= 0 {
		minCount = 1
	}
	elemMap := buildElementCardIndexMap(player)
	var out []string
	for _, ele := range elementOrderForPrompt() {
		if len(elemMap[ele]) >= minCount {
			out = append(out, string(ele))
		}
	}
	return out
}

func allHandIndices(player *model.Player) []int {
	if player == nil {
		return nil
	}
	out := make([]int, 0, len(player.Hand))
	for i := range player.Hand {
		out = append(out, i)
	}
	return out
}

func removeElementIndices(indices []int, player *model.Player, element model.Element, keepIndex int) []int {
	if len(indices) == 0 {
		return nil
	}
	var out []int
	for _, idx := range indices {
		if idx == keepIndex {
			continue
		}
		if idx < 0 || player == nil || idx >= len(player.Hand) {
			continue
		}
		if player.Hand[idx].Element == element {
			continue
		}
		out = append(out, idx)
	}
	return out
}

func parseStringSliceContextValue(v interface{}) []string {
	var out []string
	switch arr := v.(type) {
	case []string:
		out = append(out, arr...)
	case []interface{}:
		for _, item := range arr {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
	}
	return out
}

func dedupeIDs(ids []string) []string {
	if len(ids) == 0 {
		return nil
	}
	out := make([]string, 0, len(ids))
	seen := map[string]bool{}
	for _, id := range ids {
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		out = append(out, id)
	}
	return out
}

func idsToSet(ids []string) map[string]bool {
	set := make(map[string]bool, len(ids))
	for _, id := range ids {
		if id == "" {
			continue
		}
		set[id] = true
	}
	return set
}

func (e *GameEngine) campEnemyIDs(camp model.Camp) []string {
	var ids []string
	for _, pid := range e.State.PlayerOrder {
		p := e.State.Players[pid]
		if p == nil || p.Camp == camp {
			continue
		}
		ids = append(ids, p.ID)
	}
	return ids
}

// 吟游诗人：记录“当前回合我方已对哪些敌方角色造成过法术伤害”，并在满足条件时触发沉沦协奏曲。
func (e *GameEngine) tryTriggerBardDescentAfterMagicDamage(pd *model.PendingDamage) bool {
	if pd == nil || pd.Damage <= 0 {
		return false
	}
	source := e.State.Players[pd.SourceID]
	target := e.State.Players[pd.TargetID]
	if source == nil || target == nil || source.Camp == target.Camp {
		return false
	}

	if e.turnMagicDamageTargets == nil {
		e.resetTurnMagicDamageTracker()
	}
	if _, ok := e.turnMagicDamageTargets[source.Camp]; !ok {
		e.turnMagicDamageTargets[source.Camp] = map[string]bool{}
	}
	e.turnMagicDamageTargets[source.Camp][target.ID] = true
	if len(e.turnMagicDamageTargets[source.Camp]) < 2 {
		return false
	}

	for _, pid := range e.State.PlayerOrder {
		bard := e.State.Players[pid]
		if bard == nil || !e.isBard(bard) || bard.Camp != source.Camp {
			continue
		}
		if bard.Tokens == nil {
			bard.Tokens = map[string]int{}
		}
		// 仅普通形态，且每回合仅触发一次选择机会。
		if bard.Tokens["bd_prisoner_form"] > 0 || bard.Tokens["bd_descent_used_turn"] > 0 {
			continue
		}
		if bardMaxSameElementCount(bard) < 2 {
			continue
		}
		enemyIDs := e.campEnemyIDs(bard.Camp)
		if len(enemyIDs) == 0 {
			continue
		}
		e.PushInterrupt(&model.Interrupt{
			Type:     model.InterruptChoice,
			PlayerID: bard.ID,
			Context: map[string]interface{}{
				"choice_type": "bd_descent_confirm",
				"user_id":     bard.ID,
				"target_ids":  enemyIDs,
			},
		})
		e.Log(fmt.Sprintf("%s 满足 [沉沦协奏曲] 触发条件，可选择是否发动", bard.Name))
		return true
	}
	return false
}

func (e *GameEngine) maybeTriggerBardRousingAtTurnStart(current *model.Player) bool {
	if current == nil {
		return false
	}
	for _, pid := range e.State.PlayerOrder {
		bard := e.State.Players[pid]
		if bard == nil || !e.isBard(bard) {
			continue
		}
		holderID := e.bardEternalHolderID(bard)
		if holderID == "" || holderID != current.ID {
			continue
		}
		enemyIDs := e.campEnemyIDs(bard.Camp)
		if len(enemyIDs) == 0 {
			continue
		}
		e.PushInterrupt(&model.Interrupt{
			Type:     model.InterruptChoice,
			PlayerID: bard.ID,
			Context: map[string]interface{}{
				"choice_type": "bd_rousing_mode",
				"user_id":     bard.ID,
				"holder_id":   current.ID,
				"target_ids":  enemyIDs,
			},
		})
		e.Log(fmt.Sprintf("%s 持有永恒乐章，%s 可发动 [激昂狂想曲]", current.Name, bard.Name))
		return true
	}
	return false
}

func (e *GameEngine) maybeTriggerBardVictoryAtTurnEnd(current *model.Player) bool {
	if current == nil {
		return false
	}
	for _, pid := range e.State.PlayerOrder {
		bard := e.State.Players[pid]
		if bard == nil || !e.isBard(bard) {
			continue
		}
		holderID := e.bardEternalHolderID(bard)
		if holderID == "" || holderID != current.ID {
			continue
		}
		e.PushInterrupt(&model.Interrupt{
			Type:     model.InterruptChoice,
			PlayerID: bard.ID,
			Context: map[string]interface{}{
				"choice_type": "bd_victory_mode",
				"user_id":     bard.ID,
				"holder_id":   current.ID,
			},
		})
		e.Log(fmt.Sprintf("%s 持有永恒乐章，%s 可发动 [胜利交响诗]", current.Name, bard.Name))
		return true
	}
	return false
}

func (e *GameEngine) resolveBardForbiddenVerseAfterSong(bard *model.Player, songName string) {
	if bard == nil || !e.isBard(bard) {
		return
	}
	if bard.Tokens == nil {
		bard.Tokens = map[string]int{}
	}
	if bardInspiration(bard) < bardInspirationCapEngine {
		now := addBardInspiration(bard, 1)
		removed := e.removeBardEternalMovement(bard)
		if removed {
			e.Log(fmt.Sprintf("%s 的 [禁忌诗篇] 生效：灵感+1（当前%d），并移除永恒乐章", bard.Name, now))
		} else {
			e.Log(fmt.Sprintf("%s 的 [禁忌诗篇] 生效：灵感+1（当前%d）", bard.Name, now))
		}
		return
	}

	if bard.Tokens["bd_prisoner_form"] <= 0 {
		bard.Tokens["bd_prisoner_form"] = 1
		e.Log(fmt.Sprintf("%s 的 [禁忌诗篇] 生效：转为永恒囚徒形态", bard.Name))
	}
	e.AddPendingDamage(model.PendingDamage{
		SourceID:   bard.ID,
		TargetID:   bard.ID,
		Damage:     3,
		DamageType: "magic",
		Stage:      0,
	})
	e.Log(fmt.Sprintf("%s 的 [禁忌诗篇] 生效：灵感已满，对自己造成3点法术伤害（来源：%s）", bard.Name, songName))
}

func (e *GameEngine) bardAlliesExcluding(camp model.Camp, excludeID string) []string {
	var ids []string
	for _, pid := range e.State.PlayerOrder {
		p := e.State.Players[pid]
		if p == nil || p.Camp != camp || p.ID == excludeID {
			continue
		}
		ids = append(ids, p.ID)
	}
	return ids
}

func (e *GameEngine) enqueueDeferredFollowup(f model.DeferredFollowup) {
	e.State.DeferredFollowups = append(e.State.DeferredFollowups, f)
}

func (e *GameEngine) processDeferredFollowups() bool {
	if len(e.State.DeferredFollowups) == 0 {
		return false
	}
	// 逐条出队，避免同一后续被重复执行。
	f := e.State.DeferredFollowups[0]
	e.State.DeferredFollowups = e.State.DeferredFollowups[1:]
	switch f.Type {
	case "spirit_caster_talisman":
		if err := e.resolveSpiritCasterTalismanFollowup(f); err != nil {
			e.Log(fmt.Sprintf("[SpiritCaster] 延迟结算失败: %v", err))
		}
	case "blood_priestess_shared_life_place":
		if err := e.resolveBloodPriestessSharedLifePlaceFollowup(f); err != nil {
			e.Log(fmt.Sprintf("[BloodPriestess] 同生共死延迟放置失败: %v", err))
		}
	case "blood_priestess_wail_damage":
		if err := e.resolveBloodPriestessWailDamageFollowup(f); err != nil {
			e.Log(fmt.Sprintf("[BloodPriestess] 血之悲鸣延迟伤害失败: %v", err))
		}
	default:
		e.Log(fmt.Sprintf("[Warn] 未知的延迟后续类型: %s", f.Type))
	}
	return true
}

func (e *GameEngine) beginSpiritCasterTalisman(user *model.Player, skillID string, targetIDs []string, discardedCards []model.Card) error {
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	if !e.isSpiritCaster(user) {
		return fmt.Errorf("仅灵符师可发动灵符技能")
	}
	targetIDs = dedupeIDs(targetIDs)
	if len(targetIDs) != 2 {
		return fmt.Errorf("灵符技能需要且仅需指定2名角色")
	}
	for _, tid := range targetIDs {
		if e.State.Players[tid] == nil {
			return fmt.Errorf("目标玩家不存在: %s", tid)
		}
	}

	e.enqueueDeferredFollowup(model.DeferredFollowup{
		Type:      "spirit_caster_talisman",
		UserID:    user.ID,
		SkillID:   skillID,
		TargetIDs: append([]string{}, targetIDs...),
	})
	// 弃牌“展示/封印”已由弃牌通知统一处理

	// 若封印产生了延迟伤害，优先进入伤害阶段，后续技能由 DeferredFollowups 在伤害后继续。
	if len(e.State.PendingDamageQueue) > 0 && e.State.Phase != model.PhasePendingDamageResolution {
		if e.State.ReturnPhase == "" {
			e.State.ReturnPhase = model.PhaseExtraAction
		}
		e.State.Phase = model.PhasePendingDamageResolution
	}
	return nil
}

func (e *GameEngine) resolveSpiritCasterTalismanFollowup(f model.DeferredFollowup) error {
	user := e.State.Players[f.UserID]
	if user == nil {
		return fmt.Errorf("执行者不存在: %s", f.UserID)
	}
	targetIDs := dedupeIDs(f.TargetIDs)
	if len(targetIDs) != 2 {
		return fmt.Errorf("灵符后续目标数量无效: %d", len(targetIDs))
	}

	// 念咒：每次发动灵符时可将1张手牌盖为妖力（上限2）。
	if spiritCasterPowerCount(user, "") < spiritCasterPowerCapEngine && len(user.Hand) > 0 {
		e.PushInterrupt(&model.Interrupt{
			Type:     model.InterruptChoice,
			PlayerID: user.ID,
			Context: map[string]interface{}{
				"choice_type": "sc_incant_confirm",
				"user_id":     user.ID,
				"skill_id":    f.SkillID,
				"target_ids":  append([]string{}, targetIDs...),
			},
		})
		return nil
	}
	if spiritCasterPowerCount(user, "") >= spiritCasterPowerCapEngine {
		e.Log(fmt.Sprintf("%s 的 [念咒] 未触发：妖力已达上限%d", user.Name, spiritCasterPowerCapEngine))
	}
	return e.continueSpiritCasterTalisman(user, f.SkillID, targetIDs)
}

func (e *GameEngine) continueSpiritCasterTalisman(user *model.Player, skillID string, targetIDs []string) error {
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	switch skillID {
	case "sc_talisman_thunder":
		if e.CanPayCrystalCost(user.ID, 1) {
			e.PushInterrupt(&model.Interrupt{
				Type:     model.InterruptChoice,
				PlayerID: user.ID,
				Context: map[string]interface{}{
					"choice_type": "sc_spiritual_collapse_confirm",
					"user_id":     user.ID,
					"mode":        "sc_talisman_thunder",
					"target_ids":  append([]string{}, targetIDs...),
				},
			})
			return nil
		}
		e.resolveSpiritCasterThunderDamage(user, targetIDs, 0)
	case "sc_talisman_wind":
		return e.startSpiritCasterWindDiscardFlow(user, targetIDs)
	default:
		return fmt.Errorf("未知灵符技能: %s", skillID)
	}
	return nil
}

func (e *GameEngine) resolveSpiritCasterThunderDamage(user *model.Player, targetIDs []string, bonus int) {
	if user == nil {
		return
	}
	damage := 1 + bonus
	if damage < 0 {
		damage = 0
	}
	targetSet := idsToSet(dedupeIDs(targetIDs))
	ordered := e.reverseOrderTargetIDsFrom(user.ID, true)
	hitCount := 0
	for _, tid := range ordered {
		if !targetSet[tid] {
			continue
		}
		e.AddPendingDamage(model.PendingDamage{
			SourceID:   user.ID,
			TargetID:   tid,
			Damage:     damage,
			DamageType: "magic",
			Stage:      0,
		})
		hitCount++
	}
	e.Log(fmt.Sprintf("%s 发动 [灵符-雷鸣]：对%d名角色各造成%d点法术伤害", user.Name, hitCount, damage))
	if len(e.State.PendingDamageQueue) > 0 {
		e.State.Phase = model.PhasePendingDamageResolution
		if e.State.ReturnPhase == "" {
			e.State.ReturnPhase = model.PhaseExtraAction
		}
	}
}

func (e *GameEngine) startSpiritCasterWindDiscardFlow(user *model.Player, targetIDs []string) error {
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	targetSet := idsToSet(dedupeIDs(targetIDs))
	orderedAll := e.reverseOrderTargetIDsFrom(user.ID, true)
	ordered := make([]string, 0, len(targetIDs))
	for _, pid := range orderedAll {
		if !targetSet[pid] {
			continue
		}
		ordered = append(ordered, pid)
	}
	if len(ordered) == 0 {
		e.Log(fmt.Sprintf("%s 的 [灵符-风行]：无有效目标", user.Name))
		return nil
	}

	cursor := 0
	for cursor < len(ordered) {
		target := e.State.Players[ordered[cursor]]
		if target == nil || len(target.Hand) == 0 {
			if target != nil {
				e.Log(fmt.Sprintf("%s 的 [灵符-风行]：%s 无手牌可弃置", user.Name, target.Name))
			}
			cursor++
			continue
		}
		break
	}
	if cursor >= len(ordered) {
		e.Log(fmt.Sprintf("%s 的 [灵符-风行]：所有目标均无手牌可弃置", user.Name))
		return nil
	}

	currentTargetID := ordered[cursor]
	e.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: currentTargetID,
		Context: map[string]interface{}{
			"choice_type":        "sc_talisman_wind_discard",
			"user_id":            user.ID,
			"ordered_target_ids": ordered,
			"cursor":             cursor,
			"current_target_id":  currentTargetID,
		},
	})
	return nil
}

func (e *GameEngine) resolveSpiritCasterHundredNightSingle(user *model.Player, targetID string, bonus int) error {
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	target := e.State.Players[targetID]
	if target == nil {
		return fmt.Errorf("目标不存在")
	}
	damage := 1 + bonus
	e.AddPendingDamage(model.PendingDamage{
		SourceID:   user.ID,
		TargetID:   target.ID,
		Damage:     damage,
		DamageType: "magic",
		Stage:      0,
	})
	e.Log(fmt.Sprintf("%s 发动 [百鬼夜行]：对 %s 造成%d点法术伤害", user.Name, target.Name, damage))
	return nil
}

func (e *GameEngine) resolveSpiritCasterHundredNightFireAOE(user *model.Player, excludeIDs []string, bonus int) error {
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	exclude := idsToSet(dedupeIDs(excludeIDs))
	damage := 1 + bonus
	ordered := e.reverseOrderTargetIDsFrom(user.ID, true)
	hitCount := 0
	for _, pid := range ordered {
		if exclude[pid] {
			continue
		}
		target := e.State.Players[pid]
		if target == nil {
			continue
		}
		e.AddPendingDamage(model.PendingDamage{
			SourceID:   user.ID,
			TargetID:   target.ID,
			Damage:     damage,
			DamageType: "magic",
			Stage:      0,
		})
		hitCount++
	}
	e.Log(fmt.Sprintf("%s 发动 [百鬼夜行·火]：对除2名指定角色外的其他角色各造成%d点法术伤害（命中%d名）", user.Name, damage, hitCount))
	return nil
}

func (e *GameEngine) resolveBloodPriestessSharedLifePlaceFollowup(f model.DeferredFollowup) error {
	user := e.State.Players[f.UserID]
	if user == nil {
		return fmt.Errorf("执行者不存在: %s", f.UserID)
	}
	if !e.isBloodPriestess(user) {
		return fmt.Errorf("仅血之巫女可执行同生共死后续")
	}
	if len(f.TargetIDs) != 1 {
		return fmt.Errorf("同生共死后续目标数量错误: %d", len(f.TargetIDs))
	}
	target := e.State.Players[f.TargetIDs[0]]
	if target == nil {
		return fmt.Errorf("同生共死目标不存在: %s", f.TargetIDs[0])
	}

	card := bloodPriestessSharedLifeCard(user)
	if f.Data != nil {
		if v, ok := f.Data["card"]; ok {
			switch c := v.(type) {
			case model.Card:
				card = c
			case *model.Card:
				if c != nil {
					card = *c
				}
			}
		}
	}

	if err := e.placeBloodPriestessSharedLife(user, target, card); err != nil {
		user.RestoreExclusiveCard(card)
		return err
	}
	e.Log(fmt.Sprintf("%s 的 [同生共死] 生效：放置于 %s 面前", user.Name, target.Name))

	// 放置后手牌上限可能变化，立即检查爆牌。
	e.checkHandLimit(user, nil)
	if target.ID != user.ID {
		e.checkHandLimit(target, nil)
	}
	return nil
}

func (e *GameEngine) resolveBloodPriestessWailDamageFollowup(f model.DeferredFollowup) error {
	user := e.State.Players[f.UserID]
	if user == nil {
		return fmt.Errorf("执行者不存在: %s", f.UserID)
	}
	if len(f.TargetIDs) != 1 {
		return fmt.Errorf("血之悲鸣后续目标数量错误: %d", len(f.TargetIDs))
	}
	target := e.State.Players[f.TargetIDs[0]]
	if target == nil {
		return fmt.Errorf("血之悲鸣目标不存在: %s", f.TargetIDs[0])
	}
	damage := 1
	if f.Data != nil {
		if v, ok := f.Data["damage"]; ok {
			damage = toIntContextValue(v)
			if damage <= 0 {
				damage = 1
			}
		}
	}
	e.AddPendingDamage(model.PendingDamage{
		SourceID:   user.ID,
		TargetID:   target.ID,
		Damage:     damage,
		DamageType: "magic",
		Stage:      0,
	})
	e.AddPendingDamage(model.PendingDamage{
		SourceID:   user.ID,
		TargetID:   user.ID,
		Damage:     damage,
		DamageType: "magic",
		Stage:      0,
	})
	e.Log(fmt.Sprintf("%s 的 [血之悲鸣] 延迟生效：对 %s 和自己各造成%d点法术伤害", user.Name, target.Name, damage))
	if len(e.State.PendingDamageQueue) > 0 {
		e.State.Phase = model.PhasePendingDamageResolution
		if e.State.ReturnPhase == "" {
			e.State.ReturnPhase = model.PhaseExtraAction
		}
	}
	return nil
}

// prepareMagicLancerFullnessStep 在“充盈”结算过程中推进到下一个需要选择的角色。
// 返回 true 表示所有角色都已处理完。
func (e *GameEngine) prepareMagicLancerFullnessStep(ctxData map[string]interface{}, user *model.Player) (bool, error) {
	if ctxData == nil || user == nil {
		return true, fmt.Errorf("充盈上下文无效")
	}
	orderIDs := parseStringSliceContextValue(ctxData["order_ids"])
	if len(orderIDs) == 0 {
		return true, nil
	}
	idx := toIntContextValue(ctxData["order_index"])
	if idx < 0 {
		idx = 0
	}
	for idx < len(orderIDs) {
		pid := orderIDs[idx]
		target := e.State.Players[pid]
		if target == nil {
			idx++
			continue
		}
		allowSkip := target.Camp == user.Camp
		candidates := allHandIndices(target)
		if len(candidates) == 0 {
			e.Log(fmt.Sprintf("%s 的 [充盈] 结算：%s 无手牌可弃，跳过", user.Name, target.Name))
			idx++
			continue
		}
		ctxData["order_index"] = idx
		ctxData["current_player_id"] = pid
		ctxData["allow_skip"] = allowSkip
		ctxData["candidates"] = candidates
		return false, nil
	}
	return true, nil
}

func (e *GameEngine) prependPendingDamages(pds []model.PendingDamage) {
	if len(pds) == 0 {
		return
	}
	// 使用“后进先出”顺序前插，满足嵌套法术伤害的栈式结算语义。
	reversed := make([]model.PendingDamage, 0, len(pds))
	for i := len(pds) - 1; i >= 0; i-- {
		reversed = append(reversed, pds[i])
	}
	e.State.PendingDamageQueue = append(reversed, e.State.PendingDamageQueue...)
	for _, pd := range reversed {
		e.Log(fmt.Sprintf("[System] 延迟伤害已前插: Source: %s, Target: %s, Damage: %d, Type: %s",
			pd.SourceID, pd.TargetID, pd.Damage, pd.DamageType))
	}
}

func (e *GameEngine) blazeWitchAttackElement(player *model.Player, card model.Card) model.Element {
	if player == nil || player.Tokens == nil {
		return card.Element
	}
	if !e.isBlazeWitch(player) || player.Tokens["bw_flame_form"] <= 0 {
		return card.Element
	}
	if card.Type != model.CardTypeAttack {
		return card.Element
	}
	if card.Element == model.ElementWater || card.Element == model.ElementDark {
		return card.Element
	}
	return model.ElementFire
}

func (e *GameEngine) applyBlazeWitchAttackCardTransform(player *model.Player, card model.Card) model.Card {
	card.Element = e.blazeWitchAttackElement(player, card)
	return card
}

func (e *GameEngine) isForcedAdventurerParadiseResponse(playerID string) bool {
	intr := e.State.PendingInterrupt
	if intr == nil || intr.Type != model.InterruptResponseSkill || intr.PlayerID != playerID {
		return false
	}
	player := e.State.Players[playerID]
	if player == nil || player.Tokens == nil || player.Tokens["adventurer_extract_requires_paradise"] <= 0 {
		return false
	}
	for _, sid := range intr.SkillIDs {
		if sid == "adventurer_paradise" {
			return true
		}
	}
	return false
}

// drawForAction 行动摸牌并处理爆牌
func (e *GameEngine) drawForAction(p *model.Player, count int) {
	cards, newDeck, newDiscard := rules.DrawCards(e.State.Deck, e.State.DiscardPile, count)
	e.State.Deck = newDeck
	e.State.DiscardPile = newDiscard
	p.Hand = append(p.Hand, cards...)
	e.NotifyDrawCards(p.ID, count, "action")
	e.checkHandLimit(p, nil)
}

// checkHandLimit 检查手牌上限
// GetMaxHand 计算玩家的动态手牌上限
func (e *GameEngine) GetMaxHand(p *model.Player) int {
	if p == nil {
		return 0
	}
	// 流血形态手牌<3时强制重置（优先于后续上限修正）。
	_ = e.maybeAutoReleaseBloodPriestessByHand(p, "手牌<3强制脱离流血形态")

	// 固定上限角色/状态：不受同生共死等动态修正影响。
	if e.hasFixedMaxHandCap(p) {
		if e.isMagicLancer(p) && p.Tokens != nil && p.Tokens["ml_phantom_form"] > 0 {
			return 5
		}
		if e.isHero(p) && p.Tokens != nil && p.Tokens["hero_exhaustion_form"] > 0 {
			return 4
		}
		return 7 // 怜悯
	}

	// 基础手牌上限
	maxHand := p.MaxHand
	if e.isWarHomunculus(p) && p.Tokens != nil && p.Tokens["hom_burst_form"] > 0 {
		maxHand++
	}
	if e.isBlazeWitch(p) && p.Tokens != nil && p.Tokens["bw_flame_form"] > 0 {
		maxHand += p.Tokens["bw_rebirth"] - 2
	}
	// 蝶舞者：生命之火，手牌上限 = 基础上限 - 蛹数，最低为3。
	if e.isButterflyDancer(p) {
		maxHand -= butterflyPupa(p)
		if maxHand < 3 {
			maxHand = 3
		}
	}
	// 血之巫女：同生共死根据形态动态修正手牌上限。
	maxHand += e.bloodPriestessSharedLifeDeltaFor(p)
	if maxHand < 0 {
		maxHand = 0
	}

	return maxHand
}

func (e *GameEngine) checkHandLimit(p *model.Player, ctx *model.Context) {
	if ctx != nil && ctx.Flags["preventOverflow"] {
		e.Log(fmt.Sprintf("[System] %s 的本次摸牌忽略手牌上限检查", p.Name))
		return
	}

	heroDeadDuelPending := false
	if p != nil && e.isHero(p) && p.Tokens != nil && p.Tokens["hero_dead_duel_pending"] > 0 {
		heroDeadDuelPending = true
	}

	over := len(p.Hand) - e.GetMaxHand(p)
	if over > 0 {
		isMagic := false
		if ctx != nil && ctx.Flags["IsMagicDamage"] {
			isMagic = true
		}
		fromDamageDraw := false
		if ctx != nil && ctx.Flags["FromDamageDraw"] {
			fromDamageDraw = true
		}
		noMoraleLoss := false
		if ctx != nil && ctx.Flags["NoMoraleLoss"] {
			noMoraleLoss = true
		}
		stayInTurn := false
		if ctx != nil && ctx.Flags["StayInTurn"] {
			stayInTurn = true
		}
		heroDeadDuelFloor := heroDeadDuelPending && fromDamageDraw && isMagic
		// 检查是否在伤害结算阶段（含延迟伤害结算）。
		isDamageResolution := e.State.Phase == model.PhaseDamageResolution || e.State.Phase == model.PhasePendingDamageResolution

		// 推送弃牌中断
		e.PushInterrupt(&model.Interrupt{
			Type:     model.InterruptDiscard,
			PlayerID: p.ID,
			Context: map[string]interface{}{
				"discard_count":        over,    // 必须弃掉所有溢出的牌
				"is_magic":             isMagic, // 传递伤害来源信息
				"from_damage_draw":     fromDamageDraw,
				"no_morale_loss":       noMoraleLoss,
				"victim_id":            p.ID,
				"stay_in_turn":         stayInTurn,         // 传递是否留在当前回合的标志
				"is_damage_resolution": isDamageResolution, // 标记是否在伤害结算阶段
				"hero_dead_duel_floor": heroDeadDuelFloor,
			},
		})
		if heroDeadDuelFloor && p.Tokens != nil {
			p.Tokens["hero_dead_duel_pending"] = 0
		}
		e.State.Phase = model.PhaseDiscardSelection
		e.Log(fmt.Sprintf("[System] %s 手牌超出上限 %d 张！需要选择 %d 张牌丢弃", p.Name, len(p.Hand), over))
	} else if heroDeadDuelPending && ctx != nil && ctx.Flags["FromDamageDraw"] && ctx.Flags["IsMagicDamage"] {
		// 死斗只作用于“本次法术伤害导致的爆牌士气下降”：若未爆牌，标记直接失效。
		p.Tokens["hero_dead_duel_pending"] = 0
	}
}

// Notify 统一通知方法 (替换所有的 fmt.Printf)
func (e *GameEngine) Notify(eventType model.GameEventType, msg string, data interface{}) {
	if e.observer != nil {
		e.observer.OnGameEvent(model.GameEvent{
			Type:    eventType,
			Message: msg,
			Data:    data,
		})
	}
}

// Log 快捷日志方法
func (e *GameEngine) Log(msg string) {
	e.Notify(model.EventLog, msg, nil)
}

// NotifyCardRevealed 通知明牌展示（出牌/弃牌等），供前端做动画
func (e *GameEngine) NotifyCardRevealed(playerID string, cards []model.Card, actionType string) {
	e.notifyCards(playerID, cards, actionType, false)
}

// NotifyCardHidden 通知暗弃展示（不展示牌面内容），供前端显示牌背
func (e *GameEngine) NotifyCardHidden(playerID string, cards []model.Card, actionType string) {
	e.notifyCards(playerID, cards, actionType, true)
}

// notifyCards 通用的牌展示通知方法
func (e *GameEngine) notifyCards(playerID string, cards []model.Card, actionType string, hidden bool) {
	if e.observer == nil || len(cards) == 0 {
		return
	}
	switch actionType {
	case "discard":
		e.addActionDiscard(playerID, len(cards))
	case "defend":
		if p := e.State.Players[playerID]; p != nil {
			cardNames := make([]string, 0, len(cards))
			for _, c := range cards {
				cardNames = append(cardNames, c.Name)
			}
			if len(cardNames) > 0 {
				e.addActionResponse(fmt.Sprintf("%s 防御【%s】", p.Name, strings.Join(cardNames, "、")))
			}
		}
	case "counter":
		if p := e.State.Players[playerID]; p != nil {
			cardNames := make([]string, 0, len(cards))
			for _, c := range cards {
				cardNames = append(cardNames, c.Name)
			}
			if len(cardNames) > 0 {
				e.addActionResponse(fmt.Sprintf("%s 应战【%s】", p.Name, strings.Join(cardNames, "、")))
			}
		}
	}
	p := e.State.Players[playerID]
	// 封印规则：打出/展示/丢弃都会触发封印（爆牌弃牌除外）。
	// 明弃触发“展示”类技能，暗弃仅触发封印。
	if actionType == "discard" && !e.suppressSealOnDiscard && p != nil {
		if hidden {
			for i := range cards {
				card := cards[i]
				e.triggerSealDamageForCardUse(p, &card)
			}
		} else if e.dispatcher != nil {
			for i := range cards {
				card := cards[i]
				cardCtx := &model.EventContext{
					Type:     model.EventCardUsed,
					SourceID: playerID,
					Card:     &card,
				}
				revealCtx := e.buildContext(p, nil, model.TriggerOnCardRevealed, cardCtx)
				e.dispatcher.OnTrigger(model.TriggerOnCardRevealed, revealCtx)
			}
		}
	}
	playerName := playerID
	if p != nil {
		playerName = p.Name
	}
	e.Notify(model.EventCardRevealed, "", map[string]interface{}{
		"player_id":   playerID,
		"player_name": playerName,
		"cards":       cards,
		"action_type": actionType,
		"hidden":      hidden,
	})
}

// NotifyDamageDealt 通知伤害结算，供前端暴血特效
func (e *GameEngine) NotifyDamageDealt(sourceID, targetID string, damage int, damageType string) {
	if e.observer == nil || damage <= 0 {
		return
	}
	e.addActionDamage(targetID, damage)
	source := e.State.Players[sourceID]
	target := e.State.Players[targetID]
	sourceName := sourceID
	targetName := targetID
	if source != nil {
		sourceName = source.Name
	}
	if target != nil {
		targetName = target.Name
	}
	e.Notify(model.EventDamageDealt, "", map[string]interface{}{
		"source_id":   sourceID,
		"source_name": sourceName,
		"target_id":   targetID,
		"target_name": targetName,
		"damage":      damage,
		"damage_type": damageType,
	})
}

// NotifyActionStep 通知行动步骤，供桌面区域展示行动流程
func (e *GameEngine) NotifyActionStep(line string) {
	if e.observer == nil || line == "" {
		return
	}
	if e.actionSummary != nil && e.actionSummary.active {
		e.addActionNote(line)
		return
	}
	e.Notify(model.EventActionStep, "", map[string]interface{}{
		"line": line,
		"kind": "detail",
	})
}

// NotifyActionSummary 发送行动汇总信息（战斗播报只展示此类）。
func (e *GameEngine) NotifyActionSummary(line string) {
	if e.observer == nil || line == "" {
		return
	}
	e.Notify(model.EventActionStep, "", map[string]interface{}{
		"line": line,
		"kind": "summary",
	})
}

// NotifyCombatCue 通知战斗双方与阶段，供前端在战区播放对战动画
func (e *GameEngine) NotifyCombatCue(attackerID, targetID, phase string) {
	if e.observer == nil || attackerID == "" || targetID == "" || phase == "" {
		return
	}
	e.Notify(model.EventCombatCue, "", map[string]interface{}{
		"attacker_id": attackerID,
		"target_id":   targetID,
		"phase":       phase, // attack/defend/take/counter
	})
}

// NotifyDrawCards 通知摸牌事件，供前端播放公共牌堆到角色区的摸牌动画
func (e *GameEngine) NotifyDrawCards(playerID string, count int, reason string) {
	if e.observer == nil || playerID == "" || count <= 0 {
		return
	}
	e.addActionDraw(playerID, count)
	p := e.State.Players[playerID]
	playerName := playerID
	if p != nil {
		playerName = p.Name
	}
	e.Notify(model.EventDrawCards, "", map[string]interface{}{
		"player_id":   playerID,
		"player_name": playerName,
		"draw_count":  count,
		"reason":      reason,
	})
}

// CheckHandLimit 提供给技能处理器的手牌上限检查入口。
func (e *GameEngine) CheckHandLimit(playerID string, stayInTurn bool) {
	player := e.State.Players[playerID]
	if player == nil {
		return
	}
	ctx := e.buildContext(player, nil, model.TriggerNone, nil)
	if stayInTurn {
		ctx.Flags["StayInTurn"] = true
	}
	e.checkHandLimit(player, ctx)
}

// GetAllPlayers 返回所有玩家的切片
func (e *GameEngine) GetAllPlayers() []*model.Player {
	players := make([]*model.Player, 0, len(e.State.Players))
	seen := make(map[string]struct{}, len(e.State.PlayerOrder))
	for _, pid := range e.State.PlayerOrder {
		p := e.State.Players[pid]
		if p == nil {
			continue
		}
		players = append(players, p)
		seen[pid] = struct{}{}
	}

	// 兜底：若存在未进 PlayerOrder 的玩家，按 ID 排序追加，保证结果稳定。
	extraIDs := make([]string, 0)
	for pid := range e.State.Players {
		if _, ok := seen[pid]; ok {
			continue
		}
		extraIDs = append(extraIDs, pid)
	}
	sort.Strings(extraIDs)
	for _, pid := range extraIDs {
		if p := e.State.Players[pid]; p != nil {
			players = append(players, p)
		}
	}

	return players
}

// resumePendingDraw 恢复暂停的扣卡流程
func (e *GameEngine) resumePendingDraw(ctx *model.Context) {
	// 检查上下文是否为摸牌前事件
	if ctx == nil || ctx.Trigger != model.TriggerBeforeDraw || ctx.TriggerCtx == nil || ctx.TriggerCtx.DrawCount == nil {
		e.Log("[Draw] 跳过恢复摸牌：上下文不完整")
		return
	}

	drawCount := *ctx.TriggerCtx.DrawCount
	target := ctx.User

	// 检查是否取消扣卡
	if ctx.Flags["cancelDraw"] {
		e.Log(fmt.Sprintf("[Draw] %s 的扣卡被取消", target.Name))
		return
	}
	if ctx.Flags["capToHandLimit"] {
		room := e.GetMaxHand(target) - len(target.Hand)
		if room < 0 {
			room = 0
		}
		if drawCount > room {
			e.Log(fmt.Sprintf("[Draw] %s 的伤害摸牌受上限保护：%d -> %d", target.Name, drawCount, room))
			drawCount = room
			*ctx.TriggerCtx.DrawCount = drawCount
		}
	}
	if drawCount <= 0 {
		e.Log(fmt.Sprintf("[Draw] %s 本次无需扣卡", target.Name))
		return
	}

	// 执行扣卡
	e.Log(fmt.Sprintf("[Draw] %s 扣卡 %d 张", target.Name, drawCount))
	cards, newDeck, newDiscard := rules.DrawCards(e.State.Deck, e.State.DiscardPile, drawCount)
	e.State.Deck = newDeck
	e.State.DiscardPile = newDiscard
	target.Hand = append(target.Hand, cards...)
	e.NotifyDrawCards(target.ID, drawCount, "resume_draw")

	// 爆牌检查 (传入上下文以支持preventOverflow标记)
	e.checkHandLimit(target, ctx)
}

// ConfirmStartupSkill 确认发动启动技能
func (e *GameEngine) ConfirmStartupSkill(playerID string, skillID string) error {
	return e.dispatcher.ConfirmStartupSkill(playerID, skillID)
}

// SkipStartupSkill 跳过启动技能
func (e *GameEngine) SkipStartupSkill(playerID string) error {
	return e.dispatcher.SkipStartupSkill(playerID)
}

// ConfirmResponseSkill 确认发动响应技能
func (e *GameEngine) ConfirmResponseSkill(playerID string, skillID string) error {
	return e.dispatcher.ConfirmResponseSkill(playerID, skillID)
}

// ConfirmDiscard 确认执行弃牌
func (e *GameEngine) ConfirmDiscard(playerID string, indices []int) error {
	// ... (校验代码保持不变) ...
	if e.State.PendingInterrupt == nil || e.State.PendingInterrupt.Type != model.InterruptDiscard {
		return fmt.Errorf("当前没有待处理的弃牌操作")
	}

	// 获取上下文数据
	data, _ := e.State.PendingInterrupt.Context.(map[string]interface{})
	skillID, hasSkillID := data["skill_id"].(string)

	// 处理技能交互回调
	if hasSkillID && skillID != "" {
		// 这是技能触发的弃牌交互，执行技能逻辑
		minSelect, _ := data["min"].(int)
		maxSelect, _ := data["max"].(int)

		// 验证选择数量
		if len(indices) < minSelect {
			return fmt.Errorf("至少需要选择 %d 张牌，你选择了 %d 张", minSelect, len(indices))
		}
		if len(indices) > maxSelect {
			return fmt.Errorf("最多只能选择 %d 张牌，你选择了 %d 张", maxSelect, len(indices))
		}

		// 获取原始上下文
		userCtx, hasCtx := data["user_ctx"]
		if !hasCtx {
			return fmt.Errorf("技能上下文丢失")
		}
		ctx, ok := userCtx.(*model.Context)
		if !ok {
			return fmt.Errorf("技能上下文格式错误")
		}

		// 注入选择结果
		if ctx.Selections == nil {
			ctx.Selections = make(map[string]any)
		}
		ctx.Selections["discard_indices"] = indices

		// 执行技能逻辑
		handler := skills.GetHandler(skillID)
		if handler == nil {
			return fmt.Errorf("技能处理器不存在")
		}

		err := handler.Execute(ctx)
		if err != nil {
			return fmt.Errorf("技能执行失败: %v", err)
		}

		// 处理技能执行后的弃牌
		if discardedCards, ok := ctx.Selections["discardedCards"]; ok {
			if cards, ok := discardedCards.([]model.Card); ok {
				e.State.DiscardPile = append(e.State.DiscardPile, cards...)
			}
		}

		// 检查是否需要恢复暂停的扣卡流程
		if ctx.Trigger == model.TriggerBeforeDraw {
			e.resumePendingDraw(ctx)
		}

		var nextSkillIDs []string
		if rawList, ok := data["remaining_skills"].([]string); ok {
			// 再次过滤，确保剩余技能在弃牌消耗资源后依然可用
			// 注意：这里需要 SkillDispatcher 的实例或者辅助函数，
			// 但 GameEngine 通常不直接持有 Dispatcher 的逻辑方法。
			// 简化处理：假设 ConfirmResponseSkill 传过来时已经过滤了一遍，
			// 或者在这里简单信任列表，等玩家选的时候再报错(资源不足)。
			// 为了严谨，建议在 SkillDispatcher 里做过滤，但这里我们直接使用列表：
			nextSkillIDs = rawList
		}

		if len(nextSkillIDs) > 0 {
			// 还有技能没放完：不 PopInterrupt，而是将状态切回 ResponseSkill
			e.State.PendingInterrupt.Type = model.InterruptResponseSkill
			e.State.PendingInterrupt.SkillIDs = nextSkillIDs
			// 移除 Context 里的弃牌专用数据，恢复为技能选择上下文 (其实复用 userCtx 即可)
			e.State.PendingInterrupt.Context = ctx

			e.Log("[System] 弃牌技能执行完毕，你还可以选择发动其他技能")

			// 保持 Phase 不变 (通常是 PhaseDiscardSelection 切回 PhaseResponse)
			e.State.Phase = model.PhaseResponse
			return nil
		}

		// 弹出中断，继续游戏流程
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil && ctx.Trigger == model.TriggerOnAttackMiss {
			if e.resumePendingAttackMiss(ctx) {
				return nil
			}
		}
		if e.State.PendingInterrupt == nil {
			if len(e.State.ActionStack) > 0 {
				// 还在战斗响应堆栈中
				e.State.Phase = model.PhaseResponse
			} else if len(e.State.ActionQueue) > 0 {
				// 还有被中断的行动 (如攻击前触发技能)
				e.State.Phase = model.PhaseBeforeAction
			} else {
				// 没什么事了，检查回合结束 (处理额外行动 Token)
				e.State.Phase = model.PhaseTurnEnd
			}
		}
		return nil
	}

	// 下面是原有的爆牌弃牌逻辑
	discardCount := data["discard_count"].(int)

	if len(indices) != discardCount {
		return fmt.Errorf("需要选择 %d 张牌丢弃，你选择了 %d 张", discardCount, len(indices))
	}

	player := e.State.Players[playerID]
	if player == nil {
		return fmt.Errorf("玩家不存在")
	}

	// 验证索引
	for _, idx := range indices {
		if idx < 0 || idx >= len(player.Hand) {
			return fmt.Errorf("无效的牌索引: %d", idx)
		}
	}

	// 检查是否有重复索引
	seen := make(map[int]bool)
	for _, idx := range indices {
		if seen[idx] {
			return fmt.Errorf("不能重复选择同一张牌")
		}
		seen[idx] = true
	}

	// 执行弃牌
	sort.Sort(sort.Reverse(sort.IntSlice(indices))) // 从大到小排序，避免索引变化
	var discardedCards []model.Card
	for _, idx := range indices {
		discardedCards = append(discardedCards, player.Hand[idx])
		player.Hand = append(player.Hand[:idx], player.Hand[idx+1:]...)
	}

	// 爆牌弃牌为暗弃，不展示牌面内容
	e.suppressSealOnDiscard = true
	e.NotifyCardHidden(playerID, discardedCards, "discard")
	e.suppressSealOnDiscard = false

	// 扣除士气
	moraleLoss := len(discardedCards)
	finalLoss := moraleLoss
	noMoraleLoss, _ := data["no_morale_loss"].(bool)
	heroDeadDuelFloor, _ := data["hero_dead_duel_floor"].(bool)
	stayInTurn, _ := data["stay_in_turn"].(bool)
	isDamageResolution, _ := data["is_damage_resolution"].(bool)
	if noMoraleLoss {
		moraleLoss = 0
		finalLoss = 0
	}
	if heroDeadDuelFloor && moraleLoss > 0 {
		// 死斗：若本次法术伤害导致实际士气下降，则该次下降值恒定为1。
		moraleLoss = 1
	}

	// 仅当有 victim_id 且为爆牌/伤害结算等扣除士气场景才处理士气
	if moraleLoss > 0 {
		victimID, _ := data["victim_id"].(string)
		fromDamageDraw, _ := data["from_damage_draw"].(bool)
		victim := e.State.Players[victimID]
		// 圣剑摸X弃X、魔爆冲击等技能弃牌不扣士气，无 victim_id
		if victim != nil {
			var lossCtx *model.Context
			// 红莲骑士热血形态：对“伤害结算导致的爆牌”免疫士气下降。
			if (fromDamageDraw || isDamageResolution) && e.isCrimsonKnight(victim) {
				if victim.Tokens == nil {
					victim.Tokens = map[string]int{}
				}
				if victim.Tokens["crk_hot_form"] > 0 {
					moraleLoss = 0
				}
			}
			isMagic, _ := data["is_magic"].(bool)
			// 蝶舞者【凋零】士气下限：在触发“士气下降前”技能前先裁剪可下降值，避免无实际下降时仍触发相关技能。
			allowedByFloor := e.campMorale(victim.Camp) - e.moraleFloorForCamp(victim.Camp)
			if allowedByFloor < 0 {
				allowedByFloor = 0
			}
			if moraleLoss > allowedByFloor {
				moraleLoss = allowedByFloor
			}

			if moraleLoss > 0 {
				lossEventCtx := &model.EventContext{
					Type:      model.EventDamage,
					DamageVal: &moraleLoss,
				}
				lossCtx = e.buildContext(victim, nil, model.TriggerBeforeMoraleLoss, lossEventCtx)
				lossCtx.Flags["IsMagicDamage"] = isMagic
				if lossCtx.Selections == nil {
					lossCtx.Selections = map[string]any{}
				}
				lossCtx.Selections["discarded_cards"] = append([]model.Card{}, discardedCards...)
				lossCtx.Selections["from_damage_draw"] = fromDamageDraw
				lossCtx.Selections["victim_id"] = victimID
				lossCtx.Selections["discard_player_id"] = player.ID
				lossCtx.Selections["morale_loss_stay_in_turn"] = stayInTurn
				lossCtx.Selections["morale_loss_is_damage_resolution"] = isDamageResolution

				e.dispatcher.OnTrigger(model.TriggerBeforeMoraleLoss, lossCtx)
				// 若触发了响应技能（如神之庇护），延后到响应结束后再结算士气损失
				pendingResponse := false
				for _, intr := range e.State.InterruptQueue {
					if intr != nil && intr.Type == model.InterruptResponseSkill {
						pendingResponse = true
						break
					}
				}
				if pendingResponse {
					lossCtx.Selections["morale_loss_pending"] = true
					lossCtx.Selections["morale_loss_value"] = moraleLoss
					lossCtx.Selections["is_magic"] = isMagic
					lossCtx.Selections["hero_dead_duel_floor"] = heroDeadDuelFloor
					// 弹出当前弃牌中断，让响应技能进入处理
					e.PopInterrupt()
					return nil
				}
			} else {
				finalLoss = 0
			}
			finalLoss = e.applyMoraleLossAfterTrigger(victim, moraleLoss, isMagic, fromDamageDraw, heroDeadDuelFloor, discardedCards, lossCtx)
		} else {
			e.State.DiscardPile = append(e.State.DiscardPile, discardedCards...)
			finalLoss = 0
		}
	} else {
		// 无士气损失结算时，弃牌正常进入弃牌堆
		e.State.DiscardPile = append(e.State.DiscardPile, discardedCards...)
	}

	e.Log(fmt.Sprintf("[System] %s 丢弃了 %d 张牌！士气 -%d", player.Name, len(discardedCards), finalLoss))
	// 魔枪：幻影星尘若因本次自伤进入爆牌弃牌，需要在此处完成“完全结算后”的后续判定。
	if e.isMagicLancer(player) && player.Tokens != nil && player.Tokens["ml_stardust_wait_discard"] > 0 {
		e.resolveMagicLancerStardustAfterSelf(player)
	}

	// 【新增】使用 PopInterrupt 处理队列
	e.PopInterrupt()
	if e.State.PendingInterrupt == nil {
		// 检查是否在伤害结算阶段（通过中断上下文中的标志）
		if isDamageResolution {
			// 伤害结算阶段的弃牌完成，进入 PhaseExtraAction
			e.State.Phase = model.PhaseExtraAction

		} else if stayInTurn {
			// 比如中毒摸牌导致的爆牌，弃完牌后继续回合
			e.Log("[System] 弃牌完成，继续当前回合")
			// 如果有 ReturnPhase 设置（如从 PendingDamageResolution 阶段来），使用它
			if e.State.ReturnPhase != "" {
				e.State.Phase = e.State.ReturnPhase
				e.State.ReturnPhase = ""
			} else if len(e.State.ActionQueue) > 0 {
				e.State.Phase = model.PhaseBeforeAction
			} else if len(e.State.PendingDamageQueue) > 0 {
				// 还有待处理的伤害，继续处理
				e.State.Phase = model.PhasePendingDamageResolution
			} else {
				// 默认回到启动阶段，让回合继续正常流程
				e.State.Phase = model.PhaseStartup
			}
		} else {
			e.State.Phase = model.PhaseTurnEnd
		}
	}

	// 检查游戏结束条件
	e.checkGameEnd()

	return nil
}

func (e *GameEngine) resolveCrimsonKnightBloodyPrayer(user *model.Player, x int, allocations map[string]int) error {
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	if x <= 0 {
		return fmt.Errorf("无效的X值")
	}
	if user.Heal < x {
		return fmt.Errorf("治疗不足，无法结算血腥祷言")
	}

	user.Heal -= x
	for _, pid := range e.State.PlayerOrder {
		amt := allocations[pid]
		if amt <= 0 {
			continue
		}
		e.Heal(pid, amt)
	}
	e.AddPendingDamage(model.PendingDamage{
		SourceID:              user.ID,
		TargetID:              user.ID,
		Damage:                x,
		DamageType:            "magic",
		AllowCrimsonFaithHeal: true,
		Stage:                 0,
	})
	if user.Tokens == nil {
		user.Tokens = map[string]int{}
	}
	user.Tokens["crk_blood_mark"]++
	if user.Tokens["crk_blood_mark"] > 3 {
		user.Tokens["crk_blood_mark"] = 3
	}

	var parts []string
	for _, pid := range e.State.PlayerOrder {
		amt := allocations[pid]
		if amt <= 0 {
			continue
		}
		if p := e.State.Players[pid]; p != nil {
			parts = append(parts, fmt.Sprintf("%s +%d治疗", p.Name, amt))
		}
	}
	allocText := "未分配治疗"
	if len(parts) > 0 {
		allocText = strings.Join(parts, "，")
	}
	e.Log(fmt.Sprintf("%s 发动 [血腥祷言]：移除%d治疗并自伤%d，%s，血印+1", user.Name, x, x, allocText))
	return nil
}

// checkGameEnd 检查游戏是否结束
func (e *GameEngine) checkGameEnd() {
	// 星杯胜利：任一方星杯达到 5
	if e.State.RedCups >= 5 {
		e.Notify(model.EventGameEnd, "红方胜利！星杯达到 5", nil)
		e.State.Phase = model.PhaseEnd
		return
	}
	if e.State.BlueCups >= 5 {
		e.Notify(model.EventGameEnd, "蓝方胜利！星杯达到 5", nil)
		e.State.Phase = model.PhaseEnd
		return
	}
	// 检查是否有玩家的士气归零
	for _, player := range e.State.Players {
		if player.Camp == model.RedCamp && e.State.RedMorale <= 0 {
			e.Notify(model.EventGameEnd, "蓝方胜利！红方士气归零", nil)
			e.State.Phase = model.PhaseEnd
			return
		}
		if player.Camp == model.BlueCamp && e.State.BlueMorale <= 0 {
			e.Notify(model.EventGameEnd, "红方胜利！蓝方士气归零", nil)
			e.State.Phase = model.PhaseEnd
			return
		}
	}
}

// GetCurrentPrompt 获取当前用户交互提示
func (e *GameEngine) GetCurrentPrompt() *model.Prompt {
	// 如果有中断，优先处理中断相关的Prompt
	if e.State.PendingInterrupt != nil {
		switch e.State.PendingInterrupt.Type {
		case model.InterruptResponseSkill:
			if e.prunePendingResponseSkills() {
				_ = e.SkipResponse()
				return nil
			}
			return e.buildResponseSkillPrompt()
		case model.InterruptStartupSkill:
			return e.buildStartupSkillPrompt()
		case model.InterruptDiscard:
			return e.buildDiscardPrompt()
		case model.InterruptChoice:
			return e.buildChoicePrompt()
		case model.InterruptMagicMissile:
			return e.buildMagicMissilePrompt()
		case model.InterruptGiveCards:
			return e.buildGiveCardsPrompt()
		case model.InterruptMagicBulletFusion:
			return e.buildMagicBulletFusionPrompt()
		case model.InterruptMagicBulletDirection:
			return e.buildMagicBulletDirectionPrompt()
		case model.InterruptHolySwordDraw:
			return e.buildHolySwordDrawPrompt()
		case model.InterruptSaintHeal:
			return e.buildSaintHealPrompt()
		case model.InterruptMagicBlast:
			return e.buildMagicBlastPrompt()
		}
	}
	// 2. 【新增】处理普通的响应阶段提示
	if e.State.Phase == model.PhaseResponse && len(e.State.ActionStack) > 0 {
		lastAction := e.State.ActionStack[len(e.State.ActionStack)-1]
		targetID := lastAction.TargetID

		// 只有目标玩家能看到提示
		return &model.Prompt{
			Type:     model.PromptConfirm, // 或者定义一个新的 PromptTypeStandardResponse
			PlayerID: targetID,
			Message:  fmt.Sprintf("你成为了 %s 的目标，请做出响应 (take/counter/defend)", lastAction.Type),
			Options: []model.PromptOption{
				{ID: "take", Label: "承受 (take) - 结算伤害/效果"},
				{ID: "counter", Label: "应战 (counter <idx>) - 尝试反击"},
				{ID: "defend", Label: "防御 (defend) - 使用圣光（圣盾需提前放置）"},
			},
			// 如果是强制命中，可以在 Message 里提示 "攻击强制命中，无法防御/应战，只能 take"
		}
	}

	return nil
}

// PushInterrupt 向引擎推送一个中断
func (e *GameEngine) PushInterrupt(interrupt *model.Interrupt) {
	if e.State.PendingInterrupt == nil {
		e.State.PendingInterrupt = interrupt
		e.updatePhaseByInterrupt(interrupt)
		choiceType := ""
		if data, ok := interrupt.Context.(map[string]interface{}); ok {
			if ct, ok := data["choice_type"].(string); ok {
				choiceType = ct
			}
		}
		if choiceType != "" {
			e.Log(fmt.Sprintf("[Interrupt] Pending=%s Player=%s Choice=%s", interrupt.Type, interrupt.PlayerID, choiceType))
		} else {
			e.Log(fmt.Sprintf("[Interrupt] Pending=%s Player=%s", interrupt.Type, interrupt.PlayerID))
		}
		// 立即发送 AskInput 事件
		e.notifyInterruptPrompt()
	} else {
		// 2. 否则进入队列排队
		e.State.InterruptQueue = append(e.State.InterruptQueue, interrupt)
		e.Log(fmt.Sprintf("新中断入队等待: %s (Player: %s)", interrupt.Type, interrupt.PlayerID))
	}
}

// 辅助方法：根据中断类型自动更新游戏阶段
func (e *GameEngine) updatePhaseByInterrupt(interrupt *model.Interrupt) {
	switch interrupt.Type {
	case model.InterruptResponseSkill:
		// 响应技能 -> 进入响应阶段
		e.State.Phase = model.PhaseResponse
	case model.InterruptDiscard:
		// 弃牌 -> 进入弃牌选择阶段
		e.State.Phase = model.PhaseDiscardSelection
	case model.InterruptStartupSkill:
		// 启动技 -> 保持在启动阶段 (或专门的 Trigger 阶段)
		e.State.Phase = model.PhaseStartup
	case model.InterruptChoice:
		if data, ok := interrupt.Context.(map[string]interface{}); ok {
			if typeVal, ok := data["choice_type"].(string); ok && typeVal == "weak" {
				e.State.Phase = model.PhaseBuffResolve
				return
			}
			if typeVal, ok := data["choice_type"].(string); ok && typeVal == "heal" {
				e.State.Phase = model.PhasePendingDamageResolution
				return
			}
			if typeVal, ok := data["choice_type"].(string); ok && typeVal == "buy_resource" {
				e.State.Phase = model.PhaseTurnEnd // 购买选择后直接进入回合结束，用此 phase 仅表示等待
				return
			}
			if typeVal, ok := data["choice_type"].(string); ok && typeVal == "extract" {
				e.State.Phase = model.PhaseActionSelection // 提炼选择，仍在行动选择流程
				return
			}
			if typeVal, ok := data["choice_type"].(string); ok && typeVal == "five_elements_bind" {
				e.State.Phase = model.PhaseBuffResolve // 五系束缚选择在BuffResolve阶段
				return
			}
			if typeVal, ok := data["choice_type"].(string); ok && typeVal == "fighter_psi_bullet_target" {
				e.State.Phase = model.PhaseExtraAction
				return
			}
			if typeVal, ok := data["choice_type"].(string); ok &&
				(typeVal == "mb_magic_pierce_hit_confirm" ||
					typeVal == "hb_holy_shard_miss_confirm" ||
					typeVal == "hb_holy_shard_miss_x" ||
					typeVal == "hb_holy_shard_miss_ally_target" ||
					typeVal == "ml_black_spear_x" ||
					typeVal == "ml_dark_barrier_mode" ||
					typeVal == "ml_dark_barrier_x" ||
					typeVal == "ml_dark_barrier_cards" ||
					typeVal == "sc_hundred_night_power" ||
					typeVal == "sc_hundred_night_fire_reveal" ||
					typeVal == "sc_hundred_night_target" ||
					typeVal == "sc_hundred_night_exclude_pick") {
				e.State.Phase = model.PhasePendingDamageResolution
				return
			}
			if typeVal, ok := data["choice_type"].(string); ok &&
				(typeVal == "hb_meteor_bullet_cost" || typeVal == "hb_meteor_bullet_target") {
				e.State.Phase = model.PhaseResponse
				return
			}
			if typeVal, ok := data["choice_type"].(string); ok && typeVal == "ss_convert_color" {
				e.State.Phase = model.PhaseResponse
				return
			}
			if typeVal, ok := data["choice_type"].(string); ok && typeVal == "ss_link_target" {
				e.State.Phase = model.PhaseStartup
				return
			}
			if typeVal, ok := data["choice_type"].(string); ok && typeVal == "ss_recall_pick" {
				e.State.Phase = model.PhaseExtraAction
				return
			}
			if typeVal, ok := data["choice_type"].(string); ok && typeVal == "ss_link_transfer_x" {
				e.State.Phase = model.PhasePendingDamageResolution
				return
			}
			if typeVal, ok := data["choice_type"].(string); ok &&
				(typeVal == "bp_shared_life_target" || typeVal == "bp_curse_discard") {
				e.State.Phase = model.PhaseExtraAction
				return
			}
			if typeVal, ok := data["choice_type"].(string); ok &&
				(typeVal == "bp_blood_sorrow_mode" || typeVal == "bp_blood_sorrow_target") {
				e.State.Phase = model.PhaseStartup
				return
			}
			if typeVal, ok := data["choice_type"].(string); ok && typeVal == "angel_song_pick" {
				e.State.Phase = model.PhaseStartup
				return
			}
			if typeVal, ok := data["choice_type"].(string); ok && typeVal == "angel_bond_heal_target" {
				// 保持当前阶段不变
				return
			}
			if typeVal, ok := data["choice_type"].(string); ok &&
				(typeVal == "bt_dance_mode" ||
					typeVal == "bt_dance_discard" ||
					typeVal == "bt_chrysalis_resolve" ||
					typeVal == "bt_cocoon_overflow_discard" ||
					typeVal == "bt_reverse_discard" ||
					typeVal == "bt_reverse_mode" ||
					typeVal == "bt_reverse_target" ||
					typeVal == "bt_reverse_branch2_cost" ||
					typeVal == "bt_reverse_branch2_pick") {
				e.State.Phase = model.PhaseExtraAction
				return
			}
			if typeVal, ok := data["choice_type"].(string); ok &&
				(typeVal == "bt_pilgrimage_pick" ||
					typeVal == "bt_poison_pick" ||
					typeVal == "bt_mirror_pair" ||
					typeVal == "bt_wither_confirm" ||
					typeVal == "bt_wither_target") {
				e.State.Phase = model.PhasePendingDamageResolution
				return
			}
			if typeVal, ok := data["choice_type"].(string); ok && typeVal == "bp_blood_wail_x" {
				e.State.Phase = model.PhasePendingDamageResolution
				return
			}
			if typeVal, ok := data["choice_type"].(string); ok &&
				(typeVal == "mg_medusa_darkmoon_pick" ||
					typeVal == "mg_medusa_magic_discard" ||
					typeVal == "mg_medusa_magic_target") {
				e.State.Phase = model.PhaseResponse
				return
			}
			if typeVal, ok := data["choice_type"].(string); ok && typeVal == "mg_darkmoon_slash_x" {
				e.State.Phase = model.PhasePendingDamageResolution
				return
			}
			if typeVal, ok := data["choice_type"].(string); ok &&
				(typeVal == "mg_moon_cycle_mode" || typeVal == "mg_moon_cycle_heal_target") {
				e.State.Phase = model.PhaseTurnEnd
				return
			}
			if typeVal, ok := data["choice_type"].(string); ok && typeVal == "mg_blasphemy_target" {
				e.State.Phase = model.PhasePendingDamageResolution
				return
			}
			if typeVal, ok := data["choice_type"].(string); ok &&
				(typeVal == "mg_pale_moon_mode" ||
					typeVal == "mg_pale_moon_x" ||
					typeVal == "mg_pale_moon_target" ||
					typeVal == "mg_pale_moon_discard") {
				e.State.Phase = model.PhaseActionSelection
				return
			}
			if typeVal, ok := data["choice_type"].(string); ok &&
				(typeVal == "hb_auto_fill_resource" || typeVal == "hb_auto_fill_gain") {
				e.State.Phase = model.PhaseTurnEnd
				return
			}
			if typeVal, ok := data["choice_type"].(string); ok && typeVal == "sc_spiritual_collapse_confirm" {
				if mode, _ := data["mode"].(string); strings.HasPrefix(mode, "sc_hundred_night") {
					e.State.Phase = model.PhasePendingDamageResolution
					return
				}
			}
		}
		// 默认回退到行动选择（或者你定义的 PhaseGeneralChoice）
		e.State.Phase = model.PhaseActionSelection
	case model.InterruptMagicMissile:
		// 魔弹响应 -> 进入响应阶段
		e.State.Phase = model.PhaseResponse
	case model.InterruptGiveCards:
		e.State.Phase = model.PhaseDiscardSelection
	case model.InterruptMagicBulletFusion, model.InterruptMagicBulletDirection:
		// 魔弹融合/掌控询问 -> 保持在行动执行阶段
		e.State.Phase = model.PhaseActionExecution
	case model.InterruptHolySwordDraw:
		// 圣剑摸X弃X -> 保持在当前阶段
		e.State.Phase = model.PhaseActionExecution
	case model.InterruptSaintHeal:
		// 圣疗分配治疗 -> 保持在当前阶段
		e.State.Phase = model.PhaseActionExecution
	case model.InterruptMagicBlast:
		// 魔爆冲击弃牌 -> 保持在响应阶段
		e.State.Phase = model.PhaseResponse
	}
}

// PopInterrupt 弹出当前中断并处理下一个
func (e *GameEngine) PopInterrupt() {
	e.State.PendingInterrupt = nil
	// 2. 检查队列
	if len(e.State.InterruptQueue) > 0 {
		// 取出队首
		nextInterrupt := e.State.InterruptQueue[0]
		e.State.InterruptQueue = e.State.InterruptQueue[1:]

		// 设置为当前中断
		e.State.PendingInterrupt = nextInterrupt
		e.Log(fmt.Sprintf("[System] 队列弹出中断: %s", nextInterrupt.Type))

		// 更新阶段并通知
		e.updatePhaseByInterrupt(nextInterrupt)
		e.notifyInterruptPrompt()
	} else {
		// 队列为空，什么都不做
		// Drive() 循环会在下一次运行时接管流程，根据 Phase 进行自动流转
		e.Log("[System] 所有中断处理完毕，恢复主流程")
	}
}

// notifyInterruptPrompt 发送中断提示事件
func (e *GameEngine) notifyInterruptPrompt() {
	if e.State.PendingInterrupt == nil {
		return
	}
	var prompt *model.Prompt
	switch e.State.PendingInterrupt.Type {
	case model.InterruptResponseSkill:
		if e.prunePendingResponseSkills() {
			if err := e.SkipResponse(); err != nil {
				e.Log(fmt.Sprintf("[System] 自动跳过无可用响应失败: %v", err))
			}
			return
		}
		prompt = e.buildResponseSkillPrompt()
	case model.InterruptDiscard:
		prompt = e.buildDiscardPrompt()
	case model.InterruptStartupSkill:
		prompt = e.buildStartupSkillPrompt()
	case model.InterruptChoice:
		prompt = e.buildChoicePrompt()
	case model.InterruptMagicMissile:
		prompt = e.buildMagicMissilePrompt()
	case model.InterruptGiveCards:
		prompt = e.buildGiveCardsPrompt()
	case model.InterruptMagicBulletFusion:
		prompt = e.buildMagicBulletFusionPrompt()
	case model.InterruptMagicBulletDirection:
		prompt = e.buildMagicBulletDirectionPrompt()
	case model.InterruptHolySwordDraw:
		prompt = e.buildHolySwordDrawPrompt()
	case model.InterruptSaintHeal:
		prompt = e.buildSaintHealPrompt()
	case model.InterruptMagicBlast:
		prompt = e.buildMagicBlastPrompt()
	default:
		return
	}
	if prompt != nil {
		e.Notify(model.EventAskInput, "", prompt)
	}
}

// prunePendingResponseSkills 重新校验响应技能列表，移除当前已不满足条件的技能。
// 返回 true 表示已无可用技能。
func (e *GameEngine) prunePendingResponseSkills() bool {
	intr := e.State.PendingInterrupt
	if intr == nil || intr.Type != model.InterruptResponseSkill {
		return false
	}
	if len(intr.SkillIDs) == 0 {
		return true
	}

	player := e.State.Players[intr.PlayerID]
	if player == nil || e.dispatcher == nil {
		intr.SkillIDs = nil
		return true
	}

	var ctx *model.Context
	switch data := intr.Context.(type) {
	case *model.Context:
		ctx = data
	case map[string]interface{}:
		if userCtx, ok := data["user_ctx"].(*model.Context); ok {
			ctx = userCtx
		}
	}
	if ctx == nil {
		ctx = &model.Context{}
	}
	ctx.Game = e
	ctx.User = player

	filtered := make([]string, 0, len(intr.SkillIDs))
	for _, skillID := range intr.SkillIDs {
		if skillID == "" {
			continue
		}
		if e.dispatcher.isSkillStillUsable(skillID, player, ctx) {
			filtered = append(filtered, skillID)
		}
	}

	if len(filtered) != len(intr.SkillIDs) {
		e.Log(fmt.Sprintf("[System] 响应技能实时校验：%d -> %d", len(intr.SkillIDs), len(filtered)))
		intr.SkillIDs = filtered
	}

	return len(intr.SkillIDs) == 0
}

// buildResponseSkillPrompt 构建响应技能选择提示
func (e *GameEngine) buildResponseSkillPrompt() *model.Prompt {
	playerID := e.State.PendingInterrupt.PlayerID
	player := e.State.Players[playerID]

	skillIDs := e.State.PendingInterrupt.SkillIDs
	n := len(skillIDs)
	// 提示语中明确说明输入方式：choose 1 / choose 2 / ... / choose N 跳过
	message := fmt.Sprintf("你触发了多个响应机会，请选择发动 (剩余 %d 个)。输入 choose 1 发动第一项，choose 2 发动第二项，…，choose %d 跳过：", n, n+1)
	var options []model.PromptOption

	for i, skillID := range skillIDs {
		for _, skill := range player.Character.Skills {
			if skill.ID == skillID {
				costStr := ""
				if skill.CostGem > 0 || skill.CostCrystal > 0 {
					costStr = fmt.Sprintf(" [💎%d 🏆%d]", skill.CostGem, skill.CostCrystal)
				}
				options = append(options, model.PromptOption{
					ID:    skill.ID,
					Label: fmt.Sprintf("%d. %s%s: %s", i+1, skill.Title, costStr, skill.Description),
				})
				break
			}
		}
	}

	// 跳过选项：序号为 len(skillIDs)+1，即 choose N+1
	options = append(options, model.PromptOption{
		ID:    "skip",
		Label: fmt.Sprintf("%d. 跳过 / 结束响应", n+1),
	})

	return &model.Prompt{
		Type:     model.PromptChooseSkill,
		PlayerID: playerID,
		Message:  message,
		Options:  options,
		Min:      1,
		Max:      1,
	}
}

// buildStartupSkillPrompt 构建启动技能确认提示
func (e *GameEngine) buildStartupSkillPrompt() *model.Prompt {
	playerID := e.State.PendingInterrupt.PlayerID
	player := e.State.Players[playerID]

	skillIDs := e.State.PendingInterrupt.SkillIDs
	n := len(skillIDs)
	message := fmt.Sprintf("你可以发动启动技能。输入 choose 1 发动第一项，…，choose %d 跳过：", n+1)
	var options []model.PromptOption

	for i, skillID := range skillIDs {
		for _, skill := range player.Character.Skills {
			if skill.ID == skillID {
				costStr := ""
				if skill.CostGem > 0 || skill.CostCrystal > 0 {
					costStr = fmt.Sprintf(" (消耗: 宝石%d 水晶%d)", skill.CostGem, skill.CostCrystal)
				}
				options = append(options, model.PromptOption{
					ID:    skill.ID,
					Label: fmt.Sprintf("%d. %s%s - %s", i+1, skill.Title, costStr, skill.Description),
				})
				break
			}
		}
	}

	options = append(options, model.PromptOption{
		ID:    "skip",
		Label: fmt.Sprintf("%d. 跳过 - 不发动启动技能", n+1),
	})

	return &model.Prompt{
		Type:     model.PromptChooseSkill,
		PlayerID: playerID,
		Message:  message,
		Options:  options,
		Min:      1,
		Max:      1,
	}
}

// buildDiscardPrompt 构建弃牌选择提示
// formatCardInfo 格式化卡牌信息 (复用printHand的逻辑)
func formatCardInfo(card model.Card) string {
	// 基础信息
	elementLabel := elementNameForPrompt(string(card.Element))
	if elementLabel == "" {
		info := fmt.Sprintf("[%s] %s", card.Element, card.Name)
		// 类型和伤害
		if card.Type != "" {
			info += fmt.Sprintf(" (%s", card.Type)
			if card.Damage > 0 {
				info += fmt.Sprintf(" Dmg:%d", card.Damage)
			}
			info += ")"
		}

		// 命格
		if card.Faction != "" {
			info += fmt.Sprintf(" [%s命格]", card.Faction)
		}

		// 独有技信息
		exclusiveInfo := []string{}
		if card.ExclusiveChar1 != "" && card.ExclusiveSkill1 != "" {
			exclusiveInfo = append(exclusiveInfo, fmt.Sprintf("%s:%s", card.ExclusiveChar1, card.ExclusiveSkill1))
		}
		if card.ExclusiveChar2 != "" && card.ExclusiveSkill2 != "" {
			exclusiveInfo = append(exclusiveInfo, fmt.Sprintf("%s:%s", card.ExclusiveChar2, card.ExclusiveSkill2))
		}
		if len(exclusiveInfo) > 0 {
			info += fmt.Sprintf(" [独有技:%s]", strings.Join(exclusiveInfo, " | "))
		}

		return info
	}
	info := fmt.Sprintf("[%s系] %s", elementLabel, card.Name)

	// 类型和伤害
	if card.Type != "" {
		info += fmt.Sprintf(" (%s", card.Type)
		if card.Damage > 0 {
			info += fmt.Sprintf(" Dmg:%d", card.Damage)
		}
		info += ")"
	}

	// 命格
	if card.Faction != "" {
		info += fmt.Sprintf(" [%s命格]", card.Faction)
	}

	// 独有技信息
	exclusiveInfo := []string{}
	if card.ExclusiveChar1 != "" && card.ExclusiveSkill1 != "" {
		exclusiveInfo = append(exclusiveInfo, fmt.Sprintf("%s:%s", card.ExclusiveChar1, card.ExclusiveSkill1))
	}
	if card.ExclusiveChar2 != "" && card.ExclusiveSkill2 != "" {
		exclusiveInfo = append(exclusiveInfo, fmt.Sprintf("%s:%s", card.ExclusiveChar2, card.ExclusiveSkill2))
	}
	if len(exclusiveInfo) > 0 {
		info += fmt.Sprintf(" [独有技:%s]", strings.Join(exclusiveInfo, " | "))
	}

	return info
}

func (e *GameEngine) buildDiscardPrompt() *model.Prompt {
	playerID := e.State.PendingInterrupt.PlayerID
	player := e.State.Players[playerID]

	data := e.State.PendingInterrupt.Context.(map[string]interface{})
	var message string
	var min, max int

	// 场景 A: 爆牌 (固定数量) - 通过上下文传递的 discard_count 处理
	if count, ok := data["discard_count"].(int); ok && count > 0 {
		min = count
		max = count
		message = fmt.Sprintf("手牌上限溢出！请弃置 %d 张牌：", count)
		if customMsg, ok := data["prompt"].(string); ok && customMsg != "" {
			message = customMsg
		}
	} else {
		// 场景 B: 技能交互 (范围数量)
		// 安全地获取 min/max，获取失败则给默认值
		if v, ok := data["min"].(int); ok {
			min = v
		} else {
			min = 1
		}

		// max 为 0 或 -1 通常代表不限制上限（即最多弃全手牌）
		if v, ok := data["max"].(int); ok && v > 0 {
			max = v
		} else {
			max = len(player.Hand)
		}

		// 优先使用技能配置的提示语 (例如: "请选择水系牌...")
		if customMsg, ok := data["prompt"].(string); ok && customMsg != "" {
			message = customMsg
		} else {
			message = fmt.Sprintf("请选择 %d-%d 张牌弃置：", min, max)
		}
	}

	var options []model.PromptOption
	discardType, _ := data["discard_type"].(model.CardType)
	discardElement, _ := data["discard_element"].(model.Element)
	excludeBlessings, _ := data["exclude_blessings"].(bool)
	for i, card := range player.Hand {
		// 技能弃牌：按类型/元素过滤
		if discardType != "" && card.Type != discardType {
			continue
		}
		if discardElement != "" && card.Element != discardElement {
			continue
		}
		if excludeBlessings && isElfBlessingCard(player, card.ID) {
			continue
		}
		options = append(options, model.PromptOption{
			ID:    strconv.Itoa(i), // 选项ID就是手牌索引
			Label: fmt.Sprintf("%d: %s", i+1, formatCardInfo(card)),
		})
	}

	return &model.Prompt{
		Type:     model.PromptChooseCards,
		PlayerID: playerID,
		Message:  message,
		Options:  options,
		Min:      min,
		Max:      max,
	}
}

// buildGiveCardsPrompt 构建选牌交给他人提示（天使祝福等）
func (e *GameEngine) buildGiveCardsPrompt() *model.Prompt {
	playerID := e.State.PendingInterrupt.PlayerID
	player := e.State.Players[playerID]
	if player == nil {
		return nil
	}

	data, ok := e.State.PendingInterrupt.Context.(map[string]interface{})
	if !ok {
		return nil
	}

	// giveCount 可能是 int 或 float64（JSON反序列化后）
	var giveCount int
	if gc, ok := data["give_count"].(int); ok {
		giveCount = gc
	} else if gcf, ok := data["give_count"].(float64); ok {
		giveCount = int(gcf)
	}
	receiverID, _ := data["receiver_id"].(string)
	if giveCount <= 0 || receiverID == "" {
		return nil
	}

	receiver := e.State.Players[receiverID]
	receiverName := receiverID
	if receiver != nil {
		receiverName = receiver.Name
	}

	message := fmt.Sprintf("请选择 %d 张牌交给 %s：", giveCount, receiverName)

	var options []model.PromptOption
	for i, card := range player.Hand {
		options = append(options, model.PromptOption{
			ID:    strconv.Itoa(i),
			Label: fmt.Sprintf("%d: %s", i+1, formatCardInfo(card)),
		})
	}

	return &model.Prompt{
		Type:     model.PromptChooseCards,
		PlayerID: playerID,
		Message:  message,
		Options:  options,
		Min:      giveCount,
		Max:      giveCount,
	}
}

// ConfirmGiveCards 确认选牌交给他人（天使祝福等技能）
func (e *GameEngine) ConfirmGiveCards(giverID, receiverID string, indices []int) error {
	if e.State.PendingInterrupt == nil || e.State.PendingInterrupt.Type != model.InterruptGiveCards {
		return fmt.Errorf("当前没有待处理的给牌操作")
	}

	if e.State.PendingInterrupt.PlayerID != giverID {
		return fmt.Errorf("当前不是你的给牌回合")
	}

	data, ok := e.State.PendingInterrupt.Context.(map[string]interface{})
	if !ok {
		return fmt.Errorf("中断上下文错误")
	}

	// giveCount 可能是 int 或 float64（JSON反序列化后）
	var giveCount int
	if gc, ok := data["give_count"].(int); ok {
		giveCount = gc
	} else if gcf, ok := data["give_count"].(float64); ok {
		giveCount = int(gcf)
	}
	ctxReceiverID, _ := data["receiver_id"].(string)
	if ctxReceiverID != receiverID {
		return fmt.Errorf("接收者不匹配")
	}

	giver := e.State.Players[giverID]
	receiver := e.State.Players[receiverID]
	if giver == nil || receiver == nil {
		return fmt.Errorf("玩家不存在")
	}

	if len(indices) != giveCount {
		return fmt.Errorf("需要选择 %d 张牌，你选择了 %d 张", giveCount, len(indices))
	}

	seen := make(map[int]bool)
	for _, idx := range indices {
		if idx < 0 || idx >= len(giver.Hand) {
			return fmt.Errorf("无效的牌索引: %d", idx)
		}
		if seen[idx] {
			return fmt.Errorf("不能重复选择同一张牌")
		}
		seen[idx] = true
	}

	// 从大到小排序，避免移除时索引错乱
	sort.Sort(sort.Reverse(sort.IntSlice(indices)))
	var givenCards []model.Card
	for _, idx := range indices {
		givenCards = append(givenCards, giver.Hand[idx])
		giver.Hand = append(giver.Hand[:idx], giver.Hand[idx+1:]...)
	}

	receiver.Hand = append(receiver.Hand, givenCards...)
	e.Log(fmt.Sprintf("[Skill] %s 将 %d 张牌交给了 %s", giver.Name, len(givenCards), receiver.Name))

	// 检查是否还有更多给牌中断在队列中
	queueLen := len(e.State.InterruptQueue)
	e.Log(fmt.Sprintf("[Debug] 给牌完成，队列中还有 %d 个中断", queueLen))

	e.PopInterrupt()

	// PopInterrupt 后检查新的中断
	if e.State.PendingInterrupt != nil {
		e.Log(fmt.Sprintf("[Debug] 新的中断已设置: Type=%s, PlayerID=%s", e.State.PendingInterrupt.Type, e.State.PendingInterrupt.PlayerID))
	} else {
		e.Log("[Debug] 所有给牌中断已处理完毕")
	}

	return nil
}

// buildChoicePrompt 构建选择提示
func (e *GameEngine) buildChoicePrompt() *model.Prompt {
	// 1. 安全检查
	if e.State.PendingInterrupt == nil {
		return nil
	}

	// 2. 获取上下文数据
	// 我们在 resolveBuffs 里 push 的时候存了 map[string]interface{}{"choice_type": "weak"}
	data, ok := e.State.PendingInterrupt.Context.(map[string]interface{})
	if !ok {
		return nil
	}

	// 3. 判断选择类型
	choiceType, _ := data["choice_type"].(string)
	playerID := e.State.PendingInterrupt.PlayerID
	player := e.State.Players[playerID]

	// === 处理虚弱 (Weak) 逻辑 ===
	if choiceType == "weak" {
		return &model.Prompt{
			Type:     model.PromptConfirm, // 或者叫 PromptChoice，前端显示按钮
			PlayerID: playerID,
			Message:  fmt.Sprintf("【虚弱状态】%s，你需要做出选择：", player.Name),
			// 选项顺序：1=摸3张牌（继续游戏），2=跳过回合（放弃），符合用户直觉
			Options: []model.PromptOption{
				{
					ID:    "0", // choose 1 → 摸3张牌
					Label: "摸3张牌 (移除虚弱)",
				},
				{
					ID:    "1", // choose 2 → 跳过回合
					Label: "跳过回合 (移除虚弱)",
				},
			},
			Min: 1, // 必选其一
			Max: 1,
		}
	}
	// === 处理购买时战绩区4星石选择 (BuyResource) ===
	if choiceType == "buy_resource" {
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "战绩区已有4个星石，选择添加宝石或水晶：",
			Options: []model.PromptOption{
				{ID: "0", Label: "添加宝石"},
				{ID: "1", Label: "添加水晶"},
			},
			Min: 1,
			Max: 1,
		}
	}
	// === 处理治疗选择 (Heal) ===
	if choiceType == "heal" {
		maxHeal, _ := data["max_heal"].(int)
		if maxHeal < 0 {
			maxHeal = 0
		}
		var options []model.PromptOption
		for i := 0; i <= maxHeal; i++ {
			label := fmt.Sprintf("使用 %d 点治疗", i)
			if i == 0 {
				label = "不使用治疗"
			}
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", i),
				Label: label,
			})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  fmt.Sprintf("%s 受到伤害，可选择使用治疗抵消：", player.Name),
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	// === 勇者：怒吼摸牌选择 ===
	if choiceType == "hero_roar_draw" {
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【怒吼】请选择摸牌数量：",
			Options: []model.PromptOption{
				{ID: "0", Label: "摸0张"},
				{ID: "1", Label: "摸1张"},
			},
			Min: 1,
			Max: 1,
		}
	}
	if choiceType == "angel_song_pick" {
		type songPickOption struct {
			ID    string
			Label string
		}
		var raw []songPickOption
		if arr, ok := data["options"].([]songPickOption); ok {
			raw = append(raw, arr...)
		} else if arr, ok := data["options"].([]interface{}); ok {
			for _, v := range arr {
				m, ok := v.(map[string]interface{})
				if !ok || m == nil {
					continue
				}
				id, _ := m["id"].(string)
				label, _ := m["label"].(string)
				if id == "" || label == "" {
					continue
				}
				raw = append(raw, songPickOption{ID: id, Label: label})
			}
		} else if arr, ok := data["options"].([]map[string]interface{}); ok {
			for _, m := range arr {
				if m == nil {
					continue
				}
				id, _ := m["id"].(string)
				label, _ := m["label"].(string)
				if id == "" || label == "" {
					continue
				}
				raw = append(raw, songPickOption{ID: id, Label: label})
			}
		}
		if len(raw) == 0 {
			return nil
		}
		options := make([]model.PromptOption, 0, len(raw))
		for _, item := range raw {
			options = append(options, model.PromptOption{
				ID:    item.ID,
				Label: item.Label,
			})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【天使之歌】请选择要移除的基础效果：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "seal_break_pick_effect" {
		type pickOption struct {
			ID    string
			Label string
		}
		var raw []pickOption
		if arr, ok := data["options"].([]pickOption); ok {
			raw = append(raw, arr...)
		} else if arr, ok := data["options"].([]interface{}); ok {
			for _, v := range arr {
				m, ok := v.(map[string]interface{})
				if !ok || m == nil {
					continue
				}
				id, _ := m["id"].(string)
				label, _ := m["label"].(string)
				if id == "" {
					// 兼容后端仅传 target/effect/field_index 的结构
					tid, _ := m["target_id"].(string)
					eff, _ := m["effect"].(string)
					if tid != "" && eff != "" {
						id = tid + "|" + eff
					}
				}
				if id == "" || label == "" {
					continue
				}
				raw = append(raw, pickOption{ID: id, Label: label})
			}
		} else if arr, ok := data["options"].([]map[string]interface{}); ok {
			for _, m := range arr {
				if m == nil {
					continue
				}
				id, _ := m["id"].(string)
				label, _ := m["label"].(string)
				if id == "" {
					tid, _ := m["target_id"].(string)
					eff, _ := m["effect"].(string)
					if tid != "" && eff != "" {
						id = tid + "|" + eff
					}
				}
				if id == "" || label == "" {
					continue
				}
				raw = append(raw, pickOption{ID: id, Label: label})
			}
		}
		if len(raw) == 0 {
			return nil
		}
		options := make([]model.PromptOption, 0, len(raw))
		for _, item := range raw {
			options = append(options, model.PromptOption{ID: item.ID, Label: item.Label})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【封印破碎】请选择要收回的基础效果：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "angel_bond_heal_target" {
		var targetIDs []string
		if arr, ok := data["target_ids"].([]string); ok {
			targetIDs = append(targetIDs, arr...)
		} else if arr, ok := data["target_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					targetIDs = append(targetIDs, s)
				}
			}
		}
		var options []model.PromptOption
		for _, tid := range targetIDs {
			p := e.State.Players[tid]
			if p == nil {
				continue
			}
			options = append(options, model.PromptOption{ID: tid, Label: p.Name})
		}
		if len(options) == 0 {
			return nil
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【天使羁绊】请选择1名角色获得+1治疗：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "frost_prayer_target" {
		var targetIDs []string
		if arr, ok := data["target_ids"].([]string); ok {
			targetIDs = append(targetIDs, arr...)
		} else if arr, ok := data["target_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					targetIDs = append(targetIDs, s)
				}
			}
		}
		var options []model.PromptOption
		for _, tid := range targetIDs {
			p := e.State.Players[tid]
			if p == nil {
				continue
			}
			options = append(options, model.PromptOption{
				ID:    tid,
				Label: p.Name,
			})
		}
		if len(options) == 0 {
			return nil
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【冰霜祷言】请选择1名角色获得+1治疗：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	// === 女武神：军威神光（二选一）===
	if choiceType == "valkyrie_military_glory_mode" {
		maxX, _ := data["max_x"].(int)
		var options []model.PromptOption
		options = append(options, model.PromptOption{
			ID:    "0",
			Label: "你+1治疗并脱离英灵形态",
		})
		if maxX > 0 {
			options = append(options, model.PromptOption{
				ID:    "1",
				Label: fmt.Sprintf("移除我方战绩区星石（1~%d）并指定角色+X治疗", maxX),
			})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【军神威光】请选择效果：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "arbiter_forced_doomsday_target" {
		var targetIDs []string
		if arr, ok := data["target_ids"].([]string); ok {
			targetIDs = arr
		} else if arr, ok := data["target_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					targetIDs = append(targetIDs, s)
				}
			}
		}
		var options []model.PromptOption
		for _, tid := range targetIDs {
			p := e.State.Players[tid]
			if p == nil {
				continue
			}
			options = append(options, model.PromptOption{ID: tid, Label: p.Name})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【末日审判（强制）】请选择目标角色：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "arbiter_balance_mode" {
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【判决天平】请选择一个分支：",
			Options: []model.PromptOption{
				{ID: "0", Label: "弃掉当前手上的所有手牌"},
				{ID: "1", Label: "将手牌补到上限，并我方战绩区+1红宝石"},
			},
			Min: 1,
			Max: 1,
		}
	}
	if choiceType == "valkyrie_military_glory_x" {
		maxX, _ := data["max_x"].(int)
		var options []model.PromptOption
		for x := 1; x <= maxX; x++ {
			options = append(options, model.PromptOption{
				ID:    strconv.Itoa(x),
				Label: fmt.Sprintf("X=%d", x),
			})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【军神威光】请选择X：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "valkyrie_military_glory_target" {
		var targetIDs []string
		if arr, ok := data["target_ids"].([]string); ok {
			targetIDs = append(targetIDs, arr...)
		} else if arr, ok := data["target_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					targetIDs = append(targetIDs, s)
				}
			}
		}
		var options []model.PromptOption
		for _, tid := range targetIDs {
			p := e.State.Players[tid]
			if p == nil {
				continue
			}
			options = append(options, model.PromptOption{
				ID:    tid,
				Label: p.Name,
			})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【军神威光】请选择目标角色：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	// === 女武神：英灵召唤额外弃法术 ===
	if choiceType == "valkyrie_heroic_extra_confirm" {
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【英灵召唤】是否额外弃1张法术牌并选择任意角色+1治疗？",
			Options: []model.PromptOption{
				{ID: "0", Label: "是"},
				{ID: "1", Label: "否"},
			},
			Min: 1,
			Max: 1,
		}
	}
	if choiceType == "valkyrie_heroic_discard_card" {
		var indices []int
		if arr, ok := data["magic_indices"].([]int); ok {
			indices = arr
		} else if arr, ok := data["magic_indices"].([]interface{}); ok {
			for _, v := range arr {
				if f, ok := v.(float64); ok {
					indices = append(indices, int(f))
				}
			}
		}
		var options []model.PromptOption
		for _, idx := range indices {
			if idx < 0 || idx >= len(player.Hand) {
				continue
			}
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", idx),
				Label: fmt.Sprintf("%d: %s", idx+1, formatCardInfo(player.Hand[idx])),
			})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【英灵召唤】请选择要额外弃置的1张法术牌：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "valkyrie_heroic_heal_target" {
		var targetIDs []string
		if arr, ok := data["target_ids"].([]string); ok {
			targetIDs = arr
		} else if arr, ok := data["target_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					targetIDs = append(targetIDs, s)
				}
			}
		}
		var options []model.PromptOption
		for _, tid := range targetIDs {
			p := e.State.Players[tid]
			if p == nil {
				continue
			}
			options = append(options, model.PromptOption{ID: tid, Label: p.Name})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【英灵召唤】选择1名角色+1治疗：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	// === 元素师：独有技额外弃同系牌 ===
	if choiceType == "elementalist_bonus_confirm" {
		skillName, _ := data["skill_display_name"].(string)
		ele, _ := data["bonus_element"].(string)
		eleZh := elementNameForPrompt(ele)
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  fmt.Sprintf("【%s】是否额外弃1张%s系牌，使本次法术伤害+1？", skillName, eleZh),
			Options: []model.PromptOption{
				{ID: "0", Label: "是"},
				{ID: "1", Label: "否"},
			},
			Min: 1,
			Max: 1,
		}
	}
	if choiceType == "elementalist_bonus_card" {
		var indices []int
		if arr, ok := data["matching_indices"].([]int); ok {
			indices = arr
		} else if arr, ok := data["matching_indices"].([]interface{}); ok {
			for _, v := range arr {
				if f, ok := v.(float64); ok {
					indices = append(indices, int(f))
				}
			}
		}
		var options []model.PromptOption
		for _, idx := range indices {
			if idx < 0 || idx >= len(player.Hand) {
				continue
			}
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", idx),
				Label: fmt.Sprintf("%d: %s", idx+1, formatCardInfo(player.Hand[idx])),
			})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "请选择额外弃置的同系牌：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	// === 冒险家：欺诈 ===
	if choiceType == "adventurer_fraud_mode" {
		can2, _ := data["can2"].(bool)
		can3, _ := data["can3"].(bool)
		var options []model.PromptOption
		if can2 {
			options = append(options, model.PromptOption{ID: "0", Label: "弃2张同系牌，视为非暗灭任意系主动攻击"})
		}
		if can3 {
			options = append(options, model.PromptOption{ID: "1", Label: "弃3张同系牌，视为暗灭主动攻击"})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【欺诈】请选择发动方式：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "adventurer_fraud_attack_element" {
		var options []model.PromptOption
		for _, ele := range []string{
			string(model.ElementWater), string(model.ElementFire), string(model.ElementEarth),
			string(model.ElementWind), string(model.ElementThunder),
		} {
			options = append(options, model.PromptOption{
				ID:    ele,
				Label: fmt.Sprintf("%s系", elementNameForPrompt(ele)),
			})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【欺诈】请选择本次攻击系别（不可选光/暗）：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "adventurer_fraud_discard_element" {
		var options []model.PromptOption
		var elems []string
		if arr, ok := data["discard_elements"].([]string); ok {
			elems = arr
		} else if arr, ok := data["discard_elements"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					elems = append(elems, s)
				}
			}
		}
		for _, ele := range elems {
			options = append(options, model.PromptOption{
				ID:    ele,
				Label: fmt.Sprintf("弃%s系同系2张", elementNameForPrompt(ele)),
			})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【欺诈】请选择用于弃置的同系牌：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "adventurer_fraud_discard_combo" {
		var combos []string
		if arr, ok := data["combos"].([]string); ok {
			combos = arr
		} else if arr, ok := data["combos"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					combos = append(combos, s)
				}
			}
		}
		var options []model.PromptOption
		for _, combo := range combos {
			parts := strings.Split(combo, ":")
			if len(parts) < 2 {
				continue
			}
			ele := parts[0]
			eleZh := elementNameForPrompt(ele)
			label := fmt.Sprintf("%s系 组合", eleZh)
			if len(parts) == 2 {
				idxStrs := strings.Split(parts[1], ",")
				var cardLabels []string
				for _, s := range idxStrs {
					idx, err := strconv.Atoi(s)
					if err != nil || idx < 0 || idx >= len(player.Hand) {
						continue
					}
					cardLabels = append(cardLabels, fmt.Sprintf("%d:%s", idx+1, player.Hand[idx].Name))
				}
				if len(cardLabels) > 0 {
					label = fmt.Sprintf("%s系 [%s]", eleZh, strings.Join(cardLabels, " + "))
				}
			}
			options = append(options, model.PromptOption{ID: combo, Label: label})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【欺诈】请选择要弃置的同系牌组合：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	// === 冒险家：冒险者天堂 ===
	if choiceType == "adventurer_paradise_target" {
		var allyIDs []string
		if arr, ok := data["ally_ids"].([]string); ok {
			allyIDs = arr
		} else if arr, ok := data["ally_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					allyIDs = append(allyIDs, s)
				}
			}
		}
		var options []model.PromptOption
		for _, aid := range allyIDs {
			if p := e.State.Players[aid]; p != nil {
				options = append(options, model.PromptOption{ID: aid, Label: p.Name})
			}
		}
		transferGem := toIntContextValue(data["transfer_gem"])
		transferCrystal := toIntContextValue(data["transfer_crystal"])
		transferTotal := toIntContextValue(data["transfer_total"])
		if transferTotal <= 0 {
			transferTotal = transferGem + transferCrystal
		}
		msg := "【冒险者天堂】请选择接收能量的队友："
		if transferTotal > 0 {
			msg = fmt.Sprintf("【冒险者天堂】请选择接收提炼结果的队友（共%d点：宝石%d / 水晶%d）：", transferTotal, transferGem, transferCrystal)
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  msg,
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	// === 神官：神圣领域 ===
	if choiceType == "priest_divine_contract_x" {
		maxX := toIntContextValue(data["max_x"])
		targetID, _ := data["target_id"].(string)
		targetName := targetID
		targetHeal := -1
		if target := e.State.Players[targetID]; target != nil {
			targetName = target.Name
			targetHeal = target.Heal
		}
		var options []model.PromptOption
		for x := 1; x <= maxX; x++ {
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", x),
				Label: fmt.Sprintf("转移 %d 点治疗", x),
			})
		}
		msg := "【神圣契约】请选择转移治疗值X："
		if targetName != "" {
			if targetHeal >= 0 {
				msg = fmt.Sprintf("【神圣契约】请选择转移治疗值X（目标：%s，当前治疗%d）：", targetName, targetHeal)
			} else {
				msg = fmt.Sprintf("【神圣契约】请选择转移治疗值X（目标：%s）：", targetName)
			}
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  msg,
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "priest_divine_domain_mode" {
		var modeOptions []string
		if arr, ok := data["mode_options"].([]string); ok {
			modeOptions = append(modeOptions, arr...)
		} else if arr, ok := data["mode_options"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					modeOptions = append(modeOptions, s)
				}
			}
		}
		var options []model.PromptOption
		for _, mode := range modeOptions {
			switch mode {
			case "damage":
				options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", len(options)), Label: "分支①：移除1治疗，对任意角色造成2点法术伤害"})
			case "heal":
				options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", len(options)), Label: "分支②：你+2治疗，1名队友+1治疗"})
			}
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【神圣领域】请选择发动分支：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	// === 冒险家：偷天换日 ===
	if choiceType == "adventurer_steal_sky_mode" {
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【偷天换日】请选择效果：",
			Options: []model.PromptOption{
				{ID: "0", Label: "转移对方战绩区1红宝石到我方"},
				{ID: "1", Label: "将我方战绩区全部蓝水晶转换成红宝石"},
			},
			Min: 1,
			Max: 1,
		}
	}
	if choiceType == "adventurer_steal_sky_extra_action" {
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【偷天换日】请选择额外行动类型：",
			Options: []model.PromptOption{
				{ID: "0", Label: "额外+1攻击行动"},
				{ID: "1", Label: "额外+1法术行动"},
			},
			Min: 1,
			Max: 1,
		}
	}
	// === 圣枪骑士：地枪 X ===
	if choiceType == "holy_lancer_earth_spear_x" {
		maxX, _ := data["max_x"].(int)
		if maxX == 0 {
			if f, ok := data["max_x"].(float64); ok {
				maxX = int(f)
			}
		}
		var options []model.PromptOption
		for x := 1; x <= maxX; x++ {
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", x-1),
				Label: fmt.Sprintf("移除%d点治疗，本次伤害+%d", x, x),
			})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【地枪】请选择X值：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	// === 祈祷师：赐福触发 ===
	if choiceType == "prayer_power_blessing_trigger" {
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【威力赐福】是否移除该赐福，使本次攻击伤害+2？",
			Options: []model.PromptOption{
				{ID: "0", Label: "是"},
				{ID: "1", Label: "否"},
			},
			Min: 1,
			Max: 1,
		}
	}
	if choiceType == "prayer_swift_blessing_trigger" {
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【迅捷赐福】是否移除该赐福，获得额外1次攻击行动？",
			Options: []model.PromptOption{
				{ID: "0", Label: "是"},
				{ID: "1", Label: "否"},
			},
			Min: 1,
			Max: 1,
		}
	}
	// === 红莲骑士：血腥祷言 ===
	if choiceType == "crk_bloody_prayer_x" {
		maxX, _ := data["max_x"].(int)
		if maxX == 0 {
			if f, ok := data["max_x"].(float64); ok {
				maxX = int(f)
			}
		}
		var options []model.PromptOption
		for x := 1; x <= maxX; x++ {
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", x),
				Label: fmt.Sprintf("X=%d（移除%d治疗并对自己造成%d法伤）", x, x, x),
			})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【血腥祷言】请选择X值：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "crk_bloody_prayer_ally_count" {
		var allyIDs []string
		if arr, ok := data["ally_ids"].([]string); ok {
			allyIDs = arr
		} else if arr, ok := data["ally_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					allyIDs = append(allyIDs, s)
				}
			}
		}
		x := 1
		if xv, ok := data["x_value"].(int); ok && xv > 0 {
			x = xv
		} else if xf, ok := data["x_value"].(float64); ok && int(xf) > 0 {
			x = int(xf)
		}
		var options []model.PromptOption
		options = append(options, model.PromptOption{ID: "0", Label: "选择1名队友"})
		if len(allyIDs) >= 2 && x >= 2 {
			options = append(options, model.PromptOption{ID: "1", Label: "选择2名队友（治疗将分配）"})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【血腥祷言】请选择要分配治疗的队友数量：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "crk_bloody_prayer_target" {
		var allyIDs []string
		if arr, ok := data["ally_ids"].([]string); ok {
			allyIDs = arr
		} else if arr, ok := data["ally_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					allyIDs = append(allyIDs, s)
				}
			}
		}
		selectedSet := map[string]bool{}
		if arr, ok := data["selected_ally_ids"].([]string); ok {
			for _, s := range arr {
				selectedSet[s] = true
			}
		} else if arr, ok := data["selected_ally_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					selectedSet[s] = true
				}
			}
		}
		allyCount := 1
		if v, ok := data["ally_count"].(int); ok && v > 0 {
			allyCount = v
		} else if f, ok := data["ally_count"].(float64); ok && int(f) > 0 {
			allyCount = int(f)
		}
		var options []model.PromptOption
		for _, aid := range allyIDs {
			if selectedSet[aid] {
				continue
			}
			if p := e.State.Players[aid]; p != nil {
				options = append(options, model.PromptOption{
					ID:    aid,
					Label: p.Name,
				})
			}
		}
		pickIndex := len(selectedSet) + 1
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  fmt.Sprintf("【血腥祷言】请选择第 %d/%d 名队友：", pickIndex, allyCount),
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "crk_bloody_prayer_split" {
		var selected []string
		if arr, ok := data["selected_ally_ids"].([]string); ok {
			selected = append(selected, arr...)
		} else if arr, ok := data["selected_ally_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					selected = append(selected, s)
				}
			}
		}
		if len(selected) != 2 {
			return nil
		}
		x := 0
		if v, ok := data["x_value"].(int); ok {
			x = v
		} else if f, ok := data["x_value"].(float64); ok {
			x = int(f)
		}
		if x < 2 {
			return nil
		}
		p1 := e.State.Players[selected[0]]
		p2 := e.State.Players[selected[1]]
		if p1 == nil || p2 == nil {
			return nil
		}
		var options []model.PromptOption
		for v1 := 1; v1 < x; v1++ {
			v2 := x - v1
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", v1-1),
				Label: fmt.Sprintf("%s +%d，%s +%d", p1.Name, v1, p2.Name, v2),
			})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【血腥祷言】请选择治疗分配：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "crk_calm_mind_action" {
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【戒骄戒躁】请选择额外行动类型：",
			Options: []model.PromptOption{
				{ID: "0", Label: "额外攻击行动"},
				{ID: "1", Label: "额外法术行动"},
			},
			Min: 1,
			Max: 1,
		}
	}
	if choiceType == "hom_rune_reforge_distribution" {
		total := 3
		if v, ok := data["total_runes"].(int); ok && v > 0 {
			total = v
		} else if f, ok := data["total_runes"].(float64); ok && int(f) > 0 {
			total = int(f)
		}
		var options []model.PromptOption
		for war := 0; war <= total; war++ {
			magic := total - war
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", war),
				Label: fmt.Sprintf("战纹 %d / 魔纹 %d", war, magic),
			})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  fmt.Sprintf("【符文改造】请选择战纹/魔纹分配（总计%d）：", total),
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	// === 阴阳师：生命结界 ===
	if choiceType == "onmyoji_life_barrier_mode" {
		ghostFire := 0
		if v, ok := data["ghost_fire"].(int); ok {
			ghostFire = v
		} else if f, ok := data["ghost_fire"].(float64); ok {
			ghostFire = int(f)
		}
		var options []model.PromptOption
		options = append(options, model.PromptOption{
			ID:    "0",
			Label: "分支①：1名队友+1宝石+1治疗，自己承受X点法伤",
		})
		releaseCombos := 0
		if arr, ok := data["release_card_combos"].([]string); ok {
			releaseCombos = len(arr)
		} else if arr, ok := data["release_card_combos"].([]interface{}); ok {
			releaseCombos = len(arr)
		}
		if releaseCombos > 0 {
			options = append(options, model.PromptOption{
				ID:    "1",
				Label: "分支②：弃2张同命格手牌并脱离式神形态，令1名队友弃1张手牌",
			})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  fmt.Sprintf("【生命结界】当前鬼火=%d，请选择发动分支：", ghostFire),
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "onmyoji_life_barrier_release_combo" {
		var combos []string
		if arr, ok := data["release_card_combos"].([]string); ok {
			combos = append(combos, arr...)
		} else if arr, ok := data["release_card_combos"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					combos = append(combos, s)
				}
			}
		}
		var options []model.PromptOption
		for _, combo := range combos {
			parts := strings.Split(combo, ",")
			if len(parts) != 2 {
				continue
			}
			i, err1 := strconv.Atoi(parts[0])
			j, err2 := strconv.Atoi(parts[1])
			if err1 != nil || err2 != nil || i < 0 || j < 0 || i >= len(player.Hand) || j >= len(player.Hand) || i == j {
				continue
			}
			options = append(options, model.PromptOption{
				ID:    combo,
				Label: fmt.Sprintf("%d:%s + %d:%s（%s命格）", i+1, player.Hand[i].Name, j+1, player.Hand[j].Name, player.Hand[i].Faction),
			})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【生命结界·分支②】请选择要弃置的2张同命格手牌：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	// === 阴阳师：式神咒束 ===
	if choiceType == "onmyoji_yinyang_confirm" {
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【阴阳转换】你可使用同命格攻击牌应战，是否发动？",
			Options: []model.PromptOption{
				{ID: "0", Label: "是"},
				{ID: "1", Label: "否"},
			},
			Min: 1,
			Max: 1,
		}
	}
	if choiceType == "onmyoji_yinyang_card" {
		var rawOptions []map[string]interface{}
		if arr, ok := data["card_options"].([]map[string]interface{}); ok {
			rawOptions = append(rawOptions, arr...)
		} else if arr, ok := data["card_options"].([]interface{}); ok {
			for _, item := range arr {
				if m, ok := item.(map[string]interface{}); ok && m != nil {
					rawOptions = append(rawOptions, m)
				}
			}
		}
		var options []model.PromptOption
		for _, m := range rawOptions {
			cardID, _ := m["card_id"].(string)
			label, _ := m["label"].(string)
			if cardID == "" || label == "" {
				continue
			}
			options = append(options, model.PromptOption{ID: cardID, Label: label})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【阴阳转换】请选择用于同命格应战的攻击牌：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "onmyoji_yinyang_counter_target" {
		var targetIDs []string
		if arr, ok := data["counter_target_ids"].([]string); ok {
			targetIDs = arr
		} else if arr, ok := data["counter_target_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					targetIDs = append(targetIDs, s)
				}
			}
		}
		var options []model.PromptOption
		for _, tid := range targetIDs {
			if p := e.State.Players[tid]; p != nil {
				options = append(options, model.PromptOption{ID: tid, Label: p.Name})
			}
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【阴阳转换】请选择应战反弹目标：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "onmyoji_binding_confirm" {
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【式神咒束】是否代替队友执行应战？",
			Options: []model.PromptOption{
				{ID: "0", Label: "是"},
				{ID: "1", Label: "否"},
			},
			Min: 1,
			Max: 1,
		}
	}
	if choiceType == "onmyoji_binding_card" {
		var rawOptions []map[string]interface{}
		if arr, ok := data["card_options"].([]map[string]interface{}); ok {
			rawOptions = append(rawOptions, arr...)
		} else if arr, ok := data["card_options"].([]interface{}); ok {
			for _, item := range arr {
				if m, ok := item.(map[string]interface{}); ok && m != nil {
					rawOptions = append(rawOptions, m)
				}
			}
		}
		var options []model.PromptOption
		for _, m := range rawOptions {
			cardID, _ := m["card_id"].(string)
			label, _ := m["label"].(string)
			if cardID == "" || label == "" {
				continue
			}
			options = append(options, model.PromptOption{ID: cardID, Label: label})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【式神咒束】请选择用于代应战的攻击牌：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "onmyoji_binding_counter_target" {
		var targetIDs []string
		if arr, ok := data["counter_target_ids"].([]string); ok {
			targetIDs = arr
		} else if arr, ok := data["counter_target_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					targetIDs = append(targetIDs, s)
				}
			}
		}
		var options []model.PromptOption
		for _, tid := range targetIDs {
			if p := e.State.Players[tid]; p != nil {
				options = append(options, model.PromptOption{ID: tid, Label: p.Name})
			}
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【式神咒束】请选择应战反弹目标：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	// === 英灵人形：战纹碎击 / 魔纹融合 ===
	if choiceType == "hom_rune_smash_x" || choiceType == "hom_glyph_fusion_x" {
		maxX := 0
		if v, ok := data["max_x"].(int); ok {
			maxX = v
		} else if f, ok := data["max_x"].(float64); ok {
			maxX = int(f)
		}
		minX := 1
		if choiceType == "hom_glyph_fusion_x" {
			minX = 2
		}
		var options []model.PromptOption
		for x := minX; x <= maxX; x++ {
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", x),
				Label: fmt.Sprintf("X=%d", x),
			})
		}
		msg := "【战纹碎击】请选择X（弃置同系牌数量）："
		if choiceType == "hom_glyph_fusion_x" {
			msg = "【魔纹融合】请选择X（弃置异系且元素互不相同的牌数量）："
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  msg,
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "hom_rune_smash_cards" || choiceType == "hom_glyph_fusion_cards" {
		var remaining []int
		if arr, ok := data["remaining_indices"].([]int); ok {
			remaining = arr
		} else if arr, ok := data["remaining_indices"].([]interface{}); ok {
			for _, v := range arr {
				if f, ok := v.(float64); ok {
					remaining = append(remaining, int(f))
				}
			}
		}
		xVal := 0
		if v, ok := data["x_value"].(int); ok {
			xVal = v
		} else if f, ok := data["x_value"].(float64); ok {
			xVal = int(f)
		}
		selectedCount := 0
		if arr, ok := data["selected_indices"].([]int); ok {
			selectedCount = len(arr)
		} else if arr, ok := data["selected_indices"].([]interface{}); ok {
			selectedCount = len(arr)
		}
		var options []model.PromptOption
		for _, idx := range remaining {
			if idx < 0 || idx >= len(player.Hand) {
				continue
			}
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", idx),
				Label: fmt.Sprintf("%d: %s", idx+1, formatCardInfo(player.Hand[idx])),
			})
		}
		msg := fmt.Sprintf("【战纹碎击】请选择第 %d/%d 张弃牌：", selectedCount+1, xVal)
		if choiceType == "hom_glyph_fusion_cards" {
			msg = fmt.Sprintf("【魔纹融合】请选择第 %d/%d 张弃牌（元素不可重复）：", selectedCount+1, xVal)
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  msg,
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "hom_rune_smash_y" || choiceType == "hom_glyph_fusion_y" {
		maxY := 0
		if v, ok := data["max_y"].(int); ok {
			maxY = v
		} else if f, ok := data["max_y"].(float64); ok {
			maxY = int(f)
		}
		var options []model.PromptOption
		for y := 0; y <= maxY; y++ {
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", y),
				Label: fmt.Sprintf("Y=%d", y),
			})
		}
		msg := "【战纹碎击】请选择Y（额外翻转战纹数）："
		if choiceType == "hom_glyph_fusion_y" {
			msg = "【魔纹融合】请选择Y（额外翻转魔纹数）："
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  msg,
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}

	// === 处理提炼选择 (Extract) ===
	if choiceType == "extract" {
		optsRaw, _ := data["extract_options"]
		optsIfaces, ok := optsRaw.([]interface{})
		if !ok {
			return nil
		}
		minSel := 1
		maxSel := 2
		if m, ok := data["extract_min"].(int); ok && m > 0 {
			minSel = m
		}
		if m, ok := data["extract_max"].(int); ok && m > 0 {
			maxSel = m
		}
		var options []model.PromptOption
		for i, o := range optsIfaces {
			om, _ := o.(map[string]interface{})
			if om == nil {
				continue
			}
			typ, _ := om["type"].(string)
			label := "红宝石"
			if typ == "crystal" {
				label = "蓝水晶"
			}
			options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", i), Label: label})
		}
		// 战绩区摘要
		msg := "战绩区可提炼的星石：请选择 1-2 个提炼到能量区（点击切换选择）："
		if len(options) > 0 {
			msg = fmt.Sprintf("战绩区可提炼的星石（共 %d 个）：请选择 %d-%d 个提炼到能量区：", len(options), minSel, maxSel)
		}
		return &model.Prompt{
			Type:     model.PromptChooseExtract,
			PlayerID: playerID,
			Message:  msg,
			Options:  options,
			Min:      minSel,
			Max:      maxSel,
		}
	}
	// === 处理五系束缚选择 (FiveElementsBind) ===
	if choiceType == "five_elements_bind" {
		drawCountAny, _ := data["draw_count"]
		drawCount, ok := drawCountAny.(int)
		if !ok {
			if f, ok := drawCountAny.(float64); ok {
				drawCount = int(f)
			}
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  fmt.Sprintf("【五系束缚】%s，你需要做出选择：", player.Name),
			Options: []model.PromptOption{
				{
					ID:    "0",
					Label: fmt.Sprintf("摸 %d 张牌 (继续行动)", drawCount),
				},
				{
					ID:    "1",
					Label: "放弃行动 (移除五系束缚)",
				},
			},
			Min: 1,
			Max: 1,
		}
	}
	// === 精灵射手：元素射击 ===
	if choiceType == "elf_elemental_shot_cost" {
		canMagic, _ := data["can_discard_magic"].(bool)
		canBless, _ := data["can_remove_bless"].(bool)
		var options []model.PromptOption
		if canMagic {
			options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", len(options)), Label: "弃1张法术牌发动"})
		}
		if canBless {
			options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", len(options)), Label: "移除1个祝福发动"})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【元素射击】请选择发动消耗：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "elf_elemental_shot_discard_magic" {
		var idxs []int
		if arr, ok := data["magic_indices"].([]int); ok {
			idxs = arr
		} else if arr, ok := data["magic_indices"].([]interface{}); ok {
			for _, v := range arr {
				if f, ok := v.(float64); ok {
					idxs = append(idxs, int(f))
				}
			}
		}
		var options []model.PromptOption
		for _, idx := range idxs {
			if idx < 0 || idx >= len(player.Hand) {
				continue
			}
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", idx),
				Label: fmt.Sprintf("%d: %s", idx+1, formatCardInfo(player.Hand[idx])),
			})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【元素射击】请选择弃置的法术牌：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "elf_elemental_shot_remove_blessing" {
		var idxs []int
		if arr, ok := data["blessing_indices"].([]int); ok {
			idxs = arr
		} else if arr, ok := data["blessing_indices"].([]interface{}); ok {
			for _, v := range arr {
				if f, ok := v.(float64); ok {
					idxs = append(idxs, int(f))
				}
			}
		}
		var options []model.PromptOption
		for _, idx := range idxs {
			if idx < 0 || idx >= len(player.Blessings) {
				continue
			}
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", idx),
				Label: fmt.Sprintf("%d: %s", idx+1, formatCardInfo(player.Blessings[idx])),
			})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【元素射击】请选择要移除的祝福：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "elf_animal_companion_confirm" {
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【动物伙伴】是否发动（摸1弃1）？",
			Options: []model.PromptOption{
				{ID: "0", Label: "是"},
				{ID: "1", Label: "否"},
			},
			Min: 1,
			Max: 1,
		}
	}
	if choiceType == "elf_pet_empower_confirm" {
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【宠物强化】是否消耗1蓝水晶，将效果改为任意角色摸1弃1？",
			Options: []model.PromptOption{
				{ID: "0", Label: "是"},
				{ID: "1", Label: "否"},
			},
			Min: 1,
			Max: 1,
		}
	}
	// === 瘟疫法师：死亡之触 ===
	if choiceType == "plague_death_touch_element" {
		var elements []string
		if arr, ok := data["elements"].([]string); ok {
			elements = arr
		} else if arr, ok := data["elements"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					elements = append(elements, s)
				}
			}
		}
		var options []model.PromptOption
		for i, ele := range elements {
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", i),
				Label: fmt.Sprintf("%s系", elementNameForPrompt(ele)),
			})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【死亡之触】请选择弃置同系牌的元素：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "plague_death_touch_x" {
		maxHeal := 0
		if v, ok := data["max_heal"].(int); ok {
			maxHeal = v
		} else if f, ok := data["max_heal"].(float64); ok {
			maxHeal = int(f)
		}
		var options []model.PromptOption
		for x := 2; x <= maxHeal; x++ {
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", x-2),
				Label: fmt.Sprintf("X=%d（移除%d点治疗）", x, x),
			})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【死亡之触】请选择X值：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "plague_death_touch_y" {
		maxCards := 0
		if v, ok := data["max_cards"].(int); ok {
			maxCards = v
		} else if f, ok := data["max_cards"].(float64); ok {
			maxCards = int(f)
		}
		var options []model.PromptOption
		for y := 2; y <= maxCards; y++ {
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", y-2),
				Label: fmt.Sprintf("Y=%d（弃%d张同系牌）", y, y),
			})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【死亡之触】请选择Y值：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "plague_death_touch_cards" {
		var remaining []int
		if arr, ok := data["remaining_indices"].([]int); ok {
			remaining = arr
		} else if arr, ok := data["remaining_indices"].([]interface{}); ok {
			for _, v := range arr {
				if f, ok := v.(float64); ok {
					remaining = append(remaining, int(f))
				}
			}
		}
		yNeed := 0
		if v, ok := data["y_value"].(int); ok {
			yNeed = v
		} else if f, ok := data["y_value"].(float64); ok {
			yNeed = int(f)
		}
		selectedCount := 0
		if arr, ok := data["selected_indices"].([]int); ok {
			selectedCount = len(arr)
		} else if arr, ok := data["selected_indices"].([]interface{}); ok {
			selectedCount = len(arr)
		}
		var options []model.PromptOption
		for _, idx := range remaining {
			if idx < 0 || idx >= len(player.Hand) {
				continue
			}
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", idx),
				Label: fmt.Sprintf("%d: %s", idx+1, formatCardInfo(player.Hand[idx])),
			})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  fmt.Sprintf("【死亡之触】请选择第 %d/%d 张弃牌：", selectedCount+1, yNeed),
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	// === 魔剑士：暗影流星 ===
	if choiceType == "ms_shadow_meteor_discard" {
		var remaining []int
		if arr, ok := data["remaining_indices"].([]int); ok {
			remaining = arr
		} else if arr, ok := data["remaining_indices"].([]interface{}); ok {
			for _, v := range arr {
				if f, ok := v.(float64); ok {
					remaining = append(remaining, int(f))
				}
			}
		}
		if len(remaining) == 0 {
			if arr, ok := data["magic_indices"].([]int); ok {
				remaining = arr
			} else if arr, ok := data["magic_indices"].([]interface{}); ok {
				for _, v := range arr {
					if f, ok := v.(float64); ok {
						remaining = append(remaining, int(f))
					}
				}
			}
		}
		selectedCount := 0
		if arr, ok := data["selected_indices"].([]int); ok {
			selectedCount = len(arr)
		} else if arr, ok := data["selected_indices"].([]interface{}); ok {
			selectedCount = len(arr)
		}
		var options []model.PromptOption
		for _, idx := range remaining {
			if idx < 0 || idx >= len(player.Hand) {
				continue
			}
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", idx),
				Label: fmt.Sprintf("%d: %s", idx+1, formatCardInfo(player.Hand[idx])),
			})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  fmt.Sprintf("【暗影流星】请选择第 %d/2 张法术牌：", selectedCount+1),
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "ms_shadow_meteor_release_confirm" {
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【暗影流星】是否额外移除我方战绩区2个星石，转正并+1红宝石？",
			Options: []model.PromptOption{
				{ID: "0", Label: "是"},
				{ID: "1", Label: "否"},
			},
			Min: 1,
			Max: 1,
		}
	}
	// === 血色剑灵 ===
	if choiceType == "css_blood_barrier_counter_confirm" {
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【血气屏障】是否额外对一名对手造成1点法术伤害？",
			Options: []model.PromptOption{
				{ID: "0", Label: "是"},
				{ID: "1", Label: "否"},
				{ID: "cancel", Label: "取消"},
			},
			Min: 1,
			Max: 1,
		}
	}
	if choiceType == "css_dance_mode" {
		canCrystal, _ := data["can_crystal"].(bool)
		canGem, _ := data["can_gem"].(bool)
		var options []model.PromptOption
		if canCrystal {
			options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", len(options)), Label: "消耗1蓝水晶（可用红宝石替代）：放置庭院并+2鲜血"})
		}
		if canGem {
			options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", len(options)), Label: "消耗1红宝石：放置庭院并+2鲜血（上限4）且弃牌至4"})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【散华轮舞】请选择发动分支：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	// === 苍炎魔女 ===
	if choiceType == "bw_witch_wrath_draw" {
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【魔女之怒】请选择摸牌数量：",
			Options: []model.PromptOption{
				{ID: "0", Label: "摸0张"},
				{ID: "1", Label: "摸1张"},
				{ID: "2", Label: "摸2张"},
			},
			Min: 1,
			Max: 1,
		}
	}
	if choiceType == "bw_substitute_doll_card" {
		var magicIndices []int
		if arr, ok := data["magic_indices"].([]int); ok {
			magicIndices = append(magicIndices, arr...)
		} else if arr, ok := data["magic_indices"].([]interface{}); ok {
			for _, v := range arr {
				if f, ok := v.(float64); ok {
					magicIndices = append(magicIndices, int(f))
				}
			}
		}
		var options []model.PromptOption
		for _, idx := range magicIndices {
			if idx < 0 || idx >= len(player.Hand) {
				continue
			}
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", idx),
				Label: fmt.Sprintf("%d: %s", idx+1, formatCardInfo(player.Hand[idx])),
			})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【替身玩偶】请选择弃置1张法术牌：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "bw_mana_inversion_x" {
		maxX := toIntContextValue(data["max_x"])
		var options []model.PromptOption
		for x := 2; x <= maxX; x++ {
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", x-2),
				Label: fmt.Sprintf("X=%d（弃%d张法术牌，造成%d点法术伤害）", x, x, x-1),
			})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【魔能反转】请选择X值：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "bw_mana_inversion_cards" {
		var remaining []int
		if arr, ok := data["remaining_indices"].([]int); ok {
			remaining = append(remaining, arr...)
		} else if arr, ok := data["remaining_indices"].([]interface{}); ok {
			for _, v := range arr {
				if f, ok := v.(float64); ok {
					remaining = append(remaining, int(f))
				}
			}
		}
		selectedCount := 0
		if arr, ok := data["selected_indices"].([]int); ok {
			selectedCount = len(arr)
		} else if arr, ok := data["selected_indices"].([]interface{}); ok {
			selectedCount = len(arr)
		}
		targetCount := toIntContextValue(data["x_value"])
		var options []model.PromptOption
		for _, idx := range remaining {
			if idx < 0 || idx >= len(player.Hand) {
				continue
			}
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", idx),
				Label: fmt.Sprintf("%d: %s", idx+1, formatCardInfo(player.Hand[idx])),
			})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  fmt.Sprintf("【魔能反转】请选择第 %d/%d 张法术牌：", selectedCount+1, targetCount),
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	// === 贤者 ===
	if choiceType == "sage_wisdom_codex_discard_confirm" {
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【智慧法典】是否额外弃置1张手牌？",
			Options: []model.PromptOption{
				{ID: "0", Label: "是"},
				{ID: "1", Label: "否"},
			},
			Min: 1,
			Max: 1,
		}
	}
	if choiceType == "sage_magic_rebound_confirm" {
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【法术反弹】是否发动？",
			Options: []model.PromptOption{
				{ID: "0", Label: "是"},
				{ID: "1", Label: "否"},
			},
			Min: 1,
			Max: 1,
		}
	}
	if choiceType == "sage_magic_rebound_x" {
		maxX := toIntContextValue(data["max_x"])
		var options []model.PromptOption
		for x := 2; x <= maxX; x++ {
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", x-2),
				Label: fmt.Sprintf("X=%d（弃%d张同系牌）", x, x),
			})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【法术反弹】请选择X值：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "sage_magic_rebound_element" {
		xValue := toIntContextValue(data["x_value"])
		elements := availableElementsByMinCount(player, xValue)
		var options []model.PromptOption
		for i, ele := range elements {
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", i),
				Label: fmt.Sprintf("%s系", elementNameForPrompt(ele)),
			})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【法术反弹】请选择弃置同系牌的元素：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "sage_magic_rebound_cards" || choiceType == "sage_arcane_cards" || choiceType == "sage_holy_cards" {
		var remaining []int
		if arr, ok := data["remaining_indices"].([]int); ok {
			remaining = append(remaining, arr...)
		} else if arr, ok := data["remaining_indices"].([]interface{}); ok {
			for _, v := range arr {
				if f, ok := v.(float64); ok {
					remaining = append(remaining, int(f))
				}
			}
		}
		selectedCount := 0
		if arr, ok := data["selected_indices"].([]int); ok {
			selectedCount = len(arr)
		} else if arr, ok := data["selected_indices"].([]interface{}); ok {
			selectedCount = len(arr)
		}
		targetCount := toIntContextValue(data["x_value"])
		var options []model.PromptOption
		for _, idx := range remaining {
			if idx < 0 || idx >= len(player.Hand) {
				continue
			}
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", idx),
				Label: fmt.Sprintf("%d: %s", idx+1, formatCardInfo(player.Hand[idx])),
			})
		}
		msg := fmt.Sprintf("请选择第 %d/%d 张牌：", selectedCount+1, targetCount)
		switch choiceType {
		case "sage_magic_rebound_cards":
			msg = fmt.Sprintf("【法术反弹】请选择第 %d/%d 张同系牌：", selectedCount+1, targetCount)
		case "sage_arcane_cards":
			msg = fmt.Sprintf("【魔道法典】请选择第 %d/%d 张异系牌：", selectedCount+1, targetCount)
		case "sage_holy_cards":
			msg = fmt.Sprintf("【圣洁法典】请选择第 %d/%d 张异系牌：", selectedCount+1, targetCount)
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  msg,
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "sage_arcane_x" || choiceType == "sage_holy_x" {
		maxX := toIntContextValue(data["max_x"])
		minX := 2
		if choiceType == "sage_holy_x" {
			minX = 3
		}
		var options []model.PromptOption
		for x := minX; x <= maxX; x++ {
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", x-minX),
				Label: fmt.Sprintf("X=%d（弃%d张异系牌）", x, x),
			})
		}
		msg := "【魔道法典】请选择X值："
		if choiceType == "sage_holy_x" {
			msg = "【圣洁法典】请选择X值："
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  msg,
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "sage_holy_target_count" {
		maxCount := toIntContextValue(data["max_target_count"])
		if maxCount < 0 {
			maxCount = 0
		}
		var options []model.PromptOption
		for c := 0; c <= maxCount; c++ {
			label := fmt.Sprintf("选择%d名角色", c)
			if c == 0 {
				label = "不选择角色"
			}
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", c),
				Label: label,
			})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【圣洁法典】请选择要获得治疗的角色数量：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "sage_holy_targets" {
		var targetIDs []string
		if arr, ok := data["target_ids"].([]string); ok {
			targetIDs = append(targetIDs, arr...)
		} else if arr, ok := data["target_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					targetIDs = append(targetIDs, s)
				}
			}
		}
		selectedSet := map[string]bool{}
		if arr, ok := data["selected_target_ids"].([]string); ok {
			for _, s := range arr {
				selectedSet[s] = true
			}
		} else if arr, ok := data["selected_target_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					selectedSet[s] = true
				}
			}
		}
		targetCount := toIntContextValue(data["target_count"])
		var options []model.PromptOption
		for _, tid := range targetIDs {
			if selectedSet[tid] {
				continue
			}
			if p := e.State.Players[tid]; p != nil {
				options = append(options, model.PromptOption{
					ID:    fmt.Sprintf("%d", len(options)),
					Label: p.Name,
				})
			}
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  fmt.Sprintf("【圣洁法典】请选择第 %d/%d 名治疗目标：", len(selectedSet)+1, targetCount),
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	// === 魔弓 ===
	if choiceType == "mb_magic_pierce_hit_confirm" {
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【魔贯冲击】攻击命中，是否额外移除1个火系充能使本次伤害再+1？",
			Options: []model.PromptOption{
				{ID: "0", Label: "是"},
				{ID: "1", Label: "否"},
			},
			Min: 1,
			Max: 1,
		}
	}
	if choiceType == "mb_charge_draw_x" {
		maxDraw := toIntContextValue(data["max_draw"])
		if maxDraw <= 0 {
			maxDraw = 4
		}
		var options []model.PromptOption
		for x := 0; x <= maxDraw; x++ {
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", x),
				Label: fmt.Sprintf("X=%d（摸%d张）", x, x),
			})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【充能】请选择摸牌数量X（0~4）：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "mb_charge_place_count" {
		maxPlace := toIntContextValue(data["max_place"])
		if maxPlace < 0 {
			maxPlace = 0
		}
		var options []model.PromptOption
		for c := 0; c <= maxPlace; c++ {
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", c),
				Label: fmt.Sprintf("放置%d张充能", c),
			})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【充能】请选择要放置为充能的手牌数量：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "mb_charge_place_cards" || choiceType == "mb_demon_eye_charge_card" {
		var remaining []int
		if arr, ok := data["remaining_indices"].([]int); ok {
			remaining = append(remaining, arr...)
		} else if arr, ok := data["remaining_indices"].([]interface{}); ok {
			for _, v := range arr {
				if f, ok := v.(float64); ok {
					remaining = append(remaining, int(f))
				}
			}
		}
		if len(remaining) == 0 && choiceType == "mb_demon_eye_charge_card" {
			remaining = allHandIndices(player)
		}
		selectedCount := 0
		if arr, ok := data["selected_indices"].([]int); ok {
			selectedCount = len(arr)
		} else if arr, ok := data["selected_indices"].([]interface{}); ok {
			selectedCount = len(arr)
		}
		needCount := toIntContextValue(data["need_count"])
		if choiceType == "mb_demon_eye_charge_card" && needCount <= 0 {
			needCount = 1
		}
		if needCount <= 0 {
			needCount = 1
		}
		var options []model.PromptOption
		for _, idx := range remaining {
			if idx < 0 || idx >= len(player.Hand) {
				continue
			}
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", idx),
				Label: fmt.Sprintf("%d: %s", idx+1, formatCardInfo(player.Hand[idx])),
			})
		}
		msg := fmt.Sprintf("【充能】请选择第 %d/%d 张作为充能的手牌：", selectedCount+1, needCount)
		if choiceType == "mb_demon_eye_charge_card" {
			msg = "【魔眼】请选择1张手牌作为充能："
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  msg,
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "mb_thunder_scatter_extra" {
		maxExtra := toIntContextValue(data["max_extra"])
		if maxExtra < 0 {
			maxExtra = 0
		}
		var options []model.PromptOption
		for x := 0; x <= maxExtra; x++ {
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", x),
				Label: fmt.Sprintf("额外移除%d个雷系充能", x),
			})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【雷光散射】请选择额外移除雷系充能数量X：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "mb_demon_eye_mode" {
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【魔眼】请选择发动分支：",
			Options: []model.PromptOption{
				{ID: "0", Label: "选择1名角色弃1张牌"},
				{ID: "1", Label: "你摸3张牌"},
			},
			Min: 1,
			Max: 1,
		}
	}
	// === 圣弓 ===
	if choiceType == "hb_holy_shard_combo" {
		var combos []string
		if arr, ok := data["combos"].([]string); ok {
			combos = append(combos, arr...)
		} else if arr, ok := data["combos"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					combos = append(combos, s)
				}
			}
		}
		var options []model.PromptOption
		for _, combo := range combos {
			parts := strings.Split(combo, ":")
			if len(parts) != 2 {
				continue
			}
			idxParts := strings.Split(parts[1], ",")
			if len(idxParts) != 2 {
				continue
			}
			i, err1 := strconv.Atoi(strings.TrimSpace(idxParts[0]))
			j, err2 := strconv.Atoi(strings.TrimSpace(idxParts[1]))
			if err1 != nil || err2 != nil || i < 0 || j < 0 || i >= len(player.Hand) || j >= len(player.Hand) || i == j {
				continue
			}
			options = append(options, model.PromptOption{
				ID:    combo,
				Label: fmt.Sprintf("%s系：%d:%s + %d:%s", elementNameForPrompt(parts[0]), i+1, player.Hand[i].Name, j+1, player.Hand[j].Name),
			})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【圣屑飓暴】请选择要弃置的2张同系攻击牌：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "hb_holy_shard_target" {
		var targetIDs []string
		if arr, ok := data["target_ids"].([]string); ok {
			targetIDs = append(targetIDs, arr...)
		} else if arr, ok := data["target_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					targetIDs = append(targetIDs, s)
				}
			}
		}
		var options []model.PromptOption
		for _, tid := range targetIDs {
			if p := e.State.Players[tid]; p != nil {
				options = append(options, model.PromptOption{ID: tid, Label: p.Name})
			}
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【圣屑飓暴】请选择主动攻击目标：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "hb_holy_shard_miss_confirm" {
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【圣屑飓暴】未命中：是否移除治疗并令1名队友弃牌？",
			Options: []model.PromptOption{
				{ID: "0", Label: "是"},
				{ID: "1", Label: "否"},
			},
			Min: 1,
			Max: 1,
		}
	}
	if choiceType == "hb_holy_shard_miss_x" {
		maxX := toIntContextValue(data["max_x"])
		if maxX <= 0 {
			maxX = 1
		}
		var options []model.PromptOption
		for x := 1; x <= maxX; x++ {
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", x),
				Label: fmt.Sprintf("移除%d点治疗，并令队友弃%d张牌", x, x),
			})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【圣屑飓暴】请选择移除治疗点数X：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "hb_holy_shard_miss_ally_target" {
		var allyIDs []string
		if arr, ok := data["ally_ids"].([]string); ok {
			allyIDs = append(allyIDs, arr...)
		} else if arr, ok := data["ally_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					allyIDs = append(allyIDs, s)
				}
			}
		}
		var options []model.PromptOption
		for _, aid := range allyIDs {
			if p := e.State.Players[aid]; p != nil {
				options = append(options, model.PromptOption{ID: aid, Label: p.Name})
			}
		}
		x := toIntContextValue(data["x_value"])
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  fmt.Sprintf("【圣屑飓暴】请选择1名队友弃置%d张手牌：", x),
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "hb_radiant_descent_cost" {
		var modes []string
		if arr, ok := data["cost_modes"].([]string); ok {
			modes = append(modes, arr...)
		} else if arr, ok := data["cost_modes"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					modes = append(modes, s)
				}
			}
		}
		var options []model.PromptOption
		for _, m := range modes {
			switch m {
			case "heal":
				options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", len(options)), Label: "移除2点治疗"})
			case "faith":
				options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", len(options)), Label: "移除2点信仰"})
			}
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【圣煌降临】请选择支付方式：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "hb_light_burst_mode" {
		var allyIDs []string
		if arr, ok := data["ally_ids"].([]string); ok {
			allyIDs = append(allyIDs, arr...)
		} else if arr, ok := data["ally_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					allyIDs = append(allyIDs, s)
				}
			}
		}
		var enemyIDs []string
		if arr, ok := data["enemy_ids"].([]string); ok {
			enemyIDs = append(enemyIDs, arr...)
		} else if arr, ok := data["enemy_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					enemyIDs = append(enemyIDs, s)
				}
			}
		}
		maxX := toIntContextValue(data["max_x"])
		canModeA := player.Heal >= 1 && len(allyIDs) > 0
		canModeB := false
		if maxX > 0 && len(enemyIDs) > 0 {
			handCount := len(player.Hand)
			for x := 1; x <= maxX; x++ {
				limit := handCount - x
				eligible := 0
				for _, eid := range enemyIDs {
					if ep := e.State.Players[eid]; ep != nil && len(ep.Hand) <= limit {
						eligible++
					}
				}
				if eligible > 0 {
					canModeB = true
					break
				}
			}
		}
		var options []model.PromptOption
		if canModeA {
			options = append(options, model.PromptOption{ID: "0", Label: "分支①：摸1、移除1治疗、+1信仰、我方1人+1治疗"})
		}
		if canModeB {
			options = append(options, model.PromptOption{ID: "1", Label: "分支②：移除X治疗并弃X牌，至多X名对手各受攻击伤害"})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【圣光爆裂】请选择发动分支：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "hb_light_burst_mode_a_target" || choiceType == "hb_meteor_bullet_target" {
		var allyIDs []string
		if arr, ok := data["ally_ids"].([]string); ok {
			allyIDs = append(allyIDs, arr...)
		} else if arr, ok := data["ally_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					allyIDs = append(allyIDs, s)
				}
			}
		}
		var options []model.PromptOption
		for _, aid := range allyIDs {
			if p := e.State.Players[aid]; p != nil {
				options = append(options, model.PromptOption{ID: aid, Label: p.Name})
			}
		}
		msg := "请选择我方角色："
		if choiceType == "hb_light_burst_mode_a_target" {
			msg = "【圣光爆裂】分支①请选择获得治疗的我方角色："
		} else {
			msg = "【流星圣弹】请选择获得治疗的我方角色："
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  msg,
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "hb_light_burst_mode_b_x" {
		var enemyIDs []string
		if arr, ok := data["enemy_ids"].([]string); ok {
			enemyIDs = append(enemyIDs, arr...)
		} else if arr, ok := data["enemy_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					enemyIDs = append(enemyIDs, s)
				}
			}
		}
		maxX := toIntContextValue(data["max_x"])
		var options []model.PromptOption
		for x := 1; x <= maxX; x++ {
			limit := len(player.Hand) - x
			eligible := 0
			for _, eid := range enemyIDs {
				if ep := e.State.Players[eid]; ep != nil && len(ep.Hand) <= limit {
					eligible++
				}
			}
			if eligible <= 0 {
				continue
			}
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", x),
				Label: fmt.Sprintf("X=%d（移除%d治疗并弃%d张牌）", x, x, x),
			})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【圣光爆裂】分支②请选择X值：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "hb_light_burst_mode_b_target_count" {
		x := toIntContextValue(data["x_value"])
		eligible := toIntContextValue(data["eligible_count"])
		maxCount := x
		if eligible < maxCount {
			maxCount = eligible
		}
		if maxCount < 1 {
			maxCount = 1
		}
		var options []model.PromptOption
		for c := 1; c <= maxCount; c++ {
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", c),
				Label: fmt.Sprintf("选择%d名对手", c),
			})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【圣光爆裂】分支②请选择目标人数：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "hb_light_burst_mode_b_targets" {
		var candidates []string
		if arr, ok := data["candidate_target_ids"].([]string); ok {
			candidates = append(candidates, arr...)
		} else if arr, ok := data["candidate_target_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					candidates = append(candidates, s)
				}
			}
		}
		selectedSet := map[string]bool{}
		if arr, ok := data["selected_target_ids"].([]string); ok {
			for _, s := range arr {
				selectedSet[s] = true
			}
		} else if arr, ok := data["selected_target_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					selectedSet[s] = true
				}
			}
		}
		var options []model.PromptOption
		for _, tid := range candidates {
			if selectedSet[tid] {
				continue
			}
			if p := e.State.Players[tid]; p != nil {
				options = append(options, model.PromptOption{ID: tid, Label: p.Name})
			}
		}
		targetCount := toIntContextValue(data["target_count"])
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  fmt.Sprintf("【圣光爆裂】分支②请选择第 %d/%d 名目标：", len(selectedSet)+1, targetCount),
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "hb_light_burst_mode_b_discard" {
		var remaining []int
		if arr, ok := data["remaining_indices"].([]int); ok {
			remaining = append(remaining, arr...)
		} else if arr, ok := data["remaining_indices"].([]interface{}); ok {
			for _, v := range arr {
				if f, ok := v.(float64); ok {
					remaining = append(remaining, int(f))
				}
			}
		}
		selectedCount := 0
		if arr, ok := data["selected_indices"].([]int); ok {
			selectedCount = len(arr)
		} else if arr, ok := data["selected_indices"].([]interface{}); ok {
			selectedCount = len(arr)
		}
		x := toIntContextValue(data["x_value"])
		var options []model.PromptOption
		for _, idx := range remaining {
			if idx < 0 || idx >= len(player.Hand) {
				continue
			}
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", idx),
				Label: fmt.Sprintf("%d: %s", idx+1, formatCardInfo(player.Hand[idx])),
			})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  fmt.Sprintf("【圣光爆裂】分支②请选择第 %d/%d 张弃牌：", selectedCount+1, x),
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "hb_meteor_bullet_cost" {
		var modes []string
		if arr, ok := data["cost_modes"].([]string); ok {
			modes = append(modes, arr...)
		} else if arr, ok := data["cost_modes"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					modes = append(modes, s)
				}
			}
		}
		var options []model.PromptOption
		for _, m := range modes {
			switch m {
			case "heal":
				options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", len(options)), Label: "移除1点治疗"})
			case "faith":
				options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", len(options)), Label: "移除1点信仰"})
			}
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【流星圣弹】请选择要移除的资源：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "hb_radiant_cannon_side" {
		requiredFaith := toIntContextValue(data["required_faith"])
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  fmt.Sprintf("【圣煌辉光炮】将消耗1辉光炮与%d点信仰。请选择士气对齐方向：", requiredFaith),
			Options: []model.PromptOption{
				{ID: "0", Label: "将红方士气调整为蓝方士气"},
				{ID: "1", Label: "将蓝方士气调整为红方士气"},
			},
			Min: 1,
			Max: 1,
		}
	}
	if choiceType == "hb_auto_fill_resource" {
		var modes []string
		if arr, ok := data["resource_modes"].([]string); ok {
			modes = append(modes, arr...)
		} else if arr, ok := data["resource_modes"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					modes = append(modes, s)
				}
			}
		}
		var options []model.PromptOption
		for _, mode := range modes {
			switch mode {
			case "crystal":
				options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", len(options)), Label: "分支①：消耗1蓝水晶（红宝石可替代）"})
			case "gem":
				options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", len(options)), Label: "分支②：消耗1红宝石并获得1蓝水晶"})
			}
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【自动填充】请选择要发动的分支：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "hb_auto_fill_gain" {
		branch, _ := data["branch"].(string)
		var options []model.PromptOption
		msg := "【自动填充】请选择增益："
		if branch == "gem" {
			options = append(options, model.PromptOption{ID: "0", Label: "+2信仰"})
			options = append(options, model.PromptOption{ID: "1", Label: "+2治疗"})
		} else {
			options = append(options, model.PromptOption{ID: "0", Label: "+1信仰"})
			options = append(options, model.PromptOption{ID: "1", Label: "+1治疗"})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  msg,
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	// === 灵魂术士 ===
	if choiceType == "ss_convert_color" {
		var modeOrder []string
		if arr, ok := data["mode_order"].([]string); ok {
			modeOrder = append(modeOrder, arr...)
		} else if arr, ok := data["mode_order"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					modeOrder = append(modeOrder, s)
				}
			}
		}
		var options []model.PromptOption
		for _, mode := range modeOrder {
			switch mode {
			case "y2b":
				options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", len(options)), Label: "黄魂 -> 蓝魂（转换1点）"})
			case "b2y":
				options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", len(options)), Label: "蓝魂 -> 黄魂（转换1点）"})
			}
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【灵魂转换】请选择转换方向：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "ss_link_target" {
		var allyIDs []string
		if arr, ok := data["ally_ids"].([]string); ok {
			allyIDs = append(allyIDs, arr...)
		} else if arr, ok := data["ally_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					allyIDs = append(allyIDs, s)
				}
			}
		}
		var options []model.PromptOption
		for _, aid := range allyIDs {
			if p := e.State.Players[aid]; p != nil {
				options = append(options, model.PromptOption{ID: aid, Label: p.Name})
			}
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【灵魂链接】请选择要放置灵魂链接的队友：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "ss_link_transfer_x" {
		maxX := toIntContextValue(data["max_x"])
		if maxX < 0 {
			maxX = 0
		}
		var options []model.PromptOption
		for x := 0; x <= maxX; x++ {
			label := fmt.Sprintf("移除%d点蓝魂并转移%d点伤害", x, x)
			if x == 0 {
				label = "不转移伤害"
			}
			options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", x), Label: label})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【灵魂链接】请选择要转移的伤害点数X：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "ss_recall_pick" {
		var remaining []int
		if arr, ok := data["remaining_indices"].([]int); ok {
			remaining = append(remaining, arr...)
		} else if arr, ok := data["remaining_indices"].([]interface{}); ok {
			for _, v := range arr {
				if f, ok := v.(float64); ok {
					remaining = append(remaining, int(f))
				}
			}
		}
		selectedCount := 0
		if arr, ok := data["selected_indices"].([]int); ok {
			selectedCount = len(arr)
		} else if arr, ok := data["selected_indices"].([]interface{}); ok {
			selectedCount = len(arr)
		}
		var options []model.PromptOption
		doneLabel := "完成选择并结算"
		if selectedCount == 0 {
			doneLabel = "完成选择并结算（至少选择1张）"
		}
		options = append(options, model.PromptOption{ID: "-1", Label: doneLabel})
		for _, idx := range remaining {
			if idx < 0 || idx >= len(player.Hand) {
				continue
			}
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", idx),
				Label: fmt.Sprintf("%d: %s", idx+1, formatCardInfo(player.Hand[idx])),
			})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  fmt.Sprintf("【灵魂召还】已选择%d张法术牌。继续选择或结束：", selectedCount),
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	// === 血之巫女 ===
	if choiceType == "bp_shared_life_target" {
		var targetIDs []string
		if arr, ok := data["target_ids"].([]string); ok {
			targetIDs = append(targetIDs, arr...)
		} else if arr, ok := data["target_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					targetIDs = append(targetIDs, s)
				}
			}
		}
		var options []model.PromptOption
		for _, tid := range targetIDs {
			if p := e.State.Players[tid]; p != nil {
				options = append(options, model.PromptOption{ID: tid, Label: p.Name})
			}
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【同生共死】请选择放置目标（先摸2张牌，再放置）：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "bp_blood_sorrow_mode" {
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【血之哀伤】请选择效果：",
			Options: []model.PromptOption{
				{ID: "0", Label: "转移同生共死目标"},
				{ID: "1", Label: "移除同生共死"},
			},
			Min: 1,
			Max: 1,
		}
	}
	if choiceType == "bp_blood_sorrow_target" {
		var targetIDs []string
		if arr, ok := data["target_ids"].([]string); ok {
			targetIDs = append(targetIDs, arr...)
		} else if arr, ok := data["target_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					targetIDs = append(targetIDs, s)
				}
			}
		}
		var options []model.PromptOption
		for _, tid := range targetIDs {
			if p := e.State.Players[tid]; p != nil {
				options = append(options, model.PromptOption{ID: tid, Label: p.Name})
			}
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【血之哀伤】请选择新的同生共死目标：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "bp_blood_wail_x" {
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【血之悲鸣】请选择X值：",
			Options: []model.PromptOption{
				{ID: "0", Label: "X=0（伤害=1）"},
				{ID: "1", Label: "X=1（伤害=2）"},
				{ID: "2", Label: "X=2（伤害=3）"},
			},
			Min: 1,
			Max: 1,
		}
	}
	if choiceType == "bp_curse_discard" {
		discardCount := toIntContextValue(data["discard_count"])
		if discardCount < 0 {
			discardCount = 0
		}
		var remaining []int
		if arr, ok := data["remaining_indices"].([]int); ok {
			remaining = append(remaining, arr...)
		} else if arr, ok := data["remaining_indices"].([]interface{}); ok {
			for _, v := range arr {
				if f, ok := v.(float64); ok {
					remaining = append(remaining, int(f))
				}
			}
		}
		selectedCount := 0
		if arr, ok := data["selected_indices"].([]int); ok {
			selectedCount = len(arr)
		} else if arr, ok := data["selected_indices"].([]interface{}); ok {
			selectedCount = len(arr)
		}
		var options []model.PromptOption
		doneLabel := fmt.Sprintf("完成弃牌（需弃%d张）", discardCount)
		if selectedCount < discardCount {
			doneLabel = fmt.Sprintf("完成弃牌（还需%d张）", discardCount-selectedCount)
		}
		options = append(options, model.PromptOption{ID: "-1", Label: doneLabel})
		for _, idx := range remaining {
			if idx < 0 || idx >= len(player.Hand) {
				continue
			}
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", idx),
				Label: fmt.Sprintf("%d: %s", idx+1, formatCardInfo(player.Hand[idx])),
			})
		}
		return &model.Prompt{
			Type:     model.PromptChooseCards,
			PlayerID: playerID,
			Message:  fmt.Sprintf("【血之诅咒】已选择%d/%d张弃牌，继续选择或完成：", selectedCount, discardCount),
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	// === 蝶舞者 ===
	if choiceType == "bt_dance_mode" {
		canDiscard := toBoolContextValue(data["can_discard"])
		options := []model.PromptOption{
			{ID: "0", Label: "摸1张牌"},
		}
		if canDiscard {
			options = append(options, model.PromptOption{ID: "1", Label: "弃1张牌"})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【舞动】请选择先执行的动作：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "bt_dance_discard" {
		var options []model.PromptOption
		for idx, c := range player.Hand {
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", idx),
				Label: fmt.Sprintf("%d: %s", idx+1, formatCardInfo(c)),
			})
		}
		return &model.Prompt{
			Type:     model.PromptChooseCards,
			PlayerID: playerID,
			Message:  "【舞动】请选择要弃置的1张手牌：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "bt_chrysalis_resolve" {
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【蛹化】确认结算：+1蛹，并将牌库顶4张牌作为茧放置？",
			Options: []model.PromptOption{
				{ID: "0", Label: "确认结算"},
			},
			Min: 1,
			Max: 1,
		}
	}
	if choiceType == "bt_cocoon_overflow_discard" {
		discardCount := toIntContextValue(data["discard_count"])
		var remaining []int
		if arr, ok := data["remaining_indices"].([]int); ok {
			remaining = append(remaining, arr...)
		} else if arr, ok := data["remaining_indices"].([]interface{}); ok {
			for _, v := range arr {
				if f, ok := v.(float64); ok {
					remaining = append(remaining, int(f))
				}
			}
		}
		selectedCount := 0
		if arr, ok := data["selected_indices"].([]int); ok {
			selectedCount = len(arr)
		} else if arr, ok := data["selected_indices"].([]interface{}); ok {
			selectedCount = len(arr)
		}
		var options []model.PromptOption
		doneLabel := fmt.Sprintf("完成舍弃（需弃%d个茧）", discardCount)
		if selectedCount < discardCount {
			doneLabel = fmt.Sprintf("完成舍弃（还需%d个茧）", discardCount-selectedCount)
		}
		options = append(options, model.PromptOption{ID: "-1", Label: doneLabel})
		for _, idx := range remaining {
			if idx < 0 || idx >= len(player.Field) || player.Field[idx] == nil {
				continue
			}
			fc := player.Field[idx]
			if fc.Mode != model.FieldCover || fc.Effect != model.EffectButterflyCocoon {
				continue
			}
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", idx),
				Label: fmt.Sprintf("茧[%d]: %s", idx, formatCardInfo(fc.Card)),
			})
		}
		return &model.Prompt{
			Type:     model.PromptChooseCards,
			PlayerID: playerID,
			Message:  fmt.Sprintf("【茧上限】已选择%d/%d个茧，继续选择或完成：", selectedCount, discardCount),
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "bt_reverse_discard" {
		discardCount := toIntContextValue(data["discard_count"])
		var remaining []int
		if arr, ok := data["remaining_indices"].([]int); ok {
			remaining = append(remaining, arr...)
		} else if arr, ok := data["remaining_indices"].([]interface{}); ok {
			for _, v := range arr {
				if f, ok := v.(float64); ok {
					remaining = append(remaining, int(f))
				}
			}
		}
		selectedCount := 0
		if arr, ok := data["selected_indices"].([]int); ok {
			selectedCount = len(arr)
		} else if arr, ok := data["selected_indices"].([]interface{}); ok {
			selectedCount = len(arr)
		}
		var options []model.PromptOption
		doneLabel := fmt.Sprintf("完成弃牌（需弃%d张）", discardCount)
		if selectedCount < discardCount {
			doneLabel = fmt.Sprintf("完成弃牌（还需%d张）", discardCount-selectedCount)
		}
		options = append(options, model.PromptOption{ID: "-1", Label: doneLabel})
		for _, idx := range remaining {
			if idx < 0 || idx >= len(player.Hand) {
				continue
			}
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", idx),
				Label: fmt.Sprintf("%d: %s", idx+1, formatCardInfo(player.Hand[idx])),
			})
		}
		return &model.Prompt{
			Type:     model.PromptChooseCards,
			PlayerID: playerID,
			Message:  fmt.Sprintf("【倒逆之蝶】已选择%d/%d张弃牌，继续选择或完成：", selectedCount, discardCount),
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "bt_reverse_mode" {
		canBranch2 := toBoolContextValue(data["can_branch2"])
		options := []model.PromptOption{
			{ID: "0", Label: "分支①：对目标造成1点不可治疗抵御的法术伤害"},
		}
		if canBranch2 {
			options = append(options, model.PromptOption{ID: "1", Label: "分支②：移除2个茧或自伤4，然后移除1个蛹"})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【倒逆之蝶】请选择发动分支：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "bt_reverse_target" {
		var targetIDs []string
		if arr, ok := data["target_ids"].([]string); ok {
			targetIDs = append(targetIDs, arr...)
		} else if arr, ok := data["target_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					targetIDs = append(targetIDs, s)
				}
			}
		}
		var options []model.PromptOption
		for _, tid := range targetIDs {
			if p := e.State.Players[tid]; p != nil {
				options = append(options, model.PromptOption{ID: tid, Label: p.Name})
			}
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【倒逆之蝶】请选择分支①伤害目标：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "bt_reverse_branch2_cost" {
		canRemove := toBoolContextValue(data["can_remove_cocoon"])
		options := []model.PromptOption{}
		if canRemove {
			options = append(options, model.PromptOption{ID: "0", Label: "移除2个茧"})
		}
		options = append(options, model.PromptOption{ID: "1", Label: "对自己造成4点法术伤害"})
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【倒逆之蝶】请选择分支②代价：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "bt_reverse_branch2_pick" {
		var remaining []int
		if arr, ok := data["remaining_indices"].([]int); ok {
			remaining = append(remaining, arr...)
		} else if arr, ok := data["remaining_indices"].([]interface{}); ok {
			for _, v := range arr {
				if f, ok := v.(float64); ok {
					remaining = append(remaining, int(f))
				}
			}
		}
		selectedCount := 0
		if arr, ok := data["selected_indices"].([]int); ok {
			selectedCount = len(arr)
		} else if arr, ok := data["selected_indices"].([]interface{}); ok {
			selectedCount = len(arr)
		}
		var options []model.PromptOption
		doneLabel := "完成选择（需2个茧）"
		if selectedCount < 2 {
			doneLabel = fmt.Sprintf("完成选择（还需%d个茧）", 2-selectedCount)
		}
		options = append(options, model.PromptOption{ID: "-1", Label: doneLabel})
		for _, idx := range remaining {
			if idx < 0 || idx >= len(player.Field) || player.Field[idx] == nil {
				continue
			}
			fc := player.Field[idx]
			if fc.Mode != model.FieldCover || fc.Effect != model.EffectButterflyCocoon {
				continue
			}
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", idx),
				Label: fmt.Sprintf("茧[%d]: %s", idx, formatCardInfo(fc.Card)),
			})
		}
		return &model.Prompt{
			Type:     model.PromptChooseCards,
			PlayerID: playerID,
			Message:  fmt.Sprintf("【倒逆之蝶】分支②已选择%d/2个茧，继续选择或完成：", selectedCount),
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "bt_pilgrimage_pick" || choiceType == "bt_poison_pick" {
		var cocoonIndices []int
		if arr, ok := data["cocoon_indices"].([]int); ok {
			cocoonIndices = append(cocoonIndices, arr...)
		} else if arr, ok := data["cocoon_indices"].([]interface{}); ok {
			for _, v := range arr {
				if f, ok := v.(float64); ok {
					cocoonIndices = append(cocoonIndices, int(f))
				}
			}
		}
		var options []model.PromptOption
		options = append(options, model.PromptOption{ID: "-1", Label: "不发动"})
		for _, idx := range cocoonIndices {
			if idx < 0 || idx >= len(player.Field) || player.Field[idx] == nil {
				continue
			}
			fc := player.Field[idx]
			if fc.Mode != model.FieldCover || fc.Effect != model.EffectButterflyCocoon {
				continue
			}
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", len(options)),
				Label: fmt.Sprintf("移除茧[%d]: %s", idx, formatCardInfo(fc.Card)),
			})
		}
		msg := "【朝圣】是否移除1个茧抵御1点伤害？"
		if choiceType == "bt_poison_pick" {
			msg = "【毒粉】是否移除1个茧使该次法术伤害+1？"
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  msg,
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "bt_mirror_pair" {
		var labels []string
		if arr, ok := data["pair_labels"].([]string); ok {
			labels = append(labels, arr...)
		} else if arr, ok := data["pair_labels"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					labels = append(labels, s)
				}
			}
		}
		options := []model.PromptOption{{ID: "-1", Label: "不发动"}}
		for i, label := range labels {
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", i+1),
				Label: fmt.Sprintf("移除并展示：%s", label),
			})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【镜花水月】是否发动并改写该次2点法术伤害？",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "bt_wither_confirm" {
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【凋零】可发动：是否对目标造成1点法术伤害并对自己造成2点法术伤害？",
			Options: []model.PromptOption{
				{ID: "0", Label: "发动凋零"},
				{ID: "1", Label: "不发动"},
			},
			Min: 1,
			Max: 1,
		}
	}
	if choiceType == "bt_wither_target" {
		var targetIDs []string
		if arr, ok := data["target_ids"].([]string); ok {
			targetIDs = append(targetIDs, arr...)
		} else if arr, ok := data["target_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					targetIDs = append(targetIDs, s)
				}
			}
		}
		var options []model.PromptOption
		for _, tid := range targetIDs {
			if p := e.State.Players[tid]; p != nil {
				options = append(options, model.PromptOption{ID: tid, Label: p.Name})
			}
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【凋零】请选择1名目标角色：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	// === 月之女神 ===
	if choiceType == "mg_medusa_darkmoon_pick" {
		var indices []int
		if arr, ok := data["darkmoon_indices"].([]int); ok {
			indices = append(indices, arr...)
		} else if arr, ok := data["darkmoon_indices"].([]interface{}); ok {
			for _, v := range arr {
				if f, ok := v.(float64); ok {
					indices = append(indices, int(f))
				}
			}
		}
		var options []model.PromptOption
		for _, idx := range indices {
			if idx < 0 || idx >= len(player.Field) || player.Field[idx] == nil {
				continue
			}
			fc := player.Field[idx]
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", idx),
				Label: fmt.Sprintf("移除暗月[%s/%s/%s]", fc.Card.Name, fc.Card.Type, fc.Card.Element),
			})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【美杜莎之眼】请选择要展示并移除的同系暗月：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "mg_medusa_magic_discard" {
		var options []model.PromptOption
		for idx, c := range player.Hand {
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", idx),
				Label: fmt.Sprintf("%d: %s", idx+1, formatCardInfo(c)),
			})
		}
		return &model.Prompt{
			Type:     model.PromptChooseCards,
			PlayerID: playerID,
			Message:  "【美杜莎之眼】因移除了法术暗月，请弃1张手牌：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "mg_medusa_magic_target" {
		var targetIDs []string
		if arr, ok := data["target_ids"].([]string); ok {
			targetIDs = append(targetIDs, arr...)
		} else if arr, ok := data["target_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					targetIDs = append(targetIDs, s)
				}
			}
		}
		var options []model.PromptOption
		for _, tid := range targetIDs {
			if p := e.State.Players[tid]; p != nil {
				options = append(options, model.PromptOption{ID: tid, Label: p.Name})
			}
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【美杜莎之眼】请选择1名对手造成1点法术伤害：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "mg_moon_cycle_mode" {
		var modes []string
		if arr, ok := data["modes"].([]string); ok {
			modes = append(modes, arr...)
		} else if arr, ok := data["modes"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					modes = append(modes, s)
				}
			}
		}
		var options []model.PromptOption
		for _, mode := range modes {
			switch mode {
			case "branch1":
				options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", len(options)), Label: "分支①：移除1个暗月，令任意角色+1治疗"})
			case "branch2":
				options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", len(options)), Label: "分支②：移除1点治疗，你+1新月"})
			}
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【月之轮回】请选择发动分支：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "mg_moon_cycle_heal_target" {
		var targetIDs []string
		if arr, ok := data["target_ids"].([]string); ok {
			targetIDs = append(targetIDs, arr...)
		} else if arr, ok := data["target_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					targetIDs = append(targetIDs, s)
				}
			}
		}
		var options []model.PromptOption
		for _, tid := range targetIDs {
			if p := e.State.Players[tid]; p != nil {
				options = append(options, model.PromptOption{ID: tid, Label: p.Name})
			}
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【月之轮回】请选择获得1点治疗的角色：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "mg_blasphemy_target" {
		var targetIDs []string
		if arr, ok := data["target_ids"].([]string); ok {
			targetIDs = append(targetIDs, arr...)
		} else if arr, ok := data["target_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					targetIDs = append(targetIDs, s)
				}
			}
		}
		options := []model.PromptOption{{ID: "0", Label: "跳过月渎"}}
		for _, tid := range targetIDs {
			if p := e.State.Players[tid]; p != nil {
				options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", len(options)), Label: fmt.Sprintf("对 %s 造成1点法术伤害", p.Name)})
			}
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【月渎】请选择目标（或跳过）：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "mg_darkmoon_slash_x" {
		maxX := toIntContextValue(data["max_x"])
		if maxX < 0 {
			maxX = 0
		}
		var options []model.PromptOption
		for x := 0; x <= maxX; x++ {
			label := fmt.Sprintf("移除%d个暗月，本次攻击伤害额外+%d", x, x)
			if x == 0 {
				label = "不移除暗月（伤害不增加）"
			}
			options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", x), Label: label})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【暗月斩】请选择X值：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "mg_pale_moon_mode" {
		var modes []string
		if arr, ok := data["modes"].([]string); ok {
			modes = append(modes, arr...)
		} else if arr, ok := data["modes"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					modes = append(modes, s)
				}
			}
		}
		var options []model.PromptOption
		for _, mode := range modes {
			switch mode {
			case "branch1":
				options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", len(options)), Label: "分支①：移除3石化，强化下次攻击并获得额外回合"})
			case "branch2":
				options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", len(options)), Label: "分支②：移除X新月，弃1张牌并造成(X+1)法术伤害"})
			}
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【苍白之月】请选择分支：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "mg_pale_moon_x" {
		maxX := toIntContextValue(data["max_x"])
		if maxX < 0 {
			maxX = 0
		}
		var options []model.PromptOption
		for x := 0; x <= maxX; x++ {
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", x),
				Label: fmt.Sprintf("X=%d（目标法术伤害=%d）", x, x+1),
			})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【苍白之月】分支②请选择X值：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "mg_pale_moon_target" {
		var targetIDs []string
		if arr, ok := data["target_ids"].([]string); ok {
			targetIDs = append(targetIDs, arr...)
		} else if arr, ok := data["target_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					targetIDs = append(targetIDs, s)
				}
			}
		}
		var options []model.PromptOption
		for _, tid := range targetIDs {
			if p := e.State.Players[tid]; p != nil {
				options = append(options, model.PromptOption{ID: tid, Label: p.Name})
			}
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【苍白之月】分支②请选择目标对手：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "mg_pale_moon_discard" {
		var options []model.PromptOption
		for idx, c := range player.Hand {
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", idx),
				Label: fmt.Sprintf("%d: %s", idx+1, formatCardInfo(c)),
			})
		}
		return &model.Prompt{
			Type:     model.PromptChooseCards,
			PlayerID: playerID,
			Message:  "【苍白之月】分支②请弃1张牌：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	// === 魔枪 ===
	if choiceType == "ml_black_spear_x" {
		maxX := toIntContextValue(data["max_x"])
		if maxX < 1 {
			maxX = 1
		}
		var options []model.PromptOption
		for x := 1; x <= maxX; x++ {
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", x),
				Label: fmt.Sprintf("X=%d（消耗%d蓝水晶，伤害额外+%d）", x, x, x+2),
			})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【漆黑之枪】请选择X值：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "ml_dark_barrier_mode" {
		maxMagic := toIntContextValue(data["max_magic"])
		maxThunder := toIntContextValue(data["max_thunder"])
		var options []model.PromptOption
		if maxMagic > 0 {
			options = append(options, model.PromptOption{ID: "0", Label: "弃法术牌"})
		}
		if maxThunder > 0 {
			options = append(options, model.PromptOption{ID: "1", Label: "弃雷系牌"})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【暗之障壁】请选择本次弃牌类型：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "ml_dark_barrier_x" {
		maxX := toIntContextValue(data["max_x"])
		if maxX < 1 {
			maxX = 1
		}
		var options []model.PromptOption
		for x := 1; x <= maxX; x++ {
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", x),
				Label: fmt.Sprintf("弃置%d张", x),
			})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【暗之障壁】请选择X值：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "ml_dark_barrier_cards" {
		var remaining []int
		if arr, ok := data["remaining_indices"].([]int); ok {
			remaining = append(remaining, arr...)
		} else if arr, ok := data["remaining_indices"].([]interface{}); ok {
			for _, v := range arr {
				if f, ok := v.(float64); ok {
					remaining = append(remaining, int(f))
				}
			}
		}
		selectedCount := 0
		if arr, ok := data["selected_indices"].([]int); ok {
			selectedCount = len(arr)
		} else if arr, ok := data["selected_indices"].([]interface{}); ok {
			selectedCount = len(arr)
		}
		xValue := toIntContextValue(data["x_value"])
		var options []model.PromptOption
		for _, idx := range remaining {
			if idx < 0 || idx >= len(player.Hand) {
				continue
			}
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", idx),
				Label: fmt.Sprintf("%d: %s", idx+1, formatCardInfo(player.Hand[idx])),
			})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  fmt.Sprintf("【暗之障壁】请选择第 %d/%d 张弃牌：", selectedCount+1, xValue),
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "ml_fullness_cost_card" {
		var options []model.PromptOption
		for idx, c := range player.Hand {
			if c.Type != model.CardTypeMagic && c.Element != model.ElementThunder {
				continue
			}
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", idx),
				Label: fmt.Sprintf("%d: %s", idx+1, formatCardInfo(c)),
			})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【充盈】请选择要弃置的1张法术牌或雷系牌：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "ml_fullness_discard_step" {
		currentID, _ := data["current_player_id"].(string)
		target := e.State.Players[currentID]
		if target == nil {
			return nil
		}
		allowSkip, _ := data["allow_skip"].(bool)
		var options []model.PromptOption
		if allowSkip {
			options = append(options, model.PromptOption{ID: "skip", Label: "不弃置"})
		}
		var candidates []int
		if arr, ok := data["candidates"].([]int); ok {
			candidates = append(candidates, arr...)
		} else if arr, ok := data["candidates"].([]interface{}); ok {
			for _, v := range arr {
				if f, ok := v.(float64); ok {
					candidates = append(candidates, int(f))
				}
			}
		}
		for _, idx := range candidates {
			if idx < 0 || idx >= len(target.Hand) {
				continue
			}
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", idx),
				Label: fmt.Sprintf("%s 的第%d张：%s", target.Name, idx+1, formatCardInfo(target.Hand[idx])),
			})
		}
		msg := fmt.Sprintf("【充盈】请选择 %s 的弃牌：", target.Name)
		if allowSkip {
			msg = fmt.Sprintf("【充盈】请选择 %s 是否弃牌：", target.Name)
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  msg,
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	// === 灵符师 ===
	if choiceType == "sc_incant_confirm" {
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【念咒】是否将1张手牌面朝下放置为妖力？",
			Options: []model.PromptOption{
				{ID: "0", Label: "是"},
				{ID: "1", Label: "否"},
			},
			Min: 1,
			Max: 1,
		}
	}
	if choiceType == "sc_incant_card" {
		var options []model.PromptOption
		for idx, c := range player.Hand {
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", idx),
				Label: fmt.Sprintf("%d: %s", idx+1, formatCardInfo(c)),
			})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【念咒】请选择要作为妖力盖放的手牌：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "sc_hundred_night_power" {
		powers := spiritCasterPowerCovers(player)
		var options []model.PromptOption
		for i, fc := range powers {
			if fc == nil {
				continue
			}
			eleZh := elementNameForPrompt(string(fc.Card.Element))
			if eleZh == "" {
				eleZh = string(fc.Card.Element)
			}
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", i),
				Label: fmt.Sprintf("%s（%s系）", fc.Card.Name, eleZh),
			})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【百鬼夜行】请选择要移除的1个妖力：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "sc_hundred_night_fire_reveal" {
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【百鬼夜行】移除的是火系妖力，是否展示并改为范围伤害？",
			Options: []model.PromptOption{
				{ID: "0", Label: "展示并改为范围伤害"},
				{ID: "1", Label: "不展示，改为单体伤害"},
			},
			Min: 1,
			Max: 1,
		}
	}
	if choiceType == "sc_hundred_night_exclude_pick" {
		var targetIDs []string
		if arr, ok := data["target_ids"].([]string); ok {
			targetIDs = append(targetIDs, arr...)
		} else if arr, ok := data["target_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					targetIDs = append(targetIDs, s)
				}
			}
		}
		selectedSet := map[string]bool{}
		if arr, ok := data["selected_exclude_ids"].([]string); ok {
			for _, s := range arr {
				selectedSet[s] = true
			}
		} else if arr, ok := data["selected_exclude_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					selectedSet[s] = true
				}
			}
		}
		var options []model.PromptOption
		for _, tid := range targetIDs {
			if selectedSet[tid] {
				continue
			}
			if p := e.State.Players[tid]; p != nil {
				options = append(options, model.PromptOption{ID: tid, Label: p.Name})
			}
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  fmt.Sprintf("【百鬼夜行】请选择第 %d/2 名排除目标：", len(selectedSet)+1),
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "sc_spiritual_collapse_confirm" {
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【灵力崩解】是否消耗1蓝水晶（红宝石可替代），使本次每段伤害额外+1？",
			Options: []model.PromptOption{
				{ID: "0", Label: "是"},
				{ID: "1", Label: "否"},
			},
			Min: 1,
			Max: 1,
		}
	}
	if choiceType == "sc_talisman_wind_discard" {
		currentTargetID, _ := data["current_target_id"].(string)
		target := e.State.Players[currentTargetID]
		if target == nil {
			return nil
		}
		var options []model.PromptOption
		for idx, c := range target.Hand {
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", idx),
				Label: fmt.Sprintf("%d: %s", idx+1, formatCardInfo(c)),
			})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  fmt.Sprintf("【灵符-风行】请 %s 选择1张手牌弃置：", target.Name),
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	// === 吟游诗人 ===
	if choiceType == "bd_descent_confirm" {
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【沉沦协奏曲】是否发动？（发动后需弃2张同系牌）",
			Options: []model.PromptOption{
				{ID: "0", Label: "发动"},
				{ID: "1", Label: "跳过"},
			},
			Min: 1,
			Max: 1,
		}
	}
	if choiceType == "bd_descent_element" {
		elemCounts := getSameElementCounts(player)
		var elems []model.Element
		for _, ele := range elementOrderForPrompt() {
			if elemCounts[ele] >= 2 {
				elems = append(elems, ele)
			}
		}
		var options []model.PromptOption
		for _, ele := range elems {
			options = append(options, model.PromptOption{
				ID:    string(ele),
				Label: fmt.Sprintf("%s系", elementNameForPrompt(string(ele))),
			})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【沉沦协奏曲】请选择要弃置的同系元素：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "bd_descent_cards" {
		chosenEle, _ := data["chosen_element"].(string)
		chosenEleZh := elementNameForPrompt(chosenEle)
		var remaining []int
		if arr, ok := data["remaining_indices"].([]int); ok {
			remaining = append(remaining, arr...)
		} else if arr, ok := data["remaining_indices"].([]interface{}); ok {
			for _, v := range arr {
				if f, ok := v.(float64); ok {
					remaining = append(remaining, int(f))
				}
			}
		}
		selected := 0
		if arr, ok := data["selected_indices"].([]int); ok {
			selected = len(arr)
		} else if arr, ok := data["selected_indices"].([]interface{}); ok {
			selected = len(arr)
		}
		var options []model.PromptOption
		for _, idx := range remaining {
			if idx < 0 || idx >= len(player.Hand) {
				continue
			}
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", idx),
				Label: fmt.Sprintf("%d: %s", idx+1, formatCardInfo(player.Hand[idx])),
			})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  fmt.Sprintf("【沉沦协奏曲】请选择第 %d/2 张%s系牌：", selected+1, chosenEleZh),
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "bd_dissonance_x" {
		maxX := toIntContextValue(data["max_x"])
		if maxX < 2 {
			maxX = 2
		}
		var options []model.PromptOption
		for x := 2; x <= maxX; x++ {
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", x),
				Label: fmt.Sprintf("X=%d", x),
			})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【不谐和弦】请选择X值：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "bd_dissonance_mode" {
		xValue := toIntContextValue(data["x_value"])
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  fmt.Sprintf("【不谐和弦】请选择分支（X=%d）：", xValue),
			Options: []model.PromptOption{
				{ID: "0", Label: fmt.Sprintf("你与目标各摸%d张牌", xValue-1)},
				{ID: "1", Label: fmt.Sprintf("你与目标各弃%d张牌", xValue-1)},
			},
			Min: 1,
			Max: 1,
		}
	}
	if choiceType == "bd_dissonance_discard_step" {
		currentActorID, _ := data["current_actor_id"].(string)
		actor := e.State.Players[currentActorID]
		if actor == nil {
			return nil
		}
		need := toIntContextValue(data["need_count"])
		selected := toIntContextValue(data["selected_count"])
		var options []model.PromptOption
		for idx, c := range actor.Hand {
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", idx),
				Label: fmt.Sprintf("%d: %s", idx+1, formatCardInfo(c)),
			})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  fmt.Sprintf("【不谐和弦】请 %s 选择第 %d/%d 张弃牌：", actor.Name, selected+1, need),
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "bd_rousing_mode" {
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【激昂狂想曲】请选择效果：",
			Options: []model.PromptOption{
				{ID: "0", Label: "对2名对手各造成1点法术伤害"},
				{ID: "1", Label: "弃2张牌"},
			},
			Min: 1,
			Max: 1,
		}
	}
	if choiceType == "bd_rousing_targets" {
		targetIDs := parseStringSliceContextValue(data["target_ids"])
		selectedSet := idsToSet(parseStringSliceContextValue(data["selected_target_ids"]))
		var options []model.PromptOption
		for _, tid := range targetIDs {
			if selectedSet[tid] {
				continue
			}
			if p := e.State.Players[tid]; p != nil {
				options = append(options, model.PromptOption{
					ID:    tid,
					Label: p.Name,
				})
			}
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  fmt.Sprintf("【激昂狂想曲】请选择第 %d/2 名目标：", len(selectedSet)+1),
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "bd_rousing_discard_cards" {
		selected := 0
		if arr, ok := data["selected_indices"].([]int); ok {
			selected = len(arr)
		} else if arr, ok := data["selected_indices"].([]interface{}); ok {
			selected = len(arr)
		}
		var remaining []int
		if arr, ok := data["remaining_indices"].([]int); ok {
			remaining = append(remaining, arr...)
		} else if arr, ok := data["remaining_indices"].([]interface{}); ok {
			for _, v := range arr {
				if f, ok := v.(float64); ok {
					remaining = append(remaining, int(f))
				}
			}
		}
		var options []model.PromptOption
		for _, idx := range remaining {
			if idx < 0 || idx >= len(player.Hand) {
				continue
			}
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", idx),
				Label: fmt.Sprintf("%d: %s", idx+1, formatCardInfo(player.Hand[idx])),
			})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  fmt.Sprintf("【激昂狂想曲】请选择第 %d/2 张弃牌：", selected+1),
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "bd_victory_mode" {
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【胜利交响诗】请选择效果：",
			Options: []model.PromptOption{
				{ID: "0", Label: "将我方战绩区1个星石提炼为你的能量"},
				{ID: "1", Label: "我方战绩区+1宝石，你+1治疗"},
			},
			Min: 1,
			Max: 1,
		}
	}
	if choiceType == "bd_hope_draw_confirm" {
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【希望赋格曲】是否先摸1张牌？",
			Options: []model.PromptOption{
				{ID: "0", Label: "是"},
				{ID: "1", Label: "否"},
			},
			Min: 1,
			Max: 1,
		}
	}
	if choiceType == "bd_hope_mode" {
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【希望赋格曲】请选择分支：",
			Options: []model.PromptOption{
				{ID: "0", Label: "将永恒乐章放置于目标队友面前"},
				{ID: "1", Label: "将永恒乐章转移给我方另一名角色"},
			},
			Min: 1,
			Max: 1,
		}
	}
	if choiceType == "bd_hope_transfer_discard" {
		var options []model.PromptOption
		for idx, c := range player.Hand {
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", idx),
				Label: fmt.Sprintf("%d: %s", idx+1, formatCardInfo(c)),
			})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【希望赋格曲】请选择弃置1张手牌：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	if choiceType == "bd_hope_transfer_gain" {
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【希望赋格曲】请选择获得效果：",
			Options: []model.PromptOption{
				{ID: "0", Label: "+1治疗"},
				{ID: "1", Label: "+1灵感"},
			},
			Min: 1,
			Max: 1,
		}
	}
	// === 通用目标选择类 ===
	if choiceType == "elf_elemental_shot_water_target" ||
		choiceType == "elf_elemental_shot_earth_target" ||
		choiceType == "elf_pet_empower_target" ||
		choiceType == "elf_ritual_release_target" ||
		choiceType == "bw_substitute_doll_target" ||
		choiceType == "bw_mana_inversion_target" ||
		choiceType == "sage_magic_rebound_target" ||
		choiceType == "sage_arcane_target" ||
		choiceType == "priest_divine_domain_damage_target" ||
		choiceType == "priest_divine_domain_heal_target" ||
		choiceType == "onmyoji_dark_ritual_target" ||
		choiceType == "onmyoji_life_barrier_support_target" ||
		choiceType == "onmyoji_life_barrier_release_target" ||
		choiceType == "plague_death_touch_target" ||
		choiceType == "ms_shadow_meteor_target" ||
		choiceType == "css_blood_barrier_target" ||
		choiceType == "hom_dual_echo_target" ||
		choiceType == "mb_thunder_scatter_target" ||
		choiceType == "mb_multi_shot_target" ||
		choiceType == "mb_demon_eye_target" ||
		choiceType == "sc_hundred_night_target" ||
		choiceType == "bd_descent_target" ||
		choiceType == "bd_dissonance_target" ||
		choiceType == "bd_hope_place_target" ||
		choiceType == "bd_hope_transfer_target" ||
		choiceType == "ml_stardust_target" ||
		choiceType == "fighter_psi_bullet_target" {
		var targetIDs []string
		if arr, ok := data["target_ids"].([]string); ok {
			targetIDs = arr
		} else if arr, ok := data["target_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					targetIDs = append(targetIDs, s)
				}
			}
		}
		var options []model.PromptOption
		for _, tid := range targetIDs {
			if p := e.State.Players[tid]; p != nil {
				options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", len(options)), Label: p.Name})
			}
		}
		msg := "请选择目标："
		switch choiceType {
		case "elf_elemental_shot_water_target":
			msg = "【水之矢】请选择+1治疗目标："
		case "elf_elemental_shot_earth_target":
			msg = "【地之矢】请选择1点法术伤害目标："
		case "elf_pet_empower_target":
			msg = "【宠物强化】请选择摸1弃1目标："
		case "elf_ritual_release_target":
			msg = "【精灵密仪】你已无祝福，转正并请选择2点法术伤害目标："
		case "bw_substitute_doll_target":
			msg = "【替身玩偶】请选择摸1张牌的队友："
		case "bw_mana_inversion_target":
			msg = "【魔能反转】请选择法术伤害目标："
		case "sage_magic_rebound_target":
			msg = "【法术反弹】请选择法术伤害目标："
		case "sage_arcane_target":
			msg = "【魔道法典】请选择法术伤害目标："
		case "priest_divine_domain_damage_target":
			msg = "【神圣领域·分支①】请选择2点法术伤害目标："
		case "priest_divine_domain_heal_target":
			msg = "【神圣领域·分支②】请选择+1治疗的队友："
		case "onmyoji_dark_ritual_target":
			msg = "【黑暗祭礼】请选择2点法术伤害目标："
		case "onmyoji_life_barrier_support_target":
			msg = "【生命结界·分支①】请选择获得+1宝石/+1治疗的队友："
		case "onmyoji_life_barrier_release_target":
			msg = "【生命结界·分支②】请选择弃1张手牌的队友："
		case "plague_death_touch_target":
			msg = "【死亡之触】请选择法术伤害目标："
		case "ms_shadow_meteor_target":
			msg = "【暗影流星】请选择2点法术伤害目标："
		case "css_blood_barrier_target":
			msg = "【血气屏障】请选择1点法术伤害目标："
		case "hom_dual_echo_target":
			damage := 0
			if v, ok := data["damage"].(int); ok {
				damage = v
			} else if f, ok := data["damage"].(float64); ok {
				damage = int(f)
			}
			msg = fmt.Sprintf("【双重回响】请选择额外造成%d点法术伤害的目标：", damage)
		case "mb_thunder_scatter_target":
			extraX := toIntContextValue(data["extra_x"])
			msg = fmt.Sprintf("【雷光散射】请选择额外受到%d点法术伤害的目标：", extraX)
		case "mb_multi_shot_target":
			msg = "【多重射击】请选择暗系追加攻击目标："
		case "mb_demon_eye_target":
			msg = "【魔眼】请选择弃1张牌的目标角色："
		case "sc_hundred_night_target":
			msg = "【百鬼夜行】请选择1点法术伤害目标："
		case "bd_descent_target":
			msg = "【沉沦协奏曲】请选择1点法术伤害目标："
		case "bd_dissonance_target":
			msg = "【不谐和弦】请选择目标角色："
		case "bd_hope_place_target":
			msg = "【希望赋格曲】请选择放置永恒乐章的目标队友："
		case "bd_hope_transfer_target":
			msg = "【希望赋格曲】请选择转移永恒乐章的目标角色："
		case "ml_stardust_target":
			msg = "【幻影星尘】请选择2点法术伤害目标："
		case "fighter_psi_bullet_target":
			msg = "【念弹】请选择1名目标对手："
		}
		if choiceType == "hom_dual_echo_target" || choiceType == "css_blood_barrier_target" {
			options = append(options, model.PromptOption{ID: "cancel", Label: "取消"})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  msg,
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}

	return nil
}

// SkipResponse 跳过响应阶段
func (e *GameEngine) SkipResponse() error {
	// 获取当前中断的上下文（用于恢复扣卡）
	// InterruptResponseSkill 的 Context 直接是 *model.Context（如 TriggerBeforeDraw 的 drawCtx）
	var resumeCtx *model.Context
	var resumeAttackHitCtx *model.Context
	var resumeAttackMissCtx *model.Context
	var resumeMoraleCtx *model.Context
	var resumePhaseEndCtx *model.Context
	holyLancerEarthSkipped := false
	holyLancerEarthPlayerID := ""
	if e.State.PendingInterrupt != nil {
		if e.State.PendingInterrupt.Type == model.InterruptResponseSkill {
			if ctx, ok := e.State.PendingInterrupt.Context.(*model.Context); ok && ctx != nil && ctx.Trigger == model.TriggerOnAttackHit {
				for _, sid := range e.State.PendingInterrupt.SkillIDs {
					if sid == "holy_lancer_earth_spear" {
						holyLancerEarthSkipped = true
						holyLancerEarthPlayerID = e.State.PendingInterrupt.PlayerID
						break
					}
				}
			}
		}
		if ctx, ok := e.State.PendingInterrupt.Context.(*model.Context); ok {
			if ctx.Trigger == model.TriggerBeforeDraw {
				resumeCtx = ctx
			}
			if ctx.Trigger == model.TriggerOnAttackHit {
				resumeAttackHitCtx = ctx
			}
			if ctx.Trigger == model.TriggerOnAttackMiss {
				resumeAttackMissCtx = ctx
			}
			if ctx.Trigger == model.TriggerBeforeMoraleLoss {
				resumeMoraleCtx = ctx
			}
			if ctx.Trigger == model.TriggerOnPhaseEnd {
				resumePhaseEndCtx = ctx
			}
		}
		if resumeCtx == nil || resumeAttackHitCtx == nil || resumeAttackMissCtx == nil || resumeMoraleCtx == nil {
			if data, ok := e.State.PendingInterrupt.Context.(map[string]interface{}); ok {
				if userCtx, hasCtx := data["user_ctx"]; hasCtx {
					if ctx, ok := userCtx.(*model.Context); ok {
						if ctx.Trigger == model.TriggerBeforeDraw {
							resumeCtx = ctx
						}
						if ctx.Trigger == model.TriggerOnAttackHit {
							resumeAttackHitCtx = ctx
						}
						if ctx.Trigger == model.TriggerOnAttackMiss {
							resumeAttackMissCtx = ctx
						}
						if ctx.Trigger == model.TriggerBeforeMoraleLoss {
							resumeMoraleCtx = ctx
						}
						if ctx.Trigger == model.TriggerOnPhaseEnd {
							resumePhaseEndCtx = ctx
						}
					}
				}
			}
		}
	}

	// 使用 PopInterrupt 处理队列
	e.PopInterrupt()
	if holyLancerEarthSkipped && holyLancerEarthPlayerID != "" {
		if user := e.State.Players[holyLancerEarthPlayerID]; user != nil && e.isHolyLancer(user) {
			if user.Tokens == nil {
				user.Tokens = map[string]int{}
			}
			// 未发动地枪时，补触发圣击（若本次攻击未被天枪/地枪阻断）。
			if user.Tokens["holy_lancer_block_sacred_strike"] == 0 {
				e.Heal(user.ID, 1)
				e.Log(fmt.Sprintf("%s 未发动 [地枪]，触发 [圣击]：+1治疗", user.Name))
			}
		}
	}

	// 如果有暂停的扣卡流程，需要恢复
	if resumeCtx != nil {
		e.resumePendingDraw(resumeCtx)
		// 该伤害的摸牌已执行完毕，从队列中移除当前项，避免 processPendingDamages 再次结算同一条
		if len(e.State.PendingDamageQueue) > 0 {
			e.State.PendingDamageQueue = e.State.PendingDamageQueue[1:]
		}
	}
	// OnAttackHit 响应取消：推进延迟伤害到下一阶段，避免同一次命中再次弹响应框
	if resumeAttackHitCtx != nil && e.advancePendingAttackDamageStageAfterHit(resumeAttackHitCtx) {
		if e.State.PendingInterrupt == nil {
			e.State.Phase = model.PhasePendingDamageResolution
		}
		return nil
	}
	// OnAttackMiss 响应取消：恢复被中断的未命中后续流程（如防御成功/应战反弹）。
	if resumeAttackMissCtx != nil {
		if e.resumePendingAttackMiss(resumeAttackMissCtx) {
			return nil
		}
	}
	// TriggerBeforeMoraleLoss 响应取消：恢复士气损失结算。
	if resumeMoraleCtx != nil {
		if e.State.PendingInterrupt == nil && e.resumePendingMoraleLoss(resumeMoraleCtx) {
			return nil
		}
	}

	// 只有在没有待处理中断时才继续流程
	if e.State.PendingInterrupt == nil {
		// 若刚恢复的是摸牌前响应（伤害结算中的水影等），保持当前 Phase，让 Drive 继续结算伤害队列
		if resumeCtx != nil {
			// InterruptResponseSkill 会把阶段切到 Response，这里需要显式切回伤害结算。
			e.State.Phase = model.PhasePendingDamageResolution
			return nil
		}
		if resumePhaseEndCtx != nil {
			// 阶段结束响应（如法力潮汐）跳过后，需回到 ExtraAction 继续执行
			// 行动结束后的其余链路（如迅捷赐福）。
			e.State.Phase = model.PhaseExtraAction
			return nil
		}
		if len(e.State.ActionStack) > 0 {
			e.State.Phase = model.PhaseResponse
		} else if len(e.State.ActionQueue) > 0 {
			e.State.Phase = model.PhaseBeforeAction
		} else {
			e.State.Phase = model.PhaseTurnEnd
		}
	}

	return nil
}

// advancePendingAttackDamageStageAfterHit 将命中后响应取消的延迟伤害从 Stage0 推进到 Stage1
// 这样后续只继续受伤结算，不会再次触发 OnAttackHit 的响应弹框。
func (e *GameEngine) advancePendingAttackDamageStageAfterHit(ctx *model.Context) bool {
	if ctx == nil || ctx.TriggerCtx == nil || len(e.State.PendingDamageQueue) == 0 {
		return false
	}
	for i := range e.State.PendingDamageQueue {
		pd := &e.State.PendingDamageQueue[i]
		if !strings.EqualFold(pd.DamageType, "Attack") {
			continue
		}
		if pd.SourceID != ctx.TriggerCtx.SourceID || pd.TargetID != ctx.TriggerCtx.TargetID {
			continue
		}
		if pd.Stage == 0 {
			pd.Stage = 1
		}
		// Stage>=1 也视作“可恢复”，用于命中后多段响应（如先确认技能再选X）的续流程。
		return pd.Stage >= 1
	}
	return false
}

func (e *GameEngine) resolveHeroRoarMiss(attackerID string) {
	attacker := e.State.Players[attackerID]
	if attacker == nil || !e.isHero(attacker) {
		return
	}
	if attacker.Tokens == nil || attacker.Tokens["hero_roar_active"] <= 0 {
		return
	}
	attacker.Tokens["hero_roar_active"] = 0
	attacker.Tokens["hero_roar_damage_pending"] = 0
	wisdom := attacker.Tokens["hero_wisdom"] + 1
	if wisdom > heroTokenCapEngine {
		wisdom = heroTokenCapEngine
	}
	attacker.Tokens["hero_wisdom"] = wisdom
	e.Log(fmt.Sprintf("%s 的 [怒吼] 未命中分支生效：知性+1（当前%d）", attacker.Name, wisdom))
}

func (e *GameEngine) resolveFighterChargeMiss(attackerID string) {
	attacker := e.State.Players[attackerID]
	if attacker == nil || !e.isFighter(attacker) {
		return
	}
	if attacker.Tokens == nil || attacker.Tokens["fighter_charge_pending"] <= 0 {
		return
	}
	attacker.Tokens["fighter_charge_pending"] = 0
	attacker.Tokens["fighter_charge_damage_pending"] = 0
	damage := attacker.Tokens["fighter_qi"]
	if damage > 0 {
		e.AddPendingDamage(model.PendingDamage{
			SourceID:   attacker.ID,
			TargetID:   attacker.ID,
			Damage:     damage,
			DamageType: "magic",
			Stage:      0,
		})
	}
	e.Log(fmt.Sprintf("%s 的 [蓄力一击] 未命中分支生效：对自己造成%d点法术伤害", attacker.Name, damage))
}

func (e *GameEngine) resolveMagicBowPierceMiss(attackerID, targetID string) {
	e.resolveHeroRoarMiss(attackerID)
	e.resolveFighterChargeMiss(attackerID)
	e.resolveHolyBowShardMiss(attackerID, targetID)
	attacker := e.State.Players[attackerID]
	target := e.State.Players[targetID]
	if attacker == nil || target == nil {
		return
	}
	if attacker.Tokens == nil || attacker.Tokens["mb_magic_pierce_pending"] <= 0 {
		return
	}
	attacker.Tokens["mb_magic_pierce_pending"] = 0
	e.AddPendingDamage(model.PendingDamage{
		SourceID:   attackerID,
		TargetID:   targetID,
		Damage:     3,
		DamageType: "magic",
		Stage:      0,
	})
	e.Log(fmt.Sprintf("%s 的 [魔贯冲击] 未命中：对 %s 造成3点法术伤害", attacker.Name, target.Name))
}

func (e *GameEngine) resolveHolyBowShardMiss(attackerID, targetID string) {
	attacker := e.State.Players[attackerID]
	target := e.State.Players[targetID]
	if attacker == nil || target == nil || !e.isHolyBow(attacker) {
		return
	}
	if attacker.Tokens == nil || attacker.Tokens["hb_shard_miss_pending"] <= 0 {
		return
	}
	attacker.Tokens["hb_shard_miss_pending"] = 0
	maxX := attacker.Heal
	if maxX > 2 {
		maxX = 2
	}
	if maxX <= 0 {
		e.Log(fmt.Sprintf("%s 的 [圣屑飓暴] 未命中，但治疗不足，未触发后续效果", attacker.Name))
		return
	}
	allyIDs := make([]string, 0)
	for _, pid := range e.State.PlayerOrder {
		p := e.State.Players[pid]
		if p == nil || p.Camp != attacker.Camp || p.ID == attacker.ID {
			continue
		}
		allyIDs = append(allyIDs, p.ID)
	}
	if len(allyIDs) == 0 {
		e.Log(fmt.Sprintf("%s 的 [圣屑飓暴] 未命中，但无可选队友执行弃牌", attacker.Name))
		return
	}
	e.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: attacker.ID,
		Context: map[string]interface{}{
			"choice_type": "hb_holy_shard_miss_confirm",
			"user_id":     attacker.ID,
			"target_id":   targetID,
			"max_x":       maxX,
			"ally_ids":    allyIDs,
		},
	})
	e.Log(fmt.Sprintf("%s 的 [圣屑飓暴] 未命中：可移除治疗并令队友弃牌", attacker.Name))
}

// applyMoraleLossAfterTrigger 在 TriggerBeforeMoraleLoss 后应用士气损失与联动效果。
func (e *GameEngine) applyMoraleLossAfterTrigger(victim *model.Player, moraleLoss int, isMagic bool, fromDamageDraw bool, heroDeadDuelFloor bool, discardedCards []model.Card, lossCtx *model.Context) int {
	if victim == nil {
		if len(discardedCards) > 0 {
			e.State.DiscardPile = append(e.State.DiscardPile, discardedCards...)
		}
		return 0
	}

	finalLoss := moraleLoss
	if lossCtx != nil && lossCtx.TriggerCtx != nil && lossCtx.TriggerCtx.DamageVal != nil {
		finalLoss = *lossCtx.TriggerCtx.DamageVal
	}
	if finalLoss < 0 {
		finalLoss = 0
	}
	if heroDeadDuelFloor && finalLoss > 0 {
		finalLoss = 1
	}

	finalLoss = e.applyCampMoraleLoss(victim.Camp, finalLoss)

	// 血之巫女：普通形态下，因承伤导致我方士气下降时，强制进入流血形态并+1治疗。
	if fromDamageDraw && finalLoss > 0 && e.isBloodPriestess(victim) {
		if victim.Tokens == nil {
			victim.Tokens = map[string]int{}
		}
		if victim.Tokens["bp_bleed_form"] <= 0 {
			victim.Tokens["bp_bleed_form"] = 1
			e.Heal(victim.ID, 1)
			e.Log(fmt.Sprintf("%s 的 [流血] 触发：进入流血形态并获得1点治疗", victim.Name))
			_ = e.maybeAutoReleaseBloodPriestessByHand(victim, "手牌<3强制脱离流血形态")
		}
	}
	if fromDamageDraw && isMagic && finalLoss > 0 && e.isBlazeWitch(victim) {
		if victim.Tokens == nil {
			victim.Tokens = map[string]int{}
		}
		before := victim.Tokens["bw_rebirth"]
		victim.Tokens["bw_rebirth"]++
		if victim.Tokens["bw_rebirth"] > 4 {
			victim.Tokens["bw_rebirth"] = 4
		}
		if victim.Tokens["bw_rebirth"] != before {
			e.Log(fmt.Sprintf("%s 的 [永生银时计] 触发，重生+1（当前%d）", victim.Name, victim.Tokens["bw_rebirth"]))
		}
	}
	// 红莲骑士：仅当“伤害导致且实际发生士气下降”时，强制进入热血沸腾形态。
	if fromDamageDraw && finalLoss > 0 && e.isCrimsonKnight(victim) {
		if victim.Tokens == nil {
			victim.Tokens = map[string]int{}
		}
		if victim.Tokens["crk_hot_form"] == 0 {
			victim.Tokens["crk_hot_form"] = 1
			e.Log(fmt.Sprintf("%s 的 [热血沸腾] 触发，进入热血沸腾形态", victim.Name))
		}
	}
	if moraleLoss != finalLoss {
		e.Log(fmt.Sprintf("[System] 士气损失被抵御！原损失: %d, 实际损失: %d", moraleLoss, finalLoss))
	}
	absorbByMoonID := ""
	if lossCtx != nil && lossCtx.Selections != nil {
		absorbByMoonID, _ = lossCtx.Selections["mg_new_moon_absorb_by"].(string)
	}
	if absorbByMoonID == "" {
		e.State.DiscardPile = append(e.State.DiscardPile, discardedCards...)
	} else {
		e.Log(fmt.Sprintf("[Skill] %s 的爆牌被 [新月庇护] 吸收为暗月（未进入弃牌堆）", victim.Name))
	}
	return finalLoss
}

// resumePendingMoraleLoss 恢复被响应中断的士气损失结算。
func (e *GameEngine) resumePendingMoraleLoss(ctx *model.Context) bool {
	if ctx == nil || ctx.Selections == nil {
		return false
	}
	pending, _ := ctx.Selections["morale_loss_pending"].(bool)
	if !pending {
		return false
	}

	// 恢复上下文数据
	victimID, _ := ctx.Selections["victim_id"].(string)
	victim := e.State.Players[victimID]

	moraleLoss := 0
	switch v := ctx.Selections["morale_loss_value"].(type) {
	case int:
		moraleLoss = v
	case float64:
		moraleLoss = int(v)
	}
	isMagic, _ := ctx.Selections["is_magic"].(bool)
	fromDamageDraw, _ := ctx.Selections["from_damage_draw"].(bool)
	heroDeadDuelFloor, _ := ctx.Selections["hero_dead_duel_floor"].(bool)

	var discardedCards []model.Card
	switch v := ctx.Selections["discarded_cards"].(type) {
	case []model.Card:
		discardedCards = append(discardedCards, v...)
	case []interface{}:
		for _, item := range v {
			if c, ok := item.(model.Card); ok {
				discardedCards = append(discardedCards, c)
			} else if m, ok := item.(map[string]interface{}); ok {
				var c model.Card
				if name, _ := m["name"].(string); name != "" {
					c.Name = name
				}
				if element, _ := m["element"].(string); element != "" {
					c.Element = model.Element(element)
				}
				if c.Name != "" {
					discardedCards = append(discardedCards, c)
				}
			}
		}
	}

	finalLoss := e.applyMoraleLossAfterTrigger(victim, moraleLoss, isMagic, fromDamageDraw, heroDeadDuelFloor, discardedCards, ctx)
	mbChargeResume, _ := ctx.Selections["mb_charge_resume"].(bool)
	discardPlayerID, _ := ctx.Selections["discard_player_id"].(string)
	discardPlayer := e.State.Players[discardPlayerID]
	if discardPlayer == nil {
		discardPlayer = victim
	}
	if mbChargeResume {
		if victim != nil {
			e.Log(fmt.Sprintf("%s 的 [充能] 爆士气结算完成：士气-%d（本次不弃牌）", victim.Name, finalLoss))
		}
	} else if discardPlayer != nil {
		e.Log(fmt.Sprintf("[System] %s 丢弃了 %d 张牌！士气 -%d", discardPlayer.Name, len(discardedCards), finalLoss))
		// 魔枪：幻影星尘若因本次自伤进入爆牌弃牌，需要在此处完成“完全结算后”的后续判定。
		if e.isMagicLancer(discardPlayer) && discardPlayer.Tokens != nil && discardPlayer.Tokens["ml_stardust_wait_discard"] > 0 {
			e.resolveMagicLancerStardustAfterSelf(discardPlayer)
		}
	}
	ctx.Selections["morale_loss_pending"] = false
	e.checkGameEnd()
	if mbChargeResume {
		if e.State.Phase == model.PhaseEnd {
			return true
		}
		maxPlace := toIntContextValue(ctx.Selections["mb_charge_max_place"])
		userID, _ := ctx.Selections["mb_charge_user_id"].(string)
		user := e.State.Players[userID]
		if e.State.PendingInterrupt == nil {
			if user != nil && maxPlace > 0 {
				e.PushInterrupt(&model.Interrupt{
					Type:     model.InterruptChoice,
					PlayerID: user.ID,
					Context: map[string]interface{}{
						"choice_type": "mb_charge_place_count",
						"user_id":     user.ID,
						"max_place":   maxPlace,
					},
				})
			} else {
				e.State.Phase = model.PhaseStartup
			}
		}
		return true
	}

	stayInTurn, _ := ctx.Selections["morale_loss_stay_in_turn"].(bool)
	isDamageResolution, _ := ctx.Selections["morale_loss_is_damage_resolution"].(bool)
	if e.State.PendingInterrupt == nil {
		if isDamageResolution {
			e.State.Phase = model.PhaseExtraAction
		} else if stayInTurn {
			e.Log("[System] 弃牌完成，继续当前回合")
			if e.State.ReturnPhase != "" {
				e.State.Phase = e.State.ReturnPhase
				e.State.ReturnPhase = ""
			} else if len(e.State.ActionQueue) > 0 {
				e.State.Phase = model.PhaseBeforeAction
			} else if len(e.State.PendingDamageQueue) > 0 {
				e.State.Phase = model.PhasePendingDamageResolution
			} else {
				e.State.Phase = model.PhaseStartup
			}
		} else {
			e.State.Phase = model.PhaseTurnEnd
		}
	}
	return true
}

// resumePendingAttackMiss 恢复被响应中断打断的“攻击未命中后续流程”。
// 返回 true 表示已完成恢复并设置了下一阶段。
func (e *GameEngine) resumePendingAttackMiss(ctx *model.Context) bool {
	if ctx == nil || ctx.Selections == nil || len(e.State.CombatStack) == 0 {
		return false
	}
	raw := ctx.Selections["attack_miss_resume"]
	data, ok := raw.(map[string]interface{})
	if !ok || data == nil {
		return false
	}
	mode, _ := data["mode"].(string)
	if mode == "" {
		return false
	}
	attackerID, _ := data["attacker_id"].(string)
	targetID, _ := data["target_id"].(string)
	top := e.State.CombatStack[len(e.State.CombatStack)-1]
	if attackerID != "" && top.AttackerID != attackerID {
		return false
	}
	if targetID != "" && top.TargetID != targetID {
		return false
	}

	switch mode {
	case "defend":
		defender := e.State.Players[top.TargetID]
		if defender != nil {
			e.Log(fmt.Sprintf("[Combat] %s 防御成功，攻击未命中", defender.Name))
		}
		e.resolveMagicBowPierceMiss(top.AttackerID, top.TargetID)
		if atk := e.State.Players[top.AttackerID]; atk != nil && atk.Tokens != nil {
			atk.Tokens["elf_elemental_shot_thunder_pending"] = 0
		}
		e.clearCombatStack()
		if len(e.State.PendingDamageQueue) > 0 {
			e.State.Phase = model.PhasePendingDamageResolution
			e.State.ReturnPhase = model.PhaseExtraAction
		} else {
			e.State.Phase = model.PhaseExtraAction
		}
		return true
	case "counter":
		counterPlayerID, _ := data["counter_player_id"].(string)
		counterTargetID, _ := data["counter_target_id"].(string)
		var counterCard model.Card
		switch v := data["counter_card"].(type) {
		case model.Card:
			counterCard = v
		case *model.Card:
			if v != nil {
				counterCard = *v
			}
		}
		if counterPlayerID == "" || counterTargetID == "" || counterCard.Name == "" {
			return false
		}
		counterPlayer := e.State.Players[counterPlayerID]
		counterTarget := e.State.Players[counterTargetID]
		if counterPlayer != nil && counterTarget != nil {
			e.Log(fmt.Sprintf("[Combat] %s 使用 %s 应战成功！攻击反弹给 %s",
				counterPlayer.Name, counterCard.Name, counterTarget.Name))
		}
		e.resolveMagicBowPierceMiss(top.AttackerID, top.TargetID)
		if atk := e.State.Players[top.AttackerID]; atk != nil && atk.Tokens != nil {
			atk.Tokens["elf_elemental_shot_thunder_pending"] = 0
		}
		e.State.CombatStack = e.State.CombatStack[:len(e.State.CombatStack)-1]
		e.initCombat(counterPlayerID, counterTargetID, &counterCard, false, true, true)
		if counterPlayer != nil && counterTarget != nil {
			e.Log(fmt.Sprintf("[Combat] %s 应战成功！攻击转移向 %s", counterPlayer.Name, counterTarget.Name))
		}
		return true
	default:
		return false
	}
}

// tryStartOnmyojiBindingInterrupt 检查并触发“式神咒束”代应战确认。
// 返回 true 表示已进入中断等待（应暂停当前 Drive）。
func (e *GameEngine) tryStartOnmyojiBindingInterrupt(combatReq *model.CombatRequest) bool {
	if combatReq == nil {
		return false
	}
	if combatReq.OnmyojiBindingChecked {
		return false
	}
	combatReq.OnmyojiBindingChecked = true

	if combatReq.IsCounter || combatReq.IsForcedHit || !combatReq.CanBeResponded || combatReq.Card == nil {
		return false
	}
	if combatReq.Card.Element == model.ElementDark {
		return false
	}
	target := e.State.Players[combatReq.TargetID]
	attacker := e.State.Players[combatReq.AttackerID]
	if target == nil || attacker == nil {
		return false
	}
	// 仅“敌方主动攻击队友”场景可触发。
	if attacker.Camp == target.Camp {
		return false
	}
	// 阴阳师代应战仅在“队友成为攻击目标”时触发。
	if e.isOnmyoji(target) {
		return false
	}

	var counterTargetIDs []string
	for _, pid := range e.State.PlayerOrder {
		if pid == attacker.ID {
			continue
		}
		p := e.State.Players[pid]
		if p == nil || p.Camp != attacker.Camp {
			continue
		}
		counterTargetIDs = append(counterTargetIDs, pid)
	}
	if len(counterTargetIDs) == 0 {
		return false
	}

	for _, pid := range e.State.PlayerOrder {
		actor := e.State.Players[pid]
		if actor == nil || actor.ID == target.ID {
			continue
		}
		if !e.isOnmyoji(actor) || actor.Camp != target.Camp {
			continue
		}
		if actor.Tokens == nil || actor.Tokens["onmyoji_form"] <= 0 {
			continue
		}
		if !e.canPayOnmyojiBindingCost(actor.Camp) {
			continue
		}
		cardOptions := collectOnmyojiCounterOptions(actor, combatReq.Card)
		if len(cardOptions) == 0 {
			continue
		}
		e.PushInterrupt(&model.Interrupt{
			Type:     model.InterruptChoice,
			PlayerID: actor.ID,
			Context: map[string]interface{}{
				"choice_type":        "onmyoji_binding_confirm",
				"actor_id":           actor.ID,
				"attacker_id":        combatReq.AttackerID,
				"target_id":          combatReq.TargetID,
				"card_options":       cardOptions,
				"counter_target_ids": counterTargetIDs,
			},
		})
		e.Log(fmt.Sprintf("%s 可发动 [式神咒束] 代应战，等待其确认", actor.Name))
		return true
	}
	return false
}

// tryStartOnmyojiYinYangInterrupt 检查并触发“阴阳转换”优先确认。
// 规则：目标阴阳师若手里有“与来袭攻击同命格”的攻击牌，则先询问是否发动；
// 若选择不发动，才进入常规 承受/防御/应战 弹框。
func (e *GameEngine) tryStartOnmyojiYinYangInterrupt(combatReq *model.CombatRequest) bool {
	if combatReq == nil {
		return false
	}
	if combatReq.OnmyojiYinYangChecked {
		return false
	}
	combatReq.OnmyojiYinYangChecked = true

	if combatReq.IsCounter || combatReq.IsForcedHit || !combatReq.CanBeResponded || combatReq.Card == nil {
		return false
	}
	if combatReq.Card.Element == model.ElementDark {
		// 暗灭无法应战，不询问阴阳转换。
		return false
	}

	target := e.State.Players[combatReq.TargetID]
	attacker := e.State.Players[combatReq.AttackerID]
	if target == nil || attacker == nil || !e.isOnmyoji(target) {
		return false
	}
	if !onmyojiCanUseFactionCounter(combatReq.Card) {
		return false
	}

	// 阴阳转换只看“同命格应战”分支（不含普通同系/暗灭应战）。
	allOptions := collectOnmyojiCounterOptions(target, combatReq.Card)
	var factionOptions []map[string]interface{}
	for _, opt := range allOptions {
		useFaction, _ := opt["use_faction"].(bool)
		if useFaction {
			factionOptions = append(factionOptions, opt)
		}
	}
	if len(factionOptions) == 0 {
		return false
	}

	// 应战反弹目标：攻击方的队友（不含攻击者本人）。
	var counterTargetIDs []string
	for _, pid := range e.State.PlayerOrder {
		if pid == attacker.ID {
			continue
		}
		p := e.State.Players[pid]
		if p == nil || p.Camp != attacker.Camp {
			continue
		}
		counterTargetIDs = append(counterTargetIDs, pid)
	}
	if len(counterTargetIDs) == 0 {
		return false
	}

	e.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: target.ID,
		Context: map[string]interface{}{
			"choice_type":        "onmyoji_yinyang_confirm",
			"actor_id":           target.ID,
			"attacker_id":        combatReq.AttackerID,
			"target_id":          combatReq.TargetID,
			"card_options":       factionOptions,
			"counter_target_ids": counterTargetIDs,
		},
	})
	e.Log(fmt.Sprintf("%s 可发动 [阴阳转换]，等待其确认", target.Name))
	return true
}

// executeOnmyojiBindingCounter 在战斗阶段自动执行已确认的“式神咒束应战”。
// 返回 true 表示已推进流程（可能进入中断），当前 Drive 应暂停。
func (e *GameEngine) executeOnmyojiBindingCounter(combatReq *model.CombatRequest) bool {
	if combatReq == nil {
		return false
	}
	actorID := combatReq.OnmyojiBindingActorID
	cardID := combatReq.OnmyojiBindingCounterID
	counterTargetID := combatReq.OnmyojiBindingTargetID
	if actorID == "" || cardID == "" || counterTargetID == "" {
		return false
	}
	actor := e.State.Players[actorID]
	if actor == nil || combatReq.Card == nil {
		combatReq.OnmyojiBindingActorID = ""
		combatReq.OnmyojiBindingCounterID = ""
		combatReq.OnmyojiBindingTargetID = ""
		combatReq.OnmyojiBindingUseFaction = false
		return false
	}
	cardIdx := findPlayableCardIndexByID(actor, cardID)
	card, _, _, ok := getPlayableCardByIndex(actor, cardIdx)
	if !ok || card.Type != model.CardTypeAttack {
		combatReq.OnmyojiBindingActorID = ""
		combatReq.OnmyojiBindingCounterID = ""
		combatReq.OnmyojiBindingTargetID = ""
		combatReq.OnmyojiBindingUseFaction = false
		return false
	}
	useFaction := combatReq.OnmyojiBindingUseFaction
	canCounter := card.Element == combatReq.Card.Element || card.Element == model.ElementDark
	if !canCounter && useFaction {
		canCounter = onmyojiCanUseFactionCounter(combatReq.Card) &&
			card.Faction != "" && card.Faction == combatReq.Card.Faction
	}
	if !canCounter {
		combatReq.OnmyojiBindingActorID = ""
		combatReq.OnmyojiBindingCounterID = ""
		combatReq.OnmyojiBindingTargetID = ""
		combatReq.OnmyojiBindingUseFaction = false
		return false
	}

	e.NotifyCardRevealed(actor.ID, []model.Card{card}, "counter")
	e.NotifyCombatCue(combatReq.AttackerID, combatReq.TargetID, "counter")
	if _, err := consumePlayableCardByIndex(actor, cardIdx); err != nil {
		combatReq.OnmyojiBindingActorID = ""
		combatReq.OnmyojiBindingCounterID = ""
		combatReq.OnmyojiBindingTargetID = ""
		combatReq.OnmyojiBindingUseFaction = false
		return false
	}
	e.State.DiscardPile = append(e.State.DiscardPile, card)

	if useFaction {
		if actor.Tokens == nil {
			actor.Tokens = map[string]int{}
		}
		actor.Tokens["onmyoji_ghost_fire"]++
		if actor.Tokens["onmyoji_ghost_fire"] > 3 {
			actor.Tokens["onmyoji_ghost_fire"] = 3
		}
		e.Log(fmt.Sprintf("%s 的 [阴阳转换] 触发，鬼火+1", actor.Name))
		// 式神形态内联动触发式神转换：摸1并再+1鬼火，然后脱离形态。
		if actor.Tokens["onmyoji_form"] > 0 {
			e.DrawCards(actor.ID, 1)
			actor.Tokens["onmyoji_ghost_fire"]++
			if actor.Tokens["onmyoji_ghost_fire"] > 3 {
				actor.Tokens["onmyoji_ghost_fire"] = 3
			}
			actor.Tokens["onmyoji_form"] = 0
			e.Log(fmt.Sprintf("%s 的 [式神转换] 触发：摸1并鬼火+1，然后脱离式神形态", actor.Name))
		}
		card.Damage = actor.Tokens["onmyoji_ghost_fire"]
		if card.Damage < 0 {
			card.Damage = 0
		}
	}

	missCtx := &model.EventContext{
		Type:     model.EventAttack,
		SourceID: combatReq.AttackerID,
		TargetID: combatReq.TargetID,
		Card:     combatReq.Card,
		AttackInfo: &model.AttackEventInfo{
			ActionType: string(model.ActionAttack),
			CounterInitiator: func() string {
				if combatReq.IsCounter {
					return combatReq.AttackerID
				}
				return ""
			}(),
		},
	}
	skillCtx := e.buildContext(e.State.Players[combatReq.AttackerID], e.State.Players[combatReq.TargetID], model.TriggerOnAttackMiss, missCtx)
	skillCtx.Selections["attack_miss_resume"] = map[string]interface{}{
		"mode":              "counter",
		"attacker_id":       combatReq.AttackerID,
		"target_id":         combatReq.TargetID,
		"counter_player_id": actor.ID,
		"counter_target_id": counterTargetID,
		"counter_card":      card,
	}
	e.dispatcher.OnTrigger(model.TriggerOnAttackMiss, skillCtx)
	if e.State.PendingInterrupt != nil {
		return true
	}

	e.resolveMagicBowPierceMiss(combatReq.AttackerID, combatReq.TargetID)
	if attacker := e.State.Players[combatReq.AttackerID]; attacker != nil && attacker.Tokens != nil {
		attacker.Tokens["elf_elemental_shot_thunder_pending"] = 0
	}
	e.Log(fmt.Sprintf("[Combat] %s 通过[式神咒束]代应战成功，攻击反弹给 %s", actor.Name, model.GetPlayerDisplayName(e.State.Players[counterTargetID])))
	e.State.CombatStack = e.State.CombatStack[:len(e.State.CombatStack)-1]
	e.initCombat(actor.ID, counterTargetID, &card, false, true, true)
	combatReq.OnmyojiBindingActorID = ""
	combatReq.OnmyojiBindingCounterID = ""
	combatReq.OnmyojiBindingTargetID = ""
	combatReq.OnmyojiBindingUseFaction = false
	return true
}

func (e *GameEngine) hasUsableShieldForCombat(target *model.Player, combatReq model.CombatRequest) bool {
	if target == nil {
		return false
	}
	if attacker := e.State.Players[combatReq.AttackerID]; attacker != nil && attacker.TurnState.GaleSlashActive {
		// 烈风技：无视圣盾
		return false
	}
	if attacker := e.State.Players[combatReq.AttackerID]; attacker != nil && attacker.Tokens != nil &&
		attacker.Tokens["berserker_blood_roar_ignore_shield"] > 0 {
		// 血腥咆哮：本次攻击无视圣盾
		return false
	}
	for _, fc := range target.Field {
		if fc != nil && fc.Mode == model.FieldEffect && fc.Effect == model.EffectShield {
			return true
		}
	}
	return false
}

func (e *GameEngine) consumeShieldForCombatTake(target *model.Player, combatReq model.CombatRequest) bool {
	if !e.hasUsableShieldForCombat(target, combatReq) {
		return false
	}
	if target == nil {
		return false
	}

	removed := false
	for _, fc := range target.Field {
		if fc == nil || fc.Mode != model.FieldEffect || fc.Effect != model.EffectShield {
			continue
		}
		target.RemoveFieldCard(fc)
		e.State.DiscardPile = append(e.State.DiscardPile, fc.Card)
		removed = true
		break
	}
	if !removed {
		return false
	}

	e.addActionResponse(fmt.Sprintf("%s 的【圣盾】自动抵挡本次攻击", target.Name))
	e.NotifyActionStep(fmt.Sprintf("%s 的【圣盾】触发，自动抵挡了本次攻击", target.Name))
	e.NotifyCombatCue(combatReq.AttackerID, combatReq.TargetID, "defend")
	e.Log(fmt.Sprintf("[Combat] %s 选择承受伤害，触发【圣盾】抵挡本次攻击！", target.Name))
	e.resolveMagicBowPierceMiss(combatReq.AttackerID, combatReq.TargetID)
	if atk := e.State.Players[combatReq.AttackerID]; atk != nil && atk.Tokens != nil {
		atk.Tokens["elf_elemental_shot_thunder_pending"] = 0
	}
	e.clearCombatStack()
	if len(e.State.PendingDamageQueue) > 0 {
		e.State.Phase = model.PhasePendingDamageResolution
		e.State.ReturnPhase = model.PhaseExtraAction
	} else {
		e.State.Phase = model.PhaseExtraAction
	}
	return true
}

// handleCombatResponse 处理战斗交互阶段的响应
func (e *GameEngine) handleCombatResponse(act model.PlayerAction) error {
	if len(e.State.CombatStack) == 0 {
		return errors.New("响应时，战斗栈为空")
	}

	if len(act.ExtraArgs) == 0 {
		return errors.New("缺少响应类型")
	}

	respType := act.ExtraArgs[0]                                 // take, defend, counter
	combatReq := e.State.CombatStack[len(e.State.CombatStack)-1] // 查看栈顶

	// 验证响应者是否是当前目标
	if act.PlayerID != combatReq.TargetID {
		return fmt.Errorf("不是 %s 的响应回合", e.State.Players[combatReq.TargetID].Name)
	}

	player := e.State.Players[act.PlayerID]
	if player == nil {
		return errors.New("玩家不存在")
	}

	switch respType {
	case "take", "hit":
		// 优先给玩家应战/防御机会；只有在其明确选择承受后，才触发场上圣盾抵挡。
		if e.consumeShieldForCombatTake(player, combatReq) {
			return nil
		}

		// 承受伤害：将伤害事件推入 PendingDamageQueue 进行统一处理
		// 这样可以支持多阶段触发 (AttackHit -> DamageTaken) 和中断恢复
		if attacker := e.State.Players[combatReq.AttackerID]; attacker != nil && attacker.Tokens != nil {
			attacker.Tokens["elf_elemental_shot_thunder_pending"] = 0
		}
		e.clearCombatStack()

		pd := model.PendingDamage{
			SourceID:   combatReq.AttackerID,
			TargetID:   combatReq.TargetID,
			Damage:     combatReq.Card.Damage,
			DamageType: "Attack",
			Card:       combatReq.Card,
			Stage:      0,
			IsCounter:  combatReq.IsCounter, // 应战命中→加水晶，主动命中→加宝石
		}

		// 战斗伤害优先处理，插入到队列头部
		e.State.PendingDamageQueue = append([]model.PendingDamage{pd}, e.State.PendingDamageQueue...)

		e.addActionResponse(fmt.Sprintf("%s 承受伤害", player.Name))
		e.NotifyActionStep(fmt.Sprintf("%s承受伤害", model.GetPlayerDisplayName(player)))
		e.NotifyCombatCue(combatReq.AttackerID, combatReq.TargetID, "take")
		e.Log(fmt.Sprintf("[Combat] %s 选择承受伤害，进入伤害结算流程", player.Name))

		// 设置阶段为延迟伤害结算
		e.State.Phase = model.PhasePendingDamageResolution
		// 战斗结束后，应进入额外行动阶段 (检查风怒等)
		e.State.ReturnPhase = model.PhaseExtraAction
		return nil

	case "defend":
		// 防御：仅允许打出【圣光】；【圣盾】必须提前放置为场上效果并自动生效。
		if e.isMagicLancer(player) {
			return errors.New("魔枪受[黑暗束缚]影响，不能使用法术牌防御")
		}
		card, _, _, ok := getPlayableCardByIndex(player, act.CardIndex)
		if !ok {
			return errors.New("无效的卡牌索引")
		}
		if card.Type != model.CardTypeMagic {
			return errors.New("只能使用法术牌进行防御")
		}
		if card.Name == "圣盾" {
			return errors.New("【圣盾】不能在防御时打出，请提前放置到场上触发")
		}
		if card.Name != "圣光" {
			return errors.New("防御只能使用【圣光】；【圣盾】需提前放置到场上")
		}

		e.triggerSealDamageForCardUse(player, &card)
		e.NotifyCardRevealed(act.PlayerID, []model.Card{card}, "defend")
		e.NotifyCombatCue(combatReq.AttackerID, combatReq.TargetID, "defend")
		// 消耗防御牌
		if _, err := consumePlayableCardByIndex(player, act.CardIndex); err != nil {
			return err
		}
		e.State.DiscardPile = append(e.State.DiscardPile, card)

		// 触发攻击者的 AttackMiss 事件（防御=攻击未命中）
		missCtx := &model.EventContext{
			Type:     model.EventAttack,
			SourceID: combatReq.AttackerID,
			TargetID: combatReq.TargetID,
			Card:     combatReq.Card,
			AttackInfo: &model.AttackEventInfo{
				ActionType: string(model.ActionAttack),
				CounterInitiator: func() string {
					if combatReq.IsCounter {
						return combatReq.AttackerID
					}
					return ""
				}(),
			},
		}
		skillCtx := e.buildContext(e.State.Players[combatReq.AttackerID], e.State.Players[combatReq.TargetID], model.TriggerOnAttackMiss, missCtx)
		skillCtx.Selections["attack_miss_resume"] = map[string]interface{}{
			"mode":        "defend",
			"attacker_id": combatReq.AttackerID,
			"target_id":   combatReq.TargetID,
		}
		e.dispatcher.OnTrigger(model.TriggerOnAttackMiss, skillCtx)
		// 若产生中断（如贯穿射击），需要立即 return，不执行后续 clearCombatStack
		if e.State.PendingInterrupt != nil {
			return nil
		}

		e.Log(fmt.Sprintf("[Combat] %s 使用 %s 防御成功！", player.Name, card.Name))
		e.resolveMagicBowPierceMiss(combatReq.AttackerID, combatReq.TargetID)
		if attacker := e.State.Players[combatReq.AttackerID]; attacker != nil && attacker.Tokens != nil {
			attacker.Tokens["elf_elemental_shot_thunder_pending"] = 0
		}
		e.clearCombatStack()
		if len(e.State.PendingDamageQueue) > 0 {
			e.State.Phase = model.PhasePendingDamageResolution
			e.State.ReturnPhase = model.PhaseExtraAction
		} else {
			e.State.Phase = model.PhaseExtraAction
		}

		return nil

	case "counter":
		// 应战：验证目标，推入新的 CombatRequest，调用 Drive()（形成递归）
		if !combatReq.CanBeResponded {
			return errors.New("此攻击无法被应战")
		}

		card, _, _, ok := getPlayableCardByIndex(player, act.CardIndex)
		if !ok {
			return errors.New("无效的卡牌索引")
		}
		if card.Type != model.CardTypeAttack {
			return errors.New("只能使用攻击牌进行应战")
		}
		card = e.applyBlazeWitchAttackCardTransform(player, card)

		// 验证应战卡牌元素（规则：同系或暗灭）
		// 暗灭不可被应战，只能承受伤害或使用圣光（若有场上圣盾会自动生效）
		if combatReq.Card.Element == model.ElementDark {
			return errors.New("暗灭无法被应战，只能承受伤害或使用圣光抵挡（场上圣盾会自动生效）")
		}
		useFactionCounter := false
		// 非暗灭攻击：只能用同系攻击牌或暗灭应战
		if card.Element != combatReq.Card.Element && card.Element != model.ElementDark {
			// 阴阳师可通过“阴阳转换”以同命格应战（非欺诈）。
			if e.isOnmyoji(player) && onmyojiCanUseFactionCounter(combatReq.Card) &&
				card.Faction != "" && card.Faction == combatReq.Card.Faction {
				useFactionCounter = true
			} else {
				return fmt.Errorf("应战必须使用同系攻击牌或暗灭，对方为 %s 系", combatReq.Card.Element)
			}
		}

		// 应战只能反弹给攻击方的队友，不能选择攻击者本人
		targetID := act.TargetID
		if targetID == "" {
			return errors.New("应战必须指定反弹目标（从攻击方队友中选择）")
		}
		if targetID == combatReq.AttackerID {
			return errors.New("不能选择攻击者本人，只能选择攻击方的队友进行反弹")
		}

		target := e.State.Players[targetID]
		if target == nil {
			return errors.New("目标不存在")
		}

		attacker := e.State.Players[combatReq.AttackerID]
		if attacker == nil {
			return errors.New("攻击者信息异常")
		}
		// 目标必须是攻击方的队友
		if target.Camp != attacker.Camp {
			return errors.New("应战反弹目标必须是攻击方的队友")
		}

		e.triggerSealDamageForCardUse(player, &card)
		e.NotifyCardRevealed(act.PlayerID, []model.Card{card}, "counter")
		e.NotifyCombatCue(combatReq.AttackerID, combatReq.TargetID, "counter")
		// 消耗应战牌
		if _, err := consumePlayableCardByIndex(player, act.CardIndex); err != nil {
			return err
		}
		e.State.DiscardPile = append(e.State.DiscardPile, card)
		if useFactionCounter {
			if player.Tokens == nil {
				player.Tokens = map[string]int{}
			}
			player.Tokens["onmyoji_ghost_fire"]++
			if player.Tokens["onmyoji_ghost_fire"] > 3 {
				player.Tokens["onmyoji_ghost_fire"] = 3
			}
			e.Log(fmt.Sprintf("%s 的 [阴阳转换] 触发，鬼火+1", player.Name))
			// 处于式神形态时联动式神转换：摸1并再+1鬼火，然后脱离式神形态。
			if player.Tokens["onmyoji_form"] > 0 {
				e.DrawCards(player.ID, 1)
				player.Tokens["onmyoji_ghost_fire"]++
				if player.Tokens["onmyoji_ghost_fire"] > 3 {
					player.Tokens["onmyoji_ghost_fire"] = 3
				}
				player.Tokens["onmyoji_form"] = 0
				e.Log(fmt.Sprintf("%s 的 [式神转换] 触发：摸1并鬼火+1，然后脱离式神形态", player.Name))
			}
			card.Damage = player.Tokens["onmyoji_ghost_fire"]
			if card.Damage < 0 {
				card.Damage = 0
			}
		}

		// [新增] 触发原攻击者的 AttackMiss 事件
		missCtx := &model.EventContext{
			Type:     model.EventAttack,
			SourceID: combatReq.AttackerID,
			TargetID: combatReq.TargetID,
			Card:     combatReq.Card,
			AttackInfo: &model.AttackEventInfo{
				ActionType: string(model.ActionAttack),
				CounterInitiator: func() string {
					if combatReq.IsCounter {
						return combatReq.AttackerID
					}
					return ""
				}(),
			},
		}
		skillCtx := e.buildContext(e.State.Players[combatReq.AttackerID], e.State.Players[combatReq.TargetID], model.TriggerOnAttackMiss, missCtx)
		skillCtx.Selections["attack_miss_resume"] = map[string]interface{}{
			"mode":              "counter",
			"attacker_id":       combatReq.AttackerID,
			"target_id":         combatReq.TargetID,
			"counter_player_id": act.PlayerID,
			"counter_target_id": targetID,
			"counter_card":      card,
		}
		e.dispatcher.OnTrigger(model.TriggerOnAttackMiss, skillCtx)
		if e.State.PendingInterrupt != nil {
			return nil
		}

		e.Log(fmt.Sprintf("[Combat] %s 使用 %s 应战成功！攻击反弹给 %s",
			player.Name, card.Name, target.Name))
		e.resolveMagicBowPierceMiss(combatReq.AttackerID, combatReq.TargetID)
		if origAttacker := e.State.Players[combatReq.AttackerID]; origAttacker != nil && origAttacker.Tokens != nil {
			origAttacker.Tokens["elf_elemental_shot_thunder_pending"] = 0
		}

		// 弹出当前战斗请求
		e.State.CombatStack = e.State.CombatStack[:len(e.State.CombatStack)-1]

		// 创建新的战斗请求（应战反弹，IsCounter=true）
		e.initCombat(act.PlayerID, targetID, &card, false, true, true)

		e.Log(fmt.Sprintf("[Combat] %s 应战成功！攻击转移向 %s", player.Name, target.Name))

		if len(e.State.PendingDamageQueue) > 0 {
			e.State.Phase = model.PhasePendingDamageResolution
			e.State.ReturnPhase = model.PhaseCombatInteraction
		}

		return nil

	default:
		return fmt.Errorf("未知的响应类型: %s", respType)
	}
}

func (e *GameEngine) forceTurnTo(targetPID string) error {
	// 寻找玩家索引
	foundIdx := -1
	for i, pid := range e.State.PlayerOrder {
		if pid == targetPID {
			foundIdx = i
			break
		}
	}
	if foundIdx == -1 {
		return fmt.Errorf("玩家不存在: %s", targetPID)
	}

	// 先结束当前玩家的回合状态
	currentPID := e.State.PlayerOrder[e.State.CurrentTurn]
	if curr := e.State.Players[currentPID]; curr != nil {
		curr.IsActive = false
	}

	// 设置新玩家
	e.State.CurrentTurn = foundIdx
	newPlayer := e.State.Players[targetPID]
	newPlayer.IsActive = true
	newPlayer.TurnState = model.NewPlayerTurnState()

	e.State.Phase = model.PhaseActionSelection // 重置到行动选择阶段
	e.State.HasPerformedStartup = false

	return nil
}

func (e *GameEngine) debugFindCharacter(roleID string) *model.Character {
	if roleID == "" {
		return nil
	}
	characters := data.GetCharacters()
	for _, c := range characters {
		if c.ID == roleID || c.Name == roleID {
			charCopy := c
			return &charCopy
		}
	}
	return nil
}

func (e *GameEngine) debugResetPlayerForRole(player *model.Player, char *model.Character) {
	if player == nil || char == nil {
		return
	}
	player.Role = char.ID
	player.Character = char
	player.MaxHand = char.MaxHand
	player.MaxHeal = 2
	player.Heal = 0
	player.Gem = 0
	player.Crystal = 0
	player.Hand = []model.Card{}
	player.Blessings = []model.Card{}
	player.ExclusiveCards = []model.Card{}
	player.Field = []*model.FieldCard{}
	player.Buffs = []model.Buff{}
	player.Tokens = map[string]int{}
	player.CharaZone = nil
	player.TurnState = model.NewPlayerTurnState()
	e.applyRoleDefaults(player)
	e.ensureStarterRoleCards(player)
}

func (e *GameEngine) debugPickOtherPlayer(excludeID string) *model.Player {
	for _, pid := range e.State.PlayerOrder {
		if pid == excludeID {
			continue
		}
		if p := e.State.Players[pid]; p != nil {
			return p
		}
	}
	return nil
}

func (e *GameEngine) debugFindSkill(player *model.Player, skillID string) (model.SkillDefinition, bool) {
	if player == nil || player.Character == nil || skillID == "" {
		return model.SkillDefinition{}, false
	}
	for _, s := range player.Character.Skills {
		if s.ID == skillID {
			return s, true
		}
	}
	return model.SkillDefinition{}, false
}

func debugFindCardTemplate(deck []model.Card, element model.Element, cardType model.CardType) *model.Card {
	for _, c := range deck {
		if c.Type == cardType && c.Element == element {
			card := c
			return &card
		}
	}
	for _, c := range deck {
		if c.Type == cardType {
			card := c
			return &card
		}
	}
	return nil
}

func debugBuildExclusiveCardForCharacter(ownerID string, char *model.Character, skillTitle string) model.Card {
	charName := ""
	faction := ""
	if char != nil {
		charName = char.Name
		faction = char.Faction
	}
	card := model.Card{
		ID:              fmt.Sprintf("debug-exclusive-%s-%d", ownerID, time.Now().UnixNano()),
		Name:            skillTitle,
		Type:            model.CardTypeMagic,
		Element:         model.ElementLight,
		Faction:         faction,
		Damage:          0,
		Description:     "调试专属牌",
		ExclusiveChar1:  charName,
		ExclusiveSkill1: skillTitle,
	}
	return card
}

func (e *GameEngine) debugBuildExclusiveCard(player *model.Player, skillTitle string) model.Card {
	if player == nil {
		return model.Card{}
	}
	return debugBuildExclusiveCardForCharacter(player.ID, player.Character, skillTitle)
}

func (e *GameEngine) debugBuildExclusiveCardByRole(player *model.Player, char *model.Character, skillTitle string) model.Card {
	if player == nil {
		return model.Card{}
	}
	return debugBuildExclusiveCardForCharacter(player.ID, char, skillTitle)
}

func debugRemoveCardIndices(src []model.Card, indices []int) []model.Card {
	if len(indices) == 0 {
		return src
	}
	// 索引按升序传入，这里倒序删除避免位移问题。
	for i := len(indices) - 1; i >= 0; i-- {
		idx := indices[i]
		if idx < 0 || idx >= len(src) {
			continue
		}
		src = append(src[:idx], src[idx+1:]...)
	}
	return src
}

// debugDrawExclusiveCardsFromStock 从当前牌库/弃牌堆中抽取满足独有标记的卡牌。
// 约束：调试模式下不再构造自定义独有牌，必须来自实际牌堆。
func (e *GameEngine) debugDrawExclusiveCardsFromStock(characterName, skillTitle string, count int) ([]model.Card, error) {
	if count <= 0 {
		return nil, nil
	}
	if characterName == "" || skillTitle == "" {
		return nil, fmt.Errorf("独有牌检索参数无效")
	}

	deckIndices := make([]int, 0, count)
	for i, c := range e.State.Deck {
		if c.MatchExclusive(characterName, skillTitle) {
			deckIndices = append(deckIndices, i)
			if len(deckIndices) >= count {
				break
			}
		}
	}

	remain := count - len(deckIndices)
	discardIndices := make([]int, 0, remain)
	if remain > 0 {
		for i, c := range e.State.DiscardPile {
			if c.MatchExclusive(characterName, skillTitle) {
				discardIndices = append(discardIndices, i)
				if len(discardIndices) >= remain {
					break
				}
			}
		}
	}

	if len(deckIndices)+len(discardIndices) < count {
		return nil, fmt.Errorf(
			"牌库/弃牌堆中独有牌不足：需要%d张 [%s·%s]，仅找到%d张",
			count, characterName, skillTitle, len(deckIndices)+len(discardIndices),
		)
	}

	picked := make([]model.Card, 0, count)
	for _, idx := range deckIndices {
		picked = append(picked, e.State.Deck[idx])
	}
	for _, idx := range discardIndices {
		picked = append(picked, e.State.DiscardPile[idx])
	}

	e.State.Deck = debugRemoveCardIndices(e.State.Deck, deckIndices)
	e.State.DiscardPile = debugRemoveCardIndices(e.State.DiscardPile, discardIndices)
	return picked, nil
}

func (e *GameEngine) debugEnsureExclusiveCard(player *model.Player, skill model.SkillDefinition, toHand bool) {
	if player == nil || player.Character == nil || skill.Title == "" {
		return
	}
	charName := player.Character.Name
	if player.HasExclusiveCard(charName, skill.Title) {
		return
	}
	cards, err := e.debugDrawExclusiveCardsFromStock(charName, skill.Title, 1)
	if err != nil || len(cards) == 0 {
		e.Log(fmt.Sprintf("[Cheat] 独有牌补齐失败 [%s·%s]: %v", charName, skill.Title, err))
		return
	}
	card := cards[0]
	if toHand {
		player.Hand = append(player.Hand, card)
	} else {
		player.ExclusiveCards = append(player.ExclusiveCards, card)
	}
}

func (e *GameEngine) debugAddExclusiveCopies(player *model.Player, skill model.SkillDefinition, count int) error {
	if player == nil || player.Character == nil || count <= 0 {
		return nil
	}
	cards, err := e.debugDrawExclusiveCardsFromStock(player.Character.Name, skill.Title, count)
	if err != nil {
		return err
	}
	player.Hand = append(player.Hand, cards...)
	return nil
}

func (e *GameEngine) debugAddCardCopies(player *model.Player, card *model.Card, count int) {
	if player == nil || card == nil || count <= 0 {
		return
	}
	for i := 0; i < count; i++ {
		newCard := *card
		newCard.ID = fmt.Sprintf("debug-%s-%d-%d", card.Name, time.Now().UnixNano(), i)
		player.Hand = append(player.Hand, newCard)
	}
}

func debugParseElement(raw string) (model.Element, error) {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case "water", "水", "水系":
		return model.ElementWater, nil
	case "fire", "火", "火系":
		return model.ElementFire, nil
	case "earth", "土", "地", "土系", "地系":
		return model.ElementEarth, nil
	case "wind", "风", "风系":
		return model.ElementWind, nil
	case "thunder", "雷", "雷系":
		return model.ElementThunder, nil
	case "light", "光", "光系":
		return model.ElementLight, nil
	case "dark", "暗", "暗系", "暗灭":
		return model.ElementDark, nil
	default:
		return "", fmt.Errorf("未知系别: %s", raw)
	}
}

// elementNameForPrompt 将系别值标准化为中文提示词（不带“系”后缀）。
// 例如 Fire -> 火, Thunder -> 雷, 水系 -> 水。
func elementNameForPrompt(raw string) string {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case "water", "水", "水系":
		return "水"
	case "fire", "火", "火系":
		return "火"
	case "earth", "土", "地", "土系", "地系":
		return "地"
	case "wind", "风", "风系":
		return "风"
	case "thunder", "雷", "雷系":
		return "雷"
	case "light", "光", "光系":
		return "光"
	case "dark", "暗", "暗系", "暗灭":
		return "暗"
	default:
		trimmed := strings.TrimSpace(raw)
		if strings.HasSuffix(trimmed, "系") {
			return strings.TrimSuffix(trimmed, "系")
		}
		return trimmed
	}
}

func debugNormalizeFaction(raw string) (string, error) {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case "圣", "holy":
		return "圣", nil
	case "血", "blood":
		return "血", nil
	case "幻", "phantom":
		return "幻", nil
	case "咏", "chant":
		return "咏", nil
	case "技", "technique":
		return "技", nil
	default:
		return "", fmt.Errorf("未知命格: %s", raw)
	}
}

func debugParseEffectType(raw string) (model.EffectType, error) {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case "shield", "圣盾":
		return model.EffectShield, nil
	case "poison", "中毒":
		return model.EffectPoison, nil
	case "weak", "虚弱":
		return model.EffectWeak, nil
	case "powerblessing", "power_blessing", "威力赐福":
		return model.EffectPowerBlessing, nil
	case "swiftblessing", "swift_blessing", "迅捷赐福":
		return model.EffectSwiftBlessing, nil
	case "sealfire", "seal_fire", "火之封印":
		return model.EffectSealFire, nil
	case "sealwater", "seal_water", "水之封印":
		return model.EffectSealWater, nil
	case "sealearth", "seal_earth", "地之封印":
		return model.EffectSealEarth, nil
	case "sealwind", "seal_wind", "风之封印":
		return model.EffectSealWind, nil
	case "sealthunder", "seal_thunder", "雷之封印":
		return model.EffectSealThunder, nil
	case "stealth", "潜行":
		return model.EffectStealth, nil
	case "mercy", "怜悯":
		return model.EffectMercy, nil
	default:
		return "", fmt.Errorf("未知效果: %s", raw)
	}
}

func debugEffectTrigger(effect model.EffectType) model.EffectTrigger {
	switch effect {
	case model.EffectPoison, model.EffectWeak:
		return model.EffectTriggerOnTurnStart
	case model.EffectShield:
		return model.EffectTriggerOnDamaged
	case model.EffectSealFire, model.EffectSealWater, model.EffectSealEarth, model.EffectSealWind, model.EffectSealThunder:
		return model.EffectTriggerOnAttack
	default:
		return model.EffectTriggerManual
	}
}

func (e *GameEngine) debugSetEffectCount(player *model.Player, effect model.EffectType, count int) {
	if player == nil {
		return
	}
	filtered := make([]*model.FieldCard, 0, len(player.Field))
	for _, fc := range player.Field {
		if fc == nil || fc.Mode != model.FieldEffect || fc.Effect != effect {
			filtered = append(filtered, fc)
		}
	}
	player.Field = filtered
	if count <= 0 {
		return
	}

	trigger := debugEffectTrigger(effect)
	for i := 0; i < count; i++ {
		card := model.Card{
			ID:          fmt.Sprintf("debug-effect-%s-%d-%d", effect, time.Now().UnixNano(), i),
			Name:        string(effect),
			Type:        model.CardTypeMagic,
			Element:     model.ElementLight,
			Damage:      0,
			Description: "调试效果牌",
		}
		player.AddFieldCard(&model.FieldCard{
			Card:     card,
			OwnerID:  player.ID,
			SourceID: player.ID,
			Mode:     model.FieldEffect,
			Effect:   effect,
			Trigger:  trigger,
			Duration: -1,
		})
	}
}

func debugFindCardsByFilter(filter func(model.Card) bool) []model.Card {
	deck := rules.InitDeck()
	templates := make([]model.Card, 0)
	seen := make(map[string]bool)
	for _, card := range deck {
		if !filter(card) {
			continue
		}
		key := fmt.Sprintf("%s|%s|%s|%s|%s", card.Name, card.Type, card.Element, card.Faction, card.Description)
		if seen[key] {
			continue
		}
		seen[key] = true
		templates = append(templates, card)
	}
	return templates
}

func (e *GameEngine) debugAddCardsFromTemplates(player *model.Player, templates []model.Card, count int) error {
	if player == nil {
		return fmt.Errorf("玩家不存在")
	}
	if count <= 0 {
		return fmt.Errorf("数量必须大于0")
	}
	if len(templates) == 0 {
		return fmt.Errorf("未找到满足条件的卡牌")
	}
	for i := 0; i < count; i++ {
		template := templates[i%len(templates)]
		newCard := template
		newCard.ID = fmt.Sprintf("debug-filter-%s-%d-%d", template.Name, time.Now().UnixNano(), i)
		player.Hand = append(player.Hand, newCard)
	}
	return nil
}

func (e *GameEngine) debugPrepareSkillResources(player *model.Player, skill model.SkillDefinition) {
	if player == nil {
		return
	}
	if skill.CostGem > 0 && player.Gem < skill.CostGem {
		player.Gem = skill.CostGem
	}
	if skill.CostCrystal > 0 && player.Crystal < skill.CostCrystal {
		player.Crystal = skill.CostCrystal
	}

	if player.TurnState.UsedSkillCounts == nil {
		player.TurnState.UsedSkillCounts = make(map[string]int)
	}
	player.TurnState.UsedSkillCounts[skill.ID] = 0
	player.TurnState.HasUsedTriggerSkill = false
}

func (e *GameEngine) debugPrepareSkillCards(player *model.Player, skill model.SkillDefinition) error {
	if player == nil {
		return nil
	}
	if skill.RequireExclusive {
		required := skill.CostDiscards
		if required <= 0 {
			required = 1
		}
		// 需要专属牌时，补足到手牌，便于弃置/展示
		if err := e.debugAddExclusiveCopies(player, skill, required); err != nil {
			return err
		}
		// 若弃牌就是专属牌，则不再额外补普通牌
		if skill.CostDiscards > 0 {
			return nil
		}
	}

	if skill.CostDiscards <= 0 {
		return nil
	}

	deck := rules.InitDeck()
	element := skill.DiscardElement
	if element == "" {
		element = model.ElementFire
	}
	card := debugFindCardTemplate(deck, element, model.CardTypeMagic)
	if card == nil {
		card = debugFindCardTemplate(deck, element, model.CardTypeAttack)
	}
	if card == nil {
		return nil
	}
	e.debugAddCardCopies(player, card, skill.CostDiscards)
	return nil
}

func (e *GameEngine) debugBuildContext(user *model.Player, skill model.SkillDefinition) *model.Context {
	if user == nil {
		return nil
	}
	target := e.debugPickOtherPlayer(user.ID)
	if target == nil {
		target = user
	}

	attacker := user
	defender := target
	if skill.RequiredRole == model.RoleDefender {
		attacker = target
		defender = user
	}

	actionType := model.ActionAttack
	cardType := model.CardTypeAttack
	element := model.ElementFire
	if skill.DiscardElement != "" {
		element = skill.DiscardElement
	}
	if skill.Trigger == model.TriggerOnCardUsed || skill.Trigger == model.TriggerOnCardRevealed {
		actionType = model.ActionMagic
		cardType = model.CardTypeMagic
		if skill.DiscardElement == "" {
			element = model.ElementWater
		}
	}

	deck := rules.InitDeck()
	card := debugFindCardTemplate(deck, element, cardType)
	if card == nil {
		card = &model.Card{
			ID:          fmt.Sprintf("debug-card-%s-%d", element, time.Now().UnixNano()),
			Name:        "调试卡",
			Type:        cardType,
			Element:     element,
			Damage:      1,
			Description: "调试卡牌",
		}
	}
	if skill.RequireExclusive || model.ContainsSkillTag(skill.Tags, model.TagUnique) {
		card.ExclusiveChar1 = user.Character.Name
		card.ExclusiveSkill1 = skill.Title
	}

	damageVal := 1
	drawCount := 1
	eventType := model.EventNone
	switch skill.Trigger {
	case model.TriggerOnAttackStart, model.TriggerOnAttackHit, model.TriggerOnAttackMiss:
		eventType = model.EventAttack
	case model.TriggerOnDamageTaken:
		eventType = model.EventDamage
	case model.TriggerOnCardUsed:
		eventType = model.EventCardUsed
	case model.TriggerBeforeDraw:
		eventType = model.EventBeforeDraw
	case model.TriggerAfterDraw:
		eventType = model.EventAfterDraw
	case model.TriggerOnTurnStart:
		eventType = model.EventTurnStart
	case model.TriggerOnPhaseEnd:
		eventType = model.EventPhaseEnd
	}

	attackInfo := &model.AttackEventInfo{
		IsHit:          true,
		IsHitForced:    false,
		Element:        string(element),
		CanBeResponded: true,
		ActionType:     string(actionType),
	}
	if skill.Trigger == model.TriggerOnAttackMiss || skill.Trigger == model.TriggerOnAttackStart {
		attackInfo.IsHit = false
	}

	ctx := &model.Context{
		Game:    e,
		User:    user,
		Target:  target,
		Trigger: skill.Trigger,
		TriggerCtx: &model.EventContext{
			Type:       eventType,
			SourceID:   attacker.ID,
			TargetID:   defender.ID,
			Card:       card,
			ActionType: actionType,
			DamageVal:  &damageVal,
			AttackInfo: attackInfo,
			DrawCount:  &drawCount,
		},
		Selections: map[string]any{},
		Flags:      map[string]bool{},
	}
	return ctx
}

// handleCheat 处理作弊指令 (用于测试)
func (e *GameEngine) handleCheat(act model.PlayerAction) error {
	// cheat <pid> <card_name> [count]
	// cheat turn <pid> (强制切换回合)
	targetStr := act.TargetID
	if targetStr == "turn" {
		// cheat turn <pid>
		if len(act.ExtraArgs) == 0 {
			return fmt.Errorf("未指定目标玩家ID")
		}
		targetPID := act.ExtraArgs[0]
		if err := e.forceTurnTo(targetPID); err != nil {
			return err
		}
		e.Log(fmt.Sprintf("[Cheat] 强制切换回合到 %s", e.State.Players[targetPID].Name))
		return nil
	}
	if targetStr == "role" {
		// cheat role <pid> <role_id>
		if len(act.ExtraArgs) < 2 {
			return fmt.Errorf("用法: cheat role <pid> <role_id>")
		}
		pid := act.ExtraArgs[0]
		roleID := act.ExtraArgs[1]
		player := e.State.Players[pid]
		if player == nil {
			return fmt.Errorf("玩家不存在: %s", pid)
		}
		char := e.debugFindCharacter(roleID)
		if char == nil {
			return fmt.Errorf("角色不存在: %s", roleID)
		}
		e.debugResetPlayerForRole(player, char)
		e.Log(fmt.Sprintf("[Cheat] %s 切换角色为 %s", player.Name, char.Name))
		return nil
	}
	if targetStr == "token" {
		// cheat token <pid> <token_key> <value>
		if len(act.ExtraArgs) < 3 {
			return fmt.Errorf("用法: cheat token <pid> <token_key> <value>")
		}
		pid := act.ExtraArgs[0]
		tokenKey := act.ExtraArgs[1]
		val, err := strconv.Atoi(act.ExtraArgs[2])
		if err != nil {
			return fmt.Errorf("token 值无效: %s", act.ExtraArgs[2])
		}
		player := e.State.Players[pid]
		if player == nil {
			return fmt.Errorf("玩家不存在: %s", pid)
		}
		if player.Tokens == nil {
			player.Tokens = map[string]int{}
		}
		player.Tokens[tokenKey] = val
		e.Log(fmt.Sprintf("[Cheat] %s 指示物 %s=%d", player.Name, tokenKey, val))
		return nil
	}
	if targetStr == "set" {
		// cheat set <pid> <field> <value>
		if len(act.ExtraArgs) < 3 {
			return fmt.Errorf("用法: cheat set <pid> <field> <value>")
		}
		pid := act.ExtraArgs[0]
		field := act.ExtraArgs[1]
		val, err := strconv.Atoi(act.ExtraArgs[2])
		if err != nil {
			return fmt.Errorf("数值无效: %s", act.ExtraArgs[2])
		}
		player := e.State.Players[pid]
		if player == nil {
			return fmt.Errorf("玩家不存在: %s", pid)
		}
		switch field {
		case "gem":
			player.Gem = val
		case "crystal":
			player.Crystal = val
		case "heal":
			player.Heal = val
		case "max_heal":
			player.MaxHeal = val
		default:
			return fmt.Errorf("未知字段: %s", field)
		}
		e.Log(fmt.Sprintf("[Cheat] %s 设置 %s=%d", player.Name, field, val))
		return nil
	}
	if targetStr == "effect" {
		// cheat effect <pid> <effect_type> <count>
		if len(act.ExtraArgs) < 3 {
			return fmt.Errorf("用法: cheat effect <pid> <effect_type> <count>")
		}
		pid := act.ExtraArgs[0]
		rawEffect := act.ExtraArgs[1]
		count, err := strconv.Atoi(act.ExtraArgs[2])
		if err != nil {
			return fmt.Errorf("效果数量无效: %s", act.ExtraArgs[2])
		}
		player := e.State.Players[pid]
		if player == nil {
			return fmt.Errorf("玩家不存在: %s", pid)
		}
		effectType, err := debugParseEffectType(rawEffect)
		if err != nil {
			return err
		}
		e.debugSetEffectCount(player, effectType, count)
		e.Log(fmt.Sprintf("[Cheat] %s 基础效果 %s 设置为 %d 层", player.Name, effectType, count))
		return nil
	}
	if targetStr == "card_exclusive" {
		// cheat card_exclusive <pid> <role_id> <skill_id> [count]
		if len(act.ExtraArgs) < 3 {
			return fmt.Errorf("用法: cheat card_exclusive <pid> <role_id> <skill_id> [count]")
		}
		pid := act.ExtraArgs[0]
		roleID := act.ExtraArgs[1]
		skillID := act.ExtraArgs[2]
		count := 1
		if len(act.ExtraArgs) > 3 {
			c, err := strconv.Atoi(act.ExtraArgs[3])
			if err != nil {
				return fmt.Errorf("数量无效: %s", act.ExtraArgs[3])
			}
			count = c
		}
		if count <= 0 {
			return fmt.Errorf("数量必须大于0")
		}
		player := e.State.Players[pid]
		if player == nil {
			return fmt.Errorf("玩家不存在: %s", pid)
		}
		char := e.debugFindCharacter(roleID)
		if char == nil {
			return fmt.Errorf("角色不存在: %s", roleID)
		}
		var skill *model.SkillDefinition
		for _, s := range char.Skills {
			if s.ID == skillID || s.Title == skillID {
				copySkill := s
				skill = &copySkill
				break
			}
		}
		if skill == nil {
			return fmt.Errorf("角色[%s]不存在该技能: %s", char.Name, skillID)
		}
		cards, err := e.debugDrawExclusiveCardsFromStock(char.Name, skill.Title, count)
		if err != nil {
			return err
		}
		player.Hand = append(player.Hand, cards...)
		e.Log(fmt.Sprintf("[Cheat] %s 获得 %d 张独有技手牌 [%s·%s]", player.Name, count, char.Name, skill.Title))
		return nil
	}
	if targetStr == "card_element" {
		// cheat card_element <pid> <element> [count]
		if len(act.ExtraArgs) < 2 {
			return fmt.Errorf("用法: cheat card_element <pid> <element> [count]")
		}
		pid := act.ExtraArgs[0]
		rawElement := act.ExtraArgs[1]
		count := 1
		if len(act.ExtraArgs) > 2 {
			c, err := strconv.Atoi(act.ExtraArgs[2])
			if err != nil {
				return fmt.Errorf("数量无效: %s", act.ExtraArgs[2])
			}
			count = c
		}
		player := e.State.Players[pid]
		if player == nil {
			return fmt.Errorf("玩家不存在: %s", pid)
		}
		element, err := debugParseElement(rawElement)
		if err != nil {
			return err
		}
		templates := debugFindCardsByFilter(func(card model.Card) bool {
			return card.Element == element && (card.Type == model.CardTypeAttack || card.Type == model.CardTypeMagic)
		})
		if err := e.debugAddCardsFromTemplates(player, templates, count); err != nil {
			return err
		}
		e.Log(fmt.Sprintf("[Cheat] %s 获得 %d 张%s手牌", player.Name, count, element))
		return nil
	}
	if targetStr == "card_faction" {
		// cheat card_faction <pid> <faction> [count]
		if len(act.ExtraArgs) < 2 {
			return fmt.Errorf("用法: cheat card_faction <pid> <faction> [count]")
		}
		pid := act.ExtraArgs[0]
		rawFaction := act.ExtraArgs[1]
		count := 1
		if len(act.ExtraArgs) > 2 {
			c, err := strconv.Atoi(act.ExtraArgs[2])
			if err != nil {
				return fmt.Errorf("数量无效: %s", act.ExtraArgs[2])
			}
			count = c
		}
		player := e.State.Players[pid]
		if player == nil {
			return fmt.Errorf("玩家不存在: %s", pid)
		}
		faction, err := debugNormalizeFaction(rawFaction)
		if err != nil {
			return err
		}
		templates := debugFindCardsByFilter(func(card model.Card) bool {
			return strings.TrimSpace(card.Faction) == faction && (card.Type == model.CardTypeAttack || card.Type == model.CardTypeMagic)
		})
		if err := e.debugAddCardsFromTemplates(player, templates, count); err != nil {
			return err
		}
		e.Log(fmt.Sprintf("[Cheat] %s 获得 %d 张%s命格手牌", player.Name, count, faction))
		return nil
	}
	if targetStr == "card_magic" {
		// cheat card_magic <pid> <card_name> [count]
		if len(act.ExtraArgs) < 2 {
			return fmt.Errorf("用法: cheat card_magic <pid> <card_name> [count]")
		}
		pid := act.ExtraArgs[0]
		cardName := strings.TrimSpace(act.ExtraArgs[1])
		count := 1
		if len(act.ExtraArgs) > 2 {
			c, err := strconv.Atoi(act.ExtraArgs[2])
			if err != nil {
				return fmt.Errorf("数量无效: %s", act.ExtraArgs[2])
			}
			count = c
		}
		player := e.State.Players[pid]
		if player == nil {
			return fmt.Errorf("玩家不存在: %s", pid)
		}
		if cardName == "" {
			return fmt.Errorf("法术牌名称不能为空")
		}
		templates := debugFindCardsByFilter(func(card model.Card) bool {
			return card.Type == model.CardTypeMagic && strings.TrimSpace(card.Name) == cardName
		})
		if err := e.debugAddCardsFromTemplates(player, templates, count); err != nil {
			return err
		}
		e.Log(fmt.Sprintf("[Cheat] %s 获得 %d 张法术牌 [%s]", player.Name, count, cardName))
		return nil
	}
	if targetStr == "skill" {
		// cheat skill <pid> [role_id] <skill_id>
		if len(act.ExtraArgs) < 2 {
			return fmt.Errorf("用法: cheat skill <pid> [role_id] <skill_id>")
		}
		pid := act.ExtraArgs[0]
		player := e.State.Players[pid]
		if player == nil {
			return fmt.Errorf("玩家不存在: %s", pid)
		}

		roleID := ""
		skillID := ""
		if len(act.ExtraArgs) == 2 {
			skillID = act.ExtraArgs[1]
		} else {
			roleID = act.ExtraArgs[1]
			skillID = act.ExtraArgs[2]
		}

		if roleID != "" {
			char := e.debugFindCharacter(roleID)
			if char == nil {
				return fmt.Errorf("角色不存在: %s", roleID)
			}
			e.debugResetPlayerForRole(player, char)
		}

		// 若当前不是该玩家回合，强制切换到该玩家
		if err := e.forceTurnTo(pid); err != nil {
			return err
		}
		e.Log(fmt.Sprintf("[Cheat] 强制切换回合到 %s", player.Name))

		// 清理中断/队列，避免旧流程影响调试
		e.State.PendingInterrupt = nil
		e.State.InterruptQueue = nil
		e.State.ActionQueue = []model.QueuedAction{}
		e.State.ActionStack = []model.Action{}
		e.State.CombatStack = []model.CombatRequest{}
		e.State.HasPerformedStartup = false
		player.TurnState = model.NewPlayerTurnState()

		skill, ok := e.debugFindSkill(player, skillID)
		if !ok {
			return fmt.Errorf("技能不存在: %s", skillID)
		}

		e.debugPrepareSkillResources(player, skill)
		if err := e.debugPrepareSkillCards(player, skill); err != nil {
			return err
		}

		// 行动技能：仅准备资源，手动在 UI 内发动
		if skill.Type == model.SkillTypeAction {
			e.Log(fmt.Sprintf("[Cheat] 已准备技能 %s（行动技），请在 UI 手动发动", skill.Title))
			return nil
		}

		// 启动/响应/被动：构造触发上下文，直接进入响应流程
		ctx := e.debugBuildContext(player, skill)
		if ctx == nil {
			return fmt.Errorf("无法构建技能上下文")
		}
		e.dispatcher.processSkills([]model.SkillDefinition{skill}, ctx)
		e.Log(fmt.Sprintf("[Cheat] 已触发调试技能 %s", skill.Title))
		return nil
	}

	pid := act.TargetID
	if pid == "" {
		return fmt.Errorf("未指定玩家ID")
	}
	player := e.State.Players[pid]
	if player == nil {
		return fmt.Errorf("玩家不存在: %s", pid)
	}

	if len(act.ExtraArgs) == 0 {
		return fmt.Errorf("未指定卡牌名称")
	}
	cardName := act.ExtraArgs[0]

	count := 1
	if len(act.ExtraArgs) > 1 {
		if c, err := strconv.Atoi(act.ExtraArgs[1]); err == nil {
			count = c
		}
	}

	// 查找卡牌模版 (简单遍历 Deck 或者构造)
	// 这里为了简单，直接从完整牌库中找一个同名的复制
	// 注意：这可能会产生 ID 重复的卡牌，但在简单测试中通常可以接受
	// 或者我们新建一个 Card

	// 为了更严谨，我们可以扫描整个 CardData
	// 但这里我们假设 e.State.Deck 里有所有类型的牌（初始洗牌后）
	// 或者我们预定义一些常见牌的属性

	var template *model.Card

	// 尝试从 rules.InitDeck() 获取一个新的牌库来查找模版
	tempDeck := rules.InitDeck()
	for _, c := range tempDeck {
		if c.Name == cardName {
			template = &c
			break
		}
	}

	if template == nil {
		return fmt.Errorf("未找到卡牌: %s", cardName)
	}

	for i := 0; i < count; i++ {
		newCard := *template
		newCard.ID = fmt.Sprintf("cheat-%s-%d-%d", cardName, time.Now().UnixNano(), i)
		player.Hand = append(player.Hand, newCard)
	}

	e.Log(fmt.Sprintf("[Cheat] 给 %s 添加了 %d 张 %s", player.Name, count, cardName))
	return nil
}

// executeSpecialAction 执行特殊行动（不结束回合）
func (e *GameEngine) executeSpecialAction(p *model.Player, actType model.ActionType) error {
	switch actType {
	case model.ActionBuy:
		return e.handleBuy(p)
	case model.ActionSynthesize:
		return e.handleSynthesize(p)
	case model.ActionExtract:
		return e.handleExtract(p)
	default:
		return fmt.Errorf("未知的特殊行动类型: %s", actType)
	}
}

// handleActionSelection 处理行动选择阶段的行动
func (e *GameEngine) handleActionSelection(act model.PlayerAction) error {
	currentPid := e.State.PlayerOrder[e.State.CurrentTurn]
	player := e.State.Players[currentPid]

	// 验证回合权
	if act.PlayerID != currentPid {
		return fmt.Errorf("不是你的回合")
	}

	if err := e.validateExtraActionConstraint(player, act); err != nil {
		return err
	}

	tauntSourceID := ""
	if tauntCard := getHeroTauntCard(player); tauntCard != nil {
		if src := e.State.Players[tauntCard.SourceID]; src != nil && src.Camp != player.Camp {
			tauntSourceID = src.ID
		} else {
			e.RemoveFieldCard(player.ID, model.EffectHeroTaunt)
		}
	}
	if tauntSourceID != "" {
		if act.Type != model.CmdAttack {
			e.Log(fmt.Sprintf("[Taunt] %s 未按挑衅要求发起攻击，跳过本次行动阶段", player.Name))
			e.RemoveFieldCard(player.ID, model.EffectHeroTaunt)
			e.State.Phase = model.PhaseTurnEnd
			return nil
		}
		targetID := act.TargetID
		if targetID == "" && len(act.TargetIDs) > 0 {
			targetID = act.TargetIDs[0]
		}
		if targetID != tauntSourceID {
			srcName := tauntSourceID
			if src := e.State.Players[tauntSourceID]; src != nil {
				srcName = model.GetPlayerDisplayName(src)
			}
			e.Log(fmt.Sprintf("[Taunt] %s 未攻击挑衅来源 %s，跳过本次行动阶段", player.Name, srcName))
			e.RemoveFieldCard(player.ID, model.EffectHeroTaunt)
			e.State.Phase = model.PhaseTurnEnd
			return nil
		}
		// 挑衅要求已触发（成功宣告攻击来源），立刻移除该效果。
		e.RemoveFieldCard(player.ID, model.EffectHeroTaunt)
	}

	if player.Tokens == nil {
		player.Tokens = map[string]int{}
	}
	if e.isFighter(player) && player.Tokens["fighter_hundred_dragon_form"] > 0 {
		if act.Type == model.CmdCannotAct {
			player.Tokens["fighter_hundred_dragon_form"] = 0
			player.Tokens["fighter_hundred_dragon_target_order"] = 0
			e.Log(fmt.Sprintf("%s 选择【无法行动】，取消 [百式幻龙拳] 并转正", player.Name))
		} else if act.Type != model.CmdAttack {
			player.Tokens["fighter_hundred_dragon_form"] = 0
			player.Tokens["fighter_hundred_dragon_target_order"] = 0
			e.Log(fmt.Sprintf("%s 未执行攻击，取消 [百式幻龙拳] 并转正", player.Name))
			return fmt.Errorf("百式幻龙拳状态下仅可执行攻击；已取消该状态，请重新选择行动")
		} else {
			targetID := act.TargetID
			if targetID == "" && len(act.TargetIDs) > 0 {
				targetID = act.TargetIDs[0]
			}
			if targetID == "" {
				return fmt.Errorf("百式幻龙拳状态下攻击必须指定目标")
			}
			targetPlayer := e.State.Players[targetID]
			if targetPlayer == nil {
				return fmt.Errorf("目标不存在")
			}
			if targetPlayer.Camp == player.Camp {
				return fmt.Errorf("攻击目标必须是敌方角色")
			}
			targetOrder := 0
			for i, pid := range e.State.PlayerOrder {
				if pid == targetID {
					targetOrder = i + 1
					break
				}
			}
			if targetOrder == 0 {
				return fmt.Errorf("目标不存在")
			}
			lockedOrder := player.Tokens["fighter_hundred_dragon_target_order"]
			if lockedOrder == 0 {
				player.Tokens["fighter_hundred_dragon_target_order"] = targetOrder
			} else if lockedOrder != targetOrder {
				player.Tokens["fighter_hundred_dragon_form"] = 0
				player.Tokens["fighter_hundred_dragon_target_order"] = 0
				e.Log(fmt.Sprintf("%s 攻击目标变化，取消 [百式幻龙拳] 并转正", player.Name))
				return fmt.Errorf("百式幻龙拳要求主动攻击同一目标；已取消该状态，请重新选择行动")
			}
		}
	}

	switch act.Type {
	case model.CmdBuy, model.CmdSynthesize, model.CmdExtract, model.CmdSkill:
		// 特殊行动：立即执行，然后进入 TurnEnd
		if e.State.HasPerformedStartup &&
			(act.Type == model.CmdBuy || act.Type == model.CmdSynthesize || act.Type == model.CmdExtract) {
			return fmt.Errorf("你本回合已执行启动技能，不能执行特殊行动")
		}

		var actionType model.ActionType
		switch act.Type {
		case model.CmdBuy:
			actionType = model.ActionBuy
		case model.CmdSynthesize:
			actionType = model.ActionSynthesize
		case model.CmdExtract:
			actionType = model.ActionExtract
		case model.CmdSkill:
			// 1. 基础校验
			if act.SkillID == "" {
				return fmt.Errorf("未指定技能ID")
			}

			// 2. 将 PlayerAction 中的 TargetIDs 和 Selections 直接传递给 UseSkill
			if err := e.UseSkill(act.PlayerID, act.SkillID, act.TargetIDs, act.Selections); err != nil {
				return fmt.Errorf("技能发动失败: %v", err)
			}
			skillTitle := act.SkillID
			if player.Character != nil {
				for _, s := range player.Character.Skills {
					if s.ID == act.SkillID {
						skillTitle = s.Title
						break
					}
				}
			}
			targets := []string{}
			if act.TargetID != "" {
				targets = append(targets, act.TargetID)
			}
			if len(act.TargetIDs) > 0 {
				targets = append(targets, act.TargetIDs...)
			}
			e.beginActionSummary("skill", player.ID, skillTitle, targets)
			// 3. 【关键】状态流转
			// 主动技能通常消耗一次行动机会。
			// 执行成功后，逻辑类似 Magic/Attack 结束，进入额外行动检查阶段
			if len(e.State.PendingDamageQueue) > 0 {
				e.State.Phase = model.PhasePendingDamageResolution
				e.State.ReturnPhase = model.PhaseExtraAction
			} else {
				e.State.Phase = model.PhaseExtraAction
			}

			// 注意：这里 return nil，由最外层的 HandleAction 去调用 Drive()
			return nil
		}

		// 执行特殊行动
		specialName := ""
		switch actionType {
		case model.ActionBuy:
			specialName = "购买"
		case model.ActionSynthesize:
			specialName = "合成"
		case model.ActionExtract:
			specialName = "提炼"
		}
		if specialName != "" {
			e.beginActionSummary("special", player.ID, specialName, nil)
		}
		if err := e.executeSpecialAction(player, actionType); err != nil {
			return err
		}
		if player.Tokens == nil {
			player.Tokens = map[string]int{}
		}
		player.Tokens["hb_special_used_turn"] = 1
		if e.isHolyBow(player) && player.Tokens["hb_form"] > 0 {
			player.Tokens["hb_form"] = 0
			e.Heal(player.ID, 1)
			e.Log(fmt.Sprintf("%s 在圣煌形态下执行特殊行动，脱离圣煌形态并获得1点治疗", player.Name))
		}
		player.TurnState.LastActionType = string(actionType)

		phaseEventCtx := &model.EventContext{
			Type:       model.EventPhaseEnd,
			SourceID:   player.ID,
			ActionType: actionType,
		}
		phaseCtx := e.buildContext(player, nil, model.TriggerOnPhaseEnd, phaseEventCtx)
		e.dispatcher.OnTrigger(model.TriggerOnPhaseEnd, phaseCtx)
		// 标记：本次特殊行动的 OnPhaseEnd 已在 ActionSelection 触发，
		// ExtraAction 阶段不应重复触发。
		if player.Tokens == nil {
			player.Tokens = map[string]int{}
		}
		player.Tokens["special_phase_end_dispatched"] = 1

		// 若购买触发了战绩区4星石选择（PendingInterrupt），不覆盖阶段，等待选择
		if e.State.PendingInterrupt == nil {
			e.State.Phase = model.PhaseExtraAction
		}
		return nil

	case model.CmdAttack, model.CmdMagic:
		// 普通行动：创建 QueuedAction 并推入队列
		if act.CardIndex < 0 {
			return fmt.Errorf("需要指定卡牌索引")
		}

		// 验证卡牌索引
		card, _, _, ok := getPlayableCardByIndex(player, act.CardIndex)
		if !ok {
			return fmt.Errorf("无效的卡牌索引")
		}

		// 验证卡牌类型
		if act.Type == model.CmdAttack && card.Type != model.CardTypeAttack {
			return fmt.Errorf("只能使用攻击牌进行攻击")
		}
		if act.Type == model.CmdMagic && card.Type != model.CardTypeMagic {
			return fmt.Errorf("只能使用法术牌进行法术")
		}
		if act.Type == model.CmdMagic && !e.canCastMagicInAction(player) {
			return fmt.Errorf("当前形态不能在行动阶段使用法术牌")
		}

		// 目标校验：
		// - 攻击必须指定目标
		// - 普通法术必须指定目标
		// - 仅“魔弹”允许不指定目标（由后端按传递顺序自动寻找目标）
		needTarget := act.Type == model.CmdAttack || (act.Type == model.CmdMagic && card.Name != "魔弹")
		if needTarget && act.TargetID == "" && len(act.TargetIDs) == 0 {
			if act.Type == model.CmdAttack {
				return fmt.Errorf("攻击需要指定目标")
			}
			return fmt.Errorf("该法术需要指定目标")
		}

		if act.TargetID != "" {
			if e.State.Players[act.TargetID] == nil {
				return fmt.Errorf("目标玩家 [%s] 不存在，请检查 ID", act.TargetID)
			}
		}
		if len(act.TargetIDs) > 0 {
			for _, tid := range act.TargetIDs {
				if e.State.Players[tid] == nil {
					return fmt.Errorf("目标玩家 [%s] 不存在，请检查 ID", tid)
				}
			}
		}
		if act.Type == model.CmdAttack {
			attackTargetID := act.TargetID
			if attackTargetID == "" && len(act.TargetIDs) > 0 {
				attackTargetID = act.TargetIDs[0]
			}
			if attackTargetID != "" {
				target := e.State.Players[attackTargetID]
				if target == nil {
					return fmt.Errorf("目标玩家 [%s] 不存在，请检查 ID", attackTargetID)
				}
				if target.Camp == player.Camp {
					return fmt.Errorf("攻击目标必须是敌方角色")
				}
				if target.HasFieldEffect(model.EffectStealth) {
					return fmt.Errorf("目标处于潜行状态，不能成为主动攻击目标")
				}
			}
		}

		// 创建 QueuedAction
		var actionType model.ActionType
		if act.Type == model.CmdAttack {
			actionType = model.ActionAttack
		} else if act.Type == model.CmdMagic {
			actionType = model.ActionMagic
		} else {
			return fmt.Errorf("无效的行动类型")
		}

		queuedAction := model.QueuedAction{
			SourceID:    currentPid,
			TargetID:    act.TargetID,
			TargetIDs:   act.TargetIDs,
			Type:        actionType,
			Element:     card.Element,
			Card:        &card,
			CardIndex:   act.CardIndex,
			SourceSkill: "", // 普通行动没有来源技能
		}
		if actionType == model.ActionAttack {
			card = e.applyBlazeWitchAttackCardTransform(player, card)
			queuedAction.Element = card.Element
			queuedAction.Card = &card
		}
		targets := []string{}
		if act.TargetID != "" {
			targets = append(targets, act.TargetID)
		}
		if len(act.TargetIDs) > 0 {
			targets = append(targets, act.TargetIDs...)
		}
		if actionType == model.ActionAttack {
			e.beginActionSummary("attack", player.ID, card.Name, targets)
		} else {
			e.beginActionSummary("magic", player.ID, card.Name, targets)
		}

		// 推入队列（或直接设置为当前行动）
		e.State.ActionQueue = append(e.State.ActionQueue, queuedAction)

		// 设置阶段为 BeforeAction
		e.State.Phase = model.PhaseBeforeAction
		return nil

	case model.CmdCannotAct:
		// 额外行动受限且无合法动作时，允许用“无法行动”主动跳过本次额外行动。
		if player.TurnState.CurrentExtraAction != "" {
			if e.checkExtraActionCards(player, player.TurnState.CurrentExtraAction, player.TurnState.CurrentExtraElement) {
				return errors.New("当前额外行动仍有可执行动作，不能跳过")
			}
			constraintInfo := e.buildConstraintInfo(player.TurnState.CurrentExtraAction, player.TurnState.CurrentExtraElement)
			e.beginActionSummary("cannot_act", player.ID, "跳过额外行动", nil)
			e.Log(fmt.Sprintf("[Turn] %s 宣告【无法行动】，跳过本次额外行动%s", player.Name, constraintInfo))
			player.TurnState.CurrentExtraAction = ""
			player.TurnState.CurrentExtraElement = nil
			e.State.Phase = model.PhaseTurnEnd
			return nil
		}

		// 常规阶段的“无法行动”：展示手牌、弃掉所有手牌、摸等量牌、本回合禁止特殊行动
		e.beginActionSummary("cannot_act", player.ID, "无法行动", nil)
		handCount := len(player.Hand)
		if handCount == 0 {
			// 无手牌时允许直接结束本回合行动阶段，避免在“已禁特殊行动”场景下无操作可做而卡死。
			e.Log(fmt.Sprintf("[Action] %s 宣告【无法行动】（无手牌），结束本回合行动阶段", player.Name))
			e.State.HasPerformedStartup = true
			e.State.Phase = model.PhaseTurnEnd
			return nil
		}
		// 再次校验：如果有攻击或法术牌，不允许宣告
		canUseMagic := e.canCastMagicInAction(player)
		for idx := 0; idx < playableCardCount(player); idx++ {
			c, _, _, ok := getPlayableCardByIndex(player, idx)
			if !ok {
				continue
			}
			if c.Type == model.CardTypeAttack || (c.Type == model.CardTypeMagic && canUseMagic) {
				return errors.New("你还有可用的攻击/法术牌，无法宣告无法行动")
			}
		}
		// 弃掉全部手牌
		e.Log(fmt.Sprintf("[Action] %s 宣告【无法行动】，展示并弃掉全部手牌(%d张)", player.Name, handCount))
		e.NotifyCardRevealed(player.ID, append([]model.Card{}, player.Hand...), "discard")
		for _, c := range player.Hand {
			e.State.DiscardPile = append(e.State.DiscardPile, c)
		}
		player.Hand = player.Hand[:0]
		// 摸等量的牌
		cards, newDeck, newDiscard := rules.DrawCards(e.State.Deck, e.State.DiscardPile, handCount)
		e.State.Deck = newDeck
		e.State.DiscardPile = newDiscard
		player.Hand = append(player.Hand, cards...)
		e.NotifyDrawCards(player.ID, handCount, "cannot_act_redraw")
		// 魔剑士备注：若重摸后仍全是法术牌，则展示弃掉并继续重摸，直到手牌中出现攻击牌。
		if e.isMagicSwordsman(player) {
			for len(player.Hand) > 0 {
				hasAttack := false
				allMagic := true
				for _, c := range player.Hand {
					if c.Type == model.CardTypeAttack {
						hasAttack = true
						break
					}
					if c.Type != model.CardTypeMagic {
						allMagic = false
					}
				}
				if hasAttack || !allMagic {
					break
				}
				redrawCount := len(player.Hand)
				e.NotifyCardRevealed(player.ID, append([]model.Card{}, player.Hand...), "discard")
				e.State.DiscardPile = append(e.State.DiscardPile, player.Hand...)
				player.Hand = player.Hand[:0]
				nextCards, deck2, discard2 := rules.DrawCards(e.State.Deck, e.State.DiscardPile, redrawCount)
				e.State.Deck = deck2
				e.State.DiscardPile = discard2
				player.Hand = append(player.Hand, nextCards...)
				e.NotifyDrawCards(player.ID, redrawCount, "magic_swordsman_redraw")
				e.Log(fmt.Sprintf("[Action] %s 触发魔剑士重摸：全法术手牌已弃置并重摸%d张", player.Name, redrawCount))
			}
		}
		e.Log(fmt.Sprintf("[Action] %s 重新摸了%d张牌，且本回合不可执行特殊行动", player.Name, handCount))
		// 标记已执行启动（禁止特殊行动）
		e.State.HasPerformedStartup = true
		// 重新进入行动选择
		e.State.Phase = model.PhaseActionSelection
		return nil

	default:
		return fmt.Errorf("无效的行动类型: %s", act.Type)
	}
	return nil
}

// 【新增辅助函数】校验额外行动约束
func (e *GameEngine) validateExtraActionConstraint(p *model.Player, act model.PlayerAction) error {
	// 1. 校验行动类型约束
	// 如果规定必须 Attack，但你用了 Magic 或 Buy
	if p.TurnState.CurrentExtraAction != "" {
		requiredType := p.TurnState.CurrentExtraAction

		// 受限额外行动下允许“无法行动”仅用于跳过：
		// 仅当当前确实不存在任何符合约束的可执行牌时生效。
		if act.Type == model.CmdCannotAct {
			if e.checkExtraActionCards(p, requiredType, p.TurnState.CurrentExtraElement) {
				return fmt.Errorf("当前额外行动仍有可执行动作，不能跳过")
			}
			return nil
		}

		// 将 Cmd 转换为 string 进行比较 (需要简单的映射逻辑)
		isMatch := false

		// 根据要求的类型进行匹配
		if requiredType == "Attack" {
			// 如果要求攻击：必须是 Attack 指令
			// 注意：这里故意不包含 CmdSkill，因此技能会被拦截
			if act.Type == model.CmdAttack {
				isMatch = true
			}
		} else if requiredType == "Magic" {
			// 如果要求法术：允许 Magic 指令，也允许主动技能（视为法术行动）
			if act.Type == model.CmdMagic || act.Type == model.CmdSkill {
				isMatch = true
			}
		}
		// 额外行动通常禁止特殊行动(Buy/Syn/Ext)，除非规则特殊说明

		if !isMatch {
			// 生成具体的错误提示
			if requiredType == "Attack" && act.Type == model.CmdSkill {
				return fmt.Errorf("当前额外行动必须是 [Attack]，不能使用技能")
			}
			return fmt.Errorf("当前额外行动必须是 [%s]", requiredType)
		}
	}

	// 2. 校验元素约束 (仅针对 Attack/Magic)
	// 如果规定必须用水系，但你用了火系
	if len(p.TurnState.CurrentExtraElement) > 0 && (act.Type == model.CmdAttack || act.Type == model.CmdMagic) {
		if card, _, _, ok := getPlayableCardByIndex(p, act.CardIndex); ok {
			if act.Type == model.CmdAttack {
				card = e.applyBlazeWitchAttackCardTransform(p, card)
			}
			isAllowed := false
			for _, allowed := range p.TurnState.CurrentExtraElement {
				if card.Element == allowed {
					isAllowed = true
					break
				}
			}
			if !isAllowed {
				var allowed []string
				for _, ele := range p.TurnState.CurrentExtraElement {
					if ele == "" {
						continue
					}
					allowed = append(allowed, fmt.Sprintf("%s系", elementNameForPrompt(string(ele))))
				}
				chosen := fmt.Sprintf("%s系", elementNameForPrompt(string(card.Element)))
				if len(allowed) == 0 {
					return fmt.Errorf("当前行动限制元素，你选择了 %s", chosen)
				}
				return fmt.Errorf("当前行动限制元素为 %s，你选择了 %s", strings.Join(allowed, " / "), chosen)
			}
		}
	}

	return nil
}

// checkExtraActionCards 检查玩家是否有符合额外行动约束的牌
func (e *GameEngine) checkExtraActionCards(p *model.Player, mustType string, mustElement []model.Element) bool {
	total := playableCardCount(p)
	for idx := 0; idx < total; idx++ {
		card, _, _, ok := getPlayableCardByIndex(p, idx)
		if !ok {
			continue
		}
		// 检查类型约束
		if mustType == "Attack" && card.Type != model.CardTypeAttack {
			continue
		}
		if mustType == "Magic" && card.Type != model.CardTypeMagic {
			continue
		}
		if mustType == "Magic" && !e.canCastMagicInAction(p) {
			continue
		}
		if mustType == "Attack" {
			card = e.applyBlazeWitchAttackCardTransform(p, card)
		}

		// 检查元素约束
		if len(mustElement) > 0 {
			elementMatch := false
			for _, elem := range mustElement {
				if card.Element == elem {
					elementMatch = true
					break
				}
			}
			if !elementMatch {
				continue
			}
		}

		// 找到符合条件的牌
		return true
	}
	// 额外法术行动允许发动主动技能（视为法术行动）。
	if mustType == "Magic" && e.hasUsableActionSkillForExtraMagic(p) {
		return true
	}
	return false
}

func (e *GameEngine) hasUsableActionSkillForExtraMagic(p *model.Player) bool {
	if p == nil || p.Character == nil {
		return false
	}

	for _, sd := range p.Character.Skills {
		if sd.Type != model.SkillTypeAction {
			continue
		}
		if !e.isActionSkillUsableForExtraMagic(p, sd) {
			continue
		}
		return true
	}
	return false
}

func (e *GameEngine) isActionSkillUsableForExtraMagic(p *model.Player, sd model.SkillDefinition) bool {
	// 回合限定：本回合已用过则不可再用。
	if model.ContainsSkillTag(sd.Tags, model.TagTurnLimit) && p.TurnState.UsedSkillCounts[sd.ID] > 0 {
		return false
	}
	// 资源校验（宝石/水晶）。
	if !canPaySkillEnergyCost(p, sd.CostGem, sd.CostCrystal) {
		return false
	}
	// 独有技：需拥有对应独有牌（手牌或专属卡区）。
	if sd.RequireExclusive && !p.HasExclusiveCard(p.Character.Name, sd.Title) {
		return false
	}
	// 弃牌成本可达成性。
	if !e.canSatisfyActionSkillDiscardRequirement(p, sd) {
		return false
	}
	// 目标可达成性（仅做最小目标数校验）。
	if !e.hasActionSkillValidTarget(p, sd) {
		return false
	}

	// 与前端/房间可用技能筛选保持一致的技能特例。
	switch sd.ID {
	case "ms_shadow_meteor":
		// 魔剑士【暗影流星】需要处于暗影形态，且至少可弃2张法术牌。
		if p.Tokens == nil || p.Tokens["ms_shadow_form"] <= 0 {
			return false
		}
		magicCnt := 0
		for _, c := range p.Hand {
			if c.Type == model.CardTypeMagic {
				magicCnt++
			}
		}
		if magicCnt < 2 {
			return false
		}
	case "adventurer_fraud":
		elemCount := map[model.Element]int{}
		for _, c := range p.Hand {
			elemCount[c.Element]++
		}
		canUseFraud := false
		for ele, n := range elemCount {
			if ele != "" && n >= 2 {
				canUseFraud = true
				break
			}
			if n >= 3 {
				canUseFraud = true
				break
			}
		}
		if !canUseFraud {
			return false
		}
	case "onmyoji_shikigami_descend":
		factionCount := map[string]int{}
		hasSameFactionPair := false
		for _, c := range p.Hand {
			if c.Faction == "" {
				continue
			}
			factionCount[c.Faction]++
			if factionCount[c.Faction] >= 2 {
				hasSameFactionPair = true
				break
			}
		}
		if !hasSameFactionPair {
			return false
		}
	case "mb_thunder_scatter":
		if p.TurnState.UsedSkillCounts["mb_charge_lock_turn"] > 0 {
			return false
		}
		if e.countCoverCardsByEffectAndElement(p, model.EffectMagicBowCharge, model.ElementThunder) <= 0 {
			return false
		}
	case "bd_dissonance_chord":
		inspiration := 0
		if p.Tokens != nil {
			inspiration = p.Tokens["bd_inspiration"]
		}
		if inspiration <= 1 {
			return false
		}
	case "elementalist_ignite":
		element := 0
		if p.Tokens != nil {
			element = p.Tokens["element"]
		}
		if element < 3 {
			return false
		}
	case "angel_cleanse":
		if !e.hasAnyBasicFieldEffectTarget() {
			return false
		}
	}

	return true
}

func (e *GameEngine) canSatisfyActionSkillDiscardRequirement(p *model.Player, sd model.SkillDefinition) bool {
	if sd.ID == "priest_water_power" {
		hasWater := false
		for _, card := range p.Hand {
			if card.Element == model.ElementWater {
				hasWater = true
				break
			}
		}
		if !hasWater {
			return false
		}
		required := 2
		if required > len(p.Hand) {
			required = len(p.Hand)
		}
		return len(p.Hand) >= required && required > 0
	}

	requiredDiscards := sd.CostDiscards
	// 神官-神圣领域：弃牌数量为“2 或当前全部手牌（当手牌<2）”。
	if sd.ID == "priest_divine_domain" && requiredDiscards > len(p.Hand) {
		requiredDiscards = len(p.Hand)
	}
	if requiredDiscards <= 0 {
		return true
	}

	matched := 0
	for _, card := range p.Hand {
		effectiveElement := card.Element
		if sd.DiscardElement != "" {
			effectiveElement = e.blazeWitchAttackElement(p, card)
		}
		if sd.DiscardElement != "" && effectiveElement != sd.DiscardElement {
			continue
		}
		if sd.ID == "magic_bullet_fusion" &&
			card.Element != model.ElementFire &&
			card.Element != model.ElementEarth {
			continue
		}
		if sd.DiscardType != "" && card.Type != sd.DiscardType {
			continue
		}
		if sd.DiscardFate != "" && card.Faction != sd.DiscardFate {
			continue
		}
		if sd.RequireExclusive && !card.MatchExclusive(p.Character.Name, sd.Title) {
			continue
		}
		matched++
		if matched >= requiredDiscards {
			return true
		}
	}
	return false
}

func (e *GameEngine) hasActionSkillValidTarget(p *model.Player, sd model.SkillDefinition) bool {
	switch sd.TargetType {
	case model.TargetNone, model.TargetSelf:
		return true
	}

	minTargets := sd.MinTargets
	if minTargets <= 0 {
		if sd.TargetType >= model.TargetEnemy {
			minTargets = 1
		} else {
			minTargets = 0
		}
	}
	if minTargets <= 0 {
		return true
	}

	candidates := 0
	for _, pid := range e.State.PlayerOrder {
		target := e.State.Players[pid]
		if target == nil {
			continue
		}
		switch sd.TargetType {
		case model.TargetEnemy:
			if target.Camp != p.Camp {
				candidates++
			}
		case model.TargetAlly:
			if target.Camp == p.Camp && target.ID != p.ID {
				candidates++
			}
		case model.TargetAllySelf:
			if target.Camp == p.Camp {
				candidates++
			}
		case model.TargetAny, model.TargetSpecific:
			candidates++
		}
		if candidates >= minTargets {
			return true
		}
	}
	return false
}

func (e *GameEngine) hasAnyBasicFieldEffectTarget() bool {
	isBasicEffect := func(effect model.EffectType) bool {
		switch effect {
		case model.EffectShield,
			model.EffectWeak,
			model.EffectPoison,
			model.EffectSealFire,
			model.EffectSealWater,
			model.EffectSealEarth,
			model.EffectSealWind,
			model.EffectSealThunder,
			model.EffectPowerBlessing,
			model.EffectSwiftBlessing:
			return true
		default:
			return false
		}
	}

	for _, p := range e.State.Players {
		if p == nil {
			continue
		}
		for _, fc := range p.Field {
			if fc.Mode == model.FieldEffect && isBasicEffect(fc.Effect) {
				return true
			}
		}
	}
	return false
}

func (e *GameEngine) countCoverCardsByEffectAndElement(p *model.Player, effect model.EffectType, element model.Element) int {
	if p == nil {
		return 0
	}
	count := 0
	for _, fc := range p.Field {
		if fc == nil || fc.Mode != model.FieldCover || fc.Effect != effect {
			continue
		}
		if element != "" && fc.Card.Element != element {
			continue
		}
		count++
	}
	return count
}

func queuedActionUsesVirtualCard(sourceSkill string) bool {
	switch sourceSkill {
	case "adventurer_fraud", "mb_multi_shot", "hb_holy_shard_storm":
		return true
	default:
		return false
	}
}

func (e *GameEngine) repairQueuedActionCard(player *model.Player, qa *model.QueuedAction) bool {
	if player == nil || qa == nil {
		return false
	}

	requiredType := model.CardType("")
	switch qa.Type {
	case model.ActionAttack:
		requiredType = model.CardTypeAttack
	case model.ActionMagic:
		requiredType = model.CardTypeMagic
	default:
		return false
	}

	// 优先使用当前索引（若仍然有效）。
	if card, _, _, ok := getPlayableCardByIndex(player, qa.CardIndex); ok {
		if card.Type == requiredType {
			if requiredType == model.CardTypeAttack {
				card = e.applyBlazeWitchAttackCardTransform(player, card)
			}
			cardCopy := card
			qa.Card = &cardCopy
			return true
		}
	}

	// 其次尝试按原卡 ID 对齐。
	if qa.Card != nil {
		if idx := findPlayableCardIndexByID(player, qa.Card.ID); idx >= 0 {
			if card, _, _, ok := getPlayableCardByIndex(player, idx); ok && card.Type == requiredType {
				if requiredType == model.CardTypeAttack {
					card = e.applyBlazeWitchAttackCardTransform(player, card)
				}
				qa.CardIndex = idx
				cardCopy := card
				qa.Card = &cardCopy
				return true
			}
		}
	}

	// 再按类型 + 元素约束寻找替代牌。
	total := playableCardCount(player)
	for idx := 0; idx < total; idx++ {
		card, _, _, ok := getPlayableCardByIndex(player, idx)
		if !ok {
			continue
		}
		if card.Type != requiredType {
			continue
		}
		if requiredType == model.CardTypeAttack {
			card = e.applyBlazeWitchAttackCardTransform(player, card)
		}
		if qa.Element != "" && card.Element != qa.Element {
			continue
		}
		qa.CardIndex = idx
		cardCopy := card
		qa.Card = &cardCopy
		return true
	}

	// 最后退化为任意同类型牌。
	for idx := 0; idx < total; idx++ {
		card, _, _, ok := getPlayableCardByIndex(player, idx)
		if !ok {
			continue
		}
		if card.Type == requiredType {
			qa.CardIndex = idx
			cardCopy := card
			qa.Card = &cardCopy
			return true
		}
	}

	return false
}

// buildConstraintInfo 构建约束信息字符串
func (e *GameEngine) buildConstraintInfo(mustType string, mustElement []model.Element) string {
	constraintInfo := ""
	if len(mustElement) > 0 {
		labels := make([]string, 0, len(mustElement))
		for _, ele := range mustElement {
			if ele == "" {
				continue
			}
			labels = append(labels, fmt.Sprintf("%s系", elementNameForPrompt(string(ele))))
		}
		if len(labels) > 0 {
			constraintInfo += fmt.Sprintf("[%s]", strings.Join(labels, "/"))
		}
	}
	if mustType != "" {
		constraintInfo += fmt.Sprintf("[%s行动]", mustType)
	}
	return constraintInfo
}

// HandleAction 核心路由器：处理所有 Action
func (e *GameEngine) HandleAction(act model.PlayerAction) error {
	e.Log(fmt.Sprintf("[Debug] HandleAction 收到指令: %s", act.Type))
	// === 1. 第一优先级：系统指令 (随时可执行) ===
	// 允许玩家在任何时候退出或查看帮助，哪怕是在选择弃牌的时候
	if act.Type == model.CmdQuit {
		// e.Notify(model.EventGameEnd, "玩家强制退出", nil)
		return fmt.Errorf("EXIT_GAME") // 或者特定的退出逻辑
	}
	if act.Type == model.CmdHelp {
		// 帮助信息通常由 CLI 直接处理，Engine 也可以返回特定的提示
		return nil
	}

	// 作弊指令 (Debug用)
	if act.Type == model.CmdCheat {
		if err := e.handleCheat(act); err != nil {
			return err
		}
		// 作弊成功后也驱动一次状态机，让回合和提示立即更新
		if e.State.PendingInterrupt == nil {
			e.Drive()
		}
		return nil
	}

	// === 2. 第二优先级：中断处理 (Interrupt) ===
	// 如果当前有挂起的中断，**必须** 先处理中断，禁止执行其他普通指令
	if e.State.PendingInterrupt != nil {
		// 处理中断输入
		err := e.handleInterruptAction(act)
		if err != nil {
			return err // 处理失败（如输入非法），直接返回错误，不驱动引擎
		}

		// 【关键】中断处理成功后，驱动状态机继续运行
		// 因为 handleInterruptAction 内部调用了 PopInterrupt，现在的状态可能已经变了
		e.Drive()
		return nil // 中断处理完直接返回，不要往下执行普通逻辑
	}

	// === 3. 第三优先级：游戏结束拦截 ===
	if e.State.Phase == model.PhaseEnd {
		return fmt.Errorf("游戏已结束")
	}

	// 3. 回合权校验
	currentPlayer := e.State.PlayerOrder[e.State.CurrentTurn]
	// 特殊情况：战斗响应阶段，允许目标玩家操作
	if e.State.Phase == model.PhaseCombatInteraction {
		// 在战斗响应逻辑内部校验目标ID，这里先放行
	} else {
		// 其他阶段，必须是当前回合玩家
		if act.PlayerID != currentPlayer && act.Type != model.CmdStart {
			return fmt.Errorf("不是你的回合")
		}
	}

	// 这里只调用逻辑处理函数，不要在这里调用 Drive
	var err error

	switch e.State.Phase {
	case model.PhaseActionSelection:
		// 行动选择阶段：处理攻击、法术、特殊行动
		err = e.handleActionSelection(act)

	case model.PhaseCombatInteraction:
		// 战斗交互阶段：处理响应 (take/defend/counter)
		if act.Type == model.CmdRespond {
			err = e.handleCombatResponse(act)
		} else {
			err = fmt.Errorf("当前必须响应战斗 (使用 take/defend/counter)")
		}

	// 以前的 Start 逻辑、Confirm 逻辑等，可以根据 Phase 归类
	// 如果 Start 只能在游戏未开始时用，可以在这里加一个 case model.PhaseInit
	default:
		// 处理一些尚未归类的全局指令（如 Start）
		if act.Type == model.CmdStart {
			err = e.StartGame()
		} else {
			err = fmt.Errorf("当前阶段 (%s) 不支持该指令", e.State.Phase)
		}
	}

	// === 6. 统一驱动 ===
	// 如果逻辑执行出错，直接返回错误，不驱动引擎
	if err != nil {
		return err
	}
	e.Log(fmt.Sprintf("[Debug] 指令执行成功，准备 Drive. Phase: %s, Interrupt: %v", e.State.Phase, e.State.PendingInterrupt))

	// 如果逻辑执行成功（err == nil），说明状态已经改变（ActionQueue加了东西，或者Phase变了）
	// 这时候踩一脚油门，让自动流程跑起来
	if e.State.PendingInterrupt == nil {
		e.Drive()
	} else {
		e.Log("[Debug] 存在挂起中断，暂不 Drive")
	}

	return nil
}

// handleInterruptAction 专门处理中断状态下的输入
func (e *GameEngine) handleInterruptAction(act model.PlayerAction) error {
	if act.PlayerID != e.State.PendingInterrupt.PlayerID {
		return fmt.Errorf("当前不是等待你的响应")
	}

	switch e.State.PendingInterrupt.Type {
	case model.InterruptResponseSkill:
		if e.prunePendingResponseSkills() {
			if p := e.State.Players[act.PlayerID]; p != nil && p.Tokens != nil && p.Tokens["adventurer_extract_requires_paradise"] > 0 {
				return fmt.Errorf("本次提炼结果需先发动[冒险者天堂]分配给队友")
			}
			e.clearAdventurerExtractState(e.State.Players[act.PlayerID])
			return e.SkipResponse()
		}
		forceParadise := e.isForcedAdventurerParadiseResponse(act.PlayerID)
		if act.Type == model.CmdCancel {
			if forceParadise {
				return fmt.Errorf("本次提炼结果需先发动[冒险者天堂]分配给队友")
			}
			e.clearAdventurerExtractState(e.State.Players[act.PlayerID])
			return e.SkipResponse()
		}
		if act.Type == model.CmdSelect {
			if len(act.Selections) != 1 {
				return fmt.Errorf("请选择一个选项")
			}
			idx := act.Selections[0]
			// 选项列表: [技能1, 技能2, ..., 跳过]
			// indices are 0-based from CLI
			if idx < 0 || idx > len(e.State.PendingInterrupt.SkillIDs) {
				return fmt.Errorf("无效的选择")
			}
			if idx == len(e.State.PendingInterrupt.SkillIDs) {
				if forceParadise {
					return fmt.Errorf("本次提炼结果需先发动[冒险者天堂]分配给队友")
				}
				e.clearAdventurerExtractState(e.State.Players[act.PlayerID])
				return e.SkipResponse()
			}
			skillID := e.State.PendingInterrupt.SkillIDs[idx]
			return e.ConfirmResponseSkill(act.PlayerID, skillID)
		}

	case model.InterruptStartupSkill:
		if act.Type == model.CmdCancel {
			return e.SkipStartupSkill(act.PlayerID)
		}
		if act.Type == model.CmdSelect {
			if len(act.Selections) != 1 {
				return fmt.Errorf("请选择一个选项")
			}
			idx := act.Selections[0]
			if idx < 0 || idx > len(e.State.PendingInterrupt.SkillIDs) {
				return fmt.Errorf("无效的选择")
			}
			if idx == len(e.State.PendingInterrupt.SkillIDs) {
				return e.SkipStartupSkill(act.PlayerID)
			}
			skillID := e.State.PendingInterrupt.SkillIDs[idx]
			return e.ConfirmStartupSkill(act.PlayerID, skillID)
		}

	case model.InterruptDiscard:
		if act.Type == model.CmdCancel {
			data, _ := e.State.PendingInterrupt.Context.(map[string]interface{})
			skillID, _ := data["skill_id"].(string)
			if skillID == "" {
				return fmt.Errorf("当前弃牌为强制操作，不能取消")
			}
			if skillID == "mb_charge_followup_discard" {
				return fmt.Errorf("【充能】弃牌为强制步骤，不能取消")
			}
			// 响应技的弃牌交互：复用 SkipResponse 恢复被中断的流程（如水影恢复摸牌）。
			if _, hasUserCtx := data["user_ctx"]; hasUserCtx {
				return e.SkipResponse()
			}

			// 主动技能的弃牌交互：回到行动选择阶段，允许重新选择行动。
			e.PopInterrupt()
			e.Log(fmt.Sprintf("[System] %s 取消了技能 [%s] 的弃牌发动", act.PlayerID, skillID))
			if e.State.PendingInterrupt == nil {
				e.State.Phase = model.PhaseActionSelection
			}
			return nil
		}
		if act.Type == model.CmdSelect {
			return e.ConfirmDiscard(act.PlayerID, act.Selections)
		}

	case model.InterruptGiveCards:
		if act.Type == model.CmdSelect {
			data, ok := e.State.PendingInterrupt.Context.(map[string]interface{})
			if !ok {
				return fmt.Errorf("给牌中断上下文错误")
			}
			receiverID, _ := data["receiver_id"].(string)
			return e.ConfirmGiveCards(act.PlayerID, receiverID, act.Selections)
		}

	case model.InterruptChoice:
		if act.Type == model.CmdCancel {
			if data, ok := e.State.PendingInterrupt.Context.(map[string]interface{}); ok {
				if ct, _ := data["choice_type"].(string); ct == "extract" {
					e.PopInterrupt()
					if e.State.PendingInterrupt == nil {
						e.State.Phase = model.PhaseActionSelection
					}
					if p := e.State.Players[act.PlayerID]; p != nil {
						e.Log(fmt.Sprintf("[System] %s 取消了提炼操作", p.Name))
					} else {
						e.Log(fmt.Sprintf("[System] %s 取消了提炼操作", act.PlayerID))
					}
					return nil
				} else if ct == "hom_dual_echo_target" {
					e.PopInterrupt()
					if p := e.State.Players[act.PlayerID]; p != nil {
						e.Log(fmt.Sprintf("[System] %s 取消了 [双重回响] 的目标选择", p.Name))
					} else {
						e.Log(fmt.Sprintf("[System] %s 取消了 [双重回响] 的目标选择", act.PlayerID))
					}
					if e.State.PendingInterrupt == nil && len(e.State.PendingDamageQueue) > 0 {
						e.State.Phase = model.PhasePendingDamageResolution
					}
					return nil
				} else if ct == "css_blood_barrier_counter_confirm" || ct == "css_blood_barrier_target" {
					e.PopInterrupt()
					if p := e.State.Players[act.PlayerID]; p != nil {
						e.Log(fmt.Sprintf("[System] %s 取消了 [血气屏障] 的追加效果", p.Name))
					} else {
						e.Log(fmt.Sprintf("[System] %s 取消了 [血气屏障] 的追加效果", act.PlayerID))
					}
					if e.State.PendingInterrupt == nil {
						if len(e.State.PendingDamageQueue) > 0 {
							e.State.Phase = model.PhasePendingDamageResolution
						} else {
							e.State.Phase = model.PhaseExtraAction
						}
					}
					return nil
				}
			}
		}
		if act.Type == model.CmdSelect {
			if data, ok := e.State.PendingInterrupt.Context.(map[string]interface{}); ok {
				if ct, _ := data["choice_type"].(string); ct == "extract" {
					return e.handleExtractChoiceResponse(act)
				}
			}
			if len(act.Selections) != 1 {
				return fmt.Errorf("请选择一个选项")
			}
			idx := act.Selections[0]
			return e.handleWeakChoiceInput(act.PlayerID, idx)
		}

	case model.InterruptMagicMissile:
		// 支持 CmdRespond (take/defend/counter)
		if act.Type == model.CmdRespond {
			return e.handleMagicMissileResponse(act)
		}

	case model.InterruptMagicBulletFusion:
		// 魔弹融合询问：地系/火系牌当魔弹使用
		if act.Type == model.CmdSelect {
			return e.handleMagicBulletFusionResponse(act)
		}

	case model.InterruptMagicBulletDirection:
		// 魔弹掌控询问：选择传递方向
		if act.Type == model.CmdSelect {
			return e.handleMagicBulletDirectionResponse(act)
		}

	case model.InterruptHolySwordDraw:
		// 圣剑摸X弃X
		if act.Type == model.CmdSelect {
			return e.handleHolySwordDrawResponse(act)
		}

	case model.InterruptSaintHeal:
		// 圣疗分配治疗
		if act.Type == model.CmdSelect {
			return e.handleSaintHealResponse(act)
		}

	case model.InterruptMagicBlast:
		// 魔爆冲击弃牌选择
		if act.Type == model.CmdSelect || act.Type == model.CmdCancel {
			return e.handleMagicBlastResponse(act)
		}
	}

	return fmt.Errorf("当前中断类型不支持该指令")
}

func (e *GameEngine) consumeShieldForMagicMissileTake(target *model.Player, chain *model.MagicBulletChain) bool {
	if target == nil || chain == nil || !target.HasFieldEffect(model.EffectShield) {
		return false
	}
	removed := false
	for _, fc := range target.Field {
		if fc == nil || fc.Mode != model.FieldEffect || fc.Effect != model.EffectShield {
			continue
		}
		target.RemoveFieldCard(fc)
		e.State.DiscardPile = append(e.State.DiscardPile, fc.Card)
		removed = true
		break
	}
	if !removed {
		return false
	}

	e.addActionResponse(fmt.Sprintf("%s 的【圣盾】自动抵挡魔弹", target.Name))
	e.NotifyActionStep(fmt.Sprintf("%s 的【圣盾】触发，自动抵挡了魔弹", target.Name))
	e.Log(fmt.Sprintf("[Magic] %s 选择承受，触发【圣盾】自动抵挡魔弹", target.Name))
	e.State.MagicBulletChain = nil
	e.PopInterrupt()
	return true
}

// handleMagicMissileResponse 处理魔弹响应
func (e *GameEngine) handleMagicMissileResponse(act model.PlayerAction) error {
	chain := e.State.MagicBulletChain
	if chain == nil {
		return fmt.Errorf("魔弹链条不存在")
	}

	if act.PlayerID != chain.TargetID {
		return fmt.Errorf("不是你的响应回合")
	}

	respType := ""
	if len(act.ExtraArgs) > 0 {
		respType = act.ExtraArgs[0]
	} else {
		return fmt.Errorf("缺少响应类型")
	}

	player := e.State.Players[act.PlayerID]

	switch respType {
	case "take":
		// 承受伤害（若有场上圣盾，则此处触发抵挡）
		if e.consumeShieldForMagicMissileTake(player, chain) {
			return nil
		}

		damage := chain.CurrentDamage
		e.Log(fmt.Sprintf("[Magic] %s 选择承受魔弹伤害 (%d点)", player.Name, damage))

		// 构造临时卡牌用于伤害结算
		magicCard := &model.Card{
			Name:        "魔弹",
			Type:        model.CardTypeMagic,
			Damage:      damage,
			Description: "魔弹伤害",
		}

		// 优先弹出中断，否则 ResolveDamage 会检测到当前还存在中断而暂停执行
		e.PopInterrupt()

		// 使用 AddPendingDamage 代替直接调用 ResolveDamage，以支持中断和恢复
		e.AddPendingDamage(model.PendingDamage{
			SourceID:   chain.SourcePlayerID,
			TargetID:   player.ID,
			Damage:     damage,
			DamageType: "magic",
			Card:       magicCard,
			Stage:      0,
		})

		// 设置阶段为延迟伤害结算
		e.State.Phase = model.PhasePendingDamageResolution
		// 魔弹结算后，通常本回合法术结束，进入额外行动(触发PhaseEnd)或直接TurnEnd
		e.State.ReturnPhase = model.PhaseExtraAction

		// 魔弹结束
		e.State.MagicBulletChain = nil
		return nil

	case "counter":
		// 传递 (需打出魔弹)
		card, _, _, ok := getPlayableCardByIndex(player, act.CardIndex)
		if !ok {
			return fmt.Errorf("无效的卡牌索引")
		}
		if e.isMagicLancer(player) {
			return fmt.Errorf("魔枪受[黑暗束缚]影响，不能使用法术牌")
		}
		if card.Name != "魔弹" {
			return fmt.Errorf("必须使用【魔弹】进行传递")
		}

		// 检查是否已参与过
		hasParticipated := false
		for _, pid := range chain.InvolvedIDs {
			if pid == player.ID {
				hasParticipated = true
				break
			}
		}

		// 计算参与本轮传递的玩家数量（当前简化为所有在座玩家）
		aliveCount := len(e.State.PlayerOrder)

		if hasParticipated {
			return fmt.Errorf("你在本轮传递中已参与过，无法再次传递")
		}

		// 消耗卡牌
		if _, err := consumePlayableCardByIndex(player, act.CardIndex); err != nil {
			return err
		}
		e.State.DiscardPile = append(e.State.DiscardPile, card)

		e.Log(fmt.Sprintf("[Magic] %s 打出魔弹，将伤害传递给下一位！伤害+1", player.Name))

		// 更新链条
		chain.CurrentDamage += 1
		chain.SourcePlayerID = player.ID
		chain.InvolvedIDs = append(chain.InvolvedIDs, player.ID)

		// 当本轮传递已覆盖全员时，魔弹链条直接结束，不再开启下一轮。
		if len(chain.InvolvedIDs) >= aliveCount {
			e.Log("[Magic] 本轮魔弹传递已覆盖所有角色，魔弹结算结束")
			e.State.MagicBulletChain = nil
			e.PopInterrupt()
			return nil
		}

		// 寻找下一个目标
		nextTargetID := e.findNextMagicBulletTarget(player.ID)
		if nextTargetID == "" {
			e.Log("[Magic] 没有下一个目标，魔弹失效")
			e.State.MagicBulletChain = nil
			e.PopInterrupt()
			return nil
		}

		nextTarget := e.State.Players[nextTargetID]

		chain.TargetID = nextTargetID

		// 更新中断
		e.State.PendingInterrupt.PlayerID = nextTargetID
		if ctx, ok := e.State.PendingInterrupt.Context.(map[string]interface{}); ok {
			ctx["damage"] = chain.CurrentDamage
			ctx["source_id"] = player.ID
		}

		// 通知新的响应者
		e.notifyInterruptPrompt()

		e.Log(fmt.Sprintf("[Magic] 魔弹指向 %s (伤害: %d)，等待响应...",
			nextTarget.Name, chain.CurrentDamage))

		return nil

	case "defend":
		// 抵挡：仅允许打出【圣光】；【圣盾】必须提前放置并在被指向时自动触发。
		if e.isMagicLancer(player) {
			return fmt.Errorf("魔枪受[黑暗束缚]影响，不能使用法术牌防御")
		}
		if card, _, _, ok := getPlayableCardByIndex(player, act.CardIndex); ok {
			if card.Name == "圣盾" {
				return fmt.Errorf("【圣盾】不能在防御时打出，请提前放置到场上触发")
			}
			if card.Name != "圣光" {
				return fmt.Errorf("必须使用【圣光】抵挡")
			}
			e.Log(fmt.Sprintf("[Magic] %s 使用【圣光】，抵挡了魔弹", player.Name))
			if _, err := consumePlayableCardByIndex(player, act.CardIndex); err != nil {
				return err
			}
			e.State.DiscardPile = append(e.State.DiscardPile, card)
		} else {
			holyIdx := -1
			for i := 0; i < playableCardCount(player); i++ {
				c, _, _, ok := getPlayableCardByIndex(player, i)
				if !ok {
					continue
				}
				if c.Name == "圣光" {
					holyIdx = i
					break
				}
			}
			if holyIdx < 0 {
				return fmt.Errorf("没有【圣光】可以抵挡（若有场上【圣盾】，可选择承受伤害来自动触发）")
			}
			card, _, _, _ := getPlayableCardByIndex(player, holyIdx)
			e.Log(fmt.Sprintf("[Magic] %s 使用【圣光】，抵挡了魔弹", player.Name))
			if _, err := consumePlayableCardByIndex(player, holyIdx); err != nil {
				return err
			}
			e.State.DiscardPile = append(e.State.DiscardPile, card)
		}

		// 抵挡成功，魔弹结束
		e.State.MagicBulletChain = nil
		e.PopInterrupt()
		return nil

	default:
		return fmt.Errorf("未知的响应类型: %s", respType)
	}
}

// buildMagicBulletFusionPrompt 构建魔弹融合询问提示
func (e *GameEngine) buildMagicBulletFusionPrompt() *model.Prompt {
	interrupt := e.State.PendingInterrupt
	if interrupt == nil {
		return nil
	}
	playerID := interrupt.PlayerID
	player := e.State.Players[playerID]
	if player == nil {
		return nil
	}

	data, ok := interrupt.Context.(map[string]interface{})
	if !ok {
		return nil
	}
	cardIdx, _ := data["card_idx"].(int)
	card, _, _, cardOK := getPlayableCardByIndex(player, cardIdx)
	if !cardOK {
		return nil
	}

	return &model.Prompt{
		Type:     model.PromptConfirm,
		PlayerID: playerID,
		Message:  fmt.Sprintf("【魔弹融合】是否将 %s (%s系) 当魔弹使用？", card.Name, elementNameForPrompt(string(card.Element))),
		Options: []model.PromptOption{
			{ID: "yes", Label: "是 - 当魔弹使用"},
			{ID: "no", Label: "否 - 正常使用"},
		},
		Min: 1,
		Max: 1,
	}
}

// buildMagicBulletDirectionPrompt 构建魔弹掌控询问提示
func (e *GameEngine) buildMagicBulletDirectionPrompt() *model.Prompt {
	interrupt := e.State.PendingInterrupt
	if interrupt == nil {
		return nil
	}
	playerID := interrupt.PlayerID

	return &model.Prompt{
		Type:     model.PromptConfirm,
		PlayerID: playerID,
		Message:  "【魔弹掌控】选择魔弹传递方向：",
		Options: []model.PromptOption{
			{ID: "normal", Label: "默认方向 (右手边，前一位对手)"},
			{ID: "reverse", Label: "逆向传递 (左手边，后一位对手)"},
		},
		Min: 1,
		Max: 1,
	}
}

// buildHolySwordDrawPrompt 构建圣剑摸X弃X提示
func (e *GameEngine) buildHolySwordDrawPrompt() *model.Prompt {
	interrupt := e.State.PendingInterrupt
	if interrupt == nil {
		return nil
	}
	playerID := interrupt.PlayerID

	return &model.Prompt{
		Type:     model.PromptConfirm,
		PlayerID: playerID,
		Message:  "【圣剑】第3次攻击结束！选择摸X张牌然后弃X张牌 (X=0-3)：",
		Options: []model.PromptOption{
			{ID: "0", Label: "X=0 (不摸不弃)"},
			{ID: "1", Label: "X=1 (摸1弃1)"},
			{ID: "2", Label: "X=2 (摸2弃2)"},
			{ID: "3", Label: "X=3 (摸3弃3)"},
		},
		Min: 1,
		Max: 1,
	}
}

// buildSaintHealPrompt 构建圣疗分配治疗提示
func (e *GameEngine) buildSaintHealPrompt() *model.Prompt {
	interrupt := e.State.PendingInterrupt
	if interrupt == nil {
		return nil
	}
	playerID := interrupt.PlayerID

	// 获取所有玩家作为可选目标
	var options []model.PromptOption
	for _, p := range e.State.Players {
		options = append(options, model.PromptOption{
			ID:    p.ID,
			Label: p.Name,
		})
	}

	return &model.Prompt{
		Type:     model.PromptConfirm,
		PlayerID: playerID,
		Message:  "【圣疗】选择1-3名角色分配共3点治疗。先选择目标，再分配点数：",
		Options:  options,
		Min:      1,
		Max:      3,
	}
}

// buildMagicBlastPrompt 构建魔爆冲击弃牌提示
func (e *GameEngine) buildMagicBlastPrompt() *model.Prompt {
	interrupt := e.State.PendingInterrupt
	if interrupt == nil {
		return nil
	}

	playerID := interrupt.PlayerID
	player := e.State.Players[playerID]
	if player == nil {
		return nil
	}
	data, _ := interrupt.Context.(map[string]interface{})
	stage, _ := data["stage"].(string)
	if stage == "" {
		stage = "target_discard"
	}

	// 施法者可选弃1张任意牌（或取消跳过）
	if stage == "caster_optional_discard" {
		var options []model.PromptOption
		for i, card := range player.Hand {
			options = append(options, model.PromptOption{
				ID:    strconv.Itoa(i),
				Label: fmt.Sprintf("%d: %s", i+1, formatCardInfo(card)),
			})
		}
		return &model.Prompt{
			Type:     model.PromptChooseCards,
			PlayerID: playerID,
			Message:  "【魔爆冲击】你可选择弃1张牌（任意类型），或取消跳过：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}

	// 目标阶段：收集法术牌选项
	var options []model.PromptOption
	for i, card := range player.Hand {
		if card.Type == model.CardTypeMagic {
			options = append(options, model.PromptOption{
				ID:    strconv.Itoa(i),
				Label: fmt.Sprintf("%d: %s", i+1, formatCardInfo(card)),
			})
		}
	}

	// 添加"不弃牌"选项
	options = append(options, model.PromptOption{
		ID:    "refuse",
		Label: "不弃牌 (受到2点伤害)",
	})

	return &model.Prompt{
		Type:     model.PromptChooseCards,
		PlayerID: playerID,
		Message:  "【魔爆冲击】请选择弃一张法术牌，否则受到2点伤害：",
		Options:  options,
		Min:      1,
		Max:      1,
	}
}

// handleMagicBlastResponse 处理魔爆冲击弃牌响应
func (e *GameEngine) handleMagicBlastResponse(act model.PlayerAction) error {
	interrupt := e.State.PendingInterrupt
	if interrupt == nil {
		return fmt.Errorf("没有待处理的中断")
	}

	data, ok := interrupt.Context.(map[string]interface{})
	if !ok {
		return fmt.Errorf("中断上下文格式错误")
	}
	stage, _ := data["stage"].(string)
	if stage == "" {
		stage = "target_discard"
	}

	player := e.State.Players[act.PlayerID]
	if player == nil {
		return fmt.Errorf("玩家不存在")
	}

	casterID, _ := data["caster_id"].(string)
	caster := e.State.Players[casterID]

	targetsRaw, _ := data["targets"].([]string)
	// 处理 JSON 反序列化后可能的类型
	if targetsRaw == nil {
		if targetsIface, ok := data["targets"].([]interface{}); ok {
			targetsRaw = make([]string, len(targetsIface))
			for i, v := range targetsIface {
				targetsRaw[i], _ = v.(string)
			}
		}
	}

	currentTargetIdx := 0
	if ct, ok := data["current_target"].(int); ok {
		currentTargetIdx = ct
	} else if ctf, ok := data["current_target"].(float64); ok {
		currentTargetIdx = int(ctf)
	}

	failedCount := 0
	if fc, ok := data["failed_count"].(int); ok {
		failedCount = fc
	} else if fcf, ok := data["failed_count"].(float64); ok {
		failedCount = int(fcf)
	}

	// 阶段2：施法者可选弃1张牌（任意类型）
	if stage == "caster_optional_discard" {
		if act.Type == model.CmdCancel {
			e.Log(fmt.Sprintf("[Skill] %s 选择不弃牌", player.Name))
			e.PopInterrupt()
			return nil
		}
		if act.Type != model.CmdSelect || len(act.Selections) == 0 {
			return fmt.Errorf("请选择1张牌，或取消跳过")
		}

		selection := act.Selections[0]
		if selection < 0 || selection >= len(player.Hand) {
			return fmt.Errorf("无效的卡牌索引")
		}
		card := player.Hand[selection]
		e.NotifyCardRevealed(player.ID, []model.Card{card}, "discard")
		player.Hand = append(player.Hand[:selection], player.Hand[selection+1:]...)
		e.State.DiscardPile = append(e.State.DiscardPile, card)
		e.Log(fmt.Sprintf("[Skill] %s 选择弃掉了 %s", player.Name, card.Name))
		e.PopInterrupt()
		return nil
	}

	// 阶段1：目标玩家选择弃法术牌或受伤
	discarded := false
	if act.Type == model.CmdCancel {
		// 拒绝弃牌
		discarded = false
	} else if len(act.Selections) > 0 {
		selection := act.Selections[0]
		// 检查是否是有效的牌索引
		if selection >= 0 && selection < len(player.Hand) {
			card := player.Hand[selection]
			if card.Type == model.CardTypeMagic {
				// 弃掉法术牌
				player.Hand = append(player.Hand[:selection], player.Hand[selection+1:]...)
				e.State.DiscardPile = append(e.State.DiscardPile, card)
				e.Log(fmt.Sprintf("[Skill] %s 弃掉了法术牌 %s", player.Name, card.Name))
				discarded = true
			}
		}
	}

	if !discarded {
		// 未弃牌，受到2点伤害
		e.InflictDamage(casterID, player.ID, 2, "magic")
		e.Log(fmt.Sprintf("[Skill] %s 未弃法术牌，受到2点伤害", player.Name))
		failedCount++
	}

	// 检查是否还有下一个目标
	currentTargetIdx++
	if currentTargetIdx < len(targetsRaw) {
		// 还有下一个目标，更新中断
		data["current_target"] = currentTargetIdx
		data["failed_count"] = failedCount
		e.State.PendingInterrupt.PlayerID = targetsRaw[currentTargetIdx]
		e.State.PendingInterrupt.Context = data

		nextTarget := e.State.Players[targetsRaw[currentTargetIdx]]
		if nextTarget != nil {
			e.Log(fmt.Sprintf("[Skill] %s 需要选择弃一张法术牌或受到2点伤害", nextTarget.Name))
		}
		e.notifyInterruptPrompt()
		return nil
	}

	// 所有目标处理完毕 -> 切换到施法者可选弃牌阶段
	if caster != nil {
		data["stage"] = "caster_optional_discard"
		data["failed_count"] = failedCount
		e.State.PendingInterrupt.PlayerID = caster.ID
		e.State.PendingInterrupt.Context = data
		e.Log(fmt.Sprintf("[Skill] 目标处理完毕：未弃牌人数=%d，%s 可选择弃1张牌或跳过", failedCount, caster.Name))
		e.notifyInterruptPrompt()
		return nil
	}

	e.PopInterrupt()

	return nil
}

// handleMagicBulletFusionResponse 处理魔弹融合响应
func (e *GameEngine) handleMagicBulletFusionResponse(act model.PlayerAction) error {
	interrupt := e.State.PendingInterrupt
	if interrupt == nil {
		return fmt.Errorf("没有待处理的中断")
	}

	data, ok := interrupt.Context.(map[string]interface{})
	if !ok {
		return fmt.Errorf("中断上下文格式错误")
	}

	player := e.State.Players[act.PlayerID]
	if player == nil {
		return fmt.Errorf("玩家不存在")
	}

	cardIdx, _ := data["card_idx"].(int)
	targetID, _ := data["target_id"].(string)

	card, _, _, cardOK := getPlayableCardByIndex(player, cardIdx)
	if !cardOK {
		return fmt.Errorf("无效的卡牌索引")
	}

	// 选项索引：0=是(当魔弹)，1=否(正常使用)
	choice := 1 // 默认否
	if len(act.Selections) > 0 {
		choice = act.Selections[0]
	}

	// 弹出当前中断
	e.PopInterrupt()

	if choice == 0 {
		// 选择当魔弹使用
		e.Log(fmt.Sprintf("[Skill] %s 发动【魔弹融合】，将 %s 当魔弹使用！", player.Name, card.Name))

		// 从可打出牌区移除卡牌
		if _, err := consumePlayableCardByIndex(player, cardIdx); err != nil {
			return err
		}
		e.State.DiscardPile = append(e.State.DiscardPile, card)

		// 继续询问是否逆向传递（魔弹掌控）
		e.PushInterrupt(&model.Interrupt{
			Type:     model.InterruptMagicBulletDirection,
			PlayerID: player.ID,
			Context: map[string]interface{}{
				"source_id":   player.ID,
				"is_fusion":   true,
				"fusion_card": card,
			},
		})
		return nil
	}

	// 选择正常使用，继续原来的法术逻辑
	e.Log(fmt.Sprintf("[Magic] %s 选择正常使用 %s", player.Name, card.Name))

	// 重新调用 PerformMagic，但需要跳过融合检查
	// 这里直接执行原始法术效果
	player.TurnState.SkipFusionCheck = true
	err := e.PerformMagic(act.PlayerID, targetID, cardIdx)
	player.TurnState.SkipFusionCheck = false
	return err
}

// handleMagicBulletDirectionResponse 处理魔弹掌控响应
func (e *GameEngine) handleMagicBulletDirectionResponse(act model.PlayerAction) error {
	interrupt := e.State.PendingInterrupt
	if interrupt == nil {
		return fmt.Errorf("没有待处理的中断")
	}

	data, ok := interrupt.Context.(map[string]interface{})
	if !ok {
		return fmt.Errorf("中断上下文格式错误")
	}

	player := e.State.Players[act.PlayerID]
	if player == nil {
		return fmt.Errorf("玩家不存在")
	}

	// 选项索引：0=顺时针，1=逆时针
	reverse := false
	if len(act.Selections) > 0 && act.Selections[0] == 1 {
		reverse = true
	}

	// 检查是否是融合触发的
	isFusion, _ := data["is_fusion"].(bool)
	var fusionCard *model.Card
	if isFusion {
		if fc, ok := data["fusion_card"].(model.Card); ok {
			fusionCard = &fc
		}
	}

	// 弹出当前中断
	e.PopInterrupt()

	direction := "顺时针"
	if reverse {
		direction = "逆时针"
		e.Log(fmt.Sprintf("[Skill] %s 发动【魔弹掌控】，魔弹将%s传递！", player.Name, direction))
	}

	// 执行魔弹效果
	return e.executeMagicBullet(player, reverse, isFusion, fusionCard)
}

// handleHolySwordDrawResponse 处理圣剑摸X弃X响应
func (e *GameEngine) handleHolySwordDrawResponse(act model.PlayerAction) error {
	interrupt := e.State.PendingInterrupt
	if interrupt == nil {
		return fmt.Errorf("没有待处理的中断")
	}

	player := e.State.Players[act.PlayerID]
	if player == nil {
		return fmt.Errorf("玩家不存在")
	}

	// 选项索引：0=X=0, 1=X=1, 2=X=2, 3=X=3
	x := 0
	if len(act.Selections) > 0 {
		x = act.Selections[0]
	}
	if x < 0 || x > 3 {
		x = 0
	}

	e.PopInterrupt()

	if x == 0 {
		e.Log(fmt.Sprintf("[Skill] %s 选择不摸不弃", player.Name))
	} else {
		// 摸X张牌
		e.DrawCards(player.ID, x)
		e.Log(fmt.Sprintf("[Skill] %s 摸了 %d 张牌", player.Name, x))

		// 推送弃牌中断
		e.PushInterrupt(&model.Interrupt{
			Type:     model.InterruptDiscard,
			PlayerID: player.ID,
			Context: map[string]interface{}{
				"discard_count":        x,
				"is_holy_sword":        true,
				"stay_in_turn":         false,
				"is_damage_resolution": false,
			},
		})
		e.Log(fmt.Sprintf("[Skill] %s 需要弃 %d 张牌", player.Name, x))
		return nil
	}

	// 继续游戏流程
	e.NextTurn()
	return nil
}

// handleSaintHealResponse 处理圣疗分配治疗响应
func (e *GameEngine) handleSaintHealResponse(act model.PlayerAction) error {
	interrupt := e.State.PendingInterrupt
	if interrupt == nil {
		return fmt.Errorf("没有待处理的中断")
	}

	data, ok := interrupt.Context.(map[string]interface{})
	if !ok {
		return fmt.Errorf("中断上下文格式错误")
	}

	player := e.State.Players[act.PlayerID]
	if player == nil {
		return fmt.Errorf("玩家不存在")
	}

	// 获取当前阶段（选择目标 or 分配点数）
	stage, _ := data["stage"].(string)

	if stage == "" || stage == "choose_targets" {
		// 阶段1：选择目标
		// 前端传递选项索引，需要转换为玩家ID
		targetIDs := make([]string, 0, len(act.Selections))

		// 如果前端传的是选项ID（玩家ID），需要从 ExtraArgs 获取
		if len(act.ExtraArgs) > 0 {
			targetIDs = act.ExtraArgs
		} else {
			// 否则从 Selections 获取（前端可能传递的是玩家列表的索引）
			playerList := e.GetAllPlayers()
			for _, idx := range act.Selections {
				if idx >= 0 && idx < len(playerList) {
					targetIDs = append(targetIDs, playerList[idx].ID)
				}
			}
		}

		if len(targetIDs) == 0 || len(targetIDs) > 3 {
			return fmt.Errorf("请选择1-3名目标")
		}

		// 保存目标，进入分配阶段
		data["targets"] = targetIDs
		data["stage"] = "allocate_heal"
		data["remaining_heal"] = 3
		e.State.PendingInterrupt.Context = data

		// 重新发送提示
		e.notifyInterruptPrompt()
		return nil
	}

	if stage == "allocate_heal" {
		// 阶段2：分配治疗点数
		targets, _ := data["targets"].([]string)
		if len(targets) == 0 {
			return fmt.Errorf("没有目标")
		}

		// 简化处理：平均分配3点治疗
		// 1个目标: +3
		// 2个目标: +2, +1
		// 3个目标: +1, +1, +1
		points := 3
		for i, targetID := range targets {
			healAmount := 1
			if len(targets) == 1 {
				healAmount = 3
			} else if len(targets) == 2 && i == 0 {
				healAmount = 2
			}
			if healAmount > points {
				healAmount = points
			}
			if healAmount > 0 {
				e.Heal(targetID, healAmount)
				target := e.State.Players[targetID]
				if target != nil {
					e.Log(fmt.Sprintf("[Skill] %s 获得 %d 点治疗", target.Name, healAmount))
				}
				points -= healAmount
			}
		}

		e.PopInterrupt()

		// 额外攻击行动
		token := model.ActionContext{
			Source:   "圣疗",
			MustType: "Attack",
		}
		player.TurnState.PendingActions = append(player.TurnState.PendingActions, token)
		e.Log(fmt.Sprintf("[Skill] %s 发动 [圣疗]，获得额外攻击行动", player.Name))

		return nil
	}

	return fmt.Errorf("未知的圣疗阶段")
}

// handleWeakChoiceInput 处理虚弱/选择中断
func (e *GameEngine) handleWeakChoiceInput(playerID string, selectionIndex int) error {
	// 1. 安全获取上下文数据
	if e.State.PendingInterrupt == nil {
		return fmt.Errorf("没有待处理的中断")
	}

	ctxData, ok := e.State.PendingInterrupt.Context.(map[string]interface{})
	if !ok {
		return fmt.Errorf("中断上下文格式错误")
	}

	choiceType, _ := ctxData["choice_type"].(string)

	if choiceType == "weak" {
		player := e.State.Players[playerID]

		// 2. 移除虚弱状态 (FieldCard)
		// 遍历场上牌，找到 EffectWeak 并移除
		newField := make([]*model.FieldCard, 0)
		foundWeak := false
		for _, fc := range player.Field {
			if fc.Mode == model.FieldEffect && fc.Effect == model.EffectWeak {
				// 将虚弱牌放入弃牌堆
				e.State.DiscardPile = append(e.State.DiscardPile, fc.Card)
				foundWeak = true
				// 不加入 newField，即从场上移除
				continue
			}
			newField = append(newField, fc)
		}
		player.Field = newField

		if foundWeak {
			e.Log(fmt.Sprintf("[System] %s 的虚弱状态已移除", player.Name))
		}

		// 3. 处理选项逻辑 (选项顺序：0=摸3张牌, 1=跳过回合)

		// 选项 0: 摸3张牌 (choose 1)
		if selectionIndex == 0 {
			e.Log(fmt.Sprintf("[Weak] %s 选择摸3张牌", player.Name))

			// 执行摸牌
			cards, newDeck, newDiscard := rules.DrawCards(e.State.Deck, e.State.DiscardPile, 3)
			e.State.Deck = newDeck
			e.State.DiscardPile = newDiscard
			player.Hand = append(player.Hand, cards...)
			e.NotifyDrawCards(player.ID, 3, "weak_choice")

			// 【核心修改】构造上下文，标记“留在当前回合”
			checkCtx := e.buildContext(player, nil, model.TriggerNone, nil)
			checkCtx.Flags["StayInTurn"] = true

			// 检查手牌上限 (可能会触发新的 InterruptDiscard)
			// 传入 nil Context，表示这是一个普通的检查
			e.checkHandLimit(player, checkCtx)

			// 弹出当前的虚弱中断
			e.PopInterrupt()

			// 如果没有触发爆牌中断，则进入正常的启动阶段 (Startup)
			if e.State.PendingInterrupt == nil {
				e.State.Phase = model.PhaseStartup
			}
			// 如果触发了爆牌，PopInterrupt 会将 Phase 设置为 DiscardSelection，这里就不需要动了

			return nil
		}

		// 选项 1: 跳过回合 (choose 2)
		if selectionIndex == 1 {
			e.Log(fmt.Sprintf("[Weak] %s 选择跳过回合", player.Name))
			if player.Tokens == nil {
				player.Tokens = map[string]int{}
			}
			player.Tokens["arbiter_skip_forced_doomsday"] = 1

			// 弹出当前中断
			e.PopInterrupt()

			// 设置阶段为 TurnEnd，Drive 会在下一次循环调用 NextTurn
			if e.State.PendingInterrupt == nil {
				e.State.Phase = model.PhaseTurnEnd
			}
			return nil
		}

		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	if choiceType == "buy_resource" {
		camp, _ := ctxData["camp"].(string)
		if camp == "" {
			if p := e.State.Players[playerID]; p != nil {
				camp = string(p.Camp)
			}
		}
		if camp == "" {
			return fmt.Errorf("购买资源选择缺少阵营信息")
		}

		switch selectionIndex {
		case 0:
			e.ModifyGem(camp, 1)
			e.Log(fmt.Sprintf("[Action] 购买结算：%s 阵营战绩区 +1 宝石", camp))
		case 1:
			e.ModifyCrystal(camp, 1)
			e.Log(fmt.Sprintf("[Action] 购买结算：%s 阵营战绩区 +1 水晶", camp))
		default:
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}

		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			e.State.Phase = model.PhaseExtraAction
		}
		return nil
	}
	if choiceType == "heal" {
		damageIdxAny, _ := ctxData["damage_index"]
		damageIdx, _ := damageIdxAny.(int)
		if damageIdx < 0 || damageIdx >= len(e.State.PendingDamageQueue) {
			return fmt.Errorf("伤害上下文不存在")
		}
		pd := &e.State.PendingDamageQueue[damageIdx]
		target := e.State.Players[pd.TargetID]
		if target == nil {
			return fmt.Errorf("目标不存在")
		}
		// selectionIndex 表示使用的治疗数量（0..max）
		healToUse := selectionIndex
		if healToUse < 0 {
			healToUse = 0
		}
		if healToUse > target.Heal {
			healToUse = target.Heal
		}
		if healToUse > pd.Damage {
			healToUse = pd.Damage
		}
		if healToUse > 0 {
			target.Heal -= healToUse
			pd.Damage -= healToUse
			e.Log(fmt.Sprintf("[Combat] %s 使用 %d 点治疗抵消伤害", target.Name, healToUse))
		} else {
			e.Log(fmt.Sprintf("[Combat] %s 选择不使用治疗", target.Name))
		}
		pd.HealResolved = true

		// 弹出当前中断
		e.PopInterrupt()
		// 继续留在伤害结算阶段
		if e.State.PendingInterrupt == nil {
			e.State.Phase = model.PhasePendingDamageResolution
		}
		return nil
	}
	if choiceType == "hero_roar_draw" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		drawCount := 0
		switch selectionIndex {
		case 0:
			drawCount = 0
		case 1:
			drawCount = 1
		default:
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		if drawCount > 0 {
			e.DrawCards(user.ID, drawCount)
		}
		e.Log(fmt.Sprintf("%s 的 [怒吼] 结算：摸%d张牌", user.Name, drawCount))
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			e.State.Phase = model.PhaseBeforeAction
		}
		return nil
	}
	if choiceType == "angel_song_pick" {
		type songPickOption struct {
			TargetID string
			Effect   string
			Label    string
		}
		var options []songPickOption
		if arr, ok := ctxData["options"].([]songPickOption); ok {
			options = append(options, arr...)
		} else if arr, ok := ctxData["options"].([]interface{}); ok {
			for _, v := range arr {
				m, ok := v.(map[string]interface{})
				if !ok || m == nil {
					continue
				}
				targetID, _ := m["target_id"].(string)
				effect, _ := m["effect"].(string)
				label, _ := m["label"].(string)
				if targetID == "" || effect == "" {
					continue
				}
				options = append(options, songPickOption{
					TargetID: targetID,
					Effect:   effect,
					Label:    label,
				})
			}
		} else if arr, ok := ctxData["options"].([]map[string]interface{}); ok {
			for _, m := range arr {
				if m == nil {
					continue
				}
				targetID, _ := m["target_id"].(string)
				effect, _ := m["effect"].(string)
				label, _ := m["label"].(string)
				if targetID == "" || effect == "" {
					continue
				}
				options = append(options, songPickOption{
					TargetID: targetID,
					Effect:   effect,
					Label:    label,
				})
			}
		}
		if selectionIndex < 0 || selectionIndex >= len(options) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		selected := options[selectionIndex]
		target := e.State.Players[selected.TargetID]
		if target == nil {
			return fmt.Errorf("目标不存在")
		}
		if !e.RemoveFieldCardBy(selected.TargetID, model.EffectType(selected.Effect), playerID) {
			return fmt.Errorf("所选基础效果已不存在")
		}
		user := e.State.Players[playerID]
		userName := playerID
		if user != nil {
			userName = user.Name
		}
		e.Log(fmt.Sprintf("%s 的 [天使之歌]：移除了 %s", userName, selected.Label))
		e.NotifyActionStep(fmt.Sprintf("%s 的【天使之歌】移除了 %s", userName, selected.Label))
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			e.State.Phase = model.PhaseBeforeAction
		}
		return nil
	}
	if choiceType == "seal_break_pick_effect" {
		type sealBreakOption struct {
			TargetID   string
			FieldIndex int
			Effect     string
			Label      string
		}
		var options []sealBreakOption
		if arr, ok := ctxData["options"].([]sealBreakOption); ok {
			options = append(options, arr...)
		} else if arr, ok := ctxData["options"].([]interface{}); ok {
			for _, v := range arr {
				m, ok := v.(map[string]interface{})
				if !ok || m == nil {
					continue
				}
				targetID, _ := m["target_id"].(string)
				effect, _ := m["effect"].(string)
				label, _ := m["label"].(string)
				fieldIndex := -1
				if iv, ok := m["field_index"].(int); ok {
					fieldIndex = iv
				} else if fv, ok := m["field_index"].(float64); ok {
					fieldIndex = int(fv)
				}
				if targetID == "" || effect == "" || fieldIndex < 0 {
					continue
				}
				options = append(options, sealBreakOption{
					TargetID:   targetID,
					FieldIndex: fieldIndex,
					Effect:     effect,
					Label:      label,
				})
			}
		} else if arr, ok := ctxData["options"].([]map[string]interface{}); ok {
			for _, m := range arr {
				if m == nil {
					continue
				}
				targetID, _ := m["target_id"].(string)
				effect, _ := m["effect"].(string)
				label, _ := m["label"].(string)
				fieldIndex := -1
				if iv, ok := m["field_index"].(int); ok {
					fieldIndex = iv
				} else if fv, ok := m["field_index"].(float64); ok {
					fieldIndex = int(fv)
				}
				if targetID == "" || effect == "" || fieldIndex < 0 {
					continue
				}
				options = append(options, sealBreakOption{
					TargetID:   targetID,
					FieldIndex: fieldIndex,
					Effect:     effect,
					Label:      label,
				})
			}
		}
		if selectionIndex < 0 || selectionIndex >= len(options) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		selected := options[selectionIndex]
		user := e.State.Players[playerID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}

		takenCard, err := e.TakeFieldCard(selected.TargetID, selected.FieldIndex, playerID)
		if err != nil {
			return err
		}
		user.Hand = append(user.Hand, takenCard)

		target := e.State.Players[selected.TargetID]
		targetName := selected.TargetID
		if target != nil {
			targetName = target.Name
		}
		display := selected.Label
		if strings.TrimSpace(display) == "" {
			display = fmt.Sprintf("%s：%s", targetName, selected.Effect)
		}
		e.Log(fmt.Sprintf("%s 的 [封印破碎]：收回了 %s，并将该牌加入手牌", user.Name, display))
		e.NotifyActionStep(fmt.Sprintf("%s 的【封印破碎】收回了 %s", user.Name, display))

		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			e.State.Phase = model.PhaseBeforeAction
		}
		return nil
	}
	if choiceType == "angel_bond_heal_target" {
		var targetIDs []string
		if arr, ok := ctxData["target_ids"].([]string); ok {
			targetIDs = append(targetIDs, arr...)
		} else if arr, ok := ctxData["target_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					targetIDs = append(targetIDs, s)
				}
			}
		}
		if selectionIndex < 0 || selectionIndex >= len(targetIDs) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		targetID := targetIDs[selectionIndex]
		target := e.State.Players[targetID]
		if target == nil {
			return fmt.Errorf("目标不存在")
		}
		e.Heal(target.ID, 1)
		userName := playerID
		if user := e.State.Players[playerID]; user != nil {
			userName = user.Name
		}
		e.Log(fmt.Sprintf("%s 的 [天使羁绊] 生效：%s 获得 +1 治疗", userName, target.Name))
		e.PopInterrupt()
		return nil
	}

	if choiceType == "five_elements_bind" {
		// 五系束缚选择处理
		// 选项 0: 摸 drawCount 张牌
		// 选项 1: 放弃行动，移除五系束缚
		drawCountAny, _ := ctxData["draw_count"]
		playerIDAny, _ := ctxData["player_id"]
		drawCount, ok := drawCountAny.(int)
		if !ok {
			// 兼容 JSON 反序列化后的 float64
			if f, ok := drawCountAny.(float64); ok {
				drawCount = int(f)
			}
		}
		targetPlayerID, _ := playerIDAny.(string)

		player := e.State.Players[targetPlayerID]
		if player == nil {
			e.PopInterrupt()
			return fmt.Errorf("五系束缚目标玩家不存在")
		}

		// 移除五系束缚效果
		newField := make([]*model.FieldCard, 0)
		for _, fc := range player.Field {
			if fc.Mode == model.FieldEffect && fc.Effect == model.EffectFiveElementsBind {
				// 将牌放入弃牌堆
				e.State.DiscardPile = append(e.State.DiscardPile, fc.Card)
				continue
			}
			newField = append(newField, fc)
		}
		player.Field = newField

		if selectionIndex == 0 {
			// 选择摸牌
			e.Log(fmt.Sprintf("[FiveElementsBind] %s 选择摸 %d 张牌", player.Name, drawCount))
			cards, newDeck, newDiscard := rules.DrawCards(e.State.Deck, e.State.DiscardPile, drawCount)
			e.State.Deck = newDeck
			e.State.DiscardPile = newDiscard
			player.Hand = append(player.Hand, cards...)
			e.NotifyDrawCards(player.ID, drawCount, "five_elements_bind")

			// 检查手牌上限
			checkCtx := e.buildContext(player, nil, model.TriggerNone, nil)
			checkCtx.Flags["StayInTurn"] = true
			e.checkHandLimit(player, checkCtx)

			e.PopInterrupt()

			if e.State.PendingInterrupt == nil {
				e.State.Phase = model.PhaseStartup
			}
			return nil
		}

		if selectionIndex == 1 {
			// 选择放弃行动
			e.Log(fmt.Sprintf("[FiveElementsBind] %s 选择放弃行动", player.Name))
			if player.Tokens == nil {
				player.Tokens = map[string]int{}
			}
			player.Tokens["arbiter_skip_forced_doomsday"] = 1

			e.PopInterrupt()

			if e.State.PendingInterrupt == nil {
				e.State.Phase = model.PhaseTurnEnd
			}
			return nil
		}

		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	if choiceType == "arbiter_forced_doomsday_target" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		var targetIDs []string
		if arr, ok := ctxData["target_ids"].([]string); ok {
			targetIDs = arr
		} else if arr, ok := ctxData["target_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					targetIDs = append(targetIDs, s)
				}
			}
		}
		if selectionIndex < 0 || selectionIndex >= len(targetIDs) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		targetID := targetIDs[selectionIndex]
		target := e.State.Players[targetID]
		if target == nil {
			return fmt.Errorf("目标不存在")
		}
		judgment := 0
		if user.Tokens != nil {
			judgment = user.Tokens["judgment"]
		}
		user.Tokens["judgment"] = 0
		user.Tokens["arbiter_forced_doomsday_done_turn"] = 1
		if judgment > 0 {
			e.InflictDamage(userID, targetID, judgment, "magic")
		}
		e.Log(fmt.Sprintf("%s 触发强制 [末日审判]，对 %s 造成%d点法术伤害", user.Name, target.Name, judgment))
		e.PopInterrupt()
		if len(e.State.PendingDamageQueue) > 0 && e.State.PendingInterrupt == nil {
			e.State.Phase = model.PhasePendingDamageResolution
			e.State.ReturnPhase = model.PhaseTurnEnd
		}
		return nil
	}
	if choiceType == "arbiter_balance_mode" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		switch selectionIndex {
		case 0:
			for _, c := range user.Hand {
				e.State.DiscardPile = append(e.State.DiscardPile, c)
			}
			user.Hand = nil
			e.Log(fmt.Sprintf("%s 选择判决天平分支1：弃掉所有手牌", user.Name))
		case 1:
			maxHand := user.MaxHand
			if len(user.Hand) < maxHand {
				e.DrawCards(user.ID, maxHand-len(user.Hand))
			}
			e.ModifyGem(string(user.Camp), 1)
			e.Log(fmt.Sprintf("%s 选择判决天平分支2：补牌到上限并我方战绩区+1红宝石", user.Name))
		default:
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		e.PopInterrupt()
		return nil
	}
	if choiceType == "frost_prayer_target" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		var targetIDs []string
		if arr, ok := ctxData["target_ids"].([]string); ok {
			targetIDs = append(targetIDs, arr...)
		} else if arr, ok := ctxData["target_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					targetIDs = append(targetIDs, s)
				}
			}
		}
		if selectionIndex < 0 || selectionIndex >= len(targetIDs) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		targetID := targetIDs[selectionIndex]
		target := e.State.Players[targetID]
		if target == nil {
			return fmt.Errorf("目标不存在")
		}
		e.Heal(targetID, 1)
		e.Log(fmt.Sprintf("%s 的 [冰霜祷言] 生效：%s +1治疗", user.Name, target.Name))
		e.PopInterrupt()
		return nil
	}
	if choiceType == "valkyrie_military_glory_mode" {
		userID, _ := ctxData["user_id"].(string)
		camp, _ := ctxData["camp"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		if selectionIndex == 0 {
			e.Heal(userID, 1)
			if user.Tokens == nil {
				user.Tokens = map[string]int{}
			}
			user.Tokens["valkyrie_spirit"] = 0
			e.Log(fmt.Sprintf("%s 选择军神威光选项1：+1治疗并脱离英灵形态", user.Name))
			e.PopInterrupt()
			return nil
		}
		if selectionIndex == 1 {
			maxX := 0
			if v, ok := ctxData["max_x"].(int); ok {
				maxX = v
			} else if f, ok := ctxData["max_x"].(float64); ok {
				maxX = int(f)
			}
			if maxX <= 0 {
				return fmt.Errorf("当前阵营无可用能量")
			}
			e.State.PendingInterrupt.Context = map[string]interface{}{
				"choice_type": "valkyrie_military_glory_x",
				"user_id":     userID,
				"camp":        camp,
				"max_x":       maxX,
			}
			e.notifyInterruptPrompt()
			return nil
		}
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	if choiceType == "valkyrie_military_glory_x" {
		userID, _ := ctxData["user_id"].(string)
		camp, _ := ctxData["camp"].(string)
		maxX := 0
		if v, ok := ctxData["max_x"].(int); ok {
			maxX = v
		} else if f, ok := ctxData["max_x"].(float64); ok {
			maxX = int(f)
		}
		if maxX <= 0 {
			return fmt.Errorf("当前阵营无可用能量")
		}
		x := selectionIndex + 1
		if x <= 0 || x > maxX || x >= 3 {
			return fmt.Errorf("无效的X值")
		}
		targetIDs := make([]string, 0, len(e.State.PlayerOrder))
		for _, pid := range e.State.PlayerOrder {
			if e.State.Players[pid] == nil {
				continue
			}
			targetIDs = append(targetIDs, pid)
		}
		e.State.PendingInterrupt.Context = map[string]interface{}{
			"choice_type": "valkyrie_military_glory_target",
			"user_id":     userID,
			"camp":        camp,
			"x":           x,
			"target_ids":  targetIDs,
		}
		e.notifyInterruptPrompt()
		return nil
	}
	if choiceType == "valkyrie_military_glory_target" {
		userID, _ := ctxData["user_id"].(string)
		camp, _ := ctxData["camp"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		var targetIDs []string
		if arr, ok := ctxData["target_ids"].([]string); ok {
			targetIDs = append(targetIDs, arr...)
		} else if arr, ok := ctxData["target_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					targetIDs = append(targetIDs, s)
				}
			}
		}
		if selectionIndex < 0 || selectionIndex >= len(targetIDs) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		targetID := targetIDs[selectionIndex]
		target := e.State.Players[targetID]
		if target == nil {
			return fmt.Errorf("目标不存在")
		}
		x := 0
		if v, ok := ctxData["x"].(int); ok {
			x = v
		} else if f, ok := ctxData["x"].(float64); ok {
			x = int(f)
		}
		if x <= 0 || x >= 3 {
			return fmt.Errorf("无效的X值")
		}
		total := e.GetCampCrystals(camp) + e.GetCampGems(camp)
		if x > total {
			return fmt.Errorf("阵营能量不足")
		}
		useCrystal := x
		if useCrystal > e.GetCampCrystals(camp) {
			useCrystal = e.GetCampCrystals(camp)
		}
		if useCrystal > 0 {
			e.ModifyCrystal(camp, -useCrystal)
		}
		remain := x - useCrystal
		if remain > 0 {
			e.ModifyGem(camp, -remain)
		}
		e.Heal(targetID, x)
		if user.Tokens == nil {
			user.Tokens = map[string]int{}
		}
		user.Tokens["valkyrie_spirit"] = 0
		e.Log(fmt.Sprintf("%s 选择军神威光选项2：移除%d星石并使 %s +%d治疗", user.Name, x, target.Name, x))
		e.PopInterrupt()
		return nil
	}
	if choiceType == "valkyrie_heroic_extra_confirm" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		if selectionIndex == 1 {
			e.PopInterrupt()
			e.resumePendingAttackHit(ctxData)
			return nil
		}
		if selectionIndex == 0 {
			var magicIndices []int
			for i, c := range user.Hand {
				if c.Type == model.CardTypeMagic {
					magicIndices = append(magicIndices, i)
				}
			}
			if len(magicIndices) == 0 {
				e.PopInterrupt()
				e.resumePendingAttackHit(ctxData)
				return nil
			}
			e.State.PendingInterrupt.Context = map[string]interface{}{
				"choice_type":   "valkyrie_heroic_discard_card",
				"user_id":       userID,
				"magic_indices": magicIndices,
				"user_ctx":      ctxData["user_ctx"],
			}
			e.notifyInterruptPrompt()
			return nil
		}
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	if choiceType == "valkyrie_heroic_discard_card" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		var magicIndices []int
		if arr, ok := ctxData["magic_indices"].([]int); ok {
			magicIndices = arr
		} else if arr, ok := ctxData["magic_indices"].([]interface{}); ok {
			for _, v := range arr {
				if f, ok := v.(float64); ok {
					magicIndices = append(magicIndices, int(f))
				}
			}
		}
		cardIdx, ok := resolveSelectionToCandidate(selectionIndex, magicIndices)
		if !ok {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		if cardIdx < 0 || cardIdx >= len(user.Hand) || user.Hand[cardIdx].Type != model.CardTypeMagic {
			return fmt.Errorf("请选择法术牌")
		}
		card := user.Hand[cardIdx]
		e.NotifyCardRevealed(userID, []model.Card{card}, "discard")
		user.Hand = append(user.Hand[:cardIdx], user.Hand[cardIdx+1:]...)
		e.State.DiscardPile = append(e.State.DiscardPile, card)
		e.State.PendingInterrupt.Context = map[string]interface{}{
			"choice_type": "valkyrie_heroic_heal_target",
			"user_id":     userID,
			"target_ids":  append([]string{}, e.State.PlayerOrder...),
			"user_ctx":    ctxData["user_ctx"],
		}
		e.notifyInterruptPrompt()
		return nil
	}
	if choiceType == "valkyrie_heroic_heal_target" {
		var targetIDs []string
		if arr, ok := ctxData["target_ids"].([]string); ok {
			targetIDs = arr
		} else if arr, ok := ctxData["target_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					targetIDs = append(targetIDs, s)
				}
			}
		}
		if selectionIndex < 0 || selectionIndex >= len(targetIDs) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		targetID := targetIDs[selectionIndex]
		target := e.State.Players[targetID]
		if target == nil {
			return fmt.Errorf("目标不存在")
		}
		e.Heal(targetID, 1)
		e.Log(fmt.Sprintf("%s 因英灵召唤额外效果，获得1点治疗", target.Name))
		e.PopInterrupt()
		e.resumePendingAttackHit(ctxData)
		return nil
	}
	if choiceType == "elementalist_bonus_confirm" {
		if selectionIndex == 1 {
			return e.resolveElementalistBonus(ctxData, false, -1)
		}
		if selectionIndex == 0 {
			userID, _ := ctxData["user_id"].(string)
			user := e.State.Players[userID]
			if user == nil {
				return fmt.Errorf("玩家不存在")
			}
			bonusElement, _ := ctxData["bonus_element"].(string)
			var matching []int
			for i, c := range user.Hand {
				if string(c.Element) == bonusElement {
					matching = append(matching, i)
				}
			}
			if len(matching) == 0 {
				return e.resolveElementalistBonus(ctxData, false, -1)
			}
			ctxData["choice_type"] = "elementalist_bonus_card"
			ctxData["matching_indices"] = matching
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return nil
		}
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	if choiceType == "elementalist_bonus_card" {
		var matching []int
		if arr, ok := ctxData["matching_indices"].([]int); ok {
			matching = arr
		} else if arr, ok := ctxData["matching_indices"].([]interface{}); ok {
			for _, v := range arr {
				if f, ok := v.(float64); ok {
					matching = append(matching, int(f))
				}
			}
		}
		cardIdx, ok := resolveSelectionToCandidate(selectionIndex, matching)
		if !ok {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		return e.resolveElementalistBonus(ctxData, true, cardIdx)
	}
	if choiceType == "adventurer_fraud_mode" {
		can2, _ := ctxData["can2"].(bool)
		can3, _ := ctxData["can3"].(bool)
		var modeList []string
		if can2 {
			modeList = append(modeList, "2")
		}
		if can3 {
			modeList = append(modeList, "3")
		}
		if selectionIndex < 0 || selectionIndex >= len(modeList) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		mode := modeList[selectionIndex]
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		if mode == "2" {
			counts := map[string]int{}
			for _, c := range user.Hand {
				ele := string(c.Element)
				counts[ele]++
			}
			var discardElems []string
			for _, ele := range []string{
				string(model.ElementWater), string(model.ElementFire), string(model.ElementEarth),
				string(model.ElementWind), string(model.ElementThunder), string(model.ElementLight), string(model.ElementDark),
			} {
				if counts[ele] >= 2 {
					discardElems = append(discardElems, ele)
				}
			}
			if len(discardElems) == 0 {
				return fmt.Errorf("没有可用于弃2同系的元素")
			}
			ctxData["choice_type"] = "adventurer_fraud_attack_element"
			ctxData["fraud_mode"] = "2"
			ctxData["discard_elements"] = discardElems
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return nil
		}
		// mode == 3
		combos := e.buildFraudCombos(user, model.ElementDark, 3, true)
		if len(combos) == 0 {
			return fmt.Errorf("没有可用于弃3同系的组合")
		}
		ctxData["choice_type"] = "adventurer_fraud_discard_combo"
		ctxData["fraud_mode"] = "3"
		ctxData["chosen_element"] = string(model.ElementDark)
		ctxData["combos"] = combos
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return nil
	}
	if choiceType == "adventurer_fraud_attack_element" {
		attackElems := []string{
			string(model.ElementWater), string(model.ElementFire), string(model.ElementEarth),
			string(model.ElementWind), string(model.ElementThunder),
		}
		if selectionIndex < 0 || selectionIndex >= len(attackElems) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		ctxData["chosen_element"] = attackElems[selectionIndex]
		ctxData["choice_type"] = "adventurer_fraud_discard_element"
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return nil
	}
	if choiceType == "adventurer_fraud_discard_element" {
		var elems []string
		if arr, ok := ctxData["discard_elements"].([]string); ok {
			elems = arr
		} else if arr, ok := ctxData["discard_elements"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					elems = append(elems, s)
				}
			}
		}
		if selectionIndex < 0 || selectionIndex >= len(elems) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		ele := elems[selectionIndex]
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		combos := e.buildFraudCombos(user, model.Element(ele), 2, false)
		if len(combos) == 0 {
			return fmt.Errorf("该元素无可弃组合")
		}
		ctxData["choice_type"] = "adventurer_fraud_discard_combo"
		ctxData["discard_element"] = ele
		ctxData["combos"] = combos
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return nil
	}
	if choiceType == "adventurer_fraud_discard_combo" {
		var combos []string
		if arr, ok := ctxData["combos"].([]string); ok {
			combos = arr
		} else if arr, ok := ctxData["combos"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					combos = append(combos, s)
				}
			}
		}
		if selectionIndex < 0 || selectionIndex >= len(combos) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		parts := strings.Split(combos[selectionIndex], ":")
		if len(parts) != 2 {
			return fmt.Errorf("组合格式错误")
		}
		ele := parts[0]
		idxStrs := strings.Split(parts[1], ",")
		var idxs []int
		for _, s := range idxStrs {
			i, err := strconv.Atoi(s)
			if err != nil {
				return fmt.Errorf("组合索引错误")
			}
			idxs = append(idxs, i)
		}
		sort.Sort(sort.Reverse(sort.IntSlice(idxs)))
		for _, idx := range idxs {
			if idx < 0 || idx >= len(user.Hand) {
				return fmt.Errorf("弃牌索引越界")
			}
			card := user.Hand[idx]
			e.NotifyCardRevealed(userID, []model.Card{card}, "discard")
			user.Hand = append(user.Hand[:idx], user.Hand[idx+1:]...)
			e.State.DiscardPile = append(e.State.DiscardPile, card)
		}
		mode, _ := ctxData["fraud_mode"].(string)
		attackElement := model.Element(ele)
		canBeResponded := true
		if mode == "3" {
			attackElement = model.ElementDark
			canBeResponded = false
		} else {
			if chosen, ok := ctxData["chosen_element"].(string); ok && chosen != "" {
				attackElement = model.Element(chosen)
			}
		}

		// 兼容两种流程：
		// 1) 老流程：攻击开始时响应，直接改写当前攻击事件
		// 2) 新流程：作为主动技能发动，创建一次“欺诈攻击”行动
		rawCtx, ok := ctxData["user_ctx"].(*model.Context)
		if ok && rawCtx != nil && rawCtx.TriggerCtx != nil && rawCtx.TriggerCtx.Card != nil && rawCtx.TriggerCtx.AttackInfo != nil {
			rawCtx.TriggerCtx.Card.Faction = ""
			rawCtx.TriggerCtx.Card.Element = attackElement
			rawCtx.TriggerCtx.Card.Damage = 2
			rawCtx.TriggerCtx.AttackInfo.CanBeResponded = canBeResponded
			e.Log(fmt.Sprintf("%s 发动[欺诈]完成，弃同系牌并将本次攻击改为 %s", user.Name, attackElement))
			// 兼容“攻击开始时改写”老流程：此时 AttackStart 不会再次进入调度，直接结算强运。
			e.resolveAdventurerLuckyFortuneFromFraud(user)
			e.PopInterrupt()
			return nil
		}

		targetID, _ := ctxData["fraud_target_id"].(string)
		if targetID == "" || e.State.Players[targetID] == nil {
			return fmt.Errorf("欺诈目标无效")
		}
		virtualCard := model.Card{
			ID:      "fraud_virtual_attack",
			Name:    "欺诈",
			Type:    model.CardTypeAttack,
			Element: attackElement,
			Faction: "",
			Damage:  2,
		}
		e.State.ActionQueue = append(e.State.ActionQueue, model.QueuedAction{
			SourceID:    userID,
			TargetID:    targetID,
			Type:        model.ActionAttack,
			Element:     attackElement,
			Card:        &virtualCard,
			CardIndex:   -1,
			SourceSkill: "adventurer_fraud",
		})
		e.State.Phase = model.PhaseBeforeAction
		e.Log(fmt.Sprintf("%s 发动[欺诈]完成，弃同系牌并对 %s 发起%s系主动攻击", user.Name, e.State.Players[targetID].Name, attackElement))
		e.PopInterrupt()
		return nil
	}
	if choiceType == "adventurer_paradise_target" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		var allyIDs []string
		if arr, ok := ctxData["ally_ids"].([]string); ok {
			allyIDs = arr
		} else if arr, ok := ctxData["ally_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					allyIDs = append(allyIDs, s)
				}
			}
		}
		if selectionIndex < 0 || selectionIndex >= len(allyIDs) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		ally := e.State.Players[allyIDs[selectionIndex]]
		if ally == nil {
			return fmt.Errorf("队友不存在")
		}

		transferGem := toIntContextValue(ctxData["transfer_gem"])
		transferCrystal := toIntContextValue(ctxData["transfer_crystal"])
		transferTotal := toIntContextValue(ctxData["transfer_total"])
		if transferTotal <= 0 {
			transferTotal = transferGem + transferCrystal
		}
		fromPending, _ := ctxData["from_pending"].(bool)
		if transferTotal <= 0 {
			e.clearAdventurerExtractState(user)
			e.PopInterrupt()
			return nil
		}

		capLeft := e.getPlayerEnergyCap(ally) - (ally.Gem + ally.Crystal)
		if capLeft < transferTotal {
			return fmt.Errorf("%s 能量空间不足，无法接收全部提炼结果", ally.Name)
		}
		if !fromPending {
			if user.Gem < transferGem || user.Crystal < transferCrystal {
				return fmt.Errorf("自身提炼结果异常，无法转移")
			}
			user.Gem -= transferGem
			user.Crystal -= transferCrystal
		}
		ally.Gem += transferGem
		ally.Crystal += transferCrystal

		removedEnergy := false
		if user.Crystal > 0 {
			user.Crystal--
			removedEnergy = true
		} else if user.Gem > 0 {
			user.Gem--
			removedEnergy = true
		}
		e.clearAdventurerExtractState(user)
		if removedEnergy {
			e.Log(fmt.Sprintf("%s 发动[冒险者天堂]，将提炼结果交给 %s（宝石%d/水晶%d），并移除自身1点能量", user.Name, ally.Name, transferGem, transferCrystal))
		} else {
			e.Log(fmt.Sprintf("%s 发动[冒险者天堂]，将提炼结果交给 %s（宝石%d/水晶%d）", user.Name, ally.Name, transferGem, transferCrystal))
		}
		e.PopInterrupt()
		return nil
	}
	if choiceType == "adventurer_steal_sky_mode" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		enemyCamp, _ := ctxData["enemy_camp"].(string)
		selfCamp, _ := ctxData["self_camp"].(string)
		switch selectionIndex {
		case 0:
			if e.GetCampGems(enemyCamp) > 0 {
				e.ModifyGem(enemyCamp, -1)
				e.ModifyGem(selfCamp, 1)
			}
		case 1:
			cr := e.GetCampCrystals(selfCamp)
			if cr > 0 {
				e.ModifyCrystal(selfCamp, -cr)
				e.ModifyGem(selfCamp, cr)
			}
		default:
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		e.State.PendingInterrupt.Context = map[string]interface{}{
			"choice_type": "adventurer_steal_sky_extra_action",
			"user_id":     userID,
		}
		e.notifyInterruptPrompt()
		e.Log(fmt.Sprintf("%s 完成[偷天换日]主效果，等待选择额外行动", user.Name))
		return nil
	}
	if choiceType == "adventurer_steal_sky_extra_action" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		switch selectionIndex {
		case 0:
			user.TurnState.PendingActions = append(user.TurnState.PendingActions, model.ActionContext{Source: "偷天换日", MustType: "Attack"})
		case 1:
			user.TurnState.PendingActions = append(user.TurnState.PendingActions, model.ActionContext{Source: "偷天换日", MustType: "Magic"})
		default:
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		e.PopInterrupt()
		return nil
	}
	if choiceType == "priest_divine_contract_x" {
		userID, _ := ctxData["user_id"].(string)
		targetID, _ := ctxData["target_id"].(string)
		user := e.State.Players[userID]
		target := e.State.Players[targetID]
		if user == nil || target == nil {
			return fmt.Errorf("神圣契约目标不存在")
		}
		if target.Camp != user.Camp {
			return fmt.Errorf("神圣契约目标必须是队友")
		}
		maxX := 0
		if v, ok := ctxData["max_x"].(int); ok {
			maxX = v
		} else if f, ok := ctxData["max_x"].(float64); ok {
			maxX = int(f)
		}
		x := selectionIndex + 1
		if x < 1 || x > maxX {
			return fmt.Errorf("无效的X值")
		}
		if x > user.Heal {
			return fmt.Errorf("当前治疗不足，无法转移%d点治疗", x)
		}

		before := target.Heal
		user.Heal -= x
		if before <= 4 {
			target.Heal = before + x
			if target.Heal > 4 {
				target.Heal = 4
			}
		}
		after := target.Heal
		if before > 4 {
			e.Log(fmt.Sprintf("%s 的 [神圣契约] 生效：移除自身%d点治疗；目标 %s 当前治疗已超过4（%d），保持不变",
				user.Name, x, target.Name, before))
		} else {
			e.Log(fmt.Sprintf("%s 的 [神圣契约] 生效：移除自身%d点治疗并转移给 %s（%d -> %d）",
				user.Name, x, target.Name, before, after))
		}

		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			e.State.Phase = model.PhaseExtraAction
		}
		return nil
	}
	if choiceType == "priest_divine_domain_mode" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		var modeOptions []string
		if arr, ok := ctxData["mode_options"].([]string); ok {
			modeOptions = append(modeOptions, arr...)
		} else if arr, ok := ctxData["mode_options"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					modeOptions = append(modeOptions, s)
				}
			}
		}
		if selectionIndex < 0 || selectionIndex >= len(modeOptions) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		mode := modeOptions[selectionIndex]
		switch mode {
		case "damage":
			var allTargets []string
			if arr, ok := ctxData["all_target_ids"].([]string); ok {
				allTargets = append(allTargets, arr...)
			} else if arr, ok := ctxData["all_target_ids"].([]interface{}); ok {
				for _, v := range arr {
					if s, ok := v.(string); ok {
						allTargets = append(allTargets, s)
					}
				}
			}
			if len(allTargets) == 0 {
				return fmt.Errorf("无可选伤害目标")
			}
			ctxData["choice_type"] = "priest_divine_domain_damage_target"
			ctxData["target_ids"] = allTargets
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return nil
		case "heal":
			var allyTargets []string
			if arr, ok := ctxData["ally_target_ids"].([]string); ok {
				allyTargets = append(allyTargets, arr...)
			} else if arr, ok := ctxData["ally_target_ids"].([]interface{}); ok {
				for _, v := range arr {
					if s, ok := v.(string); ok {
						allyTargets = append(allyTargets, s)
					}
				}
			}
			if len(allyTargets) == 0 {
				return fmt.Errorf("无可选队友目标")
			}
			ctxData["choice_type"] = "priest_divine_domain_heal_target"
			ctxData["target_ids"] = allyTargets
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return nil
		default:
			return fmt.Errorf("无效的神圣领域分支")
		}
	}
	if choiceType == "priest_divine_domain_damage_target" || choiceType == "priest_divine_domain_heal_target" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		var targetIDs []string
		if arr, ok := ctxData["target_ids"].([]string); ok {
			targetIDs = append(targetIDs, arr...)
		} else if arr, ok := ctxData["target_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					targetIDs = append(targetIDs, s)
				}
			}
		}
		if selectionIndex < 0 || selectionIndex >= len(targetIDs) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		targetID := targetIDs[selectionIndex]
		target := e.State.Players[targetID]
		if target == nil {
			return fmt.Errorf("目标不存在")
		}
		if choiceType == "priest_divine_domain_damage_target" {
			if user.Heal <= 0 {
				return fmt.Errorf("神圣领域分支①需要至少1点治疗")
			}
			user.Heal--
			e.AddPendingDamage(model.PendingDamage{
				SourceID:   user.ID,
				TargetID:   targetID,
				Damage:     2,
				DamageType: "magic",
				Stage:      0,
			})
			e.Log(fmt.Sprintf("%s 的 [神圣领域] 分支①生效：移除1点治疗，对 %s 造成2点法术伤害", user.Name, target.Name))
			e.PopInterrupt()
			if e.State.PendingInterrupt == nil {
				if len(e.State.PendingDamageQueue) > 0 {
					e.State.Phase = model.PhasePendingDamageResolution
					e.State.ReturnPhase = model.PhaseExtraAction
				} else {
					e.State.Phase = model.PhaseExtraAction
				}
			}
			return nil
		}
		if target.Camp != user.Camp || target.ID == user.ID {
			return fmt.Errorf("神圣领域分支②目标必须是其他队友")
		}
		e.Heal(user.ID, 2)
		e.Heal(targetID, 1)
		e.Log(fmt.Sprintf("%s 的 [神圣领域] 分支②生效：自身+2治疗，%s +1治疗", user.Name, target.Name))
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			e.State.Phase = model.PhaseExtraAction
		}
		return nil
	}
	if choiceType == "holy_lancer_earth_spear_x" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		maxX := 0
		if v, ok := ctxData["max_x"].(int); ok {
			maxX = v
		} else if f, ok := ctxData["max_x"].(float64); ok {
			maxX = int(f)
		}
		x := selectionIndex + 1
		if x < 1 || x > maxX || x > user.Heal {
			return fmt.Errorf("无效的X值")
		}
		userCtx, ok := ctxData["user_ctx"].(*model.Context)
		if !ok || userCtx == nil || userCtx.TriggerCtx == nil || userCtx.TriggerCtx.DamageVal == nil {
			return fmt.Errorf("地枪上下文丢失")
		}
		user.Heal -= x
		*userCtx.TriggerCtx.DamageVal += x
		if user.Tokens == nil {
			user.Tokens = map[string]int{}
		}
		user.Tokens["holy_lancer_block_sacred_strike"] = 1
		e.Log(fmt.Sprintf("%s 发动 [地枪]，移除%d治疗，本次伤害+%d", user.Name, x, x))
		e.PopInterrupt()
		e.resumePendingAttackHit(ctxData)
		return nil
	}
	if choiceType == "prayer_power_blessing_trigger" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		if selectionIndex == 0 {
			e.RemoveFieldCard(user.ID, model.EffectPowerBlessing)
			sourceID, _ := ctxData["source_id"].(string)
			targetID, _ := ctxData["target_id"].(string)
			for i := range e.State.PendingDamageQueue {
				pd := &e.State.PendingDamageQueue[i]
				if pd.SourceID != sourceID || pd.TargetID != targetID {
					continue
				}
				if !strings.EqualFold(pd.DamageType, "Attack") {
					continue
				}
				pd.Damage += 2
				e.Log(fmt.Sprintf("%s 的 [威力赐福] 生效，本次攻击伤害+2", user.Name))
				break
			}
		}
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil && len(e.State.PendingDamageQueue) > 0 {
			e.State.Phase = model.PhasePendingDamageResolution
		}
		return nil
	}
	if choiceType == "prayer_swift_blessing_trigger" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		if selectionIndex == 0 {
			e.RemoveFieldCard(user.ID, model.EffectSwiftBlessing)
			user.TurnState.PendingActions = append(user.TurnState.PendingActions, model.ActionContext{
				Source:   "迅捷赐福",
				MustType: "Attack",
			})
			e.Log(fmt.Sprintf("%s 的 [迅捷赐福] 生效，获得额外攻击行动", user.Name))
		}
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil && e.State.Phase != model.PhaseExtraAction {
			e.State.Phase = model.PhaseExtraAction
		}
		return nil
	}
	if choiceType == "crk_bloody_prayer_x" {
		maxX := 0
		if v, ok := ctxData["max_x"].(int); ok {
			maxX = v
		} else if f, ok := ctxData["max_x"].(float64); ok {
			maxX = int(f)
		}
		x := selectionIndex + 1
		if x < 1 || x > maxX {
			return fmt.Errorf("无效的X值")
		}
		ctxData["x_value"] = x
		ctxData["selected_ally_ids"] = []string{}
		var allyIDs []string
		if arr, ok := ctxData["ally_ids"].([]string); ok {
			allyIDs = arr
		} else if arr, ok := ctxData["ally_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					allyIDs = append(allyIDs, s)
				}
			}
		}
		if len(allyIDs) == 0 {
			return fmt.Errorf("没有可分配治疗的队友")
		}
		if len(allyIDs) >= 2 && x >= 2 {
			ctxData["choice_type"] = "crk_bloody_prayer_ally_count"
		} else {
			ctxData["ally_count"] = 1
			ctxData["choice_type"] = "crk_bloody_prayer_target"
		}
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return nil
	}
	if choiceType == "crk_bloody_prayer_ally_count" {
		x := 0
		if v, ok := ctxData["x_value"].(int); ok {
			x = v
		} else if f, ok := ctxData["x_value"].(float64); ok {
			x = int(f)
		}
		var allyIDs []string
		if arr, ok := ctxData["ally_ids"].([]string); ok {
			allyIDs = arr
		} else if arr, ok := ctxData["ally_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					allyIDs = append(allyIDs, s)
				}
			}
		}
		maxCount := 1
		if len(allyIDs) >= 2 && x >= 2 {
			maxCount = 2
		}
		allyCount := selectionIndex + 1
		if allyCount < 1 || allyCount > maxCount {
			return fmt.Errorf("无效的队友数量选择")
		}
		ctxData["ally_count"] = allyCount
		ctxData["selected_ally_ids"] = []string{}
		ctxData["choice_type"] = "crk_bloody_prayer_target"
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return nil
	}
	if choiceType == "crk_bloody_prayer_target" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		var allyIDs []string
		if arr, ok := ctxData["ally_ids"].([]string); ok {
			allyIDs = arr
		} else if arr, ok := ctxData["ally_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					allyIDs = append(allyIDs, s)
				}
			}
		}
		selected := make([]string, 0, 2)
		selectedSet := map[string]bool{}
		if arr, ok := ctxData["selected_ally_ids"].([]string); ok {
			for _, s := range arr {
				if s == "" || selectedSet[s] {
					continue
				}
				selected = append(selected, s)
				selectedSet[s] = true
			}
		} else if arr, ok := ctxData["selected_ally_ids"].([]interface{}); ok {
			for _, v := range arr {
				s, ok := v.(string)
				if !ok || s == "" || selectedSet[s] {
					continue
				}
				selected = append(selected, s)
				selectedSet[s] = true
			}
		}
		remaining := make([]string, 0, len(allyIDs))
		for _, aid := range allyIDs {
			if !selectedSet[aid] {
				remaining = append(remaining, aid)
			}
		}
		if selectionIndex < 0 || selectionIndex >= len(remaining) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		allyCount := 1
		if v, ok := ctxData["ally_count"].(int); ok && v > 0 {
			allyCount = v
		} else if f, ok := ctxData["ally_count"].(float64); ok && int(f) > 0 {
			allyCount = int(f)
		}
		chosenID := remaining[selectionIndex]
		selected = append(selected, chosenID)
		ctxData["selected_ally_ids"] = selected

		x := 0
		if v, ok := ctxData["x_value"].(int); ok {
			x = v
		} else if f, ok := ctxData["x_value"].(float64); ok {
			x = int(f)
		}
		if x <= 0 || user.Heal < x {
			return fmt.Errorf("治疗不足，无法结算血腥祷言")
		}

		// 目标还没选满：继续选择下一名队友。
		if len(selected) < allyCount {
			ctxData["choice_type"] = "crk_bloody_prayer_target"
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return nil
		}

		// 单目标：全部治疗给该队友；双目标：进入分配步骤。
		if allyCount <= 1 {
			alloc := map[string]int{selected[0]: x}
			if err := e.resolveCrimsonKnightBloodyPrayer(user, x, alloc); err != nil {
				return err
			}
		} else {
			if x < 2 {
				return fmt.Errorf("X不足以分配给2名队友")
			}
			ctxData["choice_type"] = "crk_bloody_prayer_split"
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return nil
		}

		e.PopInterrupt()
		if e.State.PendingInterrupt == nil && len(e.State.PendingDamageQueue) > 0 {
			e.State.Phase = model.PhasePendingDamageResolution
		}
		return nil
	}
	if choiceType == "crk_bloody_prayer_split" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		x := 0
		if v, ok := ctxData["x_value"].(int); ok {
			x = v
		} else if f, ok := ctxData["x_value"].(float64); ok {
			x = int(f)
		}
		if x < 2 || user.Heal < x {
			return fmt.Errorf("治疗不足，无法结算血腥祷言")
		}
		var selected []string
		if arr, ok := ctxData["selected_ally_ids"].([]string); ok {
			selected = append(selected, arr...)
		} else if arr, ok := ctxData["selected_ally_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					selected = append(selected, s)
				}
			}
		}
		if len(selected) != 2 {
			return fmt.Errorf("血腥祷言分配目标数量异常")
		}
		if selectionIndex < 0 || selectionIndex >= x-1 {
			return fmt.Errorf("无效的分配选项")
		}
		firstHeal := selectionIndex + 1
		secondHeal := x - firstHeal
		alloc := map[string]int{
			selected[0]: firstHeal,
			selected[1]: secondHeal,
		}
		if err := e.resolveCrimsonKnightBloodyPrayer(user, x, alloc); err != nil {
			return err
		}
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil && len(e.State.PendingDamageQueue) > 0 {
			e.State.Phase = model.PhasePendingDamageResolution
		}
		return nil
	}
	if choiceType == "crk_calm_mind_action" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		switch selectionIndex {
		case 0:
			user.TurnState.PendingActions = append(user.TurnState.PendingActions, model.ActionContext{
				Source:   "戒骄戒躁",
				MustType: "Attack",
			})
			e.Log(fmt.Sprintf("%s 的 [戒骄戒躁] 生效：额外获得1次攻击行动", user.Name))
		case 1:
			user.TurnState.PendingActions = append(user.TurnState.PendingActions, model.ActionContext{
				Source:   "戒骄戒躁",
				MustType: "Magic",
			})
			e.Log(fmt.Sprintf("%s 的 [戒骄戒躁] 生效：额外获得1次法术行动", user.Name))
		default:
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			e.State.Phase = model.PhaseExtraAction
		}
		return nil
	}
	if choiceType == "hom_rune_reforge_distribution" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		total := 3
		if v, ok := ctxData["total_runes"].(int); ok && v > 0 {
			total = v
		} else if f, ok := ctxData["total_runes"].(float64); ok && int(f) > 0 {
			total = int(f)
		}
		warRunes := selectionIndex
		if warRunes < 0 || warRunes > total {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		if user.Tokens == nil {
			user.Tokens = map[string]int{}
		}
		user.Tokens["hom_war_rune"] = warRunes
		user.Tokens["hom_magic_rune"] = total - warRunes
		e.Log(fmt.Sprintf("%s 的 [符文改造]：战纹=%d，魔纹=%d", user.Name, user.Tokens["hom_war_rune"], user.Tokens["hom_magic_rune"]))
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			e.State.Phase = model.PhaseStartup
		}
		return nil
	}
	if choiceType == "onmyoji_life_barrier_mode" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		if selectionIndex == 0 {
			var targetIDs []string
			if arr, ok := ctxData["support_target_ids"].([]string); ok {
				targetIDs = arr
			} else if arr, ok := ctxData["support_target_ids"].([]interface{}); ok {
				for _, v := range arr {
					if s, ok := v.(string); ok {
						targetIDs = append(targetIDs, s)
					}
				}
			}
			if len(targetIDs) == 0 {
				return fmt.Errorf("生命结界分支①没有可选队友")
			}
			ctxData["choice_type"] = "onmyoji_life_barrier_support_target"
			ctxData["target_ids"] = targetIDs
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return nil
		}
		if selectionIndex == 1 {
			if user.Tokens == nil || user.Tokens["onmyoji_form"] <= 0 {
				return fmt.Errorf("不在式神形态，无法选择生命结界分支②")
			}
			var combos []string
			if arr, ok := ctxData["release_card_combos"].([]string); ok {
				combos = append(combos, arr...)
			} else if arr, ok := ctxData["release_card_combos"].([]interface{}); ok {
				for _, v := range arr {
					if s, ok := v.(string); ok {
						combos = append(combos, s)
					}
				}
			}
			if len(combos) == 0 {
				return fmt.Errorf("分支②需要弃2张同命格手牌")
			}
			var targetIDs []string
			if arr, ok := ctxData["release_target_ids"].([]string); ok {
				targetIDs = arr
			} else if arr, ok := ctxData["release_target_ids"].([]interface{}); ok {
				for _, v := range arr {
					if s, ok := v.(string); ok {
						targetIDs = append(targetIDs, s)
					}
				}
			}
			if len(targetIDs) == 0 {
				return fmt.Errorf("分支②没有可选队友目标")
			}
			ctxData["choice_type"] = "onmyoji_life_barrier_release_combo"
			ctxData["target_ids"] = targetIDs
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return nil
		}
		return fmt.Errorf("无效的生命结界分支选择")
	}
	if choiceType == "onmyoji_life_barrier_release_combo" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		var combos []string
		if arr, ok := ctxData["release_card_combos"].([]string); ok {
			combos = append(combos, arr...)
		} else if arr, ok := ctxData["release_card_combos"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					combos = append(combos, s)
				}
			}
		}
		if selectionIndex < 0 || selectionIndex >= len(combos) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		parts := strings.Split(combos[selectionIndex], ",")
		if len(parts) != 2 {
			return fmt.Errorf("无效的弃牌组合")
		}
		i, errI := strconv.Atoi(parts[0])
		j, errJ := strconv.Atoi(parts[1])
		if errI != nil || errJ != nil {
			return fmt.Errorf("无效的弃牌组合索引")
		}
		if i < 0 || j < 0 || i >= len(user.Hand) || j >= len(user.Hand) || i == j {
			return fmt.Errorf("弃牌组合越界")
		}
		c1 := user.Hand[i]
		c2 := user.Hand[j]
		if c1.Faction == "" || c2.Faction == "" || c1.Faction != c2.Faction {
			return fmt.Errorf("分支②需要弃2张同命格手牌")
		}
		e.NotifyCardRevealed(user.ID, []model.Card{c1, c2}, "discard")
		if i < j {
			i, j = j, i
		}
		user.Hand = append(user.Hand[:i], user.Hand[i+1:]...)
		user.Hand = append(user.Hand[:j], user.Hand[j+1:]...)
		e.State.DiscardPile = append(e.State.DiscardPile, c1, c2)
		user.Tokens["onmyoji_form"] = 0
		e.Log(fmt.Sprintf("%s 的 [生命结界] 分支②：弃2张同命格手牌并脱离式神形态", user.Name))
		ctxData["choice_type"] = "onmyoji_life_barrier_release_target"
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return nil
	}
	if choiceType == "onmyoji_life_barrier_support_target" || choiceType == "onmyoji_life_barrier_release_target" || choiceType == "onmyoji_dark_ritual_target" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		var targetIDs []string
		if arr, ok := ctxData["target_ids"].([]string); ok {
			targetIDs = arr
		} else if arr, ok := ctxData["target_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					targetIDs = append(targetIDs, s)
				}
			}
		}
		if selectionIndex < 0 || selectionIndex >= len(targetIDs) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		targetID := targetIDs[selectionIndex]
		target := e.State.Players[targetID]
		if target == nil {
			return fmt.Errorf("目标不存在")
		}
		switch choiceType {
		case "onmyoji_life_barrier_support_target":
			ghostFire := 0
			if v, ok := ctxData["ghost_fire"].(int); ok {
				ghostFire = v
			} else if f, ok := ctxData["ghost_fire"].(float64); ok {
				ghostFire = int(f)
			}
			target.Gem++
			e.Heal(targetID, 1)
			if ghostFire > 0 {
				damageType := "magic"
				if ghostFire >= 3 {
					damageType = "magic_no_morale"
				}
				e.AddPendingDamage(model.PendingDamage{
					SourceID:   user.ID,
					TargetID:   user.ID,
					Damage:     ghostFire,
					DamageType: damageType,
					Stage:      0,
				})
			}
			e.Log(fmt.Sprintf("%s 的 [生命结界] 分支①生效：%s +1宝石+1治疗，自身承受%d点法术伤害", user.Name, target.Name, ghostFire))
			e.PopInterrupt()
			if e.State.PendingInterrupt == nil {
				if len(e.State.PendingDamageQueue) > 0 {
					e.State.Phase = model.PhasePendingDamageResolution
					e.State.ReturnPhase = model.PhaseTurnEnd
				} else {
					e.State.Phase = model.PhaseTurnEnd
				}
			}
			return nil
		case "onmyoji_life_barrier_release_target":
			if target.Camp != user.Camp || target.ID == user.ID {
				return fmt.Errorf("分支②目标必须是其他队友")
			}
			if len(target.Hand) == 0 {
				return fmt.Errorf("目标队友没有手牌可弃置")
			}
			e.PushInterrupt(&model.Interrupt{
				Type:     model.InterruptDiscard,
				PlayerID: target.ID,
				Context: map[string]interface{}{
					"discard_count": 1,
					"prompt":        fmt.Sprintf("【生命结界】请弃置1张手牌（由 %s 指定）", user.Name),
				},
			})
			e.Log(fmt.Sprintf("%s 的 [生命结界] 分支②生效：指定 %s 弃置1张手牌", user.Name, target.Name))
			e.PopInterrupt()
			return nil
		case "onmyoji_dark_ritual_target":
			ghostFire := 0
			if v, ok := ctxData["ghost_fire"].(int); ok {
				ghostFire = v
			} else if f, ok := ctxData["ghost_fire"].(float64); ok {
				ghostFire = int(f)
			}
			user.Tokens["onmyoji_ghost_fire"] = 0
			e.AddPendingDamage(model.PendingDamage{
				SourceID:   user.ID,
				TargetID:   targetID,
				Damage:     2,
				DamageType: "magic",
				Stage:      0,
			})
			e.Log(fmt.Sprintf("%s 的 [黑暗祭礼] 生效：移除%d点鬼火，对 %s 造成2点法术伤害", user.Name, ghostFire, target.Name))
			e.PopInterrupt()
			if e.State.PendingInterrupt == nil {
				e.State.Phase = model.PhasePendingDamageResolution
				e.State.ReturnPhase = model.PhaseTurnEnd
			}
			return nil
		}
	}
	if choiceType == "onmyoji_binding_confirm" {
		if selectionIndex != 0 && selectionIndex != 1 {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		if selectionIndex == 1 {
			e.PopInterrupt()
			if e.State.PendingInterrupt == nil {
				e.State.Phase = model.PhaseCombatInteraction
			}
			return nil
		}
		var cardOptions []map[string]interface{}
		if arr, ok := ctxData["card_options"].([]map[string]interface{}); ok {
			cardOptions = append(cardOptions, arr...)
		} else if arr, ok := ctxData["card_options"].([]interface{}); ok {
			for _, item := range arr {
				if m, ok := item.(map[string]interface{}); ok && m != nil {
					cardOptions = append(cardOptions, m)
				}
			}
		}
		if len(cardOptions) == 0 {
			e.PopInterrupt()
			if e.State.PendingInterrupt == nil {
				e.State.Phase = model.PhaseCombatInteraction
			}
			return nil
		}
		ctxData["choice_type"] = "onmyoji_binding_card"
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return nil
	}
	if choiceType == "onmyoji_yinyang_confirm" {
		if selectionIndex != 0 && selectionIndex != 1 {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		if selectionIndex == 1 {
			e.PopInterrupt()
			if e.State.PendingInterrupt == nil {
				e.State.Phase = model.PhaseCombatInteraction
			}
			return nil
		}
		var cardOptions []map[string]interface{}
		if arr, ok := ctxData["card_options"].([]map[string]interface{}); ok {
			cardOptions = append(cardOptions, arr...)
		} else if arr, ok := ctxData["card_options"].([]interface{}); ok {
			for _, item := range arr {
				if m, ok := item.(map[string]interface{}); ok && m != nil {
					cardOptions = append(cardOptions, m)
				}
			}
		}
		if len(cardOptions) == 0 {
			e.PopInterrupt()
			if e.State.PendingInterrupt == nil {
				e.State.Phase = model.PhaseCombatInteraction
			}
			return nil
		}
		ctxData["choice_type"] = "onmyoji_yinyang_card"
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return nil
	}
	if choiceType == "onmyoji_yinyang_card" {
		var cardOptions []map[string]interface{}
		if arr, ok := ctxData["card_options"].([]map[string]interface{}); ok {
			cardOptions = append(cardOptions, arr...)
		} else if arr, ok := ctxData["card_options"].([]interface{}); ok {
			for _, item := range arr {
				if m, ok := item.(map[string]interface{}); ok && m != nil {
					cardOptions = append(cardOptions, m)
				}
			}
		}
		if selectionIndex < 0 || selectionIndex >= len(cardOptions) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		cardID, _ := cardOptions[selectionIndex]["card_id"].(string)
		if cardID == "" {
			return fmt.Errorf("无效的应战卡牌")
		}
		ctxData["selected_card_id"] = cardID
		ctxData["choice_type"] = "onmyoji_yinyang_counter_target"
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return nil
	}
	if choiceType == "onmyoji_yinyang_counter_target" {
		var counterTargets []string
		if arr, ok := ctxData["counter_target_ids"].([]string); ok {
			counterTargets = arr
		} else if arr, ok := ctxData["counter_target_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					counterTargets = append(counterTargets, s)
				}
			}
		}
		if selectionIndex < 0 || selectionIndex >= len(counterTargets) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		actorID, _ := ctxData["actor_id"].(string)
		cardID, _ := ctxData["selected_card_id"].(string)
		if actorID == "" || cardID == "" {
			return fmt.Errorf("阴阳转换上下文缺失")
		}
		actor := e.State.Players[actorID]
		if actor == nil {
			return fmt.Errorf("阴阳师不存在")
		}
		cardIdx := findPlayableCardIndexByID(actor, cardID)
		if cardIdx < 0 {
			return fmt.Errorf("应战牌已不存在")
		}
		targetID := counterTargets[selectionIndex]
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			e.State.Phase = model.PhaseCombatInteraction
		}
		return e.handleCombatResponse(model.PlayerAction{
			PlayerID:  actorID,
			Type:      model.CmdRespond,
			ExtraArgs: []string{"counter"},
			CardIndex: cardIdx,
			TargetID:  targetID,
		})
	}
	if choiceType == "onmyoji_binding_card" {
		var cardOptions []map[string]interface{}
		if arr, ok := ctxData["card_options"].([]map[string]interface{}); ok {
			cardOptions = append(cardOptions, arr...)
		} else if arr, ok := ctxData["card_options"].([]interface{}); ok {
			for _, item := range arr {
				if m, ok := item.(map[string]interface{}); ok && m != nil {
					cardOptions = append(cardOptions, m)
				}
			}
		}
		if selectionIndex < 0 || selectionIndex >= len(cardOptions) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		cardID, _ := cardOptions[selectionIndex]["card_id"].(string)
		useFaction, _ := cardOptions[selectionIndex]["use_faction"].(bool)
		if cardID == "" {
			return fmt.Errorf("无效的应战卡牌")
		}
		ctxData["selected_card_id"] = cardID
		ctxData["selected_use_faction"] = useFaction
		ctxData["choice_type"] = "onmyoji_binding_counter_target"
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return nil
	}
	if choiceType == "onmyoji_binding_counter_target" {
		var counterTargets []string
		if arr, ok := ctxData["counter_target_ids"].([]string); ok {
			counterTargets = arr
		} else if arr, ok := ctxData["counter_target_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					counterTargets = append(counterTargets, s)
				}
			}
		}
		if selectionIndex < 0 || selectionIndex >= len(counterTargets) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		actorID, _ := ctxData["actor_id"].(string)
		cardID, _ := ctxData["selected_card_id"].(string)
		if actorID == "" || cardID == "" {
			return fmt.Errorf("式神咒束上下文缺失")
		}
		actor := e.State.Players[actorID]
		if actor == nil {
			return fmt.Errorf("阴阳师不存在")
		}
		if len(e.State.CombatStack) == 0 {
			return fmt.Errorf("当前没有可代应战的战斗请求")
		}
		combatReq := &e.State.CombatStack[len(e.State.CombatStack)-1]
		if !e.payOnmyojiBindingCost(actor.Camp) {
			return fmt.Errorf("战绩区资源不足，无法发动式神咒束")
		}
		useFaction, _ := ctxData["selected_use_faction"].(bool)
		combatReq.OnmyojiBindingChecked = true
		combatReq.OnmyojiBindingActorID = actorID
		combatReq.OnmyojiBindingCounterID = cardID
		combatReq.OnmyojiBindingTargetID = counterTargets[selectionIndex]
		combatReq.OnmyojiBindingUseFaction = useFaction
		// 伤害承担重定向：后续“攻击未命中”流程中的伤害承担者应变更为阴阳师。
		combatReq.TargetID = actorID

		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			e.State.Phase = model.PhaseCombatInteraction
		}
		return nil
	}
	if choiceType == "hom_rune_smash_x" || choiceType == "hom_glyph_fusion_x" {
		maxX := 0
		if v, ok := ctxData["max_x"].(int); ok {
			maxX = v
		} else if f, ok := ctxData["max_x"].(float64); ok {
			maxX = int(f)
		}
		minX := 1
		nextChoice := "hom_rune_smash_cards"
		if choiceType == "hom_glyph_fusion_x" {
			minX = 2
			nextChoice = "hom_glyph_fusion_cards"
		}
		xVal := selectionIndex
		if xVal < minX || xVal > maxX {
			xVal = selectionIndex + minX
		}
		if xVal < minX || xVal > maxX {
			return fmt.Errorf("无效的X值")
		}
		var candidates []int
		if arr, ok := ctxData["candidate_indices"].([]int); ok {
			candidates = append(candidates, arr...)
		} else if arr, ok := ctxData["candidate_indices"].([]interface{}); ok {
			for _, v := range arr {
				if f, ok := v.(float64); ok {
					candidates = append(candidates, int(f))
				}
			}
		}
		if xVal > len(candidates) {
			return fmt.Errorf("可选牌数量不足")
		}
		ctxData["choice_type"] = nextChoice
		ctxData["x_value"] = xVal
		ctxData["selected_indices"] = []int{}
		ctxData["remaining_indices"] = append([]int{}, candidates...)
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return nil
	}
	if choiceType == "hom_rune_smash_cards" || choiceType == "hom_glyph_fusion_cards" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		var remaining []int
		if arr, ok := ctxData["remaining_indices"].([]int); ok {
			remaining = arr
		} else if arr, ok := ctxData["remaining_indices"].([]interface{}); ok {
			for _, v := range arr {
				if f, ok := v.(float64); ok {
					remaining = append(remaining, int(f))
				}
			}
		}
		var selected []int
		if arr, ok := ctxData["selected_indices"].([]int); ok {
			selected = arr
		} else if arr, ok := ctxData["selected_indices"].([]interface{}); ok {
			for _, v := range arr {
				if f, ok := v.(float64); ok {
					selected = append(selected, int(f))
				}
			}
		}
		xVal := 0
		if v, ok := ctxData["x_value"].(int); ok {
			xVal = v
		} else if f, ok := ctxData["x_value"].(float64); ok {
			xVal = int(f)
		}
		cardIdx, ok := resolveSelectionToCandidate(selectionIndex, remaining)
		if !ok {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		if cardIdx < 0 || cardIdx >= len(user.Hand) {
			return fmt.Errorf("无效的手牌索引: %d", cardIdx)
		}
		attackElement, _ := ctxData["attack_element"].(string)
		if choiceType == "hom_rune_smash_cards" && attackElement != "" && string(user.Hand[cardIdx].Element) != attackElement {
			return fmt.Errorf("战纹碎击需选择与攻击同系的牌")
		}
		if choiceType == "hom_glyph_fusion_cards" && attackElement != "" && string(user.Hand[cardIdx].Element) == attackElement {
			return fmt.Errorf("魔纹融合需选择与攻击异系的牌")
		}
		if choiceType == "hom_glyph_fusion_cards" {
			for _, idx := range selected {
				if idx >= 0 && idx < len(user.Hand) && user.Hand[idx].Element == user.Hand[cardIdx].Element {
					return fmt.Errorf("魔纹融合需选择元素互不相同的异系牌")
				}
			}
		}
		selected = append(selected, cardIdx)
		var nextRemaining []int
		for _, v := range remaining {
			if v == cardIdx {
				continue
			}
			if choiceType == "hom_glyph_fusion_cards" {
				if v >= 0 && v < len(user.Hand) && user.Hand[v].Element == user.Hand[cardIdx].Element {
					continue
				}
			}
			nextRemaining = append(nextRemaining, v)
		}
		if len(selected) < xVal {
			ctxData["selected_indices"] = selected
			ctxData["remaining_indices"] = nextRemaining
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return nil
		}
		ctxData["selected_indices"] = selected
		maxY := 0
		if v, ok := ctxData["max_y"].(int); ok {
			maxY = v
		} else if f, ok := ctxData["max_y"].(float64); ok {
			maxY = int(f)
		}
		if maxY > 0 {
			if choiceType == "hom_rune_smash_cards" {
				ctxData["choice_type"] = "hom_rune_smash_y"
			} else {
				ctxData["choice_type"] = "hom_glyph_fusion_y"
			}
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return nil
		}
		ctxData["y_value"] = 0
		return e.resolveHomunculusRuneChoice(ctxData, choiceType == "hom_glyph_fusion_cards")
	}
	if choiceType == "hom_rune_smash_y" || choiceType == "hom_glyph_fusion_y" {
		maxY := 0
		if v, ok := ctxData["max_y"].(int); ok {
			maxY = v
		} else if f, ok := ctxData["max_y"].(float64); ok {
			maxY = int(f)
		}
		yVal := selectionIndex
		if yVal < 0 || yVal > maxY {
			return fmt.Errorf("无效的Y值")
		}
		ctxData["y_value"] = yVal
		return e.resolveHomunculusRuneChoice(ctxData, choiceType == "hom_glyph_fusion_y")
	}
	if choiceType == "elf_elemental_shot_cost" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		canMagic, _ := ctxData["can_discard_magic"].(bool)
		canBless, _ := ctxData["can_remove_bless"].(bool)
		var modeList []int
		if canMagic {
			modeList = append(modeList, 0)
		}
		if canBless {
			modeList = append(modeList, 1)
		}
		if selectionIndex < 0 || selectionIndex >= len(modeList) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		mode := modeList[selectionIndex]
		if mode == 0 {
			ctxData["choice_type"] = "elf_elemental_shot_discard_magic"
			ctxData["magic_indices"] = getCardIndicesByType(user, model.CardTypeMagic)
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return nil
		}
		ctxData["choice_type"] = "elf_elemental_shot_remove_blessing"
		ctxData["blessing_indices"] = elfBlessingHandIndices(user)
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return nil
	}
	if choiceType == "elf_elemental_shot_discard_magic" || choiceType == "elf_elemental_shot_remove_blessing" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		var candidates []int
		key := "magic_indices"
		if choiceType == "elf_elemental_shot_remove_blessing" {
			key = "blessing_indices"
		}
		if arr, ok := ctxData[key].([]int); ok {
			candidates = arr
		} else if arr, ok := ctxData[key].([]interface{}); ok {
			for _, v := range arr {
				if f, ok := v.(float64); ok {
					candidates = append(candidates, int(f))
				}
			}
		}
		cardIdx, ok := resolveSelectionToCandidate(selectionIndex, candidates)
		if !ok {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		var card model.Card
		if choiceType == "elf_elemental_shot_remove_blessing" {
			if cardIdx < 0 || cardIdx >= len(user.Blessings) {
				return fmt.Errorf("无效的祝福索引: %d", selectionIndex)
			}
			card = user.Blessings[cardIdx]
			removeElfBlessingByCardID(user, card.ID)
		} else {
			if cardIdx < 0 || cardIdx >= len(user.Hand) {
				return fmt.Errorf("无效的手牌索引: %d", selectionIndex)
			}
			card = user.Hand[cardIdx]
			user.Hand = append(user.Hand[:cardIdx], user.Hand[cardIdx+1:]...)
		}
		e.NotifyCardRevealed(user.ID, []model.Card{card}, "discard")
		e.State.DiscardPile = append(e.State.DiscardPile, card)

		if user.Tokens == nil {
			user.Tokens = map[string]int{}
		}
		user.Tokens["elf_elemental_shot_fire_pending"] = 0
		user.Tokens["elf_elemental_shot_water_pending"] = 0
		user.Tokens["elf_elemental_shot_earth_pending"] = 0
		user.Tokens["elf_elemental_shot_thunder_pending"] = 0
		rawCtx, _ := ctxData["user_ctx"].(*model.Context)
		attackElement, _ := ctxData["attack_element"].(string)
		switch attackElement {
		case string(model.ElementFire):
			user.Tokens["elf_elemental_shot_fire_pending"] = 1
		case string(model.ElementWater):
			user.Tokens["elf_elemental_shot_water_pending"] = 1
		case string(model.ElementWind):
			user.TurnState.PendingActions = append(user.TurnState.PendingActions, model.ActionContext{
				Source:   "风之矢",
				MustType: "Attack",
			})
		case string(model.ElementThunder):
			user.Tokens["elf_elemental_shot_thunder_pending"] = 1
			if rawCtx != nil && rawCtx.TriggerCtx != nil && rawCtx.TriggerCtx.AttackInfo != nil {
				rawCtx.TriggerCtx.AttackInfo.CanBeResponded = false
			}
		case string(model.ElementEarth):
			user.Tokens["elf_elemental_shot_earth_pending"] = 1
		}
		e.Log(fmt.Sprintf("%s 发动 [元素射击]（%s）", user.Name, attackElement))
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			if len(e.State.ActionQueue) > 0 {
				e.State.Phase = model.PhaseBeforeAction
			} else if len(e.State.PendingDamageQueue) > 0 {
				e.State.Phase = model.PhasePendingDamageResolution
			}
		}
		return nil
	}
	if choiceType == "elf_animal_companion_confirm" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		if selectionIndex == 1 {
			e.PopInterrupt()
			if e.State.PendingInterrupt == nil && len(e.State.PendingDamageQueue) > 0 {
				e.State.Phase = model.PhasePendingDamageResolution
			}
			return nil
		}
		if selectionIndex != 0 {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		if e.CanPayCrystalCost(user.ID, 1) {
			e.State.PendingInterrupt.Context = map[string]interface{}{
				"choice_type": "elf_pet_empower_confirm",
				"user_id":     userID,
			}
			e.notifyInterruptPrompt()
			return nil
		}
		e.DrawCards(user.ID, 1)
		e.PushInterrupt(&model.Interrupt{
			Type:     model.InterruptDiscard,
			PlayerID: user.ID,
			Context: map[string]interface{}{
				"discard_count":     1,
				"stay_in_turn":      true,
				"prompt":            "【动物伙伴】请选择弃置1张牌：",
				"exclude_blessings": true,
			},
		})
		e.PopInterrupt()
		return nil
	}
	if choiceType == "elf_pet_empower_confirm" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		if selectionIndex == 0 && e.ConsumeCrystalCost(user.ID, 1) {
			e.State.PendingInterrupt.Context = map[string]interface{}{
				"choice_type": "elf_pet_empower_target",
				"user_id":     userID,
				"target_ids":  append([]string{}, e.State.PlayerOrder...),
			}
			e.notifyInterruptPrompt()
			return nil
		}
		// 否则按普通动物伙伴结算
		e.DrawCards(user.ID, 1)
		e.PushInterrupt(&model.Interrupt{
			Type:     model.InterruptDiscard,
			PlayerID: user.ID,
			Context: map[string]interface{}{
				"discard_count":     1,
				"stay_in_turn":      true,
				"prompt":            "【动物伙伴】请选择弃置1张牌：",
				"exclude_blessings": true,
			},
		})
		e.PopInterrupt()
		return nil
	}
	if choiceType == "elf_pet_empower_target" {
		var targetIDs []string
		if arr, ok := ctxData["target_ids"].([]string); ok {
			targetIDs = arr
		} else if arr, ok := ctxData["target_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					targetIDs = append(targetIDs, s)
				}
			}
		}
		if selectionIndex < 0 || selectionIndex >= len(targetIDs) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		target := e.State.Players[targetIDs[selectionIndex]]
		if target == nil {
			return fmt.Errorf("目标不存在")
		}
		e.DrawCards(target.ID, 1)
		// 若摸牌已触发爆牌弃牌（手牌仍超上限），则本次“摸1弃1”中的“弃1”由该爆牌弃牌结算承担，
		// 不再额外追加一次弃牌中断，避免出现“连续弃两次”。
		if len(target.Hand) > e.GetMaxHand(target) {
			e.Log(fmt.Sprintf("[宠物强化] %s 摸牌后触发爆牌，本次弃1由爆牌弃牌结算承担", target.Name))
			e.PopInterrupt()
			return nil
		}
		e.PushInterrupt(&model.Interrupt{
			Type:     model.InterruptDiscard,
			PlayerID: target.ID,
			Context: map[string]interface{}{
				"discard_count":     1,
				"stay_in_turn":      true,
				"prompt":            fmt.Sprintf("【宠物强化】%s 请弃置1张牌：", target.Name),
				"exclude_blessings": e.isElfArcher(target),
			},
		})
		e.PopInterrupt()
		return nil
	}
	if choiceType == "bw_witch_wrath_draw" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		if selectionIndex < 0 || selectionIndex > 2 {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		if selectionIndex > 0 {
			e.DrawCards(user.ID, selectionIndex)
		}
		e.Log(fmt.Sprintf("%s 的 [魔女之怒]：选择摸%d张牌", user.Name, selectionIndex))
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			e.State.Phase = model.PhaseStartup
		}
		return nil
	}
	if choiceType == "bw_substitute_doll_card" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		var magicIndices []int
		if arr, ok := ctxData["magic_indices"].([]int); ok {
			magicIndices = append(magicIndices, arr...)
		} else if arr, ok := ctxData["magic_indices"].([]interface{}); ok {
			for _, v := range arr {
				if f, ok := v.(float64); ok {
					magicIndices = append(magicIndices, int(f))
				}
			}
		}
		cardIdx, ok := resolveSelectionToCandidate(selectionIndex, magicIndices)
		if !ok || cardIdx < 0 || cardIdx >= len(user.Hand) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		if user.Hand[cardIdx].Type != model.CardTypeMagic {
			return fmt.Errorf("替身玩偶需弃置法术牌")
		}
		ctxData["selected_card_index"] = cardIdx
		ctxData["choice_type"] = "bw_substitute_doll_target"
		if arr, ok := ctxData["ally_ids"].([]string); ok {
			ctxData["target_ids"] = append([]string{}, arr...)
		} else if arr, ok := ctxData["ally_ids"].([]interface{}); ok {
			var targetIDs []string
			for _, v := range arr {
				if s, ok := v.(string); ok {
					targetIDs = append(targetIDs, s)
				}
			}
			ctxData["target_ids"] = targetIDs
		}
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return nil
	}
	if choiceType == "bw_mana_inversion_x" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		maxX := toIntContextValue(ctxData["max_x"])
		xValue := selectionIndex + 2
		if xValue < 2 || xValue > maxX {
			return fmt.Errorf("无效的X值")
		}
		var magicIndices []int
		for i, c := range user.Hand {
			if c.Type == model.CardTypeMagic {
				magicIndices = append(magicIndices, i)
			}
		}
		if len(magicIndices) < xValue {
			return fmt.Errorf("法术牌不足，无法弃置X=%d张", xValue)
		}
		ctxData["choice_type"] = "bw_mana_inversion_cards"
		ctxData["x_value"] = xValue
		ctxData["selected_indices"] = []int{}
		ctxData["remaining_indices"] = magicIndices
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return nil
	}
	if choiceType == "bw_mana_inversion_cards" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		var remaining []int
		if arr, ok := ctxData["remaining_indices"].([]int); ok {
			remaining = append(remaining, arr...)
		} else if arr, ok := ctxData["remaining_indices"].([]interface{}); ok {
			for _, v := range arr {
				if f, ok := v.(float64); ok {
					remaining = append(remaining, int(f))
				}
			}
		}
		var selected []int
		if arr, ok := ctxData["selected_indices"].([]int); ok {
			selected = append(selected, arr...)
		} else if arr, ok := ctxData["selected_indices"].([]interface{}); ok {
			for _, v := range arr {
				if f, ok := v.(float64); ok {
					selected = append(selected, int(f))
				}
			}
		}
		xValue := toIntContextValue(ctxData["x_value"])
		cardIdx, ok := resolveSelectionToCandidate(selectionIndex, remaining)
		if !ok || cardIdx < 0 || cardIdx >= len(user.Hand) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		if user.Hand[cardIdx].Type != model.CardTypeMagic {
			return fmt.Errorf("魔能反转需弃置法术牌")
		}
		selected = append(selected, cardIdx)
		var nextRemaining []int
		for _, v := range remaining {
			if v != cardIdx {
				nextRemaining = append(nextRemaining, v)
			}
		}
		if len(selected) < xValue {
			ctxData["selected_indices"] = selected
			ctxData["remaining_indices"] = nextRemaining
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return nil
		}
		var enemyIDs []string
		for _, p := range e.State.PlayerOrder {
			target := e.State.Players[p]
			if target == nil || target.Camp == user.Camp {
				continue
			}
			enemyIDs = append(enemyIDs, target.ID)
		}
		if len(enemyIDs) == 0 {
			return fmt.Errorf("无可选敌方目标")
		}
		ctxData["selected_indices"] = selected
		ctxData["choice_type"] = "bw_mana_inversion_target"
		ctxData["target_ids"] = enemyIDs
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return nil
	}
	if choiceType == "bw_substitute_doll_target" || choiceType == "bw_mana_inversion_target" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		var targetIDs []string
		if arr, ok := ctxData["target_ids"].([]string); ok {
			targetIDs = append(targetIDs, arr...)
		} else if arr, ok := ctxData["target_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					targetIDs = append(targetIDs, s)
				}
			}
		}
		if selectionIndex < 0 || selectionIndex >= len(targetIDs) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		targetID := targetIDs[selectionIndex]
		target := e.State.Players[targetID]
		if target == nil {
			return fmt.Errorf("目标不存在")
		}
		switch choiceType {
		case "bw_substitute_doll_target":
			cardIdx := toIntContextValue(ctxData["selected_card_index"])
			if cardIdx < 0 || cardIdx >= len(user.Hand) {
				return fmt.Errorf("无效的弃牌索引")
			}
			if user.Hand[cardIdx].Type != model.CardTypeMagic {
				return fmt.Errorf("替身玩偶需弃置法术牌")
			}
			card := user.Hand[cardIdx]
			user.Hand = append(user.Hand[:cardIdx], user.Hand[cardIdx+1:]...)
			e.NotifyCardRevealed(user.ID, []model.Card{card}, "discard")
			e.State.DiscardPile = append(e.State.DiscardPile, card)
			e.DrawCards(targetID, 1)
			e.Log(fmt.Sprintf("%s 的 [替身玩偶] 生效：%s 摸1张牌", user.Name, target.Name))
		case "bw_mana_inversion_target":
			var selected []int
			if arr, ok := ctxData["selected_indices"].([]int); ok {
				selected = append(selected, arr...)
			} else if arr, ok := ctxData["selected_indices"].([]interface{}); ok {
				for _, v := range arr {
					if f, ok := v.(float64); ok {
						selected = append(selected, int(f))
					}
				}
			}
			xValue := toIntContextValue(ctxData["x_value"])
			if xValue < 2 || len(selected) != xValue {
				return fmt.Errorf("魔能反转弃牌参数错误")
			}
			for _, idx := range selected {
				if idx < 0 || idx >= len(user.Hand) || user.Hand[idx].Type != model.CardTypeMagic {
					return fmt.Errorf("魔能反转弃牌必须为法术牌")
				}
			}
			removed, err := removeCardsByIndicesFromHand(user, append([]int{}, selected...))
			if err != nil {
				return err
			}
			e.NotifyCardRevealed(user.ID, removed, "discard")
			e.State.DiscardPile = append(e.State.DiscardPile, removed...)
			damage := xValue - 1
			if damage > 0 {
				e.AddPendingDamage(model.PendingDamage{
					SourceID:   user.ID,
					TargetID:   targetID,
					Damage:     damage,
					DamageType: "magic",
					Stage:      0,
				})
			}
			e.Log(fmt.Sprintf("%s 的 [魔能反转] 生效：弃%d张法术牌，对 %s 造成%d点法术伤害", user.Name, xValue, target.Name, damage))
		}
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			if len(e.State.PendingDamageQueue) > 0 {
				e.State.Phase = model.PhasePendingDamageResolution
			}
		}
		return nil
	}
	if choiceType == "ml_black_spear_x" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		maxX := toIntContextValue(ctxData["max_x"])
		x := selectionIndex + 1
		if x < 1 || x > maxX {
			return fmt.Errorf("无效的X值")
		}
		if !e.ConsumeCrystalCost(user.ID, x) {
			return fmt.Errorf("漆黑之枪需要%d点蓝水晶（红宝石可替代）", x)
		}
		targetID, _ := ctxData["target_id"].(string)
		bonus := x + 2
		applied := false
		for i := range e.State.PendingDamageQueue {
			pd := &e.State.PendingDamageQueue[i]
			if !strings.EqualFold(pd.DamageType, "Attack") {
				continue
			}
			if pd.SourceID != user.ID {
				continue
			}
			if targetID != "" && pd.TargetID != targetID {
				continue
			}
			pd.Damage += bonus
			applied = true
			break
		}
		if !applied {
			e.Log("[Warn] 漆黑之枪未找到可叠加的攻击伤害条目")
		}
		e.Log(fmt.Sprintf("%s 的 [漆黑之枪] 生效：消耗%d点蓝水晶，本次攻击伤害额外+%d", user.Name, x, bonus))
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			e.State.Phase = model.PhasePendingDamageResolution
		}
		return nil
	}
	if choiceType == "ml_dark_barrier_mode" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		maxMagic := toIntContextValue(ctxData["max_magic"])
		maxThunder := toIntContextValue(ctxData["max_thunder"])
		modes := make([]string, 0, 2)
		if maxMagic > 0 {
			modes = append(modes, "magic")
		}
		if maxThunder > 0 {
			modes = append(modes, "thunder")
		}
		if selectionIndex < 0 || selectionIndex >= len(modes) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		mode := modes[selectionIndex]
		maxX := maxMagic
		if mode == "thunder" {
			maxX = maxThunder
		}
		if maxX <= 0 {
			return fmt.Errorf("可弃牌数量不足")
		}
		ctxData["mode"] = mode
		ctxData["max_x"] = maxX
		ctxData["choice_type"] = "ml_dark_barrier_x"
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return nil
	}
	if choiceType == "ml_dark_barrier_x" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		mode, _ := ctxData["mode"].(string)
		maxX := toIntContextValue(ctxData["max_x"])
		x := selectionIndex + 1
		if x < 1 || x > maxX {
			return fmt.Errorf("无效的X值")
		}
		remaining := make([]int, 0)
		for idx, c := range user.Hand {
			if mode == "magic" {
				if c.Type == model.CardTypeMagic {
					remaining = append(remaining, idx)
				}
			} else if mode == "thunder" {
				if c.Element == model.ElementThunder {
					remaining = append(remaining, idx)
				}
			}
		}
		if len(remaining) < x {
			return fmt.Errorf("可选弃牌不足，无法选择X=%d", x)
		}
		ctxData["x_value"] = x
		ctxData["selected_indices"] = []int{}
		ctxData["remaining_indices"] = remaining
		ctxData["choice_type"] = "ml_dark_barrier_cards"
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return nil
	}
	if choiceType == "ml_dark_barrier_cards" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		mode, _ := ctxData["mode"].(string)
		xValue := toIntContextValue(ctxData["x_value"])
		var remaining []int
		if arr, ok := ctxData["remaining_indices"].([]int); ok {
			remaining = append(remaining, arr...)
		} else if arr, ok := ctxData["remaining_indices"].([]interface{}); ok {
			for _, v := range arr {
				if f, ok := v.(float64); ok {
					remaining = append(remaining, int(f))
				}
			}
		}
		var selected []int
		if arr, ok := ctxData["selected_indices"].([]int); ok {
			selected = append(selected, arr...)
		} else if arr, ok := ctxData["selected_indices"].([]interface{}); ok {
			for _, v := range arr {
				if f, ok := v.(float64); ok {
					selected = append(selected, int(f))
				}
			}
		}
		cardIdx, ok := resolveSelectionToCandidate(selectionIndex, remaining)
		if !ok || cardIdx < 0 || cardIdx >= len(user.Hand) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		card := user.Hand[cardIdx]
		if mode == "magic" && card.Type != model.CardTypeMagic {
			return fmt.Errorf("暗之障壁当前需要弃法术牌")
		}
		if mode == "thunder" && card.Element != model.ElementThunder {
			return fmt.Errorf("暗之障壁当前需要弃雷系牌")
		}
		selected = append(selected, cardIdx)
		nextRemaining := make([]int, 0, len(remaining))
		for _, idx := range remaining {
			if idx != cardIdx {
				nextRemaining = append(nextRemaining, idx)
			}
		}
		if len(selected) < xValue {
			ctxData["selected_indices"] = selected
			ctxData["remaining_indices"] = nextRemaining
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return nil
		}
		removed, err := removeCardsByIndicesFromHand(user, append([]int{}, selected...))
		if err != nil {
			return err
		}
		e.NotifyCardRevealed(user.ID, removed, "discard")
		e.State.DiscardPile = append(e.State.DiscardPile, removed...)
		e.Log(fmt.Sprintf("%s 的 [暗之障壁] 生效：弃置%d张%s牌", user.Name, xValue, map[string]string{
			"magic":   "法术",
			"thunder": "雷系",
		}[mode]))
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			if len(e.State.PendingDamageQueue) > 0 {
				e.State.Phase = model.PhasePendingDamageResolution
			} else if len(e.State.ActionStack) > 0 {
				e.State.Phase = model.PhaseResponse
			}
		}
		return nil
	}
	if choiceType == "ml_fullness_cost_card" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		candidates := make([]int, 0)
		for idx, c := range user.Hand {
			if c.Type == model.CardTypeMagic || c.Element == model.ElementThunder {
				candidates = append(candidates, idx)
			}
		}
		cardIdx, ok := resolveSelectionToCandidate(selectionIndex, candidates)
		if !ok || cardIdx < 0 || cardIdx >= len(user.Hand) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		costCard := user.Hand[cardIdx]
		user.Hand = append(user.Hand[:cardIdx], user.Hand[cardIdx+1:]...)
		e.NotifyCardRevealed(user.ID, []model.Card{costCard}, "discard")
		e.State.DiscardPile = append(e.State.DiscardPile, costCard)

		ctxData["order_ids"] = e.reverseOrderTargetIDsFrom(user.ID, true)
		ctxData["order_index"] = 0
		ctxData["bonus"] = 0
		ctxData["choice_type"] = "ml_fullness_discard_step"

		done, err := e.prepareMagicLancerFullnessStep(ctxData, user)
		if err != nil {
			return err
		}
		if done {
			if user.TurnState.UsedSkillCounts == nil {
				user.TurnState.UsedSkillCounts = map[string]int{}
			}
			user.TurnState.PendingActions = append(user.TurnState.PendingActions, model.ActionContext{Source: "充盈", MustType: "Attack"})
			e.Log(fmt.Sprintf("%s 的 [充盈] 生效：无可处理弃牌目标，获得额外1次攻击行动", user.Name))
			e.PopInterrupt()
			if e.State.PendingInterrupt == nil {
				e.State.Phase = model.PhaseExtraAction
			}
			return nil
		}
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return nil
	}
	if choiceType == "ml_fullness_discard_step" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		currentID, _ := ctxData["current_player_id"].(string)
		target := e.State.Players[currentID]
		if target == nil {
			return fmt.Errorf("弃牌目标不存在")
		}
		allowSkip, _ := ctxData["allow_skip"].(bool)
		var candidates []int
		if arr, ok := ctxData["candidates"].([]int); ok {
			candidates = append(candidates, arr...)
		} else if arr, ok := ctxData["candidates"].([]interface{}); ok {
			for _, v := range arr {
				if f, ok := v.(float64); ok {
					candidates = append(candidates, int(f))
				}
			}
		}
		if len(candidates) == 0 {
			allowSkip = true
		}
		skipped := false
		chosenCard := model.Card{}
		if allowSkip && selectionIndex == 0 {
			skipped = true
		} else {
			optionIdx := selectionIndex
			if allowSkip {
				optionIdx--
			}
			cardIdx, ok := resolveSelectionToCandidate(optionIdx, candidates)
			if !ok || cardIdx < 0 || cardIdx >= len(target.Hand) {
				return fmt.Errorf("无效的选项索引: %d", selectionIndex)
			}
			chosenCard = target.Hand[cardIdx]
			target.Hand = append(target.Hand[:cardIdx], target.Hand[cardIdx+1:]...)
			e.NotifyCardRevealed(target.ID, []model.Card{chosenCard}, "discard")
			e.State.DiscardPile = append(e.State.DiscardPile, chosenCard)
		}
		if skipped {
			e.Log(fmt.Sprintf("%s 的 [充盈]：%s 选择不弃牌", user.Name, target.Name))
		} else {
			e.Log(fmt.Sprintf("%s 的 [充盈]：%s 弃置了 %s", user.Name, target.Name, chosenCard.Name))
			if target.ID != user.ID && (chosenCard.Type == model.CardTypeMagic || chosenCard.Element == model.ElementThunder) {
				ctxData["bonus"] = toIntContextValue(ctxData["bonus"]) + 1
			}
		}

		ctxData["order_index"] = toIntContextValue(ctxData["order_index"]) + 1
		done, err := e.prepareMagicLancerFullnessStep(ctxData, user)
		if err != nil {
			return err
		}
		if !done {
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return nil
		}

		bonus := toIntContextValue(ctxData["bonus"])
		if user.TurnState.UsedSkillCounts == nil {
			user.TurnState.UsedSkillCounts = map[string]int{}
		}
		if bonus > 0 {
			user.TurnState.UsedSkillCounts["ml_fullness_next_attack_bonus"] += bonus
		}
		user.TurnState.PendingActions = append(user.TurnState.PendingActions, model.ActionContext{Source: "充盈", MustType: "Attack"})
		e.Log(fmt.Sprintf("%s 的 [充盈] 结算完成：本回合下次主动攻击伤害额外+%d，额外获得1次攻击行动", user.Name, bonus))
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			if len(e.State.PendingDamageQueue) > 0 {
				e.State.Phase = model.PhasePendingDamageResolution
				e.State.ReturnPhase = model.PhaseExtraAction
			} else {
				e.State.Phase = model.PhaseExtraAction
			}
		}
		return nil
	}
	if choiceType == "sc_incant_confirm" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		skillID, _ := ctxData["skill_id"].(string)
		targetIDs := dedupeIDs(parseStringSliceContextValue(ctxData["target_ids"]))
		switch selectionIndex {
		case 0:
			if spiritCasterPowerCount(user, "") >= spiritCasterPowerCapEngine || len(user.Hand) == 0 {
				e.Log(fmt.Sprintf("%s 的 [念咒] 未触发：妖力已满或无手牌", user.Name))
				e.PopInterrupt()
				return e.continueSpiritCasterTalisman(user, skillID, targetIDs)
			}
			ctxData["choice_type"] = "sc_incant_card"
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return nil
		case 1:
			e.PopInterrupt()
			return e.continueSpiritCasterTalisman(user, skillID, targetIDs)
		default:
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
	}
	if choiceType == "sc_incant_card" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		skillID, _ := ctxData["skill_id"].(string)
		targetIDs := dedupeIDs(parseStringSliceContextValue(ctxData["target_ids"]))
		if spiritCasterPowerCount(user, "") >= spiritCasterPowerCapEngine {
			return fmt.Errorf("妖力已达上限，无法继续念咒")
		}
		if selectionIndex < 0 || selectionIndex >= len(user.Hand) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		card := user.Hand[selectionIndex]
		user.Hand = append(user.Hand[:selectionIndex], user.Hand[selectionIndex+1:]...)
		if !addSpiritCasterPowerCard(user, card) {
			user.Hand = append(user.Hand, card)
			return fmt.Errorf("妖力已达上限，无法放置")
		}
		e.Log(fmt.Sprintf("%s 发动 [念咒]：将1张手牌盖放为妖力（当前妖力%d/%d）", user.Name, spiritCasterPowerCount(user, ""), spiritCasterPowerCapEngine))
		e.PopInterrupt()
		return e.continueSpiritCasterTalisman(user, skillID, targetIDs)
	}
	if choiceType == "sc_hundred_night_power" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		powers := spiritCasterPowerCovers(user)
		if len(powers) == 0 {
			return fmt.Errorf("没有可移除的妖力")
		}
		powerIdx, ok := resolveSelectionToCandidate(selectionIndex, func() []int {
			idxs := make([]int, 0, len(powers))
			for i := range powers {
				idxs = append(idxs, i)
			}
			return idxs
		}())
		if !ok || powerIdx < 0 || powerIdx >= len(powers) || powers[powerIdx] == nil {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		selectedPower := powers[powerIdx]
		card := selectedPower.Card
		user.RemoveFieldCard(selectedPower)
		syncSpiritCasterPowerToken(user)
		e.State.DiscardPile = append(e.State.DiscardPile, card)
		e.Log(fmt.Sprintf("%s 发动 [百鬼夜行]：移除1个妖力", user.Name))

		if card.Element == model.ElementFire {
			ctxData["removed_element"] = string(card.Element)
			ctxData["removed_name"] = card.Name
			ctxData["choice_type"] = "sc_hundred_night_fire_reveal"
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return nil
		}
		ctxData["choice_type"] = "sc_hundred_night_target"
		ctxData["target_ids"] = append([]string{}, e.State.PlayerOrder...)
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return nil
	}
	if choiceType == "sc_hundred_night_fire_reveal" {
		switch selectionIndex {
		case 0:
			userID, _ := ctxData["user_id"].(string)
			user := e.State.Players[userID]
			if user != nil {
				e.Log(fmt.Sprintf("%s 展示了火系妖力，触发 [百鬼夜行] 范围分支", user.Name))
			}
			ctxData["choice_type"] = "sc_hundred_night_exclude_pick"
			ctxData["target_ids"] = append([]string{}, e.State.PlayerOrder...)
			ctxData["selected_exclude_ids"] = []string{}
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return nil
		case 1:
			ctxData["choice_type"] = "sc_hundred_night_target"
			ctxData["target_ids"] = append([]string{}, e.State.PlayerOrder...)
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return nil
		default:
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
	}
	if choiceType == "sc_hundred_night_exclude_pick" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		allTargetIDs := parseStringSliceContextValue(ctxData["target_ids"])
		if len(allTargetIDs) < 2 {
			return fmt.Errorf("可选目标不足2名")
		}
		selected := dedupeIDs(parseStringSliceContextValue(ctxData["selected_exclude_ids"]))
		selectedSet := idsToSet(selected)
		remaining := make([]string, 0, len(allTargetIDs))
		for _, tid := range allTargetIDs {
			if !selectedSet[tid] {
				remaining = append(remaining, tid)
			}
		}
		if selectionIndex < 0 || selectionIndex >= len(remaining) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		selected = append(selected, remaining[selectionIndex])
		if len(selected) < 2 {
			ctxData["selected_exclude_ids"] = selected
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return nil
		}
		if e.CanPayCrystalCost(user.ID, 1) {
			ctxData["choice_type"] = "sc_spiritual_collapse_confirm"
			ctxData["mode"] = "sc_hundred_night_fire_aoe"
			ctxData["exclude_ids"] = selected
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return nil
		}
		if err := e.resolveSpiritCasterHundredNightFireAOE(user, selected, 0); err != nil {
			return err
		}
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			e.State.Phase = model.PhasePendingDamageResolution
		}
		return nil
	}
	if choiceType == "sc_hundred_night_target" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		targetIDs := parseStringSliceContextValue(ctxData["target_ids"])
		if selectionIndex < 0 || selectionIndex >= len(targetIDs) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		targetID := targetIDs[selectionIndex]
		if e.CanPayCrystalCost(user.ID, 1) {
			ctxData["choice_type"] = "sc_spiritual_collapse_confirm"
			ctxData["mode"] = "sc_hundred_night_single"
			ctxData["target_id"] = targetID
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return nil
		}
		if err := e.resolveSpiritCasterHundredNightSingle(user, targetID, 0); err != nil {
			return err
		}
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			e.State.Phase = model.PhasePendingDamageResolution
		}
		return nil
	}
	if choiceType == "sc_spiritual_collapse_confirm" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		bonus := 0
		if selectionIndex == 0 {
			if !e.ConsumeCrystalCost(user.ID, 1) {
				return fmt.Errorf("灵力崩解需要1点蓝水晶（红宝石可替代）")
			}
			bonus = 1
			e.Log(fmt.Sprintf("%s 发动 [灵力崩解]：本次每段伤害额外+1", user.Name))
		} else if selectionIndex != 1 {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		mode, _ := ctxData["mode"].(string)
		switch mode {
		case "sc_talisman_thunder":
			targetIDs := dedupeIDs(parseStringSliceContextValue(ctxData["target_ids"]))
			e.PopInterrupt()
			e.resolveSpiritCasterThunderDamage(user, targetIDs, bonus)
		case "sc_hundred_night_single":
			targetID, _ := ctxData["target_id"].(string)
			if targetID == "" {
				return fmt.Errorf("百鬼夜行目标缺失")
			}
			if err := e.resolveSpiritCasterHundredNightSingle(user, targetID, bonus); err != nil {
				return err
			}
			e.PopInterrupt()
		case "sc_hundred_night_fire_aoe":
			excludeIDs := dedupeIDs(parseStringSliceContextValue(ctxData["exclude_ids"]))
			if len(excludeIDs) != 2 {
				return fmt.Errorf("百鬼夜行火系分支需要2名排除目标")
			}
			if err := e.resolveSpiritCasterHundredNightFireAOE(user, excludeIDs, bonus); err != nil {
				return err
			}
			e.PopInterrupt()
		default:
			return fmt.Errorf("灵力崩解上下文无效")
		}
		if e.State.PendingInterrupt == nil && len(e.State.PendingDamageQueue) > 0 {
			e.State.Phase = model.PhasePendingDamageResolution
		}
		return nil
	}
	if choiceType == "sc_talisman_wind_discard" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("灵符师不存在")
		}
		ordered := parseStringSliceContextValue(ctxData["ordered_target_ids"])
		if len(ordered) == 0 {
			return fmt.Errorf("灵符-风行上下文无效")
		}
		cursor := toIntContextValue(ctxData["cursor"])
		if cursor < 0 || cursor >= len(ordered) {
			return fmt.Errorf("灵符-风行游标无效")
		}
		currentTargetID, _ := ctxData["current_target_id"].(string)
		if currentTargetID == "" {
			currentTargetID = ordered[cursor]
		}
		target := e.State.Players[currentTargetID]
		if target == nil {
			return fmt.Errorf("弃牌目标不存在")
		}
		if len(target.Hand) == 0 {
			e.Log(fmt.Sprintf("%s 的 [灵符-风行]：%s 已无手牌，跳过", user.Name, target.Name))
		} else {
			candidates := allHandIndices(target)
			cardIdx, ok := resolveSelectionToCandidate(selectionIndex, candidates)
			if !ok || cardIdx < 0 || cardIdx >= len(target.Hand) {
				return fmt.Errorf("无效的选项索引: %d", selectionIndex)
			}
			card := target.Hand[cardIdx]
			target.Hand = append(target.Hand[:cardIdx], target.Hand[cardIdx+1:]...)
			e.NotifyCardHidden(target.ID, []model.Card{card}, "discard")
			e.State.DiscardPile = append(e.State.DiscardPile, card)
			e.Log(fmt.Sprintf("%s 的 [灵符-风行]：%s 选择弃置了1张手牌", user.Name, target.Name))
		}

		nextCursor := cursor + 1
		for nextCursor < len(ordered) {
			nextTarget := e.State.Players[ordered[nextCursor]]
			if nextTarget == nil {
				nextCursor++
				continue
			}
			if len(nextTarget.Hand) <= 0 {
				e.Log(fmt.Sprintf("%s 的 [灵符-风行]：%s 无手牌可弃置", user.Name, nextTarget.Name))
				nextCursor++
				continue
			}
			ctxData["cursor"] = nextCursor
			ctxData["current_target_id"] = nextTarget.ID
			e.State.PendingInterrupt.PlayerID = nextTarget.ID
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return nil
		}

		e.Log(fmt.Sprintf("%s 的 [灵符-风行] 结算完成", user.Name))
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil && len(e.State.PendingDamageQueue) > 0 {
			e.State.Phase = model.PhasePendingDamageResolution
		}
		return nil
	}
	if choiceType == "bd_descent_confirm" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		if user.Tokens == nil {
			user.Tokens = map[string]int{}
		}
		if selectionIndex == 1 {
			user.Tokens["bd_descent_used_turn"] = 1
			e.PopInterrupt()
			if e.State.PendingInterrupt == nil && len(e.State.PendingDamageQueue) > 0 {
				e.State.Phase = model.PhasePendingDamageResolution
			}
			return nil
		}
		if selectionIndex != 0 {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		if bardMaxSameElementCount(user) < 2 {
			return fmt.Errorf("同系手牌不足2张，无法发动沉沦协奏曲")
		}
		ctxData["choice_type"] = "bd_descent_element"
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return nil
	}
	if choiceType == "bd_descent_element" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		elemCounts := getSameElementCounts(user)
		var elems []model.Element
		for _, ele := range elementOrderForPrompt() {
			if elemCounts[ele] >= 2 {
				elems = append(elems, ele)
			}
		}
		if selectionIndex < 0 || selectionIndex >= len(elems) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		chosen := elems[selectionIndex]
		ctxData["chosen_element"] = string(chosen)
		ctxData["selected_indices"] = []int{}
		ctxData["remaining_indices"] = getCardIndicesByElement(user, chosen)
		ctxData["choice_type"] = "bd_descent_cards"
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return nil
	}
	if choiceType == "bd_descent_cards" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		chosenElement, _ := ctxData["chosen_element"].(string)
		var remaining []int
		if arr, ok := ctxData["remaining_indices"].([]int); ok {
			remaining = append(remaining, arr...)
		} else if arr, ok := ctxData["remaining_indices"].([]interface{}); ok {
			for _, v := range arr {
				if f, ok := v.(float64); ok {
					remaining = append(remaining, int(f))
				}
			}
		}
		var selected []int
		if arr, ok := ctxData["selected_indices"].([]int); ok {
			selected = append(selected, arr...)
		} else if arr, ok := ctxData["selected_indices"].([]interface{}); ok {
			for _, v := range arr {
				if f, ok := v.(float64); ok {
					selected = append(selected, int(f))
				}
			}
		}
		cardIdx, ok := resolveSelectionToCandidate(selectionIndex, remaining)
		if !ok || cardIdx < 0 || cardIdx >= len(user.Hand) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		if string(user.Hand[cardIdx].Element) != chosenElement {
			return fmt.Errorf("沉沦协奏曲需弃置同系牌")
		}
		selected = append(selected, cardIdx)
		var nextRemaining []int
		for _, idx := range remaining {
			if idx != cardIdx {
				nextRemaining = append(nextRemaining, idx)
			}
		}
		if len(selected) < 2 {
			ctxData["selected_indices"] = selected
			ctxData["remaining_indices"] = nextRemaining
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return nil
		}
		removed, err := removeCardsByIndicesFromHand(user, append([]int{}, selected...))
		if err != nil {
			return err
		}
		e.NotifyCardRevealed(user.ID, removed, "discard")
		e.State.DiscardPile = append(e.State.DiscardPile, removed...)
		if user.Tokens == nil {
			user.Tokens = map[string]int{}
		}
		user.Tokens["bd_descent_used_turn"] = 1
		now := addBardInspiration(user, 1)
		e.Log(fmt.Sprintf("%s 发动 [沉沦协奏曲]：弃2张%s系牌，灵感+1（当前%d）", user.Name, chosenElement, now))

		hasMagic := false
		for _, c := range removed {
			if c.Type == model.CardTypeMagic {
				hasMagic = true
				break
			}
		}
		if hasMagic {
			ctxData["choice_type"] = "bd_descent_target"
			ctxData["target_ids"] = e.campEnemyIDs(user.Camp)
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return nil
		}
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil && len(e.State.PendingDamageQueue) > 0 {
			e.State.Phase = model.PhasePendingDamageResolution
		}
		return nil
	}
	if choiceType == "bd_descent_target" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		targetIDs := parseStringSliceContextValue(ctxData["target_ids"])
		if selectionIndex < 0 || selectionIndex >= len(targetIDs) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		targetID := targetIDs[selectionIndex]
		target := e.State.Players[targetID]
		if target == nil {
			return fmt.Errorf("目标不存在")
		}
		e.AddPendingDamage(model.PendingDamage{
			SourceID:   user.ID,
			TargetID:   target.ID,
			Damage:     1,
			DamageType: "magic",
			Stage:      0,
		})
		e.Log(fmt.Sprintf("%s 的 [沉沦协奏曲] 追加效果：对 %s 造成1点法术伤害", user.Name, target.Name))
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			e.State.Phase = model.PhasePendingDamageResolution
		}
		return nil
	}
	if choiceType == "bd_dissonance_x" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		maxX := toIntContextValue(ctxData["max_x"])
		xValue := selectionIndex + 2
		if xValue < 2 || xValue > maxX {
			return fmt.Errorf("无效的X值")
		}
		if bardInspiration(user) < xValue {
			return fmt.Errorf("灵感不足")
		}
		addBardInspiration(user, -xValue)
		if user.Tokens != nil && user.Tokens["bd_prisoner_form"] > 0 {
			user.Tokens["bd_prisoner_form"] = 0
			e.Log(fmt.Sprintf("%s 发动 [不谐和弦]：脱离永恒囚徒形态", user.Name))
		}
		ctxData["x_value"] = xValue
		ctxData["choice_type"] = "bd_dissonance_mode"
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return nil
	}
	if choiceType == "bd_dissonance_mode" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		if selectionIndex != 0 && selectionIndex != 1 {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		ctxData["mode"] = selectionIndex
		ctxData["choice_type"] = "bd_dissonance_target"
		ctxData["target_ids"] = append([]string{}, e.State.PlayerOrder...)
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return nil
	}
	if choiceType == "bd_dissonance_target" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		targetIDs := parseStringSliceContextValue(ctxData["target_ids"])
		if selectionIndex < 0 || selectionIndex >= len(targetIDs) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		targetID := targetIDs[selectionIndex]
		target := e.State.Players[targetID]
		if target == nil {
			return fmt.Errorf("目标不存在")
		}
		xValue := toIntContextValue(ctxData["x_value"])
		mode := toIntContextValue(ctxData["mode"])
		n := xValue - 1
		if n < 0 {
			n = 0
		}
		if mode == 0 {
			if n > 0 {
				e.DrawCards(user.ID, n)
				e.DrawCards(target.ID, n)
			}
			e.Log(fmt.Sprintf("%s 发动 [不谐和弦]：与 %s 各摸%d张牌", user.Name, target.Name, n))
			e.PopInterrupt()
			if e.State.PendingInterrupt == nil {
				if len(e.State.PendingDamageQueue) > 0 {
					e.State.Phase = model.PhasePendingDamageResolution
					e.State.ReturnPhase = model.PhaseExtraAction
				} else {
					e.State.Phase = model.PhaseExtraAction
				}
			}
			return nil
		}
		actors := []string{user.ID, target.ID}
		startCursor := 0
		for startCursor < len(actors) {
			actor := e.State.Players[actors[startCursor]]
			if actor != nil && len(actor.Hand) > 0 && n > 0 {
				break
			}
			startCursor++
		}
		if n <= 0 || startCursor >= len(actors) {
			e.Log(fmt.Sprintf("%s 发动 [不谐和弦]：弃牌分支无可执行弃牌", user.Name))
			e.PopInterrupt()
			if e.State.PendingInterrupt == nil {
				e.State.Phase = model.PhaseExtraAction
			}
			return nil
		}
		currentActor := e.State.Players[actors[startCursor]]
		ctxData["choice_type"] = "bd_dissonance_discard_step"
		ctxData["actor_ids"] = actors
		ctxData["cursor"] = startCursor
		ctxData["current_actor_id"] = currentActor.ID
		ctxData["need_count"] = n
		ctxData["selected_count"] = 0
		ctxData["selected_indices"] = []int{}
		ctxData["remaining_indices"] = allHandIndices(currentActor)
		e.State.PendingInterrupt.PlayerID = currentActor.ID
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return nil
	}
	if choiceType == "bd_dissonance_discard_step" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		actorIDs := parseStringSliceContextValue(ctxData["actor_ids"])
		cursor := toIntContextValue(ctxData["cursor"])
		if cursor < 0 || cursor >= len(actorIDs) {
			return fmt.Errorf("弃牌游标无效")
		}
		currentActorID, _ := ctxData["current_actor_id"].(string)
		if currentActorID == "" {
			currentActorID = actorIDs[cursor]
		}
		actor := e.State.Players[currentActorID]
		if actor == nil {
			return fmt.Errorf("弃牌角色不存在")
		}
		needCount := toIntContextValue(ctxData["need_count"])
		selectedCount := toIntContextValue(ctxData["selected_count"])
		var remaining []int
		if arr, ok := ctxData["remaining_indices"].([]int); ok {
			remaining = append(remaining, arr...)
		} else if arr, ok := ctxData["remaining_indices"].([]interface{}); ok {
			for _, v := range arr {
				if f, ok := v.(float64); ok {
					remaining = append(remaining, int(f))
				}
			}
		}
		var selected []int
		if arr, ok := ctxData["selected_indices"].([]int); ok {
			selected = append(selected, arr...)
		} else if arr, ok := ctxData["selected_indices"].([]interface{}); ok {
			for _, v := range arr {
				if f, ok := v.(float64); ok {
					selected = append(selected, int(f))
				}
			}
		}
		cardIdx, ok := resolveSelectionToCandidate(selectionIndex, remaining)
		if !ok || cardIdx < 0 || cardIdx >= len(actor.Hand) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		selected = append(selected, cardIdx)
		selectedCount++
		nextRemaining := make([]int, 0, len(remaining))
		for _, idx := range remaining {
			if idx != cardIdx {
				nextRemaining = append(nextRemaining, idx)
			}
		}
		if selectedCount < needCount && len(nextRemaining) > 0 {
			ctxData["selected_count"] = selectedCount
			ctxData["selected_indices"] = selected
			ctxData["remaining_indices"] = nextRemaining
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return nil
		}
		removed, err := removeCardsByIndicesFromHand(actor, append([]int{}, selected...))
		if err != nil {
			return err
		}
		e.NotifyCardHidden(actor.ID, removed, "discard")
		e.State.DiscardPile = append(e.State.DiscardPile, removed...)
		e.Log(fmt.Sprintf("%s 的 [不谐和弦]：%s 弃置了%d张手牌", user.Name, actor.Name, len(removed)))

		nextCursor := cursor + 1
		for nextCursor < len(actorIDs) {
			nextActor := e.State.Players[actorIDs[nextCursor]]
			if nextActor == nil || len(nextActor.Hand) == 0 || needCount <= 0 {
				nextCursor++
				continue
			}
			ctxData["cursor"] = nextCursor
			ctxData["current_actor_id"] = nextActor.ID
			ctxData["selected_count"] = 0
			ctxData["selected_indices"] = []int{}
			ctxData["remaining_indices"] = allHandIndices(nextActor)
			e.State.PendingInterrupt.PlayerID = nextActor.ID
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return nil
		}
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			if len(e.State.PendingDamageQueue) > 0 {
				e.State.Phase = model.PhasePendingDamageResolution
				e.State.ReturnPhase = model.PhaseExtraAction
			} else {
				e.State.Phase = model.PhaseExtraAction
			}
		}
		return nil
	}
	if choiceType == "bd_rousing_mode" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("吟游诗人不存在")
		}
		if selectionIndex == 0 {
			ctxData["choice_type"] = "bd_rousing_targets"
			ctxData["selected_target_ids"] = []string{}
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return nil
		}
		if selectionIndex == 1 {
			if len(user.Hand) < 2 {
				return fmt.Errorf("手牌不足2张，无法执行弃2张牌分支")
			}
			ctxData["choice_type"] = "bd_rousing_discard_cards"
			ctxData["selected_indices"] = []int{}
			ctxData["remaining_indices"] = allHandIndices(user)
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return nil
		}
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	if choiceType == "bd_rousing_targets" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("吟游诗人不存在")
		}
		targetIDs := parseStringSliceContextValue(ctxData["target_ids"])
		selected := dedupeIDs(parseStringSliceContextValue(ctxData["selected_target_ids"]))
		selectedSet := idsToSet(selected)
		remaining := make([]string, 0, len(targetIDs))
		for _, tid := range targetIDs {
			if !selectedSet[tid] {
				remaining = append(remaining, tid)
			}
		}
		if selectionIndex < 0 || selectionIndex >= len(remaining) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		selected = append(selected, remaining[selectionIndex])
		if len(selected) < 2 {
			ctxData["selected_target_ids"] = selected
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return nil
		}
		for _, tid := range selected {
			e.AddPendingDamage(model.PendingDamage{
				SourceID:   user.ID,
				TargetID:   tid,
				Damage:     1,
				DamageType: "magic",
				Stage:      0,
			})
		}
		e.Log(fmt.Sprintf("%s 发动 [激昂狂想曲]：对2名目标各造成1点法术伤害", user.Name))
		e.resolveBardForbiddenVerseAfterSong(user, "激昂狂想曲")
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			if len(e.State.PendingDamageQueue) > 0 {
				e.State.Phase = model.PhasePendingDamageResolution
			} else {
				e.State.Phase = model.PhaseStartup
			}
		}
		return nil
	}
	if choiceType == "bd_rousing_discard_cards" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("吟游诗人不存在")
		}
		var remaining []int
		if arr, ok := ctxData["remaining_indices"].([]int); ok {
			remaining = append(remaining, arr...)
		} else if arr, ok := ctxData["remaining_indices"].([]interface{}); ok {
			for _, v := range arr {
				if f, ok := v.(float64); ok {
					remaining = append(remaining, int(f))
				}
			}
		}
		var selected []int
		if arr, ok := ctxData["selected_indices"].([]int); ok {
			selected = append(selected, arr...)
		} else if arr, ok := ctxData["selected_indices"].([]interface{}); ok {
			for _, v := range arr {
				if f, ok := v.(float64); ok {
					selected = append(selected, int(f))
				}
			}
		}
		cardIdx, ok := resolveSelectionToCandidate(selectionIndex, remaining)
		if !ok || cardIdx < 0 || cardIdx >= len(user.Hand) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		selected = append(selected, cardIdx)
		nextRemaining := make([]int, 0, len(remaining))
		for _, idx := range remaining {
			if idx != cardIdx {
				nextRemaining = append(nextRemaining, idx)
			}
		}
		if len(selected) < 2 {
			ctxData["selected_indices"] = selected
			ctxData["remaining_indices"] = nextRemaining
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return nil
		}
		removed, err := removeCardsByIndicesFromHand(user, append([]int{}, selected...))
		if err != nil {
			return err
		}
		e.NotifyCardRevealed(user.ID, removed, "discard")
		e.State.DiscardPile = append(e.State.DiscardPile, removed...)
		e.Log(fmt.Sprintf("%s 发动 [激昂狂想曲]：选择弃2张牌", user.Name))
		e.resolveBardForbiddenVerseAfterSong(user, "激昂狂想曲")
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			if len(e.State.PendingDamageQueue) > 0 {
				e.State.Phase = model.PhasePendingDamageResolution
			} else {
				e.State.Phase = model.PhaseStartup
			}
		}
		return nil
	}
	if choiceType == "bd_victory_mode" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("吟游诗人不存在")
		}
		switch selectionIndex {
		case 0:
			// 从我方战绩区提炼1个星石为吟游诗人的能量。
			camp := string(user.Camp)
			gems := e.GetCampGems(camp)
			crystals := e.GetCampCrystals(camp)
			if gems+crystals <= 0 {
				e.Log(fmt.Sprintf("%s 的 [胜利交响诗] 分支①失败：我方战绩区无星石可提炼", user.Name))
			} else {
				addGem, addCrystal := 0, 0
				if crystals > 0 {
					e.ModifyCrystal(camp, -1)
					addCrystal = 1
				} else {
					e.ModifyGem(camp, -1)
					addGem = 1
				}
				maxEnergy := e.getPlayerEnergyCap(user)
				room := maxEnergy - (user.Gem + user.Crystal)
				if room <= 0 {
					e.Log(fmt.Sprintf("%s 的 [胜利交响诗]：提炼成功但能量已满，未增加个人能量", user.Name))
				} else {
					if addGem > room {
						addGem = room
						addCrystal = 0
					}
					if addCrystal > room {
						addCrystal = room
						addGem = 0
					}
					user.Gem += addGem
					user.Crystal += addCrystal
					e.Log(fmt.Sprintf("%s 发动 [胜利交响诗]：提炼1个星石为个人能量（+%d宝石 +%d水晶）", user.Name, addGem, addCrystal))
				}
			}
		case 1:
			e.addCampResource(user.Camp, "gem")
			e.Heal(user.ID, 1)
			e.Log(fmt.Sprintf("%s 发动 [胜利交响诗]：我方战绩区+1宝石，自己+1治疗", user.Name))
		default:
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		e.resolveBardForbiddenVerseAfterSong(user, "胜利交响诗")
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			if len(e.State.PendingDamageQueue) > 0 {
				e.State.Phase = model.PhasePendingDamageResolution
				e.State.ReturnPhase = model.PhaseTurnEnd
			} else {
				e.State.Phase = model.PhaseTurnEnd
			}
		}
		return nil
	}
	if choiceType == "bd_hope_draw_confirm" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("吟游诗人不存在")
		}
		if selectionIndex == 0 {
			e.DrawCards(user.ID, 1)
		} else if selectionIndex != 1 {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		ctxData["choice_type"] = "bd_hope_mode"
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return nil
	}
	if choiceType == "bd_hope_mode" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("吟游诗人不存在")
		}
		switch selectionIndex {
		case 0:
			targetIDs := e.bardAlliesExcluding(user.Camp, user.ID)
			if len(targetIDs) == 0 {
				return fmt.Errorf("无可选队友目标")
			}
			ctxData["choice_type"] = "bd_hope_place_target"
			ctxData["target_ids"] = targetIDs
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return nil
		case 1:
			holderID := e.bardEternalHolderID(user)
			if holderID == "" {
				return fmt.Errorf("当前没有永恒乐章可转移")
			}
			if len(user.Hand) == 0 {
				return fmt.Errorf("手牌不足，无法执行转移分支")
			}
			targetIDs := e.bardAlliesExcluding(user.Camp, holderID)
			if len(targetIDs) == 0 {
				return fmt.Errorf("没有可转移的目标角色")
			}
			ctxData["holder_id"] = holderID
			ctxData["choice_type"] = "bd_hope_transfer_target"
			ctxData["target_ids"] = targetIDs
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return nil
		default:
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
	}
	if choiceType == "bd_hope_place_target" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("吟游诗人不存在")
		}
		targetIDs := parseStringSliceContextValue(ctxData["target_ids"])
		if selectionIndex < 0 || selectionIndex >= len(targetIDs) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		target := e.State.Players[targetIDs[selectionIndex]]
		if target == nil {
			return fmt.Errorf("目标不存在")
		}
		if err := e.placeBardEternalMovement(user, target); err != nil {
			return err
		}
		e.Log(fmt.Sprintf("%s 发动 [希望赋格曲]：将永恒乐章放置于 %s 面前", user.Name, target.Name))
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			e.State.Phase = model.PhaseStartup
		}
		return nil
	}
	if choiceType == "bd_hope_transfer_target" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("吟游诗人不存在")
		}
		targetIDs := parseStringSliceContextValue(ctxData["target_ids"])
		if selectionIndex < 0 || selectionIndex >= len(targetIDs) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		target := e.State.Players[targetIDs[selectionIndex]]
		if target == nil {
			return fmt.Errorf("目标不存在")
		}
		if err := e.placeBardEternalMovement(user, target); err != nil {
			return err
		}
		ctxData["choice_type"] = "bd_hope_transfer_discard"
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return nil
	}
	if choiceType == "bd_hope_transfer_discard" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("吟游诗人不存在")
		}
		if selectionIndex < 0 || selectionIndex >= len(user.Hand) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		card := user.Hand[selectionIndex]
		user.Hand = append(user.Hand[:selectionIndex], user.Hand[selectionIndex+1:]...)
		e.NotifyCardRevealed(user.ID, []model.Card{card}, "discard")
		e.State.DiscardPile = append(e.State.DiscardPile, card)
		ctxData["choice_type"] = "bd_hope_transfer_gain"
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return nil
	}
	if choiceType == "bd_hope_transfer_gain" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("吟游诗人不存在")
		}
		if selectionIndex == 0 {
			e.Heal(user.ID, 1)
			e.Log(fmt.Sprintf("%s 的 [希望赋格曲] 转移分支：+1治疗", user.Name))
		} else if selectionIndex == 1 {
			now := addBardInspiration(user, 1)
			e.Log(fmt.Sprintf("%s 的 [希望赋格曲] 转移分支：灵感+1（当前%d）", user.Name, now))
		} else {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			e.State.Phase = model.PhaseStartup
		}
		return nil
	}
	if choiceType == "sage_wisdom_codex_discard_confirm" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		if selectionIndex == 0 && len(user.Hand) > 0 {
			e.PushInterrupt(&model.Interrupt{
				Type:     model.InterruptDiscard,
				PlayerID: userID,
				Context: map[string]interface{}{
					"discard_count": 1,
					"stay_in_turn":  true,
					"prompt":        "【智慧法典】请选择弃置1张手牌：",
				},
			})
		} else if selectionIndex != 1 && selectionIndex != 0 {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil && len(e.State.PendingDamageQueue) > 0 {
			e.State.Phase = model.PhasePendingDamageResolution
		}
		return nil
	}
	if choiceType == "sage_magic_rebound_confirm" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		if selectionIndex == 1 {
			e.PopInterrupt()
			if e.State.PendingInterrupt == nil && len(e.State.PendingDamageQueue) > 0 {
				e.State.Phase = model.PhasePendingDamageResolution
			}
			return nil
		}
		if selectionIndex != 0 {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		maxX := maxSameElementCount(user)
		if maxX < 2 {
			return fmt.Errorf("同系手牌不足2张，无法发动法术反弹")
		}
		ctxData["choice_type"] = "sage_magic_rebound_x"
		ctxData["max_x"] = maxX
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return nil
	}
	if choiceType == "sage_magic_rebound_x" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		xValue := selectionIndex + 2
		maxX := maxSameElementCount(user)
		if xValue < 2 || xValue > maxX {
			return fmt.Errorf("无效的X值")
		}
		ctxData["x_value"] = xValue
		ctxData["choice_type"] = "sage_magic_rebound_element"
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return nil
	}
	if choiceType == "sage_magic_rebound_element" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		xValue := toIntContextValue(ctxData["x_value"])
		elements := availableElementsByMinCount(user, xValue)
		if selectionIndex < 0 || selectionIndex >= len(elements) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		chosenElement := model.Element(elements[selectionIndex])
		ctxData["chosen_element"] = string(chosenElement)
		ctxData["choice_type"] = "sage_magic_rebound_cards"
		ctxData["selected_indices"] = []int{}
		ctxData["remaining_indices"] = getCardIndicesByElement(user, chosenElement)
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return nil
	}
	if choiceType == "sage_magic_rebound_cards" || choiceType == "sage_arcane_cards" || choiceType == "sage_holy_cards" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		xValue := toIntContextValue(ctxData["x_value"])
		if xValue <= 0 {
			return fmt.Errorf("X值无效")
		}
		var remaining []int
		if arr, ok := ctxData["remaining_indices"].([]int); ok {
			remaining = append(remaining, arr...)
		} else if arr, ok := ctxData["remaining_indices"].([]interface{}); ok {
			for _, v := range arr {
				if f, ok := v.(float64); ok {
					remaining = append(remaining, int(f))
				}
			}
		}
		var selected []int
		if arr, ok := ctxData["selected_indices"].([]int); ok {
			selected = append(selected, arr...)
		} else if arr, ok := ctxData["selected_indices"].([]interface{}); ok {
			for _, v := range arr {
				if f, ok := v.(float64); ok {
					selected = append(selected, int(f))
				}
			}
		}
		cardIdx, ok := resolveSelectionToCandidate(selectionIndex, remaining)
		if !ok || cardIdx < 0 || cardIdx >= len(user.Hand) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		chosenCard := user.Hand[cardIdx]
		for _, idx := range selected {
			if choiceType != "sage_magic_rebound_cards" &&
				idx >= 0 &&
				idx < len(user.Hand) &&
				user.Hand[idx].Element == chosenCard.Element {
				return fmt.Errorf("需弃置异系牌，不能重复选择同系")
			}
		}
		if choiceType == "sage_magic_rebound_cards" {
			ele, _ := ctxData["chosen_element"].(string)
			if string(chosenCard.Element) != ele {
				return fmt.Errorf("法术反弹需弃置同系牌")
			}
		}
		selected = append(selected, cardIdx)
		var nextRemaining []int
		if choiceType == "sage_magic_rebound_cards" {
			for _, v := range remaining {
				if v != cardIdx {
					nextRemaining = append(nextRemaining, v)
				}
			}
		} else {
			nextRemaining = removeElementIndices(remaining, user, chosenCard.Element, cardIdx)
		}
		if len(selected) < xValue {
			ctxData["selected_indices"] = selected
			ctxData["remaining_indices"] = nextRemaining
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return nil
		}
		ctxData["selected_indices"] = selected
		if choiceType == "sage_magic_rebound_cards" {
			ctxData["choice_type"] = "sage_magic_rebound_target"
		} else if choiceType == "sage_arcane_cards" {
			ctxData["choice_type"] = "sage_arcane_target"
		} else {
			maxTargetCount := xValue - 2
			if maxTargetCount < 0 {
				maxTargetCount = 0
			}
			ctxData["choice_type"] = "sage_holy_target_count"
			ctxData["max_target_count"] = maxTargetCount
		}
		ctxData["target_ids"] = append([]string{}, e.State.PlayerOrder...)
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return nil
	}
	if choiceType == "sage_arcane_x" || choiceType == "sage_holy_x" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		maxX := toIntContextValue(ctxData["max_x"])
		minX := 2
		if choiceType == "sage_holy_x" {
			minX = 3
		}
		xValue := selectionIndex + minX
		if xValue < minX || xValue > maxX {
			return fmt.Errorf("无效的X值")
		}
		ctxData["x_value"] = xValue
		ctxData["selected_indices"] = []int{}
		ctxData["remaining_indices"] = allHandIndices(user)
		if choiceType == "sage_arcane_x" {
			ctxData["choice_type"] = "sage_arcane_cards"
		} else {
			ctxData["choice_type"] = "sage_holy_cards"
		}
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return nil
	}
	if choiceType == "sage_holy_target_count" {
		targetCount := selectionIndex
		maxCount := toIntContextValue(ctxData["max_target_count"])
		if targetCount < 0 || targetCount > maxCount {
			return fmt.Errorf("无效的治疗目标数量")
		}
		if targetCount == 0 {
			userID, _ := ctxData["user_id"].(string)
			user := e.State.Players[userID]
			if user == nil {
				return fmt.Errorf("玩家不存在")
			}
			var selectedCards []int
			if arr, ok := ctxData["selected_indices"].([]int); ok {
				selectedCards = append(selectedCards, arr...)
			} else if arr, ok := ctxData["selected_indices"].([]interface{}); ok {
				for _, v := range arr {
					if f, ok := v.(float64); ok {
						selectedCards = append(selectedCards, int(f))
					}
				}
			}
			xValue := toIntContextValue(ctxData["x_value"])
			if xValue <= 2 || len(selectedCards) != xValue {
				return fmt.Errorf("圣洁法典弃牌参数无效")
			}
			removed, err := removeCardsByIndicesFromHand(user, append([]int{}, selectedCards...))
			if err != nil {
				return err
			}
			e.NotifyCardRevealed(user.ID, removed, "discard")
			e.State.DiscardPile = append(e.State.DiscardPile, removed...)
			damage := xValue - 1
			if damage > 0 {
				e.AddPendingDamage(model.PendingDamage{
					SourceID:   user.ID,
					TargetID:   user.ID,
					Damage:     damage,
					DamageType: "magic",
					Stage:      0,
				})
			}
			e.Log(fmt.Sprintf("%s 发动 [圣洁法典]：未选择治疗目标，对自己造成%d点法术伤害", user.Name, damage))
			e.PopInterrupt()
			if e.State.PendingInterrupt == nil {
				if len(e.State.PendingDamageQueue) > 0 {
					e.State.Phase = model.PhasePendingDamageResolution
					e.State.ReturnPhase = model.PhaseExtraAction
				} else {
					e.State.Phase = model.PhaseExtraAction
				}
			}
			return nil
		}
		ctxData["target_count"] = targetCount
		ctxData["selected_target_ids"] = []string{}
		ctxData["choice_type"] = "sage_holy_targets"
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return nil
	}
	if choiceType == "sage_holy_targets" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		targetCount := toIntContextValue(ctxData["target_count"])
		if targetCount <= 0 {
			return fmt.Errorf("治疗目标数量无效")
		}
		var allTargetIDs []string
		if arr, ok := ctxData["target_ids"].([]string); ok {
			allTargetIDs = append(allTargetIDs, arr...)
		} else if arr, ok := ctxData["target_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					allTargetIDs = append(allTargetIDs, s)
				}
			}
		}
		var selected []string
		selectedSet := map[string]bool{}
		if arr, ok := ctxData["selected_target_ids"].([]string); ok {
			for _, s := range arr {
				if s != "" && !selectedSet[s] {
					selected = append(selected, s)
					selectedSet[s] = true
				}
			}
		} else if arr, ok := ctxData["selected_target_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok && s != "" && !selectedSet[s] {
					selected = append(selected, s)
					selectedSet[s] = true
				}
			}
		}
		var remaining []string
		for _, tid := range allTargetIDs {
			if !selectedSet[tid] {
				remaining = append(remaining, tid)
			}
		}
		if selectionIndex < 0 || selectionIndex >= len(remaining) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		selected = append(selected, remaining[selectionIndex])
		ctxData["selected_target_ids"] = selected
		if len(selected) < targetCount {
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return nil
		}
		var selectedCards []int
		if arr, ok := ctxData["selected_indices"].([]int); ok {
			selectedCards = append(selectedCards, arr...)
		} else if arr, ok := ctxData["selected_indices"].([]interface{}); ok {
			for _, v := range arr {
				if f, ok := v.(float64); ok {
					selectedCards = append(selectedCards, int(f))
				}
			}
		}
		xValue := toIntContextValue(ctxData["x_value"])
		if xValue <= 2 || len(selectedCards) != xValue {
			return fmt.Errorf("圣洁法典弃牌参数无效")
		}
		removed, err := removeCardsByIndicesFromHand(user, append([]int{}, selectedCards...))
		if err != nil {
			return err
		}
		e.NotifyCardRevealed(user.ID, removed, "discard")
		e.State.DiscardPile = append(e.State.DiscardPile, removed...)

		for _, tid := range selected {
			e.Heal(tid, 2)
		}
		damage := xValue - 1
		if damage > 0 {
			e.AddPendingDamage(model.PendingDamage{
				SourceID:   user.ID,
				TargetID:   user.ID,
				Damage:     damage,
				DamageType: "magic",
				Stage:      0,
			})
		}
		e.Log(fmt.Sprintf("%s 发动 [圣洁法典]：为%d名角色各+2治疗，并对自己造成%d点法术伤害", user.Name, len(selected), damage))
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			if len(e.State.PendingDamageQueue) > 0 {
				e.State.Phase = model.PhasePendingDamageResolution
				e.State.ReturnPhase = model.PhaseExtraAction
			} else {
				e.State.Phase = model.PhaseExtraAction
			}
		}
		return nil
	}
	if choiceType == "sage_magic_rebound_target" || choiceType == "sage_arcane_target" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		var targetIDs []string
		if arr, ok := ctxData["target_ids"].([]string); ok {
			targetIDs = append(targetIDs, arr...)
		} else if arr, ok := ctxData["target_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					targetIDs = append(targetIDs, s)
				}
			}
		}
		if selectionIndex < 0 || selectionIndex >= len(targetIDs) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		var selected []int
		if arr, ok := ctxData["selected_indices"].([]int); ok {
			selected = append(selected, arr...)
		} else if arr, ok := ctxData["selected_indices"].([]interface{}); ok {
			for _, v := range arr {
				if f, ok := v.(float64); ok {
					selected = append(selected, int(f))
				}
			}
		}
		xValue := toIntContextValue(ctxData["x_value"])
		if xValue <= 1 || len(selected) != xValue {
			return fmt.Errorf("弃牌参数无效")
		}
		removed, err := removeCardsByIndicesFromHand(user, append([]int{}, selected...))
		if err != nil {
			return err
		}
		e.NotifyCardRevealed(user.ID, removed, "discard")
		e.State.DiscardPile = append(e.State.DiscardPile, removed...)
		targetID := targetIDs[selectionIndex]
		target := e.State.Players[targetID]
		if target == nil {
			return fmt.Errorf("目标不存在")
		}
		switch choiceType {
		case "sage_magic_rebound_target":
			damageToTarget := xValue - 1
			damageToSelf := xValue
			var pds []model.PendingDamage
			if damageToTarget > 0 {
				pds = append(pds, model.PendingDamage{
					SourceID:   user.ID,
					TargetID:   targetID,
					Damage:     damageToTarget,
					DamageType: "magic",
					Stage:      0,
				})
			}
			if damageToSelf > 0 {
				pds = append(pds, model.PendingDamage{
					SourceID:   user.ID,
					TargetID:   user.ID,
					Damage:     damageToSelf,
					DamageType: "magic",
					Stage:      0,
				})
			}
			e.prependPendingDamages(pds)
			e.Log(fmt.Sprintf("%s 发动 [法术反弹]：弃%d张同系牌，对 %s 造成%d点法术伤害，并对自己造成%d点法术伤害", user.Name, xValue, target.Name, damageToTarget, damageToSelf))
		case "sage_arcane_target":
			damage := xValue - 1
			if damage > 0 {
				e.AddPendingDamage(model.PendingDamage{
					SourceID:   user.ID,
					TargetID:   targetID,
					Damage:     damage,
					DamageType: "magic",
					Stage:      0,
				})
				e.AddPendingDamage(model.PendingDamage{
					SourceID:   user.ID,
					TargetID:   user.ID,
					Damage:     damage,
					DamageType: "magic",
					Stage:      0,
				})
			}
			e.Log(fmt.Sprintf("%s 发动 [魔道法典]：弃%d张异系牌，对 %s 与自己各造成%d点法术伤害", user.Name, xValue, target.Name, damage))
		}
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			if len(e.State.PendingDamageQueue) > 0 {
				e.State.Phase = model.PhasePendingDamageResolution
				e.State.ReturnPhase = model.PhaseExtraAction
			} else {
				e.State.Phase = model.PhaseExtraAction
			}
		}
		return nil
	}
	if choiceType == "hb_holy_shard_combo" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		var combos []string
		if arr, ok := ctxData["combos"].([]string); ok {
			combos = append(combos, arr...)
		} else if arr, ok := ctxData["combos"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					combos = append(combos, s)
				}
			}
		}
		if selectionIndex < 0 || selectionIndex >= len(combos) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		choice := combos[selectionIndex]
		parts := strings.Split(choice, ":")
		if len(parts) != 2 {
			return fmt.Errorf("同系组合格式错误")
		}
		element := strings.TrimSpace(parts[0])
		idxParts := strings.Split(parts[1], ",")
		if len(idxParts) != 2 {
			return fmt.Errorf("同系组合索引格式错误")
		}
		i, err1 := strconv.Atoi(strings.TrimSpace(idxParts[0]))
		j, err2 := strconv.Atoi(strings.TrimSpace(idxParts[1]))
		if err1 != nil || err2 != nil || i < 0 || j < 0 || i >= len(user.Hand) || j >= len(user.Hand) || i == j {
			return fmt.Errorf("无效的弃牌索引")
		}
		c1 := user.Hand[i]
		c2 := user.Hand[j]
		if c1.Type != model.CardTypeAttack || c2.Type != model.CardTypeAttack || c1.Element != c2.Element {
			return fmt.Errorf("圣屑飓暴需要弃置2张同系攻击牌")
		}
		removed, err := removeCardsByIndicesFromHand(user, []int{i, j})
		if err != nil {
			return err
		}
		e.NotifyCardRevealed(user.ID, removed, "discard")
		e.State.DiscardPile = append(e.State.DiscardPile, removed...)
		ctxData["selected_element"] = element
		ctxData["choice_type"] = "hb_holy_shard_target"
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return nil
	}
	if choiceType == "hb_holy_shard_target" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		var targetIDs []string
		if arr, ok := ctxData["target_ids"].([]string); ok {
			targetIDs = append(targetIDs, arr...)
		} else if arr, ok := ctxData["target_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					targetIDs = append(targetIDs, s)
				}
			}
		}
		if selectionIndex < 0 || selectionIndex >= len(targetIDs) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		targetID := targetIDs[selectionIndex]
		target := e.State.Players[targetID]
		if target == nil {
			return fmt.Errorf("目标不存在")
		}
		if target.Camp == user.Camp {
			return fmt.Errorf("圣屑飓暴只能指定敌方目标")
		}
		eleStr, _ := ctxData["selected_element"].(string)
		ele := model.Element(eleStr)
		if ele == "" {
			return fmt.Errorf("圣屑飓暴攻击元素缺失")
		}
		virtualCard := model.Card{
			ID:          fmt.Sprintf("hb_holy_shard_%s_%d", user.ID, len(e.State.DiscardPile)+len(e.State.ActionQueue)+1),
			Name:        "圣屑飓暴",
			Type:        model.CardTypeAttack,
			Element:     ele,
			Faction:     "圣",
			Damage:      2,
			Description: "由圣屑飓暴视为的圣命格主动攻击",
		}
		if user.Tokens == nil {
			user.Tokens = map[string]int{}
		}
		user.Tokens["hb_shard_miss_pending"] = 1
		e.State.ActionQueue = append(e.State.ActionQueue, model.QueuedAction{
			SourceID:    user.ID,
			TargetID:    target.ID,
			Type:        model.ActionAttack,
			Element:     ele,
			Card:        &virtualCard,
			CardIndex:   -1,
			SourceSkill: "hb_holy_shard_storm",
		})
		e.Log(fmt.Sprintf("%s 发动 [圣屑飓暴]：对 %s 发起1次%s系圣命格主动攻击", user.Name, target.Name, ele))
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			e.State.Phase = model.PhaseBeforeAction
		}
		return nil
	}
	if choiceType == "hb_holy_shard_miss_confirm" {
		if selectionIndex == 1 {
			e.PopInterrupt()
			if e.State.PendingInterrupt == nil {
				if len(e.State.PendingDamageQueue) > 0 {
					e.State.Phase = model.PhasePendingDamageResolution
				} else if len(e.State.CombatStack) > 0 {
					e.State.Phase = model.PhaseCombatInteraction
				} else {
					e.State.Phase = model.PhaseExtraAction
				}
			}
			return nil
		}
		if selectionIndex != 0 {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		maxX := toIntContextValue(ctxData["max_x"])
		if maxX <= 0 {
			e.PopInterrupt()
			return nil
		}
		ctxData["choice_type"] = "hb_holy_shard_miss_x"
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return nil
	}
	if choiceType == "hb_holy_shard_miss_x" {
		maxX := toIntContextValue(ctxData["max_x"])
		xValue := selectionIndex + 1
		if xValue < 1 || xValue > maxX {
			return fmt.Errorf("无效的X值")
		}
		ctxData["x_value"] = xValue
		ctxData["choice_type"] = "hb_holy_shard_miss_ally_target"
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return nil
	}
	if choiceType == "hb_holy_shard_miss_ally_target" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		var allyIDs []string
		if arr, ok := ctxData["ally_ids"].([]string); ok {
			allyIDs = append(allyIDs, arr...)
		} else if arr, ok := ctxData["ally_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					allyIDs = append(allyIDs, s)
				}
			}
		}
		if selectionIndex < 0 || selectionIndex >= len(allyIDs) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		targetID := allyIDs[selectionIndex]
		target := e.State.Players[targetID]
		if target == nil {
			return fmt.Errorf("目标队友不存在")
		}
		xValue := toIntContextValue(ctxData["x_value"])
		if xValue <= 0 {
			return fmt.Errorf("无效的X值")
		}
		if user.Heal < xValue {
			return fmt.Errorf("治疗不足，无法移除%d点治疗", xValue)
		}
		user.Heal -= xValue
		discardNeed := xValue
		if len(target.Hand) < discardNeed {
			discardNeed = len(target.Hand)
		}
		e.Log(fmt.Sprintf("%s 的 [圣屑飓暴] 未命中分支生效：移除%d点治疗，指定 %s 弃置%d张手牌", user.Name, xValue, target.Name, discardNeed))
		e.PopInterrupt()
		if discardNeed > 0 {
			e.PushInterrupt(&model.Interrupt{
				Type:     model.InterruptDiscard,
				PlayerID: target.ID,
				Context: map[string]interface{}{
					"discard_count":        discardNeed,
					"prompt":               fmt.Sprintf("【圣屑飓暴】请弃置%d张手牌：", discardNeed),
					"stay_in_turn":         true,
					"is_damage_resolution": true,
				},
			})
			return nil
		}
		if e.State.PendingInterrupt == nil {
			if len(e.State.PendingDamageQueue) > 0 {
				e.State.Phase = model.PhasePendingDamageResolution
			} else {
				e.State.Phase = model.PhaseExtraAction
			}
		}
		return nil
	}
	if choiceType == "hb_radiant_descent_cost" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		var modes []string
		if arr, ok := ctxData["cost_modes"].([]string); ok {
			modes = append(modes, arr...)
		} else if arr, ok := ctxData["cost_modes"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					modes = append(modes, s)
				}
			}
		}
		if selectionIndex < 0 || selectionIndex >= len(modes) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		switch modes[selectionIndex] {
		case "heal":
			if user.Heal < 2 {
				return fmt.Errorf("治疗不足2点")
			}
			user.Heal -= 2
		case "faith":
			if holyBowFaith(user) < 2 {
				return fmt.Errorf("信仰不足2点")
			}
			addHolyBowFaith(user, -2)
		default:
			return fmt.Errorf("无效的支付方式")
		}
		if user.Tokens == nil {
			user.Tokens = map[string]int{}
		}
		user.Tokens["hb_form"] = 1
		user.TurnState.PendingActions = append(user.TurnState.PendingActions, model.ActionContext{
			Source:   "圣煌降临",
			MustType: "Magic",
		})
		e.Log(fmt.Sprintf("%s 发动 [圣煌降临]：进入圣煌形态并获得额外法术行动", user.Name))
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			e.State.Phase = model.PhaseExtraAction
		}
		return nil
	}
	if choiceType == "hb_light_burst_mode" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		var allyIDs []string
		if arr, ok := ctxData["ally_ids"].([]string); ok {
			allyIDs = append(allyIDs, arr...)
		} else if arr, ok := ctxData["ally_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					allyIDs = append(allyIDs, s)
				}
			}
		}
		var enemyIDs []string
		if arr, ok := ctxData["enemy_ids"].([]string); ok {
			enemyIDs = append(enemyIDs, arr...)
		} else if arr, ok := ctxData["enemy_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					enemyIDs = append(enemyIDs, s)
				}
			}
		}
		maxX := toIntContextValue(ctxData["max_x"])
		var modeOrder []string
		if user.Heal >= 1 && len(allyIDs) > 0 {
			modeOrder = append(modeOrder, "a")
		}
		if maxX > 0 && len(enemyIDs) > 0 {
			handCount := len(user.Hand)
			canB := false
			for x := 1; x <= maxX; x++ {
				limit := handCount - x
				for _, eid := range enemyIDs {
					if ep := e.State.Players[eid]; ep != nil && len(ep.Hand) <= limit {
						canB = true
						break
					}
				}
				if canB {
					break
				}
			}
			if canB {
				modeOrder = append(modeOrder, "b")
			}
		}
		if selectionIndex < 0 || selectionIndex >= len(modeOrder) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		switch modeOrder[selectionIndex] {
		case "a":
			ctxData["choice_type"] = "hb_light_burst_mode_a_target"
		case "b":
			ctxData["choice_type"] = "hb_light_burst_mode_b_x"
		default:
			return fmt.Errorf("无效的分支")
		}
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return nil
	}
	if choiceType == "hb_light_burst_mode_a_target" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		var allyIDs []string
		if arr, ok := ctxData["ally_ids"].([]string); ok {
			allyIDs = append(allyIDs, arr...)
		} else if arr, ok := ctxData["ally_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					allyIDs = append(allyIDs, s)
				}
			}
		}
		if selectionIndex < 0 || selectionIndex >= len(allyIDs) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		targetID := allyIDs[selectionIndex]
		if user.Heal < 1 {
			return fmt.Errorf("治疗不足，无法发动分支①")
		}
		e.DrawCards(user.ID, 1)
		user.Heal--
		faith := addHolyBowFaith(user, 1)
		e.Heal(targetID, 1)
		target := e.State.Players[targetID]
		targetName := targetID
		if target != nil {
			targetName = target.Name
		}
		e.Log(fmt.Sprintf("%s 的 [圣光爆裂] 分支①生效：摸1、移除1治疗、信仰+1（当前%d），%s +1治疗", user.Name, faith, targetName))
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			if len(e.State.PendingDamageQueue) > 0 {
				e.State.Phase = model.PhasePendingDamageResolution
				e.State.ReturnPhase = model.PhaseExtraAction
			} else {
				e.State.Phase = model.PhaseExtraAction
			}
		}
		return nil
	}
	if choiceType == "hb_light_burst_mode_b_x" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		var enemyIDs []string
		if arr, ok := ctxData["enemy_ids"].([]string); ok {
			enemyIDs = append(enemyIDs, arr...)
		} else if arr, ok := ctxData["enemy_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					enemyIDs = append(enemyIDs, s)
				}
			}
		}
		maxX := toIntContextValue(ctxData["max_x"])
		validX := make([]int, 0)
		for x := 1; x <= maxX; x++ {
			limit := len(user.Hand) - x
			eligible := 0
			for _, eid := range enemyIDs {
				if ep := e.State.Players[eid]; ep != nil && len(ep.Hand) <= limit {
					eligible++
				}
			}
			if eligible > 0 {
				validX = append(validX, x)
			}
		}
		if selectionIndex < 0 || selectionIndex >= len(validX) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		xValue := validX[selectionIndex]
		limit := len(user.Hand) - xValue
		candidateTargets := make([]string, 0)
		for _, eid := range enemyIDs {
			if ep := e.State.Players[eid]; ep != nil && len(ep.Hand) <= limit {
				candidateTargets = append(candidateTargets, eid)
			}
		}
		if len(candidateTargets) == 0 {
			return fmt.Errorf("没有满足手牌条件的目标")
		}
		ctxData["x_value"] = xValue
		ctxData["eligible_count"] = len(candidateTargets)
		ctxData["candidate_target_ids"] = candidateTargets
		ctxData["choice_type"] = "hb_light_burst_mode_b_target_count"
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return nil
	}
	if choiceType == "hb_light_burst_mode_b_target_count" {
		xValue := toIntContextValue(ctxData["x_value"])
		eligibleCount := toIntContextValue(ctxData["eligible_count"])
		maxCount := xValue
		if eligibleCount < maxCount {
			maxCount = eligibleCount
		}
		targetCount := selectionIndex + 1
		if targetCount < 1 || targetCount > maxCount {
			return fmt.Errorf("无效的目标数量")
		}
		ctxData["target_count"] = targetCount
		ctxData["selected_target_ids"] = []string{}
		ctxData["choice_type"] = "hb_light_burst_mode_b_targets"
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return nil
	}
	if choiceType == "hb_light_burst_mode_b_targets" {
		var candidates []string
		if arr, ok := ctxData["candidate_target_ids"].([]string); ok {
			candidates = append(candidates, arr...)
		} else if arr, ok := ctxData["candidate_target_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					candidates = append(candidates, s)
				}
			}
		}
		targetCount := toIntContextValue(ctxData["target_count"])
		if targetCount <= 0 {
			return fmt.Errorf("目标数量无效")
		}
		var selected []string
		selectedSet := map[string]bool{}
		if arr, ok := ctxData["selected_target_ids"].([]string); ok {
			for _, s := range arr {
				if s != "" && !selectedSet[s] {
					selected = append(selected, s)
					selectedSet[s] = true
				}
			}
		} else if arr, ok := ctxData["selected_target_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok && s != "" && !selectedSet[s] {
					selected = append(selected, s)
					selectedSet[s] = true
				}
			}
		}
		var remaining []string
		for _, tid := range candidates {
			if !selectedSet[tid] {
				remaining = append(remaining, tid)
			}
		}
		if selectionIndex < 0 || selectionIndex >= len(remaining) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		selected = append(selected, remaining[selectionIndex])
		ctxData["selected_target_ids"] = selected
		if len(selected) < targetCount {
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return nil
		}
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		ctxData["selected_indices"] = []int{}
		ctxData["remaining_indices"] = allHandIndices(user)
		ctxData["choice_type"] = "hb_light_burst_mode_b_discard"
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return nil
	}
	if choiceType == "hb_light_burst_mode_b_discard" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		xValue := toIntContextValue(ctxData["x_value"])
		if xValue <= 0 {
			return fmt.Errorf("X值无效")
		}
		var remaining []int
		if arr, ok := ctxData["remaining_indices"].([]int); ok {
			remaining = append(remaining, arr...)
		} else if arr, ok := ctxData["remaining_indices"].([]interface{}); ok {
			for _, v := range arr {
				if f, ok := v.(float64); ok {
					remaining = append(remaining, int(f))
				}
			}
		}
		var selected []int
		if arr, ok := ctxData["selected_indices"].([]int); ok {
			selected = append(selected, arr...)
		} else if arr, ok := ctxData["selected_indices"].([]interface{}); ok {
			for _, v := range arr {
				if f, ok := v.(float64); ok {
					selected = append(selected, int(f))
				}
			}
		}
		cardIdx, ok := resolveSelectionToCandidate(selectionIndex, remaining)
		if !ok || cardIdx < 0 || cardIdx >= len(user.Hand) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		selected = append(selected, cardIdx)
		var nextRemaining []int
		for _, idx := range remaining {
			if idx != cardIdx {
				nextRemaining = append(nextRemaining, idx)
			}
		}
		if len(selected) < xValue {
			ctxData["selected_indices"] = selected
			ctxData["remaining_indices"] = nextRemaining
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return nil
		}
		if user.Heal < xValue {
			return fmt.Errorf("治疗不足，无法移除%d点治疗", xValue)
		}
		removed, err := removeCardsByIndicesFromHand(user, append([]int{}, selected...))
		if err != nil {
			return err
		}
		e.NotifyCardRevealed(user.ID, removed, "discard")
		e.State.DiscardPile = append(e.State.DiscardPile, removed...)
		user.Heal -= xValue

		var targetIDs []string
		if arr, ok := ctxData["selected_target_ids"].([]string); ok {
			targetIDs = append(targetIDs, arr...)
		} else if arr, ok := ctxData["selected_target_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					targetIDs = append(targetIDs, s)
				}
			}
		}
		y := 0
		for _, tid := range targetIDs {
			if p := e.State.Players[tid]; p != nil && p.Heal > 0 {
				y++
			}
		}
		damage := y + 2
		for _, tid := range targetIDs {
			if e.State.Players[tid] == nil {
				continue
			}
			e.AddPendingDamage(model.PendingDamage{
				SourceID:   user.ID,
				TargetID:   tid,
				Damage:     damage,
				DamageType: "Attack",
				Stage:      0,
			})
		}
		e.Log(fmt.Sprintf("%s 的 [圣光爆裂] 分支②生效：移除%d治疗并弃%d张牌，对%d名目标各造成%d点攻击伤害（Y=%d）", user.Name, xValue, xValue, len(targetIDs), damage, y))
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			if len(e.State.PendingDamageQueue) > 0 {
				e.State.Phase = model.PhasePendingDamageResolution
				e.State.ReturnPhase = model.PhaseExtraAction
			} else {
				e.State.Phase = model.PhaseExtraAction
			}
		}
		return nil
	}
	if choiceType == "hb_meteor_bullet_cost" {
		var modes []string
		if arr, ok := ctxData["cost_modes"].([]string); ok {
			modes = append(modes, arr...)
		} else if arr, ok := ctxData["cost_modes"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					modes = append(modes, s)
				}
			}
		}
		if selectionIndex < 0 || selectionIndex >= len(modes) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		ctxData["chosen_cost_mode"] = modes[selectionIndex]
		ctxData["choice_type"] = "hb_meteor_bullet_target"
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return nil
	}
	if choiceType == "hb_meteor_bullet_target" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		var allyIDs []string
		if arr, ok := ctxData["ally_ids"].([]string); ok {
			allyIDs = append(allyIDs, arr...)
		} else if arr, ok := ctxData["ally_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					allyIDs = append(allyIDs, s)
				}
			}
		}
		if selectionIndex < 0 || selectionIndex >= len(allyIDs) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		targetID := allyIDs[selectionIndex]
		mode, _ := ctxData["chosen_cost_mode"].(string)
		switch mode {
		case "heal":
			if user.Heal <= 0 {
				return fmt.Errorf("治疗不足，无法发动流星圣弹")
			}
			user.Heal--
		case "faith":
			if holyBowFaith(user) <= 0 {
				return fmt.Errorf("信仰不足，无法发动流星圣弹")
			}
			addHolyBowFaith(user, -1)
		default:
			return fmt.Errorf("流星圣弹资源选择无效")
		}
		e.Heal(targetID, 1)
		target := e.State.Players[targetID]
		targetName := targetID
		if target != nil {
			targetName = target.Name
		}
		e.Log(fmt.Sprintf("%s 发动 [流星圣弹]：移除1点%s，令 %s +1治疗", user.Name, map[string]string{"heal": "治疗", "faith": "信仰"}[mode], targetName))
		rawCtx, _ := ctxData["user_ctx"].(*model.Context)
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			if rawCtx != nil && rawCtx.Trigger == model.TriggerOnAttackStart {
				if len(e.State.ActionQueue) > 0 {
					e.State.Phase = model.PhaseBeforeAction
				} else if len(e.State.CombatStack) > 0 {
					e.State.Phase = model.PhaseCombatInteraction
				} else {
					e.State.Phase = model.PhaseTurnEnd
				}
			} else {
				e.State.Phase = model.PhaseExtraAction
			}
		}
		return nil
	}
	if choiceType == "hb_radiant_cannon_side" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		if !e.isHolyBow(user) {
			return fmt.Errorf("仅圣弓可发动圣煌辉光炮")
		}
		if user.Tokens == nil {
			user.Tokens = map[string]int{}
		}
		requiredFaith := toIntContextValue(ctxData["required_faith"])
		if requiredFaith <= 0 {
			requiredFaith = 4
		}
		if holyBowCannon(user) <= 0 {
			return fmt.Errorf("圣煌辉光炮指示物不足")
		}
		if holyBowFaith(user) < requiredFaith {
			return fmt.Errorf("信仰不足，需要%d点", requiredFaith)
		}
		if selectionIndex != 0 && selectionIndex != 1 {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		user.Tokens["hb_cannon"] = holyBowCannon(user) - 1
		addHolyBowFaith(user, -requiredFaith)

		for _, pid := range e.State.PlayerOrder {
			p := e.State.Players[pid]
			if p == nil {
				continue
			}
			if len(p.Hand) > 4 {
				discarded := append([]model.Card{}, p.Hand[4:]...)
				p.Hand = append([]model.Card{}, p.Hand[:4]...)
				e.NotifyCardRevealed(p.ID, discarded, "discard")
				e.State.DiscardPile = append(e.State.DiscardPile, discarded...)
			} else if len(p.Hand) < 4 {
				drawN := 4 - len(p.Hand)
				cards, newDeck, newDiscard := rules.DrawCards(e.State.Deck, e.State.DiscardPile, drawN)
				e.State.Deck = newDeck
				e.State.DiscardPile = newDiscard
				p.Hand = append(p.Hand, cards...)
				e.NotifyDrawCards(p.ID, drawN, "hb_radiant_cannon_adjust")
			}
		}
		if user.Camp == model.RedCamp {
			e.State.RedCups++
			if e.State.RedCups > 5 {
				e.State.RedCups = 5
			}
		} else {
			e.State.BlueCups++
			if e.State.BlueCups > 5 {
				e.State.BlueCups = 5
			}
		}
		if selectionIndex == 0 {
			e.State.RedMorale = e.State.BlueMorale
		} else {
			e.State.BlueMorale = e.State.RedMorale
		}
		e.Log(fmt.Sprintf("%s 发动 [圣煌辉光炮]：全员手牌调整至4，我方星杯+1，并完成士气对齐", user.Name))
		e.checkGameEnd()
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			if len(e.State.PendingDamageQueue) > 0 {
				e.State.Phase = model.PhasePendingDamageResolution
				e.State.ReturnPhase = model.PhaseExtraAction
			} else {
				e.State.Phase = model.PhaseExtraAction
			}
		}
		return nil
	}
	if choiceType == "hb_auto_fill_resource" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		var modes []string
		if arr, ok := ctxData["resource_modes"].([]string); ok {
			modes = append(modes, arr...)
		} else if arr, ok := ctxData["resource_modes"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					modes = append(modes, s)
				}
			}
		}
		if selectionIndex < 0 || selectionIndex >= len(modes) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		branch := modes[selectionIndex]
		switch branch {
		case "crystal":
			if !e.ConsumeCrystalCost(user.ID, 1) {
				return fmt.Errorf("自动填充分支①需要1点蓝水晶（红宝石可替代）")
			}
		case "gem":
			if user.Gem <= 0 {
				return fmt.Errorf("自动填充分支②需要1点红宝石")
			}
			user.Gem--
			maxEnergy := e.getPlayerEnergyCap(user)
			if user.Gem+user.Crystal < maxEnergy {
				user.Crystal++
				if user.Gem+user.Crystal > maxEnergy {
					user.Crystal -= (user.Gem + user.Crystal - maxEnergy)
				}
			}
		default:
			return fmt.Errorf("无效分支")
		}
		ctxData["branch"] = branch
		ctxData["choice_type"] = "hb_auto_fill_gain"
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return nil
	}
	if choiceType == "hb_auto_fill_gain" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		branch, _ := ctxData["branch"].(string)
		if selectionIndex != 0 && selectionIndex != 1 {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		if branch == "gem" {
			if selectionIndex == 0 {
				now := addHolyBowFaith(user, 2)
				e.Log(fmt.Sprintf("%s 的 [自动填充] 分支②生效：+2信仰（当前%d）", user.Name, now))
			} else {
				e.Heal(user.ID, 2)
				e.Log(fmt.Sprintf("%s 的 [自动填充] 分支②生效：+2治疗", user.Name))
			}
		} else {
			if selectionIndex == 0 {
				now := addHolyBowFaith(user, 1)
				e.Log(fmt.Sprintf("%s 的 [自动填充] 分支①生效：+1信仰（当前%d）", user.Name, now))
			} else {
				e.Heal(user.ID, 1)
				e.Log(fmt.Sprintf("%s 的 [自动填充] 分支①生效：+1治疗", user.Name))
			}
		}
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			e.State.Phase = model.PhaseTurnEnd
		}
		return nil
	}
	if choiceType == "ss_convert_color" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		var modeOrder []string
		if arr, ok := ctxData["mode_order"].([]string); ok {
			modeOrder = append(modeOrder, arr...)
		} else if arr, ok := ctxData["mode_order"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					modeOrder = append(modeOrder, s)
				}
			}
		}
		if selectionIndex < 0 || selectionIndex >= len(modeOrder) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		mode := modeOrder[selectionIndex]
		switch mode {
		case "y2b":
			if soulSorcererYellow(user) <= 0 {
				return fmt.Errorf("黄色灵魂不足")
			}
			if soulSorcererBlue(user) >= soulSorcererBlueCapEngine {
				return fmt.Errorf("蓝色灵魂已满")
			}
			addSoulSorcererYellow(user, -1)
			addSoulSorcererBlue(user, 1)
			e.Log(fmt.Sprintf("%s 的 [灵魂转换] 生效：黄魂-1，蓝魂+1（黄:%d 蓝:%d）", user.Name, soulSorcererYellow(user), soulSorcererBlue(user)))
		case "b2y":
			if soulSorcererBlue(user) <= 0 {
				return fmt.Errorf("蓝色灵魂不足")
			}
			if soulSorcererYellow(user) >= soulSorcererYellowCapEngine {
				return fmt.Errorf("黄色灵魂已满")
			}
			addSoulSorcererBlue(user, -1)
			addSoulSorcererYellow(user, 1)
			e.Log(fmt.Sprintf("%s 的 [灵魂转换] 生效：蓝魂-1，黄魂+1（黄:%d 蓝:%d）", user.Name, soulSorcererYellow(user), soulSorcererBlue(user)))
		default:
			return fmt.Errorf("无效的灵魂转换模式")
		}
		rawCtx, _ := ctxData["user_ctx"].(*model.Context)
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			if rawCtx != nil && rawCtx.Trigger == model.TriggerOnAttackStart {
				if len(e.State.ActionQueue) > 0 {
					e.State.Phase = model.PhaseBeforeAction
				} else if len(e.State.CombatStack) > 0 {
					e.State.Phase = model.PhaseCombatInteraction
				} else {
					e.State.Phase = model.PhaseTurnEnd
				}
			} else {
				e.State.Phase = model.PhaseResponse
			}
		}
		return nil
	}
	if choiceType == "ss_link_target" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		var allyIDs []string
		if arr, ok := ctxData["ally_ids"].([]string); ok {
			allyIDs = append(allyIDs, arr...)
		} else if arr, ok := ctxData["ally_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					allyIDs = append(allyIDs, s)
				}
			}
		}
		if selectionIndex < 0 || selectionIndex >= len(allyIDs) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		target := e.State.Players[allyIDs[selectionIndex]]
		if target == nil {
			return fmt.Errorf("目标队友不存在")
		}
		if target.Camp != user.Camp || target.ID == user.ID {
			return fmt.Errorf("灵魂链接只能指定队友")
		}
		if soulSorcererYellow(user) < 1 || soulSorcererBlue(user) < 1 {
			return fmt.Errorf("灵魂不足，无法放置灵魂链接")
		}
		if user.Character == nil {
			return fmt.Errorf("角色信息缺失")
		}
		linkCard, ok := user.ConsumeExclusiveCard(user.Character.Name, "灵魂链接")
		if !ok {
			return fmt.Errorf("未找到【灵魂链接】专属技能卡")
		}
		addSoulSorcererYellow(user, -1)
		addSoulSorcererBlue(user, -1)
		if err := e.placeSoulLink(user, target, linkCard); err != nil {
			user.RestoreExclusiveCard(linkCard)
			addSoulSorcererYellow(user, 1)
			addSoulSorcererBlue(user, 1)
			return err
		}
		e.Log(fmt.Sprintf("%s 发动 [灵魂链接]：移除1黄魂+1蓝魂，并将灵魂链接放置于 %s 面前", user.Name, target.Name))
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			e.State.Phase = model.PhaseStartup
		}
		return nil
	}
	if choiceType == "ss_link_transfer_x" {
		sorcererID, _ := ctxData["sorcerer_id"].(string)
		sorcerer := e.State.Players[sorcererID]
		if sorcerer == nil {
			return fmt.Errorf("灵魂术士不存在")
		}
		damageIdx := toIntContextValue(ctxData["damage_index"])
		if damageIdx < 0 || damageIdx >= len(e.State.PendingDamageQueue) {
			return fmt.Errorf("伤害上下文不存在")
		}
		pd := &e.State.PendingDamageQueue[damageIdx]
		sourceID, _ := ctxData["source_id"].(string)
		targetID, _ := ctxData["target_id"].(string)
		if sourceID != "" && sourceID != pd.SourceID {
			return fmt.Errorf("伤害来源已变化")
		}
		if targetID != "" && targetID != pd.TargetID {
			return fmt.Errorf("伤害目标已变化")
		}
		maxX := toIntContextValue(ctxData["max_x"])
		if maxX < 0 {
			maxX = 0
		}
		x := selectionIndex
		if x < 0 || x > maxX {
			return fmt.Errorf("无效的X值")
		}
		if x > soulSorcererBlue(sorcerer) {
			x = soulSorcererBlue(sorcerer)
		}
		if x > pd.Damage {
			x = pd.Damage
		}
		counterpartID, _ := ctxData["counterpart_id"].(string)
		counterpart := e.State.Players[counterpartID]
		if x > 0 && counterpart != nil {
			addSoulSorcererBlue(sorcerer, -x)
			pd.Damage -= x
			if pd.Damage < 0 {
				pd.Damage = 0
			}
			e.AddPendingDamage(model.PendingDamage{
				SourceID:     pd.SourceID,
				TargetID:     counterpart.ID,
				Damage:       x,
				DamageType:   "magic",
				Stage:        0,
				FromSoulLink: true,
			})
			e.Log(fmt.Sprintf("%s 的 [灵魂链接] 生效：移除%d点蓝魂，将%d点伤害转移给 %s（法术伤害）", sorcerer.Name, x, x, counterpart.Name))
		} else {
			e.Log(fmt.Sprintf("%s 的 [灵魂链接] 选择不转移伤害", sorcerer.Name))
		}
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			e.State.Phase = model.PhasePendingDamageResolution
		}
		return nil
	}
	if choiceType == "ss_recall_pick" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		var remaining []int
		if arr, ok := ctxData["remaining_indices"].([]int); ok {
			remaining = append(remaining, arr...)
		} else if arr, ok := ctxData["remaining_indices"].([]interface{}); ok {
			for _, v := range arr {
				if f, ok := v.(float64); ok {
					remaining = append(remaining, int(f))
				}
			}
		}
		var selected []int
		if arr, ok := ctxData["selected_indices"].([]int); ok {
			selected = append(selected, arr...)
		} else if arr, ok := ctxData["selected_indices"].([]interface{}); ok {
			for _, v := range arr {
				if f, ok := v.(float64); ok {
					selected = append(selected, int(f))
				}
			}
		}
		// 兼容两种前端：选项序号或直接 option ID=-1
		if selectionIndex == -1 || selectionIndex == 0 {
			if len(selected) == 0 {
				return fmt.Errorf("灵魂召还至少选择1张法术牌")
			}
			removed, err := removeCardsByIndicesFromHand(user, append([]int{}, selected...))
			if err != nil {
				return err
			}
			e.NotifyCardRevealed(user.ID, removed, "discard")
			e.State.DiscardPile = append(e.State.DiscardPile, removed...)
			gain := len(removed)
			before := soulSorcererBlue(user)
			after := addSoulSorcererBlue(user, gain)
			e.Log(fmt.Sprintf("%s 发动 [灵魂召还]：弃置%d张法术牌，蓝色灵魂 +%d（%d→%d）", user.Name, gain, gain, before, after))
			e.PopInterrupt()
			if e.State.PendingInterrupt == nil {
				if len(e.State.PendingDamageQueue) > 0 {
					e.State.Phase = model.PhasePendingDamageResolution
					e.State.ReturnPhase = model.PhaseExtraAction
				} else {
					e.State.Phase = model.PhaseExtraAction
				}
			}
			return nil
		}
		cardIdx := -1
		if selectionIndex >= 1 && selectionIndex <= len(remaining) {
			cardIdx = remaining[selectionIndex-1]
		} else {
			for _, idx := range remaining {
				if idx == selectionIndex {
					cardIdx = idx
					break
				}
			}
		}
		if cardIdx < 0 || cardIdx >= len(user.Hand) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		if user.Hand[cardIdx].Type != model.CardTypeMagic {
			return fmt.Errorf("灵魂召还只能选择法术牌")
		}
		selected = append(selected, cardIdx)
		nextRemaining := make([]int, 0, len(remaining))
		for _, idx := range remaining {
			if idx != cardIdx {
				nextRemaining = append(nextRemaining, idx)
			}
		}
		ctxData["selected_indices"] = selected
		ctxData["remaining_indices"] = nextRemaining
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return nil
	}
	if choiceType == "bp_shared_life_target" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		var targetIDs []string
		if arr, ok := ctxData["target_ids"].([]string); ok {
			targetIDs = append(targetIDs, arr...)
		} else if arr, ok := ctxData["target_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					targetIDs = append(targetIDs, s)
				}
			}
		}
		if selectionIndex < 0 || selectionIndex >= len(targetIDs) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		target := e.State.Players[targetIDs[selectionIndex]]
		if target == nil {
			return fmt.Errorf("同生共死目标不存在")
		}
		if user.Character == nil {
			return fmt.Errorf("角色信息缺失")
		}
		linkCard, ok := user.ConsumeExclusiveCard(user.Character.Name, "同生共死")
		if !ok {
			return fmt.Errorf("未找到【同生共死】专属技能卡")
		}

		// 先移除旧目标，再摸2张牌；放置动作通过延迟后续在爆牌结算后执行。
		_ = e.removeBloodPriestessSharedLife(user, false)
		cards, newDeck, newDiscard := rules.DrawCards(e.State.Deck, e.State.DiscardPile, 2)
		e.State.Deck = newDeck
		e.State.DiscardPile = newDiscard
		user.Hand = append(user.Hand, cards...)
		e.NotifyDrawCards(user.ID, 2, "bp_shared_life_draw")
		e.enqueueDeferredFollowup(model.DeferredFollowup{
			Type:      "blood_priestess_shared_life_place",
			UserID:    user.ID,
			TargetIDs: []string{target.ID},
			Data: map[string]interface{}{
				"card": linkCard,
			},
		})
		e.checkHandLimit(user, nil)
		e.Log(fmt.Sprintf("%s 发动 [同生共死]：先摸2张牌，待爆牌结算后放置于 %s 面前", user.Name, target.Name))

		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			if len(e.State.PendingDamageQueue) > 0 {
				e.State.Phase = model.PhasePendingDamageResolution
				e.State.ReturnPhase = model.PhaseExtraAction
			} else {
				e.State.Phase = model.PhaseExtraAction
			}
		}
		return nil
	}
	if choiceType == "bp_blood_sorrow_mode" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		if selectionIndex < 0 || selectionIndex > 1 {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		if !toBoolContextValue(ctxData["damage_queued"]) {
			e.AddPendingDamage(model.PendingDamage{
				SourceID:   user.ID,
				TargetID:   user.ID,
				Damage:     2,
				DamageType: "magic",
				Stage:      0,
			})
			ctxData["damage_queued"] = true
		}
		if selectionIndex == 0 {
			var targetIDs []string
			if arr, ok := ctxData["target_ids"].([]string); ok {
				targetIDs = append(targetIDs, arr...)
			} else if arr, ok := ctxData["target_ids"].([]interface{}); ok {
				for _, v := range arr {
					if s, ok := v.(string); ok {
						targetIDs = append(targetIDs, s)
					}
				}
			}
			if len(targetIDs) == 0 {
				return fmt.Errorf("无可转移目标")
			}
			ctxData["choice_type"] = "bp_blood_sorrow_target"
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return nil
		}

		if !e.removeBloodPriestessSharedLife(user, true) {
			return fmt.Errorf("当前没有可移除的同生共死")
		}
		e.Log(fmt.Sprintf("%s 发动 [血之哀伤]：对自己造成2点法术伤害，并移除【同生共死】", user.Name))
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			e.State.Phase = model.PhasePendingDamageResolution
			e.State.ReturnPhase = model.PhaseStartup
		}
		return nil
	}
	if choiceType == "bp_blood_sorrow_target" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		var targetIDs []string
		if arr, ok := ctxData["target_ids"].([]string); ok {
			targetIDs = append(targetIDs, arr...)
		} else if arr, ok := ctxData["target_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					targetIDs = append(targetIDs, s)
				}
			}
		}
		if selectionIndex < 0 || selectionIndex >= len(targetIDs) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		target := e.State.Players[targetIDs[selectionIndex]]
		if target == nil {
			return fmt.Errorf("转移目标不存在")
		}
		holder, fc := e.findBloodPriestessSharedLife(user)
		if holder == nil || fc == nil {
			return fmt.Errorf("当前没有可转移的同生共死")
		}
		card := fc.Card
		holder.RemoveFieldCard(fc)
		if err := e.placeBloodPriestessSharedLife(user, target, card); err != nil {
			return err
		}
		e.Log(fmt.Sprintf("%s 发动 [血之哀伤]：对自己造成2点法术伤害，并将【同生共死】转移至 %s", user.Name, target.Name))
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			e.State.Phase = model.PhasePendingDamageResolution
			e.State.ReturnPhase = model.PhaseStartup
		}
		return nil
	}
	if choiceType == "bp_blood_wail_x" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		targetID, _ := ctxData["target_id"].(string)
		target := e.State.Players[targetID]
		if target == nil {
			return fmt.Errorf("目标角色不存在")
		}
		if selectionIndex < 0 || selectionIndex > 2 {
			return fmt.Errorf("无效的X值")
		}
		damage := selectionIndex + 1
		_ = e.maybeAutoReleaseBloodPriestessByHand(user, "手牌<3强制脱离流血形态")
		queueBefore := len(e.State.InterruptQueue)
		e.checkHandLimit(user, nil)
		if len(e.State.InterruptQueue) > queueBefore {
			e.enqueueDeferredFollowup(model.DeferredFollowup{
				Type:      "blood_priestess_wail_damage",
				UserID:    user.ID,
				TargetIDs: []string{target.ID},
				Data: map[string]interface{}{
					"damage": damage,
				},
			})
			e.Log(fmt.Sprintf("%s 的 [血之悲鸣] 延迟：先结算手牌上限变化，再造成伤害", user.Name))
			e.PopInterrupt()
			return nil
		}
		e.AddPendingDamage(model.PendingDamage{
			SourceID:   user.ID,
			TargetID:   target.ID,
			Damage:     damage,
			DamageType: "magic",
			Stage:      0,
		})
		e.AddPendingDamage(model.PendingDamage{
			SourceID:   user.ID,
			TargetID:   user.ID,
			Damage:     damage,
			DamageType: "magic",
			Stage:      0,
		})
		e.Log(fmt.Sprintf("%s 发动 [血之悲鸣]：对 %s 和自己各造成%d点法术伤害", user.Name, target.Name, damage))
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			e.State.Phase = model.PhasePendingDamageResolution
			e.State.ReturnPhase = model.PhaseExtraAction
		}
		return nil
	}
	if choiceType == "bp_curse_discard" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		discardNeed := toIntContextValue(ctxData["discard_count"])
		if discardNeed < 0 {
			discardNeed = 0
		}
		var selected []int
		if arr, ok := ctxData["selected_indices"].([]int); ok {
			selected = append(selected, arr...)
		} else if arr, ok := ctxData["selected_indices"].([]interface{}); ok {
			for _, v := range arr {
				if f, ok := v.(float64); ok {
					selected = append(selected, int(f))
				}
			}
		}
		var remaining []int
		if arr, ok := ctxData["remaining_indices"].([]int); ok {
			remaining = append(remaining, arr...)
		} else if arr, ok := ctxData["remaining_indices"].([]interface{}); ok {
			for _, v := range arr {
				if f, ok := v.(float64); ok {
					remaining = append(remaining, int(f))
				}
			}
		}
		if selectionIndex == -1 {
			if len(selected) < discardNeed {
				return fmt.Errorf("还需选择 %d 张弃牌", discardNeed-len(selected))
			}
			sort.Sort(sort.Reverse(sort.IntSlice(selected)))
			var discarded []model.Card
			for _, idx := range selected {
				if idx < 0 || idx >= len(user.Hand) {
					return fmt.Errorf("无效的弃牌索引: %d", idx)
				}
				discarded = append(discarded, user.Hand[idx])
				user.Hand = append(user.Hand[:idx], user.Hand[idx+1:]...)
			}
			if len(discarded) > 0 {
				e.NotifyCardRevealed(user.ID, discarded, "discard")
				e.State.DiscardPile = append(e.State.DiscardPile, discarded...)
			}
			_ = e.maybeAutoReleaseBloodPriestessByHand(user, "手牌<3强制脱离流血形态")
			e.Log(fmt.Sprintf("%s 的 [血之诅咒] 后续：弃置%d张牌", user.Name, len(discarded)))
			e.PopInterrupt()
			if e.State.PendingInterrupt == nil {
				if len(e.State.PendingDamageQueue) > 0 {
					e.State.Phase = model.PhasePendingDamageResolution
					e.State.ReturnPhase = model.PhaseExtraAction
				} else {
					e.State.Phase = model.PhaseExtraAction
				}
			}
			return nil
		}

		cardIdx := -1
		if selectionIndex >= 1 && selectionIndex <= len(remaining) {
			cardIdx = remaining[selectionIndex-1]
		} else {
			for _, idx := range remaining {
				if idx == selectionIndex {
					cardIdx = idx
					break
				}
			}
		}
		if cardIdx < 0 || cardIdx >= len(user.Hand) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		for _, idx := range selected {
			if idx == cardIdx {
				return fmt.Errorf("不能重复选择同一张牌")
			}
		}
		selected = append(selected, cardIdx)
		nextRemaining := make([]int, 0, len(remaining))
		for _, idx := range remaining {
			if idx != cardIdx {
				nextRemaining = append(nextRemaining, idx)
			}
		}
		ctxData["selected_indices"] = selected
		ctxData["remaining_indices"] = nextRemaining
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return nil
	}
	if choiceType == "mg_medusa_darkmoon_pick" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		var indices []int
		if arr, ok := ctxData["darkmoon_indices"].([]int); ok {
			indices = append(indices, arr...)
		} else if arr, ok := ctxData["darkmoon_indices"].([]interface{}); ok {
			for _, v := range arr {
				if f, ok := v.(float64); ok {
					indices = append(indices, int(f))
				}
			}
		}
		if selectionIndex < 0 || selectionIndex >= len(indices) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		fieldIdx := indices[selectionIndex]
		card, ok := e.removeMoonGoddessDarkMoonByFieldIndex(user, fieldIdx)
		if !ok {
			return fmt.Errorf("请选择可用的暗月")
		}
		e.Heal(user.ID, 1)
		nowPetrify := addMoonGoddessPetrify(user, 1)
		e.Log(fmt.Sprintf("%s 发动 [美杜莎之眼]：移除%s系暗月，治疗+1，石化+1（当前%d）",
			user.Name, card.Element, nowPetrify))

		rawCtx, _ := ctxData["user_ctx"].(*model.Context)
		if card.Type == model.CardTypeMagic && len(user.Hand) > 0 {
			targetIDs := e.moonGoddessEnemyIDs(user)
			if len(targetIDs) > 0 {
				ctxData["choice_type"] = "mg_medusa_magic_discard"
				ctxData["target_ids"] = targetIDs
				ctxData["user_ctx"] = rawCtx
				e.State.PendingInterrupt.Context = ctxData
				e.notifyInterruptPrompt()
				return nil
			}
		}
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			if rawCtx != nil && rawCtx.Trigger == model.TriggerOnAttackStart {
				if len(e.State.ActionQueue) > 0 {
					e.State.Phase = model.PhaseBeforeAction
				} else {
					e.State.Phase = model.PhaseResponse
				}
			} else {
				e.State.Phase = model.PhaseResponse
			}
		}
		return nil
	}
	if choiceType == "mg_medusa_magic_discard" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		cardIdx := selectionIndex
		if cardIdx < 0 || cardIdx >= len(user.Hand) {
			return fmt.Errorf("无效的弃牌索引: %d", selectionIndex)
		}
		card := user.Hand[cardIdx]
		user.Hand = append(user.Hand[:cardIdx], user.Hand[cardIdx+1:]...)
		e.NotifyCardRevealed(user.ID, []model.Card{card}, "discard")
		e.State.DiscardPile = append(e.State.DiscardPile, card)
		e.Log(fmt.Sprintf("%s 的 [美杜莎之眼] 额外效果：弃置1张手牌", user.Name))
		ctxData["choice_type"] = "mg_medusa_magic_target"
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return nil
	}
	if choiceType == "mg_medusa_magic_target" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		var targetIDs []string
		if arr, ok := ctxData["target_ids"].([]string); ok {
			targetIDs = append(targetIDs, arr...)
		} else if arr, ok := ctxData["target_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					targetIDs = append(targetIDs, s)
				}
			}
		}
		if selectionIndex < 0 || selectionIndex >= len(targetIDs) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		target := e.State.Players[targetIDs[selectionIndex]]
		if target == nil {
			return fmt.Errorf("目标角色不存在")
		}
		e.AddPendingDamage(model.PendingDamage{
			SourceID:   user.ID,
			TargetID:   target.ID,
			Damage:     1,
			DamageType: "magic",
			Stage:      0,
		})
		e.Log(fmt.Sprintf("%s 的 [美杜莎之眼] 额外效果：对 %s 造成1点法术伤害", user.Name, target.Name))

		rawCtx, _ := ctxData["user_ctx"].(*model.Context)
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			if rawCtx != nil && rawCtx.Trigger == model.TriggerOnAttackStart {
				e.State.Phase = model.PhasePendingDamageResolution
				e.State.ReturnPhase = model.PhaseBeforeAction
			} else {
				e.State.Phase = model.PhasePendingDamageResolution
			}
		}
		return nil
	}
	if choiceType == "mg_moon_cycle_mode" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		var modes []string
		if arr, ok := ctxData["modes"].([]string); ok {
			modes = append(modes, arr...)
		} else if arr, ok := ctxData["modes"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					modes = append(modes, s)
				}
			}
		}
		if selectionIndex < 0 || selectionIndex >= len(modes) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		mode := modes[selectionIndex]
		if mode == "branch1" {
			if moonGoddessDarkMoonCount(user) <= 0 {
				return fmt.Errorf("暗月不足，无法发动分支①")
			}
			ctxData["choice_type"] = "mg_moon_cycle_heal_target"
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return nil
		}
		if mode == "branch2" {
			if user.Heal <= 0 {
				return fmt.Errorf("治疗不足，无法发动分支②")
			}
			user.Heal--
			now := addMoonGoddessNewMoon(user, 1)
			e.Log(fmt.Sprintf("%s 发动 [月之轮回] 分支②：移除1治疗，+1新月（当前%d）", user.Name, now))
			e.PopInterrupt()
			if e.State.PendingInterrupt == nil {
				e.State.Phase = model.PhaseTurnEnd
			}
			return nil
		}
		return fmt.Errorf("无效分支")
	}
	if choiceType == "mg_moon_cycle_heal_target" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		if moonGoddessDarkMoonCount(user) <= 0 {
			return fmt.Errorf("暗月不足，无法发动分支①")
		}
		var targetIDs []string
		if arr, ok := ctxData["target_ids"].([]string); ok {
			targetIDs = append(targetIDs, arr...)
		} else if arr, ok := ctxData["target_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					targetIDs = append(targetIDs, s)
				}
			}
		}
		if selectionIndex < 0 || selectionIndex >= len(targetIDs) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		target := e.State.Players[targetIDs[selectionIndex]]
		if target == nil {
			return fmt.Errorf("目标角色不存在")
		}
		e.removeMoonGoddessDarkMoonAny(user, 1)
		e.Heal(target.ID, 1)
		e.Log(fmt.Sprintf("%s 发动 [月之轮回] 分支①：移除1暗月并令 %s +1治疗", user.Name, target.Name))
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			e.State.Phase = model.PhaseTurnEnd
		}
		return nil
	}
	if choiceType == "mg_blasphemy_target" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		var targetIDs []string
		if arr, ok := ctxData["target_ids"].([]string); ok {
			targetIDs = append(targetIDs, arr...)
		} else if arr, ok := ctxData["target_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					targetIDs = append(targetIDs, s)
				}
			}
		}
		if selectionIndex == 0 {
			if user.Tokens != nil {
				user.Tokens["mg_blasphemy_pending"] = 0
			}
			e.Log(fmt.Sprintf("%s 选择跳过 [月渎]", user.Name))
			e.PopInterrupt()
			if e.State.PendingInterrupt == nil {
				e.State.Phase = model.PhasePendingDamageResolution
			}
			return nil
		}
		choice := selectionIndex - 1
		if choice < 0 || choice >= len(targetIDs) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		target := e.State.Players[targetIDs[choice]]
		if target == nil {
			return fmt.Errorf("目标角色不存在")
		}
		if user.Heal <= 0 {
			return fmt.Errorf("治疗不足，无法发动月渎")
		}
		user.Heal--
		if user.Tokens == nil {
			user.Tokens = map[string]int{}
		}
		user.Tokens["mg_blasphemy_pending"] = 0
		user.Tokens["mg_blasphemy_used_turn"] = 1
		e.AddPendingDamage(model.PendingDamage{
			SourceID:   user.ID,
			TargetID:   target.ID,
			Damage:     1,
			DamageType: "magic",
			Stage:      0,
		})
		e.Log(fmt.Sprintf("%s 发动 [月渎]：移除1治疗，对 %s 造成1点法术伤害", user.Name, target.Name))
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			e.State.Phase = model.PhasePendingDamageResolution
		}
		return nil
	}
	if choiceType == "mg_darkmoon_slash_x" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		maxX := toIntContextValue(ctxData["max_x"])
		if selectionIndex < 0 || selectionIndex > maxX {
			return fmt.Errorf("无效的X值")
		}
		x := selectionIndex
		if x > moonGoddessDarkMoonCount(user) {
			x = moonGoddessDarkMoonCount(user)
		}
		if x > 0 {
			e.removeMoonGoddessDarkMoonAny(user, x)
			applied := false
			for i := range e.State.PendingDamageQueue {
				pd := &e.State.PendingDamageQueue[i]
				if !strings.EqualFold(pd.DamageType, "Attack") {
					continue
				}
				pd.Damage += x
				applied = true
				break
			}
			if applied {
				e.Log(fmt.Sprintf("%s 的 [暗月斩] 生效：移除%d个暗月，本次攻击伤害额外+%d", user.Name, x, x))
			}
		} else {
			e.Log(fmt.Sprintf("%s 的 [暗月斩]：选择X=0，不增加伤害", user.Name))
		}
		rawCtx, _ := ctxData["user_ctx"].(*model.Context)
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			if rawCtx != nil && rawCtx.Trigger == model.TriggerOnAttackHit {
				e.advancePendingAttackDamageStageAfterHit(rawCtx)
				e.State.Phase = model.PhasePendingDamageResolution
			} else {
				e.State.Phase = model.PhasePendingDamageResolution
			}
		}
		return nil
	}
	if choiceType == "mg_pale_moon_mode" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		var modes []string
		if arr, ok := ctxData["modes"].([]string); ok {
			modes = append(modes, arr...)
		} else if arr, ok := ctxData["modes"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					modes = append(modes, s)
				}
			}
		}
		if selectionIndex < 0 || selectionIndex >= len(modes) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		mode := modes[selectionIndex]
		if mode == "branch1" {
			if moonGoddessPetrify(user) < 3 {
				return fmt.Errorf("石化不足3点，无法发动分支①")
			}
			addMoonGoddessPetrify(user, -3)
			if user.Tokens == nil {
				user.Tokens = map[string]int{}
			}
			user.Tokens["mg_next_attack_no_counter"]++
			user.Tokens["mg_extra_turn_pending"]++
			user.TurnState.PendingActions = append(user.TurnState.PendingActions, model.ActionContext{
				Source:   "苍白之月",
				MustType: "Attack",
			})
			e.Log(fmt.Sprintf("%s 发动 [苍白之月] 分支①：移除3石化，下次主动攻击不可应战，额外+1攻击行动并获得额外回合", user.Name))
			e.PopInterrupt()
			if e.State.PendingInterrupt == nil {
				e.State.Phase = model.PhaseExtraAction
			}
			return nil
		}
		if mode == "branch2" {
			if len(user.Hand) <= 0 {
				return fmt.Errorf("手牌不足，无法发动分支②")
			}
			maxX := moonGoddessNewMoon(user)
			ctxData["choice_type"] = "mg_pale_moon_x"
			ctxData["max_x"] = maxX
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return nil
		}
		return fmt.Errorf("无效分支")
	}
	if choiceType == "mg_pale_moon_x" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		maxX := toIntContextValue(ctxData["max_x"])
		if selectionIndex < 0 || selectionIndex > maxX {
			return fmt.Errorf("无效的X值")
		}
		targetIDs := e.moonGoddessEnemyIDs(user)
		if len(targetIDs) == 0 {
			return fmt.Errorf("没有可选对手")
		}
		ctxData["x"] = selectionIndex
		ctxData["target_ids"] = targetIDs
		ctxData["choice_type"] = "mg_pale_moon_target"
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return nil
	}
	if choiceType == "mg_pale_moon_target" {
		var targetIDs []string
		if arr, ok := ctxData["target_ids"].([]string); ok {
			targetIDs = append(targetIDs, arr...)
		} else if arr, ok := ctxData["target_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					targetIDs = append(targetIDs, s)
				}
			}
		}
		if selectionIndex < 0 || selectionIndex >= len(targetIDs) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		ctxData["target_id"] = targetIDs[selectionIndex]
		ctxData["choice_type"] = "mg_pale_moon_discard"
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return nil
	}
	if choiceType == "mg_pale_moon_discard" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		cardIdx := selectionIndex
		if cardIdx < 0 || cardIdx >= len(user.Hand) {
			return fmt.Errorf("无效的弃牌索引: %d", selectionIndex)
		}
		targetID, _ := ctxData["target_id"].(string)
		target := e.State.Players[targetID]
		if target == nil {
			return fmt.Errorf("目标角色不存在")
		}
		x := toIntContextValue(ctxData["x"])
		card := user.Hand[cardIdx]
		user.Hand = append(user.Hand[:cardIdx], user.Hand[cardIdx+1:]...)
		e.NotifyCardRevealed(user.ID, []model.Card{card}, "discard")
		e.State.DiscardPile = append(e.State.DiscardPile, card)
		if x > moonGoddessNewMoon(user) {
			x = moonGoddessNewMoon(user)
		}
		if x > 0 {
			addMoonGoddessNewMoon(user, -x)
		}
		nowPetrify := addMoonGoddessPetrify(user, 1)
		damage := x + 1
		e.AddPendingDamage(model.PendingDamage{
			SourceID:   user.ID,
			TargetID:   target.ID,
			Damage:     damage,
			DamageType: "magic",
			Stage:      0,
		})
		e.Log(fmt.Sprintf("%s 发动 [苍白之月] 分支②：移除%d新月，石化+1（当前%d），弃1张牌并对 %s 造成%d点法术伤害",
			user.Name, x, nowPetrify, target.Name, damage))
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			e.State.Phase = model.PhasePendingDamageResolution
			e.State.ReturnPhase = model.PhaseExtraAction
		}
		return nil
	}
	if choiceType == "mb_magic_pierce_hit_confirm" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		sourceID, _ := ctxData["source_id"].(string)
		if sourceID != "" && sourceID != user.ID {
			return fmt.Errorf("魔贯冲击上下文玩家不匹配")
		}
		if selectionIndex == 0 {
			if _, ok := removeMagicBowChargeByElement(user, model.ElementFire); ok {
				applied := false
				for i := range e.State.PendingDamageQueue {
					pd := &e.State.PendingDamageQueue[i]
					if !strings.EqualFold(pd.DamageType, "Attack") {
						continue
					}
					pd.Damage++
					applied = true
					break
				}
				e.Log(fmt.Sprintf("%s 的 [魔贯冲击] 命中追加生效：额外移除1个火系充能，本次攻击伤害+1", user.Name))
				if !applied {
					e.Log("[Warn] 魔贯冲击命中追加未找到对应伤害条目，未能叠加伤害")
				}
			} else {
				e.Log(fmt.Sprintf("%s 的 [魔贯冲击] 命中追加失败：火系充能不足", user.Name))
			}
		} else if selectionIndex != 1 {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		if user.Tokens == nil {
			user.Tokens = map[string]int{}
		}
		user.Tokens["mb_magic_pierce_pending"] = 0
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			e.State.Phase = model.PhasePendingDamageResolution
		}
		return nil
	}
	if choiceType == "mb_charge_draw_x" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		maxDraw := toIntContextValue(ctxData["max_draw"])
		if maxDraw <= 0 {
			maxDraw = 4
		}
		x := selectionIndex
		if x < 0 || x > maxDraw {
			return fmt.Errorf("无效的X值")
		}

		if x > 0 {
			cards, newDeck, newDiscard := rules.DrawCards(e.State.Deck, e.State.DiscardPile, x)
			e.State.Deck = newDeck
			e.State.DiscardPile = newDiscard
			user.Hand = append(user.Hand, cards...)
			e.NotifyDrawCards(user.ID, len(cards), "mb_charge")
		}
		room := magicBowChargeCapEngine - magicBowChargeCount(user, "")
		maxPlace := x
		if maxPlace > len(user.Hand) {
			maxPlace = len(user.Hand)
		}
		if maxPlace > room {
			maxPlace = room
		}

		// 【充能】特殊规则：摸牌后若超手牌上限，仅按超出值爆士气，不触发弃牌。
		overflow := len(user.Hand) - e.GetMaxHand(user)
		if overflow > 0 {
			moraleLoss := overflow
			allowedByFloor := e.campMorale(user.Camp) - e.moraleFloorForCamp(user.Camp)
			if allowedByFloor < 0 {
				allowedByFloor = 0
			}
			if moraleLoss > allowedByFloor {
				moraleLoss = allowedByFloor
			}
			if moraleLoss > 0 {
				lossEventCtx := &model.EventContext{
					Type:      model.EventDamage,
					DamageVal: &moraleLoss,
				}
				lossCtx := e.buildContext(user, nil, model.TriggerBeforeMoraleLoss, lossEventCtx)
				lossCtx.Flags["IsMagicDamage"] = false
				if lossCtx.Selections == nil {
					lossCtx.Selections = map[string]any{}
				}
				lossCtx.Selections["discarded_cards"] = []model.Card{}
				lossCtx.Selections["from_damage_draw"] = false
				lossCtx.Selections["victim_id"] = user.ID
				lossCtx.Selections["discard_player_id"] = user.ID
				lossCtx.Selections["morale_loss_stay_in_turn"] = true
				lossCtx.Selections["morale_loss_is_damage_resolution"] = false
				lossCtx.Selections["mb_charge_resume"] = true
				lossCtx.Selections["mb_charge_user_id"] = user.ID
				lossCtx.Selections["mb_charge_max_place"] = maxPlace

				e.dispatcher.OnTrigger(model.TriggerBeforeMoraleLoss, lossCtx)

				pendingResponse := false
				for _, intr := range e.State.InterruptQueue {
					if intr != nil && intr.Type == model.InterruptResponseSkill {
						pendingResponse = true
						break
					}
				}
				if pendingResponse {
					lossCtx.Selections["morale_loss_pending"] = true
					lossCtx.Selections["morale_loss_value"] = moraleLoss
					lossCtx.Selections["is_magic"] = false
					lossCtx.Selections["hero_dead_duel_floor"] = false
					// 当前仍处于“选择X”的中断，弹出后先进入士气响应链，随后在 resumePendingMoraleLoss 中续接充能流程。
					e.PopInterrupt()
					return nil
				}

				finalLoss := e.applyMoraleLossAfterTrigger(user, moraleLoss, false, false, false, []model.Card{}, lossCtx)
				e.Log(fmt.Sprintf("%s 的 [充能] 摸牌后超出手牌上限%d：士气-%d（本次不弃牌）", user.Name, overflow, finalLoss))
				e.checkGameEnd()
			}
		}

		if maxPlace <= 0 {
			e.Log(fmt.Sprintf("%s 的 [充能] 生效：摸%d张，不放置充能", user.Name, x))
			e.PopInterrupt()
			if e.State.PendingInterrupt == nil {
				e.State.Phase = model.PhaseStartup
			}
			return nil
		}

		ctxData["choice_type"] = "mb_charge_place_count"
		ctxData["max_place"] = maxPlace
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		e.Log(fmt.Sprintf("%s 的 [充能] 生效：摸%d张，可放置最多%d张充能", user.Name, x, maxPlace))
		return nil
	}
	if choiceType == "mb_charge_place_count" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		maxPlace := toIntContextValue(ctxData["max_place"])
		if maxPlace < 0 {
			maxPlace = 0
		}
		needCount := selectionIndex
		if needCount < 0 || needCount > maxPlace {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		if needCount == 0 {
			e.Log(fmt.Sprintf("%s 选择不放置充能", user.Name))
			e.PopInterrupt()
			if e.State.PendingInterrupt == nil {
				e.State.Phase = model.PhaseStartup
			}
			return nil
		}
		ctxData["choice_type"] = "mb_charge_place_cards"
		ctxData["need_count"] = needCount
		ctxData["selected_indices"] = []int{}
		ctxData["remaining_indices"] = allHandIndices(user)
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return nil
	}
	if choiceType == "mb_charge_place_cards" || choiceType == "mb_demon_eye_charge_card" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		var remaining []int
		if arr, ok := ctxData["remaining_indices"].([]int); ok {
			remaining = append(remaining, arr...)
		} else if arr, ok := ctxData["remaining_indices"].([]interface{}); ok {
			for _, v := range arr {
				if f, ok := v.(float64); ok {
					remaining = append(remaining, int(f))
				}
			}
		}
		if len(remaining) == 0 {
			remaining = allHandIndices(user)
		}
		var selected []int
		if arr, ok := ctxData["selected_indices"].([]int); ok {
			selected = append(selected, arr...)
		} else if arr, ok := ctxData["selected_indices"].([]interface{}); ok {
			for _, v := range arr {
				if f, ok := v.(float64); ok {
					selected = append(selected, int(f))
				}
			}
		}
		needCount := toIntContextValue(ctxData["need_count"])
		if choiceType == "mb_demon_eye_charge_card" && needCount <= 0 {
			needCount = 1
		}
		if needCount <= 0 {
			needCount = 1
		}
		cardIdx, ok := resolveSelectionToCandidate(selectionIndex, remaining)
		if !ok || cardIdx < 0 || cardIdx >= len(user.Hand) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		selected = append(selected, cardIdx)
		var nextRemaining []int
		for _, idx := range remaining {
			if idx != cardIdx {
				nextRemaining = append(nextRemaining, idx)
			}
		}
		if len(selected) < needCount {
			ctxData["selected_indices"] = selected
			ctxData["remaining_indices"] = nextRemaining
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return nil
		}
		removed, err := removeCardsByIndicesFromHand(user, append([]int{}, selected...))
		if err != nil {
			return err
		}
		added := addMagicBowChargeCards(user, removed)
		if added < len(removed) {
			e.State.DiscardPile = append(e.State.DiscardPile, removed[added:]...)
		}
		if choiceType == "mb_demon_eye_charge_card" {
			maxEnergy := e.getPlayerEnergyCap(user)
			if user.Gem+user.Crystal < maxEnergy {
				user.Crystal++
				if user.Gem+user.Crystal > maxEnergy {
					user.Crystal -= (user.Gem + user.Crystal - maxEnergy)
				}
			}
			e.Log(fmt.Sprintf("%s 的 [魔眼] 生效：放置1张充能并获得1点蓝水晶", user.Name))
		} else {
			e.Log(fmt.Sprintf("%s 的 [充能] 生效：放置%d张充能", user.Name, added))
		}
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			e.State.Phase = model.PhaseStartup
		}
		return nil
	}
	if choiceType == "mb_thunder_scatter_extra" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		var targetIDs []string
		if arr, ok := ctxData["target_ids"].([]string); ok {
			targetIDs = append(targetIDs, arr...)
		} else if arr, ok := ctxData["target_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					targetIDs = append(targetIDs, s)
				}
			}
		}
		if len(targetIDs) == 0 {
			return fmt.Errorf("雷光散射没有可选目标")
		}
		maxExtra := toIntContextValue(ctxData["max_extra"])
		extraX := selectionIndex
		if extraX < 0 || extraX > maxExtra {
			return fmt.Errorf("无效的X值")
		}
		// 基础效果：对所有对手各造成1点法术伤害
		for _, tid := range targetIDs {
			e.AddPendingDamage(model.PendingDamage{
				SourceID:   user.ID,
				TargetID:   tid,
				Damage:     1,
				DamageType: "magic",
				Stage:      0,
			})
		}
		actualExtra := 0
		for i := 0; i < extraX; i++ {
			if _, ok := removeMagicBowChargeByElement(user, model.ElementThunder); !ok {
				break
			}
			actualExtra++
		}
		if actualExtra <= 0 {
			e.Log(fmt.Sprintf("%s 的 [雷光散射] 生效：对所有对手各造成1点法术伤害", user.Name))
			e.PopInterrupt()
			if e.State.PendingInterrupt == nil && len(e.State.PendingDamageQueue) > 0 {
				e.State.Phase = model.PhasePendingDamageResolution
				e.State.ReturnPhase = model.PhaseExtraAction
			}
			return nil
		}
		ctxData["choice_type"] = "mb_thunder_scatter_target"
		ctxData["extra_x"] = actualExtra
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return nil
	}
	if choiceType == "mb_demon_eye_mode" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		if selectionIndex == 0 {
			ctxData["choice_type"] = "mb_demon_eye_target"
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return nil
		}
		if selectionIndex == 1 {
			e.DrawCards(user.ID, 3)
			if len(user.Hand) == 0 {
				maxEnergy := e.getPlayerEnergyCap(user)
				if user.Gem+user.Crystal < maxEnergy {
					user.Crystal++
					if user.Gem+user.Crystal > maxEnergy {
						user.Crystal -= (user.Gem + user.Crystal - maxEnergy)
					}
				}
				e.Log(fmt.Sprintf("%s 的 [魔眼] 生效：摸3后无手牌可充能，改为仅获得1点蓝水晶", user.Name))
				e.PopInterrupt()
				if e.State.PendingInterrupt == nil {
					e.State.Phase = model.PhaseStartup
				}
				return nil
			}
			ctxData["choice_type"] = "mb_demon_eye_charge_card"
			ctxData["need_count"] = 1
			ctxData["selected_indices"] = []int{}
			ctxData["remaining_indices"] = allHandIndices(user)
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return nil
		}
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	if choiceType == "mb_thunder_scatter_target" || choiceType == "mb_multi_shot_target" || choiceType == "mb_demon_eye_target" || choiceType == "ml_stardust_target" || choiceType == "fighter_psi_bullet_target" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		var targetIDs []string
		if arr, ok := ctxData["target_ids"].([]string); ok {
			targetIDs = append(targetIDs, arr...)
		} else if arr, ok := ctxData["target_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					targetIDs = append(targetIDs, s)
				}
			}
		}
		if selectionIndex < 0 || selectionIndex >= len(targetIDs) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		targetID := targetIDs[selectionIndex]
		target := e.State.Players[targetID]
		if target == nil {
			return fmt.Errorf("目标不存在")
		}
		switch choiceType {
		case "mb_thunder_scatter_target":
			extraX := toIntContextValue(ctxData["extra_x"])
			if extraX > 0 {
				e.AddPendingDamage(model.PendingDamage{
					SourceID:   user.ID,
					TargetID:   targetID,
					Damage:     extraX,
					DamageType: "magic",
					Stage:      0,
				})
			}
			e.Log(fmt.Sprintf("%s 的 [雷光散射] 生效：对所有对手各1点，并对 %s 额外造成%d点法术伤害", user.Name, target.Name, extraX))
			e.PopInterrupt()
			if e.State.PendingInterrupt == nil && len(e.State.PendingDamageQueue) > 0 {
				e.State.Phase = model.PhasePendingDamageResolution
				e.State.ReturnPhase = model.PhaseExtraAction
			}
			return nil
		case "mb_multi_shot_target":
			prevOrder := user.TurnState.UsedSkillCounts["mb_last_attack_target_order"]
			if prevOrder > 0 {
				for i, pid := range e.State.PlayerOrder {
					if i+1 == prevOrder && pid == targetID {
						return fmt.Errorf("多重射击不能选择上次攻击目标")
					}
				}
			}
			virtualCard := model.Card{
				ID:          fmt.Sprintf("mb_multi_shot_%s_%d", user.ID, len(e.State.DiscardPile)+len(e.State.ActionQueue)+1),
				Name:        "多重射击",
				Type:        model.CardTypeAttack,
				Element:     model.ElementDark,
				Damage:      1,
				Description: "由多重射击视为的暗系主动攻击（伤害-1）",
			}
			e.State.ActionQueue = append(e.State.ActionQueue, model.QueuedAction{
				SourceID:    user.ID,
				TargetID:    target.ID,
				Type:        model.ActionAttack,
				Element:     model.ElementDark,
				Card:        &virtualCard,
				CardIndex:   -1,
				SourceSkill: "mb_multi_shot",
			})
			e.Log(fmt.Sprintf("%s 的 [多重射击] 生效：对 %s 发起1次暗系追加攻击（伤害-1）", user.Name, target.Name))
			e.PopInterrupt()
			if e.State.PendingInterrupt == nil {
				e.State.Phase = model.PhaseBeforeAction
			}
			return nil
		case "mb_demon_eye_target":
			if len(target.Hand) > 0 {
				discarded := target.Hand[0]
				target.Hand = target.Hand[1:]
				e.NotifyCardRevealed(target.ID, []model.Card{discarded}, "discard")
				e.State.DiscardPile = append(e.State.DiscardPile, discarded)
				e.Log(fmt.Sprintf("%s 的 [魔眼] 生效：%s 弃置了1张手牌", user.Name, target.Name))
			} else {
				e.Log(fmt.Sprintf("%s 的 [魔眼] 生效：%s 无手牌可弃", user.Name, target.Name))
			}
			if len(user.Hand) == 0 {
				maxEnergy := e.getPlayerEnergyCap(user)
				if user.Gem+user.Crystal < maxEnergy {
					user.Crystal++
					if user.Gem+user.Crystal > maxEnergy {
						user.Crystal -= (user.Gem + user.Crystal - maxEnergy)
					}
				}
				e.Log(fmt.Sprintf("%s 的 [魔眼] 生效：无手牌可充能，改为仅获得1点蓝水晶", user.Name))
				e.PopInterrupt()
				if e.State.PendingInterrupt == nil {
					e.State.Phase = model.PhaseStartup
				}
				return nil
			}
			ctxData["choice_type"] = "mb_demon_eye_charge_card"
			ctxData["need_count"] = 1
			ctxData["selected_indices"] = []int{}
			ctxData["remaining_indices"] = allHandIndices(user)
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return nil
		case "ml_stardust_target":
			e.AddPendingDamage(model.PendingDamage{
				SourceID:   user.ID,
				TargetID:   targetID,
				Damage:     2,
				DamageType: "magic",
				Stage:      0,
			})
			e.Log(fmt.Sprintf("%s 的 [幻影星尘] 生效：对 %s 造成2点法术伤害", user.Name, target.Name))
			e.PopInterrupt()
			if e.State.PendingInterrupt == nil {
				if len(e.State.PendingDamageQueue) > 0 {
					e.State.Phase = model.PhasePendingDamageResolution
				} else {
					e.State.Phase = model.PhaseStartup
				}
			}
			return nil
		case "fighter_psi_bullet_target":
			e.AddPendingDamage(model.PendingDamage{
				SourceID:   user.ID,
				TargetID:   targetID,
				Damage:     1,
				DamageType: "magic",
				Stage:      0,
			})
			selfDamage := 0
			if target.Heal <= 0 {
				selfDamage = user.Tokens["fighter_qi"]
				if selfDamage > 0 {
					e.AddPendingDamage(model.PendingDamage{
						SourceID:   user.ID,
						TargetID:   user.ID,
						Damage:     selfDamage,
						DamageType: "magic",
						Stage:      0,
					})
				}
			}
			if selfDamage > 0 {
				e.Log(fmt.Sprintf("%s 的 [念弹] 生效：对 %s 造成1点法术伤害；目标治疗为0，自己额外承受%d点法术伤害", user.Name, target.Name, selfDamage))
			} else {
				e.Log(fmt.Sprintf("%s 的 [念弹] 生效：对 %s 造成1点法术伤害", user.Name, target.Name))
			}
			e.PopInterrupt()
			if e.State.PendingInterrupt == nil {
				if len(e.State.PendingDamageQueue) > 0 {
					e.State.Phase = model.PhasePendingDamageResolution
					e.State.ReturnPhase = model.PhaseExtraAction
				} else {
					e.State.Phase = model.PhaseExtraAction
				}
			}
			return nil
		}
	}
	if choiceType == "elf_elemental_shot_water_target" || choiceType == "elf_elemental_shot_earth_target" || choiceType == "elf_ritual_release_target" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		var targetIDs []string
		if arr, ok := ctxData["target_ids"].([]string); ok {
			targetIDs = arr
		} else if arr, ok := ctxData["target_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					targetIDs = append(targetIDs, s)
				}
			}
		}
		if selectionIndex < 0 || selectionIndex >= len(targetIDs) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		targetID := targetIDs[selectionIndex]
		switch choiceType {
		case "elf_elemental_shot_water_target":
			e.Heal(targetID, 1)
		case "elf_elemental_shot_earth_target":
			e.AddPendingDamage(model.PendingDamage{
				SourceID:   user.ID,
				TargetID:   targetID,
				Damage:     1,
				DamageType: "magic",
				Stage:      0,
			})
		case "elf_ritual_release_target":
			user.Tokens["elf_ritual_form"] = 0
			user.Tokens["elf_ritual_release_waiting"] = 0
			e.AddPendingDamage(model.PendingDamage{
				SourceID:   user.ID,
				TargetID:   targetID,
				Damage:     2,
				DamageType: "magic",
				Stage:      0,
			})
		}
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil && len(e.State.PendingDamageQueue) > 0 {
			e.State.Phase = model.PhasePendingDamageResolution
		}
		return nil
	}
	if choiceType == "plague_death_touch_element" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		var elements []string
		if arr, ok := ctxData["elements"].([]string); ok {
			elements = arr
		} else if arr, ok := ctxData["elements"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					elements = append(elements, s)
				}
			}
		}
		if selectionIndex < 0 || selectionIndex >= len(elements) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		ele := elements[selectionIndex]
		maxCards := len(getCardIndicesByElement(user, model.Element(ele)))
		e.State.PendingInterrupt.Context = map[string]interface{}{
			"choice_type":      "plague_death_touch_x",
			"user_id":          userID,
			"chosen_element":   ele,
			"max_heal":         user.Heal,
			"max_cards":        maxCards,
			"selected_indices": []int{},
		}
		e.notifyInterruptPrompt()
		return nil
	}
	if choiceType == "plague_death_touch_x" {
		maxHeal := 0
		if v, ok := ctxData["max_heal"].(int); ok {
			maxHeal = v
		} else if f, ok := ctxData["max_heal"].(float64); ok {
			maxHeal = int(f)
		}
		x := selectionIndex + 2
		if x < 2 || x > maxHeal {
			return fmt.Errorf("无效的X值")
		}
		ctxData["choice_type"] = "plague_death_touch_y"
		ctxData["x_value"] = x
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return nil
	}
	if choiceType == "plague_death_touch_y" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		maxCards := 0
		if v, ok := ctxData["max_cards"].(int); ok {
			maxCards = v
		} else if f, ok := ctxData["max_cards"].(float64); ok {
			maxCards = int(f)
		}
		y := selectionIndex + 2
		if y < 2 || y > maxCards {
			return fmt.Errorf("无效的Y值")
		}
		ele, _ := ctxData["chosen_element"].(string)
		indices := getCardIndicesByElement(user, model.Element(ele))
		ctxData["choice_type"] = "plague_death_touch_cards"
		ctxData["y_value"] = y
		ctxData["selected_indices"] = []int{}
		ctxData["remaining_indices"] = indices
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return nil
	}
	if choiceType == "plague_death_touch_cards" {
		var remaining []int
		if arr, ok := ctxData["remaining_indices"].([]int); ok {
			remaining = arr
		} else if arr, ok := ctxData["remaining_indices"].([]interface{}); ok {
			for _, v := range arr {
				if f, ok := v.(float64); ok {
					remaining = append(remaining, int(f))
				}
			}
		}
		var selected []int
		if arr, ok := ctxData["selected_indices"].([]int); ok {
			selected = arr
		} else if arr, ok := ctxData["selected_indices"].([]interface{}); ok {
			for _, v := range arr {
				if f, ok := v.(float64); ok {
					selected = append(selected, int(f))
				}
			}
		}
		yValue := 0
		if v, ok := ctxData["y_value"].(int); ok {
			yValue = v
		} else if f, ok := ctxData["y_value"].(float64); ok {
			yValue = int(f)
		}
		cardIdx, ok := resolveSelectionToCandidate(selectionIndex, remaining)
		if !ok {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		selected = append(selected, cardIdx)
		var nextRemaining []int
		for _, v := range remaining {
			if v != cardIdx {
				nextRemaining = append(nextRemaining, v)
			}
		}
		if len(selected) < yValue {
			ctxData["selected_indices"] = selected
			ctxData["remaining_indices"] = nextRemaining
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return nil
		}
		ctxData["selected_indices"] = selected
		ctxData["choice_type"] = "plague_death_touch_target"
		ctxData["target_ids"] = append([]string{}, e.State.PlayerOrder...)
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return nil
	}
	if choiceType == "plague_death_touch_target" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		var targetIDs []string
		if arr, ok := ctxData["target_ids"].([]string); ok {
			targetIDs = arr
		} else if arr, ok := ctxData["target_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					targetIDs = append(targetIDs, s)
				}
			}
		}
		if selectionIndex < 0 || selectionIndex >= len(targetIDs) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		var selected []int
		if arr, ok := ctxData["selected_indices"].([]int); ok {
			selected = arr
		} else if arr, ok := ctxData["selected_indices"].([]interface{}); ok {
			for _, v := range arr {
				if f, ok := v.(float64); ok {
					selected = append(selected, int(f))
				}
			}
		}
		xVal := 0
		if v, ok := ctxData["x_value"].(int); ok {
			xVal = v
		} else if f, ok := ctxData["x_value"].(float64); ok {
			xVal = int(f)
		}
		yVal := 0
		if v, ok := ctxData["y_value"].(int); ok {
			yVal = v
		} else if f, ok := ctxData["y_value"].(float64); ok {
			yVal = int(f)
		}
		removed, err := removeCardsByIndicesFromHand(user, append([]int{}, selected...))
		if err != nil {
			return err
		}
		e.NotifyCardRevealed(user.ID, removed, "discard")
		e.State.DiscardPile = append(e.State.DiscardPile, removed...)
		if user.Heal < xVal {
			return fmt.Errorf("治疗不足，无法移除X=%d", xVal)
		}
		user.Heal -= xVal
		damage := xVal + yVal - 3
		if damage < 0 {
			damage = 0
		}
		if user.Tokens == nil {
			user.Tokens = map[string]int{}
		}
		user.Tokens["plague_block_immortal"] = 1
		e.AddPendingDamage(model.PendingDamage{
			SourceID:           user.ID,
			TargetID:           targetIDs[selectionIndex],
			Damage:             damage,
			DamageType:         "magic",
			CapDrawToHandLimit: true,
			Stage:              0,
		})
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			if len(e.State.PendingDamageQueue) > 0 {
				e.State.Phase = model.PhasePendingDamageResolution
				e.State.ReturnPhase = model.PhaseExtraAction
			} else {
				e.State.Phase = model.PhaseExtraAction
			}
		}
		return nil
	}
	if choiceType == "ms_shadow_meteor_discard" {
		var remaining []int
		if arr, ok := ctxData["remaining_indices"].([]int); ok {
			remaining = arr
		} else if arr, ok := ctxData["remaining_indices"].([]interface{}); ok {
			for _, v := range arr {
				if f, ok := v.(float64); ok {
					remaining = append(remaining, int(f))
				}
			}
		}
		if len(remaining) == 0 {
			if arr, ok := ctxData["magic_indices"].([]int); ok {
				remaining = arr
			} else if arr, ok := ctxData["magic_indices"].([]interface{}); ok {
				for _, v := range arr {
					if f, ok := v.(float64); ok {
						remaining = append(remaining, int(f))
					}
				}
			}
		}
		var selected []int
		if arr, ok := ctxData["selected_indices"].([]int); ok {
			selected = arr
		} else if arr, ok := ctxData["selected_indices"].([]interface{}); ok {
			for _, v := range arr {
				if f, ok := v.(float64); ok {
					selected = append(selected, int(f))
				}
			}
		}
		cardIdx, ok := resolveSelectionToCandidate(selectionIndex, remaining)
		if !ok {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		selected = append(selected, cardIdx)
		var nextRemaining []int
		for _, v := range remaining {
			if v != cardIdx {
				nextRemaining = append(nextRemaining, v)
			}
		}
		if len(selected) < 2 {
			ctxData["selected_indices"] = selected
			ctxData["remaining_indices"] = nextRemaining
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return nil
		}
		ctxData["selected_indices"] = selected
		ctxData["choice_type"] = "ms_shadow_meteor_target"
		ctxData["target_ids"] = append([]string{}, e.State.PlayerOrder...)
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return nil
	}
	if choiceType == "ms_shadow_meteor_target" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		var targetIDs []string
		if arr, ok := ctxData["target_ids"].([]string); ok {
			targetIDs = arr
		} else if arr, ok := ctxData["target_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					targetIDs = append(targetIDs, s)
				}
			}
		}
		if selectionIndex < 0 || selectionIndex >= len(targetIDs) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		var selected []int
		if arr, ok := ctxData["selected_indices"].([]int); ok {
			selected = arr
		} else if arr, ok := ctxData["selected_indices"].([]interface{}); ok {
			for _, v := range arr {
				if f, ok := v.(float64); ok {
					selected = append(selected, int(f))
				}
			}
		}
		removed, err := removeCardsByIndicesFromHand(user, append([]int{}, selected...))
		if err != nil {
			return err
		}
		e.NotifyCardRevealed(user.ID, removed, "discard")
		e.State.DiscardPile = append(e.State.DiscardPile, removed...)
		e.AddPendingDamage(model.PendingDamage{
			SourceID:   user.ID,
			TargetID:   targetIDs[selectionIndex],
			Damage:     2,
			DamageType: "magic",
			Stage:      0,
		})
		camp := string(user.Camp)
		total := e.GetCampGems(camp) + e.GetCampCrystals(camp)
		if total >= 2 {
			e.State.PendingInterrupt.Context = map[string]interface{}{
				"choice_type": "ms_shadow_meteor_release_confirm",
				"user_id":     user.ID,
				"camp":        camp,
			}
			e.notifyInterruptPrompt()
			return nil
		}
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			if len(e.State.PendingDamageQueue) > 0 {
				e.State.Phase = model.PhasePendingDamageResolution
				e.State.ReturnPhase = model.PhaseExtraAction
			} else {
				e.State.Phase = model.PhaseExtraAction
			}
		}
		return nil
	}
	if choiceType == "ms_shadow_meteor_release_confirm" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		if selectionIndex == 0 {
			camp, _ := ctxData["camp"].(string)
			need := 2
			useCrystal := need
			if useCrystal > e.GetCampCrystals(camp) {
				useCrystal = e.GetCampCrystals(camp)
			}
			if useCrystal > 0 {
				e.ModifyCrystal(camp, -useCrystal)
			}
			remain := need - useCrystal
			if remain > 0 {
				e.ModifyGem(camp, -remain)
			}
			user.Tokens["ms_shadow_form"] = 0
			user.Tokens["ms_shadow_release_pending"] = 0
			user.Gem++
			e.Log(fmt.Sprintf("%s 通过[暗影流星]额外效果转正并获得1红宝石", user.Name))
		}
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			if len(e.State.PendingDamageQueue) > 0 {
				e.State.Phase = model.PhasePendingDamageResolution
				e.State.ReturnPhase = model.PhaseExtraAction
			} else {
				e.State.Phase = model.PhaseExtraAction
			}
		}
		return nil
	}
	if choiceType == "css_blood_barrier_counter_confirm" {
		if selectionIndex == 1 {
			e.PopInterrupt()
			if e.State.PendingInterrupt == nil && len(e.State.PendingDamageQueue) > 0 {
				e.State.Phase = model.PhasePendingDamageResolution
			}
			return nil
		}
		var enemyIDs []string
		if arr, ok := ctxData["enemy_ids"].([]string); ok {
			enemyIDs = arr
		} else if arr, ok := ctxData["enemy_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					enemyIDs = append(enemyIDs, s)
				}
			}
		}
		ctxData["choice_type"] = "css_blood_barrier_target"
		ctxData["target_ids"] = enemyIDs
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return nil
	}
	if choiceType == "css_blood_barrier_target" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		var targetIDs []string
		if arr, ok := ctxData["target_ids"].([]string); ok {
			targetIDs = arr
		} else if arr, ok := ctxData["target_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					targetIDs = append(targetIDs, s)
				}
			}
		}
		if selectionIndex < 0 || selectionIndex >= len(targetIDs) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		e.AddPendingDamage(model.PendingDamage{
			SourceID:   user.ID,
			TargetID:   targetIDs[selectionIndex],
			Damage:     1,
			DamageType: "magic",
			Stage:      0,
		})
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil && len(e.State.PendingDamageQueue) > 0 {
			e.State.Phase = model.PhasePendingDamageResolution
		}
		return nil
	}
	if choiceType == "bt_dance_mode" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		canDiscard := toBoolContextValue(ctxData["can_discard"])
		var modes []string
		modes = append(modes, "draw")
		if canDiscard {
			modes = append(modes, "discard")
		}
		if selectionIndex < 0 || selectionIndex >= len(modes) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		mode := modes[selectionIndex]
		if mode == "discard" {
			if len(user.Hand) <= 0 {
				return fmt.Errorf("手牌不足，无法弃牌")
			}
			ctxData["choice_type"] = "bt_dance_discard"
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return nil
		}
		cards, newDeck, newDiscard := rules.DrawCards(e.State.Deck, e.State.DiscardPile, 1)
		e.State.Deck = newDeck
		e.State.DiscardPile = newDiscard
		user.Hand = append(user.Hand, cards...)
		e.NotifyDrawCards(user.ID, len(cards), "bt_dance_draw")

		cocoons, deckAfter, discardAfter := rules.DrawCards(e.State.Deck, e.State.DiscardPile, 1)
		e.State.Deck = deckAfter
		e.State.DiscardPile = discardAfter
		added := addButterflyCocoonCards(user, cocoons)
		e.Log(fmt.Sprintf("%s 发动 [舞动]：摸1张牌，并将牌库顶%d张牌放置为茧", user.Name, added))

		e.checkHandLimit(user, nil)
		overflow := butterflyCocoonCount(user) - butterflyCocoonCapEngine
		if overflow > 0 {
			e.PushInterrupt(&model.Interrupt{
				Type:     model.InterruptChoice,
				PlayerID: user.ID,
				Context: map[string]interface{}{
					"choice_type":       "bt_cocoon_overflow_discard",
					"user_id":           user.ID,
					"discard_count":     overflow,
					"remaining_indices": butterflyCocoonFieldIndices(user),
					"selected_indices":  []int{},
				},
			})
		}
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			if len(e.State.PendingDamageQueue) > 0 {
				e.State.Phase = model.PhasePendingDamageResolution
				if e.State.ReturnPhase == "" {
					e.State.ReturnPhase = model.PhaseExtraAction
				}
			} else {
				e.State.Phase = model.PhaseExtraAction
			}
		}
		return nil
	}
	if choiceType == "bt_dance_discard" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		if selectionIndex < 0 || selectionIndex >= len(user.Hand) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		card := user.Hand[selectionIndex]
		user.Hand = append(user.Hand[:selectionIndex], user.Hand[selectionIndex+1:]...)
		e.NotifyCardRevealed(user.ID, []model.Card{card}, "discard")
		e.State.DiscardPile = append(e.State.DiscardPile, card)

		cocoons, deckAfter, discardAfter := rules.DrawCards(e.State.Deck, e.State.DiscardPile, 1)
		e.State.Deck = deckAfter
		e.State.DiscardPile = discardAfter
		added := addButterflyCocoonCards(user, cocoons)
		e.Log(fmt.Sprintf("%s 发动 [舞动]：弃1张牌，并将牌库顶%d张牌放置为茧", user.Name, added))

		overflow := butterflyCocoonCount(user) - butterflyCocoonCapEngine
		if overflow > 0 {
			e.PushInterrupt(&model.Interrupt{
				Type:     model.InterruptChoice,
				PlayerID: user.ID,
				Context: map[string]interface{}{
					"choice_type":       "bt_cocoon_overflow_discard",
					"user_id":           user.ID,
					"discard_count":     overflow,
					"remaining_indices": butterflyCocoonFieldIndices(user),
					"selected_indices":  []int{},
				},
			})
		}
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			e.State.Phase = model.PhaseExtraAction
		}
		return nil
	}
	if choiceType == "bt_chrysalis_resolve" {
		if selectionIndex != 0 {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		now := addButterflyPupa(user, 1)
		cards, newDeck, newDiscard := rules.DrawCards(e.State.Deck, e.State.DiscardPile, 4)
		e.State.Deck = newDeck
		e.State.DiscardPile = newDiscard
		added := addButterflyCocoonCards(user, cards)
		e.Log(fmt.Sprintf("%s 发动 [蛹化]：蛹+1（当前%d），获得%d个茧", user.Name, now, added))
		overflow := butterflyCocoonCount(user) - butterflyCocoonCapEngine
		if overflow > 0 {
			e.PushInterrupt(&model.Interrupt{
				Type:     model.InterruptChoice,
				PlayerID: user.ID,
				Context: map[string]interface{}{
					"choice_type":       "bt_cocoon_overflow_discard",
					"user_id":           user.ID,
					"discard_count":     overflow,
					"remaining_indices": butterflyCocoonFieldIndices(user),
					"selected_indices":  []int{},
				},
			})
		}
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			e.State.Phase = model.PhaseExtraAction
		}
		return nil
	}
	if choiceType == "bt_cocoon_overflow_discard" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		discardNeed := toIntContextValue(ctxData["discard_count"])
		var remaining []int
		if arr, ok := ctxData["remaining_indices"].([]int); ok {
			remaining = append(remaining, arr...)
		} else if arr, ok := ctxData["remaining_indices"].([]interface{}); ok {
			for _, v := range arr {
				if f, ok := v.(float64); ok {
					remaining = append(remaining, int(f))
				}
			}
		}
		var selected []int
		if arr, ok := ctxData["selected_indices"].([]int); ok {
			selected = append(selected, arr...)
		} else if arr, ok := ctxData["selected_indices"].([]interface{}); ok {
			for _, v := range arr {
				if f, ok := v.(float64); ok {
					selected = append(selected, int(f))
				}
			}
		}
		if selectionIndex == -1 {
			if len(selected) < discardNeed {
				return fmt.Errorf("还需选择 %d 个茧", discardNeed-len(selected))
			}
			removed, err := removeButterflyCocoonByFieldIndices(user, append([]int{}, selected...))
			if err != nil {
				return err
			}
			if len(removed) > 0 {
				e.NotifyCardHidden(user.ID, removed, "discard")
				e.State.DiscardPile = append(e.State.DiscardPile, removed...)
			}
			e.Log(fmt.Sprintf("%s 的 [茧上限] 结算：舍弃%d个茧", user.Name, len(removed)))
			e.PopInterrupt()
			if e.State.PendingInterrupt == nil {
				e.State.Phase = model.PhaseExtraAction
			}
			return nil
		}
		candidate, ok := resolveSelectionToCandidate(selectionIndex, remaining)
		if !ok {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		for _, v := range selected {
			if v == candidate {
				return fmt.Errorf("不能重复选择同一个茧")
			}
		}
		selected = append(selected, candidate)
		var nextRemaining []int
		for _, v := range remaining {
			if v != candidate {
				nextRemaining = append(nextRemaining, v)
			}
		}
		ctxData["selected_indices"] = selected
		ctxData["remaining_indices"] = nextRemaining
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return nil
	}
	if choiceType == "bt_reverse_discard" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		discardNeed := toIntContextValue(ctxData["discard_count"])
		var remaining []int
		if arr, ok := ctxData["remaining_indices"].([]int); ok {
			remaining = append(remaining, arr...)
		} else if arr, ok := ctxData["remaining_indices"].([]interface{}); ok {
			for _, v := range arr {
				if f, ok := v.(float64); ok {
					remaining = append(remaining, int(f))
				}
			}
		}
		var selected []int
		if arr, ok := ctxData["selected_indices"].([]int); ok {
			selected = append(selected, arr...)
		} else if arr, ok := ctxData["selected_indices"].([]interface{}); ok {
			for _, v := range arr {
				if f, ok := v.(float64); ok {
					selected = append(selected, int(f))
				}
			}
		}
		if selectionIndex == -1 {
			if len(selected) < discardNeed {
				return fmt.Errorf("还需选择 %d 张弃牌", discardNeed-len(selected))
			}
			removed, err := removeCardsByIndicesFromHand(user, append([]int{}, selected...))
			if err != nil {
				return err
			}
			if len(removed) > 0 {
				e.NotifyCardRevealed(user.ID, removed, "discard")
				e.State.DiscardPile = append(e.State.DiscardPile, removed...)
			}
			ctxData["choice_type"] = "bt_reverse_mode"
			ctxData["can_branch2"] = butterflyPupa(user) > 0
			ctxData["can_remove_cocoon"] = butterflyCocoonCount(user) >= 2
			ctxData["target_ids"] = append([]string{}, e.State.PlayerOrder...)
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return nil
		}
		cardIdx, ok := resolveSelectionToCandidate(selectionIndex, remaining)
		if !ok || cardIdx < 0 || cardIdx >= len(user.Hand) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		for _, v := range selected {
			if v == cardIdx {
				return fmt.Errorf("不能重复选择同一张牌")
			}
		}
		selected = append(selected, cardIdx)
		var nextRemaining []int
		for _, v := range remaining {
			if v != cardIdx {
				nextRemaining = append(nextRemaining, v)
			}
		}
		ctxData["selected_indices"] = selected
		ctxData["remaining_indices"] = nextRemaining
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return nil
	}
	if choiceType == "bt_reverse_mode" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		canBranch2 := toBoolContextValue(ctxData["can_branch2"])
		var modes []string
		modes = append(modes, "branch1")
		if canBranch2 {
			modes = append(modes, "branch2")
		}
		if selectionIndex < 0 || selectionIndex >= len(modes) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		mode := modes[selectionIndex]
		if mode == "branch1" {
			ctxData["choice_type"] = "bt_reverse_target"
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return nil
		}
		if butterflyPupa(user) <= 0 {
			return fmt.Errorf("蛹不足，无法发动分支②")
		}
		ctxData["choice_type"] = "bt_reverse_branch2_cost"
		ctxData["can_remove_cocoon"] = butterflyCocoonCount(user) >= 2
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return nil
	}
	if choiceType == "bt_reverse_target" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		var targetIDs []string
		if arr, ok := ctxData["target_ids"].([]string); ok {
			targetIDs = append(targetIDs, arr...)
		} else if arr, ok := ctxData["target_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					targetIDs = append(targetIDs, s)
				}
			}
		}
		if selectionIndex < 0 || selectionIndex >= len(targetIDs) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		targetID := targetIDs[selectionIndex]
		e.AddPendingDamage(model.PendingDamage{
			SourceID:   user.ID,
			TargetID:   targetID,
			Damage:     1,
			DamageType: "magic",
			IgnoreHeal: true,
			Stage:      0,
		})
		if target := e.State.Players[targetID]; target != nil {
			e.Log(fmt.Sprintf("%s 的 [倒逆之蝶] 分支①：对 %s 造成1点不可治疗抵御的法术伤害", user.Name, target.Name))
		}
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			if len(e.State.PendingDamageQueue) > 0 {
				e.State.Phase = model.PhasePendingDamageResolution
				e.State.ReturnPhase = model.PhaseExtraAction
			} else {
				e.State.Phase = model.PhaseExtraAction
			}
		}
		return nil
	}
	if choiceType == "bt_reverse_branch2_cost" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		canRemove := toBoolContextValue(ctxData["can_remove_cocoon"])
		var modes []string
		if canRemove {
			modes = append(modes, "remove_cocoon")
		}
		modes = append(modes, "self_damage")
		if selectionIndex < 0 || selectionIndex >= len(modes) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		mode := modes[selectionIndex]
		if mode == "remove_cocoon" {
			ctxData["choice_type"] = "bt_reverse_branch2_pick"
			ctxData["remaining_indices"] = butterflyCocoonFieldIndices(user)
			ctxData["selected_indices"] = []int{}
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return nil
		}
		e.AddPendingDamage(model.PendingDamage{
			SourceID:   user.ID,
			TargetID:   user.ID,
			Damage:     4,
			DamageType: "magic",
			Stage:      0,
		})
		now := addButterflyPupa(user, -1)
		e.Log(fmt.Sprintf("%s 的 [倒逆之蝶] 分支②：对自己造成4点法术伤害并移除1个蛹（当前蛹=%d）", user.Name, now))
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			if len(e.State.PendingDamageQueue) > 0 {
				e.State.Phase = model.PhasePendingDamageResolution
				e.State.ReturnPhase = model.PhaseExtraAction
			} else {
				e.State.Phase = model.PhaseExtraAction
			}
		}
		return nil
	}
	if choiceType == "bt_reverse_branch2_pick" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		var remaining []int
		if arr, ok := ctxData["remaining_indices"].([]int); ok {
			remaining = append(remaining, arr...)
		} else if arr, ok := ctxData["remaining_indices"].([]interface{}); ok {
			for _, v := range arr {
				if f, ok := v.(float64); ok {
					remaining = append(remaining, int(f))
				}
			}
		}
		var selected []int
		if arr, ok := ctxData["selected_indices"].([]int); ok {
			selected = append(selected, arr...)
		} else if arr, ok := ctxData["selected_indices"].([]interface{}); ok {
			for _, v := range arr {
				if f, ok := v.(float64); ok {
					selected = append(selected, int(f))
				}
			}
		}
		if selectionIndex == -1 {
			if len(selected) < 2 {
				return fmt.Errorf("还需选择 %d 个茧", 2-len(selected))
			}
			removed, err := removeButterflyCocoonByFieldIndices(user, append([]int{}, selected...))
			if err != nil {
				return err
			}
			if len(removed) > 0 {
				e.NotifyCardRevealed(user.ID, removed, "discard")
				e.State.DiscardPile = append(e.State.DiscardPile, removed...)
			}
			for _, c := range removed {
				if c.Type == model.CardTypeMagic {
					e.queueButterflyWitherTrigger(user)
				}
			}
			now := addButterflyPupa(user, -1)
			e.Log(fmt.Sprintf("%s 的 [倒逆之蝶] 分支②：移除2个茧并移除1个蛹（当前蛹=%d）", user.Name, now))
			e.PopInterrupt()
			if e.State.PendingInterrupt == nil {
				if len(e.State.PendingDamageQueue) > 0 {
					e.State.Phase = model.PhasePendingDamageResolution
					e.State.ReturnPhase = model.PhaseExtraAction
				} else {
					e.State.Phase = model.PhaseExtraAction
				}
			}
			return nil
		}
		candidate, ok := resolveSelectionToCandidate(selectionIndex, remaining)
		if !ok {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		for _, v := range selected {
			if v == candidate {
				return fmt.Errorf("不能重复选择同一个茧")
			}
		}
		selected = append(selected, candidate)
		var nextRemaining []int
		for _, v := range remaining {
			if v != candidate {
				nextRemaining = append(nextRemaining, v)
			}
		}
		ctxData["selected_indices"] = selected
		ctxData["remaining_indices"] = nextRemaining
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return nil
	}
	if choiceType == "bt_pilgrimage_pick" || choiceType == "bt_poison_pick" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		var cocoonIndices []int
		if arr, ok := ctxData["cocoon_indices"].([]int); ok {
			cocoonIndices = append(cocoonIndices, arr...)
		} else if arr, ok := ctxData["cocoon_indices"].([]interface{}); ok {
			for _, v := range arr {
				if f, ok := v.(float64); ok {
					cocoonIndices = append(cocoonIndices, int(f))
				}
			}
		}
		if selectionIndex == -1 || selectionIndex == 0 {
			e.PopInterrupt()
			if e.State.PendingInterrupt == nil {
				if len(e.State.PendingDamageQueue) > 0 {
					e.State.Phase = model.PhasePendingDamageResolution
				} else {
					e.State.Phase = model.PhaseExtraAction
				}
			}
			return nil
		}
		pickIdx := -1
		if selectionIndex >= 1 && selectionIndex <= len(cocoonIndices) {
			pickIdx = cocoonIndices[selectionIndex-1]
		} else if idx, ok := resolveSelectionToCandidate(selectionIndex, cocoonIndices); ok {
			pickIdx = idx
		}
		if pickIdx < 0 {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		removed, ok := removeButterflyCocoonByFieldIndex(user, pickIdx)
		if !ok {
			return fmt.Errorf("选择的茧无效")
		}
		e.NotifyCardRevealed(user.ID, []model.Card{removed}, "discard")
		e.State.DiscardPile = append(e.State.DiscardPile, removed)
		damageIdx := toIntContextValue(ctxData["damage_index"])
		if damageIdx < 0 || damageIdx >= len(e.State.PendingDamageQueue) {
			return fmt.Errorf("伤害上下文不存在")
		}
		pd := &e.State.PendingDamageQueue[damageIdx]
		if choiceType == "bt_pilgrimage_pick" {
			if pd.Damage > 0 {
				pd.Damage--
			}
			e.Log(fmt.Sprintf("%s 发动 [朝圣]：移除1个茧，抵御1点伤害（剩余伤害=%d）", user.Name, pd.Damage))
		} else {
			pd.Damage++
			e.Log(fmt.Sprintf("%s 发动 [毒粉]：移除1个茧，本次法术伤害+1（当前伤害=%d）", user.Name, pd.Damage))
		}
		if removed.Type == model.CardTypeMagic {
			e.queueButterflyWitherTrigger(user)
		}
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			e.State.Phase = model.PhasePendingDamageResolution
		}
		return nil
	}
	if choiceType == "bt_mirror_pair" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		if selectionIndex == -1 || selectionIndex == 0 {
			e.PopInterrupt()
			if e.State.PendingInterrupt == nil {
				if len(e.State.PendingDamageQueue) > 0 {
					e.State.Phase = model.PhasePendingDamageResolution
				} else {
					e.State.Phase = model.PhaseExtraAction
				}
			}
			return nil
		}
		var pairDefs []string
		if arr, ok := ctxData["pair_defs"].([]string); ok {
			pairDefs = append(pairDefs, arr...)
		} else if arr, ok := ctxData["pair_defs"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					pairDefs = append(pairDefs, s)
				}
			}
		}
		pairChoice := -1
		if selectionIndex >= 1 && selectionIndex <= len(pairDefs) {
			pairChoice = selectionIndex - 1
		} else if selectionIndex >= 0 && selectionIndex < len(pairDefs) {
			pairChoice = selectionIndex
		}
		if pairChoice < 0 || pairChoice >= len(pairDefs) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		parts := strings.Split(pairDefs[pairChoice], ",")
		if len(parts) != 2 {
			return fmt.Errorf("镜花水月配对参数无效")
		}
		left, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
		right, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
		if err1 != nil || err2 != nil {
			return fmt.Errorf("镜花水月配对索引无效")
		}
		removed, err := removeButterflyCocoonByFieldIndices(user, []int{left, right})
		if err != nil {
			return err
		}
		if len(removed) != 2 || removed[0].Element != removed[1].Element {
			return fmt.Errorf("镜花水月需要移除2张同系茧")
		}
		e.NotifyCardRevealed(user.ID, removed, "discard")
		e.State.DiscardPile = append(e.State.DiscardPile, removed...)
		damageIdx := toIntContextValue(ctxData["damage_index"])
		if damageIdx < 0 || damageIdx >= len(e.State.PendingDamageQueue) {
			return fmt.Errorf("伤害上下文不存在")
		}
		pd := &e.State.PendingDamageQueue[damageIdx]
		originTargetID := pd.TargetID
		pd.Damage = 0
		e.AddPendingDamage(model.PendingDamage{
			SourceID:   user.ID,
			TargetID:   originTargetID,
			Damage:     1,
			DamageType: "magic",
			Stage:      0,
		})
		e.AddPendingDamage(model.PendingDamage{
			SourceID:   user.ID,
			TargetID:   originTargetID,
			Damage:     1,
			DamageType: "magic",
			Stage:      0,
		})
		for _, c := range removed {
			if c.Type == model.CardTypeMagic {
				e.queueButterflyWitherTrigger(user)
			}
		}
		if target := e.State.Players[originTargetID]; target != nil {
			e.Log(fmt.Sprintf("%s 发动 [镜花水月]：抵御原伤害，并改为对 %s 造成2次1点法术伤害", user.Name, target.Name))
		}
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			e.State.Phase = model.PhasePendingDamageResolution
		}
		return nil
	}
	if choiceType == "bt_wither_confirm" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		if user.Tokens == nil {
			user.Tokens = map[string]int{}
		}
		if selectionIndex == 0 {
			ctxData["choice_type"] = "bt_wither_target"
			ctxData["target_ids"] = e.butterflyEnemyIDs(user)
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return nil
		}
		if selectionIndex != 1 {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		if user.Tokens["bt_wither_pending"] > 0 {
			user.Tokens["bt_wither_pending"]--
		}
		if user.Tokens["bt_wither_pending"] > 0 {
			ctxData["choice_type"] = "bt_wither_confirm"
			ctxData["target_ids"] = e.butterflyEnemyIDs(user)
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return nil
		}
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			if len(e.State.PendingDamageQueue) > 0 {
				e.State.Phase = model.PhasePendingDamageResolution
			} else {
				e.State.Phase = model.PhaseExtraAction
			}
		}
		return nil
	}
	if choiceType == "bt_wither_target" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		var targetIDs []string
		if arr, ok := ctxData["target_ids"].([]string); ok {
			targetIDs = append(targetIDs, arr...)
		} else if arr, ok := ctxData["target_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					targetIDs = append(targetIDs, s)
				}
			}
		}
		if selectionIndex < 0 || selectionIndex >= len(targetIDs) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		targetID := targetIDs[selectionIndex]
		e.AddPendingDamage(model.PendingDamage{
			SourceID:   user.ID,
			TargetID:   targetID,
			Damage:     1,
			DamageType: "magic",
			Stage:      0,
		})
		e.AddPendingDamage(model.PendingDamage{
			SourceID:   user.ID,
			TargetID:   user.ID,
			Damage:     2,
			DamageType: "magic",
			Stage:      0,
		})
		if user.Tokens == nil {
			user.Tokens = map[string]int{}
		}
		user.Tokens["bt_wither_active"] = 1
		if user.Tokens["bt_wither_pending"] > 0 {
			user.Tokens["bt_wither_pending"]--
		}
		if target := e.State.Players[targetID]; target != nil {
			e.Log(fmt.Sprintf("%s 发动 [凋零]：对 %s 造成1点法术伤害，并对自己造成2点法术伤害；对方士气最低为1直到其下回合开始前", user.Name, target.Name))
		}
		if user.Tokens["bt_wither_pending"] > 0 {
			ctxData["choice_type"] = "bt_wither_confirm"
			ctxData["target_ids"] = e.butterflyEnemyIDs(user)
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return nil
		}
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			e.State.Phase = model.PhasePendingDamageResolution
		}
		return nil
	}
	if choiceType == "hom_dual_echo_target" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		var targetIDs []string
		if arr, ok := ctxData["target_ids"].([]string); ok {
			targetIDs = arr
		} else if arr, ok := ctxData["target_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					targetIDs = append(targetIDs, s)
				}
			}
		}
		if selectionIndex < 0 || selectionIndex >= len(targetIDs) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		costPending := toIntContextValue(ctxData["cost_pending"])
		if costPending > 0 {
			if !e.ConsumeCrystalCost(user.ID, costPending) {
				return fmt.Errorf("双重回响需要1蓝水晶（红宝石可替代）")
			}
			ctxData["cost_pending"] = 0
		}
		damage := 0
		if v, ok := ctxData["damage"].(int); ok {
			damage = v
		} else if f, ok := ctxData["damage"].(float64); ok {
			damage = int(f)
		}
		if damage < 0 {
			damage = 0
		}
		e.AddPendingDamage(model.PendingDamage{
			SourceID:           user.ID,
			TargetID:           targetIDs[selectionIndex],
			Damage:             damage,
			DamageType:         "magic",
			CapDrawToHandLimit: true,
			Stage:              0,
		})
		if target := e.State.Players[targetIDs[selectionIndex]]; target != nil {
			e.Log(fmt.Sprintf("%s 的 [双重回响] 对 %s 造成%d点法术伤害", user.Name, target.Name, damage))
		}
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil && len(e.State.PendingDamageQueue) > 0 {
			e.State.Phase = model.PhasePendingDamageResolution
		}
		return nil
	}
	if choiceType == "css_dance_mode" {
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("玩家不存在")
		}
		canCrystal, _ := ctxData["can_crystal"].(bool)
		canGem, _ := ctxData["can_gem"].(bool)
		var modeList []int
		if canCrystal {
			modeList = append(modeList, 0)
		}
		if canGem {
			modeList = append(modeList, 1)
		}
		if selectionIndex < 0 || selectionIndex >= len(modeList) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		mode := modeList[selectionIndex]
		if user.Character == nil || user.Character.Name == "" {
			return fmt.Errorf("角色信息缺失")
		}
		courtyardCard, ok := user.ConsumeExclusiveCard(user.Character.Name, "血蔷薇庭院")
		if !ok {
			return fmt.Errorf("未找到【血蔷薇庭院】专属技能卡")
		}
		user.AddFieldCard(&model.FieldCard{
			Card:     courtyardCard,
			OwnerID:  user.ID,
			SourceID: user.ID,
			Mode:     model.FieldEffect,
			Effect:   model.EffectRoseCourtyard,
			Trigger:  model.EffectTriggerManual,
		})
		user.Tokens["css_rose_courtyard_active"] = 1
		if mode == 0 {
			if !e.ConsumeCrystalCost(user.ID, 1) {
				return fmt.Errorf("蓝水晶不足（红宝石可替代）")
			}
			user.Tokens["css_blood_cap"] = 3
			addBlood(user, 2)
		} else {
			if user.Gem <= 0 {
				return fmt.Errorf("红宝石不足")
			}
			user.Gem--
			user.Tokens["css_blood_cap"] = 4
			addBlood(user, 2)
			overflow := len(user.Hand) - 4
			if overflow > 0 {
				e.PushInterrupt(&model.Interrupt{
					Type:     model.InterruptDiscard,
					PlayerID: user.ID,
					Context: map[string]interface{}{
						"discard_count": overflow,
						"stay_in_turn":  true,
						"prompt":        fmt.Sprintf("【散华轮舞】请弃置 %d 张手牌至4张：", overflow),
					},
				})
			}
		}
		e.PopInterrupt()
		return nil
	}

	return fmt.Errorf("未知的选择类型: %s", choiceType)
}

func (e *GameEngine) resolveElementalistBonus(ctxData map[string]interface{}, bonus bool, discardIdx int) error {
	userID, _ := ctxData["user_id"].(string)
	user := e.State.Players[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	targetID, _ := ctxData["damage_target_id"].(string)
	target := e.State.Players[targetID]
	if target == nil {
		return fmt.Errorf("目标不存在")
	}
	baseDamage := 0
	if v, ok := ctxData["base_damage"].(int); ok {
		baseDamage = v
	} else if f, ok := ctxData["base_damage"].(float64); ok {
		baseDamage = int(f)
	}
	skillName, _ := ctxData["skill_display_name"].(string)
	bonusElement, _ := ctxData["bonus_element"].(string)

	damage := baseDamage
	if bonus {
		if discardIdx < 0 || discardIdx >= len(user.Hand) {
			return fmt.Errorf("无效的弃牌索引")
		}
		card := user.Hand[discardIdx]
		if string(card.Element) != bonusElement {
			return fmt.Errorf("弃牌元素不匹配")
		}
		e.NotifyCardRevealed(userID, []model.Card{card}, "discard")
		user.Hand = append(user.Hand[:discardIdx], user.Hand[discardIdx+1:]...)
		e.State.DiscardPile = append(e.State.DiscardPile, card)
		damage++
	}
	e.InflictDamage(userID, targetID, damage, "magic")

	if healTargetID, ok := ctxData["heal_target_id"].(string); ok && healTargetID != "" {
		if hp := e.State.Players[healTargetID]; hp != nil {
			e.Heal(healTargetID, 1)
			e.Log(fmt.Sprintf("[元素师] %s 为 %s 提供了1点治疗", skillName, hp.Name))
		}
	}
	campGemBonus := 0
	if v, ok := ctxData["camp_gem_bonus"].(int); ok {
		campGemBonus = v
	} else if f, ok := ctxData["camp_gem_bonus"].(float64); ok {
		campGemBonus = int(f)
	}
	if campGemBonus > 0 {
		e.ModifyGem(string(user.Camp), campGemBonus)
	}
	grantAttack := false
	if v, ok := ctxData["grant_attack"].(bool); ok {
		grantAttack = v
	}
	grantMagic := false
	if v, ok := ctxData["grant_magic"].(bool); ok {
		grantMagic = v
	}
	if grantAttack {
		user.TurnState.PendingActions = append(user.TurnState.PendingActions, model.ActionContext{
			Source:   skillName,
			MustType: "Attack",
		})
	}
	if grantMagic {
		user.TurnState.PendingActions = append(user.TurnState.PendingActions, model.ActionContext{
			Source:   skillName,
			MustType: "Magic",
		})
	}
	e.Log(fmt.Sprintf("%s 发动 [%s]，对 %s 造成%d点法术伤害", user.Name, skillName, target.Name, damage))
	e.PopInterrupt()
	return nil
}

// resumePendingAttackHit 恢复被响应技能选择打断的“攻击命中后续结算”
func (e *GameEngine) resumePendingAttackHit(ctxData map[string]interface{}) {
	rawCtx, ok := ctxData["user_ctx"].(*model.Context)
	if !ok || rawCtx == nil || rawCtx.Trigger != model.TriggerOnAttackHit || rawCtx.TriggerCtx == nil {
		return
	}
	// OnAttackHit 的可选响应结束后，继续走 PendingDamage 的 Stage1/Stage2，
	// 避免直接 finishTakeHit 导致与队列重复结算（重复触发/重复伤害）。
	if e.advancePendingAttackDamageStageAfterHit(rawCtx) {
		if e.State.PendingInterrupt == nil {
			e.State.Phase = model.PhasePendingDamageResolution
		}
	}
}

// resolveHomunculusRuneChoice 结算英灵人形“战纹碎击/魔纹融合”的X/Y交互结果。
func (e *GameEngine) resolveHomunculusRuneChoice(ctxData map[string]interface{}, glyph bool) error {
	toInt := func(v interface{}) int {
		if n, ok := v.(int); ok {
			return n
		}
		if f, ok := v.(float64); ok {
			return int(f)
		}
		return 0
	}
	toIntSlice := func(v interface{}) []int {
		if arr, ok := v.([]int); ok {
			return append([]int{}, arr...)
		}
		var out []int
		if arr, ok := v.([]interface{}); ok {
			for _, item := range arr {
				if f, ok := item.(float64); ok {
					out = append(out, int(f))
				}
			}
		}
		return out
	}

	userID, _ := ctxData["user_id"].(string)
	user := e.State.Players[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	rawCtx, _ := ctxData["user_ctx"].(*model.Context)
	if rawCtx == nil || rawCtx.TriggerCtx == nil {
		return fmt.Errorf("英灵人形技能上下文丢失")
	}
	xVal := toInt(ctxData["x_value"])
	yVal := toInt(ctxData["y_value"])
	if xVal <= 0 || yVal < 0 {
		return fmt.Errorf("X/Y 参数无效")
	}
	selected := toIntSlice(ctxData["selected_indices"])
	if len(selected) != xVal {
		return fmt.Errorf("弃牌数量与X不一致")
	}

	attackElement, _ := ctxData["attack_element"].(string)
	glyphSelectedElements := map[model.Element]bool{}
	for _, idx := range selected {
		if idx < 0 || idx >= len(user.Hand) {
			return fmt.Errorf("无效的手牌索引: %d", idx)
		}
		if glyph {
			if attackElement != "" && string(user.Hand[idx].Element) == attackElement {
				return fmt.Errorf("魔纹融合需弃置异系牌")
			}
			if glyphSelectedElements[user.Hand[idx].Element] {
				return fmt.Errorf("魔纹融合需弃置元素互不相同的异系牌")
			}
			glyphSelectedElements[user.Hand[idx].Element] = true
		} else if attackElement != "" && string(user.Hand[idx].Element) != attackElement {
			return fmt.Errorf("战纹碎击需弃置同系牌")
		}
	}

	if user.Tokens == nil {
		user.Tokens = map[string]int{}
	}
	flipCount := 1 + yVal
	if glyph {
		if user.Tokens["hom_magic_rune"] < flipCount {
			return fmt.Errorf("魔纹不足，至少需要%d个", flipCount)
		}
		user.Tokens["hom_magic_rune"] -= flipCount
		user.Tokens["hom_war_rune"] += flipCount
	} else {
		if user.Tokens["hom_war_rune"] < flipCount {
			return fmt.Errorf("战纹不足，至少需要%d个", flipCount)
		}
		user.Tokens["hom_war_rune"] -= flipCount
		user.Tokens["hom_magic_rune"] += flipCount
	}

	removed, err := removeCardsByIndicesFromHand(user, append([]int{}, selected...))
	if err != nil {
		return err
	}
	e.NotifyCardRevealed(user.ID, removed, "discard")
	e.State.DiscardPile = append(e.State.DiscardPile, removed...)

	targetID := rawCtx.TriggerCtx.TargetID
	if glyph {
		damage := xVal - 1 + yVal
		if damage < 0 {
			damage = 0
		}
		if damage > 0 && targetID != "" {
			e.AddPendingDamage(model.PendingDamage{
				SourceID:   user.ID,
				TargetID:   targetID,
				Damage:     damage,
				DamageType: "magic",
				Stage:      0,
			})
		}
		e.Log(fmt.Sprintf("%s 发动 [魔纹融合]：弃%d张异系牌，翻转%d个魔纹为战纹，额外造成%d点法术伤害", user.Name, xVal, flipCount, damage))
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil && rawCtx.Trigger == model.TriggerOnAttackMiss {
			if e.resumePendingAttackMiss(rawCtx) {
				return nil
			}
		}
		if e.State.PendingInterrupt == nil && len(e.State.PendingDamageQueue) > 0 {
			e.State.Phase = model.PhasePendingDamageResolution
		}
		return nil
	}

	bonusDamage := xVal - 1
	if bonusDamage < 0 {
		bonusDamage = 0
	}
	if rawCtx.TriggerCtx.DamageVal != nil && bonusDamage > 0 {
		*rawCtx.TriggerCtx.DamageVal += bonusDamage
	}
	if yVal > 0 && targetID != "" {
		e.AddPendingDamage(model.PendingDamage{
			SourceID:   user.ID,
			TargetID:   targetID,
			Damage:     yVal,
			DamageType: "magic",
			Stage:      0,
		})
	}
	e.Log(fmt.Sprintf("%s 发动 [战纹碎击]：弃%d张同系牌，翻转%d个战纹为魔纹，本次攻击伤害+%d", user.Name, xVal, flipCount, bonusDamage))
	e.PopInterrupt()
	e.resumePendingAttackHit(ctxData)
	return nil
}

// resolveSelectionToCandidate 兼容两种前端回传：
// 1) 选项序号（0..n-1）
// 2) 选项ID就是候选值本身（如手牌原始索引）
func resolveSelectionToCandidate(selection int, candidates []int) (int, bool) {
	if len(candidates) == 0 {
		return 0, false
	}
	for _, v := range candidates {
		if v == selection {
			return v, true
		}
	}
	if selection >= 0 && selection < len(candidates) {
		return candidates[selection], true
	}
	return 0, false
}

func (e *GameEngine) resolveAdventurerLuckyFortuneFromFraud(user *model.Player) {
	if user == nil {
		return
	}
	user.Crystal++
	e.Log(fmt.Sprintf("%s 的 [强运] 触发，获得1蓝水晶", user.Name))
	e.Log(fmt.Sprintf("[Skill] %s 使用了技能: 强运", user.Name))
}

func (e *GameEngine) buildFraudCombos(user *model.Player, element model.Element, need int, allowAnyElementForDark bool) []string {
	if user == nil || need <= 0 {
		return nil
	}
	elemToIdx := map[model.Element][]int{}
	for i, c := range user.Hand {
		elemToIdx[c.Element] = append(elemToIdx[c.Element], i)
	}

	var targets []model.Element
	if allowAnyElementForDark {
		for ele, idxs := range elemToIdx {
			if len(idxs) >= need {
				targets = append(targets, ele)
			}
		}
	} else {
		if len(elemToIdx[element]) >= need {
			targets = append(targets, element)
		}
	}

	var combos []string
	for _, ele := range targets {
		idxs := elemToIdx[ele]
		for _, picked := range pickKIndices(idxs, need) {
			parts := make([]string, 0, len(picked))
			for _, v := range picked {
				parts = append(parts, fmt.Sprintf("%d", v))
			}
			combos = append(combos, fmt.Sprintf("%s:%s", ele, strings.Join(parts, ",")))
		}
	}
	return combos
}

func pickKIndices(src []int, k int) [][]int {
	var out [][]int
	var dfs func(start int, cur []int)
	dfs = func(start int, cur []int) {
		if len(cur) == k {
			cp := append([]int{}, cur...)
			out = append(out, cp)
			return
		}
		for i := start; i < len(src); i++ {
			cur = append(cur, src[i])
			dfs(i+1, cur)
			cur = cur[:len(cur)-1]
		}
	}
	dfs(0, nil)
	return out
}

func (e *GameEngine) buildContext(user *model.Player, target *model.Player, trigger model.TriggerType, eventCtx *model.EventContext) *model.Context {
	ctx := &model.Context{
		Game:       e,
		User:       user,
		Target:     target,
		Trigger:    trigger,
		TriggerCtx: eventCtx,
		// 初始化 map 避免 handler 写入时 panic
		Selections: make(map[string]any),
		Flags:      make(map[string]bool),
		// 当前PendingInterrupt （仅供Handler读取，不要修改）
		PendingInterrupt: e.State.PendingInterrupt,
		// 自动将单个 Target 包装进 Targets 切片，方便多目标技能处理
		Targets: []*model.Player{},
	}

	if target != nil {
		ctx.Targets = append(ctx.Targets, target)
	}

	return ctx
}

// AddPendingDamage 将延迟伤害添加到队列
func (e *GameEngine) AddPendingDamage(pd model.PendingDamage) {
	e.State.PendingDamageQueue = append(e.State.PendingDamageQueue, pd)
	e.Log(fmt.Sprintf("[System] 延迟伤害已添加: Source: %s, Target: %s, Damage: %d, Type: %s",
		pd.SourceID, pd.TargetID, pd.Damage, pd.DamageType))

	// 确保引擎会处理此伤害
	// 如果当前不在处理伤害阶段，切换状态并在处理后返回原阶段
	if e.State.Phase != model.PhasePendingDamageResolution {
		if e.State.ReturnPhase == "" {
			e.State.ReturnPhase = e.State.Phase
		}
		e.State.Phase = model.PhasePendingDamageResolution
	}
}

// AddPendingDamageFront 将延迟伤害插入队列头部（用于“必须先结算”的伤害）。
func (e *GameEngine) AddPendingDamageFront(pd model.PendingDamage) {
	e.State.PendingDamageQueue = append([]model.PendingDamage{pd}, e.State.PendingDamageQueue...)
	e.Log(fmt.Sprintf("[System] 延迟伤害已前插: Source: %s, Target: %s, Damage: %d, Type: %s",
		pd.SourceID, pd.TargetID, pd.Damage, pd.DamageType))

	if e.State.Phase != model.PhasePendingDamageResolution {
		if e.State.ReturnPhase == "" {
			e.State.ReturnPhase = e.State.Phase
		}
		e.State.Phase = model.PhasePendingDamageResolution
	}
}

// processPendingDamages 处理伤害队列中的所有伤害
// 返回 true 如果产生了中断需要暂停 Drive
func (e *GameEngine) processPendingDamages() bool {
	for len(e.State.PendingDamageQueue) > 0 {
		// Peek: 取出队列中第一个延迟伤害（暂不弹出，等待所有步骤完成）
		pd := &e.State.PendingDamageQueue[0]

		// Stage 0: 初始化 & 攻击命中触发 (OnAttackHit)
		if pd.Stage == 0 {
			// 如果是攻击伤害，且有卡牌上下文
			if pd.DamageType == "Attack" && pd.Card != nil {
				attacker := e.State.Players[pd.SourceID]
				victim := e.State.Players[pd.TargetID]

				if attacker != nil && victim != nil {
					// 1. 应用被动效果 (如精准射击、狂化) - 这里可能会修改 pd.Damage
					action := model.Action{
						SourceID: pd.SourceID,
						TargetID: pd.TargetID,
						Type:     model.ActionAttack,
						Card:     pd.Card,
						CounterInitiator: func() string {
							if pd.IsCounter {
								return pd.SourceID
							}
							return ""
						}(),
					}
					pd.Damage = e.applyPassiveAttackEffects(attacker, victim, pd.Damage, action)
				}

				// 2. 攻击命中加星石：主动攻击→宝石，应战→水晶（战绩区上限5）
				if pd.IsCounter {
					e.addCampResource(attacker.Camp, "crystal")
					e.Log(fmt.Sprintf("[Combat] 应战攻击命中！%s 方战绩区+1水晶", attacker.Camp))
				} else {
					e.addCampResource(attacker.Camp, "gem")
					e.Log(fmt.Sprintf("[Combat] 主动攻击命中！%s 方战绩区+1宝石", attacker.Camp))
				}

				// 3. 触发 OnAttackHit (如撕裂)
				hitEventCtx := &model.EventContext{
					Type:      model.EventAttack,
					SourceID:  pd.SourceID,
					TargetID:  pd.TargetID,
					DamageVal: &pd.Damage, // 允许技能修改伤害
					Card:      pd.Card,
					AttackInfo: &model.AttackEventInfo{
						ActionType: "Attack",
						IsHit:      true,
						CounterInitiator: func() string {
							if pd.IsCounter {
								return pd.SourceID
							}
							return ""
						}(),
					},
				}
				hitCtx := e.buildContext(e.State.Players[pd.SourceID], e.State.Players[pd.TargetID], model.TriggerOnAttackHit, hitEventCtx)
				e.dispatcher.OnTrigger(model.TriggerOnAttackHit, hitCtx)

				// 如果触发了中断 (例如询问是否发动撕裂)，暂停处理
				if e.State.PendingInterrupt != nil {
					return true
				}
				// 处理攻击命中后的附加技能分支（如元素射击后续/黄泉震颤）。
				if e.handlePostAttackHitEffects(pd) {
					// 该命中已进入伤害阶段，避免恢复后再次重复触发 OnAttackHit。
					pd.Stage = 1
					return true
				}
			}
			// 完成 Stage 0，进入下一阶段
			pd.Stage = 1
		}

		// Stage 1: 受伤触发 (OnDamageTaken / 圣盾 / 减伤)
		if pd.Stage == 1 {
			// 灵魂术士：灵魂链接在“承受伤害前”可转移部分伤害。
			if e.maybeTriggerSoulLinkTransfer(pd) {
				return true
			}

			damageEventCtx := &model.EventContext{
				Type:      model.EventDamage,
				SourceID:  pd.SourceID,
				TargetID:  pd.TargetID,
				DamageVal: &pd.Damage, // 允许技能修改伤害
				Card:      pd.Card,
			}
			damageCtx := e.buildContext(e.State.Players[pd.TargetID], e.State.Players[pd.SourceID], model.TriggerOnDamageTaken, damageEventCtx)
			damageCtx.Flags["IsMagicDamage"] = (pd.DamageType != "Attack" && pd.DamageType != "attack")
			if strings.Contains(strings.ToLower(pd.DamageType), "no_absorb") {
				damageCtx.Flags["NoElementAbsorb"] = true
			}

			e.dispatcher.OnTrigger(model.TriggerOnDamageTaken, damageCtx)

			// 如果触发了中断 (例如询问是否发动减伤技能)，暂停处理
			if e.State.PendingInterrupt != nil {
				return true
			}
			// 治疗选择阶段：允许受伤方选择是否使用治疗抵消
			if !pd.HealResolved {
				target := e.State.Players[pd.TargetID]
				if target != nil && pd.Damage > 0 && e.canUseHealToResist(target, pd.SourceID, pd.DamageType, pd.IgnoreHeal, pd.AllowCrimsonFaithHeal) {
					maxHeal := target.Heal
					if pd.Damage < maxHeal {
						maxHeal = pd.Damage
					}
					// 神官：每次仅可使用1点治疗抵挡伤害。
					if e.isPriest(target) && maxHeal > 1 {
						maxHeal = 1
					}
					e.PushInterrupt(&model.Interrupt{
						Type:     model.InterruptChoice,
						PlayerID: pd.TargetID,
						Context: map[string]interface{}{
							"choice_type":  "heal",
							"max_heal":     maxHeal,
							"damage_index": 0,
						},
					})
					return true
				}
				pd.HealResolved = true
			}
			// 蝶舞者：伤害应用前的时点响应（朝圣/毒粉/镜花水月）。
			if e.maybeTriggerButterflyDamageResponses(pd) {
				return true
			}
			// 完成 Stage 1
			pd.Stage = 2
		}

		// Stage 2: 应用伤害 & 移除效果
		if pd.Stage == 2 {
			if pd.Damage < 0 {
				pd.Damage = 0
			}

			target := e.State.Players[pd.TargetID]
			source := e.State.Players[pd.SourceID]
			if target != nil && pd.Damage > 0 {
				if pd.DamageType == "Attack" && source != nil {
					e.NotifyActionStep(fmt.Sprintf("总共对%s造成%d点伤害", model.GetPlayerDisplayName(target), pd.Damage))
				}
				e.NotifyDamageDealt(pd.SourceID, pd.TargetID, pd.Damage, pd.DamageType)
			}
			if target != nil {
				// 执行实际扣血/摸牌逻辑
				e.applyDamageWithOptions(target, pd.Damage, pd.DamageType, pd.CapDrawToHandLimit)
				if target.Tokens != nil {
					target.Tokens["css_blood_barrier_lock"] = 0
					target.Tokens["bw_substitute_lock"] = 0
					target.Tokens["bw_mana_inversion_lock"] = 0
				}

				// 如果指定了 EffectTypeToRemove，在伤害结算后移除场上效果
				if pd.EffectTypeToRemove != "" {
					e.RemoveFieldCard(target.ID, pd.EffectTypeToRemove)
					e.Log(fmt.Sprintf("[System] 移除了 %s 的场上效果: %s", target.Name, pd.EffectTypeToRemove))
				}
			}
			resolved := *pd

			// 处理完毕，从队列中弹出
			e.State.PendingDamageQueue = e.State.PendingDamageQueue[1:]
			// 伤害结算后触发额外技能（例如动物伙伴）。
			if e.handlePostDamageResolved(&resolved) {
				return true
			}

			// 伤害结算可能产生新的中断 (例如爆牌弃牌)，如果有中断，暂停
			if e.State.PendingInterrupt != nil {
				return true
			}
		}
	}
	return false
}
