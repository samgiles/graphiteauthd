/**
 * The MIT License (MIT)
 *
 * Copyright (c) 2015 Samuel Giles
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in
 * all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
 * THE SOFTWARE.
 */
package main

import (
	"flag"
	"errors"
	"net"
	"crypto/tls"
	"fmt"
	"os"
	"io"
	"github.com/samgiles/bytetrie"
)

var numberOfBackendConnections = flag.Int("beconnections", 1, "Number of concurrent connections to open to the Graphite backend")
var listenAddressString = flag.String("listen", ":9090", "Address and port to bind listener to")
var remoteAddressString = flag.String("remote", "", "Address and port of remote graphite instance")
var acceptKeyString     = flag.String("apikey", "MYAPIKEY", "API key to accept")

var certificateFile = flag.String("cert", "", "Path to .pem certificate")
var keyFile         = flag.String("key", "", "Path to .pem key file")

var searchRoot = new(bytetrie.Node)

func main() {
	flag.Parse()

	if *remoteAddressString == "" {
		warn("Flag: 'remote' is required")
		flag.Usage()
		os.Exit(1)
	}

	if *acceptKeyString == "" {
		warn("Flag: 'apiKey' is required")
		flag.Usage()
		os.Exit(1)
	}


	searchRoot.Insert([]byte(*acceptKeyString))
	multiplexers := initialiseMultiplexers(*numberOfBackendConnections, *remoteAddressString)


	if *certificateFile == "" || *keyFile == "" {
		warn("No certificate file or key file found, not using TLS")
		listener, err := net.Listen("tcp", *listenAddressString)
		check(err)
		startListening(listener, multiplexers)
	} else {
		cert, err := tls.LoadX509KeyPair(*certificateFile, *keyFile)
		check(err)
		tlsConfig := tls.Config{Certificates: []tls.Certificate{cert}}
		listener, err := tls.Listen("tcp", *listenAddressString, &tlsConfig)
		check(err)
		startListening(listener, multiplexers)
	}

}


func initialiseMultiplexers(count int, remoteAddress string) []multiplexer {
	multiplexers := make([]multiplexer, count)

	fmt.Printf("Starting %d backend connection(s)\n", count)
	for i := 0; i < count; i++ {
		remoteConnection, err := net.Dial("tcp", remoteAddress)
		check(err)

		multiplexers[i] = multiplexer {
			remoteConnection: remoteConnection,
			erred:              false,
			closesig:           make(chan bool),
			id:                 i,
		}
	}

	return multiplexers
}

func startListening(listener net.Listener, multiplexers []multiplexer) {
	var nextMultiplexer = 0
	var connectionId = 0
	fmt.Printf("Starting listener\n");
	for {
		connection, err := listener.Accept()
		if err != nil {
			warn("Failed to accept connection '%s'\n", err)
			continue
		}

		multiplexer := multiplexers[nextMultiplexer]
		p := &proxy {
			multiplexer:      multiplexer,
			localConnection:  connection,
			erred:            false,
			closesig:         make(chan bool),
			prefix:           connectionId,
		}

		connectionId++

		// Round robin select the next connection to multiplex this new
		// incoming connection onto
		nextMultiplexer++

		if nextMultiplexer == len(multiplexers) {
			nextMultiplexer = 0
		}

		go p.start()
	}
}

// A multiplexer represents a single TCP connection used to multiplex many
// connections
type multiplexer struct {
	remoteConnection  net.Conn
	erred              bool
	closesig           chan bool
	id                 int
}

func (m *multiplexer) Write(bytes []byte) (int, error) {
	return m.remoteConnection.Write(bytes)
}

type proxy struct {
	multiplexer multiplexer
	localConnection  net.Conn
	erred            bool
	closesig         chan bool
	prefix           int
}

func (p *proxy) start() {
	defer p.localConnection.Close()

	// We only care about one way communication, Graphite never replies
	go p.pipe()

	//wait for close...
	<-p.closesig
}

func (p *proxy) pipe() {
	// 64k buffer
	buffer := make([]byte, 0xffff)
	src := p.localConnection
	dest := p.multiplexer
	var offset = 0

	for {
		numberOfBytes, err := src.Read(buffer[offset:])

		if err != nil {
			if err == io.EOF {
				p.closesig <-true
			} else {
				p.err("Read failed '%s'\n", err)
			}

			return
		}

		receivedBytes := buffer[:numberOfBytes]

		metrics, remaining, err := ParseBuffer(receivedBytes, searchRoot)
		if err != nil {
			p.err("Connection dropped terminated '%s'\n", err)
			return
		}

		offset = len(remaining)
		copy(buffer[:offset], remaining)

		//write out result
		_, err = dest.Write(metrics)
		if err != nil {
			p.err("Write failed '%s'\n", err)
			return
		}
	}
}

func (p *proxy) err(s string, err error) {
	if p.erred {
		return
	}

	warn("Connection Error[%d]: " + s, p.prefix, err)

	p.closesig <- true
	p.erred = true
}

func check(err error) {
	if err != nil {
		warn(err.Error())
		os.Exit(1)
	}
}

func warn(f string, args ...interface{}) {
	fmt.Printf(f + "\n", args...)
}


// Split the buffer by '\n' (0x0A) characters, return an byte[][] of
// indicating each metric, and byte[] of the remaining parts of the buffer
func ParseBuffer(buffer []byte, apiKeys *bytetrie.Node) ([]byte, []byte, error) {
	var metricBufferCapacity uint = uint(len(buffer))
	metricBuffer := make([]byte, metricBufferCapacity)

	var metricSize uint =  0
	var metricBufferUsage uint = 0
	var totalMetrics int = 0
	var currentSearchNode = apiKeys
	var byteAccepted = false

	var rootNamespaceAccepted = false

	for _, b := range buffer {

		if !(b == '.' || rootNamespaceAccepted) {
			currentSearchNode, byteAccepted = currentSearchNode.Accepts(b)
			if !byteAccepted {
				return nil, nil,  errors.New(fmt.Sprintf("Invalid API key: %s*", metricBuffer[metricBufferUsage - metricSize:metricBufferUsage]))
			}
		} else {
			if !currentSearchNode.IsLeaf {
				return nil, nil, errors.New(fmt.Sprintf("Invalid API key: %s", metricBuffer[metricBufferUsage - metricSize:metricBufferUsage]))
			} else {
				rootNamespaceAccepted = true
			}
		}

		metricBuffer[metricBufferUsage] = b
		metricSize++
		metricBufferUsage++

		if b == '\n' {
			totalMetrics++
			metricSize = 0;
			currentSearchNode = apiKeys;
			rootNamespaceAccepted = false
		}
	}

	return metricBuffer[:metricBufferUsage - metricSize], metricBuffer[(metricBufferUsage - metricSize):metricBufferUsage], nil
}
