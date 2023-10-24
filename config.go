package dgotorrent

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"os"

	"github.com/spf13/viper"
)

type DConfig struct{}

var (
	dconfig *DConfig
	config  *viper.Viper
)

const (
	_CONFIG_PATH = "./.dgo_torrent.toml"

	KEY_PEER_ID = "client.peer_id"
)

func init() {
	// check config file
	_, err := os.Stat(_CONFIG_PATH)
	if err != nil {
		if os.IsNotExist(err) {
			_, err = os.Create(_CONFIG_PATH)
			if err != nil {
				panic(fmt.Sprintf("[dgo_torrent] create config file failed with error: %v\n", err))
			}
		} else {
			panic(fmt.Sprintf("[dgo_torrent] check config file failed with error: %v\n", err))
		}
	}

	// init viper
	config = viper.New()
	config.SetConfigFile(_CONFIG_PATH)

	err = config.ReadInConfig()
	if err != nil {
		panic(fmt.Sprintf("[dgo_torrent] viper read config failed with error: %v", err))
	}
}

func generatePeerID() string {
	charSet := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	charSetLength := big.NewInt(int64(len(charSet)))

	randomString := make([]byte, 14)
	for i := 0; i < 14; i++ {
		index, err := rand.Int(rand.Reader, charSetLength)
		if err != nil {
			panic(err)
		}
		randomString[i] = charSet[index.Int64()]
	}

	return fmt.Sprintf("DT-%s-%s", "00", string(randomString))
}

func Config() *DConfig {
	return dconfig
}

func (cf *DConfig) GetConfig() *viper.Viper {
	return config
}

func (cf *DConfig) GetPeerID() string {
	if !config.IsSet(KEY_PEER_ID) {
		id := generatePeerID()

		config.Set(KEY_PEER_ID, id)
		config.WriteConfig()

		return id
	}

	id := config.GetString(KEY_PEER_ID)
	idBytes := []byte(id)
	if len(idBytes) != 20 {
		// malformed peer id, regenerate
		id = generatePeerID()

		config.Set(KEY_PEER_ID, id)
		config.WriteConfig()
	}

	return id
}
