package main

import (
	"log"
	"time"

	"github.com/Akilakuma/rolling"
)

func bossEvent() error {
	log.Println("🍐🍐this is A event🍐🍐")
	return nil
}

func evEvent() error {
	log.Println("🍊🍊this is B event🍊🍊")
	return nil
}

func fishwave() error {
	log.Println("🍋🍋this is C event🍋🍋")
	return nil
}

func main() {

	eventManger := rolling.NewEM(true, 100)

	// A -> -> -> -> A -> -> -> -> C -> empty
	// 		-> B 		 -> B

	// (主事件1) A
	// start: 00:00 ~ 00:15 +-10s
	// period: 60s
	bossEvent1 := &rolling.Event{
		Name:           "A_event",
		Period:         15,
		IsRepeat:       true,
		Action:         bossEvent,
		PNRandPlusTime: 10,
		IsTripBegan:    true,
	}
	// (主事件1的延伸事件) B
	// start: 01:00 ~ 01:15
	// period: A 出現之後第45秒出現
	extend1 := &rolling.Event{
		Name:             "B_event",
		Period:           45,
		IsRepeat:         false,
		Action:           evEvent,
		PositivePlusTime: 15,
	}
	bossEvent1.ExtendEvent = append(bossEvent1.ExtendEvent, extend1)

	// (主事件2) A
	// start: 02:00 ~ 03:00 +-10s
	// period: 60s
	bossEvent := &rolling.Event{
		Name:           "A_event",
		Period:         105,
		IsRepeat:       true,
		Action:         bossEvent,
		PNRandPlusTime: 10,
	}
	// (主事件2的延伸事件) B
	// period: A 出現之後第45秒出現
	extend2 := &rolling.Event{
		Name:             "B_event",
		Period:           45,
		IsRepeat:         false,
		Action:           evEvent,
		PositivePlusTime: 15,
	}
	bossEvent.ExtendEvent = append(bossEvent.ExtendEvent, extend2)

	// (主事件3) C
	// start: 04:00 ~ 05:00
	// period: 60s
	fishWaveEvent := &rolling.Event{
		Name:      "C_event",
		Period:    120,
		IsRepeat:  true,
		Action:    fishwave,
		PatchTime: 120,
	}

	emptyEvent := &rolling.Event{
		Name:     "empty_Event",
		Period:   60,
		IsRepeat: true,
		Action:   nil,
	}

	eventManger.PushEvent(bossEvent1)
	eventManger.PushEvent(bossEvent)
	eventManger.PushEvent(fishWaveEvent)
	eventManger.PushEvent(emptyEvent)

	log.Println("===trip began ===")
	go eventManger.Running()

	for {
		time.Sleep(100 * time.Minute)
	}
}
