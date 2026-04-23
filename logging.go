/*
Copyright © 2021 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

// Generated documentation is available at:
// https://pkg.go.dev/github.com/RedHatInsights/insights-results-aggregator-exporter
//
// Documentation in literate-programming-style is available at:
// https://redhatinsights.github.io/insights-results-aggregator-exporter/packages/logging.html

import (
	"fmt"
	"io"
	"os"

	sentry "github.com/getsentry/sentry-go"
	sentryzerolog "github.com/getsentry/sentry-go/zerolog"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// InitLogging add more writers to zerolog log object. This way the logging can be sent to
// many targets. For the moment just STDOUT and Sentry are configured.
func InitLogging(config *ConfigStruct) (func(), error) {
	var (
		writers       []io.Writer
		writeClosers  []io.WriteCloser
		consoleWriter io.Writer
	)

	loggingConf := GetLoggingConfiguration(config)
	sentryConf := GetSentryConfiguration(config)

	stdOut := os.Stdout
	consoleWriter = stdOut

	if loggingConf.Debug {
		// nice colored output
		consoleWriter = zerolog.ConsoleWriter{Out: stdOut}
	}

	writers = append(writers, consoleWriter)

	if sentryConf.SentryDSN != "" {
		sentryWriter, err := setupSentryLogging(sentryConf)
		if err != nil {
			err = fmt.Errorf("Error initializing Sentry logging: %s", err.Error())
			return func() {}, err
		}
		writers = append(writers, sentryWriter)
		writeClosers = append(writeClosers, sentryWriter)
	}

	logsWriter := zerolog.MultiLevelWriter(writers...)
	log.Logger = zerolog.New(logsWriter).With().Timestamp().Logger()

	return func() {
		log.Info().Msg("Closing logging writers")
		for _, w := range writeClosers {
			err := w.Close()
			if err != nil {
				log.Error().Err(err).Msg("unable to close writer")
			}
		}
	}, nil
}

func setupSentryLogging(conf SentryConfiguration) (io.WriteCloser, error) {
	sentryWriter, err := sentryzerolog.New(sentryzerolog.Config{
		ClientOptions: sentry.ClientOptions{
			Dsn:         conf.SentryDSN,
			Environment: conf.SentryEnvironment,
		},
	})
	if err != nil {
		return nil, err
	}

	return sentryWriter, nil
}
