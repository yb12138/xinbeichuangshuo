// gameflow: 角色默认指示物、形态等。

package engine

import "starcup-engine/internal/model"

type roleDefaultConfig struct {
	setMaxHeal bool
	maxHeal    int
	addMaxHeal int
	addCrystal int
	tokens     map[string]int
}

func (cfg roleDefaultConfig) apply(player *model.Player) {
	if player == nil {
		return
	}
	if cfg.setMaxHeal {
		player.MaxHeal = cfg.maxHeal
	}
	if cfg.addMaxHeal != 0 {
		player.MaxHeal += cfg.addMaxHeal
	}
	if cfg.addCrystal != 0 {
		player.Crystal += cfg.addCrystal
	}
	if len(cfg.tokens) == 0 {
		return
	}
	if player.Tokens == nil {
		player.Tokens = map[string]int{}
	}
	for key, value := range cfg.tokens {
		player.Tokens[key] = value
	}
}

var roleDefaultConfigs = map[string]roleDefaultConfig{
	"moon_goddess": {
		tokens: map[string]int{
			"mg_new_moon":           0,
			"mg_petrify":            0,
			"mg_extra_turn_pending": 0,
		},
	},
}

// applyRoleDefaults 初始化角色的基础指示物/上限等（与 AddPlayer 保持一致）
func (e *GameEngine) applyRoleDefaults(player *model.Player) {
	if player == nil || player.Character == nil {
		return
	}
	player.Orientation = model.OrientationNormal
	player.Form = ""
	if player.Tokens == nil {
		player.Tokens = map[string]int{}
	}
	if entry := roleRegistry.Entry(player.Character.ID); entry.ID != "" && entry.Defaults != nil {
		entry.ApplyDefaults(player)
		return
	}
	if cfg, ok := roleDefaultConfigs[player.Character.ID]; ok {
		cfg.apply(player)
	}
}
