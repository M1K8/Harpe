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

// TODO - read / write  Max | Min. Here and repo
package db

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

func (d *DB) CreateCrypto(coin, author string, spt, ept, poi, stop float32, channelType int, guildID string, starting float32) (chan bool, bool, error) {
	contxt := context.Background()

	s := &Crypto{
		CryptoAlertID:  coin + "_" + guildID,
		CryptoGuildID:  guildID,
		CryptoCoin:     coin,
		CryptoStarting: starting,
		CryptoHighest:  starting,
		ChannelType:    channelType,
		CryptoCallTime: time.Now(),
		CryptoEPt:      ept,
		CryptoSPt:      spt,
		CryptoStop:     stop,
		CryptoPoI:      poi,
		CryptoPOIHit:   false,
		Caller:         author,
	}

	_, err := d.db.NewInsert().Model(s).On("CONFLICT (crypto_alert_id) DO UPDATE").Exec(contxt)

	if err != nil {
		log.Println(fmt.Sprintf("Unable to create Crypto %v : %v", coin, err.Error()))
		return nil, false, err
	}

	exitChan := make(chan bool, 1)
	chanMap.LoadOrStore(guildID, &sync.Map{})
	exists, exitChan := d.getExitChanExists(guildID, "c_"+coin, exitChan)

	return exitChan, exists, nil
}

func (d *DB) RemoveCrypto(guildID, coin string) error {

	contxt := context.Background()

	s := &Crypto{
		CryptoGuildID:  guildID,
		CryptoCoin:     coin,
		CryptoStarting: 0,
		CryptoCallTime: time.Time{},
	}
	_, err := d.db.NewDelete().Model(s).Where("crypto_alert_id = ?", coin+"_"+guildID).Exec(contxt)

	if err != nil {
		log.Println(fmt.Sprintf("Unable to remove Crypto %v : %v", coin, err.Error()))
		return err
	}
	clearFromSyncMap(chanMap, guildID, "c_"+coin)
	return nil
}

func (d *DB) GetCrypto(guildID, coin string) (*Crypto, error) {

	contxt := context.Background()

	s := &Crypto{
		CryptoGuildID:  guildID,
		CryptoCoin:     coin,
		CryptoStarting: 0,
		CryptoCallTime: time.Time{},
	}
	err := d.db.NewSelect().Model(s).Where("crypto_alert_id = ?", coin+"_"+guildID).Scan(contxt)

	if err != nil {
		log.Println(fmt.Sprintf("Unable to get Crypto %v : %v", coin, err.Error()))
		return nil, err
	}
	gMap, ok := chanMap.Load(guildID)
	if ok {
		gMapCast := gMap.(*sync.Map)
		_, ok := gMapCast.Load("c_" + coin)
		if !ok {
			log.Println(fmt.Sprintf("Unable to get alert channel for crypto %v. Please try recreating this alert, or calling !refresh then running this command again.", coin))
			return nil, err
		}
	}

	return s, nil
}

func (d *DB) CryptoPOIHit(guildID, coin string) error {
	contxt := context.Background()

	s, err := d.GetCrypto(guildID, coin)
	if err != nil {
		log.Println(fmt.Sprintf("Unable to get Crypto %v : %v", coin, err.Error()))
		return err
	}

	s.CryptoPOIHit = true

	_, err = d.db.NewInsert().Model(s).On("CONFLICT (crypto_alert_id) DO UPDATE").Exec(contxt)

	if err != nil {
		log.Println(fmt.Sprintf("Unable to update Crypto %v : %v", coin, err.Error()))
		return err
	}

	return nil
}

func (d *DB) CryptoSetNewHigh(guildID, coin string, price float32) error {
	contxt := context.Background()

	s, err := d.GetCrypto(guildID, coin)
	if err != nil {
		log.Println(fmt.Sprintf("Unable to get Crypto %v : %v", coin, err.Error()))
		return err
	}

	s.CryptoHighest = price

	_, err = d.db.NewInsert().Model(s).On("CONFLICT (crypto_alert_id) DO UPDATE").Exec(contxt)

	if err != nil {
		log.Println(fmt.Sprintf("Unable to update Crypto %v : %v", coin, err.Error()))
		return err
	}

	return nil
}

func (c Crypto) GetPctGain(highest float32) float32 {
	return ((highest - c.CryptoStarting) / c.CryptoStarting) * 100
}
