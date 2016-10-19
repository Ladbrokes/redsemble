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
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/tcpassembly"
)

var (
	flogDebug      *bool
	flogTimestamps *bool
	fdelimiter     *string
)

func mergeSources(pm []*gopacket.PacketSource) chan gopacket.Packet {
	inputs := len(pm)
	heads := make([]gopacket.Packet, inputs)

	for i, v := range pm {
		packet, err := v.NextPacket()
		if err != nil {
			heads[i] = nil
			continue
		}
		heads[i] = packet
	}

	ch := make(chan gopacket.Packet, 1000)

	go func() {
		defer close(ch)
		for {
			var oldest gopacket.Packet
			var oldestIndex = 0

			nils := 0
			for i, n := range heads {
				if n == nil {
					nils++
					continue
				}
				if ts := n.Metadata().Timestamp; oldest == nil || ts.Before(oldest.Metadata().Timestamp) {
					oldestIndex = i
					oldest = n
				}
			}

			if nils == inputs {
				ch <- nil
				return
			}

			ch <- oldest

			packet, err := pm[oldestIndex].NextPacket()
			if err != nil {
				heads[oldestIndex] = nil
				continue
			}
			heads[oldestIndex] = packet
		}
	}()

	return ch
}

func main() {
	flogAllPackets := flag.Bool("l", false, "log all packets")
	flogTimestamps = flag.Bool("t", false, "output time stamps")
	flogDebug = flag.Bool("d", false, "debug redis packets")
	fdelimiter = flag.String("s", "|", "output delimiter")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] *.pcap[ *.pcap[ ... ]]\n\n", filepath.Base(os.Args[0]))
		flag.PrintDefaults()
		os.Exit(0)
	}

	flag.Parse()

	files := flag.Args()

	if len(files) == 0 {
		flag.Usage()
	}

	packetSources := make([]*gopacket.PacketSource, len(files))

	for i, filename := range files {
		handle, err := pcap.OpenOffline(filename)
		if err != nil {
			log.Fatal(err)
		}

		packetSources[i] = gopacket.NewPacketSource(handle, handle.LinkType())
	}

	streamFactory := &redisStreamFactory{}
	streamPool := tcpassembly.NewStreamPool(streamFactory)
	assembler := tcpassembly.NewAssembler(streamPool)

	packets := mergeSources(packetSources)
	for {
		select {
		case packet := <-packets:
			if packet == nil {
				assembler.FlushAll()
				return
			}
			if packet.NetworkLayer() == nil || packet.TransportLayer() == nil || packet.TransportLayer().LayerType() != layers.LayerTypeTCP {
				if *flogDebug {
					log.Println("Unusable packet")
				}
				continue
			}
			if *flogAllPackets {
				log.Println(packet)
			}
			tcp := packet.TransportLayer().(*layers.TCP)
			assembler.AssembleWithTimestamp(packet.NetworkLayer().NetworkFlow(), tcp, packet.Metadata().Timestamp)
		}
	}
}
