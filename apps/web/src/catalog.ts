// Device identity is fixed. Observations always come from homehub-api.
// Keep addresses aligned with home-server/inventory.yml.
export const devices = [
  { address: "E7:50:54:EF:DF:F8", name: "温湿度計", model: "Meter Plus", hasCO2: false },
  { address: "B0:E9:FE:54:40:90", name: "CO₂モニター", model: "Meter Pro CO₂", hasCO2: true },
] as const;

export const applications = [
  { name: "勤怠管理", description: "勤務時間の入力・月次集計", href: "/kintai/", action: "勤怠を開く" },
  { name: "料理・レシピ", description: "レシピの検索・管理", href: "/cooking/", action: "料理アプリを開く" },
] as const;
