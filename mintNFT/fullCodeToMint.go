package mintNFT

import (
	"context"
	"fmt"
	"log"

	"github.com/blocto/solana-go-sdk/client"
	"github.com/blocto/solana-go-sdk/common"
	"github.com/blocto/solana-go-sdk/program/associated_token_account"
	"github.com/blocto/solana-go-sdk/program/metaplex/token_metadata"
	"github.com/blocto/solana-go-sdk/program/system"
	"github.com/blocto/solana-go-sdk/program/token"
	"github.com/blocto/solana-go-sdk/rpc"
	"github.com/blocto/solana-go-sdk/types"
)

// FUarP2p5EnxD66vVDL4PWRoWMzA56ZVHG24hpEDFShEz
var feePayer, _ = types.AccountFromBytes([]byte{162, 128, 223, 203, 33, 217, 35, 50, 114, 79, 106, 50, 93, 174, 66, 2, 47, 22, 191, 158, 233, 41, 109, 52, 49, 255, 214, 3, 3, 182, 50, 185, 2, 212, 203, 14, 73, 174, 65, 37, 136, 138, 5, 84, 53, 62, 136, 198, 69, 3, 211, 0, 20, 214, 9, 140, 211, 24, 14, 197, 109, 104, 35, 65})

// 9aE476sH92Vz7DMPyq5WLPkrKWivxeuTKEFKd2sZZcde
//var alice, _ = types.AccountFromBase58("4voSPg3tYuWbKzimpQK9EbXHmuyy5fUrtXvpLDMLkmY6TRncaTHAKGD8jUg3maB5Jbrd9CkQg4qjJMyN6sQvnEF2")

func MainCode() (string, string, error) {
	c := client.NewClient(rpc.TestnetRPCEndpoint)
	fmt.Println("--------------------fee payer-----------------------------------------", feePayer.PublicKey)
	mint := types.NewAccount()
	fmt.Printf("NFT: %v\n", mint.PublicKey.ToBase58())
	var mintKey = mint.PublicKey.ToBase58()
	//var mintSign = string(mint.PrivateKey)

	ata, _, err := common.FindAssociatedTokenAddress(feePayer.PublicKey, mint.PublicKey)
	if err != nil {
		log.Fatalf("failed to find a valid ata, err: %v", err)
	}
	var source = ata.ToBase58()

	tokenMetadataPubkey, err := token_metadata.GetTokenMetaPubkey(mint.PublicKey)
	if err != nil {
		log.Fatalf("failed to find a valid token metadata, err: %v", err)

	}

	tokenMasterEditionPubkey, err := token_metadata.GetMasterEdition(mint.PublicKey)
	if err != nil {
		log.Fatalf("failed to find a valid master edition, err: %v", err)
	}

	mintAccountRent, err := c.GetMinimumBalanceForRentExemption(context.Background(), token.MintAccountSize)
	if err != nil {
		log.Fatalf("failed to get mint account rent, err: %v", err)
	}

	recentBlockhashResponse, err := c.GetLatestBlockhash(context.Background())
	if err != nil {
		log.Fatalf("failed to get recent blockhash, err: %v", err)
	}
	u64zero := uint64(0)
	tx, err := types.NewTransaction(types.NewTransactionParam{
		Signers: []types.Account{mint, feePayer},
		Message: types.NewMessage(types.NewMessageParam{
			FeePayer:        feePayer.PublicKey,
			RecentBlockhash: recentBlockhashResponse.Blockhash,
			Instructions: []types.Instruction{
				system.CreateAccount(system.CreateAccountParam{
					From:     feePayer.PublicKey,
					New:      mint.PublicKey,
					Owner:    common.TokenProgramID,
					Lamports: mintAccountRent,
					Space:    token.MintAccountSize,
				}),
				token.InitializeMint(token.InitializeMintParam{
					Decimals: 0,
					Mint:     mint.PublicKey,
					MintAuth: feePayer.PublicKey,
				}),
				token_metadata.CreateMetadataAccount(token_metadata.CreateMetadataAccountParam{
					Metadata:                tokenMetadataPubkey,
					Mint:                    mint.PublicKey,
					MintAuthority:           feePayer.PublicKey,
					Payer:                   feePayer.PublicKey,
					UpdateAuthority:         feePayer.PublicKey,
					UpdateAuthorityIsSigner: true,
					IsMutable:               false,
					MintData: token_metadata.Data{
						Name:                 "Tracified NFT",
						Symbol:               "Kels",
						Uri:                  "https://tillit-explorer.netlify.app/proof-verification?type=pobl&txn=241bf3d832f9f73efd66abc1468b7ab10364c46aeb473fd4638f31043f976585",
						SellerFeeBasisPoints: 500,
						Creators: &[]token_metadata.Creator{
							{
								Address:  feePayer.PublicKey,
								Verified: true,
								Share:    100,
							},
						},
					},
				}),
				associated_token_account.CreateAssociatedTokenAccount(associated_token_account.CreateAssociatedTokenAccountParam{
					Funder:                 feePayer.PublicKey,
					Owner:                  feePayer.PublicKey,
					Mint:                   mint.PublicKey,
					AssociatedTokenAccount: ata,
				}),
				token.MintTo(token.MintToParam{
					Mint:   mint.PublicKey,
					To:     ata,
					Auth:   feePayer.PublicKey,
					Amount: 1,
				}),

				token_metadata.CreateMasterEdition(token_metadata.CreateMasterEditionParam{
					Edition:         tokenMasterEditionPubkey,
					Mint:            mint.PublicKey,
					UpdateAuthority: feePayer.PublicKey,
					MintAuthority:   feePayer.PublicKey,
					Metadata:        tokenMetadataPubkey,
					Payer:           feePayer.PublicKey,
					MaxSupply:       &u64zero,
				}),
			},
		}),
	})
	if err != nil {
		log.Fatalf("failed to new a tx, err: %v", err)
	}

	sig, err := c.SendTransaction(context.Background(), tx)
	if err != nil {
		log.Fatalf("failed to send tx, err: %v", err)
	}

	fmt.Println(sig)
	return mintKey, source, nil
}
