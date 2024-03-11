package logger

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
	DONE
	CRITICAL
	ERROR
	WARNING
	INFO
	INTSIGNAL
	QUIT
)

var (
	crit, err, warn, info, debug, sigs, quit, done ch      // various channels used to receive logs.
	verboseDefault                                 = false // verbose is set to false by default.
	debugColor                                     = BLUE
	critColor                                      = PURPLE
	errColor                                       = RED
	warnColor                                      = YELLOW
	baseColor                                      = WHITE
	timeFormat                                     = "2006-01-02 15:04:05"
)

func (e errorType) String() string {
	timeNow := func() string {
		return time.Now().Format(timeFormat)
	}
	switch e {
	case DEBUG:
		return fmt.Sprintf(timeNow() + ":" + colorWrap(e.Color(), "DEBUG:"))
	case CRITICAL:
		return fmt.Sprintf(timeNow() + ":" + colorWrap(e.Color(), "CRITICAL:"))
	case ERROR:
		return fmt.Sprintf(timeNow() + ":" + colorWrap(e.Color(), "ERROR:"))
	case WARNING:
		return fmt.Sprintf(timeNow() + ":" + colorWrap(e.Color(), "WARNING:"))
	case INFO:
		return fmt.Sprintf(timeNow() + ":" + colorWrap(e.Color(), "INFO:"))
	}
	return fmt.Sprintf(timeNow() + ":" + colorWrap(e.Color(), "INFO:"))
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

func (e errorType) initChan() ch {
	switch e {
	case INTSIGNAL:
		sigs = make(ch, 1)
		return sigs
	case QUIT:
		quit = make(ch, 1)
		return quit
	case DEBUG:
		debug = make(ch, chBufSize)
		return debug
	case CRITICAL:
		crit = make(ch, chBufSize)
		return crit
	case ERROR:
		err = make(ch, chBufSize)
		return err
	case WARNING:
		warn = make(ch, chBufSize)
		return warn
	case DONE:
		return make(ch, chBufSize)
	case INFO:
		info = make(ch, chBufSize)
		return info
	}
	return make(ch, chBufSize)
}

func (e errorType) initLog(f *os.File) *log.Logger {
	return log.New(f, fmt.Sprintf("%v", e), log.Lshortfile)
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
	case DONE:
		return done
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
	crit  ch
	err   ch
	warn  ch
	info  ch
	debug ch
	done  ch
	sigs  chan os.Signal
	quit  chan interface{}
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
	verbose  bool
}

// Drain all log channels
func (l *Mylogger) drainLogChannels() {
	defer l.wg.Done()
	chList := []errorType{
		ERROR,
		WARNING,
		INFO,
		DEBUG,
	}
	// define function used to drain channels
	drainAndClose := func(e errorType) {
		select {
		case m := <-e.channel():
			switch e {
			case ERROR:
				l.errlog.Println(cioe(m).Error())
			case WARNING:
				l.warnlog.Println(cioe(m).Error())
			case INFO:
				l.infolog.Println(cioe(m).Error())
			case DEBUG:
				l.debuglog.Println(cioe(m).Error())
			}
		default:
			close(e.channel())
			return
		}
	}

	// create a waitgroup to wait for all channels to be drained.
	// once all channels are drained, close the channels.
	// this is done to ensure that all logs are drained before the channels are closed.
	for _, e := range chList {
		drainAndClose(e)
	}
}

// generic shutdown sequence, return true at end of shutdown
func (l *Mylogger) genericshutdownSequence(e error) bool {
	// close done channel, signaling the intention to shutdown to listening applications.
	close(l.chans.done)
	// and listening applications should decrement from the wait group. Once the waitgroup
	// is zero ensuring that everything is closed, we continue
	l.wg.Wait()
	if l.verbose {
		l.debuglog.Println("All tracked Routines stopped")
	}
	l.infolog.Printf("Server ran for %s", time.Since(l.StartTime()))
	if e != nil {
		l.warnlog.Println("Server exited with error: ", e.Error())
		os.Exit(1)
	}
	l.infolog.Printf("Shutting Down...")
	// after all routines have stopped, drain the channels of logs.
	l.AddToWaitGroup()
	go l.drainLogChannels()
	l.wg.Wait()
	// exit with status 0
	return true
}

// Begin the logging process
// Returns a pointer to a Mylogger struct
// Example:
// l := StartLogger(log.Default())
// l.Debug("Debug message")
// l.Error("Error message")...
func StartLogger(f *os.File, isVerbose ...bool) *Mylogger {
	wg := &sync.WaitGroup{} // waitgroup is intended to track the number of active goroutines.
	quit := make(chan any, 1)
	sigs := make(chan os.Signal, 1)
	crit = make(ch, chBufSize)
	err = make(ch, chBufSize)
	warn = make(ch, chBufSize)
	info = make(ch, chBufSize)
	debug = make(ch, chBufSize)
	done = make(ch, chBufSize)
	l := Mylogger{
		wg:       wg,
		start:    time.Now(), // Set start time of the server.
		warnlog:  WARNING.initLog(f),
		errlog:   ERROR.initLog(f),
		critlog:  CRITICAL.initLog(f),
		debuglog: DEBUG.initLog(f),
		infolog:  INFO.initLog(f),
		verbose: func() bool {
			if len(isVerbose) > 0 {
				return isVerbose[0]
			} else {
				return verboseDefault
			}
		}(),
	}
	l.chans = channels{
		crit:  crit,
		err:   err,
		warn:  warn,
		info:  info,
		debug: debug,
		done:  done,
		sigs:  sigs,
		quit:  quit,
	}
	go func() {
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		// mediate channels
		l.AddToWaitGroup()
		mediateChannels(&l)
	}()
	return &l
}

// Signal the start of a new goroutine to the WaitGroup.
func (l *Mylogger) AddToWaitGroup() {
	l.wg.Add(1)
}

// decerement waitgroup
func (l *Mylogger) Done() {
	l.wg.Done()
}

// mediates Log messages between the various channels.
func mediateChannels(l *Mylogger) {
	for {
		select {
		case <-l.chans.done:
			l.Done()
			return
		case <-l.chans.quit:
			l.warnlog.Println("Received Quit Signal, shutting down logger")
			return
		case e := <-l.chans.err:
			l.errlog.Println(cioe(e).Error())
		case e := <-l.chans.warn:
			l.warnlog.Println(cioe(e).Error())
		case e := <-l.chans.info:
			l.infolog.Println(cioe(e).Error())
		case e := <-l.chans.debug:
			l.debuglog.Println(cioe(e).Error())
		case s := <-l.chans.sigs:
			l.infolog.Println("Received Signal: ", s.String())
			l.genericshutdownSequence(nil)
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
func (l Mylogger) Shutdown(e error) bool {
	return l.genericshutdownSequence(e)
}

// Returns start time of server.
func (l Mylogger) StartTime() time.Time {
	return l.start
}

// Log Critical Error and shutdown
func (l *Mylogger) Critical(a any) {
	// Abort all operations and shutdown server.
	err := cioe(a)
	l.critlog.Fatal(err.Error())
}

// Log Error
func (l *Mylogger) Error(a any) {
	l.chans.err <- a
}

// Log Debug Message
func (l *Mylogger) Debug(a any) {
	// if verbose is set, send to debug channel, else return.
	if l.verbose {
		l.chans.debug <- a
	} else {
		return
	}
}

// Log Warning
func (l *Mylogger) Warning(a any) {
	l.chans.warn <- a
}

// Log Information
func (l *Mylogger) Info(a any) {
	l.chans.info <- a
}

// shutsdown logger routine. This is not a graceful exit.
func (l *Mylogger) Quit(a any) {
	l.chans.quit <- a
}
