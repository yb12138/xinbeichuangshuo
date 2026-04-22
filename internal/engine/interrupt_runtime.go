// gameflow: 中断 Push/Pop、类型分派、与 Prompt 绑定。

package engine

import (
	"fmt"
	"sort"

	"starcup-engine/internal/engine/core/runtimeutil"
	playerpkg "starcup-engine/internal/engine/player"
	adventurer "starcup-engine/internal/engine/player/adventurer"
	"starcup-engine/internal/model"
)

// ConfirmStartupSkill 确认发动启动技能。
func (e *GameEngine) ConfirmStartupSkill(playerID string, skillID string) error {
	return e.dispatcher.ConfirmStartupSkill(playerID, skillID)
}

// SkipStartupSkill 跳过启动技能。
func (e *GameEngine) SkipStartupSkill(playerID string) error {
	return e.dispatcher.SkipStartupSkill(playerID)
}

// ConfirmResponseSkill 确认发动响应技能。
func (e *GameEngine) ConfirmResponseSkill(playerID string, skillID string) error {
	return e.dispatcher.ConfirmResponseSkill(playerID, skillID)
}

func (e *GameEngine) handleInterruptResponseSkillAction(act model.PlayerAction) error {
	if e.prunePendingResponseSkills() {
		if p := e.State.Players[act.PlayerID]; p != nil && p.TurnState.SkillFlowState != nil && p.TurnState.SkillFlowState["adventurer_extract_requires_paradise"] > 0 {
			return fmt.Errorf("本次提炼结果需先发动[冒险者天堂]分配给队友")
		}
		e.clearAdventurerExtractState(e.State.Players[act.PlayerID])
		return e.SkipResponse()
	}
	forceParadise := adventurer.IsForcedParadiseResponse(e.State.PendingInterrupt, e.State.Players, act.PlayerID)
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
		return e.ConfirmResponseSkill(act.PlayerID, e.State.PendingInterrupt.SkillIDs[idx])
	}
	return fmt.Errorf("当前中断类型不支持该指令")
}

func (e *GameEngine) handleInterruptStartupSkillAction(act model.PlayerAction) error {
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
		return e.ConfirmStartupSkill(act.PlayerID, e.State.PendingInterrupt.SkillIDs[idx])
	}
	return fmt.Errorf("当前中断类型不支持该指令")
}

func (e *GameEngine) handleInterruptGiveCardsAction(act model.PlayerAction) error {
	if act.Type != model.CmdSelect {
		return fmt.Errorf("当前中断类型不支持该指令")
	}
	data, ok := e.State.PendingInterrupt.Context.(map[string]interface{})
	if !ok {
		return fmt.Errorf("给牌中断上下文错误")
	}
	receiverID, _ := data["receiver_id"].(string)
	return e.ConfirmGiveCards(act.PlayerID, receiverID, act.Selections)
}

// GetCurrentPrompt 获取当前用户交互提示。
func (e *GameEngine) GetCurrentPrompt() *model.Prompt {
	var prompt *model.Prompt
	if e.State.PendingInterrupt != nil {
		if e.State.PendingInterrupt.Type == model.InterruptResponseSkill && e.prunePendingResponseSkills() {
			_ = e.SkipResponse()
			return nil
		}
		prompt = e.buildPendingInterruptPrompt()
	}
	if prompt != nil {
		return e.decoratePromptForClient(prompt)
	}
	if prompt = e.buildStandardResponsePrompt(); prompt != nil {
		return e.decoratePromptForClient(prompt)
	}

	return nil
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

// ConfirmGiveCards 确认选牌交给他人。
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

	sort.Sort(sort.Reverse(sort.IntSlice(indices)))
	var givenCards []model.Card
	for _, idx := range indices {
		givenCards = append(givenCards, giver.Hand[idx])
		giver.Hand = append(giver.Hand[:idx], giver.Hand[idx+1:]...)
	}

	receiver.Hand = append(receiver.Hand, givenCards...)
	e.Log(fmt.Sprintf("[Skill] %s 将 %d 张牌交给了 %s", giver.Name, len(givenCards), receiver.Name))
	overflowCtx := e.buildContext(receiver, nil, model.TimingActive, nil)
	if runtimeutil.ToBoolContextValue(data["stay_in_turn"]) {
		overflowCtx.Flags["StayInTurn"] = true
	}
	if point, ok := choiceResumePointValue(data["resume_phase"]); ok {
		overflowCtx.Flags["StayInTurn"] = true
		overflowCtx.Selections["draw_resume_phase"] = point
	}
	e.checkHandLimit(receiver, overflowCtx)
	e.Log(fmt.Sprintf("[Debug] 给牌完成，队列中还有 %d 个中断", len(e.State.InterruptQueue)))

	e.PopInterrupt()
	if e.State.PendingInterrupt != nil {
		e.Log(fmt.Sprintf("[Debug] 新的中断已设置: Type=%s, PlayerID=%s", e.State.PendingInterrupt.Type, e.State.PendingInterrupt.PlayerID))
	} else {
		e.Log("[Debug] 所有给牌中断已处理完毕")
	}

	return nil
}

// SkipResponse 跳过响应阶段。
func (e *GameEngine) SkipResponse() error {
	if !isBeforeDrawResponseInterrupt(e.State.PendingInterrupt) && e.maybeAdvanceResponseSkillSelection() {
		return nil
	}
	state := e.captureResponseResumeStateFromInterrupt(responseCompletionSkip, "", e.State.PendingInterrupt)
	e.PopInterrupt()
	e.runTimingOnResponseSkipEffects(&state)
	e.restoreSkippedResponseAfterPop(state)
	return nil
}

func isBeforeDrawResponseInterrupt(intr *model.Interrupt) bool {
	if intr == nil || intr.Type != model.InterruptResponseSkill {
		return false
	}
	switch raw := intr.Context.(type) {
	case *model.Context:
		return raw.BeforeDrawPhase()
	case map[string]interface{}:
		if ctx, ok := raw["user_ctx"].(*model.Context); ok {
			return ctx.BeforeDrawPhase()
		}
	}
	return false
}

// maybeAdvanceResponseSkillSelection 在跳过当前响应技能时，根据 policy 推进到下一批可展示技能。
func (e *GameEngine) maybeAdvanceResponseSkillSelection() bool {
	intr := e.State.PendingInterrupt
	if intr == nil || intr.Type != model.InterruptResponseSkill || intr.PlayerID == "" {
		return false
	}
	player := e.State.Players[intr.PlayerID]
	if player == nil || e.dispatcher == nil {
		return false
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
		return false
	}
	// 摸牌前响应（如水影）跳过后应直接恢复摸牌，不在同一窗口反复询问。
	if ctx.BeforeDrawPhase() {
		return false
	}

	if len(intr.SkillIDs) == 1 && intr.SkillIDs[0] == "fighter_charge_strike" &&
		ctx.Timing == model.TimingOnAttackDeclared && ctx.EventCtx != nil && ctx.EventCtx.AttackInfo != nil &&
		ctx.EventCtx.AttackInfo.CounterInitiator == "" && playerpkg.IsCharacter(player, "fighter") &&
		e.dispatcher.isSkillStillUsable("fighter_burst_crash", player, ctx) {
		intr.SkillIDs = []string{"fighter_burst_crash"}
		e.Log(fmt.Sprintf("%s 放弃 [蓄力一击]，继续询问是否发动 [气绝崩击]", player.Name))
		e.notifyInterruptPrompt()
		return true
	}

	nextSkillIDs := e.dispatcher.getOtherUsableSkills("", player, ctx)
	nextSkillIDs = e.dispatcher.applyTimingOnHitCheckResponseSkillNormalize(nextSkillIDs, ctx)
	if len(nextSkillIDs) == 0 {
		return false
	}
	if len(nextSkillIDs) == len(intr.SkillIDs) {
		same := true
		for i := range nextSkillIDs {
			if intr.SkillIDs[i] != nextSkillIDs[i] {
				same = false
				break
			}
		}
		if same {
			return false
		}
	}

	intr.SkillIDs = nextSkillIDs
	e.notifyInterruptPrompt()
	return true
}
