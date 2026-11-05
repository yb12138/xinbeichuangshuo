// gameflow: 精灵射手技能处理器。

package elf_archer

import (
	"fmt"
	engineplayer "starcup-engine/internal/engine/player"
	"strings"

	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// ElfElementalShotHandler 元素射击
type ElfElementalShotHandler struct{ engineplayer.BaseHandler }

// ElfAnimalCompanionHandler 动物伙伴
type ElfAnimalCompanionHandler struct{ engineplayer.BaseHandler }

// ElfRitualHandler 精灵密仪
type ElfRitualHandler struct{ engineplayer.BaseHandler }

// ElfPetEmpowerHandler 宠物强化
type ElfPetEmpowerHandler struct{ engineplayer.BaseHandler }

// CanUse implements engineplayer.BaseHandler for ElfElementalShotHandler.
func (h *ElfElementalShotHandler) CanUse(ctx *model.Context) bool {
	if !ctx.AttackDeclarePhase() || ctx.EventCtx == nil || ctx.EventCtx.Card == nil {
		return false
	}
	if ctx.EventCtx.Card.Element == model.ElementDark {
		return false
	}
	if ctx.EventCtx.AttackInfo != nil && ctx.EventCtx.AttackInfo.CounterInitiator != "" {
		return false
	}
	attackCardID := ctx.EventCtx.Card.ID
	// 手牌中排除攻击牌后是否还有法术牌
	for _, c := range ctx.User.Hand {
		if c.Type == model.CardTypeMagic && c.ID != attackCardID {
			return true
		}
	}
	// 祝福中排除攻击牌后是否还有可用祝福
	for _, c := range elfBlessingCardsLocal(ctx.User) {
		if c.ID != attackCardID {
			return true
		}
	}
	return false
}

// Execute implements engineplayer.BaseHandler for ElfElementalShotHandler.
func (h *ElfElementalShotHandler) Execute(ctx *model.Context) error {
	if ctx.EventCtx == nil || ctx.EventCtx.Card == nil {
		return nil
	}
	attackCardID := ""
	if ctx.EventCtx.Card != nil {
		attackCardID = ctx.EventCtx.Card.ID
	}
	candidateCount := 0
	for _, c := range ctx.User.Hand {
		if c.Type == model.CardTypeMagic && c.ID != attackCardID {
			candidateCount++
		}
	}
	for _, c := range elfBlessingCardsLocal(ctx.User) {
		if c.ID != attackCardID {
			candidateCount++
		}
	}
	if candidateCount == 0 {
		return fmt.Errorf("元素射击需要至少1张法术牌或祝福")
	}
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type":    "elf_archer_elemental_shot_pick",
			"user_id":        ctx.User.ID,
			"attack_element": string(ctx.EventCtx.Card.Element),
			"attack_card_id": attackCardID,
			"user_ctx":       ctx,
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 可发动 [元素射击]，请选择1张法术牌或祝福发动", ctx.User.Name))
	return nil
}

// CanUse implements engineplayer.BaseHandler for ElfAnimalCompanionHandler.
func (h *ElfAnimalCompanionHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.Game == nil || ctx.Target == nil || ctx.EventCtx == nil || ctx.EventCtx.DamageVal == nil {
		return false
	}
	// 仅在⑥"承受伤害后"时点（由 postDamageResolvedHook 打标）触发，不在伤害结算③时点触发。
	if !ctx.DamageTakenPhase() || !ctx.Flags[elfPostDamageResolvedFlag] || *ctx.EventCtx.DamageVal <= 0 {
		return false
	}
	if ctx.EventCtx.SourceID != ctx.User.ID || ctx.EventCtx.TargetID == "" || ctx.EventCtx.TargetID == ctx.User.ID {
		return false
	}
	// 自己回合内造成的任意伤害（含法术伤害）均可触发。
	return ctx.User.IsActive
}

// Execute implements engineplayer.BaseHandler for ElfAnimalCompanionHandler.
func (h *ElfAnimalCompanionHandler) Execute(ctx *model.Context) error {
	return resolveElfForcedDrawDiscard(ctx.Game, ctx.User, "elf_animal_companion")
}

// CanUse implements engineplayer.BaseHandler for ElfRitualHandler.
func (h *ElfRitualHandler) CanUse(ctx *model.Context) bool {
	return ctx.User.Gem > 0 && !player.HasForm(ctx.User, model.FormElfArcherRitual)
}

// Execute implements engineplayer.BaseHandler for ElfRitualHandler.
func (h *ElfRitualHandler) Execute(ctx *model.Context) error {
	// CostGem 已在 ConfirmStartupSkillAction 由框架统一扣减（见 skill definition CostGem: 1）
	// 牌库顶3张直接作为祝福，不走摸牌流程（避免触发摸牌钩子 / 被摸牌中断打断）。
	cards, _ := ctx.Game.DrawRawCards(3)
	if len(cards) < 3 {
		// 失败回滚：退还宝石，不进入形态，翻出的牌进弃牌堆。
		if len(cards) > 0 {
			ctx.Game.AppendToDiscard(cards)
		}
		ctx.User.Gem++
		return fmt.Errorf("精灵密仪抽取祝福数量不足")
	}
	player.SetForm(ctx.User, model.FormElfArcherRitual)
	markElfBlessingCards(ctx.User, cards)
	ctx.Game.Log(fmt.Sprintf("%s 发动 [精灵密仪]，进入精灵祝福形态并获得3张祝福", ctx.User.Name))
	return nil
}

// CanUse implements engineplayer.BaseHandler for ElfPetEmpowerHandler.
func (h *ElfPetEmpowerHandler) CanUse(ctx *model.Context) bool {
	if !engineplayer.CanPayCrystalLike(ctx, 1) {
		return false
	}
	if ctx == nil || ctx.User == nil || ctx.Target == nil || ctx.EventCtx == nil || ctx.EventCtx.DamageVal == nil {
		return false
	}
	// 仅在⑥"承受伤害后"时点（由 postDamageResolvedHook 打标）触发，不在伤害结算③时点触发。
	if !ctx.DamageTakenPhase() || !ctx.Flags[elfPostDamageResolvedFlag] || *ctx.EventCtx.DamageVal <= 0 {
		return false
	}
	if ctx.EventCtx.SourceID != ctx.User.ID || ctx.EventCtx.TargetID != ctx.Target.ID {
		return false
	}
	// 自己回合内对对手造成的任意伤害（含法术伤害）均可触发。
	return ctx.User.IsActive && ctx.Target.Camp != ctx.User.Camp
}

// Execute implements engineplayer.BaseHandler for ElfPetEmpowerHandler.
func (h *ElfPetEmpowerHandler) Execute(ctx *model.Context) error {
	if !engineplayer.CanPayCrystalLike(ctx, 1) {
		return fmt.Errorf("宠物强化需要至少1个蓝水晶")
	}
	if ctx == nil || ctx.Target == nil {
		return fmt.Errorf("宠物强化缺少受伤目标")
	}
	if !engineplayer.SpendCrystalLike(ctx, 1) {
		return fmt.Errorf("宠物强化结算失败：水晶不足（红宝石可替代）")
	}
	ctx.Game.Log(fmt.Sprintf("%s 发动 [宠物强化]，动物伙伴效果改为目标摸1弃1", ctx.User.Name))
	return resolveElfForcedDrawDiscard(ctx.Game, ctx.Target, "elf_pet_empower")
}

// resolveElfForcedDrawDiscard 精灵射手动物伙伴/宠物强化强制摸弃牌。
// 摸牌后若爆牌，先结算爆牌弃牌（弃到上限，可能掉士气），之后再结算技能本身的"弃1张"：
// 在爆牌弃牌中断上挂续接，由 FlowContinuationAfterDiscard -> handleElfArcherAfterDiscard 推出技能弃牌。
func resolveElfForcedDrawDiscard(game model.IGameEngine, target *model.Player, skillID string) error {
	if game == nil || target == nil {
		return fmt.Errorf("动物伙伴结算目标无效")
	}
	game.DrawCards(target.ID, 1)
	if len(target.Hand) > game.GetMaxHand(target) {
		if attachElfOverflowDiscardContinuation(game, target.ID, skillID) {
			game.Log(fmt.Sprintf("[精灵射手] %s 摸牌后爆牌，先结算爆牌弃牌，之后再弃1张", target.Name))
			return nil
		}
		// 未找到可挂续接的爆牌弃牌中断（理论上不应发生）：按队列顺序追加技能弃牌，仍排在爆牌弃牌之后。
	}
	pushElfSkillDiscard(game, target, skillID)
	return nil
}

// pushElfSkillDiscard 推出技能本身的"弃1张牌"；手牌为空时跳过。
func pushElfSkillDiscard(game model.IGameEngine, target *model.Player, skillID string) {
	if game == nil || target == nil {
		return
	}
	if len(target.Hand) == 0 {
		game.Log(fmt.Sprintf("[精灵射手] %s 已无手牌，跳过技能弃牌", target.Name))
		return
	}
	prompt := "【动物伙伴】请选择弃置1张牌："
	if skillID == "elf_pet_empower" {
		prompt = fmt.Sprintf("【宠物强化】%s 请弃置1张牌：", target.Name)
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
			"exclude_blessings": engineplayer.IsCharacter(target, "elf_archer"),
		},
	})
}

// attachElfOverflowDiscardContinuation 找到目标刚被推出的爆牌弃牌中断（当前中断或队列中），挂上精灵射手续接。
func attachElfOverflowDiscardContinuation(game model.IGameEngine, targetID, skillID string) bool {
	stateHolder, ok := game.(interface{ GetState() *model.GameState })
	if !ok || stateHolder.GetState() == nil {
		return false
	}
	st := stateHolder.GetState()
	candidates := make([]*model.Interrupt, 0, len(st.InterruptQueue)+1)
	candidates = append(candidates, st.PendingInterrupt)
	candidates = append(candidates, st.InterruptQueue...)
	// 从后往前找：最近推入的那条才是本次摸牌产生的爆牌弃牌。
	for i := len(candidates) - 1; i >= 0; i-- {
		intr := candidates[i]
		if intr == nil || intr.Type != model.InterruptChoice || intr.PlayerID != targetID {
			continue
		}
		data, ok := intr.Context.(map[string]interface{})
		if !ok || data == nil {
			continue
		}
		if reason, _ := data["discard_reason"].(string); reason != "hand_overflow" {
			continue
		}
		if victimID, _ := data["victim_id"].(string); victimID != targetID {
			continue
		}
		if existing, _ := data["flow_continuation_role_id"].(string); existing != "" {
			return false
		}
		data["flow_continuation_role_id"] = "elf_archer"
		data["flow_continuation_player_id"] = targetID
		data["flow_continuation_skill_id"] = skillID
		return true
	}
	return false
}

// handleElfArcherAfterDiscard 爆牌弃牌（含士气结算）完成后，结算动物伙伴/宠物强化的"弃1张"。
func handleElfArcherAfterDiscard(rt engineplayer.ChoiceRuntime, cont model.FlowContinuation) error {
	if cont.SkillID != "elf_animal_companion" && cont.SkillID != "elf_pet_empower" {
		return nil
	}
	target := rt.GetPlayers()[cont.PlayerID]
	if target == nil {
		return fmt.Errorf("动物伙伴弃牌目标不存在: %s", cont.PlayerID)
	}
	pushElfSkillDiscard(rt, target, cont.SkillID)
	return nil
}

// markElfBlessingCards 将卡牌标记为精灵祝福盖牌。
func markElfBlessingCards(p *model.Player, cards []model.Card) {
	if p == nil || len(cards) == 0 {
		return
	}
	existsBless := map[string]bool{}
	for _, c := range elfBlessingCardsLocal(p) {
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
	blessings := elfBlessingCardsLocal(p)
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

func elfBlessingCardsLocal(p *model.Player) []model.Card {
	covers := p.GetCoverCardsByEffect(model.EffectElfBlessing)
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
