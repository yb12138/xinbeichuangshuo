#!/usr/bin/env node
import { spawn } from 'node:child_process'

const DEFAULT_BACKEND = 'http://127.0.0.1:8080'
const DEFAULT_FRONTEND = 'http://127.0.0.1:5173'
const DEFAULT_NAMES = ['测试玩家1', '测试玩家2', '测试玩家3', '测试玩家4']

function parseArgs(argv) {
  const options = {
    backend: DEFAULT_BACKEND,
    frontend: DEFAULT_FRONTEND,
    names: DEFAULT_NAMES,
    openPages: true,
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
      options.names = splitNames(next)
      i += 1
    } else if (arg.startsWith('--names=')) {
      options.names = splitNames(arg.slice('--names='.length))
    } else if (arg === '--no-open') {
      options.openPages = false
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
    names: ensureFourNames(options.names),
  }
}

function splitNames(value) {
  return String(value)
    .split(',')
    .map((name) => name.trim())
    .filter(Boolean)
}

function ensureFourNames(names) {
  const result = names.slice(0, 4)
  for (let i = result.length; i < 4; i += 1) {
    result.push(DEFAULT_NAMES[i])
  }
  return result
}

function trimTrailingSlash(value) {
  return String(value).replace(/\/+$/, '')
}

function printHelp() {
  console.log(`一键创建四页面测试房间

Usage:
  npm run dev:quad
  npm run dev:quad -- --frontend http://127.0.0.1:5173 --backend http://127.0.0.1:8080 --names A,B,C,D

Options:
  --backend <url>   后端地址，默认 ${DEFAULT_BACKEND}
  --frontend <url>  前端地址，默认 ${DEFAULT_FRONTEND}
  --names <list>    逗号分隔的 4 个玩家名
  --no-open         只打印 URL，不打开浏览器
`)
}

async function fetchJson(url, errorLabel) {
  let response
  try {
    response = await fetch(url)
  } catch (error) {
    throw new Error(`${errorLabel}: ${error.message}`)
  }

  if (!response.ok) {
    throw new Error(`${errorLabel}: HTTP ${response.status} ${await response.text()}`)
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

function buildPlayerUrls(frontend, roomCode, names) {
  return names.map((name) => {
    const url = new URL(frontend)
    url.searchParams.set('room', roomCode)
    url.searchParams.set('name', name)
    return url.toString()
  })
}

function openUrls(urls) {
  if (process.platform === 'darwin') {
    spawn('open', urls, { stdio: 'ignore', detached: true }).unref()
    return
  }

  if (process.platform === 'win32') {
    for (const url of urls) {
      spawn('cmd', ['/c', 'start', '', url], { stdio: 'ignore', detached: true }).unref()
    }
    return
  }

  for (const url of urls) {
    spawn('xdg-open', [url], { stdio: 'ignore', detached: true }).unref()
  }
}

async function main() {
  const options = parseArgs(process.argv.slice(2))

  const health = await isReachable(`${options.backend}/api/health`)
  if (!health) {
    console.error(`后端未就绪: ${options.backend}/api/health`)
    console.error('请先在仓库根目录运行: go run ./cmd/server')
    process.exit(1)
  }

  const frontendReady = await isReachable(options.frontend)
  if (!frontendReady) {
    console.warn(`前端未检测到: ${options.frontend}`)
    console.warn('请先在 web 目录运行: npm run dev')
    console.warn('仍会继续创建房间并打印 URL。')
  }

  const room = await fetchJson(`${options.backend}/api/room/create`, '创建房间失败')
  const roomCode = room.room_code
  if (!roomCode) {
    throw new Error(`创建房间失败: 响应缺少 room_code (${JSON.stringify(room)})`)
  }

  const urls = buildPlayerUrls(options.frontend, roomCode, options.names)

  console.log(`房间码: ${roomCode}`)
  console.log('四个测试页面:')
  urls.forEach((url, index) => {
    console.log(`${index + 1}. ${options.names[index]} -> ${url}`)
  })

  if (options.openPages) {
    openUrls(urls)
    console.log('已请求浏览器打开 4 个页面。')
  }
}

main().catch((error) => {
  console.error(error.message || error)
  process.exit(1)
})
