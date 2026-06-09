import { Telegraf } from "telegraf";
import type { Pilot } from "../agent/pilot.js";

export function startTelegram(token: string, pilot: Pilot) {
  const bot = new Telegraf(token);

  bot.start((ctx) =>
    ctx.reply("👋 I'm Pilot, your Aeroplane copilot. Ask me anything about your deployments.")
  );

  bot.help((ctx) =>
    ctx.reply(
      "Commands:\n" +
      "/status — list all services and their health\n" +
      "/health — daily health summary\n" +
      "/recent — last 10 deployments\n" +
      "/failures — recent failed deployments\n" +
      "Or just chat with me naturally!"
    )
  );

  bot.command("status", async (ctx) => {
    const result = await pilot.chat("List all services and their current status. Be brief.");
    await ctx.reply(result.text);
  });

  bot.command("health", async (ctx) => {
    const result = await pilot.dailyHealth();
    await ctx.reply(result.text);
  });

  bot.command("recent", async (ctx) => {
    const result = await pilot.chat("List the 10 most recent deployments with their status.");
    await ctx.reply(result.text);
  });

  bot.command("failures", async (ctx) => {
    const result = await pilot.chat("List recent failed deployments and what went wrong.");
    await ctx.reply(result.text);
  });

  bot.on("text", async (ctx) => {
    const result = await pilot.chat(ctx.message.text);
    await ctx.reply(result.text);
  });

  bot.launch();
  console.log("Telegram gateway started");
  return bot;
}
