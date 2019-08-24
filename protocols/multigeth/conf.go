/*
	Copyright 2019 whiteblock Inc.
	This file is a part of the genesis.

	Genesis is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	Genesis is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU General Public License
	along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package multigeth

import (
	"encoding/json"
	"fmt"

	"github.com/whiteblock/genesis/protocols/ethereum"
	"github.com/whiteblock/genesis/protocols/helpers"
	"github.com/whiteblock/genesis/testnet"
	"github.com/whiteblock/genesis/util"
)

type ethConf struct {
	ethereum.BaseConfig
	ChainID        int64 `json:"chainId"`
	EIP158Block    int64 `json:"eip158Block"`
	ByzantiumBlock int64 `json:"byzantiumBlock"`
	DisposalBlock  int64 `json:"disposalBlock"`
	//	ConstantinopleBlock int64  `json:"constantinopleBlock"`
	ECIP1017EraRounds  int64 `json:"ecip1017EraRounds"`
	ECIP1010PauseBlock int64 `json:"ecip1010PauseBlock"`
	//	TrustedCheckpoint   int64  `json:"trustedCheckpoint"`
	ExposedAccounts int64 `json:"exposedAccounts"`
	Verbosity       int   `json:"verbosity"`
}

/**
 * Fills in the defaults for missing parts,
 */
func newConf(tn *testnet.TestNet) (*ethConf, error) {
	data := tn.LDD.Params
	out := new(ethConf)
	err := helpers.HandleBlockchainConfig(blockchain, data, out)
	if err != nil || data == nil {
		return out, util.LogError(err)
	}

	initBalance, exists := data["initBalance"]
	if exists && initBalance != nil {
		switch initBalance.(type) {
		case json.Number:
			out.InitBalance = initBalance.(json.Number).String()
		case string:
			out.InitBalance = initBalance.(string)
		default:
			return nil, fmt.Errorf("incorrect type for initBalance given")
		}
	}
	if out.ExposedAccounts != -1 && out.ExposedAccounts > out.ExtraAccounts+int64(tn.LDD.Nodes) {
		out.ExtraAccounts = out.ExposedAccounts - int64(tn.LDD.Nodes)
	}

	return out, nil
}

func restoreConf(tn *testnet.TestNet) (*ethConf, error) {
	out, err := newConf(tn)
	if err != nil {
		return nil, util.LogError(err)
	}
	return out, ethereum.FetchConfigParameters(tn, out)
}
