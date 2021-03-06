package controllers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/splitio/go-split-commons/v3/dtos"
	"github.com/splitio/go-toolkit/v4/logging"
	"github.com/splitio/split-synchronizer/v4/conf"
	"github.com/splitio/split-synchronizer/v4/log"
	"github.com/splitio/split-synchronizer/v4/splitio/proxy/interfaces"
)

func TestImpressionsBufferCounter(t *testing.T) {
	var p = impressionPoolBufferSizeStruct{size: 0}

	p.Addition(1)
	p.Addition(2)
	if !p.GreaterThan(2) || p.GreaterThan(4) {
		t.Error("Error on Addition method")
	}

	p.Reset()
	if !p.GreaterThan(-1) || p.GreaterThan(1) {
		t.Error("Error on Reset")
	}

}

func TestAddImpressions(t *testing.T) {
	conf.Initialize()
	if log.Instance == nil {
		stdoutWriter := ioutil.Discard //os.Stdout
		log.Initialize(stdoutWriter, stdoutWriter, stdoutWriter, stdoutWriter, stdoutWriter, logging.LevelNone)
	}
	interfaces.Initialize()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		sdkVersion := r.Header.Get("SplitSDKVersion")
		sdkMachine := r.Header.Get("SplitSDKMachineIP")
		sdkMachineName := r.Header.Get("SplitSDKMachineName")

		if sdkVersion != "test-1.0.0" {
			t.Error("SDK Version HEADER not match")
		}

		if sdkMachine != "127.0.0.1" {
			t.Error("SDK Machine HEADER not match")
		}

		if sdkMachineName != "SOME_MACHINE_NAME" {
			t.Error("SDK Machine Name HEADER not match", sdkMachineName)
		}

		impressionsMode := r.Header.Get("SplitSDKImpressionsMode")
		if impressionsMode != "debug" {
			t.Error("SplitSDKImpressionsMode HEADER not match", impressionsMode)
		}

		rBody, _ := ioutil.ReadAll(r.Body)

		var impressionsInPost []dtos.ImpressionsDTO
		err := json.Unmarshal(rBody, &impressionsInPost)
		if err != nil {
			t.Error(err)
			return
		}

		if impressionsInPost[0].TestName != "some_test" ||
			impressionsInPost[0].KeyImpressions[0].KeyName != "some_key_1" ||
			impressionsInPost[0].KeyImpressions[1].KeyName != "some_key_2" {
			t.Error("Posted impressions arrived mal-formed")
		}

		fmt.Fprintln(w, "ok!!")
	}))
	defer ts.Close()

	imp1 := dtos.ImpressionDTO{KeyName: "some_key_1", Treatment: "on", Time: 1234567890, ChangeNumber: 9876543210, Label: "some_label_1", BucketingKey: "some_bucket_key_1"}
	imp2 := dtos.ImpressionDTO{KeyName: "some_key_2", Treatment: "off", Time: 1234567890, ChangeNumber: 9876543210, Label: "some_label_2", BucketingKey: "some_bucket_key_2"}

	keyImpressions := make([]dtos.ImpressionDTO, 0)
	keyImpressions = append(keyImpressions, imp1, imp2)
	impressionsTest := dtos.ImpressionsDTO{TestName: "some_test", KeyImpressions: keyImpressions}

	impressions := make([]dtos.ImpressionsDTO, 0)
	impressions = append(impressions, impressionsTest)

	data, err := json.Marshal(impressions)
	if err != nil {
		t.Error(err)
		return
	}

	os.Setenv("SPLITIO_SDK_URL", ts.URL)
	os.Setenv("SPLITIO_EVENTS_URL", ts.URL)

	// Init Impressions controller.
	wg := &sync.WaitGroup{}
	InitializeImpressionWorkers(200, 2, wg)
	AddImpressions(data, "test-1.0.0", "127.0.0.1", "SOME_MACHINE_NAME", "DEBUG")

	// Lets async function post impressions
	time.Sleep(time.Duration(4) * time.Second)
}

func TestAddImpressionsOptimized(t *testing.T) {
	conf.Initialize()
	if log.Instance == nil {
		stdoutWriter := ioutil.Discard //os.Stdout
		log.Initialize(stdoutWriter, stdoutWriter, stdoutWriter, stdoutWriter, stdoutWriter, logging.LevelNone)
	}
	interfaces.Initialize()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		sdkVersion := r.Header.Get("SplitSDKVersion")
		sdkMachine := r.Header.Get("SplitSDKMachineIP")
		sdkMachineName := r.Header.Get("SplitSDKMachineName")

		if sdkVersion != "test-1.0.0" {
			t.Error("SDK Version HEADER not match")
		}

		if sdkMachine != "127.0.0.1" {
			t.Error("SDK Machine HEADER not match")
		}

		if sdkMachineName != "SOME_MACHINE_NAME" {
			t.Error("SDK Machine Name HEADER not match", sdkMachineName)
		}

		impressionsMode := r.Header.Get("SplitSDKImpressionsMode")
		if impressionsMode != "optimized" {
			t.Error("SplitSDKImpressionsMode HEADER not match", impressionsMode)
		}

		rBody, _ := ioutil.ReadAll(r.Body)

		var impressionsInPost []dtos.ImpressionsDTO
		err := json.Unmarshal(rBody, &impressionsInPost)
		if err != nil {
			t.Error(err)
			return
		}

		if impressionsInPost[0].TestName != "some_test" ||
			impressionsInPost[0].KeyImpressions[0].KeyName != "some_key_1" ||
			impressionsInPost[0].KeyImpressions[1].KeyName != "some_key_2" {
			t.Error("Posted impressions arrived mal-formed")
		}

		fmt.Fprintln(w, "ok!!")
	}))
	defer ts.Close()

	imp1 := dtos.ImpressionDTO{KeyName: "some_key_1", Treatment: "on", Time: 1234567890, ChangeNumber: 9876543210, Label: "some_label_1", BucketingKey: "some_bucket_key_1"}
	imp2 := dtos.ImpressionDTO{KeyName: "some_key_2", Treatment: "off", Time: 1234567890, ChangeNumber: 9876543210, Label: "some_label_2", BucketingKey: "some_bucket_key_2"}

	keyImpressions := make([]dtos.ImpressionDTO, 0)
	keyImpressions = append(keyImpressions, imp1, imp2)
	impressionsTest := dtos.ImpressionsDTO{TestName: "some_test", KeyImpressions: keyImpressions}

	impressions := make([]dtos.ImpressionsDTO, 0)
	impressions = append(impressions, impressionsTest)

	data, err := json.Marshal(impressions)
	if err != nil {
		t.Error(err)
		return
	}

	os.Setenv("SPLITIO_SDK_URL", ts.URL)
	os.Setenv("SPLITIO_EVENTS_URL", ts.URL)

	// Init Impressions controller.
	wg := &sync.WaitGroup{}
	InitializeImpressionWorkers(200, 2, wg)
	AddImpressions(data, "test-1.0.0", "127.0.0.1", "SOME_MACHINE_NAME", "OPTIMIZED")

	// Lets async function post impressions
	time.Sleep(time.Duration(4) * time.Second)
}

func TestAddImpressionsWithoutMode(t *testing.T) {
	conf.Initialize()
	if log.Instance == nil {
		stdoutWriter := ioutil.Discard //os.Stdout
		log.Initialize(stdoutWriter, stdoutWriter, stdoutWriter, stdoutWriter, stdoutWriter, logging.LevelNone)
	}
	interfaces.Initialize()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		sdkVersion := r.Header.Get("SplitSDKVersion")
		sdkMachine := r.Header.Get("SplitSDKMachineIP")
		sdkMachineName := r.Header.Get("SplitSDKMachineName")

		if sdkVersion != "test-1.0.0" {
			t.Error("SDK Version HEADER not match")
		}

		if sdkMachine != "127.0.0.1" {
			t.Error("SDK Machine HEADER not match")
		}

		if sdkMachineName != "SOME_MACHINE_NAME" {
			t.Error("SDK Machine Name HEADER not match", sdkMachineName)
		}

		impressionsMode := r.Header.Get("SplitSDKImpressionsMode")
		if impressionsMode != "" {
			t.Error("SplitSDKImpressionsMode HEADER not match", impressionsMode)
		}

		rBody, _ := ioutil.ReadAll(r.Body)

		var impressionsInPost []dtos.ImpressionsDTO
		err := json.Unmarshal(rBody, &impressionsInPost)
		if err != nil {
			t.Error(err)
			return
		}

		if impressionsInPost[0].TestName != "some_test" ||
			impressionsInPost[0].KeyImpressions[0].KeyName != "some_key_1" ||
			impressionsInPost[0].KeyImpressions[1].KeyName != "some_key_2" {
			t.Error("Posted impressions arrived mal-formed")
		}

		fmt.Fprintln(w, "ok!!")
	}))
	defer ts.Close()

	imp1 := dtos.ImpressionDTO{KeyName: "some_key_1", Treatment: "on", Time: 1234567890, ChangeNumber: 9876543210, Label: "some_label_1", BucketingKey: "some_bucket_key_1"}
	imp2 := dtos.ImpressionDTO{KeyName: "some_key_2", Treatment: "off", Time: 1234567890, ChangeNumber: 9876543210, Label: "some_label_2", BucketingKey: "some_bucket_key_2"}

	keyImpressions := make([]dtos.ImpressionDTO, 0)
	keyImpressions = append(keyImpressions, imp1, imp2)
	impressionsTest := dtos.ImpressionsDTO{TestName: "some_test", KeyImpressions: keyImpressions}

	impressions := make([]dtos.ImpressionsDTO, 0)
	impressions = append(impressions, impressionsTest)

	data, err := json.Marshal(impressions)
	if err != nil {
		t.Error(err)
		return
	}

	os.Setenv("SPLITIO_SDK_URL", ts.URL)
	os.Setenv("SPLITIO_EVENTS_URL", ts.URL)

	// Init Impressions controller.
	wg := &sync.WaitGroup{}
	InitializeImpressionWorkers(200, 2, wg)
	AddImpressions(data, "test-1.0.0", "127.0.0.1", "SOME_MACHINE_NAME", "")

	// Lets async function post impressions
	time.Sleep(time.Duration(4) * time.Second)
}
