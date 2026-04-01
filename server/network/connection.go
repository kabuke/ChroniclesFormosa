package network

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"

	"github.com/google/uuid"
	"github.com/kabuke/ChroniclesFormosa/common/crypto"
	"github.com/kabuke/ChroniclesFormosa/server/handler"
	"github.com/kabuke/ChroniclesFormosa/server/session"
	pb "github.com/kabuke/ChroniclesFormosa/resource"
	"github.com/xtaci/kcp-go/v5"
	"google.golang.org/protobuf/proto"
)

// HandleConnection 處理單個 KCP/TCP 連接
func HandleConnection(conn net.Conn) {
	defer conn.Close()

	// 1. 初始化一個匿名的臨時 Session
	sid := uuid.New().String()
	userSession := session.GetManager().CreateSession(sid, nil)
	defer session.GetManager().UnregisterSession(sid)

	// 連結連線與 Session
	kcpConn := conn.(*kcp.UDPSession)
	userSession.SetConn(kcpConn)
	userSession.TriggerFlush = func() {
		if len(userSession.SharedSecret) != 32 {
			// 尚未握手完成或金鑰遺失，不發送加密封包
			return
		}
		userSession.FlushOutbox(128, func(env *pb.Envelope) error {
			if userSession.Conn == nil { return fmt.Errorf("connection lost") }
			plaintext, _ := proto.Marshal(env)
			ciphertext, nonce, err := crypto.EncryptAESGCM(plaintext, userSession.SharedSecret)
			if err != nil { return err }
			
			packet := &pb.Packet{
				Payload: &pb.Packet_Encrypted{
					Encrypted: &pb.TransferEncrypted{
						Codes: pb.SystemCodes_EncryptedData,
						Data:  ciphertext,
						Nonce: nonce,
					},
				},
			}
			b, _ := proto.Marshal(packet)
			sendRaw(conn, b)
			return nil
		})
	}

	log.Printf("[Network] New Connection: %s (ID: %s)", conn.RemoteAddr(), sid)

	for {
		// 讀取 Raw Packet (含 4-byte 長度頭)
		data, err := readRaw(conn)
		if err != nil {
			log.Printf("[Network] Session %s Disconnected: %v", sid, err)
			break
		}

		var packet pb.Packet
		if err := proto.Unmarshal(data, &packet); err != nil {
			log.Printf("[Network] Invalid Packet: %v", err)
			continue
		}

		// 處理握手或加密數據
		switch p := packet.Payload.(type) {
		case *pb.Packet_KeyExchangeReq:
			handleHandshake(conn, userSession, p.KeyExchangeReq)
		case *pb.Packet_Encrypted:
			handleEncrypted(userSession, p.Encrypted)
		case *pb.Packet_ResumeReq:
			// TODO: 實作 Resume 邏輯
		}
	}
}

func readRaw(conn net.Conn) ([]byte, error) {
	lenBuf := make([]byte, 4)
	if _, err := conn.Read(lenBuf); err != nil {
		return nil, err
	}
	length := binary.BigEndian.Uint32(lenBuf)
	data := make([]byte, length)
	_, err := conn.Read(data)
	return data, err
}

func handleHandshake(conn net.Conn, s *session.UserSession, req *pb.KeyExchangeRequest) {
	priv, pub, _ := crypto.GenerateECDHKeys()
	shared, _ := crypto.DeriveSharedSecret(priv, req.PublicKey)

	s.SharedSecret = shared
	
	resp := &pb.Packet{
		Payload: &pb.Packet_KeyExchangeResp{
			KeyExchangeResp: &pb.KeyExchangeResponse{
				PublicKey:    pub,
				SessionId:    s.SessionID,
				MaxClientSeq: 0,
			},
		},
	}
	b, _ := proto.Marshal(resp)
	sendRaw(conn, b)
	log.Printf("[Network] Handshake completed for session %s", s.SessionID)
}

func handleEncrypted(s *session.UserSession, enc *pb.TransferEncrypted) {
	plaintext, err := crypto.DecryptAESGCM(enc.Data, enc.Nonce, s.SharedSecret)
	if err != nil {
		log.Printf("[Network] Decrypt failed: %v", err)
		return
	}

	var env pb.Envelope
	if err := proto.Unmarshal(plaintext, &env); err == nil {
		handler.HandleEnvelope(&env, s)
	}
}

func sendRaw(conn net.Conn, data []byte) {
	lenBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lenBuf, uint32(len(data)))
	conn.Write(lenBuf)
	conn.Write(data)
}
