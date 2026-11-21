import assert from "node:assert/strict";
import { test } from "node:test";
import { diskSnapshotStale, formatBytes, loadDisks } from "../src/disks.ts";

const snapshot = { collected_at: "2026-11-21T00:00:00Z", stale: false, filesystems: [
  { name: "Samba", path: "/share", status: "ok", total_bytes: 1024, used_bytes: 1000, available_bytes: 0 },
  { name: "missing", path: "/missing", status: "error", total_bytes: null, used_bytes: null, available_bytes: null },
] };

test("loads capacity independently and preserves zero and unavailable entries", async t => {
  t.mock.method(globalThis, "fetch", async (path, options) => {
    assert.equal(path, "./api/disks");
    assert.equal(options.cache, "no-store");
    return Response.json(snapshot);
  });
  assert.deepEqual(await loadDisks(), snapshot);
});

test("rejects unavailable, incomplete and impossible disk snapshots", async t => {
  for (const response of [new Response("failure", { status: 503 }), Response.json({}),
    Response.json({ ...snapshot, filesystems: [] }),
    Response.json({ ...snapshot, filesystems: [{ ...snapshot.filesystems[0], available_bytes: 1024 }] }),
    Response.json({ ...snapshot, collected_at: "invalid" }),
  ]) {
    t.mock.method(globalThis, "fetch", async () => response);
    await assert.rejects(loadDisks(), /取得できません/);
  }
});

test("cached results become stale even when the subsequent request fails", () => {
  const collectedAt = Date.parse(snapshot.collected_at);
  assert.equal(diskSnapshotStale(snapshot, collectedAt + 899_999), false);
  assert.equal(diskSnapshotStale(snapshot, collectedAt + 900_000), true);
  assert.equal(diskSnapshotStale({ ...snapshot, stale: true }, collectedAt), true);
});

test("capacity units distinguish missing values from an actually full filesystem", () => {
  assert.equal(formatBytes(null), "未取得");
  assert.equal(formatBytes(0), "0 B");
  assert.equal(formatBytes(1024), "1 KiB");
  assert.equal(formatBytes(1.5 * 1024 ** 3), "1.5 GiB");
});
