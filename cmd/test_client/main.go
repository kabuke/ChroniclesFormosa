package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"io"
	"log"
	"os"
	"time"

	"github.com/kabuke/ChroniclesFormosa/common/crypto"
	pb "github.com/kabuke/ChroniclesFormosa/resource"
	"github.com/xtaci/kcp-go/v5"
	"google.golang.org/protobuf/proto"
)

func sendProto(conn *kcp.UDPSession, pkt *pb.Packet) {
	data, _ := proto.Marshal(pkt)
	length := uint32(len(data))
	buf := make([]byte, 4+length)
	binary.BigEndian.PutUint32(buf[:4], length)
	copy(buf[4:], data)
	_, _ = conn.Write(buf)
}

func readLengthPrefixed(conn *kcp.UDPSession) ([]byte, error) {
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

func main() {
	addr := "127.0.0.1:8999"
	if len(os.Args) > 1 {
		addr = os.Args[1]
	}

	conn, err := kcp.DialWithOptions(addr, nil, 10, 3)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	log.Println("[Client] Connected to", addr)

	// ==========================================
	// 1. Handshake Phase (明文交換)
	// ==========================================
	priv, pubKeyBytes, err := crypto.GenerateECDHKeys()
	if err != nil {
		log.Fatal("Gen key error:", err)
	}

	req := &pb.Packet{
		Payload: &pb.Packet_KeyExchangeReq{
			KeyExchangeReq: &pb.KeyExchangeRequest{
				PublicKey: pubKeyBytes,
			},
		},
	}
	sendProto(conn, req)
	log.Println("[Client] 🟢 Sent KeyExchangeReq (My PublicKey)")

	respData, err := readLengthPrefixed(conn)
	if err != nil {
		log.Fatal("Read handshake resp error:", err)
	}

	var respPkt pb.Packet
	if err := proto.Unmarshal(respData, &respPkt); err != nil {
		log.Fatal("Unmarshal error:", err)
	}

	var sessionID string
	var secret []byte
	switch p := respPkt.Payload.(type) {
	case *pb.Packet_KeyExchangeResp:
		sessionID = p.KeyExchangeResp.SessionId
		serverPubKey := p.KeyExchangeResp.PublicKey
		log.Printf("[Client] 🟢 Handshake matched. SessionID=%s\n", sessionID)
		
		secret, err = crypto.DeriveSharedSecret(priv, serverPubKey)
		if err != nil {
			log.Fatal("Secret derivation failed:", err)
		}
	default:
		log.Fatal("Unexpected packet response format in handshake.")
	}

	// 驗證生成的 SessionID 規則相同 (Server Hash)
	hash := sha256.Sum256(pubKeyBytes)
	hashStr := base64.RawURLEncoding.EncodeToString(hash[:])
	if len(hashStr) > 8 {
		hashStr = hashStr[:8]
	}
	log.Printf("[Client] 🟢 Derived Shared Secret AESGCM Key. (Client Prefix: %s)", hashStr)

	// ==========================================
	// 2. Encryption Phase (AESGCM 發送 Login/Register Envelope)
	// ==========================================
	now := time.Now().UnixMilli()
	
	// 先嘗試註冊 (長度 8-32)
	regEnv := &pb.Envelope{
		Header: &pb.Header{
			Seq:       1,
			SessionId: sessionID,
		},
		Payload: &pb.Envelope_Register{
			Register: &pb.Register{
				Username:        "DummyTest8",
				Password:        "Password123",
				ConfirmPassword: "Password123",
				FactionId:       1,
			},
		},
	}
	sendEncryptedPayload(conn, regEnv, secret)
	log.Println("[Client] 🟢 Sent Encrypted Register Payload")

	// 接著發起 Login
	loginEnv := &pb.Envelope{
		Header: &pb.Header{
			Seq:       2,
			SessionId: sessionID,
		},
		Payload: &pb.Envelope_Login{
			Login: &pb.Login{
				Username: "DummyTest8",
				Password: "Password123",
			},
		},
	}
	sendEncryptedPayload(conn, loginEnv, secret)
	log.Println("[Client] 🟢 Sent Encrypted Login Payload")

	// ==========================================
	// 3. 收取回音、解密並計算 RTT
	// ==========================================
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	for {
		recvData, err := readLengthPrefixed(conn)
		if err != nil {
			log.Fatal("Read encrypted error:", err)
		}
		
		var recvPkt pb.Packet
		if err := proto.Unmarshal(recvData, &recvPkt); err != nil {
			continue
		}

		if enc, ok := recvPkt.Payload.(*pb.Packet_Encrypted); ok {
			plain, err := crypto.DecryptAESGCM(enc.Encrypted.Data, enc.Encrypted.Nonce, secret)
			if err != nil {
				log.Println("[Client] Decryption error:", err)
				continue
			}
			
			var recvEnv pb.Envelope
			if err := proto.Unmarshal(plain, &recvEnv); err == nil {
				
				if loginResp, ok := recvEnv.Payload.(*pb.Envelope_LoginResponse); ok {
					log.Printf("[Client] 🟢 Decrypted LoginResponse: Success=%v, Msg=%s\n", loginResp.LoginResponse.Success, loginResp.LoginResponse.Message)
					if loginResp.LoginResponse.Success {
						// 登入成功後，發送加入村莊請求（替換原先的世界聊天）
						joinEnv := &pb.Envelope{
							Header: &pb.Header{
								Seq:       3,
								SessionId: sessionID,
							},
							Payload: &pb.Envelope_Village{
								Village: &pb.VillageAction{
									Action: &pb.VillageAction_JoinReq{
										JoinReq: &pb.VillageJoinReq{
											VillageId: 1, // 嘗試加入「打狗」
										},
									},
								},
							},
						}
						sendEncryptedPayload(conn, joinEnv, secret)
					}
				} else if chatMsg, ok := recvEnv.Payload.(*pb.Envelope_Chat); ok {
					log.Printf("[Client] 🟢 Decrypted Server Response: [%s]: %s\n", chatMsg.Chat.Sender, chatMsg.Chat.Content)
				} else if villageResp, ok := recvEnv.Payload.(*pb.Envelope_Village); ok {
					if joinResp, ok2 := villageResp.Village.Action.(*pb.VillageAction_JoinResp); ok2 {
						log.Printf("[Client] 🟢 Decrypted VillageJoinResp: Success=%v, Msg=%s\n", joinResp.JoinResp.Success, joinResp.JoinResp.Message)
						rtt := time.Now().UnixMilli() - now
						log.Printf("==== 🎉 SUCCESS (Village Joined): RTT %d ms ====\n", rtt)
						return
					}
				}
			}
		}
	}
}

func sendEncryptedPayload(conn *kcp.UDPSession, env *pb.Envelope, secret []byte) {
	innerData, _ := proto.Marshal(env)
	cipherData, nonce, _ := crypto.EncryptAESGCM(innerData, secret)

	encPkt := &pb.Packet{
		Payload: &pb.Packet_Encrypted{
			Encrypted: &pb.TransferEncrypted{
				Codes: pb.SystemCodes_EncryptedData,
				Data:  cipherData,
				Nonce: nonce,
			},
		},
	}
	sendProto(conn, encPkt)
}
