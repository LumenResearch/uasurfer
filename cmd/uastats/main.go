package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/LumenResearch/uasurfer"
)

func main() {
	var count int
	ua := &uasurfer.UserAgent{}
	stats := stats{
		BrowserNames: make(map[uasurfer.BrowserName]int),
		OSNames:      make(map[uasurfer.OSName]int),
		DeviceTypes:  make(map[uasurfer.DeviceType]int),
	}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		count++
		ua.Reset()
		uasurfer.ParseUserAgent(scanner.Text(), ua)
		stats.BrowserNames[ua.Browser.Name]++
		stats.OSNames[ua.OS.Name]++
		stats.DeviceTypes[ua.DeviceType]++
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}
	fmt.Printf("Read %d useragents\n", count)
	fmt.Println()
	stats.Summary(count, os.Stdout)
}

type stats struct {
	OSNames      map[uasurfer.OSName]int
	BrowserNames map[uasurfer.BrowserName]int
	DeviceTypes  map[uasurfer.DeviceType]int
}

func (s *stats) Summary(total int, dest io.Writer) {
	browserCounts := make([]stringCount, 0, len(s.BrowserNames))
	for k, v := range s.BrowserNames {
		browserCounts = append(browserCounts, stringCount{name: k.String(), count: v})
	}
	sort.Slice(browserCounts, func(i, j int) bool { return browserCounts[j].count < browserCounts[i].count }) // by count reversed
	_, _ = fmt.Fprintln(dest, "Browsers")
	err := writeTable(browserCounts, total, dest)
	if err != nil {
		fmt.Fprintf(os.Stderr, "writing summary: %v", err)
		return
	}

	_, _ = fmt.Fprintln(dest)
	osCounts := make([]stringCount, 0, len(s.OSNames))
	for k, v := range s.OSNames {
		osCounts = append(osCounts, stringCount{name: k.String(), count: v})
	}
	sort.Slice(osCounts, func(i, j int) bool { return osCounts[j].count < osCounts[i].count }) // by count reversed
	_, _ = fmt.Fprintln(dest, "Operating Systems")
	err = writeTable(osCounts, total, dest)
	if err != nil {
		fmt.Fprintf(os.Stderr, "writing summary: %v", err)
		return
	}

	_, _ = fmt.Fprintln(dest)
	deviceCounts := make([]stringCount, 0, len(s.DeviceTypes))
	for k, v := range s.DeviceTypes {
		deviceCounts = append(deviceCounts, stringCount{name: k.String(), count: v})
	}
	sort.Slice(deviceCounts, func(i, j int) bool { return deviceCounts[j].count < deviceCounts[i].count }) // by count reversed
	_, _ = fmt.Fprintln(dest, "Device Types")
	err = writeTable(deviceCounts, total, dest)
	if err != nil {
		fmt.Fprintf(os.Stderr, "writing summary: %v", err)
		return
	}
}

func writeTable(counts []stringCount, total int, dest io.Writer) error {
	tw := tabwriter.NewWriter(dest, 10, 1, 2, ' ', 0)
	for i := range counts {
		// tabwriter buffers, so any write error surfaces from Flush below.
		_, _ = fmt.Fprintf(tw, "%s\t%d (%.2f%%)\n", counts[i].name, counts[i].count, percent(counts[i].count, total))
	}
	return tw.Flush()
}

type stringCount struct {
	name  string
	count int
}

func percent(num, den int) float64 {
	return float64(num) / float64(den) * 100.0
}
