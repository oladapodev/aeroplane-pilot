import Database from "better-sqlite3";
import type { Config } from "../shared/config.js";

export type ServiceOverview = {
  id: string;
  name: string;
  status: string;
  runtimeMode: string;
  lastDeployedAt: string | null;
};

export type DeploymentInfo = {
  id: string;
  serviceId: string;
  status: string;
  trigger: string;
  startedAt: string | null;
  finishedAt: string | null;
};

export class AeroplaneClient {
  private db: Database.Database;

  constructor(config: Config) {
    this.db = new Database(config.dbPath);
  }

  listServices(): ServiceOverview[] {
    return this.db
      .prepare("SELECT id, name, status, runtime_mode, last_deployed_at FROM projects ORDER BY name")
      .all() as ServiceOverview[];
  }

  getService(id: string): ServiceOverview | null {
    return this.db
      .prepare("SELECT id, name, status, runtime_mode, last_deployed_at FROM projects WHERE id = ?")
      .get(id) as ServiceOverview | null;
  }

  latestDeployments(limit = 10): DeploymentInfo[] {
    return this.db
      .prepare("SELECT id, service_id, status, trigger, started_at, finished_at FROM deployments ORDER BY created_at DESC LIMIT ?")
      .all(limit) as DeploymentInfo[];
  }

  failedDeployments(limit = 5): DeploymentInfo[] {
    return this.db
      .prepare("SELECT id, service_id, status, trigger, started_at, finished_at FROM deployments WHERE status = 'failed' ORDER BY created_at DESC LIMIT ?")
      .all(limit) as DeploymentInfo[];
  }

  deploymentLogs(deploymentId: string, maxLines = 100): string[] {
    const rows = this.db
      .prepare("SELECT line FROM deployment_logs WHERE deployment_id = ? ORDER BY id DESC LIMIT ?")
      .all(deploymentId, maxLines) as { line: string }[];
    return rows.map((r) => r.line).reverse();
  }

  close() {
    this.db.close();
  }
}
