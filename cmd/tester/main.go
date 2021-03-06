/**
 * Copyright 2019 Whiteblock Inc. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package main

import (
	"encoding/json"
	"fmt"

	"github.com/whiteblock/genesis/pkg/config"

	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	queue "github.com/whiteblock/amqp"
	"github.com/whiteblock/definition/command"
	"github.com/whiteblock/utility/utils"
)

var (
	conn *amqp.Connection
)

func mintCommand(i interface{}, orderType command.OrderType) command.Command {
	raw, err := json.Marshal(i)
	if err != nil {
		panic(err)
	}
	cmd := command.Command{
		ID:     "TEST",
		Target: command.Target{IP: "0.0.0.0"},
		Order: command.Order{
			Type: orderType,
		},
	}
	err = json.Unmarshal(raw, &cmd.Order.Payload)
	if err != nil {
		panic(err)
	}
	return cmd
}

func createVolume(name string) command.Command {
	vol := command.Volume{
		Name: name,
		Labels: map[string]string{
			"FOO": "BAR",
		},
	}

	return mintCommand(vol, command.Createvolume)
}

func removeVolume(name string) command.Command {
	return mintCommand(map[string]string{
		"name": name,
	}, command.Removevolume)
}

func removeContainer(name string) command.Command {
	return mintCommand(command.SimpleName{
		Name: name,
	}, command.Removecontainer)
}

func createNetwork(name string, num int) command.Command {
	testNetwork := command.Network{
		Name:   name,
		Global: true,
		Labels: map[string]string{
			"FOO": "BAR",
		},
		Gateway: fmt.Sprintf("10.%d.0.1", num),
		Subnet:  fmt.Sprintf("10.%d.0.0/16", num),
	}
	return mintCommand(testNetwork, command.Createnetwork)
}

func attachNetwork(networkName string, containerName string) command.Command {
	return mintCommand(map[string]string{
		"container": containerName,
		"network":   networkName,
	}, command.Attachnetwork)
}

func detachNetwork(networkName string, containerName string) command.Command {
	return mintCommand(map[string]string{
		"container": containerName,
		"network":   networkName,
	}, command.Detachnetwork)
}

func pullImage() command.Command {
	return mintCommand(map[string]string{
		"image": "debian:latest",
	}, command.Pullimage)
}

func removeNetwork(name string) command.Command {
	return mintCommand(map[string]string{"name": name}, command.Removenetwork)
}

func createContainer(name string, network string, volume string, args []string) command.Command {
	testContainer := command.Container{
		BoundCPUs: nil,
		Environment: map[string]string{
			"FOO": "BAR",
		},
		Labels: map[string]string{
			"FOO": "BAR",
		},
		Name:    name,
		Network: network,
		Volumes: []command.Mount{{
			Name:      volume,
			Directory: "/foo/bar",
			ReadOnly:  false,
		}},
		Image:      "nettools/ubuntools",
		EntryPoint: "ping",
		Args:       args,
	}
	testContainer.Cpus = "1"
	testContainer.Memory = "1gb"
	return mintCommand(testContainer, "createContainer")
}

func startContainer(name string, attach bool) command.Command {
	return mintCommand(map[string]interface{}{
		"name":   name,
		"attach": attach,
	}, command.Startcontainer)
}

func emulate(containerName string, networkName string) command.Command {
	return mintCommand(command.Netconf{
		Container: containerName,
		Network:   networkName,
		Delay:     100000,
	}, command.Emulation)
}

func getAMQPService() (queue.AMQPService, error) {
	conf, err := config.NewConfig()
	if err != nil {
		return nil, err
	}

	cmdConf, err := conf.CommandAMQP()
	if err != nil {
		return nil, err
	}
	log.WithField("commandConf", cmdConf).Debug("got the config")
	cmdConn, err := queue.OpenAMQPConnection(cmdConf.Endpoint)
	if err != nil {
		return nil, err
	}
	return queue.NewAMQPService(cmdConf, queue.NewAMQPRepository(cmdConn), conf.GetLogger()), nil
}

func main() {
	conf, err := config.NewConfig()
	if err != nil {
		panic(err)
	}

	lvl, err := log.ParseLevel(conf.Verbosity)
	if err != nil {
		panic(err)
	}
	log.SetLevel(lvl)

	serv, err := getAMQPService()
	if err != nil {
		panic(err)
	}
	err = runTest(serv)
	if err != nil {
		log.Fatal(err)
	}
}

func runTest(serv queue.AMQPService) error {
	networkNames := []string{
		utils.GetUUIDString() + "-testnet",
		utils.GetUUIDString() + "-testnet",
	}
	containerNames := []string{
		utils.GetUUIDString() + "-tester",
		utils.GetUUIDString() + "-tester",
		utils.GetUUIDString() + "-tester",
	}
	volumeNames := []string{
		utils.GetUUIDString() + "-volume",
	}
	cmds := [][]command.Command{
		[]command.Command{
			createVolume(volumeNames[0]),
			createNetwork(networkNames[0], 14),
			createNetwork(networkNames[1], 15),
			pullImage(),
		},
		[]command.Command{
			createContainer(containerNames[0], networkNames[0], volumeNames[0], []string{"localhost"}),
			createContainer(containerNames[1], networkNames[0], volumeNames[0], []string{"localhost"}),
			createContainer(containerNames[2], networkNames[0], volumeNames[0], []string{"-c", "10", "localhost"}),
		},
		[]command.Command{
			startContainer(containerNames[0], false),
			startContainer(containerNames[1], false),
			startContainer(containerNames[2], true),
		},
		[]command.Command{
			attachNetwork(networkNames[1], containerNames[0]),
			attachNetwork(networkNames[1], containerNames[1]),
		},
		[]command.Command{
			detachNetwork(networkNames[0], containerNames[0]),
			emulate(containerNames[0], networkNames[1]),
			emulate(containerNames[1], networkNames[1]),
		},
		[]command.Command{
			removeContainer(containerNames[0]),
			removeContainer(containerNames[1]),
			removeContainer(containerNames[2]),
		},
		[]command.Command{
			removeVolume(volumeNames[0]),
			removeNetwork(networkNames[0]),
			removeNetwork(networkNames[1]),
		},
	}

	rawBytes, err := json.Marshal(cmds)
	if err != nil {
		return err
	}
	log.Info("Sending message")
	return serv.Send(amqp.Publishing{
		ContentType: "application/json",
		Body:        rawBytes,
	})

}
