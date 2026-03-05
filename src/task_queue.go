package plugin

import (
	timer "astroboxplugin/bindings/astrobox_psys_host_timer"
	"strings"
	"sync"
)

type queueTask struct {
	name string
	fn   func() error
}

var (
	taskQueueMu      sync.Mutex
	taskQueue        []queueTask
	taskQueueRunning bool

	queueDrainMu        sync.Mutex
	queueDrainScheduled bool
)

func EnqueueRpcTask(name string, fn func() error) {
	taskQueueMu.Lock()
	taskQueue = append(taskQueue, queueTask{name: name, fn: fn})
	queueLen := len(taskQueue)
	taskQueueMu.Unlock()

	withState(func(state *DebugState) {
		state.TaskQueueLength = queueLen
	})
	scheduleQueueDrain()
}

func scheduleQueueDrain() {
	queueDrainMu.Lock()
	if queueDrainScheduled {
		queueDrainMu.Unlock()
		return
	}
	queueDrainScheduled = true
	queueDrainMu.Unlock()
	_ = timer.SetTimeout(10, queueDrainPayload).Read()
}

func handleQueueDrainPayloadText(payloadText string) bool {
	if strings.TrimSpace(payloadText) != queueDrainPayload {
		return false
	}

	queueDrainMu.Lock()
	queueDrainScheduled = false
	queueDrainMu.Unlock()

	DrainTaskQueue()

	if !hasPendingRequest() {
		taskQueueMu.Lock()
		needsMore := len(taskQueue) > 0 && !taskQueueRunning
		taskQueueMu.Unlock()
		if needsMore {
			scheduleQueueDrain()
		}
	}
	return true
}

func handleQueueDrainEventPayload(eventPayload string) bool {
	text, err := ExtractPayloadText(eventPayload)
	if err != nil {
		return false
	}
	return handleQueueDrainPayloadText(text)
}

func DrainTaskQueue() {
	if hasPendingRequest() {
		return
	}

	taskQueueMu.Lock()
	if taskQueueRunning {
		taskQueueMu.Unlock()
		return
	}
	if len(taskQueue) == 0 {
		taskQueueMu.Unlock()
		withState(func(state *DebugState) {
			state.TaskQueueLength = 0
			state.TaskQueueBusy = false
		})
		return
	}

	task := taskQueue[0]
	taskQueue = taskQueue[1:]
	taskQueueRunning = true
	queueLen := len(taskQueue)
	taskQueueMu.Unlock()

	withState(func(state *DebugState) {
		state.TaskQueueBusy = true
		state.TaskQueueLength = queueLen
	})

	err := task.fn()
	if err != nil {
		appendLogf("ERROR", "任务失败 [%s]: %v", task.name, err)
	}

	taskQueueMu.Lock()
	taskQueueRunning = false
	queueLen = len(taskQueue)
	taskQueueMu.Unlock()

	withState(func(state *DebugState) {
		state.TaskQueueBusy = false
		state.TaskQueueLength = queueLen
	})

	if hasPendingRequest() {
		return
	}
	DrainTaskQueue()
}

func hasPendingRequest() bool {
	return readState(func(state DebugState) bool {
		return state.Pending != nil
	})
}

func resetTaskQueue() {
	taskQueueMu.Lock()
	taskQueue = nil
	taskQueueRunning = false
	taskQueueMu.Unlock()
	withState(func(state *DebugState) {
		state.TaskQueueLength = 0
		state.TaskQueueBusy = false
	})
}
