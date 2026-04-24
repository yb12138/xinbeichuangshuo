// gameflow: 暗杀者延迟后续处理。

package assassin

import (
	"fmt"

	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// FollowupSpecs 导出角色延迟后续声明。
func FollowupSpecs() map[string]engineplayer.FollowupSpec {
	return map[string]engineplayer.FollowupSpec{
		"assassin_stealth_apply": {Label: "Assassin", Resolve: resolveStealthApply},
	}
}

func resolveStealthApply(rt engineplayer.ChoiceRuntime, f model.DeferredFollowup) error {
	user := rt.LookupPlayer(f.UserID)
	if user == nil {
		return fmt.Errorf("暗杀者潜行后续执行者不存在: %s", f.UserID)
	}
	if !engineplayer.IsCharacter(user, "assassin") {
		return fmt.Errorf("仅暗杀者可执行潜行后续")
	}
	rt.ApplyStealthEffect(user)
	return nil
}
