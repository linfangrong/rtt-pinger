package pinger

import (
	"container/list"
	"sort"
	"sync"
	"time"
)

var (
	RoundSuccessCount int           = 3
	RoundDuration     time.Duration = 5 * time.Millisecond

	RttDataManagerDuration time.Duration = 2 * time.Minute
)

type RttData struct {
	timeStamp time.Time
	rtt       time.Duration
}

func NewRttData(rtt time.Duration) (rd *RttData) {
	return &RttData{
		timeStamp: time.Now(),
		rtt:       rtt,
	}
}

type RttDataMapItem struct {
	key   string
	value *list.List
}

func NewRttDataMapItem(key string) (rdmi *RttDataMapItem) {
	return &RttDataMapItem{
		key:   key,
		value: list.New(),
	}
}

func (rdmi *RttDataMapItem) Add(rtt time.Duration) {
	rdmi.value.PushBack(NewRttData(rtt))
	var (
		timeStamp time.Time = time.Now().Add(-RttDataManagerDuration)
		item      *list.Element
	)
	for item = rdmi.value.Front(); item != nil; item = item.Next() {
		if item.Value.(*RttData).timeStamp.Before(timeStamp) {
			rdmi.value.Remove(item)
			continue
		}
		break
	}
}

func (rdmi *RttDataMapItem) SuccessCount() int {
	return rdmi.value.Len()
}

func (rdmi *RttDataMapItem) AvgRtt() time.Duration {
	var (
		total  time.Duration
		item   *list.Element
		length time.Duration = time.Duration(rdmi.value.Len())
	)
	for item = rdmi.value.Front(); item != nil; item = item.Next() {
		total += item.Value.(*RttData).rtt
	}
	return total / length
}

type RttDataMapItemSortByStrategy []*RttDataMapItem

func (s RttDataMapItemSortByStrategy) Len() int { return len(s) }

func (s RttDataMapItemSortByStrategy) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

func (s RttDataMapItemSortByStrategy) Less(i, j int) bool {
	var (
		successCountI int           = s[i].SuccessCount()
		successCountJ int           = s[j].SuccessCount()
		avgRttI       time.Duration = s[i].AvgRtt()
		avgRttJ       time.Duration = s[j].AvgRtt()
	)
	switch {
	case successCountI >= successCountJ && avgRttI <= avgRttJ:
		return true
	case successCountI <= successCountJ && avgRttI >= avgRttJ:
		return false

	case successCountI <= successCountJ && avgRttI <= avgRttJ:
		if successCountJ-successCountI < RoundSuccessCount {
			return true
		}
		return false
	default: // successCountI >= successCountJ && avgRttI >= avgRttJ
		if avgRttI-avgRttJ < RoundDuration {
			return true
		}
		return false
	}
}

type RttDataManager struct {
	sync.RWMutex
	data  map[string]*RttDataMapItem
	slice RttDataMapItemSortByStrategy
}

func NewRttDataManager() (rdm *RttDataManager) {
	return &RttDataManager{
		data: make(map[string]*RttDataMapItem),
	}
}

func (rdm *RttDataManager) Add(key string, rtt time.Duration) {
	rdm.Lock()
	var (
		rdmi *RttDataMapItem
		ok   bool
	)
	if rdmi, ok = rdm.data[key]; !ok {
		rdmi = NewRttDataMapItem(key)
		rdm.data[key] = rdmi
		rdm.slice = append(rdm.slice, rdmi)
	}
	rdmi.Add(rtt)
	rdm.Unlock()
}

func (rdm *RttDataManager) TopN(n int) (top []string) {
	rdm.Lock()
	sort.Sort(rdm.slice)
	var (
		index int
		rdmi  *RttDataMapItem
	)
	for index, rdmi = range rdm.slice {
		if index >= n {
			break
		}
		top = append(top, rdmi.key)
	}
	rdm.Unlock()
	return
}
