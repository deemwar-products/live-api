package main

import "testing"

func TestMain_WhenLogAccessed_ThenNotNil(t *testing.T) {
 if log == nil {
 t.Error("log variable is nil")
 }
}