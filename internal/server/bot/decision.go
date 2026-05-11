package bot

import (
	"math"
	"sort"
	"strconv"
	"strings"
	"sync"

	"starcup-engine/internal/data"
	"starcup-engine/internal/model"
)

func DecideAction(input DecisionInput) (model.PlayerAction, bool) {
	me, ok := input.State.Players[input.PlayerID]
	if !ok {
		return model.PlayerAction{}, false
	}

	optionSet := map[string]bool{}
	for _, o := range input.Prompt.Options {
		optionSet[o.ID] = true
	}

	// 1) 对战响应优先
	if optionSet["take"] || optionSet["defend"] || optionSet["counter"] {
		return decideCombatResponse(input.PlayerID, input.State, input.Prompt, me, input.Memory)
	}

	// 2) 行动选择（攻击/法术/特殊行动/无法行动）
	if optionSet["attack"] || optionSet["magic"] || optionSet["special"] || optionSet["cannot_act"] ||
		hasPromptOption(input.Prompt, "buy") || hasPromptOption(input.Prompt, "extract") || hasPromptOption(input.Prompt, "synthesize") {
		return decideActionSelection(input.PlayerID, input.State, input.Prompt, me, input.AvailableSkills, input.Memory)
	}

	// 3) 选择类中断
	switch input.Prompt.Type {
	case model.PromptChooseCards:
		return decideChooseCards(input.PlayerID, input.State, input.Prompt, me)
	case model.PromptChooseSkill:
		return decideChooseSkill(input.PlayerID, input.State, input.Prompt, me)
	case model.PromptChooseExtract:
		return decideChooseExtract(input.PlayerID, input.State, input.Prompt, me)
	case model.PromptConfirm:
		return decideConfirmChoice(input.PlayerID, input.State, input.Prompt, me)
	default:
		return BuildFallbackAction(input.PlayerID, input.Prompt, input.State)
	}
}

func decideActionSelection(playerID string, state StateSnapshot, prompt *model.Prompt, me PlayerSnapshot, availableSkills []AvailableSkill, memory *Memory) (model.PlayerAction, bool) {
	playable := botPlayableCards(me)
	attackCards := collectCardIndices(playable, model.CardTypeAttack)
	magicCards := collectCardIndices(playable, model.CardTypeMagic)

	maxHand := estimateMaxHand(me.Role)
	handPressure := float64(len(me.Hand)) / float64(maxHand)
	enemyThreat := estimateEnemyThreat(state, playerID)
	campNeed := estimateCampResourceNeed(state, me.Camp)

	bestAttack, bestAttackScore := pickBestAttack(playerID, state, me, attackCards, playable, campNeed, handPressure, enemyThreat, memory)
	bestMagic, bestMagicScore := pickBestMagic(playerID, state, me, magicCards, playable, handPressure, enemyThreat)

	canAttack := hasPromptOption(prompt, "attack") && bestAttack.Type == model.CmdAttack
	canMagic := hasPromptOption(prompt, "magic") && (bestMagic.Type == model.CmdMagic || bestMagic.Type == model.CmdSkill)

	// 若当前可打出高收益伤害，优先打伤害
	if canAttack && (bestAttackScore >= bestMagicScore || !canMagic) && bestAttackScore > 1.2 {
		return bestAttack, true
	}
	if canMagic && bestMagicScore >= bestAttackScore && bestMagicScore > 1.0 {
		return bestMagic, true
	}

	// 手牌压力较高且本回合难以打高收益时，优先“减手牌/保命节奏”
	if handPressure >= 0.83 {
		if canAttack && bestAttack.Type == model.CmdAttack {
			return bestAttack, true
		}
		if canMagic && bestMagic.Type == model.CmdMagic {
			return bestMagic, true
		}
		if hasPromptOption(prompt, "extract") {
			return model.PlayerAction{PlayerID: playerID, Type: model.CmdExtract}, true
		}
	}

	// 资源运营动作
	if hasPromptOption(prompt, "synthesize") && campNeed < 0.45 {
		return model.PlayerAction{PlayerID: playerID, Type: model.CmdSynthesize}, true
	}
	if hasPromptOption(prompt, "extract") && me.Gem+me.Crystal < estimateMaxEnergy(me.Role) {
		return model.PlayerAction{PlayerID: playerID, Type: model.CmdExtract}, true
	}
	if hasPromptOption(prompt, "buy") && handPressure <= 0.6 {
		return model.PlayerAction{PlayerID: playerID, Type: model.CmdBuy}, true
	}
	if hasPromptOption(prompt, "cannot_act") {
		return model.PlayerAction{PlayerID: playerID, Type: model.CmdCannotAct}, true
	}

	if canAttack {
		return bestAttack, true
	}
	if canMagic {
		return bestMagic, true
	}

	// 特殊兜底：无牌可打但有可发动技能
	if hasPromptOption(prompt, "magic") {
		if skillAct, ok := chooseSkillAction(playerID, state, me, availableSkills); ok {
			return skillAct, true
		}
	}

	return model.PlayerAction{PlayerID: playerID, Type: model.CmdPass}, true
}

func pickBestAttack(playerID string, state StateSnapshot, me PlayerSnapshot, attackCards []int, hand []model.Card, campNeed, handPressure, enemyThreat float64, memory *Memory) (model.PlayerAction, float64) {
	best := model.PlayerAction{}
	bestScore := -999.0
	for _, cardIdx := range attackCards {
		card := hand[cardIdx]
		for targetID, target := range state.Players {
			if targetID == playerID || target.Camp == me.Camp {
				continue
			}
			hitProb := estimateHitProbability(state, me, target, card.Element, memory)
			damage := float64(card.Damage)
			damage += estimateBerserkerBonus(me, card, target, len(hand))
			score := damage * hitProb

			// 阵营缺资源时，命中收益上调（命中可产出红宝石/蓝水晶）
			score += 0.9 * campNeed * hitProb

			// 打手牌少的目标更容易命中（你提到的重点策略）
			if target.HandCount <= 2 {
				score += 0.5
			}
			if target.HandCount <= 1 {
				score += 0.3
			}

			// 若敌方后续威胁高，优先压制高威胁目标
			score += 0.4 * enemyThreat * roleThreatWeight(target.Role)

			// 手牌压力大时，主动出牌本身有价值
			score += 0.2 * handPressure

			if score > bestScore {
				bestScore = score
				best = model.PlayerAction{
					PlayerID:  playerID,
					Type:      model.CmdAttack,
					TargetID:  targetID,
					CardIndex: cardIdx,
				}
			}
		}
	}
	return best, bestScore
}

func pickBestMagic(playerID string, state StateSnapshot, me PlayerSnapshot, magicCards []int, hand []model.Card, handPressure, enemyThreat float64) (model.PlayerAction, float64) {
	best := model.PlayerAction{}
	bestScore := -999.0
	for _, cardIdx := range magicCards {
		card := hand[cardIdx]
		targets := pickMagicTargets(state, me, card)
		for _, targetID := range targets {
			target := state.Players[targetID]
			score := float64(card.Damage)
			name := card.Name

			if strings.Contains(name, "虚弱") || strings.Contains(name, "中毒") || strings.Contains(name, "封印") {
				score += 1.4 + 0.4*enemyThreat*roleThreatWeight(target.Role)
			}
			if strings.Contains(name, "圣盾") || strings.Contains(name, "圣光") {
				// 防御向法术：对自己或高威胁局面价值更高
				if targetID == playerID {
					score += 1.2 + enemyThreat*0.6
				} else {
					score += 0.6
				}
			}
			score += 0.15 * handPressure

			if score > bestScore {
				bestScore = score
				best = model.PlayerAction{
					PlayerID:  playerID,
					Type:      model.CmdMagic,
					TargetID:  targetID,
					CardIndex: cardIdx,
				}
			}
		}
	}
	return best, bestScore
}

func chooseSkillAction(playerID string, state StateSnapshot, me PlayerSnapshot, availableSkills []AvailableSkill) (model.PlayerAction, bool) {
	if len(availableSkills) == 0 {
		return model.PlayerAction{}, false
	}
	// 优先使用无弃牌成本的输出/控制技能，避免复杂交互卡顿
	for _, sk := range availableSkills {
		if sk.CostDiscards > 0 {
			continue
		}
		targetIDs := pickTargetsForSkill(state, me, sk)
		return model.PlayerAction{
			PlayerID:  playerID,
			Type:      model.CmdSkill,
			SkillID:   sk.ID,
			TargetIDs: targetIDs,
			TargetID:  firstOrEmpty(targetIDs),
		}, true
	}
	return model.PlayerAction{}, false
}

func decideCombatResponse(playerID string, state StateSnapshot, prompt *model.Prompt, me PlayerSnapshot, memory *Memory) (model.PlayerAction, bool) {
	playable := botPlayableCards(me)
	handPressure := float64(len(me.Hand)) / float64(estimateMaxHand(me.Role))
	enemyThreat := estimateEnemyThreat(state, playerID)

	defendCards := collectDefendCards(playable)
	counterCards := collectCounterCards(playable, model.Element(prompt.AttackElement))

	// 先尝试防御：高压局、暗灭攻击、手牌接近上限时更偏防御
	if hasPromptOption(prompt, "defend") && len(defendCards) > 0 {
		shouldDefend := handPressure >= 0.8 || enemyThreat >= 0.75 || strings.EqualFold(prompt.AttackElement, string(model.ElementDark))
		if shouldDefend || (!hasPromptOption(prompt, "counter") || len(counterCards) == 0) {
			idx := defendCards[0]
			return model.PlayerAction{
				PlayerID:  playerID,
				Type:      model.CmdRespond,
				CardIndex: idx,
				ExtraArgs: []string{"defend"},
			}, true
		}
	}

	// 再考虑应战：优先选择同系牌，目标选攻击方队友中“手牌最少”的对象，提升命中概率。
	if hasPromptOption(prompt, "counter") && len(counterCards) > 0 {
		targetID := pickCounterTarget(state, prompt.CounterTargetIDs)
		if len(prompt.CounterTargetIDs) == 0 && strings.Contains(prompt.Message, "魔弹") {
			// 魔弹链条可无反弹目标
			targetID = ""
		}
		if targetID != "" || strings.Contains(prompt.Message, "魔弹") {
			idx := pickBestCounterCard(counterCards, playable, model.Element(prompt.AttackElement))
			action := model.PlayerAction{
				PlayerID:  playerID,
				Type:      model.CmdRespond,
				CardIndex: idx,
				ExtraArgs: []string{"counter"},
			}
			if targetID != "" {
				action.TargetID = targetID
			}
			return action, true
		}
	}

	return model.PlayerAction{
		PlayerID:  playerID,
		Type:      model.CmdRespond,
		ExtraArgs: []string{"take"},
	}, true
}

func decideChooseCards(playerID string, state StateSnapshot, prompt *model.Prompt, me PlayerSnapshot) (model.PlayerAction, bool) {
	hand := me.Hand
	candidates := make([]int, 0, len(prompt.Options))
	for _, opt := range prompt.Options {
		idx, err := strconv.Atoi(opt.ID)
		if err != nil {
			continue
		}
		if idx >= 0 && idx < len(hand) {
			candidates = append(candidates, idx)
		}
	}
	if len(candidates) == 0 {
		return BuildFallbackAction(playerID, prompt, state)
	}

	minPick := prompt.Min
	maxPick := prompt.Max
	if minPick <= 0 {
		minPick = 1
	}
	if maxPick <= 0 || maxPick > len(candidates) {
		maxPick = len(candidates)
	}

	// 暗杀者【水影】特化：优先弃水系，潜行状态可额外弃1法术
	if prompt.SkillID == "water_shadow" {
		selections := pickWaterShadowDiscards(candidates, hand, me, minPick, maxPick)
		return model.PlayerAction{
			PlayerID:   playerID,
			Type:       model.CmdSelect,
			Selections: selections,
		}, true
	}

	targetCount := minPick
	if float64(len(hand))/float64(estimateMaxHand(me.Role)) >= 0.9 && maxPick > minPick {
		targetCount = minPick + 1
		if targetCount > maxPick {
			targetCount = maxPick
		}
	}

	sort.Slice(candidates, func(i, j int) bool {
		return cardKeepValue(me.Role, hand[candidates[i]], len(hand)) < cardKeepValue(me.Role, hand[candidates[j]], len(hand))
	})
	selections := append([]int{}, candidates[:targetCount]...)

	return model.PlayerAction{
		PlayerID:   playerID,
		Type:       model.CmdSelect,
		Selections: selections,
	}, true
}

func pickWaterShadowDiscards(candidates []int, hand []model.Card, me PlayerSnapshot, minPick, maxPick int) []int {
	isStealth := me.Form == model.FormAssassinStealth

	var water, magic, other []int
	for _, idx := range candidates {
		card := hand[idx]
		switch {
		case card.Element == model.ElementWater:
			water = append(water, idx)
		case card.Type == model.CardTypeMagic:
			magic = append(magic, idx)
		default:
			other = append(other, idx)
		}
	}

	// 水影至少弃1张水系牌
	targetCount := minPick
	pressure := float64(len(hand)) / float64(estimateMaxHand(me.Role))
	if pressure >= 0.85 && maxPick >= minPick+1 {
		targetCount = minPick + 1
	}
	if targetCount > maxPick {
		targetCount = maxPick
	}

	selections := make([]int, 0, targetCount)
	sort.Slice(water, func(i, j int) bool {
		return cardKeepValue(me.Role, hand[water[i]], len(hand)) < cardKeepValue(me.Role, hand[water[j]], len(hand))
	})
	for _, idx := range water {
		if len(selections) >= targetCount {
			break
		}
		selections = append(selections, idx)
	}
	if isStealth {
		sort.Slice(magic, func(i, j int) bool {
			return cardKeepValue(me.Role, hand[magic[i]], len(hand)) < cardKeepValue(me.Role, hand[magic[j]], len(hand))
		})
		for _, idx := range magic {
			if len(selections) >= targetCount {
				break
			}
			selections = append(selections, idx)
		}
	}

	// 兜底，避免数量不足导致中断卡住
	if len(selections) < minPick {
		sort.Slice(other, func(i, j int) bool {
			return cardKeepValue(me.Role, hand[other[i]], len(hand)) < cardKeepValue(me.Role, hand[other[j]], len(hand))
		})
		for _, idx := range append(water, append(magic, other...)...) {
			if len(selections) >= minPick {
				break
			}
			if !containsInt(selections, idx) {
				selections = append(selections, idx)
			}
		}
	}
	return selections
}

func decideChooseSkill(playerID string, state StateSnapshot, prompt *model.Prompt, me PlayerSnapshot) (model.PlayerAction, bool) {
	// 默认保守：倾向跳过，避免复杂连锁把自己锁死。
	skipIndex := -1
	for idx, opt := range prompt.Options {
		if strings.Contains(opt.ID, "skip") || strings.Contains(opt.Label, "跳过") {
			skipIndex = idx
			break
		}
	}

	// 少量角色特化：暗杀者启动技“潜行”在手牌偏多时优先发动
	if me.Role == "assassin" && me.Gem >= 1 {
		for idx, opt := range prompt.Options {
			if strings.Contains(opt.Label, "潜行") || strings.Contains(opt.ID, "stealth") {
				return model.PlayerAction{PlayerID: playerID, Type: model.CmdSelect, Selections: []int{idx}}, true
			}
		}
	}
	// 狂战士“撕裂”在可选响应中优先（高爆发）
	if me.Role == "berserker" && me.Gem >= 1 {
		for idx, opt := range prompt.Options {
			if strings.Contains(opt.Label, "撕裂") || strings.Contains(opt.ID, "berserker_tear") {
				return model.PlayerAction{PlayerID: playerID, Type: model.CmdSelect, Selections: []int{idx}}, true
			}
		}
	}

	if skipIndex >= 0 {
		return model.PlayerAction{PlayerID: playerID, Type: model.CmdSelect, Selections: []int{skipIndex}}, true
	}
	if len(prompt.Options) > 0 {
		return model.PlayerAction{PlayerID: playerID, Type: model.CmdSelect, Selections: []int{0}}, true
	}
	return BuildFallbackAction(playerID, prompt, state)
}

func decideChooseExtract(playerID string, state StateSnapshot, prompt *model.Prompt, me PlayerSnapshot) (model.PlayerAction, bool) {
	if len(prompt.Options) == 0 {
		return BuildFallbackAction(playerID, prompt, state)
	}
	minPick := prompt.Min
	maxPick := prompt.Max
	if minPick <= 0 {
		minPick = 1
	}
	if maxPick <= 0 {
		maxPick = minPick
	}

	preferGem := me.Gem <= me.Crystal
	var gemIdx, crystalIdx []int
	for idx, opt := range prompt.Options {
		if strings.Contains(opt.Label, "红宝石") {
			gemIdx = append(gemIdx, idx)
		} else {
			crystalIdx = append(crystalIdx, idx)
		}
	}
	pool := crystalIdx
	backup := gemIdx
	if preferGem {
		pool = gemIdx
		backup = crystalIdx
	}

	selections := make([]int, 0, maxPick)
	for _, idx := range pool {
		if len(selections) >= maxPick {
			break
		}
		selections = append(selections, idx)
	}
	for _, idx := range backup {
		if len(selections) >= maxPick {
			break
		}
		selections = append(selections, idx)
	}
	if len(selections) < minPick {
		for i := 0; i < len(prompt.Options) && len(selections) < minPick; i++ {
			if !containsInt(selections, i) {
				selections = append(selections, i)
			}
		}
	}
	return model.PlayerAction{PlayerID: playerID, Type: model.CmdSelect, Selections: selections[:minPick]}, true
}

func decideConfirmChoice(playerID string, state StateSnapshot, prompt *model.Prompt, me PlayerSnapshot) (model.PlayerAction, bool) {
	handPressure := float64(len(me.Hand)) / float64(estimateMaxHand(me.Role))
	enemyThreat := estimateEnemyThreat(state, playerID)

	if prompt.ChoiceType == "assassin_stealth_draw" {
		// 暗杀者【潜行】：手牌压力高时优先不摸，避免刚入潜行就因上限-1触发爆牌；
		// 否则偏向摸1，保留技能的资源转换收益。
		if handPressure >= 0.9 && len(prompt.Options) >= 2 {
			return model.PlayerAction{PlayerID: playerID, Type: model.CmdSelect, Selections: []int{1}}, true
		}
		return model.PlayerAction{PlayerID: playerID, Type: model.CmdSelect, Selections: []int{0}}, true
	}

	// 通用偏好：优先非跳过项
	bestIdx := -1
	bestScore := -999.0
	for idx, opt := range prompt.Options {
		score := 0.0
		label := opt.Label
		id := opt.ID

		if strings.Contains(label, "跳过") || strings.Contains(label, "不发动") || strings.Contains(label, "不使用") || id == "skip" || id == "cancel" {
			score -= 0.8
		}
		if strings.Contains(label, "伤害") || strings.Contains(label, "攻击") || strings.Contains(label, "额外") {
			score += 0.6
		}
		if strings.Contains(label, "摸") {
			// 手牌接近满时，降低摸牌偏好
			score += 0.2 - handPressure
		}
		if strings.Contains(label, "放弃行动") || strings.Contains(label, "跳过回合") {
			score -= 0.5
			if handPressure >= 0.92 && enemyThreat >= 0.7 {
				score += 0.8
			}
		}
		if strings.Contains(label, "使用 0 点治疗") || strings.Contains(label, "不使用治疗") {
			score -= 0.4
			if handPressure < 0.65 && enemyThreat < 0.6 {
				score += 0.3
			}
		}
		// 治疗选择：倾向于较高治疗值（降低即刻伤害与扣卡）
		if strings.Contains(label, "使用 ") && strings.Contains(label, "点治疗") {
			x := extractLeadingInt(label)
			score += float64(x) * (0.25 + handPressure*0.2)
		}

		// 战绩区4星石分配：按阵营短板选
		if strings.Contains(prompt.Message, "战绩区已有4个星石") {
			needGem := campNeedGem(state, me.Camp)
			if strings.Contains(label, "宝石") && needGem {
				score += 0.8
			}
			if strings.Contains(label, "水晶") && !needGem {
				score += 0.8
			}
		}

		if score > bestScore {
			bestScore = score
			bestIdx = idx
		}
	}

	if bestIdx < 0 {
		return BuildFallbackAction(playerID, prompt, state)
	}
	return model.PlayerAction{PlayerID: playerID, Type: model.CmdSelect, Selections: []int{bestIdx}}, true
}

func estimateHitProbability(state StateSnapshot, me PlayerSnapshot, target PlayerSnapshot, attackElement model.Element, memory *Memory) float64 {
	// 从“手牌数量 + 公开行为偏好 + 场上圣盾”估算命中概率
	pDefend := clamp(0.12*float64(target.HandCount), 0.05, 0.58)
	pCounter := clamp(0.10*float64(target.HandCount), 0.03, 0.48)

	// 暗灭无法被应战，只保留防御概率
	if attackElement == model.ElementDark {
		pCounter = 0
	}

	// 场上已有圣盾时，命中概率显著下降（非烈风等特例）
	for _, fc := range target.Field {
		if fc.Effect == model.EffectShield {
			pDefend = math.Max(pDefend, 0.78)
			break
		}
	}

	// 结合公开信息偏差
	if memory != nil {
		pDefend += memory.DefendBias(target.ID)
		pCounter += memory.AttackBias(target.ID)
	}

	// 同阵营资源紧缺时，偏向高命中目标
	hit := 1.0 - pDefend - 0.7*pCounter
	return clamp(hit, 0.05, 0.95)
}

func estimateEnemyThreat(state StateSnapshot, playerID string) float64 {
	me, ok := state.Players[playerID]
	if !ok {
		return 0.5
	}
	maxThreat := 0.0
	for _, p := range state.Players {
		if p.ID == playerID || p.Camp == me.Camp {
			continue
		}
		base := roleThreatWeight(p.Role)
		base += 0.08 * float64(p.HandCount)
		base += 0.05 * float64(p.Gem+p.Crystal)
		if base > maxThreat {
			maxThreat = base
		}
	}
	return clamp(maxThreat, 0.2, 1.2)
}

func roleThreatWeight(role string) float64 {
	switch role {
	case "berserker", "holy_lancer", "archer", "arbiter", "crimson_knight":
		return 1.0
	case "assassin", "valkyrie", "elementalist", "war_homunculus", "onmyoji":
		return 0.85
	case "blaze_witch":
		return 0.9
	case "sage":
		return 0.92
	case "magic_bow":
		return 0.9
	case "magic_lancer":
		return 0.93
	case "spirit_caster":
		return 0.88
	case "bard":
		return 0.8
	case "hero":
		return 0.94
	case "fighter":
		return 0.91
	case "holy_bow":
		return 0.9
	case "soul_sorcerer":
		return 0.88
	case "moon_goddess":
		return 0.9
	case "blood_priestess":
		return 0.89
	case "butterfly_dancer":
		return 0.9
	case "prayer_master", "priest":
		return 0.75
	default:
		return 0.65
	}
}

func estimateCampResourceNeed(state StateSnapshot, camp string) float64 {
	var gems, crystals int
	if camp == string(model.RedCamp) {
		gems, crystals = state.RedGems, state.RedCrystals
	} else {
		gems, crystals = state.BlueGems, state.BlueCrystals
	}
	total := gems + crystals
	switch {
	case total <= 1:
		return 1.0
	case total == 2:
		return 0.8
	case total == 3:
		return 0.5
	default:
		return 0.2
	}
}

func campNeedGem(state StateSnapshot, camp string) bool {
	if camp == string(model.RedCamp) {
		return state.RedGems <= state.RedCrystals
	}
	return state.BlueGems <= state.BlueCrystals
}

func estimateBerserkerBonus(me PlayerSnapshot, card model.Card, target PlayerSnapshot, handLen int) float64 {
	if me.Role != "berserker" {
		return 0
	}
	bonus := 1.0 // 狂化基础+1
	// 攻击牌会先打出，命中判定时手牌减少1，因此当前>=5才能触发额外+1
	if handLen >= 5 {
		bonus += 1.0
	}
	// 独有牌“血影狂刀”额外伤害（命中后看目标手牌）
	if card.MatchExclusive("狂战士", "血影狂刀") {
		if target.HandCount == 2 {
			bonus += 2.0
		} else if target.HandCount == 3 {
			bonus += 1.0
		}
	}
	return bonus
}

func pickMagicTargets(state StateSnapshot, me PlayerSnapshot, card model.Card) []string {
	// 默认：伤害/控制打敌人，防御类打自己。
	if strings.Contains(card.Name, "圣盾") || strings.Contains(card.Name, "圣光") {
		return []string{me.ID}
	}
	var enemyIDs []string
	for pid, p := range state.Players {
		if p.Camp != me.Camp && pid != me.ID {
			enemyIDs = append(enemyIDs, pid)
		}
	}
	if len(enemyIDs) == 0 {
		return []string{me.ID}
	}
	return enemyIDs
}

func pickTargetsForSkill(state StateSnapshot, me PlayerSnapshot, sk AvailableSkill) []string {
	all := make([]PlayerSnapshot, 0, len(state.Players))
	for _, p := range state.Players {
		all = append(all, p)
	}
	// target_type: 0=None, 1=Self, 2=Enemy, 3=Ally, 4=AllySelf, 5=Any, 6=Specific
	switch sk.TargetType {
	case int(model.TargetNone):
		return nil
	case int(model.TargetSelf):
		return []string{me.ID}
	case int(model.TargetEnemy):
		for _, p := range all {
			if p.Camp != me.Camp {
				return []string{p.ID}
			}
		}
	case int(model.TargetAlly):
		for _, p := range all {
			if p.Camp == me.Camp && p.ID != me.ID {
				return []string{p.ID}
			}
		}
	case int(model.TargetAllySelf):
		return []string{me.ID}
	default:
		return []string{me.ID}
	}
	return nil
}

func firstOrEmpty(arr []string) string {
	if len(arr) == 0 {
		return ""
	}
	return arr[0]
}

func botPlayableCards(me PlayerSnapshot) []model.Card {
	blessings := make([]model.Card, 0)
	for _, fc := range me.Field {
		if fc == nil || fc.Mode != model.FieldCover || fc.Effect != model.EffectElfBlessing {
			continue
		}
		blessings = append(blessings, fc.Card)
	}
	if len(blessings) == 0 {
		return me.Hand
	}
	out := make([]model.Card, 0, len(me.Hand)+len(blessings))
	out = append(out, me.Hand...)
	out = append(out, blessings...)
	return out
}

func pickFirstEnemy(state StateSnapshot, playerID string) string {
	me, ok := state.Players[playerID]
	if !ok {
		return ""
	}
	for pid, p := range state.Players {
		if pid == playerID {
			continue
		}
		if p.Camp != me.Camp {
			return pid
		}
	}
	return ""
}

func BuildFallbackAction(playerID string, prompt *model.Prompt, state StateSnapshot) (model.PlayerAction, bool) {
	if prompt == nil {
		return model.PlayerAction{}, false
	}
	optionIDs := map[string]bool{}
	for _, o := range prompt.Options {
		optionIDs[o.ID] = true
	}
	if optionIDs["take"] {
		return model.PlayerAction{
			PlayerID:  playerID,
			Type:      model.CmdRespond,
			ExtraArgs: []string{"take"},
		}, true
	}
	if prompt.Type == model.PromptChooseCards {
		if len(prompt.Options) == 0 {
			return model.PlayerAction{}, false
		}
		need := prompt.Min
		if need <= 0 {
			need = 1
		}
		selections := make([]int, 0, need)
		for optIdx, opt := range prompt.Options {
			if len(selections) >= need {
				break
			}
			idx, err := strconv.Atoi(opt.ID)
			if err != nil {
				// 部分提示项不是“手牌索引ID”，退化为选项序号。
				idx = optIdx
			}
			if !containsInt(selections, idx) {
				selections = append(selections, idx)
			}
		}
		if len(selections) == 0 {
			return model.PlayerAction{}, false
		}
		return model.PlayerAction{
			PlayerID:   playerID,
			Type:       model.CmdSelect,
			Selections: selections,
		}, true
	}
	if len(prompt.Options) > 0 {
		if optionIDs["special"] {
			if hasPromptOption(prompt, "extract") {
				return model.PlayerAction{PlayerID: playerID, Type: model.CmdExtract}, true
			}
			if hasPromptOption(prompt, "buy") {
				return model.PlayerAction{PlayerID: playerID, Type: model.CmdBuy}, true
			}
			if hasPromptOption(prompt, "synthesize") {
				return model.PlayerAction{PlayerID: playerID, Type: model.CmdSynthesize}, true
			}
		}
		return model.PlayerAction{
			PlayerID:   playerID,
			Type:       model.CmdSelect,
			Selections: []int{0},
		}, true
	}
	me, hasMe := state.Players[playerID]
	if hasMe {
		for idx, card := range botPlayableCards(me) {
			if card.Type == model.CardTypeAttack {
				target := pickFirstEnemy(state, playerID)
				if target != "" {
					return model.PlayerAction{PlayerID: playerID, Type: model.CmdAttack, CardIndex: idx, TargetID: target}, true
				}
			}
		}
	}
	return model.PlayerAction{PlayerID: playerID, Type: model.CmdPass}, true
}

func collectCardIndices(hand []model.Card, t model.CardType) []int {
	out := make([]int, 0, len(hand))
	for i, c := range hand {
		if c.Type == t {
			out = append(out, i)
		}
	}
	return out
}

func collectDefendCards(hand []model.Card) []int {
	out := make([]int, 0)
	for i, c := range hand {
		if c.Type == model.CardTypeMagic && strings.Contains(c.Name, "圣光") {
			out = append(out, i)
		}
	}
	return out
}

func collectCounterCards(hand []model.Card, attackElement model.Element) []int {
	out := make([]int, 0)
	for i, c := range hand {
		if c.Type != model.CardTypeAttack {
			continue
		}
		if attackElement == model.ElementDark {
			continue
		}
		if c.Element == attackElement || c.Element == model.ElementDark {
			out = append(out, i)
		}
	}
	return out
}

func pickBestCounterCard(candidates []int, hand []model.Card, required model.Element) int {
	if len(candidates) == 0 {
		return -1
	}
	// 优先同系，保留暗灭作为通用应战牌
	for _, idx := range candidates {
		if hand[idx].Element == required {
			return idx
		}
	}
	return candidates[0]
}

func pickCounterTarget(state StateSnapshot, candidateIDs []string) string {
	bestID := ""
	bestHand := 999
	for _, id := range candidateIDs {
		p, ok := state.Players[id]
		if !ok {
			continue
		}
		if p.HandCount < bestHand {
			bestHand = p.HandCount
			bestID = id
		}
	}
	return bestID
}

func estimateMaxHand(role string) int {
	maxHandByRoleOnce.Do(func() {
		maxHandByRole = map[string]int{}
		for _, ch := range data.GetCharacters() {
			maxHandByRole[ch.ID] = ch.MaxHand
		}
	})
	if v, ok := maxHandByRole[role]; ok && v > 0 {
		return v
	}
	return 6
}

func estimateMaxEnergy(role string) int {
	if role == "sage" {
		return 4
	}
	return 3
}

var (
	maxHandByRoleOnce sync.Once
	maxHandByRole     map[string]int
)

func cardKeepValue(role string, c model.Card, handLen int) float64 {
	v := 0.0
	if c.Type == model.CardTypeAttack {
		v += 1.0
	}
	if c.Type == model.CardTypeMagic {
		v += 0.8
	}
	if c.Element == model.ElementDark {
		v += 1.4
	}
	if strings.Contains(c.Name, "圣光") || strings.Contains(c.Name, "圣盾") {
		v += 1.8
	}
	if c.ExclusiveSkill1 != "" || c.ExclusiveSkill2 != "" {
		v += 0.8
	}
	// 角色特化：暗杀者保留水系，狂战士偏保留攻击牌
	if role == "assassin" && c.Element == model.ElementWater {
		v += 1.0
	}
	if role == "berserker" && c.Type == model.CardTypeAttack && handLen <= 5 {
		v += 0.4
	}
	return v
}

func containsInt(arr []int, x int) bool {
	for _, v := range arr {
		if v == x {
			return true
		}
	}
	return false
}

func extractLeadingInt(s string) int {
	// 例如 "使用 2 点治疗"
	fields := strings.Fields(s)
	for i := 0; i < len(fields); i++ {
		if n, err := strconv.Atoi(fields[i]); err == nil {
			return n
		}
	}
	return 0
}

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
