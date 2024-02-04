# Custom Logger for Go Applications
A custom logger in golang with the following features. 

## Key Features:

- <b>Level-based logging:</b> Logs messages with different severities (`critical`, `error`, `warning`, `info`, `debug`) to separate channels.
- <b>Colored output:</b> Differentiates log levels with colors for better readability.
- <b>Graceful shutdown:</b> Manages cleanup of resources and ensures remaining logs are written before exiting.
- <b>Signal handling:</b> Responds to system signals (SIGINT, SIGTERM) for graceful shutdown.
- <b>Asynchronous logging:</b> Uses channels to prevent blocking of main program execution.
- <b>Server uptime tracking:</b> Records server start time for performance insights.

## Usage:

### Start the logger:

```Go
 logger := StartLogger()
```

 ### Log messages:

```Go
logger.Critical("Critical error occurred!")
logger.Error("An error happened.")
logger.Warning("This is a warning.")
logger.Info("Informational message.")
logger.Debug("Debugging details.")
```

### Initiate shutdown:
```Go
logger.Shutdown()  // Graceful shutdown
logger.Quit()      // Forced, non-graceful shutdown
```



## Run Time Example
![](logger.gif)

