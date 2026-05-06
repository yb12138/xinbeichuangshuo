// gameflow: 开局专属牌测试辅助函数（仅用于测试）。

package engine

import (
	"fmt"

	"starcup-engine/internal/model"
)

// Test helper functions for starter cards (only used in tests).

func makeStarterBardRousingRhapsodyCard(player *model.Player) model.Card {
	return model.Card{
		ID:              fmt.Sprintf("starter-%s-bd_rousing_rhapsody", player.ID),
		Name:            "激昂狂想曲",
		Type:            model.CardTypeMagic,
		Element:         model.ElementDark,
		Faction:         player.Character.Faction,
		Description:     "吟游诗人开局自带专属技能卡",
		ExclusiveChar1:  player.Character.ID,
		ExclusiveSkill1: "激昂狂想曲",
	}
}

func makeStarterBardVictorySymphonyCard(player *model.Player) model.Card {
	return model.Card{
		ID:              fmt.Sprintf("starter-%s-bd_victory_symphony", player.ID),
		Name:            "胜利交响诗",
		Type:            model.CardTypeMagic,
		Element:         model.ElementDark,
		Faction:         player.Character.Faction,
		Description:     "吟游诗人开局自带专属技能卡",
		ExclusiveChar1:  player.Character.ID,
		ExclusiveSkill1: "胜利交响诗",
	}
}

func makeStarterBardHopeFugueCard(player *model.Player) model.Card {
	return model.Card{
		ID:              fmt.Sprintf("starter-%s-bd_hope_fugue", player.ID),
		Name:            "希望赋格曲",
		Type:            model.CardTypeMagic,
		Element:         model.ElementDark,
		Faction:         player.Character.Faction,
		Description:     "吟游诗人开局自带专属技能卡",
		ExclusiveChar1:  player.Character.ID,
		ExclusiveSkill1: "希望赋格曲",
	}
}

func makeStarterSoulLinkCard(player *model.Player) model.Card {
	return model.Card{
		ID:              fmt.Sprintf("starter-%s-soul_link", player.ID),
		Name:            "灵魂链接",
		Type:            model.CardTypeMagic,
		Element:         model.ElementDark,
		Faction:         player.Character.Faction,
		Description:     "灵魂术士开局自带专属技能卡",
		ExclusiveChar1:  player.Character.ID,
		ExclusiveSkill1: "灵魂链接",
	}
}

func makeStarterBloodSharedLifeCard(player *model.Player) model.Card {
	return model.Card{
		ID:              fmt.Sprintf("starter-%s-bp_shared_life", player.ID),
		Name:            "同生共死",
		Type:            model.CardTypeMagic,
		Element:         model.ElementDark,
		Faction:         player.Character.Faction,
		Description:     "血之巫女开局自带专属技能卡",
		ExclusiveChar1:  player.Character.ID,
		ExclusiveSkill1: "同生共死",
	}
}
