// gameflow: 英灵人形相关技能流。

package engine

import (
	"fmt"

	"starcup-engine/internal/engine/core/runtimeutil"
	"starcup-engine/internal/model"
)

func validateHomRuneCardSelection(user *model.Player, selected []int, attackElement string, glyph bool, mismatchErr, duplicateErr string) error {
	seen := map[model.Element]bool{}
	for _, idx := range selected {
		if idx < 0 || idx >= len(user.Hand) {
			return fmt.Errorf("无效的手牌索引: %d", idx)
		}
		elem := user.Hand[idx].Element
		if glyph {
			if attackElement != "" && string(elem) == attackElement {
				return fmt.Errorf(mismatchErr)
			}
			if duplicateErr != "" && seen[elem] {
				return fmt.Errorf(duplicateErr)
			}
			seen[elem] = true
			continue
		}
		if attackElement != "" && string(elem) != attackElement {
			return fmt.Errorf(mismatchErr)
		}
	}
	return nil
}

func applyHomRuneFlip(user *model.Player, glyph bool, flipCount int) error {
	if user.Tokens == nil {
		user.Tokens = map[string]int{}
	}
	if glyph {
		if user.Tokens["hom_magic_rune"] < flipCount {
			return fmt.Errorf("魔纹不足，至少需要%d个", flipCount)
		}
		user.Tokens["hom_magic_rune"] -= flipCount
		user.Tokens["hom_war_rune"] += flipCount
		return nil
	}
	if user.Tokens["hom_war_rune"] < flipCount {
		return fmt.Errorf("战纹不足，至少需要%d个", flipCount)
	}
	user.Tokens["hom_war_rune"] -= flipCount
	user.Tokens["hom_magic_rune"] += flipCount
	return nil
}

func filterHomRuneRemainingCandidates(user *model.Player, remaining []int, picked int, glyph bool) []int {
	nextRemaining := make([]int, 0, len(remaining))
	for _, idx := range remaining {
		if idx == picked {
			continue
		}
		if glyph && idx >= 0 && idx < len(user.Hand) && picked >= 0 && picked < len(user.Hand) && user.Hand[idx].Element == user.Hand[picked].Element {
			continue
		}
		nextRemaining = append(nextRemaining, idx)
	}
	return nextRemaining
}

func (e *GameEngine) updateHomRuneChoiceContext(ctxData map[string]interface{}) {
	e.State.PendingInterrupt.Context = ctxData
	e.notifyInterruptPrompt()
}

// resolveHomunculusRuneChoice 结算英灵人形"战纹碎击/魔纹融合"的X/Y交互结果。
func (e *GameEngine) resolveHomunculusRuneChoice(ctxData map[string]interface{}, glyph bool) error {
	userID, _ := ctxData["user_id"].(string)
	user := e.State.Players[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	rawCtx, _ := ctxData["user_ctx"].(*model.Context)
	if rawCtx == nil || rawCtx.EventCtx == nil {
		return fmt.Errorf("英灵人形技能上下文丢失")
	}
	xVal := runtimeutil.ToIntContextValue(ctxData["x_value"])
	yVal := runtimeutil.ToIntContextValue(ctxData["y_value"])
	if xVal <= 0 || yVal < 0 {
		return fmt.Errorf("X/Y 参数无效")
	}
	selected := runtimeutil.ParseChoiceIntSlice(ctxData["selected_indices"])
	if len(selected) != xVal {
		return fmt.Errorf("弃牌数量与X不一致")
	}

	attackElement, _ := ctxData["attack_element"].(string)
	mismatchErr := "战纹碎击需弃置同系牌"
	duplicateErr := ""
	if glyph {
		mismatchErr = "魔纹融合需弃置异系牌"
		duplicateErr = "魔纹融合需弃置元素互不相同的异系牌"
	}
	if err := validateHomRuneCardSelection(user, selected, attackElement, glyph, mismatchErr, duplicateErr); err != nil {
		return err
	}

	flipCount := 1 + yVal
	if err := applyHomRuneFlip(user, glyph, flipCount); err != nil {
		return err
	}

	removed, err := removeCardsByIndicesFromHand(user, append([]int{}, selected...))
	if err != nil {
		return err
	}
	e.NotifyCardRevealed(user.ID, removed, "discard")
	e.State.DiscardPile = append(e.State.DiscardPile, removed...)

	targetID := rawCtx.EventCtx.TargetID
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
				DamageType: model.MagicAttack,
			})
		}
		e.Log(fmt.Sprintf("%s 发动 [魔纹融合]：弃%d张异系牌，翻转%d个魔纹为战纹，额外造成%d点法术伤害", user.Name, xVal, flipCount, damage))
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil && rawCtx.ResumeAttackMissPhase() {
			if e.resumePendingAttackMiss(rawCtx) {
				return nil
			}
		}
		if e.State.PendingInterrupt == nil && len(e.State.PendingDamageQueue) > 0 {
			e.enterDamageResolution(nil)
		}
		return nil
	}

	bonusDamage := xVal - 1
	if bonusDamage < 0 {
		bonusDamage = 0
	}
	if rawCtx.EventCtx.DamageVal != nil && bonusDamage > 0 {
		*rawCtx.EventCtx.DamageVal += bonusDamage
	}
	if yVal > 0 && targetID != "" {
		e.AddPendingDamage(model.PendingDamage{
			SourceID:   user.ID,
			TargetID:   targetID,
			Damage:     yVal,
			DamageType: model.MagicAttack,
		})
	}
	e.Log(fmt.Sprintf("%s 发动 [战纹碎击]：弃%d张同系牌，翻转%d个战纹为魔纹，本次攻击伤害+%d", user.Name, xVal, flipCount, bonusDamage))
	e.PopInterrupt()
	e.resumePendingAttackHit(ctxData)
	return nil
}
