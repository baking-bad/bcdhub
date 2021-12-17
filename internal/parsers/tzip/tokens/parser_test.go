package tokens

import (
	"testing"
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/domains"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
	"github.com/baking-bad/bcdhub/internal/models/tokenmetadata"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/stretchr/testify/assert"

	mock_general "github.com/baking-bad/bcdhub/internal/models/mock"
	mock_bmd "github.com/baking-bad/bcdhub/internal/models/mock/bigmapdiff"
	mock_block "github.com/baking-bad/bcdhub/internal/models/mock/block"
	mock_token_metadata "github.com/baking-bad/bcdhub/internal/models/mock/tokenmetadata"
	"github.com/golang/mock/gomock"
)

func ptrInt64(val int64) *int64 {
	return &val
}

func TestParser_ParseBigMapDiff(t *testing.T) {
	timestamp := time.Now()

	ctrlStorage := gomock.NewController(t)
	defer ctrlStorage.Finish()
	generalRepo := mock_general.NewMockGeneralRepository(ctrlStorage)

	ctrlBmdRepo := gomock.NewController(t)
	defer ctrlBmdRepo.Finish()
	bmdRepo := mock_bmd.NewMockRepository(ctrlBmdRepo)

	ctrlBlockRepo := gomock.NewController(t)
	defer ctrlBlockRepo.Finish()
	blocksRepo := mock_block.NewMockRepository(ctrlBlockRepo)

	ctrlTokenMetadataRepo := gomock.NewController(t)
	defer ctrlTokenMetadataRepo.Finish()
	tmRepo := mock_token_metadata.NewMockRepository(ctrlTokenMetadataRepo)

	ctrlRPC := gomock.NewController(t)
	defer ctrlRPC.Finish()
	rpc := noderpc.NewMockINode(ctrlRPC)

	rpc.EXPECT().GetScriptStorageRaw(
		gomock.Eq("KT1QaDvkDe1sLXGL9rqmDMtNCmvNyPfUTYWK"),
		gomock.Eq(int64(519894)),
	).Return(
		[]byte(`{"prim":"Pair","args":[{"prim":"Pair","args":[{"prim":"Pair","args":[{"int":"5"},{"int":"5000000"}]},{"string":"tz1cJywnhho2iGwfrs5gHCQs7stAVFMnRHc1"},{"int":"3"},{"int":"66048"}]},{"prim":"Pair","args":[{"int":"66049"},{"int":"3"}]},{"int":"66050"},{"prim":"False"},{"int":"66051"}]}`),
		nil,
	).AnyTimes()

	blocksRepo.EXPECT().GetNetworkAlias(
		"NetXz969SFaFn8k",
	).Return("granadanet", nil).AnyTimes()

	bmdRepo.EXPECT().Current(
		types.Granadanet, "expruc4MqoCyxFbogqrZumAraAzt3BXw7rZYeWkaXPLC27nfhMd7pt", int64(66049),
	).Return(bigmapdiff.BigMapState{
		Ptr:             66049,
		Key:             types.MustNewBytes("7b22737472696e67223a22746f6b656e5f325f6d65746164617461227d"),
		KeyHash:         "expruc4MqoCyxFbogqrZumAraAzt3BXw7rZYeWkaXPLC27nfhMd7pt",
		Value:           types.MustNewBytes("7b226279746573223a22376230613232363137323734363936363631363337343535373236393232336132303232363837343734373037333361326632663732363536653634363537323635373232653638333337303265363436353633366636653633363537303734326536333666366432663366373436663662363536653639363433643332323232633061323236343635363336393664363136633733323233613230333032633061323236333732363536313734366637323733323233613230356232323437363536663636363632303533373436353631373236653733323032323564326330613232363436353733363337323639373037343639366636653232336132303232343836313733363832303534363837323635363532303530366636393665373437333230363937333230363132303663366636653637326436363666373236643230363736353665363537323631373436393736363532303631373237343230363336663663366336353633373436393666366532303633366636653733363937333734363936653637323036663636323033313330333233343230373736663732366237333230366636653230373436383635323035343635376136663733323036323663366636333662363336383631363936653265323232633061323236353738373436353732366536313663353537323639323233613230323236383734373437303733336132663266363837303333326536343635363336663665363336353730373432653633366636643266323232633061323236363666373236643631373437333232336132303762306132323735373236393232336132303232363837343734373037333361326632663732363536653634363537323635373232653638333337303265363436353633366636653633363537303734326536333666366432663366373436663662363536653639363433643332323232633061323236643639366436353534373937303635323233613230323237343635373837343266363837343664366332323061376432633061323236393733343236663666366336353631366534313664366637353665373432323361323037343732373536353263306132323665363136643635323233613230323234383631373336383230353436383732363536353230353036663639366537343733323032333332323232633061323237343631363737333232336132303562323236373635366536353732363137343639373636353232326332303232363137323734323232633230323236633666366536373264363636663732366432323263323032323633366336393633366232323263323032323638363137333638323232633230323237343638373236353635323232633230323237303666363936653734373332323564326330613232373436663662363536653566363836313733363832323361323032323330373833313635363633323632363133303635363533343336363133323339363336333333333433373337363333353634363633383330333936333636333136313336333836363330333833313339363233393633363136343636363533393633333136333636333133373632363633343337333333343339333633323339363233323232326330613232373236393637363837343733323233613230323234633639363336353665373336353361323034333433323034323539326434653433323033343265333032323263306132323732366637393631366337343639363537333232336132303762306132323634363536333639366436313663373332323361323033323263306132323733363836313732363537333232336132303762306132323734376133313464333436623464353233333733353534323434373236613331333836643737363435363534373534623533333136363638333837353661353936653535333133313232336132303331333030613764306137643263306132323733373936643632366636633232336132303232343833333530323230613764227d"),
		LastUpdateLevel: 519894,
		Contract:        "KT1QaDvkDe1sLXGL9rqmDMtNCmvNyPfUTYWK",
		Network:         types.Granadanet,
		LastUpdateTime:  timestamp,
	}, nil).AnyTimes()

	tests := []struct {
		name       string
		sharePath  string
		network    types.Network
		bmd        *domains.BigMapDiff
		storageAST string
		want       []tokenmetadata.TokenMetadata
		wantErr    bool
	}{
		{
			name:      "Token metadata in tezos storage",
			sharePath: "./test",
			network:   types.Granadanet,
			bmd: &domains.BigMapDiff{
				BigMapDiff: &bigmapdiff.BigMapDiff{
					Ptr:         66051,
					Key:         types.MustNewBytes("7b22696e74223a2232227d"),
					KeyHash:     "expruDuAZnFKqmLoisJqUGqrNzXTvw7PJM2rYk97JErM5FHCerQqgn",
					Value:       types.MustNewBytes("7b227072696d223a2250616972222c2261726773223a5b7b22696e74223a2232227d2c5b7b227072696d223a22456c74222c2261726773223a5b7b22737472696e67223a22227d2c7b226279746573223a22373436353761366637333264373337343666373236313637363533613734366636623635366535663332356636643635373436313634363137343631227d5d7d2c7b227072696d223a22456c74222c2261726773223a5b7b22737472696e67223a22746f6b656e5f68617368227d2c7b226279746573223a2231656632626130656534366132396363333437376335646638303963663161363866303831396239636164666539633163663137626634373334393632396232227d5d7d5d5d7d"),
					Level:       519894,
					Contract:    "KT1QaDvkDe1sLXGL9rqmDMtNCmvNyPfUTYWK",
					Network:     types.Granadanet,
					Timestamp:   timestamp,
					ProtocolID:  17,
					OperationID: 27283107,
					KeyStrings:  []string{"tezos-storage:token_2_metadata", "token_hash"},
				},
				Operation: &operation.Operation{
					DeffatedStorage: types.MustNewBytes("5b5b7b227072696d223a2250616972222c2261726773223a5b7b22696e74223a2235227d2c7b22696e74223a2235303030303030227d5d7d2c7b226279746573223a223030303062366466663132306339656266326462333364653566666666623031643765383536616464383132227d2c7b22696e74223a2233227d2c7b22696e74223a223636303438227d5d2c7b227072696d223a2250616972222c2261726773223a5b7b22696e74223a223636303439227d2c7b22696e74223a2233227d5d7d2c7b22696e74223a223636303530227d2c7b227072696d223a2246616c7365227d2c7b22696e74223a223636303531227d5d"),
				},
				Protocol: &protocol.Protocol{
					Hash: "PtGRANADsDU8R9daYKAgWnQYAJ64omN1o3KMGVCykShA97vQbvV",
				},
			},
			storageAST: `{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%MAX_SUPPLY"]},{"prim":"mutez","annots":["%PURCHASE_PRICE_MUTEZ"]}]},{"prim":"pair","args":[{"prim":"address","annots":["%administrator"]},{"prim":"pair","args":[{"prim":"nat","annots":["%all_tokens"]},{"prim":"big_map","args":[{"prim":"pair","args":[{"prim":"address"},{"prim":"nat"}]},{"prim":"nat"}],"annots":["%ledger"]}]}]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"string"},{"prim":"bytes"}],"annots":["%metadata"]},{"prim":"nat","annots":["%next_id"]}]},{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"unit"}],"annots":["%operators"]},{"prim":"pair","args":[{"prim":"bool","annots":["%paused"]},{"prim":"big_map","args":[{"prim":"nat"},{"prim":"pair","args":[{"prim":"nat","annots":["%token_id"]},{"prim":"map","args":[{"prim":"string"},{"prim":"bytes"}],"annots":["%token_info"]}]}],"annots":["%token_metadata"]}]}]}]}]}`,
			want: []tokenmetadata.TokenMetadata{
				{
					Network:         types.Granadanet,
					Contract:        "KT1QaDvkDe1sLXGL9rqmDMtNCmvNyPfUTYWK",
					TokenID:         2,
					Level:           519894,
					Timestamp:       timestamp,
					Symbol:          "H3P",
					Decimals:        ptrInt64(0),
					Name:            "Hash Three Points #2",
					IsBooleanAmount: true,
					IsTransferable:  true,
					Description:     "Hash Three Points is a long-form generative art collection consisting of 1024 works on the Tezos blockchain.",
					ArtifactURI:     "https://renderer.h3p.deconcept.com/?tokenid=2",
					ExternalURI:     "https://hp3.deconcept.com/",
					Creators:        []string{"Geoff Stearns "},
					Formats:         []byte(`{"mimeType":"text/html","uri":"https://renderer.h3p.deconcept.com/?tokenid=2"}`),
					Tags: []string{
						"generative",
						"art",
						"long-form",
						"click",
						"hash",
						"three",
						"points",
					},
					Extras: map[string]interface{}{
						"@@empty":    "tezos-storage:token_2_metadata",
						"rights":     "License: CC BY-NC 4.0",
						"token_hash": "0x1ef2ba0ee46a29cc3477c5df809cf1a68f0819b9cadfe9c1cf17bf47349629b2",
						"royalties": map[string]interface{}{
							"decimals": float64(2),
							"shares": map[string]interface{}{
								"tz1M4kMR3sUBDrj18mwdVTuKS1fh8ujYnU11": float64(10),
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := Parser{
				bmdRepo:    bmdRepo,
				blocksRepo: blocksRepo,
				tmRepo:     tmRepo,
				storage:    generalRepo,
				rpc:        rpc,
				network:    tt.network,
				sharePath:  tt.sharePath,
			}

			storageAST, err := ast.NewTypedAstFromString(tt.storageAST)
			if err != nil {
				t.Errorf("NewTypedAstFromString() error = %v", err)
				return
			}
			got, err := parser.ParseBigMapDiff(tt.bmd, storageAST)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parser.ParseBigMapDiff() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
