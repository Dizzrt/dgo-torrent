package config

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"sync"

	"github.com/Dizzrt/dgo-torrent/common"
	"github.com/Dizzrt/dgo-torrent/db"
	"github.com/Dizzrt/dgo-torrent/dlog"
	"github.com/spf13/viper"
)

type config struct {
	V *viper.Viper
}

var (
	_cfg *config
	once sync.Once
)

const (
	_CONFIG_PATH = "./.dgo_torrent.toml"

	KEY_PEER_ID = "client.peer_id"

	KEY_DEFAULT_DOWNLOAD_PATH = "client.settings.default_download_path"
)

func init() {
	// init logger
	dlog.Init()

	// init sqlite
	db.Init()
}

func Instance() *config {
	once.Do(func() {
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
		v := viper.New()
		v.SetConfigFile(_CONFIG_PATH)

		err = v.ReadInConfig()
		if err != nil {
			panic(fmt.Sprintf("[dgo_torrent] viper read config failed with error: %v", err))
		}

		c := &config{
			V: v,
		}

		_cfg = c
	})

	return _cfg
}

func (cfg *config) GetPeerID() string {
	if !cfg.V.IsSet(KEY_PEER_ID) {
		id := common.GeneratePeerID()

		cfg.V.Set(KEY_PEER_ID, id)
		cfg.V.WriteConfig()

		return id
	}

	id := cfg.V.GetString(KEY_PEER_ID)
	idBytes := []byte(id)
	if len(idBytes) != 20 {
		// malformed peer id, regenerate
		id = common.GeneratePeerID()

		cfg.V.Set(KEY_PEER_ID, id)
		cfg.V.WriteConfig()
	}

	return id
}

func (cfg *config) GetDefaultDonwloadPath() string {
	if !cfg.V.IsSet(KEY_DEFAULT_DOWNLOAD_PATH) {
		var path string
		user, err := user.Current()
		if err != nil {
			path, err = os.Getwd()
			if err != nil {
				dlog.Fatalf("Failed to get default download path with error: %v", err)
			}
		} else {
			path = filepath.Join(user.HomeDir, "Downloads")
		}

		cfg.V.Set(KEY_DEFAULT_DOWNLOAD_PATH, path)
		cfg.V.WriteConfig()

		return path
	}

	path := cfg.V.GetString(KEY_DEFAULT_DOWNLOAD_PATH)
	return path
}
