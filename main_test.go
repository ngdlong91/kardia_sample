// Package main
package main

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFloatToBigInt(t *testing.T) {
	var Hydro = big.NewFloat(1000000000000000000)
	rewardInFloat := new(big.Float).SetFloat64(0.2)

	rewardInFloat.Mul(rewardInFloat, Hydro)
	reward := new(big.Int)
	rewardInFloat.Int(reward)
	fmt.Println("Reward", reward)
}

func TestReward_RewardUser(t *testing.T) {
	ctx := context.Background()
	err := RewardToUser(ctx, "testUser", "orderTest", 0.2)
	assert.Nil(t, err)
}
