package data

import "starcup-engine/internal/model"

// GetCharacters 返回所有角色定义
func GetCharacters() []model.Character {
	characters := []model.Character{
		// 1. 天使
		{
			ID:      "angel",
			Name:    "天使",
			Title:   "圣",
			Faction: "圣",
			MaxHand: 6,
			Skills: []model.SkillDefinition{
				{
					ID: "angel_bond", Timings: []model.FlowTiming{model.TimingOnFieldMarkChanged}, Title: "天使羁绊",
					Type:        model.SkillTypePassive,
					Description: "（每当你移除一个基础效果或是使用［圣盾］时）目标角色+1［治疗］。",
					// TimingOnFieldMarkChanged：移除/新增场标（含圣盾放置）均落在同一窗口。
					LogicHandler: "angel_bond",
					// 这是一个被动技能，TargetType 为 None，具体 Target 由运行时 Context 决定
					TargetType:   model.TargetNone,
					ResponseType: model.ResponseSilent, // 自动触发
				},
				{
					ID: "angel_blessing", Timings: []model.FlowTiming{model.TimingActive}, Title: "天使祝福",
					Type:           model.SkillTypeAction,
					Tags:           []model.SkillTag{},
					Description:    "（弃1张水系牌［展示］）指定目标玩家给你2张牌或指定2名角色各给你1张牌。",
					CostDiscards:   1,
					DiscardElement: model.ElementWater, // 需要弃水系牌
					LogicHandler:   "angel_blessing",
					TargetType:     model.TargetSpecific, // 1个玩家 or 2个角色
					MaxTargets:     2,
				},
				{
					ID: "angel_cleanse", Timings: []model.FlowTiming{model.TimingActive}, Title: "风之洁净",
					Type:           model.SkillTypeAction,
					Tags:           []model.SkillTag{},
					Description:    "（弃1张风系牌［展示］）移除场上任意1个基础效果。",
					CostDiscards:   1,
					DiscardElement: model.ElementWind, // 需要弃风系牌
					LogicHandler:   "angel_cleanse",
					TargetType:     model.TargetAny, // 选人移除Buff
					MaxTargets:     1,
				},
				{
					ID: "angel_song", Timings: []model.FlowTiming{model.TimingOnTurnStart}, Title: "天使之歌",
					Type:        model.SkillTypeResponse,
					Tags:        []model.SkillTag{model.TagTurnLimit, model.TagCrystal, model.TagUltimate},
					Description: "［回合限定］［水晶］（在你的回合开始前发动）移除场上任意1个基础效果。",
					CostCrystal: 1,

					LogicHandler: "angel_song",
					TargetType:   model.TargetAny,
					ResponseType: model.ResponseOptional,
				},
				{
					ID: "god_protection", Timings: []model.FlowTiming{model.TimingBeforeMoraleLoss}, Title: "神之庇护",
					Type:        model.SkillTypeResponse,
					Tags:        []model.SkillTag{model.TagCrystal, model.TagUltimate},
					Description: "X个［水晶］为我方抵御X点因法术伤害而造成的士气下降。",
					// CostCrystal: -1, // 逻辑里手动扣除
					// [修改]
					RequiredRole: model.RoleDefender,     // Activated on damage
					ResponseType: model.ResponseOptional, // 玩家选择是否发动
					LogicHandler: "god_protection",
					TargetType:   model.TargetNone,
				},
				{
					ID: "angel_wall", Timings: []model.FlowTiming{model.TimingActive}, Title: "天使之墙",
					Type:             model.SkillTypeAction,
					Tags:             []model.SkillTag{model.TagUnique},
					Description:      "此牌可以当作［圣盾］使用。",
					CostDiscards:     1,                        // 需要指定一张牌
					RequireExclusive: true,                     // 必须使用独有牌
					PlaceCard:        true,                     // 放置场上牌
					PlaceMode:        model.FieldEffect,        // 效果牌
					PlaceEffect:      model.EffectShield,       // 圣盾效果
					PlaceHook:        model.FieldHookOnDamaged, // 受到伤害时触发
					LogicHandler:     "angel_wall",
					TargetType:       model.TargetAny,
				},
			},
			ExclusiveCards: []string{"angel_wall"},
		},
		// 2. 狂战士
		{
			ID:      "berserker",
			Name:    "狂战士",
			Title:   "血",
			Faction: "血",
			MaxHand: 6,
			Skills: []model.SkillDefinition{
				{
					ID: "berserker_frenzy", Timings: []model.FlowTiming{model.TimingOnDamageCalculated, model.TimingOnHitCheck}, Title: "狂化",
					Type:        model.SkillTypePassive,
					Tags:        []model.SkillTag{},
					Description: "你发动的所有攻击伤害额外+1。（攻击命中时②，若你的手牌>3）本次攻击伤害额外+1。",

					RequiredRole: model.RoleAttacker,
					LogicHandler: "berserker_frenzy",
					TargetType:   model.TargetNone,
				},
				{
					ID: "berserker_tear", Timings: []model.FlowTiming{model.TimingOnHitCheck}, Title: "撕裂",
					Type:        model.SkillTypeResponse,
					Tags:        []model.SkillTag{model.TagGem, model.TagUltimate, model.TagOptional},
					Description: "［宝石］攻击命中后发动②，本次攻击伤害额外+2。",
					CostGem:     1,

					RequiredRole: model.RoleAttacker,     // Only runs when attacking
					ResponseType: model.ResponseOptional, // 攻击命中后弹框确认是否发动
					LogicHandler: "berserker_tear",
					TargetType:   model.TargetNone,
				},
				{
					ID: "blood_roar", Timings: []model.FlowTiming{model.TimingOnAttackDeclared}, Title: "血腥咆哮",
					Type:        model.SkillTypeResponse,
					Tags:        []model.SkillTag{model.TagUnique},
					Description: "作为主动攻击打出时发动，若攻击的目标拥有的［治疗］为2，则本次攻击强制命中。",

					RequiredRole: model.RoleAttacker,   // Only runs when attacking
					ResponseType: model.ResponseSilent, // 静默执行，满足条件自动发动
					LogicHandler: "blood_roar",
					TargetType:   model.TargetNone,
				},
				{
					ID: "blood_blade", Timings: []model.FlowTiming{model.TimingOnHitCheck}, Title: "血影狂刀",
					Type:        model.SkillTypeResponse,
					Tags:        []model.SkillTag{model.TagUnique},
					Description: "作为主动攻击打出时发动●若命中后②对手的手牌为2，本次攻击伤害额外+2。●若命中后②对手的手牌为3，本次攻击伤害额外+1。",

					RequiredRole: model.RoleAttacker,
					ResponseType: model.ResponseSilent,
					LogicHandler: "blood_blade",
					TargetType:   model.TargetNone,
				},
			},
			ExclusiveCards: []string{"blood_roar", "blood_blade"},
		},
		// 3. 封印师
		{
			ID:      "sealer",
			Name:    "封印师",
			Title:   "幻",
			Faction: "幻",
			MaxHand: 6,
			Skills: []model.SkillDefinition{
				{
					ID: "magic_surge", Timings: []model.FlowTiming{model.TimingOnActionEnd}, Title: "法术激荡",
					Type:        model.SkillTypeResponse,
					Tags:        []model.SkillTag{},
					Description: "（［法术行动］结束时发动）额外+1［攻击行动］。",
					// 法术行动结束
					ResponseType: model.ResponseOptional, // 可选响应，需要玩家确认
					LogicHandler: "magic_surge",
					TargetType:   model.TargetNone,
				},
				{
					ID: "seal_break", Timings: []model.FlowTiming{model.TimingActive}, Title: "封印破碎",
					Type:         model.SkillTypeAction,
					Tags:         []model.SkillTag{model.TagCrystal, model.TagUltimate},
					Description:  "［水晶］将场上任意一张基础效果牌收入自己手中。",
					CostCrystal:  1,
					LogicHandler: "seal_break",
					TargetType:   model.TargetAny, // 选Buff
				},
				{
					ID: "five_elements_bind", Timings: []model.FlowTiming{model.TimingActive}, Title: "五系束缚",
					Type:             model.SkillTypeAction,
					Tags:             []model.SkillTag{model.TagCrystal},
					Description:      "［水晶］使用开局自带的五系束缚专属技能卡，将其放置于目标对手面前，该对手跳过其下个行动阶段。在其下个行动阶段开始前他可以选择摸（2+X）张牌来取消五系束缚的效果。X为场上封印的数量，X最高为2。不论效果是否发动，触发后移除此牌。",
					CostCrystal:      1,
					CostDiscards:     0,
					RequireExclusive: true,                          // 必须使用独有牌
					PlaceCard:        true,                          // 放置场上牌
					PlaceMode:        model.FieldEffect,             // 效果牌
					PlaceEffect:      model.EffectFiveElementsBind,  // 五系束缚效果
					PlaceHook:        model.FieldHookOnBeforeAction, // 行动阶段开始前触发
					LogicHandler:     "five_elements_bind",
					TargetType:       model.TargetEnemy,
					MinTargets:       1,
					MaxTargets:       1,
				},
				{
					ID: "water_seal", Timings: []model.FlowTiming{model.TimingActive}, Title: "水之封印",
					Type:             model.SkillTypeAction,
					Tags:             []model.SkillTag{model.TagUnique},
					Description:      "（将水之封印放置于目标对手面前）该对手获得（直到他从手中打出或展示出水系牌时强制触发）对他造成3点法术伤害③，触发后移除此牌。",
					CostDiscards:     1,
					DiscardElement:   model.ElementWater,                    // 需要弃水系牌
					DiscardType:      "",                                    // 不限制类型
					RequireExclusive: true,                                  // 必须使用独有牌
					PlaceCard:        true,                                  // 放置场上牌
					PlaceMode:        model.FieldEffect,                     // 效果牌
					PlaceEffect:      model.EffectSealWater,                 // 水之封印效果
					PlaceHook:        model.FieldHookOnCardPlayedOrRevealed, // 打出或展示对应元素牌时触发
					LogicHandler:     "water_seal",
					TargetType:       model.TargetEnemy,
					MinTargets:       1,
					MaxTargets:       1,
				},
				{
					ID: "fire_seal", Timings: []model.FlowTiming{model.TimingActive}, Title: "火之封印",
					Type:             model.SkillTypeAction,
					Tags:             []model.SkillTag{model.TagUnique},
					Description:      "（将火之封印放置于目标对手面前）该对手获得（直到他从手中打出或展示出火系牌时强制触发）对他造成3点法术伤害③，触发后移除此牌。",
					CostDiscards:     1,
					DiscardElement:   model.ElementFire,                     // 需要弃火系牌
					DiscardType:      "",                                    // 不限制类型
					RequireExclusive: true,                                  // 必须使用独有牌
					PlaceCard:        true,                                  // 放置场上牌
					PlaceMode:        model.FieldEffect,                     // 效果牌
					PlaceEffect:      model.EffectSealFire,                  // 火之封印效果
					PlaceHook:        model.FieldHookOnCardPlayedOrRevealed, // 打出或展示对应元素牌时触发
					LogicHandler:     "fire_seal",
					TargetType:       model.TargetEnemy,
					MinTargets:       1,
					MaxTargets:       1,
				},
				{
					ID: "earth_seal", Timings: []model.FlowTiming{model.TimingActive}, Title: "地之封印",
					Type:             model.SkillTypeAction,
					Tags:             []model.SkillTag{model.TagUnique},
					Description:      "（将地之封印放置于目标对手面前）该对手获得（直到他从手中打出或展示出地系牌时强制触发）对他造成3点法术伤害③，触发后移除此牌。",
					CostDiscards:     1,
					DiscardElement:   model.ElementEarth,                    // 需要弃地系牌
					DiscardType:      "",                                    // 不限制类型
					RequireExclusive: true,                                  // 必须使用独有牌
					PlaceCard:        true,                                  // 放置场上牌
					PlaceMode:        model.FieldEffect,                     // 效果牌
					PlaceEffect:      model.EffectSealEarth,                 // 地之封印效果
					PlaceHook:        model.FieldHookOnCardPlayedOrRevealed, // 打出或展示对应元素牌时触发
					LogicHandler:     "earth_seal",
					TargetType:       model.TargetEnemy,
					MinTargets:       1,
					MaxTargets:       1,
				},
				{
					ID: "wind_seal", Timings: []model.FlowTiming{model.TimingActive}, Title: "风之封印",
					Type:             model.SkillTypeAction,
					Tags:             []model.SkillTag{model.TagUnique},
					Description:      "（将风之封印放置于目标对手面前）该对手获得（直到他从手中打出或展示出风系牌时强制触发）对他造成3点法术伤害③，触发后移除此牌。",
					CostDiscards:     1,
					DiscardElement:   model.ElementWind,                     // 需要弃风系牌
					DiscardType:      "",                                    // 不限制类型
					RequireExclusive: true,                                  // 必须使用独有牌
					PlaceCard:        true,                                  // 放置场上牌
					PlaceMode:        model.FieldEffect,                     // 效果牌
					PlaceEffect:      model.EffectSealWind,                  // 风之封印效果
					PlaceHook:        model.FieldHookOnCardPlayedOrRevealed, // 打出或展示对应元素牌时触发
					LogicHandler:     "wind_seal",
					TargetType:       model.TargetEnemy,
					MinTargets:       1,
					MaxTargets:       1,
				},
				{
					ID: "thunder_seal", Timings: []model.FlowTiming{model.TimingActive}, Title: "雷之封印",
					Type:             model.SkillTypeAction,
					Tags:             []model.SkillTag{model.TagUnique},
					Description:      "（将雷之封印放置于目标对手面前）该对手获得（直到他从手中打出或展示出雷系牌时强制触发）对他造成3点法术伤害③，触发后移除此牌。",
					CostDiscards:     1,
					DiscardElement:   model.ElementThunder,                  // 需要弃雷系牌
					DiscardType:      "",                                    // 不限制类型
					RequireExclusive: true,                                  // 必须使用独有牌
					PlaceCard:        true,                                  // 放置场上牌
					PlaceMode:        model.FieldEffect,                     // 效果牌
					PlaceEffect:      model.EffectSealThunder,               // 雷之封印效果
					PlaceHook:        model.FieldHookOnCardPlayedOrRevealed, // 打出或展示对应元素牌时触发
					LogicHandler:     "thunder_seal",
					TargetType:       model.TargetEnemy,
					MinTargets:       1,
					MaxTargets:       1,
				},
			},
			ExclusiveCards: []string{"five_elements_bind", "water_seal", "fire_seal", "earth_seal", "wind_seal", "thunder_seal"},
		},
		// 4. 风之剑圣
		{
			ID:      "blade_master",
			Name:    "风之剑圣",
			Title:   "技",
			Faction: "技",
			MaxHand: 6,
			Skills: []model.SkillDefinition{
				{
					ID: "wind_fury", Timings: []model.FlowTiming{model.TimingOnActionEnd}, Title: "风怒追击",
					Type:        model.SkillTypeResponse,
					Tags:        []model.SkillTag{model.TagTurnLimit},
					Description: "［回合限定］（［攻击行动］结束时发动）额外+1风系［攻击行动］。",
					// 攻击行动结束
					ResponseType: model.ResponseOptional, // 可选响应，需要玩家确认
					RequiredRole: model.RoleAttacker,
					LogicHandler: "wind_fury",
					TargetType:   model.TargetNone,
				},
				{
					ID: "holy_sword", Timings: []model.FlowTiming{model.TimingOnAttackDeclared}, Title: "圣剑",
					Type:        model.SkillTypePassive,
					Tags:        []model.SkillTag{},
					Description: "若你的主动攻击为你本次行动阶段的第三次［攻击行动］，则此攻击强制命中。本次［攻击行动］结束后，你摸X张牌，弃X张牌（X<4）。",

					RequiredRole: model.RoleAttacker, // Only runs when attacking
					LogicHandler: "holy_sword",
					TargetType:   model.TargetNone,
				},
				{
					ID: "sword_shadow", Timings: []model.FlowTiming{model.TimingOnActionEnd}, Title: "剑影",
					Type:        model.SkillTypeResponse,
					Tags:        []model.SkillTag{model.TagTurnLimit, model.TagCrystal, model.TagUltimate},
					Description: "［回合限定］［蓝水晶］（［攻击行动］结束时发动）额外+1［攻击行动］。",
					CostCrystal: 1,

					RequiredRole: model.RoleAttacker,
					ResponseType: model.ResponseOptional, // 可选响应，需要玩家确认
					LogicHandler: "sword_shadow",
					TargetType:   model.TargetNone,
				},
				{
					ID: "gale_skill", Timings: []model.FlowTiming{model.TimingOnAttackDeclared}, Title: "疾风技",
					Type:        model.SkillTypeResponse,
					Tags:        []model.SkillTag{model.TagUnique},
					Description: "（作为主动攻击打出时发动）额外+1［攻击行动］。",

					RequiredRole: model.RoleAttacker,   // Only runs when attacking
					ResponseType: model.ResponseSilent, // 静默执行，独有技自动发动
					LogicHandler: "gale_skill",
					TargetType:   model.TargetNone,
				},
				{
					ID: "gale_slash", Timings: []model.FlowTiming{model.TimingOnAttackDeclared}, Title: "列风技",
					Type:        model.SkillTypeResponse,
					Tags:        []model.SkillTag{model.TagUnique},
					Description: "（攻击目标拥有圣盾时发动）无视对手圣盾的效果，且此攻击对手无法应战。",

					RequiredRole: model.RoleAttacker,   // Only runs when attacking
					ResponseType: model.ResponseSilent, // 静默执行，独有技自动发动
					LogicHandler: "gale_slash",
					TargetType:   model.TargetNone,
				},
			},
			ExclusiveCards: []string{"gale_skill", "gale_slash"},
		},
		// 5. 神箭手
		{
			ID:      "archer",
			Name:    "神箭手",
			Title:   "技",
			Faction: "技",
			MaxHand: 6,
			Skills: []model.SkillDefinition{
				{
					ID: "piercing_shot", Timings: []model.FlowTiming{model.TimingOnHitCheck}, Title: "贯穿射击",
					Type:           model.SkillTypeResponse,
					Description:    "（主动攻击未命中时发动②，弃1张法术牌［展示］）对你所攻击的目标造成2点法术伤害③。",
					CostDiscards:   1,
					DiscardType:    model.CardTypeMagic, // 需要弃法术牌
					DiscardElement: "",                  // 不限制元素

					RequiredRole:    model.RoleAttacker, // Only runs when attacking
					ResponseType:    model.ResponseOptional,
					LogicHandler:    "piercing_shot",
					TargetType:      model.TargetNone,
					InteractionType: model.InteractionDiscard, // 新增
					InteractionConfig: model.InteractionConfig{ // 新增
						MinSelect: 1,
						MaxSelect: 1,
						Prompt:    "请选择一张法术牌发动贯穿射击",
					},
				},
				{
					ID: "lightning_arrow", Timings: []model.FlowTiming{model.TimingOnAttackDeclared}, Title: "闪电箭",
					Type:        model.SkillTypePassive,
					Tags:        []model.SkillTag{},
					Description: "你的雷系攻击对手无法应战。",

					RequiredRole: model.RoleAttacker, // Only runs when attacking
					LogicHandler: "lightning_arrow",
					TargetType:   model.TargetNone,
					ResponseType: model.ResponseSilent, // 自动触发
				},
				{
					ID: "snipe", Timings: []model.FlowTiming{model.TimingActive}, Title: "狙击",
					Type:         model.SkillTypeAction,
					Tags:         []model.SkillTag{model.TagCrystal, model.TagUltimate},
					Description:  "［水晶］目标角色手牌补到5张［强制］，额外+1［攻击行动］。",
					CostCrystal:  1,
					LogicHandler: "snipe",
					TargetType:   model.TargetAny,
				},
				{
					ID: "precise_shot", Timings: []model.FlowTiming{model.TimingOnAttackDeclared}, Title: "精准射击",
					Type:        model.SkillTypeResponse,
					Tags:        []model.SkillTag{model.TagUnique},
					Description: "此攻击强制命中，但本次攻击伤害-1。",

					RequiredRole:     model.RoleAttacker,     // Only runs when attacking
					RequireExclusive: true,                   // 必须打出独有牌
					ResponseType:     model.ResponseOptional, // 打出独有牌时询问是否发动
					LogicHandler:     "precise_shot",
					TargetType:       model.TargetNone,
				},
				{
					ID: "flash_trap", Timings: []model.FlowTiming{model.TimingActive}, Title: "闪光陷阱",
					Type:             model.SkillTypeAction,
					Tags:             []model.SkillTag{model.TagUnique},
					Description:      "对目标角色造成2点法术伤害③。",
					CostDiscards:     1,
					RequireExclusive: true,
					LogicHandler:     "flash_trap",
					TargetType:       model.TargetAny,
				},
			},
			ExclusiveCards: []string{"precise_shot", "flash_trap"},
		},
		// 6. 暗杀者
		{
			ID:      "assassin",
			Name:    "暗杀者",
			Title:   "技",
			Faction: "技",
			MaxHand: 6,
			Skills: []model.SkillDefinition{
				{
					ID: "backlash", Timings: []model.FlowTiming{model.TimingOnDamageTaken}, Title: "反噬",
					Type:        model.SkillTypePassive,
					Tags:        []model.SkillTag{},
					Description: "（承受攻击伤害时发动⑥）攻击你的对手摸1张牌［强制］。",

					RequiredRole: model.RoleDefender, // Only runs when taking damage
					LogicHandler: "backlash",
					TargetType:   model.TargetNone,
					ResponseType: model.ResponseSilent, // 自动触发
				},
				{
					ID: "water_shadow", Timings: []model.FlowTiming{model.TimingBeforeCardDrawn}, Title: "水影",
					Type:        model.SkillTypeResponse,
					Tags:        []model.SkillTag{},
					Description: "摸牌前可弃X张水系牌；潜行状态下可额外弃1张法术牌。",

					RequiredRole:    model.RoleAny,
					ResponseType:    model.ResponseOptional, // 可选触发
					LogicHandler:    "water_shadow",
					TargetType:      model.TargetNone,
					InteractionType: model.InteractionDiscard,
					InteractionConfig: model.InteractionConfig{
						MinSelect: 1,
						MaxSelect: 99, // 允许选择任意数量
						Prompt:    "请选择要弃置的卡牌发动水影技能",
					},
				},
				{
					ID: "stealth", Timings: []model.FlowTiming{model.TimingOnTurnStart}, Title: "潜行",
					Type:        model.SkillTypeStartup,
					Tags:        []model.SkillTag{model.TagGem, model.TagUltimate},
					Description: "［宝石］你可选择摸1张牌，［横置］持续到你的下个行动阶段开始，你的手牌上限-1；你不能成为主动攻击的目标；你的主动攻击对方无法应战且伤害额外+X，X为你剩余的能量数。潜行的效果结束时角色［转正］。",
					CostGem:     1,

					LogicHandler: "stealth",
					TargetType:   model.TargetNone,
				},
			},
			ExclusiveCards: []string{},
		},
		// 7. 圣女
		{
			ID:      "saintess",
			Name:    "圣女",
			Title:   "圣",
			Faction: "圣",
			MaxHand: 6,
			Skills: []model.SkillDefinition{
				{
					ID: "frost_prayer", Timings: []model.FlowTiming{model.TimingOnCardPlayedOrRevealed}, Title: "冰霜祷言",
					Type:        model.SkillTypePassive,
					Tags:        []model.SkillTag{},
					Description: "（每当你打出或展示水系牌或圣光时发动）目标角色+1［治疗］。",

					// 展示卡牌时也触发
					LogicHandler: "frost_prayer",
					TargetType:   model.TargetAny,
					ResponseType: model.ResponseSilent, // 自动触发
				},
				{
					ID: "healing_light", Timings: []model.FlowTiming{model.TimingActive}, Title: "治愈之光",
					Type:             model.SkillTypeAction,
					Tags:             []model.SkillTag{model.TagUnique},
					Description:      "指定最多3名角色各+1［治疗］。",
					CostDiscards:     1,
					RequireExclusive: true,
					LogicHandler:     "healing_light",
					TargetType:       model.TargetAny,
					MaxTargets:       3,
				},
				{
					ID: "heal", Timings: []model.FlowTiming{model.TimingActive}, Title: "治疗术",
					Type:             model.SkillTypeAction,
					Tags:             []model.SkillTag{model.TagUnique},
					Description:      "目标角色+2［治疗］。",
					CostDiscards:     1,
					RequireExclusive: true,
					LogicHandler:     "heal",
					TargetType:       model.TargetAny,
				},
				{
					ID: "saint_heal", Timings: []model.FlowTiming{model.TimingActive}, Title: "圣疗",
					Type:         model.SkillTypeAction,
					Tags:         []model.SkillTag{model.TagTurnLimit, model.TagCrystal, model.TagUltimate},
					Description:  "［回合限定］［水晶］任意分配3点［治疗］给1~3名角色，额外+1［攻击行动］或［法术行动］。",
					CostCrystal:  1,
					LogicHandler: "saint_heal",
					TargetType:   model.TargetAny,
					MaxTargets:   3,
				},
				{
					ID: "mercy", Timings: []model.FlowTiming{model.TimingOnTurnStart}, Title: "怜悯",
					Type:        model.SkillTypeStartup,
					Tags:        []model.SkillTag{model.TagGem, model.TagUltimate},
					Description: "［持续］［宝石］［横置］，你的手牌上限恒定为7［恒定］，你+1［水晶］。",
					CostGem:     1,

					LogicHandler: "mercy",
					TargetType:   model.TargetNone,
				},
			},
			ExclusiveCards: []string{"healing_light", "heal"},
		},
		// 8. 魔法少女
		{
			ID:      "magical_girl",
			Name:    "魔法少女",
			Title:   "咏",
			Faction: "咏",
			MaxHand: 6,
			Skills: []model.SkillDefinition{
				{
					ID: "magic_bullet_control", Timings: []model.FlowTiming{model.TimingActive}, Title: "魔弹掌控",
					Type:        model.SkillTypeResponse,
					Tags:        []model.SkillTag{},
					Description: "你主动使用魔弹时可以选择逆向传递。",
					// 由 PerformMagic 的魔弹链路主动触发
					LogicHandler: "magic_bullet_control",
					TargetType:   model.TargetNone,
					ResponseType: model.ResponseOptional,
				},
				{
					ID: "magic_bullet_fusion", Timings: []model.FlowTiming{model.TimingActive}, Title: "魔弹融合",
					Type:         model.SkillTypeAction,
					Tags:         []model.SkillTag{},
					Description:  "你的地系或火系牌可以当魔弹使用。",
					LogicHandler: "magic_bullet_fusion",
					TargetType:   model.TargetNone,
					CostDiscards: 1,
				},
				{
					ID: "magic_bullet_fusion_chain", Timings: []model.FlowTiming{model.TimingOnHitCheck}, Title: "魔弹融合",
					Type:            model.SkillTypeResponse,
					Tags:            []model.SkillTag{},
					Description:     "当你成为魔弹目标时，可以打出1张地系或火系牌视为魔弹继续传递。",
					LogicHandler:    "magic_bullet_fusion_chain",
					TargetType:      model.TargetNone,
					ResponseType:    model.ResponseOptional,
					InteractionType: model.InteractionDiscard,
					InteractionConfig: model.InteractionConfig{
						MinSelect: 1,
						MaxSelect: 1,
						Prompt:    "请选择1张地系或火系牌发动魔弹融合",
					},
				},
				{
					ID: "magic_blast", Timings: []model.FlowTiming{model.TimingActive}, Title: "魔爆冲击",
					Type:         model.SkillTypeAction,
					Tags:         []model.SkillTag{},
					Description:  "（弃1张法术牌［展示］）我方战绩区+1颗［宝石］；选择2名目标对手各弃1张法术牌［展示］，每有人不如此做，你对他造成2点法术伤害，然后你弃1张牌。",
					CostDiscards: 1,
					DiscardType:  model.CardTypeMagic,
					LogicHandler: "magic_blast",
					TargetType:   model.TargetEnemy,
					MinTargets:   2,
					MaxTargets:   2,
				},
				{
					ID: "destruction_storm", Timings: []model.FlowTiming{model.TimingActive}, Title: "毁灭风暴",
					Type:         model.SkillTypeAction,
					Tags:         []model.SkillTag{model.TagGem, model.TagUltimate},
					Description:  "［宝石］对任2名目标对手各造成2点法术伤害③。",
					CostGem:      1,
					LogicHandler: "destruction_storm",
					TargetType:   model.TargetEnemy,
					MinTargets:   2,
					MaxTargets:   2,
				},
			},
			ExclusiveCards: []string{},
		},
		// 9. 女武神
		{
			ID:      "valkyrie",
			Name:    "女武神",
			Title:   "圣",
			Faction: "圣",
			MaxHand: 6,
			Skills: []model.SkillDefinition{
				{
					ID: "valkyrie_divine_pursuit", Timings: []model.FlowTiming{model.TimingOnActionEnd}, Title: "神圣追击",
					Type:        model.SkillTypeResponse,
					Description: "攻击/法术行动结束时，若你有治疗，可移除1点治疗，额外+1攻击行动。",

					RequiredRole: model.RoleAttacker,
					ResponseType: model.ResponseOptional,
					LogicHandler: "valkyrie_divine_pursuit",
					TargetType:   model.TargetNone,
				},
				{
					ID: "valkyrie_order_seal", Timings: []model.FlowTiming{model.TimingActive}, Title: "秩序之印",
					Type:         model.SkillTypeAction,
					Description:  "摸2张牌，然后自身+1治疗和+1蓝水晶。",
					LogicHandler: "valkyrie_order_seal",
					TargetType:   model.TargetNone,
				},
				{
					ID: "valkyrie_peace_walker", Timings: []model.FlowTiming{model.TimingOnAttackDeclared}, Title: "和平行者",
					Type:        model.SkillTypePassive,
					Description: "你的回合内发动英灵召唤后进入英灵形态；当你执行主动攻击行动时，转正并脱离英灵形态。",

					RequiredRole: model.RoleAttacker,
					ResponseType: model.ResponseSilent,
					LogicHandler: "valkyrie_peace_walker",
					TargetType:   model.TargetNone,
				},
				{
					ID: "valkyrie_military_glory", Timings: []model.FlowTiming{model.TimingOnTurnStart}, Title: "军威神光",
					Type:        model.SkillTypeStartup,
					Description: "（回合开始时，若你处于［英灵形态］）选择以下1项发动：●你+1［治疗］，［转正］脱离［英灵型态］。●（移除我方战绩区X个星石，X<3）目标角色+X［治疗］。",

					ResponseType: model.ResponseOptional,
					LogicHandler: "valkyrie_military_glory",
					TargetType:   model.TargetNone,
				},
				{
					ID: "valkyrie_heroic_summon", Timings: []model.FlowTiming{model.TimingOnHitCheck}, Title: "英灵召唤",
					Type:        model.SkillTypeResponse,
					Tags:        []model.SkillTag{model.TagCrystal, model.TagUltimate},
					Description: "［水晶］（攻击命中时发动②）本次攻击伤害额外+1，（若你额外弃1张法术牌［展示］）目标角色+1［治疗］。",

					RequiredRole: model.RoleAttacker,
					ResponseType: model.ResponseOptional,
					LogicHandler: "valkyrie_heroic_summon",
					TargetType:   model.TargetNone,
				},
			},
			ExclusiveCards: []string{},
		},
		// 10. 元素师
		{
			ID:      "elementalist",
			Name:    "元素师",
			Title:   "咏",
			Faction: "咏",
			MaxHand: 6,
			Skills: []model.SkillDefinition{
				{
					ID: "elementalist_absorb", Timings: []model.FlowTiming{model.TimingOnDamageTaken}, Title: "元素吸收",
					Type:        model.SkillTypeResponse,
					Description: "你造成法术伤害后，元素+1（上限3）。元素点燃造成的伤害不触发。",

					RequiredRole: model.RoleAttacker,
					ResponseType: model.ResponseSilent,
					LogicHandler: "elementalist_absorb",
					TargetType:   model.TargetNone,
				},
				{
					ID: "elementalist_ignite", Timings: []model.FlowTiming{model.TimingActive}, Title: "元素点燃",
					Type:         model.SkillTypeAction,
					Description:  "移除3点元素，对目标敌方角色造成2点法术伤害，并额外+1法术行动。",
					LogicHandler: "elementalist_ignite",
					TargetType:   model.TargetEnemy,
					MinTargets:   1,
					MaxTargets:   1,
				},
				{
					ID: "elementalist_thunder_strike", Timings: []model.FlowTiming{model.TimingActive}, Title: "雷击",
					Type:             model.SkillTypeAction,
					Tags:             []model.SkillTag{model.TagUnique},
					Description:      "独有技法术：对目标敌方角色造成1点法术伤害；可额外弃1张雷系牌使伤害+1；阵营+1宝石。",
					CostDiscards:     1,
					RequireExclusive: true,
					LogicHandler:     "elementalist_thunder_strike",
					TargetType:       model.TargetEnemy,
					MinTargets:       1,
					MaxTargets:       1,
				},
				{
					ID: "elementalist_freeze", Timings: []model.FlowTiming{model.TimingActive}, Title: "冰冻",
					Type:             model.SkillTypeAction,
					Tags:             []model.SkillTag{model.TagUnique},
					Description:      "独有技法术：选择1名角色受1点法术伤害，再选择1名角色+1治疗；可额外弃1张水系牌使伤害+1。",
					CostDiscards:     1,
					RequireExclusive: true,
					LogicHandler:     "elementalist_freeze",
					TargetType:       model.TargetAny,
					MinTargets:       2,
					MaxTargets:       2,
				},
				{
					ID: "elementalist_wind_blade", Timings: []model.FlowTiming{model.TimingActive}, Title: "风刃",
					Type:             model.SkillTypeAction,
					Tags:             []model.SkillTag{model.TagUnique},
					Description:      "独有技法术：对目标敌方角色造成1点法术伤害，额外+1攻击行动；可额外弃1张风系牌使伤害+1。",
					CostDiscards:     1,
					RequireExclusive: true,
					LogicHandler:     "elementalist_wind_blade",
					TargetType:       model.TargetEnemy,
					MinTargets:       1,
					MaxTargets:       1,
				},
				{
					ID: "elementalist_meteor", Timings: []model.FlowTiming{model.TimingActive}, Title: "陨石",
					Type:             model.SkillTypeAction,
					Tags:             []model.SkillTag{model.TagUnique},
					Description:      "独有技法术：对目标敌方角色造成1点法术伤害，额外+1法术行动；可额外弃1张地系牌使伤害+1。",
					CostDiscards:     1,
					RequireExclusive: true,
					LogicHandler:     "elementalist_meteor",
					TargetType:       model.TargetEnemy,
					MinTargets:       1,
					MaxTargets:       1,
				},
				{
					ID: "elementalist_fireball", Timings: []model.FlowTiming{model.TimingActive}, Title: "火球",
					Type:             model.SkillTypeAction,
					Tags:             []model.SkillTag{model.TagUnique},
					Description:      "独有技法术：对目标敌方角色造成2点法术伤害；可额外弃1张火系牌使伤害+1。",
					CostDiscards:     1,
					RequireExclusive: true,
					LogicHandler:     "elementalist_fireball",
					TargetType:       model.TargetEnemy,
					MinTargets:       1,
					MaxTargets:       1,
				},
				{
					ID: "elementalist_moonlight", Timings: []model.FlowTiming{model.TimingActive}, Title: "月光",
					Type:         model.SkillTypeAction,
					Tags:         []model.SkillTag{model.TagGem, model.TagUltimate},
					Description:  "［宝石］对目标敌方角色造成（X+1）点法术伤害，X为发动后剩余能量数。",
					CostGem:      1,
					LogicHandler: "elementalist_moonlight",
					TargetType:   model.TargetEnemy,
					MinTargets:   1,
					MaxTargets:   1,
				},
			},
			ExclusiveCards: []string{"elementalist_thunder_strike", "elementalist_freeze", "elementalist_wind_blade", "elementalist_meteor", "elementalist_fireball"},
		},
		// 11. 仲裁者
		{
			ID:      "arbiter",
			Name:    "仲裁者",
			Title:   "幻",
			Faction: "幻",
			MaxHand: 6,
			Skills: []model.SkillDefinition{
				{
					ID: "arbiter_law", Timings: []model.FlowTiming{model.TimingActive}, Title: "仲裁法则",
					Type:         model.SkillTypePassive,
					Description:  "游戏开始时获得2个蓝水晶。",
					ResponseType: model.ResponseSilent,
					LogicHandler: "arbiter_law",
					TargetType:   model.TargetNone,
				},
				{
					ID: "arbiter_judgment_tide", Timings: []model.FlowTiming{model.TimingOnDamageTaken}, Title: "审判浪潮",
					Type:        model.SkillTypePassive,
					Description: "每次承受伤害时，审判+1（上限4）。",

					RequiredRole: model.RoleDefender,
					ResponseType: model.ResponseSilent,
					LogicHandler: "arbiter_judgment_tide",
					TargetType:   model.TargetNone,
				},
				{
					ID: "arbiter_ritual", Timings: []model.FlowTiming{model.TimingOnTurnStart}, Title: "仲裁仪式",
					Type:        model.SkillTypeStartup,
					Tags:        []model.SkillTag{model.TagGem, model.TagUltimate},
					Description: "启动阶段可消耗1宝石进入审判形态：横置，手牌上限恒定为5；审判形态下每次自己回合开始审判+1。",

					ResponseType: model.ResponseOptional,
					LogicHandler: "arbiter_ritual",
					TargetType:   model.TargetNone,
				},
				{
					ID: "arbiter_ritual_break", Timings: []model.FlowTiming{model.TimingOnTurnStart}, Title: "仪式中断",
					Type:        model.SkillTypeStartup,
					Description: "启动阶段若处于审判形态，可转正脱离并使我方战绩区+1宝石。",

					ResponseType: model.ResponseOptional,
					LogicHandler: "arbiter_ritual_break",
					TargetType:   model.TargetNone,
				},
				{
					ID: "arbiter_doomsday", Timings: []model.FlowTiming{model.TimingActive}, Title: "末日审判",
					Type:         model.SkillTypeAction,
					Description:  "移除当前所有审判点数，对1名敌方角色造成等量法术伤害；若行动阶段开始时审判已达上限，则该阶段必须发动此技能。",
					LogicHandler: "arbiter_doomsday",
					TargetType:   model.TargetEnemy,
					MinTargets:   1,
					MaxTargets:   1,
				},
				{
					ID: "arbiter_balance", Timings: []model.FlowTiming{model.TimingActive}, Title: "判决天平",
					Type:         model.SkillTypeAction,
					Tags:         []model.SkillTag{model.TagCrystal, model.TagUltimate},
					Description:  "［水晶］你+1［审判］，再选择以下一项发动：●弃掉你的所有手牌。●将你的手牌补到上限［强制］，我方战绩区+1［宝石］。",
					CostCrystal:  1,
					LogicHandler: "arbiter_balance",
					TargetType:   model.TargetNone,
				},
			},
			ExclusiveCards: []string{},
		},
		// 12. 冒险家
		{
			ID:      "adventurer",
			Name:    "冒险家",
			Title:   "技",
			Faction: "技",
			MaxHand: 6,
			Skills: []model.SkillDefinition{
				{
					ID: "adventurer_fraud", Timings: []model.FlowTiming{model.TimingActive}, Title: "欺诈",
					Type:         model.SkillTypeAction,
					Description:  "主动技能：选择1名敌方角色，弃同系牌将本次视为一次主动攻击（弃2张同系可选五系攻击〔不含暗灭〕；弃3张同系视为暗灭）。",
					LogicHandler: "adventurer_fraud",
					TargetType:   model.TargetEnemy,
					MinTargets:   1,
					MaxTargets:   1,
				},
				{
					ID: "adventurer_lucky_fortune", Timings: []model.FlowTiming{model.TimingOnAttackDeclared}, Title: "强运",
					Type:         model.SkillTypePassive,
					Description:  "发动欺诈后，+1蓝水晶。",
					LogicHandler: "adventurer_lucky_fortune",
					TargetType:   model.TargetNone,
				},
				{
					ID: "adventurer_underground_law", Timings: []model.FlowTiming{model.TimingActive}, Title: "地下法则",
					Type:         model.SkillTypePassive,
					Description:  "你执行购买时，改为我方战绩区+2宝石。",
					LogicHandler: "adventurer_underground_law",
					TargetType:   model.TargetNone,
				},
				{
					ID: "adventurer_paradise", Timings: []model.FlowTiming{model.TimingActive}, Title: "冒险者天堂",
					Type:         model.SkillTypeResponse,
					Description:  "你执行提炼时，可将本次提炼出的［宝石］和［水晶］全部交给1名队友（不能拆分），然后移除你的1［能量］。",
					ResponseType: model.ResponseOptional,
					LogicHandler: "adventurer_paradise",
					TargetType:   model.TargetNone,
				},
				{
					ID: "adventurer_steal_sky", Timings: []model.FlowTiming{model.TimingActive}, Title: "偷天换日",
					Type:         model.SkillTypeAction,
					Tags:         []model.SkillTag{model.TagTurnLimit, model.TagCrystal, model.TagUltimate},
					Description:  "［回合限定］［水晶］二选一：转移对方战绩区1个［宝石］到我方；或将我方战绩区所有［水晶］转为［宝石］。随后额外+1［攻击行动］或［法术行动］。",
					CostCrystal:  1,
					LogicHandler: "adventurer_steal_sky",
					TargetType:   model.TargetNone,
				},
			},
			ExclusiveCards: []string{},
		},
		// 13. 圣枪骑士
		{
			ID:      "holy_lancer",
			Name:    "圣枪骑士",
			Title:   "圣",
			Faction: "圣",
			MaxHand: 6,
			Skills: []model.SkillDefinition{
				{
					ID: "holy_lancer_revelation", Timings: []model.FlowTiming{model.TimingActive}, Title: "神圣启示",
					Type:         model.SkillTypePassive,
					Description:  "当我方星杯数不小于对方时，治疗上限+1。",
					ResponseType: model.ResponseSilent,
					LogicHandler: "holy_lancer_revelation",
					TargetType:   model.TargetNone,
				},
				{
					ID: "holy_lancer_radiance", Timings: []model.FlowTiming{model.TimingActive}, Title: "辉耀",
					Type:           model.SkillTypeAction,
					Description:    "弃1张水系牌，全场各+1治疗，同时额外+1攻击行动。",
					CostDiscards:   1,
					DiscardElement: model.ElementWater,
					LogicHandler:   "holy_lancer_radiance",
					TargetType:     model.TargetNone,
				},
				{
					ID: "holy_lancer_punishment", Timings: []model.FlowTiming{model.TimingActive}, Title: "惩戒",
					Type:         model.SkillTypeAction,
					Description:  "弃1张法术牌，将任意其他角色1点治疗转移给你，并额外+1攻击行动。",
					CostDiscards: 1,
					DiscardType:  model.CardTypeMagic,
					LogicHandler: "holy_lancer_punishment",
					TargetType:   model.TargetAny,
					MinTargets:   1,
					MaxTargets:   1,
				},
				{
					ID: "holy_lancer_holy_strike", Timings: []model.FlowTiming{model.TimingOnHitCheck}, Title: "圣击",
					Type:        model.SkillTypeResponse,
					Description: "攻击命中后，若本次未发动天枪/地枪，则+1治疗。",

					RequiredRole: model.RoleAttacker,
					ResponseType: model.ResponseSilent,
					LogicHandler: "holy_lancer_holy_strike",
					TargetType:   model.TargetNone,
				},
				{
					ID: "holy_lancer_sky_spear", Timings: []model.FlowTiming{model.TimingOnAttackDeclared}, Title: "天枪",
					Type:        model.SkillTypeResponse,
					Description: "主动攻击前，若治疗≥2且本回合未发动圣光祈愈，可移除2治疗使本次攻击不可应战。",

					RequiredRole: model.RoleAttacker,
					ResponseType: model.ResponseOptional,
					LogicHandler: "holy_lancer_sky_spear",
					TargetType:   model.TargetNone,
				},
				{
					ID: "holy_lancer_earth_spear", Timings: []model.FlowTiming{model.TimingOnHitCheck}, Title: "地枪",
					Type:        model.SkillTypeResponse,
					Description: "（主动攻击命中后发动②）移除你的X点［治疗］，本次攻击伤害额外+X，X最高为4；不能和［圣击］同时发动。",

					RequiredRole: model.RoleAttacker,
					ResponseType: model.ResponseOptional,
					LogicHandler: "holy_lancer_earth_spear",
					TargetType:   model.TargetNone,
				},
				{
					ID: "holy_lancer_prayer", Timings: []model.FlowTiming{model.TimingActive}, Title: "圣光祈愈",
					Type:         model.SkillTypeAction,
					Tags:         []model.SkillTag{model.TagGem, model.TagUltimate},
					Description:  "［宝石］无视治疗上限为你+2治疗，但治疗数最高为5；额外+1攻击行动；本回合你不能再发动天枪。",
					CostGem:      1,
					LogicHandler: "holy_lancer_prayer",
					TargetType:   model.TargetNone,
				},
			},
			ExclusiveCards: []string{},
		},
		// 14. 精灵射手
		{
			ID:      "elf_archer",
			Name:    "精灵射手",
			Title:   "灵",
			Faction: "灵",
			MaxHand: 6,
			Skills: []model.SkillDefinition{
				{
					ID: "elf_elemental_shot", Timings: []model.FlowTiming{model.TimingOnAttackDeclared}, Title: "元素射击",
					Type:        model.SkillTypeResponse,
					Tags:        []model.SkillTag{model.TagTurnLimit, model.TagOptional},
					Description: "[回合限定]（主动攻击时①，若攻击牌非暗系，弃1张法术牌[展示]或移除1个[祝福]）根据攻击牌系别附加以下[元素箭]效果：\n[火之矢]：本次攻击伤害额外+1。\n[水之矢]：（主动攻击命中时②）目标角色+1[治疗]。\n[风之矢]：（[攻击行动]结束后）+1[攻击行动]。\n[雷之矢]：本次攻击无法应战。\n[地之矢]：（主动攻击命中时②）对目标角色造成1点法术伤害③。",

					RequiredRole:    model.RoleAttacker,
					ResponseType:    model.ResponseOptional,
					LogicHandler:    "elf_elemental_shot",
					TargetType:      model.TargetNone,
					InteractionType: model.InteractionNone,
				},
				{
					ID: "elf_animal_companion", Timings: []model.FlowTiming{model.TimingOnDamageTaken}, Title: "动物伙伴",
					Type:        model.SkillTypeResponse,
					Tags:        []model.SkillTag{model.TagOptional},
					Description: "你的回合内，当目标角色承受你造成的主动攻击伤害后，可摸1弃1。",

					RequiredRole: model.RoleAttacker,
					ResponseType: model.ResponseOptional,
					LogicHandler: "elf_animal_companion",
					TargetType:   model.TargetNone,
				},
				{
					ID: "elf_ritual", Timings: []model.FlowTiming{model.TimingOnTurnStart}, Title: "精灵密仪",
					Type:        model.SkillTypeStartup,
					Tags:        []model.SkillTag{model.TagGem, model.TagUltimate},
					Description: "启动技：消耗1红宝石，进入精灵祝福形态并将牌库顶3张作为祝福。",
					CostGem:     1,

					ResponseType: model.ResponseOptional,
					LogicHandler: "elf_ritual",
					TargetType:   model.TargetNone,
				},
				{
					ID: "elf_pet_empower", Timings: []model.FlowTiming{model.TimingOnDamageTaken}, Title: "宠物强化",
					Type:        model.SkillTypeResponse,
					Tags:        []model.SkillTag{model.TagOptional, model.TagCrystal, model.TagUltimate},
					Description: "触发动物伙伴时可消耗1蓝水晶，将效果改为受伤目标摸1弃1。",
					CostCrystal: 1,

					RequiredRole: model.RoleAttacker,
					ResponseType: model.ResponseOptional,
					LogicHandler: "elf_pet_empower",
					TargetType:   model.TargetEnemy,
					MinTargets:   1,
					MaxTargets:   1,
				},
			},
			ExclusiveCards: []string{},
		},
		// 15. 瘟疫法师
		{
			ID:      "plague_mage",
			Name:    "瘟疫法师",
			Title:   "灾",
			Faction: "灾",
			MaxHand: 6,
			Skills: []model.SkillDefinition{
				{
					ID: "plague_immortal", Timings: []model.FlowTiming{model.TimingOnActionEnd}, Title: "不朽",
					Type:        model.SkillTypeResponse,
					Description: "你的法术行动结束后，+1治疗；若本次法术行动为死亡之触，则不触发。",

					ResponseType: model.ResponseSilent,
					LogicHandler: "plague_immortal",
					TargetType:   model.TargetNone,
				},
				{
					ID: "plague_blasphemy", Timings: []model.FlowTiming{model.TimingActive}, Title: "圣渎",
					Type:         model.SkillTypePassive,
					Description:  "你的治疗不能抵挡攻击伤害，但可抵挡法术伤害；治疗上限初始+3。",
					LogicHandler: "plague_blasphemy",
					TargetType:   model.TargetNone,
				},
				{
					ID: "plague_outbreak", Timings: []model.FlowTiming{model.TimingActive}, Title: "瘟疫",
					Type:           model.SkillTypeAction,
					Description:    "弃1张地系牌，对除自己外所有角色各造成1点法术伤害（逆序结算）；若因此造成士气下降，则回合结束时你+1治疗。",
					CostDiscards:   1,
					DiscardElement: model.ElementEarth,
					LogicHandler:   "plague_outbreak",
					TargetType:     model.TargetNone,
				},
				{
					ID: "plague_death_touch", Timings: []model.FlowTiming{model.TimingActive}, Title: "死亡之触",
					Type:         model.SkillTypeAction,
					Description:  "选择X点治疗与Y张同系牌（X,Y均≥2）并弃置，指定目标造成(X+Y-3)点法术伤害；本次不触发不朽。",
					LogicHandler: "plague_death_touch",
					TargetType:   model.TargetEnemy,
					MinTargets:   1,
					MaxTargets:   1,
				},
				{
					ID: "plague_toxic_nova", Timings: []model.FlowTiming{model.TimingActive}, Title: "剧毒新星",
					Type:         model.SkillTypeAction,
					Tags:         []model.SkillTag{model.TagGem, model.TagUltimate},
					Description:  "消耗1红宝石，对除自己外所有角色造成2点法术伤害（逆序结算），自身+1治疗。",
					CostGem:      1,
					LogicHandler: "plague_toxic_nova",
					TargetType:   model.TargetNone,
				},
			},
			ExclusiveCards: []string{},
		},
		// 16. 魔剑士
		{
			ID:      "magic_swordsman",
			Name:    "魔剑士",
			Title:   "影",
			Faction: "影",
			MaxHand: 6,
			Skills: []model.SkillDefinition{
				{
					ID: "ms_asura_combo", Timings: []model.FlowTiming{model.TimingOnActionEnd}, Title: "修罗连斩",
					Type:        model.SkillTypeResponse,
					Tags:        []model.SkillTag{model.TagTurnLimit, model.TagOptional},
					Description: "攻击行动结束后，可额外+1次火系攻击行动。",

					ResponseType: model.ResponseOptional,
					LogicHandler: "ms_asura_combo",
					TargetType:   model.TargetNone,
				},
				{
					ID: "ms_shadow_gather", Timings: []model.FlowTiming{model.TimingOnTurnStart}, Title: "暗影凝聚",
					Type:        model.SkillTypeStartup,
					Description: "启动阶段可对自己造成1点法术伤害并进入暗影形态，持续至下个己方行动阶段开始前转正。",

					ResponseType: model.ResponseOptional,
					LogicHandler: "ms_shadow_gather",
					TargetType:   model.TargetSelf,
					MinTargets:   1,
					MaxTargets:   1,
				},
				{
					ID: "ms_shadow_power", Timings: []model.FlowTiming{model.TimingActive}, Title: "暗影之力",
					Type:         model.SkillTypePassive,
					Description:  "暗影形态下，你发起的所有攻击伤害额外+1。",
					LogicHandler: "ms_shadow_power",
					TargetType:   model.TargetNone,
				},
				{
					ID: "ms_shadow_reject", Timings: []model.FlowTiming{model.TimingActive}, Title: "暗影抗拒",
					Type:         model.SkillTypePassive,
					Description:  "行动阶段不能使用法术牌；非自己行动阶段可使用魔弹/圣光响应。",
					LogicHandler: "ms_shadow_reject",
					TargetType:   model.TargetNone,
				},
				{
					ID: "ms_shadow_meteor", Timings: []model.FlowTiming{model.TimingActive}, Title: "暗影流星",
					Type:         model.SkillTypeAction,
					Description:  "暗影形态下弃2张法术牌，对1名敌方角色造成2点法术伤害；可额外移除我方战绩区2星石转正并+1红宝石。",
					CostDiscards: 2,
					DiscardType:  model.CardTypeMagic,
					LogicHandler: "ms_shadow_meteor",
					TargetType:   model.TargetEnemy,
					MinTargets:   1,
					MaxTargets:   1,
				},
				{
					ID: "ms_yellow_spring", Timings: []model.FlowTiming{model.TimingOnAttackDeclared}, Title: "黄泉震颤",
					Type:        model.SkillTypeResponse,
					Tags:        []model.SkillTag{model.TagTurnLimit, model.TagGem, model.TagUltimate, model.TagOptional},
					Description: "每回合一次：主动攻击前消耗1红宝石，本次攻击不可应战；命中后手牌补至上限并弃2。",
					CostGem:     1,

					RequiredRole: model.RoleAttacker,
					ResponseType: model.ResponseOptional,
					LogicHandler: "ms_yellow_spring",
					TargetType:   model.TargetNone,
				},
			},
			ExclusiveCards: []string{},
		},
		// 17. 血色剑灵
		{
			ID:      "crimson_sword_spirit",
			Name:    "血色剑灵",
			Title:   "血",
			Faction: "血",
			MaxHand: 6,
			Skills: []model.SkillDefinition{
				{
					ID: "css_blood_thorns", Timings: []model.FlowTiming{model.TimingOnHitCheck}, Title: "血色荆棘",
					Type:        model.SkillTypePassive,
					Description: "攻击命中时自动+1鲜血（上限3）。",

					RequiredRole: model.RoleAttacker,
					ResponseType: model.ResponseSilent,
					LogicHandler: "css_blood_thorns",
					TargetType:   model.TargetNone,
				},
				{
					ID: "css_crimson_flash", Timings: []model.FlowTiming{model.TimingOnActionEnd}, Title: "赤色一闪",
					Type:        model.SkillTypeResponse,
					Tags:        []model.SkillTag{model.TagOptional},
					Description: "攻击行动结束后若有鲜血，可移除1鲜血并对自己造成2点法术伤害，额外+1攻击行动。",

					ResponseType: model.ResponseOptional,
					LogicHandler: "css_crimson_flash",
					TargetType:   model.TargetNone,
				},
				{
					ID: "css_blood_rose", Timings: []model.FlowTiming{model.TimingActive}, Title: "血染蔷薇",
					Type:         model.SkillTypeAction,
					Description:  "移除2点［鲜血］发动，移除目标角色2点［治疗］，将我方阵营的1［水晶］翻面为1［宝石］，再选择任意1名队友+1［治疗］。（若［血蔷薇庭院］在场）额外对所有角色各造成1点法术伤害③。",
					LogicHandler: "css_blood_rose",
					TargetType:   model.TargetAny,
					MinTargets:   2,
					MaxTargets:   2,
				},
				{
					ID: "css_blood_barrier", Timings: []model.FlowTiming{model.TimingOnDamageTaken}, Title: "血气屏障",
					Type:        model.SkillTypeResponse,
					Tags:        []model.SkillTag{model.TagOptional},
					Description: "受到其他角色造成的法术伤害时可移除1鲜血，使本次伤害-1，并对伤害来源造成1点法术伤害。",

					RequiredRole: model.RoleDefender,
					ResponseType: model.ResponseOptional,
					LogicHandler: "css_blood_barrier",
					TargetType:   model.TargetNone,
				},
				{
					ID: "css_rose_courtyard", Timings: []model.FlowTiming{model.TimingActive}, Title: "血蔷薇庭院",
					Type:         model.SkillTypePassive,
					Description:  "专属卡在场时，所有人的治疗均不能用于抵挡伤害；你的回合结束时移回手牌区。",
					LogicHandler: "css_rose_courtyard",
					TargetType:   model.TargetNone,
				},
				{
					ID: "css_dance", Timings: []model.FlowTiming{model.TimingOnTurnStart}, Title: "散华轮舞",
					Type:        model.SkillTypeStartup,
					Tags:        []model.SkillTag{model.TagUltimate},
					Description: "启动阶段二选一：1)耗蓝放置庭院并+2鲜血；2)耗红放置庭院并+2鲜血（上限可达4）且手牌弃至4。",

					ResponseType: model.ResponseOptional,
					LogicHandler: "css_dance",
					TargetType:   model.TargetNone,
				},
			},
			ExclusiveCards: []string{"css_rose_courtyard"},
		},
		// 18. 祈祷师
		{
			ID:      "prayer_master",
			Name:    "祈祷师",
			Title:   "咏",
			Faction: "咏",
			MaxHand: 6,
			Skills: []model.SkillDefinition{
				{
					ID: "prayer_enter_form", Timings: []model.FlowTiming{model.TimingOnTurnStart}, Title: "祈祷",
					Type:        model.SkillTypeStartup,
					Tags:        []model.SkillTag{model.TagGem, model.TagUltimate},
					CostGem:     1,
					Description: "启动阶段消耗1红宝石进入祈祷形态；主动攻击会累计祈祷符文。",

					ResponseType: model.ResponseOptional,
					LogicHandler: "prayer_enter_form",
					TargetType:   model.TargetNone,
				},
				{
					ID: "prayer_rune_gain", Timings: []model.FlowTiming{model.TimingOnAttackDeclared}, Title: "祈祷·攻击增符",
					Type:        model.SkillTypePassive,
					Description: "祈祷形态下，每次主动攻击时+2祈祷符文（上限3）。",

					RequiredRole: model.RoleAttacker,
					ResponseType: model.ResponseSilent,
					LogicHandler: "prayer_rune_gain",
					TargetType:   model.TargetNone,
				},
				{
					ID: "prayer_radiant_faith", Timings: []model.FlowTiming{model.TimingActive}, Title: "光辉信仰",
					Type:         model.SkillTypeAction,
					CostDiscards: 2,
					Description:  "祈祷形态下可发动：移除1祈祷符文，弃2张牌，我方战绩区+1宝石（若未满），并使1名队友+1治疗。",
					LogicHandler: "prayer_radiant_faith",
					TargetType:   model.TargetAlly,
					MinTargets:   1,
					MaxTargets:   1,
				},
				{
					ID: "prayer_dark_curse", Timings: []model.FlowTiming{model.TimingActive}, Title: "黑暗诅咒",
					Type:         model.SkillTypeAction,
					Description:  "祈祷形态下可发动：移除1祈祷符文，对任意1名角色造成2点法术伤害，再对自己造成2点法术伤害。",
					LogicHandler: "prayer_dark_curse",
					TargetType:   model.TargetAny,
					MinTargets:   1,
					MaxTargets:   1,
				},
				{
					ID: "prayer_power_blessing", Timings: []model.FlowTiming{model.TimingActive}, Title: "威力赐福",
					Type:             model.SkillTypeAction,
					Tags:             []model.SkillTag{model.TagUnique},
					Description:      "将独有牌当法术牌打出并放置于1名队友面前；该队友攻击命中后可移除此牌，本次伤害+2。",
					RequireExclusive: true,
					PlaceCard:        true,
					PlaceMode:        model.FieldEffect,
					PlaceEffect:      model.EffectPowerBlessing,
					PlaceHook:        model.FieldHookManual,
					LogicHandler:     "prayer_power_blessing",
					TargetType:       model.TargetAlly,
					MinTargets:       1,
					MaxTargets:       1,
				},
				{
					ID: "prayer_swift_blessing", Timings: []model.FlowTiming{model.TimingActive}, Title: "迅捷赐福",
					Type:             model.SkillTypeAction,
					Tags:             []model.SkillTag{model.TagUnique},
					Description:      "将独有牌当法术牌打出并放置于1名队友面前；该队友攻击/法术行动结束后可移除此牌，额外+1攻击行动。",
					RequireExclusive: true,
					PlaceCard:        true,
					PlaceMode:        model.FieldEffect,
					PlaceEffect:      model.EffectSwiftBlessing,
					PlaceHook:        model.FieldHookManual,
					LogicHandler:     "prayer_swift_blessing",
					TargetType:       model.TargetAlly,
					MinTargets:       1,
					MaxTargets:       1,
				},
				{
					ID: "prayer_mana_tide", Timings: []model.FlowTiming{model.TimingOnActionEnd}, Title: "法力潮汐",
					Type:        model.SkillTypeResponse,
					Tags:        []model.SkillTag{model.TagTurnLimit, model.TagCrystal, model.TagUltimate, model.TagOptional},
					CostCrystal: 1,
					Description: "回合限定：法术行动结束后可消耗1蓝水晶，额外+1法术行动。",

					ResponseType: model.ResponseOptional,
					LogicHandler: "prayer_mana_tide",
					TargetType:   model.TargetNone,
				},
			},
			ExclusiveCards: []string{},
		},
		// 19. 红莲骑士
		{
			ID:      "crimson_knight",
			Name:    "红莲骑士",
			Title:   "血",
			Faction: "血",
			MaxHand: 6,
			Skills: []model.SkillDefinition{
				{
					ID: "crk_crimson_pact", Timings: []model.FlowTiming{model.TimingOnAttackDeclared}, Title: "腥红圣约",
					Type:        model.SkillTypeResponse,
					Tags:        []model.SkillTag{model.TagTurnLimit, model.TagOptional},
					Description: "回合限定：主动攻击时可响应，+1治疗。",

					RequiredRole: model.RoleAttacker,
					ResponseType: model.ResponseOptional,
					LogicHandler: "crk_crimson_pact",
					TargetType:   model.TargetNone,
				},
				{
					ID: "crk_crimson_faith", Timings: []model.FlowTiming{model.TimingActive}, Title: "腥红信仰",
					Type:         model.SkillTypePassive,
					Description:  "你的治疗只能抵御自己对自己造成的伤害；治疗上限初始+2。",
					LogicHandler: "crk_crimson_faith",
					TargetType:   model.TargetNone,
				},
				{
					ID: "crk_bloody_prayer", Timings: []model.FlowTiming{model.TimingOnTurnStart}, Title: "血腥祷言",
					Type:        model.SkillTypeStartup,
					Description: "当你有治疗时可发动：移除你的X点［治疗］，对自己造成X点法术伤害③；选择1~2名队友并任意分配这X点［治疗］，你+1［血印］。",

					ResponseType: model.ResponseOptional,
					LogicHandler: "crk_bloody_prayer",
					TargetType:   model.TargetNone,
				},
				{
					ID: "crk_killing_feast", Timings: []model.FlowTiming{model.TimingOnHitCheck}, Title: "杀戮盛宴",
					Type:        model.SkillTypeResponse,
					Tags:        []model.SkillTag{model.TagOptional},
					Description: "主动攻击命中且有血印时可响应：移除1血印并对自己造成4法术伤害，本次攻击伤害+2。",

					RequiredRole: model.RoleAttacker,
					ResponseType: model.ResponseOptional,
					LogicHandler: "crk_killing_feast",
					TargetType:   model.TargetNone,
				},
				{
					ID: "crk_hot_blood", Timings: []model.FlowTiming{model.TimingActive}, Title: "热血沸腾",
					Type:         model.SkillTypePassive,
					Description:  "因伤害导致我方士气下降时进入热血沸腾形态；该形态下伤害导致的士气下降被免疫。回合结束时脱离并+2治疗。",
					LogicHandler: "crk_hot_blood",
					TargetType:   model.TargetNone,
				},
				{
					ID: "crk_calm_mind", Timings: []model.FlowTiming{model.TimingOnActionEnd}, Title: "戒骄戒躁",
					Type:        model.SkillTypeResponse,
					Tags:        []model.SkillTag{model.TagCrystal, model.TagUltimate, model.TagOptional},
					Description: "热血沸腾形态下，攻击/法术行动结束后可消耗1蓝水晶（红宝石可替代），转正脱离并额外获得1次与刚结束行动同类型的行动。",

					ResponseType: model.ResponseOptional,
					LogicHandler: "crk_calm_mind",
					TargetType:   model.TargetNone,
				},
				{
					ID: "crk_crimson_cross", Timings: []model.FlowTiming{model.TimingActive}, Title: "腥红十字",
					Type:         model.SkillTypeAction,
					Tags:         []model.SkillTag{model.TagCrystal, model.TagUltimate},
					CostCrystal:  1,
					CostDiscards: 2,
					DiscardType:  model.CardTypeMagic,
					Description:  "有蓝水晶与血印且手牌中至少2法术牌时可发动：消耗1蓝水晶与1血印，弃2法术牌，对自己造成4点法术伤害，并对1名敌方角色造成3点法术伤害。",
					LogicHandler: "crk_crimson_cross",
					TargetType:   model.TargetEnemy,
					MinTargets:   1,
					MaxTargets:   1,
				},
			},
			ExclusiveCards: []string{},
		},
		// 20. 英灵人形
		{
			ID:      "war_homunculus",
			Name:    "英灵人形",
			Title:   "幻",
			Faction: "幻",
			MaxHand: 6,
			Skills: []model.SkillDefinition{
				{
					ID: "hom_battle_pattern", Timings: []model.FlowTiming{model.TimingActive}, Title: "战纹掌控",
					Type:         model.SkillTypePassive,
					Description:  "开局获得3战纹（当前实现为战纹/魔纹指示物）。",
					LogicHandler: "hom_battle_pattern",
					TargetType:   model.TargetNone,
				},
				{
					ID: "hom_rage_suppress", Timings: []model.FlowTiming{model.TimingOnHitCheck}, Title: "怒火压制",
					Type:        model.SkillTypeResponse,
					Tags:        []model.SkillTag{model.TagOptional},
					Description: "主动攻击未命中时可响应：翻转1战纹为魔纹。",

					RequiredRole: model.RoleAttacker,
					ResponseType: model.ResponseOptional,
					LogicHandler: "hom_rage_suppress",
					TargetType:   model.TargetNone,
				},
				{
					ID: "hom_rune_smash", Timings: []model.FlowTiming{model.TimingOnHitCheck}, Title: "战纹碎击",
					Type:        model.SkillTypeResponse,
					Tags:        []model.SkillTag{model.TagOptional},
					Description: "主动攻击命中时可响应：翻转1战纹为魔纹，弃X张同系牌，本次攻击伤害额外+(X-1)；若处于蓄势迸发形态，可额外翻转Y个战纹，本次额外法术伤害+Y。",

					RequiredRole: model.RoleAttacker,
					ResponseType: model.ResponseOptional,
					LogicHandler: "hom_rune_smash",
					TargetType:   model.TargetNone,
				},
				{
					ID: "hom_glyph_fusion", Timings: []model.FlowTiming{model.TimingOnHitCheck}, Title: "魔纹融合",
					Type:        model.SkillTypeResponse,
					Tags:        []model.SkillTag{model.TagOptional},
					Description: "主动攻击未命中时可响应：翻转1魔纹为战纹，弃X张异系牌（X>1），对本次攻击目标造成(X-1)点法术伤害；若处于蓄势迸发形态，可额外翻转Y个魔纹，本次法术伤害额外+Y。",

					RequiredRole: model.RoleAttacker,
					ResponseType: model.ResponseOptional,
					LogicHandler: "hom_glyph_fusion",
					TargetType:   model.TargetNone,
				},
				{
					ID: "hom_rune_reforge", Timings: []model.FlowTiming{model.TimingOnTurnStart}, Title: "符文改造",
					Type:        model.SkillTypeStartup,
					Tags:        []model.SkillTag{model.TagGem, model.TagUltimate},
					CostGem:     1,
					Description: "启动阶段可消耗1红宝石进入蓄势迸发形态，手牌上限+1并强制摸1张牌；可重新分配战纹/魔纹（总数保持3）；回合结束时转正脱离该形态。",

					ResponseType: model.ResponseOptional,
					LogicHandler: "hom_rune_reforge",
					TargetType:   model.TargetNone,
				},
				{
					ID: "hom_dual_echo", Timings: []model.FlowTiming{model.TimingOnDamageTaken}, Title: "双重回响",
					Type:        model.SkillTypeResponse,
					Tags:        []model.SkillTag{model.TagTurnLimit, model.TagCrystal, model.TagUltimate, model.TagOptional},
					CostCrystal: 1,
					Description: "回合限定：对目标角色造成攻击或法术伤害时可消耗1蓝水晶，对另一名角色造成等量（最多3）法术伤害；该伤害不会造成士气下降。",

					ResponseType: model.ResponseOptional,
					LogicHandler: "hom_dual_echo",
					TargetType:   model.TargetAny,
					MinTargets:   1,
					MaxTargets:   1,
				},
			},
			ExclusiveCards: []string{},
		},
		// 21. 神官
		{
			ID:      "priest",
			Name:    "神官",
			Title:   "圣",
			Faction: "圣",
			MaxHand: 6,
			Skills: []model.SkillDefinition{
				{
					ID: "priest_divine_revelation", Timings: []model.FlowTiming{model.TimingOnActionEnd}, Title: "神圣启示",
					Type:        model.SkillTypePassive,
					Description: "特殊行动结束时触发，+1治疗。",

					ResponseType: model.ResponseSilent,
					LogicHandler: "priest_divine_revelation",
					TargetType:   model.TargetNone,
				},
				{
					ID: "priest_divine_bless", Timings: []model.FlowTiming{model.TimingActive}, Title: "神圣祈福",
					Type:         model.SkillTypeAction,
					CostDiscards: 2,
					DiscardType:  model.CardTypeMagic,
					Description:  "弃2张法术牌，自己+2治疗。",
					LogicHandler: "priest_divine_bless",
					TargetType:   model.TargetNone,
				},
				{
					ID: "priest_water_power", Timings: []model.FlowTiming{model.TimingActive}, Title: "水之神力",
					Type:         model.SkillTypeAction,
					CostDiscards: 2,
					Description:  "弃1张水系牌，并将另一张手牌交给1名队友；你与该队友各+1治疗。",
					LogicHandler: "priest_water_power",
					TargetType:   model.TargetAlly,
					MinTargets:   1,
					MaxTargets:   1,
				},
				{
					ID: "priest_guardian", Timings: []model.FlowTiming{model.TimingActive}, Title: "圣使守护",
					Type:         model.SkillTypePassive,
					Description:  "治疗上限+4；每次抵御伤害时最多使用1点治疗。",
					LogicHandler: "priest_guardian",
					TargetType:   model.TargetNone,
				},
				{
					ID: "priest_divine_contract", Timings: []model.FlowTiming{model.TimingOnTurnStart}, Title: "神圣契约",
					Type:        model.SkillTypeStartup,
					Tags:        []model.SkillTag{model.TagCrystal, model.TagUltimate},
					CostCrystal: 1,
					Description: "消耗1蓝水晶，将自身治疗转移给1名队友（目标治疗上限按4封顶）。",

					LogicHandler: "priest_divine_contract",
					TargetType:   model.TargetAlly,
					MinTargets:   1,
					MaxTargets:   1,
				},
				{
					ID: "priest_divine_domain", Timings: []model.FlowTiming{model.TimingActive}, Title: "神圣领域",
					Type:         model.SkillTypeAction,
					Tags:         []model.SkillTag{model.TagCrystal, model.TagUltimate},
					CostCrystal:  1,
					CostDiscards: 2,
					Description:  "［水晶］弃2张牌，再选择以下1项：①移除你1点治疗并对1名其他角色造成2点法术伤害；②你+2治疗并令1名其他队友+1治疗。",
					LogicHandler: "priest_divine_domain",
					TargetType:   model.TargetNone,
				},
			},
			ExclusiveCards: []string{},
		},
		// 22. 阴阳师
		{
			ID:      "onmyoji",
			Name:    "阴阳师",
			Title:   "咏",
			Faction: "咏",
			MaxHand: 6,
			Skills: []model.SkillDefinition{
				{
					ID: "onmyoji_shikigami_descend", Timings: []model.FlowTiming{model.TimingActive}, Title: "式神降临",
					Type:         model.SkillTypeAction,
					CostDiscards: 2,
					Description:  "［持续］（弃2张命格相同的手牌［展示］）［横置］转为［式神形态］，你+1［鬼火］，额外+1［攻击行动］。",
					LogicHandler: "onmyoji_shikigami_descend",
					TargetType:   model.TargetNone,
				},
				{
					ID: "onmyoji_yinyang_shift", Timings: []model.FlowTiming{model.TimingOnAttackDeclared}, Title: "阴阳转换",
					Type:        model.SkillTypeResponse,
					Description: "你应战时可展示1张与来袭攻击同命格的攻击牌，视为你应战此次攻击并将其系别转为该牌系别；你+1［鬼火］。若处于式神形态则转正脱离，本次攻击伤害=X（X为你的鬼火数）。",

					ResponseType: model.ResponseSilent,
					LogicHandler: "onmyoji_yinyang_shift",
					TargetType:   model.TargetNone,
				},
				{
					ID: "onmyoji_shikigami_shift", Timings: []model.FlowTiming{model.TimingOnAttackDeclared}, Title: "式神转换",
					Type:        model.SkillTypeResponse,
					Description: "当阴阳转换生效时自动触发：你强制摸1张牌，然后+1［鬼火］。",

					ResponseType: model.ResponseSilent,
					LogicHandler: "onmyoji_shikigami_shift",
					TargetType:   model.TargetNone,
				},
				{
					ID: "onmyoji_dark_ritual", Timings: []model.FlowTiming{model.TimingOnActionEnd}, Title: "黑暗祭礼",
					Type:        model.SkillTypePassive,
					Description: "回合结束时若鬼火达上限，强制发动：选择1名敌方角色，移除全部鬼火并对其造成2点法术伤害。",

					ResponseType: model.ResponseSilent,
					LogicHandler: "onmyoji_dark_ritual",
					TargetType:   model.TargetEnemy,
					MinTargets:   1,
					MaxTargets:   1,
				},
				{
					ID: "onmyoji_binding", Timings: []model.FlowTiming{model.TimingOnAttackDeclared}, Title: "式神咒束",
					Type:        model.SkillTypeResponse,
					Description: "（目标队友受到主动攻击时①，若此攻击可应战且你处于［式神形态］，打出1张合理的应战攻击牌［展示］，移除我方［战绩区］1［宝石］1［水晶］）将本次攻击目标变更为你，且视为你使用此牌执行应战攻击。",

					ResponseType: model.ResponseSilent,
					LogicHandler: "onmyoji_binding",
					TargetType:   model.TargetNone,
				},
				{
					ID: "onmyoji_life_barrier", Timings: []model.FlowTiming{model.TimingActive}, Title: "生命结界",
					Type:         model.SkillTypeAction,
					Tags:         []model.SkillTag{model.TagUltimate, model.TagCrystal},
					CostCrystal:  1,
					Description:  "消耗1蓝水晶并+1鬼火后选择：①队友+1宝石+1治疗，自己受X点法伤（X为鬼火，X=3时该伤害不导致士气下降）；②若处于式神形态，弃2张同命格手牌并脱离式神形态，令1名队友弃1张手牌。",
					LogicHandler: "onmyoji_life_barrier",
					TargetType:   model.TargetNone,
					MinTargets:   0,
					MaxTargets:   0,
				},
			},
			ExclusiveCards: []string{},
		},
		// 23. 苍炎魔女
		{
			ID:      "blaze_witch",
			Name:    "苍炎魔女",
			Title:   "血",
			Faction: "血",
			MaxHand: 6,
			Skills: []model.SkillDefinition{
				{
					ID: "bw_rebirth_clock", Timings: []model.FlowTiming{model.TimingBeforeMoraleLoss}, Title: "永生银时计",
					Type:        model.SkillTypePassive,
					Description: "［重生］上限4；当你因承受法术伤害导致士气下降时，你+1［重生］。",

					ResponseType: model.ResponseSilent,
					LogicHandler: "bw_rebirth_clock",
					TargetType:   model.TargetNone,
				},
				{
					ID: "bw_blazing_codex", Timings: []model.FlowTiming{model.TimingActive}, Title: "苍炎法典",
					Type:           model.SkillTypeAction,
					CostDiscards:   1,
					DiscardElement: model.ElementFire,
					Description:    "弃1张火系牌［展示］，对目标角色和自己各造成2点法术伤害（目标先结算）。",
					LogicHandler:   "bw_blazing_codex",
					TargetType:     model.TargetAny,
					MinTargets:     1,
					MaxTargets:     1,
				},
				{
					ID: "bw_heavenfire_cleave", Timings: []model.FlowTiming{model.TimingActive}, Title: "天火断空",
					Type:           model.SkillTypeAction,
					CostDiscards:   2,
					DiscardElement: model.ElementFire,
					Description:    "弃2张火系牌并移除1点［重生］（烈焰形态下免移除），对目标角色和自己各造成3点法术伤害；若我方士气落后于目标方，本次伤害额外+1。",
					LogicHandler:   "bw_heavenfire_cleave",
					TargetType:     model.TargetAny,
					MinTargets:     1,
					MaxTargets:     1,
				},
				{
					ID: "bw_witch_wrath", Timings: []model.FlowTiming{model.TimingOnTurnStart}, Title: "魔女之怒",
					Type:        model.SkillTypeStartup,
					Description: "手牌<4时可发动：［横置］进入烈焰形态并选择摸0~2张牌；持续到下个行动阶段开始前。烈焰形态下：非水/暗攻击牌视为火系；发动天火断空无需消耗重生；手牌上限+(重生-2)。到时转正脱离。",

					ResponseType: model.ResponseOptional,
					LogicHandler: "bw_witch_wrath",
					TargetType:   model.TargetNone,
				},
				{
					ID: "bw_substitute_doll", Timings: []model.FlowTiming{model.TimingOnDamageTaken}, Title: "替身玩偶",
					Type:         model.SkillTypeResponse,
					CostDiscards: 1,
					DiscardType:  model.CardTypeMagic,
					Description:  "任何人对你造成攻击伤害时可响应：弃1张法术牌［展示］，令1名队友摸1张牌。",

					RequiredRole: model.RoleDefender,
					ResponseType: model.ResponseOptional,
					LogicHandler: "bw_substitute_doll",
					TargetType:   model.TargetAlly,
					MinTargets:   1,
					MaxTargets:   1,
				},
				{
					ID: "bw_pain_link", Timings: []model.FlowTiming{model.TimingActive}, Title: "痛苦链接",
					Type:         model.SkillTypeAction,
					Tags:         []model.SkillTag{model.TagCrystal, model.TagUltimate},
					CostCrystal:  1,
					Description:  "［水晶］对目标对手和自己各造成1点法术伤害，然后你弃到3张手牌。",
					LogicHandler: "bw_pain_link",
					TargetType:   model.TargetEnemy,
					MinTargets:   1,
					MaxTargets:   1,
				},
				{
					ID: "bw_mana_inversion", Timings: []model.FlowTiming{model.TimingOnDamageTaken}, Title: "魔能反转",
					Type:        model.SkillTypeResponse,
					Tags:        []model.SkillTag{model.TagCrystal, model.TagUltimate},
					CostCrystal: 1,
					Description: "［水晶］任何人对你造成法术伤害时可响应：弃X张法术牌［展示］（X>1），对目标对手造成(X-1)点法术伤害。",

					RequiredRole: model.RoleDefender,
					ResponseType: model.ResponseOptional,
					LogicHandler: "bw_mana_inversion",
					TargetType:   model.TargetEnemy,
					MinTargets:   1,
					MaxTargets:   1,
				},
			},
			ExclusiveCards: []string{},
		},
		// 24. 贤者
		{
			ID:      "sage",
			Name:    "贤者",
			Title:   "咏",
			Faction: "咏",
			MaxHand: 6,
			Skills: []model.SkillDefinition{
				{
					ID: "sage_wisdom_codex", Timings: []model.FlowTiming{model.TimingActive}, Title: "智慧法典",
					Type:         model.SkillTypePassive,
					Description:  "你的能量上限+1；你每次承受法术伤害时，若该伤害>3：你+2红宝石并弃1张牌。",
					LogicHandler: "sage_wisdom_codex",
					TargetType:   model.TargetNone,
				},
				{
					ID: "sage_magic_rebound", Timings: []model.FlowTiming{model.TimingActive}, Title: "法术反弹",
					Type:         model.SkillTypeResponse,
					Description:  "你每次承受法术伤害时，若该伤害仅为1点：可弃X张同系牌（X>1），对目标角色造成(X-1)点法术伤害，并对自己造成X点法术伤害。",
					LogicHandler: "sage_magic_rebound",
					TargetType:   model.TargetAny,
					MinTargets:   1,
					MaxTargets:   1,
				},
				{
					ID: "sage_arcane_codex", Timings: []model.FlowTiming{model.TimingActive}, Title: "魔道法典",
					Type:         model.SkillTypeAction,
					Tags:         []model.SkillTag{model.TagGem, model.TagUltimate},
					CostGem:      1,
					Description:  "［宝石］弃X张异系牌（X>1），对目标角色与自己各造成(X-1)点法术伤害。",
					LogicHandler: "sage_arcane_codex",
					TargetType:   model.TargetAny,
					MinTargets:   1,
					MaxTargets:   1,
				},
				{
					ID: "sage_holy_codex", Timings: []model.FlowTiming{model.TimingActive}, Title: "圣洁法典",
					Type:         model.SkillTypeAction,
					Tags:         []model.SkillTag{model.TagGem, model.TagUltimate},
					CostGem:      1,
					Description:  "［宝石］弃X张异系牌（X>2），最多(X-2)名角色各+2治疗，然后对自己造成(X-1)点法术伤害。",
					LogicHandler: "sage_holy_codex",
					TargetType:   model.TargetAny,
					MinTargets:   1,
					MaxTargets:   6,
				},
			},
			ExclusiveCards: []string{},
		},
		// 25. 魔弓
		{
			ID:      "magic_bow",
			Name:    "魔弓",
			Title:   "幻",
			Faction: "幻",
			MaxHand: 6,
			Skills: []model.SkillDefinition{
				{
					ID: "mb_magic_pierce", Timings: []model.FlowTiming{model.TimingOnAttackDeclared}, Title: "魔贯冲击",
					Type:        model.SkillTypeResponse,
					Description: "主动攻击前可发动：移除1个火系充能，本次攻击伤害+1；若命中可再移除1个火系充能使伤害再+1；若未命中则对目标造成3点法术伤害。本回合与多重射击互斥。",

					RequiredRole: model.RoleAttacker,
					ResponseType: model.ResponseOptional,
					LogicHandler: "mb_magic_pierce",
					TargetType:   model.TargetNone,
				},
				{
					ID: "mb_thunder_scatter", Timings: []model.FlowTiming{model.TimingActive}, Title: "雷光散射",
					Type:         model.SkillTypeAction,
					Description:  "移除1个雷系充能：对所有对手各造成1点法术伤害；可额外移除X个雷系充能并指定1名对手，本次对其伤害额外+X。",
					LogicHandler: "mb_thunder_scatter",
					TargetType:   model.TargetEnemy,
					MinTargets:   0,
					MaxTargets:   1,
				},
				{
					ID: "mb_multi_shot", Timings: []model.FlowTiming{model.TimingOnActionEnd}, Title: "多重射击",
					Type:        model.SkillTypeResponse,
					Description: "攻击行动结束时可发动：移除1个风系充能，视为1次暗系主动攻击（不能攻击上次目标，且本次伤害-1）。本回合与魔贯冲击互斥。",

					RequiredRole: model.RoleAttacker,
					ResponseType: model.ResponseOptional,
					LogicHandler: "mb_multi_shot",
					TargetType:   model.TargetNone,
				},
				{
					ID: "mb_charge", Timings: []model.FlowTiming{model.TimingOnTurnStart}, Title: "充能",
					Type:        model.SkillTypeStartup,
					Tags:        []model.SkillTag{model.TagCrystal, model.TagUltimate},
					CostCrystal: 1,
					Description: "［水晶］弃到4张牌后摸X张牌（X<5），可将最多X张手牌作为充能盖牌（上限8）。本回合不能发动魔贯冲击与雷光散射。",

					ResponseType: model.ResponseOptional,
					LogicHandler: "mb_charge",
					TargetType:   model.TargetNone,
				},
				{
					ID: "mb_demon_eye", Timings: []model.FlowTiming{model.TimingOnTurnStart}, Title: "魔眼",
					Type:        model.SkillTypeStartup,
					Tags:        []model.SkillTag{model.TagGem, model.TagUltimate},
					CostGem:     1,
					Description: "［宝石］选择一项：令1名角色弃1张牌；或你摸3张牌。然后将1张手牌作为充能，并获得1蓝水晶。",

					ResponseType: model.ResponseOptional,
					LogicHandler: "mb_demon_eye",
					TargetType:   model.TargetNone,
					MinTargets:   0,
					MaxTargets:   0,
				},
			},
			ExclusiveCards: []string{},
		},
		// 26. 魔枪
		{
			ID:      "magic_lancer",
			Name:    "魔枪",
			Title:   "幻",
			Faction: "幻",
			MaxHand: 6,
			Skills: []model.SkillDefinition{
				{
					ID: "ml_dark_release", Timings: []model.FlowTiming{model.TimingOnTurnStart}, Title: "暗之解放",
					Type:        model.SkillTypeStartup,
					Description: "［横置］转为幻影形态，手牌上限恒定为5；本回合下一次主动攻击伤害+1，且本回合不能发动漆黑之枪与充盈。",

					ResponseType: model.ResponseOptional,
					LogicHandler: "ml_dark_release",
					TargetType:   model.TargetNone,
				},
				{
					ID: "ml_phantom_stardust", Timings: []model.FlowTiming{model.TimingOnTurnStart}, Title: "幻影星尘",
					Type:        model.SkillTypeStartup,
					Description: "仅幻影形态可发动：先对自己造成2点法术伤害并完全结算，随后转正脱离幻影形态；若未因此导致我方士气下降，则对目标角色造成2点法术伤害。",

					ResponseType: model.ResponseOptional,
					LogicHandler: "ml_phantom_stardust",
					TargetType:   model.TargetEnemy,
					MinTargets:   1,
					MaxTargets:   1,
				},
				{
					ID: "ml_dark_bind", Timings: []model.FlowTiming{model.TimingActive}, Title: "黑暗束缚",
					Type:         model.SkillTypePassive,
					Description:  "你始终不能使用法术牌。",
					LogicHandler: "ml_dark_bind",
					TargetType:   model.TargetNone,
				},
				{
					ID: "ml_dark_barrier", Timings: []model.FlowTiming{model.TimingOnDamageTaken}, Title: "暗之障壁",
					Type:        model.SkillTypeResponse,
					Description: "任何人对你造成伤害时可发动：弃X张法术牌或X张雷系牌（同次发动不可混弃）。",

					RequiredRole: model.RoleDefender,
					ResponseType: model.ResponseOptional,
					LogicHandler: "ml_dark_barrier",
					TargetType:   model.TargetNone,
				},
				{
					ID: "ml_fullness", Timings: []model.FlowTiming{model.TimingActive}, Title: "充盈",
					Type:         model.SkillTypeAction,
					Description:  "弃1张法术牌或雷系牌：全体角色按逆时针各弃1张牌（我方可选择不弃）；除你外每有1名角色以此法弃置法术牌或雷系牌，你本回合下次主动攻击伤害+1；额外+1次攻击行动。",
					LogicHandler: "ml_fullness",
					TargetType:   model.TargetNone,
				},
				{
					ID: "ml_black_spear", Timings: []model.FlowTiming{model.TimingOnHitCheck}, Title: "漆黑之枪",
					Type:        model.SkillTypeResponse,
					Tags:        []model.SkillTag{model.TagCrystal, model.TagUltimate},
					Description: "X［水晶］（仅幻影形态下，主动攻击手牌为1或2的对手并命中后）本次攻击伤害额外+（X+2）。",

					RequiredRole: model.RoleAttacker,
					ResponseType: model.ResponseOptional,
					LogicHandler: "ml_black_spear",
					TargetType:   model.TargetNone,
				},
			},
			ExclusiveCards: []string{},
		},
		// 27. 灵符师
		{
			ID:      "spirit_caster",
			Name:    "灵符师",
			Title:   "咏",
			Faction: "咏",
			MaxHand: 6,
			Skills: []model.SkillDefinition{
				{
					ID: "sc_talisman_thunder", Timings: []model.FlowTiming{model.TimingActive}, Title: "灵符-雷鸣",
					Type:           model.SkillTypeAction,
					CostDiscards:   1,
					DiscardElement: model.ElementThunder,
					Description:    "弃1张雷系牌［展示］，对任意2名角色各造成1点法术伤害。若触发封印，按“封印伤害→念咒→技能效果”顺序结算。",
					LogicHandler:   "sc_talisman_thunder",
					TargetType:     model.TargetAny,
					MinTargets:     2,
					MaxTargets:     2,
				},
				{
					ID: "sc_talisman_wind", Timings: []model.FlowTiming{model.TimingActive}, Title: "灵符-风行",
					Type:           model.SkillTypeAction,
					CostDiscards:   1,
					DiscardElement: model.ElementWind,
					Description:    "弃1张风系牌［展示］，指定2名角色各弃1张牌。若触发封印，按“封印伤害→念咒→技能效果”顺序结算。",
					LogicHandler:   "sc_talisman_wind",
					TargetType:     model.TargetAny,
					MinTargets:     2,
					MaxTargets:     2,
				},
				{
					ID: "sc_incantation", Timings: []model.FlowTiming{model.TimingActive}, Title: "念咒",
					Type:         model.SkillTypeResponse,
					Description:  "你每次发动灵符时，可将1张手牌面朝下放置在角色旁，作为［妖力］。",
					LogicHandler: "sc_incantation",
					TargetType:   model.TargetNone,
				},
				{
					ID: "sc_hundred_night", Timings: []model.FlowTiming{model.TimingOnHitCheck}, Title: "百鬼夜行",
					Type:        model.SkillTypeResponse,
					Description: "主动攻击命中后可发动：移除1个妖力。默认对1名角色造成1点法术伤害；若移除的是火系妖力，可展示并改为指定2名角色，对其余所有角色各造成1点法术伤害。",

					RequiredRole: model.RoleAttacker,
					ResponseType: model.ResponseOptional,
					LogicHandler: "sc_hundred_night",
					TargetType:   model.TargetAny,
					MinTargets:   1,
					MaxTargets:   2,
				},
				{
					ID: "sc_spiritual_collapse", Timings: []model.FlowTiming{model.TimingActive}, Title: "灵力崩解",
					Type:         model.SkillTypeResponse,
					Tags:         []model.SkillTag{model.TagCrystal, model.TagUltimate},
					CostCrystal:  1,
					Description:  "［水晶］可与【灵符-雷鸣】或【百鬼夜行】同发动，使本次每段伤害额外+1。",
					LogicHandler: "sc_spiritual_collapse",
					TargetType:   model.TargetNone,
				},
			},
			ExclusiveCards: []string{},
		},
		// 28. 吟游诗人
		{
			ID:      "bard",
			Name:    "吟游诗人",
			Title:   "幻",
			Faction: "幻",
			MaxHand: 6,
			Skills: []model.SkillDefinition{
				{
					ID: "bd_descent_concerto", Timings: []model.FlowTiming{model.TimingActive}, Title: "沉沦协奏曲",
					Type:         model.SkillTypeResponse,
					Tags:         []model.SkillTag{model.TagTurnLimit},
					Description:  "［回合限定］仅普通形态：本回合你对至少2名不同对手造成法术伤害并结算完后，强制弃2张同系牌并公开展示；你+1灵感。若弃牌中含法术牌，则再对1名目标对手造成1点法术伤害。",
					LogicHandler: "bd_descent_concerto",
					TargetType:   model.TargetEnemy,
					MinTargets:   1,
					MaxTargets:   1,
				},
				{
					ID: "bd_dissonance_chord", Timings: []model.FlowTiming{model.TimingActive}, Title: "不谐和弦",
					Type:         model.SkillTypeAction,
					Description:  "移除X点灵感（X>1）；若处于永恒囚徒形态则转正脱离。然后选择一项：你与目标角色各摸(X-1)张牌；或你与目标角色各弃(X-1)张牌。",
					LogicHandler: "bd_dissonance_chord",
					TargetType:   model.TargetAny,
					MinTargets:   1,
					MaxTargets:   1,
				},
				{
					ID: "bd_forbidden_verse", Timings: []model.FlowTiming{model.TimingActive}, Title: "禁忌诗篇",
					Type:         model.SkillTypePassive,
					Description:  "激昂狂想曲或胜利交响诗结算后：若灵感未满则你+1灵感并移除永恒乐章；若灵感已满则对自己造成3点法术伤害，且普通形态下转为永恒囚徒形态。",
					LogicHandler: "bd_forbidden_verse",
					TargetType:   model.TargetNone,
				},
				{
					ID: "bd_rousing_rhapsody", Timings: []model.FlowTiming{model.TimingActive}, Title: "激昂狂想曲",
					Type:         model.SkillTypeResponse,
					Description:  "回合开始时，若我方存在永恒乐章：选择一项——吟游诗人对2名目标对手各造成1点法术伤害；或你弃2张牌。",
					LogicHandler: "bd_rousing_rhapsody",
					TargetType:   model.TargetEnemy,
					MinTargets:   0,
					MaxTargets:   2,
				},
				{
					ID: "bd_victory_symphony", Timings: []model.FlowTiming{model.TimingActive}, Title: "胜利交响诗",
					Type:         model.SkillTypeResponse,
					Description:  "回合结束时，若我方存在永恒乐章：选择一项——将我方战绩区1个指定星石提炼为你的能量；或我方战绩区+1宝石且你+1治疗。",
					LogicHandler: "bd_victory_symphony",
					TargetType:   model.TargetNone,
				},
				{
					ID: "bd_hope_fugue", Timings: []model.FlowTiming{model.TimingOnTurnStart}, Title: "希望赋格曲",
					Type:        model.SkillTypeStartup,
					Tags:        []model.SkillTag{model.TagCrystal, model.TagUltimate},
					CostCrystal: 1,
					Description: "［水晶］你可以先摸1张牌；然后选择：将永恒乐章放置于目标队友面前；或将永恒乐章转移给我方另一名目标角色，你弃1张牌并获得+1治疗或+1灵感。",

					ResponseType: model.ResponseOptional,
					LogicHandler: "bd_hope_fugue",
					TargetType:   model.TargetAlly,
					MinTargets:   1,
					MaxTargets:   1,
				},
			},
			ExclusiveCards: []string{"bd_eternal_movement"},
		},
		// 29. 勇者
		{
			ID:      "hero",
			Name:    "勇者",
			Title:   "血",
			Faction: "血",
			MaxHand: 6,
			Skills: []model.SkillDefinition{
				{
					ID: "hero_heart", Timings: []model.FlowTiming{model.TimingActive}, Title: "勇者之心",
					Type:         model.SkillTypePassive,
					Description:  "游戏初始时，你+2［水晶］。",
					LogicHandler: "hero_heart",
					TargetType:   model.TargetNone,
				},
				{
					ID: "hero_roar", Timings: []model.FlowTiming{model.TimingOnAttackDeclared}, Title: "怒吼",
					Type:        model.SkillTypeResponse,
					Description: "主动攻击前可发动：移除1点［怒气］，你摸0~1张牌，本次攻击伤害额外+2；若未命中，你+1［知性］。",

					RequiredRole: model.RoleAttacker,
					ResponseType: model.ResponseOptional,
					LogicHandler: "hero_roar",
					TargetType:   model.TargetNone,
				},
				{
					ID: "hero_forbidden_power", Timings: []model.FlowTiming{model.TimingOnHitCheck}, Title: "禁断之力",
					Type:        model.SkillTypeResponse,
					Tags:        []model.SkillTag{model.TagCrystal, model.TagUltimate},
					CostCrystal: 1,
					Description: "［水晶］主动攻击命中或未命中后可发动：展示并弃掉所有手牌；每有1张法术牌你+1怒气；未命中时每有1张水系牌你+1知性；命中时每有1张火系牌本次攻击伤害额外+1，并对自己造成等同火系牌数量的法术伤害。",

					RequiredRole: model.RoleAttacker,
					ResponseType: model.ResponseOptional,
					LogicHandler: "hero_forbidden_power",
					TargetType:   model.TargetNone,
				},
				{
					ID: "hero_exhaustion", Timings: []model.FlowTiming{model.TimingActive}, Title: "精疲力竭",
					Type:         model.SkillTypePassive,
					Description:  "发动禁断之力后强制触发：［横置］额外+1攻击行动；持续到你的下个行动阶段开始，手牌上限恒定为4。效果结束时转正并对自己造成3点法术伤害。",
					LogicHandler: "hero_exhaustion",
					TargetType:   model.TargetNone,
				},
				{
					ID: "hero_calm_mind", Timings: []model.FlowTiming{model.TimingOnAttackDeclared}, Title: "明镜止水",
					Type:        model.SkillTypeResponse,
					Description: "主动攻击前可发动：移除4点［知性］，本次攻击对手无法应战；本次攻击结束时你+1［水晶］。",

					RequiredRole: model.RoleAttacker,
					ResponseType: model.ResponseOptional,
					LogicHandler: "hero_calm_mind",
					TargetType:   model.TargetNone,
				},
				{
					ID: "hero_taunt", Timings: []model.FlowTiming{model.TimingActive}, Title: "挑衅",
					Type:             model.SkillTypeAction,
					Tags:             []model.SkillTag{model.TagUnique},
					Description:      "移除1点［怒气］：将【挑衅】放置于目标对手面前，你+1［知性］；该对手在其下个行动阶段必须且只能主动攻击你，否则跳过该行动阶段。触发后移除此牌。",
					RequireExclusive: true,
					PlaceCard:        true,
					PlaceMode:        model.FieldEffect,
					PlaceEffect:      model.EffectHeroTaunt,
					PlaceHook:        model.FieldHookManual,
					LogicHandler:     "hero_taunt",
					TargetType:       model.TargetEnemy,
					MinTargets:       1,
					MaxTargets:       1,
				},
				{
					ID: "hero_dead_duel", Timings: []model.FlowTiming{model.TimingOnDamageTaken}, Title: "死斗",
					Type:        model.SkillTypeResponse,
					Tags:        []model.SkillTag{model.TagGem, model.TagUltimate},
					CostGem:     1,
					Description: "［宝石］每当你承受法术伤害时可发动：你+3［怒气］；若此伤害造成士气实际下降，本次士气下降值恒定为1。",

					RequiredRole: model.RoleDefender,
					ResponseType: model.ResponseOptional,
					LogicHandler: "hero_dead_duel",
					TargetType:   model.TargetNone,
				},
			},
			ExclusiveCards: []string{"hero_taunt"},
		},
		// 30. 格斗家
		{
			ID:      "fighter",
			Name:    "格斗家",
			Title:   "技",
			Faction: "技",
			MaxHand: 6,
			Skills: []model.SkillDefinition{
				{
					ID: "fighter_psi_field", Timings: []model.FlowTiming{model.TimingOnDamageTaken}, Title: "念气力场",
					Type:        model.SkillTypePassive,
					Description: "所有对你造成的伤害每次最高为4点。",

					RequiredRole: model.RoleDefender,
					ResponseType: model.ResponseSilent,
					LogicHandler: "fighter_psi_field",
					TargetType:   model.TargetNone,
				},
				{
					ID: "fighter_charge_strike", Timings: []model.FlowTiming{model.TimingOnAttackDeclared}, Title: "蓄力一击",
					Type:        model.SkillTypeResponse,
					Description: "主动攻击前可发动（斗气未满）：+1斗气，本次攻击伤害额外+1；若未命中，对自己造成X点法术伤害（X为当前斗气）。",

					RequiredRole: model.RoleAttacker,
					ResponseType: model.ResponseOptional,
					LogicHandler: "fighter_charge_strike",
					TargetType:   model.TargetNone,
				},
				{
					ID: "fighter_psi_bullet", Timings: []model.FlowTiming{model.TimingOnActionEnd}, Title: "念弹",
					Type:        model.SkillTypeResponse,
					Description: "法术行动结束时可发动（斗气未满）：+1斗气，对1名目标对手造成1点法术伤害；若其治疗为0，则你再承受X点法术伤害（X为当前斗气）。",

					RequiredRole: model.RoleAttacker,
					ResponseType: model.ResponseOptional,
					LogicHandler: "fighter_psi_bullet",
					TargetType:   model.TargetEnemy,
					MinTargets:   1,
					MaxTargets:   1,
				},
				{
					ID: "fighter_hundred_dragon", Timings: []model.FlowTiming{model.TimingOnTurnStart}, Title: "百式幻龙拳",
					Type:        model.SkillTypeStartup,
					Description: "持续：移除3斗气并横置。主动攻击伤害+2、应战攻击伤害+1；主动攻击需锁定同一目标，且不能发动蓄力一击。若改为执行法术行动或特殊行动，或主动攻击更换目标，则立即退出该状态。",

					ResponseType: model.ResponseOptional,
					LogicHandler: "fighter_hundred_dragon",
					TargetType:   model.TargetEnemy,
					MinTargets:   1,
					MaxTargets:   1,
				},
				{
					ID: "fighter_burst_crash", Timings: []model.FlowTiming{model.TimingOnAttackDeclared}, Title: "气绝崩击",
					Type:        model.SkillTypeResponse,
					Description: "主动攻击前可发动：移除1斗气，本次攻击无法应战；然后对自己造成X点法术伤害（X为当前斗气）。不能与蓄力一击同时发动。",

					RequiredRole: model.RoleAttacker,
					ResponseType: model.ResponseOptional,
					LogicHandler: "fighter_burst_crash",
					TargetType:   model.TargetNone,
				},
				{
					ID: "fighter_war_god_drive", Timings: []model.FlowTiming{model.TimingOnTurnStart}, Title: "斗神天驱",
					Type:        model.SkillTypeStartup,
					Tags:        []model.SkillTag{model.TagCrystal, model.TagUltimate},
					CostCrystal: 1,
					Description: "［水晶］你弃到3张牌，然后+2治疗。",

					ResponseType: model.ResponseOptional,
					LogicHandler: "fighter_war_god_drive",
					TargetType:   model.TargetNone,
				},
			},
			ExclusiveCards: []string{},
		},
		// 31. 圣弓
		{
			ID:      "holy_bow",
			Name:    "圣弓",
			Title:   "圣",
			Faction: "圣",
			MaxHand: 6,
			Skills: []model.SkillDefinition{
				{
					ID: "hb_heavenly_bow", Timings: []model.FlowTiming{model.TimingActive}, Title: "天之弓",
					Type:         model.SkillTypePassive,
					Description:  "初始+1圣煌辉光炮、+2水晶、治疗上限+1；主动攻击若非圣命格伤害-1；主动攻击命中且为圣命格时+1信仰。",
					LogicHandler: "hb_heavenly_bow",
					TargetType:   model.TargetNone,
				},
				{
					ID: "hb_holy_shard_storm", Timings: []model.FlowTiming{model.TimingActive}, Title: "圣屑飓暴",
					Type:         model.SkillTypeAction,
					Description:  "弃2张同系攻击牌并展示，视为1次圣命格同系主动攻击；若未命中，可移除最多2点治疗，并令1名队友弃置等量手牌。",
					LogicHandler: "hb_holy_shard_storm",
					TargetType:   model.TargetNone,
				},
				{
					ID: "hb_radiant_descent", Timings: []model.FlowTiming{model.TimingActive}, Title: "圣煌降临",
					Type:         model.SkillTypeAction,
					Description:  "移除2点治疗或2点信仰，进入持续的圣煌形态，并额外+1法术行动；圣煌形态下执行特殊行动会退场并+1治疗。",
					LogicHandler: "hb_radiant_descent",
					TargetType:   model.TargetNone,
				},
				{
					ID: "hb_light_burst", Timings: []model.FlowTiming{model.TimingActive}, Title: "圣光爆裂",
					Type:         model.SkillTypeAction,
					Description:  "仅圣煌形态可发动：①摸1，移除1治疗，+1信仰，1名其他我方角色+1治疗；②移除X治疗并弃X牌，至多选择X名手牌数不大于你手牌数-X的对手，各受(Y+2)点攻击伤害。",
					LogicHandler: "hb_light_burst",
					TargetType:   model.TargetNone,
				},
				{
					ID: "hb_meteor_bullet", Timings: []model.FlowTiming{model.TimingOnAttackDeclared}, Title: "流星圣弹",
					Type:        model.SkillTypeResponse,
					Description: "仅圣煌形态下，主动攻击前可发动：移除1点治疗或1点信仰，令1名我方角色+1治疗。",

					RequiredRole: model.RoleAttacker,
					ResponseType: model.ResponseOptional,
					LogicHandler: "hb_meteor_bullet",
					TargetType:   model.TargetAlly,
					MinTargets:   1,
					MaxTargets:   1,
				},
				{
					ID: "hb_radiant_cannon", Timings: []model.FlowTiming{model.TimingActive}, Title: "圣煌辉光炮",
					Type:         model.SkillTypeAction,
					Description:  "仅圣煌形态可发动：移除1圣煌辉光炮与(4+士气落后值)信仰，所有角色手牌调整为4，我方星杯+1，然后选择将一方士气调整为与另一方相同。",
					LogicHandler: "hb_radiant_cannon",
					TargetType:   model.TargetNone,
				},
				{
					ID: "hb_auto_fill", Timings: []model.FlowTiming{model.TimingActive}, Title: "自动填充",
					Type:         model.SkillTypeResponse,
					Tags:         []model.SkillTag{model.TagUltimate},
					Description:  "回合结束时，若你未执行特殊行动：可选择①消耗1水晶，+1信仰或+1治疗；②消耗1宝石，+1水晶并+2信仰或+2治疗。",
					ResponseType: model.ResponseOptional,
					LogicHandler: "hb_auto_fill",
					TargetType:   model.TargetNone,
				},
			},
			ExclusiveCards: []string{},
		},
		// 32. 剑帝
		{
			ID:      "sword_emperor",
			Name:    "剑帝",
			Title:   "技",
			Faction: "技",
			MaxHand: 6,
			Skills: []model.SkillDefinition{
				{
					ID: "se_sword_soul_guard", Timings: []model.FlowTiming{model.TimingOnHitCheck}, Title: "剑魂守护",
					Type:        model.SkillTypePassive,
					Description: "主动攻击未命中时：若剑魂未达上限，则将本次打出的攻击牌作为剑魂置于角色旁；当前能量为单数时剑魂视为天使之魂，双数时视为恶魔之魂，无能量时不属于任何一种。",

					ResponseType: model.ResponseSilent,
					LogicHandler: "se_sword_soul_guard",
					TargetType:   model.TargetNone,
				},
				{
					ID: "se_feint", Timings: []model.FlowTiming{model.TimingOnHitCheck}, Title: "佯攻",
					Type:        model.SkillTypePassive,
					Description: "主动攻击未命中时，你+1剑气。",

					ResponseType: model.ResponseSilent,
					LogicHandler: "se_feint",
					TargetType:   model.TargetNone,
				},
				{
					ID: "se_sword_qi_slash", Timings: []model.FlowTiming{model.TimingOnHitCheck}, Title: "剑气斩",
					Type:        model.SkillTypeResponse,
					Description: "主动攻击命中后可发动：移除X点剑气（X最高为3），对除当前攻击目标外的任意1名角色造成X点法术伤害。",

					RequiredRole: model.RoleAttacker,
					ResponseType: model.ResponseOptional,
					LogicHandler: "se_sword_qi_slash",
					TargetType:   model.TargetAny,
					MinTargets:   1,
					MaxTargets:   1,
				},
				{
					ID: "se_angel_soul", Timings: []model.FlowTiming{model.TimingOnAttackDeclared}, Title: "天使之魂",
					Type:        model.SkillTypeResponse,
					Description: "主动攻击前可发动：当前能量为单数且至少有1张剑魂时，移除1张天使之魂；本次攻击若命中则你+2治疗，若未命中则我方士气+1；不能与剑魂守护同时发动。",

					RequiredRole: model.RoleAttacker,
					ResponseType: model.ResponseOptional,
					LogicHandler: "se_angel_soul",
					TargetType:   model.TargetNone,
				},
				{
					ID: "se_demon_soul", Timings: []model.FlowTiming{model.TimingOnAttackDeclared}, Title: "恶魔之魂",
					Type:        model.SkillTypeResponse,
					Description: "主动攻击前可发动：当前能量为双数且至少有1张剑魂时，移除1张恶魔之魂；本次攻击伤害额外+1，若未命中则你+2剑气；不能与剑魂守护同时发动。",

					RequiredRole: model.RoleAttacker,
					ResponseType: model.ResponseOptional,
					LogicHandler: "se_demon_soul",
					TargetType:   model.TargetNone,
				},
				{
					ID: "se_angel_soul_hit", Timings: []model.FlowTiming{model.TimingOnHitCheck}, Title: "天使之魂·命中结算",
					Type:        model.SkillTypePassive,
					Description: "若本次攻击由天使之魂挂载且命中，则你+2治疗。",

					ResponseType: model.ResponseSilent,
					LogicHandler: "se_angel_soul_hit",
					TargetType:   model.TargetNone,
				},
				{
					ID: "se_angel_soul_miss", Timings: []model.FlowTiming{model.TimingOnHitCheck}, Title: "天使之魂·未命中结算",
					Type:        model.SkillTypePassive,
					Description: "若本次攻击由天使之魂挂载且未命中，则我方士气+1。",

					ResponseType: model.ResponseSilent,
					LogicHandler: "se_angel_soul_miss",
					TargetType:   model.TargetNone,
				},
				{
					ID: "se_demon_soul_miss", Timings: []model.FlowTiming{model.TimingOnHitCheck}, Title: "恶魔之魂·未命中结算",
					Type:        model.SkillTypePassive,
					Description: "若本次攻击由恶魔之魂挂载且未命中，则你+2剑气。",

					ResponseType: model.ResponseSilent,
					LogicHandler: "se_demon_soul_miss",
					TargetType:   model.TargetNone,
				},
				{
					ID: "se_indomitable_will", Timings: []model.FlowTiming{model.TimingOnActionEnd}, Title: "不屈意志",
					Type:        model.SkillTypeResponse,
					Tags:        []model.SkillTag{model.TagCrystal, model.TagUltimate},
					CostCrystal: 1,
					Description: "［水晶］攻击行动结束时可发动：摸1张牌、+1剑气，并额外+1攻击行动。",

					ResponseType: model.ResponseOptional,
					LogicHandler: "se_indomitable_will",
					TargetType:   model.TargetNone,
				},
			},
			ExclusiveCards: []string{},
		},
		// 33. 兽灵武士
		{
			ID:      "beast_samurai",
			Name:    "兽灵武士",
			Title:   "技",
			Faction: "技",
			MaxHand: 6,
			Skills: []model.SkillDefinition{
				{
					ID: "bs_warrior_zanshin", Timings: []model.FlowTiming{model.TimingOnActionEnd}, Title: "武者残心",
					Type:        model.SkillTypeResponse,
					Tags:        []model.SkillTag{model.TagTurnLimit},
					Description: "［回合限定］攻击行动结束时，你+1残心。",

					ResponseType: model.ResponseSilent,
					LogicHandler: "bs_warrior_zanshin",
					TargetType:   model.TargetNone,
				},
				{
					ID: "bs_one_strike_no_thought", Timings: []model.FlowTiming{model.TimingOnActionEnd}, Title: "一击无念",
					Type:        model.SkillTypeResponse,
					Description: "攻击行动结束后可发动：移除4点残心，额外+1攻击行动，并挂载下次主动攻击的无视圣盾/无视圣光/技命格强制命中效果。",

					ResponseType: model.ResponseOptional,
					LogicHandler: "bs_one_strike_no_thought",
					TargetType:   model.TargetNone,
				},
				{
					ID: "bs_one_strike_intercept", Timings: []model.FlowTiming{model.TimingOnAttackDeclared}, Title: "一击无念·下次攻击劫持",
					Type:        model.SkillTypePassive,
					Description: "当一击无念已挂载时，你的下一次主动攻击无视圣盾且无法用圣光抵挡；若攻击牌为技命格，则强制命中。",

					ResponseType: model.ResponseSilent,
					LogicHandler: "bs_one_strike_intercept",
					TargetType:   model.TargetNone,
				},
				{
					ID: "bs_beast_soul_will", Timings: []model.FlowTiming{model.TimingOnHitCheck}, Title: "兽魂意念",
					Type:        model.SkillTypePassive,
					Description: "每移除1点兽魂，你+1残心；仅普通形态下，主动攻击命中时你+1兽魂。",

					RequiredRole: model.RoleAttacker,
					ResponseType: model.ResponseSilent,
					LogicHandler: "bs_beast_soul_will",
					TargetType:   model.TargetNone,
				},
				{
					ID: "bs_beast_soul_alert", Timings: []model.FlowTiming{model.TimingOnOrientationChanged}, Title: "兽魂警戒",
					Type:        model.SkillTypeResponse,
					Description: "其他角色横置效果结算完成后可发动：移除1点兽魂并进入御魂流居合形态，令该角色展示并弃置1张牌；若其弃的是法术牌，则你+1兽魂。",

					ResponseType: model.ResponseOptional,
					LogicHandler: "bs_beast_soul_alert",
					TargetType:   model.TargetNone,
				},
				{
					ID: "bs_beast_return", Timings: []model.FlowTiming{model.TimingOnDamageTaken}, Title: "兽返",
					Type:        model.SkillTypeResponse,
					Description: "当其他角色对你造成法术伤害时可发动：移除X点兽魂，你弃X张牌，他弃1张牌；若其弃的是法术牌，则你+1兽魂。",

					RequiredRole: model.RoleDefender,
					ResponseType: model.ResponseOptional,
					LogicHandler: "bs_beast_return",
					TargetType:   model.TargetNone,
				},
				{
					ID: "bs_iaijutsu_turn_end_drain", Timings: []model.FlowTiming{model.TimingOnActionEnd}, Title: "御魂流居合形态·回合结束扣魂",
					Type:        model.SkillTypePassive,
					Description: "处于御魂流居合形态时，你的回合结束前-1兽魂，并同步+1残心。",

					ResponseType: model.ResponseSilent,
					LogicHandler: "bs_iaijutsu_turn_end_drain",
					TargetType:   model.TargetNone,
				},
				{
					ID: "bs_iaijutsu_exit_on_deal_damage", Timings: []model.FlowTiming{model.TimingOnDamageTaken}, Title: "御魂流居合形态·造成伤害退场",
					Type:        model.SkillTypePassive,
					Description: "处于御魂流居合形态时，只要你造成过伤害，立即转正并脱离该形态。",

					RequiredRole: model.RoleAttacker,
					ResponseType: model.ResponseSilent,
					LogicHandler: "bs_iaijutsu_exit_on_deal_damage",
					TargetType:   model.TargetNone,
				},
				{
					ID: "bs_iaijutsu_exit_on_zero", Timings: []model.FlowTiming{model.TimingOnActionEnd}, Title: "御魂流居合形态·兽魂归零退场",
					Type:        model.SkillTypePassive,
					Description: "你的回合结束时，若仍处于御魂流居合形态且兽魂为0，则转正并脱离该形态。",

					ResponseType: model.ResponseSilent,
					LogicHandler: "bs_iaijutsu_exit_on_zero",
					TargetType:   model.TargetNone,
				},
				{
					ID: "bs_iaijutsu_tapped_target_boost", Timings: []model.FlowTiming{model.TimingOnDamageCalculated}, Title: "御魂流居合形态·横置目标增伤",
					Type:        model.SkillTypePassive,
					Description: "处于御魂流居合形态时，你对横置目标角色的主动攻击伤害+1。",

					RequiredRole: model.RoleAttacker,
					ResponseType: model.ResponseSilent,
					LogicHandler: "bs_iaijutsu_tapped_target_boost",
					TargetType:   model.TargetNone,
				},
				{
					ID: "bs_reversal_iaijutsu", Timings: []model.FlowTiming{model.TimingOnHitCheck}, Title: "逆反居合斩",
					Type:        model.SkillTypeResponse,
					Description: "仅御魂流居合形态下，主动攻击命中手牌<4的对手时可发动：移除X点兽魂，本次攻击改为目标弃置(X+2)张手牌；若实际弃牌数小于X+2，则对方士气-1。",

					RequiredRole: model.RoleAttacker,
					ResponseType: model.ResponseOptional,
					LogicHandler: "bs_reversal_iaijutsu",
					TargetType:   model.TargetNone,
				},
				{
					ID: "bs_iaijutsu_style", Timings: []model.FlowTiming{model.TimingActive}, Title: "御魂流居合式",
					Type:         model.SkillTypeStartup,
					Tags:         []model.SkillTag{model.TagGem, model.TagUltimate},
					CostGem:      1,
					Description:  "［宝石］无视兽魂上限+1兽魂，并选择摸1或弃1；若已处于御魂流居合形态，则+1残心；若处于普通形态，则横置进入御魂流居合形态。",
					LogicHandler: "bs_iaijutsu_style",
					TargetType:   model.TargetNone,
				},
			},
			ExclusiveCards: []string{},
		},
		// 34. 灵魂术士
		{
			ID:      "soul_sorcerer",
			Name:    "灵魂术士",
			Title:   "幻",
			Faction: "幻",
			MaxHand: 6,
			Skills: []model.SkillDefinition{
				{
					ID: "ss_soul_devour", Timings: []model.FlowTiming{model.TimingBeforeMoraleLoss}, Title: "灵魂吞噬",
					Type:        model.SkillTypePassive,
					Description: "（我方角色因承受伤害导致士气每下降1点）你+1［黄色灵魂］（上限6）。",

					ResponseType: model.ResponseSilent,
					LogicHandler: "ss_soul_devour",
					TargetType:   model.TargetNone,
				},
				{
					ID: "ss_soul_recall", Timings: []model.FlowTiming{model.TimingActive}, Title: "灵魂召还",
					Type:         model.SkillTypeAction,
					Description:  "弃X张法术牌［展示］，你+X点［蓝色灵魂］（上限6）。",
					LogicHandler: "ss_soul_recall",
					TargetType:   model.TargetNone,
				},
				{
					ID: "ss_soul_convert", Timings: []model.FlowTiming{model.TimingOnAttackDeclared}, Title: "灵魂转换",
					Type:        model.SkillTypeResponse,
					Description: "（你每发动1次主动攻击①）可转换1点你拥有的［灵魂］颜色。",

					RequiredRole: model.RoleAttacker,
					ResponseType: model.ResponseOptional,
					LogicHandler: "ss_soul_convert",
					TargetType:   model.TargetNone,
				},
				{
					ID: "ss_soul_mirror", Timings: []model.FlowTiming{model.TimingActive}, Title: "灵魂镜像",
					Type:         model.SkillTypeAction,
					CostDiscards: 2,
					Description:  "（移除2点［黄色灵魂］）你弃2张牌，目标角色摸2张牌［强制］，但最多补到其手牌上限。",
					LogicHandler: "ss_soul_mirror",
					TargetType:   model.TargetAny,
					MinTargets:   1,
					MaxTargets:   1,
				},
				{
					ID: "ss_soul_blast", Timings: []model.FlowTiming{model.TimingActive}, Title: "灵魂震爆",
					Type:             model.SkillTypeAction,
					Tags:             []model.SkillTag{model.TagUnique},
					CostDiscards:     1,
					RequireExclusive: true,
					Description:      "（移除3点［黄色灵魂］）对目标角色造成3点法术伤害；若其手牌<3且手牌上限>5，本次伤害额外+2。",
					LogicHandler:     "ss_soul_blast",
					TargetType:       model.TargetAny,
					MinTargets:       1,
					MaxTargets:       1,
				},
				{
					ID: "ss_soul_grant", Timings: []model.FlowTiming{model.TimingActive}, Title: "灵魂赐予",
					Type:             model.SkillTypeAction,
					Tags:             []model.SkillTag{model.TagUnique},
					CostDiscards:     1,
					RequireExclusive: true,
					Description:      "（移除3点［蓝色灵魂］）目标角色+2［宝石］。",
					LogicHandler:     "ss_soul_grant",
					TargetType:       model.TargetAny,
					MinTargets:       1,
					MaxTargets:       1,
				},
				{
					ID: "ss_soul_link", Timings: []model.FlowTiming{model.TimingOnTurnStart}, Title: "灵魂链接",
					Type:             model.SkillTypeStartup,
					Tags:             []model.SkillTag{model.TagExclusive},
					RequireExclusive: true,
					Description:      "（仅你队友数>1时可发动，移除1黄魂+1蓝魂）将灵魂链接放置于目标队友面前；你或其承受伤害前可移除X蓝魂，将X点伤害转移给另一方（转移伤害为法术伤害）。",

					ResponseType: model.ResponseOptional,
					LogicHandler: "ss_soul_link",
					TargetType:   model.TargetNone,
				},
				{
					ID: "ss_soul_amp", Timings: []model.FlowTiming{model.TimingOnTurnStart}, Title: "灵魂增幅",
					Type:        model.SkillTypeStartup,
					Tags:        []model.SkillTag{model.TagGem, model.TagUltimate},
					CostGem:     1,
					Description: "［宝石］你+2［黄色灵魂］和+2［蓝色灵魂］。",

					ResponseType: model.ResponseOptional,
					LogicHandler: "ss_soul_amp",
					TargetType:   model.TargetNone,
				},
			},
			ExclusiveCards: []string{"soul_link"},
		},
		// 34. 月之女神
		{
			ID:      "moon_goddess",
			Name:    "月之女神",
			Title:   "圣",
			Faction: "圣",
			MaxHand: 6,
			Skills: []model.SkillDefinition{
				{
					ID: "mg_new_moon_shelter", Timings: []model.FlowTiming{model.TimingBeforeMoraleLoss}, Title: "新月庇护",
					Type:        model.SkillTypeResponse,
					Description: "（我方角色因承受伤害导致爆牌并将士气下降时）转为暗月形态，并将本次爆牌改为你的暗月；本次士气不下降。",

					ResponseType: model.ResponseOptional,
					LogicHandler: "mg_new_moon_shelter",
					TargetType:   model.TargetNone,
				},
				{
					ID: "mg_dark_moon_curse", Timings: []model.FlowTiming{model.TimingActive}, Title: "闇月诅咒",
					Type:         model.SkillTypePassive,
					Description:  "你每次移除1个闇月，我方士气-1；当闇月数为0时，脱离闇月形态。",
					LogicHandler: "mg_dark_moon_curse",
					TargetType:   model.TargetNone,
				},
				{
					ID: "mg_medusa_eye", Timings: []model.FlowTiming{model.TimingOnAttackDeclared}, Title: "美杜莎之眼",
					Type:         model.SkillTypeResponse,
					Description:  "目标对手攻击时，可移除1个同系闇月：你+1治疗、+1石化；若移除的是法术牌，再弃1张牌并对该目标对手造成1点法术伤害。",
					LogicHandler: "mg_medusa_eye",
					TargetType:   model.TargetNone,
				},
				{
					ID: "mg_moon_cycle", Timings: []model.FlowTiming{model.TimingActive}, Title: "月之轮回",
					Type:         model.SkillTypeResponse,
					Description:  "你的回合结束时，选择其一：①移除1闇月，令目标角色+1治疗；②移除1治疗，你+1新月。",
					LogicHandler: "mg_moon_cycle",
					TargetType:   model.TargetNone,
				},
				{
					ID: "mg_blasphemy", Timings: []model.FlowTiming{model.TimingActive}, Title: "月渎",
					Type:         model.SkillTypeResponse,
					Tags:         []model.SkillTag{model.TagTurnLimit},
					Description:  "［回合限定］目标对手承受你造成的法术伤害后，可移除1治疗，对该目标对手造成1点法术伤害。",
					LogicHandler: "mg_blasphemy",
					TargetType:   model.TargetNone,
				},
				{
					ID: "mg_darkmoon_slash", Timings: []model.FlowTiming{model.TimingOnHitCheck}, Title: "闇月斩",
					Type:        model.SkillTypeResponse,
					Tags:        []model.SkillTag{model.TagCrystal, model.TagUltimate},
					CostCrystal: 1,
					Description: "仅闇月形态下，主动攻击命中时可发动：移除X个闇月（1<=X<3），本次攻击伤害额外+X。",

					RequiredRole: model.RoleAttacker,
					ResponseType: model.ResponseOptional,
					LogicHandler: "mg_darkmoon_slash",
					TargetType:   model.TargetNone,
				},
				{
					ID: "mg_pale_moon", Timings: []model.FlowTiming{model.TimingActive}, Title: "苍白之月",
					Type:         model.SkillTypeAction,
					Tags:         []model.SkillTag{model.TagGem, model.TagUltimate},
					CostGem:      1,
					Description:  "［宝石］选择其一：①移除3石化，下次主动攻击不可应战、额外+1攻击行动，并额外获得一个回合。②移除X新月、你+1石化、弃1张牌，对目标对手造成(X+1)点法术伤害（X>=1）。",
					LogicHandler: "mg_pale_moon",
					TargetType:   model.TargetNone,
				},
			},
			ExclusiveCards: []string{},
		},
		// 35. 血之巫女
		{
			ID:      "blood_priestess",
			Name:    "血之巫女",
			Title:   "血",
			Faction: "血",
			MaxHand: 6,
			Skills: []model.SkillDefinition{
				{
					ID: "bp_blood_sorrow", Timings: []model.FlowTiming{model.TimingOnTurnStart}, Title: "血之哀伤",
					Type: model.SkillTypeStartup,

					ResponseType: model.ResponseOptional,
					Description:  "（对自己造成2点法术伤害）选择：转移【同生共死】目标，或移除【同生共死】。",
					LogicHandler: "bp_blood_sorrow",
					TargetType:   model.TargetNone,
				},
				{
					ID: "bp_bleeding", Timings: []model.FlowTiming{model.TimingActive}, Title: "流血",
					Type:         model.SkillTypePassive,
					Description:  "普通形态下因承伤导致我方士气下降时，强制进入流血形态并+1治疗；流血形态下回合开始对自己造成1点法术伤害；当手牌<3时强制脱离流血形态。",
					LogicHandler: "bp_bleeding",
					TargetType:   model.TargetNone,
				},
				{
					ID: "bp_backflow", Timings: []model.FlowTiming{model.TimingActive}, Title: "逆流",
					Type:         model.SkillTypeAction,
					CostDiscards: 2,
					Description:  "仅流血形态下可发动：你弃2张牌，你+1治疗。",
					LogicHandler: "bp_backflow",
					TargetType:   model.TargetNone,
				},
				{
					ID: "bp_blood_wail", Timings: []model.FlowTiming{model.TimingActive}, Title: "血之悲鸣",
					Type:             model.SkillTypeAction,
					Tags:             []model.SkillTag{model.TagUnique},
					RequireExclusive: true,
					Description:      "仅流血形态下可发动：选择X（X<3），对目标角色和自己各造成（X+1）点法术伤害（先目标后自己）。",
					LogicHandler:     "bp_blood_wail",
					TargetType:       model.TargetAny,
					MinTargets:       1,
					MaxTargets:       1,
				},
				{
					ID: "bp_shared_life", Timings: []model.FlowTiming{model.TimingActive}, Title: "同生共死",
					Type:         model.SkillTypeAction,
					Tags:         []model.SkillTag{model.TagExclusive},
					Description:  "你摸2张牌（强制），将【同生共死】放置于目标角色面前：普通形态下你和其手牌上限各-2；流血形态下你和其手牌上限各+1。",
					LogicHandler: "bp_shared_life",
					TargetType:   model.TargetNone,
				},
				{
					ID: "bp_blood_curse", Timings: []model.FlowTiming{model.TimingActive}, Title: "血之诅咒",
					Type:         model.SkillTypeAction,
					Tags:         []model.SkillTag{model.TagGem, model.TagUltimate},
					CostGem:      1,
					Description:  "［宝石］对目标角色造成2点法术伤害，然后你弃3张牌（手牌不足则全弃）。",
					LogicHandler: "bp_blood_curse",
					TargetType:   model.TargetAny,
					MinTargets:   1,
					MaxTargets:   1,
				},
			},
			ExclusiveCards: []string{"shared_life"},
		},
		// 37. 蝶舞者
		{
			ID:      "butterfly_dancer",
			Name:    "蝶舞者",
			Title:   "咏",
			Faction: "咏",
			MaxHand: 6,
			Skills: []model.SkillDefinition{
				{
					ID: "bt_life_fire", Timings: []model.FlowTiming{model.TimingActive}, Title: "生命之火",
					Type:         model.SkillTypePassive,
					Description:  "你的手牌上限-X（X为蛹数量），但最低为3。",
					LogicHandler: "bt_life_fire",
					TargetType:   model.TargetNone,
				},
				{
					ID: "bt_dance", Timings: []model.FlowTiming{model.TimingActive}, Title: "舞动",
					Type:         model.SkillTypeAction,
					Description:  "选择：摸1张牌（强制）或弃1张牌（强制）；然后将牌库顶1张牌面朝下放置为茧。",
					LogicHandler: "bt_dance",
					TargetType:   model.TargetNone,
				},
				{
					ID: "bt_poison_powder", Timings: []model.FlowTiming{model.TimingActive}, Title: "毒粉",
					Type:         model.SkillTypeResponse,
					Description:  "每当有角色产生1点实际法术伤害时，可移除1个茧，使该次伤害额外+1。",
					LogicHandler: "bt_poison_powder",
					TargetType:   model.TargetNone,
				},
				{
					ID: "bt_pilgrimage", Timings: []model.FlowTiming{model.TimingActive}, Title: "朝圣",
					Type:         model.SkillTypeResponse,
					Description:  "每当你承受伤害时，可移除1个茧，抵御1点该来源伤害（每次伤害最多发动1次）。",
					LogicHandler: "bt_pilgrimage",
					TargetType:   model.TargetNone,
				},
				{
					ID: "bt_mirror", Timings: []model.FlowTiming{model.TimingActive}, Title: "镜花水月",
					Type:         model.SkillTypeResponse,
					Description:  "每当有角色产生2点实际法术伤害时，可移除2张同系茧并展示：抵御该次伤害，改为你对伤害来源造成2次1点法术伤害。",
					LogicHandler: "bt_mirror",
					TargetType:   model.TargetNone,
				},
				{
					ID: "bt_wither", Timings: []model.FlowTiming{model.TimingActive}, Title: "凋零",
					Type:         model.SkillTypeResponse,
					Description:  "你每次移除茧时，若该茧为法术牌，可展示并发动：对目标造成1点法术伤害，再对自己造成2点法术伤害；直到你下回合开始前，对方士气最低为1。",
					LogicHandler: "bt_wither",
					TargetType:   model.TargetNone,
				},
				{
					ID: "bt_chrysalis", Timings: []model.FlowTiming{model.TimingActive}, Title: "蛹化",
					Type:         model.SkillTypeAction,
					Tags:         []model.SkillTag{model.TagGem, model.TagUltimate},
					CostGem:      1,
					Description:  "［宝石］你+1蛹，并将牌库顶4张牌面朝下放置为茧。",
					LogicHandler: "bt_chrysalis",
					TargetType:   model.TargetNone,
				},
				{
					ID: "bt_reverse_butterfly", Timings: []model.FlowTiming{model.TimingActive}, Title: "倒逆之蝶",
					Type:         model.SkillTypeAction,
					Tags:         []model.SkillTag{model.TagCrystal, model.TagUltimate},
					CostCrystal:  1,
					CostDiscards: 2,
					Description:  "［水晶］你弃2张牌，再选择以下1项发动：①对目标造成1点不可用治疗抵御的法术伤害；②移除2个茧或对自己造成4点法术伤害，然后移除1个蛹。",
					LogicHandler: "bt_reverse_butterfly",
					TargetType:   model.TargetNone,
				},
			},
			ExclusiveCards: []string{},
		},
	}
	return normalizeSkillTimings(characters)
}
