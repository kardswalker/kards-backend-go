package config

import "time"

var (
	Version      = "Kards 1.50"
	Host         = "127.0.0.1"
	Port         = "5231"
	WSPort       = "5232"
	DatabaseURL  = "root:1234567890@tcp(127.0.0.1:3306)/users?charset=utf8mb4&parseTime=True&loc=Local"
	JWTKey       = []byte("CometKards-is-a-help-much-kards-players-that-can't-find-gameuser-or-baned")
	JWTAlgorithm = "HS256"
	JWTExpiry    = time.Hour * 24
)
func GetKardsTime() string {
	return time.Now().UTC().Format("2006-01-02T15:04:05.000") + "Z"
}
