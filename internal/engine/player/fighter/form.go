package fighter

import (
	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

func InHundredDragonForm(p *model.Player) bool {
	return player.HasForm(p, model.FormFighterHundredDragon)
}

func LeaveHundredDragonForm(p *model.Player) bool {
	return player.ClearForm(p, model.FormFighterHundredDragon)
}

func BlocksMagicCasting(p *model.Player) bool {
	// 格斗家百式幻龙拳：形态期间不能执行法术行动
	return player.IsCharacter(p, "fighter") && InHundredDragonForm(p)
}
