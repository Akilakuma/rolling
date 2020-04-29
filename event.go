package rolling

import (
	"math/rand"
	"time"
)

// Event 事件資料
// 1. 新增指定隨機秒數(範圍是指定值的正負值)
// 2. 新增指定隨機秒數(正值)
// 3. 新增補正秒數(0的話不需補正)
type Event struct {
	Name             string       // 事件名稱
	Period           int          // timer時間
	IsRepeat         bool         // 是否重複，是的話會塞回channel的末端
	Action           func() error // timer時間到，要執行的method
	PNRandPlusTime   int          // 隨機增加時間，+-指定秒數，ex:10  ---> -10 ~ 10
	PositivePlusTime int          // 隨機增加時間，ex:10  ---> 0 ~ 10
	totalTime        int          // 總時長 = Period + PNRandPlusTime + PositivePlusTime
	// 補正秒數
	// 到這個事件之前預計經過多少時間，差值在這個事件發生時間做補正
	// ex:1 預期到這事件之前，N個事件是總共經過120秒，實際上因為前面事件有新增隨機秒數的關係，只花了115秒，
	// 則在這個事件timer多+5秒
	// ex2: 預期到這事件之前，N個事件是總共經過120秒，實際上因為前面事件有新增隨機秒數的關係，花了132秒，
	// 則在這個事件timer做-12秒的處理
	PatchTime   int
	ExtendEvent []*Event // 延伸事件
	IsTripBegan bool     // 是否為每個trip的起點
}

// getPlusRangeTime 事件timer時間挑整
func (e *Event) getPlusRangeTime(em *EventManager, t int) int {

	// 加隨機時間(正負差)
	if e.PNRandPlusTime != 0 {
		rand.Seed(time.Now().UnixNano())

		// 0 --> 負   1--> 正
		pn := rand.Intn(1)
		x := rand.Intn(e.PNRandPlusTime)

		if pn == 0 {
			t = t - x
		} else {
			t = t + x
		}
	}

	// 加隨機時間(正差)
	if e.PositivePlusTime != 0 {
		rand.Seed(time.Now().UnixNano())

		x := rand.Intn(e.PositivePlusTime)
		t = t + x
	}

	// 時間需要補正
	if e.PatchTime != 0 {
		patchTime := e.PatchTime - em.tripProcessingTime
		t = t + patchTime
		if t < 0 {
			t = 0
		}
	}

	return t
}

// subEventExecuting 延伸事件執行
func (e *Event) subEventExecuting(em *EventManager, parentTime int) {

	t := e.getPlusRangeTime(em, e.Period)
	e.totalTime = t + parentTime

	subEventTimer := time.NewTimer(time.Duration(e.totalTime) * time.Second)
	select {
	case <-subEventTimer.C:
		subEventTimer.Stop()
		if e.Action != nil {
			e.Action()
		}

		// 往下擴張延伸事件
		// 請注意goroutine leak的問題
		for _, graChildEvent := range e.ExtendEvent {
			go graChildEvent.subEventExecuting(em, e.totalTime)
		}
	}
}
