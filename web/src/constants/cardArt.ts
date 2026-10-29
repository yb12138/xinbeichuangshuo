import type { Card } from '../types/game'

/** 卡名 → public/image 下的卡图文件名（拼音） */
const CARD_IMAGE_PINYIN_MAP: Record<string, string> = {
  火焰斩: 'huoYanZhan',
  水涟斩: 'shuiLianZhan',
  地裂斩: 'diLieZhan',
  风神斩: 'fengShenZhan',
  雷光斩: 'leiGuangZhan',
  暗灭: 'anMie',
  魔弹: 'moDan',
  圣光: 'shengGuang',
  圣盾: 'shengDun',
  虚弱: 'xuRuo',
  中毒: 'zhongDu'
}

function fallbackPinyinByCard(card: Card): string {
  if (card.type === 'Attack') {
    if (card.name.includes('暗')) return 'anMie'
    if (card.name.includes('火')) return 'huoYanZhan'
    if (card.name.includes('水')) return 'shuiLianZhan'
    if (card.name.includes('地')) return 'diLieZhan'
    if (card.name.includes('风')) return 'fengShenZhan'
    if (card.name.includes('雷')) return 'leiGuangZhan'
    if (card.element === 'Fire') return 'huoYanZhan'
    if (card.element === 'Water') return 'shuiLianZhan'
    if (card.element === 'Earth') return 'diLieZhan'
    if (card.element === 'Wind') return 'fengShenZhan'
    if (card.element === 'Thunder') return 'leiGuangZhan'
    if (card.element === 'Dark') return 'anMie'
    return 'moRen'
  }
  if (card.name.includes('圣盾')) return 'shengDun'
  if (card.name.includes('圣光')) return 'shengGuang'
  if (card.name.includes('魔弹')) return 'moDan'
  if (card.name.includes('中毒')) return 'zhongDu'
  if (card.name.includes('虚弱')) return 'xuRuo'
  return 'shiShuCard'
}

/** 卡图文件名；加载失败时用通用牌面兜底 */
export function cardArtFile(card: Card, loadFailed = false): string {
  if (loadFailed) return card.type === 'Attack' ? 'moRen' : 'shiShuCard'
  return CARD_IMAGE_PINYIN_MAP[card.name] ?? fallbackPinyinByCard(card)
}

export function cardArtSrc(card: Card, loadFailed = false): string {
  return `/image/${cardArtFile(card, loadFailed)}.png`
}
