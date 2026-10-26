<script setup lang="ts">
/**
 * 对战标记徽章：场上状态（圆形）、盖牌（竖版牌背）、指示物（菱形）。
 * 外形与金属边框统一绘制，内部剪影取自 markGlyphs；颜色只来自 constants/battleMarks。
 */
import { computed } from 'vue'
import { COVER_MARKS, STATUS_MARKS, TOKEN_MARKS, statusRingColor } from '../../constants/battleMarks'
import { FALLBACK_GLYPH, MARK_GLYPHS } from './markGlyphs'

const props = defineProps<{
  kind: 'status' | 'cover' | 'token'
  /** status/cover = FieldCard.effect；token = 指示物键。 */
  markKey: string
  /** 叠加层数 / 盖牌数量；>1 时显示角标。 */
  count?: number
}>()

const view = computed(() => {
  if (props.kind === 'status') {
    const def = STATUS_MARKS[props.markKey]
    return {
      label: def?.label ?? props.markKey,
      glyph: (def && MARK_GLYPHS[def.glyph]) || FALLBACK_GLYPH,
      ring: def ? statusRingColor(def) : '#b9c1cd',
    }
  }
  if (props.kind === 'cover') {
    const def = COVER_MARKS[props.markKey]
    return {
      label: def?.label ?? '盖牌',
      glyph: (def && MARK_GLYPHS[def.glyph]) || FALLBACK_GLYPH,
      ring: def?.accent ?? '#b9c1cd',
    }
  }
  const def = TOKEN_MARKS[props.markKey]
  return {
    label: def?.label ?? props.markKey,
    glyph: (def && MARK_GLYPHS[def.glyph]) || FALLBACK_GLYPH,
    ring: def?.accent ?? '#b9c1cd',
  }
})

const showCount = computed(() => typeof props.count === 'number' && props.count > 1)
</script>

<template>
  <span
    class="battle-mark"
    :class="`battle-mark--${kind}`"
    :style="{ '--mark-ring': view.ring }"
    role="img"
    :aria-label="showCount ? `${view.label} ×${count}` : view.label"
  >
    <svg class="battle-mark__svg" :viewBox="kind === 'cover' ? '0 0 24 32' : '0 0 24 24'" aria-hidden="true">
      <!-- 外框 -->
      <circle v-if="kind === 'status'" class="battle-mark__frame" cx="12" cy="12" r="10.6" />
      <rect v-else-if="kind === 'cover'" class="battle-mark__frame" x="1.4" y="1.4" width="21.2" height="29.2" rx="3" />
      <path v-else class="battle-mark__frame" d="M12 1.2L22.8 12 12 22.8 1.2 12z" />
      <!-- 内圈细线 -->
      <circle v-if="kind === 'status'" class="battle-mark__inner" cx="12" cy="12" r="8.6" />
      <rect v-else-if="kind === 'cover'" class="battle-mark__inner" x="3.4" y="3.4" width="17.2" height="25.2" rx="2" />
      <!-- 剪影 -->
      <g :transform="kind === 'cover' ? 'translate(3.6 6.8) scale(0.7)' : kind === 'token' ? 'translate(4.8 4.8) scale(0.6)' : 'translate(3.6 3.6) scale(0.7)'">
        <path class="battle-mark__glyph" :d="view.glyph" />
      </g>
    </svg>
    <span v-if="showCount" class="battle-mark__count">{{ count }}</span>
  </span>
</template>

<style scoped>
.battle-mark {
  position: relative;
  display: inline-flex;
  width: 100%;
  height: 100%;
  line-height: 0;
}
.battle-mark__svg {
  width: 100%;
  height: 100%;
  overflow: visible;
  filter: drop-shadow(0 1px 1.5px rgba(0, 0, 0, 0.7));
}
.battle-mark__frame {
  fill: var(--bt-bg-0, #070b14);
  stroke: var(--mark-ring);
  stroke-width: 1.2;
}
.battle-mark__inner {
  fill: none;
  stroke: color-mix(in srgb, var(--mark-ring) 45%, transparent);
  stroke-width: 0.6;
}
.battle-mark__glyph {
  fill: var(--mark-ring);
  filter: drop-shadow(0 0 1.2px color-mix(in srgb, var(--mark-ring) 70%, transparent));
}
.battle-mark--cover .battle-mark__frame {
  fill: color-mix(in srgb, var(--mark-ring) 12%, var(--bt-bg-0, #070b14));
}
.battle-mark__count {
  position: absolute;
  right: -3px;
  bottom: -3px;
  min-width: 13px;
  height: 13px;
  padding: 0 3px;
  border-radius: 999px;
  font-size: 9px;
  font-weight: 800;
  line-height: 13px;
  text-align: center;
  color: var(--bt-text, #e8ecf4);
  background: var(--bt-bg-0, #070b14);
  border: 1px solid var(--mark-ring);
}
</style>
