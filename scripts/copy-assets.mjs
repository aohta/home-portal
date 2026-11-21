import { cpSync, rmSync } from "node:fs";
const source = new URL("../apps/web/dist/", import.meta.url);
const destination = new URL("../internal/webui/dist/", import.meta.url);
// Only the generated, ignored bundle is replaced.
rmSync(destination, { recursive: true, force: true });
cpSync(source, destination, { recursive: true });
