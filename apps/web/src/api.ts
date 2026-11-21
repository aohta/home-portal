export interface Observation {
  address: string;
  observed_at: string;
  temperature_c: number;
  humidity_percent: number;
  battery_percent: number | null;
  co2_ppm: number | null;
}

const finiteNumber = (value: unknown): value is number => typeof value === "number" && Number.isFinite(value);
const nullableNumber = (value: unknown): boolean => value === null || finiteNumber(value);

function isObservation(value: unknown): value is Observation {
  if (typeof value !== "object" || value === null) return false;
  const row = value as Record<string, unknown>;
  return typeof row.address === "string" && /^([0-9a-f]{2}:){5}[0-9a-f]{2}$/i.test(row.address) &&
    typeof row.observed_at === "string" && Number.isFinite(Date.parse(row.observed_at)) &&
    finiteNumber(row.temperature_c) && finiteNumber(row.humidity_percent) &&
    nullableNumber(row.battery_percent) && nullableNumber(row.co2_ppm);
}

export async function loadDevices(signal?: AbortSignal): Promise<Observation[]> {
  const response = await fetch("./api/devices", { signal, cache: "no-store", credentials: "same-origin" });
  if (!response.ok) throw new Error("機器のデータを取得できませんでした。時間をおいて再読み込みしてください。");
  let body: unknown;
  try { body = await response.json(); }
  catch { throw new Error("機器のデータを正しく読み取れませんでした。再読み込みしてください。"); }
  if (typeof body !== "object" || body === null || !("devices" in body) ||
      !Array.isArray(body.devices) || !body.devices.every(isObservation)) {
    throw new Error("機器のデータを正しく読み取れませんでした。再読み込みしてください。");
  }
  return body.devices;
}
