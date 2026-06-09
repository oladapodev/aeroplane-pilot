import { Client, GatewayIntentBits, Events } from "discord.js";
import type { Pilot } from "../agent/pilot.js";

export async function startDiscord(token: string, pilot: Pilot) {
  const client = new Client({
    intents: [GatewayIntentBits.Guilds, GatewayIntentBits.GuildMessages, GatewayIntentBits.MessageContent],
  });

  client.once(Events.ClientReady, (c) => {
    console.log(`Discord gateway started as ${c.user.tag}`);
  });

  client.on(Events.MessageCreate, async (msg) => {
    if (msg.author.bot) return;

    const content = msg.content.toLowerCase();

    if (content === "/status") {
      const result = await pilot.chat("List all services and their current status. Be brief.");
      await msg.reply(result.text);
    } else if (content === "/health") {
      const result = await pilot.dailyHealth();
      await msg.reply(result.text);
    } else if (content.startsWith("/")) {
      // Treat as a general pilot command
      const result = await pilot.chat(msg.content.slice(1).trim());
      await msg.reply(result.text);
    }
  });

  await client.login(token);
  return client;
}
