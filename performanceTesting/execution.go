package main

import (
	"context"
	"fmt"
	"github.com/c2h5oh/datasize"
	"github.com/echocat/goxr/common"
	"github.com/echocat/goxr/log"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

const (
	bufferSize = 10 * 1024
)

var (
	responseReaderPool         sync.Pool
	executionTimePerTest       = 10 * time.Second
	numberOfParallelExecutions = uint(10)
)

func createExecutor() executor {
	return executor{
		executionTime:              executionTimePerTest,
		numberOfParallelExecutions: uint16(numberOfParallelExecutions),
		client: &http.Client{
			Transport: &http.Transport{
				MaxConnsPerHost:     int(numberOfParallelExecutions),
				MaxIdleConnsPerHost: int(numberOfParallelExecutions),
				MaxIdleConns:        int(numberOfParallelExecutions),
				IdleConnTimeout:     time.Second * 1,
			},
			Timeout: 5 * time.Second,
		},
	}
}

type executor struct {
	executionTime              time.Duration
	numberOfParallelExecutions uint16
	client                     *http.Client
}

func (instance executor) String() string {
	return fmt.Sprintf("execute for %v with %d number of executors", instance.executionTime, instance.numberOfParallelExecutions)
}

type executionResult struct {
	Target string
	URL    string

	MaxDuration   time.Duration
	MinDuration   time.Duration
	TotalDuration time.Duration

	Executions uint64
	Failures   uint64
	Errors     uint64
}

func (instance executor) execute(te test, ta target) executionResult {
	//instance.warmUp(te, ta)
	result := executionResult{
		Target: ta.name(),
		URL:    ta.createUriFor(te),
	}
	wg := new(sync.WaitGroup)
	for i := uint16(0); i < instance.numberOfParallelExecutions; i++ {
		wg.Add(1)
		go func(result *executionResult) {
			instance.executeFor(te, ta, result)
			wg.Done()
		}(&result)
	}
	wg.Wait()
	closeIdleConnections(instance.client.Transport)
	return result
}

func (instance executor) warmUp(te test, ta target) {
	for i := 0; i < 10; i++ {
		wg := new(sync.WaitGroup)
		for j := uint16(0); j < instance.numberOfParallelExecutions; j++ {
			wg.Add(1)
			go func() {
				result := executionRunResult{}
				instance.executeRun(te, ta, false, &result)
				wg.Done()
			}()
		}
		wg.Wait()
	}
	closeIdleConnections(instance.client.Transport)
}

func (instance executor) executeFor(te test, ta target, result *executionResult) {
	done := uint32(0)
	time.AfterFunc(instance.executionTime, func() {
		atomic.StoreUint32(&done, 1)
		//cancelFunc()
	})

	debug := log.IsDebugEnabled()
	runResult := executionRunResult{}
	for {
		start := time.Now()
		if !instance.executeRun(te, ta, debug, &runResult) || atomic.LoadUint32(&done) != 0 {
			break
		}
		d := time.Now().Sub(start)
		atomic.AddUint64(&result.Executions, 1)
		if !runResult.Success {
			atomic.AddUint64(&result.Failures, 1)
		}
		if runResult.Error != nil {
			atomic.AddUint64(&result.Errors, 1)
		}
		atomic.AddInt64((*int64)(&result.TotalDuration), int64(d))
		for {
			max := atomic.LoadInt64((*int64)(&result.MaxDuration))
			if d < time.Duration(max) {
				break
			}
			if atomic.CompareAndSwapInt64((*int64)(&result.MaxDuration), max, int64(d)) {
				break
			}
		}
		for {
			min := atomic.LoadInt64((*int64)(&result.MinDuration))
			if d > time.Duration(min) {
				break
			}
			if atomic.CompareAndSwapInt64((*int64)(&result.MinDuration), min, int64(d)) {
				break
			}
		}
	}
}

type executionRunResult struct {
	Error      error
	StatusCode int
	Success    bool
}

func (instance *executionRunResult) reset() {
	instance.Error = nil
	instance.StatusCode = 0
	instance.Success = false
}

func (instance executor) executeRun(te test, ta target, debug bool, result *executionRunResult) (nextRunAllowed bool) {
	result.reset()
	request := instance.createRequest(te, ta)

	resp, err := instance.client.Do(request)
	err = common.UnderlyingError(err)
	if err != nil {
		if err == context.DeadlineExceeded || err == context.Canceled {
			return false
		} else if err == http.ErrAbortHandler ||
			err == http.ErrContentLength ||
			err == http.ErrLineTooLong ||
			err == http.ErrServerClosed ||
			err == http.ErrMissingBoundary ||
			err == io.ErrNoProgress ||
			err == io.ErrClosedPipe ||
			err == io.ErrUnexpectedEOF ||
			isTemporary(err) {
			result.Error = err
			if debug {
				log.
					WithError(err).
					WithField("url", request.URL).
					Debug("request failed")
			}
			return true
		} else {
			panic(fmt.Sprintf("unepxected error while %v for target %v (req: %v): %v", instance, ta, request.URL, err))
		}
	}
	defer close(resp.Body)
	result.StatusCode = resp.StatusCode

	if result.StatusCode == 200 {
		rr := acquireResponseReader()
		rr.expected = te.getSize()
		defer releaseResponseReader(rr)
		err = rr.readFrom(resp)
		if err == io.ErrNoProgress ||
			err == io.ErrClosedPipe ||
			err == io.ErrUnexpectedEOF {
			result.Error = err
		} else if err != nil {
			panic(fmt.Sprintf("unepxected error while %v for target %v (req: %v): %v", instance, ta, request.URL, err))
		}
	}

	result.Success = result.StatusCode == 200 && result.Error == nil
	return true
}

func (instance executor) createRequest(te test, ta target) *http.Request {
	req, err := http.NewRequest("GET", ta.createUriFor(te), nil)
	must(err)
	return req
}

type responseReader struct {
	expected  datasize.ByteSize
	read      datasize.ByteSize
	buf       []byte
	bufLength datasize.ByteSize
}

func acquireResponseReader() *responseReader {
	v := responseReaderPool.Get()
	if v == nil {
		return &responseReader{
			buf:       make([]byte, bufferSize),
			bufLength: bufferSize,
		}
	}
	return v.(*responseReader)
}

func releaseResponseReader(resp *responseReader) {
	responseReaderPool.Put(resp)
}

func (instance *responseReader) readFrom(resp *http.Response) error {
	for remain := datasize.ByteSize(resp.ContentLength); remain > 0; {
		offset := instance.bufLength
		if remain < instance.bufLength {
			offset = remain
		}
		b := instance.buf[:offset]
		nn, err := resp.Body.Read(b)
		if nn > 0 {
			instance.read += datasize.ByteSize(nn)
			remain -= datasize.ByteSize(nn)
		}
		if err == io.EOF {
			if remain <= 0 {
				err = nil
			} else {
				err = io.ErrUnexpectedEOF
			}
		} else if err != nil {
			return err
		}
	}

	return nil
}
