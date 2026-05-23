// gameflow: 行动周期内资源变动汇总，供日志或技能读取。

package engine

import (
	"fmt"
	"sort"
	"strings"

	"starcup-engine/internal/model"
)

type actionSummary struct {
	active     bool
	actionType string // attack | magic | skill | special | cannot_act
	actorID    string
	actionName string
	targets    map[string]bool

	responses []string
	skills    []string

	draws    map[string]int
	discards map[string]int
	damages  map[string]int
	heals    map[string]int

	notes []string
}

func (e *GameEngine) BeginActionSummary(actionType, actorID, actionName string, targets []string) {
	if e == nil {
		return
	}
	if e.actionSummary != nil && e.actionSummary.active {
		e.FinalizeActionSummaryIfIdle()
		if e.actionSummary != nil && e.actionSummary.active {
			e.clearActionSummary()
		}
	}
	if e.actionSummaryTurn <= 0 {
		e.actionSummaryTurn = 1
	}
	sum := &actionSummary{
		active:     true,
		actionType: actionType,
		actorID:    actorID,
		actionName: actionName,
		targets:    map[string]bool{},
		responses:  []string{},
		skills:     []string{},
		draws:      map[string]int{},
		discards:   map[string]int{},
		damages:    map[string]int{},
		heals:      map[string]int{},
		notes:      []string{},
	}
	for _, tid := range targets {
		if tid == "" {
			continue
		}
		sum.targets[tid] = true
	}
	e.actionSummary = sum
	if actionType == "special" {
		e.addActionSummaryNote(fmt.Sprintf("%s 执行特殊行动【%s】", e.playerName(actorID), actionName))
	}
	if actionType == "cannot_act" {
		e.addActionSummaryNote(fmt.Sprintf("%s 宣告无法行动", e.playerName(actorID)))
	}
}

func (e *GameEngine) clearActionSummary() {
	if e == nil {
		return
	}
	e.actionSummary = nil
}

func (e *GameEngine) addActionResponse(text string) {
	if e.actionSummary == nil || !e.actionSummary.active || text == "" {
		return
	}
	for _, existing := range e.actionSummary.responses {
		if existing == text {
			return
		}
	}
	e.actionSummary.responses = append(e.actionSummary.responses, text)
}

func (e *GameEngine) addActionSkill(text string) {
	if e.actionSummary == nil || !e.actionSummary.active || text == "" {
		return
	}
	for _, existing := range e.actionSummary.skills {
		if existing == text {
			return
		}
	}
	e.actionSummary.skills = append(e.actionSummary.skills, text)
}

func (e *GameEngine) addActionSummaryNote(text string) {
	if e.actionSummary == nil || !e.actionSummary.active || text == "" {
		return
	}
	for _, existing := range e.actionSummary.notes {
		if existing == text {
			return
		}
	}
	e.actionSummary.notes = append(e.actionSummary.notes, text)
}

func (e *GameEngine) addActionDraw(playerID string, count int) {
	if e.actionSummary == nil || !e.actionSummary.active || count <= 0 || playerID == "" {
		return
	}
	e.actionSummary.draws[playerID] += count
}

func (e *GameEngine) addActionDiscard(playerID string, count int) {
	if e.actionSummary == nil || !e.actionSummary.active || count <= 0 || playerID == "" {
		return
	}
	e.actionSummary.discards[playerID] += count
}

func (e *GameEngine) addActionDamage(playerID string, amount int) {
	if e.actionSummary == nil || !e.actionSummary.active || amount <= 0 || playerID == "" {
		return
	}
	e.actionSummary.damages[playerID] += amount
}

func (e *GameEngine) addActionHeal(playerID string, amount int) {
	if e.actionSummary == nil || !e.actionSummary.active || amount <= 0 || playerID == "" {
		return
	}
	e.actionSummary.heals[playerID] += amount
}

func (e *GameEngine) recordSkillUsage(playerID, title string, skillType model.SkillType) {
	if e.actionSummary == nil || !e.actionSummary.active || playerID == "" || title == "" {
		return
	}
	userName := e.playerName(playerID)
	if skillType == model.SkillTypeResponse {
		e.addActionResponse(fmt.Sprintf("%s 响应技能【%s】", userName, title))
		return
	}
	if skillType == model.SkillTypeAction {
		if e.actionSummary.actionType == "skill" && e.actionSummary.actorID == playerID && e.actionSummary.actionName == title {
			return
		}
		e.addActionSkill(fmt.Sprintf("%s 发动技能【%s】", userName, title))
	}
}

func (e *GameEngine) playerName(playerID string) string {
	if e == nil || e.State == nil {
		return playerID
	}
	if p := e.State.Players[playerID]; p != nil {
		return p.Name
	}
	return playerID
}

func (e *GameEngine) actionSummaryMessage() string {
	if e.actionSummary == nil || !e.actionSummary.active {
		return ""
	}
	sum := e.actionSummary
	parts := make([]string, 0, len(sum.notes)+len(sum.responses)+len(sum.skills)+8)
	if title := e.actionSummaryTitle(sum); title != "" {
		parts = append(parts, title)
	}
	known := map[string]bool{}
	for _, entry := range sum.responses {
		known[normalize(entry)] = true
	}
	for _, entry := range sum.skills {
		known[normalize(entry)] = true
	}
	for _, note := range sum.notes {
		if known[normalize(note)] {
			continue
		}
		parts = append(parts, note)
	}
	for _, entry := range sum.responses {
		parts = append(parts, entry)
	}
	for _, entry := range sum.skills {
		parts = append(parts, entry)
	}
	parts = e.appendActionCounterParts(parts, sum.draws, "摸%d张牌")
	parts = e.appendActionCounterParts(parts, sum.discards, "弃%d张牌")
	parts = e.appendActionCounterParts(parts, sum.damages, "受到%d点伤害")
	parts = e.appendActionCounterParts(parts, sum.heals, "获得%d点治疗")
	if len(parts) == 0 {
		return ""
	}
	turn := e.actionSummaryTurn
	if turn <= 0 {
		turn = 1
	}
	return fmt.Sprintf("回合%d：%s", turn, strings.Join(parts, "；"))
}

func normalize(text string) string {
	replacer := strings.NewReplacer(" ", "", "，", "", "。", "", "；", "", ":", "", "：", "", "【", "", "】", "")
	return replacer.Replace(text)
}

func (e *GameEngine) actionSummaryTitle(sum *actionSummary) string {
	if sum == nil {
		return ""
	}
	actorName := e.playerName(sum.actorID)
	actionName := sum.actionName
	targetText := e.actionSummaryTargets(sum)
	withTarget := func(base string) string {
		if targetText == "" {
			return base
		}
		return fmt.Sprintf("%s -> %s", base, targetText)
	}
	switch sum.actionType {
	case "attack":
		return withTarget(fmt.Sprintf("%s 使用攻击【%s】", actorName, actionName))
	case "magic":
		return withTarget(fmt.Sprintf("%s 使用法术【%s】", actorName, actionName))
	case "skill":
		return withTarget(fmt.Sprintf("%s 发动技能【%s】", actorName, actionName))
	case "special", "cannot_act":
		return ""
	default:
		if actionName == "" {
			return actorName
		}
		return withTarget(fmt.Sprintf("%s 执行【%s】", actorName, actionName))
	}
}

func (e *GameEngine) actionSummaryTargets(sum *actionSummary) string {
	if e == nil || sum == nil || len(sum.targets) == 0 {
		return ""
	}
	ids := make([]string, 0, len(sum.targets))
	if e.State != nil {
		for _, id := range e.State.PlayerOrder {
			if sum.targets[id] {
				ids = append(ids, id)
			}
		}
	}
	for id := range sum.targets {
		found := false
		for _, existing := range ids {
			if existing == id {
				found = true
				break
			}
		}
		if !found {
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 {
		return ""
	}
	names := make([]string, 0, len(ids))
	for _, id := range ids {
		names = append(names, e.playerName(id))
	}
	return strings.Join(names, " / ")
}

func (e *GameEngine) appendActionCounterParts(parts []string, counters map[string]int, format string) []string {
	if len(counters) == 0 {
		return parts
	}
	ids := e.orderedActionCounterIDs(counters)
	for _, id := range ids {
		count := counters[id]
		if count <= 0 {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s "+format, e.playerName(id), count))
	}
	return parts
}

func (e *GameEngine) orderedActionCounterIDs(counters map[string]int) []string {
	seen := map[string]bool{}
	ids := make([]string, 0, len(counters))
	if e != nil && e.State != nil {
		for _, id := range e.State.PlayerOrder {
			if counters[id] > 0 {
				ids = append(ids, id)
				seen[id] = true
			}
		}
	}
	tail := make([]string, 0, len(counters))
	for id, count := range counters {
		if count > 0 && !seen[id] {
			tail = append(tail, id)
		}
	}
	sort.Strings(tail)
	ids = append(ids, tail...)
	return ids
}

// 判断当前行动是否结束
func (e *GameEngine) isActionFinalizeIdle() bool {
	if e == nil || e.State == nil {
		return false
	}
	if e.State.PendingInterrupt != nil {
		return false
	}
	if len(e.State.PendingDamageQueue) > 0 {
		return false
	}
	if len(e.State.CombatStack) > 0 {
		return false
	}
	if e.State.Subflow != model.SubflowNone || e.State.CombatStage != model.CombatStageNone {
		return false
	}
	switch e.State.TurnStage {
	case model.TurnStageActionEnd, model.TurnStageExtraAction, model.TurnStageTurnEnd:
		return true
	default:
		return false
	}
}

func (e *GameEngine) FinalizeActionSummaryIfIdle() {
	if e == nil || e.actionSummary == nil || !e.actionSummary.active {
		return
	}
	if !e.isActionFinalizeIdle() {
		return
	}
	msg := e.actionSummaryMessage()
	if msg != "" {
		e.NotifyActionSummary(msg)
	}
	e.clearActionSummary()
}
