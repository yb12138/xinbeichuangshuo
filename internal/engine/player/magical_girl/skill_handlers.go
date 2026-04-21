// gameflow: 魔法少女 skill handler。

package magical_girl

import (
	"fmt"

	skills "starcup-engine/internal/engine/skill"
	"starcup-engine/internal/model"
)

type MagicBulletControlHandler struct{ skills.BaseHandler }

func (h *MagicBulletControlHandler) Execute(ctx *model.Context) error {
	// 魔弹掌控由 magic.go/game.go 中的魔弹中断链路直接处理；
	// 这里保留 handler 仅用于维持技能注册表完整。
	return nil
}

type MagicBulletFusionHandler struct{ skills.BaseHandler }

func (h *MagicBulletFusionHandler) Execute(ctx *model.Context) error {
	// 魔弹融合由 PerformMagic 触发的确认中断统一处理；
	// 这里保留 handler 仅用于维持技能注册表完整。
	return nil
}

type MagicBlastHandler struct{ skills.BaseHandler }

func (h *MagicBlastHandler) CanUse(ctx *model.Context) bool {
	// 需要有法术牌可弃才能发动
	for _, card := range ctx.User.Hand {
		if card.Type == model.CardTypeMagic {
			return true
		}
	}
	return false
}

func (h *MagicBlastHandler) Execute(ctx *model.Context) error {
	// 弃牌代价已在 UseSkill 中处理；这里从"我方战绩区 +1 宝石"开始。
	ctx.Game.ModifyGem(string(ctx.User.Camp), 1)
	ctx.Game.Log(fmt.Sprintf("%s 发动 [魔爆冲击]，我方战绩区+1宝石", ctx.User.Name))

	targets := ctx.Targets
	if len(targets) == 0 && ctx.Target != nil {
		targets = []*model.Player{ctx.Target}
	}
	if len(targets) != 2 {
		return fmt.Errorf("魔爆冲击需要且只能指定2名敌方目标")
	}

	targetIDs := make([]string, len(targets))
	for i, t := range targets {
		targetIDs[i] = t.ID
	}

	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptMagicBlast,
		PlayerID: targetIDs[0], // 第一个目标先响应
		Context: map[string]interface{}{
			"choice_type":    "magic_blast",
			"stage":          "target_discard",
			"caster_id":      ctx.User.ID,
			"targets":        targetIDs,
			"current_target": 0,
		},
	})
	ctx.Game.Log(fmt.Sprintf("[Skill] %s 需要选择弃一张法术牌或受到2点伤害", targets[0].Name))

	return nil
}

type DestructionStormHandler struct{ skills.BaseHandler }

func (h *DestructionStormHandler) Execute(ctx *model.Context) error {
	// 毁灭风暴：[宝石] 对任2名目标对手各造成2点法术伤害
	targets := ctx.Targets
	if len(targets) == 0 && ctx.Target != nil {
		targets = []*model.Player{ctx.Target}
	}
	if len(targets) != 2 {
		return fmt.Errorf("毁灭风暴需要且只能指定2名敌方目标")
	}

	for _, t := range targets {
		ctx.Game.InflictDamage(ctx.User.ID, t.ID, 2, model.MagicAttack)
	}

	ctx.Game.Log(fmt.Sprintf("%s 发动 [毁灭风暴]，对 %d 名目标造成伤害", ctx.User.Name, len(targets)))
	return nil
}
