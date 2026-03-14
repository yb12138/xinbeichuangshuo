package rules

import (
	"fmt"
	"testing"

	"starcup-engine/internal/model"
)

func hasAdjacentSimilarCards(deck []model.Card) bool {
	for i := 1; i < len(deck); i++ {
		if cardsAreSimilar(deck[i-1], deck[i]) {
			return true
		}
	}
	return false
}

func maxAdjacentSimilarRun(deck []model.Card) int {
	if len(deck) == 0 {
		return 0
	}
	run := 1
	maxRun := 1
	for i := 1; i < len(deck); i++ {
		if cardsAreSimilar(deck[i-1], deck[i]) {
			run++
			if run > maxRun {
				maxRun = run
			}
			continue
		}
		run = 1
	}
	return maxRun
}

func countTypeInPrefix(deck []model.Card, cardType model.CardType, prefix int) int {
	if prefix > len(deck) {
		prefix = len(deck)
	}
	count := 0
	for i := 0; i < prefix; i++ {
		if deck[i].Type == cardType {
			count++
		}
	}
	return count
}

func countDistinctElementsInPrefix(deck []model.Card, prefix int) int {
	if prefix > len(deck) {
		prefix = len(deck)
	}
	seen := map[model.Element]struct{}{}
	for i := 0; i < prefix; i++ {
		if deck[i].Element == "" {
			continue
		}
		seen[deck[i].Element] = struct{}{}
	}
	return len(seen)
}

func countDistinctFactionsInPrefix(deck []model.Card, prefix int) int {
	if prefix > len(deck) {
		prefix = len(deck)
	}
	seen := map[string]struct{}{}
	for i := 0; i < prefix; i++ {
		if deck[i].Faction == "" {
			continue
		}
		seen[deck[i].Faction] = struct{}{}
	}
	return len(seen)
}

func countDistinctExclusiveSignaturesInPrefix(deck []model.Card, prefix int) int {
	if prefix > len(deck) {
		prefix = len(deck)
	}
	seen := map[string]struct{}{}
	for i := 0; i < prefix; i++ {
		sig := cardExclusiveSignature(deck[i])
		if sig == "" {
			continue
		}
		seen[sig] = struct{}{}
	}
	return len(seen)
}

func TestShuffle_DeclusterBalancedDeckAvoidsAdjacentSimilarity(t *testing.T) {
	base := make([]model.Card, 0, 48)
	nextID := 1
	addGroup := func(name string, ele model.Element, count int) {
		for i := 0; i < count; i++ {
			base = append(base, model.Card{
				ID:      fmt.Sprintf("%d", nextID),
				Name:    name,
				Type:    model.CardTypeAttack,
				Element: ele,
				Damage:  2,
			})
			nextID++
		}
	}
	addGroup("火焰斩", model.ElementFire, 12)
	addGroup("水涟斩", model.ElementWater, 12)
	addGroup("地裂斩", model.ElementEarth, 12)
	addGroup("风神斩", model.ElementWind, 12)

	for seed := int64(1); seed <= 16; seed++ {
		restore := SetDeterministicShuffleSeedForTesting(seed)
		shuffled := Shuffle(base)
		restore()

		if len(shuffled) != len(base) {
			t.Fatalf("seed=%d: shuffled size mismatch, got=%d want=%d", seed, len(shuffled), len(base))
		}
		if hasAdjacentSimilarCards(shuffled) {
			t.Fatalf("seed=%d: expected no adjacent similar cards for balanced deck", seed)
		}
	}
}

func TestShuffle_AllSameCardsKeepsDeckIntact(t *testing.T) {
	base := make([]model.Card, 0, 10)
	for i := 0; i < 10; i++ {
		base = append(base, model.Card{
			ID:      fmt.Sprintf("same-%d", i),
			Name:    "火焰斩",
			Type:    model.CardTypeAttack,
			Element: model.ElementFire,
			Damage:  2,
		})
	}

	restore := SetDeterministicShuffleSeedForTesting(20260310)
	shuffled := Shuffle(base)
	restore()

	if len(shuffled) != len(base) {
		t.Fatalf("shuffled size mismatch, got=%d want=%d", len(shuffled), len(base))
	}
	if !hasAdjacentSimilarCards(shuffled) {
		t.Fatalf("all-same deck should still have adjacent similar cards")
	}
}

func TestShuffle_ImbalancedDeckKeepsSimilarRunShort(t *testing.T) {
	base := make([]model.Card, 0, 15)
	nextID := 1
	addGroup := func(name string, ele model.Element, count int) {
		for i := 0; i < count; i++ {
			base = append(base, model.Card{
				ID:      fmt.Sprintf("%d", nextID),
				Name:    name,
				Type:    model.CardTypeAttack,
				Element: ele,
				Damage:  2,
			})
			nextID++
		}
	}

	addGroup("火焰斩", model.ElementFire, 9)
	addGroup("水涟斩", model.ElementWater, 3)
	addGroup("风神斩", model.ElementWind, 3)

	for seed := int64(1); seed <= 24; seed++ {
		restore := SetDeterministicShuffleSeedForTesting(seed)
		shuffled := Shuffle(base)
		restore()

		if len(shuffled) != len(base) {
			t.Fatalf("seed=%d: shuffled size mismatch, got=%d want=%d", seed, len(shuffled), len(base))
		}
		if got := maxAdjacentSimilarRun(shuffled); got > 2 {
			t.Fatalf("seed=%d: expected max similar run <=2, got=%d", seed, got)
		}
	}
}

func TestShuffle_InitDeckFrontSegmentDistributionIsBalanced(t *testing.T) {
	base := InitDeck()
	if len(base) == 0 {
		t.Fatalf("init deck should not be empty")
	}

	for seed := int64(1); seed <= 20; seed++ {
		restore := SetDeterministicShuffleSeedForTesting(seed)
		shuffled := Shuffle(base)
		restore()

		if len(shuffled) != len(base) {
			t.Fatalf("seed=%d: shuffled size mismatch, got=%d want=%d", seed, len(shuffled), len(base))
		}

		// 前 12 张不能“几乎全攻击”，至少要有 2 张法术。
		attackIn12 := countTypeInPrefix(shuffled, model.CardTypeAttack, 12)
		if attackIn12 > 10 {
			t.Fatalf("seed=%d: expected front 12 cards not attack-heavy, got attack=%d magic=%d", seed, attackIn12, 12-attackIn12)
		}

		// 前段应覆盖较丰富的系别/命格/独有技签名，避免体验过于单一。
		if got := countDistinctElementsInPrefix(shuffled, 14); got < 5 {
			t.Fatalf("seed=%d: expected >=5 distinct elements in first 14 cards, got=%d", seed, got)
		}
		if got := countDistinctFactionsInPrefix(shuffled, 15); got < 4 {
			t.Fatalf("seed=%d: expected >=4 distinct factions in first 15 cards, got=%d", seed, got)
		}
		if got := countDistinctExclusiveSignaturesInPrefix(shuffled, 16); got < 6 {
			t.Fatalf("seed=%d: expected >=6 distinct exclusive signatures in first 16 cards, got=%d", seed, got)
		}
	}
}
