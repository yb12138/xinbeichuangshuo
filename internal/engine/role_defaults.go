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
	"plague_mage": {
		setMaxHeal: true,
		maxHeal:    5,
	},
	"crimson_sword_spirit": {
		tokens: map[string]int{
			"css_blood_cap": 3,
			"css_blood":     0,
		},
	},
	"prayer_master": {
		tokens: map[string]int{
			"prayer_rune": 0,
		},
	},
	"crimson_knight": {
		setMaxHeal: true,
		maxHeal:    4,
		tokens: map[string]int{
			"crk_blood_mark": 0,
		},
	},
	"war_homunculus": {
		tokens: map[string]int{
			"hom_war_rune":   3,
			"hom_magic_rune": 0,
		},
	},
	"priest": {
		setMaxHeal: true,
		maxHeal:    6,
	},
	"onmyoji": {
		tokens: map[string]int{
			"onmyoji_ghost_fire": 0,
		},
	},
	"blaze_witch": {
		tokens: map[string]int{
			"bw_rebirth":               0,
			"bw_flame_release_pending": 0,
		},
	},
	"magic_lancer":  {},
	"spirit_caster": {},
	"bard": {
		tokens: map[string]int{
			"bd_inspiration": 0,
		},
	},
	"hero": {
		addCrystal: 2,
		tokens: map[string]int{
			"hero_anger":                      0,
			"hero_wisdom":                     0,
			"hero_exhaustion_release_pending": 0,
			"hero_calm_end_crystal_pending":   0,
		},
	},
	"fighter": {
		tokens: map[string]int{
			"fighter_qi":                          0,
			"fighter_hundred_dragon_target_order": 0,
		},
	},
	"holy_bow": {
		addCrystal: 2,
		addMaxHeal: 1,
		tokens: map[string]int{
			"hb_cannon": 1,
			"hb_faith":  0,
		},
	},
	"sword_emperor": {
		tokens: map[string]int{
			"se_sword_qi":         0,
			"se_sword_soul_count": 0,
		},
	},
	"beast_samurai": {
		tokens: map[string]int{
			"bs_zanshin":            0,
			"bs_beast_soul":         0,
			"bs_reversal_pending_x": 0,
		},
	},
	"holy_lancer": {
		setMaxHeal: true,
		maxHeal:    2,
	},
	"arbiter": {
		addCrystal: 2,
		tokens: map[string]int{
			"arbiter_law_inited": 1,
		},
	},
	"soul_sorcerer": {
		tokens: map[string]int{
			"ss_blue_soul":   0,
			"ss_yellow_soul": 0,
		},
	},
	"moon_goddess": {
		tokens: map[string]int{
			"mg_new_moon":           0,
			"mg_petrify":            0,
			"mg_extra_turn_pending": 0,
		},
	},
	"butterfly_dancer": {
		tokens: map[string]int{
			"bt_pupa":          0,
			"bt_wither_active": 0,
		},
	},
	"blood_priestess": {},
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
