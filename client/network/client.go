package network

import (
	"crypto/ecdh"
	"encoding/binary"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/kabuke/ChroniclesFormosa/common/crypto"
	pb "github.com/kabuke/ChroniclesFormosa/resource"
	"github.com/xtaci/kcp-go/v5"
	"google.golang.org/protobuf/proto"
)

type ClientState int

const (
	StateDisconnected ClientState = iota
	StateConnecting
	StateConnected
	StateResuming
)

// NetworkClient 為客戶端專用的背景網路通訊層 (Non-blocking)
type NetworkClient struct {
	State      ClientState
	serverAddr string
	conn       *kcp.UDPSession

	// 加密金鑰
	privateKey   *ecdh.PrivateKey
	publicKey    []byte
	sharedSecret []byte
	SessionID    string

	// PING / 網路品質
	RTT int64

	// Seq / Ack
	clientSeq    uint64
	lastAck      uint64 // 收過來自 Server 的最大 Seq
	maxServerAck uint64 // Server 確認收過來自我的最大 Seq

	outbox       []*pb.Envelope
	waitingQueue sync.Map // map[uint64]*pb.Envelope
	sendMutex    sync.Mutex

	// UI 排隊緩衝 (讓 Ebiten 主迴圈去吃)
	incomingQueue []*pb.Envelope
	incomingMut   sync.Mutex

	// 當回到主迴圈時，執行這個方法 (給 SceneManager 註冊用)
	OnEnvelopeReceived func(env *pb.Envelope)

	// 控制器
	stopChan chan struct{}
}

func NewNetworkClient(addr string) *NetworkClient {
	return &NetworkClient{
		State:      StateDisconnected,
		serverAddr: addr,
		stopChan:   make(chan struct{}),
	}
}

// Connect 啟動背景連線
func (c *NetworkClient) Connect() {
	go c.reconnectLoop()
}

func (c *NetworkClient) reconnectLoop() {
	for {
		c.State = StateConnecting
		log.Printf("[Network] 🌐 Connecting to %s...", c.serverAddr)

		// 嘗試建立 KCP UDP 連線
		conn, err := kcp.DialWithOptions(c.serverAddr, nil, 10, 3)
		if err != nil {
			log.Printf("[Network] ❌ Connection failed: %v, retrying in 3s...", err)
			time.Sleep(3 * time.Second)
			continue
		}

		conn.SetStreamMode(true)
		conn.SetWindowSize(1024, 1024)
		conn.SetNoDelay(1, 10, 2, 1)

		c.conn = conn

		// 如果已經有 SessionID，嘗試做 Resume 熱重連
		if c.SessionID != "" {
			c.State = StateResuming
			if err := c.resume(); err != nil {
				log.Printf("[Network] ⚠️ Resume failed (%s), falling back to Handshake.", err)
				c.SessionID = "" // 清除失效 SessionID
				if err := c.handshake(); err != nil {
					c.conn.Close()
					time.Sleep(3 * time.Second)
					continue
				}
			}
		} else {
			// 全新握手 (ECDH Key Exchange)
			if err := c.handshake(); err != nil {
				log.Printf("[Network] ❌ Handshake failed: %v", err)
				c.conn.Close()
				time.Sleep(3 * time.Second)
				continue
			}
		}

		c.State = StateConnected
		log.Printf("[Network] 🟢 Connected! SessionID = %s", c.SessionID)

		// 啟動監聽與讀取背景
		go c.readLoop()
		go c.flushLoop()
		go c.heartbeatLoop()

		// 阻塞直到這條連線中斷
		<-c.stopChan
		log.Println("[Network] 🛑 Connection Lost. Restarting reconnect loop...")
		time.Sleep(2 * time.Second) // 斷線後等一下再重試
	}
}

// ======================= KCP 底層讀寫 =======================

func (c *NetworkClient) sendRaw(data []byte) error {
	lenBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lenBuf, uint32(len(data)))
	if _, err := c.conn.Write(lenBuf); err != nil {
		return err
	}
	_, err := c.conn.Write(data)
	return err
}

func (c *NetworkClient) readPacket() ([]byte, error) {
	lenBuf := make([]byte, 4)
	if _, err := c.conn.Read(lenBuf); err != nil {
		return nil, err
	}
	length := binary.BigEndian.Uint32(lenBuf)
	if length > 10*1024*1024 { // 10MB limit
		return nil, fmt.Errorf("packet too large: %d", length)
	}

	data := make([]byte, length)
	var totalRead uint32 = 0
	for totalRead < length {
		n, err := c.conn.Read(data[totalRead:])
		if err != nil {
			return nil, err
		}
		totalRead += uint32(n)
	}
	return data, nil
}

// ======================= 握手與重連 =======================

func (c *NetworkClient) handshake() error {
	priv, pub, err := crypto.GenerateECDHKeys()
	if err != nil {
		return err
	}
	c.privateKey = priv
	c.publicKey = pub

	req := &pb.Packet{
		Payload: &pb.Packet_KeyExchangeReq{
			KeyExchangeReq: &pb.KeyExchangeRequest{
				PublicKey: pub,
			},
		},
	}
	reqData, _ := proto.Marshal(req)
	if err := c.sendRaw(reqData); err != nil {
		return err
	}

	respData, err := c.readPacket()
	if err != nil {
		return err
	}

	var packet pb.Packet
	if err := proto.Unmarshal(respData, &packet); err != nil {
		return err
	}

	resp, ok := packet.Payload.(*pb.Packet_KeyExchangeResp)
	if !ok {
		return fmt.Errorf("unexpected handshake resp")
	}

	shared, err := crypto.DeriveSharedSecret(c.privateKey, resp.KeyExchangeResp.PublicKey)
	if err != nil {
		return err
	}

	c.sharedSecret = shared
	c.SessionID = resp.KeyExchangeResp.SessionId
	c.clientSeq = resp.KeyExchangeResp.MaxClientSeq + 1 // 從接續開始
	return nil
}

func (c *NetworkClient) resume() error {
	req := &pb.Packet{
		Payload: &pb.Packet_ResumeReq{
			ResumeReq: &pb.ResumeSessionRequest{
				SessionId:     c.SessionID,
				LastAck:       c.lastAck,
				ClientNextSeq: c.clientSeq,
			},
		},
	}
	reqData, _ := proto.Marshal(req)
	if err := c.sendRaw(reqData); err != nil {
		return err
	}

	respData, err := c.readPacket()
	if err != nil {
		return err
	}

	var packet pb.Packet
	if err := proto.Unmarshal(respData, &packet); err != nil {
		return err
	}

	resp, ok := packet.Payload.(*pb.Packet_ResumeResp)
	if !ok {
		return fmt.Errorf("unexpected resume packet")
	}

	c.maxServerAck = resp.ResumeResp.MaxClientSeq
	return nil
}

// ======================= 傳輸與接收 =======================

// SendEnvelope 給前端用的安全發送（扔進 Outbox，由背景加密負責送）
func (c *NetworkClient) SendEnvelope(env *pb.Envelope) {
	if c.State != StateConnected {
		log.Println("[Network] 🚫 Cannot send packet while disconnected")
		return
	}
	c.sendMutex.Lock()
	defer c.sendMutex.Unlock()

	c.clientSeq++
	env.Header = &pb.Header{
		Seq:       c.clientSeq,
		SessionId: c.SessionID,
	}
	c.outbox = append(c.outbox, env)
	// 放進 Slide Window 等待被 Server ACK
	c.waitingQueue.Store(c.clientSeq, env)
}

// readLoop 接收所有來自伺服器的加密 Payload
func (c *NetworkClient) readLoop() {
	defer func() {
		close(c.stopChan) // 通知外層迴圈重新連線
	}()

	for {
		data, err := c.readPacket()
		if err != nil {
			log.Printf("[Network] 🤔 Read dropped: %v", err)
			return
		}

		var packet pb.Packet
		if err := proto.Unmarshal(data, &packet); err != nil {
			log.Printf("[Network] ⚠️ Invalid packet format")
			continue
		}

		encMsg, ok := packet.Payload.(*pb.Packet_Encrypted)
		if !ok {
			log.Printf("[Network] 🛡️ Ignored unencrypted mid-session packet")
			continue
		}

		plaintext, err := crypto.DecryptAESGCM(encMsg.Encrypted.Data, encMsg.Encrypted.Nonce, c.sharedSecret)
		if err != nil {
			log.Printf("[Network] 🛡️ Decrypt Error: %v", err)
			continue
		}

		var env pb.Envelope
		if err := proto.Unmarshal(plaintext, &env); err != nil {
			log.Printf("[Network] 🛡️ Unwrap Envelope Error")
			continue
		}

		// 處理 Seq 去重與 Server 回丟的 ACK
		if env.Header != nil {
			if env.Header.Seq > c.lastAck {
				c.lastAck = env.Header.Seq
			}
			if env.Header.Ack > c.maxServerAck {
				c.maxServerAck = env.Header.Ack
				// 把已經被伺服器收到的人從重傳駐列剔除
				c.clearWaitingQueueUpTo(c.maxServerAck)
			}
		}

		// Ping-Pong 計時器
		if pong, isPong := env.Payload.(*pb.Envelope_Pong); isPong {
			c.RTT = time.Now().UnixMilli() - pong.Pong.Timestamp
			continue
		}

		// 送回主執行緒
		c.incomingMut.Lock()
		c.incomingQueue = append(c.incomingQueue, &env)
		c.incomingMut.Unlock()
	}
}

// flushLoop 控制網路實際送出操作，頻率 20 TPS
func (c *NetworkClient) flushLoop() {
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-c.stopChan:
			return
		case <-ticker.C:
			c.tryFlush()
		}
	}
}

func (c *NetworkClient) tryFlush() {
	if c.conn == nil || c.State != StateConnected {
		return
	}

	c.sendMutex.Lock()
	msgs := c.outbox
	c.outbox = nil
	c.sendMutex.Unlock()

	// 這裡理論上應該處理 Window Size 與重傳，但 Phase 1 為了簡化，只要從 WaitingQueue 抓沒被 ACK 的資料重新投遞即可
	// 把過去 2 秒內沒被 ACK 的拔出來重寄
	c.waitingQueue.Range(func(key, value interface{}) bool {
		seq := key.(uint64)
		if seq > c.maxServerAck {
			// 如果大於 Server 最新收到的數量，就混進這次重傳送
			msgs = append(msgs, value.(*pb.Envelope))
		}
		return true
	})

	if len(msgs) == 0 {
		return
	}

	for _, msg := range msgs {
		msg.Header.Ack = c.lastAck // 告訴伺服器我現在收到的最新進度
		plaintext, _ := proto.Marshal(msg)

		ciphertext, nonce, err := crypto.EncryptAESGCM(plaintext, c.sharedSecret)
		if err == nil {
			wireQueue := &pb.Packet{
				Payload: &pb.Packet_Encrypted{
					Encrypted: &pb.TransferEncrypted{
						Codes: pb.SystemCodes_EncryptedData,
						Data:  ciphertext,
						Nonce: nonce,
						Tag:   nil,
					},
				},
			}
			b, _ := proto.Marshal(wireQueue)
			c.sendRaw(b)
		}
	}
}

func (c *NetworkClient) clearWaitingQueueUpTo(ack uint64) {
	c.waitingQueue.Range(func(key, value interface{}) bool {
		seq := key.(uint64)
		if seq <= ack {
			c.waitingQueue.Delete(key)
		}
		return true
	})
}

// heartbeatLoop
func (c *NetworkClient) heartbeatLoop() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.stopChan:
			return
		case <-ticker.C:
			if c.State == StateConnected {
				pingEnv := &pb.Envelope{
					Payload: &pb.Envelope_Ping{
						Ping: &pb.Heartbeat{
							Timestamp: time.Now().UnixMilli(),
						},
					},
				}
				c.SendEnvelope(pingEnv)
			}
		}
	}
}

// ProcessIncoming 是 Ebiten 的主 Update() 迴圈用來拉取資料的安全口
func (c *NetworkClient) ProcessIncoming() {
	if c.OnEnvelopeReceived == nil {
		return
	}

	c.incomingMut.Lock()
	queue := c.incomingQueue
	c.incomingQueue = nil
	c.incomingMut.Unlock()

	for _, env := range queue {
		c.OnEnvelopeReceived(env)
	}
}

