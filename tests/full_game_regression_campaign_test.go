package tests

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"testing"
)

type fullGameCampaignResult struct {
	lineups         [][]string
	triggeredSkills map[string]int
	roleTriggered   map[string]bool
	err             error
}

type directedScenarioCampaignResult struct {
	lineups               [][]string
	scenarios             []directedScenarioPlan
	triggeredSkills       map[string]int
	targetSkillSet        map[string]struct{}
	scenarioHitCounts     map[string]int
	scenarioMissingSkills map[string][]string
	err                   error
}

var (
	fullGameCampaignOnce     sync.Once
	fullGameCampaignData     fullGameCampaignResult
	directedScenarioOnce     sync.Once
	directedScenarioCampaign directedScenarioCampaignResult
)

const fullGameCampaignRounds = 3

func buildDirectedScenarios() []directedScenarioPlan {
	return []directedScenarioPlan{
		{
			Name:              "S1_startup_chain",
			Lineup:            []string{"arbiter", "saintess", "assassin", "angel", "sealer", "adventurer"},
			Runs:              10,
			TargetSkillTitles: []string{"仲裁仪式", "仪式中断", "怜悯", "潜行", "天使之歌"},
			RoleActionOrder: map[string][]string{
				"arbiter":  {autoPlanActionExtract, autoPlanActionSkill, autoPlanActionAttack, autoPlanActionMagic, autoPlanActionBuy},
				"saintess": {autoPlanActionExtract, autoPlanActionSkill, autoPlanActionMagic, autoPlanActionAttack, autoPlanActionBuy},
				"assassin": {autoPlanActionExtract, autoPlanActionAttack, autoPlanActionMagic, autoPlanActionBuy},
				"angel":    {autoPlanActionAttack, autoPlanActionBuy, autoPlanActionExtract, autoPlanActionSkill, autoPlanActionMagic},
				"sealer":   {autoPlanActionAttack, autoPlanActionBuy, autoPlanActionExtract, autoPlanActionSkill, autoPlanActionMagic},
				"adventurer": {
					autoPlanActionAttack, autoPlanActionBuy, autoPlanActionExtract, autoPlanActionSkill, autoPlanActionMagic, autoPlanActionSynthesize,
				},
			},
			RoleInterruptSkillPriority: map[string][]string{
				"arbiter":  {"arbiter_ritual", "arbiter_ritual_break"},
				"saintess": {"mercy"},
				"assassin": {"stealth"},
				"angel":    {"angel_song"},
			},
		},
		{
			Name:              "S2_seal_chain",
			Lineup:            []string{"sealer", "angel", "holy_lancer", "archer", "magical_girl", "adventurer"},
			Runs:              20,
			TargetSkillTitles: []string{"五系束缚", "雷之封印", "封印破碎"},
			RoleActionOrder: map[string][]string{
				"sealer":      {autoPlanActionAttack, autoPlanActionBuy, autoPlanActionExtract, autoPlanActionSkill, autoPlanActionMagic},
				"angel":       {autoPlanActionAttack, autoPlanActionBuy, autoPlanActionExtract, autoPlanActionSkill, autoPlanActionMagic},
				"holy_lancer": {autoPlanActionAttack, autoPlanActionBuy, autoPlanActionExtract, autoPlanActionSkill, autoPlanActionMagic},
			},
			RoleActionSkillPriority: map[string][]string{
				"sealer": {"seal_break", "thunder_seal", "five_elements_bind", "water_seal", "fire_seal", "earth_seal", "wind_seal"},
			},
		},
		{
			Name:              "S3_adventurer_economy",
			Lineup:            []string{"adventurer", "sealer", "valkyrie", "berserker", "archer", "angel"},
			Runs:              14,
			TargetSkillTitles: []string{"强运", "地下法则", "冒险者天堂"},
			RoleActionOrder: map[string][]string{
				"adventurer": {
					autoPlanActionAttack, autoPlanActionBuy, autoPlanActionExtract, autoPlanActionSkill, autoPlanActionMagic, autoPlanActionSynthesize,
				},
				"sealer":   {autoPlanActionAttack, autoPlanActionBuy, autoPlanActionExtract, autoPlanActionSkill, autoPlanActionMagic},
				"valkyrie": {autoPlanActionAttack, autoPlanActionBuy, autoPlanActionExtract, autoPlanActionSkill, autoPlanActionMagic},
			},
			RoleActionSkillPriority: map[string][]string{
				"adventurer": {"adventurer_fraud", "adventurer_steal_sky"},
			},
			RoleInterruptSkillPriority: map[string][]string{
				"adventurer": {"adventurer_paradise"},
			},
		},
		{
			Name:              "S4_berserker_damage",
			Lineup:            []string{"berserker", "saintess", "angel", "holy_lancer", "adventurer", "valkyrie"},
			Runs:              12,
			TargetSkillTitles: []string{"狂化", "撕裂", "血影狂刀"},
			RoleActionOrder: map[string][]string{
				"berserker": {autoPlanActionExtract, autoPlanActionAttack, autoPlanActionBuy, autoPlanActionMagic},
				"saintess":  {autoPlanActionBuy, autoPlanActionAttack, autoPlanActionMagic, autoPlanActionExtract, autoPlanActionSkill},
				"angel":     {autoPlanActionBuy, autoPlanActionAttack, autoPlanActionExtract, autoPlanActionSkill, autoPlanActionMagic},
			},
			RoleInterruptSkillPriority: map[string][]string{
				"berserker": {"berserker_tear"},
			},
		},
		{
			Name:              "S5_archer_miss",
			Lineup:            []string{"archer", "sealer", "blade_master", "angel", "holy_lancer", "assassin"},
			Runs:              16,
			TargetSkillTitles: []string{"狙击", "贯穿射击"},
			RoleActionOrder: map[string][]string{
				"archer":       {autoPlanActionAttack, autoPlanActionExtract, autoPlanActionSkill, autoPlanActionBuy, autoPlanActionMagic},
				"sealer":       {autoPlanActionAttack, autoPlanActionBuy, autoPlanActionExtract, autoPlanActionSkill, autoPlanActionMagic},
				"blade_master": {autoPlanActionAttack, autoPlanActionBuy, autoPlanActionExtract, autoPlanActionSkill, autoPlanActionMagic},
			},
			RoleActionSkillPriority: map[string][]string{
				"archer": {"snipe", "flash_trap"},
			},
			RoleInterruptSkillPriority: map[string][]string{
				"archer": {"piercing_shot"},
			},
		},
		{
			Name:              "S6_blademaster_break_shield",
			Lineup:            []string{"blade_master", "saintess", "sealer", "angel", "magical_girl", "adventurer"},
			Runs:              16,
			TargetSkillTitles: []string{"烈风技", "剑影"},
			RoleActionOrder: map[string][]string{
				"blade_master": {autoPlanActionAttack, autoPlanActionBuy, autoPlanActionExtract, autoPlanActionSkill, autoPlanActionMagic},
				"saintess":     {autoPlanActionAttack, autoPlanActionBuy, autoPlanActionExtract, autoPlanActionSkill, autoPlanActionMagic},
				"sealer":       {autoPlanActionAttack, autoPlanActionBuy, autoPlanActionExtract, autoPlanActionSkill, autoPlanActionMagic},
			},
			RoleInterruptSkillPriority: map[string][]string{
				"blade_master": {"sword_shadow", "wind_fury"},
			},
		},
		{
			Name:              "S7_holy_lancer_saintess",
			Lineup:            []string{"holy_lancer", "saintess", "valkyrie", "berserker", "assassin", "archer"},
			Runs:              12,
			TargetSkillTitles: []string{"圣光祈愈", "圣疗"},
			RoleActionOrder: map[string][]string{
				"holy_lancer": {autoPlanActionSkill, autoPlanActionExtract, autoPlanActionAttack, autoPlanActionMagic, autoPlanActionBuy},
				"saintess":    {autoPlanActionExtract, autoPlanActionSkill, autoPlanActionBuy, autoPlanActionAttack, autoPlanActionMagic},
			},
			RoleActionSkillPriority: map[string][]string{
				"holy_lancer": {"holy_lancer_prayer", "holy_lancer_punishment", "holy_lancer_radiance"},
				"saintess":    {"saint_heal", "healing_light", "heal"},
			},
		},
		{
			Name:              "S8_magic_bullet_chain",
			Lineup:            []string{"magical_girl", "elementalist", "sealer", "adventurer", "angel", "arbiter"},
			Runs:              6,
			TargetSkillTitles: []string{"魔弹融合", "魔弹掌控", "毁灭风暴"},
			RoleActionOrder: map[string][]string{
				"magical_girl": {autoPlanActionExtract, autoPlanActionMagic, autoPlanActionSkill, autoPlanActionAttack, autoPlanActionBuy},
			},
			RoleActionSkillPriority: map[string][]string{
				"magical_girl": {"destruction_storm", "magic_blast"},
			},
		},
		{
			Name:              "S9_angel_guard",
			Lineup:            []string{"angel", "saintess", "arbiter", "magical_girl", "elementalist", "sealer"},
			Runs:              12,
			TargetSkillTitles: []string{"神之庇护"},
			RoleActionOrder: map[string][]string{
				"angel":        {autoPlanActionExtract, autoPlanActionBuy, autoPlanActionAttack, autoPlanActionMagic, autoPlanActionSkill},
				"magical_girl": {autoPlanActionMagic, autoPlanActionSkill, autoPlanActionExtract, autoPlanActionAttack, autoPlanActionBuy},
				"elementalist": {autoPlanActionSkill, autoPlanActionMagic, autoPlanActionAttack, autoPlanActionExtract, autoPlanActionBuy},
			},
			RoleInterruptSkillPriority: map[string][]string{
				"angel": {"god_protection", "angel_song"},
			},
		},
		{
			Name:              "S10_angel_song_focus",
			Lineup:            []string{"angel", "adventurer", "sealer", "arbiter", "assassin", "holy_lancer"},
			Runs:              20,
			TargetSkillTitles: []string{"天使之歌"},
			RoleActionOrder: map[string][]string{
				"angel":      {autoPlanActionAttack, autoPlanActionBuy, autoPlanActionExtract, autoPlanActionSkill, autoPlanActionMagic},
				"adventurer": {autoPlanActionAttack, autoPlanActionBuy, autoPlanActionSkill, autoPlanActionMagic, autoPlanActionExtract},
				"sealer":     {autoPlanActionAttack, autoPlanActionBuy, autoPlanActionSkill, autoPlanActionMagic, autoPlanActionExtract},
			},
			RoleInterruptSkillPriority: map[string][]string{
				"angel": {"angel_song"},
			},
		},
		{
			Name:              "S11_bind_focus",
			Lineup:            []string{"sealer", "angel", "adventurer", "arbiter", "assassin", "holy_lancer"},
			Runs:              24,
			TargetSkillTitles: []string{"五系束缚"},
			RoleActionOrder: map[string][]string{
				"sealer":     {autoPlanActionAttack, autoPlanActionBuy, autoPlanActionExtract, autoPlanActionSkill, autoPlanActionMagic},
				"angel":      {autoPlanActionAttack, autoPlanActionBuy, autoPlanActionSkill, autoPlanActionMagic, autoPlanActionExtract},
				"adventurer": {autoPlanActionAttack, autoPlanActionBuy, autoPlanActionSkill, autoPlanActionMagic, autoPlanActionExtract},
			},
			RoleActionSkillPriority: map[string][]string{
				"sealer": {"five_elements_bind", "thunder_seal", "seal_break", "water_seal", "fire_seal", "earth_seal", "wind_seal"},
			},
		},
		{
			Name:              "S12_archer_snipe_focus",
			Lineup:            []string{"archer", "sealer", "adventurer", "angel", "holy_lancer", "assassin"},
			Runs:              24,
			TargetSkillTitles: []string{"狙击"},
			RoleActionOrder: map[string][]string{
				"archer":     {autoPlanActionAttack, autoPlanActionExtract, autoPlanActionSkill, autoPlanActionBuy, autoPlanActionMagic},
				"sealer":     {autoPlanActionAttack, autoPlanActionBuy, autoPlanActionSkill, autoPlanActionMagic, autoPlanActionExtract},
				"adventurer": {autoPlanActionAttack, autoPlanActionBuy, autoPlanActionSkill, autoPlanActionMagic, autoPlanActionExtract},
			},
			RoleActionSkillPriority: map[string][]string{
				"archer": {"snipe", "flash_trap"},
			},
		},
		{
			Name:              "S13_sword_shadow_focus",
			Lineup:            []string{"blade_master", "sealer", "adventurer", "angel", "magical_girl", "holy_lancer"},
			Runs:              36,
			TargetSkillTitles: []string{"剑影"},
			RoleActionOrder: map[string][]string{
				"blade_master": {autoPlanActionAttack, autoPlanActionBuy, autoPlanActionExtract, autoPlanActionSkill, autoPlanActionMagic},
				"sealer":       {autoPlanActionBuy, autoPlanActionAttack, autoPlanActionSkill, autoPlanActionMagic, autoPlanActionExtract},
				"adventurer":   {autoPlanActionBuy, autoPlanActionAttack, autoPlanActionSkill, autoPlanActionMagic, autoPlanActionExtract},
			},
			RoleInterruptSkillPriority: map[string][]string{
				"blade_master": {"sword_shadow"},
			},
		},
		{
			Name:              "S14_archer_piercing_focus",
			Lineup:            []string{"archer", "sealer", "adventurer", "magical_girl", "elementalist", "saintess"},
			Runs:              36,
			TargetSkillTitles: []string{"贯穿射击"},
			RoleActionOrder: map[string][]string{
				"archer":     {autoPlanActionAttack, autoPlanActionBuy, autoPlanActionExtract, autoPlanActionSkill, autoPlanActionMagic},
				"sealer":     {autoPlanActionAttack, autoPlanActionBuy, autoPlanActionSkill, autoPlanActionMagic, autoPlanActionExtract},
				"adventurer": {autoPlanActionAttack, autoPlanActionBuy, autoPlanActionSkill, autoPlanActionMagic, autoPlanActionExtract},
			},
			RoleInterruptSkillPriority: map[string][]string{
				"archer": {"piercing_shot"},
			},
		},
		{
			Name:              "S15_seal_break_focus",
			Lineup:            []string{"sealer", "angel", "archer", "saintess", "adventurer", "holy_lancer"},
			Runs:              28,
			TargetSkillTitles: []string{"封印破碎"},
			RoleActionOrder: map[string][]string{
				"sealer":     {autoPlanActionAttack, autoPlanActionBuy, autoPlanActionExtract, autoPlanActionSkill, autoPlanActionMagic},
				"angel":      {autoPlanActionSkill, autoPlanActionAttack, autoPlanActionBuy, autoPlanActionExtract, autoPlanActionMagic},
				"archer":     {autoPlanActionAttack, autoPlanActionMagic, autoPlanActionBuy, autoPlanActionExtract, autoPlanActionSkill},
				"adventurer": {autoPlanActionAttack, autoPlanActionBuy, autoPlanActionExtract, autoPlanActionSkill, autoPlanActionMagic},
			},
			RoleActionSkillPriority: map[string][]string{
				"sealer": {"seal_break", "thunder_seal", "five_elements_bind", "water_seal", "fire_seal", "earth_seal", "wind_seal"},
				"angel":  {"angel_wall"},
			},
		},
		{
			Name:              "S16_paradise_focus",
			Lineup:            []string{"adventurer", "sealer", "angel", "arbiter", "assassin", "holy_lancer"},
			Runs:              28,
			TargetSkillTitles: []string{"冒险者天堂"},
			RoleActionOrder: map[string][]string{
				"adventurer": {
					autoPlanActionExtract, autoPlanActionAttack, autoPlanActionBuy, autoPlanActionSkill, autoPlanActionMagic, autoPlanActionSynthesize,
				},
				"sealer": {autoPlanActionBuy, autoPlanActionAttack, autoPlanActionExtract, autoPlanActionSkill, autoPlanActionMagic},
				"angel":  {autoPlanActionBuy, autoPlanActionAttack, autoPlanActionExtract, autoPlanActionSkill, autoPlanActionMagic},
			},
			RoleInterruptSkillPriority: map[string][]string{
				"adventurer": {"adventurer_paradise"},
			},
		},
		{
			Name:              "S17_berserker_tear_blade_focus",
			Lineup:            []string{"berserker", "adventurer", "saintess", "angel", "holy_lancer", "valkyrie"},
			Runs:              32,
			TargetSkillTitles: []string{"撕裂", "血影狂刀"},
			RoleActionOrder: map[string][]string{
				"berserker":  {autoPlanActionExtract, autoPlanActionAttack, autoPlanActionBuy, autoPlanActionMagic},
				"adventurer": {autoPlanActionBuy, autoPlanActionAttack, autoPlanActionExtract, autoPlanActionSkill, autoPlanActionMagic},
				"saintess":   {autoPlanActionBuy, autoPlanActionAttack, autoPlanActionMagic, autoPlanActionExtract, autoPlanActionSkill},
			},
			RoleInterruptSkillPriority: map[string][]string{
				"berserker": {"berserker_tear"},
			},
		},
		{
			Name:              "S18_saint_heal_focus",
			Lineup:            []string{"saintess", "adventurer", "angel", "holy_lancer", "sealer", "archer"},
			Runs:              28,
			TargetSkillTitles: []string{"圣疗"},
			RoleActionOrder: map[string][]string{
				"saintess":   {autoPlanActionExtract, autoPlanActionSkill, autoPlanActionBuy, autoPlanActionAttack, autoPlanActionMagic},
				"adventurer": {autoPlanActionBuy, autoPlanActionAttack, autoPlanActionExtract, autoPlanActionSkill, autoPlanActionMagic},
				"angel":      {autoPlanActionBuy, autoPlanActionAttack, autoPlanActionExtract, autoPlanActionSkill, autoPlanActionMagic},
			},
			RoleActionSkillPriority: map[string][]string{
				"saintess": {"saint_heal"},
			},
		},
		{
			Name:              "S19_elf_archer_elemental_shot",
			Lineup:            []string{"elf_archer", "adventurer", "angel", "sealer", "saintess", "holy_lancer"},
			Runs:              20,
			TargetSkillTitles: []string{"元素射击"},
			RoleActionOrder: map[string][]string{
				"elf_archer": {autoPlanActionAttack, autoPlanActionBuy, autoPlanActionExtract, autoPlanActionSkill, autoPlanActionMagic},
				"adventurer": {autoPlanActionBuy, autoPlanActionAttack, autoPlanActionExtract, autoPlanActionSkill, autoPlanActionMagic},
				"angel":      {autoPlanActionBuy, autoPlanActionAttack, autoPlanActionExtract, autoPlanActionSkill, autoPlanActionMagic},
			},
			RoleInterruptSkillPriority: map[string][]string{
				"elf_archer": {"elf_elemental_shot", "elf_animal_companion", "elf_pet_empower"},
			},
		},
		{
			Name:              "S20_bard_hope_fugue",
			Lineup:            []string{"bard", "adventurer", "angel", "sealer", "holy_lancer", "saintess"},
			Runs:              20,
			TargetSkillTitles: []string{"希望赋格曲"},
			RoleActionOrder: map[string][]string{
				"bard":       {autoPlanActionExtract, autoPlanActionBuy, autoPlanActionAttack, autoPlanActionMagic, autoPlanActionSkill},
				"adventurer": {autoPlanActionBuy, autoPlanActionAttack, autoPlanActionExtract, autoPlanActionSkill, autoPlanActionMagic},
				"angel":      {autoPlanActionBuy, autoPlanActionAttack, autoPlanActionExtract, autoPlanActionSkill, autoPlanActionMagic},
			},
			RoleInterruptSkillPriority: map[string][]string{
				"bard": {"bd_hope_fugue", "bd_rousing_rhapsody", "bd_victory_symphony"},
			},
		},
	}
}

func mergedRegressionTriggeredSkills(t *testing.T) (fullGameCampaignResult, map[string]int) {
	t.Helper()
	campaign := runFullGameCampaign(t)
	directed := runDirectedScenarioCampaign(t)
	merged := make(map[string]int, len(campaign.triggeredSkills)+len(directed.triggeredSkills))
	mergeTriggered(merged, campaign.triggeredSkills)
	mergeTriggered(merged, directed.triggeredSkills)
	return campaign, merged
}

func buildRegressionLineups(roleIDs []string) [][]string {
	seen := make(map[string]struct{})
	lineups := make([][]string, 0)

	appendUnique := func(items [][]string) {
		for _, lineup := range items {
			key := strings.Join(lineup, "|")
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			lineups = append(lineups, lineup)
		}
	}

	base := buildCoverageLineups(roleIDs, autoGamePlayers)
	appendUnique(base)

	mirrors := make([][]string, 0, len(base))
	for _, lineup := range base {
		mirrors = append(mirrors, mirrorLineup(lineup))
	}
	appendUnique(mirrors)

	for _, shift := range []int{1, 2, 3} {
		rotated := rotateRoleIDs(roleIDs, shift)
		appendUnique(buildCoverageLineups(rotated, autoGamePlayers))
	}

	return lineups
}

func runFullGameCampaign(t *testing.T) fullGameCampaignResult {
	t.Helper()
	fullGameCampaignOnce.Do(func() {
		roleIDs := collectRoleIDs()
		if len(roleIDs) < autoGamePlayers {
			fullGameCampaignData.err = fmt.Errorf("need at least %d roles for 3v3 test, got %d", autoGamePlayers, len(roleIDs))
			return
		}

		lineups := make([][]string, 0)
		for round := 0; round < fullGameCampaignRounds; round++ {
			rotated := rotateRoleIDs(roleIDs, round*2)
			lineups = append(lineups, buildRegressionLineups(rotated)...)
		}
		roleSkills := collectRoleSkillTitles()

		fullGameCampaignData.lineups = lineups
		fullGameCampaignData.triggeredSkills = make(map[string]int)
		fullGameCampaignData.roleTriggered = make(map[string]bool)

		for _, roleID := range roleIDs {
			fullGameCampaignData.roleTriggered[roleID] = false
		}

		for _, lineup := range lineups {
			result, err := runAutoGame(lineup, autoGameStepLimit)
			if err != nil {
				fullGameCampaignData.err = fmt.Errorf("lineup=%v err=%w", lineup, err)
				return
			}
			mergeTriggered(fullGameCampaignData.triggeredSkills, result.triggeredSkills)

			for roleID, skills := range roleSkills {
				if fullGameCampaignData.roleTriggered[roleID] {
					continue
				}
				for skill := range skills {
					if fullGameCampaignData.triggeredSkills[skill] > 0 {
						fullGameCampaignData.roleTriggered[roleID] = true
						break
					}
				}
			}
		}
	})

	if fullGameCampaignData.err != nil {
		t.Fatal(fullGameCampaignData.err)
	}
	return fullGameCampaignData
}

func runDirectedScenarioCampaign(t *testing.T) directedScenarioCampaignResult {
	t.Helper()
	directedScenarioOnce.Do(func() {
		scenarios := buildDirectedScenarios()
		directedScenarioCampaign.scenarios = scenarios
		directedScenarioCampaign.triggeredSkills = make(map[string]int)
		directedScenarioCampaign.targetSkillSet = make(map[string]struct{})
		directedScenarioCampaign.scenarioHitCounts = make(map[string]int)
		directedScenarioCampaign.scenarioMissingSkills = make(map[string][]string)

		for i := range scenarios {
			scenario := scenarios[i]
			runs := scenario.Runs
			if runs <= 0 {
				runs = 1
			}

			localTriggered := make(map[string]int)
			for _, skill := range scenario.TargetSkillTitles {
				directedScenarioCampaign.targetSkillSet[skill] = struct{}{}
			}

			for runIdx := 0; runIdx < runs; runIdx++ {
				lineup := append([]string{}, scenario.Lineup...)
				if runIdx%2 == 1 {
					lineup = mirrorLineup(lineup)
				}
				directedScenarioCampaign.lineups = append(directedScenarioCampaign.lineups, lineup)

				seedTag := fmt.Sprintf("%s_run_%02d", scenario.Name, runIdx)
				result, err := runAutoGameWithScenarioSeedTag(lineup, autoGameStepLimit, &scenario, seedTag)
				if err != nil {
					directedScenarioCampaign.err = fmt.Errorf("directed scenario=%s lineup=%v err=%w", scenario.Name, lineup, err)
					return
				}
				mergeTriggered(localTriggered, result.triggeredSkills)
				mergeTriggered(directedScenarioCampaign.triggeredSkills, result.triggeredSkills)
			}

			hitCount := 0
			missing := make([]string, 0)
			for _, title := range scenario.TargetSkillTitles {
				if localTriggered[title] > 0 {
					hitCount++
				} else {
					missing = append(missing, title)
				}
			}
			sort.Strings(missing)
			directedScenarioCampaign.scenarioHitCounts[scenario.Name] = hitCount
			directedScenarioCampaign.scenarioMissingSkills[scenario.Name] = missing
		}
	})

	if directedScenarioCampaign.err != nil {
		t.Fatal(directedScenarioCampaign.err)
	}
	return directedScenarioCampaign
}

func TestFullGame3v3_Regression_ActionSkillCoverage(t *testing.T) {
	campaign, merged := mergedRegressionTriggeredSkills(t)
	expected := collectActionSkillTitles()
	triggeredCount, total, ratio := coverageStats(expected, merged)

	t.Logf("campaign lineups=%d", len(campaign.lineups))
	t.Logf("action-skill coverage: %d/%d (%.2f)", triggeredCount, total, ratio)
	if top := topTriggeredActionSkills(expected, merged, 15); len(top) > 0 {
		t.Logf("top action skills: %s", strings.Join(top, ", "))
	}

	if ratio < 0.45 {
		missing := missingSkillList(expected, merged)
		t.Fatalf("action-skill coverage too low: %d/%d (%.2f), missing=%v", triggeredCount, total, ratio, missing)
	}
}

func TestFullGame3v3_Regression_AllSkillCoverage(t *testing.T) {
	campaign, merged := mergedRegressionTriggeredSkills(t)
	expected := collectAllSkillTitles()
	triggeredCount, total, ratio := coverageStats(expected, merged)

	t.Logf("campaign lineups=%d", len(campaign.lineups))
	t.Logf("all-skill coverage: %d/%d (%.2f)", triggeredCount, total, ratio)

	// 角色池扩展后（含更多低频启动/响应技能），全技能覆盖率期望适度下调。
	// 该阈值仍能识别明显回归，但避免因技能总数上升导致稳定误报。
	if ratio < 0.54 {
		missing := missingSkillList(expected, merged)
		t.Fatalf("all-skill coverage too low: %d/%d (%.2f), missing=%v", triggeredCount, total, ratio, missing)
	}
}

func TestFullGame3v3_Regression_EachRoleTriggersSkill(t *testing.T) {
	_, merged := mergedRegressionTriggeredSkills(t)
	roleIDs := collectRoleIDs()
	roleSkills := collectRoleSkillTitles()
	missingRoles := make([]string, 0)
	for _, roleID := range roleIDs {
		hit := false
		for skill := range roleSkills[roleID] {
			if merged[skill] > 0 {
				hit = true
				break
			}
		}
		if !hit {
			missingRoles = append(missingRoles, roleID)
		}
	}
	sort.Strings(missingRoles)

	if len(missingRoles) > 0 {
		t.Fatalf("some roles have no triggered skills in campaign: %v", missingRoles)
	}
}

func TestFullGame3v3_DirectedScenario_EachScenarioHitsTarget(t *testing.T) {
	campaign := runDirectedScenarioCampaign(t)
	t.Logf("directed scenario runs=%d", len(campaign.lineups))

	zeroHit := make([]string, 0)
	hitScenarios := 0
	for _, scenario := range campaign.scenarios {
		hitCount := campaign.scenarioHitCounts[scenario.Name]
		t.Logf(
			"scenario=%s hits=%d/%d missing=%v",
			scenario.Name,
			hitCount,
			len(scenario.TargetSkillTitles),
			campaign.scenarioMissingSkills[scenario.Name],
		)
		if hitCount > 0 {
			hitScenarios++
		} else {
			zeroHit = append(zeroHit, scenario.Name)
		}
	}
	sort.Strings(zeroHit)
	if hitScenarios < 2 {
		t.Fatalf(
			"directed scenarios with target hits are too few: hit=%d total=%d zero_hit=%v",
			hitScenarios,
			len(campaign.scenarios),
			zeroHit,
		)
	}
}

func TestFullGame3v3_DirectedScenario_TargetSkillCoverage(t *testing.T) {
	campaign := runDirectedScenarioCampaign(t)
	triggeredCount, total, ratio := coverageStats(campaign.targetSkillSet, campaign.triggeredSkills)

	t.Logf("directed scenario runs=%d", len(campaign.lineups))
	t.Logf("directed target-skill coverage: %d/%d (%.2f)", triggeredCount, total, ratio)
	if top := topTriggeredActionSkills(campaign.targetSkillSet, campaign.triggeredSkills, 20); len(top) > 0 {
		t.Logf("directed top triggered target skills: %s", strings.Join(top, ", "))
	}

	if ratio < 1.0 {
		missing := missingSkillList(campaign.targetSkillSet, campaign.triggeredSkills)
		t.Fatalf("directed target-skill coverage too low: %d/%d (%.2f), missing=%v", triggeredCount, total, ratio, missing)
	}
}
