package main

import (
	"flint/config"
	"flint/mc"
	"flint/proxy"
	"log"
)

func main() {
	configWatcher, err := config.WatchConfig("./config.toml")
	defer configWatcher.Close()
	if err != nil {
		log.Fatalln("fatal: failed to load config:", err)
	}

	server, err := mc.Listen(configWatcher.CurrentConfig.Ip, configWatcher.CurrentConfig.Port)
	defer server.Close()
	if err != nil {
		log.Fatalln("fatal: failed to start server:", err)
	}

	proxyHandler := proxy.NewHandler(configWatcher.CurrentConfig)
	configWatcher.OnConfigChanged = proxyHandler.UpdateConfig

	log.Printf("Listening on %s\n", server.Addr().String())
	for {
		conn, err := server.Accept()
		if err != nil {
			log.Fatalln("fatal: failed to accept connection:", err)
		}

		go proxyHandler.HandleConn(conn)
	}
}
