package network

import (
	"crypto/ecdh"
	"encoding/binary"
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

type NetworkClient struct {
	State      ClientState
	serverAddr string
	conn       *kcp.UDPSession

	privateKey   *ecdh.PrivateKey
	publicKey    []byte
	sharedSecret []byte
	SessionID    string

	RTT          int64
	clientSeq    uint64
	lastAck      uint64
	maxServerAck uint64

	outbox       []*pb.Envelope
	waitingQueue sync.Map 
	sendMutex    sync.Mutex

	incomingQueue []*pb.Envelope
	incomingMut   sync.Mutex

	OnEnvelopeReceived func(env *pb.Envelope)
	stopChan           chan struct{}
}

func NewNetworkClient(addr string) *NetworkClient {
	return &NetworkClient{
		State:      StateDisconnected,
		serverAddr: addr,
		stopChan:   make(chan struct{}),
	}
}

func (c *NetworkClient) Connect() {
	go c.reconnectLoop()
}

func (c *NetworkClient) reconnectLoop() {
	for {
		c.State = StateConnecting
		log.Printf("[Network] 🌐 Connecting to %s...", c.serverAddr)

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

		if c.SessionID != "" {
			c.State = StateResuming
			if err := c.resume(); err != nil {
				c.SessionID = ""
				if err := c.handshake(); err != nil {
					conn.Close()
					continue
				}
			}
		} else {
			if err := c.handshake(); err != nil {
				conn.Close()
				continue
			}
		}

		c.State = StateConnected
		log.Printf("[Network] 🟢 Connected! SessionID = %s", c.SessionID)

		go c.readLoop()
		go c.flushLoop()
		go c.heartbeatLoop()

		<-c.stopChan
		log.Println("[Network] 🛑 Connection Lost.")
		time.Sleep(2 * time.Second)
	}
}

func (c *NetworkClient) readRaw() ([]byte, error) {
	lenBuf := make([]byte, 4)
	if _, err := c.conn.Read(lenBuf); err != nil { return nil, err }
	length := binary.BigEndian.Uint32(lenBuf)
	data := make([]byte, length)
	_, err := c.conn.Read(data)
	return data, err
}

func (c *NetworkClient) sendRaw(data []byte) error {
	lenBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lenBuf, uint32(len(data)))
	if _, err := c.conn.Write(lenBuf); err != nil { return err }
	_, err := c.conn.Write(data)
	return err
}

func (c *NetworkClient) readLoop() {
	defer func() { c.stopChan <- struct{}{} }()
	for {
		data, err := c.readRaw()
		if err != nil { return }

		var packet pb.Packet
		if err := proto.Unmarshal(data, &packet); err != nil { continue }

		encMsg, ok := packet.Payload.(*pb.Packet_Encrypted)
		if !ok { continue }

		plaintext, err := crypto.DecryptAESGCM(encMsg.Encrypted.Data, encMsg.Encrypted.Nonce, c.sharedSecret)
		if err != nil { continue }

		env := &pb.Envelope{}
		if err := proto.Unmarshal(plaintext, env); err != nil { continue }

		if env.Header != nil {
			if env.Header.Seq > c.lastAck { c.lastAck = env.Header.Seq }
			if env.Header.Ack > c.maxServerAck {
				c.maxServerAck = env.Header.Ack
				c.clearWaitingQueueUpTo(c.maxServerAck)
			}
		}

		if pong, isPong := env.Payload.(*pb.Envelope_Pong); isPong {
			c.RTT = time.Now().UnixMilli() - pong.Pong.Timestamp
			continue
		}

		c.incomingMut.Lock()
		// 關鍵修復：使用 Clone 確保每個入隊封包都是獨立實體，根治重複與雙顯 Bug
		c.incomingQueue = append(c.incomingQueue, proto.Clone(env).(*pb.Envelope))
		c.incomingMut.Unlock()
	}
}

func (c *NetworkClient) flushLoop() {
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-c.stopChan: return
		case <-ticker.C: c.tryFlush()
		}
	}
}

func (c *NetworkClient) tryFlush() {
	if c.conn == nil || c.State != StateConnected { return }
	c.sendMutex.Lock()
	msgs := c.outbox
	c.outbox = nil
	c.sendMutex.Unlock()

	for _, msg := range msgs {
		msg.Header.Ack = c.lastAck
		plaintext, _ := proto.Marshal(msg)
		ciphertext, nonce, err := crypto.EncryptAESGCM(plaintext, c.sharedSecret)
		if err == nil {
			wire := &pb.Packet{
				Payload: &pb.Packet_Encrypted{
					Encrypted: &pb.TransferEncrypted{
						Codes: pb.SystemCodes_EncryptedData,
						Data:  ciphertext,
						Nonce: nonce,
					},
				},
			}
			b, _ := proto.Marshal(wire)
			c.sendRaw(b)
		}
	}
}

func (c *NetworkClient) clearWaitingQueueUpTo(ack uint64) {
	c.waitingQueue.Range(func(key, value interface{}) bool {
		if key.(uint64) <= ack { c.waitingQueue.Delete(key) }
		return true
	})
}

func (c *NetworkClient) handshake() error {
	priv, pub, _ := crypto.GenerateECDHKeys()
	c.privateKey, c.publicKey = priv, pub
	req := &pb.Packet{Payload: &pb.Packet_KeyExchangeReq{KeyExchangeReq: &pb.KeyExchangeRequest{PublicKey: pub}}}
	b, _ := proto.Marshal(req)
	c.sendRaw(b)

	respData, err := c.readRaw()
	if err != nil { return err }
	var packet pb.Packet
	proto.Unmarshal(respData, &packet)
	resp := packet.GetKeyExchangeResp()
	shared, _ := crypto.DeriveSharedSecret(c.privateKey, resp.PublicKey)
	c.sharedSecret, c.SessionID = shared, resp.SessionId
	c.clientSeq = resp.MaxClientSeq + 1
	return nil
}

func (c *NetworkClient) resume() error {
	req := &pb.Packet{Payload: &pb.Packet_ResumeReq{ResumeReq: &pb.ResumeSessionRequest{SessionId: c.SessionID, LastAck: c.lastAck, ClientNextSeq: c.clientSeq}}}
	b, _ := proto.Marshal(req)
	c.sendRaw(b)
	respData, err := c.readRaw()
	if err != nil { return err }
	var packet pb.Packet
	proto.Unmarshal(respData, &packet)
	c.maxServerAck = packet.GetResumeResp().MaxClientSeq
	return nil
}

func (c *NetworkClient) heartbeatLoop() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-c.stopChan: return
		case <-ticker.C:
			if c.State == StateConnected {
				c.SendEnvelope(&pb.Envelope{Payload: &pb.Envelope_Ping{Ping: &pb.Heartbeat{Timestamp: time.Now().UnixMilli()}}})
			}
		}
	}
}

func (c *NetworkClient) SendEnvelope(env *pb.Envelope) {
	if c.State != StateConnected { return }
	c.sendMutex.Lock()
	defer c.sendMutex.Unlock()
	c.clientSeq++
	env.Header = &pb.Header{Seq: c.clientSeq, SessionId: c.SessionID}
	c.outbox = append(c.outbox, env)
	c.waitingQueue.Store(c.clientSeq, env)
}

func (c *NetworkClient) ProcessIncoming() {
	if c.OnEnvelopeReceived == nil { return }
	c.incomingMut.Lock()
	queue := c.incomingQueue
	c.incomingQueue = nil
	c.incomingMut.Unlock()
	for _, env := range queue { c.OnEnvelopeReceived(env) }
}
