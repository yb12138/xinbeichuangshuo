// gameflow: 封印师 handler。

package sealer

import (
	"fmt"
	engineplayer "starcup-engine/internal/engine/player"

	"starcup-engine/internal/model"
)

// --- Magic Surge Handler ---

type MagicSurgeHandler struct{ engineplayer.BaseHandler }

func (h *MagicSurgeHandler) CanUse(ctx *model.Context) bool {
	if ctx.EventCtx == nil {
		return false
	}
	// [Correction] As long as it's a magic action (including magic cards and active skills), the condition is met
	return ctx.EventCtx.ActionType == model.ActionMagic
}

func (h *MagicSurgeHandler) Execute(ctx *model.Context) error {
	// Magic Surge: (Triggered at the end of [Magic Action]) Gain an additional +1 [Attack Action]
	// Add an unrestricted attack action token to the action queue
	model.AppendAttackAction(ctx.User, "法术激荡")
	ctx.Game.Log(fmt.Sprintf("%s 发动 [法术激荡]，额外获得1次攻击行动", ctx.User.Name))
	return nil
}

// --- Seal Break Handler ---

type SealBreakHandler struct{ engineplayer.BaseHandler }

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

func (h *SealBreakHandler) Execute(ctx *model.Context) error {
	// Seal Break: Take any basic effect card from the field into your hand.
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

	// If there are multiple basic effects on the field, prompt the sealer to choose which one.
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

// --- Five Elements Bind Handler ---

type FiveElementsBindHandler struct{ engineplayer.BaseHandler }

func (h *FiveElementsBindHandler) CanUse(ctx *model.Context) bool {
	return ctx != nil && ctx.Timing == model.TimingActionDuring && ctx.User != nil && ctx.Target != nil
}

func (h *FiveElementsBindHandler) Execute(ctx *model.Context) error {
	if !h.CanUse(ctx) {
		return nil
	}
	ctx.Game.Log(fmt.Sprintf("%s 对 %s 发动五系束缚", ctx.User.Name, ctx.Target.Name))
	return nil
}

// ==========================================
// Five Elements Seal Handler Design Notes
// ==========================================
// The complete flow of five elements seal is divided into three phases:
//
// Phase 1: Place Seal (handled by skill usage flow)
//   - Sealer uses skill (Water Seal, etc.)
//   - UseSkill -> consumeSkillInputs -> placeSkillFieldCard
//   - Field card is placed in front of target player, Meta records bound element
//
// Phase 2: Trigger Seal (handled by SkillDispatcher)
//   - Target player plays/reveals matching element card
//   - Trigger TimingOnCardPlayedOrRevealed
//   - collectSkillsForTiming iterates Field, finds matching seal
//   - SealLogic.CanUse -> canResolveElementalSealStatus
//
// Phase 3: Resolve Damage (handled by processPendingDamages)
//   - SealLogic.Execute -> executeElementalSealStatus
//   - Add PendingDamage, mark EffectTypeToRemove
//   - After damage is applied, seal removal hook removes the seal
// ==========================================

// SealLogic Common handler logic for five elements seals
// Only retains the entry mapping after placement;
// The actual trigger rules are delegated to field status resolver to avoid coupling with main flow.
type SealLogic struct {
	EffectType model.EffectType // Corresponding Effect enum, used for removal
}

func (s *SealLogic) CanUse(ctx *model.Context) bool {
	return canResolveFieldStatus(ctx, s.EffectType)
}

func (s *SealLogic) Execute(ctx *model.Context) error {
	if ctx != nil && ctx.Timing == model.TimingActionDuring {
		return nil
	}
	return executeFieldStatus(ctx, s.EffectType)
}

// Field status resolver for elemental seals
type fieldStatusResolverSpec struct {
	matchesEffect func(effect model.EffectType) bool
	canUse        func(ctx *model.Context, fc *model.FieldCard) bool
	execute       func(ctx *model.Context, fc *model.FieldCard) error
}

var fieldStatusResolverSpecs = []fieldStatusResolverSpec{
	{
		matchesEffect: model.IsElementalSealEffect,
		canUse:        canResolveElementalSealStatus,
		execute:       executeElementalSealStatus,
	},
}

func canResolveFieldStatus(ctx *model.Context, effect model.EffectType) bool {
	spec, fc := resolveFieldStatusSpec(ctx, effect)
	return spec != nil && fc != nil && spec.canUse(ctx, fc)
}

func executeFieldStatus(ctx *model.Context, effect model.EffectType) error {
	spec, fc := resolveFieldStatusSpec(ctx, effect)
	if spec == nil || fc == nil {
		return nil
	}
	return spec.execute(ctx, fc)
}

func resolveFieldStatusSpec(ctx *model.Context, effect model.EffectType) (*fieldStatusResolverSpec, *model.FieldCard) {
	if ctx == nil || ctx.User == nil {
		return nil, nil
	}
	for i := range fieldStatusResolverSpecs {
		spec := &fieldStatusResolverSpecs[i]
		if !spec.matchesEffect(effect) {
			continue
		}
		for _, fc := range ctx.User.Field {
			if fc == nil || fc.Mode != model.FieldEffect || fc.Effect != effect {
				continue
			}
			return spec, fc
		}
	}
	return nil, nil
}

// canResolveElementalSealStatus checks if five elements seal can trigger
// Trigger conditions:
//  1. Timing is TimingOnCardPlayedOrRevealed (play or reveal card)
//  2. The element of the played/revealed card matches the bound element of the seal
func canResolveElementalSealStatus(ctx *model.Context, fc *model.FieldCard) bool {
	if ctx == nil || ctx.User == nil || fc == nil {
		return false
	}
	// Only trigger on "play card" or "reveal card"
	if ctx.Timing != model.TimingCardPlayedRevealed {
		return false
	}
	if ctx.EventCtx == nil || ctx.EventCtx.Card == nil {
		return false
	}
	// Get bound element (read from Meta, or infer from EffectType)
	boundElement := model.BoundElementForFieldCard(fc)
	// Check if the played/revealed card's element matches the seal element
	return boundElement != "" && ctx.EventCtx.Card.Element == boundElement
}

// executeElementalSealStatus executes the five elements seal effect
// Effect:
//  1. Deal 3 magic damage to target
//  2. Remove seal after damage is applied (via EffectTypeToRemove marker)
func executeElementalSealStatus(ctx *model.Context, fc *model.FieldCard) error {
	if !canResolveElementalSealStatus(ctx, fc) {
		return nil
	}
	// Damage source: player who placed the seal (SourceID)
	sourceID := fc.SourceID
	if sourceID == "" {
		sourceID = ctx.User.ID
	}
	actionWord := "打出"
	if ctx.Timing == model.TimingCardPlayedRevealed {
		actionWord = "展示"
	}
	ctx.Game.Log(fmt.Sprintf(
		"[Seal] %s %s了%s系牌，触发了%s",
		ctx.User.Name,
		actionWord,
		model.BoundElementForFieldCard(fc),
		elementalSealName(fc),
	))
	// Add pending damage, mark EffectTypeToRemove for removal after damage is applied
	ctx.Game.AddPendingDamage(model.PendingDamage{
		SourceID:           sourceID,
		TargetID:           ctx.User.ID,
		Damage:             3,
		DamageType:         model.MagicAttack,
		EffectTypeToRemove: fc.Effect, // [Key] Remove this seal after damage is applied
	})
	return nil
}

// elementalSealName gets the seal name
// Prefer card name on field card, otherwise return default name based on EffectType
func elementalSealName(fc *model.FieldCard) string {
	if fc != nil && fc.Card.Name != "" {
		return fc.Card.Name
	}
	switch fc.Effect {
	case model.EffectSealWater:
		return "水之封印"
	case model.EffectSealFire:
		return "火之封印"
	case model.EffectSealEarth:
		return "地之封印"
	case model.EffectSealWind:
		return "风之封印"
	case model.EffectSealThunder:
		return "雷之封印"
	default:
		return string(fc.Effect)
	}
}

// Five Elements Seal Handlers (placement handled by PlaceCard, trigger handled by generic status resolver)
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
