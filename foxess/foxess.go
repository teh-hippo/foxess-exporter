package foxess

import (
	"crypto/md5"
	"fmt"
)

const (
	BaseUrl = "https://www.foxesscloud.com"
)

func CalculateSignature(path string, apiKey string, timestamp int64) string {
	term := []byte(path + "\\r\\n" + apiKey + "\\r\\n" + fmt.Sprint(timestamp))
	return fmt.Sprintf("%x", md5.Sum(term))
}
