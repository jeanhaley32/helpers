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
	CRITICAL
	ERROR
	WARNING
	INFO
	INTSIGNAL
	SHUTDOWN
	QUIT
)

var (
	crit, err, warn, info, debug, sigs, shutdown, quit ch      // various channels used to receive logs.
	verboseDefault                                     = false // verbose is set to false by default.
	debugColor                                         = BLUE
	critColor                                          = PURPLE
	errColor                                           = RED
	warnColor                                          = YELLOW
	baseColor                                          = WHITE
	timeFormat                                         = "2006-01-02 15:04:05"
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
	case SHUTDOWN:
		shutdown = make(ch, 1)
		return shutdown
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
	case INFO:
		info = make(ch, chBufSize)
		return info
	}
	return make(ch, chBufSize)
}

func (e errorType) initLog() *log.Logger {
	return log.New(os.Stderr, fmt.Sprintf("%v", e), log.Lshortfile)
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
	verbose  bool
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
				l.debuglog.Println(cioe(m).Error())
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
func StartLogger(isVerbose ...bool) *Mylogger {
	warnlog := WARNING.initLog()
	errlog := ERROR.initLog()
	critlog := CRITICAL.initLog()
	debuglog := DEBUG.initLog()
	infolog := INFO.initLog()

	wg := &sync.WaitGroup{} // waitgroup is intended to track the number of active goroutines.
	quit := make(chan any, 1)
	sigs := make(chan os.Signal, 1)
	shutdown := make(chan interface{}, 1)
	l := Mylogger{
		wg:       wg,
		start:    time.Now(), // Set start time of the server.
		warnlog:  warnlog,
		errlog:   errlog,
		critlog:  critlog,
		debuglog: debuglog,
		infolog:  infolog,
		verbose: func() bool {
			if len(isVerbose) > 0 {
				return isVerbose[0]
			} else {
				return verboseDefault
			}
		}(),
	}
	l.chans = channels{
		crit:     CRITICAL.initChan(),
		err:      ERROR.initChan(),
		warn:     WARNING.initChan(),
		info:     INFO.initChan(),
		debug:    DEBUG.initChan(),
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
		case e := <-l.chans.err:
			l.errlog.Println(cioe(e).Error())
		case e := <-l.chans.warn:
			l.warnlog.Println(cioe(e).Error())
		case e := <-l.chans.info:
			l.infolog.Println(cioe(e).Error())
		case e := <-l.chans.debug:
			l.debuglog.Println(cioe(e).Error())
		case <-l.chans.shutdown:
			l.genericshutdownSequence(nil)
		case s := <-l.chans.sigs:
			l.infolog.Println("Received Signal: ", s.String())
			l.genericshutdownSequence(nil)
		case <-l.chans.quit:
			l.warnlog.Println("Received Quit Signal, shutting down logger")
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
func (l Mylogger) Shutdown() {
	l.chans.shutdown <- nil
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
