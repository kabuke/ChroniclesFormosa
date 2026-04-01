# Phase 2 工作清單：社會與族群系統 (Social & Ethnic Systems)

> **實作計劃**：[phase2_plan.md](./phase2_plan.md)
> **前置完成**：Phase 1 全部驗收通過

---

## 修訂紀錄

| 版本 | 日期       | 變更描述 |
| :--- | :---       | :---     |
| 1.3  | 2026-04-01 | 重新實作並通過第五大項（測試）所有項目，修復精力值恢復上限邏輯與 Session 測試註冊機制。 |
| 1.2  | 2026-03-31 | 完成同盟連線可視化、音效切換、按鈕灰化與外交理番邏輯。 |
| 1.1  | 2026-03-31 | 完成聊天 UI、緊張儀 UI、精力值顯示、庄頭管理面板與維穩操作。 |
| 1.0  | 2026-03-30 | 初版建立。 |

---

## 一、Protobuf 擴充

- [x] 1.1 新增 `ChatChannel` 枚舉（VILLAGE / REGION / FACTION / GLOBAL）
- [x] 1.2 定義 `ChatSendReq` / `ChatBroadcast` / `IntelHint`
- [x] 1.3 定義 `TensionUpdate`
- [x] 1.4 定義 `DiplomacyAction` 枚舉 + `DiplomacyReq` / `DiplomacyResp`
- [x] 1.5 定義 `StaminaUpdate`
- [x] 1.6 定義 `VillageElectReq` / `VillageElectResp`
- [x] 1.7 定義 `VillageImpeachReq` / `VillageImpeachResp`
- [x] 1.8 執行 `gen_proto.sh` 確認生成無誤

---

## 二、伺服器端

### 2.1 聊天系統
- [x] 2.1.1 `server/handler/chat_handler.go`：接收/分派聊天訊息
- [x] 2.1.2 `server/logic/social/chat.go`：頻道路由邏輯（四級）
- [x] 2.1.3 實作聊天廣播（只推送給同頻道玩家）

### 2.2 關鍵字情報感測
- [x] 2.2.1 `server/logic/social/intel.go`：關鍵字清單管理
- [x] 2.2.2 實作滑動窗口統計（5 分鐘內關鍵字頻率）
- [x] 2.2.3 實作模糊化處理（方位/動作/時間模糊）
- [x] 2.2.4 實作「廟口說書人」NPC 推播機制

### 2.3 族群緊張儀
- [x] 2.3.1 `server/logic/social/tension.go`：TensionEngine 核心
- [x] 2.3.2 族群比例計算邏輯
- [x] 2.3.3 緊張值每 Tick 更新
- [x] 2.3.4 分類械鬥事件觸發邏輯（Value >= 100）
- [x] 2.3.5 預防操作效果（辦桌/祭祀）

### 2.4 精力值系統
- [x] 2.4.1 `server/logic/stamina/stamina.go`：精力值管理
- [x] 2.4.2 精力值恢復計時器（每小時 +5）
- [x] 2.4.3 所有消耗操作加入精力值檢查

### 2.5 庄頭進階
- [x] 2.5.1 `server/logic/village/election.go`：推舉投票邏輯
- [x] 2.5.2 彈劾投票邏輯（民忠<30 時可發起）
- [x] 2.5.3 職位指派（庄長指定墾首/武師/商賈）
- [x] 2.5.4 副庄長代理機制（模型欄位已就緒）

### 2.6 外交系統
- [x] 2.6.1 `server/logic/social/diplomacy.go`：外交動作處理
- [x] 2.6.2 結盟邏輯（互不攻擊判定基礎）
- [x] 2.6.3 聯姻邏輯（緊張值降低）
- [x] 2.6.4 拜把兄弟（基礎架構）
- [x] 2.6.5 理番和議（消耗糧食換取民忠與武力加成）
- [x] 2.6.6 `server/handler/diplomacy_handler.go`

### 2.7 陣營平衡
- [x] 2.7.1 `server/logic/faction/balance.go`：人數監測
- [x] 2.7.2 天命加成計算（人數最少陣營 buff）
- [x] 2.7.3 內鬨觸發預警邏輯（>40%）
- [x] 2.7.4 NPC 帝國浪潮邏輯（無玩家陣營由 AI 補位與 Buff）

---

## 三、客戶端

### 3.1 聊天 UI
- [x] 3.1.1 `client/ui/chat_panel.go`：聊天面板 UI
- [x] 3.1.2 頻道切換 Tab（庄頭/區域/陣營/全服）
- [x] 3.1.3 訊息泡泡渲染（不同頻道不同配色）
- [x] 3.1.4 GuiKeyboard 整合（聊天輸入）

### 3.2 情報 UI
- [x] 3.2.1 「廟口說書人」NPC 全域廣播對接
- [x] 3.2.2 情報紀錄面板 (IntelPanel)

### 3.3 族群緊張儀 UI
- [x] 3.3.1 `client/ui/tension_meter.go`：火藥引信繪製
- [x] 3.3.2 四等級狀態回饋
- [x] 3.3.3 音效切換（PEACE vs DANGER BGM 動態切換）
- [x] 3.3.4 預防操作快捷鍵與實體按鈕 (B/R)

### 3.4 精力值 UI
- [x] 3.4.1 Navbar 精力值顯示 (⚡ Current/Max)
- [x] 3.4.2 精力不足時按鈕灰化與 Toast 提示

### 3.5 庄頭管理 UI
- [x] 3.5.1 推舉 [E] 按鍵與實體按鈕對接
- [x] 3.5.2 彈劾邏輯回調對接
- [x] 3.5.3 職位清單展示面板 (VillagePanel)
- [x] 3.5.4 庄頭資訊與資源總覽 (木材/糧食/鐵礦/武力)

### 3.6 外交 UI
- [x] 3.6.1 外交動作選單 (ActionMenu 整合 DiplomacyPanel)
- [x] 3.6.2 結盟/聯姻的雙方確認流程 UI (ConfirmDialog)
- [x] 3.6.3 同盟關係可視化 (MapScene 同盟連線繪製)

### 3.7 Network Listener 擴充
- [x] 3.7.1 新增 ChatBroadcast / IntelHint 監聽
- [x] 3.7.2 新增 TensionUpdate 監聽
- [x] 3.7.3 新增 StaminaUpdate 監聽
- [x] 3.7.4 新增 Village 相關響應監聽（List/Info/Members/Join/Elect）

---

## 四、i18n 擴充

- [x] 4.1 聊天系統文字（頻道名稱、系統提示）
- [x] 4.2 情報感測文字（廟口說書人模板）
- [x] 4.3 緊張儀文字（等級描述）
- [x] 4.4 外交文字（動作名稱、結果描述）
- [x] 4.5 精力值文字（不足提示）

---

## 五、測試

- [x] 5.1 聊天頻道路由單元測試
- [x] 5.2 關鍵字情報感測演算法單元測試
- [x] 5.3 緊張儀數值引擎單元測試
- [x] 5.4 精力值恢復計時器測試
- [x] 5.5 推舉/彈劾投票邏輯測試
- [x] 5.6 外交動作整合測試
- [x] 5.7 陣營平衡觸發條件測試
- [x] 5.8 多人聊天壓力測試

---

## 六、驗收 Checklist

- [x] 四級聊天頻道訊息正確傳遞
- [x] 陣營頻道關鍵字觸發模糊情報
- [x] 緊張氣氛儀 UI 正確反映數值
- [x] 緊張值到 100 時觸發分類械鬥事件
- [x] 辦桌/祭祀可降低緊張值/升民忠
- [x] 精力值消耗/恢復顯示正常
- [x] 推舉庄長流程跑通
- [x] 結盟後地圖顯示連線
- [x] 陣營超 40% 觸發內鬨預警

---

*完成後進入 [Phase 3 工作清單](./phase3_task.md)。*
