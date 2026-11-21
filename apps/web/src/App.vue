<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from "vue";
import dayjs from "dayjs";
import utc from "dayjs/plugin/utc";
import timezone from "dayjs/plugin/timezone";
import ElCard from "element-plus/es/components/card/index";
import ElConfigProvider from "element-plus/es/components/config-provider/index";
import ja from "element-plus/es/locale/lang/ja";
import { applications, devices } from "./catalog";
import { loadDevices } from "./api";
import type { Observation } from "./api";
import { findObservation, formatMeasurement, observationState } from "./observations";
import PortalButton from "./PortalButton.vue";
import PortalMessage from "./PortalMessage.vue";
import DiskUsagePanel from "./DiskUsagePanel.vue";
import arrowRight from "./assets/icons/digital-agency/arrow_right_line.svg?url";

dayjs.extend(utc);
dayjs.extend(timezone);

const observations = ref<Observation[]>([]);
const loading = ref(false);
const error = ref("");
const fetchedAt = ref<number | null>(null);
const now = ref(Date.now());
const stateLabels = { missing: "未取得", fresh: "受信済み", stale: "5分以上更新なし", future: "機器の時刻を確認" };
const cards = computed(() => devices.map(device => {
  const observation = findObservation(observations.value, device.address);
  return { ...device, observation, state: observationState(observation, now.value) };
}));
const today = computed(() => new Intl.DateTimeFormat("ja-JP", {
  timeZone: "Asia/Tokyo", year: "numeric", month: "long", day: "numeric", weekday: "short",
}).format(now.value));
const timeLabel = (value: string | number): string => dayjs(value).tz("Asia/Tokyo").format("M/D HH:mm:ss");
const arrowStyle = { "--portal-icon-image": `url("${arrowRight}")` };
let controller: AbortController | undefined;
let refreshTimer: ReturnType<typeof setInterval>;
let clockTimer: ReturnType<typeof setInterval>;
let disposed = false;

async function refresh(): Promise<void> {
  if (loading.value || disposed) return;
  loading.value = true;
  error.value = "";
  controller = new AbortController();
  const timeout = setTimeout(() => controller?.abort(), 8_000);
  try {
    observations.value = await loadDevices(controller.signal);
    fetchedAt.value = Date.now();
    now.value = fetchedAt.value;
  } catch (cause) {
    if (!disposed) error.value = cause instanceof Error && cause.name !== "AbortError" && cause.name !== "TypeError"
      ? cause.message : "機器のデータを取得できませんでした。時間をおいて再読み込みしてください。";
  } finally {
    clearTimeout(timeout);
    if (!disposed) loading.value = false;
  }
}

onMounted(() => {
  void refresh();
  refreshTimer = setInterval(() => { void refresh(); }, 60_000);
  clockTimer = setInterval(() => { now.value = Date.now(); }, 30_000);
});
onUnmounted(() => {
  disposed = true;
  controller?.abort();
  clearInterval(refreshTimer);
  clearInterval(clockTimer);
});
</script>

<template>
  <ElConfigProvider :locale="ja">
    <a class="portal-skip-link" href="#main">本文へ移動</a>
    <div class="portal-page">
      <header class="portal-topbar">
        <h1 class="portal-topbar__title">ホームポータル</h1>
        <time class="portal-date">{{ today }}</time>
      </header>
      <main id="main" class="portal-main" tabindex="-1">
        <section aria-labelledby="applications-title">
          <ElCard class="portal-surface" shadow="never">
            <h2 id="applications-title" class="portal-section-title">アプリ</h2>
            <nav class="portal-apps" aria-label="家庭内アプリ">
              <a v-for="app in applications" :key="app.href" :href="app.href" class="portal-app-link">
                <div class="portal-app-link__body">
                  <h3>{{ app.name }}</h3>
                  <p>{{ app.description }}</p>
                  <span class="portal-app-link__action">{{ app.action }}</span>
                </div>
                <span class="portal-arrow" :style="arrowStyle" aria-hidden="true"></span>
              </a>
            </nav>
          </ElCard>
        </section>

        <DiskUsagePanel />

        <section aria-labelledby="devices-title">
          <ElCard class="portal-surface" shadow="never">
            <div class="portal-section-heading">
              <div>
                <h2 id="devices-title" class="portal-section-title">機器の状態</h2>
                <p class="portal-section-note">1分ごとに自動更新 · 時刻は日本時間</p>
              </div>
              <PortalButton :loading="loading" @click="refresh">{{ loading ? '読み込み中' : '再読み込み' }}</PortalButton>
            </div>
            <PortalMessage v-if="error" severity="error">
              {{ error }}<span v-if="fetchedAt !== null"> 表示中の値は前回取得したデータです。</span>
            </PortalMessage>
            <p class="portal-fetch-status" role="status" aria-live="polite" aria-atomic="true">
              <template v-if="loading">機器のデータを読み込んでいます。</template>
              <template v-else-if="fetchedAt !== null">表示更新 <time :datetime="new Date(fetchedAt).toISOString()">{{ timeLabel(fetchedAt) }}</time></template>
              <template v-else>機器のデータはまだ取得できていません。</template>
            </p>
            <div class="portal-devices" :aria-busy="loading">
              <article v-for="device in cards" :key="device.address" class="portal-device" :aria-labelledby="`device-${device.address.replaceAll(':', '')}`">
                <div class="portal-device-heading">
                  <div>
                    <h3 :id="`device-${device.address.replaceAll(':', '')}`">{{ device.name }}</h3>
                    <p class="portal-device-model">{{ device.model }}</p>
                  </div>
                  <span class="portal-state" :class="`portal-state--${device.state}`">{{ stateLabels[device.state] }}</span>
                </div>
                <dl class="portal-measurements">
                  <div><dt>温度</dt><dd><span class="portal-value">{{ formatMeasurement(device.observation?.temperature_c, 1) }}</span><span v-if="device.observation" class="portal-unit">℃</span></dd></div>
                  <div><dt>湿度</dt><dd><span class="portal-value">{{ formatMeasurement(device.observation?.humidity_percent) }}</span><span v-if="device.observation" class="portal-unit">%</span></dd></div>
                  <div v-if="device.hasCO2"><dt>CO₂濃度</dt><dd><span class="portal-value">{{ formatMeasurement(device.observation?.co2_ppm) }}</span><span v-if="device.observation?.co2_ppm != null" class="portal-unit">ppm</span></dd></div>
                </dl>
                <dl class="portal-device-details">
                  <div><dt>観測日時</dt><dd><time v-if="device.observation" :datetime="device.observation.observed_at">{{ timeLabel(device.observation.observed_at) }}</time><span v-else>未取得</span></dd></div>
                  <div><dt>電池残量</dt><dd>{{ formatMeasurement(device.observation?.battery_percent) }}<span v-if="device.observation?.battery_percent != null"> %</span></dd></div>
                </dl>
                <p v-if="!device.observation && fetchedAt !== null" class="portal-device-note">まだ観測データがありません。</p>
              </article>
            </div>
          </ElCard>
        </section>
      </main>
    </div>
  </ElConfigProvider>
</template>
