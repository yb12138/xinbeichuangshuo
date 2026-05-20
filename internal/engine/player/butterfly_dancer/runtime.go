package butterfly_dancer

import (
	"fmt"

	"starcup-engine/internal/engine/hook/promptfmt"
	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

const CocoonCap = 8

// ActionTargetIDs returns all active player IDs.
type butterflyRuntime interface {
	GetPlayers() map[string]*model.Player
	PlayerOrder() []string
	DrawRawCards(amount int) ([]model.Card, bool)
	CheckHandLimit(playerID string, stayInTurn bool)
	PushInterrupt(intr *model.Interrupt)
	Log(msg string)
}

func ActionTargetIDs(rt butterflyRuntime) []string {
	targetIDs := make([]string, 0, len(rt.PlayerOrder()))
	for _, pid := range rt.PlayerOrder() {
		if rt.GetPlayers()[pid] != nil {
			targetIDs = append(targetIDs, pid)
		}
	}
	return targetIDs
}

// QueueWitherChoice queues the wither (凋零) confirmation interrupt.
func QueueWitherChoice(rt engineplayer.ChoiceRuntime, user *model.Player) {
	if user == nil || !engineplayer.IsCharacter(user, "butterfly_dancer") {
		return
	}
	if user.TurnState.SkillFlowState == nil {
		user.TurnState.SkillFlowState = make(map[string]int)
	}
	user.TurnState.SkillFlowState["bt_wither_pending"]++
	if user.TurnState.SkillFlowState["bt_wither_pending"] > 1 {
		return
	}
	rt.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: user.ID,
		Context: map[string]interface{}{
			"choice_type": "bt_wither_confirm",
			"user_id":     user.ID,
			"target_ids":  ActionTargetIDs(rt),
		},
	})
	rt.Log(fmt.Sprintf("%s 可发动 [凋零]：请选择是否发动", user.Name))
}

// ResolveChrysalis handles the chrysalis (蛹化) skill execution.
func ResolveChrysalis(rt butterflyRuntime, userID string) error {
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	now := AddPupa(user, 1)
	cards, ok := rt.DrawRawCards(4)
	if !ok {
		cards = nil
	}
	added := addCocoonCardsDirect(user, cards)
	rt.Log(fmt.Sprintf("%s 发动 [蛹化]：蛹+1（当前%d），获得%d个茧", user.Name, now, added))
	rt.CheckHandLimit(user.ID, true)
	overflow := CocoonCount(user) - CocoonCap
	if overflow > 0 {
		rt.PushInterrupt(&model.Interrupt{
			Type:     model.InterruptChoice,
			PlayerID: user.ID,
			Context:  butterflyCocoonOverflowContext(user.ID, overflow),
		})
	}
	return nil
}

// StartReverse handles the reverse butterfly (倒逆之蝶) skill execution.
func StartReverse(rt butterflyRuntime, userID string) error {
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	rt.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: user.ID,
		Context: map[string]interface{}{
			"choice_type": "bt_reverse_mode",
			"user_id":     user.ID,
			"can_branch2": Pupa(user) > 0,
			"target_ids":  ActionTargetIDs(rt),
		},
	})
	rt.Log(fmt.Sprintf("%s 发动 [倒逆之蝶]：已弃2张牌，请选择发动分支", user.Name))
	return nil
}

// MaybeDamageResponses checks butterfly dancer damage response triggers.
func MaybeDamageResponses(rt engineplayer.ChoiceRuntime, pd *model.PendingDamage) bool {
	if pd == nil || pd.Damage <= 0 {
		return false
	}

	// 预取来源/目标名称，供 prompt message 使用
	sourceName := ""
	if src := rt.GetPlayers()[pd.SourceID]; src != nil {
		sourceName = src.Name
	}
	targetName := ""
	if tgt := rt.GetPlayers()[pd.TargetID]; tgt != nil {
		targetName = tgt.Name
	}
	damageAmount := pd.Damage

	if !pd.HasCheck(model.PendingDamageCheckBeforeApplyDefend) {
		pd.SetCheck(model.PendingDamageCheckBeforeApplyDefend, true)
		target := rt.GetPlayers()[pd.TargetID]
		if target != nil && engineplayer.IsCharacter(target, "butterfly_dancer") && CocoonCount(target) > 0 {
			indices := CocoonFieldIndices(target)
			if len(indices) > 0 {
				rt.PushInterrupt(&model.Interrupt{
					Type:     model.InterruptChoice,
					PlayerID: target.ID,
					Context: map[string]interface{}{
						"choice_type":    "bt_pilgrimage_confirm",
						"user_id":        target.ID,
						"source_id":      pd.SourceID,
						"source_name":    sourceName,
						"target_id":      pd.TargetID,
						"target_name":    targetName,
						"damage_index":   0,
						"damage_amount":  damageAmount,
						"cocoon_indices": indices,
					},
				})
				rt.Log(fmt.Sprintf("%s 的 [朝圣] 可触发：是否移除1个茧抵御1点伤害", target.Name))
				return true
			}
		}
	}
	if pd.DamageType != model.MagicDamage {
		return false
	}
	if pd.HasCheck(model.PendingDamageCheckBeforeApplyResponse) {
		return false
	}
	pd.SetCheck(model.PendingDamageCheckBeforeApplyResponse, true)

	if pd.Damage == 1 {
		for _, pid := range rt.GetPlayerOrder() {
			user := rt.GetPlayers()[pid]
			if user == nil || !engineplayer.IsCharacter(user, "butterfly_dancer") || CocoonCount(user) <= 0 {
				continue
			}
			indices := CocoonFieldIndices(user)
			if len(indices) == 0 {
				continue
			}
			rt.PushInterrupt(&model.Interrupt{
				Type:     model.InterruptChoice,
				PlayerID: user.ID,
				Context: map[string]interface{}{
					"choice_type":    "bt_poison_pick",
					"user_id":        user.ID,
					"source_id":      pd.SourceID,
					"source_name":    sourceName,
					"target_id":      pd.TargetID,
					"target_name":    targetName,
					"damage_index":   0,
					"damage_amount":  damageAmount,
					"cocoon_indices": indices,
				},
			})
			rt.Log(fmt.Sprintf("%s 的 [毒粉] 可触发：是否移除1个茧令该次法术伤害+1", user.Name))
			return true
		}
		return false
	}

	if pd.Damage == 2 {
		for _, pid := range rt.GetPlayerOrder() {
			user := rt.GetPlayers()[pid]
			if user == nil || !engineplayer.IsCharacter(user, "butterfly_dancer") || CocoonCount(user) < 2 {
				continue
			}
			defs, labels := mirrorPairDefs(user)
			if len(defs) == 0 {
				continue
			}
			rt.PushInterrupt(&model.Interrupt{
				Type:     model.InterruptChoice,
				PlayerID: user.ID,
				Context: map[string]interface{}{
					"choice_type":   "bt_mirror_pair",
					"user_id":       user.ID,
					"source_id":     pd.SourceID,
					"source_name":   sourceName,
					"target_id":     pd.TargetID,
					"target_name":   targetName,
					"damage_index":  0,
					"damage_amount": damageAmount,
					"pair_defs":     defs,
					"pair_labels":   labels,
				},
			})
			rt.Log(fmt.Sprintf("%s 的 [镜花水月] 可触发：是否移除2张同系茧改写本次伤害来源", user.Name))
			return true
		}
	}
	return false
}

func addCocoonCardsDirect(player *model.Player, cards []model.Card) int {
	if player == nil || len(cards) == 0 {
		return 0
	}
	AddCocoonCards(player, cards)
	return len(cards)
}

func mirrorPairDefs(player *model.Player) ([]string, []string) {
	if player == nil {
		return nil, nil
	}
	elemToFieldIdx := map[model.Element][]int{}
	for i, fc := range player.Field {
		if fc == nil || fc.Mode != model.FieldCover || fc.Effect != model.EffectButterflyCocoon {
			continue
		}
		elemToFieldIdx[fc.Card.Element] = append(elemToFieldIdx[fc.Card.Element], i)
	}
	elements := []model.Element{
		model.ElementFire, model.ElementWater, model.ElementWind, model.ElementThunder,
		model.ElementEarth, model.ElementLight, model.ElementDark,
	}
	var defs []string
	var labels []string
	for _, ele := range elements {
		idxs := elemToFieldIdx[ele]
		if len(idxs) < 2 {
			continue
		}
		for i := 0; i < len(idxs); i++ {
			for j := i + 1; j < len(idxs); j++ {
				left := idxs[i]
				right := idxs[j]
				defs = append(defs, fmt.Sprintf("%d,%d", left, right))
				lc := player.Field[left].Card
				rc := player.Field[right].Card
				labels = append(labels, fmt.Sprintf("%s系茧：%s + %s", promptfmt.ElementName(string(ele)), promptfmt.FormatCardInfo(lc), promptfmt.FormatCardInfo(rc)))
			}
		}
	}
	return defs, labels
}
