/*
 *   Redsemble - Redis packet assembler
 *   Copyright (c) 2016 Shannon Wynter, Ladbrokes Digital Australia Pty Ltd.
 *
 *   This program is free software: you can redistribute it and/or modify
 *   it under the terms of the GNU General Public License as published by
 *   the Free Software Foundation, either version 3 of the License, or
 *   (at your option) any later version.
 *
 *   This program is distributed in the hope that it will be useful,
 *   but WITHOUT ANY WARRANTY; without even the implied warranty of
 *   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *   GNU General Public License for more details.
 *
 *   You should have received a copy of the GNU General Public License
 *   along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 *   Author: Shannon Wynter <http://fremnet.net/contact>
 */

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"

	"github.com/google/gopacket"
	"github.com/google/gopacket/tcpassembly"
	"github.com/tidwall/redcon"
)

type redisStreamFactory struct{}

type redisStream struct {
	net, transport gopacket.Flow
	readerStream   readerStream
}

func (r *redisStreamFactory) New(net, transport gopacket.Flow) tcpassembly.Stream {

	rstream := &redisStream{
		net:          net,
		transport:    transport,
		readerStream: newReaderStream(),
	}

	go rstream.run()

	return &rstream.readerStream
}

func (r *redisStream) run() {
	buf := bufio.NewReader(&r.readerStream)

	for {
		reader := redcon.NewReader(buf)
		cmd, err := reader.ReadCommand()
		if err == io.EOF {
			return
		} else if err != nil {
			if *flogDebug {
				log.Println("Error reading stream", r.net, r.transport, ":", err)
			}
			continue
		}
		if redisCommand(bytes.ToLower(cmd.Args[0])).Valid() {
			printing := []string{}
			if *flogTimestamps {
				printing = append(printing, r.readerStream.InitialTS.Format("2006/01/02 15:04:05"))
			}
			printing = append(printing, string(cmd.Args[0]))
			if len(cmd.Args) > 1 {
				printing = append(printing, string(cmd.Args[1]))
				if len(cmd.Args) > 2 {
					printing = append(printing, strconv.Itoa(len(bytes.Join(cmd.Args[2:], []byte(" ")))))
				}
			}
			fmt.Println(strings.Join(printing, *fdelimiter))
		}
	}
}
