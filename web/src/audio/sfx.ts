/**
 * 对战音效引擎（Web Audio）。
 *
 * SAMPLE_FILES 里登记的素材在解锁时预载：裁掉开头静音、按 TARGET_RMS_DB 统一响度后优先播放；
 * 没登记或加载失败的音效回落到振荡器 / 噪声合成的占位音。
 * 浏览器要求用户交互后才能出声，入口处调用 installAudioUnlock() 即可。
 */

export type SfxCue =
  | 'attack'
  | 'magic'
  | 'hit'
  | 'block'
  | 'miss'
  | 'heal'
  | 'skill'
  | 'trigger'
  | 'gem'
  | 'draw'
  | 'turn'
  | 'prompt'
  | 'game-end'
  | 'victory'
  | 'defeat'

/** 真实素材登记表：放进 public/assets/sfx/ 后在这里加一行；没登记的音效继续用合成音。 */
const SAMPLE_FILES: Partial<Record<SfxCue, string>> = {
  attack: '/assets/sfx/attack.mp3',
  magic: '/assets/sfx/magic.mp3',
  hit: '/assets/sfx/hit.mp3',
  block: '/assets/sfx/block.mp3',
  miss: '/assets/sfx/miss.mp3',
  heal: '/assets/sfx/heal.mp3',
  skill: '/assets/sfx/skill.mp3',
  trigger: '/assets/sfx/trigger.mp3',
  gem: '/assets/sfx/gem.mp3',
  draw: '/assets/sfx/draw.mp3',
  turn: '/assets/sfx/turn.mp3',
  prompt: '/assets/sfx/prompt.mp3',
  victory: '/assets/sfx/victory.mp3',
  defeat: '/assets/sfx/defeat.mp3',
  // game-end 没有素材：没有观战位，只在结束消息里认不出胜方时兜底，用合成音
}

/**
 * 每个音效的目标响度（有声部分的 RMS，dBFS）。素材各自生成、响度差异很大，
 * 解码后按这里统一：提示音最醒目，高频出现的触发 / 宝石最轻。
 */
const TARGET_RMS_DB: Record<SfxCue, number> = {
  prompt: -18,
  hit: -21,
  block: -23,
  attack: -24,
  skill: -24,
  turn: -24,
  magic: -25,
  draw: -26,
  miss: -27,
  heal: -25,
  gem: -27,
  trigger: -30,
  'game-end': -22,
  victory: -20,
  defeat: -22,
}

/** 归一化后的峰值上限，防止增益拉高后削波。 */
const PEAK_CEILING = 0.89
/** 低于这个电平（约 -45 dBFS）视为静音，用来裁掉素材开头的空白。 */
const SILENCE_THRESHOLD = 0.0056
/** 起音前保留一点余量，避免把瞬态切掉。 */
const PRE_ROLL_SEC = 0.004

/** 同一音效的最短间隔，避免一批步骤同时到达时叠成噪音。 */
const MIN_REPEAT_MS = 70

let ctx: AudioContext | null = null
let master: GainNode | null = null
let volume = 0.6
let muted = false
interface Sample {
  buffer: AudioBuffer
  /** 跳过开头静音的偏移（秒） */
  offset: number
  /** 归一化增益 */
  gain: number
}

const samples = new Map<SfxCue, Sample>()
const lastPlayedAt = new Map<SfxCue, number>()

function audioContextCtor(): typeof AudioContext | undefined {
  if (typeof window === 'undefined') return undefined
  return window.AudioContext ?? (window as unknown as { webkitAudioContext?: typeof AudioContext }).webkitAudioContext
}

function effectiveGain(): number {
  return muted ? 0 : volume
}

function ensureContext(): AudioContext | null {
  if (ctx) return ctx
  const Ctor = audioContextCtor()
  if (!Ctor) return null
  try {
    ctx = new Ctor()
  } catch {
    return null
  }
  master = ctx.createGain()
  master.gain.value = effectiveGain()
  master.connect(ctx.destination)
  void preloadSamples(ctx)
  return ctx
}

async function preloadSamples(context: AudioContext) {
  for (const [cue, url] of Object.entries(SAMPLE_FILES) as [SfxCue, string][]) {
    try {
      const res = await fetch(url)
      if (!res.ok) continue
      const buffer = await context.decodeAudioData(await res.arrayBuffer())
      samples.set(cue, analyzeSample(buffer, TARGET_RMS_DB[cue]))
    } catch {
      // 素材缺失或解码失败时继续用合成音
    }
  }
}

/** 找出起音位置，并按有声部分的 RMS 算出达到目标响度所需的增益（受峰值上限约束）。 */
export function analyzeSample(buffer: AudioBuffer, targetDb: number): Sample {
  const channels = Array.from({ length: buffer.numberOfChannels }, (_, i) => buffer.getChannelData(i))
  const length = buffer.length
  let first = length
  let last = -1
  let peak = 0
  for (let i = 0; i < length; i += 1) {
    let v = 0
    for (const data of channels) v = Math.max(v, Math.abs(data[i]!))
    if (v > peak) peak = v
    if (v > SILENCE_THRESHOLD) {
      if (first === length) first = i
      last = i
    }
  }
  if (last < 0 || peak === 0) return { buffer, offset: 0, gain: 1 }

  let sum = 0
  for (const data of channels) {
    for (let i = first; i <= last; i += 1) sum += data[i]! * data[i]!
  }
  const rms = Math.sqrt(sum / ((last - first + 1) * channels.length))
  const wanted = Math.pow(10, targetDb / 20) / rms
  const gain = Math.min(wanted, PEAK_CEILING / peak)
  const offset = Math.max(0, first / buffer.sampleRate - PRE_ROLL_SEC)
  return { buffer, offset, gain }
}

export function configureSfx(options: { volume?: number; muted?: boolean }) {
  if (typeof options.volume === 'number' && Number.isFinite(options.volume)) {
    volume = Math.min(1, Math.max(0, options.volume))
  }
  if (typeof options.muted === 'boolean') muted = options.muted
  if (ctx && master) master.gain.setTargetAtTime(effectiveGain(), ctx.currentTime, 0.02)
}

/** 首次 pointerdown / keydown 时创建并恢复 AudioContext（自动播放策略要求）。 */
export function installAudioUnlock(target: Pick<Window, 'addEventListener' | 'removeEventListener'> = window) {
  const unlock = () => {
    const context = ensureContext()
    if (!context) return
    if (context.state === 'suspended') void context.resume()
    if (context.state !== 'suspended') {
      target.removeEventListener('pointerdown', unlock)
      target.removeEventListener('keydown', unlock)
    }
  }
  target.addEventListener('pointerdown', unlock)
  target.addEventListener('keydown', unlock)
}

export function playSfx(cue: SfxCue) {
  if (muted || volume <= 0) return
  // 还没解锁就不出声，也不在这里创建 context（会被浏览器挂起并报警告）
  if (!ctx || !master || ctx.state !== 'running') return
  const nowMs = ctx.currentTime * 1000
  const last = lastPlayedAt.get(cue)
  if (last !== undefined && nowMs - last < MIN_REPEAT_MS) return
  lastPlayedAt.set(cue, nowMs)

  const sample = samples.get(cue)
  if (sample) {
    const src = ctx.createBufferSource()
    src.buffer = sample.buffer
    const gain = ctx.createGain()
    gain.gain.value = sample.gain
    src.connect(gain).connect(master)
    src.start(0, sample.offset)
    return
  }
  SYNTH[cue](ctx, master, ctx.currentTime)
}

// ---------------------------------------------------------------------------
// 合成占位音
// ---------------------------------------------------------------------------

interface ToneOptions {
  freq: number
  to?: number
  dur: number
  type?: OscillatorType
  gain?: number
  delay?: number
}

function tone(c: AudioContext, out: AudioNode, t0: number, o: ToneOptions) {
  const start = t0 + (o.delay ?? 0)
  const osc = c.createOscillator()
  const env = c.createGain()
  osc.type = o.type ?? 'sine'
  osc.frequency.setValueAtTime(o.freq, start)
  if (o.to) osc.frequency.exponentialRampToValueAtTime(o.to, start + o.dur)
  const peak = o.gain ?? 0.3
  env.gain.setValueAtTime(0.0001, start)
  env.gain.exponentialRampToValueAtTime(peak, start + 0.008)
  env.gain.exponentialRampToValueAtTime(0.0001, start + o.dur)
  osc.connect(env).connect(out)
  osc.start(start)
  osc.stop(start + o.dur + 0.02)
}

let noiseBuffer: AudioBuffer | null = null

function noise(c: AudioContext, out: AudioNode, t0: number, o: { dur: number; gain?: number; filter: number; q?: number; type?: BiquadFilterType; delay?: number }) {
  if (!noiseBuffer || noiseBuffer.sampleRate !== c.sampleRate) {
    noiseBuffer = c.createBuffer(1, c.sampleRate, c.sampleRate)
    const data = noiseBuffer.getChannelData(0)
    for (let i = 0; i < data.length; i += 1) data[i] = Math.random() * 2 - 1
  }
  const start = t0 + (o.delay ?? 0)
  const src = c.createBufferSource()
  src.buffer = noiseBuffer
  const filter = c.createBiquadFilter()
  filter.type = o.type ?? 'bandpass'
  filter.frequency.value = o.filter
  filter.Q.value = o.q ?? 1
  const env = c.createGain()
  env.gain.setValueAtTime(o.gain ?? 0.3, start)
  env.gain.exponentialRampToValueAtTime(0.0001, start + o.dur)
  src.connect(filter).connect(env).connect(out)
  src.start(start)
  src.stop(start + o.dur + 0.02)
}

type Synth = (c: AudioContext, out: AudioNode, t: number) => void

const SYNTH: Record<SfxCue, Synth> = {
  // 攻击出手：短促的挥砍风声
  attack: (c, o, t) => {
    noise(c, o, t, { dur: 0.18, gain: 0.35, filter: 1800, q: 0.8, type: 'highpass' })
    tone(c, o, t, { freq: 420, to: 160, dur: 0.14, type: 'triangle', gain: 0.12 })
  },
  // 法术出手：上扬的闪光
  magic: (c, o, t) => {
    tone(c, o, t, { freq: 520, to: 1560, dur: 0.28, type: 'sine', gain: 0.18 })
    tone(c, o, t, { freq: 780, to: 2100, dur: 0.24, type: 'triangle', gain: 0.08, delay: 0.04 })
  },
  // 命中：低频重击
  hit: (c, o, t) => {
    tone(c, o, t, { freq: 150, to: 50, dur: 0.22, type: 'sine', gain: 0.5 })
    noise(c, o, t, { dur: 0.1, gain: 0.3, filter: 900, q: 0.7 })
  },
  // 被挡下：金属格挡
  block: (c, o, t) => {
    tone(c, o, t, { freq: 1250, dur: 0.25, type: 'square', gain: 0.06 })
    tone(c, o, t, { freq: 1870, dur: 0.18, type: 'triangle', gain: 0.08 })
    noise(c, o, t, { dur: 0.05, gain: 0.25, filter: 4000, q: 2 })
  },
  // 未命中：轻飘的下滑
  miss: (c, o, t) => {
    noise(c, o, t, { dur: 0.22, gain: 0.15, filter: 2600, q: 0.6 })
    tone(c, o, t, { freq: 600, to: 300, dur: 0.18, type: 'sine', gain: 0.06 })
  },
  // 治疗：温和的上行三音
  heal: (c, o, t) => {
    tone(c, o, t, { freq: 523, dur: 0.3, gain: 0.14 })
    tone(c, o, t, { freq: 659, dur: 0.3, gain: 0.12, delay: 0.07 })
    tone(c, o, t, { freq: 784, dur: 0.4, gain: 0.12, delay: 0.14 })
  },
  // 主动技能：亮一点的和弦
  skill: (c, o, t) => {
    tone(c, o, t, { freq: 392, dur: 0.35, type: 'triangle', gain: 0.14 })
    tone(c, o, t, { freq: 587, dur: 0.35, type: 'triangle', gain: 0.1, delay: 0.03 })
    tone(c, o, t, { freq: 1175, to: 1400, dur: 0.25, gain: 0.05, delay: 0.06 })
  },
  // 被动触发：很轻的提示
  trigger: (c, o, t) => {
    tone(c, o, t, { freq: 880, dur: 0.12, type: 'sine', gain: 0.07 })
  },
  // 宝石 / 水晶：清脆的叮
  gem: (c, o, t) => {
    tone(c, o, t, { freq: 1568, dur: 0.22, gain: 0.1 })
    tone(c, o, t, { freq: 2349, dur: 0.16, gain: 0.05, delay: 0.03 })
  },
  // 摸牌：纸张摩擦
  draw: (c, o, t) => {
    noise(c, o, t, { dur: 0.09, gain: 0.18, filter: 3200, q: 1.2 })
    noise(c, o, t, { dur: 0.07, gain: 0.12, filter: 3800, q: 1.2, delay: 0.06 })
  },
  // 回合切换
  turn: (c, o, t) => {
    tone(c, o, t, { freq: 330, dur: 0.3, type: 'triangle', gain: 0.12 })
    tone(c, o, t, { freq: 494, dur: 0.4, type: 'triangle', gain: 0.1, delay: 0.1 })
  },
  // 轮到我操作：最醒目的两声
  prompt: (c, o, t) => {
    tone(c, o, t, { freq: 988, dur: 0.16, gain: 0.2 })
    tone(c, o, t, { freq: 1319, dur: 0.28, gain: 0.2, delay: 0.12 })
  },
  // 胜利：上行三和弦
  victory: (c, o, t) => {
    tone(c, o, t, { freq: 523, dur: 0.4, type: 'triangle', gain: 0.16 })
    tone(c, o, t, { freq: 659, dur: 0.4, type: 'triangle', gain: 0.14, delay: 0.14 })
    tone(c, o, t, { freq: 784, dur: 0.4, type: 'triangle', gain: 0.14, delay: 0.28 })
    tone(c, o, t, { freq: 1047, dur: 1.1, type: 'triangle', gain: 0.16, delay: 0.42 })
  },
  // 失败：缓慢下行
  defeat: (c, o, t) => {
    tone(c, o, t, { freq: 440, dur: 0.6, type: 'sine', gain: 0.14 })
    tone(c, o, t, { freq: 392, dur: 0.6, type: 'sine', gain: 0.13, delay: 0.32 })
    tone(c, o, t, { freq: 311, dur: 1.2, type: 'sine', gain: 0.14, delay: 0.64 })
  },
  // 对局结束（观战 / 判断不出胜负时）
  'game-end': (c, o, t) => {
    tone(c, o, t, { freq: 392, dur: 0.5, type: 'triangle', gain: 0.16 })
    tone(c, o, t, { freq: 494, dur: 0.5, type: 'triangle', gain: 0.14, delay: 0.15 })
    tone(c, o, t, { freq: 587, dur: 0.9, type: 'triangle', gain: 0.16, delay: 0.3 })
  },
}
