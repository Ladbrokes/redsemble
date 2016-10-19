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
	"time"

	"github.com/google/gopacket/tcpassembly"
	"github.com/google/gopacket/tcpassembly/tcpreader"
)

type readerStream struct {
	tcpreader.ReaderStream
	InitialTS time.Time
}

func newReaderStream() readerStream {
	return readerStream{
		ReaderStream: tcpreader.NewReaderStream(),
	}
}

func (r *readerStream) Reassembled(reassembly []tcpassembly.Reassembly) {
	if r.InitialTS.IsZero() && len(reassembly) > 0 {
		r.InitialTS = reassembly[0].Seen
	}
	r.ReaderStream.Reassembled(reassembly)
}
