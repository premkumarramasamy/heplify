package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/negbie/heplify/config"
	"github.com/negbie/heplify/sniffer"
	"github.com/negbie/logp"
	//_ "github.com/mkevac/debugcharts"
)

const version = "heplify 1.46"

func init() {

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Use %s like: %s [option]\n", version, os.Args[0])
		flag.PrintDefaults()
	}

	var (
		err         error
		ifaceConfig config.InterfacesConfig
		logging     logp.Logging
		fileRotator logp.FileRotator
		dbg         string
		std         bool
		sys         bool
	)

	flag.StringVar(&ifaceConfig.Device, "i", "any", "Listen on interface")
	flag.StringVar(&ifaceConfig.Type, "t", "pcap", "Capture types are [pcap, af_packet]")
	flag.StringVar(&ifaceConfig.ReadFile, "rf", "", "Read pcap file")
	flag.StringVar(&ifaceConfig.WriteFile, "wf", "", "Path to write pcap file")
	flag.IntVar(&ifaceConfig.RotationTime, "rt", 60, "Pcap rotation time in minutes")
	flag.BoolVar(&config.Cfg.Zip, "zf", false, "Enable pcap compression")
	flag.IntVar(&ifaceConfig.Loop, "lp", 1, "Loop count over ReadFile. Use 0 to loop forever")
	flag.BoolVar(&ifaceConfig.ReadSpeed, "rs", false, "Use packet timestamps with maximum pcap read speed")
	flag.IntVar(&ifaceConfig.Snaplen, "s", 8192, "Snaplength")
	flag.StringVar(&ifaceConfig.PortRange, "pr", "5060-5090", "Portrange to capture SIP")
	flag.BoolVar(&ifaceConfig.WithVlan, "vlan", false, "vlan")
	flag.BoolVar(&ifaceConfig.WithErspan, "erspan", false, "erspan")
	flag.IntVar(&ifaceConfig.BufferSizeMb, "b", 32, "Interface buffersize (MB)")
	flag.StringVar(&dbg, "d", "", "Enable certain debug selectors [fragment,layer,payload,rtp,rtcp,sdp]")
	flag.BoolVar(&std, "e", false, "Log to stderr and disable syslog/file output")
	flag.BoolVar(&sys, "sl", false, "Log to syslog")
	flag.StringVar(&logging.Level, "l", "info", "Log level [debug, info, warning, error]")
	flag.BoolVar(&ifaceConfig.OneAtATime, "o", false, "Read packet for packet")
	flag.StringVar(&fileRotator.Path, "p", "./", "Log filepath")
	flag.StringVar(&fileRotator.Name, "n", "heplify.log", "Log filename")
	flag.BoolVar(&config.Cfg.Bench, "bm", false, "Benchmark for the next 2 minutes and exit")
	flag.StringVar(&config.Cfg.Mode, "m", "SIPRTCP", "Capture modes [SIP, SIPDNS, SIPLOG, SIPRTP, SIPRTCP]")
	flag.BoolVar(&config.Cfg.Dedup, "dd", false, "Deduplicate packets")
	flag.StringVar(&config.Cfg.Discard, "di", "", "Discard uninteresting packets by any string")
	flag.StringVar(&config.Cfg.DiscardMethod, "dim", "", "Discard uninteresting SIP packets by CSeq [OPTIONS,NOTIFY]")
	flag.StringVar(&config.Cfg.Filter, "fi", "", "Filter interesting packets by any string")
	flag.StringVar(&config.Cfg.HepServer, "hs", "127.0.0.1:9060", "HEP server address")
	flag.StringVar(&config.Cfg.HepNodePW, "hp", "myhep", "HEP node PW")
	flag.UintVar(&config.Cfg.HepNodeID, "hi", 2002, "HEP node ID")
	flag.StringVar(&config.Cfg.Network, "nt", "udp", "Network types are [udp, tcp, tls]")
	flag.BoolVar(&config.Cfg.Protobuf, "protobuf", false, "Use Protobuf on wire")
	flag.BoolVar(&config.Cfg.Reassembly, "tcpassembly", false, "If true, tcpassembly will be enabled")
	flag.Parse()

	config.Cfg.Iface = &ifaceConfig
	logp.ToStderr = &std
	logging.ToSyslog = &sys
	logp.DebugSelectorsStr = &dbg
	logging.Files = &fileRotator
	config.Cfg.Logging = &logging

	if config.Cfg.HepNodeID > 0xFFFFFFFE {
		config.Cfg.HepNodeID = 0xFFFFFFFE
	}
	config.Cfg.Discard, err = strconv.Unquote(`"` + config.Cfg.Discard + `"`)
	checkErr(err)
	config.Cfg.Filter, err = strconv.Unquote(`"` + config.Cfg.Filter + `"`)
	checkErr(err)
}

func checkErr(err error) {
	if err != nil {
		fmt.Printf("\nError: %v\n\n", err)
	}
}

func checkCritErr(err error) {
	if err != nil {
		fmt.Printf("\nCritical: %v\n\n", err)
		os.Exit(1)
	}
}

func main() {

	/* 	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}() */

	err := logp.Init("heplify", config.Cfg.Logging)
	checkCritErr(err)

	capture, err := sniffer.New(config.Cfg.Mode, config.Cfg.Iface)
	checkCritErr(err)
	defer capture.Close()

	err = capture.Run()
	checkCritErr(err)
}
