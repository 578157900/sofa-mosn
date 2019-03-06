/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package metrics

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"net"
	"time"

	gometrics "github.com/rcrowley/go-metrics"

	"github.com/alipay/sofa-mosn/pkg/log"
	"github.com/alipay/sofa-mosn/pkg/types"
	"syscall"
)

// TransferStats
type TransferStats struct {
	Type   string
	Labels map[string]string
	Data   []TransferData
}

// TransferData keeps information for single go-metrics objects, like gauge
type TransferData struct {
	MetricsType   string
	MetricsKey    string
	MetricsValues []int64
}

const (
	metricsCounter   = "counter"
	metricsGauge     = "gauge"
	metricsHistogram = "histogram"
)

func init() {
	gob.Register(new(TransferStats))
	gob.Register(new(TransferData))
}

// makesTransferData get all registered metrics data as a map[string]map[string][]TransferData
// the map will be gob encoded to transfer
func makesTransferData() ([]byte, error) {

	metrics := GetAll()

	transfers := make([]TransferStats, len(metrics))

	for i, metric := range metrics {
		transfers[i].Type = metric.Type()
		transfers[i].Labels = metric.Labels()

		metric.Each(func(key string, val interface{}) {
			data := TransferData{
				MetricsKey: key,
			}
			switch metric := val.(type) {
			case gometrics.Counter:
				data.MetricsType = metricsCounter
				data.MetricsValues = []int64{metric.Count()}
			case gometrics.Gauge:
				data.MetricsType = metricsGauge
				data.MetricsValues = []int64{metric.Value()}
			case gometrics.Histogram:
				h := metric.Snapshot()
				data.MetricsType = metricsHistogram
				data.MetricsValues = h.Sample().Values()
			default: //unsupport metrics, ignore
				return
			}
			transfers[i].Data = append(transfers[i].Data, data)
		})

	}

	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(transfers); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// readTransferData gets the gob encoded data, makes it as go-metrics data
func readTransferData(b []byte) error {
	var transfers []TransferStats
	buf := bytes.NewBuffer(b)
	if err := gob.NewDecoder(buf).Decode(&transfers); err != nil {
		return err
	}
	for _, transfer := range transfers {
		s, _ := NewMetrics(transfer.Type, transfer.Labels)

		for _, metric := range transfer.Data {
			switch metric.MetricsType {
			case metricsCounter:
				s.Counter(metric.MetricsKey).Inc(metric.MetricsValues[0])
			case metricsGauge:
				s.Gauge(metric.MetricsKey).Update(metric.MetricsValues[0])
			case metricsHistogram:
				h := s.Histogram(metric.MetricsKey)
				for _, v := range metric.MetricsValues {
					h.Update(v)
				}
			}
		}
	}
	return nil
}

// TransferServer starts a unix socket, lasts 10 seconds and 2*$gracefultime}, receive metrics datas
// When serves a conn, sends a message to chan
func TransferServer(gracefultime time.Duration, ch chan<- bool) {
	defer func() {
		if r := recover(); r != nil {
			log.DefaultLogger.Errorf("transfer metrics server panic %v", r)
		}
	}()
	syscall.Unlink(types.TransferStatsDomainSocket)
	ln, err := net.Listen("unix", types.TransferStatsDomainSocket)
	if err != nil {
		log.DefaultLogger.Errorf("transfer metrics net listen error %v", err)
		return
	}
	defer ln.Close()
	log.DefaultLogger.Infof("transfer metrics server start")
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.DefaultLogger.Errorf("transfer metrics server panic %v", r)
			}
		}()
		for {
			conn, err := ln.Accept()
			if err != nil {
				if ope, ok := err.(*net.OpError); ok && (ope.Op == "accept") {
					log.DefaultLogger.Infof("transfer metrics server listener closed")
				} else {
					log.DefaultLogger.Errorf("transfer metrics server accept error %v", err)
				}
				return
			}
			log.DefaultLogger.Infof("transfer metrics accept")
			go func() {
				serveConn(conn)
				if ch != nil {
					select {
					case ch <- true:
					case <-time.After(10 * time.Second): // write timeout
					}
				}
			}()
		}
	}()
	select {
	case <-time.After(2*gracefultime + types.DefaultConnReadTimeout + 10*time.Second):
		log.DefaultLogger.Infof("transfer metrics server exit")
	}
}

// TransferMetrics sends metrics data to unix socket
// If wait is true, will wait server response, with ${timeout}
// If wait is false, timeout is useless
func TransferMetrics(wait bool, timeout time.Duration) {
	body, err := makesTransferData()
	if err != nil {
		log.DefaultLogger.Errorf("transfer metrics get metrics data error: %v", err)
		return
	}
	transferMetrics(body, wait, timeout)
}

func transferMetrics(body []byte, wait bool, timeout time.Duration) {
	defer func() {
		if r := recover(); r != nil {
			log.DefaultLogger.Errorf("transfer metrics send data error: %v", r)
		}
	}()
	conn, err := net.Dial("unix", types.TransferStatsDomainSocket)
	if err != nil {
		log.DefaultLogger.Errorf("transfer metrics dial unix socket failed:%v", err)
		return
	}
	defer conn.Close()
	header := make([]byte, 4)
	binary.BigEndian.PutUint32(header, uint32(len(body)))
	if _, err := conn.Write(header); err != nil {
		log.DefaultLogger.Errorf("transfer metrics send header error: %v", err)
		return
	}
	if _, err := conn.Write(body); err != nil {
		log.DefaultLogger.Errorf("transfer metrics send body error: %v", err)
	}
	if wait {
		conn.SetReadDeadline(time.Now().Add(timeout))
		resp := make([]byte, 1)
		_, err := conn.Read(resp)
		if err != nil {
			log.DefaultLogger.Errorf("transfer metrics get response error: %v", err)
		}
		log.DefaultLogger.Infof("transfer metrics get reponse status: %v", resp[0])
	}
}

/**
*  transfer protocol
*  request:
*  	header: data length (4 bytes, uint32, bigendian)
*  	body: data (data length bytes)
*  response:
* 	header: status code (1 bytes, 0 means ok, 1 means failed)
**/
func read(conn net.Conn, size int) ([]byte, error) {
	if size == 0 {
		return nil, nil
	}
	b := make([]byte, size)
	var n, off int
	var err error
	for {
		n, err = conn.Read(b[off:])
		if err != nil {
			return nil, err
		}
		off += n
		if off == size {
			return b, nil
		}
	}
}
func readHeader(conn net.Conn) (int, error) {
	b, err := read(conn, 4)
	if err != nil {
		return 0, err
	}
	return int(binary.BigEndian.Uint32(b)), nil
}

func serveConn(conn net.Conn) {
	b := make([]byte, 1)
	if err := handler(conn); err != nil {
		b[0] = 0x01
	}
	conn.Write(b)
}

func handler(conn net.Conn) error {
	size, err := readHeader(conn)
	if err != nil {
		log.DefaultLogger.Errorf("transfer metrics read header error: %v", err)
		return err
	}
	body, err := read(conn, size)
	if err != nil {
		log.DefaultLogger.Errorf("transfer metrics read body error: %v", err)
		return err
	}
	if err := readTransferData(body); err != nil {
		log.DefaultLogger.Errorf("transfer metrics parse body error: %v", err)
		return err
	}
	return nil
}
