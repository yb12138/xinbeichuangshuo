// gameflow: 精灵射手技能处理器。

package elf_archer

import (
	"fmt"
	"strings"

	"starcup-engine/internal/engine/player"
	skills "starcup-engine/internal/engine/skill"
	"starcup-engine/internal/model"
)

// ElfElementalShotHandler 元素射击
type ElfElementalShotHandler struct{ skills.BaseHandler }

// ElfAnimalCompanionHandler 动物伙伴
type ElfAnimalCompanionHandler struct{ skills.BaseHandler }

// ElfRitualHandler 精灵密仪
type ElfRitualHandler struct{ skills.BaseHandler }

// ElfPetEmpowerHandler 宠物强化
type ElfPetEmpowerHandler struct{ skills.BaseHandler }

// CanUse implements skills.BaseHandler for ElfElementalShotHandler.
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

// Execute implements skills.BaseHandler for ElfElementalShotHandler.
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

// CanUse implements skills.BaseHandler for ElfAnimalCompanionHandler.
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

// Execute implements skills.BaseHandler for ElfAnimalCompanionHandler.
func (h *ElfAnimalCompanionHandler) Execute(ctx *model.Context) error {
	return resolveElfForcedDrawDiscard(ctx.Game, ctx.User, "【动物伙伴】请选择弃置1张牌：", true)
}

// CanUse implements skills.BaseHandler for ElfRitualHandler.
func (h *ElfRitualHandler) CanUse(ctx *model.Context) bool {
	return ctx.User.Gem > 0 && !player.HasForm(ctx.User, model.FormElfArcherRitual)
}

// Execute implements skills.BaseHandler for ElfRitualHandler.
func (h *ElfRitualHandler) Execute(ctx *model.Context) error {
	if ctx.User.Gem <= 0 {
		return fmt.Errorf("精灵密仪需要至少1个红宝石")
	}
	ctx.User.Gem--
	player.SetForm(ctx.User, model.FormElfArcherRitual)
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

// CanUse implements skills.BaseHandler for ElfPetEmpowerHandler.
func (h *ElfPetEmpowerHandler) CanUse(ctx *model.Context) bool {
	if !skills.CanPayCrystalLike(ctx, 1) {
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

// Execute implements skills.BaseHandler for ElfPetEmpowerHandler.
func (h *ElfPetEmpowerHandler) Execute(ctx *model.Context) error {
	if !skills.CanPayCrystalLike(ctx, 1) {
		return fmt.Errorf("宠物强化需要至少1个蓝水晶")
	}
	if ctx == nil || ctx.Target == nil {
		return fmt.Errorf("宠物强化缺少受伤目标")
	}
	if !skills.SpendCrystalLike(ctx, 1) {
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

// resolveElfForcedDrawDiscard 精灵射手动物伙伴/宠物强化强制摸弃牌。
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
