package singleflight

import "sync"

type call struct {
	wait_group      sync.WaitGroup
	value           interface{}
	error_value     error
	duplicate_count int
}

// Group manages duplicate suppression.
type Group struct {
	mutex    sync.Mutex
	call_map map[string]*call
}

// Do executes and suppresses duplicate calls for the same key.
func (group *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error, bool) {
	group.mutex.Lock()
	if group.call_map == nil {
		group.call_map = make(map[string]*call)
	}
	if existing_call, ok := group.call_map[key]; ok {
		existing_call.duplicate_count++
		group.mutex.Unlock()
		existing_call.wait_group.Wait()
		return existing_call.value, existing_call.error_value, true
	}
	new_call := new(call)
	new_call.wait_group.Add(1)
	group.call_map[key] = new_call
	group.mutex.Unlock()

	new_call.value, new_call.error_value = fn()
	new_call.wait_group.Done()

	group.mutex.Lock()
	delete(group.call_map, key)
	group.mutex.Unlock()

	return new_call.value, new_call.error_value, new_call.duplicate_count > 0
}
