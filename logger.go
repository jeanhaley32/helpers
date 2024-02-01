package logger

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type ch chan any

var (
	chBufSize = 100
)

// various channels used to receive logs.
type channels struct {
	crit  ch
	err   ch
	warn  ch
	info  ch
	debug ch
	SIGS  chan os.Signal
	shutdown  chan interface{}
	quit chan interface{}
}

// Struct defining a Custom Logger
type Mylogger struct {
	start      time.Time
	chans      channels
	baseLogger *log.Logger
}

// genericshutdownSequence is a function that will run when the server is stopped.
func (m Mylogger) genericshutdownSequence(e error) {
	returnCode := 0
	if e != nil {
		log.Default().Printf("%v", e)
		returnCode = 1
	}
	m.baseLogger.Println("Server stopped")
	m.baseLogger.Printf("Server ran for %s", time.Since(m.StartTime()))
	fmt.Println("shutdowning...")
	os.Exit(returnCode)
}

// Begin the logging process
// Returns a pointer to a Mylogger struct
// Example:
// l := StartLogger(log.Default())
// l.Debug("Debug message")
// l.Error("Error message")...
func StartLogger(logger *log.Logger) *Mylogger {
	CRITCH := make(ch, chBufSize)
	ERRCH := make(ch, chBufSize)
	WARNCH := make(ch, chBufSize)
	INFOCH := make(ch, chBufSize)
	DEBUGCH := make(ch, chBufSize)
	SIGS := make(chan os.Signal, 1)
	SHUTDOWN := make(chan interface{}, 1)
	signal.Notify(SIGS, syscall.SIGINT, syscall.SIGTERM)
	l := Mylogger{
		baseLogger: logger,
	}
	l.chans = channels{
		crit:  CRITCH,
		err:   ERRCH,
		warn:  WARNCH,
		info:  INFOCH,
		debug: DEBUGCH,
		shutdown:  SHUTDOWN,
		SIGS:  SIGS,
	}
	go func() {
		mediateChannels(&l)
	}()
	return &l
}

// mediates Log messages between the various channels.
func mediateChannels(m *Mylogger) {
	for {
		select {
		case e := <-m.chans.crit:
			t := colorWrap(PURPLE, "CRITICAL")
			msg := errors.New(t + " " + convertToError(e).Error())
			m.genericshutdownSequence(msg)
		case e := <-m.chans.err:
			t := colorWrap(RED, "ERROR")
			msg := errors.New(t + " " + convertToError(e).Error())
			m.baseLogger.Println(msg.Error())
		case e := <-m.chans.warn:
			t := colorWrap(YELLOW, "WARNING")
			msg := errors.New(t + " " + convertToError(e).Error())
			m.baseLogger.Println(msg.Error())
		case e := <-m.chans.info:
			t := colorWrap(WHITE, "INFO")
			msg := errors.New(t + " " + convertToError(e).Error())
			m.baseLogger.Printf(msg.Error())
		case e := <-m.chans.debug:
			t := colorWrap(BLUE, "DEBUG")
			msg := errors.New(t + " " + convertToError(e).Error())
			m.baseLogger.Printf(msg.Error())
		case <-m.chans.shutdown:
			m.genericshutdownSequence(nil)
		case s := <-m.chans.SIGS:
			t := colorWrap(PURPLE, "INTSIGNAL")
			msg := errors.New(t + " " + convertToError(s.String()).Error())
			m.baseLogger.Println(msg.Error())
			m.genericshutdownSequence(nil)
		case s := <-m.chan.quit:
			return
		}
		case <-m.
	}
}

// Ensure that the argument is an error.
func convertToError(a any) error {
	switch t := a.(type) {
	case error:
		return t
	case string:
		return errors.New(t)
	default:
		return nil
	}
}

// Kill the server.
func (m Mylogger) shutdown() {
	m.chans.shutdown <- nil
}

// Returns start time of server.
func (m Mylogger) StartTime() time.Time {
	return m.start
}

// Log Critical Message and shutdown
func (s *Mylogger) Critical(a any) {
	s.chans.crit <- a
}

// Log Error
func (s *Mylogger) Error(a any) {
	s.chans.err <- a
}

// Log Debug Message
func (s *Mylogger) Debug(a any) {
	s.chans.debug <- a
}

// Log Warning
func (s *Mylogger) Warning(a any) {
	s.chans.warn <- a
}

// Log Information
func (s *Mylogger) Info(a any) {
	s.chans.info <- a
}

// shutsdown logger routine. This is not a graceful exit. 
func (s *Mylogger) Quit(a any){
	s.chan.quit <- a
}
