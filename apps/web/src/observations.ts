import type { Observation } from "./api.ts";

export function findObservation(rows: readonly Observation[], address: string): Observation | undefined {
  return rows.find(row => row.address.toUpperCase() === address.toUpperCase());
}

export function observationState(observation: Observation | undefined, now: number): "missing" | "fresh" | "stale" | "future" {
  if (!observation) return "missing";
  const age = now - Date.parse(observation.observed_at);
  if (age < -30_000) return "future";
  return age >= 300_000 ? "stale" : "fresh";
}

export function formatMeasurement(value: number | null | undefined, digits = 0): string {
  if (value === null || value === undefined || !Number.isFinite(value)) return "未取得";
  return new Intl.NumberFormat("ja-JP", { minimumFractionDigits: digits, maximumFractionDigits: digits }).format(value);
}
