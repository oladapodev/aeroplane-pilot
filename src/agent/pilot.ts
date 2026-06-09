import { generateText, type CoreTool } from "ai";
import { createOpenAI } from "@ai-sdk/openai";
import { createAnthropic } from "@ai-sdk/anthropic";
import { createGoogleGenerativeAI } from "@ai-sdk/google";
import { z } from "zod";
import type { Config } from "../shared/config.js";
import { AeroplaneClient } from "../integration/aeroplane.js";

export type PilotResult = { text: string };

export class Pilot {
  private client: AeroplaneClient;
  private config: Config;

  constructor(config: Config) {
    this.config = config;
    this.client = new AeroplaneClient(config);
  }

  private model() {
    const { llmApiKey, llmModel } = this.config;
    if (this.config.llmProvider === "anthropic") {
      return createAnthropic({ apiKey: llmApiKey })(llmModel);
    }
    if (this.config.llmProvider === "google") {
      return createGoogleGenerativeAI({ apiKey: llmApiKey })(llmModel);
    }
    return createOpenAI({ apiKey: llmApiKey })(llmModel);
  }

  async chat(message: string): Promise<PilotResult> {
    const tools: Record<string, CoreTool> = {
      listServices: {
        description: "List all services on the Aeroplane instance with their status",
        parameters: z.object({}),
        execute: async () => this.client.listServices(),
      },
      getService: {
        description: "Get details about a specific service by ID or name",
        parameters: z.object({ serviceId: z.string() }),
        execute: async ({ serviceId }) => this.client.getService(serviceId),
      },
      latestDeployments: {
        description: "Get the most recent deployments",
        parameters: z.object({ limit: z.number().optional().default(10) }),
        execute: async ({ limit }) => this.client.latestDeployments(limit),
      },
      failedDeployments: {
        description: "Get recent failed deployments",
        parameters: z.object({ limit: z.number().optional().default(5) }),
        execute: async ({ limit }) => this.client.failedDeployments(limit),
      },
      deploymentLogs: {
        description: "Get logs for a specific deployment",
        parameters: z.object({ deploymentId: z.string() }),
        execute: async ({ deploymentId }) => this.client.deploymentLogs(deploymentId),
      },
    };

    const systemPrompt = `You are Pilot, the AI copilot for the Aeroplane deployment platform.
You run on the user's VPS and have direct access to their Aeroplane database.

You can:
- List services and their status
- Check recent and failed deployments
- Show deployment logs
- Diagnose why things failed
- Give health summaries

Be concise and direct. When you see issues (failed deploys, services down), flag them clearly.
Use the tools available. If you don't have enough info, ask.`;

    const result = await generateText({
      model: this.model(),
      system: systemPrompt,
      prompt: message,
      tools,
      maxSteps: 5,
    });

    return { text: result.text };
  }

  async dailyHealth(): Promise<PilotResult> {
    const services = this.client.listServices();
    const recent = this.client.latestDeployments(5);
    const failed = this.client.failedDeployments(5);

    const report = [
      `**Services:** ${services.length} total`,
      ...services.map((s) => `  - ${s.name}: ${s.status} (${s.runtimeMode})`),
      "",
      `**Recent deployments:** ${recent.length}`,
      ...recent.map((d) => `  - ${d.status} triggered by ${d.trigger}`),
      "",
      failed.length > 0
        ? `**Failed deployments:** ${failed.length}`
        : "**No failed deployments.**",
    ].join("\n");

    return this.chat(
      `Here's the Aeroplane state:\n\n${report}\n\nWrite a brief health summary in plain text. Flag anything needing attention.`
    );
  }

  close() {
    this.client.close();
  }
}
