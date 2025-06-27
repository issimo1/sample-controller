package cfg

type MysqlConfig struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Database string `json:"database"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
}
