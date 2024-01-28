package logger

// A logger used to log handle errors of different severities.
import (
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type ch chan any

var (
	CRITCH, ERRCH, WARNCH, INFOCH, DEBUGCH ch
)

// LogChan is a type that represents a channel that can be used to send logs
type LogChan int

// channels is a type that represents a collection of channels that can be used
type channels struct {
	crit  ch
	err   ch
	warn  ch
	info  ch
	debug ch
}

const (
	CRITICAL LogChan = iota
	ERROR
	WARNING
	INFO
	DEBUG
)

func (l LogChan) String() string {
	switch l {
	case CRITICAL:
		return colorWrap(PURPLE, "CRITICAL")
	case ERROR:
		return colorWrap(RED, "ERROR")
	case WARNING:
		return colorWrap(YELLOW, "WARNING")
	case INFO:
		return colorWrap(GRAY, "INFO")
	case DEBUG:
		return colorWrap(BLUE, "DEBUG")
	}
	return ""
}

func (l LogChan) GetChannel() ch {
	switch l {
	case CRITICAL:
		return CRITCH
	case ERROR:
		return ERRCH
	case WARNING:
		return WARNCH
	case INFO:
		return INFOCH
	case DEBUG:
		return DEBUGCH
	}
	return nil
}

// Mylogger defines a logger that can be used to log messages to the console.
type Mylogger struct {
	start time.Time
	chans channels
}

// genericExitSequence is a function that will run when the server is stopped.
func (m Mylogger) genericExitSequence(e error) {
	returnCode := 0
	if e != nil {
		log.Default().Printf("Error: %v", e)
		returnCode = 1
	}
	log.Default().Println("Server stopped")
	log.Default().Printf("Server ran for %s", time.Since(m.StartTime()))
	os.Exit(returnCode)
}

// StartLogging starts the logging process.
func (m Mylogger) StartLogging(l *log.Logger) {
	CRITCH := make(ch)
	ERRCH := make(ch)
	WARNCH := make(ch)
	INFOCH := make(ch)
	DEBUGCH := make(ch)
	m.chans = channels{
		crit:  CRITCH,
		err:   ERRCH,
		warn:  WARNCH,
		info:  INFOCH,
		debug: DEBUGCH,
	}
	m.Interruptlog()
	go func() {
		for {
			select {
			case e := <-m.chans.crit:
				t := colorWrap(PURPLE, "CRITICAL")
				msg := errors.New(t + " " + convertToError(e).Error())
				m.genericExitSequence(msg)
			case e := <-m.chans.err:
				t := colorWrap(RED, "ERROR")
				m := errors.New(t + " " + convertToError(e).Error())
				l.Printf(m.Error())
			case e := <-m.chans.warn:
				t := colorWrap(YELLOW, "WARNING")
				m := errors.New(t + " " + convertToError(e).Error())
				l.Printf(m.Error())
			case e := <-m.chans.info:
				t := colorWrap(WHITE, "INFO")
				m := errors.New(t + " " + convertToError(e).Error())
				l.Printf(m.Error())
			case e := <-m.chans.debug:
				t := colorWrap(BLUE, "DEBUG")
				m := errors.New(t + " " + convertToError(e).Error())
				l.Printf(m.Error())
			}
		}
	}()
}

// Ensure that the passed value is an error, and return it as an error.
func convertToError(a any) error {
	switch a.(type) {
	case error:
		return a.(error)
	case string:
		return errors.New(a.(string))
	default:
		return nil
	}
}

// Returns start time of server.
func (m Mylogger) StartTime() time.Time {
	return m.start
}

// Critical logs a critical error and exits the program.
func (s *Mylogger) Critical(e error) {
	s.chans.crit <- e
}

// Error logs an error.
func (s *Mylogger) Error(e error) {
	s.chans.err <- e
}

// Debug logs a debug message.
func (s *Mylogger) Debug(e error) {
	s.chans.debug <- e
}

// Warning logs a warning message.
func (s *Mylogger) Warning(e error) {
	s.chans.warn <- e
}

// Info logs an info message.
func (s *Mylogger) Info(e error) {
	s.chans.info <- e
}

// interruptlog is a goroutine that will listen for SIGINT and SIGTERM signals,
// and will run exitSeq when it recieves one.
func (m Mylogger) Interruptlog() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		log.Default().Printf("Recieved %v signal", sig)
		m.genericExitSequence(nil)
	}()
}
