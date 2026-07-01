import assert from "node:assert";
import { test } from "node:test";
import { fmtSize } from "./fmt.ts";

test("fmtSize", () => {
  assert.strictEqual(fmtSize(0), "0 B");
  assert.strictEqual(fmtSize(512), "512 B");
  assert.strictEqual(fmtSize(1024), "1.0 KB");
  assert.strictEqual(fmtSize(1536), "1.5 KB");
  assert.strictEqual(fmtSize(1048576), "1.0 MB");
  assert.strictEqual(fmtSize(1073741824), "1.0 GB");
  assert.strictEqual(fmtSize(1073741824 * 5), "5.0 GB");
});
