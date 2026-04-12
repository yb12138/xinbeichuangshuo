// gameflow: 互斥响应技能对（与旧 dispatcher 行为一致）。

package skill

// MutuallyExclusiveResponseSkill 若两技能互斥则返回 true。
func MutuallyExclusiveResponseSkill(currentSkillID, otherSkillID string) bool {
	if currentSkillID == "" || otherSkillID == "" {
		return false
	}
	if (currentSkillID == "elf_animal_companion" || currentSkillID == "elf_pet_empower") &&
		(otherSkillID == "elf_animal_companion" || otherSkillID == "elf_pet_empower") {
		return true
	}
	if (currentSkillID == "hom_rage_suppress" || currentSkillID == "hom_glyph_fusion") &&
		(otherSkillID == "hom_rage_suppress" || otherSkillID == "hom_glyph_fusion") {
		return true
	}
	return false
}
