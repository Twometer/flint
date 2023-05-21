package main

import (
	"flint/config"
	"flint/mc"
	"fmt"
	"log"
	"net"
)

func handlePacket(conn *mc.Conn, packet mc.Packet) error {
	//log.Printf("Received packet %d with len %d\n", packet.Id, packet.Data.Len())

	if conn.State == mc.StateInitial && packet.Id == 0x00 { // handshake
		handshakePacket, err := mc.DecodeHandshakePacket(packet)
		if err != nil {
			return err
		}

		expectedPort := conn.LocalPort()
		if handshakePacket.ServerPort != expectedPort {
			return fmt.Errorf("client connected to port %d but specified port %d", expectedPort, handshakePacket.ServerPort)
		}

		conn.State = handshakePacket.NextState
	} else if conn.State == mc.StateStatus && packet.Id == 0x00 {
		responsePacket, err := mc.EncodeStatusResponsePacket(mc.StatusResponsePacket{
			Version: mc.StatusVersion{
				Name:     "1.7.10",
				Protocol: 5,
			},
			Players: mc.StatusPlayers{
				Max:    0,
				Online: 0,
			},
			Description: mc.ChatComponent{
				Text: "§cNo Minecraft server under this address.",
			},
		})
		if err != nil {
			return err
		}

		err = conn.WritePacket(responsePacket)
		if err != nil {
			return err
		}
	} else if conn.State == mc.StateStatus && packet.Id == 0x01 {
		pingPacket, err := mc.DecodePingPacket(packet)
		if err != nil {
			return err
		}

		pongPacket, err := mc.EncodePingPacket(pingPacket)
		if err != nil {
			return err
		}

		err = conn.WritePacket(pongPacket)
		if err != nil {
			return err
		}
	} else if conn.State == mc.StateLogin && packet.Id == 0x00 {
		disconnectPacket, err := mc.EncodeDisconnectPacket(mc.DisconnectPacket{
			Reason: mc.ChatComponent{Text: "No server running under this address!"},
		})
		if err != nil {
			return err
		}

		err = conn.WritePacket(disconnectPacket)
		if err != nil {
			return err
		}
	}

	return nil
}

func handleConnection(conn mc.Conn) {
	defer conn.Close()
	for {
		packet, err := conn.ReadPacket()
		if err != nil {
			log.Printf("failed to read packet: %v\n", err)
			return
		}

		err = handlePacket(&conn, packet)
		if err != nil {
			log.Printf("failed to handle packet: %v\n", err)
			return
		}
	}
}

func main() {
	configWatcher, err := config.WatchConfig("./config.toml")
	defer configWatcher.Close()
	if err != nil {
		log.Fatalln("failed to load config:", err)
	}

	conf := configWatcher.CurrentConfig
	address := fmt.Sprintf("%s:%d", conf.Ip, conf.Port)
	server, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalln("failed to start server:", err)
	}

	log.Printf("Server running at %s\n", address)
	for {
		conn, err := server.Accept()
		if err != nil {
			log.Fatalln("failed to accept connection:", err)
		}

		mcConn := mc.NewConn(conn)

		log.Println("Accepted new connection from", conn.RemoteAddr())
		go handleConnection(mcConn)
	}
}
