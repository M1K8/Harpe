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

func (d *DB) CreateShort(stock, author string, channelType int, spt, ept, poi, stop float32, expiry int64, starting float32) (chan bool, bool, error) {

	chanMap.LoadOrStore(d.Guild, &sync.Map{})
	exists, exitChan := d.GetExitChanExists("sh_" + stock)

	if exists {
		return exitChan, exists, nil
	}

	contxt := context.Background()

	s := &Short{
		ShortAlertID:  d.Guild + "_" + stock,
		ShortGuildID:  d.Guild,
		ShortTicker:   stock,
		ShortSPt:      spt,
		ShortEPt:      ept,
		ShortExpiry:   expiry,
		ShortStarting: starting,
		ShortPoI:      poi,
		ShortStop:     stop,
		ChannelType:   channelType,
		ShortCallTime: time.Now(),
		ShortPOIHit:   false,
		ShortLowest:   starting,
		Caller:        author,
	}

	_, err := d.db.NewInsert().Model(s).On("CONFLICT (short_alert_id) DO UPDATE").Exec(contxt)

	if err != nil {
		log.Println(fmt.Sprintf("Unable to create short %v : %v", stock, err.Error()))
		return nil, false, err
	}

	return exitChan, exists, nil
}

func (d *DB) RemoveShort(stock string) error {

	contxt := context.Background()

	s := &Short{
		ShortGuildID:  d.Guild,
		ShortTicker:   stock,
		ShortStarting: 0,
		ShortCallTime: time.Time{},
	}
	_, err := d.db.NewDelete().Model(s).Where("short_alert_id = ?", stock+"_"+d.Guild).Exec(contxt)

	if err != nil {
		log.Println(fmt.Sprintf("Unable to delete short %v : %v", stock, err.Error()))
		return err
	}
	clearFromSyncMap(chanMap, d.Guild, "sh_"+stock)
	return nil
}

func (d *DB) GetShort(stock string) (*Short, error) {

	contxt := context.Background()
	stock = strings.ToUpper(stock)
	s := &Short{
		ShortGuildID:  d.Guild,
		ShortTicker:   stock,
		ShortStarting: 0,
		ShortCallTime: time.Time{},
	}
	err := d.db.NewSelect().Model(s).Where("short_alert_id = ?", stock+"_"+d.Guild).Scan(contxt)

	if err != nil {
		log.Println(fmt.Sprintf("Unable to get short %v : %v", stock, err.Error()))
		return nil, err
	}
	gMap, ok := chanMap.Load(d.Guild)
	if ok {
		gMapCast := gMap.(*sync.Map)
		_, ok := gMapCast.Load("sh_" + stock)
		if !ok {
			log.Println(fmt.Sprintf("Unable to get alert channel for short %v. Please try recreating this alert, or calling !refresh then running this command again.", stock))
			return nil, err
		}
	}
	return s, nil
}

func (d *DB) ShortPOIHit(ticker string) error {
	contxt := context.Background()

	s, err := d.GetShort(ticker)
	if err != nil {
		log.Println(fmt.Sprintf("Unable to get short %v : %v", ticker, err.Error()))
		return err
	}

	s.ShortPOIHit = true

	_, err = d.db.NewInsert().Model(s).On("CONFLICT (short_alert_id) DO UPDATE").Exec(contxt)

	if err != nil {
		log.Println(fmt.Sprintf("Unable to update short %v : %v", ticker, err.Error()))
		return err
	}

	return nil
}

func (d *DB) ShortSetNewHigh(uid string, price float32) error {
	contxt := context.Background()

	s, err := d.GetShort(uid)
	if err != nil {
		log.Println(fmt.Sprintf("Unable to get short %v : %v", uid, err.Error()))
		return err
	}

	s.ShortLowest = price

	_, err = d.db.NewInsert().Model(s).On("CONFLICT (short_alert_id) DO UPDATE").Exec(contxt)

	if err != nil {
		log.Println(fmt.Sprintf("Unable to update short %v : %v", uid, err.Error()))
		return err
	}

	return nil
}

func (s Short) GetPctGain(highest float32) float32 {
	return ((highest - s.ShortStarting) / s.ShortStarting) * 100
}
