#!/usr/bin/env node
import { spawn } from 'node:child_process'

const DEFAULT_BACKEND = 'http://127.0.0.1:8080'
const DEFAULT_FRONTEND = 'http://127.0.0.1:5173'
const PLAYER_COUNT = 6
const DEFAULT_NAMES = [
  '测试玩家1',
  '测试玩家2',
  '测试玩家3',
  '测试玩家4',
  '测试玩家5',
  '测试玩家6',
]
const DEFAULT_ROLES = [
  'assassin',
  'angel',
  'blade_master',
  'sealer',
  'saintess',
  'elementalist',
]
const DEFAULT_CAMPS = ['Blue', 'Blue', 'Blue', 'Red', 'Red', 'Red']

function parseArgs(argv) {
  const options = {
    backend: DEFAULT_BACKEND,
    frontend: DEFAULT_FRONTEND,
    names: DEFAULT_NAMES,
    roles: DEFAULT_ROLES,
    camps: DEFAULT_CAMPS,
    firstTurn: 'human',
    botsPaused: true,
    openPage: true,
    printJson: false,
  }

  for (let i = 0; i < argv.length; i += 1) {
    const arg = argv[i]
    const next = argv[i + 1]

    if (arg === '--backend' && next) {
      options.backend = next
      i += 1
    } else if (arg.startsWith('--backend=')) {
      options.backend = arg.slice('--backend='.length)
    } else if (arg === '--frontend' && next) {
      options.frontend = next
      i += 1
    } else if (arg.startsWith('--frontend=')) {
      options.frontend = arg.slice('--frontend='.length)
    } else if (arg === '--names' && next) {
      options.names = splitList(next)
      i += 1
    } else if (arg.startsWith('--names=')) {
      options.names = splitList(arg.slice('--names='.length))
    } else if (arg === '--roles' && next) {
      options.roles = splitList(next)
      i += 1
    } else if (arg.startsWith('--roles=')) {
      options.roles = splitList(arg.slice('--roles='.length))
    } else if (arg === '--camps' && next) {
      options.camps = splitList(next)
      i += 1
    } else if (arg.startsWith('--camps=')) {
      options.camps = splitList(arg.slice('--camps='.length))
    } else if (arg === '--first-turn' && next) {
      options.firstTurn = next
      i += 1
    } else if (arg.startsWith('--first-turn=')) {
      options.firstTurn = arg.slice('--first-turn='.length)
    } else if (arg === '--bots-live') {
      options.botsPaused = false
    } else if (arg === '--no-open') {
      options.openPage = false
    } else if (arg === '--json') {
      options.printJson = true
    } else if (arg === '--help' || arg === '-h') {
      printHelp()
      process.exit(0)
    } else {
      console.warn(`忽略未知参数: ${arg}`)
    }
  }

  return {
    ...options,
    backend: trimTrailingSlash(options.backend),
    frontend: trimTrailingSlash(options.frontend),
    names: ensureList(options.names, DEFAULT_NAMES),
    roles: ensureList(options.roles, DEFAULT_ROLES),
    camps: ensureList(options.camps, DEFAULT_CAMPS),
  }
}

function splitList(value) {
  return String(value)
    .split(',')
    .map(item => item.trim())
    .filter(Boolean)
}

function ensureList(values, defaults) {
  const result = values.slice(0, PLAYER_COUNT)
  for (let i = result.length; i < PLAYER_COUNT; i += 1) {
    result.push(defaults[i])
  }
  return result
}

function trimTrailingSlash(value) {
  return String(value).replace(/\/+$/, '')
}

function printHelp() {
  console.log(`创建一个已选定 6 个角色并自动开始的 3v3 测试对局。

Usage:
  npm run dev:battle
  npm run dev:battle -- --frontend http://127.0.0.1:5174 --roles assassin,angel,blade_master,sealer,saintess,elementalist

Prerequisite:
  后端需要以测试模式启动:
  STARCUP_TEST_MODE=1 go run ./cmd/server

Options:
  --backend <url>      后端地址，默认 ${DEFAULT_BACKEND}
  --frontend <url>     前端地址，默认 ${DEFAULT_FRONTEND}
  --names <list>       逗号分隔的 ${PLAYER_COUNT} 个玩家名
  --roles <list>       逗号分隔的 ${PLAYER_COUNT} 个角色 id
  --camps <list>       逗号分隔的 ${PLAYER_COUNT} 个阵营，默认 Blue,Blue,Blue,Red,Red,Red
  --first-turn <who>   human 或 bot 序号 0-4，默认 human
  --bots-live          不暂停 bot 行动
  --no-open            只打印 URL，不打开浏览器
  --json               输出机器可读 JSON
`)
}

async function fetchJson(url, init, errorLabel) {
  let response
  try {
    response = await fetch(url, init)
  } catch (error) {
    throw new Error(`${errorLabel}: ${error.message}`)
  }

  if (!response.ok) {
    const body = await response.text()
    throw new Error(`${errorLabel}: HTTP ${response.status} ${body}`)
  }

  return response.json()
}

async function isReachable(url) {
  try {
    const response = await fetch(url)
    return response.ok
  } catch {
    return false
  }
}

function scenarioPayload(options) {
  const players = options.names.map((name, index) => ({
    name,
    camp: options.camps[index],
    char_role: options.roles[index],
  }))

  return {
    human_player: players[0],
    bot_players: players.slice(1),
    setup: {
      first_turn_player: options.firstTurn,
      bots_paused: options.botsPaused,
      cheats: [
        { target: 'human', command: 'card_element', args: ['Earth', '2'] },
        { target: 'human', command: 'card_element', args: ['Light', '1'] },
      ],
    },
  }
}

function buildFrontendUrl(frontend, backend, roomCode, playerName) {
  const url = new URL(frontend)
  url.searchParams.set('room', roomCode)
  url.searchParams.set('name', playerName)
  url.searchParams.set('debug', '1')
  url.searchParams.set('ws', backend)
  return url.toString()
}

function openUrl(url) {
  if (process.platform === 'darwin') {
    spawn('open', [url], { stdio: 'ignore', detached: true }).unref()
    return
  }

  if (process.platform === 'win32') {
    spawn('cmd', ['/c', 'start', '', url], { stdio: 'ignore', detached: true }).unref()
    return
  }

  spawn('xdg-open', [url], { stdio: 'ignore', detached: true }).unref()
}

async function main() {
  const options = parseArgs(process.argv.slice(2))

  const health = await isReachable(`${options.backend}/api/health`)
  if (!health) {
    console.error(`后端未就绪: ${options.backend}/api/health`)
    console.error('请先在仓库根目录运行: STARCUP_TEST_MODE=1 go run ./cmd/server')
    process.exit(1)
  }

  const frontendReady = await isReachable(options.frontend)
  if (!frontendReady) {
    console.warn(`前端未检测到: ${options.frontend}`)
    console.warn('请先在 web 目录运行: npm run dev')
    console.warn('仍会继续创建测试对局并打印 URL。')
  }

  const payload = scenarioPayload(options)
  let scenario
  try {
    scenario = await fetchJson(
      `${options.backend}/api/test/setup-scenario`,
      {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
      },
      '创建测试对局失败',
    )
  } catch (error) {
    if (String(error.message || error).includes('404')) {
      console.error('测试接口不可用。请确认后端用 STARCUP_TEST_MODE=1 启动。')
    }
    throw error
  }

  const url = buildFrontendUrl(options.frontend, options.backend, scenario.room_code, options.names[0])

  const result = {
    roomCode: scenario.room_code,
    humanPlayerId: scenario.human_player_id,
    botPlayerIds: scenario.bot_player_ids,
    url,
    roles: options.roles,
    names: options.names,
    camps: options.camps,
  }

  if (options.printJson) {
    console.log(JSON.stringify(result, null, 2))
  } else {
    console.log(`测试对局: ${scenario.room_code}`)
    console.log(`真人视角: ${options.names[0]} (${options.roles[0]})`)
    console.log(`URL: ${url}`)
    console.log('6 个角色:')
    options.roles.forEach((role, index) => {
      console.log(`${index + 1}. ${options.names[index]} / ${options.camps[index]} / ${role}`)
    })
  }

  if (options.openPage) {
    openUrl(url)
    if (!options.printJson) console.log('已请求浏览器打开真人视角页面。')
  }
}

main().catch((error) => {
  console.error(error.message || error)
  process.exit(1)
})
