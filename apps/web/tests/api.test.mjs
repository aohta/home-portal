import assert from "node:assert/strict";
import { test } from "node:test";
import { loadDevices } from "../src/api.ts";

const row = { address: "E7:50:54:EF:DF:F8", observed_at: "2026-11-18T00:00:00Z", temperature_c: 0, humidity_percent: 0, battery_percent: 0, co2_ppm: null };

test("requests the same-origin API and preserves zero, null and empty lists", async t => {
  t.mock.method(globalThis, "fetch", async (url, options) => {
    assert.equal(url, "./api/devices");
    assert.equal(options.cache, "no-store");
    assert.equal(options.credentials, "same-origin");
    return Response.json({ devices: [row] });
  });
  assert.deepEqual(await loadDevices(), [row]);
  t.mock.method(globalThis, "fetch", async () => Response.json({ devices: [] }));
  assert.deepEqual(await loadDevices(), []);
});

test("rejects failed responses and incomplete observations", async t => {
  for (const response of [
    new Response("private error", { status: 503 }), Response.json({}),
    new Response("<html>unexpected page</html>", { status: 200 }),
    Response.json({ devices: null }), Response.json({ devices: [{ ...row, temperature_c: "22" }] }),
    Response.json({ devices: [{ ...row, observed_at: "invalid" }] }),
  ]) {
    t.mock.method(globalThis, "fetch", async () => response);
    await assert.rejects(loadDevices(), /取得できません|正しく読み取れません/);
  }
});
