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

	draws        map[string]int
	discards     map[string]int
	damages      map[string]int
	heals        map[string]int
	healUses     map[string]int
	moraleLosses map[model.Camp]int

	playerGemGains     map[string]int
	playerGemCosts     map[string]int
	playerCrystalGains map[string]int
	playerCrystalCosts map[string]int
	campGemGains       map[model.Camp]int
	campGemCosts       map[model.Camp]int
	campCrystalGains   map[model.Camp]int
	campCrystalCosts   map[model.Camp]int
	tokenGains         map[actionTokenKey]int
	tokenCosts         map[actionTokenKey]int
	resourceCursor     *actionResourceSnapshot

	notes []string
}

type actionTokenKey struct {
	playerID string
	token    string
}

type actionResourceSnapshot struct {
	playerGems     map[string]int
	playerCrystals map[string]int
	playerTokens   map[string]map[string]int
	campGems       map[model.Camp]int
	campCrystals   map[model.Camp]int
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
		active:             true,
		actionType:         actionType,
		actorID:            actorID,
		actionName:         actionName,
		targets:            map[string]bool{},
		responses:          []string{},
		skills:             []string{},
		draws:              map[string]int{},
		discards:           map[string]int{},
		damages:            map[string]int{},
		heals:              map[string]int{},
		healUses:           map[string]int{},
		moraleLosses:       map[model.Camp]int{},
		playerGemGains:     map[string]int{},
		playerGemCosts:     map[string]int{},
		playerCrystalGains: map[string]int{},
		playerCrystalCosts: map[string]int{},
		campGemGains:       map[model.Camp]int{},
		campGemCosts:       map[model.Camp]int{},
		campCrystalGains:   map[model.Camp]int{},
		campCrystalCosts:   map[model.Camp]int{},
		tokenGains:         map[actionTokenKey]int{},
		tokenCosts:         map[actionTokenKey]int{},
		notes:              []string{},
	}
	sum.resourceCursor = e.captureActionResourceSnapshot()
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

func (e *GameEngine) addActionHealUse(playerID string, amount int) {
	if e.actionSummary == nil || !e.actionSummary.active || amount <= 0 || playerID == "" {
		return
	}
	e.actionSummary.healUses[playerID] += amount
}

func (e *GameEngine) addActionMoraleLoss(camp model.Camp, amount int) {
	if e.actionSummary == nil || !e.actionSummary.active || amount <= 0 || camp == "" {
		return
	}
	e.actionSummary.moraleLosses[camp] += amount
}

func (e *GameEngine) recordActionResourceDelta() {
	if e == nil || e.actionSummary == nil || !e.actionSummary.active {
		return
	}
	current := e.captureActionResourceSnapshot()
	previous := e.actionSummary.resourceCursor
	if previous == nil {
		e.actionSummary.resourceCursor = current
		return
	}
	sum := e.actionSummary
	e.recordPlayerResourceDelta(previous.playerGems, current.playerGems, sum.playerGemGains, sum.playerGemCosts)
	e.recordPlayerResourceDelta(previous.playerCrystals, current.playerCrystals, sum.playerCrystalGains, sum.playerCrystalCosts)
	e.recordCampResourceDelta(previous.campGems, current.campGems, sum.campGemGains, sum.campGemCosts)
	e.recordCampResourceDelta(previous.campCrystals, current.campCrystals, sum.campCrystalGains, sum.campCrystalCosts)
	e.recordTokenResourceDelta(previous.playerTokens, current.playerTokens, sum.tokenGains, sum.tokenCosts)
	sum.resourceCursor = current
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
	e.recordActionResourceDelta()
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
	parts = e.appendActionResourceParts(parts, sum)
	parts = e.appendActionCounterParts(parts, sum.draws, "摸%d张牌")
	parts = e.appendActionCounterParts(parts, sum.discards, "弃%d张牌")
	parts = e.appendActionMoraleLossParts(parts, sum.moraleLosses)
	parts = e.appendActionCounterParts(parts, sum.healUses, "使用%d点治疗抵挡伤害")
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

func (e *GameEngine) appendActionMoraleLossParts(parts []string, losses map[model.Camp]int) []string {
	for _, camp := range []model.Camp{model.RedCamp, model.BlueCamp} {
		amount := losses[camp]
		if amount <= 0 {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s士气-%d", campDisplayName(camp), amount))
	}
	tail := make([]string, 0, len(losses))
	for camp, amount := range losses {
		if amount <= 0 || camp == model.RedCamp || camp == model.BlueCamp {
			continue
		}
		tail = append(tail, string(camp))
	}
	sort.Strings(tail)
	for _, camp := range tail {
		amount := losses[model.Camp(camp)]
		parts = append(parts, fmt.Sprintf("%s士气-%d", campDisplayName(model.Camp(camp)), amount))
	}
	return parts
}

func campDisplayName(camp model.Camp) string {
	switch camp {
	case model.RedCamp:
		return "红方"
	case model.BlueCamp:
		return "蓝方"
	default:
		return string(camp)
	}
}

func (e *GameEngine) appendActionResourceParts(parts []string, sum *actionSummary) []string {
	if sum == nil {
		return parts
	}
	parts = e.appendActionPlayerResourceParts(parts, sum.playerGemCosts, "消耗%d红宝石")
	parts = e.appendActionPlayerResourceParts(parts, sum.playerCrystalCosts, "消耗%d蓝水晶")
	parts = e.appendActionTokenParts(parts, sum.tokenCosts, "消耗%d个%s指示物")
	parts = appendActionCampResourceParts(parts, sum.campGemCosts, "战绩区-%d红宝石")
	parts = appendActionCampResourceParts(parts, sum.campCrystalCosts, "战绩区-%d蓝水晶")
	parts = e.appendActionPlayerResourceParts(parts, sum.playerGemGains, "获得%d红宝石")
	parts = e.appendActionPlayerResourceParts(parts, sum.playerCrystalGains, "获得%d蓝水晶")
	parts = e.appendActionTokenParts(parts, sum.tokenGains, "获得%d个%s指示物")
	parts = appendActionCampResourceParts(parts, sum.campGemGains, "战绩区+%d红宝石")
	parts = appendActionCampResourceParts(parts, sum.campCrystalGains, "战绩区+%d蓝水晶")
	return parts
}

func (e *GameEngine) appendActionPlayerResourceParts(parts []string, counters map[string]int, format string) []string {
	return e.appendActionCounterParts(parts, counters, format)
}

func appendActionCampResourceParts(parts []string, counters map[model.Camp]int, format string) []string {
	if len(counters) == 0 {
		return parts
	}
	for _, camp := range []model.Camp{model.RedCamp, model.BlueCamp} {
		amount := counters[camp]
		if amount <= 0 {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s"+format, campDisplayName(camp), amount))
	}
	tail := make([]string, 0, len(counters))
	for camp, amount := range counters {
		if amount <= 0 || camp == model.RedCamp || camp == model.BlueCamp {
			continue
		}
		tail = append(tail, string(camp))
	}
	sort.Strings(tail)
	for _, camp := range tail {
		amount := counters[model.Camp(camp)]
		parts = append(parts, fmt.Sprintf("%s"+format, campDisplayName(model.Camp(camp)), amount))
	}
	return parts
}

func (e *GameEngine) appendActionTokenParts(parts []string, counters map[actionTokenKey]int, format string) []string {
	if len(counters) == 0 {
		return parts
	}
	keys := e.orderedActionTokenKeys(counters)
	for _, key := range keys {
		count := counters[key]
		if count <= 0 {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s "+format, e.playerName(key.playerID), count, tokenDisplayName(key.token)))
	}
	return parts
}

func (e *GameEngine) orderedActionTokenKeys(counters map[actionTokenKey]int) []actionTokenKey {
	seen := map[actionTokenKey]bool{}
	keys := make([]actionTokenKey, 0, len(counters))
	if e != nil && e.State != nil {
		for _, id := range e.State.PlayerOrder {
			playerKeys := make([]string, 0)
			for key, count := range counters {
				if key.playerID == id && count > 0 {
					playerKeys = append(playerKeys, key.token)
				}
			}
			sort.Strings(playerKeys)
			for _, token := range playerKeys {
				key := actionTokenKey{playerID: id, token: token}
				keys = append(keys, key)
				seen[key] = true
			}
		}
	}
	tail := make([]actionTokenKey, 0, len(counters))
	for key, count := range counters {
		if count > 0 && !seen[key] {
			tail = append(tail, key)
		}
	}
	sort.Slice(tail, func(i, j int) bool {
		if tail[i].playerID == tail[j].playerID {
			return tail[i].token < tail[j].token
		}
		return tail[i].playerID < tail[j].playerID
	})
	keys = append(keys, tail...)
	return keys
}

func tokenDisplayName(token string) string {
	if token == "" {
		return "未知"
	}
	if name, ok := actionTokenDisplayNames[token]; ok {
		return name
	}
	return token
}

var actionTokenDisplayNames = map[string]string{
	"arbiter_forced_doomsday":         "强制末日",
	"bd_inspiration":                  "灵感",
	"bs_beast_soul":                   "兽魂",
	"bs_reversal_pending_x":           "逆反待定",
	"bs_zanshin":                      "残心",
	"bt_pupa":                         "蝶蛹",
	"bw_flame_release_pending":        "炎爆待发",
	"bw_mana_inversion_lock":          "魔能反转锁定",
	"bw_pain_link_pending_discard":    "痛楚联结弃牌",
	"bw_pain_link_pending_hits":       "痛楚联结伤害",
	"bw_rebirth":                      "重燃",
	"bw_substitute_lock":              "替身锁定",
	"crk_blood_mark":                  "血印",
	"css_blood":                       "血",
	"css_blood_cap":                   "血上限",
	"element":                         "元素",
	"fighter_qi":                      "斗气",
	"hb_cannon":                       "辉光炮",
	"hb_faith":                        "信仰",
	"hero_anger":                      "怒气",
	"hero_exhaustion_release_pending": "疲劳解除",
	"hero_wisdom":                     "智慧",
	"hom_magic_rune":                  "魔纹",
	"hom_war_rune":                    "战纹",
	"judgment":                        "审判",
	"mg_new_moon":                     "新月",
	"mg_petrify":                      "石化",
	"onmyoji_ghost_fire":              "鬼火",
	"prayer_rune":                     "祈愿符文",
	"sc_power_count":                  "蓄力",
	"se_sword_qi":                     "剑气",
	"ss_blue_soul":                    "蓝魂",
	"ss_yellow_soul":                  "黄魂",
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

func (e *GameEngine) captureActionResourceSnapshot() *actionResourceSnapshot {
	snap := &actionResourceSnapshot{
		playerGems:     map[string]int{},
		playerCrystals: map[string]int{},
		playerTokens:   map[string]map[string]int{},
		campGems:       map[model.Camp]int{},
		campCrystals:   map[model.Camp]int{},
	}
	if e == nil || e.State == nil {
		return snap
	}
	snap.campGems[model.RedCamp] = e.State.RedGems
	snap.campGems[model.BlueCamp] = e.State.BlueGems
	snap.campCrystals[model.RedCamp] = e.State.RedCrystals
	snap.campCrystals[model.BlueCamp] = e.State.BlueCrystals

	for id, p := range e.State.Players {
		if p == nil {
			continue
		}
		snap.playerGems[id] = p.Gem
		snap.playerCrystals[id] = p.Crystal
		tokens := map[string]int{}
		for token, value := range p.Tokens {
			tokens[token] = value
		}
		snap.playerTokens[id] = tokens
	}
	return snap
}

func (e *GameEngine) recordPlayerResourceDelta(previous, current map[string]int, gains, costs map[string]int) {
	for _, id := range orderedStringCounterUnion(previous, current) {
		delta := current[id] - previous[id]
		if delta > 0 {
			gains[id] += delta
		}
		if delta < 0 {
			costs[id] += -delta
		}
	}
}

func (e *GameEngine) recordCampResourceDelta(previous, current map[model.Camp]int, gains, costs map[model.Camp]int) {
	for _, camp := range orderedCampCounterUnion(previous, current) {
		delta := current[camp] - previous[camp]
		if delta > 0 {
			gains[camp] += delta
		}
		if delta < 0 {
			costs[camp] += -delta
		}
	}
}

func (e *GameEngine) recordTokenResourceDelta(previous, current map[string]map[string]int, gains, costs map[actionTokenKey]int) {
	playerIDs := map[string]bool{}
	for id := range previous {
		playerIDs[id] = true
	}
	for id := range current {
		playerIDs[id] = true
	}
	ids := make([]string, 0, len(playerIDs))
	for id := range playerIDs {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	for _, id := range ids {
		before := previous[id]
		after := current[id]
		for _, token := range orderedStringCounterUnion(before, after) {
			delta := after[token] - before[token]
			key := actionTokenKey{playerID: id, token: token}
			if delta > 0 {
				gains[key] += delta
			}
			if delta < 0 {
				costs[key] += -delta
			}
		}
	}
}

func orderedStringCounterUnion(a, b map[string]int) []string {
	seen := map[string]bool{}
	keys := make([]string, 0, len(a)+len(b))
	for key := range a {
		if seen[key] {
			continue
		}
		seen[key] = true
		keys = append(keys, key)
	}
	for key := range b {
		if seen[key] {
			continue
		}
		seen[key] = true
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func orderedCampCounterUnion(a, b map[model.Camp]int) []model.Camp {
	seen := map[model.Camp]bool{}
	camps := make([]model.Camp, 0, len(a)+len(b))
	for _, camp := range []model.Camp{model.RedCamp, model.BlueCamp} {
		if _, ok := a[camp]; ok {
			seen[camp] = true
			camps = append(camps, camp)
			continue
		}
		if _, ok := b[camp]; ok {
			seen[camp] = true
			camps = append(camps, camp)
		}
	}
	tail := make([]string, 0)
	for camp := range a {
		if !seen[camp] {
			tail = append(tail, string(camp))
			seen[camp] = true
		}
	}
	for camp := range b {
		if !seen[camp] {
			tail = append(tail, string(camp))
			seen[camp] = true
		}
	}
	sort.Strings(tail)
	for _, camp := range tail {
		camps = append(camps, model.Camp(camp))
	}
	return camps
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
