<script setup lang="ts">
import { computed } from 'vue'
import { useGameStore } from '../stores/gameStore'

const store = useGameStore()

const currentEffect = computed(() => {
  if (store.damageNotifications.length === 0) return null
  return store.damageNotifications[0]
})

const isVisible = computed(() => currentEffect.value !== null)

function getDamageTypeLabel(type: string): string {
  const labels: Record<string, string> = {
    Attack: '攻击伤害',
    Magic: '法术伤害',
    magic: '法术伤害',
    Poison: '中毒伤害',
    poison: '中毒伤害',
    counter: '应战伤害',
    backlash: '反噬伤害',
    Backlash: '反噬伤害'
  }
  return labels[type] ?? '伤害'
}

function confirmDamage() {
  store.confirmDamageNotification()
}
</script>

<template>
  <Teleport to="body">
    <Transition name="damage-modal">
      <div 
        v-if="isVisible && currentEffect" 
        class="damage-overlay fixed inset-0 z-[2100] bg-black/70 backdrop-blur-sm"
      >
        <div class="damage-center">
          <div class="damage-notification-card rounded-2xl shadow-2xl max-w-xs w-full mx-4 overflow-hidden">
            <!-- 标题栏 -->
            <div class="damage-header px-4 py-3 flex items-center gap-3">
              <span class="text-3xl">⚔️</span>
              <h3 class="text-lg font-bold text-white">伤害结算</h3>
            </div>

            <!-- 内容区域 -->
            <div class="p-4 space-y-4 text-center">
              <!-- 伤害数值 -->
              <div class="damage-display">
                <div class="damage-number text-3xl font-black text-amber-400 mb-2">
                  -{{ currentEffect.damage }}
                </div>
                <div class="text-gray-400 text-base">
                  {{ getDamageTypeLabel(currentEffect.damageType) }}
                </div>
              </div>

              <!-- 受伤信息 -->
              <div class="damage-info space-y-2 text-gray-200">
                <div class="text-lg">
                  <span class="text-yellow-400 font-bold">{{ currentEffect.targetName }}</span>
                  <span> 受到了伤害</span>
                </div>
                <div class="text-sm text-gray-400">
                  需要摸 <span class="text-red-400 font-bold">{{ currentEffect.damage }}</span> 张牌
                </div>
              </div>

              <!-- 分隔线 -->
              <div class="damage-divider my-4"></div>

              <!-- 确认按钮 -->
              <button
                class="w-full py-4 rounded-xl font-bold text-lg btn-danger transition-all transform hover:scale-[1.02] active:scale-[0.98] shadow-lg shadow-red-900/50"
                @click="confirmDamage"
              >
                确认 (摸 {{ currentEffect.damage }} 张牌)
              </button>
            </div>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.damage-overlay {
  background:
    radial-gradient(320px 180px at 50% 44%, rgba(185, 93, 84, 0.22), transparent 70%),
    rgba(0, 0, 0, 0.72);
}

.damage-center {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: min(100%, 360px);
}

.damage-modal-enter-active,
.damage-modal-leave-active {
  transition: opacity 0.28s ease;
}

.damage-modal-enter-from,
.damage-modal-leave-to {
  opacity: 0;
}

.damage-modal-enter-from .damage-notification-card,
.damage-modal-leave-to .damage-notification-card {
  transform: scale(0.84) translateY(12px);
}

.damage-modal-enter-active .damage-notification-card,
.damage-modal-leave-active .damage-notification-card {
  transition: transform 0.3s ease;
}

.damage-number {
  text-shadow: 0 0 8px rgba(255, 164, 145, 0.52);
  animation: damageGlow 1.5s ease-in-out infinite;
}

@keyframes damageGlow {
  0%, 100% {
    text-shadow: 0 0 20px rgba(250, 118, 101, 0.76), 0 0 40px rgba(203, 57, 48, 0.5);
  }
  50% {
    text-shadow: 0 0 30px rgba(250, 118, 101, 0.96), 0 0 60px rgba(203, 57, 48, 0.72);
  }
}

.damage-notification-card {
  animation: cardShake 0.52s ease-out;
  border: 1px solid rgba(181, 122, 114, 0.44);
  background:
    linear-gradient(180deg, rgba(40, 17, 20, 0.92), rgba(23, 10, 13, 0.95)),
    url('/assets/ui/modal-aura.svg') center/cover no-repeat;
  box-shadow:
    inset 0 1px 0 rgba(255, 215, 208, 0.09),
    0 22px 50px rgba(0, 0, 0, 0.6);
}

.damage-header {
  background: linear-gradient(120deg, rgba(119, 42, 37, 0.88), rgba(84, 30, 27, 0.9));
  border-bottom: 1px solid rgba(220, 153, 144, 0.28);
}

.damage-divider {
  border-top: 1px solid rgba(184, 122, 114, 0.3);
}

@keyframes cardShake {
  0%, 100% { transform: translateX(0); }
  10%, 30%, 50%, 70%, 90% { transform: translateX(-5px); }
  20%, 40%, 60%, 80% { transform: translateX(5px); }
}
</style>
