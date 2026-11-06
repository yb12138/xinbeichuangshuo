package elf_archer_test

import (
	"testing"

	elfarcher "starcup-engine/internal/engine/player/elf_archer"
	skills "starcup-engine/internal/engine/skill"
	"starcup-engine/internal/model"
)

// 审计 精灵射手-5：牌库+弃牌堆不足 3 张时，精灵密仪失败要退还宝石、不进入形态、不留下祝福。
func TestElfRitual_NotEnoughCardsRollsBackGemAndForm(t *testing.T) {
	game := buildElfBlessingGame(t)
	p1 := game.State.Players["p1"]
	p1.Gem = 1
	game.State.Deck = []model.Card{{ID: "d1", Name: "火斩", Type: model.CardTypeAttack, Element: model.ElementFire}}
	game.State.DiscardPile = nil
	handler := skills.GetHandler("elf_ritual")
	ctx := &model.Context{Game: game, User: p1, Flags: map[string]bool{}}
	p1.Gem-- // 模拟框架预扣 CostGem
	if err := handler.Execute(ctx); err == nil {
		t.Fatalf("expected ritual to fail with insufficient cards")
	}
	if p1.Gem != 1 {
		t.Fatalf("expected gem refunded, got %d", p1.Gem)
	}
	if p1.Form == model.FormElfArcherRitual {
		t.Fatalf("ritual form must not be entered on failure")
	}
	if got := elfarcher.CountBlessings(p1); got != 0 {
		t.Fatalf("expected no blessings on failure, got %d", got)
	}
}
