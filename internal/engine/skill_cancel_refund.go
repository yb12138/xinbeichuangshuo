// gameflow: 技能"取消退费"——先扣费/计次再推选择提示的技能，在提示里取消时回滚费用与回合限定次数。
//
// 机制：
//  1. 技能发动前（主动技 UseSkill、启动技确认、响应技确认）对发动者做一次快照；
//  2. 技能执行期间推入的第一个中断与快照绑定（只在执行本身没有产生可观察效果时才绑定）；
//  3. 玩家对该中断"取消"后，若中断已被弹出（流程放弃），且期间局面没有可观察变化，则按快照回滚：
//     宝石/水晶、阵营星石、回合限定计数、技能中间态，以及主动技写入的行动收尾标记。
//
// Execute 直接完成结算的技能不会推中断，也就不会产生退费记录，计数与扣费照常保留。

package engine

import (
	"fmt"

	"starcup-engine/internal/model"
)

type skillRefundSnapshot struct {
	playerID           string
	skillID            string
	skillTitle         string
	gem                int
	crystal            int
	redGems            int
	redCrystals        int
	blueGems           int
	blueCrystals       int
	usedSkillCounts    map[string]int
	skillFlowState     map[string]int
	hasActed           bool
	hasUsedActionSkill bool
	lastActionType     string
	lastActionCard     *model.Card
	// restoreActionWindow 主动技取消后需要回到行动选择。
	restoreActionWindow bool
	// stableFingerprint 不含费用的局面指纹：用于判断 Execute 本身是否已经产生了效果。
	stableFingerprint string
}

type skillRefundCapture struct {
	snapshot *skillRefundSnapshot
	pushed   *model.Interrupt
}

type skillRefundEntry struct {
	snapshot *skillRefundSnapshot
	// fingerprint 绑定时刻的完整局面指纹（含费用）：取消时必须一致才退费。
	fingerprint string
}

func copyIntMap(src map[string]int) map[string]int {
	out := make(map[string]int, len(src))
	for k, v := range src {
		out[k] = v
	}
	return out
}

// skillRefundFingerprint 生成局面指纹。includeCost=false 时忽略发动者个人能量与阵营星石（技能费用）。
func (e *GameEngine) skillRefundFingerprint(userID string, includeCost bool) string {
	if e == nil || e.State == nil {
		return ""
	}
	s := e.State
	fp := fmt.Sprintf("dp=%d|pdq=%d|mo=%d/%d|cups=%d/%d|aq=%d",
		len(s.DiscardPile), len(s.PendingDamageQueue), s.RedMorale, s.BlueMorale, s.RedCups, s.BlueCups, len(s.ActionQueue))
	if includeCost {
		fp += fmt.Sprintf("|camp=%d/%d/%d/%d", s.RedGems, s.RedCrystals, s.BlueGems, s.BlueCrystals)
	}
	for _, pid := range s.PlayerOrder {
		p := s.Players[pid]
		if p == nil {
			continue
		}
		fp += fmt.Sprintf("|%s:h=%d,heal=%d,f=%d,ex=%d,form=%s", pid, len(p.Hand), p.Heal, len(p.Field), len(p.ExclusiveCards), p.Form)
		if includeCost || pid != userID {
			fp += fmt.Sprintf(",g=%d,c=%d", p.Gem, p.Crystal)
		}
	}
	return fp
}

// beginSkillRefundCapture 在技能付费/计次前调用，记录可回滚的快照并开始捕获该技能推入的中断。
// 返回值需交给 commitSkillRefundCapture；嵌套捕获时返回 nil（只跟踪最外层技能）。
func (e *GameEngine) beginSkillRefundCapture(player *model.Player, skillDef *model.SkillDefinition, restoreActionWindow bool) *skillRefundCapture {
	if e == nil || e.State == nil || player == nil || skillDef == nil || e.skillRefundCapture != nil {
		return nil
	}
	snap := &skillRefundSnapshot{
		playerID:            player.ID,
		skillID:             skillDef.ID,
		skillTitle:          skillDef.Title,
		gem:                 player.Gem,
		crystal:             player.Crystal,
		redGems:             e.State.RedGems,
		redCrystals:         e.State.RedCrystals,
		blueGems:            e.State.BlueGems,
		blueCrystals:        e.State.BlueCrystals,
		usedSkillCounts:     copyIntMap(player.TurnState.UsedSkillCounts),
		skillFlowState:      copyIntMap(player.TurnState.SkillFlowState),
		hasActed:            player.TurnState.HasActed,
		hasUsedActionSkill:  player.TurnState.HasUsedActionSkill,
		lastActionType:      player.TurnState.LastActionType,
		restoreActionWindow: restoreActionWindow,
		stableFingerprint:   e.skillRefundFingerprint(player.ID, false),
	}
	if player.TurnState.LastActionCard != nil {
		card := *player.TurnState.LastActionCard
		snap.lastActionCard = &card
	}
	capture := &skillRefundCapture{snapshot: snap}
	e.skillRefundCapture = capture
	return capture
}

// commitSkillRefundCapture 结束捕获；若技能推入了中断且执行本身没有产生效果，则登记退费记录。
func (e *GameEngine) commitSkillRefundCapture(capture *skillRefundCapture) {
	if e == nil || capture == nil {
		return
	}
	if e.skillRefundCapture == capture {
		e.skillRefundCapture = nil
	}
	if capture.pushed == nil || !e.interruptIsLive(capture.pushed) {
		return
	}
	snap := capture.snapshot
	if e.skillRefundFingerprint(snap.playerID, false) != snap.stableFingerprint {
		return
	}
	if e.skillRefunds == nil {
		e.skillRefunds = map[*model.Interrupt]*skillRefundEntry{}
	}
	e.skillRefunds[capture.pushed] = &skillRefundEntry{
		snapshot:    snap,
		fingerprint: e.skillRefundFingerprint(snap.playerID, true),
	}
}

// abortSkillRefundCapture 技能执行失败时放弃捕获。
func (e *GameEngine) abortSkillRefundCapture(capture *skillRefundCapture) {
	if e != nil && capture != nil && e.skillRefundCapture == capture {
		e.skillRefundCapture = nil
	}
}

// noteInterruptPushedForRefund 由 PushInterrupt 调用：记录捕获期间推入的第一个中断。
func (e *GameEngine) noteInterruptPushedForRefund(intr *model.Interrupt) {
	if e == nil || e.skillRefundCapture == nil || e.skillRefundCapture.pushed != nil || intr == nil {
		return
	}
	e.skillRefundCapture.pushed = intr
}

func (e *GameEngine) interruptIsLive(intr *model.Interrupt) bool {
	if e == nil || e.State == nil || intr == nil {
		return false
	}
	if e.State.PendingInterrupt == intr {
		return true
	}
	for _, queued := range e.State.InterruptQueue {
		if queued == intr {
			return true
		}
	}
	return false
}

// pruneSkillRefunds 清理已离场中断的退费记录（中断被正常结算后就不再可退）。
func (e *GameEngine) pruneSkillRefunds() {
	for intr := range e.skillRefunds {
		if !e.interruptIsLive(intr) {
			delete(e.skillRefunds, intr)
		}
	}
}

// handleInterruptActionWithRefund 包装中断输入：取消导致流程放弃时执行退费。
func (e *GameEngine) handleInterruptActionWithRefund(act model.PlayerAction, dispatch func() error) error {
	var entry *skillRefundEntry
	var target *model.Interrupt
	if act.Type == model.CmdCancel && e.State != nil && e.State.PendingInterrupt != nil {
		target = e.State.PendingInterrupt
		entry = e.skillRefunds[target]
	}
	err := dispatch()
	if err == nil && entry != nil && !e.interruptIsLive(target) &&
		e.skillRefundFingerprint(entry.snapshot.playerID, true) == entry.fingerprint {
		delete(e.skillRefunds, target)
		e.applySkillRefund(entry.snapshot)
	}
	e.pruneSkillRefunds()
	return err
}

func (e *GameEngine) applySkillRefund(snap *skillRefundSnapshot) {
	player := e.State.Players[snap.playerID]
	if player == nil {
		return
	}
	refundGem := snap.gem - player.Gem
	refundCrystal := snap.crystal - player.Crystal
	player.Gem = snap.gem
	player.Crystal = snap.crystal
	e.State.RedGems = snap.redGems
	e.State.RedCrystals = snap.redCrystals
	e.State.BlueGems = snap.blueGems
	e.State.BlueCrystals = snap.blueCrystals
	player.TurnState.UsedSkillCounts = copyIntMap(snap.usedSkillCounts)
	player.TurnState.SkillFlowState = copyIntMap(snap.skillFlowState)
	player.TurnState.HasUsedActionSkill = snap.hasUsedActionSkill
	if snap.restoreActionWindow {
		player.TurnState.HasActed = snap.hasActed
		player.TurnState.LastActionType = snap.lastActionType
		player.TurnState.LastActionCard = snap.lastActionCard
	}
	e.recordActionResourceDelta()
	e.Log(fmt.Sprintf("[Skill] %s 取消了 [%s]，退回费用（宝石%d 水晶%d）与本回合发动次数", player.Name, snap.skillTitle, refundGem, refundCrystal))
	if snap.restoreActionWindow && e.State.PendingInterrupt == nil && len(e.State.PendingDamageQueue) == 0 {
		e.enterActionExecutionStage()
	}
}

// findSkillDefForRefund 查找玩家身上的技能定义（吟游诗人给持有者的响应技能找不到时返回 nil，不做退费）。
func (e *GameEngine) findSkillDefForRefund(playerID, skillID string) *model.SkillDefinition {
	p := e.State.Players[playerID]
	if p == nil || p.Character == nil {
		return nil
	}
	for i := range p.Character.Skills {
		if p.Character.Skills[i].ID == skillID {
			return &p.Character.Skills[i]
		}
	}
	return nil
}
