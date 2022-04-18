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
	"errors"
	"fmt"
	"log"
	"sync"
	"time"
)

func (d *DB) CreateStock(uid, stock, author string, channelType int, spt, ept, poi, stop, tstop float32, expiry int64, starting float32) (chan bool, bool, error) {

	chanMap.LoadOrStore(d.Guild, &sync.Map{})
	exists, exitChan := d.GetExitChanExists(uid)

	if exists {
		return exitChan, exists, nil
	}

	contxt := context.Background()
	s := &Stock{
		StockAlertID:      uid,
		StockGuildID:      d.Guild,
		StockTicker:       stock,
		StockEPt:          ept,
		StockSPt:          spt,
		StockExpiry:       expiry,
		StockStarting:     starting,
		StockStop:         stop,
		StockLastHigh:     starting,
		StockPoI:          poi,
		StockTrailingStop: tstop,
		ChannelType:       channelType,
		StockCallTime:     time.Now(),
		StockPOIHit:       false,
		StockHighest:      starting,
		Caller:            author,
	}

	_, err := d.db.NewInsert().Model(s).On("CONFLICT (stock_alert_id) DO UPDATE").Exec(contxt)

	if err != nil {
		log.Println(fmt.Sprintf("Unable to create Crypto %v : %v", uid, err.Error()))
		return nil, false, err
	}

	return exitChan, exists, nil
}

func (d *DB) RemoveStock(uid string) error {

	contxt := context.Background()

	s := &Stock{
		StockGuildID: d.Guild,
		StockAlertID: uid,
	}
	_, err := d.db.NewDelete().Model(s).Where("stock_alert_id = ?", uid).Exec(contxt)

	if err != nil {
		log.Println(fmt.Sprintf("Unable to delete Stock %v : %v", uid, err.Error()))
		return err
	}
	clearFromSyncMap(chanMap, d.Guild, uid)
	return nil
}

func (d *DB) GetStock(uid string) (*Stock, error) {

	contxt := context.Background()

	s := &Stock{
		StockGuildID: d.Guild,
		StockAlertID: uid,
	}
	err := d.db.NewSelect().Model(s).Where("stock_alert_id = ?", uid).Scan(contxt)

	if err != nil {
		log.Println(fmt.Sprintf("Unable to get Stock %v : %v", uid, err.Error()))
		return nil, err
	}
	gMap, ok := chanMap.Load(d.Guild)
	if ok {
		gMapCast := gMap.(*sync.Map)
		_, ok := gMapCast.Load(uid)
		if !ok {
			log.Println(fmt.Sprintf("Unable to get alert channel for stock %v. Please try recreating this alert, or calling !refresh then running this command again.", uid))
			return nil, err
		}
	}
	return s, nil
}

func (d *DB) StockPOIHit(uid string) error {
	contxt := context.Background()

	s, err := d.GetStock(uid)
	if err != nil {
		log.Println(fmt.Sprintf("Unable to get stock %v : %v", uid, err.Error()))
		return err
	}

	s.StockPOIHit = true

	_, err = d.db.NewInsert().Model(s).On("CONFLICT (stock_alert_id) DO UPDATE").Exec(contxt)

	if err != nil {
		log.Println(fmt.Sprintf("Unable to update stock %v : %v", uid, err.Error()))
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

	res, err := d.db.NewInsert().Model(s).On("CONFLICT (stock_alert_id) DO UPDATE").Exec(contxt)
	rowsAffected, _ := res.RowsAffected()

	if rowsAffected == 0 {
		err = errors.New(fmt.Sprintf("Unable to remove Stock %v : NOT FOUND", uid))
		return err
	}

	if err != nil {
		log.Println(fmt.Sprintf("Unable to remove Stock %v : %v", uid, err.Error()))
		return err
	}

	if err != nil {
		log.Println(fmt.Sprintf("Unable to update stock %v : %v", uid, err.Error()))
		return err
	}

	return nil
}

func (d *DB) StockSetNewAvg(uid string, price float32) error {
	contxt := context.Background()

	s, err := d.GetStock(uid)
	if err != nil {
		log.Println(fmt.Sprintf("Unable to get Stock %v : %v", uid, err.Error()))
		return err
	}

	s.StockStarting = price

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
