package main

import (
	"log"
	"os"

	"./libpcap"
)

func extract(filename string, filter string) {

	pcap, err := libpcap.PcapOpenOffline(filename)
	if err != nil {
		log.Panic(err)
	}

	if filter != "" {
		err := pcap.SetFilter(filter)
		if err != nil {
			log.Panic("failed to set filter: ", err)
		}
	}

	for {
		hdr, _, err := pcap.Next()
		if err != nil {
			log.Fatal(err)
		} else if hdr != nil {
			log.Println("got packet")
		} else {
			log.Println("end of file")
			return
		}
	}
}

func main() {
	extract(os.Args[1], os.Args[2])
}
