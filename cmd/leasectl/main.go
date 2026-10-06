// Command leasectl 读写租约状态文件。
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"example.com/lease-renew/lease"
)

func emit(v any) {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
}

func load(path string) *lease.State {
	st := &lease.State{}
	raw, err := os.ReadFile(path)
	if err != nil {
		return st
	}
	if err := json.Unmarshal(raw, st); err != nil {
		panic(err)
	}
	return st
}

func save(path string, st *lease.State) {
	raw, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		panic(err)
	}
}

func sampleRun(name string) {
	switch name {
	case "early":
		st := &lease.State{}
		lease.Grant(st, "L1", "h1", 0, 3000, &lease.Counter{})
		emit(lease.Renew(st, "L1", "r1", 1000, &lease.Counter{}))
	case "expired":
		st := &lease.State{}
		lease.Grant(st, "L1", "h1", 0, 3000, &lease.Counter{})
		emit(lease.Renew(st, "L1", "r1", 5000, &lease.Counter{}))
	case "replay":
		st := &lease.State{}
		lease.Grant(st, "L1", "h1", 0, 3000, &lease.Counter{})
		lease.Renew(st, "L1", "r1", 2500, &lease.Counter{})
		second := lease.Renew(st, "L1", "r1", 2500, &lease.Counter{})
		emit(map[string]any{"status": second.Status, "renewals": second.Renewals})
	case "limit":
		st := &lease.State{}
		lease.Grant(st, "L1", "h1", 0, 3000, &lease.Counter{})
		at := 2500
		last := lease.Result{}
		for i := 0; i < 4; i++ {
			last = lease.Renew(st, "L1", fmt.Sprintf("r%d", i), at, &lease.Counter{})
			at += 2500
		}
		emit(map[string]any{"status": last.Status, "renewals": last.Renewals})
	case "work":
		st := &lease.State{}
		for i := 0; i < 2000; i++ {
			lease.Grant(st, fmt.Sprintf("L%04d", i), "h", 0, 1000000, &lease.Counter{})
		}
		c := &lease.Counter{}
		for i := 0; i < 2000; i++ {
			lease.Renew(st, fmt.Sprintf("L%04d", i), fmt.Sprintf("r%d", i), 999000, c)
		}
		emit(map[string]int{"leases": 2000, "scanned": c.Scanned})
	default:
		fmt.Fprintln(os.Stderr, "未知场景："+name)
		os.Exit(2)
	}
}

func main() {
	sample := flag.String("sample", "", "early|expired|replay|limit|work")
	state := flag.String("state", "", "状态文件")
	id := flag.String("id", "", "租约 id")
	holder := flag.String("holder", "h1", "持有者")
	reqID := flag.String("req-id", "", "请求号")
	at := flag.Int("at", 0, "当前时刻（毫秒）")
	ttl := flag.Int("ttl", 3000, "租约时长（毫秒）")
	flag.Parse()

	if *sample != "" {
		sampleRun(*sample)
		return
	}
	if *state == "" || *id == "" || flag.NArg() == 0 {
		flag.Usage()
		os.Exit(2)
	}

	st := load(*state)
	c := &lease.Counter{}
	var out lease.Result
	switch flag.Arg(0) {
	case "grant":
		out = lease.Grant(st, *id, *holder, *at, *ttl, c)
	case "renew":
		out = lease.Renew(st, *id, *reqID, *at, c)
	case "status":
		out = lease.Status(st, *id, *at, c)
	default:
		fmt.Fprintln(os.Stderr, "未知子命令："+flag.Arg(0))
		os.Exit(2)
	}
	save(*state, st)
	emit(out)
}
