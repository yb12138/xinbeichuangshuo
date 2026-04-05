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

func (e *GameEngine) debugFindCharacter(roleID string) *model.Character {
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
	charID := ""
	faction := ""
	if char != nil {
		charID = char.ID
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
		ExclusiveChar1:  charID,
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

func (e *GameEngine) debugEnsureExclusiveCard(player *model.Player, skill model.SkillDefinition, toHand bool) {
	if player == nil || player.Character == nil || skill.Title == "" {
		return
	}
	charID := player.Character.ID
	if player.HasExclusiveCard(charID, skill.Title) {
		return
	}
	cards, err := e.debugDrawExclusiveCardsFromStock(charID, skill.Title, 1)
	if err != nil || len(cards) == 0 {
		e.Log(fmt.Sprintf("[Cheat] 独有牌补齐失败 [%s·%s]: %v", charID, skill.Title, err))
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
		card.ExclusiveChar1 = user.Character.ID
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
	targetStr := act.TargetID
	if targetStr == "turn" {
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
		cards, err := e.debugDrawExclusiveCardsFromStock(char.ID, skill.Title, count)
		if err != nil {
			return err
		}
		player.Hand = append(player.Hand, cards...)
		e.Log(fmt.Sprintf("[Cheat] %s 获得 %d 张独有技手牌 [%s·%s]", player.Name, count, char.Name, skill.Title))
		return nil
	}
	if targetStr == "card_element" {
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

		if err := e.forceTurnTo(pid); err != nil {
			return err
		}
		e.Log(fmt.Sprintf("[Cheat] 强制切换回合到 %s", player.Name))

		e.State.PendingInterrupt = nil
		e.State.InterruptQueue = nil
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

	var template *model.Card
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
