/*=============================================================================
#     FileName: map.go
#       Author: sunminghong, allen.fantasy@gmail.com, http://weibo.com/5d13
#         Team: http://1201.us
#   LastChange: 2015-09-03 16:23:11
#      History:
=============================================================================*/

package gotools

import (
    "sync"
)

type Map struct {
    lock *sync.RWMutex
    bm   map[int]interface{}
}

func NewMap() *Map {
    return &Map{
        lock: new(sync.RWMutex),
        bm: make(map[int]interface{}),
    }
}

//Get from maps return the k's value
func (m *Map) Get(k int) (interface{},bool) {
    //m.lock.RLock()
    //defer m.lock.RUnlock()

    val, ok := m.bm[k]
    return val,ok
}

// if the key is already in the map and changes nothing.
func (m *Map) Set(k int, v interface{}) {
    m.lock.Lock()
    defer m.lock.Unlock()

    m.bm[k] = v

    //if val, ok := m.bm[k]; !ok {
    //    m.bm[k] = v
    //} else if val != v {
    //    m.bm[k] = v
    //}
}

func (m *Map) All() map[int]interface{} {
    return m.bm
}

func (m *Map) Length() int {
    return len(m.bm)
}

// Returns true if k is exist in the map.
func (m *Map) Check(k int) bool {
    //m.lock.RLock()
    //defer m.lock.RUnlock()

    if _, ok := m.bm[k]; !ok {
        return false
    }
    return true
}

func (m *Map) Delete(k int) {
    m.lock.Lock()
    defer m.lock.Unlock()

    delete(m.bm, k)
}

func (m *Map) Clear() {
    m.lock.Lock()
    defer m.lock.Unlock()

    m.bm = make(map[int]interface{})
}

