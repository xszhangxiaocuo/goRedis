package config

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"goRedis/lib/utils"

	"goRedis/lib/logger"
)

var (
	ClusterMode    = "cluster"
	StandaloneMode = "standalone"
)

// ServerProperties 定义服务器的全局配置属性
type ServerProperties struct {
	// 公共配置
	RunID             string `cfg:"runid"`                // 每次执行时都不同的运行ID。
	Bind              string `cfg:"bind"`                 // 服务器绑定的IP地址。
	Port              int    `cfg:"port"`                 // 服务器监听的端口号。
	Dir               string `cfg:"dir"`                  // 服务器的工作目录。
	AnnounceHost      string `cfg:"announce-host"`        // 用于集群模式下，节点间通信的主机地址。
	AppendOnly        bool   `cfg:"appendonly"`           // 是否开启追加模式。
	AppendFilename    string `cfg:"appendfilename"`       // 追加模式下的文件名。
	AppendFsync       string `cfg:"appendfsync"`          // 追加模式下的同步策略。
	AofUseRdbPreamble bool   `cfg:"aof-use-rdb-preamble"` // 是否在AOF文件开头使用RDB格式数据。
	MaxClients        int    `cfg:"maxclients"`           // 最大客户端连接数。
	RequirePass       string `cfg:"requirepass"`          // 访问密码。
	Databases         int    `cfg:"databases"`            // 数据库数量。
	RDBFilename       string `cfg:"dbfilename"`           // RDB文件名。
	MasterAuth        string `cfg:"masterauth"`           // 主节点认证密码。
	SlaveAnnouncePort int    `cfg:"slave-announce-port"`  // 从节点宣告端口。
	SlaveAnnounceIP   string `cfg:"slave-announce-ip"`    // 从节点宣告IP。
	ReplTimeout       int    `cfg:"repl-timeout"`         // 复制超时时间。
	ClusterEnable     bool   `cfg:"cluster-enable"`       // 是否启用集群模式。
	ClusterAsSeed     bool   `cfg:"cluster-as-seed"`      // 是否作为种子节点。
	ClusterSeed       string `cfg:"cluster-seed"`         // 集群种子节点。
	ClusterConfigFile string `cfg:"cluster-config-file"`  // 集群配置文件。

	// 集群模式配置
	ClusterEnabled string   `cfg:"cluster-enabled"` // 目前未使用。
	Peers          []string `cfg:"peers"`           // 集群中的其他节点。
	Self           string   `cfg:"self"`            // 本节点的地址。

	// 配置文件路径
	CfPath string `cfg:"cf,omitempty"`
}

type ServerInfo struct {
	StartUpTime time.Time // 服务器启动时间。
}

func (p *ServerProperties) AnnounceAddress() string {
	return p.AnnounceHost + ":" + strconv.Itoa(p.Port)
}

func (p *ServerProperties) GetConfig(cfgName string) string {
	cfgName = strings.ToLower(cfgName)
	switch cfgName {
	case "bind":
		return p.Bind
	case "port":
		return strconv.Itoa(p.Port)
	case "databases":
		return strconv.Itoa(p.Databases)
	default:
		return ""
	}
}

// Properties holds global config properties
var Properties *ServerProperties   // 全局配置属性
var EachTimeServerInfo *ServerInfo // 服务器信息

func init() {
	// 初始化服务器信息
	EachTimeServerInfo = &ServerInfo{
		StartUpTime: time.Now(),
	}

	// 默认配置
	Properties = &ServerProperties{
		Bind:       "127.0.0.1",
		Port:       6379,
		AppendOnly: false,
		RunID:      utils.RandString(40),
	}
}

func parse(src io.Reader) *ServerProperties {
	config := &ServerProperties{}

	// 读取解析配置文件
	rawMap := make(map[string]string)
	scanner := bufio.NewScanner(src)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) > 0 && strings.TrimLeft(line, " ")[0] == '#' {
			continue
		}
		pivot := strings.IndexAny(line, " ")
		if pivot > 0 && pivot < len(line)-1 {
			key := line[0:pivot]
			value := strings.Trim(line[pivot+1:], " ")
			rawMap[strings.ToLower(key)] = value
		}
	}
	if err := scanner.Err(); err != nil {
		logger.Fatal(err)
	}

	t := reflect.TypeOf(config)
	v := reflect.ValueOf(config)
	n := t.Elem().NumField()
	for i := 0; i < n; i++ {
		field := t.Elem().Field(i)
		fieldVal := v.Elem().Field(i)
		key, ok := field.Tag.Lookup("cfg")
		if !ok || strings.TrimLeft(key, " ") == "" {
			key = field.Name
		}
		value, ok := rawMap[strings.ToLower(key)]
		if ok {
			// 设置配置
			switch field.Type.Kind() {
			case reflect.String:
				fieldVal.SetString(value)
			case reflect.Int:
				intValue, err := strconv.ParseInt(value, 10, 64)
				if err == nil {
					fieldVal.SetInt(intValue)
				}
			case reflect.Bool:
				boolValue := "yes" == value
				fieldVal.SetBool(boolValue)
			case reflect.Slice:
				if field.Type.Elem().Kind() == reflect.String {
					slice := strings.Split(value, ",")
					fieldVal.Set(reflect.ValueOf(slice))
				}
			}
		}
	}
	return config
}

// SetupConfig 读取配置文件并将属性存储到Properties中
func SetupConfig(configFilename string) {
	file, err := os.Open(configFilename)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	Properties = parse(file)
	Properties.RunID = utils.RandString(40)
	configFilePath, err := filepath.Abs(configFilename)
	if err != nil {
		return
	}
	Properties.CfPath = configFilePath
	if Properties.Dir == "" {
		Properties.Dir = "."
	}
}

func GetTmpDir() string {
	return Properties.Dir + "/tmp"
}
