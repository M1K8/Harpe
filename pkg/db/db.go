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
var chanMap = sync.Map{}

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
	})

	if client == nil {
		panic("db not set!")
	}

	return &DB{
		Guild: guildID,
		db:    client,
	}
}

func (d *DB) RmAll() error {
	contxt := context.Background()

	allStocks := make([]*Stock, 0)
	allShorts := make([]*Short, 0)
	allOptions := make([]*Option, 0)
	allCrypto := make([]*Crypto, 0)
	log.Println("Nuke called!!!!!!!!!!!!!!!!!!!!!!")

	err := d.db.NewSelect().Model((*Stock)(nil)).Where("stock_guild_id = ?", d.Guild).Scan(contxt, &allStocks)

	if err != nil {
		log.Println(fmt.Sprintf("Unable to get stocks. There is probably a serious issue: %v.", err.Error()))
		return err
	}

	for _, v := range allStocks {
		log.Println("removing " + v.StockTicker)
		d.RemoveStock(v.StockTicker)
	}

	err = d.db.NewSelect().Model((*Short)(nil)).Where("short_guild_id = ?", d.Guild).Scan(contxt, &allShorts)

	if err != nil {
		log.Println(fmt.Sprintf("Unable to get shorts. There is probably a serious issue: %v.", err.Error()))
		return err
	}

	for _, v := range allShorts {
		log.Println("removing " + v.ShortTicker)
		d.RemoveShort(v.ShortTicker)
	}

	err = d.db.NewSelect().Model((*Crypto)(nil)).Where("crypto_guild_id = ?", d.Guild).Scan(contxt, &allCrypto)

	if err != nil {
		log.Println(fmt.Sprintf("Unable to get crypto. There is probably a serious issue: %v.", err.Error()))
		return err
	}

	for _, v := range allCrypto {
		log.Println("removing " + v.CryptoCoin)
		d.RemoveCrypto(v.CryptoCoin)
	}

	err = d.db.NewSelect().Model((*Option)(nil)).Where("option_guild_id = ?", d.Guild).Scan(contxt, &allOptions)

	if err != nil {
		log.Println(fmt.Sprintf("Unable to get options. There is probably a serious issue: %v.", err.Error()))
		return err
	}

	for _, v := range allOptions {
		log.Println("removing " + v.OptionUid)
		d.RemoveOptionByCode(v.OptionUid)
	}

	log.Println("Nuke completed!!!!!!!!!!!!!!!!!!!!!!")

	return nil
}

func (d *DB) RmAllDay() error {
	contxt := context.Background()

	allStocks := make([]*Stock, 0)
	allShorts := make([]*Short, 0)
	allOptions := make([]*Option, 0)
	allCrypto := make([]*Crypto, 0)
	log.Println("Nuke called!!!!!!!!!!!!!!!!!!!!!!")

	err := d.db.NewSelect().Model((*Stock)(nil)).Where("stock_guild_id = ?", d.Guild).Where("channel_type = ?", 1).Scan(contxt, &allStocks)

	if err != nil {
		log.Println(fmt.Sprintf("Unable to get stocks. There is probably a serious issue: %v.", err.Error()))
		return err
	}

	for _, v := range allStocks {
		log.Println("removing " + v.StockTicker)
		d.RemoveStock(v.StockTicker)
	}

	err = d.db.NewSelect().Model((*Short)(nil)).Where("short_guild_id = ?", d.Guild).Where("channel_type = ?", 1).Scan(contxt, &allShorts)

	if err != nil {
		log.Println(fmt.Sprintf("Unable to get shorts. There is probably a serious issue: %v.", err.Error()))
		return err
	}

	for _, v := range allShorts {
		log.Println("removing " + v.ShortTicker)
		d.RemoveShort(v.ShortTicker)
	}

	err = d.db.NewSelect().Model((*Crypto)(nil)).Where("crypto_guild_id = ?", d.Guild).Where("channel_type = ?", 1).Scan(contxt, &allCrypto)

	if err != nil {
		log.Println(fmt.Sprintf("Unable to get crypto. There is probably a serious issue: %v.", err.Error()))
		return err
	}

	for _, v := range allCrypto {
		log.Println("removing " + v.CryptoCoin)
		d.RemoveCrypto(v.CryptoCoin)
	}

	err = d.db.NewSelect().Model((*Option)(nil)).Where("option_guild_id = ?", d.Guild).Where("channel_type = ?", 1).Scan(contxt, &allOptions)

	if err != nil {
		log.Println(fmt.Sprintf("Unable to get options. There is probably a serious issue: %v.", err.Error()))
		return err
	}

	for _, v := range allOptions {
		log.Println("removing " + v.OptionUid)
		d.RemoveOptionByCode(v.OptionUid)
	}

	log.Println("Nuke completed!!!!!!!!!!!!!!!!!!!!!!")

	return nil
}

func (d *DB) RmAllLong() error {
	contxt := context.Background()

	allStocks := make([]*Stock, 0)
	allShorts := make([]*Short, 0)
	allOptions := make([]*Option, 0)
	allCrypto := make([]*Crypto, 0)
	log.Println("Nuke called!!!!!!!!!!!!!!!!!!!!!!")

	err := d.db.NewSelect().Model((*Stock)(nil)).Where("stock_guild_id = ?", d.Guild).Where("channel_type = ?", 0).Scan(contxt, &allStocks)

	if err != nil {
		log.Println(fmt.Sprintf("Unable to get stocks. There is probably a serious issue: %v.", err.Error()))
		return err
	}

	for _, v := range allStocks {
		log.Println("removing " + v.StockTicker)
		d.RemoveStock(v.StockTicker)
	}

	err = d.db.NewSelect().Model((*Short)(nil)).Where("short_guild_id = ?", d.Guild).Where("channel_type = ?", 0).Scan(contxt, &allShorts)

	if err != nil {
		log.Println(fmt.Sprintf("Unable to get shorts. There is probably a serious issue: %v.", err.Error()))
		return err
	}

	for _, v := range allShorts {
		log.Println("removing " + v.ShortTicker)
		d.RemoveShort(v.ShortTicker)
	}

	err = d.db.NewSelect().Model((*Crypto)(nil)).Where("crypto_guild_id = ?", d.Guild).Where("channel_type = ?", 0).Scan(contxt, &allCrypto)

	if err != nil {
		log.Println(fmt.Sprintf("Unable to get crypto. There is probably a serious issue: %v.", err.Error()))
		return err
	}

	for _, v := range allCrypto {
		log.Println("removing " + v.CryptoCoin)
		d.RemoveCrypto(v.CryptoCoin)
	}

	err = d.db.NewSelect().Model((*Option)(nil)).Where("option_guild_id = ?", d.Guild).Where("channel_type = ?", 0).Scan(contxt, &allOptions)

	if err != nil {
		log.Println(fmt.Sprintf("Unable to get options. There is probably a serious issue: %v.", err.Error()))
		return err
	}

	for _, v := range allOptions {
		log.Println("removing " + v.OptionUid)
		d.RemoveOptionByCode(v.OptionUid)
	}

	log.Println("Nuke completed!!!!!!!!!!!!!!!!!!!!!!")

	return nil
}

func (d *DB) RmOpts() error {
	contxt := context.Background()

	allOptions := make([]*Option, 0)
	log.Println("Nuke called!!!!!!!!!!!!!!!!!!!!!!")

	err := d.db.NewSelect().Model((*Option)(nil)).Where("option_guild_id = ?", d.Guild).Scan(contxt, &allOptions)

	if err != nil {
		log.Println(fmt.Sprintf("Unable to get options. There is probably a serious issue: %v.", err.Error()))
		return err
	}

	for _, v := range allOptions {
		log.Println("removing " + v.OptionUid)
		d.RemoveOptionByCode(v.OptionUid)
	}

	log.Println("Nuke completed!!!!!!!!!!!!!!!!!!!!!!")

	return nil
}
func (d *DB) RmStocks() error {
	contxt := context.Background()

	allStocks := make([]*Stock, 0)

	log.Println("Nuke called!!!!!!!!!!!!!!!!!!!!!!")

	err := d.db.NewSelect().Model((*Stock)(nil)).Where("stock_guild_id = ?", d.Guild).Scan(contxt, &allStocks)

	if err != nil {
		log.Println(fmt.Sprintf("Unable to get stocks. There is probably a serious issue: %v.", err.Error()))
		return err
	}

	for _, v := range allStocks {
		log.Println("removing " + v.StockTicker)
		d.RemoveStock(v.StockTicker)
	}

	log.Println("Nuke completed!!!!!!!!!!!!!!!!!!!!!!")

	return nil
}

func (d *DB) GetAll() ([]*Stock, []*Short, []*Crypto, []*Option, error) {
	contxt := context.Background()
	allStocks := make([]*Stock, 0)
	allShorts := make([]*Short, 0)
	allOptions := make([]*Option, 0)
	allCrypto := make([]*Crypto, 0)

	err := d.db.NewSelect().Model((*Stock)(nil)).Where("stock_guild_id = ?", d.Guild).Scan(contxt, &allStocks)

	if err != nil {
		log.Println(fmt.Sprintf("Unable to get stocks. There is probably a serious issue: %v.", err.Error()))
		return nil, nil, nil, nil, err
	}

	err = d.db.NewSelect().Model((*Short)(nil)).Where("short_guild_id = ?", d.Guild).Scan(contxt, &allShorts)

	if err != nil {
		log.Println(fmt.Sprintf("Unable to get shorts. There is probably a serious issue: %v.", err.Error()))
		return nil, nil, nil, nil, err
	}

	err = d.db.NewSelect().Model((*Crypto)(nil)).Where("crypto_guild_id = ?", d.Guild).Scan(contxt, &allCrypto)

	if err != nil {
		log.Println(fmt.Sprintf("Unable to get crypto. There is probably a serious issue: %v.", err.Error()))
		return nil, nil, nil, nil, err
	}

	err = d.db.NewSelect().Model((*Option)(nil)).Where("option_guild_id = ?", d.Guild).Scan(contxt, &allOptions)

	if err != nil {
		log.Println(fmt.Sprintf("Unable to get options. There is probably a serious issue: %v.", err.Error()))
		return nil, nil, nil, nil, err
	}

	return allStocks, allShorts, allCrypto, allOptions, nil

}

func (d *DB) GetExitChan(index string) chan bool {
	val, _ := chanMap.Load(d.Guild)

	gMap := val.(*sync.Map)
	gVal, ok := gMap.Load(index)

	if ok {
		return gVal.(chan bool)
	} else {
		return nil
	}
}

func (d *DB) GetExitChanExists(index string) (bool, chan bool) {
	val, _ := chanMap.Load(d.Guild)
	exitChan := make(chan bool, 1)

	gMap := val.(*sync.Map)
	gVal, ok := gMap.LoadOrStore(index, exitChan)

	if ok {
		log.Println("Alert " + index + " exists!")
		return true, gVal.(chan bool)
	} else {
		return false, exitChan
	}
}

func (d *DB) SetAndReturnNewExitChan(index string, exitChan chan bool) chan bool {
	val, _ := chanMap.LoadOrStore(d.Guild, &sync.Map{})

	gMap := val.(*sync.Map)
	gMap.Store(index, exitChan)

	return exitChan
}

func (d *DB) RefreshFromDB() ([]*Stock, []*Short, []*Crypto, []*Option, error) {
	contxt := context.Background()
	allStocks := make([]*Stock, 0)
	allShorts := make([]*Short, 0)
	allOptions := make([]*Option, 0)
	allCrypto := make([]*Crypto, 0)

	err := d.db.NewSelect().Model(&allStocks).Where("stock_guild_id = ?", d.Guild).Scan(contxt)

	if err != nil {
		log.Println(fmt.Sprintf("Unable to get stocks. There is probably a serious issue: %v.", err.Error()))
		return nil, nil, nil, nil, err
	}

	err = d.db.NewSelect().Model(&allShorts).Where("short_guild_id = ?", d.Guild).Scan(contxt)

	if err != nil {
		log.Println(fmt.Sprintf("Unable to get shorts. There is probably a serious issue: %v.", err.Error()))
		return nil, nil, nil, nil, err
	}

	err = d.db.NewSelect().Model(&allCrypto).Where("crypto_guild_id = ?", d.Guild).Scan(contxt)

	if err != nil {
		log.Println(fmt.Sprintf("Unable to get crypto. There is probably a serious issue: %v.", err.Error()))
		return nil, nil, nil, nil, err
	}

	err = d.db.NewSelect().Model(&allOptions).Where("option_guild_id = ?", d.Guild).Scan(contxt)

	if err != nil {
		log.Println(fmt.Sprintf("Unable to get options. There is probably a serious issue: %v.", err.Error()))
		return nil, nil, nil, nil, err
	}

	chanMap.LoadOrStore(d.Guild, &sync.Map{})

	return allStocks, allShorts, allCrypto, allOptions, nil
}

func SplitOptionsCode(code string) (string, string, string, string, string, float32, error) {
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
