// gameflow: 集中注册所有 skill handler。

package engine

import (
	"starcup-engine/internal/engine/skill"
	"starcup-engine/internal/model"
)

type playerSkillRegistrarAdapter struct{}

func (playerSkillRegistrarAdapter) Register(id string, handler model.SkillHandler) {
	skills.Register(id, handler)
}

func registerRoleEntrySkills() {
	for _, entry := range roleRegistry.Entries() {
		for _, skill := range entry.Skills {
			if skill.ID != "" && skill.Handler != nil {
				skills.Register(skill.ID, skill.Handler)
			}
		}
	}
}

func init() {
	skills.Register("holy_shield", &skills.HolyShieldHandler{})

	// 角色入口化技能。
	registerRoleEntrySkills()

	// 以下角色已迁移到 player/<role>/module.go:
	// - 圣枪骑士（holy_lancer）
	// - 精灵射手（elf_archer）
	// - 瘟疫法师（plague_mage）
	// - 魔剑士（magic_swordsman）
	// - 血色剑灵（crimson_sword_spirit）
	// - 祈祷师（prayer_master）
	// - 红莲骑士（crimson_knight）
	// - 英灵人形（war_homunculus）
	// - 神官（priest）
	// - 鬼术师（onmyoji）
	// - 苍炎魔女（blaze_witch）
	// - 贤者（sage）
	// - 魔弓（magic_bow）
	// - 魔枪（magic_lancer）
	// - 吟游诗人（bard）
	// - 格斗家（fighter）
	// - 圣弓（holy_bow）
	// - 剑帝（sword_emperor）
	// - 兽灵武士（beast_samurai）
	// - 灵魂术士（soul_sorcerer）
	// - 月之女神（moon_goddess）
	// - 血之巫女（blood_priestess）
	// - 蝶舞者（butterfly_dancer）
}
