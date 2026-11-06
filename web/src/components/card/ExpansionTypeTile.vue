<script setup lang="ts">
/** 扩展区分组标签：圆角方块 + 种类剪影，边框与发光取该种类色调。 */
import { computed } from 'vue'
import { expansionGroupView } from '../../constants/expansionZone'

const props = defineProps<{
  /** FieldCard.effect，或 EXCLUSIVE_GROUP_KEY。 */
  groupKey: string
  /** 右下角数量；不传则不显示。 */
  count?: number
  active?: boolean
}>()

const view = computed(() => expansionGroupView(props.groupKey))
</script>

<template>
  <span
    class="xp-tile"
    :class="{ 'xp-tile--active': active }"
    :style="{ '--xp-tint': view.tint }"
    role="img"
    :aria-label="count !== undefined ? `${view.label} ×${count}` : view.label"
  >
    <svg class="xp-tile__glyph" viewBox="0 0 24 24" aria-hidden="true">
      <path :d="view.glyph" />
    </svg>
    <span v-if="count !== undefined" class="xp-tile__count" aria-hidden="true">{{ count }}</span>
  </span>
</template>

<style scoped>
.xp-tile {
  --xp-tint: #9fb3cf;
  position: relative;
  display: inline-grid;
  place-items: center;
  width: var(--xp-tile-size, 24px);
  height: var(--xp-tile-size, 24px);
  flex: none;
  border-radius: 6px;
  background:
    radial-gradient(80% 80% at 50% 30%, color-mix(in srgb, var(--xp-tint) 22%, transparent), transparent 70%),
    linear-gradient(180deg, #0d1a2e, #060c18);
  box-shadow:
    inset 0 0 0 1px color-mix(in srgb, var(--xp-tint) 60%, transparent),
    inset 0 1px 0 rgba(255, 255, 255, 0.08),
    0 0 0 1px rgba(0, 0, 0, 0.6);
}

.xp-tile--active {
  box-shadow:
    inset 0 0 0 1px var(--xp-tint),
    inset 0 0 8px color-mix(in srgb, var(--xp-tint) 35%, transparent),
    0 0 0 1px rgba(0, 0, 0, 0.6),
    0 0 10px color-mix(in srgb, var(--xp-tint) 45%, transparent);
}

.xp-tile__glyph {
  width: 62%;
  height: 62%;
  fill: color-mix(in srgb, var(--xp-tint) 70%, #ffffff);
  filter: drop-shadow(0 0 2px color-mix(in srgb, var(--xp-tint) 80%, transparent));
}

.xp-tile__count {
  position: absolute;
  right: -6px;
  bottom: -5px;
  min-width: 15px;
  height: 15px;
  padding: 0 3px;
  border-radius: 999px;
  background: #050a14;
  box-shadow: 0 0 0 1px color-mix(in srgb, var(--xp-tint) 70%, transparent);
  color: #e8eef8;
  font-size: 9px;
  font-weight: 800;
  line-height: 15px;
  text-align: center;
  font-variant-numeric: tabular-nums;
}
</style>
