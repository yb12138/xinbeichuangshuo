// gameflow: 调试作弊指令（仅测试/开发）。

package engine

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"starcup-engine/internal/data"
	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
)

func (e *GameEngine) forceTurnTo(targetPID string) error {
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

	currentPID := e.State.PlayerOrder[e.State.CurrentTurn]
	if curr := e.State.Players[currentPID]; curr != nil {
		curr.IsActive = false
	}

	e.State.CurrentTurn = foundIdx
	newPlayer := e.State.Players[targetPID]
	newPlayer.IsActive = true
	newPlayer.TurnState = model.NewPlayerTurnState()

	e.enterActionExecutionStage()

	return nil
}

func (e *GameEngine) DebugFindCharacter(roleID string) *model.Character {
	if roleID == "" {
		return nil
	}
	characters := data.GetCharacters()
	for _, c := range characters {
		if c.ID == roleID {
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
	e.ensureStarterRoleCards(player)
	e.rebuildTimingOnAttackDeclaredRegistry()
	e.refreshPlayerDerivedState(player)
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

func debugRemoveCardIndices(src []model.Card, indices []int) []model.Card {
	if len(indices) == 0 {
		return src
	}
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
func (e *GameEngine) debugDrawExclusiveCardsFromStock(characterID, skillTitle string, count int) ([]model.Card, error) {
	if count <= 0 {
		return nil, nil
	}
	if characterID == "" || skillTitle == "" {
		return nil, fmt.Errorf("独有牌检索参数无效")
	}

	deckIndices := make([]int, 0, count)
	for i, c := range e.State.Deck {
		if c.MatchExclusive(characterID, skillTitle) {
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
			if c.MatchExclusive(characterID, skillTitle) {
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
			count, characterID, skillTitle, len(deckIndices)+len(discardIndices),
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

func (e *GameEngine) debugAddExclusiveCopies(player *model.Player, skill model.SkillDefinition, count int) error {
	if player == nil || player.Character == nil || count <= 0 {
		return nil
	}
	cards, err := e.debugDrawExclusiveCardsFromStock(player.Character.ID, skill.Title, count)
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

func debugFieldHook(effect model.EffectType) model.FieldHook {
	switch effect {
	case model.EffectPoison, model.EffectWeak:
		return model.FieldHookOnBeforeAction
	case model.EffectShield:
		return model.FieldHookOnDamaged
	case model.EffectSealFire, model.EffectSealWater, model.EffectSealEarth, model.EffectSealWind, model.EffectSealThunder:
		return model.FieldHookOnAttack
	default:
		return model.FieldHookManual
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

	dispatch := debugFieldHook(effect)
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
			Hook:     dispatch,
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
	player.TurnState.HasUsedActionSkill = false
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
		if err := e.debugAddExclusiveCopies(player, skill, required); err != nil {
			return err
		}
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

// 若干历史技能仅区分攻击结果分支；调试上下文按技能 ID 保留未命中模拟语义。
var debugCheatSimulateAttackMiss = map[string]bool{
	"piercing_shot":       true,
	"hom_rage_suppress":   true,
	"hom_glyph_fusion":    true,
	"se_sword_soul_guard": true,
	"se_feint":            true,
}

func debugCheatEventTypeForTiming(t model.FlowTiming) model.EventType {
	switch t {
	case model.TimingAttackDeclare,
		model.TimingAttackSelectTarget,
		model.TimingAttackPlayCard,
		model.TimingAttackModifyCard,
		model.TimingAttackCommitted,
		model.TimingAttackForceHitCheck,
		model.TimingAttackNoResponseCheck,
		model.TimingAttackResponse,
		model.TimingAttackHit,
		model.TimingAttackMiss,
		model.TimingDamageSourceDeal:
		return model.EventAttack
	case model.TimingDamageTaken:
		return model.EventDamage
	case model.TimingOnCardPlayedOrRevealed:
		return model.EventCardUsed
	case model.TimingBeforeCardDrawn:
		return model.EventBeforeDraw
	case model.TimingOnCardDrawn:
		return model.EventAfterDraw
	case model.TimingOnTurnStart, model.TimingStartup:
		return model.EventTurnStart
	case model.TimingOnActionEnd:
		return model.EventPhaseEnd
	default:
		return model.EventNone
	}
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
	if skill.HasTiming(model.TimingOnCardPlayedOrRevealed) {
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
		card.ExclusiveChar1 = user.Character.ID
		card.ExclusiveSkill1 = skill.Title
	}

	damageVal := 1
	drawCount := 1
	timing := model.NormalizeTiming(skill.PrimaryTimingOrLegacy())
	if timing == model.TimingAttackResponse {
		if debugCheatSimulateAttackMiss[skill.ID] {
			timing = model.TimingAttackMiss
		} else {
			timing = model.TimingAttackHit
		}
	}
	eventType := debugCheatEventTypeForTiming(timing)

	attackInfo := &model.AttackEventInfo{
		IsHit:          true,
		IsHitForced:    false,
		Element:        string(element),
		CanBeResponded: true,
		ActionType:     string(actionType),
	}
	// 攻击结果类技能按 rulebook timing 明确区分命中/未命中。
	if timing == model.TimingAttackDeclare {
		attackInfo.IsHit = false
	} else if timing == model.TimingAttackMiss {
		attackInfo.IsHit = false
	}

	ctx := &model.Context{
		Game:   e,
		User:   user,
		Target: target,
		Timing: timing,
		EventCtx: &model.EventContext{
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

type cheatCommandHandler func(e *GameEngine, act model.PlayerAction) error

var cheatCommandHandlers = map[string]cheatCommandHandler{
	"turn":           (*GameEngine).HandleCheatTurn,
	"role":           (*GameEngine).HandleCheatRole,
	"token":          (*GameEngine).HandleCheatToken,
	"set":            (*GameEngine).HandleCheatSet,
	"effect":         (*GameEngine).HandleCheatEffect,
	"card_exclusive": (*GameEngine).HandleCheatCardExclusive,
	"card_element":   (*GameEngine).HandleCheatCardElement,
	"card_faction":   (*GameEngine).HandleCheatCardFaction,
	"card_magic":     (*GameEngine).HandleCheatCardMagic,
	"discard":        (*GameEngine).HandleCheatDiscard,
	"skill":          (*GameEngine).HandleCheatSkill,
}

// HandleCheat 处理作弊指令 (用于测试)
func (e *GameEngine) HandleCheat(act model.PlayerAction) error {
	if handler, ok := cheatCommandHandlers[act.TargetID]; ok {
		return handler(e, act)
	}
	return e.HandleCheatAddCardByName(act)
}

func (e *GameEngine) HandleCheatTurn(act model.PlayerAction) error {
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

func (e *GameEngine) HandleCheatRole(act model.PlayerAction) error {
	if len(act.ExtraArgs) < 2 {
		return fmt.Errorf("用法: cheat role <pid> <role_id>")
	}
	pid := act.ExtraArgs[0]
	roleID := act.ExtraArgs[1]
	player, err := e.getCheatPlayer(pid)
	if err != nil {
		return err
	}
	char := e.DebugFindCharacter(roleID)
	if char == nil {
		return fmt.Errorf("角色不存在: %s", roleID)
	}
	e.debugResetPlayerForRole(player, char)
	e.Log(fmt.Sprintf("[Cheat] %s 切换角色为 %s", player.Name, char.Name))
	return nil
}

func (e *GameEngine) HandleCheatToken(act model.PlayerAction) error {
	if len(act.ExtraArgs) < 3 {
		return fmt.Errorf("用法: cheat token <pid> <token_key> <value>")
	}
	pid := act.ExtraArgs[0]
	tokenKey := act.ExtraArgs[1]
	val, err := strconv.Atoi(act.ExtraArgs[2])
	if err != nil {
		return fmt.Errorf("token 值无效: %s", act.ExtraArgs[2])
	}
	player, err := e.getCheatPlayer(pid)
	if err != nil {
		return err
	}
	ensurePlayerTokensMap(player)
	player.Tokens[tokenKey] = val
	e.Log(fmt.Sprintf("[Cheat] %s 指示物 %s=%d", player.Name, tokenKey, val))
	return nil
}

func (e *GameEngine) HandleCheatSet(act model.PlayerAction) error {
	if len(act.ExtraArgs) < 3 {
		return fmt.Errorf("用法: cheat set <pid> <field> <value>")
	}
	pid := act.ExtraArgs[0]
	field := act.ExtraArgs[1]
	val, err := strconv.Atoi(act.ExtraArgs[2])
	if err != nil {
		return fmt.Errorf("数值无效: %s", act.ExtraArgs[2])
	}
	player, err := e.getCheatPlayer(pid)
	if err != nil {
		return err
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

func (e *GameEngine) HandleCheatEffect(act model.PlayerAction) error {
	if len(act.ExtraArgs) < 3 {
		return fmt.Errorf("用法: cheat effect <pid> <effect_type> <count>")
	}
	pid := act.ExtraArgs[0]
	rawEffect := act.ExtraArgs[1]
	count, err := strconv.Atoi(act.ExtraArgs[2])
	if err != nil {
		return fmt.Errorf("效果数量无效: %s", act.ExtraArgs[2])
	}
	player, err := e.getCheatPlayer(pid)
	if err != nil {
		return err
	}
	effectType, err := debugParseEffectType(rawEffect)
	if err != nil {
		return err
	}
	e.debugSetEffectCount(player, effectType, count)
	e.Log(fmt.Sprintf("[Cheat] %s 基础效果 %s 设置为 %d 层", player.Name, effectType, count))
	return nil
}

func (e *GameEngine) HandleCheatCardExclusive(act model.PlayerAction) error {
	if len(act.ExtraArgs) < 3 {
		return fmt.Errorf("用法: cheat card_exclusive <pid> <role_id> <skill_id> [count]")
	}
	pid := act.ExtraArgs[0]
	roleID := act.ExtraArgs[1]
	skillID := act.ExtraArgs[2]
	count, err := parseCheatOptionalCount(act.ExtraArgs, 3, 1)
	if err != nil {
		return err
	}
	if count <= 0 {
		return fmt.Errorf("数量必须大于0")
	}
	player, err := e.getCheatPlayer(pid)
	if err != nil {
		return err
	}
	char := e.DebugFindCharacter(roleID)
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
	cards, err := e.debugDrawExclusiveCardsFromStock(char.ID, skill.Title, count)
	if err != nil {
		return err
	}
	player.Hand = append(player.Hand, cards...)
	e.Log(fmt.Sprintf("[Cheat] %s 获得 %d 张独有技手牌 [%s·%s]", player.Name, count, char.Name, skill.Title))
	return nil
}

func (e *GameEngine) HandleCheatCardElement(act model.PlayerAction) error {
	if len(act.ExtraArgs) < 2 {
		return fmt.Errorf("用法: cheat card_element <pid> <element> [count]")
	}
	pid := act.ExtraArgs[0]
	rawElement := act.ExtraArgs[1]
	count, err := parseCheatOptionalCount(act.ExtraArgs, 2, 1)
	if err != nil {
		return err
	}
	player, err := e.getCheatPlayer(pid)
	if err != nil {
		return err
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

func (e *GameEngine) HandleCheatCardFaction(act model.PlayerAction) error {
	if len(act.ExtraArgs) < 2 {
		return fmt.Errorf("用法: cheat card_faction <pid> <faction> [count]")
	}
	pid := act.ExtraArgs[0]
	rawFaction := act.ExtraArgs[1]
	count, err := parseCheatOptionalCount(act.ExtraArgs, 2, 1)
	if err != nil {
		return err
	}
	player, err := e.getCheatPlayer(pid)
	if err != nil {
		return err
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

func (e *GameEngine) HandleCheatCardMagic(act model.PlayerAction) error {
	if len(act.ExtraArgs) < 2 {
		return fmt.Errorf("用法: cheat card_magic <pid> <card_name> [count]")
	}
	pid := act.ExtraArgs[0]
	cardName := strings.TrimSpace(act.ExtraArgs[1])
	count, err := parseCheatOptionalCount(act.ExtraArgs, 2, 1)
	if err != nil {
		return err
	}
	player, err := e.getCheatPlayer(pid)
	if err != nil {
		return err
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

func (e *GameEngine) HandleCheatDiscard(act model.PlayerAction) error {
	if len(act.ExtraArgs) < 2 {
		return fmt.Errorf("用法: cheat discard <pid> <count>")
	}
	pid := act.ExtraArgs[0]
	count, err := strconv.Atoi(act.ExtraArgs[1])
	if err != nil {
		return fmt.Errorf("弃牌数量无效: %s", act.ExtraArgs[1])
	}
	if count <= 0 {
		return fmt.Errorf("弃牌数量必须大于0")
	}
	player, err := e.getCheatPlayer(pid)
	if err != nil {
		return err
	}
	if len(player.Hand) == 0 {
		e.Log(fmt.Sprintf("[Cheat] %s 当前无手牌可弃", player.Name))
		return nil
	}

	if count > len(player.Hand) {
		count = len(player.Hand)
	}
	discarded := append([]model.Card{}, player.Hand[:count]...)
	player.Hand = append([]model.Card{}, player.Hand[count:]...)
	e.State.DiscardPile = append(e.State.DiscardPile, discarded...)
	e.NotifyCardHidden(player.ID, discarded, "discard")
	e.Log(fmt.Sprintf("[Cheat] %s 强制弃置 %d 张手牌", player.Name, count))
	return nil
}

func (e *GameEngine) HandleCheatSkill(act model.PlayerAction) error {
	if len(act.ExtraArgs) < 2 {
		return fmt.Errorf("用法: cheat skill <pid> [role_id] <skill_id>")
	}
	pid := act.ExtraArgs[0]
	player, err := e.getCheatPlayer(pid)
	if err != nil {
		return err
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
		char := e.DebugFindCharacter(roleID)
		if char == nil {
			return fmt.Errorf("角色不存在: %s", roleID)
		}
		e.debugResetPlayerForRole(player, char)
	}

	if err := e.forceTurnTo(pid); err != nil {
		return err
	}
	e.Log(fmt.Sprintf("[Cheat] 强制切换回合到 %s", player.Name))

	e.State.SetPendingInterrupt(nil)
	if len(e.State.InterruptQueue) > 0 {
		e.State.InterruptQueue = nil
		e.State.TouchInterruptRevision()
	} else {
		e.State.InterruptQueue = nil
	}
	e.State.ActionQueue = []model.QueuedAction{}
	e.State.ActionStack = []model.Action{}
	e.State.CombatStack = []model.CombatRequest{}
	player.TurnState = model.NewPlayerTurnState()

	skill, ok := e.debugFindSkill(player, skillID)
	if !ok {
		return fmt.Errorf("技能不存在: %s", skillID)
	}
	e.debugPrepareSkillResources(player, skill)
	if err := e.debugPrepareSkillCards(player, skill); err != nil {
		return err
	}
	if skill.Type == model.SkillTypeAction {
		e.Log(fmt.Sprintf("[Cheat] 已准备技能 %s（行动技），请在 UI 手动发动", skill.Title))
		return nil
	}

	ctx := e.debugBuildContext(player, skill)
	if ctx == nil {
		return fmt.Errorf("无法构建技能上下文")
	}
	e.dispatcher.processSkills([]model.SkillDefinition{skill}, ctx)
	e.Log(fmt.Sprintf("[Cheat] 已触发调试技能 %s", skill.Title))
	return nil
}

// 未命中预定义 cheat 子命令时，按“给玩家添加指定名称卡牌”处理。
func (e *GameEngine) HandleCheatAddCardByName(act model.PlayerAction) error {
	pid := act.TargetID
	if pid == "" {
		return fmt.Errorf("未指定玩家ID")
	}
	player, err := e.getCheatPlayer(pid)
	if err != nil {
		return err
	}

	if len(act.ExtraArgs) == 0 {
		return fmt.Errorf("未指定卡牌名称")
	}
	cardName := act.ExtraArgs[0]

	count := 1
	if len(act.ExtraArgs) > 1 {
		if c, parseErr := strconv.Atoi(act.ExtraArgs[1]); parseErr == nil {
			count = c
		}
	}

	var template *model.Card
	for _, c := range rules.InitDeck() {
		if c.Name == cardName {
			card := c
			template = &card
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

func (e *GameEngine) getCheatPlayer(pid string) (*model.Player, error) {
	player := e.State.Players[pid]
	if player == nil {
		return nil, fmt.Errorf("玩家不存在: %s", pid)
	}
	return player, nil
}

func parseCheatOptionalCount(extraArgs []string, idx int, defaultCount int) (int, error) {
	if len(extraArgs) <= idx {
		return defaultCount, nil
	}
	count, err := strconv.Atoi(extraArgs[idx])
	if err != nil {
		return 0, fmt.Errorf("数量无效: %s", extraArgs[idx])
	}
	return count, nil
}
