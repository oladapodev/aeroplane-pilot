# Pilot

Pilot is the AI copilot for [Aeroplane](https://github.com/xt42io/aeroplane). It monitors your deployments, answers questions about your infrastructure, and sends daily health reports — all through Telegram, Discord, or other messaging platforms.

## How it works

Pilot runs alongside Aeroplane on your VPS. It reads the same SQLite database and talks to LLMs (OpenAI, Anthropic, etc.) to answer questions and generate reports. When you message it on Telegram or Discord, it can:

- **Check status**: "What's running right now?"
- **Diagnose failures**: "Why did the last deploy fail?"
- **Daily health reports**: Scheduled summary of all services, deployments, disk usage, and system health
- **Alerts**: "Service `api` just crashed" with relevant logs
- **Manage deployments**: Trigger redeploys, view logs, check env vars

## Installation

```bash
# On your Aeroplane VPS
curl -fsSL https://get.pilot.run | sh
```

Or run alongside Aeroplane locally:

```bash
git clone https://github.com/oladapodev/aeroplane-pilot.git
cd aeroplane-pilot
cp .env.example .env
npm install
npm run dev
```

## Configuration

Pilot is configured through environment variables or a `.env` file:

| Variable | Description | Required |
|---|---|---|
| `AEROPLANE_HOME` | Path to Aeroplane installation (reads DB and config) | Yes |
| `TELEGRAM_BOT_TOKEN` | Telegram bot token | At least one gateway |
| `DISCORD_BOT_TOKEN` | Discord bot token | At least one gateway |
| `OPENAI_API_KEY` | LLM provider key | Yes (or another provider) |
| `ANTHROPIC_API_KEY` | Alternative LLM provider | No |
| `DAILY_HEALTH_TIME` | Cron time for daily reports (default: `08:00`) | No |

## License

Apache-2.0
