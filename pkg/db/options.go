/*
 * Copyright 2022 M1K
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

func (d *DB) CreateOption(uid, oID, author string, alertType int, ticker, contractType, day, month, year string, price, starting, pt, poi, stop, tstop, underStart float32) (chan bool, string, bool, error) {

	chanMap.LoadOrStore(d.Guild, &sync.Map{})
	exists, exitChan := d.GetExitChanExists(uid)

	if exists {
		return exitChan, oID, exists, nil
	}

	if len(year) != 4 {
		return nil, "", false, errors.New("invalid Syntax - year is incorrect")
	}

	if len(month) > 2 || len(month) == 0 {
		return nil, "", false, errors.New("invalid Syntax - month is incorrect")
	}

	if len(day) > 2 || len(day) == 0 {
		return nil, "", false, errors.New("invalid Syntax - day is incorrect")
	}
	s := &Option{
		OptionAlertID:            uid,
		OptionGuildID:            d.Guild,
		OptionTicker:             ticker,
		OptionUid:                oID,
		OptionDay:                day,
		OptionContractType:       contractType,
		OptionMonth:              month,
		OptionYear:               year,
		OptionStrike:             price,
		OptionStarting:           starting,
		AlertType:                alertType,
		OptionCallTime:           time.Now(),
		OptionHighest:            starting,
		OptionLastHigh:           starting,
		OptionTrailingStop:       tstop,
		OptionUnderlyingPoI:      poi,
		OptionUnderlyingStop:     stop,
		OptionUnderlyingStarting: underStart,
		OptionUnderlyingPOIHit:   false,
		Caller:                   author,
	}

	_, err := d.db.NewInsert().Model(s).On("CONFLICT (option_alert_id) DO UPDATE").Exec(context.Background())

	if err != nil {
		log.Println(fmt.Sprintf("Unable to create option %v: %v.", uid, err.Error()))
		return nil, oID, false, err
	}

	return exitChan, oID, exists, nil
}

func (d *DB) RemoveOptionByCode(uid string) error {
	contxt := context.Background()

	s := &Option{
		OptionGuildID: d.Guild,
		OptionAlertID: uid,
	}
	_, err := d.db.NewDelete().Model(s).Where("option_alert_id = ?", uid).Exec(contxt)

	if err != nil {
		log.Println(fmt.Sprintf("Unable to remove option %v: %v.", uid, err.Error()))
		return err
	}
	clearFromSyncMap(chanMap, d.Guild, uid)
	return nil
}

func (d *DB) GetOption(uid string) (*Option, error) {

	contxt := context.Background()

	s := &Option{
		OptionGuildID: d.Guild,
		OptionAlertID: uid,
	}
	err := d.db.NewSelect().Model(s).Where("option_alert_id = ?", uid).Scan(contxt)

	if err != nil {
		log.Println(fmt.Sprintf("Unable to get option %v: %v.", uid, err.Error()))
		return nil, err
	}
	gMap, ok := chanMap.Load(d.Guild)
	if ok {
		gMapCast := gMap.(*sync.Map)
		_, ok := gMapCast.Load(uid)
		if !ok {
			log.Println(fmt.Sprintf("Unable to get alert channel for option %v. Please try recreating this alert, or calling !refresh then running this command again.", uid))
			return nil, err
		}
	}
	return s, nil
}

func (d *DB) OptionPOIHit(uid string) error {
	contxt := context.Background()
	s, err := d.GetOption(uid)
	if err != nil {
		log.Println(fmt.Sprintf("Unable to get option %v : %v", uid, err.Error()))
		return err
	}

	s.OptionUnderlyingPOIHit = true

	_, err = d.db.NewInsert().Model(s).On("CONFLICT (option_alert_id) DO UPDATE").Exec(contxt)

	if err != nil {
		log.Println(fmt.Sprintf("Unable to update option %v : %v", uid, err.Error()))
		return err
	}

	return nil
}

func (d *DB) OptionSetNewHigh(uid string, price float32) error {
	contxt := context.Background()

	s, err := d.GetOption(uid)
	if err != nil {
		log.Println(fmt.Sprintf("Unable to get Option %v : %v", uid, err.Error()))
		return err
	}

	s.OptionHighest = price

	_, err = d.db.NewInsert().Model(s).On("CONFLICT (option_alert_id) DO UPDATE").Exec(contxt)

	if err != nil {
		log.Println(fmt.Sprintf("Unable to update Option %v : %v", uid, err.Error()))
		return err
	}

	return nil
}

func (d *DB) OptionSetNewAvg(uid string, price float32) error {
	contxt := context.Background()

	s, err := d.GetOption(uid)
	if err != nil {
		log.Println(fmt.Sprintf("Unable to get Option %v : %v", uid, err.Error()))
		return err
	}

	s.OptionStarting = price

	_, err = d.db.NewInsert().Model(s).On("CONFLICT (option_alert_id) DO UPDATE").Exec(contxt)

	if err != nil {
		log.Println(fmt.Sprintf("Unable to update Option %v : %v", uid, err.Error()))
		return err
	}

	return nil
}

func (o Option) GetPctGain(highest float32) float32 {
	return ((highest - o.OptionStarting) / o.OptionStarting) * 100
}
