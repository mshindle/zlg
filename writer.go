package zlg

import (
	"context"
	"encoding/json"

	"cloud.google.com/go/logging"
	"github.com/rs/zerolog"
)

var DefaultLevelMapping = map[zerolog.Level]logging.Severity{
	zerolog.DebugLevel: logging.Debug,
	zerolog.InfoLevel:  logging.Info,
	zerolog.WarnLevel:  logging.Warning,
	zerolog.ErrorLevel: logging.Error,
	zerolog.PanicLevel: logging.Critical,
	zerolog.FatalLevel: logging.Alert,
	zerolog.NoLevel:    logging.Default,
	zerolog.TraceLevel: logging.Default,
}

// keep a reference to all loggers created in the package, so we can flush
// the buffers as the application closes.
var clients = make([]*logging.Client, 0, 1)

// NewWriter creates a LevelWriter that logs only to GCP Cloud Logging using non-blocking calls.
// Writer is created using default client options. Once logging.Client is created, NewWriter will
// execute logging.Client.Ping to ensure application can use GCP Cloud Logging. If a specific
// logging client is needed, use NewWriterWithClient instead.
func NewWriter(ctx context.Context, parent, logID string, opts ...logging.LoggerOption) (zerolog.LevelWriter, error) {
	client, err := logging.NewClient(ctx, parent)
	if err != nil {
		return nil, err
	}

	err = client.Ping(ctx)
	if err != nil {
		return nil, err
	}

	clients = append(clients, client)
	return NewWriterWithClient(client, logID, opts...)
}

// Close waits for all opened loggers for all clients to be flushed.
// Once flushed, the client is then closed as well.
func Close() {
	for _, c := range clients {
		if c != nil {
			_ = c.Close()
		}
	}
}

// NewWriterWithClient instantiates a logging.Logger that will write entries with the given log ID, such as
// "syslog". A log ID must be less than 512 characters long and can only
// include the following characters: upper and lower case alphanumeric
// characters: [A-Za-z0-9]; and punctuation characters: forward-slash,
// underscore, hyphen, and period.
func NewWriterWithClient(client *logging.Client, logID string, opts ...logging.LoggerOption) (zerolog.LevelWriter, error) {
	lw := loggingWriter{
		client:       client,
		logger:       client.Logger(logID, opts...),
		levelMapping: DefaultLevelMapping,
	}
	return &lw, nil
}

type loggingWriter struct {
	client       *logging.Client
	logger       *logging.Logger
	levelMapping map[zerolog.Level]logging.Severity
}

// WriteLevel ensures that message is mapped to the correct GCP Cloud Logging
// severity.
func (lw *loggingWriter) WriteLevel(level zerolog.Level, payload []byte) (int, error) {
	lw.logger.Log(logging.Entry{
		Severity: lw.levelMapping[level],
		Payload:  json.RawMessage(payload),
	})

	// zerolog.FatalLevel will cause application to exit immediately. We need to ensure
	// that everything is flushed and hopefully makes it to GCP
	if level == zerolog.FatalLevel {
		err := lw.logger.Flush()
		if err != nil {
			return 0, err
		}
	}
	return len(payload), nil
}

// Write implements the io.Writer interface. This is useful to set as a writer
// for the standard library log.
func (lw *loggingWriter) Write(p []byte) (int, error) {
	return lw.WriteLevel(zerolog.NoLevel, p)
}

// SetLevelMapping updates the loggingWriter with a custom level
// to severity mapping.
func (lw *loggingWriter) SetLevelMapping(lm map[zerolog.Level]logging.Severity) {
	lw.levelMapping = lm
}
