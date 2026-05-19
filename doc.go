/*
Package ethernet_ip implements an EtherNet/IP client library for communicating with
Allen-Bradley PLCs and compatible devices using the EtherNet/IP protocol.

# Overview

EtherNet/IP is an industrial communication protocol developed by Rockwell Automation
for use in manufacturing and automation environments. This library enables Go
applications to connect to and exchange data with PLCs that support the EtherNet/IP
protocol, such as ControlLogix, CompactLogix, and SLC 500 series controllers.

This library provides:

  - TCP connection management with session registration
  - Tag-based read/write operations with automatic type handling
  - Support for all standard CIP data types (BOOL, SINT, INT, DINT, LINT, REAL, etc.)
  - Connection pooling for high-performance scenarios
  - Compatibility with cpppo servers via Logix Class 2 object access
  - Thread-safe operations with mutex protection
  - Symbolic address parsing for complex tag paths

# Installation

Install the package using:

	go get github.com/anviod/ethernet-ip

# Basic Usage

## Creating a Connection

Create a TCP connection to a PLC at the specified IP address:

	conn, err := ethernet_ip.NewTCP("192.168.1.10", nil)
	if err != nil {
	    log.Fatal(err)
	}
	defer conn.Close()

	if err := conn.Connect(); err != nil {
	    log.Fatal(err)
	}

The second parameter is an optional Config struct. Pass nil to use default settings.

## Reading Tags

The library supports several ways to read tags:

Method 1: Read all tags and access individual tags

	tags, err := conn.AllTags()
	if err != nil {
	    log.Fatal(err)
	}

	tag := tags["MyTag"]
	if err := tag.Read(); err != nil {
	    log.Fatal(err)
	}

	value := tag.Int32()
	log.Printf("Tag value: %d", value)

Method 2: Initialize a specific tag by path

	tag := new(ethernet_ip.Tag)
	if err := conn.InitializeTag("Program:MainProgram.MyTag", tag); err != nil {
	    log.Fatal(err)
	}

	if err := tag.Read(); err != nil {
	    log.Fatal(err)
	}

## Writing Tags

To write a value to a tag, use a setter method followed by Write:

	tag.SetInt32(42)
	if err := tag.Write(); err != nil {
	    log.Fatal(err)
	}

Available setter methods:
  - SetBool(value bool)
  - SetInt8(value int8)
  - SetUInt8(value uint8)
  - SetInt16(value int16)
  - SetUInt16(value uint16)
  - SetInt32(value int32)
  - SetUInt32(value uint32)
  - SetInt64(value int64)
  - SetUInt64(value uint64)
  - SetFloat32(value float32)
  - SetFloat64(value float64)
  - SetString(value string)

## Reading Values

After a successful Read, use the appropriate type conversion method:

	tag.Read()

	// For numeric types
	intVal := tag.Int32()
	uintVal := tag.UInt16()
	floatVal := tag.Float32()

	// For strings
	strVal := tag.String()

	// For booleans
	boolVal := tag.Bool()

## Tag Paths

Tags can be referenced using full symbolic paths:

	// Simple tag
	tag := new(Tag)
	conn.InitializeTag("MyTag", tag)

	// Tag in program scope
	conn.InitializeTag("Program:MainProgram.MyTag", tag)

	// Array element
	conn.InitializeTag("MyArray[0]", tag)

	// Multi-dimensional array
	conn.InitializeTag("MyArray[1,0,2]", tag)

	// Tag member (UDT)
	conn.InitializeTag("MyUDT.MemberName", tag)

	// Nested UDT members
	conn.InitializeTag("ParentUDT.ChildUDT.Member", tag)

## Using Tag Groups

Tag groups allow multiple tags to be read or written simultaneously:

	lock := new(sync.Mutex)
	group := ethernet_ip.NewTagGroup(lock)

	tag1 := tags["Tag1"]
	tag2 := tags["Tag2"]
	group.Add(tag1)
	group.Add(tag2)

	// Read multiple tags
	if err := group.Read(); err != nil {
	    log.Fatal(err)
	}

	// Write multiple tags
	tag1.SetInt32(100)
	tag2.SetString("hello")
	if err := group.Write(); err != nil {
	    log.Fatal(err)
	}

## Connection Pooling

For high-performance scenarios, use a connection pool to manage multiple connections:

	pool, err := ethernet_ip.NewTCPPool("192.168.1.10", nil, 10)
	if err != nil {
	    log.Fatal(err)
	}
	defer pool.Close()

	// Get a connection from the pool
	conn, err := pool.Get()
	if err != nil {
	    log.Fatal(err)
	}

	// Use the connection
	tags, err := conn.AllTags()
	// ...

	// Return the connection to the pool
	pool.Put(conn)

## cpppo Server Compatibility

The library supports accessing tags via Logix Class 2 object attributes,
which is used by cpppo servers. This allows communication with software
simulators that emulate PLC behavior.

	// Read a Class 2 attribute (attribute ID 1 corresponds to BoolTag)
	data, err := conn.ReadClass2Attribute(1)
	if err != nil {
	    log.Fatal(err)
	}

## Discovering Devices

List identity information of devices on the network:

	identities, err := conn.ListIdentity()
	if err != nil {
	    log.Fatal(err)
	}
	for _, identity := range identities {
	    log.Printf("Device: %s, Type: %d, Vendor: %d",
	        identity.ProductName, identity.DeviceType, identity.VendorID)
	}

## Forward Open

For time-critical communications, use Forward Open to establish a dedicated
connection path:

	if err := conn.ForwardOpen(); err != nil {
	    log.Fatal(err)
	}
	defer conn.ForwardClose()

## Configuration

The Config struct allows customization of connection parameters:

	config := &ethernet_ip.Config{
	    TCPPort:     44818,      // Default EtherNet/IP port
	    UDPPort:     44818,      // Default UDP port
	    Slot:        0,           // Controller slot number
	    TimeTick:    3,           // Time tick in milliseconds
	    TimeTickOut: 250,         // Connection timeout
	}

	conn, err := ethernet_ip.NewTCP("192.168.1.10", config)

# Data Types

The library supports all standard CIP data types:

	CIP Type  | Go Type   | Size (bytes) | Description
	----------|-----------|--------------|-------------------------------------
	BOOL      | bool      | 1            | Boolean
	SINT      | int8      | 1            | Signed 8-bit integer
	INT       | int16     | 2            | Signed 16-bit integer
	DINT      | int32     | 4            | Signed 32-bit integer
	LINT      | int64     | 8            | Signed 64-bit integer
	USINT     | uint8     | 1            | Unsigned 8-bit integer
	UINT      | uint16    | 2            | Unsigned 16-bit integer
	UDINT     | uint32    | 4            | Unsigned 32-bit integer
	ULINT     | uint64    | 8            | Unsigned 64-bit integer
	REAL      | float32   | 4            | Single-precision floating point
	LREAL     | float64   | 8            | Double-precision floating point
	STRING    | string    | variable     | CIP string (up to 88 bytes)
	STRING2   | string    | variable     | CIP string (extended format)

# Error Handling

All methods return errors that should be handled appropriately:

	if err := tag.Read(); err != nil {
	    switch {
	    case errors.Is(err, ethernet_ip.ErrBufferTooShort):
	        log.Println("Buffer too short for tag data")
	    case errors.Is(err, ethernet_ip.ErrTagNotFound):
	        log.Println("Tag does not exist on device")
	    default:
	        log.Printf("Read failed: %v", err)
	    }
	    return
	}

# Thread Safety

The library is designed to be thread-safe:

  - Each Tag has its own mutex (Tag.Lock) protecting read/write operations
  - EIPTCP has a request lock protecting concurrent requests
  - Connection pools use mutexes to protect internal state

You can safely use multiple tags concurrently:

	var wg sync.WaitGroup
	for _, t := range tags {
	    wg.Add(1)
	    go func(tag *ethernet_ip.Tag) {
	        defer wg.Done()
	        tag.Read()
	    }(t)
	}
	wg.Wait()

# Performance Considerations

For high-performance scenarios:

 1. Use connection pooling (EIPTCPPool) to reduce connection overhead
 2. Use Tag groups for batch read/write operations
 3. For continuous monitoring, implement caching rather than polling
 4. Consider Forward Open for time-critical applications

# Common Issues

1. Connection Refused: Ensure the PLC is reachable and the EtherNet/IP port is not blocked by firewall.

2. Tag Not Found: Verify the tag path is correct. Tag names are case-sensitive.

3. Permission Denied: Some PLCs require appropriate access levels to read/write tags.

4. Session Expired: If a session expires, reconnect using Connect().

# Examples

See the examples directory for complete working examples:

  - Basic read/write operations
  - Tag group operations
  - Connection pooling
  - cpppo server communication

# References

For more information about the EtherNet/IP protocol, see:
  - https://www.rockwellautomation.com/en-us/technologies/industrial-protocols/ethernet-ip.html
  - https://www.odva.org/ethernet-ip

For cpppo server implementation:
  - https://github.com/pjkundert/cpppo
*/
package ethernet_ip
