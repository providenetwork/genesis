/*
	Copyright 2019 whiteblock Inc.
	This file is a part of the genesis.

	Genesis is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	Genesis is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU General Public License
	along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package service

import (
	"github.com/whiteblock/genesis/pkg/command"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	dockerVolume "github.com/docker/docker/api/types/volume"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	entityMock "github.com/whiteblock/genesis/mocks/pkg/entity"
	repoMock "github.com/whiteblock/genesis/mocks/pkg/repository"
	auxMock "github.com/whiteblock/genesis/mocks/pkg/service"
	"github.com/whiteblock/genesis/pkg/service/auxillary"
)

func TestNewDockerService(t *testing.T) {
	repo := new(repoMock.DockerRepository)
	aux := new(auxMock.DockerAuxillary)

	expected := dockerService{
		repo: repo,
		aux:  aux,
	}

	ds, err := NewDockerService(repo, aux)
	assert.NoError(t, err)

	assert.Equal(t, expected, ds)
}

func TestDockerService_CreateContainer(t *testing.T) {
	testNetwork := types.NetworkResource{Name: "Testnet", ID: "id1"}
	testContainer := command.Container{
		BoundCPUs:  nil, //TODO
		Detach:     false,
		EntryPoint: "/bin/bash",
		Environment: map[string]string{
			"FOO": "BAR",
		},
		Labels: map[string]string{
			"FOO": "BAR",
		},
		Name:    "TEST",
		Network: []string{"Testnet"},
		Ports:   map[int]int{8888: 8889},
		Volumes: []command.Mount{command.Mount{Name: "volume1", Directory: "/foo/bar", ReadOnly: false}}, //TODO
		Image:   "alpine",
		Args:    []string{"test"},
	}
	testContainer.Cpus = "2.5"
	testContainer.Memory = "5gb"

	repo := new(repoMock.DockerRepository)
	repo.On("ContainerCreate", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
		container.ContainerCreateCreatedBody{}, nil).Run(func(args mock.Arguments) {
		require.Len(t, args, 6)
		assert.Nil(t, args.Get(0))
		assert.Nil(t, args.Get(1))
		{
			config, ok := args.Get(2).(*container.Config)
			require.True(t, ok)
			require.NotNil(t, config)
			require.Len(t, config.Entrypoint, 2)
			assert.Contains(t, config.Env, "FOO=BAR")
			assert.Equal(t, testContainer.EntryPoint, config.Entrypoint[0])
			assert.Equal(t, testContainer.Args[0], config.Entrypoint[1])
			assert.Equal(t, testContainer.Name, config.Hostname)
			assert.Equal(t, testContainer.Labels, config.Labels)
			assert.Equal(t, testContainer.Image, config.Image)
			{
				_, exists := config.ExposedPorts["8889/tcp"]
				assert.True(t, exists)
			}
		}
		{
			hostConfig, ok := args.Get(3).(*container.HostConfig)
			require.True(t, ok)
			require.NotNil(t, hostConfig)
			assert.Equal(t, int64(2500000000), hostConfig.NanoCPUs)
			assert.Equal(t, int64(5000000000), hostConfig.Memory)
			{ //Port bindings
				bindings, exists := hostConfig.PortBindings["8889/tcp"]
				assert.True(t, exists)
				require.NotNil(t, bindings)
				require.Len(t, bindings, 1)
				assert.Equal(t, bindings[0].HostIP, "0.0.0.0")
				assert.Equal(t, bindings[0].HostPort, "8888")
			}

			require.NotNil(t, hostConfig.Mounts)
			require.Len(t, hostConfig.Mounts, len(testContainer.Volumes))
			for i, vol := range testContainer.Volumes {
				assert.Equal(t, vol.Name, hostConfig.Mounts[i].Source)
				assert.Equal(t, vol.Directory, hostConfig.Mounts[i].Target)
				assert.Equal(t, vol.ReadOnly, hostConfig.Mounts[i].ReadOnly)
				assert.Equal(t, mount.TypeVolume, hostConfig.Mounts[i].Type)
				assert.Nil(t, hostConfig.Mounts[i].TmpfsOptions)
				assert.Nil(t, hostConfig.Mounts[i].BindOptions)
			}
			assert.True(t, hostConfig.AutoRemove)
		}
		{
			networkingConfig, ok := args.Get(4).(*network.NetworkingConfig)
			require.True(t, ok)
			require.NotNil(t, networkingConfig)
			require.NotNil(t, networkingConfig.EndpointsConfig)
			netconf, ok := networkingConfig.EndpointsConfig[testContainer.Network[0]]
			require.True(t, ok)
			require.NotNil(t, netconf)
			assert.Equal(t, netconf.NetworkID, testNetwork.ID)
			assert.Nil(t, netconf.Links)
			assert.Nil(t, netconf.Aliases)
			assert.Nil(t, netconf.IPAMConfig)
			assert.Empty(t, netconf.IPv6Gateway)
			assert.Empty(t, netconf.GlobalIPv6Address)
			assert.Empty(t, netconf.EndpointID)
			assert.Empty(t, netconf.Gateway)
			assert.Empty(t, netconf.IPAddress)
			assert.Nil(t, netconf.DriverOpts)
		}
		containerName, ok := args.Get(5).(string)
		require.True(t, ok)
		assert.Equal(t, testContainer.Name, containerName)
	})

	aux := new(auxMock.DockerAuxillary)
	aux.On("EnsureImagePulled", mock.Anything, mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {

		require.Len(t, args, 3)
		assert.Nil(t, args.Get(0))
		assert.Nil(t, args.Get(1))
		assert.Equal(t, testContainer.Image, args.String(2))
	})
	aux.On("GetNetworkByName", mock.Anything, mock.Anything, mock.Anything).Return(
		testNetwork, nil).Run(func(args mock.Arguments) {

		require.Len(t, args, 3)
		assert.Nil(t, args.Get(0))
		assert.Nil(t, args.Get(1))
		assert.Contains(t, testContainer.Network, args.String(2))
	})

	ds, err := NewDockerService(repo, aux)
	assert.NoError(t, err)
	res := ds.CreateContainer(nil, nil, testContainer)
	assert.NoError(t, res.Error)
}

func TestDockerService_StartContainer(t *testing.T) {
	containerName := "TEST"
	repo := new(repoMock.DockerRepository)
	repo.On("ContainerStart", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Run(
		func(args mock.Arguments) {

			require.Len(t, args, 4)
			assert.Nil(t, args.Get(0))
			assert.Nil(t, args.Get(1))
			assert.Equal(t, containerName, args.String(2))
			assert.Equal(t, types.ContainerStartOptions{}, args.Get(3))
		})

	aux := new(auxMock.DockerAuxillary)
	ds, err := NewDockerService(repo, aux)
	assert.NoError(t, err)
	res := ds.StartContainer(nil, nil, containerName)
	assert.NoError(t, res.Error)
}

func TestDockerService_CreateNetwork(t *testing.T) {
	testNetwork := command.Network{
		Name:   "testnet",
		Global: true,
		Labels: map[string]string{
			"FOO": "BAR",
		},
		Gateway: "10.14.0.1",
		Subnet:  "10.14.0.0/16",
	}
	repo := new(repoMock.DockerRepository)
	repo.On("NetworkCreate", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
		types.NetworkCreateResponse{}, nil).Run(func(args mock.Arguments) {

		require.Len(t, args, 4)
		assert.Nil(t, args.Get(0))
		assert.Nil(t, args.Get(1))
		assert.Equal(t, testNetwork.Name, args.String(2))

		networkCreate, ok := args.Get(3).(types.NetworkCreate)
		require.True(t, ok)
		require.NotNil(t, networkCreate)
		assert.True(t, networkCreate.CheckDuplicate)
		assert.True(t, networkCreate.Attachable)
		assert.False(t, networkCreate.Ingress)
		assert.False(t, networkCreate.Internal)
		assert.False(t, networkCreate.EnableIPv6)
		assert.Equal(t, testNetwork.Labels, networkCreate.Labels)
		assert.Nil(t, networkCreate.ConfigFrom)

		require.NotNil(t, networkCreate.IPAM)
		assert.Equal(t, "default", networkCreate.IPAM.Driver)
		assert.Nil(t, networkCreate.IPAM.Options)
		require.NotNil(t, networkCreate.IPAM.Config)
		require.Len(t, networkCreate.IPAM.Config, 1)
		assert.Equal(t, testNetwork.Subnet, networkCreate.IPAM.Config[0].Subnet)
		assert.Equal(t, testNetwork.Gateway, networkCreate.IPAM.Config[0].Gateway)

		if testNetwork.Global {
			assert.Equal(t, "overlay", networkCreate.Driver)
			assert.Equal(t, "swarm", networkCreate.Scope)
		} else {
			assert.Equal(t, "bridge", networkCreate.Driver)
			assert.Equal(t, "local", networkCreate.Scope)
			bridgeName, ok := networkCreate.Options["com.docker.network.bridge.name"]
			assert.True(t, ok)
			assert.Equal(t, testNetwork.Name, bridgeName)
		}
	})

	aux := new(auxMock.DockerAuxillary)
	ds, err := NewDockerService(repo, aux)
	assert.NoError(t, err)
	res := ds.CreateNetwork(nil, nil, testNetwork)
	assert.NoError(t, res.Error)
	repo.AssertNumberOfCalls(t, "NetworkCreate", 1)
}

func TestDockerService_RemoveNetwork(t *testing.T) {
	repo := new(repoMock.DockerRepository)
	networks := []types.NetworkResource{
		types.NetworkResource{Name: "test1", ID: "id1"},
		types.NetworkResource{Name: "test2", ID: "id2"},
	}
	repo.On("NetworkList", mock.Anything, mock.Anything, mock.Anything).Return(networks, nil).Run(
		func(args mock.Arguments) {

			require.Len(t, args, 3)
			assert.Nil(t, args.Get(0))
			assert.Nil(t, args.Get(1))
		}).Times(len(networks))

	for _, net := range networks {
		repo.On("NetworkRemove", mock.Anything, mock.Anything, net.ID).Return(nil).Run(
			func(args mock.Arguments) {

				require.Len(t, args, 3)
				assert.Nil(t, args.Get(0))
				assert.Nil(t, args.Get(1))
			}).Once()
	}
	aux := auxillary.NewDockerAuxillary(repo)
	ds, err := NewDockerService(repo, aux)
	assert.NoError(t, err)

	for _, net := range networks {
		res := ds.RemoveNetwork(nil, nil, net.Name)
		assert.NoError(t, res.Error)
	}

	repo.AssertExpectations(t)
}

func TestDockerService_RemoveContainer(t *testing.T) {
	repo := new(repoMock.DockerRepository)
	cntrs := []types.Container{
		types.Container{Names: []string{"test1", "test3"}, ID: "id1"},
		types.Container{Names: []string{"test2", "test4"}, ID: "id2"},
	}
	repo.On("ContainerList", mock.Anything, mock.Anything, mock.Anything).Return(cntrs, nil).Run(
		func(args mock.Arguments) {

			require.Len(t, args, 3)
			assert.Nil(t, args.Get(0))
			assert.Nil(t, args.Get(1))
		}).Times(len(cntrs))

	for _, cntr := range cntrs {
		repo.On("ContainerRemove", mock.Anything, mock.Anything, cntr.ID, mock.Anything).Return(nil).Run(
			func(args mock.Arguments) {

				require.Len(t, args, 4)
				assert.Nil(t, args.Get(0))
				assert.Nil(t, args.Get(1))
				opts, ok := args.Get(3).(types.ContainerRemoveOptions)
				require.True(t, ok)
				assert.False(t, opts.RemoveVolumes)
				assert.False(t, opts.RemoveLinks)
				assert.True(t, opts.Force)

			}).Once()
	}
	aux := auxillary.NewDockerAuxillary(repo)
	ds, err := NewDockerService(repo, aux)
	assert.NoError(t, err)

	for _, cntr := range cntrs {
		res := ds.RemoveContainer(nil, nil, cntr.Names[0])
		assert.NoError(t, res.Error)
	}
}

func TestDockerService_PlaceFileInContainer(t *testing.T) {
	testDir := "/pkg"
	testContainer := types.Container{Names: []string{"test1"}, ID: "id1"}

	file := new(entityMock.IFile)
	file.On("GetDir").Return(testDir).Once()
	file.On("GetTarReader").Return(nil, nil).Once()

	repo := new(repoMock.DockerRepository)
	repo.On("CopyToContainer", mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything).Return(nil).Run(
		func(args mock.Arguments) {
			require.Len(t, args, 6)
			assert.Nil(t, args.Get(0))
			assert.Nil(t, args.Get(1))
			assert.Equal(t, testContainer.ID, args.String(2))
			assert.Equal(t, testDir, args.String(3))
			assert.Nil(t, args.Get(4))
			{
				opts, ok := args.Get(5).(types.CopyToContainerOptions)
				require.True(t, ok)
				assert.False(t, opts.AllowOverwriteDirWithFile)
				assert.False(t, opts.CopyUIDGID)
			}
		}).Once()

	aux := new(auxMock.DockerAuxillary)
	aux.On("GetContainerByName", mock.Anything, mock.Anything, mock.Anything).Return(testContainer, nil).Run(
		func(args mock.Arguments) {

			require.Len(t, args, 3)
			assert.Nil(t, args.Get(0))
			assert.Nil(t, args.Get(1))
			assert.Equal(t, testContainer.Names[0], args.String(2))
		}).Once()

	ds, err := NewDockerService(repo, aux)
	assert.NoError(t, err)

	res := ds.PlaceFileInContainer(nil, nil, testContainer.Names[0], file)
	assert.NoError(t, res.Error)

	repo.AssertExpectations(t)
	aux.AssertExpectations(t)
	file.AssertExpectations(t)
}

func TestDockerService_AttachNetwork(t *testing.T) {
	cntr := types.Container{Names: []string{"test1"}, ID: "id1"}
	net := types.NetworkResource{Name: "test2", ID: "id2"}

	aux := new(auxMock.DockerAuxillary)
	aux.On("GetContainerByName", mock.Anything, mock.Anything, mock.Anything).Return(cntr, nil).Run(
		func(args mock.Arguments) {

			require.Len(t, args, 3)
			assert.Nil(t, args.Get(0))
			assert.Nil(t, args.Get(1))
			assert.Equal(t, cntr.Names[0], args.String(2))

		}).Once()

	aux.On("GetNetworkByName", mock.Anything, mock.Anything, mock.Anything).Return(net, nil).Run(
		func(args mock.Arguments) {

			require.Len(t, args, 3)
			assert.Nil(t, args.Get(0))
			assert.Nil(t, args.Get(1))
			assert.Equal(t, net.Name, args.String(2))
		}).Once()

	repo := new(repoMock.DockerRepository)
	repo.On("NetworkConnect", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
		mock.Anything).Return(nil).Run(func(args mock.Arguments) {

		require.Len(t, args, 5)
		assert.Nil(t, args.Get(0))
		assert.Nil(t, args.Get(1))
		assert.Equal(t, net.ID, args.String(2))
		assert.Equal(t, cntr.ID, args.String(3))
		epSettings, ok := args.Get(4).(*network.EndpointSettings)
		require.True(t, ok)
		require.NotNil(t, epSettings)
		assert.Equal(t, net.ID, epSettings.NetworkID)
	}).Once()

	ds, err := NewDockerService(repo, aux)
	assert.NoError(t, err)

	res := ds.AttachNetwork(nil, nil, net.Name, cntr.Names[0])
	assert.NoError(t, res.Error)

	repo.AssertExpectations(t)
	aux.AssertExpectations(t)
}

func TestDockerService_DetachNetwork(t *testing.T) {
	cntr := types.Container{Names: []string{"test1"}, ID: "id1"}
	net := types.NetworkResource{Name: "test2", ID: "id2"}

	aux := new(auxMock.DockerAuxillary)
	aux.On("GetContainerByName", mock.Anything, mock.Anything, mock.Anything).Return(cntr, nil).Run(
		func(args mock.Arguments) {

			require.Len(t, args, 3)
			assert.Nil(t, args.Get(0))
			assert.Nil(t, args.Get(1))
			assert.Equal(t, cntr.Names[0], args.String(2))

		}).Once()

	aux.On("GetNetworkByName", mock.Anything, mock.Anything, mock.Anything).Return(net, nil).Run(
		func(args mock.Arguments) {

			require.Len(t, args, 3)
			assert.Nil(t, args.Get(0))
			assert.Nil(t, args.Get(1))
			assert.Equal(t, net.Name, args.String(2))
		}).Once()

	repo := new(repoMock.DockerRepository)
	repo.On("NetworkDisconnect", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
		mock.Anything).Return(nil).Run(func(args mock.Arguments) {

		require.Len(t, args, 5)
		assert.Nil(t, args.Get(0))
		assert.Nil(t, args.Get(1))
		assert.Equal(t, net.ID, args.String(2))
		assert.Equal(t, cntr.ID, args.String(3))
		assert.True(t, args.Bool(4))
	}).Once()

	ds, err := NewDockerService(repo, aux)
	assert.NoError(t, err)

	res := ds.DetachNetwork(nil, nil, net.Name, cntr.Names[0])
	assert.NoError(t, res.Error)

	repo.AssertExpectations(t)
	aux.AssertExpectations(t)
}

func TestDockerService_CreateVolume(t *testing.T) {
	volume := types.Volume{
		Name:   "test_volume",
		Labels: map[string]string{"foo": "bar"},
	}

	repo := new(repoMock.DockerRepository)
	repo.On("VolumeCreate", mock.Anything, mock.Anything, mock.Anything).Return(volume, nil).Run(
		func(args mock.Arguments) {
			require.Len(t, args, 3)
			assert.Nil(t, args.Get(0))
			assert.Nil(t, args.Get(1))

			vol, ok := args.Get(2).(dockerVolume.VolumeCreateBody)
			require.True(t, ok)
			require.NotNil(t, vol)
			assert.Contains(t, vol.Labels, "foo")

			assert.Equal(t, volume.Name, vol.Name)
			assert.Equal(t, volume.Labels["foo"], vol.Labels["foo"])

		}).Once()

	aux := *new(auxillary.DockerAuxillary)

	ds, err := NewDockerService(repo, aux)
	assert.NoError(t, err)

	res := ds.CreateVolume(nil, nil, command.Volume{
		Name:   "test_volume",
		Labels: map[string]string{"foo": "bar"},
	})
	assert.NoError(t, res.Error)

	repo.AssertExpectations(t)
}

func TestDockerService_RemoveVolume(t *testing.T) {
	name := "test_volume"

	repo := new(repoMock.DockerRepository)
	repo.On("VolumeRemove", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Run(
		func(args mock.Arguments) {
			require.Len(t, args, 4)
			assert.Nil(t, args.Get(0))
			assert.Nil(t, args.Get(1))
			assert.Equal(t, name, args.Get(2))
			assert.Equal(t, true, args.Get(3))
		}).Once()

	aux := *new(auxillary.DockerAuxillary)

	ds, err := NewDockerService(repo, aux)
	assert.NoError(t, err)

	res := ds.RemoveVolume(nil, nil, name)
	assert.NoError(t, res.Error)

	assert.True(t, repo.AssertNumberOfCalls(t, "VolumeRemove", 1))
	repo.AssertExpectations(t)
}
func (ds dockerService) TestDockerService_Emulation(t *testing.T) {

}
