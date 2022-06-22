package listener

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

type WatcherEvent struct {
	SkrClusterID string `json:"skrClusterID"`
	Body         []byte `json:"body"`
}

type SKREventsListener struct {
	Addr           string
	Logger         logr.Logger
	receivedEvents chan event.GenericEvent
}

const paramContractVersion = "contractVersion"

func (l *SKREventsListener) ReceivedEvents() chan event.GenericEvent {
	if l.receivedEvents == nil {
		l.receivedEvents = make(chan event.GenericEvent)
	}
	return l.receivedEvents
}

func (l *SKREventsListener) Start(ctx context.Context) error {
	//routing
	mainRouter := mux.NewRouter()
	apiRouter := mainRouter.PathPrefix("/").Subrouter()

	apiRouter.HandleFunc(
		fmt.Sprintf("/v{%s}/skr/events/create", paramContractVersion),
		l.handleCreateEvent()).
		Methods(http.MethodPost)

	apiRouter.HandleFunc(
		fmt.Sprintf("/v{%s}/skr/events/update", paramContractVersion),
		l.handleUpdateEvent()).
		Methods(http.MethodPost)

	apiRouter.HandleFunc(
		fmt.Sprintf("/v{%s}/skr/events/delete", paramContractVersion),
		l.handleDeleteEvent()).
		Methods(http.MethodPost)

	apiRouter.HandleFunc(
		fmt.Sprintf("/v{%s}/skr/events/generic", paramContractVersion),
		l.handleGenericEvent()).
		Methods(http.MethodPost)

	//start web server
	server := &http.Server{Addr: l.Addr, Handler: mainRouter}
	go func() {
		l.Logger.Info("SKR events listener is starting up...")
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			l.Logger.Error(err, "Webserver startup failed")
		}
		l.Logger.WithValues("Address:", server.Addr).
			Info("SKR events listener started up successfully")
	}()
	<-ctx.Done()
	l.Logger.Info("SKR events listener is shutting down: context got closed")
	return server.Shutdown(ctx)
}

func (l *SKREventsListener) handleCreateEvent() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l.Logger.Info("CreateEvent")

		//unmarshal received event
		skrClusterID, skrEventObject := l.unmarshalEvent(w, r)

		//add event to the channel
		genericEvtObject := &unstructured.Unstructured{}
		genericEvtObject.SetName(skrEventObject.GetName())
		genericEvtObject.SetClusterName(skrClusterID)
		l.receivedEvents <- event.GenericEvent{Object: genericEvtObject}
		w.WriteHeader(http.StatusOK)
	}
}

func (l *SKREventsListener) handleUpdateEvent() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l.Logger.Info("UpdateEvent")

		//unmarshal received event
		skrClusterID, skrEventObject := l.unmarshalEvent(w, r)

		//add event to the channel
		genericEvtObject := &unstructured.Unstructured{}
		genericEvtObject.SetName(skrEventObject.GetName())
		genericEvtObject.SetClusterName(skrClusterID)
		l.receivedEvents <- event.GenericEvent{Object: genericEvtObject}
		w.WriteHeader(http.StatusOK)
	}
}

func (l *SKREventsListener) handleDeleteEvent() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l.Logger.Info("DeleteEvent")
		w.WriteHeader(http.StatusOK)
	}
}

func (l *SKREventsListener) handleGenericEvent() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l.Logger.Info("GenericEvent")
		w.WriteHeader(http.StatusOK)
	}
}

func (l *SKREventsListener) unmarshalEvent(w http.ResponseWriter, r *http.Request) (string, unstructured.Unstructured) {
	params := mux.Vars(r)
	contractVersion, ok := params[paramContractVersion]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("contract version could not be parsed"))
		return "", unstructured.Unstructured{}
	}

	if contractVersion == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("contract version cannot be empty"))
		return "", unstructured.Unstructured{}
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("could not read request body"))
		return "", unstructured.Unstructured{}
	}

	watcherEvent := &WatcherEvent{}
	err = json.Unmarshal(body, watcherEvent)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("could not unmarshal watcher event"))
		return "", unstructured.Unstructured{}
	}

	skrEventObject := unstructured.Unstructured{}
	if err = json.Unmarshal(watcherEvent.Body, &skrEventObject); err != nil {
		l.Logger.Error(err, "error transforming new component object")
		return "", unstructured.Unstructured{}
	}

	return watcherEvent.SkrClusterID, skrEventObject
}