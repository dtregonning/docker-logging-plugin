package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/docker/docker/daemon/logger"
	"github.com/docker/docker/pkg/ioutils"
	"github.com/docker/go-plugins-helpers/sdk"
)

/*(##008##)*/
type StartLoggingRequest struct {
	File string
	Info logger.Info
}

/*(##009##)*/
type StopLoggingRequest struct {
	File string
}

/*(##010##)*/
type CapabilitiesResponse struct {
	Err string
	Cap logger.Capability
}

/*(##011##)*/
type ReadLogsRequest struct {
	Info   logger.Info
	Config logger.ReadConfig
}

/* ##007## (Setup Endpoint handlers for LogDriver)(Boiler Plate)
		   - the newDriver() has returned and supplied a driver
           - handlers will setup 4 "functions" to handle calls to the endpoints outlined (https://docs.docker.com/engine/extend/plugins_logging/#create-a-logging-plugin)
           - these functions get attached to the ServeMux
           - Note that all 4 functions have a accompanying struct
				- /LogDriver.StartLogging has StartLoggingRequest (##008##)
				- /LogDriver.StopLogging has StopLoggingRequest (##009##)
				- /LogDriver.Capabilities has CapabilitiesResponse (##010##)
				- /LogDriver.ReadLogs has ReadLogsRequest (##011##)
 */
func handlers(h *sdk.Handler, d *driver) {
		/* ##012## (Setup Endpoint handlers for /LogDriver.StartLogging)(Boiler Plate)
			- Mostly error checking for correct input
			- d.StartLogging(req.File, req.Info) is where it gets interesting, driver implementation of Start Logging
			- admin type function used for setting up logger
	 	*/
		h.HandleFunc("/LogDriver.StartLogging", func(w http.ResponseWriter, r *http.Request) {
			var req StartLoggingRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if req.Info.ContainerID == "" {
				respond(errors.New("must provide container id in log context"), w)
				return
			}

			err := d.StartLogging(req.File, req.Info)
			respond(err, w)
		})

		/* ##013## (Setup Endpoint handlers for /LogDriver.StopLogging)(Boiler Plate)
			- Mostly error checking for correct input.
			- d.StopLogging(req.File) is where it gets interesting, driver implementation of Start Logging.
			- admin type function used for stopping logger.
	 	*/
		h.HandleFunc("/LogDriver.StopLogging", func(w http.ResponseWriter, r *http.Request) {
			var req StopLoggingRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			err := d.StopLogging(req.File)
			respond(err, w)
		})

		/* ##014## (Setup Endpoint handlers for /LogDriver.Capabilities)(Custom)
				- Note this uses encoded instead of decoder.
				- responds with a ReadLogs: true value as required for LogDriver Implementation.
				- admin - required implementation
		 */
		h.HandleFunc("/LogDriver.Capabilities", func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(&CapabilitiesResponse{
				Cap: logger.Capability{ReadLogs: true},
			})
		})

		/* ##015## (Setup Endpoint handlers for /LogDriver.ReadLogs)(Boiler Plate)
			- This is the interesting section of http go file
			- defer stream.Close() means the stream will close off after the function ends example(https://tour.golang.org/flowcontrol/12)

		*/
		h.HandleFunc("/LogDriver.ReadLogs", func(w http.ResponseWriter, r *http.Request) {
			var req ReadLogsRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			stream, err := d.ReadLogs(req.Info, req.Config)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer stream.Close()

			w.Header().Set("Content-Type", "application/x-json-stream")
			wf := ioutils.NewWriteFlusher(w)
			io.Copy(wf, stream)
		})
	}

	type response struct {
		Err string
	}

	func respond(err error, w http.ResponseWriter) {
		var res response
		if err != nil {
			res.Err = err.Error()
		}
		json.NewEncoder(w).Encode(&res)
	}
