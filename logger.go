package main

import (
	"errors"
	"fmt"
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
	debugColor                         = BLUE
	critColor                          = PURPLE
	errColor                           = RED
	warnColor                          = YELLOW
	baseColor                          = WHITE
)

func (e errorType) String() string {
	return e.prefix()
}

func (e errorType) Color() Color {
	switch e {
	case DEBUG:
		return debugColor
	case CRITICAL:
		return critColor
	case ERROR:
		return errColor
	case WARNING:
		return warnColor
	case INFO:
		return baseColor
	}
	return baseColor
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
	start    time.Time
	chans    channels
	wg       *sync.WaitGroup
	warnlog  *log.Logger
	errlog   *log.Logger
	critlog  *log.Logger
	debuglog *log.Logger
	infolog  *log.Logger
}

// Close all log channels
func (l *Mylogger) closeLogs() {
	defer l.wg.Done()
	close(l.chans.debug)
	close(l.chans.crit)
	close(l.chans.err)
	close(l.chans.info)
	close(l.chans.warn)
}

// Drain all log channels.
func (l *Mylogger) drainChannel() {
	chList := []errorType{
		CRITICAL,
		ERROR,
		WARNING,
		INFO,
		DEBUG,
	}
	defer l.wg.Done()
	// define function used to drain channels
	drain := func(e errorType) {
		for m := range e.channel() {
			switch e {
			case CRITICAL:
				l.critlog.Println(cioe(m).Error())
			case ERROR:
				l.errlog.Println(cioe(m).Error())
			case WARNING:
				l.warnlog.Println(cioe(m).Error())
			case INFO:
				l.infolog.Println(cioe(m).Error())
			case DEBUG:
			}
		}
	}

	for _, e := range chList {
		drain(e)
	}
	l.infolog.Output(2, "All logs drained. Closing log channels")
	l.closeLogs()
}

// generic shutdown sequence
func (l Mylogger) genericshutdownSequence(e error) {

	// Close shutdown channel first. This should be used to signal the end of the server.
	close(l.chans.shutdown)
	l.wg.Wait()
	l.infolog.Println("All tracked Routines stopped")

	l.infolog.Printf("Server ran for %s", time.Since(l.StartTime()))
	if e != nil {
		l.warnlog.Println("Server exited with error: ", e.Error())
		os.Exit(1)
	}
	l.infolog.Printf("Shutting Down...")
	// after all routines have stopped, drain the channels of logs.
	l.AddToWaitGroup()
	go l.drainChannel()
	l.wg.Wait()
	// exit with status 0
	os.Exit(0)
}

// Begin the logging process
// Returns a pointer to a Mylogger struct
// Example:
// l := StartLogger(log.Default())
// l.Debug("Debug message")
// l.Error("Error message")...
func StartLogger() *Mylogger {
	warnlog := log.New(os.Stderr, fmt.Sprintf("%v", WARNING), log.Ldate|log.Ltime|log.Lshortfile)
	errlog := log.New(os.Stderr, fmt.Sprintf("%v", ERROR), log.Ldate|log.Ltime|log.Lshortfile)
	critlog := log.New(os.Stderr, fmt.Sprintf("%v", CRITICAL), log.Ldate|log.Ltime|log.Lshortfile)
	debuglog := log.New(os.Stderr, fmt.Sprintf("%v", DEBUG), log.Ldate|log.Ltime|log.Lshortfile)
	infolog := log.New(os.Stderr, fmt.Sprintf("%v", INFO), log.Ldate|log.Ltime|log.Lshortfile)

	wg := &sync.WaitGroup{} // waitgroup is intended to track the number of active goroutines.
	crit := make(ch, chBufSize)
	err := make(ch, chBufSize)
	warn := make(ch, chBufSize)
	info := make(ch, chBufSize)
	debug := make(ch, chBufSize)
	quit := make(chan any, 1)
	sigs := make(chan os.Signal, 1)
	shutdown := make(chan interface{}, 1)
	// signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	l := Mylogger{
		wg:       wg,
		start:    time.Now(), // Set start time of the server.
		warnlog:  warnlog,
		errlog:   errlog,
		critlog:  critlog,
		debuglog: debuglog,
		infolog:  infolog,
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
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
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
			m.critlog.Println(cioe(e).Error())
			m.genericshutdownSequence(cioe(e))
		case e := <-m.chans.err:
			m.errlog.Println(cioe(e).Error())
		case e := <-m.chans.warn:
			m.warnlog.Println(cioe(e).Error())
		case e := <-m.chans.info:
			m.infolog.Println(cioe(e).Error())
		case e := <-m.chans.debug:
			m.debuglog.Println(cioe(e).Error())
		case <-m.chans.shutdown:
			m.genericshutdownSequence(nil)
		case s := <-m.chans.sigs:
			m.infolog.Println("Received Signal: ", s.String())
			m.genericshutdownSequence(nil)
		case <-m.chans.quit:
			m.warnlog.Println("Received Quit Signal, shutting down logger")
			return
		}
	}
}

// convert into error
func cioe(a any) error {
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
