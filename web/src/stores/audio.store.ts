import { defineStore } from 'pinia'
import { ref, watch } from 'vue'
import { configureSfx } from '../audio/sfx'

const AUDIO_SETTINGS_KEY = 'starcup.audio'

function settingsStorage(): Storage | null {
  try {
    const storage = typeof window !== 'undefined' ? window.localStorage : undefined
    return storage && typeof storage.getItem === 'function' ? storage : null
  } catch {
    return null
  }
}

function loadSettings(): { muted: boolean; volume: number } {
  const fallback = { muted: false, volume: 0.6 }
  try {
    const raw = settingsStorage()?.getItem(AUDIO_SETTINGS_KEY)
    const parsed = raw ? JSON.parse(raw) : null
    if (!parsed) return fallback
    return {
      muted: typeof parsed.muted === 'boolean' ? parsed.muted : fallback.muted,
      volume: Number.isFinite(parsed.volume) ? Math.min(1, Math.max(0, parsed.volume)) : fallback.volume,
    }
  } catch {
    return fallback
  }
}

/** 音效设置（静音 / 音量），跨对局记住。 */
export const useAudioStore = defineStore('audio', () => {
  const initial = loadSettings()
  const muted = ref(initial.muted)
  const volume = ref(initial.volume)

  watch(
    [muted, volume],
    ([m, v]) => {
      configureSfx({ muted: m, volume: v })
      try {
        settingsStorage()?.setItem(AUDIO_SETTINGS_KEY, JSON.stringify({ muted: m, volume: v }))
      } catch {
        // 写入失败时仅本次会话生效
      }
    },
    { immediate: true },
  )

  function toggleMuted() {
    muted.value = !muted.value
  }

  function setVolume(value: number) {
    if (Number.isFinite(value)) volume.value = Math.min(1, Math.max(0, value))
  }

  return { muted, volume, toggleMuted, setVolume }
})
