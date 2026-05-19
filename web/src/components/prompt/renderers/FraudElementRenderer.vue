<script setup lang="ts">
type FraudElementOption = {
  id: string
  title: string
  glyph: string
  tone: string
}

defineProps<{
  visible: boolean
  title: string
  options: FraudElementOption[]
}>()

const emit = defineEmits<{
  select: [optionId: string]
}>()
</script>

<template>
  <Teleport to="body">
    <Transition name="prompt-fraud-side-pop">
      <div
        v-if="visible && options.length > 0"
        class="prompt-fraud-global-layer"
        data-testid="decision-overlay"
      >
        <div class="prompt-fraud-global-panel">
          <div class="prompt-fraud-dialog prompt-fraud-dialog--global">
            <div class="prompt-fraud-title">{{ title || '请选择本次攻击系别' }}</div>
            <div class="prompt-fraud-grid">
              <button
                v-for="option in options"
                :key="option.id"
                class="prompt-fraud-card"
                :class="option.tone"
                :title="option.title"
                :aria-label="option.title"
                :data-testid="`prompt-option-${option.id}`"
                @click="emit('select', option.id)"
              >
                <span class="prompt-fraud-card-title-banner">
                  <span class="prompt-fraud-card-title">{{ option.title }}</span>
                </span>
                <span class="prompt-fraud-card-medal">
                  <span>{{ option.glyph }}</span>
                </span>
                <span class="prompt-fraud-card-art">
                  <span class="prompt-fraud-card-glyph">{{ option.glyph }}</span>
                </span>
              </button>
            </div>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.prompt-fraud-global-layer {
  position: fixed;
  inset: 0;
  z-index: 60;
  pointer-events: auto;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(7, 14, 24, 0.38);
  backdrop-filter: blur(7px) saturate(0.98);
  -webkit-backdrop-filter: blur(7px) saturate(0.98);
  padding:
    max(16px, calc(var(--safe-top, 0px) + 8px))
    max(16px, calc(var(--safe-right, 0px) + 8px))
    max(16px, calc(var(--safe-bottom, 0px) + 8px))
    max(16px, calc(var(--safe-left, 0px) + 8px));
}

.prompt-fraud-global-panel {
  pointer-events: auto;
  width: min(860px, calc(100vw - 40px));
  max-height: calc(100vh - 40px);
  overflow: auto;
  border-radius: 14px;
  border: 1px solid rgba(146, 183, 207, 0.42);
  background:
    linear-gradient(180deg, rgba(9, 20, 34, 0.96), rgba(6, 14, 24, 0.97));
  box-shadow:
    0 18px 34px rgba(2, 8, 18, 0.52),
    inset 0 1px 0 rgba(236, 246, 254, 0.12);
  padding: 10px;
}

.prompt-fraud-dialog {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 4px 2px 2px;
}

.prompt-fraud-dialog--global {
  padding: 4px;
}

.prompt-fraud-title {
  font-size: 13px;
  line-height: 1.4;
  color: rgba(225, 238, 249, 0.96);
  text-align: center;
  letter-spacing: 0.01em;
}

.prompt-fraud-grid {
  display: grid;
  grid-template-columns: repeat(5, minmax(0, 1fr));
  gap: 8px;
}

.prompt-fraud-card {
  --fraud-edge-color: rgba(185, 152, 102, 0.78);
  --fraud-edge-glow: rgba(232, 191, 121, 0.38);
  --fraud-base-top: #2f2520;
  --fraud-base-bottom: #17120f;
  --fraud-ribbon-start: #8c5a2f;
  --fraud-ribbon-end: #60401f;
  --fraud-medal-bg: radial-gradient(circle at 32% 28%, #f1d79b, #c6924f 58%, #784d1d);
  --fraud-medal-fg: #fff7ea;
  --fraud-art-top: rgba(214, 174, 116, 0.3);
  --fraud-art-bottom: rgba(71, 45, 24, 0.82);
  display: flex;
  flex-direction: column;
  align-items: stretch;
  justify-content: flex-start;
  position: relative;
  min-height: 136px;
  border-radius: 10px;
  border: 2px solid var(--fraud-edge-color);
  background: linear-gradient(180deg, var(--fraud-base-top), var(--fraud-base-bottom));
  color: rgba(245, 250, 255, 0.97);
  box-shadow:
    0 8px 16px rgba(0, 0, 0, 0.55),
    0 0 12px var(--fraud-edge-glow),
    inset 0 0 0 1px rgba(255, 244, 214, 0.24),
    inset 0 0 10px rgba(255, 255, 255, 0.08);
  transition: transform 0.16s ease, filter 0.16s ease, border-color 0.16s ease, box-shadow 0.16s ease;
  overflow: hidden;
  text-align: center;
}

.prompt-fraud-card:hover:not(:disabled) {
  transform: translateY(-2px);
  border-color: rgba(255, 228, 170, 0.96);
  box-shadow:
    0 12px 22px rgba(0, 0, 0, 0.62),
    0 0 16px rgba(255, 214, 138, 0.44),
    inset 0 0 0 1px rgba(255, 247, 229, 0.38);
}

.prompt-fraud-card-title-banner {
  position: absolute;
  top: 6px;
  left: 16px;
  right: 16px;
  height: 20px;
  border-radius: 4px;
  border: 1px solid rgba(202, 184, 148, 0.88);
  background: linear-gradient(180deg, rgba(251, 249, 243, 0.96), rgba(222, 214, 197, 0.94));
  display: inline-flex;
  align-items: center;
  justify-content: center;
  z-index: 4;
}

.prompt-fraud-card-title {
  color: rgba(46, 34, 22, 0.94);
  font-size: 11px;
  font-weight: 820;
  letter-spacing: 0.04em;
  line-height: 1;
}

.prompt-fraud-card-medal {
  position: absolute;
  top: 2px;
  left: 3px;
  width: 28px;
  height: 28px;
  border-radius: 999px;
  border: 1px solid rgba(255, 255, 255, 0.7);
  background: radial-gradient(circle at 30% 30%, rgba(255, 255, 255, 0.95), rgba(201, 191, 176, 0.8) 42%, rgba(61, 53, 47, 0.88));
  display: inline-flex;
  align-items: center;
  justify-content: center;
  z-index: 8;
  box-shadow:
    0 2px 6px rgba(0, 0, 0, 0.55),
    0 0 8px rgba(255, 244, 189, 0.44);
}

.prompt-fraud-card-medal > span {
  width: 72%;
  height: 72%;
  border-radius: 999px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: var(--fraud-medal-bg);
  color: var(--fraud-medal-fg);
  font-size: 13px;
  font-weight: 900;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.7);
}

.prompt-fraud-card-art {
  margin: 30px 7px 8px;
  height: 88px;
  border-radius: 4px;
  border: 1px solid rgba(175, 161, 132, 0.8);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background:
    radial-gradient(100% 120% at 50% 0%, var(--fraud-art-top), transparent 56%),
    linear-gradient(180deg, rgba(255, 250, 235, 0.12), var(--fraud-art-bottom));
  box-shadow:
    inset 0 0 0 1px rgba(255, 236, 196, 0.3),
    inset 0 0 16px rgba(0, 0, 0, 0.38);
}

.prompt-fraud-card-glyph {
  font-size: 38px;
  font-weight: 900;
  line-height: 1;
  color: rgba(247, 253, 255, 0.96);
  text-shadow:
    0 0 10px rgba(255, 255, 255, 0.24),
    0 2px 5px rgba(0, 0, 0, 0.6);
}

.prompt-fraud-card--water {
  --fraud-edge-color: rgba(102, 152, 196, 0.78);
  --fraud-edge-glow: rgba(124, 196, 255, 0.38);
  --fraud-base-top: #1a2a3e;
  --fraud-base-bottom: #0f1826;
  --fraud-ribbon-start: #25689f;
  --fraud-ribbon-end: #1b446c;
  --fraud-medal-bg: radial-gradient(circle at 32% 28%, #ccf3ff, #4ea1d6 58%, #195580);
  --fraud-medal-fg: #effbff;
  --fraud-art-top: rgba(138, 206, 255, 0.4);
  --fraud-art-bottom: rgba(18, 48, 78, 0.84);
}

.prompt-fraud-card--fire {
  --fraud-edge-color: rgba(205, 123, 82, 0.78);
  --fraud-edge-glow: rgba(255, 140, 98, 0.4);
  --fraud-base-top: #3c1f18;
  --fraud-base-bottom: #1c120f;
  --fraud-ribbon-start: #c6352f;
  --fraud-ribbon-end: #8e1b17;
  --fraud-medal-bg: radial-gradient(circle at 32% 28%, #ffca7f, #f36d33 58%, #9b2e1a);
  --fraud-medal-fg: #fff8eb;
  --fraud-art-top: rgba(255, 176, 126, 0.42);
  --fraud-art-bottom: rgba(88, 30, 20, 0.84);
}

.prompt-fraud-card--earth {
  --fraud-edge-color: rgba(174, 138, 93, 0.8);
  --fraud-edge-glow: rgba(225, 186, 113, 0.34);
  --fraud-base-top: #32261a;
  --fraud-base-bottom: #1a1410;
  --fraud-ribbon-start: #8a5d2f;
  --fraud-ribbon-end: #60401f;
  --fraud-medal-bg: radial-gradient(circle at 32% 28%, #f1d79b, #c6924f 58%, #784d1d);
  --fraud-medal-fg: #fff6e4;
  --fraud-art-top: rgba(236, 199, 128, 0.36);
  --fraud-art-bottom: rgba(85, 57, 22, 0.84);
}

.prompt-fraud-card--wind {
  --fraud-edge-color: rgba(96, 169, 145, 0.78);
  --fraud-edge-glow: rgba(116, 223, 181, 0.34);
  --fraud-base-top: #183329;
  --fraud-base-bottom: #101f1b;
  --fraud-ribbon-start: #237258;
  --fraud-ribbon-end: #194e3d;
  --fraud-medal-bg: radial-gradient(circle at 32% 28%, #c8f9e6, #55b68d 58%, #216f54);
  --fraud-medal-fg: #edfff6;
  --fraud-art-top: rgba(145, 241, 205, 0.36);
  --fraud-art-bottom: rgba(25, 73, 56, 0.84);
}

.prompt-fraud-card--thunder {
  --fraud-edge-color: rgba(140, 124, 200, 0.8);
  --fraud-edge-glow: rgba(183, 148, 255, 0.36);
  --fraud-base-top: #24213d;
  --fraud-base-bottom: #171427;
  --fraud-ribbon-start: #5f4a99;
  --fraud-ribbon-end: #40306f;
  --fraud-medal-bg: radial-gradient(circle at 32% 28%, #efe2ff, #9c79dc 58%, #4e3385);
  --fraud-medal-fg: #faf3ff;
  --fraud-art-top: rgba(208, 180, 255, 0.38);
  --fraud-art-bottom: rgba(54, 36, 89, 0.84);
}

.prompt-fraud-card--generic {
  --fraud-edge-color: rgba(170, 190, 216, 0.56);
  --fraud-edge-glow: rgba(166, 193, 227, 0.34);
  --fraud-base-top: #223044;
  --fraud-base-bottom: #121c29;
  --fraud-ribbon-start: #44648a;
  --fraud-ribbon-end: #2d4159;
  --fraud-medal-bg: radial-gradient(circle at 32% 28%, #dfe9f7, #8ba5cc 58%, #4d6283);
  --fraud-medal-fg: #f2f7ff;
  --fraud-art-top: rgba(170, 200, 235, 0.34);
  --fraud-art-bottom: rgba(35, 51, 72, 0.84);
}

.prompt-fraud-side-pop-enter-active,
.prompt-fraud-side-pop-leave-active {
  transition: opacity 0.2s ease, transform 0.2s ease;
}

.prompt-fraud-side-pop-enter-from,
.prompt-fraud-side-pop-leave-to {
  opacity: 0;
  transform: translateX(18px) scale(0.98);
}

@media (max-width: 900px) {
  .prompt-fraud-global-panel {
    width: min(760px, calc(100vw - 28px));
    max-height: calc(100vh - 28px);
  }

  .prompt-fraud-grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }

  .prompt-fraud-card {
    min-height: 126px;
  }
}

@media (max-width: 560px) {
  .prompt-fraud-global-layer {
    justify-content: center;
    align-items: center;
    padding:
      max(10px, calc(var(--safe-top, 0px) + 4px))
      max(8px, calc(var(--safe-right, 0px) + 4px))
      max(10px, calc(var(--safe-bottom, 0px) + 4px))
      max(8px, calc(var(--safe-left, 0px) + 4px));
  }

  .prompt-fraud-global-panel {
    width: min(100%, 620px);
    border-radius: 12px;
    padding: 8px;
  }

  .prompt-fraud-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .prompt-fraud-card {
    min-height: 120px;
  }

  .prompt-fraud-card-glyph {
    font-size: 32px;
  }
}
</style>
