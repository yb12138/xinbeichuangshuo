/**
 * 解析行动选择提示（prompt.message）的文案，拆成可分别展示的部分。
 * 文案来源：internal/engine/action_selection_policies.go finalizeActionSelectionPromptState
 * 以及各角色的 actionRulePromptMessage 覆盖（仲裁者、格斗家、勇者等）。
 */
export type ActionHubNoticeKind =
  | 'default' // 请选择行动类型：顶栏已表达，不再显示
  | 'extra' // 额外攻击/法术/任意行动：放到顶栏
  | 'extra-none' // 额外行动但无可执行动作：提示可放弃
  | 'rule' // 强制攻击/强制技能/挑衅等规则约束：醒目警示条
  | 'other' // 未识别的文案：原样显示

export interface ActionHubNotice {
  kind: ActionHubNoticeKind
  /** 去掉"请选择行动类型"和充盈后缀之后的正文 */
  text: string
  /** 额外行动的简短限制，例如"仅可攻击"，用于顶栏 */
  extraLimit: string
  /** 【充盈】攻击加成值，0 表示无 */
  fullnessBonus: number
}

const FULLNESS_RE = /（?【充盈】下一次主动攻击伤害额外\+(\d+)）?/

export function parseActionHubNotice(raw: string | null | undefined): ActionHubNotice {
  let text = String(raw || '').trim()
  let fullnessBonus = 0
  const fullness = text.match(FULLNESS_RE)
  if (fullness) {
    fullnessBonus = Number(fullness[1]) || 0
    text = text.replace(FULLNESS_RE, '').trim()
  }
  // 规则句末尾常带"请选择行动类型"，顶栏已经说明，去掉
  text = text.replace(/[。.]?\s*请选择行动类型[。.]?$/, '').trim()

  if (!text) {
    return { kind: 'default', text: '', extraLimit: '', fullnessBonus }
  }
  if (text.includes('额外行动阶段') && text.includes('没有满足约束')) {
    return { kind: 'extra-none', text, extraLimit: '无可执行动作', fullnessBonus }
  }
  if (text.startsWith('当前为额外')) {
    let extraLimit = ''
    if (text.includes('仅可执行攻击')) extraLimit = '仅可攻击'
    else if (text.includes('仅可执行法术')) extraLimit = '仅可法术'
    else if (text.includes('攻击或法术')) extraLimit = '攻击或法术'
    return { kind: 'extra', text, extraLimit, fullnessBonus }
  }
  if (/必须|只能|约束|审判已达上限/.test(text)) {
    return { kind: 'rule', text, extraLimit: '', fullnessBonus }
  }
  return { kind: 'other', text, extraLimit: '', fullnessBonus }
}
