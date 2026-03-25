package engine

import (
	"fmt"

	"starcup-engine/internal/model"
)

func ensureExclusiveStarterCard(player *model.Player, skillTitle string, buildCard func() model.Card) bool {
	if player == nil || player.Character == nil || skillTitle == "" || buildCard == nil {
		return false
	}
	charName := player.Character.Name
	for _, c := range player.ExclusiveCards {
		if c.MatchExclusive(charName, skillTitle) {
			return false
		}
	}
	// 兼容旧状态：若该专属卡误在手牌区，迁移到专属卡区。
	for i, c := range player.Hand {
		if !c.MatchExclusive(charName, skillTitle) {
			continue
		}
		player.Hand = append(player.Hand[:i], player.Hand[i+1:]...)
		player.ExclusiveCards = append(player.ExclusiveCards, c)
		return true
	}
	player.ExclusiveCards = append(player.ExclusiveCards, buildCard())
	return true
}

func makeStarterFiveElementsBindCard(player *model.Player) model.Card {
	return model.Card{
		ID:              fmt.Sprintf("starter-%s-five_elements_bind", player.ID),
		Name:            "五系束缚",
		Type:            model.CardTypeMagic,
		Element:         model.ElementLight,
		Faction:         player.Character.Faction,
		Damage:          0,
		Description:     "封印师开局自带专属技能卡",
		ExclusiveChar1:  player.Character.Name,
		ExclusiveSkill1: "五系束缚",
	}
}

func makeStarterRoseCourtyardCard(player *model.Player) model.Card {
	return model.Card{
		ID:              fmt.Sprintf("starter-%s-css_rose_courtyard", player.ID),
		Name:            "血蔷薇庭院",
		Type:            model.CardTypeMagic,
		Element:         model.ElementDark,
		Faction:         player.Character.Faction,
		Damage:          0,
		Description:     "血色剑灵开局自带专属技能卡",
		ExclusiveChar1:  player.Character.Name,
		ExclusiveSkill1: "血蔷薇庭院",
	}
}

func makeStarterHeroTauntCard(player *model.Player) model.Card {
	return model.Card{
		ID:              fmt.Sprintf("starter-%s-hero_taunt", player.ID),
		Name:            "挑衅",
		Type:            model.CardTypeMagic,
		Element:         model.ElementFire,
		Faction:         player.Character.Faction,
		Damage:          0,
		Description:     "勇者开局自带专属技能卡",
		ExclusiveChar1:  player.Character.Name,
		ExclusiveSkill1: "挑衅",
	}
}

func makeStarterSoulLinkCard(player *model.Player) model.Card {
	return model.Card{
		ID:              fmt.Sprintf("starter-%s-soul_link", player.ID),
		Name:            "灵魂链接",
		Type:            model.CardTypeMagic,
		Element:         model.ElementDark,
		Faction:         player.Character.Faction,
		Damage:          0,
		Description:     "灵魂术士开局自带专属技能卡",
		ExclusiveChar1:  player.Character.Name,
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
		Damage:          0,
		Description:     "血之巫女开局自带专属技能卡",
		ExclusiveChar1:  player.Character.Name,
		ExclusiveSkill1: "同生共死",
	}
}

func makeStarterBardRousingRhapsodyCard(player *model.Player) model.Card {
	return model.Card{
		ID:              fmt.Sprintf("starter-%s-bd_rousing_rhapsody", player.ID),
		Name:            "激昂狂想曲",
		Type:            model.CardTypeMagic,
		Element:         model.ElementDark,
		Faction:         player.Character.Faction,
		Damage:          0,
		Description:     "吟游诗人开局自带专属技能卡",
		ExclusiveChar1:  player.Character.Name,
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
		Damage:          0,
		Description:     "吟游诗人开局自带专属技能卡",
		ExclusiveChar1:  player.Character.Name,
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
		Damage:          0,
		Description:     "吟游诗人开局自带专属技能卡",
		ExclusiveChar1:  player.Character.Name,
		ExclusiveSkill1: "希望赋格曲",
	}
}

// ensureStarterRoleCards 为特定角色补充开局自带专属技能卡（置于专属卡区，不占手牌）。
func (e *GameEngine) ensureStarterRoleCards(player *model.Player) {
	if player == nil || player.Character == nil {
		return
	}
	switch player.Character.ID {
	case "sealer":
		if ensureExclusiveStarterCard(player, "五系束缚", func() model.Card {
			return makeStarterFiveElementsBindCard(player)
		}) {
			e.Log(fmt.Sprintf("[Setup] %s 获得开局专属技能卡【五系束缚】（专属卡区）", player.Name))
		}
	case "crimson_sword_spirit":
		if ensureExclusiveStarterCard(player, "血蔷薇庭院", func() model.Card {
			return makeStarterRoseCourtyardCard(player)
		}) {
			e.Log(fmt.Sprintf("[Setup] %s 获得开局专属技能卡【血蔷薇庭院】（专属卡区）", player.Name))
		}
	case "hero":
		if ensureExclusiveStarterCard(player, "挑衅", func() model.Card {
			return makeStarterHeroTauntCard(player)
		}) {
			e.Log(fmt.Sprintf("[Setup] %s 获得开局专属技能卡【挑衅】（专属卡区）", player.Name))
		}
	case "soul_sorcerer":
		if ensureExclusiveStarterCard(player, "灵魂链接", func() model.Card {
			return makeStarterSoulLinkCard(player)
		}) {
			e.Log(fmt.Sprintf("[Setup] %s 获得开局专属技能卡【灵魂链接】（专属卡区）", player.Name))
		}
	case "blood_priestess":
		if ensureExclusiveStarterCard(player, "同生共死", func() model.Card {
			return makeStarterBloodSharedLifeCard(player)
		}) {
			e.Log(fmt.Sprintf("[Setup] %s 获得开局专属技能卡【同生共死】（专属卡区）", player.Name))
		}
	case "bard":
		if ensureExclusiveStarterCard(player, "激昂狂想曲", func() model.Card {
			return makeStarterBardRousingRhapsodyCard(player)
		}) {
			e.Log(fmt.Sprintf("[Setup] %s 获得开局专属技能卡【激昂狂想曲】（专属卡区）", player.Name))
		}
		if ensureExclusiveStarterCard(player, "胜利交响诗", func() model.Card {
			return makeStarterBardVictorySymphonyCard(player)
		}) {
			e.Log(fmt.Sprintf("[Setup] %s 获得开局专属技能卡【胜利交响诗】（专属卡区）", player.Name))
		}
		if ensureExclusiveStarterCard(player, "希望赋格曲", func() model.Card {
			return makeStarterBardHopeFugueCard(player)
		}) {
			e.Log(fmt.Sprintf("[Setup] %s 获得开局专属技能卡【希望赋格曲】（专属卡区）", player.Name))
		}
	}
}
