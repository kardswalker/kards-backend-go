package security

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
)

var keyLenArray = []int{47, 53, 73, 55, 61, 103, 47, 103, 33, 45, 73, 37, 97, 71, 39, 71, 31, 61, 83, 101, 53, 97, 79, 75, 37, 31, 33, 69, 43, 63, 39, 43, 79, 55, 49, 73, 83, 67, 59, 69, 103, 39, 47, 37, 41, 71, 89, 55, 49, 45, 33, 45, 69, 49, 43, 53, 59, 31, 59, 101, 61, 41, 79, 75, 83, 89, 75, 67, 41, 89, 63, 101, 67, 63, 97}

func DecryptPacket(packet string) (int, map[string]interface{}, error) {
	idx, _ := strconv.Atoi(packet[0:2])
	length, _ := strconv.Atoi(packet[2:8])
	keyLen := keyLenArray[idx]
	
	keyStr := packet[12 : 12+keyLen]
	bodyB64 := packet[12+keyLen:]
	
	// 解码 Body
	decodedBody, _ := base64.StdEncoding.DecodeString(bodyB64)
	
	// XOR 解密
	plainBytes := make([]byte, len(decodedBody))
	for j := 0; j < len(decodedBody) && j < length; j++ {
		plainBytes[j] = decodedBody[j] ^ keyStr[j%len(keyStr)]
	}

	// 解码 ActionID (Header 8:12)
	headerBytes, _ := base64.StdEncoding.DecodeString(packet[8:12] + "==")
	aid := (int(headerBytes[0]^keyStr[0]) << 16) | (int(headerBytes[1]^keyStr[1%len(keyStr)]) << 8) | int(headerBytes[2]^keyStr[2%len(keyStr)])

	var content map[string]interface{}
	json.Unmarshal(plainBytes, &content)
	
	return aid, content, nil
}

func EncryptPacket(actionID int, data interface{}) string {
	idx := rand.Intn(len(keyLenArray))
	keyLen := keyLenArray[idx]
	
	// 生成随机 Key
	key := make([]byte, keyLen)
	for i := range key {
		key[i] = byte(rand.Intn(93) + 33) // 避免特殊字符
	}
	keyStr := string(key)

	plaintext, _ := json.Marshal(data)
	
	// Body XOR
	encryptedBody := make([]byte, len(plaintext))
	for j := range plaintext {
		encryptedBody[j] = plaintext[j] ^ keyStr[j%len(keyStr)]
	}
	bodyB64 := base64.StdEncoding.EncodeToString(encryptedBody)

	// Header XOR
	u := []byte{
		byte(actionID >> 16),
		byte(actionID >> 8),
		byte(actionID),
	}
	for j := 0; j < 3; j++ {
		u[j] ^= keyStr[j%len(keyStr)]
	}
	headerB64 := base64.StdEncoding.EncodeToString(u)[:4]

	return fmt.Sprintf("%02d%06d%s%s%s", idx, len(plaintext), headerB64, keyStr, strings.TrimRight(bodyB64, "="))
}