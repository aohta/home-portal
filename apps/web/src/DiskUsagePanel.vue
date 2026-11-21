<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from "vue";
import ElCard from "element-plus/es/components/card/index";
import PortalButton from "./PortalButton.vue";
import PortalMessage from "./PortalMessage.vue";
import { diskSnapshotStale, formatBytes, loadDisks } from "./disks";
import type { DiskSnapshot } from "./disks";

const snapshot = ref<DiskSnapshot | null>(null);
const loading = ref(false);
const error = ref("");
const now = ref(Date.now());
const stale = computed(() => snapshot.value !== null && diskSnapshotStale(snapshot.value, now.value));
const collectedAt = computed(() => snapshot.value ? new Intl.DateTimeFormat("ja-JP", {
  timeZone: "Asia/Tokyo", month: "numeric", day: "numeric", hour: "2-digit", minute: "2-digit", second: "2-digit", hourCycle: "h23",
}).format(new Date(snapshot.value.collected_at)) : "");
let controller: AbortController | undefined;
let timer: ReturnType<typeof setInterval>;
let disposed = false;

async function refresh(): Promise<void> {
  if (loading.value || disposed) return;
  now.value = Date.now();
  loading.value = true;
  error.value = "";
  controller = new AbortController();
  const timeout = setTimeout(() => controller?.abort(), 8_000);
  try { snapshot.value = await loadDisks(controller.signal); }
  catch (cause) {
    if (!disposed) error.value = cause instanceof Error && cause.name !== "AbortError" && cause.name !== "TypeError"
      ? cause.message : "空き容量の結果を取得できませんでした。時間をおいて再読み込みしてください。";
  } finally { clearTimeout(timeout); if (!disposed) loading.value = false; }
}

onMounted(() => { void refresh(); timer = setInterval(() => { void refresh(); }, 60_000); });
onUnmounted(() => { disposed = true; controller?.abort(); clearInterval(timer); });
</script>

<template>
  <section aria-labelledby="disks-title">
    <ElCard class="portal-surface" shadow="never">
      <div class="portal-section-heading">
        <div><h2 id="disks-title" class="portal-section-title">ストレージ</h2><p class="portal-section-note">5分ごとの収集結果 · 時刻は日本時間</p></div>
        <PortalButton :loading="loading" aria-label="ストレージの空き容量を再読み込み" @click="refresh">{{ loading ? '読み込み中' : '再読み込み' }}</PortalButton>
      </div>
      <PortalMessage v-if="error" severity="error">{{ error }}<span v-if="snapshot"> 表示中の値は前回取得した結果です。</span></PortalMessage>
      <p class="portal-fetch-status" role="status" aria-live="polite" aria-atomic="true">
        <template v-if="loading">空き容量の結果を読み込んでいます。</template>
        <template v-else-if="!snapshot">空き容量の結果はまだ取得できていません。</template>
        <template v-else>収集日時 <time :datetime="snapshot.collected_at">{{ collectedAt }}</time><span v-if="stale" class="portal-stale-note"> · 15分以上更新なし</span></template>
      </p>
      <div v-if="snapshot" class="portal-disks" :aria-busy="loading">
        <article v-for="(disk, index) in snapshot.filesystems" :key="disk.path" class="portal-disk" :aria-labelledby="`disk-${index}`">
          <h3 :id="`disk-${index}`">{{ disk.name }}</h3>
          <p v-if="disk.status === 'error'" class="portal-disk-error">この収集先の容量を取得できませんでした。</p>
          <dl v-else class="portal-measurements portal-disk-measurements">
            <div><dt>空き容量</dt><dd class="portal-value">{{ formatBytes(disk.available_bytes) }}</dd></div>
            <div><dt>全体の容量</dt><dd>{{ formatBytes(disk.total_bytes) }}</dd></div>
            <div><dt>使用済み</dt><dd>{{ formatBytes(disk.used_bytes) }}</dd></div>
          </dl>
        </article>
      </div>
    </ElCard>
  </section>
</template>
