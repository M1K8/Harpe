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

import "sync"

func clearFromSyncMap(chanMap sync.Map, guildID, index string) {
	gMap, ok := chanMap.Load(guildID)
	var exitChan chan bool

	if ok {
		gMapCast := gMap.(*sync.Map)
		e, ok := gMapCast.Load(index)
		if ok {
			exitChan = e.(chan bool)
			exitChan <- true
			gMapCast.Delete(index)
		}
	}
}
