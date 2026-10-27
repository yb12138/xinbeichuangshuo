/**
 * 决策选项的展示信息：从选项文案推断图标与后果预览（如"摸1张牌 → 手牌 4→5"）。
 * 只做文案层面的推断，不改变选项本身；推断不出时只显示标题。
 */
export interface OptionPreview {
  title: string
  desc: string
  glyph: string
  tone: 'gain' | 'loss' | 'neutral'
}

function split(label: string): { title: string; rest: string } {
  // "分支①：对目标造成1点法术伤害" → 标题"分支①"，说明为冒号后的内容
  const m = label.match(/^(.{1,8}?)[：:](.+)$/)
  if (m) return { title: m[1]!.trim(), rest: m[2]!.trim() }
  // "默认方向（右手边，前一位对手）" → 括号内容作为说明
  const paren = label.match(/^(.+?)[（(]([^）)]+)[）)]\s*$/)
  if (paren) return { title: paren[1]!.trim(), rest: paren[2]!.trim() }
  return { title: label, rest: '' }
}

export function optionPreview(label: string, hint: string | undefined, handCount: number | null): OptionPreview {
  const text = String(label || '').trim()
  const { title, rest } = split(text)
  let desc = String(hint || '').trim() || rest
  let glyph = 'sun'
  let tone: OptionPreview['tone'] = 'neutral'

  const draw = text.match(/摸\s*(\d+)\s*张/)
  const discard = text.match(/弃\s*(\d+)\s*张/)
  if (draw) {
    glyph = 'cards'
    tone = 'gain'
    if (!desc && handCount !== null) desc = `手牌 ${handCount} → ${handCount + Number(draw[1])}`
  } else if (discard) {
    glyph = 'trash'
    tone = 'loss'
    if (!desc && handCount !== null) desc = `手牌 ${handCount} → ${Math.max(0, handCount - Number(discard[1]))}`
  } else if (/治疗/.test(text)) {
    glyph = 'twin-heart'
    tone = 'gain'
  } else if (/伤害|攻击/.test(text)) {
    glyph = 'sword-qi'
    tone = 'loss'
  } else if (/宝石|水晶|星石|能量/.test(text)) {
    glyph = 'crystal'
    tone = 'gain'
  } else if (/跳过|放弃|不发动|取消/.test(text)) {
    glyph = 'skip'
  }
  return { title, desc, glyph, tone }
}

/**
 * 技能描述里"选择…；然后…"的共同后果，例如舞动的"然后将牌库顶1张牌面朝下放置为茧"。
 * 只取分号或句号之后以"然后"开头的那一句；没有就返回空字符串。
 */
export function sharedConsequence(skillDescription: string | undefined): string {
  const text = String(skillDescription || '')
  const m = text.match(/[；;。]\s*(然后[^；;。]+)/)
  return m ? m[1]!.trim() : ''
}
