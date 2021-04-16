package main

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/kardiachain/go-kaiclient/kardia"
	"github.com/kardiachain/go-kardia/lib/abi"
	"github.com/kardiachain/go-kardia/lib/common"
	"github.com/kardiachain/go-kardia/lib/crypto"
	"go.uber.org/zap"
)

var logger *zap.Logger

const (
	RewardSMCAddress = `0x097da74bd636FBC91c017e03BF134a884A8C3dD1`
	RewardSMCAbi     = `[
	{
		"inputs": [
			{
				"internalType": "address",
				"name": "_fado",
				"type": "address"
			}
		],
		"payable": false,
		"stateMutability": "nonpayable",
		"type": "constructor"
	},
	{
		"anonymous": false,
		"inputs": [
			{
				"indexed": true,
				"internalType": "address",
				"name": "previousOwner",
				"type": "address"
			},
			{
				"indexed": true,
				"internalType": "address",
				"name": "newOwner",
				"type": "address"
			}
		],
		"name": "OwnershipTransferred",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{
				"indexed": false,
				"internalType": "string",
				"name": "orderId",
				"type": "string"
			},
			{
				"indexed": false,
				"internalType": "string",
				"name": "userId",
				"type": "string"
			},
			{
				"indexed": false,
				"internalType": "uint256",
				"name": "rewardPerOrder",
				"type": "uint256"
			}
		],
		"name": "RewardPlaceOrder",
		"type": "event"
	},
	{
		"constant": false,
		"inputs": [
			{
				"internalType": "uint256",
				"name": "amount",
				"type": "uint256"
			}
		],
		"name": "emergencyWithdrawal",
		"outputs": [],
		"payable": false,
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"constant": true,
		"inputs": [],
		"name": "fado",
		"outputs": [
			{
				"internalType": "address",
				"name": "",
				"type": "address"
			}
		],
		"payable": false,
		"stateMutability": "view",
		"type": "function"
	},
	{
		"constant": true,
		"inputs": [
			{
				"internalType": "string",
				"name": "_userId",
				"type": "string"
			}
		],
		"name": "getTotalRewardOfUser",
		"outputs": [
			{
				"internalType": "uint256",
				"name": "",
				"type": "uint256"
			}
		],
		"payable": false,
		"stateMutability": "view",
		"type": "function"
	},
	{
		"constant": false,
		"inputs": [
			{
				"internalType": "address",
				"name": "recipient",
				"type": "address"
			},
			{
				"internalType": "string",
				"name": "_orderId",
				"type": "string"
			},
			{
				"internalType": "string",
				"name": "_userId",
				"type": "string"
			},
			{
				"internalType": "uint256",
				"name": "_rewardOrder",
				"type": "uint256"
			}
		],
		"name": "rewardPlaceOrder",
		"outputs": [],
		"payable": false,
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"constant": true,
		"inputs": [
			{
				"internalType": "string",
				"name": "",
				"type": "string"
			}
		],
		"name": "totalRewardOfUser",
		"outputs": [
			{
				"internalType": "uint256",
				"name": "",
				"type": "uint256"
			}
		],
		"payable": false,
		"stateMutability": "view",
		"type": "function"
	},
	{
		"constant": false,
		"inputs": [
			{
				"internalType": "address",
				"name": "newOwner",
				"type": "address"
			}
		],
		"name": "transferOwnership",
		"outputs": [],
		"payable": false,
		"stateMutability": "nonpayable",
		"type": "function"
	}
]`
)

func init() {
	lgr, err := zap.NewDevelopment()
	if err != nil {
		panic("cannot create logger")
	}
	logger = lgr
}

func main() {
	fmt.Println("Start")
}

func KardiaNode() (kardia.Node, error) {
	url := "https://dev-1.kardiachain.io"
	node, err := kardia.NewNode(url, logger)
	if err != nil {
		return nil, err
	}
	return node, nil
}

func RewardContract(node kardia.Node) (*kardia.BoundContract, error) {
	rewardABI, err := abi.JSON(bytes.NewReader([]byte(RewardSMCAbi)))
	if err != nil {
		return nil, err
	}
	smc := kardia.NewBoundContract(node, &rewardABI, common.HexToAddress(RewardSMCAddress))
	return smc, nil
}

func OwnerWalletInfo() (*ecdsa.PublicKey, *ecdsa.PrivateKey, error) {
	privateKey, err := crypto.HexToECDSA("63e16b5334e76d63ee94f35bd2a81c721ebbbb27e81620be6fc1c448c767eed9")
	if err != nil {
		return nil, nil, err
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, nil, err
	}

	return publicKeyECDSA, privateKey, nil
}
func RewardToUser(ctx context.Context, userID, orderID string, amount float64) error {
	// Assume:
	// wallet `0x4f36A53DC32272b97Ae5FF511387E2741D727bdb` is Fado master wallet,
	// which owner of rewardSMC
	// And wallet `0xec6D6D84369553655fE7235c07Af8742504f8397` belong to a FADO user
	// Priv: 0xa6865dd5f02bc77c21e6e9be9fae0636a5594aa70f518c15fd3c96745f12026c
	receivedAddress := common.HexToAddress("0xec6D6D84369553655fE7235c07Af8742504f8397")

	// After FADO's user completed a order, then he should be rewarded with FADO token.

	// 1. Create connection to Kardia node
	node, err := KardiaNode()
	if err != nil {
		return err
	}
	// 2. Get rewardSMC owner info, since reward should only called by SMC owner
	pubKey, privKey, err := OwnerWalletInfo()
	if err != nil {
		return err
	}

	fromAddress := crypto.PubkeyToAddress(*pubKey)

	nonce, err := node.NonceAt(context.Background(), fromAddress.String())
	if err != nil {
		return err
	}
	gasLimit := uint64(3100000)
	gasPrice := big.NewInt(1000000000) // 1 OXY
	// Amount is a *big.Int, hence FADO token is 18 decimals, so x amount should be x*10^18 in *big.Int
	// Assume user reward 0.2 FADO token
	reward := kardia.FloatToBigInt(0.2, 18)

	auth := kardia.NewKeyedTransactor(privKey)
	auth.Nonce = nonce
	auth.Value = big.NewInt(0) // in wei
	auth.GasLimit = gasLimit   // in units
	auth.GasPrice = gasPrice

	rewardSMC, err := RewardContract(node)
	if err != nil {
		return err
	}
	/*
		{
				"constant": false,
				"inputs": [
					{
						"internalType": "address",
						"name": "recipient",
						"type": "address"
					},
					{
						"internalType": "string",
						"name": "_orderId",
						"type": "string"
					},
					{
						"internalType": "string",
						"name": "_userId",
						"type": "string"
					},
					{
						"internalType": "uint256",
						"name": "_rewardOrder",
						"type": "uint256"
					}
				],
				"name": "rewardPlaceOrder",
				"outputs": [],
				"payable": false,
				"stateMutability": "nonpayable",
				"type": "function"
			},
	*/
	tx, err := rewardSMC.Transact(auth, "rewardPlaceOrder", receivedAddress, orderID, userID, reward)
	if err != nil {
		return err
	}

	fmt.Printf("tx sent: %s", tx.Hash().Hex())
	return nil
}
