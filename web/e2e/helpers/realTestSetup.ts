export interface ScenarioPlayerConfig {
  name: string
  camp: string
  char_role: string
}

export interface ScenarioCheat {
  target: string
  command: string
  args: string[]
}

export interface ScenarioConfig {
  human_player: ScenarioPlayerConfig
  bot_players: ScenarioPlayerConfig[]
  setup: {
    first_turn_player: string
    bots_paused: boolean
    cheats: ScenarioCheat[]
  }
}

export interface ScenarioResult {
  room_code: string
  human_player_id: string
  bot_player_ids: string[]
}

export async function setupTestScenario(config: ScenarioConfig): Promise<ScenarioResult> {
  const port = process.env.E2E_BACKEND_PORT ?? '18080'
  const resp = await fetch(`http://127.0.0.1:${port}/api/test/setup-scenario`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(config),
  })
  if (!resp.ok) {
    throw new Error(`Setup failed: ${resp.status} ${await resp.text()}`)
  }
  return resp.json()
}
