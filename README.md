# zlg

![GitHub](https://img.shields.io/github/license/mshindle/zlg)

**zlg** is a fork of Mark Ignacio's [zerolog-gcp](https://github.com/mark-ignacio/zerolog-gcp) project which creates a LevelWriter for using [zerolog](github.com/rs/zerolog) with Google Cloud Operations Logging (the logging system formerly known as Stackdriver).

Some notable features and changes from zerolog-gcp:

* Creation of the logging.Logger and logging.Client has been separated. Users can now call NewWriter or NewWriterWithClient. The first call will create a default client versus passing an instantiated client. If zlg creates the client, it will execute the logging.Client.Ping function to ensure ability to connect to GCP Cloud Logging.
* All writes are non-blocking.
* Handles converting `zerolog.WarnLevel` to `logging.Warning`.
* Zerolog's TraceLevel and NoLevel maps to Cloud Logging's Default severity.
* Zerolog's FatalLevel maps to Alert severity by default.
* Ensure that all zlg-created clients are closed before program exit with `defer zlg.Close()`. A logging.Client close will flush all associated loggers.

# Getting Started

## The usual cases

Logging only to GCP Cloud Logging:

```go
import "github.com/mshindle/zlg"

// [...]

gcpWriter, err := zlg.NewWriter(ctx, parent, logID, opts...)
if err != nil {
    log.Panic().Err(err).Msg("could not create a GCP Cloud Logging writer")
}
log.Logger = log.Output(gcpWriter)
```

As zlg.NewWriter creates a zerolog.LevelWriter, all of zerolog's features are available such as creating MultiLevel writers.

```go
gcpWriter, err := zlg.NewWriter(ctx, parent, logID, opts...)
if err != nil {
    log.Panic().Err(err).Msg("could not create a GCP Cloud Logging writer")
}
log.Logger = log.Output(zerolog.MultiLevelWriter(
    gcpWriter,
	zerolog.New(os.Stderr)
))
```

To ensure that the last asynchronous logs are delivered to Cloud Logging, zlg keeps a reference to all `logging.Client` entities that zlg itself creates. `Close()` should be called on the package to ensure all created clients are flushed before termination.

```go
gcpWriter, err := zlg.NewCloudLoggingWriter(ctx, projectID, logID, zlg.CloudLoggingOptions{})
if err != nil {
    log.Panic().Err(err).Msg("could not create a CloudLoggingWriter")
}
defer zlg.CLose()
```
