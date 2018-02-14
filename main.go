package main

import (
	"fmt"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/docker/go-plugins-helpers/sdk"
)

const socketAddress = "/run/docker/plugins/splunklog.sock"

var logLevels = map[string]logrus.Level{
	"debug": logrus.DebugLevel,
	"info":  logrus.InfoLevel,
	"warn":  logrus.WarnLevel,
	"error": logrus.ErrorLevel,
}

func main() {

	/* ##001## (Set Log Level)(Boiler Plate)
	           - Main Entry Point
			   - Retrieves log level from os, if no log level is set then go with info
	           - Take log level and set it in logrus
	           - Bail out if not known log type
	 */
	levelVal := os.Getenv("LOG_LEVEL")
	if levelVal == "" {
		levelVal = "info"
	}
	if level, exists := logLevels[levelVal]; exists {
		logrus.SetLevel(level)
	} else {
		fmt.Fprintln(os.Stderr, "invalid log level: ", levelVal)
		os.Exit(1)
	}

	/* ##002## (Configure new Handler)(Boiler Plate)
		- NewHandler is part of the docker SDK(https://github.com/docker/go-plugins-helpers/blob/master/sdk/handler.go)
		- (Advanced) NewHandler creates a ServeMux (https://golang.org/pkg/net/http/#ServeMux)
		- This ServeMux is told to handle any calls to "/Plugin.Activate" (https://docs.docker.com/engine/extend/plugin_api/#handshake-api)
		- The LoggingDriver reference is interesting the above link only shows Docker support for authz, NetworkDriver, VolumeDriver.
	*/
	h := sdk.NewHandler(`{"Implements": ["LoggingDriver"]}`)

	/* ##003## (Build Driver and pass and configure required plugin handlers)(Boiler Plate)
		- handlers is part of http.go. Takes in 2 params.
		- 1st a reference to the newlly created handler(002)
		- 2nd a new driver which is yet to be created, this call will kick of the Driver being built.
	*/
	handlers(&h, newDriver())

	if err := h.ServeUnix(socketAddress, 0); err != nil {
		panic(err)
	}
}
