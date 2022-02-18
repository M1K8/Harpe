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
package config

type Config struct {
	StocksCFG   StocksConfig   `json:"stocks"`
	DiscordCFG  DiscordConfig  `json:"discord"`
	TwitterCFG  TwitterConfig  `json:"twitter"`
	ServersCfg  []ServerConfig `json:"servers"`
	PostgresCfg PGConfig       `json:"pg"`
}

type DiscordConfig struct {
	API string `json:"DISCORD_API"`
}

type StocksConfig struct {
	Finn_API string `json:"FINNHUB_API"`
	E        string `json:"ENDPOINT"`
	Key      string `json:"KEY"`
}

type TwitterConfig struct {
	Allowed_guilds []AllowedGuild `json:"allowed_guilds"`
	Keys           TwitterKeys    `json:"keys"`
}

type AllowedGuild struct {
	GID  string `json:"guildID"`
	TUID string `json:"TwitterUID"`
}

type TwitterKeys struct {
	C_K string `json:"TWITTER_C_K"`
	C_S string `json:"TWITTER_C_S"`
	A_T string `json:"TWITTER_A_T"`
	A_S string `json:"TWITTER_A_S"`
}

type ServerConfig struct {
	ID               string        `json:"id"`
	ChannelConfig    ChannelConfig `json:"channels"`
	Roles            []string      `json:"allowed_roles"`
	Whitelisted_UIDs []string      `json:"whitelisted_ids"`
	AlertRole        string        `json:"alert_role"`
}
type ChannelConfig struct {
	Day        string `json:"day_trades"`
	Swing      string `json:"swings"`
	Watchlist  string `json:"watchlist"`
	DayAlerts  string `json:"day_alerts"`
	LongAlerts string `json:"long_alerts"`
	EOD        string `json:"eod"`
}

type PGConfig struct {
	PW string `json:"pw"`
}
