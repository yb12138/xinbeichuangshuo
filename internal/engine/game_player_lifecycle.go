// gameflow: 玩家增删、开局、终局。

package engine

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"starcup-engine/internal/data"
	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
)

// AddPlayer 加入一名玩家。role 为角色配置 ID（与房间 CharRole、前端 char_role 一致），仅按 data.Character.ID 绑定，不按展示名匹配。
func (e *GameEngine) AddPlayer(id, name, role string, camp model.Camp) error {
	if len(e.State.Players) >= 6 {
		return errors.New("游戏人数已满 (6人)")
	}
	if _, exists := e.State.Players[id]; exists {
		return errors.New("玩家ID已存在")
	}

	player := &model.Player{
		ID:             id,
		Name:           name,
		Camp:           camp,
		Role:           role,
		Hand:           make([]model.Card, 0),
		Blessings:      make([]model.Card, 0),
		ExclusiveCards: make([]model.Card, 0),
		MaxHand:        6, // 初始手牌上限
		Heal:           0,
		MaxHeal:        2,
		IsActive:       false,
		Tokens:         map[string]int{},
		Orientation:    model.OrientationNormal,
		TurnState:      model.NewPlayerTurnState(),
	}

	// 查找并绑定角色数据
	characters := data.GetCharacters()
	for _, c := range characters {
		if c.ID == role {
			charCopy := c // Copy struct
			player.Character = &charCopy
			player.MaxHand = c.MaxHand
			break
		}
	}
	if player.Character == nil {
		e.Log(fmt.Sprintf("Warning: Character not found for character id %s", role))
	}
	e.runPlayerAddBootstrapTiming(player)

	e.State.Players[id] = player
	e.State.PlayerOrder = append(e.State.PlayerOrder, id)
	e.rebuildTimingOnAttackDeclaredRegistry()
	e.refreshPlayerDerivedState(player)
	return nil
}

// StartGame 开始游戏
func (e *GameEngine) StartGame() error {
	if len(e.State.Players) < 2 {
		return errors.New("玩家人数不足")
	}
	const initialHandSize = 4

	// 1. 初始化牌库
	e.State.Deck = rules.InitDeck()
	e.State.Deck = rules.Shuffle(e.State.Deck)

	// 2. 发初始手牌
	for _, pid := range e.State.PlayerOrder {
		player := e.State.Players[pid]
		cards, newDeck, _ := rules.DrawCards(e.State.Deck, e.State.DiscardPile, initialHandSize)
		player.Hand = append(player.Hand, cards...)
		e.State.Deck = newDeck
		e.runPlayerGameStartBootstrapTiming(player)
	}

	// 3. 随机决定先手
	rand.Seed(time.Now().UnixNano())
	startIndex := rand.Intn(len(e.State.PlayerOrder))
	firstPlayerID := e.State.PlayerOrder[startIndex]
	e.Log(fmt.Sprintf("[Game] 游戏开始! 首发玩家: %s (%s)",
		e.State.Players[firstPlayerID].Name,
		e.State.Players[firstPlayerID].Camp))

	e.State.CurrentTurn = startIndex

	player := e.State.Players[firstPlayerID]
	player.IsActive = true
	player.TurnState = model.NewPlayerTurnState()
	e.actionSummaryTurn = 1

	e.State.TurnStage = model.TurnStageTurnBeforeStart
	e.State.CombatStage = model.CombatStageNone
	e.State.Subflow = model.SubflowNone
	e.resetTurnMagicDamageTracker()

	// 进入第一回合
	e.Drive()

	return nil
}

// checkGameEnd 检查游戏是否结束
func (e *GameEngine) checkGameEnd() {
	// 星杯胜利：任一方星杯达到 5
	if e.State.RedCups >= 5 {
		e.Notify(model.EventGameEnd, "红方胜利！星杯达到 5", nil)
		e.setGameOver(true)
		return
	}
	if e.State.BlueCups >= 5 {
		e.Notify(model.EventGameEnd, "蓝方胜利！星杯达到 5", nil)
		e.setGameOver(true)
		return
	}
	if e.State.RedMorale <= 0 {
		e.Notify(model.EventGameEnd, "蓝方胜利！红方士气归零", nil)
		e.setGameOver(true)
	} else if e.State.BlueMorale <= 0 {
		e.Notify(model.EventGameEnd, "红方胜利！蓝方士气归零", nil)
		e.setGameOver(true)
	}
}
