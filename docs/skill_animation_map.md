# 技能动画模板映射

> 对应 `web/src/constants/skillAnimationMap.ts`。覆盖 37 个已实现角色的全部 219 个技能（`internal/data/characters.go`）。
> 开场：主动 = 金色法阵 + 技能横匾；触发 = 座位旁小胶囊。效果按顺序播放，模板为 打击/治疗/状态/资源/形态/全局。
> 自动推断：是 = skill_activated 带目标且效果有对应事件；部分 = 有技能事件但缺目标/效果，或效果在后续选择中结算；否 = 无技能事件，只有状态差值。
> 依据是对各角色 handler/choices/timing hook 的代码阅读，不是实测；标“未验证”的条目需对局确认。
> 招牌（signature）只挑 12 个盘面级大招；TagUltimate 覆盖 51 个技能（几乎所有宝石/水晶技），不能直接当筛选条件。

## 汇总

| 维度 | 取值 | 技能数 |
|---|---|---|
| 开场 | 触发 | 119 |
| 开场 | 主动 | 100 |
| 效果（含该模板） | 打击 | 91 |
| 效果（含该模板） | 治疗 | 41 |
| 效果（含该模板） | 状态 | 46 |
| 效果（含该模板） | 资源 | 105 |
| 效果（含该模板） | 形态 | 104 |
| 效果（含该模板） | 全局 | 26 |
| 自动推断 | 是 | 57 |
| 自动推断 | 部分 | 118 |
| 自动推断 | 否 | 44 |
| 招牌 | signature | 12 |

## 天使（angel）

| 技能 | ID | 开场 | 效果 | 元素 | 自动推断 | 后端缺口 |
|---|---|---|---|---|---|---|
| 天使羁绊 | `angel_bond` | 触发 | 治疗 |  | 部分 | 治疗目标在后续选择中结算，需带目标的生效事件；注：执行只推选择，治疗在 choices.go |
| 天使祝福 | `angel_blessing` | 主动 | 资源 | 水 | 部分 | 缺给牌事件（来源→去向、张数） |
| 风之洁净 | `angel_cleanse` | 主动 | 状态 | 风 | 部分 | 需带被移除效果的持有者与类型 |
| 天使之歌 | `angel_song` | 触发 | 状态 |  | 部分 | 需带被移除效果的持有者与类型；注：水晶在 Execute 内扣 |
| 神之庇护 | `god_protection` | 触发 | 全局 |  | 部分 | 需带 X 与抵御阵营的生效事件；注：X 水晶在 choices.go 扣，skill_activated 先于 X 选择 |
| 天使之墙 | `angel_wall` | 主动 | 状态 |  | 是 |  |

## 狂战士（berserker）

| 技能 | ID | 开场 | 效果 | 元素 | 自动推断 | 后端缺口 |
|---|---|---|---|---|---|---|
| 狂化 | `berserker_frenzy` | 触发 | 打击 |  | 部分 | damage_dealt 补加伤来源 |
| 撕裂 | `berserker_tear` | 触发 | 打击 |  | 部分 | damage_dealt 补加伤来源 |
| 血腥咆哮 | `blood_roar` | 触发 | 打击 |  | 部分 | 补强制命中标记；注：timing_hooks.go hitCheckHook 未注册（死代码），handler 未设无视圣盾 |
| 血影狂刀 | `blood_blade` | 触发 | 打击 |  | 部分 | damage_dealt 补加伤来源 |

## 封印师（sealer）

| 技能 | ID | 开场 | 效果 | 元素 | 自动推断 | 后端缺口 |
|---|---|---|---|---|---|---|
| 法术激荡 | `magic_surge` | 触发 | 资源 |  | 是 |  |
| 封印破碎 | `seal_break` | 主动 | 状态 → 资源 |  | 部分 | 选择结算补被收回效果的 target |
| 五系束缚 | `five_elements_bind` | 主动 | 状态 → 资源 |  | 部分 | 行动前触发时补 skill_activated |
| 水之封印 | `water_seal` | 主动 | 状态 → 打击 | 水 | 部分 | 触发时 skill_activated 归属封印师并带 target；注：触发事件记在被封印者名下 |
| 火之封印 | `fire_seal` | 主动 | 状态 → 打击 | 火 | 部分 | 触发时 skill_activated 归属封印师并带 target；注：同水之封印 |
| 地之封印 | `earth_seal` | 主动 | 状态 → 打击 | 地 | 部分 | 触发时 skill_activated 归属封印师并带 target；注：同水之封印 |
| 风之封印 | `wind_seal` | 主动 | 状态 → 打击 | 风 | 部分 | 触发时 skill_activated 归属封印师并带 target；注：同水之封印 |
| 雷之封印 | `thunder_seal` | 主动 | 状态 → 打击 | 雷 | 部分 | 触发时 skill_activated 归属封印师并带 target；注：同水之封印 |

## 风之剑圣（blade_master）

| 技能 | ID | 开场 | 效果 | 元素 | 自动推断 | 后端缺口 |
|---|---|---|---|---|---|---|
| 风怒追击 | `wind_fury` | 触发 | 资源 | 风 | 是 | 注：CanUse 另要求手中有风系攻击牌，描述未写 |
| 圣剑 | `holy_sword` | 触发 | 状态 → 资源 |  | 部分 | 强制命中标记；摸X弃X 补技能事件；注：摸弃部分在 PostActionEnd hook |
| 剑影 | `sword_shadow` | 触发 | 资源 |  | 是 | 注：水晶在 handler 内扣 |
| 疾风技 | `gale_skill` | 触发 | 资源 | 风 | 部分 | 额外行动无可视效果 |
| 列风技 | `gale_slash` | 触发 | 状态 | 风 | 部分 | 补无视圣盾/无法应战标记 |

## 神箭手（archer）

| 技能 | ID | 开场 | 效果 | 元素 | 自动推断 | 后端缺口 |
|---|---|---|---|---|---|---|
| 贯穿射击 | `piercing_shot` | 触发 | 打击 |  | 部分 | 弃牌路径补 skill_activated；注：InteractionDiscard 路径不发技能事件 |
| 闪电箭 | `lightning_arrow` | 触发 | 打击 | 雷 | 部分 | 补无法应战标记 |
| 狙击 | `snipe` | 主动 | 资源 → 形态 |  | 是 |  |
| 精准射击 | `precise_shot` | 触发 | 打击 |  | 部分 | 补强制命中标记与 -1 伤害来源 |
| 闪光陷阱 | `flash_trap` | 主动 | 打击 |  | 是 | 注：伤害类型用字面量"法术"而非常量，未验证是否按法术伤害处理 |

## 暗杀者（assassin）

| 技能 | ID | 开场 | 效果 | 元素 | 自动推断 | 后端缺口 |
|---|---|---|---|---|---|---|
| 反噬 | `backlash` | 触发 | 资源 |  | 部分 | skill_activated 补 target_ids |
| 水影 | `water_shadow` | 触发 | 资源 | 水 | 部分 | 弃牌路径补 skill_activated；注：InteractionDiscard 路径不发技能事件 |
| 潜行 | `stealth` | 主动 | 形态 → 资源 |  | 部分 | 潜行加伤/解除补事件；注：技能事件在选择推送时发出 |

## 圣女（saintess）

| 技能 | ID | 开场 | 效果 | 元素 | 自动推断 | 后端缺口 |
|---|---|---|---|---|---|---|
| 冰霜祷言 | `frost_prayer` | 触发 | 治疗 | 水 | 部分 | 选择结算补 target_ids |
| 治愈之光 | `healing_light` | 主动 | 治疗 | 光 | 是 |  |
| 治疗术 | `heal` | 主动 | 治疗 | 光 | 是 |  |
| 圣疗 ★ | `saint_heal` | 主动 | 治疗 → 资源 | 光 | 部分 | 双目标分配后补事件；注：1/3 目标路径在 skill_activated 之前进入 ActionEnd，事件顺序可能错位 |
| 怜悯 | `mercy` | 主动 | 资源 → 形态 |  | 是 |  |

## 魔法少女（magical_girl）

| 技能 | ID | 开场 | 效果 | 元素 | 自动推断 | 后端缺口 |
|---|---|---|---|---|---|---|
| 魔弹掌控 | `magic_bullet_control` | 主动 | — |  | 否 | 方向选择补 skill_activated；注：handler 为死代码，打魔弹时的方向选项；无可视效果 |
| 魔弹融合 | `magic_bullet_fusion` | 主动 | 资源 → 打击 |  | 部分 | 魔弹伤害补 skill_id |
| 魔弹融合 | `magic_bullet_fusion_chain` | 触发 | 资源 → 打击 |  | 部分 | 弃牌路径补 skill_activated；注：InteractionDiscard 路径不发技能事件 |
| 魔爆冲击 | `magic_blast` | 主动 | 资源 → 全局 → 打击 |  | 部分 | 各目标弃牌/受伤补 skill_id；注：先加阵营宝石再校验目标；施法者被迫弃牌不展示 |
| 毁灭风暴 ★ | `destruction_storm` | 主动 | 打击 |  | 是 |  |

## 女武神（valkyrie）

| 技能 | ID | 开场 | 效果 | 元素 | 自动推断 | 后端缺口 |
|---|---|---|---|---|---|---|
| 神圣追击 | `valkyrie_divine_pursuit` | 触发 | 资源 |  | 是 |  |
| 秩序之印 | `valkyrie_order_seal` | 主动 | 资源 → 治疗 |  | 是 |  |
| 和平行者 | `valkyrie_peace_walker` | 触发 | 形态 |  | 部分 | 形态变化补来源；注：描述中“进入英灵形态”实际在英灵召唤 |
| 军威神光 | `valkyrie_military_glory` | 主动 | 治疗 → 形态 → 全局 |  | 否 | 回合开始 hook 触发时补 skill_activated；注：TurnStart hook 直接调 Execute，绕过 Executor |
| 英灵召唤 | `valkyrie_heroic_summon` | 触发 | 打击 → 形态 → 治疗 |  | 部分 | 弃法术治疗补技能事件与目标；注：水晶在 handler 内扣；valkyrie_heroic_summon_pick 路由未使用 |

## 元素师（elementalist）

| 技能 | ID | 开场 | 效果 | 元素 | 自动推断 | 后端缺口 |
|---|---|---|---|---|---|---|
| 元素吸收 | `elementalist_absorb` | 触发 | 形态 |  | 部分 | 注：点燃排除判断按卡名"元素点燃"，但 InflictDamage 生成"直接伤害"，可能误触发 |
| 元素点燃 | `elementalist_ignite` | 主动 | 形态 → 打击 → 资源 | 火 | 是 | 伤害类型 magic_no_absorb 前端可能不识别 |
| 雷击 | `elementalist_thunder_strike` | 主动 | 打击 → 全局 | 雷 | 是 |  |
| 冰冻 | `elementalist_freeze` | 主动 | 打击 → 治疗 | 水 | 部分 | 逐步选目标，skill_activated 无目标 |
| 风刃 | `elementalist_wind_blade` | 主动 | 打击 → 资源 | 风 | 是 |  |
| 陨石 | `elementalist_meteor` | 主动 | 打击 → 资源 | 地 | 是 |  |
| 火球 | `elementalist_fireball` | 主动 | 打击 | 火 | 是 |  |
| 月光 ★ | `elementalist_moonlight` | 主动 | 打击 |  | 是 |  |

## 仲裁者（arbiter）

| 技能 | ID | 开场 | 效果 | 元素 | 自动推断 | 后端缺口 |
|---|---|---|---|---|---|---|
| 仲裁法则 | `arbiter_law` | 触发 | 资源 |  | 部分 | 开局事件可能早于客户端就绪 |
| 审判浪潮 | `arbiter_judgment_tide` | 触发 | 形态 |  | 部分 |  |
| 仲裁仪式 | `arbiter_ritual` | 主动 | 形态 |  | 是 | 回合开始审判+1 补事件；注：宝石在 Execute 内扣 |
| 仪式中断 | `arbiter_ritual_break` | 主动 | 形态 → 全局 |  | 是 |  |
| 末日审判 ★ | `arbiter_doomsday` | 主动 | 形态 → 打击 |  | 是 | 注：强制发动路径是否走 finishSkillUse 未验证 |
| 判决天平 | `arbiter_balance` | 主动 | 形态 → 资源 → 全局 |  | 部分 | 需带所选分支的生效事件，弃牌补 card_revealed |

## 冒险家（adventurer）

| 技能 | ID | 开场 | 效果 | 元素 | 自动推断 | 后端缺口 |
|---|---|---|---|---|---|---|
| 欺诈 | `adventurer_fraud` | 主动 | 打击 |  | 部分 | 虚拟攻击补技能来源与所选元素 |
| 强运 | `adventurer_lucky_fortune` | 触发 | 资源 |  | 部分 |  |
| 地下法则 | `adventurer_underground_law` | 触发 | 全局 |  | 否 | 购买覆盖时补技能事件；注：覆盖钩子为空实现，实际生效路径未验证 |
| 冒险者天堂 | `adventurer_paradise` | 触发 | 资源 |  | 否 | 提炼覆盖补带队友目标的技能事件；注：无 SkillEntry，仅 timing hook 实现 |
| 偷天换日 | `adventurer_steal_sky` | 主动 | 全局 → 形态 |  | 部分 | 需带所选分支的生效事件 |

## 圣枪骑士（holy_lancer）

| 技能 | ID | 开场 | 效果 | 元素 | 自动推断 | 后端缺口 |
|---|---|---|---|---|---|---|
| 神圣启示 | `holy_lancer_revelation` | 触发 | 形态 |  | 否 | 治疗上限变化补事件；注：仅 hook 同步 MaxHeal |
| 辉耀 | `holy_lancer_radiance` | 主动 | 资源 → 治疗 | 水 | 是 | skill_activated 补全场 target_ids |
| 惩戒 | `holy_lancer_punishment` | 主动 | 资源 → 治疗 |  | 是 | 注：直接改 Heal 字段，满上限时 +1 丢失 |
| 圣击 | `holy_lancer_holy_strike` | 触发 | 治疗 |  | 部分 | 跳过地枪路径补 skill_activated；注：主动攻击有治疗时走 responseSkillSkipHook |
| 天枪 | `holy_lancer_sky_spear` | 触发 | 形态 → 状态 | 暗 | 部分 | 补不可应战标记；注：attackGatingHook 静默把攻击转暗系，描述未写 |
| 地枪 | `holy_lancer_earth_spear` | 触发 | 形态 → 打击 | 地 | 部分 | 选 X 后补带数值的生效事件 |
| 圣光祈愈 | `holy_lancer_prayer` | 主动 | 治疗 → 资源 | 光 | 是 | 注：直接 Heal+=2 上限5，无视 MaxHeal |

## 精灵射手（elf_archer）

| 技能 | ID | 开场 | 效果 | 元素 | 自动推断 | 后端缺口 |
|---|---|---|---|---|---|---|
| 元素射击 | `elf_elemental_shot` | 触发 | 资源 → 状态 |  | 部分 | 需注明箭矢元素的生效事件；注：实现有地系箭，描述未列 |
| 动物伙伴 | `elf_animal_companion` | 触发 | 资源 |  | 是 |  |
| 精灵密仪 | `elf_ritual` | 主动 | 形态 → 状态 |  | 部分 | 补“祝福放置”事件；结束释放伤害补技能来源；注：祝福经 draw_cards 摸入再移到场上；释放伤害描述未写 |
| 宠物强化 | `elf_pet_empower` | 触发 | 资源 |  | 是 | 注：水晶在 handler 内扣 |

## 瘟疫法师（plague_mage）

| 技能 | ID | 开场 | 效果 | 元素 | 自动推断 | 后端缺口 |
|---|---|---|---|---|---|---|
| 不朽 | `plague_immortal` | 触发 | 治疗 |  | 是 | 被抑制时不应发 skill_activated；注：被死亡之触抑制时仍发虚假 skill_activated |
| 圣渎 | `plague_blasphemy` | 触发 | 状态 |  | 否 | 注：规则型被动，无需动画 |
| 瘟疫 | `plague_outbreak` | 主动 | 资源 → 打击 → 治疗 | 地 | 部分 | skill_activated 补 target_ids（全体其他角色） |
| 死亡之触 | `plague_death_touch` | 主动 | 治疗 → 资源 → 打击 |  | 部分 | 选择结算补伤害归因 |
| 剧毒新星 ★ | `plague_toxic_nova` | 主动 | 资源 → 打击 → 治疗 |  | 部分 | skill_activated 补 target_ids（全体其他角色）；注：自愈先于伤害结算 |

## 魔剑士（magic_swordsman）

| 技能 | ID | 开场 | 效果 | 元素 | 自动推断 | 后端缺口 |
|---|---|---|---|---|---|---|
| 修罗连斩 | `ms_asura_combo` | 触发 | 资源 | 火 | 是 |  |
| 暗影凝聚 | `ms_shadow_gather` | 主动 | 打击 → 形态 | 暗 | 是 | 注：target_ids 为空（TargetSelf） |
| 暗影之力 | `ms_shadow_power` | 触发 | 打击 | 暗 | 否 | 增伤补 skill_id；注：handler 为死代码；只给主动攻击+1，描述为所有攻击 |
| 暗影抗拒 | `ms_shadow_reject` | 触发 | 状态 |  | 否 | 注：handler 为死代码；仅暗影形态生效，与描述不一致 |
| 暗影流星 | `ms_shadow_meteor` | 主动 | 资源 → 打击 → 全局 → 形态 | 暗 | 部分 | 可选转正/扣星石补事件；注：直接 user.Form=""，可能不发转正事件 |
| 黄泉震颤 | `ms_yellow_spring` | 触发 | 资源 → 状态 |  | 部分 | 命中补牌/弃2 补 skill_id；注：宝石在 handler 内扣 |

## 血色剑灵（crimson_sword_spirit）

| 技能 | ID | 开场 | 效果 | 元素 | 自动推断 | 后端缺口 |
|---|---|---|---|---|---|---|
| 血色荆棘 | `css_blood_thorns` | 触发 | 形态 |  | 部分 | target_ids 指向攻击目标而非自身；注：上限可能被散华轮舞提到4 |
| 赤色一闪 | `css_crimson_flash` | 触发 | 形态 → 打击 → 资源 |  | 是 |  |
| 血染蔷薇 | `css_blood_rose` | 主动 | 形态 → 状态 → 资源 → 治疗 → 全局 |  | 部分 | 两步选择后补生效事件；庭院 AoE 补技能来源；注：血-2 在最后一步选择才扣 |
| 血气屏障 | `css_blood_barrier` | 触发 | 形态 → 状态 → 打击 |  | 是 |  |
| 血蔷薇庭院 | `css_rose_courtyard` | 触发 | 状态 → 全局 |  | 否 | 回合结束收回补事件；注：handler 为空壳 |
| 散华轮舞 | `css_dance` | 主动 | 状态 → 形态 → 资源 → 全局 |  | 部分 | 模式选择后补生效事件；放置补 buff 事件；注：费用在选择内扣；庭院先放置再校验费用，失败时不回滚 |

## 祈祷师（prayer_master）

| 技能 | ID | 开场 | 效果 | 元素 | 自动推断 | 后端缺口 |
|---|---|---|---|---|---|---|
| 祈祷 | `prayer_enter_form` | 主动 | 资源 → 形态 |  | 部分 | 符文增加时补技能事件；注：符文累计仅 hook |
| 光辉信仰 | `prayer_radiant_faith` | 主动 | 形态 → 资源 → 全局 → 治疗 |  | 是 |  |
| 黑暗诅咒 | `prayer_dark_curse` | 主动 | 形态 → 打击 | 暗 | 是 |  |
| 威力赐福 | `prayer_power_blessing` | 主动 | 状态 → 打击 |  | 部分 | 实现 +2 伤害并补触发事件；注：触发只移除赐福牌，+2 伤害未生效 |
| 迅捷赐福 | `prayer_swift_blessing` | 主动 | 状态 → 资源 |  | 部分 | 触发时补 skill_activated |
| 法力潮汐 | `prayer_mana_tide` | 触发 | 资源 |  | 是 |  |

## 红莲骑士（crimson_knight）

| 技能 | ID | 开场 | 效果 | 元素 | 自动推断 | 后端缺口 |
|---|---|---|---|---|---|---|
| 腥红圣约 | `crk_crimson_pact` | 触发 | 治疗 |  | 是 |  |
| 腥红信仰 | `crk_crimson_faith` | 触发 | 状态 |  | 否 | 静态规则；注：handler 为空壳，在 healResistHook |
| 血腥祷言 | `crk_bloody_prayer` | 主动 | 打击 → 治疗 → 形态 |  | 部分 | X 选择后补带队友目标的生效事件；注：直接改 user.Heal |
| 杀戮盛宴 | `crk_killing_feast` | 触发 | 形态 → 打击 |  | 是 | +2 加伤无单独事件 |
| 热血沸腾 | `crk_hot_blood` | 触发 | 形态 → 治疗 |  | 否 | 形态进出 hook 补技能事件；注：handler 为空壳；士气免疫实现位置未找到 |
| 戒骄戒躁 | `crk_calm_mind` | 触发 | 形态 → 资源 |  | 是 | 注：描述为同类型额外行动，实现给任意类型 |
| 腥红十字 ★ | `crk_crimson_cross` | 主动 | 形态 → 打击 |  | 是 |  |

## 英灵人形（war_homunculus）

| 技能 | ID | 开场 | 效果 | 元素 | 自动推断 | 后端缺口 |
|---|---|---|---|---|---|---|
| 战纹掌控 | `hom_battle_pattern` | 触发 | 形态 |  | 否 | 注：Execute 为空，开局初始化 |
| 怒火压制 | `hom_rage_suppress` | 触发 | 形态 |  | 是 |  |
| 战纹碎击 | `hom_rune_smash` | 触发 | 资源 → 形态 → 打击 |  | 部分 | 选牌结算后补 X/Y 与加伤事件；注：未校验与攻击牌同系；hom_rune_smash_x 路由疑似无用 |
| 魔纹融合 | `hom_glyph_fusion` | 触发 | 资源 → 形态 → 打击 |  | 部分 | 选牌结算后补伤害来源技能；注：未校验与攻击牌异系 |
| 符文改造 | `hom_rune_reforge` | 主动 | 形态 → 资源 |  | 部分 | 重新分配符文/退形态补事件 |
| 双重回响 | `hom_dual_echo` | 触发 | 打击 → 资源 |  | 部分 | skill_activated 补 target_ids；注：水晶在选目标后才扣；候选目标可能含自己 |

## 神官（priest）

| 技能 | ID | 开场 | 效果 | 元素 | 自动推断 | 后端缺口 |
|---|---|---|---|---|---|---|
| 神圣启示 | `priest_divine_revelation` | 触发 | 治疗 |  | 是 |  |
| 神圣祈福 | `priest_divine_bless` | 主动 | 资源 → 治疗 |  | 是 |  |
| 水之神力 | `priest_water_power` | 主动 | 资源 → 治疗 | 水 | 部分 | 补给牌事件；注：第2张牌直接进队友手牌，无事件 |
| 圣使守护 | `priest_guardian` | 触发 | 状态 |  | 否 | 注：规则型被动，无需动画 |
| 神圣契约 | `priest_divine_contract` | 主动 | 资源 → 治疗 |  | 部分 | 选择结算补 target_ids 与转移量；注：直接改 Heal 字段 |
| 神圣领域 | `priest_divine_domain` | 主动 | 资源 → 治疗 → 打击 |  | 部分 | 分支结算补 target_ids |

## 阴阳师（onmyoji）

| 技能 | ID | 开场 | 效果 | 元素 | 自动推断 | 后端缺口 |
|---|---|---|---|---|---|---|
| 式神降临 | `onmyoji_shikigami_descend` | 主动 | 资源 → 形态 |  | 是 |  |
| 阴阳转换 | `onmyoji_yinyang_shift` | 触发 | 资源 → 形态 → 打击 |  | 否 | 补 skill_activated + 攻击者 target；注：handler 为死代码，走 combatInteraction hook |
| 式神转换 | `onmyoji_shikigami_shift` | 触发 | 资源 → 形态 |  | 部分 | 补 skill_activated；注：直接调 handler.Execute，绕过 Executor |
| 黑暗祭礼 | `onmyoji_dark_ritual` | 触发 | 形态 → 打击 | 暗 | 否 | 选择结算补 skill_activated + target_ids；注：仅 hook |
| 式神咒束 | `onmyoji_binding` | 触发 | 全局 → 打击 |  | 否 | 扣除宝石水晶并补 skill_activated；注：只检查不扣除 1宝石1水晶 |
| 生命结界 ★ | `onmyoji_life_barrier` | 主动 | 形态 → 资源 → 治疗 → 打击 → 状态 |  | 部分 | 分支结算补生效事件/target_ids；注：缺“X=3 不降士气”规则 |

## 苍炎魔女（blaze_witch）

| 技能 | ID | 开场 | 效果 | 元素 | 自动推断 | 后端缺口 |
|---|---|---|---|---|---|---|
| 永生银时计 | `bw_rebirth_clock` | 触发 | 形态 |  | 否 | 士气损失 hook 补技能事件；注：handler 为空壳，全在 moraleLossHook |
| 苍炎法典 | `bw_blazing_codex` | 主动 | 打击 | 火 | 是 |  |
| 天火断空 | `bw_heavenfire_cleave` | 主动 | 打击 → 形态 | 火 | 是 |  |
| 魔女之怒 | `bw_witch_wrath` | 主动 | 形态 → 资源 | 火 | 部分 | 摸牌选择后补生效事件；解除形态补事件 |
| 替身玩偶 | `bw_substitute_doll` | 触发 | 资源 |  | 部分 | 需带所选队友的生效事件；注：定义有 CostDiscards 但 InteractionType 为 None，弃牌在角色选择流 |
| 痛苦链接 | `bw_pain_link` | 主动 | 打击 → 资源 |  | 是 | 弃至3张提示补技能来源 |
| 魔能反转 | `bw_mana_inversion` | 触发 | 打击 |  | 部分 | 选牌后补带目标的生效事件；注：水晶在 handler 内扣 |

## 贤者（sage）

| 技能 | ID | 开场 | 效果 | 元素 | 自动推断 | 后端缺口 |
|---|---|---|---|---|---|---|
| 智慧法典 | `sage_wisdom_codex` | 触发 | 资源 |  | 否 | hook 触发时补 skill_activated；注：handler 为死代码 |
| 法术反弹 | `sage_magic_rebound` | 触发 | 资源 → 打击 |  | 否 | 选择结算补 skill_activated + target_ids；注：handler 为死代码 |
| 魔道法典 ★ | `sage_arcane_codex` | 主动 | 资源 → 打击 |  | 部分 | 选择结算补 target_ids；注：Execute 忽略传入 target |
| 圣洁法典 | `sage_holy_codex` | 主动 | 资源 → 治疗 → 打击 |  | 部分 | 选择结算补 target_ids；注：Execute 忽略传入 target |

## 魔弓（magic_bow）

| 技能 | ID | 开场 | 效果 | 元素 | 自动推断 | 后端缺口 |
|---|---|---|---|---|---|---|
| 魔贯冲击 | `mb_magic_pierce` | 触发 | 形态 → 打击 | 火 | 部分 | 未命中3点法伤补 skill_id；注：CanUse 另限目标手牌未满 |
| 雷光散射 | `mb_thunder_scatter` | 主动 | 形态 → 打击 | 雷 | 部分 | 额外伤害目标补 target_id |
| 多重射击 | `mb_multi_shot` | 触发 | 形态 → 打击 | 风 | 部分 | 虚拟攻击补 skill_id；注：移除风系充能，虚拟攻击为暗系 |
| 充能 | `mb_charge` | 主动 | 资源 → 形态 |  | 部分 | 盖充能补 skill_id |
| 魔眼 | `mb_demon_eye` | 主动 | 资源 → 形态 |  | 部分 | 分支①目标补 target_id |

## 魔枪（magic_lancer）

| 技能 | ID | 开场 | 效果 | 元素 | 自动推断 | 后端缺口 |
|---|---|---|---|---|---|---|
| 暗之解放 | `ml_dark_release` | 主动 | 形态 | 暗 | 是 | 下一击加伤消费补归属 |
| 幻影星尘 | `ml_phantom_stardust` | 主动 | 打击 → 形态 | 暗 | 部分 | 追加2点法伤补 skill_id；注：启动确认不传目标 |
| 黑暗束缚 | `ml_dark_bind` | 触发 | 状态 |  | 否 | 注：handler 为死代码，纯规则限制 |
| 暗之障壁 | `ml_dark_barrier` | 触发 | 资源 | 雷 | 部分 | 弃牌结算补事件；注：只弃牌无其他效果，疑似未完成 |
| 充盈 | `ml_fullness` | 主动 | 资源 → 全局 → 形态 |  | 部分 | 各角色弃牌与增伤补 skill_id |
| 漆黑之枪 ★ | `ml_black_spear` | 触发 | 资源 → 打击 | 暗 | 部分 | 增伤数值补事件；注：水晶 X 在 choices.go 扣 |

## 灵符师（spirit_caster）

| 技能 | ID | 开场 | 效果 | 元素 | 自动推断 | 后端缺口 |
|---|---|---|---|---|---|---|
| 灵符-雷鸣 | `sc_talisman_thunder` | 主动 | 打击 | 雷 | 部分 | 伤害补 skill_id/崩解加成标记；注：伤害在念咒/崩解确认后才结算 |
| 灵符-风行 | `sc_talisman_wind` | 主动 | 资源 | 风 | 部分 | 目标弃牌补来源技能；注：无手牌分支可取消，已付风系弃牌代价却跳过效果 |
| 念咒 | `sc_incantation` | 触发 | 形态 |  | 否 | 放置妖力时补 skill_activated；注：handler CanUse 恒 false，内嵌于灵符流程 |
| 百鬼夜行 | `sc_hundred_night` | 触发 | 形态 → 打击 |  | 部分 | skill_activated 补 target_ids；火系范围补事件 |
| 灵力崩解 | `sc_spiritual_collapse` | 触发 | 资源 → 打击 |  | 否 | 确认崩解时补 skill_activated；注：handler CanUse 恒 false，只是灵符流程内的确认 |

## 吟游诗人（bard）

| 技能 | ID | 开场 | 效果 | 元素 | 自动推断 | 后端缺口 |
|---|---|---|---|---|---|---|
| 沉沦协奏曲 | `bd_descent_concerto` | 触发 | 资源 → 形态 → 打击 |  | 否 | 回合结束钩子补 skill_activated；注：handler 为空壳，实际在 hook；描述为诗人造成伤害，实现记录己方任意法术伤害 |
| 不谐和弦 | `bd_dissonance_chord` | 主动 | 形态 → 资源 |  | 部分 | 弃牌补 card_revealed |
| 禁忌诗篇 | `bd_forbidden_verse` | 触发 | 形态 → 状态 → 打击 |  | 否 | 补 skill_activated；注：handler 为空壳，在曲目选择内联结算 |
| 激昂狂想曲 | `bd_rousing_rhapsody` | 触发 | 打击 → 资源 |  | 否 | 补 skill_activated（持有者与诗人）；注：发动者是永恒乐章持有者，伤害来源为诗人 |
| 胜利交响诗 | `bd_victory_symphony` | 触发 | 资源 → 全局 → 治疗 |  | 否 | 补 skill_activated（持有者与诗人）；注：描述写“你+1治疗”，实现治疗持有者 |
| 希望赋格曲 | `bd_hope_fugue` | 主动 | 资源 → 状态 → 治疗 |  | 部分 | 放置/转移乐章的目标补进事件 |

## 勇者（hero）

| 技能 | ID | 开场 | 效果 | 元素 | 自动推断 | 后端缺口 |
|---|---|---|---|---|---|---|
| 勇者之心 | `hero_heart` | 触发 | 资源 |  | 否 | 注：handler 为死代码，开局 ApplyDefaults 生效 |
| 怒吼 | `hero_roar` | 触发 | 形态 → 资源 → 打击 |  | 部分 | 未命中知性+1 补技能事件；注：miss hook 知性上限写 3，token 上限为 4，明镜止水需 4 |
| 禁断之力 | `hero_forbidden_power` | 触发 | 资源 → 形态 → 打击 | 火 | 部分 | 火系增伤补 skill_id |
| 精疲力竭 | `hero_exhaustion` | 触发 | 形态 → 资源 → 打击 |  | 否 | 精疲力竭结束补 skill_activated；注：handler 为死代码，进入在禁断之力，结束在 hook |
| 明镜止水 | `hero_calm_mind` | 触发 | 形态 → 资源 |  | 部分 | 攻击结束水晶+1 补来源 |
| 挑衅 | `hero_taunt` | 主动 | 形态 → 状态 |  | 是 |  |
| 死斗 | `hero_dead_duel` | 触发 | 资源 → 形态 → 全局 |  | 部分 | 士气固定下降1 标注来源；注：宝石在 handler 内扣 |

## 格斗家（fighter）

| 技能 | ID | 开场 | 效果 | 元素 | 自动推断 | 后端缺口 |
|---|---|---|---|---|---|---|
| 念气力场 | `fighter_psi_field` | 触发 | 状态 |  | 部分 | 补减伤前后数值；注：skill_activated 无目标 |
| 蓄力一击 | `fighter_charge_strike` | 触发 | 形态 → 打击 |  | 部分 | 未命中自伤补 skill_id；注：未命中自伤只在 attackMissHook |
| 念弹 | `fighter_psi_bullet` | 触发 | 形态 → 打击 |  | 部分 | 选目标后补带 target_id 的生效事件；注：伤害与自伤在 choices.go |
| 百式幻龙拳 | `fighter_hundred_dragon` | 主动 | 形态 → 状态 |  | 部分 | 锁定目标后补 target_ids；注：加伤与退场在 hook |
| 气绝崩击 | `fighter_burst_crash` | 触发 | 形态 → 打击 |  | 是 |  |
| 斗神天驱 | `fighter_war_god_drive` | 主动 | 资源 → 治疗 |  | 部分 | 弃牌后治疗关联 skill_id |

## 圣弓（holy_bow）

| 技能 | ID | 开场 | 效果 | 元素 | 自动推断 | 后端缺口 |
|---|---|---|---|---|---|---|
| 天之弓 | `hb_heavenly_bow` | 触发 | 资源 → 形态 → 打击 |  | 否 | 信仰+1 补来源；注：handler 为死代码，全在 ApplyDefaults/hook |
| 圣屑飓暴 | `hb_holy_shard_storm` | 主动 | 资源 → 打击 |  | 部分 | 虚拟攻击补 skill_id；注：元素随弃牌动态决定 |
| 圣煌降临 | `hb_radiant_descent` | 主动 | 资源 → 形态 | 光 | 部分 | 支付方式选择后补事件 |
| 圣光爆裂 | `hb_light_burst` | 主动 | 资源 → 治疗 → 形态 → 打击 | 光 | 部分 | 分支结算补带 target_ids 的生效事件 |
| 流星圣弹 | `hb_meteor_bullet` | 触发 | 资源 → 治疗 |  | 部分 | 治疗目标补 target_id |
| 圣煌辉光炮 ★ | `hb_radiant_cannon` | 主动 | 形态 → 资源 → 全局 | 光 | 部分 | 星杯/士气对齐补 skill_id；注：CanUse 要求士气严格落后，描述未写 |
| 自动填充 | `hb_auto_fill` | 触发 | 资源 → 形态 → 治疗 |  | 否 | 回合末触发补 skill_activated；注：handler 为死代码，仅 TurnEndFinal hook |

## 剑帝（sword_emperor）

| 技能 | ID | 开场 | 效果 | 元素 | 自动推断 | 后端缺口 |
|---|---|---|---|---|---|---|
| 剑魂守护 | `se_sword_soul_guard` | 触发 | 形态 |  | 否 | 攻击牌转剑魂时补 skill_activated；注：handler 恒 false，逻辑在 attackMissHook；runtime.go 有无调用方的死代码 |
| 佯攻 | `se_feint` | 触发 | 形态 |  | 否 | 剑气+1 时补 skill_activated；注：handler 恒 false，逻辑在 attackMissHook |
| 剑气斩 | `se_sword_qi_slash` | 触发 | 形态 → 打击 |  | 部分 | skill_activated 补 target_ids |
| 天使之魂 | `se_angel_soul` | 触发 | 形态 → 治疗 → 全局 | 光 | 部分 | 命中/未命中分支补技能事件 |
| 恶魔之魂 | `se_demon_soul` | 触发 | 形态 → 打击 | 暗 | 部分 | 加伤/未命中加剑气补技能事件 |
| 不屈意志 | `se_indomitable_will` | 触发 | 资源 → 形态 |  | 是 |  |

## 兽灵武士（beast_samurai）

| 技能 | ID | 开场 | 效果 | 元素 | 自动推断 | 后端缺口 |
|---|---|---|---|---|---|---|
| 武者残心 | `bs_warrior_zanshin` | 触发 | 形态 |  | 部分 |  |
| 一击无念 | `bs_one_strike_no_thought` | 触发 | 形态 |  | 是 | 下一击加成消费时补事件 |
| 兽魂意念 | `bs_beast_soul_will` | 触发 | 形态 |  | 否 | 命中时兽魂+1 补技能事件；注：handler 为空壳，全在 hook |
| 兽魂警戒 | `bs_beast_soul_alert` | 触发 | 形态 → 资源 |  | 部分 | 对方弃牌在后续选择，补生效事件 |
| 兽返 | `bs_beast_return` | 触发 | 形态 → 资源 |  | 部分 | X 选择后补生效事件 |
| 逆反居合斩 | `bs_reversal_iaijutsu` | 触发 | 形态 → 资源 → 全局 |  | 部分 | 补“伤害改为弃牌”事件；注：maxX 可为 0，与描述“移除兽魂”不符 |
| 御魂流居合式 | `bs_iaijutsu_style` | 主动 | 形态 → 资源 |  | 部分 | 居合形态加伤与维持补事件 |

## 灵魂术士（soul_sorcerer）

| 技能 | ID | 开场 | 效果 | 元素 | 自动推断 | 后端缺口 |
|---|---|---|---|---|---|---|
| 灵魂吞噬 | `ss_soul_devour` | 触发 | 形态 |  | 否 | 士气下降加黄魂时补 skill_activated；注：handler CanUse 恒 false，逻辑在 AfterMoraleLossHook |
| 灵魂召还 | `ss_soul_recall` | 主动 | 资源 → 形态 |  | 部分 | 选牌结算后补弃牌数/蓝魂增量事件；注：弃牌与蓝魂在 choices.go |
| 灵魂转换 | `ss_soul_convert` | 触发 | 形态 |  | 否 | 转换生效时补 skill_activated；注：handler CanUse 恒 false，逻辑在攻击宣言 hook |
| 灵魂镜像 | `ss_soul_mirror` | 主动 | 资源 → 形态 |  | 是 |  |
| 灵魂震爆 | `ss_soul_blast` | 主动 | 形态 → 打击 |  | 是 | 注：需弃1张独有牌，描述未写 |
| 灵魂赐予 | `ss_soul_grant` | 主动 | 形态 → 资源 |  | 是 | 注：需弃1张独有牌，描述未写 |
| 灵魂链接 | `ss_soul_link` | 主动 | 形态 → 状态 |  | 部分 | skill_activated 补 target_ids；转伤时补技能事件；注：runtime.go MaybeSoulLinkTransfer 无调用方 |
| 灵魂增幅 | `ss_soul_amp` | 主动 | 形态 |  | 是 |  |

## 月之女神（moon_goddess）

| 技能 | ID | 开场 | 效果 | 元素 | 自动推断 | 后端缺口 |
|---|---|---|---|---|---|---|
| 新月庇护 | `mg_new_moon_shelter` | 触发 | 形态 → 状态 → 全局 |  | 部分 | skill_activated 补 target_ids（爆牌角色） |
| 闇月诅咒 | `mg_dark_moon_curse` | 触发 | 全局 → 形态 | 暗 | 否 | 移除闇月时补 skill_activated；注：handler 为死代码，逻辑在 RemoveDarkMoon* |
| 美杜莎之眼 | `mg_medusa_eye` | 触发 | 状态 → 治疗 → 形态 → 全局 → 打击 |  | 部分 | 选择结算后补生效事件 |
| 月之轮回 | `mg_moon_cycle` | 触发 | 状态 → 治疗 → 全局 → 形态 |  | 否 | 回合结束 hook 补 skill_activated + target_ids；注：handler 为死代码，仅 hook |
| 月渎 | `mg_blasphemy` | 触发 | 治疗 → 打击 |  | 否 | 结算时补 skill_activated + target_ids；注：handler 为死代码，仅 hook |
| 闇月斩 | `mg_darkmoon_slash` | 触发 | 形态 → 全局 → 打击 | 暗 | 部分 | 选 X 后补加伤事件 |
| 苍白之月 ★ | `mg_pale_moon` | 主动 | 形态 → 资源 → 状态 → 打击 |  | 部分 | 分支结算补生效事件/目标 |

## 血之巫女（blood_priestess）

| 技能 | ID | 开场 | 效果 | 元素 | 自动推断 | 后端缺口 |
|---|---|---|---|---|---|---|
| 血之哀伤 | `bp_blood_sorrow` | 主动 | 打击 → 状态 |  | 部分 | 移除/转移同生共死补生效事件；注：flow_handlers.go 有两份重复的移除/转移逻辑 |
| 流血 | `bp_bleeding` | 触发 | 形态 → 治疗 → 打击 |  | 否 | 流血形态 hook 补技能事件；注：handler 为空壳；turn_hooks.go BleedTick 与 hook 重复 |
| 逆流 | `bp_backflow` | 主动 | 资源 → 治疗 |  | 是 |  |
| 血之悲鸣 | `bp_blood_wail` | 主动 | 打击 |  | 部分 | X 选择后伤害才入队 |
| 同生共死 | `bp_shared_life` | 主动 | 资源 → 状态 |  | 部分 | 目标选择后补生效事件；放置补 buff 事件；注：放置未调用 emitBuffAdded |
| 血之诅咒 | `bp_blood_curse` | 主动 | 打击 → 资源 |  | 是 | 弃3张补技能来源 |

## 蝶舞者（butterfly_dancer）

| 技能 | ID | 开场 | 效果 | 元素 | 自动推断 | 后端缺口 |
|---|---|---|---|---|---|---|
| 生命之火 | `bt_life_fire` | 触发 | 状态 |  | 否 | 静态规则，无需动画；注：handler 为空壳，纯静态手牌上限 |
| 舞动 | `bt_dance` | 主动 | 资源 → 状态 |  | 部分 | 补“茧放置”事件；注：茧经 draw_cards 摸入后移到场上，摸牌事件有误导 |
| 毒粉 | `bt_poison_powder` | 触发 | 状态 → 打击 |  | 否 | damageResponseHook 补技能事件；注：handler 为空壳 |
| 朝圣 | `bt_pilgrimage` | 触发 | 状态 |  | 否 | 补技能事件与减伤数值；注：handler 为空壳 |
| 镜花水月 | `bt_mirror` | 触发 | 状态 → 打击 |  | 否 | 补“伤害抵消”事件与技能来源；注：描述写打伤害来源，实现打原目标 |
| 凋零 | `bt_wither` | 触发 | 打击 → 全局 |  | 否 | 补技能事件；士气下限状态不在 state_delta；注：handler 为空壳；turn_hooks.go WitherExpiry 与 hook 重复 |
| 蛹化 | `bt_chrysalis` | 主动 | 形态 → 状态 |  | 是 |  |
| 倒逆之蝶 ★ | `bt_reverse_butterfly` | 主动 | 资源 → 打击 → 形态 |  | 部分 | 分支选择后补带目标的生效事件 |

## 需要后端补事件的技能

按使用频率粗排（高频 = 每次攻击/伤害/回合都可能触发的无费技能；低频 = 需宝石/水晶的技能）。只列“否”级；规则型被动（生命之火、腥红信仰、血蔷薇庭院、勇者之心、战纹掌控、黑暗束缚、暗影抗拒、圣渎、圣使守护）不需要动画，未列入。“部分”级的缺口见各角色表。

- 高频：
  - 佯攻（剑帝，`se_feint`）：剑气+1 时补 skill_activated
  - 剑魂守护（剑帝，`se_sword_soul_guard`）：攻击牌转剑魂时补 skill_activated
  - 兽魂意念（兽灵武士，`bs_beast_soul_will`）：命中时兽魂+1 补技能事件
  - 暗影之力（魔剑士，`ms_shadow_power`）：增伤补 skill_id
  - 阴阳转换（阴阳师，`onmyoji_yinyang_shift`）：补 skill_activated + 攻击者 target
  - 灵魂转换（灵魂术士，`ss_soul_convert`）：转换生效时补 skill_activated
  - 灵魂吞噬（灵魂术士，`ss_soul_devour`）：士气下降加黄魂时补 skill_activated
  - 毒粉（蝶舞者，`bt_poison_powder`）：damageResponseHook 补技能事件
  - 朝圣（蝶舞者，`bt_pilgrimage`）：补技能事件与减伤数值
  - 镜花水月（蝶舞者，`bt_mirror`）：补“伤害抵消”事件与技能来源
  - 凋零（蝶舞者，`bt_wither`）：补技能事件；士气下限状态不在 state_delta
  - 念咒（灵符师，`sc_incantation`）：放置妖力时补 skill_activated
  - 永生银时计（苍炎魔女，`bw_rebirth_clock`）：士气损失 hook 补技能事件
  - 闇月诅咒（月之女神，`mg_dark_moon_curse`）：移除闇月时补 skill_activated
  - 流血（血之巫女，`bp_bleeding`）：流血形态 hook 补技能事件
  - 热血沸腾（红莲骑士，`crk_hot_blood`）：形态进出 hook 补技能事件
  - 月渎（月之女神，`mg_blasphemy`）：结算时补 skill_activated + target_ids
  - 法术反弹（贤者，`sage_magic_rebound`）：选择结算补 skill_activated + target_ids
  - 智慧法典（贤者，`sage_wisdom_codex`）：hook 触发时补 skill_activated
  - 月之轮回（月之女神，`mg_moon_cycle`）：回合结束 hook 补 skill_activated + target_ids
  - 黑暗祭礼（阴阳师，`onmyoji_dark_ritual`）：选择结算补 skill_activated + target_ids
  - 沉沦协奏曲（吟游诗人，`bd_descent_concerto`）：回合结束钩子补 skill_activated
  - 激昂狂想曲（吟游诗人，`bd_rousing_rhapsody`）：补 skill_activated（持有者与诗人）
  - 胜利交响诗（吟游诗人，`bd_victory_symphony`）：补 skill_activated（持有者与诗人）
  - 禁忌诗篇（吟游诗人，`bd_forbidden_verse`）：补 skill_activated
  - 魔弹掌控（魔法少女，`magic_bullet_control`）：方向选择补 skill_activated
  - 精疲力竭（勇者，`hero_exhaustion`）：精疲力竭结束补 skill_activated
  - 天之弓（圣弓，`hb_heavenly_bow`）：信仰+1 补来源
  - 军威神光（女武神，`valkyrie_military_glory`）：回合开始 hook 触发时补 skill_activated
  - 冒险者天堂（冒险家，`adventurer_paradise`）：提炼覆盖补带队友目标的技能事件
  - 神圣启示（圣枪骑士，`holy_lancer_revelation`）：治疗上限变化补事件
- 中频：
  - 地下法则（冒险家，`adventurer_underground_law`）：购买覆盖时补技能事件
  - 式神咒束（阴阳师，`onmyoji_binding`）：扣除宝石水晶并补 skill_activated
- 低频：
  - 灵力崩解（灵符师，`sc_spiritual_collapse`）：确认崩解时补 skill_activated
  - 自动填充（圣弓，`hb_auto_fill`）：回合末触发补 skill_activated
- 部分级中最常见的共性缺口：伤害/治疗不带 skill_id（加伤、减伤只能从数值反推）；效果在后续 choice 里结算、skill_activated 先发且无目标；弃牌代价路径（贯穿射击、水影、魔弹融合）完全不发 skill_activated；五系封印触发时技能事件记在被封印者名下。
