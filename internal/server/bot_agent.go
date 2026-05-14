package server

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"starcup-engine/internal/model"
	"starcup-engine/internal/server/bot"
	"starcup-engine/internal/server/prompting"
)

// scheduleAnyBotIfPrompt 在首回合或流程恢复时检查当前Prompt是否落在机器人身上。
func (r *Room) scheduleAnyBotIfPrompt() {
	// 优先尝试中断类 Prompt。
	r.engineMu.Lock()
	if !r.Started || r.Engine == nil {
		r.engineMu.Unlock()
		return
	}
	prompt := r.Engine.GetCurrentPrompt()
	r.engineMu.Unlock()
	if prompt != nil && prompt.PlayerID != "" {
		r.scheduleBotIfNeeded(prompt.PlayerID, prompting.ClonePrompt(prompt), 0)
		return
	}

	// 兜底：覆盖 ActionSelection / CombatInteraction 这类非中断 AskInput。
	r.mu.RLock()
	for pid, p := range r.botPromptCache {
		c := r.Clients[pid]
		if c == nil || !c.IsBot || p == nil {
			continue
		}
		r.mu.RUnlock()
		r.scheduleBotIfNeeded(pid, prompting.ClonePrompt(p), 0)
		return
	}
	r.mu.RUnlock()
}

// scheduleBotIfNeeded 延时触发机器人决策，避免与当前动作执行重入。
func (r *Room) scheduleBotIfNeeded(playerID string, prompt *model.Prompt, expectedEpoch uint64) {
	if playerID == "" {
		return
	}

	r.mu.RLock()
	if r.BotsPaused {
		r.mu.RUnlock()
		return
	}
	c := r.Clients[playerID]
	if expectedEpoch == 0 {
		expectedEpoch = r.botPromptEpoch
	}
	if prompt == nil {
		if cached, ok := r.botPromptCache[playerID]; ok {
			prompt = prompting.ClonePrompt(cached)
		}
	}
	r.mu.RUnlock()
	if c == nil || !c.IsBot {
		return
	}

	// 兜底：没有显式prompt时，尝试从引擎获取（仅覆盖中断类场景）
	if prompt == nil {
		r.engineMu.Lock()
		if r.Started && r.Engine != nil {
			if p := r.Engine.GetCurrentPrompt(); p != nil && p.PlayerID == playerID {
				prompt = prompting.ClonePrompt(p)
			}
		}
		r.engineMu.Unlock()
	}
	if prompt == nil {
		return
	}

	delay := time.Duration(220+rand.Intn(360)) * time.Millisecond
	time.AfterFunc(delay, func() {
		r.enqueueViaActor(func() error {
			return r.runBotTurn(playerID, prompt, expectedEpoch)
		}, func(err error) {
			log.Printf("[Bot] player=%s decide failed: %v", playerID, err)
			// 给一次“最新状态重试”机会，避免旧快照导致的卡局。
			time.AfterFunc(120*time.Millisecond, func() {
				r.enqueueViaActor(func() error {
					return r.runBotTurn(playerID, nil, 0)
				}, func(retryErr error) {
					log.Printf("[Bot] player=%s retry failed: %v", playerID, retryErr)
				})
			})
		})
	})

	// 守护重试：若提示仍挂在该机器人身上，补一次执行，降低“偶发漏调度”导致的卡局概率。
	time.AfterFunc(delay+1600*time.Millisecond, func() {
		if r.shouldRetryBotTurn(playerID, expectedEpoch) {
			log.Printf("[Bot] player=%s watchdog retry (epoch=%d)", playerID, expectedEpoch)
			r.enqueueViaActor(func() error {
				return r.runBotTurn(playerID, nil, expectedEpoch)
			}, func(err error) {
				log.Printf("[Bot] player=%s watchdog failed: %v", playerID, err)
			})
		}
	})
}

func (r *Room) runBotTurn(playerID string, promptSnapshot *model.Prompt, expectedEpoch uint64) error {
	r.mu.RLock()
	c := r.Clients[playerID]
	currentEpoch := r.botPromptEpoch
	r.mu.RUnlock()
	if c == nil || !c.IsBot {
		return nil
	}
	if expectedEpoch > 0 && currentEpoch != expectedEpoch {
		return nil
	}

	r.engineMu.Lock()
	defer r.engineMu.Unlock()

	if !r.Started || r.Engine == nil {
		return nil
	}
	r.mu.RLock()
	currentEpoch = r.botPromptEpoch
	r.mu.RUnlock()
	if expectedEpoch > 0 && currentEpoch != expectedEpoch {
		return nil
	}
	// 优先使用事件携带的prompt快照（可覆盖CombatInteraction等非中断提示）
	prompt := prompting.ClonePrompt(promptSnapshot)
	if prompt == nil {
		prompt = r.Engine.GetCurrentPrompt()
	}
	if prompt == nil || prompt.PlayerID != playerID {
		// 再试缓存，兼容托管接管时“已有提示但无新事件”
		r.mu.RLock()
		if cached, ok := r.botPromptCache[playerID]; ok {
			prompt = prompting.ClonePrompt(cached)
		}
		r.mu.RUnlock()
		if prompt == nil || prompt.PlayerID != playerID {
			return nil
		}
	}
	if !r.isPromptActionableLocked(playerID, prompt) {
		return nil
	}

	state := r.buildBotStateSnapshot(playerID)
	availableSkills := buildBotAvailableSkills(r.buildAvailableActionSkills(playerID))
	action, ok := r.decideBotAction(playerID, state, prompt, availableSkills)
	if !ok {
		return nil
	}
	action.PlayerID = playerID

	if err := r.Engine.HandleAction(action); err != nil {
		// 保底兜底，尽量避免卡局
		fallback, ok := bot.BuildFallbackAction(playerID, prompt, state)
		if !ok {
			return err
		}
		if fbErr := r.Engine.HandleAction(fallback); fbErr != nil {
			return fmt.Errorf("action=%+v err=%v fallback=%+v fbErr=%v", action, err, fallback, fbErr)
		}
	}
	// 已成功消费提示，清理缓存避免后续误用旧 Prompt。
	r.mu.Lock()
	// 仅在仍是同一提示版本时清理，避免误删新提示。
	if expectedEpoch == 0 || r.botPromptEpoch == expectedEpoch {
		delete(r.botPromptCache, playerID)
	}
	r.mu.Unlock()
	r.Engine.Drive()
	return nil
}

func (r *Room) isPromptActionableLocked(playerID string, prompt *model.Prompt) bool {
	if r.Engine == nil {
		return false
	}
	return bot.CanActPrompt(r.Engine.State, playerID, prompt)
}

func (r *Room) shouldRetryBotTurn(playerID string, expectedEpoch uint64) bool {
	r.mu.RLock()
	c := r.Clients[playerID]
	if c == nil || !c.IsBot {
		r.mu.RUnlock()
		return false
	}
	if expectedEpoch > 0 && r.botPromptEpoch != expectedEpoch {
		r.mu.RUnlock()
		return false
	}
	cached, hasCached := r.botPromptCache[playerID]
	r.mu.RUnlock()
	if !hasCached || cached == nil {
		return false
	}

	r.engineMu.Lock()
	defer r.engineMu.Unlock()
	if !r.Started || r.Engine == nil {
		return false
	}
	prompt := prompting.ClonePrompt(cached)
	return r.isPromptActionableLocked(playerID, prompt)
}

func (r *Room) decideBotAction(playerID string, state bot.StateSnapshot, prompt *model.Prompt, availableSkills []bot.AvailableSkill) (model.PlayerAction, bool) {
	return bot.DecideAction(bot.DecisionInput{
		PlayerID:        playerID,
		State:           state,
		Prompt:          prompt,
		AvailableSkills: availableSkills,
		Memory:          r.botIntel,
	})
}
