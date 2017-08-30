package main

import (
	"sync"
)

var filelock = &FileLock{
	status: make(map[string]byte),
	mutex: sync.Mutex{},
}

type FileLock struct {
	status map[string]byte
	mutex sync.Mutex
}

func (fl *FileLock) RLock(path string) bool {
	fl.mutex.Lock()
	defer fl.mutex.Unlock()
	if v, ok := fl.status[path]; ok {
		if v != 0 {
			v++
			fl.status[path] = v
			return true
		}
		return false
	}
	fl.status[path] = 1
	return true
}

func (fl *FileLock) RUnlock(path string) {
	fl.mutex.Lock()
	defer fl.mutex.Unlock()
	if v, ok := fl.status[path]; ok {
		if v != 0 {
			v--
			if v == 0 {
				delete(fl.status, path)
			} else {
				fl.status[path] = v
			}
			return
		}
		panic("RUnlock a Lock")
	}
	panic("RUnlock before RLock")
}

func (fl *FileLock) Lock(path string) bool {
	fl.mutex.Lock()
	defer fl.mutex.Unlock()
	if _, ok := fl.status[path]; ok {
		return false
	}
	fl.status[path] = 0
	return true
}

func (fl *FileLock) Unlock(path string) {
	fl.mutex.Lock()
	defer fl.mutex.Unlock()
	if v, ok := fl.status[path]; ok {
		if v == 0 {
			delete(fl.status, path)
			return
		}
		panic("Unlock a RLock")
	}
	panic("Unlock before Lock")
}