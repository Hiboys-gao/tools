package md5

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"time"
)

func MD5V(str []byte) string {
	h := md5.New()
	h.Write(str)
	return hex.EncodeToString(h.Sum(nil))
}

func GenerateMD5Idf(format string, args ...interface{}) string {
	h := md5.New()
	h.Write([]byte(fmt.Sprintf(format, args...)))
	return hex.EncodeToString(h.Sum(nil))
}

func MD5ByTimeNow() string {
	timeStr := time.Now().String()
	return MD5V([]byte(timeStr))
}
