import { resolve } from "node:path";

export type Config = {
  aeroplaneHome: string;
  dbPath: string;
  llmProvider: "openai" | "anthropic" | "google";
  llmApiKey: string;
  llmModel: string;
  telegramToken?: string;
  discordToken?: string;
  dailyHealthTime: string;
};

export function loadConfig(): Config {
  const aeroplaneHome = process.env.AEROPLANE_HOME || "/opt/aeroplane";
  return {
    aeroplaneHome,
    dbPath: resolve(aeroplaneHome, "source/data/aeroplane.db"),
    llmProvider: (process.env.LLM_PROVIDER as Config["llmProvider"]) || "openai",
    llmApiKey: process.env.LLM_API_KEY || process.env.OPENAI_API_KEY || "",
    llmModel: process.env.LLM_MODEL || "gpt-4o",
    telegramToken: process.env.TELEGRAM_BOT_TOKEN,
    discordToken: process.env.DISCORD_BOT_TOKEN,
    dailyHealthTime: process.env.DAILY_HEALTH_TIME || "08:00",
  };
}
