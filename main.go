
package main

import (
    "encoding/binary"
    "fmt"
    "log"
    "net"
    "time"
)

//0                   1                   2                   3
//0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
//+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//|LI | VN  |Mode |    Stratum     |     Poll      |  Precision   |
//+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//|                         Root Delay                            |
//+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//|                         Root Dispersion                       |
//+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//|                          Reference ID                         |
//+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//|                                                               |
//+                     Reference Timestamp (64)                  +
//|                                                               |
//+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//|                                                               |
//+                      Origin Timestamp (64)                    +
//|                                                               |
//+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//|                                                               |
//+                      Receive Timestamp (64)                   +
//|                                                               |
//+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//|                                                               |
//+                      Transmit Timestamp (64)                  +
//|                                                               |
//+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+

type packet struct {
    Settings          uint8 // LI, VN, Mode
    Stratum           uint8
    Poll              uint8
    Precision         uint8
    RootDelay         uint32
    RootDispersion    uint32
    ReferenceID       uint32
    ReferenceTimeSec  uint32
    ReferenceTimeFrac uint32
    OriginTimeSec     uint32
    OriginTimeFrac    uint32
    ReceiveTimeSec    uint32
    ReceiveTimeFrac   uint32
    TransmitTimeSec   uint32
    TransmitTimeFrac  uint32
}

func main() {
    host := "pool.ntp.org:123"
    conn, err := net.Dial("udp", host)
    if err != nil {
       log.Fatal("failed to connect: ", err)
    }
    defer conn.Close()

    if err := conn.SetReadDeadline(time.Now().Add(15 * time.Second)); err != nil {
       log.Fatal("failed to set deadline: ", err)
    }
    req := &packet{Settings: 0x1B} // 00011011 (00 -> no leap, 011 -> version 3, 011 -> client mode)

    if err := binary.Write(conn, binary.BigEndian, req); err != nil {
       log.Fatal("failed to send request: ", err)
    }

    rsp := &packet{}

    if err := binary.Read(conn, binary.BigEndian, rsp); err != nil {
       log.Fatal("failed to read server response: ", err)
    }

    const ntpEpochOffset = 2208988800 // (70 yrs of seconds, (1970 - 1900))

    secs := float64(rsp.ReferenceTimeSec) - ntpEpochOffset
    nanos := (int64(rsp.ReferenceTimeFrac) * 1e9) >> 32

    fmt.Printf("Reference time: %v\n", time.Unix(int64(secs), nanos))

    secs = float64(rsp.OriginTimeSec) - ntpEpochOffset
    nanos = (int64(rsp.OriginTimeFrac) * 1e9) >> 32

    fmt.Printf("Origin time: %v\n", time.Unix(int64(secs), nanos))

    secs = float64(rsp.ReceiveTimeSec) - ntpEpochOffset
    nanos = (int64(rsp.ReceiveTimeFrac) * 1e9) >> 32

    fmt.Printf("Receive time: %v\n", time.Unix(int64(secs), nanos))

    secs = float64(rsp.TransmitTimeSec) - ntpEpochOffset
    nanos = (int64(rsp.TransmitTimeFrac) * 1e9) >> 32

    fmt.Printf("Transmit time: %v\n", time.Unix(int64(secs), nanos))
}