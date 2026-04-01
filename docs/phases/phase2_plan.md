# Phase 2 實作計劃：社會與族群系統 (Social & Ethnic Systems)

> **上層文件**：[DDD.md](../DDD.md)、[GDD.md](../GDD.md)
> **前置**：[Phase 1](./phase1_plan.md)（基礎架構）必須完成
> **工作清單**：[phase2_task.md](./phase2_task.md)

---

## 修訂紀錄

| 版本 | 日期       | 變更描述 |
| :--- | :---       | :---     |
| 1.6  | 2026-03-31 | 完成外交 UI、情報面板、NPC 補位機制與精力值提示。 |
| 1.5  | 2026-03-31 | 實作社會維穩操作（辦桌、祭祀）與庄頭管理 UI。 |
| 1.3  | 2026-03-31 | 實作外交系統邏輯（結盟、聯姻、拜把）。 |
| 1.2  | 2026-03-31 | 實作庄頭推舉、彈劾與職位指派邏輯。 |
| 1.1  | 2026-03-31 | 實作聊天系統、關鍵字情報感測、族群緊張儀與精力值系統。 |
| 1.0  | 2026-03-30 | 初版建立。 |

---

## 一、目標

在 Phase 1 的基礎架構上，建立萬人互動的核心社會系統。

Phase 2 完成後，應能達成：
1. ✅ 庄頭內完整的 CRUD + 推舉/彈劾制度
2. ✅ 四級聊天頻道（庄頭/區域/陣營/全服）完整運作
3. ✅ 關鍵字情報感測演算法上線
4. ✅ 族群緊張儀（Tension Meter）可視化完成
5. ✅ 陣營動態平衡機制（天命加成 + 內鬨系統）運作
6. ✅ 基礎外交系統（結盟/聯姻/拜把）可用
7. ✅ 精力值系統限制每日行動量

---

## 二、範圍與邊界

### 包含
- 庄頭進階功能（推舉、彈劾、職位指派）
- 聊天系統（四級頻道 + 訊息過濾）
- 關鍵字情報感測（模糊情報生成）
- 族群緊張儀（數值邏輯 + UI）
- 陣營平衡（天命加成、內鬨觸發）
- 外交動作（結盟、聯姻、認祖歸宗、拜把、理番）
- 精力值系統

### 不包含
- 完整戰鬥系統（僅含基礎交戰結算）
- 天災系統（Phase 3）
- 海上貿易完整版
- 劇本系統

---

## 三、技術方案

### 3.1 聊天系統架構

```
Client ──► ChatSendReq(channel, message) ──► Server
                                               │
                                    ┌──────────┼──────────┐
                                    ▼          ▼          ▼
                               庄頭廣播   區域廣播   陣營廣播
                                    │          │          │
                                    │          │    ┌─────▼─────┐
                                    │          │    │ 關鍵字掃描 │
                                    │          │    └─────┬─────┘
                                    │          │          │
                                    ▼          ▼          ▼
                               視聽者收到  視聽者收到  視聽者收到
                                                    + 敵對陣營
                                                      模糊情報
```

#### 關鍵字情報感測演算法
```
1. 維護「敏感關鍵字」清單（進攻、夜襲、府城、出兵...）
2. 滑動窗口：統計陣營頻道最近 5 分鐘內關鍵字出現頻率
3. 當頻率超過閾值 → 生成模糊情報
4. 模糊化處理：
   - 方位模糊（「南方」而非精確座標）
   - 動作模糊（「軍事行動」而非「夜襲」）
   - 時間模糊（「近期」而非精確時間）
5. 透過 NPC「廟口說書人」推播給敵對陣營
```

### 3.2 族群緊張儀

#### 數值引擎
```go
type TensionEngine struct {
    // 每個庄頭獨立計算
    tensions map[int64]*TensionState
}

type TensionState struct {
    Value      int32   // 0 (和平) ~ 100 (爆發)
    MinNan     float32 // 閩南人口比例
    Hakka      float32 // 客家人口比例
    Indigenous float32 // 原民人口比例
    FoodRatio  float32 // 糧食充足率
    Security   float32 // 治安值
}

// 每 Tick 更新
func (e *TensionEngine) Tick() {
    for _, t := range e.tensions {
        delta := calculateDelta(t)
        t.Value = clamp(t.Value + delta, 0, 100)
        if t.Value >= 100 {
            triggerRiot(t)
        }
    }
}
```

#### UI 表現：火藥引信
- 100%：長引信 + 和平背景動畫
- 70%：縮短引信 + 人群聚集
- 40%：短引信 + 火光塗鴉 + 急促鑼鼓
- 0%：爆炸 → 分類械鬥事件

### 3.3 精力值系統

```protobuf
message PlayerStamina {
    int32 current = 1;        // 當前精力
    int32 max = 2;            // 上限 (100)
    int64 last_regen_at = 3;  // 上次恢復時間
}
```

- 伺服器端計算，客戶端僅顯示
- 每小時恢復 5 點
- 所有消耗操作先檢查精力值，不足則拒絕

---

## 四、新增 Proto 訊息

```protobuf
// 聊天
message ChatSendReq {
    ChatChannel channel = 1;  // VILLAGE, REGION, FACTION, GLOBAL
    string message = 2;
}
message ChatBroadcast {
    ChatChannel channel = 1;
    string sender_name = 2;
    string message = 3;
    int64 timestamp = 4;
}
message IntelHint {           // 模糊情報 S2C
    string hint_text = 1;     // i18n key
    string direction = 2;     // 方位
}

// 族群緊張
message TensionUpdate {       // S2C
    int64 village_id = 1;
    int32 tension_value = 2;
    string visual_level = 3;  // PEACE, UNEASY, TENSE, RIOT
}

// 外交
message DiplomacyReq {
    DiplomacyAction action = 1;
    int64 target_village_id = 2;
    int64 target_player_id = 3;
}
message DiplomacyResp {
    bool success = 1;
    string result_desc = 2;
}

// 精力值
message StaminaUpdate {       // S2C
    int32 current = 1;
    int32 max = 2;
}
```

---

## 五、驗收標準

1. [ ] 庄頭推舉投票流程完整可用
2. [ ] 四級聊天頻道訊息正確廣播
3. [ ] 關鍵字情報感測可生成模糊情報
4. [ ] 緊張氣氛儀 UI 正確反映數值（四個等級視覺效果）
5. [ ] 陣營人數超過 40% 時觸發內鬨
6. [ ] 外交動作（結盟/聯姻/拜把）流程完整
7. [ ] 精力值消耗與恢復正確

---

*Phase 2 完成後，進入 [Phase 3](./phase3_plan.md)（天災與救災系統）。*
