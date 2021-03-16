//+build !test

package cmd

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/chanify/chanify/core"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	var serveCmd = &cobra.Command{
		Use:   "serve",
		Short: "Launch chanify api server",
		Long:  `Launch service for chanify api server.`,
		Run: func(cmd *cobra.Command, args []string) {
			srv := &http.Server{
				Addr:           fmt.Sprintf("%s:%d", viper.GetString("server.host"), viper.GetInt("server.port")),
				ReadTimeout:    10 * time.Second,
				WriteTimeout:   10 * time.Second,
				MaxHeaderBytes: 1 << 20,
			}
			log.Println("Launching service...")
			go func() {
				c := core.New()
				if c == nil {
					log.Fatalln("Create service failed!")
					return
				}
				defer c.Close()
				secret := viper.GetString("server.secret")
				if len(secret) <= 0 {
					log.Fatalln("Secret not found!")
					return
				}
				c.SetSecret(secret)
				c.SetVersion(Version)
				c.SetEndpoint(GetEndpoint())
				c.SetName(viper.GetString("server.name"))
				c.InitFeatures()
				srv.Handler = c.APIHandler()
				log.Println("Launch service", srv.Addr)
				if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					log.Println("Launch service failed:", err)
				}
			}()
			quit := make(chan os.Signal, 1)
			signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
			<-quit
			log.Println("Shutting down server...")
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := srv.Shutdown(ctx); err != nil {
				log.Println("Shutdown service failed:", err)
			}
			log.Println("Shutdown service success.")
		},
	}
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().String("host", "127.0.0.1", "Http restful service hostname")
	serveCmd.Flags().Int("port", 8080, "Http restful service port")
	serveCmd.Flags().String("endpoint", "", "Http restful service endpoint")
	serveCmd.Flags().String("name", "", "Http service name")
	serveCmd.Flags().String("secret", "", "Secret for service key")
	viper.BindPFlag("server.host", serveCmd.Flags().Lookup("host"))         // nolint: errcheck
	viper.BindPFlag("server.port", serveCmd.Flags().Lookup("port"))         // nolint: errcheck
	viper.BindPFlag("server.endpoint", serveCmd.Flags().Lookup("endpoint")) // nolint: errcheck
	viper.BindPFlag("server.name", serveCmd.Flags().Lookup("name"))         // nolint: errcheck
	viper.BindPFlag("server.secret", serveCmd.Flags().Lookup("secret"))     // nolint: errcheck
}

func GetEndpoint() string {
	endpoint := viper.GetString("server.endpoint")
	if len(endpoint) <= 0 {
		hostname := viper.GetString("server.hostname")
		if len(hostname) <= 0 {
			hostname = viper.GetString("server.host")
		}
		if len(hostname) > 0 {
			port := viper.GetInt("server.port")
			if port == 80 {
				endpoint = "http://" + hostname
			} else if port == 443 {
				endpoint = "https://" + hostname
			} else {
				endpoint = fmt.Sprintf("http://%s:%d", hostname, port)
			}
		}
	}
	return endpoint
}
