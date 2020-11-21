package cmd

import (
	"log"
	"net/http"

	"github.com/aryanbaghi/logprom/internal/logprom"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(serveCmd)
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Print the version number of Hugo",
	Long:  `All software has versions. This is Hugo's`,
	Run: func(cmd *cobra.Command, args []string) {
		registery := prometheus.NewRegistry()
		var registerer prometheus.Registerer = registery
		logprom.TrackConfig(*cPath, &registerer)
		http.Handle("/metrics", promhttp.HandlerFor(
			registery,
			promhttp.HandlerOpts{},
		))
		log.Fatal(http.ListenAndServe("0.0.0.0:3030", nil))
	},
}
