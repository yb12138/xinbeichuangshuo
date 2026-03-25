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
	// 记录“当前回合内各角色已对哪些敌方角色造成过法术伤害”。
	turnMagicDamageTargets map[string]map[string]bool
	actionSummary          *actionSummary
	actionSummaryTurn      int
	suppressSealOnDiscard  bool
}

func NewGameEngine(observer model.GameObserver) *GameEngine {
	skills.InitHandlers()
	engine := &GameEngine{
		State:                  model.NewGameState(),
		observer:               observer,
		turnMagicDamageTargets: map[string]map[string]bool{},
		actionSummaryTurn:      0,
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
		Orientation:    model.OrientationNormal,
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
	e.refreshPlayerDerivedState(player)
	return nil
}

// buildMagicMissilePrompt 构建魔弹响应提示
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

	e.State.TurnStage = model.TurnStageTurnBeforeStart
	e.State.CombatStage = model.CombatStageNone
	e.State.Subflow = model.SubflowNone
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
		case model.EffectWeak:
			// 虚弱需要保留到玩家作出选择时再移除，交由后续的 InterruptChoice 处理。
			remain = append(remain, fc)
			continue
		default:
			remain = append(remain, fc)
			continue
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

// Drive 状态机驱动函数，自动在阶段间转换或等待用户输入
func (e *GameEngine) Drive() {
	const maxIterations = 100
	iterations := 0
	for {
		e.Log(fmt.Sprintf("[Debug] Drive Loop: %d, %s", iterations, e.runtimeStateLabel()))

		iterations++
		if iterations > maxIterations {
			e.Log(fmt.Sprintf("[System] 严重错误：状态机死循环检测 (最后状态: %s)", e.runtimeStateLabel()))
			// 紧急制动：强制进入等待状态，避免崩溃
			return
		}
		// 如果有待处理的中断，不自动推进
		if e.State.PendingInterrupt != nil {
			return
		}
		// 仅在没有待处理延迟伤害时推进“延迟后续”。
		// 这样可保证诸如“封印伤害先结算，再继续技能后续”的严格顺序。
		if !e.isDamageResolutionActive() &&
			len(e.State.PendingDamageQueue) == 0 &&
			len(e.State.DeferredFollowups) > 0 {
			e.processDeferredFollowups()
			if e.State.PendingInterrupt != nil {
				return
			}
			continue
		}

		// 行动收尾：先跑行动结束后的全局 hook，再输出汇总信息。
		if e.runActionFinalizeHooksIfIdle() {
			if e.State.PendingInterrupt != nil {
				return
			}
			if !e.isActionFinalizeIdle() {
				continue
			}
		}

		// 行动汇总：当系统回到可继续行动的空闲状态时输出汇总信息
		e.finalizeActionSummaryIfIdle()

		currentPid := e.State.PlayerOrder[e.State.CurrentTurn]
		player := e.State.Players[currentPid]
		if outcome := e.driveNonTurnPhase(currentPid, player); outcome != driveUnhandled {
			if outcome == driveStop {
				return
			}
			continue
		}
		if outcome := e.driveTurnFSM(currentPid, player); outcome != driveUnhandled {
			if outcome == driveStop {
				return
			}
			continue
		}
		return
	}
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
		e.addCampCup(model.RedCamp)
		e.Log(fmt.Sprintf("[Action] %s 合成星杯！红方星杯+1，蓝方士气-1", p.Name))
		e.State.BlueMorale--
	} else {
		e.addCampCup(model.BlueCamp)
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

func (e *GameEngine) adventurerParadiseEligibleAllies(user *model.Player, transferTotal int) []string {
	if user == nil || transferTotal <= 0 {
		return nil
	}
	var allyIDs []string
	for _, ally := range e.State.Players {
		if ally == nil || ally.Camp != user.Camp || ally.ID == user.ID {
			continue
		}
		room := e.getPlayerEnergyCap(ally) - (ally.Gem + ally.Crystal)
		if room >= transferTotal {
			allyIDs = append(allyIDs, ally.ID)
		}
	}
	sort.Strings(allyIDs)
	return allyIDs
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
		e.turnMagicDamageTargets = map[string]map[string]bool{}
	}
	e.turnMagicDamageTargets = map[string]map[string]bool{}
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

func (e *GameEngine) allOtherPlayerIDs(userID string) []string {
	var ids []string
	for _, pid := range e.State.PlayerOrder {
		if pid == userID {
			continue
		}
		if p := e.State.Players[pid]; p != nil {
			ids = append(ids, p.ID)
		}
	}
	return ids
}

// 吟游诗人：记录“当前回合吟游诗人自己已对哪些敌方角色造成过法术伤害”，并在满足条件时触发沉沦协奏曲。
func (e *GameEngine) tryTriggerBardDescentAfterMagicDamage(pd *model.PendingDamage) bool {
	if pd == nil || pd.Damage <= 0 {
		return false
	}
	source := e.State.Players[pd.SourceID]
	target := e.State.Players[pd.TargetID]
	if source == nil || target == nil || source.Camp == target.Camp {
		return false
	}
	if !e.isBard(source) || !source.IsActive {
		return false
	}

	if e.turnMagicDamageTargets == nil {
		e.resetTurnMagicDamageTracker()
	}
	if _, ok := e.turnMagicDamageTargets[source.ID]; !ok {
		e.turnMagicDamageTargets[source.ID] = map[string]bool{}
	}
	e.turnMagicDamageTargets[source.ID][target.ID] = true
	if len(e.turnMagicDamageTargets[source.ID]) < 2 {
		return false
	}
	if source.Tokens == nil {
		source.Tokens = map[string]int{}
	}
	// 仅普通形态，且每回合仅触发一次。
	if hasBardEternalPrisonerForm(source) || source.Tokens["bd_descent_used_turn"] > 0 {
		return false
	}
	if bardMaxSameElementCount(source) < 2 {
		return false
	}
	e.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: source.ID,
		Context: map[string]interface{}{
			"choice_type": "bd_descent_element",
			"user_id":     source.ID,
		},
	})
	e.Log(fmt.Sprintf("%s 满足 [沉沦协奏曲] 触发条件，强制进入弃2张同系牌流程", source.Name))
	return true
}

func (e *GameEngine) bardResponseContext(user *model.Player, stage string, resumePoint interface{}) *model.Context {
	ctx := e.buildContext(user, nil, model.TriggerNone, &model.EventContext{
		Type:     model.EventNone,
		SourceID: user.ID,
	})
	ctx.Selections["bd_song_stage"] = stage
	ctx.Selections["response_resume_phase"] = normalizeChoiceResumePoint(resumePoint)
	return ctx
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
	if handled, label, err := e.resolveDeferredFollowup(f); handled {
		if err != nil {
			e.Log(fmt.Sprintf("[%s] 延迟后续结算失败: %v", label, err))
		}
		return true
	}
	e.Log(fmt.Sprintf("[Warn] 未知的延迟后续类型: %s", f.Type))
	return true
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
	if !e.isBlazeWitch(player) || !hasBlazeWitchFlameForm(player) {
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

var promptButtonLabelByID = map[string]string{
	"confirm":    "发动",
	"yes":        "发动",
	"no":         "放弃",
	"cancel":     "取消",
	"skip":       "放弃",
	"take":       "承受",
	"counter":    "应战",
	"defend":     "防御",
	"normal":     "顺序",
	"reverse":    "反向",
	"attack":     "攻击",
	"magic":      "法术",
	"special":    "特殊",
	"buy":        "购买",
	"synthesize": "合成",
	"extract":    "提炼",
	"cannot_act": "放弃",
	"pass":       "放弃",
}

func parsePromptNonNegativeInt(raw string) (int, bool) {
	val, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || val < 0 {
		return 0, false
	}
	return val, true
}

func isPromptDeclineLabel(label string) bool {
	trimmed := strings.TrimSpace(label)
	if trimmed == "" {
		return false
	}
	return strings.Contains(trimmed, "不发动") ||
		strings.Contains(trimmed, "放弃") ||
		strings.Contains(trimmed, "跳过") ||
		strings.Contains(trimmed, "无法行动") ||
		strings.Contains(trimmed, "拒绝")
}

func shouldUseNumericPromptButtons(prompt *model.Prompt, options []model.PromptOption) (bool, bool) {
	if prompt == nil || len(options) < 2 {
		return false, false
	}
	if prompt.Type == model.PromptChooseCards {
		return false, false
	}
	numericIDs := make([]int, 0, len(options))
	hasLongLabel := false
	hasXHint := strings.ContainsAny(strings.ToLower(prompt.Message), "xｘ")
	for _, option := range options {
		if n, ok := parsePromptNonNegativeInt(option.ID); ok {
			numericIDs = append(numericIDs, n)
		}
		label := strings.TrimSpace(option.Label)
		if len([]rune(label)) >= 8 || strings.Contains(label, "分支") {
			hasLongLabel = true
		}
		lowLabel := strings.ToLower(label)
		if strings.Contains(lowLabel, "x=") || strings.Contains(label, "X=") || strings.Contains(lowLabel, "x值") || strings.ContainsAny(lowLabel, "xｘ") {
			hasXHint = true
		}
	}
	if len(numericIDs) < 2 || (!hasLongLabel && !hasXHint) {
		return false, false
	}
	minID := numericIDs[0]
	for _, id := range numericIDs[1:] {
		if id < minID {
			minID = id
		}
	}
	return true, minID == 0
}

func normalizePromptOptionForClient(option model.PromptOption, prompt *model.Prompt, useNumeric bool, plusOne bool) model.PromptOption {
	label := strings.TrimSpace(option.Label)
	button := strings.TrimSpace(option.ButtonLabel)
	hint := strings.TrimSpace(option.Hint)
	optionID := strings.ToLower(strings.TrimSpace(option.ID))

	if button == "" {
		if mapped, ok := promptButtonLabelByID[optionID]; ok {
			button = mapped
		}
	}
	if button == "" && prompt != nil && prompt.Type == model.PromptChooseSkill {
		button = "发动"
	}
	if button == "" && optionID == "-1" {
		if strings.Contains(label, "完成") || strings.Contains(label, "结束") {
			button = "完成"
		} else {
			button = "放弃"
		}
	}
	if button == "" && useNumeric {
		if n, ok := parsePromptNonNegativeInt(option.ID); ok {
			if plusOne {
				button = strconv.Itoa(n + 1)
			} else {
				button = strconv.Itoa(n)
			}
		}
	}
	if button == "" && isPromptDeclineLabel(label) {
		button = "放弃"
	}
	if button == "" {
		if label != "" && len([]rune(label)) <= 6 {
			button = label
		} else {
			button = "执行"
		}
	}

	isCombatResponseOption := optionID == "take" || optionID == "defend" || optionID == "counter" ||
		button == "承受" || button == "防御" || button == "应战"
	if isCombatResponseOption {
		hint = ""
	}

	if hint == "" && !isCombatResponseOption && label != "" && label != button {
		if !(button == "取消" && (label == "取消" || label == "取消/跳过")) &&
			!(button == "放弃" && isPromptDeclineLabel(label)) {
			hint = label
		}
	}

	option.ButtonLabel = button
	option.Hint = hint
	return option
}

func (e *GameEngine) decoratePromptForClient(prompt *model.Prompt) *model.Prompt {
	if prompt == nil {
		return nil
	}
	cp := *prompt
	if prompt.Options != nil {
		useNumeric, plusOne := shouldUseNumericPromptButtons(prompt, prompt.Options)
		cp.Options = make([]model.PromptOption, 0, len(prompt.Options))
		for _, option := range prompt.Options {
			cp.Options = append(cp.Options, normalizePromptOptionForClient(option, prompt, useNumeric, plusOne))
		}
	}
	if prompt.SpecialOptions != nil {
		useNumeric, plusOne := shouldUseNumericPromptButtons(prompt, prompt.SpecialOptions)
		cp.SpecialOptions = make([]model.PromptOption, 0, len(prompt.SpecialOptions))
		for _, option := range prompt.SpecialOptions {
			cp.SpecialOptions = append(cp.SpecialOptions, normalizePromptOptionForClient(option, prompt, useNumeric, plusOne))
		}
	}
	if prompt.CounterTargetIDs != nil {
		cp.CounterTargetIDs = append([]string{}, prompt.CounterTargetIDs...)
	}
	if prompt.EffectHints != nil {
		cp.EffectHints = append([]string{}, prompt.EffectHints...)
	}
	return &cp
}

// Notify 统一通知方法 (替换所有的 fmt.Printf)
func (e *GameEngine) Notify(eventType model.GameEventType, msg string, data interface{}) {
	if eventType == model.EventAskInput {
		switch p := data.(type) {
		case *model.Prompt:
			data = e.decoratePromptForClient(p)
		case model.Prompt:
			cp := p
			data = e.decoratePromptForClient(&cp)
		}
	}
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

func (e *GameEngine) dispatchCardTrigger(player *model.Player, trigger model.TriggerType, targetID string, card model.Card) {
	if e == nil || e.dispatcher == nil || player == nil {
		return
	}
	cardCopy := card
	cardCtx := &model.EventContext{
		Type:     model.EventCardUsed,
		SourceID: player.ID,
		TargetID: targetID,
		Card:     &cardCopy,
	}
	skillCtx := e.buildContext(player, nil, trigger, cardCtx)
	e.dispatcher.OnTrigger(trigger, skillCtx)
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
	// 文档口径：元素封印在“打出或展示”对应系别牌时触发，暗弃不视为展示。
	if actionType == "discard" && !hidden && !e.suppressSealOnDiscard && p != nil {
		for i := range cards {
			e.dispatchCardTrigger(p, model.TriggerOnCardRevealed, "", cards[i])
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
		"phase":       phase, // attack/defend/take/counter/shield
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
	if target == nil {
		e.Log("[Draw] 跳过恢复摸牌：目标不存在")
		return
	}

	// 检查是否取消扣卡
	if ctx.Flags["cancelDraw"] {
		e.Log(fmt.Sprintf("[Draw] %s 的摸牌被替换/取消", target.Name))
		e.enqueuePendingDrawFollowup(ctx)
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
		e.Log(fmt.Sprintf("[Draw] %s 本次无需摸牌", target.Name))
		e.enqueuePendingDrawFollowup(ctx)
		return
	}

	reason, _ := ctx.Selections["draw_reason"].(string)
	if reason == "" {
		reason = "resume_draw"
	}

	// 执行摸牌
	e.Log(fmt.Sprintf("[Draw] %s 摸牌 %d 张", target.Name, drawCount))
	e.executeResolvedDraw(ctx, drawCount, reason)
	e.enqueuePendingDrawFollowup(ctx)
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

	if handled, err := e.handleBeastSamuraiDiscardInput(playerID, indices); handled || err != nil {
		return err
	}

	if hasSkillID && skillID != "" {
		return e.handleSkillDiscardSelection(playerID, indices, data)
	}

	return e.handleDiscardSelection(playerID, indices, data)
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
		e.setGameOver(true)
		return
	}
	if e.State.BlueCups >= 5 {
		e.Notify(model.EventGameEnd, "蓝方胜利！星杯达到 5", nil)
		e.setGameOver(true)
		return
	}
	// 检查是否有玩家的士气归零
	for _, player := range e.State.Players {
		if player.Camp == model.RedCamp && e.State.RedMorale <= 0 {
			e.Notify(model.EventGameEnd, "蓝方胜利！红方士气归零", nil)
			e.setGameOver(true)
			return
		}
		if player.Camp == model.BlueCamp && e.State.BlueMorale <= 0 {
			e.Notify(model.EventGameEnd, "红方胜利！蓝方士气归零", nil)
			e.setGameOver(true)
			return
		}
	}
}

// GetCurrentPrompt 获取当前用户交互提示
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
		e.enterResponseWindow()
	case model.InterruptDiscard:
		e.enterDiscardSelection()
	case model.InterruptStartupSkill:
		e.clearSubflow()
		e.clearCombatStage()
		if e.State.TurnStage != model.TurnStageTurnStart && e.State.TurnStage != model.TurnStageActionStart {
			e.setTurnStage(model.TurnStageActionStart)
		}
	case model.InterruptChoice:
		e.clearSubflow()
		e.clearCombatStage()
		if e.State.TurnStage == "" {
			e.setTurnStage(model.TurnStageActionExecution)
		}
	case model.InterruptMagicMissile:
		e.enterResponseWindow()
	case model.InterruptGiveCards:
		e.enterDiscardSelection()
	case model.InterruptMagicBulletFusion, model.InterruptMagicBulletDirection:
		e.clearSubflow()
		e.clearCombatStage()
		e.setTurnStage(model.TurnStageActionExecution)
	case model.InterruptHolySwordDraw:
		e.clearSubflow()
		e.setCombatStage(model.CombatStageDraw)
	case model.InterruptSaintHeal:
		e.clearSubflow()
		e.setCombatStage(model.CombatStageHeal)
	case model.InterruptMagicBlast:
		e.enterResponseWindow()
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
	if e.State.PendingInterrupt.Type == model.InterruptResponseSkill && e.prunePendingResponseSkills() {
		if err := e.SkipResponse(); err != nil {
			e.Log(fmt.Sprintf("[System] 自动跳过无可用响应失败: %v", err))
		}
		return
	}
	prompt := e.buildPendingInterruptPrompt()
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
	if e.State.PendingInterrupt == nil {
		return nil
	}

	data, ok := e.State.PendingInterrupt.Context.(map[string]interface{})
	if !ok {
		return nil
	}

	choiceType, _ := data["choice_type"].(string)
	playerID := e.State.PendingInterrupt.PlayerID
	player := e.State.Players[playerID]

	return e.buildRegisteredChoicePrompt(choiceType, playerID, player, data)
}

// SkipResponse 跳过响应阶段
func (e *GameEngine) SkipResponse() error {
	// 格斗家攻击前互斥技能串行提示：
	// 若当前是【蓄力一击】确认框且玩家选择跳过，则继续弹出【气绝崩击】确认框。
	if e.maybeAdvanceFighterAttackStartResponse() {
		return nil
	}
	state := e.captureResponseResumeStateFromInterrupt(responseCompletionSkip, "", e.State.PendingInterrupt)

	// 使用 PopInterrupt 处理队列
	e.PopInterrupt()
	e.runResponseSkipHooks(&state)
	e.restoreSkippedResponseAfterPop(state)

	return nil
}

// maybeAdvanceFighterAttackStartResponse 在“跳过蓄力一击”时推进到“气绝崩击”确认框。
func (e *GameEngine) maybeAdvanceFighterAttackStartResponse() bool {
	intr := e.State.PendingInterrupt
	if intr == nil || intr.Type != model.InterruptResponseSkill || intr.PlayerID == "" {
		return false
	}
	if len(intr.SkillIDs) != 1 || intr.SkillIDs[0] != "fighter_charge_strike" {
		return false
	}
	player := e.State.Players[intr.PlayerID]
	if player == nil || !e.isFighter(player) {
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
	if ctx == nil || ctx.Trigger != model.TriggerOnAttackStart || ctx.TriggerCtx == nil || ctx.TriggerCtx.AttackInfo == nil {
		return false
	}
	// 仅处理主动攻击前链路。
	if ctx.TriggerCtx.AttackInfo.CounterInitiator != "" {
		return false
	}
	if e.dispatcher == nil || !e.dispatcher.isSkillStillUsable("fighter_burst_crash", player, ctx) {
		return false
	}

	intr.SkillIDs = []string{"fighter_burst_crash"}
	e.Log(fmt.Sprintf("%s 放弃 [蓄力一击]，继续询问是否发动 [气绝崩击]", player.Name))
	e.notifyInterruptPrompt()
	return true
}

// markPendingAttackDamageHitProcessed 将命中后响应结束的攻击伤害标记为“已完成 OnAttackHit”。
// 这样后续只继续受伤结算，不会再次触发 OnAttackHit 的响应弹框。
func (e *GameEngine) markPendingAttackDamageHitProcessed(ctx *model.Context) bool {
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
		pd.AttackHitTriggerChecked = true
		return true
	}
	return false
}

func (e *GameEngine) resolveHeroRoarMiss(attackerID string) {
	e.resolveHeroRoarMissWithOverride(attackerID, false)
}

func (e *GameEngine) resolveFighterChargeMiss(attackerID string) {
	e.resolveFighterChargeMissWithOverride(attackerID, false)
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
		e.resolveMagicBowPierceMiss(top.AttackerID, top.TargetID, top.Card, top.IsCounter)
		if atk := e.State.Players[top.AttackerID]; atk != nil && atk.Tokens != nil {
			atk.Tokens["elf_elemental_shot_thunder_pending"] = 0
		}
		e.clearCombatStack()
		if len(e.State.PendingDamageQueue) > 0 {
			e.setReturnPoint(model.TurnStageExtraAction)
			e.enterDamageResolution(nil)
		} else {
			e.enterExtraActionStage()
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
		e.resolveMagicBowPierceMiss(top.AttackerID, top.TargetID, top.Card, top.IsCounter)
		if atk := e.State.Players[top.AttackerID]; atk != nil && atk.Tokens != nil {
			atk.Tokens["elf_elemental_shot_thunder_pending"] = 0
		}
		e.State.CombatStack = e.State.CombatStack[:len(e.State.CombatStack)-1]
		e.initCombat(counterPlayerID, counterTargetID, &counterCard, false, true, false, true)
		if counterPlayer != nil && counterTarget != nil {
			e.Log(fmt.Sprintf("[Combat] %s 应战成功！攻击转移向 %s", counterPlayer.Name, counterTarget.Name))
		}
		return true
	default:
		return false
	}
}

func (e *GameEngine) hasUsableShieldForCombat(target *model.Player, combatReq model.CombatRequest) bool {
	if target == nil {
		return false
	}
	if attacker := e.State.Players[combatReq.AttackerID]; attacker != nil && attacker.Tokens != nil &&
		attacker.Tokens["bs_ignore_shield_current_attack"] > 0 {
		return false
	}
	if combatReq.IgnoreShield {
		// 列风技：无视圣盾
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
	e.NotifyCombatCue(combatReq.AttackerID, combatReq.TargetID, "shield")
	e.Log(fmt.Sprintf("[Combat] %s 选择承受伤害，触发【圣盾】抵挡本次攻击！", target.Name))
	e.resolveMagicBowPierceMiss(combatReq.AttackerID, combatReq.TargetID, combatReq.Card, combatReq.IsCounter)
	if atk := e.State.Players[combatReq.AttackerID]; atk != nil && atk.Tokens != nil {
		atk.Tokens["elf_elemental_shot_thunder_pending"] = 0
	}
	e.clearCombatStack()
	if len(e.State.PendingDamageQueue) > 0 {
		e.setReturnPoint(model.TurnStageExtraAction)
		e.enterDamageResolution(nil)
	} else {
		e.enterExtraActionStage()
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
			SourceID:     combatReq.AttackerID,
			TargetID:     combatReq.TargetID,
			Damage:       combatReq.Card.Damage,
			DamageType:   "Attack",
			Card:         combatReq.Card,
			IsCounter:    combatReq.IsCounter, // 应战命中→加水晶，主动命中→加宝石
			IgnoreShield: combatReq.IgnoreShield,
		}

		// 战斗伤害优先处理，插入到队列头部
		e.State.PendingDamageQueue = append([]model.PendingDamage{pd}, e.State.PendingDamageQueue...)

		e.addActionResponse(fmt.Sprintf("%s 承受伤害", player.Name))
		e.NotifyActionStep(fmt.Sprintf("%s承受伤害", model.GetPlayerDisplayName(player)))
		e.NotifyCombatCue(combatReq.AttackerID, combatReq.TargetID, "take")
		e.Log(fmt.Sprintf("[Combat] %s 选择承受伤害，进入伤害结算流程", player.Name))

		// 战斗结束后，先进入 ActionEnd 统一派发 OnPhaseEnd，再由状态机进入额外行动/回合结束。
		e.setReturnPoint(model.TurnStageActionEnd)
		e.enterDamageResolution(nil)
		return nil

	case "defend":
		// 防御：仅允许打出【圣光】；【圣盾】必须提前放置为场上效果并自动生效。
		if attacker := e.State.Players[combatReq.AttackerID]; attacker != nil && attacker.Tokens != nil &&
			attacker.Tokens["bs_no_holy_defend_current_attack"] > 0 {
			return errors.New("本次攻击受【一击无念】影响，不能使用【圣光】防御")
		}
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

		e.dispatchCardTrigger(player, model.TriggerOnCardUsed, "", card)
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
		e.resolveMagicBowPierceMiss(combatReq.AttackerID, combatReq.TargetID, combatReq.Card, combatReq.IsCounter)
		if attacker := e.State.Players[combatReq.AttackerID]; attacker != nil && attacker.Tokens != nil {
			attacker.Tokens["elf_elemental_shot_thunder_pending"] = 0
		}
		e.clearCombatStack()
		if len(e.State.PendingDamageQueue) > 0 {
			e.setReturnPoint(model.TurnStageActionEnd)
			e.enterDamageResolution(nil)
		} else {
			e.enterActionEndStage()
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
		useShadowMagicBulletCounter := e.canUseShadowRejectResponseMagic(player) &&
			card.Type == model.CardTypeMagic &&
			card.Name == "魔弹"
		useFactionCounter := false

		if !useShadowMagicBulletCounter {
			if card.Type != model.CardTypeAttack {
				return errors.New("只能使用攻击牌进行应战（暗影抗拒下可在非自己行动阶段使用【魔弹】）")
			}
			card = e.applyBlazeWitchAttackCardTransform(player, card)

			// 验证应战卡牌元素（规则：同系或暗灭）
			// 暗灭不可被应战，只能承受伤害或使用圣光（若有场上圣盾会自动生效）
			if combatReq.Card.Element == model.ElementDark {
				return errors.New("暗灭无法被应战，只能承受伤害或使用圣光抵挡（场上圣盾会自动生效）")
			}
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
		} else {
			e.Log(fmt.Sprintf("[Combat] %s 触发[暗影抗拒]：非自己行动阶段使用【魔弹】应战", player.Name))
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

		e.dispatchCardTrigger(player, model.TriggerOnCardUsed, "", card)
		e.NotifyCardRevealed(act.PlayerID, []model.Card{card}, "counter")
		e.NotifyCombatCue(combatReq.AttackerID, combatReq.TargetID, "counter")
		// 消耗应战牌
		if _, err := consumePlayableCardByIndex(player, act.CardIndex); err != nil {
			return err
		}
		e.State.DiscardPile = append(e.State.DiscardPile, card)
		if useFactionCounter {
			e.applyOnmyojiFactionCounterBonuses(player, &card)
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
		e.resolveMagicBowPierceMiss(combatReq.AttackerID, combatReq.TargetID, combatReq.Card, combatReq.IsCounter)
		if origAttacker := e.State.Players[combatReq.AttackerID]; origAttacker != nil && origAttacker.Tokens != nil {
			origAttacker.Tokens["elf_elemental_shot_thunder_pending"] = 0
		}

		// 弹出当前战斗请求
		e.State.CombatStack = e.State.CombatStack[:len(e.State.CombatStack)-1]

		// 创建新的战斗请求（应战反弹，IsCounter=true）
		e.initCombat(act.PlayerID, targetID, &card, false, true, false, true)

		e.Log(fmt.Sprintf("[Combat] %s 应战成功！攻击转移向 %s", player.Name, target.Name))

		if len(e.State.PendingDamageQueue) > 0 {
			e.setReturnPoint(model.CombatStageHitCheck)
			e.enterDamageResolution(nil)
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

	e.enterActionExecutionStage() // 重置到行动选择/执行入口
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
	e.refreshPlayerDerivedState(player)
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
	case "mercy", "怜悯":
		return model.EffectMercy, nil
	default:
		return "", fmt.Errorf("未知效果: %s", raw)
	}
}

func debugEffectTrigger(effect model.EffectType) model.EffectTrigger {
	switch effect {
	case model.EffectPoison, model.EffectWeak:
		return model.EffectTriggerOnBeforeAction
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
	consumeHeroTauntOnAttack := false
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
			e.enterTurnEndStage()
			return nil
		}
		targetID := act.TargetID
		if targetID == "" && len(act.TargetIDs) > 0 {
			targetID = act.TargetIDs[0]
		}
		if targetID != "" && targetID != tauntSourceID {
			if e.State.Players[targetID] == nil {
				targetID = ""
			}
		}
		if targetID != "" && targetID != tauntSourceID {
			srcName := tauntSourceID
			if src := e.State.Players[tauntSourceID]; src != nil {
				srcName = model.GetPlayerDisplayName(src)
			}
			e.Log(fmt.Sprintf("[Taunt] %s 未攻击挑衅来源 %s，跳过本次行动阶段", player.Name, srcName))
			e.RemoveFieldCard(player.ID, model.EffectHeroTaunt)
			e.enterTurnEndStage()
			return nil
		}
		if targetID == tauntSourceID {
			consumeHeroTauntOnAttack = true
		}
	}

	if player.Tokens == nil {
		player.Tokens = map[string]int{}
	}
	e.maybeReleaseMagicSwordsmanShadowAtActionStart(player)
	if e.isFighter(player) && hasFighterHundredDragonForm(player) {
		switch act.Type {
		case model.CmdAttack:
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
			targetOrder := e.playerOrderPosition(targetID)
			if targetOrder == 0 {
				return fmt.Errorf("目标不存在")
			}
			lockedOrder := player.Tokens["fighter_hundred_dragon_target_order"]
			if lockedOrder == 0 {
				e.clearFighterHundredDragon(player, fmt.Sprintf("%s 的 [百式幻龙拳] 状态异常：未锁定目标，立即转正", player.Name))
				return fmt.Errorf("百式幻龙拳未锁定目标，状态已取消，请重新选择行动")
			}
			if lockedOrder != targetOrder {
				e.clearFighterHundredDragon(player, fmt.Sprintf("%s 攻击目标变化，取消 [百式幻龙拳] 并继续本次攻击", player.Name))
			}
		case model.CmdMagic, model.CmdSkill:
			e.clearFighterHundredDragon(player, fmt.Sprintf("%s 尝试执行法术行动，取消 [百式幻龙拳] 并转正", player.Name))
			return fmt.Errorf("百式幻龙拳状态下不能执行法术行动；状态已取消，请重新选择行动")
		case model.CmdBuy, model.CmdSynthesize, model.CmdExtract:
			e.clearFighterHundredDragon(player, fmt.Sprintf("%s 尝试执行特殊行动，取消 [百式幻龙拳] 并转正", player.Name))
			return fmt.Errorf("百式幻龙拳状态下不能执行特殊行动；状态已取消，请重新选择行动")
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

			// 2. 技能统一复用 TargetIDs；若前端/测试仅传了单个 TargetID，这里兜底补齐。
			targetIDs := append([]string{}, act.TargetIDs...)
			if len(targetIDs) == 0 && act.TargetID != "" {
				targetIDs = append(targetIDs, act.TargetID)
			}
			if err := e.UseSkill(act.PlayerID, act.SkillID, targetIDs, act.Selections); err != nil {
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
			// 3. 若技能已挂起中断，则尊重技能层设置的阶段，避免把 Choice/Response 覆写回 ExtraAction。
			if e.State.PendingInterrupt != nil {
				return nil
			}
			if len(e.State.PendingDamageQueue) > 0 {
				e.enterDamageResolution(model.TurnStageExtraAction)
			} else {
				e.enterExtraActionStage()
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
		if err := e.executeSpecialActionWithRuntime(player, actionType); err != nil {
			return err
		}
		if player.Tokens == nil {
			player.Tokens = map[string]int{}
		}
		player.Tokens["hb_special_used_turn"] = 1
		e.runPostSpecialActionRuntime(player, actionType)
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
			e.enterExtraActionStage()
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
				if hasAssassinStealthForm(target) {
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
		if consumeHeroTauntOnAttack {
			e.RemoveFieldCard(player.ID, model.EffectHeroTaunt)
		}

		// 设置阶段为 BeforeAction
		e.enterActionExecutionStage()
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
			e.enterTurnEndStage()
			return nil
		}

		// 常规阶段的“无法行动”：展示手牌、弃掉所有手牌、摸等量牌、本回合禁止特殊行动
		e.beginActionSummary("cannot_act", player.ID, "无法行动", nil)
		handCount := len(player.Hand)
		if handCount == 0 {
			// 无手牌时允许直接结束本回合行动阶段，避免在“已禁特殊行动”场景下无操作可做而卡死。
			e.Log(fmt.Sprintf("[Action] %s 宣告【无法行动】（无手牌），结束本回合行动阶段", player.Name))
			e.State.HasPerformedStartup = true
			e.enterTurnEndStage()
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
		e.enterActionExecutionStage()
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
	if e.State.GameOver {
		return fmt.Errorf("游戏已结束")
	}

	// 3. 回合权校验
	currentPlayer := e.State.PlayerOrder[e.State.CurrentTurn]
	// 特殊情况：战斗响应阶段，允许目标玩家操作
	if e.isCombatInteractionWindow() {
		// 在战斗响应逻辑内部校验目标ID，这里先放行
	} else {
		// 其他阶段，必须是当前回合玩家
		if act.PlayerID != currentPlayer && act.Type != model.CmdStart {
			return fmt.Errorf("不是你的回合")
		}
	}

	// 这里只调用逻辑处理函数，不要在这里调用 Drive
	var err error

	switch {
	case e.isActionSelectionWindow():
		// 行动选择阶段：处理攻击、法术、特殊行动
		err = e.handleActionSelection(act)

	case e.isCombatInteractionWindow():
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
			err = fmt.Errorf("当前状态 (%s) 不支持该指令", e.runtimeStateLabel())
		}
	}

	// === 6. 统一驱动 ===
	// 如果逻辑执行出错，直接返回错误，不驱动引擎
	if err != nil {
		return err
	}
	e.Log(fmt.Sprintf("[Debug] 指令执行成功，准备 Drive. %s, Interrupt: %v", e.runtimeStateLabel(), e.State.PendingInterrupt))

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
				e.enterActionExecutionStage()
			}
			return nil
		}
		if act.Type == model.CmdSelect {
			return e.ConfirmDiscard(act.PlayerID, act.Selections)
		}
		if act.Type == model.CmdSkill {
			return fmt.Errorf("当前正在处理弃牌步骤，请先在手牌区选择并提交弃牌")
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
				if ct, _ := data["choice_type"].(string); ct != "" {
					if handled, err := e.handleRegisteredChoiceCancel(act.PlayerID, ct); handled || err != nil {
						return err
					}
				}
			}
		}
		if act.Type == model.CmdSelect {
			if data, ok := e.State.PendingInterrupt.Context.(map[string]interface{}); ok {
				if ct, _ := data["choice_type"].(string); ct != "" {
					if handled, err := e.handleRegisteredChoiceMultiSelect(act.PlayerID, ct, act.Selections); handled || err != nil {
						return err
					}
					if _, isLegacyCardMulti := registeredSequentialCardChoiceRemainingCount(ct, data); isLegacyCardMulti {
						return e.handleLegacySequentialCardSelections(act.PlayerID, act.Selections)
					}
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
		})

		// 魔弹结算后，通常本回合法术结束，进入额外行动(触发 ActionEnd)或直接 TurnEnd
		e.setReturnPoint(model.TurnStageExtraAction)
		e.enterDamageResolution(nil)

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

	prompt := e.buildMagicBlastPrompt()
	if prompt == nil {
		return fmt.Errorf("魔爆冲击提示构建失败")
	}

	if stage == "caster_forced_discard" {
		if act.Type != model.CmdSelect || len(act.Selections) == 0 {
			return fmt.Errorf("请选择1张牌弃置")
		}
		selection := act.Selections[0]
		if selection < 0 || selection >= len(prompt.Options) {
			return fmt.Errorf("无效的选择")
		}
		cardIdx, err := strconv.Atoi(prompt.Options[selection].ID)
		if err != nil {
			return fmt.Errorf("无效的卡牌索引")
		}
		if cardIdx < 0 || cardIdx >= len(player.Hand) {
			return fmt.Errorf("无效的卡牌索引")
		}

		card := player.Hand[cardIdx]
		player.Hand = append(player.Hand[:cardIdx], player.Hand[cardIdx+1:]...)
		e.State.DiscardPile = append(e.State.DiscardPile, card)
		e.Log(fmt.Sprintf("[Skill] %s 因【魔爆冲击】弃掉了 %s", player.Name, card.Name))

		if currentTargetIdx < len(targetsRaw) {
			data["stage"] = "target_discard"
			e.State.PendingInterrupt.PlayerID = targetsRaw[currentTargetIdx]
			e.State.PendingInterrupt.Context = data
			if nextTarget := e.State.Players[targetsRaw[currentTargetIdx]]; nextTarget != nil {
				e.Log(fmt.Sprintf("[Skill] %s 需要选择弃一张法术牌或受到2点伤害", nextTarget.Name))
			}
			e.notifyInterruptPrompt()
			return nil
		}

		e.PopInterrupt()
		return nil
	}

	discarded := false
	if act.Type == model.CmdCancel {
		discarded = false
	} else if act.Type == model.CmdSelect && len(act.Selections) > 0 {
		selection := act.Selections[0]
		if selection < 0 || selection >= len(prompt.Options) {
			return fmt.Errorf("无效的选择")
		}
		optionID := prompt.Options[selection].ID
		if optionID != "refuse" {
			cardIdx, err := strconv.Atoi(optionID)
			if err != nil {
				return fmt.Errorf("无效的卡牌索引")
			}
			if cardIdx < 0 || cardIdx >= len(player.Hand) {
				return fmt.Errorf("无效的卡牌索引")
			}

			card := player.Hand[cardIdx]
			if card.Type != model.CardTypeMagic {
				return fmt.Errorf("只能弃置法术牌")
			}
			e.NotifyCardRevealed(player.ID, []model.Card{card}, "discard")
			player.Hand = append(player.Hand[:cardIdx], player.Hand[cardIdx+1:]...)
			e.State.DiscardPile = append(e.State.DiscardPile, card)
			e.Log(fmt.Sprintf("[Skill] %s 弃掉了法术牌 %s", player.Name, card.Name))
			discarded = true
		}
	}

	currentTargetIdx++
	data["current_target"] = currentTargetIdx
	if discarded {
		if currentTargetIdx < len(targetsRaw) {
			e.State.PendingInterrupt.PlayerID = targetsRaw[currentTargetIdx]
			e.State.PendingInterrupt.Context = data
			if nextTarget := e.State.Players[targetsRaw[currentTargetIdx]]; nextTarget != nil {
				e.Log(fmt.Sprintf("[Skill] %s 需要选择弃一张法术牌或受到2点伤害", nextTarget.Name))
			}
			e.notifyInterruptPrompt()
			return nil
		}
		e.PopInterrupt()
		return nil
	}

	e.InflictDamage(casterID, player.ID, 2, "magic")
	e.Log(fmt.Sprintf("[Skill] %s 未弃法术牌，受到2点伤害", player.Name))

	if caster != nil && len(caster.Hand) > 0 {
		data["stage"] = "caster_forced_discard"
		e.State.PendingInterrupt.PlayerID = caster.ID
		e.State.PendingInterrupt.Context = data
		e.notifyInterruptPrompt()
		return nil
	}

	if currentTargetIdx < len(targetsRaw) {
		e.State.PendingInterrupt.PlayerID = targetsRaw[currentTargetIdx]
		e.State.PendingInterrupt.Context = data
		if nextTarget := e.State.Players[targetsRaw[currentTargetIdx]]; nextTarget != nil {
			e.Log(fmt.Sprintf("[Skill] %s 需要选择弃一张法术牌或受到2点伤害", nextTarget.Name))
		}
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
		e.NotifyCardRevealed(player.ID, []model.Card{card}, "magic")

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
		e.resumeHolySwordAftermath()
		return nil
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
				"stay_in_turn":         true,
				"is_damage_resolution": false,
			},
		})
		e.Log(fmt.Sprintf("[Skill] %s 需要弃 %d 张牌", player.Name, x))
		return nil
	}
}

func (e *GameEngine) resumeHolySwordAftermath() {
	if e.State.PendingInterrupt != nil {
		return
	}
	if e.routePendingDamageWithDefaultReturn(model.TurnStageExtraAction) {
		return
	}
	if e.restoreReturnPoint() {
		return
	}
	e.enterExtraActionStage()
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

	targetIDs := saintHealTargetIDsFromContext(data)
	if len(targetIDs) == 0 {
		return fmt.Errorf("圣疗缺少目标")
	}

	stage, _ := data["stage"].(string)

	if stage == "allocate_heal" {
		if len(targetIDs) != 2 {
			return fmt.Errorf("圣疗双目标分配配置无效")
		}
		if len(act.Selections) != 1 {
			return fmt.Errorf("请选择一种治疗分配方式")
		}
		choice := act.Selections[0]
		if choice != 0 && choice != 1 {
			return fmt.Errorf("无效的圣疗分配选项: %d", choice)
		}
		allocations := map[string]int{}
		if choice == 0 {
			allocations[targetIDs[0]] = 2
			allocations[targetIDs[1]] = 1
		} else {
			allocations[targetIDs[0]] = 1
			allocations[targetIDs[1]] = 2
		}
		data["allocations"] = allocations
		data["stage"] = "choose_extra_action"
		e.State.PendingInterrupt.Context = data
		e.notifyInterruptPrompt()
		return nil
	}

	allocations := saintHealAllocationsFromContext(data, targetIDs)
	if len(act.Selections) != 1 {
		return fmt.Errorf("请选择额外行动类型")
	}

	extraActionType := "Attack"
	extraActionLabel := "攻击"
	if act.Selections[0] == 1 {
		extraActionType = "Magic"
		extraActionLabel = "法术"
	} else if act.Selections[0] != 0 {
		return fmt.Errorf("无效的额外行动类型选项: %d", act.Selections[0])
	}

	for _, targetID := range targetIDs {
		healAmount := allocations[targetID]
		if healAmount <= 0 {
			continue
		}
		e.Heal(targetID, healAmount)
		if target := e.State.Players[targetID]; target != nil {
			e.Log(fmt.Sprintf("[Skill] %s 获得 %d 点治疗", target.Name, healAmount))
		}
	}

	e.PopInterrupt()

	player.TurnState.PendingActions = append(player.TurnState.PendingActions, model.ActionContext{
		Source:   "圣疗",
		MustType: extraActionType,
	})
	e.Log(fmt.Sprintf("[Skill] %s 发动 [圣疗]，获得额外%s行动", player.Name, extraActionLabel))
	player.TurnState.HasActed = true
	player.TurnState.LastActionType = string(model.ActionMagic)

	phaseEventCtx := &model.EventContext{
		Type:       model.EventPhaseEnd,
		SourceID:   player.ID,
		ActionType: model.ActionMagic,
	}
	phaseCtx := e.buildContext(player, nil, model.TriggerOnPhaseEnd, phaseEventCtx)
	e.dispatcher.OnTrigger(model.TriggerOnPhaseEnd, phaseCtx)

	if e.State.PendingInterrupt != nil {
		return nil
	}

	if len(e.State.PendingDamageQueue) > 0 {
		e.setReturnPoint(model.TurnStageExtraAction)
		e.enterDamageResolution(nil)
		return nil
	}
	e.enterExtraActionStage()
	return nil
}

func parseChoiceIntSlice(raw interface{}) []int {
	var out []int
	if arr, ok := raw.([]int); ok {
		out = append(out, arr...)
		return out
	}
	if arr, ok := raw.([]interface{}); ok {
		for _, v := range arr {
			if f, ok := v.(float64); ok {
				out = append(out, int(f))
			}
		}
	}
	return out
}

func resolveSelectionToAllowedIndex(selection int, candidates []int, allowed map[int]struct{}) (int, bool) {
	if _, ok := allowed[selection]; ok {
		return selection, true
	}
	if selection >= 0 && selection < len(candidates) {
		mapped := candidates[selection]
		if _, ok := allowed[mapped]; ok {
			return mapped, true
		}
	}
	return 0, false
}

func (e *GameEngine) handleLegacySequentialCardSelections(playerID string, selections []int) error {
	if len(selections) == 0 {
		return fmt.Errorf("请先选择手牌后再提交")
	}
	if e.State.PendingInterrupt == nil || e.State.PendingInterrupt.Type != model.InterruptChoice {
		return fmt.Errorf("当前不存在可处理的选牌中断")
	}
	ctxData, ok := e.State.PendingInterrupt.Context.(map[string]interface{})
	if !ok {
		return fmt.Errorf("选牌上下文错误")
	}
	choiceType, _ := ctxData["choice_type"].(string)
	needCount, supported := registeredSequentialCardChoiceRemainingCount(choiceType, ctxData)
	if !supported {
		return fmt.Errorf("当前选择类型不支持多选提交流程")
	}
	if needCount < 1 {
		needCount = 1
	}
	// 兼容旧客户端逐次提交：当仍为单选提交时，沿用历史流程继续推进。
	if len(selections) == 1 && needCount != 1 {
		return e.handleWeakChoiceInput(playerID, selections[0])
	}
	if len(selections) != needCount {
		return fmt.Errorf("需要选择 %d 张牌", needCount)
	}
	for _, idx := range selections {
		if err := e.handleWeakChoiceInput(playerID, idx); err != nil {
			return err
		}
	}
	return nil
}

// handleWeakChoiceInput 处理虚弱/选择中断
func (e *GameEngine) handleWeakChoiceInput(playerID string, selectionIndex int) error {
	if e.State.PendingInterrupt == nil {
		return fmt.Errorf("没有待处理的中断")
	}

	ctxData, ok := e.State.PendingInterrupt.Context.(map[string]interface{})
	if !ok {
		return fmt.Errorf("中断上下文格式错误")
	}

	choiceType, _ := ctxData["choice_type"].(string)

	if handled, err := e.handleRegisteredChoiceInput(playerID, selectionIndex, ctxData); handled || err != nil {
		return err
	}

	return fmt.Errorf("未知的选择类型: %s", choiceType)
}

// resumePendingAttackHit 恢复被响应技能选择打断的“攻击命中后续结算”
func (e *GameEngine) resumePendingAttackHit(ctxData map[string]interface{}) {
	rawCtx, ok := ctxData["user_ctx"].(*model.Context)
	if !ok || rawCtx == nil || rawCtx.Trigger != model.TriggerOnAttackHit || rawCtx.TriggerCtx == nil {
		return
	}
	// OnAttackHit 的可选响应结束后，继续走通用承伤与应用流程，
	// 避免直接 finishTakeHit 导致与队列重复结算（重复触发/重复伤害）。
	if e.markPendingAttackDamageHitProcessed(rawCtx) {
		if e.State.PendingInterrupt == nil {
			e.enterDamageResolution(nil)
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
			e.enterDamageResolution(nil)
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

func (e *GameEngine) resolveAdventurerUndergroundLaw(user *model.Player) {
	if user == nil {
		return
	}
	e.ModifyGem(string(user.Camp), 2)
	e.Log(fmt.Sprintf("%s 的 [地下法则] 生效，本次购买改为战绩区+2宝石", user.Name))
	e.Log(fmt.Sprintf("[Skill] %s 使用了技能: 地下法则", user.Name))
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
		Timing:     e.defaultTimingForTrigger(trigger),
		TriggerCtx: eventCtx,
		// 初始化 map 避免 handler 写入时 panic
		Selections: make(map[string]any),
		Flags:      make(map[string]bool),
		// 当前PendingInterrupt （仅供Handler读取，不要修改）
		PendingInterrupt: e.State.PendingInterrupt,
		// 自动将单个 Target 包装进 Targets 切片，方便多目标技能处理
		Targets: []*model.Player{},
	}
	ctx.Selections["current_resume_point"] = e.currentChoiceResumePoint()
	ctx.Selections["current_turn_stage"] = string(e.State.TurnStage)
	ctx.Selections["current_combat_stage"] = string(e.State.CombatStage)
	ctx.Selections["current_subflow"] = string(e.State.Subflow)

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

	if !e.isDamageResolutionActive() {
		if e.State.ReturnTurnStage == "" && e.State.ReturnCombatStage == model.CombatStageNone && e.State.ReturnSubflow == model.SubflowNone {
			e.setReturnPoint(e.currentChoiceResumePoint())
		}
		e.enterDamageResolution(nil)
	}
}

// AddPendingDamageFront 将延迟伤害插入队列头部（用于“必须先结算”的伤害）。
func (e *GameEngine) AddPendingDamageFront(pd model.PendingDamage) {
	e.State.PendingDamageQueue = append([]model.PendingDamage{pd}, e.State.PendingDamageQueue...)
	e.Log(fmt.Sprintf("[System] 延迟伤害已前插: Source: %s, Target: %s, Damage: %d, Type: %s",
		pd.SourceID, pd.TargetID, pd.Damage, pd.DamageType))

	if !e.isDamageResolutionActive() {
		if e.State.ReturnTurnStage == "" && e.State.ReturnCombatStage == model.CombatStageNone && e.State.ReturnSubflow == model.SubflowNone {
			e.setReturnPoint(e.currentChoiceResumePoint())
		}
		e.enterDamageResolution(nil)
	}
}

func (e *GameEngine) resolveShieldBlockedAttackAsMiss(pd *model.PendingDamage) {
	if pd == nil || pd.AttackMissResolved || !strings.EqualFold(pd.DamageType, "Attack") {
		return
	}
	attacker := e.State.Players[pd.SourceID]
	target := e.State.Players[pd.TargetID]
	if attacker == nil {
		return
	}

	if pd.AttackHitResourceGranted && pd.AttackHitResourceType != "" {
		if e.rollbackCampResource(attacker.Camp, pd.AttackHitResourceType) {
			pd.AttackHitResourceGranted = false
			e.Log(fmt.Sprintf("[Combat] 本次攻击被【圣盾】完全抵消，已回滚%s方命中战绩", attacker.Camp))
		}
	}

	targetName := pd.TargetID
	if target != nil {
		targetName = target.Name
	}
	e.NotifyCombatCue(pd.SourceID, pd.TargetID, "shield")
	e.NotifyActionStep(fmt.Sprintf("%s 的【圣盾】抵消了本次攻击，判定为未命中", targetName))
	e.Log(fmt.Sprintf("[Combat] %s 的攻击被【圣盾】完全抵消，按未命中处理", attacker.Name))

	e.resolveMagicBowPierceMissWithOverride(pd.SourceID, pd.TargetID, pd.Card, pd.HeroRoarMissArmed, pd.FighterChargeMissArmed, pd.IsCounter)
	if attacker.Tokens != nil {
		attacker.Tokens["elf_elemental_shot_thunder_pending"] = 0
	}
	pd.AttackMissResolved = true
}

func (e *GameEngine) processPendingAttackHit(pd *model.PendingDamage) bool {
	if pd == nil || !strings.EqualFold(pd.DamageType, "Attack") || pd.AttackHitTriggerChecked {
		return false
	}
	if pd.Card == nil {
		pd.AttackHitTriggerChecked = true
		return false
	}
	attacker := e.State.Players[pd.SourceID]
	victim := e.State.Players[pd.TargetID]
	if attacker == nil || victim == nil {
		pd.AttackHitTriggerChecked = true
		return false
	}

	// 角色命中态标记通过伤害运行时 hook 扩展，核心流程不做角色分支判断。
	e.runPendingDamageAttackInitHooks(pd, attacker, victim)

	// 1. 应用攻击伤害修正（技能链 + 仍未迁出的少量兼容逻辑）
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
	pd.Damage = e.applyAttackDamageModifiers(attacker, victim, pd.Damage, action)

	// 2. 攻击命中加星石：主动攻击→宝石，应战→水晶（战绩区上限5）
	resourceType := "gem"
	if pd.IsCounter {
		resourceType = "crystal"
	}
	pd.AttackHitResourceType = resourceType
	pd.AttackHitResourceGranted = e.addCampResource(attacker.Camp, resourceType)
	if pd.AttackHitResourceGranted {
		if resourceType == "crystal" {
			e.Log(fmt.Sprintf("[Combat] 应战攻击命中！%s 方战绩区+1水晶", attacker.Camp))
		} else {
			e.Log(fmt.Sprintf("[Combat] 主动攻击命中！%s 方战绩区+1宝石", attacker.Camp))
		}
	} else {
		e.Log(fmt.Sprintf("[Combat] 攻击命中，但 %s 方战绩区已满，本次不增加星石", attacker.Camp))
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
	hitCtx := e.buildContext(attacker, victim, model.TriggerOnAttackHit, hitEventCtx)
	e.dispatcher.OnTrigger(model.TriggerOnAttackHit, hitCtx)

	// 如果触发了中断 (例如询问是否发动撕裂)，暂停处理。
	// 命中链恢复时由 markPendingAttackDamageHitProcessed 标记 AttackHit 已完成。
	if e.State.PendingInterrupt != nil {
		return true
	}
	// 处理攻击命中后的附加技能分支（如元素射击后续/黄泉震颤）。
	if e.handlePostAttackHitEffects(pd) {
		pd.AttackHitTriggerChecked = true
		return true
	}

	pd.AttackHitTriggerChecked = true
	return false
}

// processPendingDamages 处理伤害队列中的所有伤害
// 返回 true 如果产生了中断需要暂停 Drive
func (e *GameEngine) processPendingDamages() bool {
	for len(e.State.PendingDamageQueue) > 0 {
		// Peek: 取出队列中第一个延迟伤害（暂不弹出，等待所有步骤完成）
		pd := &e.State.PendingDamageQueue[0]

		// 先单独处理攻击伤害命中链（OnAttackHit）。
		if e.processPendingAttackHit(pd) {
			return true
		}

		// 通用承伤前流程（所有伤害都走这里）。
		// 灵魂术士：灵魂链接在“承受伤害前”可转移部分伤害。
		if e.maybeTriggerSoulLinkTransfer(pd) {
			return true
		}

		if !pd.DamageTakenTriggerChecked {
			damageEventCtx := &model.EventContext{
				Type:      model.EventDamage,
				SourceID:  pd.SourceID,
				TargetID:  pd.TargetID,
				DamageVal: &pd.Damage, // 允许技能修改伤害
				Card:      pd.Card,
			}
			damageCtx := e.buildContext(e.State.Players[pd.TargetID], e.State.Players[pd.SourceID], model.TriggerOnDamageTaken, damageEventCtx)
			damageCtx.Flags["IsMagicDamage"] = (pd.DamageType != "Attack" && pd.DamageType != "attack")
			damageCtx.Flags["holy_shield_eligible"] = strings.EqualFold(pd.DamageType, "Attack") ||
				(pd.Card != nil && strings.TrimSpace(pd.Card.Name) == "魔弹")
			damageCtx.Flags["ignore_shield"] = pd.IgnoreShield
			if strings.Contains(strings.ToLower(pd.DamageType), "no_absorb") {
				damageCtx.Flags["NoElementAbsorb"] = true
			}
			if damageCtx.Selections == nil {
				damageCtx.Selections = map[string]any{}
			}
			damageCtx.Selections["damage_type"] = pd.DamageType
			pd.DamageTakenTriggerChecked = true

			e.dispatcher.OnTrigger(model.TriggerOnDamageTaken, damageCtx)

			// 如果触发了中断 (例如询问是否发动减伤技能)，暂停处理
			if e.State.PendingInterrupt != nil {
				return true
			}
			shieldTriggered, _ := damageCtx.Selections["holy_shield_triggered"].(bool)
			if shieldTriggered && pd.Damage <= 0 && strings.EqualFold(pd.DamageType, "Attack") {
				e.resolveShieldBlockedAttackAsMiss(pd)
			}
		}
		if strings.EqualFold(pd.DamageType, "Attack") && !pd.AttackMissResolved && !pd.AttackPostHitEffectsDone {
			e.resolveSwordEmperorAttackHitAftermath(pd)
			pd.AttackPostHitEffectsDone = true
		}
		// 治疗选择阶段：允许受伤方选择是否使用治疗抵消
		if !pd.HealResolved {
			target := e.State.Players[pd.TargetID]
			if target != nil && pd.Damage > 0 && e.canUseHealToResist(target, pd.SourceID, pd.DamageType, pd.IgnoreHeal, pd.AllowCrimsonFaithHeal) {
				maxHeal := target.Heal
				if pd.Damage < maxHeal {
					maxHeal = pd.Damage
				}
				maxHeal = e.applyPendingDamageHealCapHooks(pd, target, maxHeal)
				if maxHeal <= 0 {
					pd.HealResolved = true
				} else {
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
			}
			pd.HealResolved = true
		}
		// 蝶舞者：伤害应用前的时点响应（朝圣/毒粉/镜花水月）。
		if e.maybeTriggerButterflyDamageResponses(pd) {
			return true
		}

		// 应用伤害 & 移除效果
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
			e.applyDamageWithOptions(target, pd.Damage, pd.DamageType, pd.CapDrawToHandLimit, pd.SourceID, pd.SourceSkillID, pd.OverflowMoraleLossFixed)
			// 角色落伤后清理逻辑统一交由 hook 扩展。
			e.runPendingDamageAfterApplyHooks(pd, target)

			// ==========================================
			// 五系封印移除点：伤害结算后移除封印
			// ==========================================
			// 如果指定了 EffectTypeToRemove，在伤害结算后移除场上效果
			// 【五系封印在此处移除】
			// 封印触发时会在 PendingDamage 中设置 EffectTypeToRemove
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
	return false
}
