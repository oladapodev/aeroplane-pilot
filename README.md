# Aeroplane Pilot

Aeroplane Pilot is the Go rewrite of the Aeroplane assistant. It reads the Aeroplane SQLite database, routes commands through a shared core, and can run over Telegram or Discord.

## What it does

- `/health` and `/status` commands
- free-text replies through an LLM
- Telegram runtime adapter
- Discord runtime adapter
- SQLite-backed Aeroplane reads

## Build

```bash
make test
make build
```

Binary output:

```bash
bin/pilot
```

## Run

Set at least one messaging token and an LLM key:

- `AEROPLANE_HOME` - Aeroplane install path, default `/opt/aeroplane`
- `LLM_PROVIDER` - default `openai`
- `LLM_API_KEY` or `OPENAI_API_KEY`
- `LLM_MODEL` - default `gpt-4o`
- `TELEGRAM_BOT_TOKEN` - optional
- `DISCORD_BOT_TOKEN` - optional
- `WHATSAPP_BOT_TOKEN` - reserved for later
- `DAILY_HEALTH_TIME` - default `08:00`

If no messaging token is set, the binary prints `/health` once and exits.
