package ethernet_ip

import (
	"bytes"
	"encoding/binary"
	"math"
	"os"
	"sync"
	"testing"
)

func dialEmulator(t *testing.T) *EIPTCP {
	addr := os.Getenv("EIP_EMULATOR_ADDR")
	if addr == "" {
		addr = "127.0.0.1"
	}
	conn, err := NewTCP(addr, nil)
	if err != nil {
		t.Fatalf("NewTCP failed: %v", err)
	}
	if err := conn.Connect(); err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	return conn
}

func TestDefaultTags_Read(t *testing.T) {
	conn := dialEmulator(t)
	defer conn.Close()

	// BOOL
	tag := new(Tag)
	if err := conn.InitializeTag("Program:MainProgram.BoolTag", tag); err != nil {
		t.Fatalf("InitializeTag BoolTag failed: %v", err)
	}
	if !tag.Bool() {
		t.Fatalf("BoolTag expected true, got false")
	}

	// SINT
	tag = new(Tag)
	if err := conn.InitializeTag("Program:MainProgram.SintTag", tag); err != nil {
		t.Fatalf("InitializeTag SintTag failed: %v", err)
	}
	if tag.Int8() != 127 {
		t.Fatalf("SintTag expected 127, got %d", tag.Int8())
	}

	// INT
	tag = new(Tag)
	if err := conn.InitializeTag("Program:MainProgram.IntTag", tag); err != nil {
		t.Fatalf("InitializeTag IntTag failed: %v", err)
	}
	if tag.Int16() != 100 {
		t.Fatalf("IntTag expected 100, got %d", tag.Int16())
	}

	// DINT
	tag = new(Tag)
	if err := conn.InitializeTag("Program:MainProgram.DintTag", tag); err != nil {
		t.Fatalf("InitializeTag DintTag failed: %v", err)
	}
	if tag.Int32() != 1000 {
		t.Fatalf("DintTag expected 1000, got %d", tag.Int32())
	}

	// REAL
	tag = new(Tag)
	if err := conn.InitializeTag("Program:MainProgram.RealTag", tag); err != nil {
		t.Fatalf("InitializeTag RealTag failed: %v", err)
	}
	if math.Abs(float64(tag.Float32())-2.71828) > 1e-4 {
		t.Fatalf("RealTag expected 2.71828, got %v", tag.Float32())
	}

	// STRING
	tag = new(Tag)
	if err := conn.InitializeTag("Program:MainProgram.StringTag", tag); err != nil {
		t.Fatalf("InitializeTag StringTag failed: %v", err)
	}
	if tag.String() != "Main Program" {
		t.Fatalf("StringTag expected 'Main Program', got '%s'", tag.String())
	}

	// Global tags
	tag = new(Tag)
	if err := conn.InitializeTag("Global.DintTag", tag); err != nil {
		t.Fatalf("InitializeTag Global.DintTag failed: %v", err)
	}
	if tag.Int32() != 2147483647 {
		t.Fatalf("Global.DintTag expected 2147483647, got %d", tag.Int32())
	}
}

func TestString_WriteAndRestore(t *testing.T) {
	conn := dialEmulator(t)
	defer conn.Close()

	tag := new(Tag)
	if err := conn.InitializeTag("Program:MainProgram.StringTag", tag); err != nil {
		t.Fatalf("InitializeTag StringTag failed: %v", err)
	}
	orig := tag.XString()

	newVal := "Hello from integration test"
	tag.SetString(newVal)
	if err := tag.Write(); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	// re-read
	if err := tag.Read(); err != nil {
		t.Fatalf("read after write failed: %v", err)
	}
	if tag.String() != newVal {
		t.Fatalf("after write expected '%s', got '%s'", newVal, tag.String())
	}

	// restore
	tag.SetString(orig)
	if err := tag.Write(); err != nil {
		t.Fatalf("restore write failed: %v", err)
	}
}

func TestTagGroup_MultiRead(t *testing.T) {
	conn := dialEmulator(t)
	defer conn.Close()

	lock := new(sync.Mutex)
	tg := NewTagGroup(lock)

	t1 := new(Tag)
	if err := conn.InitializeTag("Program:MainProgram.IntTag", t1); err != nil {
		t.Fatalf("InitializeTag IntTag failed: %v", err)
	}
	t2 := new(Tag)
	if err := conn.InitializeTag("Program:MainProgram.DintTag", t2); err != nil {
		t.Fatalf("InitializeTag DintTag failed: %v", err)
	}
	tg.Add(t1)
	tg.Add(t2)

	if err := tg.Read(); err != nil {
		t.Fatalf("TagGroup.Read failed: %v", err)
	}
	if t1.Int16() != 100 {
		t.Fatalf("IntTag expected 100, got %d", t1.Int16())
	}
	if t2.Int32() != 1000 {
		t.Fatalf("DintTag expected 1000, got %d", t2.Int32())
	}
}

func TestInt_WriteAndRestore(t *testing.T) {
	conn := dialEmulator(t)
	defer conn.Close()

	tag := new(Tag)
	if err := conn.InitializeTag("Program:MainProgram.IntTag", tag); err != nil {
		t.Fatalf("InitializeTag IntTag failed: %v", err)
	}
	orig := tag.Int16()

	newVal := int16(12345)
	tag.SetInt32(int32(newVal))
	if err := tag.Write(); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	// re-read
	if err := tag.Read(); err != nil {
		t.Fatalf("read after write failed: %v", err)
	}
	if tag.Int16() != newVal {
		t.Fatalf("after write expected %d, got %d", newVal, tag.Int16())
	}

	// restore
	tag.SetInt32(int32(orig))
	if err := tag.Write(); err != nil {
		t.Fatalf("restore write failed: %v", err)
	}
}

func TestReal_WriteAndRestore(t *testing.T) {
	conn := dialEmulator(t)
	defer conn.Close()

	tag := new(Tag)
	if err := conn.InitializeTag("Program:MainProgram.RealTag", tag); err != nil {
		t.Fatalf("InitializeTag RealTag failed: %v", err)
	}
	orig := tag.Float32()

	newVal := float32(3.14159)
	// Use raw bytes to write
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, newVal)
	tag.wValue = buf.Bytes()
	if err := tag.Write(); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	// re-read
	if err := tag.Read(); err != nil {
		t.Fatalf("read after write failed: %v", err)
	}
	if math.Abs(float64(tag.Float32())-float64(newVal)) > 1e-4 {
		t.Fatalf("after write expected %v, got %v", newVal, tag.Float32())
	}

	// restore
	buf.Reset()
	binary.Write(buf, binary.LittleEndian, orig)
	tag.wValue = buf.Bytes()
	if err := tag.Write(); err != nil {
		t.Fatalf("restore write failed: %v", err)
	}
}
