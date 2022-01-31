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
package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"sync"

	"github.com/m1k8/harpe/pkg/config"
	"github.com/uniplaces/carbon"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

var once = sync.Once{}
var client *bun.DB
var chanMap sync.Map

func NewDB(guildID string) *DB {
	contxt := context.Background()

	once.Do(func() {

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

		if cfg.PostgresCfg.PW == "" {
			panic("pg pw not set in config!")
		}
		pw := cfg.PostgresCfg.PW

		dsn := "postgres://postgres:@postgres:5432/db?sslmode=disable"
		//dsn := "postgres://localhost:5432/db?sslmode=disable"
		sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn),
			pgdriver.WithPassword(pw)))

		db := bun.NewDB(sqldb, pgdialect.New())
		db.RegisterModel((*Stock)(nil))
		db.RegisterModel((*Short)(nil))
		db.RegisterModel((*Option)(nil))
		db.RegisterModel((*Crypto)(nil))

		_, err = db.NewCreateTable().Model((*Stock)(nil)).IfNotExists().Exec(contxt)
		if err != nil {
			panic("unable to create/get stocks table: " + err.Error())
		}

		_, err = db.NewCreateTable().Model((*Short)(nil)).IfNotExists().Exec(contxt)
		if err != nil {
			panic("unable to create/get shorts table: " + err.Error())
		}

		_, err = db.NewCreateTable().Model((*Crypto)(nil)).IfNotExists().Exec(contxt)
		if err != nil {
			panic("unable to create/get crypto table: " + err.Error())
		}

		_, err = db.NewCreateTable().Model((*Option)(nil)).IfNotExists().Exec(contxt)
		if err != nil {
			panic("unable to create/get options table: " + err.Error())
		}

		client = db
		chanMap = sync.Map{}
	})

	if client == nil {
		panic("db not set!")
	}

	return &DB{
		guild: guildID,
		db:    client,
	}
}

func (d *DB) RmAll(guildID string) error {
	contxt := context.Background()

	allStocks := make([]*Stock, 0)
	allShorts := make([]*Short, 0)
	allOptions := make([]*Option, 0)
	allCrypto := make([]*Crypto, 0)
	log.Println("Nuke called!!!!!!!!!!!!!!!!!!!!!!")

	err := d.db.NewSelect().Model((*Stock)(nil)).Where("stock_guild_id = ?", guildID).Scan(contxt, &allStocks)

	if err != nil {
		log.Println(fmt.Sprintf("Unable to get stocks. There is probably a serious issue: %v.", err.Error()))
		return err
	}

	for _, v := range allStocks {
		log.Println("removing " + v.StockTicker)
		d.RemoveStock(guildID, v.StockTicker)
	}

	err = d.db.NewSelect().Model((*Short)(nil)).Where("short_guild_id = ?", guildID).Scan(contxt, &allShorts)

	if err != nil {
		log.Println(fmt.Sprintf("Unable to get shorts. There is probably a serious issue: %v.", err.Error()))
		return err
	}

	for _, v := range allShorts {
		log.Println("removing " + v.ShortTicker)
		d.RemoveShort(guildID, v.ShortTicker)
	}

	err = d.db.NewSelect().Model((*Crypto)(nil)).Where("crypto_guild_id = ?", guildID).Scan(contxt, &allCrypto)

	if err != nil {
		log.Println(fmt.Sprintf("Unable to get crypto. There is probably a serious issue: %v.", err.Error()))
		return err
	}

	for _, v := range allCrypto {
		log.Println("removing " + v.CryptoCoin)
		d.RemoveCrypto(guildID, v.CryptoCoin)
	}

	err = d.db.NewSelect().Model((*Option)(nil)).Where("option_guild_id = ?", guildID).Scan(contxt, &allOptions)

	if err != nil {
		log.Println(fmt.Sprintf("Unable to get options. There is probably a serious issue: %v.", err.Error()))
		return err
	}

	for _, v := range allOptions {
		log.Println("removing " + v.OptionUid)
		d.RemoveOptionByCode(guildID, v.OptionUid)
	}

	log.Println("Nuke completed!!!!!!!!!!!!!!!!!!!!!!")

	return nil
}

func (d *DB) GetAll(guildID string) ([]*Stock, []*Short, []*Crypto, []*Option, error) {
	contxt := context.Background()
	allStocks := make([]*Stock, 0)
	allShorts := make([]*Short, 0)
	allOptions := make([]*Option, 0)
	allCrypto := make([]*Crypto, 0)

	err := d.db.NewSelect().Model((*Stock)(nil)).Where("stock_guild_id = ?", guildID).Scan(contxt, &allStocks)

	if err != nil {
		log.Println(fmt.Sprintf("Unable to get stocks. There is probably a serious issue: %v.", err.Error()))
		return nil, nil, nil, nil, err
	}

	err = d.db.NewSelect().Model((*Short)(nil)).Where("short_guild_id = ?", guildID).Scan(contxt, &allShorts)

	if err != nil {
		log.Println(fmt.Sprintf("Unable to get shorts. There is probably a serious issue: %v.", err.Error()))
		return nil, nil, nil, nil, err
	}

	err = d.db.NewSelect().Model((*Crypto)(nil)).Where("crypto_guild_id = ?", guildID).Scan(contxt, &allCrypto)

	if err != nil {
		log.Println(fmt.Sprintf("Unable to get crypto. There is probably a serious issue: %v.", err.Error()))
		return nil, nil, nil, nil, err
	}

	err = d.db.NewSelect().Model((*Option)(nil)).Where("option_guild_id = ?", guildID).Scan(contxt, &allOptions)

	if err != nil {
		log.Println(fmt.Sprintf("Unable to get options. There is probably a serious issue: %v.", err.Error()))
		return nil, nil, nil, nil, err
	}

	return allStocks, allShorts, allCrypto, allOptions, nil
	/*
		if len(allStocks) > 0 {
			stockStr = "\n**__Stocks__**\n"

			for _, v := range allStocks {
				OGPrice := v.StockStarting
				newPrice, err := stonks.GetStock(v.StockTicker)

				tradeType := ""

				switch v.ChannelType {
				case utils.DAY:
					tradeType = "Day Trade"
				case utils.SWING:
					tradeType = "Long Trade"
				case utils.WATCHLIST:
					tradeType = "Watchlist"
				}

				if err != nil {
					stockStr += fmt.Sprintf("Unable to fetch current price for **%v**. Starting price: *$%.2f* - %v", v.StockTicker, OGPrice, tradeType) + "\n"
				} else {
					stockStr += fmt.Sprintf("**%v** starting price: *$%.2f*, current price: *$%.2f* - %v", v.StockTicker, OGPrice, newPrice, tradeType)
					if v.StockEPt != 0 {
						stockStr += fmt.Sprintf(" Exit PT: $%.2f ", v.StockEPt)
					}
					if v.StockSPt != 0 {
						stockStr += fmt.Sprintf(" Scale PT: $%.2f ", v.StockSPt)
					}
					stockStr += "\n"
				}
			}
		}

		if len(allShorts) > 0 {
			shortStr += "\n**__Shorts__**\n"

			for _, v := range allShorts {
				OGPrice := v.ShortStarting

				newPrice, err := stonks.GetStock(v.ShortTicker)

				tradeType := ""

				switch v.ChannelType {
				case utils.DAY:
					tradeType = "Day Trade"
				case utils.SWING:
					tradeType = "Long Trade"
				case utils.WATCHLIST:
					tradeType = "Watchlist"
				}

				if err != nil {
					shortStr += fmt.Sprintf("Unable to fetch current price for **%v**. Starting price: *$%.2f* - %v", v.ShortTicker, OGPrice, tradeType) + "\n"
				} else {
					shortStr += fmt.Sprintf("**%v** starting price: *$%.2f*, current price: *$%.2f* - %v", v.ShortTicker, OGPrice, newPrice, tradeType)
					if v.ShortEPt != 0 {
						shortStr += fmt.Sprintf(" Exit PT: $%.2f ", v.ShortEPt)
					}
					if v.ShortSPt != 0 {
						shortStr += fmt.Sprintf(" Scale PT: $%.2f ", v.ShortSPt)
					}
					shortStr += "\n"
				}
			}
		}

		if len(allCrypto) > 0 {
			cryptoStr += "\n**__Crypto__**\n"

			for _, v := range allCrypto {
				OGPrice := v.CryptoStarting

				newPrice, err := stonks.GetCrypto(v.CryptoCoin, false)
				tradeType := ""

				switch v.ChannelType {
				case utils.DAY:
					tradeType = "Day Trade"
				case utils.SWING:
					tradeType = "Long Trade"
				case utils.WATCHLIST:
					tradeType = "Watchlist"
				}

				if err != nil {
					cryptoStr += fmt.Sprintf("Unable to fetch current price for **%v**. Starting price: *$%.8f* - %v", v.CryptoCoin, OGPrice, tradeType) + "\n"
				} else {
					cryptoStr += fmt.Sprintf("**%v** starting price: *$%.8f*, current price: *$%.8f* - %v", v.CryptoCoin, OGPrice, newPrice, tradeType)
					if v.CryptoEPt != 0 {
						cryptoStr += fmt.Sprintf(" Exit PT: $%.2f ", v.CryptoEPt)
					}
					if v.CryptoSPt != 0 {
						cryptoStr += fmt.Sprintf(" Scale PT: $%.2f ", v.CryptoSPt)
					}
					cryptoStr += "\n"
				}
			}
		}

		if len(allOptions) > 0 {
			optiStr += "\n**__Options__**\n"

			for _, v := range allOptions {
				OGPrice := v.OptionStarting
				ticker, cType, day, month, year, price, err := splitOptionsCode(v.OptionUid)
				prettyOID := utils.NiceStr(ticker, cType, day, month, year, price)
				if err != nil {
					log.Println(fmt.Sprintf("Unable to parse options code for %v. It is likely in a corrupt state and should be removed: %v.", prettyOID, err.Error()))
				}

				newPrice, _, err := stonks.GetOption(ticker, cType, day, month, year, price, 0)

				tradeType := ""

				switch v.ChannelType {
				case utils.DAY:
					tradeType = "Day Trade"
				case utils.SWING:
					tradeType = "Long Trade"
				case utils.WATCHLIST:
					tradeType = "Watchlist"
				}
				if err != nil {
					log.Println(fmt.Sprintf("Unable to get option %v. Crash?: %v.", prettyOID, err.Error()))
					optiStr += fmt.Sprintf("Unable to fetch current price for **%v**. Starting price: *$%.2f* - %v", prettyOID, OGPrice, tradeType) + "\n"
				} else {
					optiStr += fmt.Sprintf("**%v** starting price: *$%.2f*, current price: *$%.2f* - %v", prettyOID, OGPrice, newPrice, tradeType) + "\n"
				}
			}
		}

		fmt.Println("stockStr is " + stockStr)
		fmt.Println("shortStr is " + shortStr)
		fmt.Println("cryptoStr is " + cryptoStr)
		fmt.Println("optiStr is " + optiStr)

		if optiStr == "" && stockStr == "" && shortStr == "" && cryptoStr == "" {
			return []string{"**No alerts to show**"}
		} else {
			respStrs := make([]string, 0)
			if !IsTradingHours() {
				respStrs = append(respStrs, "**Warning** - currently outside of market hours - data may not be accurate\n")
			}
			if stockStr != "" {
				respStrs = append(respStrs, stockStr)
			}
			if shortStr != "" {
				respStrs = append(respStrs, shortStr)
			}
			if cryptoStr != "" {
				respStrs = append(respStrs, cryptoStr)
			}
			if optiStr != "" {
				respStrs = append(respStrs, optiStr)
			}
			return respStrs
		}*/
}

func (d *DB) GetExitChan(guildid, index string) chan bool {
	val, _ := chanMap.Load(guildid)

	gMap := val.(*sync.Map)
	gVal, ok := gMap.Load(index)

	if ok {
		return gVal.(chan bool)
	} else {
		return nil
	}
}

func (d *DB) getExitChanExists(guildid, index string, exitChan chan bool) (bool, chan bool) {
	val, _ := chanMap.Load(guildid)

	gMap := val.(*sync.Map)
	gVal, ok := gMap.LoadOrStore(index, exitChan)

	if ok {
		return true, gVal.(chan bool)
	} else {
		return false, exitChan
	}
}

func (d *DB) SetAndReturnNewExitChan(guildid, index string, exitChan chan bool) chan bool {
	val, _ := chanMap.LoadOrStore(guildid, &sync.Map{})

	gMap := val.(*sync.Map)
	gMap.Store(index, exitChan)

	return exitChan
}

func (d *DB) RefreshFromDB(guildID string) ([]*Stock, []*Short, []*Option, []*Crypto, error) {
	contxt := context.Background()
	allStocks := make([]*Stock, 0)
	allShorts := make([]*Short, 0)
	allOptions := make([]*Option, 0)
	allCrypto := make([]*Crypto, 0)

	err := d.db.NewSelect().Model(&allStocks).Where("stock_guild_id = ?", guildID).Scan(contxt)

	if err != nil {
		log.Println(fmt.Sprintf("Unable to get stocks. There is probably a serious issue: %v.", err.Error()))
		return nil, nil, nil, nil, err
	}

	err = d.db.NewSelect().Model(&allShorts).Where("short_guild_id = ?", guildID).Scan(contxt)

	if err != nil {
		log.Println(fmt.Sprintf("Unable to get shorts. There is probably a serious issue: %v.", err.Error()))
		return nil, nil, nil, nil, err
	}

	err = d.db.NewSelect().Model(&allCrypto).Where("crypto_guild_id = ?", guildID).Scan(contxt)

	if err != nil {
		log.Println(fmt.Sprintf("Unable to get crypto. There is probably a serious issue: %v.", err.Error()))
		return nil, nil, nil, nil, err
	}

	err = d.db.NewSelect().Model(&allOptions).Where("option_guild_id = ?", guildID).Scan(contxt)

	if err != nil {
		log.Println(fmt.Sprintf("Unable to get options. There is probably a serious issue: %v.", err.Error()))
		return nil, nil, nil, nil, err
	}

	chanMap.LoadOrStore(guildID, &sync.Map{})

	return allStocks, allShorts, allOptions, allCrypto, nil
}

func splitOptionsCode(code string) (string, string, string, string, string, float32, error) {
	var (
		ticker string
		day    string
		month  string
		year   string
		cType  string
		price  float32
	)

	indexOffset := 0

	switch len(code) {
	case 16:
		indexOffset = 0
	case 17:
		indexOffset = 1
	case 18:
		indexOffset = 2
	case 19:
		indexOffset = 3
	default:
		return "", "", "", "", "", -1, errors.New("invalid code - " + code)
	}

	ticker = code[:indexOffset+1] // 3
	year = "20" + code[indexOffset+1:indexOffset+3]
	month = code[indexOffset+3 : indexOffset+5]
	day = code[indexOffset+5 : indexOffset+7]
	cType = code[indexOffset+7 : indexOffset+8]

	p, err := strconv.ParseFloat(code[indexOffset+8:], 32)
	if err != nil {
		return "", "", "", "", "", -1, err
	}

	price = float32(p / 1000)

	return ticker, cType, day, month, year, price, nil
}

func IsTradingHours() bool {
	now, _ := carbon.NowInLocation("America/Detroit") //ET timezone
	if now.IsWeekend() {
		return false
	}

	nowHour := now.Hour()
	nowMinute := now.Minute()
	if (nowHour < 9) || (nowHour == 9 && nowMinute < 30) || (nowHour >= 16) { // NOTE - for 2 days a year the bot will sleep early due to DST
		return false
	}

	return true
}
