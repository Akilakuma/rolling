package main

import (
	"log"
	"time"

	"github.com/Akilakuma/rolling"
)

func aEvent() error {
	log.Println("🍐🍐this is A event🍐🍐")
	return nil
}

func bEvent() error {
	log.Println("🍊🍊this is B event🍊🍊")
	return nil
}

func cEvent() error {
	log.Println("🍋🍋this is C event🍋🍋")
	return nil
}

func main() {

	eventManger := rolling.NewEM(true, 100)

	// A -> -> -> -> A -> -> -> -> C ->  -> empty
	// -> B          -> B

	// (主事件1) A
	// start: 00:00 ~ 00:15 +-10s
	// period: 60s
	aEvent1 := &rolling.Event{
		Name:           "A_event",
		Period:         15,
		IsRepeat:       true,
		Action:         aEvent,
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
		Action:           bEvent,
		PositivePlusTime: 15,
	}
	aEvent1.ExtendEvent = append(aEvent1.ExtendEvent, extend1)

	// (主事件2) A
	// start: 02:00 ~ 03:00 +-10s
	// period: 60s
	aEvent := &rolling.Event{
		Name:           "A_event",
		Period:         105,
		IsRepeat:       true,
		Action:         aEvent,
		PNRandPlusTime: 10,
	}
	// (主事件2的延伸事件) B
	// period: A 出現之後第45秒出現
	extend2 := &rolling.Event{
		Name:             "B_event",
		Period:           45,
		IsRepeat:         false,
		Action:           bEvent,
		PositivePlusTime: 15,
	}
	aEvent.ExtendEvent = append(aEvent.ExtendEvent, extend2)

	// (主事件3) C
	// start: 04:00 ~ 05:00
	// period: 60s
	cEvent := &rolling.Event{
		Name:      "C_event",
		Period:    120,
		IsRepeat:  true,
		Action:    cEvent,
		PatchTime: 120,
	}

	emptyEvent := &rolling.Event{
		Name:     "empty_Event",
		Period:   60,
		IsRepeat: true,
		Action:   nil,
	}

	eventManger.PushEvent(aEvent1)
	eventManger.PushEvent(aEvent)
	eventManger.PushEvent(cEvent)
	eventManger.PushEvent(emptyEvent)

	log.Println("===trip began ===")
	go eventManger.Running()

	for {
		time.Sleep(100 * time.Minute)
	}
}
