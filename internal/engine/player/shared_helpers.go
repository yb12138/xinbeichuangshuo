// gameflow: 指示物（Token）读写基础设施。

package player

import (
	"fmt"

	"starcup-engine/internal/engine/core/runtimeutil"
	"starcup-engine/internal/model"
)

// CanPayCrystalLike 红宝石可替代蓝水晶（仅水晶消耗方向）。
func CanPayCrystalLike(ctx *model.Context, amount int) bool {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return false
	}
	return ctx.Game.CanPayCrystalCost(ctx.User.ID, amount)
}

// SpendCrystalLike 红宝石可替代蓝水晶消耗。
func SpendCrystalLike(ctx *model.Context, amount int) bool {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return false
	}
	return ctx.Game.ConsumeCrystalCost(ctx.User.ID, amount)
}

// PlayerEnergyTotal 返回玩家当前个人能量总量（红宝石 + 蓝水晶）。
func PlayerEnergyTotal(p *model.Player) int {
	if p == nil {
		return 0
	}
	return p.Gem + p.Crystal
}

// PlayerEnergyRoom 返回玩家距离个人能量上限的剩余空间。
func PlayerEnergyRoom(p *model.Player, cap int) int {
	room := cap - PlayerEnergyTotal(p)
	if room < 0 {
		return 0
	}
	return room
}

// AddPlayerGemCapped 增加个人红宝石，并裁剪到个人能量总上限。
// 返回实际增加量。
func AddPlayerGemCapped(p *model.Player, amount int, cap int) int {
	if p == nil || amount <= 0 {
		return 0
	}
	gain := amount
	if room := PlayerEnergyRoom(p, cap); gain > room {
		gain = room
	}
	if gain <= 0 {
		return 0
	}
	p.Gem += gain
	return gain
}

// AddPlayerCrystalCapped 增加个人蓝水晶，并裁剪到个人能量总上限。
// 返回实际增加量。
func AddPlayerCrystalCapped(p *model.Player, amount int, cap int) int {
	if p == nil || amount <= 0 {
		return 0
	}
	gain := amount
	if room := PlayerEnergyRoom(p, cap); gain > room {
		gain = room
	}
	if gain <= 0 {
		return 0
	}
	p.Crystal += gain
	return gain
}

// AddPlayerGemWithCap 从引擎读取动态个人能量上限后增加红宝石。
func AddPlayerGemWithCap(game model.IGameEngine, p *model.Player, amount int) int {
	cap := 3
	if game != nil {
		cap = game.GetPlayerEnergyCap(p)
	}
	return AddPlayerGemCapped(p, amount, cap)
}

// AddPlayerCrystalWithCap 从引擎读取动态个人能量上限后增加蓝水晶。
func AddPlayerCrystalWithCap(game model.IGameEngine, p *model.Player, amount int) int {
	cap := 3
	if game != nil {
		cap = game.GetPlayerEnergyCap(p)
	}
	return AddPlayerCrystalCapped(p, amount, cap)
}

// EnsurePlayerTokensMap 确保 player.Tokens map 已初始化。
func EnsurePlayerTokensMap(p *model.Player) {
	if p != nil && p.Tokens == nil {
		p.Tokens = map[string]int{}
	}
}

// EnsurePlayerSkillFlowState 确保 player.TurnState.SkillFlowState map 已初始化。
func EnsurePlayerSkillFlowState(p *model.Player) {
	if p != nil && p.TurnState.SkillFlowState == nil {
		p.TurnState.SkillFlowState = map[string]int{}
	}
}

// GetToken 安全读取玩家 Token 值。
func GetToken(p *model.Player, key string) int {
	if p == nil {
		return 0
	}
	EnsurePlayerTokensMap(p)
	return p.Tokens[key]
}

// SetToken 安全设置玩家 Token 值。
func SetToken(p *model.Player, key string, value int) {
	if p == nil {
		return
	}
	EnsurePlayerTokensMap(p)
	p.Tokens[key] = value
}

// GetSkillFlowState 安全读取玩家回合流程状态值。
func GetSkillFlowState(p *model.Player, key string) int {
	if p == nil {
		return 0
	}
	EnsurePlayerSkillFlowState(p)
	return p.TurnState.SkillFlowState[key]
}

// SetSkillFlowState 安全设置玩家回合流程状态值。
func SetSkillFlowState(p *model.Player, key string, value int) {
	if p == nil {
		return
	}
	EnsurePlayerSkillFlowState(p)
	p.TurnState.SkillFlowState[key] = value
}

// NotifyChoiceContext refreshes the pending choice interrupt after its context
// has been mutated by a choice handler.
func NotifyChoiceContext(rt ChoiceRuntime, ctxData map[string]interface{}) {
	if rt == nil {
		return
	}
	if intr := rt.GetPendingInterrupt(); intr != nil {
		intr.Context = ctxData
	}
	rt.NotifyInterruptPrompt()
}

// AdvancePromptFlowRuntimeChoice moves a prompt flow to another declared step
// and synchronizes the internal choice route key from that step's spec.
func AdvancePromptFlowRuntimeChoice(rt ChoiceRuntime, ctxData map[string]interface{}, flowRT *model.PromptFlowRuntime, flow *model.PromptFlowState, stepID string) error {
	if flowRT == nil {
		return fmt.Errorf("prompt flow runtime is nil")
	}
	if err := flowRT.MoveTo(flow, stepID); err != nil {
		return err
	}
	step, ok := flowRT.Step(stepID)
	if !ok || step.ChoiceType == "" {
		return fmt.Errorf("prompt flow %q missing choice route for step %q", flowRT.FlowID, stepID)
	}
	if ctxData != nil {
		ctxData["choice_type"] = step.ChoiceType
	}
	NotifyChoiceContext(rt, ctxData)
	return nil
}

// TokenValue 读取并规范化玩家 token 值：
// 小于 0 归零；cap >= 0 时按上限裁剪；并回写到玩家状态。
func TokenValue(p *model.Player, key string, cap int) int {
	if p == nil {
		return 0
	}
	EnsurePlayerTokensMap(p)
	v := p.Tokens[key]
	if v < 0 {
		v = 0
	}
	if cap >= 0 && v > cap {
		v = cap
	}
	p.Tokens[key] = v
	return v
}

// AddToken 在规范化基础上增减 token，并应用统一上限规则。
func AddToken(p *model.Player, key string, delta int, cap int) int {
	return AddTokenIgnoreCap(p, key, delta, cap, false)
}

// AddTokenIgnoreCap 允许按场景跳过上限裁剪（仅保留非负约束）。
func AddTokenIgnoreCap(p *model.Player, key string, delta int, cap int, ignoreCap bool) int {
	if p == nil {
		return 0
	}
	EnsurePlayerTokensMap(p)
	baseCap := cap
	if ignoreCap {
		baseCap = -1
	}
	v := TokenValue(p, key, baseCap) + delta
	if v < 0 {
		v = 0
	}
	if !ignoreCap && cap >= 0 && v > cap {
		v = cap
	}
	p.Tokens[key] = v
	return v
}

// ParseIntSliceContextValue 从 interface{} 解析 []int（支持 []int 和 []interface{} 两种输入）。
func ParseIntSliceContextValue(raw interface{}) []int {
	result := make([]int, 0)
	switch value := raw.(type) {
	case []int:
		result = append(result, value...)
	case []interface{}:
		for _, item := range value {
			switch v := item.(type) {
			case int:
				result = append(result, v)
			case float64:
				result = append(result, int(v))
			}
		}
	}
	return result
}

// ElementOrderForPrompt returns the canonical element order for prompts.
func ElementOrderForPrompt() []model.Element {
	return []model.Element{
		model.ElementEarth,
		model.ElementWater,
		model.ElementFire,
		model.ElementWind,
		model.ElementThunder,
		model.ElementLight,
		model.ElementDark,
	}
}

// GetFieldEffectCard 返回玩家场上指定效果类型的场地牌（纯函数，无 engine 依赖）。
func GetFieldEffectCard(p *model.Player, effect model.EffectType) *model.FieldCard {
	if p == nil {
		return nil
	}
	for _, fc := range p.Field {
		if fc == nil || fc.Mode != model.FieldEffect || fc.Effect != effect {
			continue
		}
		return fc
	}
	return nil
}

// ClearElfElementalShotCombatState 清理精灵射手元素射击战斗状态。
func ClearElfElementalShotCombatState(p *model.Player) {
	if p == nil {
		return
	}
	if p.TurnState.SkillFlowState == nil {
		p.TurnState.SkillFlowState = make(map[string]int)
	}
	p.TurnState.SkillFlowState["elf_elemental_shot_water_pending"] = 0
	p.TurnState.SkillFlowState["elf_elemental_shot_earth_pending"] = 0
	p.TurnState.SkillFlowState["elf_elemental_shot_wind_pending"] = 0
}

// MaxSameElementCount 返回玩家手牌中最大同系牌数量。
func MaxSameElementCount(p *model.Player) int {
	elemMap := map[model.Element]int{}
	for _, c := range p.Hand {
		if c.Element != "" {
			elemMap[c.Element]++
		}
	}
	maxCount := 0
	for _, cnt := range elemMap {
		if cnt > maxCount {
			maxCount = cnt
		}
	}
	return maxCount
}

// RemoveCardsByIndicesFromHand 从玩家手牌中移除指定索引的牌，返回移除的牌列表。
// 索引从大到小排序后删除，避免索引位移。重复索引会报错。
func RemoveCardsByIndicesFromHand(player *model.Player, indices []int) ([]model.Card, error) {
	if player == nil {
		return nil, fmt.Errorf("玩家不存在")
	}
	if len(indices) == 0 {
		return nil, nil
	}
	for _, idx := range indices {
		if idx < 0 || idx >= len(player.Hand) {
			return nil, fmt.Errorf("无效的手牌索引: %d", idx)
		}
	}
	seen := map[int]bool{}
	for _, idx := range indices {
		if seen[idx] {
			return nil, fmt.Errorf("不能重复选择同一张牌")
		}
		seen[idx] = true
	}
	// 从大到小删除，避免索引位移。
	sorted := make([]int, len(indices))
	copy(sorted, indices)
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] < sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	var removed []model.Card
	for _, idx := range sorted {
		removed = append(removed, player.Hand[idx])
		player.Hand = append(player.Hand[:idx], player.Hand[idx+1:]...)
	}
	return removed, nil
}

// AllHandIndices 返回玩家手牌的所有索引 [0, 1, ..., len(hand)-1]。
func AllHandIndices(p *model.Player) []int {
	if p == nil {
		return nil
	}
	out := make([]int, 0, len(p.Hand))
	for i := range p.Hand {
		out = append(out, i)
	}
	return out
}

// GetCardIndicesByElement 返回玩家手牌中指定元素牌的索引列表。
func GetCardIndicesByElement(p *model.Player, element model.Element) []int {
	if p == nil {
		return nil
	}
	var out []int
	for i, c := range p.Hand {
		if c.Element == element {
			out = append(out, i)
		}
	}
	return out
}

// MustChoiceResumePointFromMap 从选择上下文中提取恢复点，缺失则 panic。
func MustChoiceResumePointFromMap(data map[string]interface{}, key string) interface{} {
	if data == nil {
		panic(fmt.Sprintf("missing resume point map for key %q", key))
	}
	raw, ok := data[key]
	if !ok {
		panic(fmt.Sprintf("missing resume point key %q", key))
	}
	if raw == nil {
		panic(fmt.Sprintf("nil resume point for key %q", key))
	}
	return raw
}

// BuildTargetChoicePrompt 构造通用目标选择 Prompt（目标列表由 data["target_ids"] 提供）。
//
// choiceType 必须为非空字符串：会写入返回的 Prompt.ChoiceType，前端依赖该字段把
// 「连续数字 option id」豁免出手牌索引匹配（见 PromptDialog.vue 中
// NON_HAND_INDEXED_PROMPT_CHOICE_TYPES）。若未设置 choice_type，
// id 为 "0"/"1" 等的目标选项会被前端误判为手牌索引。
func BuildTargetChoicePrompt(rt ChoiceRuntime, choiceType, playerID string, message string, data map[string]interface{}, allowCancel bool) *model.Prompt {
	targetIDs := runtimeutil.ParseStringSliceContextValue(data["target_ids"])
	options := make([]model.PromptOption, 0, len(targetIDs)+1)
	for _, targetID := range targetIDs {
		if target := rt.GetPlayers()[targetID]; target != nil {
			options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", len(options)), Label: target.Name, TargetID: targetID})
		}
	}
	if allowCancel {
		options = append(options, model.PromptOption{ID: "cancel", Label: "取消"})
	}
	return &model.Prompt{
		Type:       model.PromptConfirm,
		ChoiceType: choiceType,
		PlayerID:   playerID,
		Message:    message,
		Options:    options,
		Min:        1,
		Max:        1,
		Presentation: &model.PromptPresentation{
			Kind:         model.PresentationTargetPicker,
			TargetFilter: "custom",
		},
	}
}
