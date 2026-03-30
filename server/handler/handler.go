package handler

import (
    "log"
    pb "github.com/kabuke/ChroniclesFormosa/resource"
)

// HandleEnvelope 處理業務邏輯分發
func HandleEnvelope(env *pb.Envelope, send func(*pb.Envelope), broadcast func(*pb.Envelope)) {
    // 預留供 3.3 Handler 層實作
    log.Println("Received Envelope Action:", env.Payload)
}
