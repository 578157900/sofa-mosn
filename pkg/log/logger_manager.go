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

package log

import (
	"context"
	"errors"
	"sync"

	"github.com/alipay/sofa-mosn/pkg/types"
)

var (
	DefaultLogger ErrorLogger
	StartLogger   ErrorLogger

	ErrNoLoggerFound = errors.New("no logger found in logger manager")
)

var errorLoggerManagerInstance *ErrorLoggerManager

func init() {
	errorLoggerManagerInstance = &ErrorLoggerManager{
		mutex:    sync.Mutex{},
		managers: make(map[string]ErrorLogger),
	}
	// use console as start logger
	StartLogger, _ = GetOrCreateDefaultErrorLogger("", INFO)
	// default as start before Init
	DefaultLogger = StartLogger
}

// ErrorLoggerManager manages error log can be updated dynamicly
type ErrorLoggerManager struct {
	mutex    sync.Mutex
	managers map[string]ErrorLogger
}

// GetOrCreateErrorLogger returns a ErrorLogger based on the output(p).
// If Logger not exists, and create function is not nil, creates a new logger
func (mng *ErrorLoggerManager) GetOrCreateErrorLogger(p string, level Level, f CreateErrorLoggerFunc) (ErrorLogger, error) {
	mng.mutex.Lock()
	defer mng.mutex.Unlock()
	if lg, ok := mng.managers[p]; ok {
		return lg, nil
	}
	// only find exists
	if f == nil {
		return nil, ErrNoLoggerFound
	}
	lg, err := f(p, level)
	if err != nil {
		return nil, err
	}
	mng.managers[p] = lg
	return lg, nil
}

// Default Export Functions
func GetErrorLoggerManagerInstance() *ErrorLoggerManager {
	return errorLoggerManagerInstance
}

// GetOrCreateDefaultErrorLogger used default create function
func GetOrCreateDefaultErrorLogger(p string, level Level) (ErrorLogger, error) {
	return errorLoggerManagerInstance.GetOrCreateErrorLogger(p, level, CreateDefaultErrorLogger)
}

func InitDefaultLogger(output string, level Level) (err error) {
	DefaultLogger, err = GetOrCreateDefaultErrorLogger(output, level)
	return
}

func ByContext(ctx context.Context) ErrorLogger {
	if ctx != nil {
		if lg := ctx.Value(types.ContextKeyLogger); lg != nil {
			return lg.(ErrorLogger)
		}
	}
	// if context is nil, use default Logger instead
	return DefaultLogger
}

// UpdateErrorLoggerLevel updates the exists ErrorLogger's Level
func UpdateErrorLoggerLevel(p string, level Level) bool {
	// we use a nil create function means just get exists logger
	if lg, _ := errorLoggerManagerInstance.GetOrCreateErrorLogger(p, 0, nil); lg != nil {
		lg.SetLogLevel(level)
		return true
	}
	return false
}

// ToggleLogger enable/disable the exists logger, include ErrorLogger and Logger
func ToggleLogger(p string, disable bool) bool {
	// find ErrorLogger
	if lg, _ := errorLoggerManagerInstance.GetOrCreateErrorLogger(p, 0, nil); lg != nil {
		lg.Toggle(disable)
		return true
	}
	// find Logger
	if lg, ok := loggers[p]; ok {
		lg.Toggle(disable)
		return true
	}
	return false
}

// Reopen all logger
func Reopen() error {
	for _, logger := range loggers {
		if err := logger.Reopen(); err != nil {
			return err
		}
	}
	return nil
}

// CloseAll logger
func CloseAll() error {
	for _, logger := range loggers {
		if err := logger.Close(); err != nil {
			return err
		}
	}
	return nil
}
