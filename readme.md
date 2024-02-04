# Custom Logger for Go Applications
A custom logger in Go with the following features. 

## Key Features:

- **Level-based logging:** Logs messages with different severities (`critical`, `error`, `warning`, `info`, `debug`) to separate channels.
- **Colored output:** Differentiates log levels with colors for better readability.
- **Graceful shutdown:** Manages cleanup of resources and ensures remaining logs are written before exiting.
- **Signal handling:** Responds to system signals (SIGINT, SIGTERM) for graceful shutdown.
- **Asynchronous logging:** Uses channels to prevent blocking of main program execution.
- **Server uptime tracking:** Records server start time for performance insights.

## **Usage:**

### **Start the logger:**

```Go
 logger := StartLogger()
```

 ### **Log messages:**

```Go
logger.Critical("Critical error occurred!")
logger.Error("An error happened.")
logger.Warning("This is a warning.")
logger.Info("Informational message.")
logger.Debug("Debugging details.")
```

### **Initiate shutdown:**
```Go
logger.Shutdown()  // Graceful shutdown
logger.Quit()      // Forced, non-graceful shutdown
```

## **WaitGroup Handling**

> **This package utilizes a `sync.WaitGroup` to manage concurrent goroutines and ensure proper completion before shutdown:**

- **Tracking goroutines:** The `AddToWaitGroup()` function increments the WaitGroup counter, signaling the start of a new goroutine.
- **Signaling completion:** The `Done()` function decrements the counter, indicating that a goroutine has finished.
- **Waiting for completion:** The `genericshutdownSequence` function blocks until the WaitGroup counter reaches zero, ensuring all tracked goroutines have completed before proceeding with shutdown.

**This mechanism guarantees that:**

- Logs from concurrent goroutines are properly captured and written before exiting.
- Resources used by goroutines are released appropriately.
- The application exits in a clean and orderly manner, preventing potential data loss or resource leaks.


## Run Time Example
![](logger.gif)

