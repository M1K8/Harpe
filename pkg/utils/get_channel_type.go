/*
 * Copyright 2021 M1K
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package utils

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/m1k8/harpe/pkg/config"
)

var (
	servers = make(map[string]config.ServerConfig)
)

func init() {
	cfgFile, err := os.Open("config.json")

	if err != nil {
		panic("Unable to open config.json!")
	}

	defer cfgFile.Close()

	byteValue, err := ioutil.ReadAll(cfgFile)

	if err != nil {
		panic("Error reading config.json!")
	}
	var cfg config.Config
	json.Unmarshal(byteValue, &cfg)

	if len(cfg.ServersCfg) == 0 {
		panic("Servers not confugred!")
	} else {
		for _, v := range cfg.ServersCfg {
			servers[v.ID] = v
		}
	}
}

func GetChannelType(guildID, channelID string) int {

	if s, ok := servers[guildID]; ok {
		switch channelID {
		case s.ChannelConfig.Swing:
			return SWING
		case s.ChannelConfig.Day:
			return DAY
		case s.ChannelConfig.Watchlist:
			return WATCHLIST
		default:
			return -1
		}
	} else {
		return -1
	}
}
