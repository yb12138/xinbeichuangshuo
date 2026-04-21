// gameflow: 冒险家技能处理器。

package adventurer

import (
	"fmt"
	skills "starcup-engine/internal/engine/skill"
	"starcup-engine/internal/model"
)

func getSkillFlow(p *model.Player, key string) int {
	if p == nil || p.TurnState.SkillFlowState == nil {
		return 0
	}
	return p.TurnState.SkillFlowState[key]
}

func setSkillFlow(p *model.Player, key string, v int) {
	if p == nil {
		return
	}
	if p.TurnState.SkillFlowState == nil {
		p.TurnState.SkillFlowState = make(map[string]int)
	}
	p.TurnState.SkillFlowState[key] = v
}

func playerEnergyCap(p *model.Player) int {
	if p == nil {
		return 3
	}
	cap := 3
	if p.Character != nil && p.Character.ID == "sage" {
		cap++
	}
	return cap
}

// 红宝石可替代蓝水晶（仅水晶消耗方向）
func canPayCrystalLike(ctx *model.Context, amount int) bool {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return false
	}
	return ctx.Game.CanPayCrystalCost(ctx.User.ID, amount)
}

func spendCrystalLike(ctx *model.Context, amount int) bool {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return false
	}
	return ctx.Game.ConsumeCrystalCost(ctx.User.ID, amount)
}

// --- 冒险家技能处理器 ---

type AdventurerFraudHandler struct{ skills.BaseHandler }

func (h *AdventurerFraudHandler) CanUse(ctx *model.Context) bool {
	counts := map[model.Element]int{}
	for _, c := range ctx.User.Hand {
		counts[c.Element]++
	}
	for ele, n := range counts {
		// 弃2同系仅要求有同系牌；攻击系别在后续弹窗中单独选择（不含光/暗）
		if ele != "" && n >= 2 {
			return true
		}
		if n >= 3 {
			return true
		}
	}
	return false
}

func (h *AdventurerFraudHandler) Execute(ctx *model.Context) error {
	counts := map[model.Element]int{}
	for _, c := range ctx.User.Hand {
		counts[c.Element]++
	}
	canPick := false
	for ele, n := range counts {
		if ele == "" {
			continue
		}
		if n >= 2 {
			canPick = true
			break
		}
	}
	if !canPick {
		return nil
	}
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type": "adventurer_fraud_pick",
			"user_id":     ctx.User.ID,
			"user_ctx":    ctx,
			"fraud_target_id": func() string {
				if ctx.Target != nil {
					return ctx.Target.ID
				}
				return ""
			}(),
			"fraud_from_skill": true,
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [欺诈]，请先选择同系手牌", ctx.User.Name))
	return nil
}

type AdventurerLuckyFortuneHandler struct{ skills.BaseHandler }

func (h *AdventurerLuckyFortuneHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.Timing != model.TimingOnAttackDeclared {
		return false
	}
	if ctx.EventCtx == nil || ctx.EventCtx.Card == nil {
		return false
	}
	card := ctx.EventCtx.Card
	// 强运仅在"欺诈转化出的攻击"开始时自动触发。
	return card.ID == "fraud_virtual_attack" || card.Name == "欺诈"
}

func (h *AdventurerLuckyFortuneHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return nil
	}
	ctx.User.Crystal++
	ctx.Game.Log(fmt.Sprintf("%s 的 [强运] 触发，获得1蓝水晶", ctx.User.Name))
	return nil
}

type AdventurerUndergroundLawHandler struct{ skills.BaseHandler }

func (h *AdventurerUndergroundLawHandler) CanUse(ctx *model.Context) bool {
	return ctx.EventCtx != nil && ctx.EventCtx.ActionType == model.ActionBuy
}

func (h *AdventurerUndergroundLawHandler) Execute(ctx *model.Context) error {
	ctx.Game.ModifyGem(string(ctx.User.Camp), 2)
	ctx.Game.Log(fmt.Sprintf("%s 的 [地下法则] 触发，战绩区+2红宝石", ctx.User.Name))
	return nil
}

type AdventurerParadiseHandler struct{ skills.BaseHandler }

func (h *AdventurerParadiseHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.Game == nil || ctx.EventCtx == nil {
		return false
	}
	if ctx.EventCtx.ActionType != model.ActionExtract {
		return false
	}
	all := ctx.Game.GetAllPlayers()
	for _, p := range all {
		if p != nil && p.Camp == ctx.User.Camp && p.ID != ctx.User.ID {
			return true
		}
	}
	return false
}

func (h *AdventurerParadiseHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return nil
	}
	transferGem := getSkillFlow(ctx.User, "adventurer_extract_last_gem")
	transferCrystal := getSkillFlow(ctx.User, "adventurer_extract_last_crystal")
	transferTotal := transferGem + transferCrystal
	if transferTotal <= 0 {
		setSkillFlow(ctx.User, "adventurer_extract_requires_paradise", 0)
		ctx.Game.Log(fmt.Sprintf("%s 的 [冒险者天堂] 未检测到本次提炼结果，效果取消", ctx.User.Name))
		return nil
	}

	all := ctx.Game.GetAllPlayers()
	var allyIDs []string
	for _, p := range all {
		if p == nil {
			continue
		}
		if p.Camp != ctx.User.Camp || p.ID == ctx.User.ID {
			continue
		}
		room := playerEnergyCap(p) - (p.Gem + p.Crystal)
		if room >= transferTotal {
			allyIDs = append(allyIDs, p.ID)
		}
	}
	if len(allyIDs) == 0 {
		setSkillFlow(ctx.User, "adventurer_extract_requires_paradise", 0)
		ctx.Game.Log(fmt.Sprintf("%s 的 [冒险者天堂] 无法发动：没有可完整承接%d点提炼能量的队友", ctx.User.Name, transferTotal))
		return nil
	}
	forceTransfer := getSkillFlow(ctx.User, "adventurer_extract_requires_paradise") > 0
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type":      "adventurer_paradise_target",
			"user_id":          ctx.User.ID,
			"ally_ids":         allyIDs,
			"transfer_gem":     transferGem,
			"transfer_crystal": transferCrystal,
			"transfer_total":   transferTotal,
			"from_pending":     forceTransfer,
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 的 [冒险者天堂] 触发，等待选择接收%d点提炼能量的队友", ctx.User.Name, transferTotal))
	return nil
}

type AdventurerStealSkyHandler struct{ skills.BaseHandler }

func (h *AdventurerStealSkyHandler) CanUse(ctx *model.Context) bool {
	return canPayCrystalLike(ctx, 1)
}

func (h *AdventurerStealSkyHandler) Execute(ctx *model.Context) error {
	enemy := model.BlueCamp
	if ctx.User.Camp == model.BlueCamp {
		enemy = model.RedCamp
	}
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type": "adventurer_steal_sky_mode",
			"user_id":     ctx.User.ID,
			"enemy_camp":  string(enemy),
			"self_camp":   string(ctx.User.Camp),
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [偷天换日]，等待选择效果", ctx.User.Name))
	return nil
}
