// gameflow: Prompt 文案模板与格式化（与中断展示相关）。

package promptfmt

import (
	"fmt"
	"strings"

	"starcup-engine/internal/model"
)

func ElementName(raw string) string {
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

func FormatCardInfo(card model.Card) string {
	elementLabel := ElementName(string(card.Element))
	if elementLabel == "" {
		info := fmt.Sprintf("[%s] %s", card.Element, card.Name)
		if card.Type != "" {
			info += fmt.Sprintf(" (%s", card.Type)
			if card.Damage > 0 {
				info += fmt.Sprintf(" Dmg:%d", card.Damage)
			}
			info += ")"
		}
		if card.Faction != "" {
			info += fmt.Sprintf(" [%s命格]", card.Faction)
		}

		exclusiveInfo := []string{}
		if card.ExclusiveChar1 != "" && card.ExclusiveSkill1 != "" {
			exclusiveInfo = append(exclusiveInfo, fmt.Sprintf("%s:%s", card.ExclusiveChar1, card.ExclusiveSkill1))
		}
		if card.ExclusiveChar2 != "" && card.ExclusiveSkill2 != "" {
			exclusiveInfo = append(exclusiveInfo, fmt.Sprintf("%s:%s", card.ExclusiveChar2, card.ExclusiveSkill2))
		}
		if len(exclusiveInfo) > 0 {
			info += fmt.Sprintf(" [独有技:%s]", strings.Join(exclusiveInfo, " | "))
		}

		return info
	}

	info := fmt.Sprintf("[%s系] %s", elementLabel, card.Name)
	if card.Type != "" {
		info += fmt.Sprintf(" (%s", card.Type)
		if card.Damage > 0 {
			info += fmt.Sprintf(" Dmg:%d", card.Damage)
		}
		info += ")"
	}
	if card.Faction != "" {
		info += fmt.Sprintf(" [%s命格]", card.Faction)
	}

	exclusiveInfo := []string{}
	if card.ExclusiveChar1 != "" && card.ExclusiveSkill1 != "" {
		exclusiveInfo = append(exclusiveInfo, fmt.Sprintf("%s:%s", card.ExclusiveChar1, card.ExclusiveSkill1))
	}
	if card.ExclusiveChar2 != "" && card.ExclusiveSkill2 != "" {
		exclusiveInfo = append(exclusiveInfo, fmt.Sprintf("%s:%s", card.ExclusiveChar2, card.ExclusiveSkill2))
	}
	if len(exclusiveInfo) > 0 {
		info += fmt.Sprintf(" [独有技:%s]", strings.Join(exclusiveInfo, " | "))
	}

	return info
}
