package main

import (
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"os/signal"

	"github.com/KirCute/go-smb2-alist/bonjour"
	"github.com/KirCute/go-smb2-alist/config"
	smb2 "github.com/KirCute/go-smb2-alist/server"
	"github.com/KirCute/go-smb2-alist/vfs"

	log "github.com/sirupsen/logrus"
)

var cfg config.AppConfig

func main() {

	homeDir, _ := os.UserHomeDir()
	cfg = config.NewConfig([]string{
		"smb2.ini",
		homeDir + "/.fuse-t/smb2.ini",
	})
	initLogs()

	srv := smb2.NewServer(
		&smb2.ServerConfig{
			MaxIOReads:  cfg.MaxIOReads,
			MaxIOWrites: cfg.MaxIOWrites,
			Xatrrs:      cfg.Xatrrs,
		},
		&smb2.NTLMAuthenticator{
			TargetSPN: "",
			NbDomain:  cfg.Hostname,
			NbName:    cfg.Hostname,
			DnsName:   cfg.Hostname + ".local",
			DnsDomain: ".local",
			UserPassword: func(user string) (string, bool) {
				if user == "a" {
					return "a", true
				}
				return "", false
			},
			AllowGuest: cfg.AllowGuest,
		},
		func(string) map[string]vfs.VFSFileSystem {
			return map[string]vfs.VFSFileSystem{cfg.ShareName: NewPassthroughFS(cfg.MountDir + "/_smbTest")}
		},
	)

	log.Infof("Starting server at %s", cfg.ListenAddr)
	go srv.Serve(cfg.ListenAddr)
	if cfg.Advertise {
		go bonjour.Advertise(cfg.ListenAddr, cfg.Hostname, cfg.Hostname, cfg.ShareName, true)
	}

	//go stats.StatServer(":9092")

	waitSignal()
}

func initLogs() {
	log.Infof("debug level %v", cfg.Debug)
	if cfg.Debug {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	homeDir, _ := os.UserHomeDir()
	if !cfg.Console {
		log.SetOutput(&lumberjack.Logger{
			Filename:   homeDir + "smb2-go.log",
			MaxSize:    100, // megabytes
			MaxBackups: 3,
			MaxAge:     28,   //days
			Compress:   true, // disabled by default
		})
	} else {
		log.SetOutput(os.Stdout)
	}
}

func waitSignal() {
	// Ctrl+C handling
	handler := make(chan os.Signal, 1)
	signal.Notify(handler, os.Interrupt)
	for sig := range handler {
		if sig == os.Interrupt {
			bonjour.Shutdown()
			break
		}
	}
}
