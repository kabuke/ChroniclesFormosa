# Phase 3 工作清單：天災與救災系統

> **實作計劃**：[phase3_plan.md](./phase3_plan.md)
> **前置完成**：Phase 2 全部驗收通過

---

## 修訂紀錄

| 版本 | 日期       | 變更描述 |
| :--- | :---       | :---     |
| 1.0  | 2026-03-30 | 初版建立。 |

---

## 一、Protobuf 擴充

- [x] 1.1 定義 `DisasterType` 枚舉（EARTHQUAKE / TYPHOON / PLAGUE）
- [x] 1.2 定義 `TyphoonPhase` 枚舉（WARNING / LANDING / FADING）
- [x] 1.3 定義 `ObstacleType` 枚舉（LANDSLIDE / BANDIT / BROKEN_BRIDGE）
- [x] 1.4 定義 `ReliefGrade` 枚舉（PERFECT / GOOD / FAIL）
- [x] 1.5 定義 `EarthquakeNotify` / `TyphoonNotify` / `DisasterWarning`
- [x] 1.6 定義 `ReliefGameStart` / `ReliefTarget` / `ReliefObstacle`
- [x] 1.7 定義 `ReliefDonateReq` / `ReliefRouteSubmit` / `RoutePoint`
- [x] 1.8 定義 `ReliefResult`
- [x] 1.9 Action 枚舉擴充（DISASTER 系列 20-29）
- [x] 1.10 執行 `gen_proto.sh` 確認生成無誤

---

## 二、伺服器端

### 2.1 天災排程器
- [x] 2.1.1 `server/logic/disaster/timer.go`：DisasterTimer 核心
- [x] 2.1.2 地震隨機排程（每賽季 3~5 次）
- [x] 2.1.3 颱風季節判斷（5-10 月限定，每季 2~4 次）
- [x] 2.1.4 預警推播（提前 1 小時）

### 2.2 地震邏輯
- [x] 2.2.1 `server/logic/disaster/earthquake.go`：震度計算
- [x] 2.2.2 影響範圍計算（震央 + 半徑 + 衰減）
- [x] 2.2.3 城牆/糧倉/民忠損傷套用
- [x] 2.2.4 地形改變判定（高震度→新渡口/山路）

### 2.3 颱風邏輯
- [x] 2.3.1 `server/logic/disaster/typhoon.go`：颱風路徑生成
- [x] 2.3.2 海域封鎖邏輯
- [x] 2.3.3 農田產量歸零
- [x] 2.3.4 行軍速度修正（降至 20%）
- [x] 2.3.5 颱風三階段（預警→登陸→消退）狀態機

### 2.4 瘟疫邏輯（基礎版）
- [x] 2.4.1 `server/logic/disaster/plague.go`：觸發條件判定
- [x] 2.4.2 人口持續損失計算
- [x] 2.4.3 藥鋪對抗效果

### 2.5 救災系統
- [x] 2.5.1 `server/logic/disaster/relief.go`：救災發起邏輯
- [x] 2.5.2 庄民精力捐獻 → 道具轉換
- [x] 2.5.3 路線評分演算法（距離、急迫度、障礙排除）
- [x] 2.5.4 獎勵計算（BUFF、隱藏資源發現機率）
- [x] 2.5.5 救災結果廣播

### 2.6 Handler 層
- [x] 2.6.1 `server/handler/disaster_handler.go`
- [x] 2.6.2 處理 ReliefDonateReq
- [x] 2.6.3 處理 ReliefRouteSubmit
- [x] 2.6.4 天災通知推播（全服/區域）

---

## 三、客戶端

### 3.1 視覺特效
- [x] 3.1.1 `client/ui/screen_shake.go`：螢幕震動效果
- [x] 3.1.2 `client/ui/explosion.go`：GuiExplosion 粒子（地震碎石）
- [x] 3.1.3 颱風雨滴粒子系統
- [x] 3.1.4 颱風灰暗濾鏡
- [x] 3.1.5 風向箭頭 UI

### 3.2 預警 UI
- [x] 3.2.1 Navbar 氣候警告閃爍動畫
- [x] 3.2.2 天災倒計時顯示
- [x] 3.2.3 預警 Toast 通知

### 3.3 救災小遊戲場景
- [x] 3.3.1 `client/scene/relief_scene.go`：場景骨架
- [x] 3.3.2 俯瞰圖渲染（村落 + 倉庫 + 道路網）
- [x] 3.3.3 牛車路線繪製交互（滑鼠/觸控拖曳）
- [x] 3.3.4 障礙物渲染（土石流/亂民/斷橋）
- [x] 3.3.5 道具 UI（修橋工具/護衛隊按鈕）
- [x] 3.3.6 倒計時 + 急迫度指示器
- [x] 3.3.7 結算畫面（評分 + 獎勵展示）

### 3.4 庄民支援 UI
- [x] 3.4.1 「支援」按鈕（捐精力）
- [x] 3.4.2 庄長獲得道具的動畫回饋
- [x] 3.4.3 支援者清單顯示

### 3.5 Network Listener 擴充
- [x] 3.5.1 新增 EarthquakeNotify / TyphoonNotify 監聽
- [x] 3.5.2 新增 DisasterWarning 監聽
- [x] 3.5.3 新增 ReliefGameStart 監聽
- [x] 3.5.4 新增 ReliefResult 監聽

---

## 四、i18n 擴充

- [ ] 4.1 天災名稱與描述文字
- [ ] 4.2 預警訊息模板
- [ ] 4.3 救災小遊戲 UI 文字
- [ ] 4.4 獎勵描述文字
- [ ] 4.5 隱藏資源發現文字

---

## 五、音效資源

- [ ] 5.1 地震音效（低頻隆隆聲）
- [ ] 5.2 颱風音效（風雨聲）
- [ ] 5.3 預警音效（鐘聲/銅鑼）
- [ ] 5.4 救災成功音效
- [ ] 5.5 救災失敗音效

---

## 六、測試

- [ ] 6.1 天災排程器頻率測試（模擬一個賽季）
- [ ] 6.2 地震傷害計算單元測試
- [ ] 6.3 颱風路徑生成測試
- [ ] 6.4 救災路線評分演算法單元測試
- [ ] 6.5 獎勵計算測試（BUFF 生效驗證）
- [ ] 6.6 天災 + 戰鬥交互測試（趁災進攻場景）
- [ ] 6.7 小遊戲端對端流程測試

---

## 七、驗收 Checklist

- [ ] 天災在賽季中按預期頻率觸發
- [ ] 預警 1 小時前正確顯示
- [ ] 地震螢幕震動 + 粒子特效
- [ ] 颱風雨滴 + 灰暗濾鏡
- [ ] 牛車糧道小遊戲流程完整
- [ ] 庄民支援 → 庄長道具增加
- [ ] 完美救災 → BUFF 生效 3 天
- [ ] 高震度地震 → 地形改變
- [ ] 所有文字透過 i18n 顯示

---

*Phase 3 完成後，遊戲核心循環完整成型。*
