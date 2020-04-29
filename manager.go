package rolling

import (
	"math/rand"
	"time"
)

// EventManager 事件管理
// 類樹狀結構管理器(只有一個根節點事件)
// 每個事件可以用ExtendEvent掛載延伸事件(開枝散葉)
// 有trip的概念，多個event組合成一個trip
// ex : 事件A，事件B，事件C
// 一個trip 可以是 ABABC 自由搭配組合
// 以上範例，只要在第一個A塞入時，指定isTripBegan = true
// 頭一個事件有設isTripBegan，就會形成trip，
// 每重新執行到第一個event，trip的tripProcessingTime參數會重置
type EventManager struct {
	// jobQueue 是某種程度上的環狀結構概念
	// FIFO 先進先出，取出的Event
	// Event若 IsRepeat = ture 會塞回channel
	jobQueue           chan *Event
	exitMsg            chan struct{}
	isNeedCountDown    bool   // 是否需要知道倒數
	countTime          int    // 倒數時間
	jobWaitingName     string // 目前正在等待timer時間到的主事件名稱
	isRunning          bool   // 是否正在執行
	tripProcessingTime int    // 目前trip已進行時間
}

// 概念意像圖
// o代表某個事件，最上面是主事件，_代表timer倒數時間
// 可以多個主事件構成一個trip，不一定要有trip
// 若構成trip，會記錄從trip的起點事件開始，經過多少時間
// 每個事件可以在設定延伸事件
// o___________o__________o________o_____
// 	 \o___o__   \__o___o____
// 		  \o____o____


// NewEM 新的管理器實體
func NewEM(need bool, buffer int) *EventManager {
	return &EventManager{
		jobQueue:        make(chan *Event, buffer),
		exitMsg:         make(chan struct{}),
		isNeedCountDown: need,
	}
}

// PushEvent 加入event
func (em *EventManager) PushEvent(event *Event) {
	em.jobQueue <- event
}

// PopEvent 拿出event
func (em *EventManager) PopEvent() *Event {
	event, ok := <-em.jobQueue
	if ok {
		return event
	}

	return nil
}

func newTimer(em *EventManager) (*Event, *time.Timer) {

	event := em.PopEvent()
	t := event.Period

	t = event.getPlusRangeTime(em, t)
	event.totalTime = t

	// 建立取出事件的timer
	eventTimer := time.NewTimer(time.Duration(t) * time.Second)
	em.jobWaitingName = event.Name

	// timer開始倒數計數(以秒為單位)
	if em.isNeedCountDown {
		go em.CountDown(t)
	}

	for _, subEvent := range event.ExtendEvent {
		go subEvent.subEventExecuting(em, t)
	}

	return event, eventTimer
}


// Running 啟動執行
func (em *EventManager) Running() {

	em.isRunning = true

	// 拿出事件
	// 查看事件的設定時間
	// 根據這個時間設定一個timer
	Event, EventTimer := newTimer(em)

	for {
		select {
		// timer 到了
		case <-EventTimer.C:
			EventTimer.Stop()

			// trip 的開始
			// trip 的經歷時間重置
			if Event.isTripBegan {
				em.tripProcessingTime = 0
			}

			// 設置trip 經過時間
			// 如果第一個event 沒有設isTripBegan = true
			// 則em.tripProcessingTime就是無限累加
			em.setTripProcessingTime(Event.totalTime)

			// 執行event
			if Event.Action != nil {
				go Event.Action()
			}

			// 如果需要重複的話，塞回channel末端
			if Event.IsRepeat {
				em.PushEvent(Event)
			}

			// 下一輪開始
			Event, EventTimer = newTimer(em)

		case <-em.exitMsg:
			EventTimer.Stop()
			return
		}
	}
}

// setTripProcessingTime 加經過時間
func (em *EventManager) setTripProcessingTime(goneTime int) {

	// 加累積時間
	em.tripProcessingTime += goneTime
}

// Close 關閉事件管理器
func (em *EventManager) Close() {
	if em.isRunning {
		em.exitMsg <- struct{}{}
	}
}

// CountDown 倒數計時
func (em *EventManager) CountDown(t int) {

	countDownPeriod := time.NewTicker(time.Duration(t))

	// 每秒+1的經過時間
	// 不是很聰明的土方法
	secondCount := time.NewTicker(1 * time.Second)

	defer countDownPeriod.Stop()
	defer secondCount.Stop()

	for {
		select {
		case <-countDownPeriod.C:
			em.countTime = 0
			return
		case <-secondCount.C:
			em.countTime++
		}
	}
}

// GetCountDown 取得事件倒數經過時間
func (em *EventManager) GetCountDown() int {
	return em.countTime
}

// GetJobName 取得正在倒數的事件名稱
func (em *EventManager) GetJobName() string {
	return em.jobWaitingName
}
