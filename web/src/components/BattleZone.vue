<script setup lang="ts">
import { computed } from 'vue'
import { useGameStore } from '../stores/gameStore'
import CardComponent from './CardComponent.vue'

const store = useGameStore()

const actionLabels: Record<string, string> = {
  attack: '攻击',
  magic: '法术',
  discard: '弃牌',
  defend: '防御',
  counter: '应战'
}

const duelAttacker = computed(() => {
  const cue = store.combatCue
  if (!cue) return null
  return store.players[cue.attackerId] || null
})
const duelTarget = computed(() => {
  const cue = store.combatCue
  if (!cue) return null
  return store.players[cue.targetId] || null
})
const duelPhaseLabel = computed(() => {
  const phase = store.combatCue?.phase
  if (phase === 'attack') return '攻击'
  if (phase === 'defend') return '防御'
  if (phase === 'counter') return '应战'
  return '承受'
})
function charImage(role?: string) {
  if (!role) return ''
  return `/characters/${role}.png`
}
</script>

<template>
  <div class="battle-zone battle-zone-shell min-h-[90px]">
    <div class="battle-content">
      <div v-if="duelAttacker && duelTarget && store.combatCue" :key="store.combatCue.id" class="duel-stage">
        <div class="duel-side attacker" :class="{ 'pulse-attack': store.combatCue.phase === 'attack' || store.combatCue.phase === 'counter' }">
          <img :src="charImage(duelAttacker.role)" :alt="duelAttacker.name" class="duel-portrait" />
          <div class="duel-name">{{ duelAttacker.name }}</div>
        </div>
        <div class="duel-center">
          <div class="duel-effect" :class="`phase-${store.combatCue.phase}`">{{ duelPhaseLabel }}</div>
        </div>
        <div class="duel-side target" :class="{ 'pulse-defend': store.combatCue.phase === 'defend' || store.combatCue.phase === 'take' }">
          <img :src="charImage(duelTarget.role)" :alt="duelTarget.name" class="duel-portrait" />
          <div class="duel-name">{{ duelTarget.name }}</div>
        </div>
      </div>

      <div class="fly-zone">
        <template v-for="batch in store.flyingCards" :key="batch.id">
          <div
            v-for="(card, cidx) in batch.cards"
            :key="`${batch.id}-${cidx}`"
            class="flex flex-col items-center card-fly-enter"
          >
            <CardComponent :card="card" small class="card-animate" :face-down="batch.hidden" />
            <div v-if="cidx === 0 && batch.cards.length > 0" class="text-[9px] text-gray-400 mt-0.5">
              {{ batch.playerName }} {{ actionLabels[batch.actionType] || batch.actionType }}
            </div>
          </div>
        </template>

        <div
          v-if="store.flyingCards.length === 0 && !(duelAttacker && duelTarget && store.combatCue)"
          class="battle-idle-label"
        >
          <span class="battle-idle-icon">⚔</span>
          <span>战区</span>
        </div>
      </div>
    </div>
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

.fly-zone {
  min-height: 0;
  width: 100%;
  flex: 1 1 auto;
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  justify-content: center;
  align-items: center;
}

.card-fly-enter {
  animation: cardAppear 0.36s cubic-bezier(0.22, 0.61, 0.36, 1);
}

@keyframes cardAppear {
  from {
    opacity: 0;
    transform: scale(0.75) translateY(14px);
  }
  to {
    opacity: 1;
    transform: scale(1) translateY(0);
  }
}

.card-animate {
  transition: transform 0.24s ease, filter 0.24s ease;
}
.card-animate:hover {
  transform: translateY(-2px) scale(1.03);
  filter: drop-shadow(0 6px 14px rgba(5, 20, 33, 0.4));
}

.duel-stage {
  width: min(380px, 100%);
  min-height: 90px;
  display: grid;
  grid-template-columns: 1fr auto 1fr;
  align-items: center;
  gap: 8px;
  margin-bottom: 6px;
  animation: duelShow 0.2s ease-out;
}

@media (max-width: 900px) {
  .fly-zone {
    gap: 6px;
  }
}

.duel-side {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.duel-portrait {
  width: 50px;
  height: 62px;
  object-fit: cover;
  border-radius: 8px;
  border: 1px solid rgba(150, 182, 201, 0.62);
  box-shadow: 0 8px 18px rgba(2, 8, 18, 0.44);
}

.duel-name {
  margin-top: 4px;
  font-size: 10px;
  color: #dceaf5;
  text-align: center;
  max-width: 90px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.duel-center {
  display: flex;
  align-items: center;
  justify-content: center;
}

.duel-effect {
  min-width: 78px;
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

.duel-effect.phase-take {
  color: #fbe9c5;
  border-color: rgba(227, 192, 132, 0.6);
  background: rgba(90, 63, 34, 0.82);
}

.pulse-attack {
  animation: hitPush 0.26s ease-out;
}

.pulse-defend {
  animation: shieldPulse 0.28s ease-out;
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

@keyframes hitPush {
  0% { transform: translateX(0) scale(1); }
  45% { transform: translateX(4px) scale(1.03); }
  100% { transform: translateX(0) scale(1); }
}

@keyframes shieldPulse {
  0% { transform: scale(1); filter: brightness(1); }
  50% { transform: scale(1.03); filter: brightness(1.15); }
  100% { transform: scale(1); filter: brightness(1); }
}

@media (min-width: 1600px) {
  .battle-zone-shell {
    padding: 12px 12px 10px;
  }

  .duel-stage {
    width: min(460px, 100%);
    min-height: 106px;
  }

  .duel-portrait {
    width: 58px;
    height: 72px;
  }
}

@media (min-width: 2000px) {
  .battle-zone-shell {
    padding: 14px 14px 12px;
  }

  .duel-stage {
    width: min(520px, 100%);
    min-height: 118px;
  }

  .duel-portrait {
    width: 66px;
    height: 82px;
  }

  .duel-effect {
    min-width: 92px;
    font-size: 13px;
    padding: 7px 14px;
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

@media (max-width: 640px) {
  .battle-zone-shell {
    padding: 8px 8px 6px;
  }

  .duel-stage {
    width: min(320px, 100%);
    min-height: 78px;
  }

  .duel-portrait {
    width: 42px;
    height: 52px;
  }

  .duel-effect {
    min-width: 64px;
    font-size: 11px;
    padding: 5px 9px;
  }

  .duel-name {
    font-size: 9px;
    margin-top: 3px;
  }
}
</style>
