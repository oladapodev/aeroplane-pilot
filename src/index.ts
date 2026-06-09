import { loadConfig } from "./shared/config.js";
import { Pilot } from "./agent/pilot.js";

const config = loadConfig();

if (!config.llmApiKey) {
  console.error("Missing LLM_API_KEY or OPENAI_API_KEY");
  process.exit(1);
}

const pilot = new Pilot(config);

const activeGateways: string[] = [];

if (config.telegramToken) {
  const { startTelegram } = await import("./gateway/telegram.js");
  startTelegram(config.telegramToken, pilot);
  activeGateways.push("Telegram");
}

if (config.discordToken) {
  const { startDiscord } = await import("./gateway/discord.js");
  await startDiscord(config.discordToken, pilot);
  activeGateways.push("Discord");
}

if (activeGateways.length === 0) {
  console.log("No gateways configured. Run a one-shot health check:");
  console.log("");
  const result = await pilot.dailyHealth();
  console.log(result.text);
}

console.log(`Pilot ready. Gateways: ${activeGateways.length > 0 ? activeGateways.join(", ") : "none"}`);

process.on("SIGINT", () => {
  pilot.close();
  process.exit(0);
});
