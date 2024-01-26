package helpers

// A collection of helper functions I use in my projects.
import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// interruptlog is a goroutine that will listen for SIGINT and SIGTERM signals,
// and will run exitSeq when it recieves one.
func interruptlog() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		log.Default().Printf("Recieved %v signal", sig)
		exitSeq(nil)
	}()
}

// exitSeq is a function that will log the time the server ran for, and exit the program.
func exitSeq(e error) {
	if e != nil {
		log.Default().Printf("Error: %v", e)
	}
	log.Default().Println("Server stopped")
	log.Default().Printf("Server ran for %s", time.Since(ServerstartTime))
	os.Exit(0)
}
