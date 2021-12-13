// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sentryexporter

import (
	"context"
	"time"

	"github.com/getsentry/sentry-go"
)

// transport is used by exporter to send events to Sentry
type transport interface {
	SendEvents(events []*sentry.Event)
	Configure(options sentry.ClientOptions)
	Flush(ctx context.Context) bool
}

type sentryTransport struct {
	httpTransport *sentry.HTTPTransport
}

// newSentryTransport returns a new pre-configured instance of sentryTransport.
func newSentryTransport() *sentryTransport {
	tr := sentryTransport{
		httpTransport: sentry.NewHTTPTransport(),
	}
	return &tr
}

func (t *sentryTransport) Configure(options sentry.ClientOptions) {
	t.httpTransport.Configure(options)
}

func (t *sentryTransport) Flush(ctx context.Context) bool {
	deadline, ok := ctx.Deadline()
	if ok {
		return t.httpTransport.Flush(time.Until(deadline))
	}
	return t.httpTransport.Flush(time.Second)
}

// SendEvents uses a Sentry HTTPTransport to send transaction events to Sentry
func (t *sentryTransport) SendEvents(transactions []*sentry.Event) {
	bufferCounter := 0
	for _, transaction := range transactions {
		// We should flush all events when we send transactions equal to the transport
		// buffer size so we don't drop transactions.
		if bufferCounter == t.httpTransport.BufferSize {
			t.httpTransport.Flush(time.Second)
			bufferCounter = 0
		}

		t.httpTransport.SendEvent(transaction)
		bufferCounter++
	}
}
