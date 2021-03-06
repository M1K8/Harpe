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

func (d *DB) CreateCrypto(uid, coin, author string, spt, ept, poi, stop, tstop float32, alertType int, starting float32) (chan bool, bool, error) {
	chanMap.LoadOrStore(d.Guild, &sync.Map{})
	exists, exitChan := d.GetExitChanExists(uid)

	if exists {
		return exitChan, exists, nil
	}

	contxt := context.Background()

	s := &Crypto{
		CryptoAlertID:      uid,
		CryptoGuildID:      d.Guild,
		CryptoCoin:         coin,
		CryptoStarting:     starting,
		CryptoHighest:      starting,
		CryptoLastHigh:     starting,
		AlertType:          alertType,
		CryptoCallTime:     time.Now(),
		CryptoEPt:          ept,
		CryptoSPt:          spt,
		CryptoStop:         stop,
		CryptoTrailingStop: tstop,
		CryptoPoI:          poi,
		CryptoPOIHit:       false,
		Caller:             author,
	}

	_, err := d.db.NewInsert().Model(s).On("CONFLICT (crypto_alert_id) DO UPDATE").Exec(contxt)

	if err != nil {
		log.Println(fmt.Sprintf("Unable to create Crypto %v : %v", coin, err.Error()))
		return nil, false, err
	}

	return exitChan, exists, nil
}

func (d *DB) RemoveCrypto(uid string) error {

	contxt := context.Background()

	s := &Crypto{
		CryptoGuildID: d.Guild,
		CryptoAlertID: uid,
	}
	res, err := d.db.NewDelete().Model(s).Where("crypto_alert_id = ?", uid).Exec(contxt)
	rowsAffected, _ := res.RowsAffected()

	if rowsAffected == 0 {
		err = errors.New(fmt.Sprintf("Unable to remove Crypto %v : NOT FOUND", uid))
		return err
	}

	if err != nil {
		log.Println(fmt.Sprintf("Unable to remove Crypto %v : %v", uid, err.Error()))
		return err
	}
	clearFromSyncMap(chanMap, d.Guild, uid)
	return nil
}

func (d *DB) GetCrypto(uid string) (*Crypto, error) {

	contxt := context.Background()

	s := &Crypto{
		CryptoGuildID: d.Guild,
		CryptoAlertID: uid,
	}
	err := d.db.NewSelect().Model(s).Where("crypto_alert_id = ?", uid).Scan(contxt)

	if err != nil {
		log.Println(fmt.Sprintf("Unable to get Crypto %v : %v", uid, err.Error()))
		return nil, err
	}
	gMap, ok := chanMap.Load(d.Guild)
	if ok {
		gMapCast := gMap.(*sync.Map)
		_, ok := gMapCast.Load(uid)
		if !ok {
			log.Println(fmt.Sprintf("Unable to get alert channel for crypto %v. Please try recreating this alert, or calling !refresh then running this command again.", uid))
			return nil, err
		}
	}

	return s, nil
}

func (d *DB) CryptoPOIHit(uid string) error {
	contxt := context.Background()

	s, err := d.GetCrypto(uid)
	if err != nil {
		log.Println(fmt.Sprintf("Unable to get Crypto %v : %v", uid, err.Error()))
		return err
	}

	s.CryptoPOIHit = true

	_, err = d.db.NewInsert().Model(s).On("CONFLICT (crypto_alert_id) DO UPDATE").Exec(contxt)

	if err != nil {
		log.Println(fmt.Sprintf("Unable to update Crypto %v : %v", uid, err.Error()))
		return err
	}

	return nil
}

func (d *DB) CryptoSetNewHigh(uid string, price float32) error {
	contxt := context.Background()

	s, err := d.GetCrypto(uid)
	if err != nil {
		log.Println(fmt.Sprintf("Unable to get Crypto %v : %v", uid, err.Error()))
		return err
	}

	s.CryptoHighest = price

	_, err = d.db.NewInsert().Model(s).On("CONFLICT (crypto_alert_id) DO UPDATE").Exec(contxt)

	if err != nil {
		log.Println(fmt.Sprintf("Unable to update Crypto %v : %v", uid, err.Error()))
		return err
	}

	return nil
}

func (d *DB) CryptoSetNewAvg(uid string, price float32) error {
	contxt := context.Background()

	s, err := d.GetCrypto(uid)
	if err != nil {
		log.Println(fmt.Sprintf("Unable to get Crypto %v : %v", uid, err.Error()))
		return err
	}

	s.CryptoStarting = price

	_, err = d.db.NewInsert().Model(s).On("CONFLICT (crypto_alert_id) DO UPDATE").Exec(contxt)

	if err != nil {
		log.Println(fmt.Sprintf("Unable to update Crypto %v : %v", uid, err.Error()))
		return err
	}

	return nil
}

func (c Crypto) GetPctGain(highest float32) float32 {
	return ((highest - c.CryptoStarting) / c.CryptoStarting) * 100
}
