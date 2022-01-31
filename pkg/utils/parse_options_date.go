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
package utils

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/uniplaces/carbon"
)

func ParseDate(in string) (m, d, y string, err error) {
	strs := strings.Split(in, "/")

	defer func() {
		if e := recover(); e != nil {
			log.Println(e)
		}
	}()

	if len(strs) != 2 && len(strs) != 3 {
		return "", "", "", errors.New("invalid date string provided")
	}
	m = strs[0]
	if len(m) == 1 {
		m = "0" + m
	}

	d = strs[1]
	if len(d) == 1 {
		d = "0" + d
	}

	if len(strs) == 3 {
		y = "20" + strs[2]
	} else {
		nowC := carbon.NewCarbon(time.Now()).Year()
		y = fmt.Sprintf("%v", nowC)
	}

	return
}
