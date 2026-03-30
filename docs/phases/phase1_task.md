# Phase 1 工作清單：基礎架構 (Core Infrastructure)

> **實作計劃**：[phase1_plan.md](./phase1_plan.md)

---

## 修訂紀錄

| 版本 | 日期       | 變更描述 |
| :--- | :---       | :---     |
| 1.1  | 2026-03-30 | 借鑑 DiceTower，更新為三層封包+ECDH+Seq/Ack 架構。 |
| 1.0  | 2026-03-30 | 初版建立。 |

---

## 一、專案初始化

- [x] 1.1 建立 `go.mod`（module 名稱：`github.com/kabuke/ChroniclesFormosa`）
- [x] 1.2 建立完整目錄結構（server/、client/、proto/、resource/、common/crypto/、config/）
- [x] 1.3 建立 `.gitignore`（排除 binary、*.db、.DS_Store）
- [x] 1.4 安裝依賴：`kcp-go/v5`、`ebiten/v2`、`protobuf`、`modernc.org/sqlite`、`zap`
- [x] 1.5 建立 `config/config.go`（JSON 外部化配置，借鑑 DiceTower `config.go`）
- [x] 1.6 建立 `config.json` 預設配置檔

---

## 二、Protobuf 協定

- [x] 2.1 建立 `proto/message.proto`（三層封包架構，借鑑 DiceTower）
  - [x] 2.1.1 定義 `Packet` + `TransferEncrypted` + `SystemCodes`（Wire 層）
  - [x] 2.1.2 定義 `KeyExchangeRequest/Response`、`ResumeSessionRequest/Response`（握手/重連）
  - [x] 2.1.3 定義 `Header`（Seq/Ack/SessionID/OriginSessionId/WinSize）
  - [x] 2.1.4 定義 `Heartbeat`
  - [x] 2.1.5 定義 `Envelope`（OneOf payload：Login/LoginResponse/Register/Ping/Pong/Chat/Village/...）
  - [x] 2.1.6 定義 `Login` / `LoginResponse` / `Register`
  - [x] 2.1.7 定義 `ChatMessage` / `ChatChannelType` / `IntelHint`
  - [x] 2.1.8 定義 `VillageAction` + 子訊息 (`VillageInfoReq/Resp`, `VillageJoinReq/Resp`)
  - [x] 2.1.9 定義 `AOIUpdate` / `MapSync`（基礎 AOI 同步）
- [x] 2.2 建立 `proto/gen_proto.sh` 腳本（借鑑 DiceTower 的跨平台腳本）
- [ ] 2.3 執行腳本，確認 `resource/message.pb.go` 正確生成

---

## 三、伺服器端

### 3.0 加密套件（直接沿用 DiceTower）
- [x] 3.0.1 `common/crypto/crypto.go`：複製 DiceTower 的 X25519 ECDH + AES-256-GCM 實作
- [x] 3.0.2 確認 `GenerateECDHKeys` / `DeriveSharedSecret` / `EncryptAESGCM` / `DecryptAESGCM` 可用

### 3.1 Gateway 層（借鑑 DiceTower `network/connection.go`）
- [x] 3.1.1 `server/network/connection.go`：HandleConnection 完整實作
  - [x] 握手迴圈（KeyExchange / Resume）
  - [x] KCP 參數設定（從 config.json 讀取）
  - [x] 主訊息迴圈（解密→Envelope→Dispatch）
  - [x] 心跳處理（Ping→Pong）
  - [x] Seq 去重 + Ack 確認
- [x] 3.1.2 `readPacket` / `sendPacket`：Length-Prefix Framing
- [x] 3.1.3 `sendEncrypted`：加密發送函數
- [x] 3.1.4 `tryFlush`：Flush Outbox 函數（滑動窗口控制）

### 3.2 Session 管理（借鑑 DiceTower `session/manager.go`）
- [x] 3.2.1 `server/session/manager.go`：UserSession 結構體
- [x] 3.2.2 SessionManager (Singleton)：CreateSession / GetSession / GetSessionByUsername
- [x] 3.2.3 QueueMessage：分配 Seq + 加入 History + 加入 Outbox
- [x] 3.2.4 Acknowledge：清除已確認的 Outbox 訊息
- [x] 3.2.5 FlushOutbox：滑動窗口控制 + 發送
- [x] 3.2.6 ForwardQueue + forwardLoop (20 TPS)
- [x] 3.2.7 GC：清理 60 分鐘不活動的 Session
- [x] 3.2.8 SaveSessions / LoadSessions：SQLite 持久化

### 3.3 Handler 層（HandleEnvelope 三函數模式）
- [x] 3.3.1 `server/handler/handler.go`：HandleEnvelope(env, send, broadcast)
- [x] 3.3.2 `server/handler/auth.go`：登入驗證（SHA-256 密碼比對）+ LoginResponse
- [x] 3.3.3 `server/handler/village_handler.go`：查詢/加入庄頭（VillageAction 分派）

### 3.4 Logic 層
- [x] 3.4.1 `server/logic/village/village.go`：庄頭基礎邏輯（建立、加入、查詢）
- [x] 3.4.2 `server/logic/village/economy.go`：資源產出基礎計算

### 3.5 Model 層
- [x] 3.5.1 `server/model/player.go`：玩家領域模型
- [x] 3.5.2 `server/model/village.go`：庄頭領域模型
- [x] 3.5.3 `server/model/tile.go`：地圖格位模型

### 3.6 Repository 層
- [x] 3.6.1 `server/repo/interface.go`：定義 PlayerRepo、VillageRepo、MapRepo 介面
- [x] 3.6.2 `server/repo/player_repo.go`：SQLite 實作
- [x] 3.6.3 `server/repo/village_repo.go`：SQLite 實作
- [x] 3.6.4 `server/repo/map_repo.go`：記憶體 + 快照實作

### 3.7 AOI 層
- [x] 3.7.1 `server/aoi/manager.go`：AOI Manager 基礎框架
- [x] 3.7.2 `server/aoi/hive.go`：五大節點定義（基隆/台中/台南/台東/澎湖）

### 3.8 伺服器進入點
- [x] 3.8.1 `server/main.go`：LoadConfig → InitDB → KCP ListenWithOptions → AcceptKCP Loop → go HandleConnection

---

## 四、客戶端

### 4.1 網路層（借鑑 DiceTower `client/network/client.go`）
- [x] 4.1.1 `client/network/client.go`：KCP 連線 + 狀態機
  - [x] Connect / doConnect
  - [x] handshake()：ECDH 密鑰交換
  - [x] resume()：斷線重連
  - [x] listen()：監聽迴圈（解密→去重→Ack→Callback dispatch）
  - [x] heartbeat()：每 1 秒 Ping
  - [x] reconnect()：Backoff 重試
- [x] 4.1.2 SendEnvelope：waitingQueue + 滑動窗口 + sendRaw 加密發送
- [x] 4.1.3 handleAck / handleEcho / migratePendingQueue
- [x] 4.1.4 Callback 註冊機制（RegisterLoginResponseCallback 等）

### 4.2 場景管理
- [ ] 4.2.1 `client/scene/scene_manager.go`：Scene 介面 + SceneManager
- [ ] 4.2.2 `client/scene/login_scene.go`：登入畫面（帳號/密碼輸入）
- [ ] 4.2.3 `client/scene/map_scene.go`：大地圖場景（Chunk 載入）

### 4.3 大地圖渲染
- [ ] 4.3.1 Chunk 資料結構定義（5×5 = 25 Chunks）
- [ ] 4.3.2 視野中心 3×3 Chunk 載入邏輯
- [ ] 4.3.3 Camera 控制（拖曳平移、滾輪縮放）
- [ ] 4.3.4 Tile 渲染（地形瓦片繪製）
- [ ] 4.3.5 離屏緩衝 (Offscreen Buffer) 優化

### 4.4 UI 組件
- [ ] 4.4.1 `client/ui/rounded_rect.go`：DrawFilledRoundedRect
- [ ] 4.4.2 `client/ui/theme.go`：GlobalThemeManager（日/夜模式 + 陣營配色）
- [ ] 4.4.3 `client/ui/navbar.go`：GlobalNavbar（連線狀態、場景標題）
- [ ] 4.4.4 `client/ui/toast.go`：GlobalToastManager（成功/錯誤/警告）
- [ ] 4.4.5 `client/ui/keyboard.go`：GuiKeyboard（登入用）
- [ ] 4.4.6 `client/ui/form.go`：BuildGenericForm（登入表單）

### 4.5 i18n
- [ ] 4.5.1 `client/i18n/manager.go`：LanguageManager 核心
- [ ] 4.5.2 `client/i18n/zh_TW.json`：繁體中文翻譯（登入、系統訊息）

### 4.6 客戶端進入點
- [ ] 4.6.1 `client/main.go`：Ebiten Game Loop 啟動
- [ ] 4.6.2 `client/config/client_config.go`：伺服器位址、視窗大小等

---

## 五、資源檔案

- [ ] 5.1 準備基礎地形瓦片素材（草地、山林、水域、城鎮）
- [ ] 5.2 準備宣紙底紋 UI 素材
- [ ] 5.3 準備基礎字體檔（支援繁中的明體/書法體）

---

## 六、測試

- [ ] 6.1 Logic 層單元測試
  - [ ] 6.1.1 `village_test.go`：庄頭建立/加入邏輯
  - [ ] 6.1.2 `economy_test.go`：資源計算
- [ ] 6.2 Handler 整合測試
  - [ ] 6.2.1 登入流程端對端測試
  - [ ] 6.2.2 庄頭加入流程測試
- [ ] 6.3 Client-Server 連線測試
  - [ ] 6.3.1 KCP 握手成功
  - [ ] 6.3.2 心跳與斷線重連

---

## 七、驗收 Checklist

- [ ] `go build ./server` 成功
- [ ] `go build ./client` 成功
- [ ] 客戶端 → 伺服器 KCP 握手成功
- [ ] 登入流程完整（帳號/密碼 → 認證通過 → 進入地圖）
- [ ] 大地圖流暢滾動，Chunk 自動載入
- [ ] Navbar 顯示延遲 ms
- [ ] Toast 正常運作
- [ ] i18n `zh_TW` 文字正確顯示

---

*完成後進入 [Phase 2 工作清單](./phase2_task.md)。*
