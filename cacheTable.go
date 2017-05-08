/*
 * Simple caching library with expiration capabilities
 *     Copyright (c) 2013-2017, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE.txt
 */

package Gocache

import (
	"log"
	"sort"
	"sync"
	"time"
)

// CacheTable is a table within the cache
type CacheTable struct {
	sync.RWMutex

	// The table's name.
	name string
	// All cached items.
	items map[interface{}]*CacheItem

	//执行清理时刻
	cleanupTimer *time.Timer
	// 距下一次清理时间
	cleanupInterval time.Duration

	// The logger used for this table.
	logger *log.Logger

	loadData          func(key interface{}, args ...interface{}) *CacheItem
	addedItem         func(item *CacheItem)
	aboutToDeleteItem func(item *CacheItem)
}


func (table *CacheTable) Add(key interface{}, lifeSpan time.Duration, data interface{}) *CacheItem {
	item:=NewCacheItem(key,lifeSpan,data)
	table.Lock()
	table.addInternal(item)
	retun item;
}
func (table *CacheItem)addInternal(item *CacheItem){
	table.items[item.key]=item
	expireDur=table.cleanupInterval
	addItem:=table.addItem
	table.Unlock()
	if addItem!=nil{
		addItem(item)
	}
	if item.lifeSpan>0&&(expireDur==0||item.lifeSpan<expireDur){
	 go	table.expirationCheck()
	}
}
func(table *CacheTable) expirationCheck(){
	table.Lock()
	if table.cleanupTimer !=nil{
		cleanupTimer.stop()
	}
	if table.cleanupInterval>0{
		table.log("Expiration check triggered after", table.cleanupInterval, "for table", table.name)
	}else{
		table.log("Expiration check installed for table", table.name)
	}
	items:=table.items
	now:=time.Now()
	smallestDuration:=0*time.Second
	for key,item:=range items{
		item.RLock()
		lifeSpan:=item.lifeSpan
		accessedOn := item.accessedOn
		item.RUnlock()
		if lifeSpan==0{
			continue
		}
		if(now.Sub(accessedOn)>=lifeSpan){
			table.Delete(key)
		}else{
			if(smallestDuration==0||lifeSpan-now.Sub(accessedOn)<smallestDuration){
				smallestDuration = lifeSpan - now.Sub(accessedOn)
			}
		}
	}
	table.ULock()
	//最小检视时间
	table.cleanupInterval = smallestDuration
	if smallestDuration>0 {
		table.cleanupTimer = time.AfterFunc(smallestDuration, func() {
			go table.expirationCheck()
		})
	}
	table.Unlock()	
}


func (table *CacheTable) log(v ...interface{}) {
	if table.logger == nil {
		return
	}
	table.logger.Println(v)
}
func (table *CacheTable)delete(key interface{})(*CacheItem,error){
	table.Rlock()
	r,ok:=table.items[key]
	if !ok{
		table.RUnlock()
		return nil ,ErrKeyNotFound
	}
	aboutToDeleteItem:=table.aboutToDeleteItem
	table.RUnlock()
		if aboutToDeleteItem != nil {
		aboutToDeleteItem(r)
	}
	r.RLock()
	defer r.RUnlock()
	//过期
	if r.aboutToExpire!=nil{
		r.aboutToExpire(key)
	}
		table.Lock()
	defer table.Unlock()
	table.log("Deleting item with key", key, "created on", r.createdOn, "and hit", r.accessCount, "times from table", table.name)
	delete(table.items, key)

	return r, nil
}


