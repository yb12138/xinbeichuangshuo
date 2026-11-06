/**
 * 角色形态的展示名与座位芯片配色。key = PlayerView.form（引擎 model.Form* 常量）。
 * 座位形态徽章、形态变化动画与技能结果文案（"进入烈焰形态"）共用。
 * accent = 座位常驻形态光环与徽章的描边色。
 */
export const FORM_DISPLAY: Record<string, { label: string; cls: string; accent: string }> = {
  heroic_form: { label: '英灵形态', cls: 'bg-amber-800/70 text-amber-100 border-amber-500/40', accent: '#f0b54a' },
  arbiter_judgment_form: { label: '审判形态', cls: 'bg-violet-800/70 text-violet-100 border-violet-500/40', accent: '#a98bff' },
  elf_archer_ritual_form: { label: '祝福形态', cls: 'bg-emerald-800/70 text-emerald-100 border-emerald-500/40', accent: '#4fd39a' },
  magic_swordsman_shadow_form: { label: '暗影形态', cls: 'bg-slate-800/80 text-slate-100 border-slate-500/40', accent: '#8fa3c0' },
  war_homunculus_burst_form: { label: '蓄势形态', cls: 'bg-purple-900/70 text-purple-100 border-purple-500/40', accent: '#b07cff' },
  prayer_master_prayer_form: { label: '祈祷形态', cls: 'bg-amber-900/70 text-amber-100 border-amber-500/40', accent: '#f0d27a' },
  crimson_knight_hot_blooded_form: { label: '热血形态', cls: 'bg-orange-900/70 text-orange-100 border-orange-500/40', accent: '#ff7a4d' },
  onmyoji_shikigami_form: { label: '式神形态', cls: 'bg-sky-900/70 text-sky-100 border-sky-500/40', accent: '#4fb0ff' },
  blaze_witch_flame_form: { label: '烈焰形态', cls: 'bg-orange-900/70 text-orange-100 border-orange-500/40', accent: '#ff7a4d' },
  beast_samurai_iaijutsu_form: { label: '居合形态', cls: 'bg-zinc-900/75 text-zinc-100 border-zinc-500/40', accent: '#c8c8d0' },
  holy_bow_holy_glory_form: { label: '圣煌形态', cls: 'bg-sky-900/70 text-sky-100 border-sky-500/40', accent: '#8fd0ff' },
  magic_lance_phantom_form: { label: '幻影形态', cls: 'bg-slate-900/75 text-slate-100 border-slate-500/40', accent: '#8fa3c0' },
  bard_eternal_prisoner_form: { label: '囚徒形态', cls: 'bg-purple-900/70 text-purple-100 border-purple-500/40', accent: '#b07cff' },
  hero_exhaustion_form: { label: '精疲力竭', cls: 'bg-amber-900/70 text-amber-100 border-amber-500/40', accent: '#d8a85a' },
  fighter_hundred_dragon_form: { label: '幻龙拳形态', cls: 'bg-red-900/70 text-red-100 border-red-500/40', accent: '#ff5d5d' },
  moon_goddess_dark_moon_form: { label: '暗月形态', cls: 'bg-slate-900/75 text-slate-100 border-slate-500/40', accent: '#a07ce0' },
  blood_witch_bleeding_form: { label: '流血形态', cls: 'bg-red-900/70 text-red-100 border-red-500/40', accent: '#e0435a' },
}

/** 形态展示名；未登记的形态返回空串。 */
export function formLabel(form: string | undefined): string {
  return (form && FORM_DISPLAY[form]?.label) || ''
}
