package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"starcup-engine/internal/engine"
	"starcup-engine/internal/model"
)

// CLI 实现 GameObserver 接口
type CLI struct {
	CurrentInteractingPlayer string // 当前正在交互的玩家ID (用于多视角模拟)
}

func (c *CLI) OnGameEvent(event model.GameEvent) {
	switch event.Type {
	case model.EventLog:
		fmt.Printf("[LOG] %s\n", event.Message)
	case model.EventStateUpdate:
		// 暂时只打印消息，不全量刷新
		if event.Message != "" {
			fmt.Printf("[STATE] %s\n", event.Message)
		}
	case model.EventAskInput:
		// 解析 Prompt 数据
		if prompt, ok := event.Data.(*model.Prompt); ok {
			printPrompt(prompt)
			// 更新当前交互玩家，并在界面上醒目提示
			c.CurrentInteractingPlayer = prompt.PlayerID
			fmt.Printf("\n>>> 轮到玩家 [%s] 进行操作 <<<\n", prompt.PlayerID)
		} else {
			fmt.Printf("[Prompt] %s\n", event.Message)
		}
	case model.EventError:
		fmt.Printf("[ERROR] %s\n", event.Message)
	case model.EventGameEnd:
		fmt.Printf("[GAME OVER] %s\n", event.Message)
		os.Exit(0)
	}
}

func printPrompt(p *model.Prompt) {
	fmt.Printf("\n>>> %s (Player: %s)\n", p.Message, p.PlayerID)
	if len(p.Options) > 0 {
		fmt.Println("选项:")
		for i, opt := range p.Options {
			fmt.Printf("  %d: %s (ID: %s)\n", i+1, opt.Label, opt.ID)
		}
	} else {
		// 如果没有选项，可能是纯文本提示或需要输入命令
		if p.Type == model.PromptChooseCards {
			fmt.Println("(请输入 'choose <index>' 选择手牌)")
		}
	}
}

func main() {
	cli := &CLI{}
	game := engine.NewGameEngine(cli)

	// 添加玩家 (3v3 Setup)
	// Red Team
	game.AddPlayer("p1", "Player1", "berserker", model.RedCamp)
	game.AddPlayer("p2", "Player2", "blade_master", model.RedCamp)
	game.AddPlayer("p3", "Player3", "sealer", model.RedCamp)

	// Blue Team
	game.AddPlayer("p4", "Player4", "archer", model.BlueCamp)
	game.AddPlayer("p5", "Player5", "assassin", model.BlueCamp)
	game.AddPlayer("p6", "Player6", "angel", model.BlueCamp)

	fmt.Println("=== 星杯传说 CLI (Client-Server 模式) ===")
	fmt.Println("输入 'start' 开始游戏，'help' 查看帮助")

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Intercept client-side commands
		parts := strings.Fields(line)
		cmd := strings.ToLower(parts[0])
		if cmd == "status" {
			printStatus(game)
			continue
		} else if cmd == "hand" {
			// Determine target PID
			pid := ""
			// 优先级 1: 如果有明确指定 (hand p2)
			if len(parts) > 1 {
				pid = parts[1]
			} else {
				// 优先级 2: 优先显示当前交互玩家 (Prompt请求对象)
				if cli.CurrentInteractingPlayer != "" {
					pid = cli.CurrentInteractingPlayer
				} else if game.State.PendingInterrupt != nil {
					// 优先级 3: 如果有中断 (如选技能/弃牌)，显示中断者
					pid = game.State.PendingInterrupt.PlayerID
				} else {
					// 优先级 4: 默认显示当前回合玩家
					if len(game.State.PlayerOrder) > 0 {
						pid = game.State.PlayerOrder[game.State.CurrentTurn]
					}
				}
			}
			printHand(game, pid)
			continue
		} else if cmd == "skills" {
			pid := ""
			if len(parts) > 1 {
				pid = parts[1] // 指定查看某人: skills p2
			} else {
				// 智能推断: 交互对象 > 中断者 > 当前回合玩家
				if cli.CurrentInteractingPlayer != "" {
					pid = cli.CurrentInteractingPlayer
				} else if game.State.PendingInterrupt != nil {
					pid = game.State.PendingInterrupt.PlayerID
				} else if len(game.State.PlayerOrder) > 0 {
					pid = game.State.PlayerOrder[game.State.CurrentTurn]
				}
			}

			printSkills(game, pid)
			continue
		}

		action, err := parseInput(line, game, cli)
		if err != nil {
			fmt.Printf("[CLI Error] %v\n", err)
			continue
		}

		// 如果是 help 命令， parseInput 已经处理了打印，直接 continue
		if action.Type == model.CmdHelp {
			continue
		}

		err = game.HandleAction(action)
		if err != nil {
			// 引擎返回的同步错误 (通常是校验错误)
			fmt.Printf("[Engine Error] %v\n", err)
		}
	}
}

func parseInput(line string, game *engine.GameEngine, cli *CLI) (model.PlayerAction, error) {
	parts := strings.Fields(line)
	cmd := strings.ToLower(parts[0])

	// 智能推断 ID
	pid := ""

	// 策略：始终优先使用当前被请求输入的玩家 (Prompt Target)
	// 只有在没有 Prompt 时，才回退到当前回合玩家
	if cli.CurrentInteractingPlayer != "" {
		pid = cli.CurrentInteractingPlayer
	} else if len(game.State.PlayerOrder) > 0 {
		pid = game.State.PlayerOrder[game.State.CurrentTurn]
	}

	action := model.PlayerAction{
		PlayerID: pid,
	}

	switch cmd {
	case "start":
		action.Type = model.CmdStart
	case "quit", "exit":
		action.Type = model.CmdQuit
	case "pass":
		action.Type = model.CmdPass
	case "buy":
		action.Type = model.CmdBuy
	case "syb":
		action.Type = model.CmdSynthesize
	case "ext":
		action.Type = model.CmdExtract
		case "skill", "use", "s":
			action.Type = model.CmdSkill
			if len(parts) < 2 {
				return action, fmt.Errorf("用法: skill <skill_id> [target_id1] [target_id2...] [discard_indices...]")
			}

			action.SkillID = parts[1]
			
			var targetIDs []string
			var selections []int
			
			// 从第三个参数开始解析（parts[0]是cmd，parts[1]是skill_id）
			argIndex := 2
			
			// 收集目标ID，直到遇到第一个数字或参数结束
			for ; argIndex < len(parts); argIndex++ {
				arg := parts[argIndex]
				if _, err := strconv.Atoi(arg); err != nil {
					// 不是数字，认为是目标ID
					targetIDs = append(targetIDs, arg)
				} else {
					// 遇到数字，停止收集目标ID
					break
				}
			}
			action.TargetIDs = targetIDs
			if len(targetIDs) == 1 {
				action.TargetID = targetIDs[0] // 兼容单目标
			}

			// 解析剩余的参数为弃牌索引
			for ; argIndex < len(parts); argIndex++ {
				val, err := strconv.Atoi(parts[argIndex])
				if err != nil {
					return action, fmt.Errorf("弃牌索引必须是数字: %s", parts[argIndex])
				}
				selections = append(selections, val-1) // 转换为 0-based
			}
			action.Selections = selections
	case "atk": // atk <target> <idx>
		action.Type = model.CmdAttack
		if len(parts) < 3 {
			return action, fmt.Errorf("用法: atk <target_id> <card_index>")
		}
		action.TargetID = parts[1]
		idx, err := strconv.Atoi(parts[2])
		if err != nil {
			return action, fmt.Errorf("卡牌索引必须是数字")
		}
		action.CardIndex = idx - 1 // 转换为 0-based

	// === 2. 修复 响应指令映射 ===
	case "take":
		action.Type = model.CmdRespond
		action.ExtraArgs = []string{"take"} // 标记具体响应类型

	case "defend":
		action.Type = model.CmdRespond
		action.ExtraArgs = []string{"defend"}
		// defend <card_idx> (可选)
		if len(parts) > 1 {
			idx, err := strconv.Atoi(parts[1])
			if err == nil {
				action.CardIndex = idx - 1 // 转 0-based
			} else {
				action.CardIndex = -1 // 自动/默认
			}
		} else {
			action.CardIndex = -1
		}

	case "counter":
		action.Type = model.CmdRespond
		action.ExtraArgs = []string{"counter"}
		// counter <target> <card_idx> (应战攻击)
		// counter <card_idx> (魔弹传递)
		if len(parts) == 2 {
			// counter <card_idx>
			idx, err := strconv.Atoi(parts[1])
			if err != nil {
				return action, fmt.Errorf("卡牌索引必须是数字")
			}
			action.CardIndex = idx - 1
			action.TargetID = "" // 让后端逻辑决定是否需要 TargetID
		} else if len(parts) >= 3 {
			// counter <target> <card_idx>
			action.TargetID = parts[1]
			idx, err := strconv.Atoi(parts[2])
			if err != nil {
				return action, fmt.Errorf("卡牌索引必须是数字")
			}
			action.CardIndex = idx - 1
		} else {
			return action, fmt.Errorf("用法: counter [target_id] <card_index>")
		}

	case "cheat": // cheat <player_id> <card_name> [count]
		action.Type = model.CmdCheat
		if len(parts) < 3 {
			return action, fmt.Errorf("用法: cheat <player_id> <card_name> [count]")
		}
		action.TargetID = parts[1] // Use TargetID as player_id
		action.ExtraArgs = parts[2:]

	case "magic": // magic <target> <idx>
		action.Type = model.CmdMagic
		if len(parts) < 3 {
			return action, fmt.Errorf("用法: magic <target_id> <card_index>")
		}
		action.TargetID = parts[1]
		idx, err := strconv.Atoi(parts[2])
		if err != nil {
			return action, fmt.Errorf("卡牌索引必须是数字")
		}
		action.CardIndex = idx - 1
	case "confirm":
		action.Type = model.CmdConfirm
	case "cancel", "skip":
		action.Type = model.CmdCancel
	case "choose", "select", "discard":
		action.Type = model.CmdSelect
		if len(parts) < 2 {
			return action, fmt.Errorf("用法: choose <idx1> [idx2] ...")
		}
		var selections []int
		for _, s := range parts[1:] {
			val, err := strconv.Atoi(s)
			if err != nil {
				return action, fmt.Errorf("索引必须是数字: %s", s)
			}
			selections = append(selections, val-1) // 转换为 0-based
		}
		action.Selections = selections
	case "help":
		action.Type = model.CmdHelp
		fmt.Println("指令列表:")
		fmt.Println("  start - 开始游戏")
		fmt.Println("  pass - 结束回合")
		fmt.Println("  atk <target> <idx> - 攻击 (idx从1开始)")
		fmt.Println("  magic <target> <idx> - 法术")
		fmt.Println("  buy - 购买")
		fmt.Println("  syb - 合成")
		fmt.Println("  ext - 提炼")
		fmt.Println("  confirm - 确认")
		fmt.Println("  cancel / skip - 取消/跳过")
		fmt.Println("  choose <idx...> - 选择/弃牌")
		fmt.Println("  take - 承受伤害")
		fmt.Println("  defend [idx] - 防御 (可选指定卡牌)")
		fmt.Println("  skills [pid] - 查看技能列表") // <--- 新增
		fmt.Println("  skill <id> [target] [idx...] - 发动主动技能 (如: skill heal p1 1)")
		fmt.Println("  counter <target> <idx> - 应战")
		// 返回一个空的 Help action，或者直接在 main 里处理
		return action, nil
	default:
		// 尝试直接作为数字解析，映射为 confirm (如果是选项) 或 select
		// 这里简单处理，如果输入的是数字，假设是 Select (因为选项也是通过索引选的)
		if _, err := strconv.Atoi(cmd); err == nil {
			action.Type = model.CmdSelect
			var selections []int
			for _, s := range parts {
				val, err := strconv.Atoi(s)
				if err != nil {
					return action, fmt.Errorf("索引必须是数字: %s", s)
				}
				selections = append(selections, val-1)
			}
			action.Selections = selections
		} else {
			return action, fmt.Errorf("未知指令: %s", cmd)
		}
	}

	return action, nil
}

func printStatus(g *engine.GameEngine) {
	fmt.Printf("--- Game Status (Phase: %s) ---\n", g.State.Phase)
	fmt.Printf("Red Morale: %d | Blue Morale: %d\n", g.State.RedMorale, g.State.BlueMorale)

	for _, pid := range g.State.PlayerOrder {
		player := g.State.Players[pid]
		status := fmt.Sprintf("[%s] %s (%s) [角色:%s]: Hand %d/%d, Gem %d, Cry %d, Heal %d",
			player.ID, player.Name, player.Camp, player.Character.Name,
			len(player.Hand), player.MaxHand, player.Gem, player.Crystal, player.Heal)

		// Check field effects instead of Buffs
		var effectNames []string
		for _, fc := range player.Field {
			if fc.Mode == model.FieldEffect {
				effectNames = append(effectNames, fc.Card.Name)
			}
		}

		if len(effectNames) > 0 {
			status += fmt.Sprintf(" [%s]", strings.Join(effectNames, ", "))
		}

		fmt.Println(status)
	}
	fmt.Println()
}

func printHand(g *engine.GameEngine, pid string) {
	player := g.State.Players[pid]
	if player == nil {
		fmt.Printf("Player %s not found\n", pid)
		return
	}

	fmt.Printf("%s's Hand (%d cards):\n", player.Name, len(player.Hand))
	for i, card := range player.Hand {
		fmt.Printf("%d: %s\n", i+1, formatCardInfo(card))
	}
	fmt.Println()
}

// formatCardInfo formats card information (copied from game.go)
func formatCardInfo(card model.Card) string {
	// 基础信息
	info := fmt.Sprintf("[%s] %s", card.Element, card.Name)

	// 类型和伤害
	if card.Type != "" {
		info += fmt.Sprintf(" (%s", card.Type)
		if card.Damage > 0 {
			info += fmt.Sprintf(" Dmg:%d", card.Damage)
		}
		info += ")"
	}

	// 命格
	if card.Faction != "" {
		info += fmt.Sprintf(" [%s命格]", card.Faction)
	}

	// 独有技信息
	exclusiveInfo := []string{}
	if card.ExclusiveChar1 != "" && card.ExclusiveSkill1 != "" {
		exclusiveInfo = append(exclusiveInfo, fmt.Sprintf("%s:%s", card.ExclusiveChar1, card.ExclusiveSkill1))
	}
	if card.ExclusiveChar2 != "" && card.ExclusiveSkill2 != "" {
		exclusiveInfo = append(exclusiveInfo, fmt.Sprintf("%s:%s", card.ExclusiveChar2, card.ExclusiveSkill2))
	}
	if len(exclusiveInfo) > 0 {
		info += fmt.Sprintf(" [%s]", strings.Join(exclusiveInfo, ", "))
	}

	return info
}

// main.go

func printSkills(g *engine.GameEngine, pid string) {
	player := g.State.Players[pid]
	if player == nil {
		fmt.Printf("未找到玩家: %s\n", pid)
		return
	}

	char := player.Character
	fmt.Printf("=== %s 的技能列表 (角色: %s) ===\n", player.Name, char.Name)

	if len(char.Skills) == 0 {
		fmt.Println("  (无技能)")
		return
	}

	for _, skill := range char.Skills {
		// 格式化技能类型
		typeStr := ""
		switch skill.Type {
		case model.SkillTypeAction:
			typeStr = "[主动]" // 需要消耗行动权
		case model.SkillTypePassive:
			typeStr = "[被动]" // 自动生效
		case model.SkillTypeResponse:
			typeStr = "[响应]" // 特定条件下触发
		case model.SkillTypeStartup:
			typeStr = "[启动]" // 通常指启动技或特定限制
		default:
			typeStr = fmt.Sprintf("[%v]", skill.Type)
		}

		// 格式化标签 (如：回合限定)
		tags := ""
		if len(skill.Tags) > 0 {
			var tagNames []string
			for _, t := range skill.Tags {
				tagNames = append(tagNames, string(t))
			}
			tags = fmt.Sprintf(" <%s>", strings.Join(tagNames, ","))
		}

		// 打印技能标题
		fmt.Printf("- %s %s%s (ID: %s)\n", typeStr, skill.Title, tags, skill.ID)

		// 打印描述 (缩进显示)
		fmt.Printf("    %s\n", skill.Description)

		// 如果有消耗说明，也可以打印 (假设 Skill 结构体有 CostDescription)
		// fmt.Printf("    消耗: ...\n")
	}
	fmt.Println()
}
