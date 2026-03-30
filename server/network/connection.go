package network

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/kabuke/ChroniclesFormosa/common/crypto"
	"github.com/kabuke/ChroniclesFormosa/config"
	pb "github.com/kabuke/ChroniclesFormosa/resource"
	"github.com/kabuke/ChroniclesFormosa/server/handler"
	"github.com/kabuke/ChroniclesFormosa/server/session"

	"github.com/xtaci/kcp-go/v5"
	"google.golang.org/protobuf/proto"
)

func HandleConnection(conn *kcp.UDPSession) {
	defer conn.Close()

	var currentSess *session.UserSession

	// ==========================================
	// 1. Handshake / Resume Loop (明文傳輸)
	// ==========================================
	for currentSess == nil {
		msgData, err := readPacket(conn)
		if err != nil {
			log.Printf("[Handshake] Read Error: %v", err)
			return
		}

		var pkt pb.Packet
		if err := proto.Unmarshal(msgData, &pkt); err != nil {
			log.Println("[Handshake] Unmarshal error:", err)
			return
		}

		switch payload := pkt.Payload.(type) {
		case *pb.Packet_KeyExchangeReq:
			// 收到 Client 公鑰，產生本機公私鑰並推導共享金鑰
			priv, pubBytes, err := crypto.GenerateECDHKeys()
			if err != nil {
				return
			}
			secret, _ := crypto.DeriveSharedSecret(priv, payload.KeyExchangeReq.PublicKey)

			// 衍生 Session ID
			pubKey := payload.KeyExchangeReq.PublicKey
			hash := sha256.Sum256(pubKey)
			hashStr := base64.RawURLEncoding.EncodeToString(hash[:])
			if len(hashStr) > 8 {
				hashStr = hashStr[:8]
			}
			id := fmt.Sprintf("%s.%d", hashStr, time.Now().UnixNano())

			currentSess = session.GetManager().CreateSession(id, secret)
			currentSess.SetConn(conn)
			log.Printf("[Handshake] New Session: %s", id)

			resp := &pb.Packet{
				Payload: &pb.Packet_KeyExchangeResp{
					KeyExchangeResp: &pb.KeyExchangeResponse{
						PublicKey:    pubBytes,
						SessionId:    id,
						MaxClientSeq: currentSess.GetMaxClientSeq(),
					},
				},
			}
			data, _ := proto.Marshal(resp)
			_ = sendPacket(conn, data)
			tryFlush(currentSess)

		case *pb.Packet_ResumeReq:
			// 斷線重連機制
			s := session.GetManager().GetSession(payload.ResumeReq.SessionId)
			if s == nil {
				log.Println("[Resume] Failed: session not found", payload.ResumeReq.SessionId)
				return
			}
			currentSess = s
			log.Printf("[Resume] Session Restored: %s", s.SessionID)

			s.SetConn(conn)
			s.TriggerFlush = func() { tryFlush(s) }
			s.SendEnvelope = func(env *pb.Envelope) {
				s.QueueMessage(env)
				tryFlush(s)
			}

			resp := &pb.Packet{
				Payload: &pb.Packet_ResumeResp{
					ResumeResp: &pb.ResumeSessionResponse{
						MaxClientSeq: s.GetMaxClientSeq(),
					},
				},
			}
			if data, err := proto.Marshal(resp); err == nil {
				_ = sendPacket(conn, data)
			}

			// 確認上次進度並重播歷史封包
			s.Acknowledge(payload.ResumeReq.LastAck)
			for _, env := range s.History {
				if env.Header.Seq > payload.ResumeReq.LastAck {
					_ = sendEncrypted(conn, env, s.SharedSecret, pb.SystemCodes_EncryptedData)
				}
			}
			tryFlush(s)

		default:
			log.Println("[Handshake] Expected handshake or resume, got something else")
			return
		}
	}

	// ==========================================
	// 2. Setup KCP Parameters (來自 config.json)
	// ==========================================
	conn.SetStreamMode(true)
	cfg := config.AppConfig
	if cfg != nil {
		conn.SetNoDelay(cfg.KcpNoDelay, cfg.KcpInterval, cfg.KcpResend, cfg.KcpNc)
	} else {
		conn.SetNoDelay(1, 20, 2, 1) // Fallback
	}
	conn.SetWindowSize(128, 128)
	conn.SetMtu(1350)
	conn.SetACKNoDelay(true)

	currentSess.TriggerFlush = func() { tryFlush(currentSess) }
	currentSess.SendEnvelope = func(env *pb.Envelope) {
		currentSess.QueueMessage(env)
		tryFlush(currentSess)
	}

	// ==========================================
	// 3. Main Message Loop (密文傳輸 & 業務邏輯)
	// ==========================================
	for {
		conn.SetReadDeadline(time.Now().Add(30 * time.Second))

		msgData, err := readPacket(conn)
		if err != nil {
			log.Printf("[Connection] Session %s Disconnected", currentSess.SessionID)
			break
		}

		var pkt = &pb.Packet{}
		if err := proto.Unmarshal(msgData, pkt); err != nil {
			log.Println("[Connection] Unmarshal error:", err)
			continue
		}

		var env = &pb.Envelope{}

		// 解密 AES-GCM
		if enc, ok := pkt.Payload.(*pb.Packet_Encrypted); ok {
			plainData, err := crypto.DecryptAESGCM(enc.Encrypted.Data, enc.Encrypted.Nonce, currentSess.SharedSecret)
			if err != nil {
				log.Println("[Connection] Decrypt error:", err)
				continue
			}

			if err := proto.Unmarshal(plainData, env); err != nil {
				continue
			}
		} else {
			continue // 忽略加密層以外的閒雜封包
		}

		// 處理內建心跳 (Ping->Pong)
		if env.Payload != nil {
			if _, ok := env.Payload.(*pb.Envelope_Ping); ok {
				pong := &pb.Envelope{
					Payload: &pb.Envelope_Pong{
						Pong: &pb.Heartbeat{Timestamp: time.Now().UnixMilli()},
					},
				}
				_ = sendEncrypted(conn, pong, currentSess.SharedSecret, pb.SystemCodes_Pong)
				continue
			}
		}

		// 應用層去重與可靠性 (Seq / Ack)
		if env.Header != nil && env.Header.Seq > 0 {
			maxSeq := currentSess.GetMaxClientSeq()
			if env.Header.Seq <= maxSeq {
				log.Printf("[Connection] Duplicate ignored. %s Seq=%d", currentSess.SessionID, env.Header.Seq)
				
				// 雖然忽略邏輯處理但還是回傳 Ack 以防對方推不進去而卡死
				ackEnv := &pb.Envelope{
					Header: &pb.Header{
						Ack:       maxSeq,
						SessionId: currentSess.SessionID,
					},
					Payload: &pb.Envelope_Ping{
						Ping: &pb.Heartbeat{Timestamp: time.Now().UnixMilli()},
					},
				}
				_ = sendEncrypted(conn, ackEnv, currentSess.SharedSecret, pb.SystemCodes_Ping)
				tryFlush(currentSess)
				continue
			}
			currentSess.UpdateMaxClientSeq(env.Header.Seq)
			currentSess.Acknowledge(env.Header.Ack)
		} else if env.Header != nil {
			currentSess.Acknowledge(env.Header.Ack)
		}

		// 將業務邏輯拋給 3.3 Handler 層
		handler.HandleEnvelope(env, currentSess)
	}

	// 發生 Read 超時或網路錯誤斷下層，保留 Session 的內存狀態交給重連機制或 GC 回收
	currentSess.ClearConn()
}

// readPacket 讀取 Length-Prefixed 封包
func readPacket(conn *kcp.UDPSession) ([]byte, error) {
	header := make([]byte, 4)
	if _, err := io.ReadFull(conn, header); err != nil {
		return nil, err
	}
	length := binary.BigEndian.Uint32(header)

	data := make([]byte, length)
	if _, err := io.ReadFull(conn, data); err != nil {
		return nil, err
	}
	return data, nil
}

// sendPacket 發送 Length-Prefixed 封包
func sendPacket(conn *kcp.UDPSession, data []byte) error {
	length := uint32(len(data))
	buf := make([]byte, 4+length)
	binary.BigEndian.PutUint32(buf[:4], length)
	copy(buf[4:], data)
	_, err := conn.Write(buf)
	return err
}

// tryFlush 嘗試將 Session Outbox 中的訊息發送出去
func tryFlush(s *session.UserSession) {
	winLimit := uint32(128)
	if config.AppConfig != nil {
		winLimit = uint32(config.AppConfig.AppWindowSize)
	}

	s.FlushOutbox(winLimit, func(env *pb.Envelope) error {
		if env.Header == nil {
			env.Header = &pb.Header{}
		}
		// 通知 Client Server 可接受的窗口尺寸
		env.Header.WinSize = winLimit
		return sendEncrypted(s.Conn, env, s.SharedSecret, pb.SystemCodes_EncryptedData)
	})
}

// sendEncrypted 將 Envelope 封裝為 protobuf 資料流，送往 AESGCM 加密，包成 Packet 後以 KCP 寫入
func sendEncrypted(conn *kcp.UDPSession, env *pb.Envelope, secret []byte, code pb.SystemCodes) error {
	if conn == nil {
		return fmt.Errorf("conn is nil")
	}
	innerData, _ := proto.Marshal(env)
	cipherData, nonce, _ := crypto.EncryptAESGCM(innerData, secret)

	encryptedPkt := &pb.Packet{
		Payload: &pb.Packet_Encrypted{
			Encrypted: &pb.TransferEncrypted{
				Codes: code,
				Data:  cipherData,
				Nonce: nonce,
			},
		},
	}
	data, _ := proto.Marshal(encryptedPkt)
	return sendPacket(conn, data)
}
