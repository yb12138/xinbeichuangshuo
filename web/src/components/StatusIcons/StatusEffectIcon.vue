<script setup lang="ts">
import { computed } from 'vue'
import ShieldIcon from './ShieldIcon.vue'
import PoisonIcon from './PoisonIcon.vue'
import WeakIcon from './WeakIcon.vue'
import SealIcon from './SealIcon.vue'
import FiveElementsBindIcon from './FiveElementsBindIcon.vue'
import StealthIcon from './StealthIcon.vue'
import PowerBlessingIcon from './PowerBlessingIcon.vue'
import SwiftBlessingIcon from './SwiftBlessingIcon.vue'
import HeroTauntIcon from './HeroTauntIcon.vue'
import BloodSharedLifeIcon from './BloodSharedLifeIcon.vue'
import RoseCourtyardIcon from './RoseCourtyardIcon.vue'
import BardEternalMovementIcon from './BardEternalMovementIcon.vue'

const props = defineProps<{
  effect: string
  count?: number
}>()

const iconComponent = computed(() => {
  switch (props.effect) {
    case 'Shield':
      return ShieldIcon
    case 'Poison':
      return PoisonIcon
    case 'Weak':
      return WeakIcon
    case 'SealFire':
    case 'SealWater':
    case 'SealEarth':
    case 'SealWind':
    case 'SealThunder':
      return SealIcon
    case 'FiveElementsBind':
      return FiveElementsBindIcon
    case 'Stealth':
      return StealthIcon
    case 'PowerBlessing':
      return PowerBlessingIcon
    case 'SwiftBlessing':
      return SwiftBlessingIcon
    case 'HeroTaunt':
      return HeroTauntIcon
    case 'BloodSharedLife':
      return BloodSharedLifeIcon
    case 'RoseCourtyard':
      return RoseCourtyardIcon
    case 'BardEternalMovement':
      return BardEternalMovementIcon
    default:
      return null
  }
})

const sealElement = computed(() => {
  const elementMap: Record<string, 'fire' | 'water' | 'earth' | 'wind' | 'thunder'> = {
    SealFire: 'fire',
    SealWater: 'water',
    SealEarth: 'earth',
    SealWind: 'wind',
    SealThunder: 'thunder',
  }
  return elementMap[props.effect]
})

const showCount = computed(() => {
  return props.count !== undefined && props.count > 1
})
</script>

<template>
  <div class="status-effect-icon">
    <component
      :is="iconComponent"
      v-if="iconComponent"
      :element="sealElement"
      class="icon-svg"
    />
    <div v-if="showCount" class="effect-count">
      {{ count }}
    </div>
  </div>
</template>

<style scoped>
.status-effect-icon {
  position: relative;
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.icon-svg {
  width: 100%;
  height: 100%;
}

.effect-count {
  position: absolute;
  bottom: -2px;
  right: -2px;
  background: rgba(0, 0, 0, 0.8);
  color: #fff;
  font-size: 12px;
  font-weight: bold;
  padding: 2px 5px;
  border-radius: 10px;
  min-width: 18px;
  text-align: center;
  border: 1px solid rgba(255, 255, 255, 0.3);
  line-height: 1;
}
</style>
