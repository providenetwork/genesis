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

package repository

import (
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	entityMock "github.com/whiteblock/genesis/mocks/pkg/entity"
)

func TestDockerRepository_ContainerCreate(t *testing.T) {
	cli := new(entityMock.Client)

	expectedContainer := container.ContainerCreateCreatedBody{
		ID: "test",
	}

	cli.On("ContainerCreate", mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything).Return(expectedContainer, nil).Run(
		func(args mock.Arguments) {
			require.Len(t, args, 5)
			assert.Nil(t, args.Get(0))
			assert.Equal(t, "test", args.Get(4))
		}).Once()

	repo := NewDockerRepository()

	container, err := repo.ContainerCreate(nil, cli, new(container.Config), new(container.HostConfig),
		new(network.NetworkingConfig), "test")
	assert.NoError(t, err)

	assert.Equal(t, expectedContainer.ID, container.ID)
	cli.AssertExpectations(t)
}

func TestDockerRepository_ContainerList(t *testing.T) {
	cli := new(entityMock.Client)

	expectedList := []types.Container{}
	opts := types.ContainerListOptions{
		Size: true,
	}

	cli.On("ContainerList", mock.Anything, mock.Anything).Return(expectedList, nil).Run(
		func(args mock.Arguments) {
			require.Len(t, args, 2)
			assert.Nil(t, args.Get(0))
			assert.Equal(t, opts, args.Get(1))
		}).Once()

	repo := NewDockerRepository()

	list, err := repo.ContainerList(nil, cli, opts)
	assert.NoError(t, err)

	assert.Equal(t, len(expectedList), len(list))
	assert.ElementsMatch(t, expectedList, list)
	cli.AssertExpectations(t)
}

func TestDockerRepository_ContainerRemove(t *testing.T) {
	cli := new(entityMock.Client)

	opts := types.ContainerRemoveOptions{
		RemoveVolumes: true,
	}
	containerID := "test"

	cli.On("ContainerRemove", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Run(
		func(args mock.Arguments) {
			require.Len(t, args, 3)
			assert.Nil(t, args.Get(0))
			assert.Equal(t, opts, args.Get(2))
			assert.Equal(t, containerID, args.Get(1))
		}).Once()

	repo := NewDockerRepository()

	err := repo.ContainerRemove(nil, cli, "test", opts)
	assert.NoError(t, err)

	cli.AssertExpectations(t)
}

func TestDockerRepository_ContainerStart(t *testing.T) {
	cli := new(entityMock.Client)

	containerID := "test"
	opts := types.ContainerStartOptions{}

	cli.On("ContainerStart", mock.Anything, mock.Anything, mock.Anything).Return(nil).Run(
		func(args mock.Arguments) {
			require.Len(t, args, 3)
			assert.Nil(t, args.Get(0))
			assert.Equal(t, containerID, args.Get(1))
			assert.Equal(t, opts, args.Get(2))
		}).Once()

	repo := NewDockerRepository()

	err := repo.ContainerStart(nil, cli, containerID, opts)
	assert.NoError(t, err)

	cli.AssertExpectations(t)
}

func TestDockerRepository_CopyToContainer(t *testing.T) {
	//ctx
	containerID := "id1"
	dstPath := "/foo/bar"
	//content
	options := types.CopyToContainerOptions{}

	cli := new(entityMock.Client)
	cli.On("CopyToContainer", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
		mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		require.Len(t, args, 5)
		assert.Nil(t, args.Get(0))
		assert.Equal(t, containerID, args.Get(1))
		assert.Equal(t, dstPath, args.Get(2))
		assert.Nil(t, args.Get(3))
		assert.Equal(t, options, args.Get(4))
	}).Once()
	repo := NewDockerRepository()

	err := repo.CopyToContainer(nil, cli, containerID, dstPath, nil, options)
	assert.NoError(t, err)
	cli.AssertExpectations(t)
}

func TestDockerRepository_ImageLoad(t *testing.T) {
	cli := new(entityMock.Client)

	quiet := false
	expectedRes := types.ImageLoadResponse{}

	cli.On("ImageLoad", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(expectedRes, nil).Run(
		func(args mock.Arguments) {
			require.Len(t, args, 3)
			assert.Nil(t, args.Get(0))
			assert.Equal(t, quiet, args.Get(2))
			assert.Nil(t, args.Get(1))
		}).Once()

	repo := NewDockerRepository()

	res, err := repo.ImageLoad(nil, cli, nil, quiet)
	assert.NoError(t, err)

	assert.Equal(t, expectedRes, res)
	assert.True(t, cli.AssertNumberOfCalls(t, "ImageLoad", 1))
}

func TestDockerRepository_ImagePull(t *testing.T) {
	cli := new(entityMock.Client)

	refStr := "test"
	options := types.ImagePullOptions{All: true}

	cli.On("ImagePull", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, nil).Run(
		func(args mock.Arguments) {
			require.Len(t, args, 3)
			assert.Nil(t, args.Get(0))
			assert.Equal(t, refStr, args.Get(1))
			assert.Equal(t, options, args.Get(2))
		}).Once()

	repo := NewDockerRepository()

	_, err := repo.ImagePull(nil, cli, refStr, options)
	assert.NoError(t, err)

	cli.AssertExpectations(t)
}

func TestDockerRepository_ImageList(t *testing.T) {
	cli := new(entityMock.Client)

	options := types.ImageListOptions{All: true}
	expectedList := []types.ImageSummary{}

	cli.On("ImageList", mock.Anything, mock.Anything, mock.Anything).Return(expectedList, nil).Run(
		func(args mock.Arguments) {
			require.Len(t, args, 2)
			assert.Nil(t, args.Get(0))
			assert.Equal(t, options, args.Get(1))
		}).Once()

	repo := NewDockerRepository()

	list, err := repo.ImageList(nil, cli, options)
	assert.NoError(t, err)

	assert.Equal(t, expectedList, list)
	cli.AssertExpectations(t)
}

func TestDockerRepository_NetworkConnect(t *testing.T) {
	cli := new(entityMock.Client)

	networkID := "test"
	containerID := "test"
	config := new(network.EndpointSettings)

	cli.On("NetworkConnect", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Run(
		func(args mock.Arguments) {
			require.Len(t, args, 4)
			assert.Nil(t, args.Get(0))
			assert.Equal(t, networkID, args.Get(1))
			assert.Equal(t, containerID, args.Get(2))
			assert.Equal(t, config, args.Get(3))
		}).Once()

	repo := NewDockerRepository()

	err := repo.NetworkConnect(nil, cli, networkID, containerID, config)
	assert.NoError(t, err)

	cli.AssertExpectations(t)
}

func TestDockerRepository_NetworkCreate(t *testing.T) {
	cli := new(entityMock.Client)

	name := "test"
	options := types.NetworkCreate{}

	expectedCreated := types.NetworkCreateResponse{ID: "test"}

	cli.On("NetworkCreate", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(expectedCreated, nil).Run(
		func(args mock.Arguments) {
			require.Len(t, args, 3)
			assert.Nil(t, args.Get(0))
			assert.Equal(t, name, args.Get(1))
			assert.Equal(t, options, args.Get(2))
		}).Once()

	repo := NewDockerRepository()

	created, err := repo.NetworkCreate(nil, cli, name, options)
	assert.NoError(t, err)

	assert.Equal(t, expectedCreated.ID, created.ID)
	cli.AssertExpectations(t)
}

func TestDockerRepository_NetworkDisconnect(t *testing.T) {
	cli := new(entityMock.Client)

	networkID := "test"
	containerID := "test"
	force := true

	cli.On("NetworkDisconnect", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Run(
		func(args mock.Arguments) {
			require.Len(t, args, 4)
			assert.Nil(t, args.Get(0))
			assert.Equal(t, networkID, args.Get(1))
			assert.Equal(t, containerID, args.Get(2))
			assert.Equal(t, force, args.Get(3))
		}).Once()

	repo := NewDockerRepository()

	err := repo.NetworkDisconnect(nil, cli, networkID, containerID, force)
	assert.NoError(t, err)

	cli.AssertExpectations(t)
}

func TestDockerRepository_NetworkRemove(t *testing.T) {
	cli := new(entityMock.Client)

	networkID := "test"

	cli.On("NetworkRemove", mock.Anything, mock.Anything, mock.Anything).Return(nil).Run(
		func(args mock.Arguments) {
			require.Len(t, args, 2)
			assert.Nil(t, args.Get(0))
			assert.Equal(t, networkID, args.Get(1))
		}).Once()

	repo := NewDockerRepository()

	err := repo.NetworkRemove(nil, cli, networkID)
	assert.NoError(t, err)

	cli.AssertExpectations(t)
}

func TestDockerRepository_NetworkList(t *testing.T) {
	cli := new(entityMock.Client)

	options := types.NetworkListOptions{}
	expectedList := []types.NetworkResource{}

	cli.On("NetworkList", mock.Anything, mock.Anything, mock.Anything).Return(expectedList, nil).Run(
		func(args mock.Arguments) {
			require.Len(t, args, 2)
			assert.Nil(t, args.Get(0))
			assert.Equal(t, options, args.Get(1))
		}).Once()

	repo := NewDockerRepository()

	list, err := repo.NetworkList(nil, cli, options)
	assert.NoError(t, err)

	assert.Equal(t, expectedList, list)
	cli.AssertExpectations(t)
}

func TestDockerRepository_VolumeList(t *testing.T) {
	cli := new(entityMock.Client)
	testFilters := filters.Args{}
	result := volume.VolumeListOKBody{}

	cli.On("VolumeList", mock.Anything, mock.Anything).Return(result, nil).Run(func(args mock.Arguments) {
		require.Len(t, args, 2)
		assert.Nil(t, args.Get(0))
		assert.Equal(t, testFilters, args.Get(1))
	}).Once()

	repo := NewDockerRepository()
	res, err := repo.VolumeList(nil, cli, testFilters)
	assert.NoError(t, err)

	assert.Equal(t, result, res)
	cli.AssertExpectations(t)
}

func TestDockerRepository_VolumeRemove(t *testing.T) {
	cli := new(entityMock.Client)
	volumeID := "volume1"
	isForced := true

	cli.On("VolumeRemove", mock.Anything, mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		require.Len(t, args, 3)
		assert.Nil(t, args.Get(0))
		assert.Equal(t, volumeID, args.String(1))
		assert.Equal(t, isForced, args.Bool(2))
	}).Once()

	repo := NewDockerRepository()

	err := repo.VolumeRemove(nil, cli, volumeID, isForced)
	assert.NoError(t, err)

	cli.AssertExpectations(t)
}

func TestDockerRepository_VolumeCreate(t *testing.T) {
	cli := new(entityMock.Client)
	options := volume.VolumeCreateBody{
		Name:   "test_volume",
		Labels: map[string]string{"foo": "bar"},
	}

	expectedVol := types.Volume{
		Name:   options.Name,
		Labels: options.Labels,
	}

	cli.On("VolumeCreate", mock.Anything, mock.Anything).Return(expectedVol, nil).Run(func(args mock.Arguments) {
		require.Len(t, args, 2)
		assert.Nil(t, args.Get(0))
		assert.Equal(t, options, args.Get(1))
	}).Once()

	repo := NewDockerRepository()

	vol, err := repo.VolumeCreate(nil, cli, options)
	assert.NoError(t, err)

	assert.Equal(t, expectedVol.Name, vol.Name)
	assert.Equal(t, expectedVol.Labels, vol.Labels)

	assert.True(t, cli.AssertNumberOfCalls(t, "VolumeCreate", 1))
}