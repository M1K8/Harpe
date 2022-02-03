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

func (d *DB) CreateOption(oID, author string, channelType int, ticker, contractType, day, month, year string, price, starting, pt, poi, stop, underStart float32) (chan bool, string, bool, error) {

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
		OptionAlertID:            oID + "_" + d.guild,
		OptionGuildID:            d.guild,
		OptionTicker:             ticker,
		OptionUid:                oID,
		OptionDay:                day,
		OptionContractType:       contractType,
		OptionMonth:              month,
		OptionYear:               year,
		OptionStrike:             price,
		OptionStarting:           starting,
		ChannelType:              channelType,
		OptionCallTime:           time.Now(),
		OptionHighest:            starting,
		OptionUnderlyingPoI:      poi,
		OptionUnderlyingStop:     stop,
		OptionUnderlyingStarting: underStart,
		OptionUnderlyingPOIHit:   false,
		Caller:                   author,
	}

	_, err := d.db.NewInsert().Model(s).On("CONFLICT (option_alert_id) DO UPDATE").Exec(context.Background())

	if err != nil {
		log.Println(fmt.Sprintf("Unable to create option %v: %v.", oID, err.Error()))
		return nil, oID, false, err
	}

	exitChan := make(chan bool, 1)

	chanMap.LoadOrStore(d.guild, &sync.Map{})
	exists, exitChan := d.getExitChanExists(oID, exitChan)

	return exitChan, oID, exists, nil
}

func (d *DB) RemoveOption(oID, contractType, day, month, year string, price float32) error {
	if len(year) != 4 {
		return errors.New("invalid Syntax - year is incorrect")
	}

	if len(month) > 2 || len(month) == 0 {
		return errors.New("invalid Syntax - month is incorrect")
	}

	if len(day) > 2 || len(day) == 0 {
		return errors.New("invalid Syntax - day is incorrect")
	}

	if len(month) == 1 {
		month = "0" + month
	}

	if len(day) == 1 {
		day = "0" + day
	}

	contxt := context.Background()

	s := &Option{
		OptionGuildID:  d.guild,
		OptionUid:      oID,
		OptionStarting: 0,
		OptionCallTime: time.Time{},
	}
	_, err := d.db.NewDelete().Model(s).Where("option_alert_id = ?", oID+"_"+d.guild).Exec(contxt)

	if err != nil {
		log.Println(fmt.Sprintf("Unable to delete option %v: %v.", oID, err.Error()))
		return err
	}
	clearFromSyncMap(chanMap, d.guild, oID)
	return nil
}

func (d *DB) RemoveOptionByCode(oID string) error {
	contxt := context.Background()

	s := &Option{
		OptionGuildID:  d.guild,
		OptionUid:      oID,
		OptionStarting: 0,
		OptionCallTime: time.Time{},
	}
	_, err := d.db.NewDelete().Model(s).Where("option_alert_id = ?", oID+"_"+d.guild).Exec(contxt)

	if err != nil {
		log.Println(fmt.Sprintf("Unable to remove option %v: %v.", oID, err.Error()))
		return err
	}
	clearFromSyncMap(chanMap, d.guild, oID)
	return nil
}

func (d *DB) GetOption(oID string) (*Option, error) {

	contxt := context.Background()

	s := &Option{
		OptionGuildID: d.guild,
		OptionUid:     oID,
	}
	err := d.db.NewSelect().Model(s).Where("option_alert_id = ?", oID+"_"+d.guild).Scan(contxt)

	if err != nil {
		log.Println(fmt.Sprintf("Unable to get option %v: %v.", oID, err.Error()))
		return nil, err
	}
	gMap, ok := chanMap.Load(d.guild)
	if ok {
		gMapCast := gMap.(*sync.Map)
		_, ok := gMapCast.Load(oID)
		if !ok {
			log.Println(fmt.Sprintf("Unable to get alert channel for option %v. Please try recreating this alert, or calling !refresh then running this command again.", oID))
			return nil, err
		}
	}
	return s, nil
}

func (d *DB) OptionPOIHit(oID string) error {
	contxt := context.Background()
	s, err := d.GetOption(oID)
	if err != nil {
		log.Println(fmt.Sprintf("Unable to get option %v : %v", oID, err.Error()))
		return err
	}

	s.OptionUnderlyingPOIHit = true

	_, err = d.db.NewInsert().Model(s).On("CONFLICT (option_alert_id) DO UPDATE").Exec(contxt)

	if err != nil {
		log.Println(fmt.Sprintf("Unable to update option %v : %v", oID, err.Error()))
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

func (o Option) GetPctGain(highest float32) float32 {
	return ((highest - o.OptionStarting) / o.OptionStarting) * 100
}
