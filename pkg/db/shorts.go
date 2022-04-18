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

func (d *DB) CreateShort(uid, stock, author string, channelType int, spt, ept, poi, stop, tstop float32, expiry int64, starting float32) (chan bool, bool, error) {

	chanMap.LoadOrStore(d.Guild, &sync.Map{})
	exists, exitChan := d.GetExitChanExists(uid)

	if exists {
		return exitChan, exists, nil
	}

	contxt := context.Background()

	s := &Short{
		ShortAlertID:      uid,
		ShortGuildID:      d.Guild,
		ShortTicker:       stock,
		ShortSPt:          spt,
		ShortEPt:          ept,
		ShortExpiry:       expiry,
		ShortStarting:     starting,
		ShortPoI:          poi,
		ShortStop:         stop,
		ShortTrailingStop: tstop,
		ShortLastLow:      starting,
		ChannelType:       channelType,
		ShortCallTime:     time.Now(),
		ShortPOIHit:       false,
		ShortLowest:       starting,
		Caller:            author,
	}

	_, err := d.db.NewInsert().Model(s).On("CONFLICT (short_alert_id) DO UPDATE").Exec(contxt)

	if err != nil {
		log.Println(fmt.Sprintf("Unable to create short %v : %v", stock, err.Error()))
		return nil, false, err
	}

	return exitChan, exists, nil
}

func (d *DB) RemoveShort(uid string) error {

	contxt := context.Background()

	s := &Short{
		ShortGuildID: d.Guild,
		ShortAlertID: uid,
	}
	res, err := d.db.NewDelete().Model(s).Where("short_alert_id = ?", uid).Exec(contxt)
	rowsAffected, _ := res.RowsAffected()

	if rowsAffected == 0 {
		err = errors.New(fmt.Sprintf("Unable to remove Short %v : NOT FOUND", uid))
		return err
	}

	if err != nil {
		log.Println(fmt.Sprintf("Unable to remove Short %v : %v", uid, err.Error()))
		return err
	}

	if err != nil {
		log.Println(fmt.Sprintf("Unable to delete short %v : %v", uid, err.Error()))
		return err
	}
	clearFromSyncMap(chanMap, d.Guild, uid)
	return nil
}

func (d *DB) GetShort(uid string) (*Short, error) {

	contxt := context.Background()
	s := &Short{
		ShortGuildID: d.Guild,
		ShortAlertID: uid,
	}
	err := d.db.NewSelect().Model(s).Where("short_alert_id = ?", uid).Scan(contxt)

	if err != nil {
		log.Println(fmt.Sprintf("Unable to get short %v : %v", uid, err.Error()))
		return nil, err
	}
	gMap, ok := chanMap.Load(d.Guild)
	if ok {
		gMapCast := gMap.(*sync.Map)
		_, ok := gMapCast.Load(uid)
		if !ok {
			log.Println(fmt.Sprintf("Unable to get alert channel for short %v. Please try recreating this alert, or calling !refresh then running this command again.", uid))
			return nil, err
		}
	}
	return s, nil
}

func (d *DB) ShortPOIHit(uid string) error {
	contxt := context.Background()

	s, err := d.GetShort(uid)
	if err != nil {
		log.Println(fmt.Sprintf("Unable to get short %v : %v", uid, err.Error()))
		return err
	}

	s.ShortPOIHit = true

	_, err = d.db.NewInsert().Model(s).On("CONFLICT (short_alert_id) DO UPDATE").Exec(contxt)

	if err != nil {
		log.Println(fmt.Sprintf("Unable to update short %v : %v", uid, err.Error()))
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

func (d *DB) ShortSetNewAvg(uid string, price float32) error {
	contxt := context.Background()

	s, err := d.GetShort(uid)
	if err != nil {
		log.Println(fmt.Sprintf("Unable to get short %v : %v", uid, err.Error()))
		return err
	}

	s.ShortStarting = price

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
