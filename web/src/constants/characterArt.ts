/**
 * 角色立绘路径。
 * - full：完整卡牌扫描图（含原卡框与技能文字），用于大厅、悬停大图、技能弹窗。
 * - art：裁掉卡框与文字区后的插画（public/characters/art），用于对战座位等小尺寸位置。
 */
export function characterFullCardSrc(roleId?: string | null): string {
  const role = String(roleId || '').trim()
  return role ? `/characters/${role}.png` : ''
}

export function characterArtSrc(roleId?: string | null): string {
  const role = String(roleId || '').trim()
  return role ? `/characters/art/${role}.webp` : ''
}
