package main

import (
	"crypto/tls"
	"fmt"
	"net/url"
	"time"

	"github.com/mastercactapus/gopherload"
)
import "github.com/spf13/cobra"

var (
	mainCmd = &cobra.Command{
		Use:   "gopherload <url...>",
		Short: "Print metrics on GET requests",
		Run:   run,
	}
	loadCmd = &cobra.Command{
		Use:   "load <url>",
		Short: "Perform a load test against url",
		Run:   runLoad,
	}
)

func runLoad(cmd *cobra.Command, args []string) {
	t := &gopherload.SimpleTemplate{
		URL: "http://localhost:8000/",
	}
	target := gopherload.Target("localhost:8000")

	load := &gopherload.LoadGenerator{CLimit: 100, Source: t, Profiler: target}
	resCh := load.Start(5)
	for res := range resCh {
		if res.Err != nil {
			fmt.Println("Error:", res.Err.Error())
		} else {
			fmt.Printf("ResponseTime: %s\n", res.Profile.TotalElapsed.String())
		}
	}
}

func run(cmd *cobra.Command, args []string) {
	for _, urlStr := range args {
		u, err := url.Parse(urlStr)
		if err != nil {
			panic(err)
		}
		t := &gopherload.SimpleTemplate{
			URL: urlStr,
		}
		req, err := t.NewRequest()
		if err != nil {
			panic(err)
		}

		var tlsConfig *tls.Config
		if u.Scheme == "https" {
			tlsConfig = new(tls.Config)
			tlsConfig.ServerName = u.Host
		}
		p, err := gopherload.Target(u.Host).Profile(req, tlsConfig)
		if err != nil {
			panic(err)
		}

		fmt.Printf(`
Start: %s
Dial:  %s
TLS:   %s
Send:  %s
Recv:  %s

TTFB:  %s
TTH:   %s
Total: %s

Sent:     %d
Recv:     %d
RecvBody: %d

Status: %d
`, p.Start.Format(time.RubyDate), p.DialElapsed.String(),
			(p.TLSElapsed - p.DialElapsed).String(), (p.SendElapsed - p.TLSElapsed).String(),
			(p.TotalElapsed - p.SendElapsed).String(), p.TTFBElapsed.String(),
			p.HeadersElapsed.String(), p.TotalElapsed.String(), p.SentBytes, p.RecvBytes,
			p.RecvBodyBytes, p.StatusCode)
	}

}

func main() {
	mainCmd.AddCommand(loadCmd)
	mainCmd.Execute()
}
