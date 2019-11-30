package lamport

import (
	"fmt"
	"log"
)

// =================================
//  Event logger, internal use only
// =================================

type Logger struct {
	// index = time step
	// value = events that occurred at that time step
	events [][]LogEvent
}

type LogEvent struct { //server.Id, server.Tokens, event
	serverId string
	// Number of tokens before execution of event执行事件前的tokens
	serverTokens int
	event        interface{}
}

func (event LogEvent) String() string {
	prependWithTokens := false
	switch evt := event.event.(type) {
	case SentMessageEvent:
		switch evt.message.(type) {
		case TokenMessage:
			prependWithTokens = true
		}
	case ReceivedMessageEvent:
		switch evt.message.(type) {
		case TokenMessage:
			prependWithTokens = true
		}
	case StartSnapshot:
		prependWithTokens = true
	case EndSnapshot:
	default:
		log.Fatal("Attempted to log unrecognized event: ", event.event)
	}
	if prependWithTokens {
		return fmt.Sprintf("%v has %v token(s)\n\t%v",
			event.serverId,
			event.serverTokens,
			event.event)
	} else {
		return fmt.Sprintf("%v", event.event)
	}
}

func NewLogger() *Logger {
	return &Logger{make([][]LogEvent, 0)}
}

func (log *Logger) PrettyPrint() {
	for epoch, events := range log.events {
		if len(events) != 0 {
			fmt.Printf("Time %v:\n", epoch)
		}
		for _, event := range events {
			fmt.Printf("\t%v\n", event)
		}
	}
}

func (log *Logger) NewEpoch() {
	log.events = append(log.events, make([]LogEvent, 0))
	//time step and events
}

func (logger *Logger) RecordEvent(server *Server, event interface{}) {
	//参数 例 (server, SentMessageEvent{server.Id, dest, message})
	// index = time step
	// value = events that occurred at that time step
	//events [][]LogEvent
	mostRecent := len(logger.events) - 1  //logger中的enents事件
	events := logger.events[mostRecent]
	events = append(events, LogEvent{server.Id, server.Tokens, event})
	/*
	LogEvent中的字段
	serverId string
	serverTokens int 执行事件前的tokens
	event        interface{}
	*/
	logger.events[mostRecent] = events
	log.Printf("已经记录下了事件\n , %v \n",events)
}
