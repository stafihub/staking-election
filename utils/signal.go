// Copyright 2020 tpkeeper
// SPDX-License-Identifier: LGPL-3.0-only

package utils

import (
	"context"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
)

// shutdownRequestChannel is used to initiate shutdown from one of the
// subsystems using the same code paths as when an interrupt signal is received.
var ShutdownRequestChannel = make(chan struct{})

// interruptSignals defines the default signals to catch in order to do a proper
// shutdown.  This may be modified during init depending on the platform.
var interruptSignals = []os.Signal{os.Interrupt}

// shutdowntListener listens for OS Signals such as SIGINT (Ctrl+C) and shutdown
// requests from shutdownRequestChannel.  It returns a context that is canceled
// when either signal is received.
func ShutdownListener() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		interruptChannel := make(chan os.Signal, 1)
		signal.Notify(interruptChannel, interruptSignals...)
		// Listen for initial shutdown signal and close the returned
		// channel to notify the caller.
		select {
		case sig := <-interruptChannel:
			logrus.Infof("Received signal (%s).  Shutting down...",
				sig)

		case <-ShutdownRequestChannel:
			logrus.Infof("Shutdown requested.  Shutting down...")
		}
		cancel()
		// Listen for repeated signals and display a message so the user
		// knows the shutdown is in progress and the process is not
		// hung.
		for {
			select {
			case sig := <-interruptChannel:
				logrus.Infof("Received signal (%s).  Already "+
					"shutting down...", sig)

			case <-ShutdownRequestChannel:
				logrus.Info("Shutdown requested.  Already " +
					"shutting down...")
			}
		}
	}()

	return ctx
}
