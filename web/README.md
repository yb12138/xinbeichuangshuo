# StarCup Web

Vue 3 + TypeScript + Pinia + Vite frontend for the StarCup game room and battle UI.

## Development

Run the frontend dev server from this directory:

```bash
npm run dev
```

The default frontend URL is `http://127.0.0.1:5173`.

## Frontend Test Room Entry

For frontend manual testing and browser verification, enter a real game room through the prepared battle script instead of manually creating/joining a room.

Start the backend in test mode:

```bash
STARCUP_TEST_MODE=1 go run ./cmd/server
```

In another terminal, start the frontend:

```bash
cd web
npm run dev
```

Then create and open an auto-started 3v3 test battle:

```bash
cd web
npm run dev:battle
```

`npm run dev:battle` calls `scripts/open-test-battle.mjs`. It creates a test room through the backend test API, assigns six roles, starts the battle, and opens the first player's room URL. Future frontend visual/manual tests should use this command as the default way to enter the game room.

Useful options:

```bash
npm run dev:battle -- --frontend http://127.0.0.1:5174
npm run dev:battle -- --roles assassin,angel,blade_master,sealer,saintess,elementalist
npm run dev:battle -- --no-open --json
```

## Automated Checks

```bash
npm run test
npm run build
```
