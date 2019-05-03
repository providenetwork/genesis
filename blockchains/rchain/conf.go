package rchain

import (
	"../../util"
	"../helpers"
	"encoding/base64"
	"encoding/json"
	"log"
)

type rChainConf struct {
	NoUpnp               bool   `json:"noUpnp"`
	DefaultTimeout       int64  `json:"defaultTimeout"`
	MapSize              int64  `json:"mapSize"`
	CasperBlockStoreSize int64  `json:"casperBlockStoreSize"`
	InMemoryStore        bool   `json:"inMemoryStore"`
	MaxNumOfConnections  int64  `json:"maxNumOfConnections"`
	Validators           int64  `json:"validators"`
	ValidatorCount       int64  `json:"validatorCount"`
	SigAlgorithm         string `json:"sigAlgorithm"`
	Command              string `json:"command"`
	BondsValue           int64  `json:"bondsValue"`
}

func newRChainConf(data map[string]interface{}) (*rChainConf, error) {
	out := new(rChainConf)
	err := json.Unmarshal([]byte(GetDefaults()), out)
	if data == nil {
		return out, util.LogError(err)
	}
	log.Printf("Default %+v\n", *out)
	tmp, err := json.Marshal(data)
	if err != nil {
		return nil, util.LogError(err)
	}
	return out, json.Unmarshal(tmp, out)
}

// GetServices returns the services which are used by rchain
func GetServices() []util.Service {
	return []util.Service{
		{
			Name:  "wb_influx_proxy",
			Image: "gcr.io/wb-genesis/bitbucket.org/whiteblockio/influx-proxy:master",
			Env: map[string]string{
				"BASIC_AUTH_BASE64": base64.StdEncoding.EncodeToString([]byte(conf.InfluxUser + ":" + conf.InfluxPassword)),
				"INFLUXDB_URL":      conf.Influx,
				"BIND_PORT":         "8086",
			},
		},
	}
}

// GetParams fetchs rchain related parameters
func GetParams() string {
	dat, err := helpers.GetStaticBlockchainConfig("rchain", "params.json")
	if err != nil {
		panic(err) //Missing required files is a fatal error
	}
	return string(dat)
}

// GetDefaults fetchs rchain related parameter defaults
func GetDefaults() string {
	dat, err := helpers.GetStaticBlockchainConfig("rchain", "defaults.json")
	if err != nil {
		panic(err) //Missing required files is a fatal error
	}
	return string(dat)
}
