package bard

import (
	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

func responseContext(rt engineplayer.ChoiceRuntime, user *model.Player, stage string, resumePoint interface{}) *model.Context {
	ctx := rt.BuildContext(user, nil, model.TimingActionDuring, &model.EventContext{
		Type:     model.EventNone,
		SourceID: user.ID,
	})
	ctx.Selections["bd_song_stage"] = stage
	ctx.Selections["response_resume_phase"] = resumePoint
	return ctx
}
