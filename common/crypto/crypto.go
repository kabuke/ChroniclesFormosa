package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
)

// GenerateECDHKeys 產生 X25519 金鑰對
func GenerateECDHKeys() (*ecdh.PrivateKey, []byte, error) {
	priv, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	return priv, priv.PublicKey().Bytes(), nil
}

// DeriveSharedSecret 計算共享密鑰並進行 SHA256 雜湊
func DeriveSharedSecret(priv *ecdh.PrivateKey, peerPubKeyBytes []byte) ([]byte, error) {
	peerPubKey, err := ecdh.X25519().NewPublicKey(peerPubKeyBytes)
	if err != nil {
		return nil, err
	}
	secret, err := priv.ECDH(peerPubKey)
	if err != nil {
		return nil, err
	}
	// 使用 SHA256 雜湊共享密鑰，產出 32-byte 的 AES 金鑰
	hash := sha256.Sum256(secret)
	return hash[:], nil
}

// EncryptAESGCM 使用 AES-256-GCM 加密資料
func EncryptAESGCM(plainText, key []byte) (cipherText, nonce []byte, err error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}

	nonce = make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, err
	}

	cipherText = aesGCM.Seal(nil, nonce, plainText, nil)
	return cipherText, nonce, nil
}

// DecryptAESGCM 使用 AES-256-GCM 解密資料
func DecryptAESGCM(cipherText, nonce, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	plainText, err := aesGCM.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	return plainText, nil
}

// HashSHA256 將字串進行 SHA256 雜湊並回傳十六進位字串
func HashSHA256(input string) string {
	h := sha256.Sum256([]byte(input))
	return fmt.Sprintf("%x", h)
}
