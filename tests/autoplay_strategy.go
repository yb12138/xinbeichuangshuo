package tests

import (
	"errors"
	"fmt"
	"hash/fnv"
	"sort"
	"strconv"
	"strings"
	"sync"

	"starcup-engine/internal/data"
	"starcup-engine/internal/engine"
	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
)

const (
	autoGamePlayers   = 6
	autoGameStepLimit = 15000
	autoStagnantLimit = 40
)

const (
	autoPlanActionSkill      = "skill"
	autoPlanActionAttack     = "attack"
	autoPlanActionMagic      = "magic"
	autoPlanActionBuy        = "buy"
	autoPlanActionSynthesize = "synthesize"
	autoPlanActionExtract    = "extract"
	autoPlanActionSpecial    = "special"
)

// directedScenarioPlan 定义定向整局回归场景与策略偏好。
type directedScenarioPlan struct {
	Name                       string
	Lineup                     []string
	TargetSkillTitles          []string
	Runs                       int
	RoleActionOrder            map[string][]string // roleID -> 动作优先级
	RoleActionSkillPriority    map[string][]string // roleID -> 主动技能ID优先级
	RoleInterruptSkillPriority map[string][]string // roleID -> 启动/响应技能ID优先级
}

type autoGameObserver struct {
	gameEnded     bool
	endMessage    string
	logs          []string
	skillTriggers map[string]int
}

var actionSkillCursor = map[string]int{}
var errDiscardPrereqNotMet = errors.New("discard prereq not met")
var currentDirectedScenarioPlan *directedScenarioPlan
var knownSkillTitlesOnce sync.Once
var knownSkillTitles map[string]struct{}

func (o *autoGameObserver) OnGameEvent(event model.GameEvent) {
	switch event.Type {
	case model.EventGameEnd:
		o.gameEnded = true
		o.endMessage = event.Message
	case model.EventLog:
		o.logs = append(o.logs, event.Message)
		if len(o.logs) > 240 {
			o.logs = o.logs[len(o.logs)-240:]
		}
		o.captureSkillTrigger(event.Message)
	}
}

func (o *autoGameObserver) captureSkillTrigger(msg string) {
	if o.skillTriggers == nil {
		o.skillTriggers = make(map[string]int)
	}

	seenInMsg := make(map[string]struct{})
	addSkill := func(name string) {
		name = strings.TrimSpace(name)
		if name == "" {
			return
		}
		if !isKnownSkillTitle(name) {
			return
		}
		if _, dup := seenInMsg[name]; dup {
			return
		}
		seenInMsg[name] = struct{}{}
		o.skillTriggers[name]++
	}

	// 格式一: "[Skill] xxx 使用了技能: 技能名"
	const marker = "使用了技能:"
	if idx := strings.Index(msg, marker); idx >= 0 {
		rest := strings.TrimSpace(msg[idx+len(marker):])
		if cut := strings.Index(rest, " ("); cut >= 0 {
			rest = rest[:cut]
		}
		if cut := strings.Index(rest, "（"); cut >= 0 {
			rest = rest[:cut]
		}
		addSkill(rest)
	}

	// 格式二: 识别 [] / 【】中出现的技能名（覆盖 "发动[技能]" / "发动【技能】" / "的 [技能] 触发" 等）
	for _, skillName := range extractBracketSkillCandidates(msg) {
		if strings.Contains(msg, "可以发动【"+skillName+"】") ||
			strings.Contains(msg, "可以发动["+skillName+"]") ||
			strings.Contains(msg, "可以发动 ["+skillName+"]") {
			continue
		}
		addSkill(skillName)
	}

	// 格式三: 识别无括号日志，如 "发动五系束缚"、"狂化发动"
	if strings.Contains(msg, "发动") || strings.Contains(msg, "触发") {
		for skillName := range getKnownSkillTitles() {
			if !strings.Contains(msg, skillName) {
				continue
			}
			if strings.Contains(msg, "可以发动"+skillName) {
				continue
			}
			if strings.Contains(msg, "发动"+skillName) ||
				strings.Contains(msg, skillName+"发动") ||
				strings.Contains(msg, skillName+" 触发") {
				addSkill(skillName)
			}
		}
	}
}

func getKnownSkillTitles() map[string]struct{} {
	knownSkillTitlesOnce.Do(func() {
		knownSkillTitles = collectAllSkillTitles()
	})
	return knownSkillTitles
}

func isKnownSkillTitle(name string) bool {
	_, ok := getKnownSkillTitles()[name]
	return ok
}

func extractBracketSkillCandidates(msg string) []string {
	out := make([]string, 0)
	seen := make(map[string]struct{})
	add := func(name string) {
		name = strings.TrimSpace(name)
		if name == "" {
			return
		}
		if _, dup := seen[name]; dup {
			return
		}
		seen[name] = struct{}{}
		out = append(out, name)
	}

	collect := func(left, right string) {
		start := 0
		for {
			leftIdx := strings.Index(msg[start:], left)
			if leftIdx < 0 {
				break
			}
			leftIdx += start
			rightIdx := strings.Index(msg[leftIdx+len(left):], right)
			if rightIdx < 0 {
				break
			}
			rightIdx += leftIdx + len(left)
			add(msg[leftIdx+len(left) : rightIdx])
			start = rightIdx + len(right)
		}
	}

	collect("[", "]")
	collect("【", "】")
	return out
}

type autoGameResult struct {
	triggeredSkills      map[string]int
	expectedActionSkills map[string]struct{}
	expectedAllSkills    map[string]struct{}
	logs                 []string
}

func collectRoleIDs() []string {
	characters := data.GetCharacters()
	seen := make(map[string]struct{}, len(characters))
	roleIDs := make([]string, 0, len(characters))
	for _, c := range characters {
		if c.ID == "" {
			continue
		}
		if _, ok := seen[c.ID]; ok {
			continue
		}
		seen[c.ID] = struct{}{}
		roleIDs = append(roleIDs, c.ID)
	}
	sort.Strings(roleIDs)
	return roleIDs
}

func collectAllSkillTitles() map[string]struct{} {
	out := make(map[string]struct{})
	for _, c := range data.GetCharacters() {
		for _, s := range c.Skills {
			if s.Title == "" {
				continue
			}
			out[s.Title] = struct{}{}
		}
	}
	return out
}

func collectActionSkillTitles() map[string]struct{} {
	out := make(map[string]struct{})
	for _, c := range data.GetCharacters() {
		for _, s := range c.Skills {
			if s.Type != model.SkillTypeAction || s.Title == "" {
				continue
			}
			out[s.Title] = struct{}{}
		}
	}
	return out
}

func collectRoleSkillTitles() map[string]map[string]struct{} {
	roleSkills := make(map[string]map[string]struct{})
	for _, c := range data.GetCharacters() {
		if c.ID == "" {
			continue
		}
		if _, ok := roleSkills[c.ID]; !ok {
			roleSkills[c.ID] = make(map[string]struct{})
		}
		for _, s := range c.Skills {
			if s.Title == "" {
				continue
			}
			roleSkills[c.ID][s.Title] = struct{}{}
		}
	}
	return roleSkills
}

func buildCoverageLineups(roleIDs []string, size int) [][]string {
	lineups := make([][]string, 0)
	for start := 0; start < len(roleIDs); start += size {
		lineup := make([]string, 0, size)
		for offset := 0; offset < size; offset++ {
			lineup = append(lineup, roleIDs[(start+offset)%len(roleIDs)])
		}
		lineups = append(lineups, lineup)
	}
	return lineups
}

func rotateRoleIDs(roleIDs []string, shift int) []string {
	if len(roleIDs) == 0 {
		return nil
	}
	shift = shift % len(roleIDs)
	if shift < 0 {
		shift += len(roleIDs)
	}
	out := append([]string{}, roleIDs[shift:]...)
	out = append(out, roleIDs[:shift]...)
	return out
}

func mirrorLineup(lineup []string) []string {
	if len(lineup) != autoGamePlayers {
		return append([]string{}, lineup...)
	}
	out := make([]string, 0, len(lineup))
	out = append(out, lineup[autoGamePlayers/2:]...)
	out = append(out, lineup[:autoGamePlayers/2]...)
	return out
}

func mergeTriggered(dst map[string]int, src map[string]int) {
	for name, cnt := range src {
		dst[name] += cnt
	}
}

func mergeSkillSet(dst map[string]struct{}, src map[string]struct{}) {
	for name := range src {
		dst[name] = struct{}{}
	}
}

func coverageStats(expected map[string]struct{}, triggered map[string]int) (triggeredCount, total int, ratio float64) {
	total = len(expected)
	if total == 0 {
		return 0, 0, 0
	}
	for name := range expected {
		if triggered[name] > 0 {
			triggeredCount++
		}
	}
	ratio = float64(triggeredCount) / float64(total)
	return triggeredCount, total, ratio
}

func runAutoGame(lineup []string, maxSteps int) (*autoGameResult, error) {
	return runAutoGameWithScenarioSeedTag(lineup, maxSteps, nil, "baseline")
}

func runAutoGameWithScenario(lineup []string, maxSteps int, scenario *directedScenarioPlan) (*autoGameResult, error) {
	return runAutoGameWithScenarioSeedTag(lineup, maxSteps, scenario, "scenario_default")
}

func runAutoGameWithScenarioSeedTag(
	lineup []string,
	maxSteps int,
	scenario *directedScenarioPlan,
	seedTag string,
) (*autoGameResult, error) {
	currentDirectedScenarioPlan = scenario
	defer func() { currentDirectedScenarioPlan = nil }()

	observer := &autoGameObserver{}
	game := engine.NewGameEngine(observer)
	actionSkillCursor = map[string]int{}

	seed := deterministicSimulationSeed(lineup, scenario, seedTag)
	restoreShuffle := rules.SetDeterministicShuffleSeedForTesting(seed)
	defer restoreShuffle()

	expectedActionSkills := make(map[string]struct{})
	expectedAllSkills := make(map[string]struct{})

	for i, roleID := range lineup {
		pid := fmt.Sprintf("p%d", i+1)
		camp := model.RedCamp
		if i >= autoGamePlayers/2 {
			camp = model.BlueCamp
		}
		if err := game.AddPlayer(pid, fmt.Sprintf("%s_%d", roleID, i+1), roleID, camp); err != nil {
			return nil, fmt.Errorf("add player failed: pid=%s role=%s err=%w", pid, roleID, err)
		}

		p := game.State.Players[pid]
		if p != nil && p.Character != nil {
			for _, skill := range p.Character.Skills {
				if skill.Title != "" {
					expectedAllSkills[skill.Title] = struct{}{}
				}
				if skill.Type == model.SkillTypeAction {
					expectedActionSkills[skill.Title] = struct{}{}
				}
			}
		}
	}

	if err := startAutoGameDeterministically(game, seed); err != nil {
		return nil, fmt.Errorf("start game failed: %w", err)
	}

	lastSnapshot := ""
	stagnant := 0

	for step := 0; step < maxSteps && game.State.Phase != model.PhaseEnd; step++ {
		snapshot := gameplaySnapshot(game)
		if snapshot == lastSnapshot {
			stagnant++
		} else {
			stagnant = 0
			lastSnapshot = snapshot
		}

		if game.State.PendingInterrupt != nil {
			if err := resolveInterrupt(game); err != nil {
				return nil, fmt.Errorf("step %d interrupt=%s failed: %w", step, game.State.PendingInterrupt.Type, err)
			}
			continue
		}

		if stagnant > autoStagnantLimit {
			if tryRecoverFromStall(game) {
				stagnant = 0
				lastSnapshot = ""
				continue
			}
			return nil, fmt.Errorf("stagnated state for too long: %s\nlast logs:\n%s", summarizeGameState(game), observerTailLogs(observer, 20))
		}

		switch game.State.Phase {
		case model.PhaseActionSelection:
			if err := performAggressiveActionSelection(game); err != nil {
				return nil, fmt.Errorf("step %d action selection failed: %w", step, err)
			}
		case model.PhaseCombatInteraction:
			if err := resolveCombatAsTake(game); err != nil {
				return nil, fmt.Errorf("step %d combat response failed: %w", step, err)
			}
		case model.PhaseResponse:
			if len(game.State.ActionStack) > 0 {
				top := game.State.ActionStack[len(game.State.ActionStack)-1]
				if err := game.HandleAction(model.PlayerAction{
					PlayerID:  top.TargetID,
					Type:      model.CmdRespond,
					ExtraArgs: []string{"take"},
				}); err != nil {
					return nil, fmt.Errorf("response take failed: %w", err)
				}
			} else {
				game.Drive()
			}
		default:
			game.Drive()
		}
	}

	if game.State.Phase != model.PhaseEnd {
		return nil, fmt.Errorf(
			"game did not finish within %d steps, %s\nlast logs:\n%s",
			maxSteps,
			summarizeGameState(game),
			observerTailLogs(observer, 20),
		)
	}
	if !observer.gameEnded {
		return nil, fmt.Errorf("phase reached End without GameEnd event, %s\nlast logs:\n%s", summarizeGameState(game), observerTailLogs(observer, 20))
	}

	return &autoGameResult{
		triggeredSkills:      observer.skillTriggers,
		expectedActionSkills: expectedActionSkills,
		expectedAllSkills:    expectedAllSkills,
		logs:                 append([]string{}, observer.logs...),
	}, nil
}

func deterministicSimulationSeed(lineup []string, scenario *directedScenarioPlan, seedTag string) int64 {
	hasher := fnv.New64a()
	_, _ = hasher.Write([]byte(strings.Join(lineup, "|")))
	_, _ = hasher.Write([]byte("::"))
	if scenario != nil {
		_, _ = hasher.Write([]byte(scenario.Name))
	}
	_, _ = hasher.Write([]byte("::"))
	_, _ = hasher.Write([]byte(seedTag))

	seed := int64(hasher.Sum64() & 0x7fffffffffffffff)
	if seed == 0 {
		return 1
	}
	return seed
}

func startAutoGameDeterministically(game *engine.GameEngine, seed int64) error {
	if game == nil || game.State == nil {
		return fmt.Errorf("game is nil")
	}
	if len(game.State.Players) < 2 {
		return errors.New("玩家人数不足")
	}

	game.State.Deck = rules.InitDeck()
	game.State.Deck = rules.Shuffle(game.State.Deck)

	for _, pid := range game.State.PlayerOrder {
		player := game.State.Players[pid]
		cards, newDeck, _ := rules.DrawCards(game.State.Deck, game.State.DiscardPile, 4)
		player.Hand = append(player.Hand, cards...)
		game.State.Deck = newDeck
		ensureStarterRoleCardsForAuto(player)
	}
	applyDirectedScenarioPrestartState(game)

	startIndex := deterministicStartIndex(seed, len(game.State.PlayerOrder))
	firstPlayerID := game.State.PlayerOrder[startIndex]
	game.State.CurrentTurn = startIndex

	first := game.State.Players[firstPlayerID]
	first.IsActive = true
	first.TurnState = model.NewPlayerTurnState()

	game.Log(fmt.Sprintf("[Game] 游戏开始! 首发玩家: %s (%s)", first.Name, first.Camp))

	game.State.Phase = model.PhaseBuffResolve
	game.Drive()
	return nil
}

// ensureStarterRoleCardsForAuto 与后端开局规则保持一致：
// - 封印师开局自带五系束缚专属牌（专属卡区）
// - 血色剑灵开局自带血蔷薇庭院专属牌（专属卡区）
func ensureStarterRoleCardsForAuto(player *model.Player) {
	if player == nil || player.Character == nil {
		return
	}
	ensureZoneCard := func(skillTitle string, card model.Card) {
		for _, c := range player.ExclusiveCards {
			if c.MatchExclusive(player.Character.Name, skillTitle) {
				return
			}
		}
		for i, c := range player.Hand {
			if !c.MatchExclusive(player.Character.Name, skillTitle) {
				continue
			}
			player.Hand = append(player.Hand[:i], player.Hand[i+1:]...)
			player.ExclusiveCards = append(player.ExclusiveCards, c)
			return
		}
		player.ExclusiveCards = append(player.ExclusiveCards, card)
	}

	switch player.Character.ID {
	case "sealer":
		ensureZoneCard("五系束缚", model.Card{
			ID:              "starter-" + player.ID + "-five_elements_bind",
			Name:            "五系束缚",
			Type:            model.CardTypeMagic,
			Element:         model.ElementLight,
			Faction:         player.Character.Faction,
			Damage:          0,
			Description:     "封印师开局自带专属技能卡",
			ExclusiveChar1:  player.Character.Name,
			ExclusiveSkill1: "五系束缚",
		})
	case "crimson_sword_spirit":
		ensureZoneCard("血蔷薇庭院", model.Card{
			ID:              "starter-" + player.ID + "-css_rose_courtyard",
			Name:            "血蔷薇庭院",
			Type:            model.CardTypeMagic,
			Element:         model.ElementDark,
			Faction:         player.Character.Faction,
			Damage:          0,
			Description:     "血色剑灵开局自带专属技能卡",
			ExclusiveChar1:  player.Character.Name,
			ExclusiveSkill1: "血蔷薇庭院",
		})
	}
}

// applyDirectedScenarioPrestartState 仅在定向场景下注入必要初始态，保证低频技能路径可稳定覆盖。
func applyDirectedScenarioPrestartState(game *engine.GameEngine) {
	if game == nil {
		return
	}

	if scenarioTargetsSkill("仪式中断") {
		for _, pid := range game.State.PlayerOrder {
			p := game.State.Players[pid]
			if p == nil || p.Role != "arbiter" {
				continue
			}
			if p.Tokens == nil {
				p.Tokens = map[string]int{}
			}
			// 预置审判形态，使“仪式中断”在首个启动阶段可触发；
			// 同时保留1红宝石，确保后续回合仍可触发“仲裁仪式”。
			p.Tokens["arbiter_form"] = 1
			if p.Gem < 1 {
				p.Gem = 1
			}
			break
		}
	}

	if scenarioTargetsSkill("潜行") {
		for _, pid := range game.State.PlayerOrder {
			p := game.State.Players[pid]
			if p == nil || p.Role != "assassin" {
				continue
			}
			// 潜行需要1红宝石，预置后可稳定覆盖启动触发路径。
			if p.Gem < 1 {
				p.Gem = 1
			}
			break
		}
	}

	if scenarioTargetsSkill("五系束缚") {
		for _, pid := range game.State.PlayerOrder {
			p := game.State.Players[pid]
			if p == nil || p.Role != "sealer" {
				continue
			}
			// 五系束缚需要消耗1水晶，且该定向用例以覆盖该技能为主。
			// 预置后可稳定走到技能分支，降低随机局势导致的覆盖抖动。
			if p.Crystal < 1 {
				p.Crystal = 1
			}
			break
		}
	}

	if scenarioTargetsSkill("烈风技") {
		var bladeMaster *model.Player
		for _, pid := range game.State.PlayerOrder {
			p := game.State.Players[pid]
			if p == nil || p.Role != "blade_master" {
				continue
			}
			bladeMaster = p
			break
		}
		if bladeMaster != nil && bladeMaster.Character != nil {
			// 烈风技需要“打出匹配独有牌 + 目标有圣盾”，这里在定向场景预置硬前提。
			hasGaleSlashCard := false
			for _, c := range bladeMaster.Hand {
				if c.Type == model.CardTypeAttack && c.MatchExclusive(bladeMaster.Character.Name, "烈风技") {
					hasGaleSlashCard = true
					break
				}
			}
			if !hasGaleSlashCard {
				bladeMaster.Hand = append([]model.Card{{
					ID:              "scenario-" + bladeMaster.ID + "-gale_slash",
					Name:            "烈风技",
					Type:            model.CardTypeAttack,
					Element:         model.ElementWind,
					Damage:          2,
					Faction:         bladeMaster.Character.Faction,
					ExclusiveChar1:  bladeMaster.Character.Name,
					ExclusiveSkill1: "烈风技",
				}}, bladeMaster.Hand...)
			}

			// 给所有敌方预置 1 层圣盾，提高“目标拥有圣盾”路径稳定性。
			for _, pid := range game.State.PlayerOrder {
				target := game.State.Players[pid]
				if target == nil || target.Camp == bladeMaster.Camp || playerHasShield(target) {
					continue
				}
				target.AddFieldCard(&model.FieldCard{
					Card: model.Card{
						ID:      "scenario-" + target.ID + "-shield",
						Name:    "圣盾",
						Type:    model.CardTypeMagic,
						Element: model.ElementLight,
					},
					Mode:     model.FieldEffect,
					Effect:   model.EffectShield,
					Trigger:  model.EffectTriggerOnDamaged,
					SourceID: bladeMaster.ID,
				})
			}
		}
	}

	if scenarioTargetsSkill("元素射击") {
		for _, pid := range game.State.PlayerOrder {
			p := game.State.Players[pid]
			if p == nil || p.Role != "elf_archer" || p.Character == nil {
				continue
			}
			hasMagic := false
			hasNonDarkAttack := false
			for _, c := range p.Hand {
				if c.Type == model.CardTypeMagic {
					hasMagic = true
				}
				if c.Type == model.CardTypeAttack && c.Element != model.ElementDark {
					hasNonDarkAttack = true
				}
			}
			if !hasMagic {
				p.Hand = append(p.Hand, model.Card{
					ID:      "scenario-" + p.ID + "-elf-magic",
					Name:    "元素导引",
					Type:    model.CardTypeMagic,
					Element: model.ElementFire,
					Faction: p.Character.Faction,
				})
			}
			if !hasNonDarkAttack {
				p.Hand = append(p.Hand, model.Card{
					ID:      "scenario-" + p.ID + "-elf-attack",
					Name:    "风之矢",
					Type:    model.CardTypeAttack,
					Element: model.ElementWind,
					Damage:  2,
					Faction: p.Character.Faction,
				})
			}
			break
		}
	}

	if scenarioTargetsSkill("希望赋格曲") {
		for _, pid := range game.State.PlayerOrder {
			p := game.State.Players[pid]
			if p == nil || p.Role != "bard" {
				continue
			}
			if p.Crystal < 1 {
				p.Crystal = 1
			}
			break
		}
	}
}

func deterministicStartIndex(seed int64, playerCount int) int {
	if playerCount <= 0 {
		return 0
	}
	if seed < 0 {
		seed = -seed
	}
	return int(seed % int64(playerCount))
}

func performAggressiveActionSelection(game *engine.GameEngine) error {
	if len(game.State.PlayerOrder) == 0 {
		return fmt.Errorf("empty player order")
	}
	if game.State.CurrentTurn < 0 || game.State.CurrentTurn >= len(game.State.PlayerOrder) {
		return fmt.Errorf("invalid current turn: %d", game.State.CurrentTurn)
	}

	pid := game.State.PlayerOrder[game.State.CurrentTurn]
	player := game.State.Players[pid]
	if player == nil {
		return fmt.Errorf("current player not found: %s", pid)
	}

	enemies := collectEnemyIDs(game, player.Camp)
	if len(enemies) == 0 {
		return fmt.Errorf("no enemy found for player %s", pid)
	}

	attackIdx, magicIdx := collectPlayableCards(player)

	// 定向硬前提：剑影必须先拿到蓝水晶；满足条件时优先提炼，避免被通用策略稀释。
	if currentDirectedScenarioPlan != nil &&
		player.TurnState.CurrentExtraAction == "" &&
		player.Role == "blade_master" &&
		scenarioTargetsSkill("剑影") &&
		player.Crystal == 0 &&
		campHasCrystalStock(game, player.Camp) &&
		canAttemptExtract(game, player) {
		if err := game.HandleAction(model.PlayerAction{PlayerID: pid, Type: model.CmdExtract}); err == nil {
			return nil
		}
	}

	// 定向硬前提：贯穿射击要先制造“主动攻击未命中”窗口，优先让神箭手主动攻击。
	if currentDirectedScenarioPlan != nil &&
		player.TurnState.CurrentExtraAction == "" &&
		player.Role == "archer" &&
		scenarioTargetsSkill("贯穿射击") {
		if len(attackIdx) > 0 {
			if err := tryAttackActions(game, pid, enemies, attackIdx); err == nil {
				return nil
			}
		}
		if len(attackIdx) == 0 && canAttemptBuy(game, player) {
			if err := game.HandleAction(model.PlayerAction{PlayerID: pid, Type: model.CmdBuy}); err == nil {
				return nil
			}
		}
	}

	// 定向硬前提：撕裂要求狂战士在攻击命中时持有宝石，先补资源再进攻。
	if currentDirectedScenarioPlan != nil &&
		player.TurnState.CurrentExtraAction == "" &&
		player.Role == "berserker" &&
		scenarioTargetsSkill("撕裂") &&
		player.Gem == 0 {
		if canAttemptExtract(game, player) {
			if err := game.HandleAction(model.PlayerAction{PlayerID: pid, Type: model.CmdExtract}); err == nil {
				return nil
			}
		}
		if canAttemptBuy(game, player) {
			if err := game.HandleAction(model.PlayerAction{PlayerID: pid, Type: model.CmdBuy}); err == nil {
				return nil
			}
		}
	}

	// 定向硬前提：圣疗需要先拿到水晶，优先铺出水晶资源链。
	if currentDirectedScenarioPlan != nil &&
		player.TurnState.CurrentExtraAction == "" &&
		player.Role == "saintess" &&
		scenarioTargetsSkill("圣疗") &&
		player.Crystal == 0 {
		if campHasCrystalStock(game, player.Camp) && canAttemptExtract(game, player) {
			if err := game.HandleAction(model.PlayerAction{PlayerID: pid, Type: model.CmdExtract}); err == nil {
				return nil
			}
		}
		if canAttemptBuy(game, player) {
			if err := game.HandleAction(model.PlayerAction{PlayerID: pid, Type: model.CmdBuy}); err == nil {
				return nil
			}
		}
	}

	// 基线前提：魔弓没有充能时，优先补资源，为后续启动技创造窗口。
	if currentDirectedScenarioPlan == nil &&
		player.TurnState.CurrentExtraAction == "" &&
		player.Role == "magic_bow" &&
		countMagicBowChargeCards(player) == 0 {
		if canAttemptExtract(game, player) && campHasAnyResourceStock(game, player.Camp) {
			if err := game.HandleAction(model.PlayerAction{PlayerID: pid, Type: model.CmdExtract}); err == nil {
				return nil
			}
		}
		if canAttemptBuy(game, player) {
			if err := game.HandleAction(model.PlayerAction{PlayerID: pid, Type: model.CmdBuy}); err == nil {
				return nil
			}
		}
	}

	if currentDirectedScenarioPlan != nil && player.TurnState.CurrentExtraAction == "" {
		if ok, err := tryScenarioActionPlan(game, player, enemies, attackIdx, magicIdx); err != nil {
			return err
		} else if ok {
			return nil
		}
	}

	ok, err := tryActionSkillsAggressive(game, player, enemies)
	if err != nil {
		return err
	}
	if ok {
		return nil
	}
	switch player.TurnState.CurrentExtraAction {
	case "Attack":
		if err := tryAttackActions(game, pid, enemies, attackIdx); err == nil {
			return nil
		}
		// 额外行动约束下若无合法动作，显式宣告“无法行动”以跳过本次额外行动。
		if err := game.HandleAction(model.PlayerAction{PlayerID: pid, Type: model.CmdCannotAct}); err == nil {
			return nil
		}
		return fmt.Errorf("extra attack action has no legal executable move")
	case "Magic":
		// 引擎约束：额外法术行动只能使用法术牌，不可发动技能。
		if currentDirectedScenarioPlan != nil {
			if ok, err := tryScenarioActionPlan(game, player, enemies, attackIdx, magicIdx); err != nil {
				return err
			} else if ok {
				return nil
			}
		}
		if err := tryMagicActions(game, pid, enemies, magicIdx); err == nil {
			return nil
		}
		// 额外行动约束下若无合法动作，显式宣告“无法行动”以跳过本次额外行动。
		if err := game.HandleAction(model.PlayerAction{PlayerID: pid, Type: model.CmdCannotAct}); err == nil {
			return nil
		}
		return fmt.Errorf("extra magic action has no legal executable move")
	}

	if err := tryAttackActions(game, pid, enemies, attackIdx); err == nil {
		return nil
	}
	if err := tryMagicActions(game, pid, enemies, magicIdx); err == nil {
		return nil
	}
	if err := trySpecialActions(game, pid); err == nil {
		return nil
	}

	if err := game.HandleAction(model.PlayerAction{PlayerID: pid, Type: model.CmdCannotAct}); err == nil {
		return nil
	}
	// 回归兜底：若自动策略无法找到任何合法输入，强制结束本次行动阶段，避免整局回归中断。
	// 该分支仅用于测试自动驾驶，不影响正式对局逻辑。
	game.Log(fmt.Sprintf(
		"[Auto] %s 无可执行动作（Phase=%s Extra=%s），强制结束行动阶段",
		player.Name, game.State.Phase, player.TurnState.CurrentExtraAction,
	))
	player.TurnState.CurrentExtraAction = ""
	player.TurnState.CurrentExtraElement = nil
	game.State.Phase = model.PhaseTurnEnd
	return nil
}

func tryScenarioActionPlan(
	game *engine.GameEngine,
	player *model.Player,
	enemies []string,
	attackIdx []int,
	magicIdx []int,
) (bool, error) {
	if currentDirectedScenarioPlan == nil || player == nil {
		return false, nil
	}
	order := scenarioActionOrderForPlayer(game, player)
	if len(order) == 0 {
		return false, nil
	}

	for _, action := range order {
		if !scenarioActionAllowedByExtra(player.TurnState.CurrentExtraAction, action) {
			continue
		}

		switch action {
		case autoPlanActionSkill:
			ok, err := tryScenarioActionSkills(game, player, enemies)
			if err != nil {
				return false, err
			}
			if ok {
				return true, nil
			}
		case autoPlanActionAttack:
			if err := tryAttackActions(game, player.ID, enemies, attackIdx); err == nil {
				return true, nil
			}
		case autoPlanActionMagic:
			if shouldSkipScenarioMagicAction(player) {
				continue
			}
			if err := tryMagicActions(game, player.ID, enemies, magicIdx); err == nil {
				return true, nil
			}
		case autoPlanActionBuy:
			if err := game.HandleAction(model.PlayerAction{PlayerID: player.ID, Type: model.CmdBuy}); err == nil {
				return true, nil
			}
		case autoPlanActionSynthesize:
			if err := game.HandleAction(model.PlayerAction{PlayerID: player.ID, Type: model.CmdSynthesize}); err == nil {
				return true, nil
			}
		case autoPlanActionExtract:
			if err := game.HandleAction(model.PlayerAction{PlayerID: player.ID, Type: model.CmdExtract}); err == nil {
				return true, nil
			}
		case autoPlanActionSpecial:
			if err := trySpecialActions(game, player.ID); err == nil {
				return true, nil
			}
		}
	}

	return false, nil
}

func scenarioActionAllowedByExtra(extraAction string, action string) bool {
	switch extraAction {
	case "Attack":
		return action == autoPlanActionAttack
	case "Magic":
		return action == autoPlanActionMagic
	default:
		return true
	}
}

func scenarioActionOrderForPlayer(game *engine.GameEngine, player *model.Player) []string {
	if currentDirectedScenarioPlan == nil || player == nil {
		return nil
	}
	base := append([]string{}, currentDirectedScenarioPlan.RoleActionOrder[player.Role]...)
	if len(base) == 0 {
		return nil
	}

	// 第三轮：按“角色-回合-资源-目标”做前提驱动，聚焦难触发技能。
	if player.Role == "adventurer" {
		targetUnderground := scenarioTargetsSkill("地下法则")
		targetParadise := scenarioTargetsSkill("冒险者天堂")

		if targetUnderground {
			if canAttemptBuy(game, player) {
				base = prioritizeActions(base, autoPlanActionBuy)
			} else if hasAnyCombatCard(player) {
				// 买入前提是手牌<=3，先打牌压手牌。
				base = prioritizeActions(base, autoPlanActionAttack)
			}
		}
		if targetParadise && canAttemptExtract(game, player) {
			// 第四轮：当目标包含“冒险者天堂”时，允许更积极地切入提炼链路。
			if !targetUnderground || player.Gem+player.Crystal == 0 || campHasCrystalStock(game, player.Camp) {
				base = prioritizeActions(base, autoPlanActionExtract)
			}
		}
	}

	if player.Role == "berserker" && (scenarioTargetsSkill("撕裂") || scenarioTargetsSkill("血影狂刀")) {
		hasBloodBladeCard := hasAttackExclusiveCardForSkill(player, "血影狂刀")
		needGemForTear := scenarioTargetsSkill("撕裂") && player.Gem == 0
		if needGemForTear {
			if canAttemptExtract(game, player) {
				base = prioritizeActions(base, autoPlanActionExtract)
			}
			if canAttemptBuy(game, player) {
				base = prioritizeActions(base, autoPlanActionBuy)
			}
			// 资源未就绪前先延后攻击，避免把命中窗口浪费在“无宝石无法撕裂”的状态。
			base = deprioritizeActions(base, autoPlanActionAttack)
		} else if player.Gem == 0 {
			if canAttemptExtract(game, player) {
				base = prioritizeActions(base, autoPlanActionExtract)
			} else if canAttemptBuy(game, player) {
				base = prioritizeActions(base, autoPlanActionBuy)
			}
		}
		if scenarioTargetsSkill("血影狂刀") && !hasBloodBladeCard && canAttemptBuy(game, player) {
			base = prioritizeActions(base, autoPlanActionBuy)
		}
		if !needGemForTear {
			base = prioritizeActions(base, autoPlanActionAttack)
		}
		base = deprioritizeActions(base, autoPlanActionMagic)
	}

	if player.Role == "archer" {
		targetSnipe := scenarioTargetsSkill("狙击")
		targetPiercing := scenarioTargetsSkill("贯穿射击")
		attackCount := countHandType(player, model.CardTypeAttack)
		magicCount := countHandType(player, model.CardTypeMagic)

		if targetSnipe {
			if canAttemptActionSkillByID(player, "snipe") {
				base = prioritizeActions(base, autoPlanActionSkill)
			} else if canAttemptExtract(game, player) {
				base = prioritizeActions(base, autoPlanActionExtract)
			} else if canAttemptBuy(game, player) {
				base = prioritizeActions(base, autoPlanActionBuy)
			}
		}

		if targetPiercing {
			// 贯穿射击需要“主动攻击未命中”窗口，本体应优先多打攻击。
			base = prioritizeActions(base, autoPlanActionAttack)
			if attackCount == 0 {
				// 没有攻击牌时优先补牌，避免把法术牌都打掉导致贯穿弃牌资源不足。
				if canAttemptBuy(game, player) {
					base = prioritizeActions(base, autoPlanActionBuy)
				} else if canAttemptExtract(game, player) {
					base = prioritizeActions(base, autoPlanActionExtract)
				}
			}
			if magicCount == 0 {
				if canAttemptBuy(game, player) {
					base = prioritizeActions(base, autoPlanActionBuy)
				} else if canAttemptExtract(game, player) {
					base = prioritizeActions(base, autoPlanActionExtract)
				}
			}
			// 贯穿射击依赖法术弃牌，法术行动应作为最后手段。
			base = deprioritizeActions(base, autoPlanActionMagic)
		}
	}

	if player.Role == "blade_master" {
		if scenarioTargetsSkill("剑影") {
			if player.Crystal > 0 {
				base = prioritizeActions(base, autoPlanActionAttack)
			} else if canAttemptExtract(game, player) && campHasCrystalStock(game, player.Camp) {
				base = prioritizeActions(base, autoPlanActionExtract)
			} else if canAttemptBuy(game, player) {
				base = prioritizeActions(base, autoPlanActionBuy)
			} else if hasAnyCombatCard(player) {
				base = prioritizeActions(base, autoPlanActionAttack)
			}
		}
		if scenarioTargetsSkill("剑影") || scenarioTargetsSkill("烈风技") {
			base = prioritizeActions(base, autoPlanActionAttack)
		}
	}

	if scenarioTargetsSkill("剑影") &&
		player.Role != "blade_master" &&
		campHasRole(game, player.Camp, "blade_master") {
		// 剑影定向：队友尽快清手买入，给本阵营制造并保留蓝水晶。
		if canAttemptBuy(game, player) {
			base = prioritizeActions(base, autoPlanActionBuy)
		} else if hasAnyCombatCard(player) {
			base = prioritizeActions(base, autoPlanActionAttack)
		}
		base = deprioritizeActions(base, autoPlanActionExtract)
	}

	if player.Role == "holy_lancer" {
		if scenarioTargetsSkill("圣光祈愈") && player.Gem > 0 {
			base = prioritizeActions(base, autoPlanActionSkill)
		} else if scenarioTargetsSkill("圣光祈愈") && canAttemptExtract(game, player) {
			base = prioritizeActions(base, autoPlanActionExtract)
		}
	}

	if player.Role == "angel" {
		if scenarioTargetsSkill("天使之歌") {
			if player.Crystal > 0 {
				base = prioritizeActions(base, autoPlanActionSkill)
			} else if canAttemptExtract(game, player) && campHasCrystalStock(game, player.Camp) {
				base = prioritizeActions(base, autoPlanActionExtract)
			} else if canAttemptBuy(game, player) {
				base = prioritizeActions(base, autoPlanActionBuy)
			} else if hasAnyCombatCard(player) {
				base = prioritizeActions(base, autoPlanActionAttack)
			}
		}
		if scenarioTargetsSkill("神之庇护") && player.Crystal == 0 {
			if canAttemptExtract(game, player) && campHasCrystalStock(game, player.Camp) {
				base = prioritizeActions(base, autoPlanActionExtract)
			} else if canAttemptBuy(game, player) {
				base = prioritizeActions(base, autoPlanActionBuy)
			}
		}
	}

	if player.Role == "sealer" && (scenarioTargetsSkill("五系束缚") || scenarioTargetsSkill("封印破碎")) {
		targetBind := scenarioTargetsSkill("五系束缚")
		targetSealBreak := scenarioTargetsSkill("封印破碎")
		targetThunderSeal := scenarioTargetsSkill("雷之封印")
		hasBindExclusive := hasExclusiveCardForSkill(player, "五系束缚")
		canBindNow := targetBind && hasBindExclusive && canAttemptActionSkillByID(player, "five_elements_bind")
		canSealBreakNow := targetSealBreak && canAttemptActionSkillByID(player, "seal_break")
		hasFieldEffect := hasBasicFieldEffectOnBoard(game)

		if canSealBreakNow && hasFieldEffect {
			base = prioritizeActions(base, autoPlanActionSkill)
		} else if targetSealBreak && !hasFieldEffect {
			// 先做铺垫：让场上出现基础效果牌，再尝试封印破碎回收。
			if canAttemptBuy(game, player) {
				base = prioritizeActions(base, autoPlanActionBuy)
			}
			base = prioritizeActions(base, autoPlanActionAttack)
		}

		if canBindNow {
			base = prioritizeActions(base, autoPlanActionSkill)
		} else if targetBind && !hasBindExclusive {
			if canAttemptBuy(game, player) {
				base = prioritizeActions(base, autoPlanActionBuy)
			} else if hasAnyCombatCard(player) {
				base = prioritizeActions(base, autoPlanActionAttack)
			}
		} else if (targetSealBreak || targetThunderSeal) && player.Crystal == 0 && canAttemptExtract(game, player) {
			base = prioritizeActions(base, autoPlanActionExtract)
		} else if player.Crystal == 0 && canAttemptExtract(game, player) {
			base = prioritizeActions(base, autoPlanActionExtract)
		} else {
			base = prioritizeActions(base, autoPlanActionSkill)
		}
	}

	if player.Role == "saintess" && scenarioTargetsSkill("圣疗") {
		if canAttemptActionSkillByID(player, "saint_heal") {
			base = prioritizeActions(base, autoPlanActionSkill)
		} else if player.Crystal == 0 && campHasCrystalStock(game, player.Camp) && canAttemptExtract(game, player) {
			base = prioritizeActions(base, autoPlanActionExtract)
		} else if player.Crystal == 0 && canAttemptBuy(game, player) {
			base = prioritizeActions(base, autoPlanActionBuy)
		} else if canAttemptExtract(game, player) && campHasAnyResourceStock(game, player.Camp) {
			base = prioritizeActions(base, autoPlanActionExtract)
		} else if canAttemptBuy(game, player) {
			base = prioritizeActions(base, autoPlanActionBuy)
		}
		base = deprioritizeActions(base, autoPlanActionMagic)
		if player.Crystal == 0 {
			base = deprioritizeActions(base, autoPlanActionAttack)
		}
	}

	return base
}

func prioritizeActions(actions []string, preferred string) []string {
	if len(actions) == 0 || preferred == "" {
		return actions
	}
	out := make([]string, 0, len(actions))
	seen := false
	for _, a := range actions {
		if a == preferred {
			seen = true
			continue
		}
		out = append(out, a)
	}
	if seen {
		out = append([]string{preferred}, out...)
	}
	return out
}

func deprioritizeActions(actions []string, deferred string) []string {
	if len(actions) == 0 || deferred == "" {
		return actions
	}
	out := make([]string, 0, len(actions))
	seen := false
	for _, a := range actions {
		if a == deferred {
			seen = true
			continue
		}
		out = append(out, a)
	}
	if seen {
		out = append(out, deferred)
	}
	return out
}

func scenarioTargetsSkill(title string) bool {
	if currentDirectedScenarioPlan == nil || title == "" {
		return false
	}
	for _, t := range currentDirectedScenarioPlan.TargetSkillTitles {
		if t == title {
			return true
		}
	}
	return false
}

func canAttemptBuy(game *engine.GameEngine, player *model.Player) bool {
	if game == nil || player == nil {
		return false
	}
	if game.State.HasPerformedStartup {
		return false
	}
	maxHand := player.MaxHand
	if maxHand <= 0 {
		maxHand = 6
	}
	return len(player.Hand)+3 <= maxHand
}

func canAttemptExtract(game *engine.GameEngine, player *model.Player) bool {
	if game == nil || player == nil {
		return false
	}
	if game.State.HasPerformedStartup {
		return false
	}
	if player.Gem+player.Crystal >= 3 {
		return false
	}
	if player.Camp == model.RedCamp {
		return game.State.RedGems+game.State.RedCrystals > 0
	}
	return game.State.BlueGems+game.State.BlueCrystals > 0
}

func campHasAnyResourceStock(game *engine.GameEngine, camp model.Camp) bool {
	gems, crystals := campResourceStock(game, camp)
	return gems+crystals > 0
}

func campResourceStock(game *engine.GameEngine, camp model.Camp) (gems int, crystals int) {
	if game == nil {
		return 0, 0
	}
	if camp == model.RedCamp {
		return game.State.RedGems, game.State.RedCrystals
	}
	return game.State.BlueGems, game.State.BlueCrystals
}

func campHasCrystalStock(game *engine.GameEngine, camp model.Camp) bool {
	_, crystals := campResourceStock(game, camp)
	return crystals > 0
}

func hasAnyCombatCard(player *model.Player) bool {
	if player == nil {
		return false
	}
	return countHandType(player, model.CardTypeAttack) > 0 || countHandType(player, model.CardTypeMagic) > 0
}

func shouldSkipScenarioMagicAction(player *model.Player) bool {
	if player == nil {
		return false
	}
	if player.Role == "berserker" && (scenarioTargetsSkill("撕裂") || scenarioTargetsSkill("血影狂刀")) {
		attackCount := countHandType(player, model.CardTypeAttack)
		magicCount := countHandType(player, model.CardTypeMagic)
		if attackCount > 0 && magicCount > 0 {
			return true
		}
	}
	if scenarioTargetsSkill("贯穿射击") && player.Role != "archer" {
		// 贯穿射击窗口依赖防御方手中仍有法术牌可防御。
		magicCount := countHandType(player, model.CardTypeMagic)
		attackCount := countHandType(player, model.CardTypeAttack)
		if magicCount > 0 && attackCount > 0 {
			return true
		}
	}
	if player.Role == "archer" && scenarioTargetsSkill("贯穿射击") {
		// 贯穿射击依赖法术弃牌；在有攻击牌可打时，优先保留法术牌。
		magicCount := countHandType(player, model.CardTypeMagic)
		attackCount := countHandType(player, model.CardTypeAttack)
		if attackCount > 0 && magicCount > 0 {
			return true
		}
		return magicCount <= 1
	}
	return false
}

func tryScenarioActionSkills(game *engine.GameEngine, player *model.Player, enemies []string) (bool, error) {
	if player == nil || player.Character == nil {
		return false, nil
	}
	// 引擎限制：额外攻击/额外法术行动都不能发动技能。
	if player.TurnState.CurrentExtraAction != "" {
		return false, nil
	}

	if currentDirectedScenarioPlan != nil {
		priorityIDs := currentDirectedScenarioPlan.RoleActionSkillPriority[player.Role]
		for _, skillID := range priorityIDs {
			skill, ok := findSkillDefinition(player, skillID)
			if !ok || skill.Type != model.SkillTypeAction {
				continue
			}
			ok, err := trySingleActionSkill(game, player, skill, enemies)
			if err != nil {
				return false, err
			}
			if ok {
				return true, nil
			}
		}
	}

	return tryActionSkillsAggressive(game, player, enemies)
}

func trySingleActionSkill(
	game *engine.GameEngine,
	player *model.Player,
	skill model.SkillDefinition,
	enemies []string,
) (bool, error) {
	if !canPayActionSkillResource(player, skill) {
		return false, nil
	}

	targetSets := buildTargetSetsForSkill(game, player, skill, enemies)
	if len(targetSets) == 0 {
		return false, nil
	}

	discardSets := findSkillDiscardSets(player, skill)
	if len(discardSets) == 0 {
		return false, nil
	}

	for _, targetIDs := range targetSets {
		for _, discards := range discardSets {
			err := game.HandleAction(model.PlayerAction{
				PlayerID:   player.ID,
				Type:       model.CmdSkill,
				SkillID:    skill.ID,
				TargetIDs:  targetIDs,
				Selections: discards,
			})
			if err == nil {
				return true, nil
			}
		}
	}

	return false, nil
}

func canAttemptActionSkillByID(player *model.Player, skillID string) bool {
	if player == nil || skillID == "" {
		return false
	}
	skill, ok := findSkillDefinition(player, skillID)
	if !ok || skill.Type != model.SkillTypeAction {
		return false
	}
	if !canPayActionSkillResource(player, skill) {
		return false
	}
	return len(findSkillDiscardSets(player, skill)) > 0
}

func tryActionSkillsAggressive(game *engine.GameEngine, player *model.Player, enemies []string) (bool, error) {
	if player == nil || player.Character == nil {
		return false, nil
	}
	if player.TurnState.CurrentExtraAction == "Attack" {
		return false, nil
	}

	skills := sortedActionSkills(player.Character.Skills)
	if len(skills) == 0 {
		return false, nil
	}

	start := actionSkillCursor[player.ID] % len(skills)
	for offset := 0; offset < len(skills); offset++ {
		skill := skills[(start+offset)%len(skills)]
		ok, err := trySingleActionSkill(game, player, skill, enemies)
		if err != nil {
			return false, err
		}
		if ok {
			actionSkillCursor[player.ID] = (start + offset + 1) % len(skills)
			return true, nil
		}
	}

	return false, nil
}

func sortedActionSkills(skills []model.SkillDefinition) []model.SkillDefinition {
	out := make([]model.SkillDefinition, 0, len(skills))
	for _, s := range skills {
		if s.Type == model.SkillTypeAction {
			out = append(out, s)
		}
	}

	score := func(s model.SkillDefinition) int {
		val := 0
		if s.RequireExclusive {
			val += 4
		}
		if s.PlaceCard {
			val += 3
		}
		if s.CostCrystal > 0 || s.CostGem > 0 {
			val += 2
		}
		if s.CostDiscards > 0 {
			val += 1
		}
		return val
	}

	sort.SliceStable(out, func(i, j int) bool {
		si := score(out[i])
		sj := score(out[j])
		if si != sj {
			return si > sj
		}
		return out[i].Title < out[j].Title
	})

	return out
}

func canPaySkillResource(player *model.Player, skill model.SkillDefinition) bool {
	if player == nil {
		return false
	}
	// 与引擎技能费用规则对齐：
	// 1) 宝石消耗必须由宝石支付；
	// 2) 水晶消耗可由剩余宝石替代。
	if player.Gem < skill.CostGem {
		return false
	}
	needTotal := skill.CostGem + skill.CostCrystal
	return player.Gem+player.Crystal >= needTotal
}

func canPayActionSkillResource(player *model.Player, skill model.SkillDefinition) bool {
	if player == nil {
		return false
	}
	// 主动技与引擎 UseSkill 逻辑对齐：
	// 1) 宝石消耗必须由宝石支付；
	// 2) 水晶消耗可由“剩余宝石 + 水晶”共同覆盖。
	if player.Gem < skill.CostGem {
		return false
	}
	needTotal := skill.CostGem + skill.CostCrystal
	return player.Gem+player.Crystal >= needTotal
}

func buildTargetSetsForSkill(game *engine.GameEngine, user *model.Player, skill model.SkillDefinition, enemies []string) [][]string {
	if skill.TargetType == model.TargetNone {
		return [][]string{{}}
	}

	allies := collectAllyIDs(game, user.ID, false)
	alliesWithSelf := collectAllyIDs(game, user.ID, true)
	allPlayers := append([]string{}, game.State.PlayerOrder...)
	maxTargets := skill.MaxTargets
	if maxTargets <= 0 {
		maxTargets = 1
	}
	minTargets := skill.MinTargets
	if minTargets <= 0 {
		minTargets = 1
	}

	switch skill.TargetType {
	case model.TargetSelf:
		return [][]string{{user.ID}}
	case model.TargetEnemy:
		return pickTargetSetsByType(enemies, minTargets, maxTargets)
	case model.TargetAlly:
		return pickTargetSetsByType(allies, minTargets, maxTargets)
	case model.TargetAllySelf:
		return pickTargetSetsByType(alliesWithSelf, minTargets, maxTargets)
	case model.TargetAny, model.TargetSpecific:
		return pickTargetSetsByType(allPlayers, minTargets, maxTargets)
	default:
		return [][]string{{}}
	}
}

func pickTargetSetsByType(pool []string, minCount, maxCount int) [][]string {
	if len(pool) == 0 {
		return nil
	}
	if maxCount <= 0 {
		maxCount = 1
	}
	if minCount <= 0 {
		minCount = 1
	}
	if minCount > maxCount {
		minCount = maxCount
	}
	if maxCount > len(pool) {
		maxCount = len(pool)
	}

	results := make([][]string, 0)
	for c := maxCount; c >= minCount; c-- {
		remain := 12 - len(results)
		if remain <= 0 {
			break
		}
		if c == 1 {
			for _, id := range pool {
				results = append(results, []string{id})
				if len(results) >= 12 {
					break
				}
			}
			continue
		}
		results = append(results, combineStringsLimited(pool, c, remain)...)
	}
	return results
}

func combineStringsLimited(items []string, k, limit int) [][]string {
	if k <= 0 || len(items) < k || limit <= 0 {
		return nil
	}

	out := make([][]string, 0)
	var dfs func(start int, curr []string)
	dfs = func(start int, curr []string) {
		if len(out) >= limit {
			return
		}
		if len(curr) == k {
			cp := append([]string{}, curr...)
			out = append(out, cp)
			return
		}
		for i := start; i < len(items); i++ {
			curr = append(curr, items[i])
			dfs(i+1, curr)
			curr = curr[:len(curr)-1]
			if len(out) >= limit {
				return
			}
		}
	}
	dfs(0, nil)
	return out
}

func findSkillDiscardSets(player *model.Player, skill model.SkillDefinition) [][]int {
	need := skill.CostDiscards
	if need <= 0 {
		return [][]int{{}}
	}

	candidates := make([]int, 0)
	for idx, card := range player.Hand {
		if !cardMatchesSkillDiscard(player, skill, card) {
			continue
		}
		candidates = append(candidates, idx)
	}
	if len(candidates) < need {
		return nil
	}

	combos := combineIntsLimited(candidates, need, 10)
	if len(combos) == 0 {
		return nil
	}
	return combos
}

func cardMatchesSkillDiscard(player *model.Player, skill model.SkillDefinition, card model.Card) bool {
	if skill.DiscardType != "" && card.Type != skill.DiscardType {
		return false
	}
	if skill.DiscardElement != "" && card.Element != skill.DiscardElement {
		return false
	}
	if skill.DiscardFate != "" && card.Faction != skill.DiscardFate {
		return false
	}
	if skill.RequireExclusive {
		if player == nil || player.Character == nil {
			return false
		}
		if !card.MatchExclusive(player.Character.Name, skill.Title) {
			return false
		}
	}
	return true
}

func combineIntsLimited(items []int, k, limit int) [][]int {
	if k <= 0 || len(items) < k || limit <= 0 {
		return nil
	}

	out := make([][]int, 0)
	var dfs func(start int, curr []int)
	dfs = func(start int, curr []int) {
		if len(out) >= limit {
			return
		}
		if len(curr) == k {
			cp := append([]int{}, curr...)
			out = append(out, cp)
			return
		}
		for i := start; i < len(items); i++ {
			curr = append(curr, items[i])
			dfs(i+1, curr)
			curr = curr[:len(curr)-1]
			if len(out) >= limit {
				return
			}
		}
	}
	dfs(0, nil)
	return out
}

func collectEnemyIDs(game *engine.GameEngine, camp model.Camp) []string {
	enemies := make([]string, 0)
	for _, pid := range game.State.PlayerOrder {
		p := game.State.Players[pid]
		if p != nil && p.Camp != camp {
			enemies = append(enemies, pid)
		}
	}
	return enemies
}

func collectAllyIDs(game *engine.GameEngine, userID string, includeSelf bool) []string {
	user := game.State.Players[userID]
	if user == nil {
		return nil
	}
	allies := make([]string, 0)
	for _, pid := range game.State.PlayerOrder {
		p := game.State.Players[pid]
		if p == nil || p.Camp != user.Camp {
			continue
		}
		if !includeSelf && pid == userID {
			continue
		}
		allies = append(allies, pid)
	}
	return allies
}

func orderScenarioTargetsForAttack(game *engine.GameEngine, attacker *model.Player, enemies []string) []string {
	if game == nil || attacker == nil || len(enemies) == 0 {
		return enemies
	}
	if currentDirectedScenarioPlan == nil {
		return enemies
	}

	priorityScore := func(targetID string) int {
		target := game.State.Players[targetID]
		if target == nil {
			return 0
		}
		score := 0

		if attacker.Role == "blade_master" && scenarioTargetsSkill("烈风技") && playerHasShield(target) {
			score += 4
		}
		if attacker.Role == "archer" && scenarioTargetsSkill("贯穿射击") {
			if hasCardType(target, model.CardTypeMagic) {
				score += 3
			}
			if len(target.Hand) >= 4 {
				score += 1
			}
		}
		if attacker.Role == "holy_lancer" && scenarioTargetsSkill("圣疗") && target.Heal > 0 {
			score += 1
		}
		if attacker.Role == "berserker" && scenarioTargetsSkill("血影狂刀") {
			if len(target.Hand) == 2 {
				score += 6
			} else if len(target.Hand) == 3 {
				score += 5
			} else if len(target.Hand) == 1 {
				score += 1
			}
			if playerHasShield(target) {
				score -= 2
			}
		}

		return score
	}

	sort.SliceStable(enemies, func(i, j int) bool {
		si := priorityScore(enemies[i])
		sj := priorityScore(enemies[j])
		if si != sj {
			return si > sj
		}
		return enemies[i] < enemies[j]
	})
	return enemies
}

func orderScenarioTargetsForMagic(game *engine.GameEngine, caster *model.Player, enemies []string) []string {
	if game == nil || caster == nil || len(enemies) == 0 {
		return enemies
	}
	if currentDirectedScenarioPlan == nil {
		return enemies
	}

	if scenarioTargetsSkill("神之庇护") && (caster.Role == "magical_girl" || caster.Role == "elementalist") {
		sort.SliceStable(enemies, func(i, j int) bool {
			pi := game.State.Players[enemies[i]]
			pj := game.State.Players[enemies[j]]
			si := 0
			sj := 0
			if pi != nil && pi.Role == "angel" {
				si += 3
			}
			if pj != nil && pj.Role == "angel" {
				sj += 3
			}
			if pi != nil && len(pi.Hand) >= pi.MaxHand-1 {
				si += 1
			}
			if pj != nil && len(pj.Hand) >= pj.MaxHand-1 {
				sj += 1
			}
			if si != sj {
				return si > sj
			}
			return enemies[i] < enemies[j]
		})
	}
	return enemies
}

func playerHasShield(player *model.Player) bool {
	if player == nil {
		return false
	}
	for _, fc := range player.Field {
		if fc.Mode == model.FieldEffect && fc.Effect == model.EffectShield {
			return true
		}
	}
	return false
}

func hasCardType(player *model.Player, cardType model.CardType) bool {
	if player == nil {
		return false
	}
	for _, c := range player.Hand {
		if c.Type == cardType {
			return true
		}
	}
	return false
}

func countHandType(player *model.Player, cardType model.CardType) int {
	if player == nil {
		return 0
	}
	count := 0
	for _, c := range player.Hand {
		if c.Type == cardType {
			count++
		}
	}
	return count
}

func hasExclusiveCardForSkill(player *model.Player, skillTitle string) bool {
	if player == nil || player.Character == nil || skillTitle == "" {
		return false
	}
	for _, c := range player.ExclusiveCards {
		if c.MatchExclusive(player.Character.Name, skillTitle) {
			return true
		}
	}
	for _, c := range player.Hand {
		if c.MatchExclusive(player.Character.Name, skillTitle) {
			return true
		}
	}
	return false
}

func hasAttackExclusiveCardForSkill(player *model.Player, skillTitle string) bool {
	if player == nil || player.Character == nil || skillTitle == "" {
		return false
	}
	for _, c := range player.Hand {
		if c.Type != model.CardTypeAttack {
			continue
		}
		if c.MatchExclusive(player.Character.Name, skillTitle) {
			return true
		}
	}
	return false
}

func collectPlayableCards(player *model.Player) (attackIdx []int, magicIdx []int) {
	attackIdx = make([]int, 0)
	magicIdx = make([]int, 0)
	handCount := len(player.Hand)

	playableCardAt := func(playIdx int) (model.Card, bool) {
		if playIdx < 0 {
			return model.Card{}, false
		}
		if playIdx < handCount {
			return player.Hand[playIdx], true
		}
		bIdx := playIdx - handCount
		if bIdx >= 0 && bIdx < len(player.Blessings) {
			return player.Blessings[bIdx], true
		}
		return model.Card{}, false
	}

	for idx, card := range player.Hand {
		if !matchesExtraElement(player.TurnState.CurrentExtraElement, card.Element) {
			continue
		}
		if card.Type == model.CardTypeAttack {
			attackIdx = append(attackIdx, idx)
		}
		if card.Type == model.CardTypeMagic {
			magicIdx = append(magicIdx, idx)
		}
	}
	for idx, card := range player.Blessings {
		playIdx := handCount + idx
		if !matchesExtraElement(player.TurnState.CurrentExtraElement, card.Element) {
			continue
		}
		if card.Type == model.CardTypeAttack {
			attackIdx = append(attackIdx, playIdx)
		}
		if card.Type == model.CardTypeMagic {
			magicIdx = append(magicIdx, playIdx)
		}
	}

	// 高伤害优先，尽快推进局势并制造更多响应机会。
	if player != nil && player.Role == "berserker" && scenarioTargetsSkill("血影狂刀") && player.Character != nil {
		sort.SliceStable(attackIdx, func(i, j int) bool {
			ci, okI := playableCardAt(attackIdx[i])
			cj, okJ := playableCardAt(attackIdx[j])
			if !okI || !okJ {
				return attackIdx[i] < attackIdx[j]
			}
			ei := ci.MatchExclusive(player.Character.Name, "血影狂刀")
			ej := cj.MatchExclusive(player.Character.Name, "血影狂刀")
			if ei != ej {
				return ei
			}
			return ci.Damage > cj.Damage
		})
	} else {
		sort.SliceStable(attackIdx, func(i, j int) bool {
			ci, okI := playableCardAt(attackIdx[i])
			cj, okJ := playableCardAt(attackIdx[j])
			if !okI || !okJ {
				return attackIdx[i] < attackIdx[j]
			}
			return ci.Damage > cj.Damage
		})
	}
	sort.SliceStable(magicIdx, func(i, j int) bool {
		ci, okI := playableCardAt(magicIdx[i])
		cj, okJ := playableCardAt(magicIdx[j])
		if !okI || !okJ {
			return magicIdx[i] < magicIdx[j]
		}
		return ci.Damage > cj.Damage
	})

	return attackIdx, magicIdx
}

func tryAttackActions(game *engine.GameEngine, pid string, enemies []string, attackIdx []int) error {
	if len(attackIdx) == 0 {
		return fmt.Errorf("no attack card")
	}
	orderedAttackIdx := append([]int{}, attackIdx...)
	if currentDirectedScenarioPlan != nil {
		if player := game.State.Players[pid]; player != nil && player.Role == "archer" && scenarioTargetsSkill("贯穿射击") {
			// 贯穿射击依赖“攻击可响应并未命中”，优先非雷系攻击（避免闪电箭锁响应）。
			sort.SliceStable(orderedAttackIdx, func(i, j int) bool {
				ci := player.Hand[orderedAttackIdx[i]]
				cj := player.Hand[orderedAttackIdx[j]]
				ii := ci.Element == model.ElementThunder
				jj := cj.Element == model.ElementThunder
				if ii != jj {
					return !ii
				}
				return ci.Damage > cj.Damage
			})
		}
	}

	var lastErr error
	player := game.State.Players[pid]
	for _, idx := range orderedAttackIdx {
		targetOrder := append([]string{}, enemies...)
		if player != nil {
			targetOrder = orderScenarioTargetsForAttack(game, player, targetOrder)
			if player.Role == "archer" && scenarioTargetsSkill("贯穿射击") &&
				idx >= 0 && idx < len(player.Hand) {
				targetOrder = orderPiercingTargetsForCard(game, targetOrder, player.Hand[idx].Element)
			}
		}
		for _, targetID := range targetOrder {
			err := game.HandleAction(model.PlayerAction{
				PlayerID:  pid,
				Type:      model.CmdAttack,
				TargetID:  targetID,
				CardIndex: idx,
			})
			if err == nil {
				return nil
			}
			lastErr = err
		}
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("no attack attempt was made")
	}
	return lastErr
}

func orderPiercingTargetsForCard(game *engine.GameEngine, enemies []string, attackElement model.Element) []string {
	if game == nil || len(enemies) == 0 {
		return enemies
	}
	out := append([]string{}, enemies...)
	scoreTarget := func(targetID string) int {
		target := game.State.Players[targetID]
		if target == nil {
			return -100
		}
		score := 0
		if hasCardType(target, model.CardTypeMagic) {
			score += 8
		}
		if hasCounterCardForElement(target, attackElement) {
			score += 4
		}
		if playerHasShield(target) {
			score -= 4
		}
		if len(target.Hand) >= 4 {
			score += 1
		}
		return score
	}
	sort.SliceStable(out, func(i, j int) bool {
		si := scoreTarget(out[i])
		sj := scoreTarget(out[j])
		if si != sj {
			return si > sj
		}
		return out[i] < out[j]
	})
	return out
}

func hasCounterCardForElement(player *model.Player, attackElement model.Element) bool {
	if player == nil || attackElement == "" {
		return false
	}
	for _, card := range player.Hand {
		if card.Type != model.CardTypeAttack {
			continue
		}
		if card.Element == attackElement || card.Element == model.ElementDark {
			return true
		}
	}
	return false
}

func trySpecialActions(game *engine.GameEngine, pid string) error {
	actions := []model.PlayerActionType{
		model.CmdBuy,
		model.CmdSynthesize,
		model.CmdExtract,
	}
	var lastErr error
	for _, actionType := range actions {
		err := game.HandleAction(model.PlayerAction{
			PlayerID: pid,
			Type:     actionType,
		})
		if err == nil {
			return nil
		}
		lastErr = err
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("no special action attempted")
	}
	return lastErr
}

func tryMagicActions(game *engine.GameEngine, pid string, enemies []string, magicIdx []int) error {
	if len(magicIdx) == 0 {
		return fmt.Errorf("no magic card")
	}
	targetOrder := append([]string{}, enemies...)
	if player := game.State.Players[pid]; player != nil {
		targetOrder = orderScenarioTargetsForMagic(game, player, targetOrder)
	}

	var lastErr error
	for _, idx := range magicIdx {
		for _, targetID := range targetOrder {
			err := game.HandleAction(model.PlayerAction{
				PlayerID:  pid,
				Type:      model.CmdMagic,
				TargetID:  targetID,
				CardIndex: idx,
			})
			if err == nil {
				return nil
			}
			lastErr = err
		}
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("no magic attempt was made")
	}
	return lastErr
}

func matchesExtraElement(allowed []model.Element, element model.Element) bool {
	if len(allowed) == 0 {
		return true
	}
	for _, e := range allowed {
		if e == element {
			return true
		}
	}
	return false
}

func resolveInterrupt(game *engine.GameEngine) error {
	intr := game.State.PendingInterrupt
	if intr == nil {
		return nil
	}
	pid := intr.PlayerID

	switch intr.Type {
	case model.InterruptMagicMissile:
		return game.HandleAction(model.PlayerAction{
			PlayerID:  pid,
			Type:      model.CmdRespond,
			ExtraArgs: []string{"take"},
		})

	case model.InterruptResponseSkill, model.InterruptStartupSkill:
		skillIdx := chooseInterruptSkillIndex(game, intr)
		if skillIdx >= 0 {
			if err := game.HandleAction(model.PlayerAction{PlayerID: pid, Type: model.CmdSelect, Selections: []int{skillIdx}}); err == nil {
				return nil
			}
		}
		return game.HandleAction(model.PlayerAction{PlayerID: pid, Type: model.CmdCancel})

	case model.InterruptDiscard, model.InterruptGiveCards:
		var selections []int
		var err error
		if intr.Type == model.InterruptDiscard {
			selections, err = pickDiscardSelections(game, game.GetCurrentPrompt())
			if errors.Is(err, errDiscardPrereqNotMet) {
				return game.HandleAction(model.PlayerAction{PlayerID: pid, Type: model.CmdCancel})
			}
		} else {
			selections, err = pickCardSelections(game.GetCurrentPrompt())
		}
		if err != nil {
			return err
		}
		return game.HandleAction(model.PlayerAction{PlayerID: pid, Type: model.CmdSelect, Selections: selections})

	case model.InterruptChoice,
		model.InterruptMagicBulletDirection,
		model.InterruptHolySwordDraw,
		model.InterruptSaintHeal,
		model.InterruptMagicBulletFusion:
		prompt := game.GetCurrentPrompt()
		selections, err := chooseInterruptSelections(game, intr, prompt)
		if err != nil {
			return game.HandleAction(model.PlayerAction{PlayerID: pid, Type: model.CmdCancel})
		}
		return game.HandleAction(model.PlayerAction{PlayerID: pid, Type: model.CmdSelect, Selections: selections})

	case model.InterruptMagicBlast:
		prompt := game.GetCurrentPrompt()
		if prompt == nil || len(prompt.Options) == 0 {
			return game.HandleAction(model.PlayerAction{PlayerID: pid, Type: model.CmdCancel})
		}
		selections, err := pickCardSelections(prompt)
		if err != nil {
			return game.HandleAction(model.PlayerAction{PlayerID: pid, Type: model.CmdCancel})
		}
		return game.HandleAction(model.PlayerAction{PlayerID: pid, Type: model.CmdSelect, Selections: selections})
	}

	if prompt := game.GetCurrentPrompt(); prompt != nil && len(prompt.Options) > 0 {
		selections, err := pickCardSelections(prompt)
		if err == nil {
			return game.HandleAction(model.PlayerAction{PlayerID: pid, Type: model.CmdSelect, Selections: selections})
		}
		return game.HandleAction(model.PlayerAction{PlayerID: pid, Type: model.CmdSelect, Selections: []int{0}})
	}
	return game.HandleAction(model.PlayerAction{PlayerID: pid, Type: model.CmdCancel})
}

func chooseInterruptSkillIndex(game *engine.GameEngine, intr *model.Interrupt) int {
	if game == nil || intr == nil || len(intr.SkillIDs) == 0 {
		return -1
	}
	player := game.State.Players[intr.PlayerID]
	if player == nil || player.Character == nil {
		return -1
	}

	if currentDirectedScenarioPlan != nil {
		if priorityIDs := currentDirectedScenarioPlan.RoleInterruptSkillPriority[player.Role]; len(priorityIDs) > 0 {
			for _, preferredID := range priorityIDs {
				for idx, skillID := range intr.SkillIDs {
					if skillID != preferredID {
						continue
					}
					if shouldSkipInterruptSkillInScenario(player, skillID) {
						continue
					}
					skill, ok := findSkillDefinition(player, skillID)
					if !ok {
						continue
					}
					if !canPaySkillResource(player, skill) {
						continue
					}
					if !isSkillDiscardFeasible(player, skill) {
						continue
					}
					return idx
				}
			}
		}
	}

	if player.Role == "magic_bow" {
		priorityIDs := []string{"mb_charge", "mb_demon_eye", "mb_magic_pierce", "mb_multi_shot"}
		for _, preferredID := range priorityIDs {
			for idx, skillID := range intr.SkillIDs {
				if skillID != preferredID {
					continue
				}
				skill, ok := findSkillDefinition(player, skillID)
				if !ok {
					continue
				}
				if !canPaySkillResource(player, skill) {
					continue
				}
				if !isSkillDiscardFeasible(player, skill) {
					continue
				}
				return idx
			}
		}
	}

	for idx, skillID := range intr.SkillIDs {
		if shouldSkipInterruptSkillInScenario(player, skillID) {
			continue
		}
		skill, ok := findSkillDefinition(player, skillID)
		if !ok {
			continue
		}
		if !canPaySkillResource(player, skill) {
			continue
		}
		if !isSkillDiscardFeasible(player, skill) {
			continue
		}
		return idx
	}
	return -1
}

func shouldSkipInterruptSkillInScenario(player *model.Player, skillID string) bool {
	if currentDirectedScenarioPlan == nil || player == nil || skillID == "" {
		return false
	}

	if player.Role == "archer" && scenarioTargetsSkill("贯穿射击") {
		// 为了制造“主动攻击未命中”窗口，避免在该场景里优先把攻击改为强制命中。
		if skillID == "precise_shot" {
			return true
		}
	}

	return false
}

func findSkillDefinition(player *model.Player, skillID string) (model.SkillDefinition, bool) {
	if player == nil || player.Character == nil {
		return model.SkillDefinition{}, false
	}
	for _, skill := range player.Character.Skills {
		if skill.ID == skillID {
			return skill, true
		}
	}
	return model.SkillDefinition{}, false
}

func isSkillDiscardFeasible(player *model.Player, skill model.SkillDefinition) bool {
	if skill.InteractionType != model.InteractionDiscard {
		return true
	}
	if skill.ID == "water_shadow" {
		for _, c := range player.Hand {
			if c.Element == model.ElementWater {
				return true
			}
		}
		return false
	}
	need := skill.InteractionConfig.MinSelect
	if need <= 0 {
		need = skill.CostDiscards
	}
	if need <= 0 {
		need = 1
	}

	count := 0
	for _, card := range player.Hand {
		if skill.DiscardType != "" && card.Type != skill.DiscardType {
			continue
		}
		if skill.DiscardElement != "" && card.Element != skill.DiscardElement {
			continue
		}
		count++
	}
	return count >= need
}

func pickCardSelections(prompt *model.Prompt) ([]int, error) {
	if prompt == nil {
		return nil, fmt.Errorf("missing prompt for card selection")
	}

	need := prompt.Min
	if need < 0 {
		need = 0
	}
	if need == 0 {
		return []int{}, nil
	}

	selections := make([]int, 0, need)
	for _, opt := range prompt.Options {
		idx, err := strconv.Atoi(opt.ID)
		if err != nil {
			continue
		}
		selections = append(selections, idx)
		if len(selections) == need {
			return selections, nil
		}
	}

	return nil, fmt.Errorf("not enough selectable options: required=%d got=%d", need, len(selections))
}

func pickDiscardSelections(game *engine.GameEngine, prompt *model.Prompt) ([]int, error) {
	if prompt == nil {
		return nil, fmt.Errorf("missing prompt for discard selection")
	}
	if game == nil || game.State.PendingInterrupt == nil {
		return pickCardSelections(prompt)
	}

	need := prompt.Min
	if need < 0 {
		need = 0
	}
	if need == 0 {
		return []int{}, nil
	}

	intr := game.State.PendingInterrupt
	ctx, _ := intr.Context.(map[string]interface{})
	skillID, _ := ctx["skill_id"].(string)
	player := game.State.Players[intr.PlayerID]
	if player == nil {
		return nil, fmt.Errorf("missing player for discard selection")
	}

	allowed := make(map[int]struct{}, len(prompt.Options))
	for _, opt := range prompt.Options {
		idx, err := strconv.Atoi(opt.ID)
		if err != nil {
			continue
		}
		allowed[idx] = struct{}{}
	}

	selectByMatch := func(match func(model.Card) bool) ([]int, error) {
		selections := make([]int, 0, need)
		for idx, card := range player.Hand {
			if _, ok := allowed[idx]; !ok {
				continue
			}
			if match(card) {
				selections = append(selections, idx)
				if len(selections) == need {
					return selections, nil
				}
			}
		}
		if len(selections) < need {
			return nil, errDiscardPrereqNotMet
		}
		return selections, nil
	}

	if skillID == "water_shadow" {
		return selectByMatch(func(card model.Card) bool {
			return card.Element == model.ElementWater
		})
	}

	discardType, _ := ctx["discard_type"].(model.CardType)
	discardElement, _ := ctx["discard_element"].(model.Element)

	match := func(card model.Card) bool {
		if discardType != "" && card.Type != discardType {
			return false
		}
		if discardElement != "" && card.Element != discardElement {
			return false
		}
		return true
	}

	if skillID != "" {
		if skill, ok := findSkillDefinition(player, skillID); ok {
			return selectByMatch(func(card model.Card) bool {
				return cardMatchesSkillDiscard(player, skill, card)
			})
		}
		return selectByMatch(match)
	}

	return pickCardSelections(prompt)
}

func chooseInterruptSelections(game *engine.GameEngine, intr *model.Interrupt, prompt *model.Prompt) ([]int, error) {
	if intr == nil {
		return nil, fmt.Errorf("missing interrupt")
	}
	if prompt == nil {
		return []int{0}, nil
	}

	switch intr.Type {
	case model.InterruptMagicBulletFusion:
		// 定向回归优先触发【魔弹融合】日志链路。
		return []int{0}, nil
	case model.InterruptMagicBulletDirection:
		// 定向回归优先触发【魔弹掌控】日志链路（逆向传递）。
		if scenarioTargetsSkill("魔弹掌控") {
			if len(prompt.Options) > 1 {
				return []int{1}, nil
			}
		}
		return []int{0}, nil
	case model.InterruptHolySwordDraw:
		// 避免过度摸牌导致手牌爆满，优先小X。
		if len(prompt.Options) > 1 {
			return []int{1}, nil
		}
		return []int{0}, nil
	case model.InterruptChoice:
		return chooseChoiceInterruptSelections(game, intr, prompt)
	case model.InterruptSaintHeal:
		// 圣疗默认选择第一个目标，交互链路可走通即可。
		return []int{0}, nil
	default:
		return []int{0}, nil
	}
}

func chooseChoiceInterruptSelections(game *engine.GameEngine, intr *model.Interrupt, prompt *model.Prompt) ([]int, error) {
	choiceType := interruptChoiceType(intr)
	player := game.State.Players[intr.PlayerID]

	switch choiceType {
	case "mb_magic_pierce_hit_confirm":
		// 魔贯冲击命中后默认吃满收益。
		return []int{0}, nil
	case "mb_charge_draw_x":
		// 优先摸到可转化为充能的手牌，避免 X=0 导致链路断开。
		if len(prompt.Options) >= 4 {
			return []int{3}, nil
		}
		if len(prompt.Options) > 0 {
			return []int{len(prompt.Options) - 1}, nil
		}
		return []int{0}, nil
	case "mb_charge_place_count":
		// 优先放置更多充能，提升后续技能触发概率。
		if len(prompt.Options) > 1 {
			return []int{len(prompt.Options) - 1}, nil
		}
		return []int{0}, nil
	case "mb_charge_place_cards", "mb_demon_eye_charge_card":
		return []int{0}, nil
	case "mb_thunder_scatter_extra":
		// 至少额外移除1个（若可选）以覆盖额外伤害分支。
		if len(prompt.Options) > 1 {
			return []int{1}, nil
		}
		return []int{0}, nil
	case "mb_demon_eye_mode":
		// 优先摸3张，保证后续“作为充能”可执行。
		if len(prompt.Options) > 1 {
			return []int{1}, nil
		}
		return []int{0}, nil
	case "ss_recall_pick":
		// 灵魂召还：至少先选1张法术牌，再结束选择。
		// Prompt 第0项是“完成选择并结算”，其余才是可选牌。
		selectedCount := 0
		if intr != nil {
			if data, ok := intr.Context.(map[string]interface{}); ok {
				if arr, ok := data["selected_indices"].([]int); ok {
					selectedCount = len(arr)
				} else if arr, ok := data["selected_indices"].([]interface{}); ok {
					selectedCount = len(arr)
				}
			}
		}
		if selectedCount == 0 && len(prompt.Options) > 1 {
			return []int{1}, nil
		}
		return []int{0}, nil
	case "buy_resource":
		if scenarioTargetsSkill("圣疗") && player != nil {
			if player.Role == "saintess" {
				if idx := findPromptOptionIndex(prompt, "水晶"); idx >= 0 {
					return []int{idx}, nil
				}
			} else if idx := findPromptOptionIndex(prompt, "宝石"); idx >= 0 {
				return []int{idx}, nil
			}
		}
		if scenarioTargetsSkill("撕裂") && player != nil && player.Role == "berserker" {
			if idx := findPromptOptionIndex(prompt, "宝石"); idx >= 0 {
				return []int{idx}, nil
			}
		}
		if scenarioTargetsSkill("剑影") && player != nil && game != nil && campHasRole(game, player.Camp, "blade_master") {
			// 剑影定向：优先确保本阵营至少有1个蓝水晶给剑士提炼。
			_, crystals := campResourceStock(game, player.Camp)
			if crystals == 0 {
				if idx := findPromptOptionIndex(prompt, "水晶"); idx >= 0 {
					return []int{idx}, nil
				}
			}
			if player.Role != "blade_master" {
				if idx := findPromptOptionIndex(prompt, "宝石"); idx >= 0 {
					return []int{idx}, nil
				}
			}
		}
		// 资源导向：二轮场景中多数目标技能依赖水晶，优先补水晶。
		if player != nil && rolePrefersCrystal(player.Role) && len(prompt.Options) > 1 {
			return []int{1}, nil
		}
		return []int{0}, nil
	case "extract":
		return chooseExtractSelections(player, prompt), nil
	case "holy_lancer_earth_spear_x":
		// 地枪取最大 X，提升命中后收敛速度。
		if len(prompt.Options) > 0 {
			return []int{len(prompt.Options) - 1}, nil
		}
		return []int{0}, nil
	case "adventurer_steal_sky_extra_action":
		// 冒险家定向回归优先攻击链路，制造响应窗口。
		return []int{0}, nil
	default:
		return []int{0}, nil
	}
}

func findPromptOptionIndex(prompt *model.Prompt, keyword string) int {
	if prompt == nil || keyword == "" {
		return -1
	}
	for idx, opt := range prompt.Options {
		if strings.Contains(opt.Label, keyword) {
			return idx
		}
	}
	return -1
}

func interruptChoiceType(intr *model.Interrupt) string {
	if intr == nil {
		return ""
	}
	data, ok := intr.Context.(map[string]interface{})
	if !ok {
		return ""
	}
	choiceType, _ := data["choice_type"].(string)
	return choiceType
}

func chooseExtractSelections(player *model.Player, prompt *model.Prompt) []int {
	if prompt == nil || len(prompt.Options) == 0 {
		return []int{0}
	}

	minSel := prompt.Min
	if minSel < 1 {
		minSel = 1
	}
	maxSel := prompt.Max
	if maxSel < minSel {
		maxSel = minSel
	}
	if maxSel > len(prompt.Options) {
		maxSel = len(prompt.Options)
	}

	preferCrystal := false
	preferGem := false
	if player != nil {
		preferCrystal = rolePrefersCrystal(player.Role)
		preferGem = rolePrefersGem(player.Role)
	}
	if player != nil {
		// 定向场景里，非目标角色优先抽走红宝石，尽量把蓝水晶留给目标角色。
		if scenarioTargetsSkill("剑影") && player.Role != "blade_master" {
			preferCrystal = false
			preferGem = true
		}
		if scenarioTargetsSkill("天使之歌") && player.Role != "angel" {
			preferCrystal = false
			preferGem = true
		}
		if scenarioTargetsSkill("撕裂") && player.Role == "berserker" {
			preferCrystal = false
			preferGem = true
		}
		if scenarioTargetsSkill("圣疗") {
			if player.Role == "saintess" {
				preferCrystal = true
				preferGem = false
			} else {
				// 圣疗定向：队友减少对水晶的消耗，让圣女更容易提炼到水晶。
				preferCrystal = false
				preferGem = true
			}
		}
	}

	crystalIdx := make([]int, 0)
	gemIdx := make([]int, 0)
	otherIdx := make([]int, 0)
	for i, opt := range prompt.Options {
		switch {
		case strings.Contains(opt.Label, "蓝水晶"):
			crystalIdx = append(crystalIdx, i)
		case strings.Contains(opt.Label, "红宝石"):
			gemIdx = append(gemIdx, i)
		default:
			otherIdx = append(otherIdx, i)
		}
	}

	selections := make([]int, 0, maxSel)
	appendFrom := func(src []int) {
		for _, idx := range src {
			if len(selections) >= maxSel {
				return
			}
			selections = append(selections, idx)
		}
	}

	if preferCrystal {
		appendFrom(crystalIdx)
		appendFrom(gemIdx)
	} else if preferGem {
		appendFrom(gemIdx)
		appendFrom(crystalIdx)
	} else {
		appendFrom(crystalIdx)
		appendFrom(gemIdx)
	}
	appendFrom(otherIdx)

	if len(selections) < minSel {
		for i := 0; i < len(prompt.Options) && len(selections) < minSel; i++ {
			exists := false
			for _, s := range selections {
				if s == i {
					exists = true
					break
				}
			}
			if !exists {
				selections = append(selections, i)
			}
		}
	}

	return selections
}

func rolePrefersCrystal(roleID string) bool {
	if roleID == "" {
		return false
	}
	switch roleID {
	case "angel", "sealer", "blade_master", "archer", "adventurer", "arbiter", "magic_bow", "magic_lancer":
		return true
	default:
		return false
	}
}

func rolePrefersGem(roleID string) bool {
	if roleID == "" {
		return false
	}
	switch roleID {
	case "holy_lancer", "berserker", "saintess":
		return true
	default:
		return false
	}
}

func countMagicBowChargeCards(player *model.Player) int {
	if player == nil {
		return 0
	}
	count := 0
	for _, fc := range player.Field {
		if fc.Mode != model.FieldCover {
			continue
		}
		if fc.Effect != model.EffectMagicBowCharge {
			continue
		}
		count++
	}
	return count
}

func resolveCombatAsTake(game *engine.GameEngine) error {
	if len(game.State.CombatStack) == 0 {
		return fmt.Errorf("combat stack is empty")
	}
	top := game.State.CombatStack[len(game.State.CombatStack)-1]
	defender := game.State.Players[top.TargetID]
	attacker := game.State.Players[top.AttackerID]
	prompt := game.GetCurrentPrompt()

	if shouldPreferMissResponse(game, attacker, defender) {
		tryDefend := func() bool {
			if idx, ok := findDefendCardIndex(defender); ok {
				if err := game.HandleAction(model.PlayerAction{
					PlayerID:  top.TargetID,
					Type:      model.CmdRespond,
					CardIndex: idx,
					ExtraArgs: []string{"defend"},
				}); err == nil {
					return true
				}
			}
			return false
		}
		tryCounter := func() bool {
			if !top.CanBeResponded {
				return false
			}
			attackElement := top.Card.Element
			counterTargets := []string{}
			if prompt != nil && len(prompt.CounterTargetIDs) > 0 {
				counterTargets = append(counterTargets, prompt.CounterTargetIDs...)
			}
			if len(counterTargets) == 0 {
				counterTargets = collectCounterTargetIDs(game, top.AttackerID)
			}
			if idx, ok := findCounterCardIndex(defender, attackElement); ok {
				targetID := chooseCounterTargetID(counterTargets)
				if targetID != "" {
					if err := game.HandleAction(model.PlayerAction{
						PlayerID:  top.TargetID,
						Type:      model.CmdRespond,
						CardIndex: idx,
						TargetID:  targetID,
						ExtraArgs: []string{"counter"},
					}); err == nil {
						return true
					}
				}
			}
			return false
		}

		if shouldPreferCounterForCrystal(game, attacker, defender) {
			if tryCounter() {
				return nil
			}
			if tryDefend() {
				return nil
			}
		} else {
			if tryDefend() {
				return nil
			}
			if tryCounter() {
				return nil
			}
		}
	}

	return game.HandleAction(model.PlayerAction{
		PlayerID:  top.TargetID,
		Type:      model.CmdRespond,
		ExtraArgs: []string{"take"},
	})
}

func shouldPreferMissResponse(game *engine.GameEngine, attacker *model.Player, defender *model.Player) bool {
	if currentDirectedScenarioPlan == nil || game == nil || attacker == nil || defender == nil {
		return false
	}

	// 贯穿射击：需要“主动攻击未命中”窗口。
	if scenarioTargetsSkill("贯穿射击") && attacker.Role == "archer" {
		return true
	}

	// 剑影/天使之歌：优先给对应阵营制造“应战命中->战绩区+水晶”的机会。
	if scenarioTargetsSkill("剑影") && campHasRole(game, defender.Camp, "blade_master") {
		return true
	}
	if scenarioTargetsSkill("天使之歌") && campHasRole(game, defender.Camp, "angel") {
		return true
	}

	// 冒险家经济链：通过防御/应战降低手牌，增加下回合买入触发地下法则的概率。
	if defender.Role == "adventurer" && (scenarioTargetsSkill("地下法则") || scenarioTargetsSkill("冒险者天堂")) {
		return len(defender.Hand) > 0
	}

	return false
}

func shouldPreferCounterForCrystal(game *engine.GameEngine, attacker *model.Player, defender *model.Player) bool {
	if currentDirectedScenarioPlan == nil || game == nil || attacker == nil || defender == nil {
		return false
	}
	if scenarioTargetsSkill("剑影") && campHasRole(game, defender.Camp, "blade_master") {
		return true
	}
	if scenarioTargetsSkill("天使之歌") && campHasRole(game, defender.Camp, "angel") {
		return true
	}
	return false
}

func promptHasOption(prompt *model.Prompt, id string) bool {
	if prompt == nil {
		return false
	}
	for _, opt := range prompt.Options {
		if opt.ID == id {
			return true
		}
	}
	return false
}

func findDefendCardIndex(player *model.Player) (int, bool) {
	if player == nil {
		return -1, false
	}
	for idx, card := range player.Hand {
		if card.Type == model.CardTypeMagic {
			return idx, true
		}
	}
	return -1, false
}

func findCounterCardIndex(player *model.Player, attackElement model.Element) (int, bool) {
	if player == nil {
		return -1, false
	}
	if attackElement == "" {
		return -1, false
	}
	for idx, card := range player.Hand {
		if card.Type != model.CardTypeAttack {
			continue
		}
		if card.Element == attackElement || card.Element == model.ElementDark {
			return idx, true
		}
	}
	return -1, false
}

func chooseCounterTargetID(targetIDs []string) string {
	if len(targetIDs) == 0 {
		return ""
	}
	return targetIDs[0]
}

func collectCounterTargetIDs(game *engine.GameEngine, attackerID string) []string {
	if game == nil || attackerID == "" {
		return nil
	}
	attacker := game.State.Players[attackerID]
	if attacker == nil {
		return nil
	}
	out := make([]string, 0, 2)
	for _, pid := range game.State.PlayerOrder {
		if pid == attackerID {
			continue
		}
		p := game.State.Players[pid]
		if p == nil || p.Camp != attacker.Camp {
			continue
		}
		out = append(out, pid)
	}
	return out
}

func campHasRole(game *engine.GameEngine, camp model.Camp, roleID string) bool {
	if game == nil || roleID == "" {
		return false
	}
	for _, pid := range game.State.PlayerOrder {
		p := game.State.Players[pid]
		if p == nil || p.Camp != camp {
			continue
		}
		if p.Role == roleID {
			return true
		}
	}
	return false
}

func hasBasicFieldEffectOnBoard(game *engine.GameEngine) bool {
	if game == nil {
		return false
	}
	for _, pid := range game.State.PlayerOrder {
		p := game.State.Players[pid]
		if p == nil {
			continue
		}
		for _, fc := range p.Field {
			if fc.Mode != model.FieldEffect {
				continue
			}
			if fc.Effect == model.EffectShield || fc.Effect == model.EffectWeak || fc.Effect == model.EffectPoison {
				return true
			}
		}
	}
	return false
}

func normalizeQueuedActionForBeforeAction(game *engine.GameEngine) {
	if game.State.Phase != model.PhaseBeforeAction {
		return
	}
	if len(game.State.ActionQueue) == 0 {
		return
	}

	qa := &game.State.ActionQueue[0]
	source := game.State.Players[qa.SourceID]
	if source == nil {
		dropQueuedAction(game)
		return
	}
	if qa.SourceSkill == "adventurer_fraud" {
		return
	}

	if qa.CardIndex >= 0 && qa.CardIndex < len(source.Hand) {
		curr := source.Hand[qa.CardIndex]
		if qa.Card == nil || curr.ID == qa.Card.ID {
			cardCopy := curr
			qa.Card = &cardCopy
			return
		}
	}

	if idx, ok := findFallbackCardIndex(source, qa); ok {
		qa.CardIndex = idx
		cardCopy := source.Hand[idx]
		qa.Card = &cardCopy
		return
	}

	dropQueuedAction(game)
}

func findFallbackCardIndex(player *model.Player, qa *model.QueuedAction) (int, bool) {
	needType := requiredCardTypeForAction(qa.Type)
	if needType == "" {
		return -1, false
	}

	if qa.Card != nil {
		for idx, c := range player.Hand {
			if c.ID == qa.Card.ID && c.Type == needType {
				return idx, true
			}
		}
	}

	for idx, c := range player.Hand {
		if c.Type != needType {
			continue
		}
		if qa.Element == "" || c.Element == qa.Element {
			return idx, true
		}
	}

	for idx, c := range player.Hand {
		if c.Type == needType {
			return idx, true
		}
	}

	return -1, false
}

func requiredCardTypeForAction(actionType model.ActionType) model.CardType {
	switch actionType {
	case model.ActionAttack:
		return model.CardTypeAttack
	case model.ActionMagic:
		return model.CardTypeMagic
	default:
		return ""
	}
}

func dropQueuedAction(game *engine.GameEngine) {
	if len(game.State.ActionQueue) == 0 {
		return
	}
	game.State.ActionQueue = game.State.ActionQueue[1:]
	if len(game.State.ActionQueue) > 0 {
		game.State.Phase = model.PhaseBeforeAction
		return
	}
	game.State.Phase = model.PhaseExtraAction
}

func recoverBrokenPhaseWithoutInterrupt(game *engine.GameEngine) bool {
	if game.State.PendingInterrupt != nil {
		return false
	}

	switch game.State.Phase {
	case model.PhaseDiscardSelection:
		if len(game.State.PendingDamageQueue) > 0 {
			game.State.Phase = model.PhasePendingDamageResolution
		} else if len(game.State.ActionQueue) > 0 {
			game.State.Phase = model.PhaseBeforeAction
		} else if len(game.State.CombatStack) > 0 {
			game.State.Phase = model.PhaseCombatInteraction
		} else {
			game.State.Phase = model.PhaseTurnEnd
		}
		game.Drive()
		return true

	case model.PhaseBeforeAction:
		if len(game.State.ActionQueue) == 0 {
			game.State.Phase = model.PhaseExtraAction
			game.Drive()
			return true
		}

	case model.PhaseCombatInteraction:
		if len(game.State.CombatStack) == 0 {
			if len(game.State.PendingDamageQueue) > 0 {
				game.State.Phase = model.PhasePendingDamageResolution
			} else {
				game.State.Phase = model.PhaseExtraAction
			}
			game.Drive()
			return true
		}

	case model.PhaseResponse:
		if len(game.State.ActionStack) == 0 && len(game.State.CombatStack) == 0 {
			if len(game.State.PendingDamageQueue) > 0 {
				game.State.Phase = model.PhasePendingDamageResolution
			} else {
				game.State.Phase = model.PhaseExtraAction
			}
			game.Drive()
			return true
		}
	}

	return false
}

func tryRecoverFromStall(game *engine.GameEngine) bool {
	if recoverBrokenPhaseWithoutInterrupt(game) {
		return true
	}

	if game.State.Phase == model.PhaseBeforeAction {
		normalizeQueuedActionForBeforeAction(game)
		if len(game.State.ActionQueue) == 0 {
			game.State.Phase = model.PhaseExtraAction
			game.Drive()
			return true
		}
		dropQueuedAction(game)
		game.Drive()
		return true
	}

	if len(game.State.ActionQueue) > 0 {
		dropQueuedAction(game)
		game.Drive()
		return true
	}

	if game.State.Phase == model.PhaseDiscardSelection && game.State.PendingInterrupt == nil {
		game.State.Phase = model.PhaseTurnEnd
		game.Drive()
		return true
	}

	game.Drive()
	return true
}

func gameplaySnapshot(game *engine.GameEngine) string {
	b := strings.Builder{}
	b.WriteString(string(game.State.Phase))
	b.WriteString("|")
	b.WriteString(strconv.Itoa(game.State.CurrentTurn))
	b.WriteString("|")
	if game.State.PendingInterrupt != nil {
		b.WriteString(string(game.State.PendingInterrupt.Type))
	} else {
		b.WriteString("nil")
	}
	b.WriteString("|")
	b.WriteString(strconv.Itoa(len(game.State.ActionQueue)))
	b.WriteString("|")
	b.WriteString(strconv.Itoa(len(game.State.CombatStack)))
	b.WriteString("|")
	b.WriteString(strconv.Itoa(len(game.State.PendingDamageQueue)))
	if len(game.State.ActionQueue) > 0 {
		qa := game.State.ActionQueue[0]
		b.WriteString("|")
		b.WriteString(qa.SourceID)
		b.WriteString(":")
		b.WriteString(qa.TargetID)
		b.WriteString(":")
		b.WriteString(strconv.Itoa(qa.CardIndex))
		b.WriteString(":")
		b.WriteString(string(qa.Type))
	}
	return b.String()
}

func observerTailLogs(observer *autoGameObserver, n int) string {
	if observer == nil || len(observer.logs) == 0 {
		return "(no logs)"
	}
	if n <= 0 {
		n = 20
	}
	start := len(observer.logs) - n
	if start < 0 {
		start = 0
	}
	return strings.Join(observer.logs[start:], "\n")
}

func missingSkillList(expected map[string]struct{}, triggered map[string]int) []string {
	missing := make([]string, 0)
	for sid := range expected {
		if triggered[sid] == 0 {
			missing = append(missing, sid)
		}
	}
	sort.Strings(missing)
	return missing
}

func topTriggeredActionSkills(expected map[string]struct{}, triggered map[string]int, limit int) []string {
	type kv struct {
		name  string
		count int
	}
	list := make([]kv, 0)
	for name := range expected {
		if triggered[name] > 0 {
			list = append(list, kv{name: name, count: triggered[name]})
		}
	}
	sort.Slice(list, func(i, j int) bool {
		if list[i].count != list[j].count {
			return list[i].count > list[j].count
		}
		return list[i].name < list[j].name
	})
	if limit > 0 && len(list) > limit {
		list = list[:limit]
	}
	out := make([]string, 0, len(list))
	for _, item := range list {
		out = append(out, fmt.Sprintf("%s=%d", item.name, item.count))
	}
	return out
}

func summarizeGameState(game *engine.GameEngine) string {
	pid := ""
	if len(game.State.PlayerOrder) > 0 && game.State.CurrentTurn >= 0 && game.State.CurrentTurn < len(game.State.PlayerOrder) {
		pid = game.State.PlayerOrder[game.State.CurrentTurn]
	}
	pending := "nil"
	if game.State.PendingInterrupt != nil {
		pending = string(game.State.PendingInterrupt.Type)
	}
	return fmt.Sprintf(
		"phase=%s turn=%s pending=%s queue=%d combat=%d pendingDmg=%d redMorale=%d blueMorale=%d redCups=%d blueCups=%d",
		game.State.Phase,
		pid,
		pending,
		len(game.State.ActionQueue),
		len(game.State.CombatStack),
		len(game.State.PendingDamageQueue),
		game.State.RedMorale,
		game.State.BlueMorale,
		game.State.RedCups,
		game.State.BlueCups,
	)
}
