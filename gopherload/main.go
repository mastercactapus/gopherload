package main

import log "github.com/sirupsen/logrus"

import "github.com/spf13/cobra"

var bindAddr string

var (
	mainCmd  = &cobra.Command{}
	serveCmd = &cobra.Command{
		Use:   "serve",
		Short: "Run an HTTP-controlled load test server.",
		Run:   runServe,
	}
)

func runServe(cmd *cobra.Command, args []string) {
	addr, err := cmd.Flags().GetString("bind")
	if err != nil {
		log.Fatalln(err)
	}
	Serve(addr)
}
func main() {
	serveCmd.Flags().StringP("bind", "b", ":8000", "Bind address. The address:port to bind the HTTP server to.")
	mainCmd.AddCommand(serveCmd)
	mainCmd.Execute()
}
