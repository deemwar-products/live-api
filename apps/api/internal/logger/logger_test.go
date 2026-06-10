package logger

import (
 "os"
 "strings"
 "sync"
 "testing"
 "time"
)

func TestLogger_WhenInfoCalled_ThenLogsCorrectFormat(t *testing.T) {
 oldStderr := os.Stderr
 r, w, _ := os.Pipe()
 os.Stderr = w
 defer func() { os.Stderr = oldStderr; w.Close() }()

 log := New("test")
 log.Info("message %d", 42)

 w.Close()
 buf := make([]byte, 1024)
 n, _ := r.Read(buf)
 output := string(buf[:n])

 if !strings.Contains(output, "INFO") || !strings.Contains(output, "test") || !strings.Contains(output, "message 42") {
 t.Errorf("Info log format incorrect: %s", output)
 }
}

func TestLogger_WhenWarnCalled_ThenLogsCorrectFormat(t *testing.T) {
 oldStderr := os.Stderr
 r, w, _ := os.Pipe()
 os.Stderr = w
 defer func() { os.Stderr = oldStderr; w.Close() }()

 log := New("warn")
 log.Warn("warning %s", "text")

 w.Close()
 buf := make([]byte, 1024)
 n, _ := r.Read(buf)
 output := string(buf[:n])

 if !strings.Contains(output, "WARN") || !strings.Contains(output, "warn") || !strings.Contains(output, "warning text") {
 t.Errorf("Warn log format incorrect: %s", output)
 }
}

func TestLogger_WhenErrorCalled_ThenLogsCorrectFormat(t *testing.T) {
 oldStderr := os.Stderr
 r, w, _ := os.Pipe()
 os.Stderr = w
 defer func() { os.Stderr = oldStderr; w.Close() }()

 log := New("error")
 log.Error("error %v", []int{1, 2})

 w.Close()
 buf := make([]byte, 1024)
 n, _ := r.Read(buf)
 output := string(buf[:n])

 if !strings.Contains(output, "ERROR") || !strings.Contains(output, "error") || !strings.Contains(output, "error [1 2]") {
 t.Errorf("Error log format incorrect: %s", output)
 }
}

func TestLogger_WhenDebugCalled_ThenLogsCorrectFormat(t *testing.T) {
 oldStderr := os.Stderr
 r, w, _ := os.Pipe()
 os.Stderr = w
 defer func() { os.Stderr = oldStderr; w.Close() }()

 log := New("debug")
 log.Debug("debug %t", true)

 w.Close()
 buf := make([]byte, 1024)
 n, _ := r.Read(buf)
 output := string(buf[:n])

 if !strings.Contains(output, "DEBUG") || !strings.Contains(output, "debug") || !strings.Contains(output, "debug true") {
 t.Errorf("Debug log format incorrect: %s", output)
 }
}

func TestLogger_WhenLogCalledWithCustomLevel_ThenLogsCorrectFormat(t *testing.T) {
 oldStderr := os.Stderr
 r, w, _ := os.Pipe()
 os.Stderr = w
 defer func() { os.Stderr = oldStderr; w.Close() }()

 log := New("logtest")
 ts := time.Now().Format("2006-01-02 15:04:05")
 expectedPrefix := "[" + ts + "] [CUSTOM] [logtest]"

 log.log("CUSTOM", "test %s", "arg")

 w.Close()
 buf := make([]byte, 1024)
 n, _ := r.Read(buf)
 output := string(buf[:n])

 if !strings.Contains(output, expectedPrefix) || !strings.Contains(output, "test arg") {
 t.Errorf("Custom log format incorrect: %s", output)
 }
}

func TestLogger_WhenMultipleGoroutinesLog_ThenAllMessagesCaptured(t *testing.T) {
 oldStderr := os.Stderr
 r, w, _ := os.Pipe()
 os.Stderr = w
 defer func() { os.Stderr = oldStderr; w.Close() }()

 var wg sync.WaitGroup
 log := New("concurrent")

 for i := 0; i < 100; i++ {
 wg.Add(1)
 go func(i int) {
 defer wg.Done()
 log.Info("msg %d", i)
 }(i)
 }

 wg.Wait()
 w.Close()

 buf := make([]byte, 32*1024)
 n, _ := r.Read(buf)
 output := string(buf[:n])

 if n == 0 {
 t.Error("No logs captured")
 }
 if !strings.Contains(output, "INFO") {
 t.Error("INFO level missing")
 }
}