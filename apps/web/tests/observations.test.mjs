import assert from "node:assert/strict";
import { test } from "node:test";
import { findObservation, formatMeasurement, observationState } from "../src/observations.ts";

const now = Date.parse("2026-11-18T00:05:00Z");
const row = { address: "e7:50:54:ef:df:f8", observed_at: "2026-11-18T00:00:00Z", temperature_c: 22.6, humidity_percent: 48, battery_percent: null, co2_ppm: 0 };

test("matches fixed devices without confusing other devices or address case", () => {
  assert.equal(findObservation([row], "E7:50:54:EF:DF:F8"), row);
  assert.equal(findObservation([row], "B0:E9:FE:54:40:90"), undefined);
});

test("distinguishes missing, stale, fresh and future observations", () => {
  assert.equal(observationState(undefined, now), "missing");
  assert.equal(observationState(row, now), "stale");
  assert.equal(observationState(row, now - 1), "fresh");
  assert.equal(observationState({ ...row, observed_at: "2026-11-18T00:05:31Z" }, now), "future");
});

test("missing measurements never become zero", () => {
  assert.equal(formatMeasurement(null), "未取得");
  assert.equal(formatMeasurement(undefined), "未取得");
  assert.equal(formatMeasurement(0), "0");
  assert.equal(formatMeasurement(0, 1), "0.0");
  assert.equal(formatMeasurement(22.6, 1), "22.6");
  assert.equal(formatMeasurement(1234), "1,234");
});
