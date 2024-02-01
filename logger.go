package logger

import (
	"errors"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type ch chan any

type errorType int

const (
	DEBUG errorType = iota
	CRITICAL
	ERROR
	WARNING
	INFO
	INTSIGNAL
)

var (
	crit, err, warn, info, debug, sigs ch // various channels used to receive logs.
)

func (e errorType) String() string {
	switch e {
	case DEBUG:
		return "Debug"
	case CRITICAL:
		return "Critical"
	case ERROR:
		return "Error"
	case WARNING:
		return "Warning"
	case INFO:
		return "Info"
	case INTSIGNAL:
		return "Interrupt Signal"
	}
	return "Unknown Error Type"
}

func (e errorType) Color() Color {
	switch e {
	case DEBUG:
		return BLUE
	case CRITICAL:
		return PURPLE
	case ERROR:
		return RED
	case WARNING:
		return YELLOW
	case INFO:
		return WHITE
	}
	return WHITE
}

func (e errorType) prefix() string {
	switch e {
	case DEBUG:
		return colorWrap(DEBUG.Color(), "DEBUG")
	case CRITICAL:
		return colorWrap(CRITICAL.Color(), "CRITICAL")
	case ERROR:
		return colorWrap(ERROR.Color(), "ERROR")
	case WARNING:
		return colorWrap(WARNING.Color(), "WARNING")
	case INFO:
		return colorWrap(INFO.Color(), "INFO")
	}
	return "UNKNOWN"
}

func (e errorType) channel() ch {
	switch e {
	case DEBUG:
		return debug
	case CRITICAL:
		return crit
	case ERROR:
		return err
	case WARNING:
		return warn
	case INFO:
		return info
	case INTSIGNAL:
		return sigs
	}
	return info
}

var (
	// buffer size for channels
	chBufSize = 100
)

// Struct defining the various channels used to log messages.
type channels struct {
	crit     ch
	err      ch
	warn     ch
	info     ch
	debug    ch
	sigs     chan os.Signal
	shutdown chan interface{}
	quit     chan interface{}
}

// Struct defining a Custom Logger
type Mylogger struct {
	start      time.Time
	chans      channels
	wg         *sync.WaitGroup
	baseLogger *log.Logger
}

// Close all log channels
func (l *Mylogger) closeLogs() {
	close(l.chans.debug)
	close(l.chans.crit)
	close(l.chans.err)
	close(l.chans.info)
	close(l.chans.warn)
}

// generic shutdown sequence
func (l Mylogger) genericshutdownSequence(e error) {

	// define function used to drain channels
	drain := func(e errorType) {
		for m := range e.channel() {
			switch e {
			case CRITICAL:
				err := convertToError(m)
				msg := errors.New(CRITICAL.prefix() + " " + err.Error())
				l.baseLogger.Printf(msg.Error())
			case ERROR:
				err := convertToError(m)
				msg := errors.New(ERROR.prefix() + " " + err.Error())
				l.baseLogger.Printf(msg.Error())
			case WARNING:
				err := convertToError(m)
				msg := errors.New(WARNING.prefix() + " " + err.Error())
				l.baseLogger.Printf(msg.Error())
			case INFO:
				err := convertToError(m)
				msg := errors.New(INFO.prefix() + " " + err.Error())
				l.baseLogger.Printf(msg.Error())
			case DEBUG:
				err := convertToError(m)
				msg := errors.New(DEBUG.prefix() + " " + err.Error())
				l.baseLogger.Printf(msg.Error())
			}
		}
	}

	// Close shutdown channel first. This should be used to signal the end of the server.
	close(l.chans.shutdown)

	// drain and close all channels.
	drain(CRITICAL)
	drain(ERROR)
	drain(WARNING)
	drain(INFO)
	drain(DEBUG)
	// close all channels
	l.closeLogs()
	l.baseLogger.Println("All log channels drained and closed successfully")

	// wait for wg to be cleared.
	l.wg.Wait()
	l.baseLogger.Println("All goroutines have been cleared")
	l.baseLogger.Println("Server stopped")
	l.baseLogger.Printf("Server ran for %s", time.Since(l.StartTime()))
	l.baseLogger.Printf("Shutting Down...")
	if e != nil {
		l.baseLogger.Println("Server exited with error: ", e.Error())
		os.Exit(1)
	}
	os.Exit(0)
}

// Begin the logging process
// Returns a pointer to a Mylogger struct
// Example:
// l := StartLogger(log.Default())
// l.Debug("Debug message")
// l.Error("Error message")...
func StartLogger(logger *log.Logger) *Mylogger {
	wg := &sync.WaitGroup{} // waitgroup is intended to track the number of active goroutines.
	crit := make(ch, chBufSize)
	err := make(ch, chBufSize)
	warn := make(ch, chBufSize)
	info := make(ch, chBufSize)
	debug := make(ch, chBufSize)
	quit := make(chan any)
	sigs := make(chan os.Signal, 1)
	shutdown := make(chan interface{}, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	l := Mylogger{
		baseLogger: logger,
		wg:         wg,
		start:      time.Now(), // Set start time of the server.
	}
	l.chans = channels{
		crit:     crit,
		err:      err,
		warn:     warn,
		info:     info,
		debug:    debug,
		shutdown: shutdown,
		sigs:     sigs,
		quit:     quit,
	}
	go func() {
		mediateChannels(&l)
	}()
	return &l
}

// iterate on waitgroup
func (l *Mylogger) AddToWaitGroup() {
	l.wg.Add(1)
}

// decerement waitgroup
func (l *Mylogger) Done() {
	l.wg.Done()
}

// mediates Log messages between the various channels.
func mediateChannels(m *Mylogger) {
	for {
		select {
		case e := <-m.chans.crit:
			err := convertToError(e)
			msg := errors.New(CRITICAL.prefix() + " " + err.Error())
			m.genericshutdownSequence(msg)
		case e := <-m.chans.err:
			err := convertToError(e)
			msg := errors.New(ERROR.prefix() + " " + err.Error())
			m.baseLogger.Println(msg.Error())
		case e := <-m.chans.warn:
			err := convertToError(e)
			msg := errors.New(WARNING.prefix() + " " + err.Error())
			m.baseLogger.Println(msg.Error())
		case e := <-m.chans.info:
			err := convertToError(e)
			msg := errors.New(INFO.prefix() + " " + err.Error())
			m.baseLogger.Printf(msg.Error())
		case e := <-m.chans.debug:
			err := convertToError(e)
			msg := errors.New(DEBUG.prefix() + " " + err.Error())
			m.baseLogger.Printf(msg.Error())
		case <-m.chans.shutdown:
			m.genericshutdownSequence(nil)
		case s := <-m.chans.sigs:
			t := colorWrap(PURPLE, "INTSIGNAL")
			err := convertToError(s)
			msg := errors.New(t + " " + err.Error())
			m.baseLogger.Println(msg.Error())
			m.genericshutdownSequence(nil)
		case s := <-m.chans.shutdown:
			err := convertToError(s)
			m.genericshutdownSequence(err)
		case <-m.chans.quit:
			return
		}
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
func (m Mylogger) Shutdown() {
	m.chans.shutdown <- nil
}

// Returns start time of server.
func (m Mylogger) StartTime() time.Time {
	return m.start
}

// Log Critical Error and shutdown
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
func (s *Mylogger) Quit(a any) {
	s.chans.quit <- a
}
