package rules

import (
	"fmt"
	"math"
	"math/rand"
	"starcup-engine/internal/model"
	"sync"
	"time"
)

var (
	shuffleConfigMu            sync.Mutex
	shuffleDeterministic       bool
	shuffleDeterministicSeed   int64
	shuffleDeterministicOffset int64
)

// InitDeck 初始化牌库 (根据完整卡牌设定)
func InitDeck() []model.Card {
	deck := make([]model.Card, 0, 400)
	idCounter := 1

	// 辅助函数 - 带命格和独有技的卡牌
	addExclusiveCards := func(name string, cType model.CardType, element model.Element, count int, damage int, desc, faction, char1, char2, skill1, skill2 string) {
		for i := 0; i < count; i++ {
			deck = append(deck, model.Card{
				ID:              fmt.Sprintf("%d", idCounter),
				Name:            name,
				Type:            cType,
				Element:         element,
				Damage:          damage,
				Description:     desc,
				Faction:         faction,
				ExclusiveChar1:  char1,
				ExclusiveChar2:  char2,
				ExclusiveSkill1: skill1,
				ExclusiveSkill2: skill2,
			})
			idCounter++
		}
	}

	// 1. 火焰斩 (Attack, 火)
	// 幻命格 * 5
	addExclusiveCards("火焰斩", model.CardTypeAttack, model.ElementFire, 5, 2, "基础攻击，持有者可使用独有技",
		"幻", "封印师", "灵魂术士", "火之封印", "灵魂震爆")
	// 咏命格 * 4
	addExclusiveCards("火焰斩", model.CardTypeAttack, model.ElementFire, 4, 2, "基础攻击，持有者可使用独有技",
		"咏", "祈祷师", "元素师", "威力赐福", "火球")
	// 血命格 * 4
	addExclusiveCards("火焰斩", model.CardTypeAttack, model.ElementFire, 4, 2, "基础攻击，持有者可使用独有技",
		"血", "狂战士", "血之巫女", "血腥咆哮", "血之悲鸣")
	// 技命格 * 4
	addExclusiveCards("火焰斩", model.CardTypeAttack, model.ElementFire, 4, 2, "基础攻击，持有者可使用独有技",
		"技", "风之剑圣", "神箭手", "疾风技", "闪光陷阱")
	// 圣命格 * 4
	addExclusiveCards("火焰斩", model.CardTypeAttack, model.ElementFire, 4, 2, "基础攻击，持有者可使用独有技",
		"圣", "圣女", "天使", "治疗术", "天使之墙")

	// 2. 水涟斩 (Attack, 水)
	// 幻命格 * 4 (水之封印/灵魂震爆 * 2, 水之封印/灵魂赐予 * 2)
	addExclusiveCards("水涟斩", model.CardTypeAttack, model.ElementWater, 2, 2, "基础攻击，持有者可使用独有技",
		"幻", "封印师", "灵魂术士", "水之封印", "灵魂震爆")
	addExclusiveCards("水涟斩", model.CardTypeAttack, model.ElementWater, 2, 2, "基础攻击，持有者可使用独有技",
		"幻", "封印师", "灵魂术士", "水之封印", "灵魂赐予")
	// 咏命格 * 6
	addExclusiveCards("水涟斩", model.CardTypeAttack, model.ElementWater, 6, 2, "基础攻击，持有者可使用独有技",
		"咏", "祈祷师", "元素师", "威力赐福", "冰冻")
	// 血命格 * 4
	addExclusiveCards("水涟斩", model.CardTypeAttack, model.ElementWater, 4, 2, "基础攻击，持有者可使用独有技",
		"血", "狂战士", "血之巫女", "血影狂刀", "血之悲鸣")
	// 技命格 * 4
	addExclusiveCards("水涟斩", model.CardTypeAttack, model.ElementWater, 4, 2, "基础攻击，持有者可使用独有技",
		"技", "风之剑圣", "神箭手", "疾风技", "闪光陷阱")
	// 圣命格 * 3
	addExclusiveCards("水涟斩", model.CardTypeAttack, model.ElementWater, 3, 2, "基础攻击，持有者可使用独有技",
		"圣", "圣女", "天使", "治愈之光", "天使之墙")

	// 3. 地裂斩 (Attack, 地)
	// 幻命格 * 4 (地之封印/灵魂震爆 * 2, 地之封印/灵魂赐予 * 2)
	addExclusiveCards("地裂斩", model.CardTypeAttack, model.ElementEarth, 2, 2, "基础攻击，持有者可使用独有技",
		"幻", "封印师", "灵魂术士", "地之封印", "灵魂震爆")
	addExclusiveCards("地裂斩", model.CardTypeAttack, model.ElementEarth, 2, 2, "基础攻击，持有者可使用独有技",
		"幻", "封印师", "灵魂术士", "地之封印", "灵魂赐予")
	// 咏命格 * 4 (威力赐福/陨石 * 2, 迅捷赐福/陨石 * 2)
	addExclusiveCards("地裂斩", model.CardTypeAttack, model.ElementEarth, 2, 2, "基础攻击，持有者可使用独有技",
		"咏", "祈祷师", "元素师", "威力赐福", "陨石")
	addExclusiveCards("地裂斩", model.CardTypeAttack, model.ElementEarth, 2, 2, "基础攻击，持有者可使用独有技",
		"咏", "祈祷师", "元素师", "迅捷赐福", "陨石")
	// 血命格 * 5 (血影狂刀/血之悲鸣 * 3, 血腥咆哮/血之悲鸣 * 2)
	addExclusiveCards("地裂斩", model.CardTypeAttack, model.ElementEarth, 3, 2, "基础攻击，持有者可使用独有技",
		"血", "狂战士", "血之巫女", "血影狂刀", "血之悲鸣")
	addExclusiveCards("地裂斩", model.CardTypeAttack, model.ElementEarth, 2, 2, "基础攻击，持有者可使用独有技",
		"血", "狂战士", "血之巫女", "血腥咆哮", "血之悲鸣")
	// 技命格 * 5
	addExclusiveCards("地裂斩", model.CardTypeAttack, model.ElementEarth, 5, 2, "基础攻击，持有者可使用独有技",
		"技", "风之剑圣", "神箭手", "烈风技", "精准射击")
	// 圣命格 * 3
	addExclusiveCards("地裂斩", model.CardTypeAttack, model.ElementEarth, 3, 2, "基础攻击，持有者可使用独有技",
		"圣", "圣女", "", "治疗术", "")

	// 4. 风神斩 (Attack, 风)
	// 幻命格 * 4
	addExclusiveCards("风神斩", model.CardTypeAttack, model.ElementWind, 4, 2, "基础攻击，持有者可使用独有技",
		"幻", "封印师", "灵魂术士", "风之封印", "灵魂赐予")
	// 咏命格 * 5
	addExclusiveCards("风神斩", model.CardTypeAttack, model.ElementWind, 5, 2, "基础攻击，持有者可使用独有技",
		"咏", "祈祷师", "元素师", "迅捷赐福", "风刃")
	// 血命格 * 4
	addExclusiveCards("风神斩", model.CardTypeAttack, model.ElementWind, 4, 2, "基础攻击，持有者可使用独有技",
		"血", "狂战士", "血之巫女", "血腥咆哮", "血之悲鸣")
	// 技命格 * 5 (烈风技/精准射击 * 2, 疾风技/精准射击 * 3)
	addExclusiveCards("风神斩", model.CardTypeAttack, model.ElementWind, 2, 2, "基础攻击，持有者可使用独有技",
		"技", "风之剑圣", "神箭手", "烈风技", "精准射击")
	addExclusiveCards("风神斩", model.CardTypeAttack, model.ElementWind, 3, 2, "基础攻击，持有者可使用独有技",
		"技", "风之剑圣", "神箭手", "疾风技", "精准射击")
	// 圣命格 * 3
	addExclusiveCards("风神斩", model.CardTypeAttack, model.ElementWind, 3, 2, "基础攻击，持有者可使用独有技",
		"圣", "圣女", "天使", "治愈之光", "天使之墙")

	// 5. 雷光斩 (Attack, 雷)
	// 幻命格 * 5
	addExclusiveCards("雷光斩", model.CardTypeAttack, model.ElementThunder, 5, 2, "基础攻击，持有者可使用独有技",
		"幻", "封印师", "灵魂术士", "雷之封印", "灵魂震爆")
	// 咏命格 * 4
	addExclusiveCards("雷光斩", model.CardTypeAttack, model.ElementThunder, 4, 2, "基础攻击，持有者可使用独有技",
		"咏", "祈祷师", "元素师", "迅捷赐福", "雷击")
	// 血命格 * 4
	addExclusiveCards("雷光斩", model.CardTypeAttack, model.ElementThunder, 4, 2, "基础攻击，持有者可使用独有技",
		"血", "狂战士", "血之巫女", "血影狂刀", "血之悲鸣")
	// 技命格 * 4 (烈风技/精准射击 * 2, 疾风技/精准射击 * 2)
	addExclusiveCards("雷光斩", model.CardTypeAttack, model.ElementThunder, 2, 2, "基础攻击，持有者可使用独有技",
		"技", "风之剑圣", "神箭手", "烈风技", "精准射击")
	addExclusiveCards("雷光斩", model.CardTypeAttack, model.ElementThunder, 2, 2, "基础攻击，持有者可使用独有技",
		"技", "风之剑圣", "神箭手", "疾风技", "精准射击")
	// 圣命格 * 4
	addExclusiveCards("雷光斩", model.CardTypeAttack, model.ElementThunder, 4, 2, "基础攻击，持有者可使用独有技",
		"圣", "圣女", "天使", "治疗术", "天使之墙")

	// 6. 暗灭 (Attack, 暗)
	// 咏命格 * 2
	addExclusiveCards("暗灭", model.CardTypeAttack, model.ElementDark, 2, 2, "无法应战，只能防守",
		"咏", "", "", "", "")
	// 圣命格 * 4
	addExclusiveCards("暗灭", model.CardTypeAttack, model.ElementDark, 4, 2, "无法应战，只能防守",
		"圣", "", "", "", "")

	// 7. 中毒 (Magic)
	// 地命格 * 1
	addExclusiveCards("中毒", model.CardTypeMagic, model.ElementEarth, 1, 0, "目标获得中毒状态",
		"咏", "", "", "", "")
	// 水命格 * 3 (幻 * 1, 技 * 1, 圣 * 1)
	addExclusiveCards("中毒", model.CardTypeMagic, model.ElementWater, 1, 0, "目标获得中毒状态",
		"幻", "", "", "", "")
	addExclusiveCards("中毒", model.CardTypeMagic, model.ElementWater, 1, 0, "目标获得中毒状态",
		"技", "", "", "", "")
	addExclusiveCards("中毒", model.CardTypeMagic, model.ElementWater, 1, 0, "目标获得中毒状态",
		"圣", "", "", "", "")
	// 风命格 * 1
	addExclusiveCards("中毒", model.CardTypeMagic, model.ElementWind, 1, 0, "目标获得中毒状态",
		"圣", "", "", "", "")
	// 雷命格 * 1
	addExclusiveCards("中毒", model.CardTypeMagic, model.ElementThunder, 1, 0, "目标获得中毒状态",
		"技", "", "", "", "")

	// 8. 虚弱 (Magic)
	// 地命格 * 1
	addExclusiveCards("虚弱", model.CardTypeMagic, model.ElementEarth, 1, 0, "目标获得虚弱状态",
		"技", "", "", "", "")
	// 水命格 * 2 (咏 * 1, 血 * 1)
	addExclusiveCards("虚弱", model.CardTypeMagic, model.ElementWater, 1, 0, "目标获得虚弱状态",
		"咏", "", "", "", "")
	addExclusiveCards("虚弱", model.CardTypeMagic, model.ElementWater, 1, 0, "目标获得虚弱状态",
		"血", "", "", "", "")
	// 火命格 * 2 (血 * 1, 圣 * 1)
	addExclusiveCards("虚弱", model.CardTypeMagic, model.ElementFire, 1, 0, "目标获得虚弱状态",
		"血", "", "", "", "")
	addExclusiveCards("虚弱", model.CardTypeMagic, model.ElementFire, 1, 0, "目标获得虚弱状态",
		"圣", "", "", "", "")
	// 风命格 * 1
	addExclusiveCards("虚弱", model.CardTypeMagic, model.ElementWind, 1, 0, "目标获得虚弱状态",
		"幻", "", "", "", "")

	// 9. 魔弹 (Magic)
	// 水命格 * 2 (幻 * 1, 血 * 1)
	addExclusiveCards("魔弹", model.CardTypeMagic, model.ElementWater, 1, 2, "造成2点法术伤害",
		"幻", "", "", "", "")
	addExclusiveCards("魔弹", model.CardTypeMagic, model.ElementWater, 1, 2, "造成2点法术伤害",
		"血", "", "", "", "")
	// 火命格 * 1
	addExclusiveCards("魔弹", model.CardTypeMagic, model.ElementFire, 1, 2, "造成2点法术伤害",
		"咏", "", "", "", "")
	// 风命格 * 1
	addExclusiveCards("魔弹", model.CardTypeMagic, model.ElementWind, 1, 2, "造成2点法术伤害",
		"幻", "", "", "", "")
	// 雷命格 * 2 (技 * 1, 圣 * 1)
	addExclusiveCards("魔弹", model.CardTypeMagic, model.ElementThunder, 1, 2, "造成2点法术伤害",
		"技", "", "", "", "")
	addExclusiveCards("魔弹", model.CardTypeMagic, model.ElementThunder, 1, 2, "造成2点法术伤害",
		"圣", "", "", "", "")

	// 10. 圣盾 (Magic，多系分布)
	// 地系 * 3 (幻 * 1, 咏 * 1, 圣 * 1)
	addExclusiveCards("圣盾", model.CardTypeMagic, model.ElementEarth, 1, 0, "抵挡一次伤害",
		"幻", "", "", "", "")
	addExclusiveCards("圣盾", model.CardTypeMagic, model.ElementEarth, 1, 0, "抵挡一次伤害",
		"咏", "", "", "", "")
	addExclusiveCards("圣盾", model.CardTypeMagic, model.ElementEarth, 1, 0, "抵挡一次伤害",
		"圣", "", "", "", "")
	// 火系 * 2 (血 * 1, 技 * 1)
	addExclusiveCards("圣盾", model.CardTypeMagic, model.ElementFire, 1, 0, "抵挡一次伤害",
		"血", "", "", "", "")
	addExclusiveCards("圣盾", model.CardTypeMagic, model.ElementFire, 1, 0, "抵挡一次伤害",
		"技", "", "", "", "")
	// 风系 * 3 (咏 * 1, 血 * 1, 圣 * 1)
	addExclusiveCards("圣盾", model.CardTypeMagic, model.ElementWind, 1, 0, "抵挡一次伤害",
		"咏", "", "", "", "")
	addExclusiveCards("圣盾", model.CardTypeMagic, model.ElementWind, 1, 0, "抵挡一次伤害",
		"血", "", "", "", "")
	addExclusiveCards("圣盾", model.CardTypeMagic, model.ElementWind, 1, 0, "抵挡一次伤害",
		"圣", "", "", "", "")
	// 雷系 * 2 (幻 * 1, 血 * 1)
	addExclusiveCards("圣盾", model.CardTypeMagic, model.ElementThunder, 1, 0, "抵挡一次伤害",
		"幻", "", "", "", "")
	addExclusiveCards("圣盾", model.CardTypeMagic, model.ElementThunder, 1, 0, "抵挡一次伤害",
		"血", "", "", "", "")

	// 11. 圣光 (Magic, 光)
	// 幻命格 * 2
	addExclusiveCards("圣光", model.CardTypeMagic, model.ElementLight, 2, 0, "抵挡伤害或作为响应",
		"幻", "", "", "", "")
	// 血命格 * 3
	addExclusiveCards("圣光", model.CardTypeMagic, model.ElementLight, 3, 0, "抵挡伤害或作为响应",
		"血", "", "", "", "")
	// 技命格 * 3
	addExclusiveCards("圣光", model.CardTypeMagic, model.ElementLight, 3, 0, "抵挡伤害或作为响应",
		"技", "", "", "", "")
	// 圣命格 * 3 (治疗术 * 2, 治愈之光 * 1)
	addExclusiveCards("圣光", model.CardTypeMagic, model.ElementLight, 2, 0, "抵挡伤害或作为响应",
		"圣", "圣女", "", "治疗术", "")
	addExclusiveCards("圣光", model.CardTypeMagic, model.ElementLight, 1, 0, "抵挡伤害或作为响应",
		"圣", "圣女", "", "治愈之光", "")

	return deck
}

// Shuffle 洗牌
func Shuffle(deck []model.Card) []model.Card {
	seed := time.Now().UnixNano()
	shuffleConfigMu.Lock()
	if shuffleDeterministic {
		seed = shuffleDeterministicSeed + shuffleDeterministicOffset
		shuffleDeterministicOffset++
	}
	shuffleConfigMu.Unlock()

	r := rand.New(rand.NewSource(seed))
	if len(deck) <= 1 {
		shuffled := make([]model.Card, len(deck))
		copy(shuffled, deck)
		return shuffled
	}

	const candidateCount = 4
	best := make([]model.Card, len(deck))
	bestScore := -1

	for i := 0; i < candidateCount; i++ {
		candidate := make([]model.Card, len(deck))
		copy(candidate, deck)

		candidateRand := rand.New(rand.NewSource(r.Int63()))
		candidateRand.Shuffle(len(candidate), func(i, j int) {
			candidate[i], candidate[j] = candidate[j], candidate[i]
		})

		// 先做前缀分布均衡，直接压制“前段全攻击/同系集中”等问题。
		balanceDeckDistribution(candidate, candidateRand)
		optimizeAdjacentSimilarCards(candidate, candidateRand)
		// 均衡后做一次轻量三连打散，保证局部体验。
		forceBreakTripleSimilarRuns(candidate, candidateRand)
		breakTripleSimilarRunsStrict(candidate, candidateRand)

		score := shuffleQualityScore(candidate)
		if bestScore < 0 || score < bestScore {
			bestScore = score
			copy(best, candidate)
			if bestScore == 0 {
				break
			}
		}
	}

	return best
}

func cardExclusiveSignature(card model.Card) string {
	if card.ExclusiveSkill1 == "" && card.ExclusiveSkill2 == "" &&
		card.ExclusiveChar1 == "" && card.ExclusiveChar2 == "" {
		return ""
	}
	return card.ExclusiveChar1 + "|" + card.ExclusiveSkill1 + "|" + card.ExclusiveChar2 + "|" + card.ExclusiveSkill2
}

func tailRunLenByType(cards []model.Card, value string) int {
	if len(cards) == 0 {
		return 0
	}
	run := 0
	for i := len(cards) - 1; i >= 0; i-- {
		if string(cards[i].Type) != value {
			break
		}
		run++
	}
	return run
}

func tailRunLenByElement(cards []model.Card, value string) int {
	if len(cards) == 0 || value == "" {
		return 0
	}
	run := 0
	for i := len(cards) - 1; i >= 0; i-- {
		if string(cards[i].Element) != value {
			break
		}
		run++
	}
	return run
}

func tailRunLenByFaction(cards []model.Card, value string) int {
	if len(cards) == 0 || value == "" {
		return 0
	}
	run := 0
	for i := len(cards) - 1; i >= 0; i-- {
		if cards[i].Faction != value {
			break
		}
		run++
	}
	return run
}

func tailRunLenBySignature(cards []model.Card, value string) int {
	if len(cards) == 0 || value == "" {
		return 0
	}
	run := 0
	for i := len(cards) - 1; i >= 0; i-- {
		if cardExclusiveSignature(cards[i]) != value {
			break
		}
		run++
	}
	return run
}

func squaredErrorDelta(currentCount, totalCount, prefixLen, fullLen int) float64 {
	if fullLen <= 0 || totalCount <= 0 {
		return 0
	}
	expected := float64(totalCount*prefixLen) / float64(fullLen)
	before := float64(currentCount) - expected
	after := float64(currentCount+1) - expected
	return (after * after) - (before * before)
}

// balanceDeckDistribution 通过“逐位放牌 + 前缀分布偏差最小化”策略平衡：
// - 攻击/法术比例
// - 七系分布
// - 命格分布
// - 独有技签名分布
// 同时附加相邻重复惩罚，避免局部连续同类。
func balanceDeckDistribution(deck []model.Card, r *rand.Rand) {
	if len(deck) < 2 {
		return
	}
	n := len(deck)
	remaining := make([]model.Card, len(deck))
	copy(remaining, deck)
	result := make([]model.Card, 0, len(deck))

	totalType := map[string]int{}
	totalElement := map[string]int{}
	totalFaction := map[string]int{}
	totalExclusive := map[string]int{}
	for _, c := range remaining {
		totalType[string(c.Type)]++
		totalElement[string(c.Element)]++
		if c.Faction != "" {
			totalFaction[c.Faction]++
		}
		totalExclusive[cardExclusiveSignature(c)]++
	}
	usedType := map[string]int{}
	usedElement := map[string]int{}
	usedFaction := map[string]int{}
	usedExclusive := map[string]int{}

	for len(remaining) > 0 {
		k := len(result) + 1 // 当前要放置的位置（1-based）

		bestCost := math.MaxFloat64
		bestIndices := make([]int, 0, 4)

		for idx, candidate := range remaining {
			typeKey := string(candidate.Type)
			elementKey := string(candidate.Element)
			factionKey := candidate.Faction
			exclusiveKey := cardExclusiveSignature(candidate)

			// 前缀分布偏差（权重：类型 > 系别 > 命格 > 独有技）
			cost := 0.0
			cost += 35.0 * squaredErrorDelta(usedType[typeKey], totalType[typeKey], k, n)
			cost += 12.0 * squaredErrorDelta(usedElement[elementKey], totalElement[elementKey], k, n)
			if factionKey != "" {
				cost += 16.0 * squaredErrorDelta(usedFaction[factionKey], totalFaction[factionKey], k, n)
			}
			cost += 8.0 * squaredErrorDelta(usedExclusive[exclusiveKey], totalExclusive[exclusiveKey], k, n)

			// 相邻/连段惩罚：优先避免连续同类。
			if len(result) > 0 {
				prev := result[len(result)-1]
				if cardsAreSimilar(prev, candidate) {
					cost += 24.0
				}
				if prev.Type == candidate.Type {
					cost += 6.0
					run := tailRunLenByType(result, typeKey)
					if run >= 2 {
						cost += float64(20 * (run - 1))
					}
				}
				if prev.Element == candidate.Element && candidate.Element != "" {
					cost += 10.0
					run := tailRunLenByElement(result, elementKey)
					if run >= 2 {
						cost += float64(16 * (run - 1))
					}
				}
				if prev.Faction == candidate.Faction && candidate.Faction != "" {
					cost += 4.0
					run := tailRunLenByFaction(result, factionKey)
					if run >= 3 {
						cost += float64(8 * (run - 2))
					}
				}
				if exclusiveKey != "" && cardExclusiveSignature(prev) == exclusiveKey {
					cost += 4.0
					run := tailRunLenBySignature(result, exclusiveKey)
					if run >= 2 {
						cost += float64(6 * (run - 1))
					}
				}
			}

			if cost < bestCost-1e-9 {
				bestCost = cost
				bestIndices = []int{idx}
			} else if math.Abs(cost-bestCost) <= 1e-9 {
				bestIndices = append(bestIndices, idx)
			}
		}

		chosenIdx := bestIndices[0]
		if len(bestIndices) > 1 && r != nil {
			chosenIdx = bestIndices[r.Intn(len(bestIndices))]
		}
		chosen := remaining[chosenIdx]
		result = append(result, chosen)

		usedType[string(chosen.Type)]++
		usedElement[string(chosen.Element)]++
		if chosen.Faction != "" {
			usedFaction[chosen.Faction]++
		}
		usedExclusive[cardExclusiveSignature(chosen)]++

		remaining = append(remaining[:chosenIdx], remaining[chosenIdx+1:]...)
	}

	copy(deck, result)
}

func shuffleQualityScore(deck []model.Card) int {
	// 基础相邻相似惩罚（已有）
	score := adjacentSimilarityScore(deck) * 2
	if len(deck) <= 1 {
		return score
	}

	// 连段额外惩罚（类型/系别/命格/独有技签名）
	typeRun, elemRun, factionRun, exRun := 1, 1, 1, 1
	prevEx := cardExclusiveSignature(deck[0])
	for i := 1; i < len(deck); i++ {
		curEx := cardExclusiveSignature(deck[i])
		prev := deck[i-1]
		cur := deck[i]

		if prev.Type == cur.Type {
			typeRun++
			score += 6
			if typeRun >= 3 {
				score += 14 * (typeRun - 2)
			}
		} else {
			typeRun = 1
		}

		if prev.Element != "" && prev.Element == cur.Element {
			elemRun++
			score += 7
			if elemRun >= 3 {
				score += 12 * (elemRun - 2)
			}
		} else {
			elemRun = 1
		}

		if prev.Faction != "" && prev.Faction == cur.Faction {
			factionRun++
			score += 3
			if factionRun >= 4 {
				score += 6 * (factionRun - 3)
			}
		} else {
			factionRun = 1
		}

		if prevEx != "" && prevEx == curEx {
			exRun++
			score += 3
			if exRun >= 3 {
				score += 6 * (exRun - 2)
			}
		} else {
			exRun = 1
		}
		prevEx = curEx
	}

	// 前缀比例偏差惩罚（重点压制“前段全攻击牌”）
	n := len(deck)
	totalType := map[string]int{}
	totalElement := map[string]int{}
	totalFaction := map[string]int{}
	totalExclusive := map[string]int{}
	for _, c := range deck {
		totalType[string(c.Type)]++
		totalElement[string(c.Element)]++
		if c.Faction != "" {
			totalFaction[c.Faction]++
		}
		totalExclusive[cardExclusiveSignature(c)]++
	}
	prefixType := map[string]int{}
	prefixElement := map[string]int{}
	prefixFaction := map[string]int{}
	prefixExclusive := map[string]int{}

	checkpoint := map[int]struct{}{}
	frontLimit := 24
	if frontLimit > n {
		frontLimit = n
	}
	for k := 1; k <= frontLimit; k++ {
		checkpoint[k] = struct{}{}
	}
	checkpoint[n/2] = struct{}{}
	checkpoint[(3*n)/4] = struct{}{}

	for i, c := range deck {
		k := i + 1
		prefixType[string(c.Type)]++
		prefixElement[string(c.Element)]++
		if c.Faction != "" {
			prefixFaction[c.Faction]++
		}
		prefixExclusive[cardExclusiveSignature(c)]++

		if _, ok := checkpoint[k]; !ok || k <= 0 {
			continue
		}
		for key, total := range totalType {
			if total <= 0 {
				continue
			}
			expected := float64(total*k) / float64(n)
			diff := math.Abs(float64(prefixType[key]) - expected)
			score += int(diff*diff*26.0 + 0.5)
		}
		for key, total := range totalElement {
			if total <= 0 {
				continue
			}
			expected := float64(total*k) / float64(n)
			diff := math.Abs(float64(prefixElement[key]) - expected)
			score += int(diff*diff*8.0 + 0.5)
		}
		for key, total := range totalFaction {
			if total <= 0 {
				continue
			}
			expected := float64(total*k) / float64(n)
			diff := math.Abs(float64(prefixFaction[key]) - expected)
			score += int(diff*diff*9.0 + 0.5)
		}
		for key, total := range totalExclusive {
			// 只约束出现次数>=2 的签名，避免稀有签名带来噪音惩罚。
			if total < 2 || key == "" {
				continue
			}
			expected := float64(total*k) / float64(n)
			diff := math.Abs(float64(prefixExclusive[key]) - expected)
			score += int(diff*diff*4.0 + 0.5)
		}
	}

	return score
}

func cardsAreSimilar(a, b model.Card) bool {
	if a.Name != "" && a.Name == b.Name {
		return true
	}
	return a.Type != "" && a.Type == b.Type && a.Element != "" && a.Element == b.Element
}

// declusterAdjacentSimilarCards 在随机洗牌后尽量打散相邻“相似牌”。
// 相似定义：同名，或同类型+同元素。
// 采用“逐位构造 + 相似度优先散列”策略：每一步优先放置与上一张不相似的牌，
// 并倾向先放置“剩余相似牌较多”的牌，尽量避免尾部形成连续同类。
func declusterAdjacentSimilarCards(deck []model.Card, r *rand.Rand) {
	if len(deck) < 3 || r == nil {
		return
	}
	pool := make([]model.Card, len(deck))
	copy(pool, deck)
	result := make([]model.Card, 0, len(deck))

	countSimilarInPool := func(idx int) int {
		if idx < 0 || idx >= len(pool) {
			return 0
		}
		score := 0
		for j := range pool {
			if j == idx {
				continue
			}
			if cardsAreSimilar(pool[idx], pool[j]) {
				score++
			}
		}
		return score
	}

	for len(pool) > 0 {
		candidates := make([]int, 0, len(pool))
		if len(result) == 0 {
			for i := range pool {
				candidates = append(candidates, i)
			}
		} else {
			prev := result[len(result)-1]
			for i := range pool {
				if !cardsAreSimilar(prev, pool[i]) {
					candidates = append(candidates, i)
				}
			}
		}
		if len(candidates) == 0 {
			// 无法继续打散（例如牌库剩余全是同类），退化为普通抽取。
			for i := range pool {
				candidates = append(candidates, i)
			}
		}

		bestScore := -1
		best := make([]int, 0, len(candidates))
		for _, idx := range candidates {
			score := countSimilarInPool(idx)
			if score > bestScore {
				bestScore = score
				best = []int{idx}
			} else if score == bestScore {
				best = append(best, idx)
			}
		}
		chosen := best[r.Intn(len(best))]
		result = append(result, pool[chosen])
		pool = append(pool[:chosen], pool[chosen+1:]...)
	}

	copy(deck, result)
}

func adjacentSimilarityScore(deck []model.Card) int {
	if len(deck) <= 1 {
		return 0
	}
	score := 0
	run := 1
	for i := 1; i < len(deck); i++ {
		if cardsAreSimilar(deck[i-1], deck[i]) {
			score += 10
			run++
			if run >= 3 {
				// 连续三张及以上同类是最差体验，给予更高惩罚权重。
				score += 20 * (run - 2)
			}
			continue
		}
		run = 1
	}
	return score
}

func adjacentSimilarIndices(deck []model.Card) []int {
	if len(deck) <= 1 {
		return nil
	}
	bad := make([]int, 0, len(deck)/2)
	for i := 1; i < len(deck); i++ {
		if cardsAreSimilar(deck[i-1], deck[i]) {
			bad = append(bad, i)
		}
	}
	return bad
}

func optimizeAdjacentSimilarCards(deck []model.Card, r *rand.Rand) {
	if len(deck) < 4 || r == nil {
		return
	}
	best := adjacentSimilarityScore(deck)
	if best == 0 {
		return
	}

	maxAttempts := len(deck) * 20
	for attempt := 0; attempt < maxAttempts && best > 0; attempt++ {
		bad := adjacentSimilarIndices(deck)
		if len(bad) == 0 {
			return
		}
		i := bad[r.Intn(len(bad))]
		j := r.Intn(len(deck))
		if i == j {
			continue
		}
		// 避免在同一局部反复互换导致无效抖动。
		if j+1 == i || i+1 == j {
			continue
		}

		deck[i], deck[j] = deck[j], deck[i]
		score := adjacentSimilarityScore(deck)
		if score <= best {
			best = score
			continue
		}
		deck[i], deck[j] = deck[j], deck[i]
	}
}

func forceBreakTripleSimilarRuns(deck []model.Card, r *rand.Rand) {
	if len(deck) < 3 {
		return
	}
	maxPasses := len(deck) * 2
	for pass := 0; pass < maxPasses; pass++ {
		improved := false
		for i := 2; i < len(deck); i++ {
			if !(cardsAreSimilar(deck[i-2], deck[i-1]) && cardsAreSimilar(deck[i-1], deck[i])) {
				continue
			}
			j := bestSwapIndexForImprovement(deck, i, r)
			if j < 0 {
				continue
			}
			deck[i], deck[j] = deck[j], deck[i]
			improved = true
		}
		if !improved {
			return
		}
	}
}

func bestSwapIndexForImprovement(deck []model.Card, i int, r *rand.Rand) int {
	if i < 0 || i >= len(deck) || len(deck) <= 1 {
		return -1
	}
	baseScore := adjacentSimilarityScore(deck)
	bestScore := baseScore
	bestJ := -1

	start := 0
	if r != nil {
		start = r.Intn(len(deck))
	}
	for step := 0; step < len(deck); step++ {
		j := (start + step) % len(deck)
		if j == i {
			continue
		}
		deck[i], deck[j] = deck[j], deck[i]
		score := adjacentSimilarityScore(deck)
		deck[i], deck[j] = deck[j], deck[i]
		if score < bestScore {
			bestScore = score
			bestJ = j
			if score == 0 {
				return bestJ
			}
		}
	}
	return bestJ
}

func wouldCreateTripleAt(deck []model.Card, idx int) bool {
	if idx >= 2 && cardsAreSimilar(deck[idx-2], deck[idx-1]) && cardsAreSimilar(deck[idx-1], deck[idx]) {
		return true
	}
	if idx >= 1 && idx+1 < len(deck) && cardsAreSimilar(deck[idx-1], deck[idx]) && cardsAreSimilar(deck[idx], deck[idx+1]) {
		return true
	}
	if idx+2 < len(deck) && cardsAreSimilar(deck[idx], deck[idx+1]) && cardsAreSimilar(deck[idx+1], deck[idx+2]) {
		return true
	}
	return false
}

// breakTripleSimilarRunsStrict 严格尝试打散三连“相似牌”。
// 仅关心局部连续体验：若能找到不会在交换位形成新三连的位置，则执行交换。
func breakTripleSimilarRunsStrict(deck []model.Card, r *rand.Rand) {
	if len(deck) < 3 {
		return
	}
	maxPass := len(deck)
	for pass := 0; pass < maxPass; pass++ {
		improved := false
		for i := 2; i < len(deck); i++ {
			if !(cardsAreSimilar(deck[i-2], deck[i-1]) && cardsAreSimilar(deck[i-1], deck[i])) {
				continue
			}

			start := i + 1
			if start >= len(deck) {
				continue
			}
			offset := 0
			if r != nil {
				offset = r.Intn(len(deck) - start)
			}
			bestJ := -1
			for step := 0; step < len(deck)-start; step++ {
				j := start + (offset+step)%(len(deck)-start)
				// i 位置换入 deck[j] 后必须先打破前一张连续相似
				if cardsAreSimilar(deck[i-1], deck[j]) {
					continue
				}
				deck[i], deck[j] = deck[j], deck[i]
				bad := wouldCreateTripleAt(deck, i) || wouldCreateTripleAt(deck, j)
				deck[i], deck[j] = deck[j], deck[i]
				if bad {
					continue
				}
				bestJ = j
				break
			}
			if bestJ >= 0 {
				deck[i], deck[bestJ] = deck[bestJ], deck[i]
				improved = true
			}
		}
		if !improved {
			return
		}
	}
}

// SetDeterministicShuffleSeedForTesting 在测试中启用可复现的洗牌序列。
// 返回 restore 函数用于恢复之前的配置，避免污染其它测试。
func SetDeterministicShuffleSeedForTesting(seed int64) func() {
	shuffleConfigMu.Lock()
	prevDeterministic := shuffleDeterministic
	prevSeed := shuffleDeterministicSeed
	prevOffset := shuffleDeterministicOffset

	shuffleDeterministic = true
	shuffleDeterministicSeed = seed
	shuffleDeterministicOffset = 0
	shuffleConfigMu.Unlock()

	return func() {
		shuffleConfigMu.Lock()
		shuffleDeterministic = prevDeterministic
		shuffleDeterministicSeed = prevSeed
		shuffleDeterministicOffset = prevOffset
		shuffleConfigMu.Unlock()
	}
}

// DrawCards 摸牌逻辑
// 从 deck 摸 count 张牌，如果不够，将 discardPile 洗入 deck
func DrawCards(deck []model.Card, discardPile []model.Card, count int) ([]model.Card, []model.Card, []model.Card) {
	drawn := make([]model.Card, 0, count)

	for i := 0; i < count; i++ {
		if len(deck) == 0 {
			if len(discardPile) == 0 {
				break // 没牌了
			}
			// 重洗弃牌堆
			deck = Shuffle(discardPile)
			discardPile = make([]model.Card, 0)
		}

		card := deck[0]
		deck = deck[1:]
		drawn = append(drawn, card)
	}

	return drawn, deck, discardPile
}
