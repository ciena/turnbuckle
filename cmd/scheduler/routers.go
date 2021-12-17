package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	v1 "k8s.io/api/core/v1"
	schedulerapi "k8s.io/kube-scheduler/extender/v1"
)

func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, "Welcome to constraintpolicy-scheduler-extender!\n")
}

func decodeExtenderRequest(r *http.Request) (schedulerapi.ExtenderArgs, error) {
	var args schedulerapi.ExtenderArgs
	if r.Body == nil {
		return args, errors.New("request body empty")
	}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&args); err != nil {
		return args, fmt.Errorf("error %v decoding request", err)
	}
	if err := r.Body.Close(); err != nil {
		return args, err
	}
	return args, nil
}

func prescheduleChecks(w http.ResponseWriter, r *http.Request) (schedulerapi.ExtenderArgs, http.ResponseWriter, error) {
	var args schedulerapi.ExtenderArgs
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return schedulerapi.ExtenderArgs{}, w, errors.New("method type not POST")
	}
	if r.ContentLength > 1*1000*1000*1000 {
		w.WriteHeader(http.StatusInternalServerError)
		return schedulerapi.ExtenderArgs{}, w, errors.New("request size too large")
	}
	requestContentType := r.Header.Get("Content-Type")
	if requestContentType != "application/json" {
		w.WriteHeader(http.StatusNotFound)
		return schedulerapi.ExtenderArgs{}, w, errors.New("request content type not application/json")
	}
	args, err := decodeExtenderRequest(r)
	if err != nil {
		log.Printf("cannot decode request %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return schedulerapi.ExtenderArgs{}, w, err
	}
	return args, w, nil
}

func Filter(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var extenderFilterResult *schedulerapi.ExtenderFilterResult
	log.Printf("Got filter request. Doing preschedule checks")
	extenderArgs, w, err := prescheduleChecks(w, r)
	if err != nil {
		log.Printf("prescheduleChecks returned error %v", err)
		return
	}
	if nodes, err := constraintPolicyScheduler.FindBestNodes(extenderArgs.Pod, extenderArgs.Nodes.Items); err != nil {
		log.Printf("Find best node returned error: %v", err)
		w.WriteHeader(http.StatusNotFound)
		extenderFilterResult = &schedulerapi.ExtenderFilterResult{
			Error: err.Error(),
		}
	} else {
		w.WriteHeader(http.StatusOK)
		extenderFilterResult = &schedulerapi.ExtenderFilterResult{
			Nodes: &v1.NodeList{
				Items: nodes,
			},
		}
	}
	encoder := json.NewEncoder(w)
	if err := encoder.Encode(extenderFilterResult); err != nil {
		http.Error(w, "Encode error", http.StatusBadRequest)
	}
}

//func Prioritize(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {}

// func Bind(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {}
