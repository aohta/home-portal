export interface DiskFilesystem {
  name: string;
  path: string;
  status: "ok" | "error";
  total_bytes: number | null;
  used_bytes: number | null;
  available_bytes: number | null;
}

export interface DiskSnapshot {
  collected_at: string;
  stale: boolean;
  filesystems: DiskFilesystem[];
}

const capacity = (value: unknown): value is number => typeof value === "number" && Number.isSafeInteger(value) && value >= 0;

function isFilesystem(value: unknown): value is DiskFilesystem {
  if (typeof value !== "object" || value === null) return false;
  const row = value as Record<string, unknown>;
  if (typeof row.name !== "string" || row.name.trim() === "" || typeof row.path !== "string" || !row.path.startsWith("/")) return false;
  if (row.status === "error") return row.total_bytes === null && row.used_bytes === null && row.available_bytes === null;
  return row.status === "ok" && capacity(row.total_bytes) && row.total_bytes > 0 &&
    capacity(row.used_bytes) && capacity(row.available_bytes) && row.used_bytes + row.available_bytes <= row.total_bytes;
}

export async function loadDisks(signal?: AbortSignal): Promise<DiskSnapshot> {
  const unavailable = "空き容量の結果を取得できませんでした。時間をおいて再読み込みしてください。";
  const response = await fetch("./api/disks", { signal, cache: "no-store", credentials: "same-origin" });
  if (!response.ok) throw new Error(unavailable);
  let body: unknown;
  try { body = await response.json(); } catch { throw new Error(unavailable); }
  if (typeof body !== "object" || body === null || !("collected_at" in body) || typeof body.collected_at !== "string" ||
      !Number.isFinite(Date.parse(body.collected_at)) || !("stale" in body) || typeof body.stale !== "boolean" ||
      !("filesystems" in body) || !Array.isArray(body.filesystems) || body.filesystems.length < 1 || body.filesystems.length > 16 ||
      !body.filesystems.every(isFilesystem)) throw new Error(unavailable);
  return body as DiskSnapshot;
}

export function diskSnapshotStale(snapshot: DiskSnapshot, now: number): boolean {
  return snapshot.stale || now - Date.parse(snapshot.collected_at) >= 15 * 60_000;
}

export function formatBytes(value: number | null): string {
  if (value === null) return "未取得";
  const units = ["B", "KiB", "MiB", "GiB", "TiB", "PiB"];
  const index = value === 0 ? 0 : Math.min(Math.floor(Math.log(value) / Math.log(1024)), units.length - 1);
  return new Intl.NumberFormat("ja-JP", { maximumFractionDigits: 1 }).format(value / 1024 ** index) + " " + units[index];
}
