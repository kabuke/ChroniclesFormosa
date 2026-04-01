package i18n

import (
	"encoding/json"
	"io/ioutil"
)

type Language string

const (
	LangEnUS Language = "en_US"
	LangZhTW Language = "zh_TW"
	LangZhCN Language = "zh_CN"
	LangJaJP Language = "ja_JP"
)

type LanguageManager struct {
	current Language
	bundle  map[Language]map[string]string
}

var Global = &LanguageManager{
	current: LangEnUS, // 預設為英文
	bundle:  make(map[Language]map[string]string),
}

func (m *LanguageManager) Init() {
	// 初始化四大語系
	m.bundle[LangEnUS] = make(map[string]string)
	m.bundle[LangZhTW] = make(map[string]string)
	m.bundle[LangZhCN] = make(map[string]string)
	m.bundle[LangJaJP] = make(map[string]string)

	// 填充基礎文字 (作為 Fallback)
	m.seedAll()
}

func (m *LanguageManager) SetLanguage(l Language) {
	m.current = l
}

func (m *LanguageManager) GetText(id string) string {
	if langMap, ok := m.bundle[m.current]; ok {
		if val, ok := langMap[id]; ok {
			return val
		}
	}
	// 如果當前語系找不到，退回到英文
	if val, ok := m.bundle[LangEnUS][id]; ok {
		return val
	}
	return "[" + id + "]"
}

func (m *LanguageManager) LoadJSON(l Language, path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	var temp map[string]string
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}
	for k, v := range temp {
		m.bundle[l][k] = v
	}
	return nil
}

func (m *LanguageManager) seedAll() {
	// English (Primary for debugging as it works with DebugPrint)
	e := m.bundle[LangEnUS]
	e["LOGIN_TITLE"] = "LOGIN - Chronicles Formosa"
	e["ACCOUNT"] = "Account"
	e["PASSWORD"] = "Password"
	e["LOGIN_BTN"] = "Login"
	e["STATUS_ONLINE"] = "Online"
	e["STATUS_OFFLINE"] = "Offline"
	e["STATUS_CONNECTING"] = "Connecting..."
	e["VERIFYING"] = "Verifying..."
	e["ENTER_TO_LOGIN"] = "Press ENTER to Login"
	e["SWITCH_TO_REGISTER"] = "Need account? Press TAB to Switch"
	e["SWITCH_TO_LOGIN"] = "Have account? Press TAB to Switch"
	e["REGISTER_TITLE"] = "SIGNUP - Chronicles Formosa"
	e["CONFIRM_PASSWORD"] = "Confirm Pwd"
	e["NICKNAME"] = "Nickname"
	e["FACTION"] = "Faction (1-3)"
	e["REGISTER_BTN"] = "Register"

	// 繁體中文
	tw := m.bundle[LangZhTW]
	tw["LOGIN_TITLE"] = "登入 - 台灣三國誌"
	tw["ACCOUNT"] = "帳號"
	tw["PASSWORD"] = "密碼"
	tw["LOGIN_BTN"] = "登入"
	tw["STATUS_ONLINE"] = "連線成功"
	tw["STATUS_OFFLINE"] = "斷開連線"
	tw["STATUS_CONNECTING"] = "正在重連..."
	tw["VERIFYING"] = "正在驗證身份..."
	tw["ENTER_TO_LOGIN"] = "按 ENTER 登入"
	tw["SWITCH_TO_REGISTER"] = "沒有帳號？按 TAB 切換註冊"
	tw["SWITCH_TO_LOGIN"] = "已有帳號？按 TAB 切換登入"
	tw["REGISTER_TITLE"] = "註冊 - 台灣三國誌"
	tw["CONFIRM_PASSWORD"] = "確認密碼"
	tw["NICKNAME"] = "遊戲暱稱"
	tw["FACTION"] = "陣營 (1-3)"
	tw["REGISTER_BTN"] = "註冊"
	tw["DIPLO_TITLE"] = "外交與合約"
	tw["DIPLO_TARGET"] = "目標物件"
	tw["DIPLO_ALLIANCE"] = "建立結盟"
	tw["DIPLO_MARRIAGE"] = "提議聯姻"
	tw["DIPLO_BLOOD"] = "義結金蘭"
	tw["DIPLO_CONFIRM_TITLE"] = "外交請求"
	tw["ACCEPT"] = "接受"
	tw["REJECT"] = "拒絕"
	tw["STAMINA_INSUFFICIENT"] = "精力不足，請稍事休息。"
	tw["DIPLO_ALLIANCE_SUCCESS"] = "結盟協議已達成"
	tw["DIPLO_MARRIAGE_SUCCESS"] = "聯姻慶典圓滿完成"
	tw["DIPLO_RECONCILE_SUCCESS"] = "理番和議達成，獲得勇士效忠"
	tw["INTEL_PANEL_TITLE"] = "廟口傳聞錄"

	// 簡體中文
	cn := m.bundle[LangZhCN]
	cn["LOGIN_TITLE"] = "登錄 - 台湾三国志"
	cn["ACCOUNT"] = "账号"
	cn["PASSWORD"] = "密码"
	cn["LOGIN_BTN"] = "登錄"
	cn["STATUS_ONLINE"] = "连接成功"
	cn["STATUS_OFFLINE"] = "斷開连接"
	cn["STATUS_CONNECTING"] = "正在重连..."
	cn["VERIFYING"] = "正在验证身份..."
	cn["ENTER_TO_LOGIN"] = "按 ENTER 登錄"

	// 日本語
	ja := m.bundle[LangJaJP]
	ja["LOGIN_TITLE"] = "ログイン - 台湾三国志"
	ja["ACCOUNT"] = "アカウント"
	ja["PASSWORD"] = "パスワード"
	ja["LOGIN_BTN"] = "ログイン"
	ja["STATUS_ONLINE"] = "接続済み"
	ja["STATUS_OFFLINE"] = "オフライン"
	ja["STATUS_CONNECTING"] = "再接続中..."
	ja["VERIFYING"] = "認証中..."
	ja["ENTER_TO_LOGIN"] = "ENTERでログイン"
}
