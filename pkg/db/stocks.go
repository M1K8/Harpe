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
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
)

func (d *DB) CreateStock(stock, author string, channelType int, spt, ept, poi, stop float32, expiry int64, starting float32) (chan bool, bool, error) {
	contxt := context.Background()
	s := &Stock{
		StockAlertID:  stock + "_" + d.Guild,
		StockGuildID:  d.Guild,
		StockTicker:   stock,
		StockEPt:      ept,
		StockSPt:      spt,
		StockExpiry:   expiry,
		StockStarting: starting,
		StockStop:     stop,
		StockPoI:      poi,
		ChannelType:   channelType,
		StockCallTime: time.Now(),
		StockPOIHit:   false,
		StockHighest:  starting,
		Caller:        author,
	}

	_, err := d.db.NewInsert().Model(s).On("CONFLICT (stock_alert_id) DO UPDATE").Exec(contxt)

	if err != nil {
		log.Println(fmt.Sprintf("Unable to create Crypto %v : %v", stock, err.Error()))
		return nil, false, err
	}

	exitChan := make(chan bool, 1)
	chanMap.LoadOrStore(d.Guild, &sync.Map{})
	exists, exitChan := d.getExitChanExists("s_"+stock, exitChan)

	return exitChan, exists, nil
}

func (d *DB) RemoveStock(stock string) error {

	contxt := context.Background()

	s := &Stock{
		StockGuildID:  d.Guild,
		StockTicker:   stock,
		StockStarting: 0,
		StockCallTime: time.Time{},
	}
	_, err := d.db.NewDelete().Model(s).Where("stock_alert_id = ?", stock+"_"+d.Guild).Exec(contxt)

	if err != nil {
		log.Println(fmt.Sprintf("Unable to delete Stock %v : %v", stock, err.Error()))
		return err
	}
	clearFromSyncMap(chanMap, d.Guild, "s_"+stock)
	return nil
}

func (d *DB) GetStock(stock string) (*Stock, error) {

	contxt := context.Background()
	stock = strings.ToUpper(stock)

	s := &Stock{
		StockGuildID:  d.Guild,
		StockTicker:   stock,
		StockStarting: 0,
		StockCallTime: time.Time{},
	}
	err := d.db.NewSelect().Model(s).Where("stock_alert_id = ?", stock+"_"+d.Guild).Scan(contxt)

	if err != nil {
		log.Println(fmt.Sprintf("Unable to get Stock %v : %v", stock, err.Error()))
		return nil, err
	}
	gMap, ok := chanMap.Load(d.Guild)
	if ok {
		gMapCast := gMap.(*sync.Map)
		_, ok := gMapCast.Load("s_" + stock)
		if !ok {
			log.Println(fmt.Sprintf("Unable to get alert channel for stock %v. Please try recreating this alert, or calling !refresh then running this command again.", stock))
			return nil, err
		}
	}
	return s, nil
}

func (d *DB) StockPOIHit(ticker string) error {
	contxt := context.Background()

	s, err := d.GetStock(ticker)
	if err != nil {
		log.Println(fmt.Sprintf("Unable to get stock %v : %v", ticker, err.Error()))
		return err
	}

	s.StockPOIHit = true

	_, err = d.db.NewInsert().Model(s).On("CONFLICT (stock_alert_id) DO UPDATE").Exec(contxt)

	if err != nil {
		log.Println(fmt.Sprintf("Unable to update stock %v : %v", ticker, err.Error()))
		return err
	}

	return nil
}

func (d *DB) StockSetNewHigh(uid string, price float32) error {
	contxt := context.Background()

	s, err := d.GetStock(uid)
	if err != nil {
		log.Println(fmt.Sprintf("Unable to get Stock %v : %v", uid, err.Error()))
		return err
	}

	s.StockHighest = price

	_, err = d.db.NewInsert().Model(s).On("CONFLICT (stock_alert_id) DO UPDATE").Exec(contxt)

	if err != nil {
		log.Println(fmt.Sprintf("Unable to update stock %v : %v", uid, err.Error()))
		return err
	}

	return nil
}

func (s Stock) GetPctGain(highest float32) float32 {
	return ((highest - s.StockStarting) / s.StockStarting) * 100
}
