package service

import (
	"context"
	"fmt"
	"github.com/nettyrnp/exch-rates/api/common"
	"github.com/nettyrnp/exch-rates/config"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/nettyrnp/exch-rates/api/sys/repository"
)

func TestService(t *testing.T) {
	t.Parallel()

	repo, closer, repoErr := repository.NewDockerRepo()
	defer closer()
	require.NoError(t, repoErr)
	require.NotNil(t, repo)

	t.Run("get momental", testGetMomental(repo))
	t.Run("get history", testGetHistory(repo))
}

func testGetMomental(repo *repository.RDBMSRepository) func(t *testing.T) {
	return func(t *testing.T) {
		svc := New(config.Config{}, "", repo, nil)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		t1, _ := common.ParseTime("2020-03-22 15:08:29")
		rate1, err := svc.GetMomental(ctx, "USD", t1)
		require.NoError(t, err)
		assert.Len(t, rate1, 5)

		fmt.Printf(">> rate1: %v\n", rate1)
	}
}

func testGetHistory(repo *repository.RDBMSRepository) func(t *testing.T) {
	return func(t *testing.T) {
		// todo
		// ...
	}
}
