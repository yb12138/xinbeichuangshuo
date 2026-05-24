<script setup lang="ts">
import { storeToRefs } from 'pinia'
import { computed } from 'vue'
import { useBattleFxStore } from '../stores/battlefx.store'
import { useSnapshotStore } from '../stores/snapshot.store'
import RoseCourtyardIcon from './StatusIcons/RoseCourtyardIcon.vue'

const battleFxStore = useBattleFxStore()
const snapshotStore = useSnapshotStore()
const { combatCue, skillAnnouncements } = storeToRefs(battleFxStore)
const { players } = storeToRefs(snapshotStore)

const featuredSkillAnnouncement = computed(() =>
  skillAnnouncements.value.find((item) => item.phase === 'featured') ?? null
)

const settledSkillAnnouncements = computed(() =>
  skillAnnouncements.value
    .filter((item) => item.phase === 'settled')
    .slice(-3)
)

const duelPhaseLabel = computed(() => {
  const phase = combatCue.value?.phase
  if (phase === 'attack') return '攻击'
  if (phase === 'defend') return '防御'
  if (phase === 'counter') return '应战'
  if (phase === 'shield') return '圣盾'
  return '命中'
})

const hasRoseCourtyard = computed(() => {
  return Object.values(players.value).some(p =>
    p.field?.some(fc => fc.mode === 'Effect' && fc.effect === 'RoseCourtyard')
  )
})
</script>

<template>
  <div class="battle-zone battle-zone-shell min-h-[90px]">
    <div class="battle-content">
      <div v-if="combatCue" :key="combatCue.id" class="duel-center-only">
        <div class="duel-effect" :class="`phase-${combatCue.phase}`">{{ duelPhaseLabel }}</div>
      </div>

      <div v-else-if="hasRoseCourtyard" class="rose-courtyard-display">
        <div class="rose-courtyard-icon-wrap">
          <RoseCourtyardIcon />
        </div>
        <span class="rose-courtyard-label">血蔷薇庭院</span>
        <span class="rose-courtyard-hint">玩家无法使用治疗抵消伤害</span>
      </div>

      <div
        v-else
        class="battle-idle-label"
      >
        <span class="battle-idle-icon">⚔</span>
        <span>战区</span>
      </div>
    </div>

    <Transition name="skill-plaque-featured">
      <div
        v-if="featuredSkillAnnouncement"
        :key="featuredSkillAnnouncement.id"
        class="skill-plaque skill-plaque--featured"
      >
        <div class="skill-plaque__glow"></div>
        <div class="skill-plaque__body">
          <div class="skill-plaque__eyebrow">{{ featuredSkillAnnouncement.actorName }} 发动技能</div>
          <div class="skill-plaque__title">{{ featuredSkillAnnouncement.skillName }}</div>
          <div class="skill-plaque__effect">{{ featuredSkillAnnouncement.effectText }}</div>
        </div>
      </div>
    </Transition>

    <TransitionGroup name="skill-plaque-settled" tag="div" class="skill-plaque-stack">
      <div
        v-for="item in settledSkillAnnouncements"
        :key="item.id"
        class="skill-plaque skill-plaque--settled"
      >
        <div class="skill-plaque__body">
          <div class="skill-plaque__eyebrow">{{ item.actorName }}</div>
          <div class="skill-plaque__title">{{ item.skillName }}</div>
          <div class="skill-plaque__effect">{{ item.effectText }}</div>
        </div>
      </div>
    </TransitionGroup>
  </div>
</template>

<style scoped>
.battle-zone-shell {
  width: 100%;
  height: 100%;
  min-height: 0;
  position: relative;
  overflow: hidden;
  border-radius: 14px;
  border: 1px solid rgba(100, 140, 190, 0.12);
  background: radial-gradient(ellipse 80% 70% at 50% 50%, rgba(30, 60, 100, 0.12), transparent 70%);
  padding: 10px 10px 8px;
}

.skill-plaque {
  position: absolute;
  pointer-events: none;
  color: #f8ecd1;
  isolation: isolate;
}

.skill-plaque::before {
  content: '';
  position: absolute;
  inset: 0;
  z-index: -1;
  background-image:
    linear-gradient(90deg, rgba(4, 10, 20, 0.28), rgba(4, 10, 20, 0.06), rgba(4, 10, 20, 0.28)),
    url('/assets/ui/skill-plaque-bg.png');
  background-size: cover;
  background-position: center;
  border-radius: inherit;
}

.skill-plaque__body {
  position: relative;
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
  overflow: hidden;
}

.skill-plaque__eyebrow {
  max-width: 100%;
  color: rgba(244, 219, 165, 0.86);
  font-size: 11px;
  font-weight: 800;
  line-height: 1.15;
  text-shadow: 0 1px 3px rgba(0, 0, 0, 0.85);
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}

.skill-plaque__title {
  max-width: 100%;
  color: #fff2c8;
  font-weight: 900;
  line-height: 1.05;
  text-shadow:
    0 2px 4px rgba(0, 0, 0, 0.88),
    0 0 18px rgba(235, 179, 88, 0.42);
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}

.skill-plaque__effect {
  max-width: 100%;
  color: rgba(235, 240, 246, 0.9);
  line-height: 1.25;
  text-shadow: 0 1px 3px rgba(0, 0, 0, 0.86);
  display: -webkit-box;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.skill-plaque--featured {
  left: 50%;
  top: 50%;
  width: min(520px, calc(100% - 52px));
  min-height: 126px;
  padding: 18px 46px;
  border-radius: 12px;
  transform: translate(-50%, -50%);
  z-index: 16;
  filter: drop-shadow(0 20px 34px rgba(0, 0, 0, 0.44));
}

.skill-plaque--featured::before {
  box-shadow:
    inset 0 0 0 1px rgba(251, 231, 176, 0.42),
    inset 0 0 36px rgba(255, 208, 116, 0.16),
    0 0 0 1px rgba(24, 12, 4, 0.42);
}

.skill-plaque--featured .skill-plaque__body {
  gap: 8px;
}

.skill-plaque--featured .skill-plaque__eyebrow {
  font-size: clamp(11px, 1.35vw, 13px);
}

.skill-plaque--featured .skill-plaque__title {
  font-size: clamp(25px, 4.2vw, 42px);
}

.skill-plaque--featured .skill-plaque__effect {
  width: min(390px, 100%);
  font-size: clamp(12px, 1.55vw, 15px);
  -webkit-line-clamp: 2;
}

.skill-plaque__glow {
  position: absolute;
  inset: -18px 16%;
  z-index: -2;
  border-radius: 999px;
  background: radial-gradient(ellipse at center, rgba(237, 181, 91, 0.3), rgba(237, 181, 91, 0));
  filter: blur(10px);
  animation: skillPlaqueGlow 1.1s ease-out both;
}

.skill-plaque-stack {
  position: absolute;
  right: 12px;
  top: 12px;
  z-index: 12;
  width: min(300px, calc(100% - 24px));
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 8px;
  pointer-events: none;
}

.skill-plaque--settled {
  position: relative;
  width: min(284px, 100%);
  min-height: 58px;
  padding: 8px 14px;
  border-radius: 8px;
  filter: drop-shadow(0 10px 16px rgba(0, 0, 0, 0.36));
}

.skill-plaque--settled::before {
  opacity: 0.92;
  box-shadow:
    inset 0 0 0 1px rgba(236, 211, 154, 0.28),
    inset 0 0 22px rgba(223, 168, 90, 0.08);
}

.skill-plaque--settled .skill-plaque__body {
  align-items: flex-start;
  text-align: left;
  gap: 2px;
}

.skill-plaque--settled .skill-plaque__eyebrow {
  max-width: 100%;
  font-size: 10px;
}

.skill-plaque--settled .skill-plaque__title {
  max-width: 100%;
  font-size: 15px;
}

.skill-plaque--settled .skill-plaque__effect {
  max-width: 100%;
  color: rgba(229, 236, 243, 0.78);
  font-size: 11px;
  -webkit-line-clamp: 1;
}

.skill-plaque-featured-enter-active,
.skill-plaque-featured-leave-active {
  transition:
    opacity 0.28s ease,
    transform 0.28s cubic-bezier(0.2, 0.8, 0.2, 1),
    filter 0.28s ease;
}

.skill-plaque-featured-enter-from,
.skill-plaque-featured-leave-to {
  opacity: 0;
  transform: translate(-50%, -50%) scale(0.86);
  filter: blur(2px) drop-shadow(0 10px 18px rgba(0, 0, 0, 0.3));
}

.skill-plaque-settled-enter-active,
.skill-plaque-settled-leave-active {
  transition:
    opacity 0.24s ease,
    transform 0.24s cubic-bezier(0.2, 0.8, 0.2, 1);
}

.skill-plaque-settled-enter-from,
.skill-plaque-settled-leave-to {
  opacity: 0;
  transform: translateX(16px) scale(0.96);
}

@keyframes skillPlaqueGlow {
  0% { opacity: 0; transform: scaleX(0.7); }
  30% { opacity: 1; transform: scaleX(1.05); }
  100% { opacity: 0.72; transform: scaleX(1); }
}

.battle-zone-shell > * {
  position: relative;
  z-index: 1;
}

.battle-zone-shell::before {
  content: '';
  position: absolute;
  width: min(320px, 80%);
  height: min(320px, 80%);
  left: 50%;
  top: 48%;
  transform: translate(-50%, -50%);
  border-radius: 999px;
  border: 1px solid rgba(118, 157, 229, 0.12);
  box-shadow:
    inset 0 0 0 1px rgba(118, 157, 229, 0.06),
    0 0 0 40px rgba(118, 157, 229, 0.03),
    0 0 0 80px rgba(118, 157, 229, 0.015),
    inset 0 0 60px rgba(100, 160, 230, 0.06);
  background: radial-gradient(circle, rgba(80, 140, 220, 0.04), transparent 70%);
  pointer-events: none;
  z-index: 0;
  animation: battleRingBreath 6s ease-in-out infinite;
}

.battle-zone-shell::after {
  content: '';
  position: absolute;
  width: min(180px, 50%);
  height: min(180px, 50%);
  left: 50%;
  top: 48%;
  transform: translate(-50%, -50%) rotate(18deg);
  border: 1px solid rgba(152, 184, 245, 0.08);
  border-radius: 12px;
  background: radial-gradient(circle, rgba(130, 170, 240, 0.03), transparent 60%);
  pointer-events: none;
  z-index: 0;
  animation: battleSquareBreath 8s ease-in-out infinite reverse;
}

@keyframes battleRingBreath {
  0%, 100% { opacity: 0.7; transform: translate(-50%, -50%) scale(1); }
  50% { opacity: 1; transform: translate(-50%, -50%) scale(1.03); }
}

@keyframes battleSquareBreath {
  0%, 100% { opacity: 0.6; transform: translate(-50%, -50%) rotate(18deg) scale(1); }
  50% { opacity: 0.9; transform: translate(-50%, -50%) rotate(22deg) scale(1.04); }
}

.battle-content {
  height: 100%;
  min-height: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}

.duel-center-only {
  width: min(360px, 100%);
  min-height: 90px;
  display: flex;
  align-items: center;
  justify-content: center;
  animation: duelShow 0.2s ease-out;
}

.duel-effect {
  min-width: 86px;
  text-align: center;
  font-size: 12px;
  font-weight: 700;
  border-radius: 999px;
  padding: 6px 12px;
  color: #e9f3fb;
  border: 1px solid rgba(141, 172, 190, 0.48);
  background: rgba(17, 38, 58, 0.84);
  box-shadow: 0 6px 16px rgba(2, 8, 17, 0.34);
  animation: clashPulse 0.45s ease-out;
}

.duel-effect.phase-attack,
.duel-effect.phase-counter {
  color: #ffe3dd;
  border-color: rgba(220, 123, 112, 0.66);
  background: rgba(98, 33, 30, 0.84);
}

.duel-effect.phase-defend {
  color: #d9f0ff;
  border-color: rgba(120, 188, 228, 0.62);
  background: rgba(19, 60, 92, 0.84);
}

.duel-effect.phase-shield {
  color: #f5f8ff;
  border-color: rgba(183, 216, 255, 0.9);
  background: rgba(28, 52, 86, 0.9);
  box-shadow:
    0 0 0 1px rgba(207, 230, 255, 0.32),
    0 6px 22px rgba(26, 80, 146, 0.45),
    0 0 26px rgba(148, 199, 255, 0.35);
}

.duel-effect.phase-take {
  color: #fbe9c5;
  border-color: rgba(227, 192, 132, 0.6);
  background: rgba(90, 63, 34, 0.82);
}

@keyframes duelShow {
  from { opacity: 0; transform: translateY(6px); }
  to { opacity: 1; transform: translateY(0); }
}

@keyframes clashPulse {
  0% { transform: scale(0.88); opacity: 0.55; }
  50% { transform: scale(1.04); opacity: 1; }
  100% { transform: scale(1); opacity: 1; }
}

@media (min-width: 1600px) {
  .battle-zone-shell {
    padding: 12px 12px 10px;
  }

  .duel-center-only {
    width: min(420px, 100%);
    min-height: 106px;
  }

  .duel-effect {
    min-width: 96px;
    font-size: 13px;
    padding: 7px 14px;
  }
}

@media (min-width: 2000px) {
  .battle-zone-shell {
    padding: 14px 14px 12px;
  }

  .duel-center-only {
    width: min(480px, 100%);
    min-height: 118px;
  }

  .duel-effect {
    min-width: 106px;
    font-size: 14px;
    padding: 8px 16px;
  }
}

.battle-idle-label {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  color: rgba(130, 160, 190, 0.35);
  font-size: 13px;
  font-weight: 600;
  letter-spacing: 0.2em;
  user-select: none;
}

.battle-idle-icon {
  font-size: 22px;
  opacity: 0.4;
}

/* 血蔷薇庭院 */
.rose-courtyard-display {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  animation: duelShow 0.3s ease-out;
}

.rose-courtyard-icon-wrap {
  width: 64px;
  height: 64px;
  background: rgba(0, 0, 0, 0.65);
  border-radius: 14px;
  padding: 6px;
  backdrop-filter: blur(6px);
  border: 2px solid rgba(251, 113, 133, 0.45);
  box-shadow: 0 4px 16px rgba(190, 18, 60, 0.35);
  animation: roseCourtyardPulse 3s ease-in-out infinite;
}

.rose-courtyard-label {
  font-size: 14px;
  font-weight: bold;
  color: #fb7185;
  text-shadow: 0 1px 3px rgba(0, 0, 0, 0.8);
}

.rose-courtyard-hint {
  font-size: 10px;
  color: #fca5a5;
  background: rgba(0, 0, 0, 0.7);
  padding: 2px 10px;
  border-radius: 10px;
  border: 1px solid rgba(251, 113, 133, 0.25);
  text-align: center;
  max-width: 180px;
}

@keyframes roseCourtyardPulse {
  0%, 100% {
    box-shadow: 0 4px 16px rgba(190, 18, 60, 0.35);
  }
  50% {
    box-shadow: 0 6px 22px rgba(190, 18, 60, 0.55);
  }
}

@media (max-width: 640px) {
  .battle-zone-shell {
    padding: 8px 8px 6px;
  }

  .duel-center-only {
    width: min(320px, 100%);
    min-height: 78px;
  }

  .duel-effect {
    min-width: 72px;
    font-size: 11px;
    padding: 5px 9px;
  }
}
</style>
