package main

import (
	"errors"
	"io"
	"math"
	"strings"
	"testing"
)

func TestBooleans(t *testing.T) {
	if val, _ := readRespValue(strings.NewReader("#f\r\n")); val != false {
		t.Errorf("Expected false, but got '%v'", val)
	}

	if val, _ := readRespValue(strings.NewReader("#t\r\n")); val != true {
		t.Errorf("Expected true, but got '%v'", val)
	}

	if _, err := readRespValue(strings.NewReader("#")); !errors.Is(err, io.EOF) {
		t.Errorf("Expected EOF error but got '%s'", err)
	}
}

func TestNull(t *testing.T) {
	if val, _ := readRespValue(strings.NewReader("_\r\n")); val != nil {
		t.Errorf("Expected nil, but got '%v'", val)
	}

	if _, err := readRespValue(strings.NewReader("_\r")); !strings.Contains(err.Error(), "EOF") {
		t.Errorf("Expected EOF error but got '%s'", err)
	}
}

func TestSimpleStrings(t *testing.T) {
	if val, _ := readRespValue(strings.NewReader("+Hello, World!\r\n")); val != "Hello, World!" {
		t.Errorf("Expected 'Hello, World!', but got '%s'", val)
	}

	if _, err := readRespValue(strings.NewReader("+123")); !errors.Is(err, io.EOF) {
		t.Errorf("Expected EOF error but got '%s'", err)
	}
}

func TestReadInteger(t *testing.T) {
	if val, _ := readRespValue(strings.NewReader(":10304\r\n")); val != int64(10304) {
		t.Errorf("Expected '10304', but got '%v'", val)
	}

	if _, err := readRespValue(strings.NewReader(":abc\r\n")); !strings.Contains(err.Error(), "Unexpected") {
		t.Errorf("Expected EOF error but got '%s'", err)
	}

	if _, err := readRespValue(strings.NewReader(":103")); !errors.Is(err, io.EOF) {
		t.Errorf("Expected EOF error but got '%s'", err)
	}
}

func TestReadError(t *testing.T) {
	if val, _ := readRespValue(strings.NewReader("-Something Happened\r\n")); val.(error).Error() != "Something Happened" {
		t.Errorf("Expected 'Something Happened', but got '%v'", val.(error).Error())
	}

	if val, _ := readRespValue(strings.NewReader("-\r\n")); val.(error).Error() != "" {
		t.Errorf("Expected '' but got '%v'", val.(error).Error())
	}

	if _, err := readRespValue(strings.NewReader("-abc")); !errors.Is(err, io.EOF) {
		t.Errorf("Expected EOF error but got '%s'", err)
	}
}

func TestReadBulkString(t *testing.T) {
	if val, _ := readRespValue(strings.NewReader("$5\r\na\nb\nc\r\n")); val != "a\nb\nc" {
		t.Errorf("Expected 'a\nb\nc', but got '%v'", val)
	}

	if val, _ := readRespValue(strings.NewReader("$-1\r\n")); val != nil {
		t.Errorf("Expected nil, but got '%v'", val)
	}

	if _, err := readRespValue(strings.NewReader("$10")); !errors.Is(err, io.EOF) {
		t.Errorf("Expected EOF error but got '%s'", err)
	}

	if _, err := readRespValue(strings.NewReader("$10\r\na")); !strings.Contains(err.Error(), "unexpected") {
		t.Errorf("Expected EOF error but got '%s'", err)
	}
}

func TestReadBulkError(t *testing.T) {
	if val, _ := readRespValue(strings.NewReader("!5\r\na\nb\nc\r\n")); val.(error).Error() != "a\nb\nc" {
		t.Errorf("Expected 'a\nb\nc', but got '%v'", val)
	}
}

func TestReadArray(t *testing.T) {
	if val, _ := readRespValue(strings.NewReader("*3\r\n+abc\r\n:123\r\n*1\r\n:4567890\r\n")); val.([]interface{})[0].(string) != "abc" || val.([]interface{})[1].(int64) != 123 || val.([]interface{})[2].([]interface{})[0].(int64) != 4567890 {
		t.Errorf("Expected '[\"abc\", 123, [4567890]]', but got '%v'", val)
	}

	if val, _ := readRespValue(strings.NewReader("*-1")); len(val.([]interface{})) != 0 {
		t.Errorf("Expected nil, but got '%v'", val)
	}
}

func TestReadDouble(t *testing.T) {
	if val, _ := readRespValue(strings.NewReader(",4\r\n")); val != float64(4.0) {
		t.Errorf("Expected '4.0', but got '%v'", val)
	}

	if val, _ := readRespValue(strings.NewReader(",123456.789e3\r\n")); val != float64(123456789.0) {
		t.Errorf("Expected '123456789.0', but got '%v'", val)
	}

	if val, _ := readRespValue(strings.NewReader(",+123.789e0\r\n")); val != float64(123.789) {
		t.Errorf("Expected '123.789', but got '%v'", val)
	}

	if val, _ := readRespValue(strings.NewReader(",-123.00\r\n")); val != float64(-123.00) {
		t.Errorf("Expected '-123.00', but got '%v'", val)
	}

	if val, _ := readRespValue(strings.NewReader(",-123.00e-2\r\n")); val != float64(-1.23) {
		t.Errorf("Expected '-1.23', but got '%v'", val)
	}

	if val, _ := readRespValue(strings.NewReader(",inf\r\n")); val != math.Inf(1) {
		t.Errorf("Expected 'inf', but got '%v'", val)
	}

	if val, _ := readRespValue(strings.NewReader(",-inf\r\n")); val != math.Inf(-1) {
		t.Errorf("Expected '-inf', but got '%v'", val)
	}

	if val, _ := readRespValue(strings.NewReader(",nan\r\n")); !math.IsNaN(val.(float64)) {
		t.Errorf("Expected 'nan', but got '%v'", val)
	}
}

func TestReadVerbatimString(t *testing.T) {
	if val, _ := readRespValue(strings.NewReader("=3\r\nENC:abc\r\n")); val.(VerbatimString).encoding != "ENC" && string(val.(VerbatimString).data) != "abc" {
		t.Errorf("Expected {'ENC', 'abc'}, but got '%v'", val)
	}
}

func TestReadMap(t *testing.T) {
	val, _ := readRespValue(strings.NewReader("%2\r\n+first\r\n:1\r\n+second\r\n:2\r\n"))

	mapVal := val.(map[interface{}]interface{})
	if mapVal["first"] != int64(1) || mapVal["second"] != int64(2) {
		t.Errorf("Expected {'first': 1, 'second': 2}, but got %v", mapVal)
	}
}
