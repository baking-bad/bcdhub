package operations

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/cache"
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/account"
	"github.com/baking-bad/bcdhub/internal/models/bigmapaction"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	modelContract "github.com/baking-bad/bcdhub/internal/models/contract"
	mock_general "github.com/baking-bad/bcdhub/internal/models/mock"
	mock_accounts "github.com/baking-bad/bcdhub/internal/models/mock/account"
	mock_bmd "github.com/baking-bad/bcdhub/internal/models/mock/bigmapdiff"
	mock_block "github.com/baking-bad/bcdhub/internal/models/mock/block"
	mock_contract "github.com/baking-bad/bcdhub/internal/models/mock/contract"
	mock_operations "github.com/baking-bad/bcdhub/internal/models/mock/operation"
	mock_proto "github.com/baking-bad/bcdhub/internal/models/mock/protocol"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/parsers"
	"github.com/go-pg/pg/v10"
	"github.com/golang/mock/gomock"
	"github.com/microcosm-cc/bluemonday"
)

func TestGroup_Parse(t *testing.T) {
	timestamp := time.Now()

	ctrlStorage := gomock.NewController(t)
	defer ctrlStorage.Finish()
	generalRepo := mock_general.NewMockGeneralRepository(ctrlStorage)

	ctrlBmdRepo := gomock.NewController(t)
	defer ctrlBmdRepo.Finish()
	bmdRepo := mock_bmd.NewMockRepository(ctrlBmdRepo)

	ctrlAccountsRepo := gomock.NewController(t)
	defer ctrlAccountsRepo.Finish()
	accountsRepo := mock_accounts.NewMockRepository(ctrlAccountsRepo)

	ctrlBlockRepo := gomock.NewController(t)
	defer ctrlBlockRepo.Finish()
	blockRepo := mock_block.NewMockRepository(ctrlBlockRepo)

	ctrlProtoRepo := gomock.NewController(t)
	defer ctrlProtoRepo.Finish()
	protoRepo := mock_proto.NewMockRepository(ctrlProtoRepo)

	ctrlContractRepo := gomock.NewController(t)
	defer ctrlContractRepo.Finish()
	contractRepo := mock_contract.NewMockRepository(ctrlContractRepo)

	ctrlScriptRepo := gomock.NewController(t)
	defer ctrlScriptRepo.Finish()
	scriptRepo := mock_contract.NewMockScriptRepository(ctrlScriptRepo)

	ctrlGlobalConstantRepo := gomock.NewController(t)
	defer ctrlGlobalConstantRepo.Finish()
	globalConstantRepo := mock_contract.NewMockConstantRepository(ctrlGlobalConstantRepo)

	ctrlOperationsRepo := gomock.NewController(t)
	defer ctrlOperationsRepo.Finish()
	operaitonsRepo := mock_operations.NewMockRepository(ctrlOperationsRepo)

	ctrlRPC := gomock.NewController(t)
	defer ctrlRPC.Finish()
	rpc := noderpc.NewMockINode(ctrlRPC)

	rpc.
		EXPECT().
		GetScriptJSON(gomock.Any(), "KT1PWx2mnDueood7fEmfbBDKx1D9BAnnXitn", int64(0)).
		DoAndReturn(readRPCScript).
		AnyTimes()
	rpc.
		EXPECT().
		GetScriptJSON(gomock.Any(), "KT1K9gCRgaLRFKTErYt1wVxA3Frb9FjasjTV", int64(0)).
		DoAndReturn(readRPCScript).
		AnyTimes()
	rpc.
		EXPECT().
		GetScriptJSON(gomock.Any(), "KT19at7rQUvyjxnZ2fBv7D9zc8rkyG7gAoU8", int64(0)).
		DoAndReturn(readRPCScript).
		AnyTimes()
	rpc.
		EXPECT().
		GetScriptJSON(gomock.Any(), "KT1AafHA1C1vk959wvHWBispY9Y2f3fxBUUo", int64(0)).
		DoAndReturn(readRPCScript).
		AnyTimes()
	rpc.
		EXPECT().
		GetScriptJSON(gomock.Any(), "KT1LN4LPSqTMS7Sd2CJw4bbDGRkMv2t68Fy9", int64(0)).
		DoAndReturn(readRPCScript).
		AnyTimes()
	rpc.
		EXPECT().
		GetScriptJSON(gomock.Any(), "KT1S95Dyj2QrJpSnAbHRUSUZr7DhuFqssrog", int64(0)).
		DoAndReturn(readRPCScript).
		AnyTimes()

	contractRepo.
		EXPECT().
		Get(gomock.Any()).
		DoAndReturn(readTestContractModel).
		AnyTimes()

	contractRepo.
		EXPECT().
		Script(gomock.Any(), gomock.Any()).
		DoAndReturn(readTestScriptModel).
		AnyTimes()

	contractRepo.
		EXPECT().
		ScriptPart(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(readTestScriptPart).
		AnyTimes()

	scriptRepo.
		EXPECT().
		ByHash(gomock.Any()).
		Return(modelContract.Script{}, pg.ErrNoRows).
		AnyTimes()

	globalConstantRepo.
		EXPECT().
		All(gomock.Eq("exprv5uiw7xXoEgRahR3YBn4iAVwfkNCMsrkneutuBZCGG5sS64kRw")).
		Return([]modelContract.GlobalConstant{
			{
				ID:        1,
				Timestamp: timestamp,
				Level:     707452,
				Address:   "exprv5uiw7xXoEgRahR3YBn4iAVwfkNCMsrkneutuBZCGG5sS64kRw",
				Value:     []byte(`[{"prim":"parameter","args":[{"prim":"or","args":[{"prim":"lambda","args":[{"prim":"unit"},{"prim":"list","args":[{"prim":"operation"}]}],"annots":["%lambda"]},{"prim":"option","args":[{"prim":"address"}],"annots":["%update_admin"]}]}]},{"prim":"storage","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address","annots":["%current"]},{"prim":"option","args":[{"prim":"address"}],"annots":["%pending"]}],"annots":["%admin"]},{"prim":"address","annots":["%whitelist_contract"]}]}]},{"prim":"code","args":[[{"prim":"UNPAIR"},{"prim":"IF_LEFT","args":[[{"prim":"SWAP"},{"prim":"DUP"},{"prim":"DUG","args":[{"int":"2"}]},{"prim":"CAR"},{"prim":"CAR"},{"prim":"SENDER"},{"prim":"COMPARE"},{"prim":"NEQ"},{"prim":"IF","args":[[{"prim":"PUSH","args":[{"prim":"string"},{"string":"SENDER_NOT_ADMIN"}]},{"prim":"FAILWITH"}],[]]},{"prim":"SWAP"},{"prim":"UNIT"},{"prim":"DIG","args":[{"int":"2"}]},{"prim":"SWAP"},{"prim":"EXEC"},{"prim":"PAIR"}],[{"prim":"SWAP"},{"prim":"DUP"},{"prim":"DUG","args":[{"int":"2"}]},{"prim":"CDR"},{"prim":"NIL","args":[{"prim":"address"}]},{"prim":"SENDER"},{"prim":"CONS"},{"prim":"VIEW","args":[{"string":"are_whitelisted"},{"prim":"bool"}]},{"prim":"IF_NONE","args":[[{"prim":"PUSH","args":[{"prim":"string"},{"string":"CALL_ARE_WHITELISED_VIEW_FAILED"}]},{"prim":"FAILWITH"}],[]]},{"prim":"IF","args":[[{"prim":"SWAP"},{"prim":"DUP"},{"prim":"DUG","args":[{"int":"2"}]},{"prim":"CAR"},{"prim":"SWAP"},{"prim":"IF_NONE","args":[[{"prim":"CDR"},{"prim":"IF_NONE","args":[[{"prim":"PUSH","args":[{"prim":"string"},{"string":"NO_PENDING_ADMIN"}]},{"prim":"FAILWITH"}],[{"prim":"DUP"},{"prim":"SENDER"},{"prim":"COMPARE"},{"prim":"NEQ"},{"prim":"IF","args":[[{"prim":"DROP"},{"prim":"PUSH","args":[{"prim":"string"},{"string":"NOT_PENDING_ADMIN"}]},{"prim":"FAILWITH"}],[{"prim":"NONE","args":[{"prim":"address"}]},{"prim":"SWAP"},{"prim":"PAIR"}]]}]]}],[{"prim":"SWAP"},{"prim":"DUP"},{"prim":"DUG","args":[{"int":"2"}]},{"prim":"CAR"},{"prim":"SENDER"},{"prim":"COMPARE"},{"prim":"NEQ"},{"prim":"IF","args":[[{"prim":"PUSH","args":[{"prim":"string"},{"string":"SENDER_NOT_ADMIN"}]},{"prim":"FAILWITH"}],[]]},{"prim":"SOME"},{"prim":"UPDATE","args":[{"int":"2"}]}]]},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"DUG","args":[{"int":"2"}]},{"prim":"UPDATE","args":[{"int":"1"}]},{"prim":"SWAP"},{"prim":"PAIR"}],[{"prim":"DROP","args":[{"int":"2"}]},{"prim":"PUSH","args":[{"prim":"string"},{"string":"ADDRESS_NOT_WHITELISTED"}]},{"prim":"FAILWITH"}]]}]]}]]},{"prim":"view","args":[{"string":"admin"},{"prim":"unit"},{"prim":"address"},[{"prim":"CDR"},{"prim":"CAR"},{"prim":"CAR"}]]}]`),
			},
		}, nil).
		AnyTimes()

	generalRepo.
		EXPECT().
		Save(context.Background(), gomock.AssignableToTypeOf([]models.Model{})).
		Return(nil).
		AnyTimes()

	generalRepo.
		EXPECT().
		IsRecordNotFound(gomock.Any()).
		Return(true).
		AnyTimes()

	bmdRepo.
		EXPECT().
		GetByPtr("KT1HBy1L43tiLe5MVJZ5RoxGy53Kx8kMgyoU", int64(2416)).
		Return([]bigmapdiff.BigMapState{
			{
				Ptr:             2416,
				Key:             []byte(`{"bytes": "000085ef0c18b31983603d978a152de4cd61803db881"}`),
				KeyHash:         "exprtfKNhZ1G8vMscchFjt1G1qww2P93VTLHMuhyThVYygZLdnRev2",
				Value:           []byte(`{"prim":"Pair","args":[[],{"int":"6000"}]}`),
				LastUpdateLevel: 386026,
				Contract:        "KT1HBy1L43tiLe5MVJZ5RoxGy53Kx8kMgyoU",
				LastUpdateTime:  timestamp,
			},
		}, nil).
		AnyTimes()

	for _, ptr := range []int{25167, 25166, 25165, 25164} {
		bmdRepo.
			EXPECT().
			GetByPtr("KT1C2MfcjWb5R1ZDDxVULCsGuxrf5fEn5264", int64(ptr)).
			Return([]bigmapdiff.BigMapState{}, nil).
			AnyTimes()
	}

	for _, ptr := range []int{40067, 40065} {
		bmdRepo.
			EXPECT().
			GetByPtr("KT1Jk8LRDoj6LkopYZwRq5ZEWBhYv8nVc6e6", int64(ptr)).
			Return([]bigmapdiff.BigMapState{}, nil).
			AnyTimes()
	}

	bmdRepo.
		EXPECT().
		GetByPtr("KT1Dc6A6jTY9sG4UvqKciqbJNAGtXqb4n7vZ", int64(2417)).
		Return([]bigmapdiff.BigMapState{
			{
				Ptr:             2417,
				Key:             []byte(`{"bytes": "000085ef0c18b31983603d978a152de4cd61803db881"}`),
				KeyHash:         "exprtfKNhZ1G8vMscchFjt1G1qww2P93VTLHMuhyThVYygZLdnRev2",
				Value:           nil,
				LastUpdateLevel: 386026,
				Contract:        "KT1Dc6A6jTY9sG4UvqKciqbJNAGtXqb4n7vZ",
				LastUpdateTime:  timestamp,
			},
		}, nil).
		AnyTimes()

	protoRepo.
		EXPECT().
		Get("PsDELPH1Kxsxt8f9eWbxQeRxkjfbxoqM52jvs5Y5fBxWWh4ifpo", int64(-1)).
		Return(protocol.Protocol{
			Hash:    "PsDELPH1Kxsxt8f9eWbxQeRxkjfbxoqM52jvs5Y5fBxWWh4ifpo",
			SymLink: bcd.SymLinkBabylon,
			ID:      0,
		}, nil).
		AnyTimes()

	protoRepo.
		EXPECT().
		Get("PsDELPH1Kxsxt8f9eWbxQeRxkjfbxoqM52jvs5Y5fBxWWh4ifpo", int64(-1)).
		Return(protocol.Protocol{
			Hash:    "PsDELPH1Kxsxt8f9eWbxQeRxkjfbxoqM52jvs5Y5fBxWWh4ifpo",
			SymLink: bcd.SymLinkBabylon,
			ID:      1,
		}, nil).
		AnyTimes()

	protoRepo.
		EXPECT().
		Get("PsddFKi32cMJ2qPjf43Qv5GDWLDPZb3T3bF6fLKiF5HtvHNU7aP", int64(-1)).
		Return(protocol.Protocol{
			Hash:    "PsddFKi32cMJ2qPjf43Qv5GDWLDPZb3T3bF6fLKiF5HtvHNU7aP",
			SymLink: bcd.SymLinkAlpha,
			ID:      2,
		}, nil).
		AnyTimes()

	protoRepo.
		EXPECT().
		Get("PtEdo2ZkT9oKpimTah6x2embF25oss54njMuPzkJTEi5RqfdZFA", int64(-1)).
		Return(protocol.Protocol{
			Hash:    "PtEdo2ZkT9oKpimTah6x2embF25oss54njMuPzkJTEi5RqfdZFA",
			SymLink: bcd.SymLinkBabylon,
			ID:      3,
		}, nil).
		AnyTimes()

	protoRepo.
		EXPECT().
		Get("PsFLorenaUUuikDWvMDr6fGBRG8kt3e3D3fHoXK1j1BFRxeSH4i", int64(-1)).
		Return(protocol.Protocol{
			Hash:    "PsFLorenaUUuikDWvMDr6fGBRG8kt3e3D3fHoXK1j1BFRxeSH4i",
			SymLink: bcd.SymLinkBabylon,
			ID:      4,
		}, nil).
		AnyTimes()

	protoRepo.
		EXPECT().
		Get("PtHangzHogokSuiMHemCuowEavgYTP8J5qQ9fQS793MHYFpCY3r", int64(-1)).
		Return(protocol.Protocol{
			Hash:    "PtHangzHogokSuiMHemCuowEavgYTP8J5qQ9fQS793MHYFpCY3r",
			SymLink: bcd.SymLinkBabylon,
			ID:      5,
		}, nil).
		AnyTimes()

	protoRepo.
		EXPECT().
		Get("Psithaca2MLRFYargivpo7YvUr7wUDqyxrdhC5CQq78mRvimz6A", int64(-1)).
		Return(protocol.Protocol{
			Hash:    "Psithaca2MLRFYargivpo7YvUr7wUDqyxrdhC5CQq78mRvimz6A",
			SymLink: bcd.SymLinkBabylon,
			ID:      6,
		}, nil).
		AnyTimes()

	protoRepo.
		EXPECT().
		GetByID(int64(0)).
		Return(protocol.Protocol{
			Hash:    "PsDELPH1Kxsxt8f9eWbxQeRxkjfbxoqM52jvs5Y5fBxWWh4ifpo",
			SymLink: bcd.SymLinkBabylon,
		}, nil).
		AnyTimes()

	protoRepo.
		EXPECT().
		GetByID(int64(1)).
		Return(protocol.Protocol{
			Hash:    "PsDELPH1Kxsxt8f9eWbxQeRxkjfbxoqM52jvs5Y5fBxWWh4ifpo",
			SymLink: bcd.SymLinkBabylon,
			ID:      1,
		}, nil).
		AnyTimes()

	protoRepo.
		EXPECT().
		GetByID(int64(2)).
		Return(protocol.Protocol{
			Hash:    "PsddFKi32cMJ2qPjf43Qv5GDWLDPZb3T3bF6fLKiF5HtvHNU7aP",
			SymLink: bcd.SymLinkAlpha,
			ID:      2,
		}, nil).
		AnyTimes()

	protoRepo.
		EXPECT().
		GetByID(int64(3)).
		Return(protocol.Protocol{
			Hash:    "PtEdo2ZkT9oKpimTah6x2embF25oss54njMuPzkJTEi5RqfdZFA",
			SymLink: bcd.SymLinkBabylon,
			ID:      3,
		}, nil).
		AnyTimes()

	protoRepo.
		EXPECT().
		GetByID(int64(4)).
		Return(protocol.Protocol{
			Hash:    "PsFLorenaUUuikDWvMDr6fGBRG8kt3e3D3fHoXK1j1BFRxeSH4i",
			SymLink: bcd.SymLinkBabylon,
			ID:      4,
		}, nil).
		AnyTimes()

	protoRepo.
		EXPECT().
		GetByID(int64(5)).
		Return(protocol.Protocol{
			Hash:    "PtHangzHogokSuiMHemCuowEavgYTP8J5qQ9fQS793MHYFpCY3r",
			SymLink: bcd.SymLinkBabylon,
			ID:      5,
		}, nil).
		AnyTimes()

	protoRepo.
		EXPECT().
		GetByID(int64(6)).
		Return(protocol.Protocol{
			Hash:    "Psithaca2MLRFYargivpo7YvUr7wUDqyxrdhC5CQq78mRvimz6A",
			SymLink: bcd.SymLinkBabylon,
			ID:      6,
		}, nil).
		AnyTimes()

	accountsRepo.
		EXPECT().
		Get("KT1C2MfcjWb5R1ZDDxVULCsGuxrf5fEn5264").
		Return(account.Account{
			Address: "KT1C2MfcjWb5R1ZDDxVULCsGuxrf5fEn5264",
			Type:    types.AccountTypeContract,
			ID:      6,
		}, nil).
		AnyTimes()

	operaitonsRepo.
		EXPECT().
		Last(gomock.Any(), int64(0)).
		Return(operation.Operation{
			Status:          types.OperationStatusApplied,
			DestinationID:   6,
			DeffatedStorage: []byte(`{"prim":"Pair","args":[{"string":"tz1gXhGAXgKvrXjn4t16rYUXocqbch1XXJFN"},{"prim":"Right","args":[{"prim":"Pair","args":[{"bytes":"050200001531051f02000002e807430765076507650765076003680369075e076507650765076503620760036803690765036e036207650765036b0362055f0765036e03620765076507650765036e076003680369076507610765036e03620362076103680369076507650764036c0764036e036e07610765036e036e036c0765036e036207650765076505660765036b03690761036907650765076503620760036803690765036e036207650765036b0362055f0765036e036207650362036e03620765055f036d0765076507650765036e076003680369076507610765036e03620362076103680369076507650764036c0764036e036e07610765036e036e036c0765036e036207650765076505660765036b03690761036907650765076503620760036803690765036e036207650765036b0362055f0765036e036207650362036e036207650362036207650765036203620765036203620765075e0765076503620760036803690765076507650765036e076003680369076507610765036e03620362076103680369076507650764036c0764036e036e07610765036e036e036c0765036e036207650765076505660765036b03690761036907650765076503620760036803690765036e036207650765036b0362055f0765036e036207650362036e03620359075e076507650765076503620760036803690765036e036207650765036b0362055f0765036e03620765076507650765036e076003680369076507610765036e03620362076103680369076507650764036c0764036e036e07610765036e036e036c0765036e036207650765076505660765036b03690761036907650765076503620760036803690765036e036207650765036b0362055f0765036e036207650362036e0362036207070707070707070200000000020000001a020000000d03210316051f020000000203170320053d036d0342070700b40700a80f0707070700a80f0080b4bc0207070001000107070200000008032007430359030a0200000008032007430362000003210316051f02000000020317051f0200000050051f020000000607430362000b051f0200000027074307650362036e070700020a00000016000038bb193df0965b3a87badd3600f294493b5cd608074305660765036b0369020000000003420342034203210316051f02000000020317051f0200000042051f020000002707430765036e036207070a00000016000038bb193df0965b3a87badd3600f294493b5cd608000007430764036c0764036e036e0505030b0342034203210316051f02000000020317051f0200000000034207430765036e07600368036907070a00000016000038bb193df0965b3a87badd3600f294493b5cd608020000000003420342034203420743036a0000053e035d051d020000111f0500076407640865036803690000000b2563616c6c437573746f6d0764076407640764046c00000011256163636570745f6f776e6572736869700865046e000000062566726f6d5f076504620000000925746f6b656e5f696404620000000725616d6f756e7400000005256275726e0764086407640865065f0765046e00000006256f776e657204620000000925746f6b656e5f696400000009257265717565737473065a055f07650865046e00000006256f776e657204620000000925746f6b656e5f69640000000825726571756573740462000000082562616c616e6365000000092563616c6c6261636b0000000b2562616c616e63655f6f66065f0765046e000000062566726f6d5f065f0765046e0000000425746f5f076504620000000925746f6b656e5f696404620000000725616d6f756e74000000042574787300000009257472616e73666572065f07640865046e00000006256f776e65720765046e00000009256f70657261746f7204620000000925746f6b656e5f69640000000d256164645f6f70657261746f720865046e00000006256f776e65720765046e00000009256f70657261746f7204620000000925746f6b656e5f6964000000102572656d6f76655f6f70657261746f7200000011257570646174655f6f70657261746f7273000000092563616c6c5f464132046c0000001225636f6e6669726d5f6d6967726174696f6e0764076404690000000e2564726f705f70726f706f73616c04620000000625666c75736807640865046c0000000625706172616d065a0362000000092563616c6c6261636b0000001525676574566f74655065726d6974436f756e746572046e00000008256d6967726174650764076407640865046e0000000425746f5f076504620000000925746f6b656e5f696404620000000725616d6f756e7400000005256d696e74086504620000000d2566726f7a656e5f746f6b656e086003680369000000122570726f706f73616c5f6d65746164617461000000082570726f706f73650764046200000015257365745f71756f72756d5f7468726573686f6c64046200000012257365745f766f74696e675f706572696f640764046e00000013257472616e736665725f6f776e657273686970065f0765086504690000000d2570726f706f73616c5f6b6579076504590000000a25766f74655f7479706504620000000c25766f74655f616d6f756e740000000925617267756d656e7406630765045c00000004256b657904670000000a257369676e617475726500000007257065726d69740000000525766f74650865046e0000001125636f6e74726163745f61646472657373065f0765046e000000062566726f6d5f065f0765046e0000000425746f5f076504620000000925746f6b656e5f696404620000000725616d6f756e7400000004257478730000000725706172616d7300000019257472616e736665725f636f6e74726163745f746f6b656e73050107650765076507650765046e000000062561646d696e08600368036900000006256578747261076508610765036e0362036200000007256c656467657208610368036900000009256d65746164617461076507650864046c00000011256e6f745f696e5f6d6967726174696f6e0764046e0000000c256d6967726174696e67546f046e0000000b256d69677261746564546f00000011256d6967726174696f6e5f73746174757308610765046e00000006256f776e6572046e00000009256f70657261746f72036c0000000a256f70657261746f72730765046e0000000e2570656e64696e675f6f776e6572046200000010257065726d6974735f636f756e74657207650765076506660765036b03690000001f2570726f706f73616c5f6b65795f6c6973745f736f72745f62795f646174650861036907650765076504620000000a25646f776e766f74657308600368036900000009256d657461646174610765046e000000092570726f706f7365720462000000162570726f706f7365725f66726f7a656e5f746f6b656e07650765046b0000000b2573746172745f64617465046200000008257570766f746573065f0765036e03620000000725766f746572730000000a2570726f706f73616c7307650462000000112571756f72756d5f7468726573686f6c64046e0000000e25746f6b656e5f6164647265737304620000000e25766f74696e675f706572696f6407650765076507650860036803690000001325637573746f6d5f656e747279706f696e7473085e076507650765076504620000000a25646f776e766f74657308600368036900000009256d657461646174610765046e000000092570726f706f7365720462000000162570726f706f7365725f66726f7a656e5f746f6b656e07650765046b0000000b2573746172745f64617465046200000008257570766f746573065f0765036e03620000000725766f746572730765076507650765046e000000062561646d696e08600368036900000006256578747261076508610765036e0362036200000007256c656467657208610368036900000009256d65746164617461076507650864046c00000011256e6f745f696e5f6d6967726174696f6e0764046e0000000c256d6967726174696e67546f046e0000000b256d69677261746564546f00000011256d6967726174696f6e5f73746174757308610765046e00000006256f776e6572046e00000009256f70657261746f72036c0000000a256f70657261746f72730765046e0000000e2570656e64696e675f6f776e6572046200000010257065726d6974735f636f756e74657207650765076506660765036b03690000001f2570726f706f73616c5f6b65795f6c6973745f736f72745f62795f646174650861036907650765076504620000000a25646f776e766f74657308600368036900000009256d657461646174610765046e000000092570726f706f7365720462000000162570726f706f7365725f66726f7a656e5f746f6b656e07650765046b0000000b2573746172745f64617465046200000008257570766f746573065f0765036e03620000000725766f746572730000000a2570726f706f73616c7307650462000000112571756f72756d5f7468726573686f6c64046e0000000e25746f6b656e5f6164647265737304620000000e25766f74696e675f706572696f640765055f036d0765076507650765046e000000062561646d696e08600368036900000006256578747261076508610765036e0362036200000007256c656467657208610368036900000009256d65746164617461076507650864046c00000011256e6f745f696e5f6d6967726174696f6e0764046e0000000c256d6967726174696e67546f046e0000000b256d69677261746564546f00000011256d6967726174696f6e5f73746174757308610765046e00000006256f776e6572046e00000009256f70657261746f72036c0000000a256f70657261746f72730765046e0000000e2570656e64696e675f6f776e6572046200000010257065726d6974735f636f756e74657207650765076506660765036b03690000001f2570726f706f73616c5f6b65795f6c6973745f736f72745f62795f646174650861036907650765076504620000000a25646f776e766f74657308600368036900000009256d657461646174610765046e000000092570726f706f7365720462000000162570726f706f7365725f66726f7a656e5f746f6b656e07650765046b0000000b2573746172745f64617465046200000008257570766f746573065f0765036e03620000000725766f746572730000000a2570726f706f73616c7307650462000000112571756f72756d5f7468726573686f6c64046e0000000e25746f6b656e5f6164647265737304620000000e25766f74696e675f706572696f6400000010256465636973696f6e5f6c616d626461076504620000000e256d61785f70726f706f73616c73046200000015256d61785f71756f72756d5f7468726573686f6c640765076504620000000a256d61785f766f746573046200000012256d61785f766f74696e675f706572696f640765046200000015256d696e5f71756f72756d5f7468726573686f6c64046200000012256d696e5f766f74696e675f706572696f640765085e0765076504620000000d2566726f7a656e5f746f6b656e086003680369000000122570726f706f73616c5f6d657461646174610765076507650765046e000000062561646d696e08600368036900000006256578747261076508610765036e0362036200000007256c656467657208610368036900000009256d65746164617461076507650864046c00000011256e6f745f696e5f6d6967726174696f6e0764046e0000000c256d6967726174696e67546f046e0000000b256d69677261746564546f00000011256d6967726174696f6e5f73746174757308610765046e00000006256f776e6572046e00000009256f70657261746f72036c0000000a256f70657261746f72730765046e0000000e2570656e64696e675f6f776e6572046200000010257065726d6974735f636f756e74657207650765076506660765036b03690000001f2570726f706f73616c5f6b65795f6c6973745f736f72745f62795f646174650861036907650765076504620000000a25646f776e766f74657308600368036900000009256d657461646174610765046e000000092570726f706f7365720462000000162570726f706f7365725f66726f7a656e5f746f6b656e07650765046b0000000b2573746172745f64617465046200000008257570766f746573065f0765036e03620000000725766f746572730000000a2570726f706f73616c7307650462000000112571756f72756d5f7468726573686f6c64046e0000000e25746f6b656e5f6164647265737304620000000e25766f74696e675f706572696f6403590000000f2570726f706f73616c5f636865636b085e076507650765076504620000000a25646f776e766f74657308600368036900000009256d657461646174610765046e000000092570726f706f7365720462000000162570726f706f7365725f66726f7a656e5f746f6b656e07650765046b0000000b2573746172745f64617465046200000008257570766f746573065f0765036e03620000000725766f746572730765076507650765046e000000062561646d696e08600368036900000006256578747261076508610765036e0362036200000007256c656467657208610368036900000009256d65746164617461076507650864046c00000011256e6f745f696e5f6d6967726174696f6e0764046e0000000c256d6967726174696e67546f046e0000000b256d69677261746564546f00000011256d6967726174696f6e5f73746174757308610765046e00000006256f776e6572046e00000009256f70657261746f72036c0000000a256f70657261746f72730765046e0000000e2570656e64696e675f6f776e6572046200000010257065726d6974735f636f756e74657207650765076506660765036b03690000001f2570726f706f73616c5f6b65795f6c6973745f736f72745f62795f646174650861036907650765076504620000000a25646f776e766f74657308600368036900000009256d657461646174610765046e000000092570726f706f7365720462000000162570726f706f7365725f66726f7a656e5f746f6b656e07650765046b0000000b2573746172745f64617465046200000008257570766f746573065f0765036e03620000000725766f746572730000000a2570726f706f73616c7307650462000000112571756f72756d5f7468726573686f6c64046e0000000e25746f6b656e5f6164647265737304620000000e25766f74696e675f706572696f6403620000001f2572656a65637465645f70726f706f73616c5f72657475726e5f76616c7565050202000000230743036801000000184641325f494e53554646494349454e545f42414c414e43450327053d036d034c031b034c0342"},{"prim":"Pair","args":[{"prim":"Pair","args":[{"int":"25164"},{"int":"25165"}]},{"int":"25166"}]},{"int":"25167"}]}]}]}`),
		}, nil).
		AnyTimes()

	tests := []struct {
		name       string
		rpc        noderpc.INode
		ctx        *config.Context
		paramsOpts []ParseParamsOption
		filename   string
		storage    map[string]int64
		want       *parsers.TestStore
		wantErr    bool
	}{
		{
			name: "opToHHcqFhRTQWJv2oTGAtywucj9KM1nDnk5eHsEETYJyvJLsa5",
			ctx: &config.Context{
				RPC:         rpc,
				Storage:     generalRepo,
				Contracts:   contractRepo,
				BigMapDiffs: bmdRepo,
				Blocks:      blockRepo,
				Protocols:   protoRepo,
				Operations:  operaitonsRepo,
				Scripts:     scriptRepo,
				Cache: cache.NewCache(
					rpc, accountsRepo, contractRepo, protoRepo, bluemonday.UGCPolicy(),
				),
			},
			paramsOpts: []ParseParamsOption{
				WithHead(noderpc.Header{
					Timestamp: timestamp,
					Protocol:  "PsDELPH1Kxsxt8f9eWbxQeRxkjfbxoqM52jvs5Y5fBxWWh4ifpo",
					Level:     1068669,
					ChainID:   "NetXdQprcVkpaWU",
				}),
				WithProtocol(&protocol.Protocol{
					Constants: &protocol.Constants{
						CostPerByte:                  1000,
						HardGasLimitPerOperation:     400000,
						HardStorageLimitPerOperation: 60000,
						TimeBetweenBlocks:            60,
					},
					Hash:    "PsDELPH1Kxsxt8f9eWbxQeRxkjfbxoqM52jvs5Y5fBxWWh4ifpo",
					SymLink: bcd.SymLinkBabylon,
				}),
			},
			filename: "./data/rpc/opg/opToHHcqFhRTQWJv2oTGAtywucj9KM1nDnk5eHsEETYJyvJLsa5.json",
			want:     parsers.NewTestStore(),
		}, {
			name: "opJXaAMkBrAbd1XFd23kS8vXiw63tU4rLUcLrZgqUCpCbhT1Pn9",
			ctx: &config.Context{
				RPC:         rpc,
				Storage:     generalRepo,
				Contracts:   contractRepo,
				Accounts:    accountsRepo,
				BigMapDiffs: bmdRepo,
				Blocks:      blockRepo,
				Protocols:   protoRepo,
				Operations:  operaitonsRepo,
				Scripts:     scriptRepo,
				Cache: cache.NewCache(
					rpc, accountsRepo, contractRepo, protoRepo, bluemonday.UGCPolicy(),
				),
			},
			paramsOpts: []ParseParamsOption{
				WithHead(noderpc.Header{
					Timestamp: timestamp,
					Protocol:  "PsDELPH1Kxsxt8f9eWbxQeRxkjfbxoqM52jvs5Y5fBxWWh4ifpo",
					Level:     1068669,
					ChainID:   "test",
				}),
				WithProtocol(&protocol.Protocol{
					Constants: &protocol.Constants{
						CostPerByte:                  1000,
						HardGasLimitPerOperation:     1040000,
						HardStorageLimitPerOperation: 60000,
						TimeBetweenBlocks:            60,
					},
					Hash:    "PsDELPH1Kxsxt8f9eWbxQeRxkjfbxoqM52jvs5Y5fBxWWh4ifpo",
					ID:      1,
					SymLink: bcd.SymLinkBabylon,
				}),
			},
			storage: map[string]int64{
				"KT1S5iPRQ612wcNm6mXDqDhTNegGFcvTV7vM": 1068668,
				"KT19nHqEWZxFFbbDL1b7Y86escgEN7qUShGo": 1068668,
				"KT1KemKUx79keZgFW756jQrqKcZJ21y4SPdS": 1068668,
			},
			filename: "./data/rpc/opg/opJXaAMkBrAbd1XFd23kS8vXiw63tU4rLUcLrZgqUCpCbhT1Pn9.json",
			want: &parsers.TestStore{
				Operations: []*operation.Operation{
					{
						Kind: types.OperationKindTransaction,
						Source: account.Account{
							Address: "tz1aSPEN4RTZbn4aXEsxDiix38dDmacGQ8sq",
							Type:    types.AccountTypeTz,
						},
						Fee:          37300,
						Counter:      5791164,
						GasLimit:     369423,
						StorageLimit: 90,
						Destination: account.Account{
							Address: "KT1S5iPRQ612wcNm6mXDqDhTNegGFcvTV7vM",
							Type:    types.AccountTypeContract,
						},
						Delegate: account.Account{},
						Status:   types.OperationStatusApplied,
						Level:    1068669,
						Hash:     "opJXaAMkBrAbd1XFd23kS8vXiw63tU4rLUcLrZgqUCpCbhT1Pn9",
						Entrypoint: types.NullString{
							Str:   "transfer",
							Valid: true,
						},
						Timestamp: timestamp,
						Burned:    70000,
						Initiator: account.Account{
							Address: "tz1aSPEN4RTZbn4aXEsxDiix38dDmacGQ8sq",
							Type:    types.AccountTypeTz,
						},
						ProtocolID:      1,
						Parameters:      []byte("{\"entrypoint\":\"default\",\"value\":{\"prim\":\"Right\",\"args\":[{\"prim\":\"Left\",\"args\":[{\"prim\":\"Right\",\"args\":[{\"prim\":\"Right\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"string\":\"tz1aSPEN4RTZbn4aXEsxDiix38dDmacGQ8sq\"},{\"prim\":\"Pair\",\"args\":[{\"string\":\"tz1invbJv3AEm55ct7QF2dVbWZuaDekssYkV\"},{\"int\":\"8010000\"}]}]}]}]}]}]}}"),
						DeffatedStorage: []byte("{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[[{\"bytes\":\"000056d8b91b541c9d20d51f929dcccca2f14928f1dc\"}],{\"int\":\"62\"}]},{\"prim\":\"Pair\",\"args\":[{\"int\":\"63\"},{\"string\":\"Aspen Digital Token\"}]}]},{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"prim\":\"False\"},{\"bytes\":\"0000a2560a416161def96031630886abe950c4baf036\"}]},{\"prim\":\"Pair\",\"args\":[{\"prim\":\"False\"},{\"bytes\":\"010d25f77b84dc2164a5d1ce5e8a5d3ca2b1d0cbf900\"}]}]}]},{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"bytes\":\"01796ad78734892d5ae4186e84a30290040732ada700\"},{\"string\":\"ASPD\"}]},{\"int\":\"18000000\"}]}]}"),
						Tags:            types.FA12Tag,
						BigMapDiffs: []*bigmapdiff.BigMapDiff{
							{
								Ptr:        63,
								KeyHash:    "exprum2qtFLPHdeLWVasKCDw7YD5MrdiD4ra52PY2AUazaNGKyv6tx",
								Key:        []byte(`{"bytes":"0000a2560a416161def96031630886abe950c4baf036"}`),
								Value:      []byte(`{"int":"6141000"}`),
								Level:      1068669,
								Contract:   "KT1S5iPRQ612wcNm6mXDqDhTNegGFcvTV7vM",
								ProtocolID: 1,
								Timestamp:  timestamp,
							}, {
								Ptr:        63,
								KeyHash:    "exprv2snyFbF6EDZd2YAHnnmNBoFt7bbaXhGSWGXHv4a4wnxS359ob",
								Key:        []byte(`{"bytes":"0000fdf98b65d53a9661e07f41093dcb6f3d931736ba"}`),
								Value:      []byte(`{"int":"8010000"}`),
								Level:      1068669,
								Contract:   "KT1S5iPRQ612wcNm6mXDqDhTNegGFcvTV7vM",
								ProtocolID: 1,
								Timestamp:  timestamp,
							},
						},
					}, {
						Kind: types.OperationKindTransaction,
						Source: account.Account{
							Address: "KT1S5iPRQ612wcNm6mXDqDhTNegGFcvTV7vM",
							Type:    types.AccountTypeContract,
						},
						Destination: account.Account{
							Address: "KT19nHqEWZxFFbbDL1b7Y86escgEN7qUShGo",
							Type:    types.AccountTypeContract,
						},
						Delegate: account.Account{},
						Status:   types.OperationStatusApplied,
						Level:    1068669,
						Counter:  5791164,
						Hash:     "opJXaAMkBrAbd1XFd23kS8vXiw63tU4rLUcLrZgqUCpCbhT1Pn9",
						Nonce:    getInt64Pointer(0),
						Entrypoint: types.NullString{
							Str:   "validateAccounts",
							Valid: true,
						},
						Internal:   true,
						Timestamp:  timestamp,
						ProtocolID: 1,
						Initiator: account.Account{
							Address: "tz1aSPEN4RTZbn4aXEsxDiix38dDmacGQ8sq",
							Type:    types.AccountTypeTz,
						},
						Parameters:      []byte("{\"entrypoint\":\"validateAccounts\",\"value\":{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"bytes\":\"0000a2560a416161def96031630886abe950c4baf036\"},{\"bytes\":\"0000fdf98b65d53a9661e07f41093dcb6f3d931736ba\"}]},{\"prim\":\"Pair\",\"args\":[{\"int\":\"14151000\"},{\"int\":\"0\"}]}]},{\"prim\":\"Pair\",\"args\":[{\"prim\":\"True\"},{\"prim\":\"Pair\",\"args\":[{\"int\":\"8010000\"},{\"int\":\"18000000\"}]}]}]},{\"bytes\":\"01796ad78734892d5ae4186e84a30290040732ada70076616c696461746552756c6573\"}]}}"),
						DeffatedStorage: []byte("{\"int\":\"61\"}"),
					}, {
						Kind: types.OperationKindTransaction,
						Source: account.Account{
							Address: "KT19nHqEWZxFFbbDL1b7Y86escgEN7qUShGo",
							Type:    types.AccountTypeContract,
						},
						Destination: account.Account{
							Address: "KT1KemKUx79keZgFW756jQrqKcZJ21y4SPdS",
							Type:    types.AccountTypeContract,
						},
						Delegate: account.Account{},
						Status:   types.OperationStatusApplied,
						Level:    1068669,
						Counter:  5791164,
						Hash:     "opJXaAMkBrAbd1XFd23kS8vXiw63tU4rLUcLrZgqUCpCbhT1Pn9",
						Nonce:    getInt64Pointer(1),
						Entrypoint: types.NullString{
							Str:   "validateRules",
							Valid: true,
						},
						Internal:   true,
						Timestamp:  timestamp,
						ProtocolID: 1,
						Initiator: account.Account{
							Address: "tz1aSPEN4RTZbn4aXEsxDiix38dDmacGQ8sq",
							Type:    types.AccountTypeTz,
						},
						Parameters:      []byte("{\"entrypoint\":\"validateRules\",\"value\":{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"prim\":\"None\"},{\"string\":\"US\"}]},{\"prim\":\"Pair\",\"args\":[{\"prim\":\"False\"},{\"bytes\":\"000056d8b91b541c9d20d51f929dcccca2f14928f1dc\"}]}]},{\"int\":\"2\"}]},{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"prim\":\"None\"},{\"string\":\"US\"}]},{\"prim\":\"Pair\",\"args\":[{\"prim\":\"False\"},{\"bytes\":\"0000c644b537bdb0dac40fe742010106546effd69395\"}]}]},{\"int\":\"6\"}]}]},{\"prim\":\"Pair\",\"args\":[{\"bytes\":\"0000a2560a416161def96031630886abe950c4baf036\"},{\"bytes\":\"0000fdf98b65d53a9661e07f41093dcb6f3d931736ba\"}]}]},{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"int\":\"14151000\"},{\"int\":\"0\"}]},{\"prim\":\"True\"}]}]},{\"prim\":\"Pair\",\"args\":[{\"bytes\":\"01bff38c4e363eacef338f7b2e15f00ca42fafa1ce00\"},{\"prim\":\"Pair\",\"args\":[{\"int\":\"8010000\"},{\"int\":\"18000000\"}]}]}]}}"),
						DeffatedStorage: []byte("{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"bytes\":\"000056d8b91b541c9d20d51f929dcccca2f14928f1dc\"},{\"bytes\":\"010d25f77b84dc2164a5d1ce5e8a5d3ca2b1d0cbf900\"}]},[]]}"),
					},
				},
				BigMapState: []*bigmapdiff.BigMapState{
					{
						Ptr:             63,
						KeyHash:         "exprum2qtFLPHdeLWVasKCDw7YD5MrdiD4ra52PY2AUazaNGKyv6tx",
						Key:             []byte(`{"bytes":"0000a2560a416161def96031630886abe950c4baf036"}`),
						Value:           []byte(`{"int":"6141000"}`),
						Contract:        "KT1S5iPRQ612wcNm6mXDqDhTNegGFcvTV7vM",
						LastUpdateLevel: 1068669,
						LastUpdateTime:  timestamp,
					}, {
						Ptr:             63,
						KeyHash:         "exprv2snyFbF6EDZd2YAHnnmNBoFt7bbaXhGSWGXHv4a4wnxS359ob",
						Key:             []byte(`{"bytes":"0000fdf98b65d53a9661e07f41093dcb6f3d931736ba"}`),
						Value:           []byte(`{"int":"8010000"}`),
						Contract:        "KT1S5iPRQ612wcNm6mXDqDhTNegGFcvTV7vM",
						LastUpdateLevel: 1068669,
						LastUpdateTime:  timestamp,
					},
				},
			},
		}, {
			name: "opPUPCpQu6pP38z9TkgFfwLiqVBFGSWQCH8Z2PUL3jrpxqJH5gt",
			ctx: &config.Context{
				RPC:         rpc,
				Storage:     generalRepo,
				Contracts:   contractRepo,
				Accounts:    accountsRepo,
				BigMapDiffs: bmdRepo,
				Blocks:      blockRepo,
				Protocols:   protoRepo,
				Operations:  operaitonsRepo,
				Scripts:     scriptRepo,
				Cache: cache.NewCache(
					rpc, accountsRepo, contractRepo, protoRepo, bluemonday.UGCPolicy(),
				),
			},
			paramsOpts: []ParseParamsOption{
				WithHead(noderpc.Header{
					Timestamp: timestamp,
					Protocol:  "PsDELPH1Kxsxt8f9eWbxQeRxkjfbxoqM52jvs5Y5fBxWWh4ifpo",
					Level:     1151495,
					ChainID:   "test",
				}),
				WithProtocol(&protocol.Protocol{
					Constants: &protocol.Constants{
						CostPerByte:                  1000,
						HardGasLimitPerOperation:     1040000,
						HardStorageLimitPerOperation: 60000,
						TimeBetweenBlocks:            60,
					},
					Hash:    "PsDELPH1Kxsxt8f9eWbxQeRxkjfbxoqM52jvs5Y5fBxWWh4ifpo",
					ID:      1,
					SymLink: bcd.SymLinkBabylon,
				}),
			},
			storage: map[string]int64{
				"KT1Ap287P1NzsnToSJdA4aqSNjPomRaHBZSr": 1151494,
				"KT1PWx2mnDueood7fEmfbBDKx1D9BAnnXitn": 1151494,
			},
			filename: "./data/rpc/opg/opPUPCpQu6pP38z9TkgFfwLiqVBFGSWQCH8Z2PUL3jrpxqJH5gt.json",
			want: &parsers.TestStore{
				Operations: []*operation.Operation{
					{
						ContentIndex: 0,
						ProtocolID:   1,
						Hash:         "opPUPCpQu6pP38z9TkgFfwLiqVBFGSWQCH8Z2PUL3jrpxqJH5gt",
						Internal:     false,
						Nonce:        nil,
						Status:       types.OperationStatusApplied,
						Timestamp:    timestamp,
						Level:        1151495,
						Kind:         types.OperationKindTransaction,
						Initiator: account.Account{
							Address: "tz1dMH7tW7RhdvVMR4wKVFF1Ke8m8ZDvrTTE",
							Type:    types.AccountTypeTz,
						},
						Source: account.Account{
							Address: "tz1dMH7tW7RhdvVMR4wKVFF1Ke8m8ZDvrTTE",
							Type:    types.AccountTypeTz,
						},
						Delegate:     account.Account{},
						Fee:          43074,
						Counter:      6909186,
						GasLimit:     427673,
						StorageLimit: 47,
						Destination: account.Account{
							Address: "KT1Ap287P1NzsnToSJdA4aqSNjPomRaHBZSr",
							Type:    types.AccountTypeContract,
						},
						Parameters: []byte("{\"entrypoint\":\"redeem\",\"value\":{\"bytes\":\"a874aac22777351417c9bde0920cc7ed33e54453e1dd149a1f3a60521358d19a\"}}"),
						Entrypoint: types.NullString{
							Str:   "redeem",
							Valid: true,
						},
						DeffatedStorage: []byte("{\"prim\":\"Pair\",\"args\":[{\"int\":\"32\"},{\"prim\":\"Unit\"}]}"),
						BigMapDiffs: []*bigmapdiff.BigMapDiff{
							{
								Ptr:        32,
								Key:        []byte(`{"bytes": "80729e85e284dff3a30bb24a58b37ccdf474bbbe7794aad439ba034f48d66af3"}`),
								KeyHash:    "exprvJp4s8RJpoXMwD9aQujxWQUiojrkeubesi3X9LDcU3taDfahYR",
								Level:      1151495,
								Contract:   "KT1Ap287P1NzsnToSJdA4aqSNjPomRaHBZSr",
								Timestamp:  timestamp,
								ProtocolID: 1,
							},
						},
					}, {
						ContentIndex: 0,
						ProtocolID:   1,
						Hash:         "opPUPCpQu6pP38z9TkgFfwLiqVBFGSWQCH8Z2PUL3jrpxqJH5gt",
						Internal:     true,
						Nonce:        getInt64Pointer(0),
						Status:       types.OperationStatusApplied,
						Timestamp:    timestamp,
						Level:        1151495,
						Kind:         types.OperationKindTransaction,
						Initiator: account.Account{
							Address: "tz1dMH7tW7RhdvVMR4wKVFF1Ke8m8ZDvrTTE",
							Type:    types.AccountTypeTz,
						},
						Source: account.Account{
							Address: "KT1Ap287P1NzsnToSJdA4aqSNjPomRaHBZSr",
							Type:    types.AccountTypeContract,
						},
						Counter: 6909186,
						Destination: account.Account{
							Address: "KT1PWx2mnDueood7fEmfbBDKx1D9BAnnXitn",
							Type:    types.AccountTypeContract,
						},
						Delegate:   account.Account{},
						Parameters: []byte("{\"entrypoint\":\"transfer\",\"value\":{\"prim\":\"Pair\",\"args\":[{\"bytes\":\"011871cfab6dafee00330602b4342b6500c874c93b00\"},{\"prim\":\"Pair\",\"args\":[{\"bytes\":\"0000c2473c617946ce7b9f6843f193401203851cb2ec\"},{\"int\":\"7874880\"}]}]}}"),
						Entrypoint: types.NullString{
							Str:   "transfer",
							Valid: true,
						},
						Burned:          47000,
						DeffatedStorage: []byte("{\"prim\":\"Pair\",\"args\":[{\"int\":\"31\"},{\"prim\":\"Pair\",\"args\":[[{\"prim\":\"DUP\"},{\"prim\":\"CAR\"},{\"prim\":\"DIP\",\"args\":[[{\"prim\":\"CDR\"}]]},{\"prim\":\"DUP\"},{\"prim\":\"DUP\"},{\"prim\":\"CAR\"},{\"prim\":\"DIP\",\"args\":[[{\"prim\":\"CDR\"}]]},{\"prim\":\"DIP\",\"args\":[[{\"prim\":\"DIP\",\"args\":[{\"int\":\"2\"},[{\"prim\":\"DUP\"}]]},{\"prim\":\"DIG\",\"args\":[{\"int\":\"2\"}]}]]},{\"prim\":\"PUSH\",\"args\":[{\"prim\":\"string\"},{\"string\":\"code\"}]},{\"prim\":\"PAIR\"},{\"prim\":\"PACK\"},{\"prim\":\"GET\"},{\"prim\":\"IF_NONE\",\"args\":[[{\"prim\":\"NONE\",\"args\":[{\"prim\":\"lambda\",\"args\":[{\"prim\":\"pair\",\"args\":[{\"prim\":\"bytes\"},{\"prim\":\"big_map\",\"args\":[{\"prim\":\"bytes\"},{\"prim\":\"bytes\"}]}]},{\"prim\":\"pair\",\"args\":[{\"prim\":\"list\",\"args\":[{\"prim\":\"operation\"}]},{\"prim\":\"big_map\",\"args\":[{\"prim\":\"bytes\"},{\"prim\":\"bytes\"}]}]}]}]}],[{\"prim\":\"UNPACK\",\"args\":[{\"prim\":\"lambda\",\"args\":[{\"prim\":\"pair\",\"args\":[{\"prim\":\"bytes\"},{\"prim\":\"big_map\",\"args\":[{\"prim\":\"bytes\"},{\"prim\":\"bytes\"}]}]},{\"prim\":\"pair\",\"args\":[{\"prim\":\"list\",\"args\":[{\"prim\":\"operation\"}]},{\"prim\":\"big_map\",\"args\":[{\"prim\":\"bytes\"},{\"prim\":\"bytes\"}]}]}]}]},{\"prim\":\"IF_NONE\",\"args\":[[{\"prim\":\"PUSH\",\"args\":[{\"prim\":\"string\"},{\"string\":\"UStore: failed to unpack code\"}]},{\"prim\":\"FAILWITH\"}],[]]},{\"prim\":\"SOME\"}]]},{\"prim\":\"IF_NONE\",\"args\":[[{\"prim\":\"DROP\"},{\"prim\":\"DIP\",\"args\":[[{\"prim\":\"DUP\"},{\"prim\":\"PUSH\",\"args\":[{\"prim\":\"bytes\"},{\"bytes\":\"05010000000866616c6c6261636b\"}]},{\"prim\":\"GET\"},{\"prim\":\"IF_NONE\",\"args\":[[{\"prim\":\"PUSH\",\"args\":[{\"prim\":\"string\"},{\"string\":\"UStore: no field fallback\"}]},{\"prim\":\"FAILWITH\"}],[]]},{\"prim\":\"UNPACK\",\"args\":[{\"prim\":\"lambda\",\"args\":[{\"prim\":\"pair\",\"args\":[{\"prim\":\"pair\",\"args\":[{\"prim\":\"string\"},{\"prim\":\"bytes\"}]},{\"prim\":\"big_map\",\"args\":[{\"prim\":\"bytes\"},{\"prim\":\"bytes\"}]}]},{\"prim\":\"pair\",\"args\":[{\"prim\":\"list\",\"args\":[{\"prim\":\"operation\"}]},{\"prim\":\"big_map\",\"args\":[{\"prim\":\"bytes\"},{\"prim\":\"bytes\"}]}]}]}]},{\"prim\":\"IF_NONE\",\"args\":[[{\"prim\":\"PUSH\",\"args\":[{\"prim\":\"string\"},{\"string\":\"UStore: failed to unpack fallback\"}]},{\"prim\":\"FAILWITH\"}],[]]},{\"prim\":\"SWAP\"}]]},{\"prim\":\"PAIR\"},{\"prim\":\"EXEC\"}],[{\"prim\":\"DIP\",\"args\":[[{\"prim\":\"SWAP\"},{\"prim\":\"DROP\"},{\"prim\":\"PAIR\"}]]},{\"prim\":\"SWAP\"},{\"prim\":\"EXEC\"}]]}],{\"prim\":\"Pair\",\"args\":[{\"int\":\"1\"},{\"prim\":\"False\"}]}]}]}"),
						Tags:            types.FA12Tag,
						BigMapDiffs: []*bigmapdiff.BigMapDiff{
							{
								Ptr:        31,
								Key:        []byte(`{"bytes":"05010000000b746f74616c537570706c79"}`),
								KeyHash:    "exprunzteC5uyXRHbKnqJd3hUMGTWE9Gv5EtovDZHnuqu6SaGViV3N",
								Value:      []byte(`{"bytes":"050098e1e8d78a02"}`),
								Level:      1151495,
								Contract:   "KT1PWx2mnDueood7fEmfbBDKx1D9BAnnXitn",
								Timestamp:  timestamp,
								ProtocolID: 1,
							}, {
								Ptr:        31,
								Key:        []byte(`{"bytes":"05070701000000066c65646765720a000000160000c2473c617946ce7b9f6843f193401203851cb2ec"}`),
								KeyHash:    "exprv9xaiXBb9KBi67dQoP1SchDyZeKEz3XHiFwBCtHadiKS8wkX7w",
								Value:      []byte(`{"bytes":"0507070080a5c1070200000000"}`),
								Level:      1151495,
								Contract:   "KT1PWx2mnDueood7fEmfbBDKx1D9BAnnXitn",
								Timestamp:  timestamp,
								ProtocolID: 1,
							}, {
								Ptr:        31,
								Key:        []byte(`{"bytes":"05070701000000066c65646765720a00000016011871cfab6dafee00330602b4342b6500c874c93b00"}`),
								KeyHash:    "expruiWsykU9wjNb4aV7eJULLBpGLhy1EuzgD8zB8k7eUTaCk16fyV",
								Value:      []byte(`{"bytes":"05070700ba81bb090200000000"}`),
								Level:      1151495,
								Contract:   "KT1PWx2mnDueood7fEmfbBDKx1D9BAnnXitn",
								Timestamp:  timestamp,
								ProtocolID: 1,
							},
						},
					},
				},
				BigMapState: []*bigmapdiff.BigMapState{
					{
						Ptr:             32,
						Key:             []byte(`{"bytes":"80729e85e284dff3a30bb24a58b37ccdf474bbbe7794aad439ba034f48d66af3"}`),
						KeyHash:         "exprvJp4s8RJpoXMwD9aQujxWQUiojrkeubesi3X9LDcU3taDfahYR",
						Contract:        "KT1Ap287P1NzsnToSJdA4aqSNjPomRaHBZSr",
						Removed:         true,
						LastUpdateLevel: 1151495,
						LastUpdateTime:  timestamp,
					}, {
						Ptr:             31,
						Key:             []byte(`{"bytes":"05010000000b746f74616c537570706c79"}`),
						KeyHash:         "exprunzteC5uyXRHbKnqJd3hUMGTWE9Gv5EtovDZHnuqu6SaGViV3N",
						Value:           []byte(`{"bytes":"050098e1e8d78a02"}`),
						Contract:        "KT1PWx2mnDueood7fEmfbBDKx1D9BAnnXitn",
						LastUpdateLevel: 1151495,
						LastUpdateTime:  timestamp,
					}, {
						Ptr:             31,
						Key:             []byte(`{"bytes":"05070701000000066c65646765720a000000160000c2473c617946ce7b9f6843f193401203851cb2ec"}`),
						KeyHash:         "exprv9xaiXBb9KBi67dQoP1SchDyZeKEz3XHiFwBCtHadiKS8wkX7w",
						Value:           []byte(`{"bytes":"0507070080a5c1070200000000"}`),
						Contract:        "KT1PWx2mnDueood7fEmfbBDKx1D9BAnnXitn",
						LastUpdateLevel: 1151495,
						LastUpdateTime:  timestamp,
					}, {
						Ptr:      31,
						Key:      []byte(`{"bytes":"05070701000000066c65646765720a00000016011871cfab6dafee00330602b4342b6500c874c93b00"}`),
						KeyHash:  "expruiWsykU9wjNb4aV7eJULLBpGLhy1EuzgD8zB8k7eUTaCk16fyV",
						Value:    []byte(`{"bytes":"05070700ba81bb090200000000"}`),
						Contract: "KT1PWx2mnDueood7fEmfbBDKx1D9BAnnXitn",

						LastUpdateLevel: 1151495,
						LastUpdateTime:  timestamp,
					},
				},
			},
		}, {
			name: "onzUDQhwunz2yqzfEsoURXEBz9p7Gk8DgY4QBva52Z4b3AJCZjt",
			ctx: &config.Context{
				RPC:         rpc,
				Storage:     generalRepo,
				Contracts:   contractRepo,
				BigMapDiffs: bmdRepo,
				Blocks:      blockRepo,
				Protocols:   protoRepo,
				Operations:  operaitonsRepo,
				Scripts:     scriptRepo,
				Cache: cache.NewCache(
					rpc, accountsRepo, contractRepo, protoRepo, bluemonday.UGCPolicy(),
				),
			},
			paramsOpts: []ParseParamsOption{
				WithHead(noderpc.Header{
					Timestamp: timestamp,
					Protocol:  "PsDELPH1Kxsxt8f9eWbxQeRxkjfbxoqM52jvs5Y5fBxWWh4ifpo",
					Level:     86142,
					ChainID:   "test",
				}),
				WithProtocol(&protocol.Protocol{
					Constants: &protocol.Constants{
						CostPerByte:                  250,
						HardGasLimitPerOperation:     1040000,
						HardStorageLimitPerOperation: 60000,
						TimeBetweenBlocks:            30,
					},
					Hash:    "PsDELPH1Kxsxt8f9eWbxQeRxkjfbxoqM52jvs5Y5fBxWWh4ifpo",
					SymLink: bcd.SymLinkBabylon,
				}),
			},
			storage: map[string]int64{
				"KT1NppzrgyLZD3aku7fssfhYPm5QqZwyabvR": 86142,
			},
			filename: "./data/rpc/opg/onzUDQhwunz2yqzfEsoURXEBz9p7Gk8DgY4QBva52Z4b3AJCZjt.json",
			want: &parsers.TestStore{
				Operations: []*operation.Operation{
					{
						ContentIndex: 0,
						ProtocolID:   0,
						Hash:         "onzUDQhwunz2yqzfEsoURXEBz9p7Gk8DgY4QBva52Z4b3AJCZjt",
						Internal:     false,
						Status:       types.OperationStatusApplied,
						Timestamp:    timestamp,
						Level:        86142,
						Kind:         types.OperationKindOrigination,
						Initiator: account.Account{
							Address: "tz1SX7SPdx4ZJb6uP5Hh5XBVZhh9wTfFaud3",
							Type:    types.AccountTypeTz,
						},
						Source: account.Account{
							Address: "tz1SX7SPdx4ZJb6uP5Hh5XBVZhh9wTfFaud3",
							Type:    types.AccountTypeTz,
						},
						Fee:          510,
						Counter:      654594,
						GasLimit:     1870,
						StorageLimit: 371,
						Amount:       0,
						Destination: account.Account{
							Address: "KT1NppzrgyLZD3aku7fssfhYPm5QqZwyabvR",
							Type:    types.AccountTypeContract,
						},
						Delegate:                           account.Account{},
						Burned:                             87750,
						AllocatedDestinationContractBurned: 64250,
						DeffatedStorage:                    []byte("{\"int\":\"0\"}\n"),
					},
				},
				Contracts: []*modelContract.Contract{
					{
						Level:     86142,
						Timestamp: timestamp,
						Account: account.Account{
							Address: "KT1NppzrgyLZD3aku7fssfhYPm5QqZwyabvR",
							Type:    types.AccountTypeContract,
						},
						Manager: account.Account{
							Address: "tz1SX7SPdx4ZJb6uP5Hh5XBVZhh9wTfFaud3",
							Type:    types.AccountTypeTz,
						},
						Delegate: account.Account{},
						Babylon: modelContract.Script{
							Entrypoints: []string{"decrement", "increment"},
							Annotations: []string{"%decrement", "%increment"},
							Hash:        "97a40c7ff3bad5edb92c8e1dcfd4bfc778da8166a7632c1bcecbf8d8f9e4490b",
							Code:        []byte(`[[{"prim":"DUP"},{"prim":"CDR"},{"prim":"SWAP"},{"prim":"CAR"},{"prim":"IF_LEFT","args":[[{"prim":"SWAP"},{"prim":"SUB"}],[{"prim":"ADD"}]]},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]`),
							Parameter:   []byte(`[{"prim":"or","args":[{"prim":"int","annots":["%decrement"]},{"prim":"int","annots":["%increment"]}]}]`),
							Storage:     []byte(`[{"prim":"int"}]`),
						},
					},
				},
			},
		}, {
			name: "onv6Q1dNejAGEJeQzwRannWsDSGw85FuFdhLnBrY18TBcC9p8kC",
			ctx: &config.Context{
				RPC:         rpc,
				Storage:     generalRepo,
				Contracts:   contractRepo,
				BigMapDiffs: bmdRepo,
				Blocks:      blockRepo,
				Protocols:   protoRepo,
				Operations:  operaitonsRepo,
				Scripts:     scriptRepo,
				Cache: cache.NewCache(
					rpc, accountsRepo, contractRepo, protoRepo, bluemonday.UGCPolicy(),
				),
			},
			paramsOpts: []ParseParamsOption{
				WithHead(noderpc.Header{
					Timestamp: timestamp,
					Protocol:  "PsddFKi32cMJ2qPjf43Qv5GDWLDPZb3T3bF6fLKiF5HtvHNU7aP",
					Level:     301436,
					ChainID:   "test",
				}),
				WithProtocol(&protocol.Protocol{
					Constants: &protocol.Constants{
						CostPerByte:                  1000,
						HardGasLimitPerOperation:     400000,
						HardStorageLimitPerOperation: 60000,
						TimeBetweenBlocks:            60,
					},
					Hash: "PsddFKi32cMJ2qPjf43Qv5GDWLDPZb3T3bF6fLKiF5HtvHNU7aP",

					SymLink: bcd.SymLinkAlpha,
					ID:      2,
				}),
			},
			storage: map[string]int64{
				"KT1AbjG7vtpV8osdoJXcMRck8eTwst8dWoz4": 301436,
			},
			filename: "./data/rpc/opg/onv6Q1dNejAGEJeQzwRannWsDSGw85FuFdhLnBrY18TBcC9p8kC.json",
			want: &parsers.TestStore{
				Operations: []*operation.Operation{
					{
						Kind: types.OperationKindOrigination,
						Source: account.Account{

							Address: "tz1MXrEgDNnR8PDryN8sq4B2m9Pqcf57wBqM",
							Type:    types.AccountTypeTz,
						},
						Fee:          1555,
						Counter:      983250,
						GasLimit:     12251,
						StorageLimit: 351,
						Destination: account.Account{

							Address: "KT1AbjG7vtpV8osdoJXcMRck8eTwst8dWoz4",
							Type:    types.AccountTypeContract,
						},
						Delegate: account.Account{},
						Status:   types.OperationStatusApplied,
						Level:    301436,

						Hash:      "onv6Q1dNejAGEJeQzwRannWsDSGw85FuFdhLnBrY18TBcC9p8kC",
						Timestamp: timestamp,
						Burned:    331000,
						Initiator: account.Account{

							Address: "tz1MXrEgDNnR8PDryN8sq4B2m9Pqcf57wBqM",
							Type:    types.AccountTypeTz,
						},
						ProtocolID:                         2,
						DeffatedStorage:                    []byte("[]"),
						AllocatedDestinationContractBurned: 257000,
					},
				},
				Contracts: []*modelContract.Contract{
					{

						Level:     301436,
						Timestamp: timestamp,
						Account: account.Account{

							Address: "KT1AbjG7vtpV8osdoJXcMRck8eTwst8dWoz4",
							Type:    types.AccountTypeContract,
						},
						Manager: account.Account{

							Address: "tz1MXrEgDNnR8PDryN8sq4B2m9Pqcf57wBqM",
							Type:    types.AccountTypeTz,
						},
						Delegate: account.Account{},
						Alpha: modelContract.Script{
							Hash:        "c4915a55dbe0a3dfc8feb77e46f3e32828f80730a506fab277d8d6c0d5e2f1ec",
							Tags:        types.Tags(0),
							Entrypoints: []string{"default"},
							Code:        []byte(`[[[[{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]}]],{"prim":"CONS"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]`),
							Parameter:   []byte(`[{"prim":"pair","args":[{"prim":"string"},{"prim":"nat"}]}]`),
							Storage:     []byte(`[{"prim":"list","args":[{"prim":"pair","args":[{"prim":"string"},{"prim":"nat"}]}]}]`),
						},
					},
				},
			},
		}, {
			name: "op4fFMvYsxvSUKZmLWC7aUf25VMYqigaDwTZCAoBBi8zACbHTNg",
			ctx: &config.Context{
				RPC:         rpc,
				Accounts:    accountsRepo,
				Storage:     generalRepo,
				Contracts:   contractRepo,
				BigMapDiffs: bmdRepo,
				Blocks:      blockRepo,
				Protocols:   protoRepo,
				Operations:  operaitonsRepo,
				Scripts:     scriptRepo,
				Cache: cache.NewCache(
					rpc, accountsRepo, contractRepo, protoRepo, bluemonday.UGCPolicy(),
				),
			},
			paramsOpts: []ParseParamsOption{
				WithHead(noderpc.Header{
					Timestamp: timestamp,
					Protocol:  "PtEdo2ZkT9oKpimTah6x2embF25oss54njMuPzkJTEi5RqfdZFA",
					Level:     72207,
					ChainID:   "test",
				}),
				WithProtocol(&protocol.Protocol{
					Constants: &protocol.Constants{
						CostPerByte:                  1000,
						HardGasLimitPerOperation:     400000,
						HardStorageLimitPerOperation: 60000,
						TimeBetweenBlocks:            60,
					},
					Hash: "PtEdo2ZkT9oKpimTah6x2embF25oss54njMuPzkJTEi5RqfdZFA",

					SymLink: bcd.SymLinkBabylon,
					ID:      3,
				}),
			},
			storage: map[string]int64{
				"KT1C2MfcjWb5R1ZDDxVULCsGuxrf5fEn5264": 72206,
				"KT1JgHoXtZPjVfG82BY3FSys2VJhKVZo2EJU": 72207,
			},
			filename: "./data/rpc/opg/op4fFMvYsxvSUKZmLWC7aUf25VMYqigaDwTZCAoBBi8zACbHTNg.json",
			want: &parsers.TestStore{
				Operations: []*operation.Operation{
					{
						Kind: types.OperationKindTransaction,
						Source: account.Account{
							Address: "tz1gXhGAXgKvrXjn4t16rYUXocqbch1XXJFN",
							Type:    types.AccountTypeTz,
						},
						Fee:          4045,
						Counter:      155670,
						GasLimit:     37831,
						StorageLimit: 5265,
						Destination: account.Account{
							Address: "KT1C2MfcjWb5R1ZDDxVULCsGuxrf5fEn5264",
							Type:    types.AccountTypeContract,
						},
						Status:    types.OperationStatusApplied,
						Level:     72207,
						Hash:      "op4fFMvYsxvSUKZmLWC7aUf25VMYqigaDwTZCAoBBi8zACbHTNg",
						Timestamp: timestamp,
						Entrypoint: types.NullString{
							Str:   "@entrypoint_1",
							Valid: true,
						},
						Initiator: account.Account{
							Address: "tz1gXhGAXgKvrXjn4t16rYUXocqbch1XXJFN",
							Type:    types.AccountTypeTz,
						},
						Delegate:        account.Account{},
						Parameters:      []byte("{\"entrypoint\":\"default\",\"value\":{\"prim\":\"Right\",\"args\":[{\"prim\":\"Unit\"}]}}"),
						ProtocolID:      3,
						DeffatedStorage: []byte("{\"prim\":\"Pair\",\"args\":[{\"bytes\":\"0000e527ed176ccf8f8297f674a9886a2ba8a55818d9\"},{\"prim\":\"Left\",\"args\":[{\"bytes\":\"016ebc941b2ae4e305470f392fa050e41ca1e52b4500\"}]}]}"),
						BigMapActions: []*bigmapaction.BigMapAction{
							{
								Action:    types.BigMapActionRemove,
								SourcePtr: getInt64Pointer(25167),
								Level:     72207,
								Address:   "KT1C2MfcjWb5R1ZDDxVULCsGuxrf5fEn5264",
								Timestamp: timestamp,
							}, {
								Action:    types.BigMapActionRemove,
								SourcePtr: getInt64Pointer(25166),
								Level:     72207,
								Address:   "KT1C2MfcjWb5R1ZDDxVULCsGuxrf5fEn5264",
								Timestamp: timestamp,
							}, {
								Action:    types.BigMapActionRemove,
								SourcePtr: getInt64Pointer(25165),
								Level:     72207,
								Address:   "KT1C2MfcjWb5R1ZDDxVULCsGuxrf5fEn5264",
								Timestamp: timestamp,
							}, {
								Action:    types.BigMapActionRemove,
								SourcePtr: getInt64Pointer(25164),
								Level:     72207,
								Address:   "KT1C2MfcjWb5R1ZDDxVULCsGuxrf5fEn5264",
								Timestamp: timestamp,
							},
						},
					}, {
						Kind: types.OperationKindOrigination,
						Source: account.Account{
							Address: "KT1C2MfcjWb5R1ZDDxVULCsGuxrf5fEn5264",
							Type:    types.AccountTypeContract,
						},
						Nonce: getInt64Pointer(0),
						Destination: account.Account{
							Address: "KT1JgHoXtZPjVfG82BY3FSys2VJhKVZo2EJU",
							Type:    types.AccountTypeContract,
						},
						Status:    types.OperationStatusApplied,
						Level:     72207,
						Hash:      "op4fFMvYsxvSUKZmLWC7aUf25VMYqigaDwTZCAoBBi8zACbHTNg",
						Timestamp: timestamp,
						Burned:    5245000,
						Counter:   155670,
						Internal:  true,
						Initiator: account.Account{
							Address: "tz1gXhGAXgKvrXjn4t16rYUXocqbch1XXJFN",
							Type:    types.AccountTypeTz,
						},
						Delegate:                           account.Account{},
						ProtocolID:                         3,
						AllocatedDestinationContractBurned: 257000,
						Tags:                               types.LedgerTag | types.FA2Tag,
						DeffatedStorage:                    []byte("{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"string\":\"tz1QozfhaUW4wLnohDo6yiBUmh7cPCSXE9Af\"},[]]},{\"int\":\"25168\"},{\"int\":\"25169\"}]},{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Left\",\"args\":[{\"prim\":\"Unit\"}]},{\"int\":\"25170\"}]},{\"string\":\"tz1QozfhaUW4wLnohDo6yiBUmh7cPCSXE9Af\"},{\"int\":\"0\"}]},{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[[],{\"int\":\"25171\"}]},{\"int\":\"2\"},{\"string\":\"tz1QozfhaUW4wLnohDo6yiBUmh7cPCSXE9Af\"}]},{\"int\":\"11\"}]},{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[[],[[{\"prim\":\"DUP\"},{\"prim\":\"CAR\"},{\"prim\":\"DIP\",\"args\":[[{\"prim\":\"CDR\"}]]}],{\"prim\":\"DROP\"},{\"prim\":\"NIL\",\"args\":[{\"prim\":\"operation\"}]},{\"prim\":\"PAIR\"}]]},{\"int\":\"500\"},{\"int\":\"1000\"}]},{\"prim\":\"Pair\",\"args\":[{\"int\":\"1000\"},{\"int\":\"2592000\"}]},{\"int\":\"1\"},{\"int\":\"1\"}]},[{\"prim\":\"DROP\"},{\"prim\":\"PUSH\",\"args\":[{\"prim\":\"bool\"},{\"prim\":\"True\"}]}],[{\"prim\":\"DROP\"},{\"prim\":\"PUSH\",\"args\":[{\"prim\":\"nat\"},{\"int\":\"0\"}]}]]}"),
						BigMapActions: []*bigmapaction.BigMapAction{
							{
								Action:         types.BigMapActionCopy,
								SourcePtr:      getInt64Pointer(25167),
								DestinationPtr: getInt64Pointer(25171),
								Level:          72207,
								Address:        "KT1JgHoXtZPjVfG82BY3FSys2VJhKVZo2EJU",
								Timestamp:      timestamp,
							}, {
								Action:         types.BigMapActionCopy,
								SourcePtr:      getInt64Pointer(25166),
								DestinationPtr: getInt64Pointer(25170),
								Level:          72207,
								Address:        "KT1JgHoXtZPjVfG82BY3FSys2VJhKVZo2EJU",
								Timestamp:      timestamp,
							}, {
								Action:         types.BigMapActionCopy,
								SourcePtr:      getInt64Pointer(25165),
								DestinationPtr: getInt64Pointer(25169),
								Level:          72207,
								Address:        "KT1JgHoXtZPjVfG82BY3FSys2VJhKVZo2EJU",
								Timestamp:      timestamp,
							}, {
								Action:         types.BigMapActionCopy,
								SourcePtr:      getInt64Pointer(25164),
								DestinationPtr: getInt64Pointer(25168),
								Level:          72207,
								Address:        "KT1JgHoXtZPjVfG82BY3FSys2VJhKVZo2EJU",
								Timestamp:      timestamp,
							},
						},
					},
				},
				Contracts: []*modelContract.Contract{
					{
						Level:     72207,
						Timestamp: timestamp,
						Account: account.Account{
							Address: "KT1JgHoXtZPjVfG82BY3FSys2VJhKVZo2EJU",
							Type:    types.AccountTypeContract,
						},
						Manager: account.Account{
							Address: "KT1C2MfcjWb5R1ZDDxVULCsGuxrf5fEn5264",
							Type:    types.AccountTypeContract,
						},
						Delegate: account.Account{},
						Babylon: modelContract.Script{
							Hash:        "b82a20d0647f5ec74ef2daf404cd365a894f6868da0cd623ed07c6b85977b8db",
							Tags:        types.LedgerTag | types.FA2Tag,
							FailStrings: []string{"FA2_INSUFFICIENT_BALANCE"},
							Annotations: []string{"%token_address", "%drop_proposal", "%transfer_contract_tokens", "%permits_counter", "%remove_operator", "%mint", "%ledger", "%voters", "%owner", "%balance", "%transfer", "%from_", "%max_voting_period", "%not_in_migration", "%start_date", "%custom_entrypoints", "%proposal_check", "%accept_ownership", "%migrate", "%set_quorum_threshold", "%amount", "%proposals", "%min_voting_period", "%rejected_proposal_return_value", "%burn", "%flush", "%max_quorum_threshold", "%migratingTo", "%operators", "%proposer", "%call_FA2", "%argument", "%params", "%transfer_ownership", "%voting_period", "%request", "%confirm_migration", "%frozen_token", "%param", "%admin", "%migration_status", "%proposal_key_list_sort_by_date", "%requests", "%update_operators", "%add_operator", "%getVotePermitCounter", "%propose", "%vote", "%vote_amount", "%proposer_frozen_token", "%callCustom", "%txs", "%operator", "%quorum_threshold", "%to_", "%set_voting_period", "%callback", "%contract_address", "%downvotes", "%max_votes", "%balance_of", "%proposal_key", "%vote_type", "%signature", "%decision_lambda", "%token_id", "%permit", "%key", "%extra", "%pending_owner", "%upvotes", "%max_proposals", "%min_quorum_threshold", "%proposal_metadata", "%metadata", "%migratedTo"},
							Entrypoints: []string{"callCustom", "accept_ownership", "burn", "balance_of", "transfer", "update_operators", "confirm_migration", "drop_proposal", "flush", "getVotePermitCounter", "migrate", "mint", "propose", "set_quorum_threshold", "set_voting_period", "transfer_ownership", "vote", "transfer_contract_tokens"},
							Code:        []byte(`[[{"prim":"PUSH","args":[{"prim":"string"},{"string":"FA2_INSUFFICIENT_BALANCE"}]},{"prim":"FAILWITH"}]]`),
							Parameter:   []byte(`[{"prim":"or","args":[{"prim":"or","args":[{"prim":"pair","args":[{"prim":"string"},{"prim":"bytes"}],"annots":["%callCustom"]},{"prim":"or","args":[{"prim":"or","args":[{"prim":"or","args":[{"prim":"or","args":[{"prim":"unit","annots":["%accept_ownership"]},{"prim":"pair","args":[{"prim":"address","annots":["%from_"]},{"prim":"nat","annots":["%token_id"]},{"prim":"nat","annots":["%amount"]}],"annots":["%burn"]}]},{"prim":"or","args":[{"prim":"or","args":[{"prim":"or","args":[{"prim":"pair","args":[{"prim":"list","args":[{"prim":"pair","args":[{"prim":"address","annots":["%owner"]},{"prim":"nat","annots":["%token_id"]}]}],"annots":["%requests"]},{"prim":"contract","args":[{"prim":"list","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address","annots":["%owner"]},{"prim":"nat","annots":["%token_id"]}],"annots":["%request"]},{"prim":"nat","annots":["%balance"]}]}]}],"annots":["%callback"]}],"annots":["%balance_of"]},{"prim":"list","args":[{"prim":"pair","args":[{"prim":"address","annots":["%from_"]},{"prim":"list","args":[{"prim":"pair","args":[{"prim":"address","annots":["%to_"]},{"prim":"nat","annots":["%token_id"]},{"prim":"nat","annots":["%amount"]}]}],"annots":["%txs"]}]}],"annots":["%transfer"]}]},{"prim":"list","args":[{"prim":"or","args":[{"prim":"pair","args":[{"prim":"address","annots":["%owner"]},{"prim":"address","annots":["%operator"]},{"prim":"nat","annots":["%token_id"]}],"annots":["%add_operator"]},{"prim":"pair","args":[{"prim":"address","annots":["%owner"]},{"prim":"address","annots":["%operator"]},{"prim":"nat","annots":["%token_id"]}],"annots":["%remove_operator"]}]}],"annots":["%update_operators"]}],"annots":["%call_FA2"]},{"prim":"unit","annots":["%confirm_migration"]}]}]},{"prim":"or","args":[{"prim":"or","args":[{"prim":"bytes","annots":["%drop_proposal"]},{"prim":"nat","annots":["%flush"]}]},{"prim":"or","args":[{"prim":"pair","args":[{"prim":"unit","annots":["%param"]},{"prim":"contract","args":[{"prim":"nat"}],"annots":["%callback"]}],"annots":["%getVotePermitCounter"]},{"prim":"address","annots":["%migrate"]}]}]}]},{"prim":"or","args":[{"prim":"or","args":[{"prim":"or","args":[{"prim":"pair","args":[{"prim":"address","annots":["%to_"]},{"prim":"nat","annots":["%token_id"]},{"prim":"nat","annots":["%amount"]}],"annots":["%mint"]},{"prim":"pair","args":[{"prim":"nat","annots":["%frozen_token"]},{"prim":"map","args":[{"prim":"string"},{"prim":"bytes"}],"annots":["%proposal_metadata"]}],"annots":["%propose"]}]},{"prim":"or","args":[{"prim":"nat","annots":["%set_quorum_threshold"]},{"prim":"nat","annots":["%set_voting_period"]}]}]},{"prim":"or","args":[{"prim":"address","annots":["%transfer_ownership"]},{"prim":"list","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"bytes","annots":["%proposal_key"]},{"prim":"bool","annots":["%vote_type"]},{"prim":"nat","annots":["%vote_amount"]}],"annots":["%argument"]},{"prim":"option","args":[{"prim":"pair","args":[{"prim":"key","annots":["%key"]},{"prim":"signature","annots":["%signature"]}]}],"annots":["%permit"]}]}],"annots":["%vote"]}]}]}]}]},{"prim":"pair","args":[{"prim":"address","annots":["%contract_address"]},{"prim":"list","args":[{"prim":"pair","args":[{"prim":"address","annots":["%from_"]},{"prim":"list","args":[{"prim":"pair","args":[{"prim":"address","annots":["%to_"]},{"prim":"nat","annots":["%token_id"]},{"prim":"nat","annots":["%amount"]}]}],"annots":["%txs"]}]}],"annots":["%params"]}],"annots":["%transfer_contract_tokens"]}]}]}]`),
							Storage:     []byte(`[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address","annots":["%admin"]},{"prim":"map","args":[{"prim":"string"},{"prim":"bytes"}],"annots":["%extra"]}]},{"prim":"big_map","args":[{"prim":"pair","args":[{"prim":"address"},{"prim":"nat"}]},{"prim":"nat"}],"annots":["%ledger"]},{"prim":"big_map","args":[{"prim":"string"},{"prim":"bytes"}],"annots":["%metadata"]}]},{"prim":"pair","args":[{"prim":"or","args":[{"prim":"unit","annots":["%not_in_migration"]},{"prim":"or","args":[{"prim":"address","annots":["%migratingTo"]},{"prim":"address","annots":["%migratedTo"]}]}],"annots":["%migration_status"]},{"prim":"big_map","args":[{"prim":"pair","args":[{"prim":"address","annots":["%owner"]},{"prim":"address","annots":["%operator"]}]},{"prim":"unit"}],"annots":["%operators"]}]},{"prim":"address","annots":["%pending_owner"]},{"prim":"nat","annots":["%permits_counter"]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"set","args":[{"prim":"pair","args":[{"prim":"timestamp"},{"prim":"bytes"}]}],"annots":["%proposal_key_list_sort_by_date"]},{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%downvotes"]},{"prim":"map","args":[{"prim":"string"},{"prim":"bytes"}],"annots":["%metadata"]}]},{"prim":"address","annots":["%proposer"]},{"prim":"nat","annots":["%proposer_frozen_token"]}]},{"prim":"pair","args":[{"prim":"timestamp","annots":["%start_date"]},{"prim":"nat","annots":["%upvotes"]}]},{"prim":"list","args":[{"prim":"pair","args":[{"prim":"address"},{"prim":"nat"}]}],"annots":["%voters"]}]}],"annots":["%proposals"]}]},{"prim":"nat","annots":["%quorum_threshold"]},{"prim":"address","annots":["%token_address"]}]},{"prim":"nat","annots":["%voting_period"]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"map","args":[{"prim":"string"},{"prim":"bytes"}],"annots":["%custom_entrypoints"]},{"prim":"lambda","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%downvotes"]},{"prim":"map","args":[{"prim":"string"},{"prim":"bytes"}],"annots":["%metadata"]}]},{"prim":"address","annots":["%proposer"]},{"prim":"nat","annots":["%proposer_frozen_token"]}]},{"prim":"pair","args":[{"prim":"timestamp","annots":["%start_date"]},{"prim":"nat","annots":["%upvotes"]}]},{"prim":"list","args":[{"prim":"pair","args":[{"prim":"address"},{"prim":"nat"}]}],"annots":["%voters"]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address","annots":["%admin"]},{"prim":"map","args":[{"prim":"string"},{"prim":"bytes"}],"annots":["%extra"]}]},{"prim":"big_map","args":[{"prim":"pair","args":[{"prim":"address"},{"prim":"nat"}]},{"prim":"nat"}],"annots":["%ledger"]},{"prim":"big_map","args":[{"prim":"string"},{"prim":"bytes"}],"annots":["%metadata"]}]},{"prim":"pair","args":[{"prim":"or","args":[{"prim":"unit","annots":["%not_in_migration"]},{"prim":"or","args":[{"prim":"address","annots":["%migratingTo"]},{"prim":"address","annots":["%migratedTo"]}]}],"annots":["%migration_status"]},{"prim":"big_map","args":[{"prim":"pair","args":[{"prim":"address","annots":["%owner"]},{"prim":"address","annots":["%operator"]}]},{"prim":"unit"}],"annots":["%operators"]}]},{"prim":"address","annots":["%pending_owner"]},{"prim":"nat","annots":["%permits_counter"]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"set","args":[{"prim":"pair","args":[{"prim":"timestamp"},{"prim":"bytes"}]}],"annots":["%proposal_key_list_sort_by_date"]},{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%downvotes"]},{"prim":"map","args":[{"prim":"string"},{"prim":"bytes"}],"annots":["%metadata"]}]},{"prim":"address","annots":["%proposer"]},{"prim":"nat","annots":["%proposer_frozen_token"]}]},{"prim":"pair","args":[{"prim":"timestamp","annots":["%start_date"]},{"prim":"nat","annots":["%upvotes"]}]},{"prim":"list","args":[{"prim":"pair","args":[{"prim":"address"},{"prim":"nat"}]}],"annots":["%voters"]}]}],"annots":["%proposals"]}]},{"prim":"nat","annots":["%quorum_threshold"]},{"prim":"address","annots":["%token_address"]}]},{"prim":"nat","annots":["%voting_period"]}]},{"prim":"pair","args":[{"prim":"list","args":[{"prim":"operation"}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address","annots":["%admin"]},{"prim":"map","args":[{"prim":"string"},{"prim":"bytes"}],"annots":["%extra"]}]},{"prim":"big_map","args":[{"prim":"pair","args":[{"prim":"address"},{"prim":"nat"}]},{"prim":"nat"}],"annots":["%ledger"]},{"prim":"big_map","args":[{"prim":"string"},{"prim":"bytes"}],"annots":["%metadata"]}]},{"prim":"pair","args":[{"prim":"or","args":[{"prim":"unit","annots":["%not_in_migration"]},{"prim":"or","args":[{"prim":"address","annots":["%migratingTo"]},{"prim":"address","annots":["%migratedTo"]}]}],"annots":["%migration_status"]},{"prim":"big_map","args":[{"prim":"pair","args":[{"prim":"address","annots":["%owner"]},{"prim":"address","annots":["%operator"]}]},{"prim":"unit"}],"annots":["%operators"]}]},{"prim":"address","annots":["%pending_owner"]},{"prim":"nat","annots":["%permits_counter"]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"set","args":[{"prim":"pair","args":[{"prim":"timestamp"},{"prim":"bytes"}]}],"annots":["%proposal_key_list_sort_by_date"]},{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%downvotes"]},{"prim":"map","args":[{"prim":"string"},{"prim":"bytes"}],"annots":["%metadata"]}]},{"prim":"address","annots":["%proposer"]},{"prim":"nat","annots":["%proposer_frozen_token"]}]},{"prim":"pair","args":[{"prim":"timestamp","annots":["%start_date"]},{"prim":"nat","annots":["%upvotes"]}]},{"prim":"list","args":[{"prim":"pair","args":[{"prim":"address"},{"prim":"nat"}]}],"annots":["%voters"]}]}],"annots":["%proposals"]}]},{"prim":"nat","annots":["%quorum_threshold"]},{"prim":"address","annots":["%token_address"]}]},{"prim":"nat","annots":["%voting_period"]}]}],"annots":["%decision_lambda"]}]},{"prim":"nat","annots":["%max_proposals"]},{"prim":"nat","annots":["%max_quorum_threshold"]}]},{"prim":"pair","args":[{"prim":"nat","annots":["%max_votes"]},{"prim":"nat","annots":["%max_voting_period"]}]},{"prim":"nat","annots":["%min_quorum_threshold"]},{"prim":"nat","annots":["%min_voting_period"]}]},{"prim":"lambda","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%frozen_token"]},{"prim":"map","args":[{"prim":"string"},{"prim":"bytes"}],"annots":["%proposal_metadata"]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address","annots":["%admin"]},{"prim":"map","args":[{"prim":"string"},{"prim":"bytes"}],"annots":["%extra"]}]},{"prim":"big_map","args":[{"prim":"pair","args":[{"prim":"address"},{"prim":"nat"}]},{"prim":"nat"}],"annots":["%ledger"]},{"prim":"big_map","args":[{"prim":"string"},{"prim":"bytes"}],"annots":["%metadata"]}]},{"prim":"pair","args":[{"prim":"or","args":[{"prim":"unit","annots":["%not_in_migration"]},{"prim":"or","args":[{"prim":"address","annots":["%migratingTo"]},{"prim":"address","annots":["%migratedTo"]}]}],"annots":["%migration_status"]},{"prim":"big_map","args":[{"prim":"pair","args":[{"prim":"address","annots":["%owner"]},{"prim":"address","annots":["%operator"]}]},{"prim":"unit"}],"annots":["%operators"]}]},{"prim":"address","annots":["%pending_owner"]},{"prim":"nat","annots":["%permits_counter"]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"set","args":[{"prim":"pair","args":[{"prim":"timestamp"},{"prim":"bytes"}]}],"annots":["%proposal_key_list_sort_by_date"]},{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%downvotes"]},{"prim":"map","args":[{"prim":"string"},{"prim":"bytes"}],"annots":["%metadata"]}]},{"prim":"address","annots":["%proposer"]},{"prim":"nat","annots":["%proposer_frozen_token"]}]},{"prim":"pair","args":[{"prim":"timestamp","annots":["%start_date"]},{"prim":"nat","annots":["%upvotes"]}]},{"prim":"list","args":[{"prim":"pair","args":[{"prim":"address"},{"prim":"nat"}]}],"annots":["%voters"]}]}],"annots":["%proposals"]}]},{"prim":"nat","annots":["%quorum_threshold"]},{"prim":"address","annots":["%token_address"]}]},{"prim":"nat","annots":["%voting_period"]}]},{"prim":"bool"}],"annots":["%proposal_check"]},{"prim":"lambda","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%downvotes"]},{"prim":"map","args":[{"prim":"string"},{"prim":"bytes"}],"annots":["%metadata"]}]},{"prim":"address","annots":["%proposer"]},{"prim":"nat","annots":["%proposer_frozen_token"]}]},{"prim":"pair","args":[{"prim":"timestamp","annots":["%start_date"]},{"prim":"nat","annots":["%upvotes"]}]},{"prim":"list","args":[{"prim":"pair","args":[{"prim":"address"},{"prim":"nat"}]}],"annots":["%voters"]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address","annots":["%admin"]},{"prim":"map","args":[{"prim":"string"},{"prim":"bytes"}],"annots":["%extra"]}]},{"prim":"big_map","args":[{"prim":"pair","args":[{"prim":"address"},{"prim":"nat"}]},{"prim":"nat"}],"annots":["%ledger"]},{"prim":"big_map","args":[{"prim":"string"},{"prim":"bytes"}],"annots":["%metadata"]}]},{"prim":"pair","args":[{"prim":"or","args":[{"prim":"unit","annots":["%not_in_migration"]},{"prim":"or","args":[{"prim":"address","annots":["%migratingTo"]},{"prim":"address","annots":["%migratedTo"]}]}],"annots":["%migration_status"]},{"prim":"big_map","args":[{"prim":"pair","args":[{"prim":"address","annots":["%owner"]},{"prim":"address","annots":["%operator"]}]},{"prim":"unit"}],"annots":["%operators"]}]},{"prim":"address","annots":["%pending_owner"]},{"prim":"nat","annots":["%permits_counter"]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"set","args":[{"prim":"pair","args":[{"prim":"timestamp"},{"prim":"bytes"}]}],"annots":["%proposal_key_list_sort_by_date"]},{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%downvotes"]},{"prim":"map","args":[{"prim":"string"},{"prim":"bytes"}],"annots":["%metadata"]}]},{"prim":"address","annots":["%proposer"]},{"prim":"nat","annots":["%proposer_frozen_token"]}]},{"prim":"pair","args":[{"prim":"timestamp","annots":["%start_date"]},{"prim":"nat","annots":["%upvotes"]}]},{"prim":"list","args":[{"prim":"pair","args":[{"prim":"address"},{"prim":"nat"}]}],"annots":["%voters"]}]}],"annots":["%proposals"]}]},{"prim":"nat","annots":["%quorum_threshold"]},{"prim":"address","annots":["%token_address"]}]},{"prim":"nat","annots":["%voting_period"]}]},{"prim":"nat"}],"annots":["%rejected_proposal_return_value"]}]}]}]`),
						},
						Tags: types.LedgerTag | types.FA2Tag,
					},
				},
			},
		}, {
			name: "ooz1bkCQeYsZYP7vb4Dx7pYPRpWN11Z3G3yP1v4HAfdNXuHRv9c",
			ctx: &config.Context{
				RPC:         rpc,
				Storage:     generalRepo,
				Contracts:   contractRepo,
				BigMapDiffs: bmdRepo,
				Blocks:      blockRepo,
				Protocols:   protoRepo,
				Operations:  operaitonsRepo,
				Scripts:     scriptRepo,
				Cache: cache.NewCache(
					rpc, accountsRepo, contractRepo, protoRepo, bluemonday.UGCPolicy(),
				),
			},
			paramsOpts: []ParseParamsOption{
				WithHead(noderpc.Header{
					Timestamp: timestamp,
					Protocol:  "PsFLorenaUUuikDWvMDr6fGBRG8kt3e3D3fHoXK1j1BFRxeSH4i",
					Level:     1516349,
					ChainID:   "NetXdQprcVkpaWU",
				}),
				WithProtocol(&protocol.Protocol{
					Constants: &protocol.Constants{
						CostPerByte:                  1000,
						HardGasLimitPerOperation:     400000,
						HardStorageLimitPerOperation: 60000,
						TimeBetweenBlocks:            60,
					},
					Hash:    "PsFLorenaUUuikDWvMDr6fGBRG8kt3e3D3fHoXK1j1BFRxeSH4i",
					ID:      4,
					SymLink: bcd.SymLinkBabylon,
				}),
			},
			storage: map[string]int64{
				"KT1QcxwB4QyPKfmSwjH1VRxa6kquUjeDWeEy": 1516349,
			},
			filename: "./data/rpc/opg/ooz1bkCQeYsZYP7vb4Dx7pYPRpWN11Z3G3yP1v4HAfdNXuHRv9c.json",
			want: &parsers.TestStore{
				Operations: []*operation.Operation{
					{
						Kind: types.OperationKindTransaction,
						Source: account.Account{

							Address: "tz1aCzsYRUgDZBV7zb7Si6q2AobrocFW5qwb",
							Type:    types.AccountTypeTz,
						},
						Fee:      2235,
						Counter:  9432992,
						GasLimit: 18553,
						Destination: account.Account{
							Address: "KT1QcxwB4QyPKfmSwjH1VRxa6kquUjeDWeEy",
							Type:    types.AccountTypeContract,
						},
						Status: types.OperationStatusApplied,
						Level:  1516349,

						Hash:      "ooz1bkCQeYsZYP7vb4Dx7pYPRpWN11Z3G3yP1v4HAfdNXuHRv9c",
						Timestamp: timestamp,
						Entrypoint: types.NullString{
							Str:   "transfer",
							Valid: true,
						},
						Tags: types.FA2Tag | types.LedgerTag,
						Initiator: account.Account{
							Address: "tz1aCzsYRUgDZBV7zb7Si6q2AobrocFW5qwb",
							Type:    types.AccountTypeTz,
						},
						Delegate:        account.Account{},
						Parameters:      []byte(`{"entrypoint":"transfer","value":[{"prim":"Pair","args":[{"string":"tz1aCzsYRUgDZBV7zb7Si6q2AobrocFW5qwb"},[{"prim":"Pair","args":[{"string":"tz1a6ZKyEoCmfpsY74jEq6uKBK8RQXdj1aVi"},{"prim":"Pair","args":[{"int":"12"},{"int":"1"}]}]}]]}]}`),
						ProtocolID:      4,
						DeffatedStorage: []byte(`{"prim":"Pair","args":[{"prim":"Pair","args":[{"prim":"Pair","args":[{"int":"746"},{"int":"4992269"}]},{"prim":"Pair","args":[{"int":"747"},{"int":"748"}]}]},{"int":"749"}]}`),
						BigMapDiffs: []*bigmapdiff.BigMapDiff{
							{
								Ptr:        746,
								KeyHash:    "expruSKSLw7MS3ou3pPd7MUXy5QDPtVvkUNF4yWS2g6n8mXGzDJCG7",
								Key:        []byte(`{"int":"12" }`),
								Value:      []byte(`{"bytes":"00009e96262b1bfc9a709603668843d52994358be677"}`),
								Contract:   "KT1QcxwB4QyPKfmSwjH1VRxa6kquUjeDWeEy",
								Level:      1516349,
								Timestamp:  timestamp,
								ProtocolID: 4,
							},
						},
					},
				},
				BigMapState: []*bigmapdiff.BigMapState{
					{
						Ptr:             746,
						KeyHash:         "expruSKSLw7MS3ou3pPd7MUXy5QDPtVvkUNF4yWS2g6n8mXGzDJCG7",
						Key:             []byte(`{"int":"12"}`),
						Value:           []byte(`{"bytes":"00009e96262b1bfc9a709603668843d52994358be677"}`),
						Contract:        "KT1QcxwB4QyPKfmSwjH1VRxa6kquUjeDWeEy",
						LastUpdateLevel: 1516349,
						LastUpdateTime:  timestamp,
					},
				},
			},
		}, {
			name: "oocFt4vkkgQGfoRH54328cJUbDdWvj3x6KEs5Arm4XhqwwJmnJ8",
			ctx: &config.Context{
				RPC:         rpc,
				Storage:     generalRepo,
				Contracts:   contractRepo,
				BigMapDiffs: bmdRepo,
				Blocks:      blockRepo,
				Protocols:   protoRepo,
				Operations:  operaitonsRepo,
				Scripts:     scriptRepo,
				Cache: cache.NewCache(
					rpc, accountsRepo, contractRepo, protoRepo, bluemonday.UGCPolicy(),
				),
			},
			paramsOpts: []ParseParamsOption{
				WithHead(noderpc.Header{
					Timestamp: timestamp,
					Protocol:  "PsFLorenaUUuikDWvMDr6fGBRG8kt3e3D3fHoXK1j1BFRxeSH4i",
					Level:     1520888,
					ChainID:   "NetXdQprcVkpaWU",
				}),
				WithProtocol(&protocol.Protocol{
					Constants: &protocol.Constants{
						CostPerByte:                  1000,
						HardGasLimitPerOperation:     400000,
						HardStorageLimitPerOperation: 60000,
						TimeBetweenBlocks:            60,
					},
					Hash: "PsFLorenaUUuikDWvMDr6fGBRG8kt3e3D3fHoXK1j1BFRxeSH4i",

					ID:      4,
					SymLink: bcd.SymLinkBabylon,
				}),
			},
			filename: "./data/rpc/opg/oocFt4vkkgQGfoRH54328cJUbDdWvj3x6KEs5Arm4XhqwwJmnJ8.json",
			storage: map[string]int64{
				"KT1GBZmSxmnKJXGMdMLbugPfLyUPmuLSMwKS": 1520888,
				"KT1H1MqmUM4aK9i1833EBmYCCEfkbt6ZdSBc": 1520888,
			},
			want: &parsers.TestStore{
				BigMapState: []*bigmapdiff.BigMapState{
					{

						Ptr:             1264,
						Contract:        "KT1GBZmSxmnKJXGMdMLbugPfLyUPmuLSMwKS",
						LastUpdateLevel: 1520888,
						LastUpdateTime:  timestamp,
						KeyHash:         "exprvKwnhi4q3tSmdvgqXACxfN6zARGkoikHv7rqohvQKg4cWdgsii",
						Key:             []byte(`{"bytes":"62616c6c732e74657a"}`),
						Value:           []byte(`{"prim":"Pair","args":[{"prim":"Pair","args":[{"prim":"Pair","args":[{"prim":"Some","args":[{"bytes":"0000c0ca282a775946b5ecbe02e5cf73e25f6b62b70c"}]},[]]},{"prim":"Pair","args":[{"prim":"Some","args":[{"bytes":"62616c6c732e74657a"}]},[]]}]},{"prim":"Pair","args":[{"prim":"Pair","args":[{"int":"2"},{"bytes":"0000753f63893674b6d523f925f0d787bf9270b95c33"}]},{"prim":"Some","args":[{"int":"3223"}]}]}]}`),
					},
				},
				Operations: []*operation.Operation{
					{
						Kind: types.OperationKindTransaction,
						Hash: "oocFt4vkkgQGfoRH54328cJUbDdWvj3x6KEs5Arm4XhqwwJmnJ8",
						Source: account.Account{

							Address: "tz1WKygtstVY96oyc6Rmk945dMf33LeihgWT",
							Type:    types.AccountTypeTz,
						},
						Initiator: account.Account{

							Address: "tz1WKygtstVY96oyc6Rmk945dMf33LeihgWT",
							Type:    types.AccountTypeTz,
						},
						Delegate:     account.Account{},
						Status:       types.OperationStatusApplied,
						Fee:          5043,
						Counter:      10671622,
						GasLimit:     46511,
						StorageLimit: 400,
						Amount:       0,
						Timestamp:    timestamp,
						Level:        1520888,

						Entrypoint: types.NullString{
							Str:   "update_record",
							Valid: true,
						},
						ProtocolID: 4,
						Destination: account.Account{

							Address: "KT1H1MqmUM4aK9i1833EBmYCCEfkbt6ZdSBc",
							Type:    types.AccountTypeContract,
						},
						Parameters:      []byte(`{"entrypoint":"update_record","value":{"prim":"Pair","args":[{"bytes":"62616c6c732e74657a"},{"prim":"Pair","args":[{"prim":"Some","args":[{"string":"tz1dDQc4KsTHEFe3USc66Wti2pBatZ3UDbD4"}]},{"prim":"Pair","args":[{"string":"tz1WKygtstVY96oyc6Rmk945dMf33LeihgWT"},[]]}]}]}}`),
						DeffatedStorage: []byte(`{"prim":"Pair","args":[{"prim":"Pair","args":[{"bytes":"01535d971759846a1f2be8610e36f2db40fe8ce40800"},{"int":"1268"}]},{"bytes":"01ebb657570e494e8a7bd43ac3bf7cfd0267a32a9f00"}]}`),
					},
					{
						Kind:       types.OperationKindTransaction,
						Hash:       "oocFt4vkkgQGfoRH54328cJUbDdWvj3x6KEs5Arm4XhqwwJmnJ8",
						Internal:   true,
						Timestamp:  timestamp,
						Status:     types.OperationStatusApplied,
						Level:      1520888,
						Nonce:      newInt64Ptr(0),
						Counter:    10671622,
						ProtocolID: 4,
						Burned:     27000,
						Entrypoint: types.NullString{
							Str:   "execute",
							Valid: true,
						},
						Initiator: account.Account{
							Address: "tz1WKygtstVY96oyc6Rmk945dMf33LeihgWT",
							Type:    types.AccountTypeTz,
						},
						Source: account.Account{
							Address: "KT1H1MqmUM4aK9i1833EBmYCCEfkbt6ZdSBc",
							Type:    types.AccountTypeContract,
						},
						Destination: account.Account{
							Address: "KT1GBZmSxmnKJXGMdMLbugPfLyUPmuLSMwKS",
							Type:    types.AccountTypeContract,
						},
						Delegate:        account.Account{},
						Parameters:      []byte(`{"entrypoint":"execute","value":{"prim":"Pair","args":[{"string":"UpdateRecord"},{"prim":"Pair","args":[{"bytes":"0507070a0000000962616c6c732e74657a070705090a000000160000c0ca282a775946b5ecbe02e5cf73e25f6b62b70c07070a000000160000753f63893674b6d523f925f0d787bf9270b95c330200000000"},{"bytes":"0000753f63893674b6d523f925f0d787bf9270b95c33"}]}]}}`),
						DeffatedStorage: []byte(`{"prim":"Pair","args":[[{"int":"1260"},{"prim":"Pair","args":[{"prim":"Pair","args":[{"int":"1261"},{"int":"1262"}]},{"prim":"Pair","args":[{"int":"1263"},{"int":"9824"}]}]},{"prim":"Pair","args":[{"bytes":"01ebb657570e494e8a7bd43ac3bf7cfd0267a32a9f00"},{"int":"1264"}]},{"int":"1265"},{"int":"1266"}],[{"bytes":"014796e76af90e6327adfab057bbbe0375cd2c8c1000"},{"bytes":"015c6799f783b8d118b704267f634c5d24d19e9a9f00"},{"bytes":"0168e9b7d86646e312c76dfbedcbcdb24320875a3600"},{"bytes":"019178a76f3c41a9541d2291cad37dd5fb96a6850500"},{"bytes":"01ac3638385caa4ad8126ea84e061f4f49baa44d3c00"},{"bytes":"01d2a0974172cf6fc8b1eefdebd5bea681616f7c6f00"}]]}`),
						BigMapDiffs: []*bigmapdiff.BigMapDiff{
							{
								ProtocolID: 4,
								Contract:   "KT1GBZmSxmnKJXGMdMLbugPfLyUPmuLSMwKS",
								Ptr:        1264,
								Level:      1520888,
								Timestamp:  timestamp,
								KeyHash:    "exprvKwnhi4q3tSmdvgqXACxfN6zARGkoikHv7rqohvQKg4cWdgsii",
								Key:        []byte(`{"bytes":"62616c6c732e74657a"}`),
								Value:      []byte(`{"prim":"Pair","args":[{"prim":"Pair","args":[{"prim":"Pair","args":[{"prim":"Some","args":[{"bytes":"0000c0ca282a775946b5ecbe02e5cf73e25f6b62b70c"}]},[]]},{"prim":"Pair","args":[{"prim":"Some","args":[{"bytes":"62616c6c732e74657a"}]},[]]}]},{"prim":"Pair","args":[{"prim":"Pair","args":[{"int":"2"},{"bytes":"0000753f63893674b6d523f925f0d787bf9270b95c33"}]},{"prim":"Some","args":[{"int":"3223"}]}]}]}`),
							},
						},
					},
				},
			},
		}, {
			name: "ooffKPL6WmMgqzLGtRtLp2HdEbVL3K2fVzKQLyxsBFMC84wpjRt",
			ctx: &config.Context{
				RPC:         rpc,
				Storage:     generalRepo,
				Contracts:   contractRepo,
				BigMapDiffs: bmdRepo,
				Blocks:      blockRepo,
				Protocols:   protoRepo,
				Operations:  operaitonsRepo,
				Scripts:     scriptRepo,
				Cache: cache.NewCache(
					rpc, accountsRepo, contractRepo, protoRepo, bluemonday.UGCPolicy(),
				),
			},
			paramsOpts: []ParseParamsOption{
				WithHead(noderpc.Header{
					Timestamp: timestamp,
					Protocol:  "PtHangzHogokSuiMHemCuowEavgYTP8J5qQ9fQS793MHYFpCY3r",
					Level:     15400,
					ChainID:   "NetXuXoGoLxNK6o",
				}),
				WithProtocol(&protocol.Protocol{
					Constants: &protocol.Constants{
						CostPerByte:                  1000,
						HardGasLimitPerOperation:     400000,
						HardStorageLimitPerOperation: 60000,
						TimeBetweenBlocks:            60,
					},
					Hash:    "PtHangzHogokSuiMHemCuowEavgYTP8J5qQ9fQS793MHYFpCY3r",
					ID:      5,
					SymLink: bcd.SymLinkBabylon,
				}),
			},
			filename: "./data/rpc/opg/ooffKPL6WmMgqzLGtRtLp2HdEbVL3K2fVzKQLyxsBFMC84wpjRt.json",
			want: &parsers.TestStore{
				Operations: []*operation.Operation{
					{
						Kind: types.OperationKindRegisterGlobalConstant,
						Hash: "ooffKPL6WmMgqzLGtRtLp2HdEbVL3K2fVzKQLyxsBFMC84wpjRt",
						Source: account.Account{
							Address: "tz1SMARcpWCydHsGgz4MRoK9NkbpBmmUAfNe",
							Type:    types.AccountTypeTz,
						},
						Initiator: account.Account{
							Address: "tz1SMARcpWCydHsGgz4MRoK9NkbpBmmUAfNe",
							Type:    types.AccountTypeTz,
						},
						Status:       types.OperationStatusApplied,
						Fee:          377,
						Counter:      1,
						GasLimit:     1333,
						ConsumedGas:  1233,
						StorageSize:  80,
						StorageLimit: 100,
						Timestamp:    timestamp,
						Level:        15400,
						ProtocolID:   5,
					},
				},
				GlobalConstants: []*contract.GlobalConstant{
					{
						Level:     15400,
						Timestamp: timestamp,
						Address:   "expru54tk2k4E81xQy63P6x3RijnTz51s2m7BV7pr3fDQH8YDqiYvR",
						Value:     []byte(`[{"prim":"PUSH","args":[{"prim":"int"},{"int":"10"}]},{"prim":"SWAP"},{"prim":"MUL"}]`),
					},
				},
			},
		}, {
			name: "oozvzXiZmVW9QtYjKmDuYqoHNCEvt32FwM2cUgQee2S1SGWgumA",
			ctx: &config.Context{
				RPC:         rpc,
				Storage:     generalRepo,
				Contracts:   contractRepo,
				BigMapDiffs: bmdRepo,
				Blocks:      blockRepo,
				Protocols:   protoRepo,
				Operations:  operaitonsRepo,
				Scripts:     scriptRepo,
				Cache: cache.NewCache(
					rpc, accountsRepo, contractRepo, protoRepo, bluemonday.UGCPolicy(),
				),
			},
			storage: map[string]int64{
				"KT1BM1SyQnTzNU1J8TZv5Mdj4ScuTgNKH5uj": 381735,
			},
			paramsOpts: []ParseParamsOption{
				WithHead(noderpc.Header{
					Timestamp: timestamp,
					Protocol:  "Psithaca2MLRFYargivpo7YvUr7wUDqyxrdhC5CQq78mRvimz6A",
					Level:     381735,
					ChainID:   "NetXnHfVqm9iesp",
				}),
				WithProtocol(&protocol.Protocol{
					Constants: &protocol.Constants{
						CostPerByte:                  1000,
						HardGasLimitPerOperation:     400000,
						HardStorageLimitPerOperation: 60000,
						TimeBetweenBlocks:            60,
					},
					Hash:    "Psithaca2MLRFYargivpo7YvUr7wUDqyxrdhC5CQq78mRvimz6A",
					ID:      6,
					SymLink: bcd.SymLinkBabylon,
				}),
			},
			filename: "./data/rpc/opg/oozvzXiZmVW9QtYjKmDuYqoHNCEvt32FwM2cUgQee2S1SGWgumA.json",
			want: &parsers.TestStore{
				Operations: []*operation.Operation{
					{
						Kind: types.OperationKindTransaction,
						Hash: "oozvzXiZmVW9QtYjKmDuYqoHNCEvt32FwM2cUgQee2S1SGWgumA",
						Source: account.Account{
							Address: "tz1RiUE3Ao53juAz4uDYx1J3tHJMye6jPfhp",
							Type:    types.AccountTypeTz,
						},
						Destination: account.Account{
							Address: "KT1Jk8LRDoj6LkopYZwRq5ZEWBhYv8nVc6e6",
							Type:    types.AccountTypeContract,
						},
						Initiator: account.Account{
							Address: "tz1RiUE3Ao53juAz4uDYx1J3tHJMye6jPfhp",
							Type:    types.AccountTypeTz,
						},
						Status:              types.OperationStatusApplied,
						Fee:                 4175,
						Counter:             34005,
						GasLimit:            36150,
						ConsumedGas:         12916,
						StorageSize:         60070,
						StorageLimit:        27255,
						Burned:              92000,
						PaidStorageSizeDiff: 92,
						Timestamp:           timestamp,
						Level:               381735,
						ProtocolID:          6,
						Parameters:          []byte(`{"entrypoint":"add_pool","value":{"prim":"Pair","args":[{"int":"100000"},{"prim":"Pair","args":[[{"prim":"Left","args":[{"string":"KT1CR5crmVrJntzwv5XVrBv8Dk1mxZFKHj9z"}]},{"prim":"Right","args":[{"prim":"Pair","args":[{"string":"KT19H9YbHqsxFTayap7aTEfbcnyPeALKYgt9"},{"int":"0"}]}]}],{"prim":"Pair","args":[[{"prim":"Elt","args":[{"int":"0"},{"prim":"Pair","args":[{"int":"1000000000000000000"},{"int":"1000000000000000000000000"}]}]},{"prim":"Elt","args":[{"int":"1"},{"prim":"Pair","args":[{"int":"1000000000000000000"},{"int":"100000000000000000000000000"}]}]}],{"prim":"Pair","args":[{"string":"tz1RiUE3Ao53juAz4uDYx1J3tHJMye6jPfhp"},[{"string":"tz1RiUE3Ao53juAz4uDYx1J3tHJMye6jPfhp"}]]}]}]}]}}`),
						Entrypoint:          types.NullString{Str: "add_pool", Valid: true},
						DeffatedStorage:     []byte(`[[{"prim":"Pair","args":[{"bytes":"000042a7bb84edce2af4cc8ab0bc83ded699efc9300a"},{"prim":"Pair","args":[{"int":"0"},{"int":"40063"}]}]},{"int":"0"},{"int":"0"},{"int":"1"},{"int":"40064"},{"prim":"Pair","args":[{"bytes":"0107a2fc7b796ea23ad82e768221a78c86c77c640500"},{"int":"0"}]},{"int":"0"},[]],{"int":"40065"},{"int":"40066"},{"int":"40067"},{"prim":"Some","args":[{"bytes":"0502000057490743076503680362070701000000096e6f742d746f6b656e00a40103420200005727037a037a05700002037a032105290003034507430368010000001277726f6e672d746f6b656e732d636f756e740521000305290005034505210003031903250743036200020521000403190328074303620004057000040319033203140314072c0200000002032002000000020327074303620000072303620764036e0765036e03620342034c0321057100020529000305520200000036034c0321032105710002031605700003057000030317034c0346034c0350055000010321074303620001057000020317031205500002031603210348034c0342030c05210004031605290009034c0321057100020329072f02000000000200000017032007430368010000000a706f6f6c2d6578697374032705210003052900050538020000001c0317074303620000034c03210571000203170570000203160542000307430368010000000a612d746f6f2d68696768052100050316074303620080897a03190328072c02000000020320020000000203270743036200000743036200000723036203620342074303620000074303620000074303620000054200030570000303400521000b052100090316033a03400570000c0521000b0316033a0542000807430359030303770521000703160529000b07720765036e0362076503620760036207650362036207720765036e0764036e0765036e0362036207720764036e0765036e0362036207720765036e03620566036e07720765036e0362036207720362076503620765036b076503620765036b07650760036207650362076503620362076507650362076503620362076507650760036203620362036205700009034607430362000003500772036903620743036200000521000d030c034c0346034c035007720362076003620764036e0765036e03620570000c034607430362000003500743036200010521000e052900080570000e0529000703480542000f052100030529000707720362036905210005052900030772036207650362076003680369072303680369074303690a00000035697066733a2f2f516d5558464a787747456a6b6e5035444b4e774d777357587672593847326a554d62786a64424533656b50384450034607430368010000000c7468756d626e61696c5572690350074303690a00000006736451504c50034607430368010000000673796d626f6c0350074303690a0000000474727565034607430368010000001273686f756c6450726566657253796d626f6c0350074303690a0000001d537461626c652044455820517569707553776170204c5020746f6b656e03460743036801000000046e616d650350074303690a0000002c4c697175696469747920506f6f6c20746f6b656e206f662051756970755377617020537461626c6520444558034607430368010000000b6465736372697074696f6e0350074303690a0000000231380346074303680100000008646563696d616c730350074303620000034203460743036200000350077203680369074303690a000001ad7b226e616d65223a2251756970755377617020537461626c652044455820706f6f6c222c2276657273696f6e223a2276312e302e30222c226465736372697074696f6e223a22506f6f6c20666f72207377617070696e6720746f6b656e732077697468206c6f7720736c697070616765222c22617574686f7273223a5b224d6164666973682e536f6c7574696f6e73203c68747470733a2f2f7777772e6d6164666973682e736f6c7574696f6e733e225d2c22736f75726365223a7b22746f6f6c73223a5b224c69676f222c22466c657874657361225d2c226c6f636174696f6e223a2268747470733a2f2f6769746875622e636f6d2f6d6164666973682d736f6c7574696f6e732f7175697075737761702d737461626c652d636f72652f626c6f622f6d61696e2f636f6e7472616374732f6d61696e2f6465782e6c69676f227d2c22686f6d6570616765223a2268747470733a2f2f7175697075737761702e636f6d222c22696e7465726661636573223a5b22545a49502d313220676974203137323866636665222c22545a49502d3136225d2c226572726f7273223a5b5d2c227669657773223a5b5d7d03460743036801000000036465780350074303690a0000001174657a6f732d73746f726167653a6465780346074303680100000000035005700005054200060743036a0000053e035d034203420200000004037a037a051d0200004cf005000764076408640861036203690000001225636f70795f6465785f66756e6374696f6e046c0000000725667265657a650000000f25666163746f72795f616374696f6e086407640764086504590000000425616464046e0000000a2563616e64696461746500000011256164645f72656d5f6d616e616765727308650864046e0000000525666131320865046e0000000e25746f6b656e5f6164647265737304620000000925746f6b656e5f696400000004256661320000000625746f6b656e04620000000725616d6f756e740000001025636c61696d5f646576656c6f7065720764086504620000000825706f6f6c5f69640765046200000009256675747572655f41046b0000000c256675747572655f74696d65000000072572616d705f41046e0000000a257365745f61646d696e07640764046e00000015257365745f64656661756c745f726566657272616c086504620000000825706f6f6c5f69640865046200000005256c705f66076504620000000a257374616b6572735f66046200000006257265665f66000000042566656500000009257365745f6665657304620000000c2573746f705f72616d705f410000000a257573655f61646d696e086408640764076408650864046e0000000525666131320865046e0000000e25746f6b656e5f6164647265737304620000000925746f6b656e5f696400000004256661320000000625746f6b656e04620000000725616d6f756e740000000f25636c61696d5f726566657272616c086504620000000825706f6f6c5f6964076508600362036200000010256d696e5f616d6f756e74735f6f75740765046200000007257368617265730765046b0000000925646561646c696e650663036e0000000925726563656976657200000007256469766573740764086504620000000825706f6f6c5f696407650860036203620000000c25616d6f756e74735f6f7574076504620000000b256d61785f7368617265730765046b0000000925646561646c696e6507650663036e000000092572656365697665720663036e0000000925726566657272616c00000012256469766573745f696d62616c616e636564086504620000000825706f6f6c5f6964076504620000000725736861726573076504620000000c25746f6b656e5f696e646578076504620000000f256d696e5f616d6f756e745f6f75740765046b0000000925646561646c696e6507650663036e000000092572656365697665720663036e0000000925726566657272616c00000010256469766573745f6f6e655f636f696e07640764086504620000000825706f6f6c5f696407650462000000072573686172657307650860036203620000000b25696e5f616d6f756e74730765046b0000000925646561646c696e6507650663036e000000092572656365697665720663036e0000000925726566657272616c0000000725696e766573740864086504620000000825706f6f6c5f696404620000000725616d6f756e740000000425616464086504620000000825706f6f6c5f696404620000000725616d6f756e74000000072572656d6f766500000006257374616b65086504620000000825706f6f6c5f69640765046200000009256964785f66726f6d0765046200000007256964785f746f076504620000000725616d6f756e74076504620000000f256d696e5f616d6f756e745f6f75740765046b0000000925646561646c696e6507650663036e000000092572656365697665720663036e0000000925726566657272616c00000005257377617000000008257573655f6465780864076407640865065f0765046e00000006256f776e657204620000000925746f6b656e5f696400000009257265717565737473065a055f07650865046e00000006256f776e657204620000000925746f6b656e5f69640000000825726571756573740462000000082562616c616e6365000000092563616c6c6261636b0000000b2562616c616e63655f6f66086504620000000925746f6b656e5f6964065a0362000000092572656365697665720000000d25746f74616c5f737570706c790764065f0765046e000000062566726f6d5f065f0765046e0000000425746f5f076504620000000925746f6b656e5f696404620000000725616d6f756e74000000042574787300000009257472616e73666572086504620000000925746f6b656e5f69640860036803690000000b25746f6b656e5f696e666f00000010257570646174655f6d65746164617461065f07640865046e00000006256f776e65720765046e00000009256f70657261746f7204620000000925746f6b656e5f69640000000d256164645f6f70657261746f720865046e00000006256f776e65720765046e00000009256f70657261746f7204620000000925746f6b656e5f6964000000102572656d6f76655f6f70657261746f7200000011257570646174655f6f70657261746f72730000000a257573655f746f6b656e0000000c25757365725f616374696f6e050107650865046e000000062561646d696e0765046e000000112564656661756c745f726566657272616c07650666036e00000009256d616e6167657273076504620000000c25706f6f6c735f636f756e74076508610362076003620764046e0000000525666131320865046e0000000e25746f6b656e5f6164647265737304620000000925746f6b656e5f696400000004256661320000000725746f6b656e7307650861036903620000000b25706f6f6c5f746f5f6964076508610362076504620000000c25696e697469616c5f415f660765046b0000000f25696e697469616c5f415f74696d65076504620000000b256675747572655f415f660765046b0000000e256675747572655f415f74696d65076508600362076504620000000725726174655f66076504620000001725707265636973696f6e5f6d756c7469706c6965725f660462000000092572657365727665730000000c25746f6b656e735f696e666f07650865046200000005256c705f66076504620000000a257374616b6572735f66046200000006257265665f660000000425666565076508650860036203620000000e25616363756d756c61746f725f6604620000000d25746f74616c5f7374616b656400000013257374616b65725f616363756d756c61746f7204620000000d25746f74616c5f737570706c790000000625706f6f6c73076508610765036e0362036200000007256c6564676572076508610765036e03620566036e0000000b25616c6c6f77616e636573076508610764046e0000000525666131320865046e0000000e25746f6b656e5f6164647265737304620000000925746f6b656e5f6964000000042566613203620000000c256465765f72657761726473076508610765036e0764046e0000000525666131320865046e0000000e25746f6b656e5f6164647265737304620000000925746f6b656e5f6964000000042566613203620000001125726566657272616c5f72657761726473076508610765036e036207650462000000082562616c616e6365086003620765046200000009257265776172645f6604620000000925666f726d65725f6600000009256561726e696e677300000010257374616b6572735f62616c616e636507650865046e0000000e25746f6b656e5f6164647265737304620000000925746f6b656e5f69640000000c2571756970755f746f6b656e0765046e0000001025666163746f72795f616464726573730459000000082573746172746564000000082573746f72616765076508610368036900000009256d65746164617461076508610362076504620000000925746f6b656e5f69640860036803690000000b25746f6b656e5f696e666f0000000f25746f6b656e5f6d6574616461746107650861036203690000000e2561646d696e5f6c616d6264617307650861036203690000000c256465785f6c616d626461730861036203690000000e25746f6b656e5f6c616d626461730502020000247707430368010000001066756e6374696f6e2d6e6f742d73657407430368010000001663616e742d756e7061636b2d7573652d6c616d6264610743036801000000126e6f742d636f6e74726163742d61646d696e07430368010000000b6e6f742d7374617274656405700004037a0321072e020000007d072e020000006d0570000203210571000203160529001b034803190325072c0200000000020000001b0743036801000000106661696c656420617373657274696f6e0327072e02000000040550000902000000200320032103210571000203160570000203160529001c033f0550001c0550000102000000040320034c02000000040320034c034c072e0200000cae072e020000001a0570000205700003057000040570000505200005053d036d03420200000c880571000203210571000303160529001c072c02000000020320020000000203270321072e02000000a6072e020000003e072e020000002803200571000203210571000303160316034803190325072c02000000040320034f02000000020327020000000a0570000305200002034f020000005c072e020000002803200571000203210571000303160316034803190325072c02000000040320034f02000000020327020000002803200571000203210571000303160316034803190325072c02000000040320034f020000000203270200000090072e020000005c072e020000002803200571000203210571000303160316034803190325072c02000000040320034f02000000020327020000002803200571000203210571000303160316034803190325072c02000000040320034f02000000020327020000002803200571000203210571000303160316034803190325072c02000000040320034f0200000002032703200321072e0200000044072e020000001c072e0200000008032007430362000002000000080320074303620002020000001c072e02000000080320074303620003020000000803200743036200010200000030072e020000001c072e020000000803200743036200060200000008032007430362000502000000080320074303620004057000040521000405290007052100030329072f020000000203270200000004034c032005700004034c050d075e0765076407640764086504590000000425616464046e0000000a2563616e64696461746500000011256164645f72656d5f6d616e616765727308650864046e0000000525666131320865046e0000000e25746f6b656e5f6164647265737304620000000925746f6b656e5f696400000004256661320000000625746f6b656e04620000000725616d6f756e740000001025636c61696d5f646576656c6f7065720764086504620000000825706f6f6c5f69640765046200000009256675747572655f41046b0000000c256675747572655f74696d65000000072572616d705f41046e0000000a257365745f61646d696e07640764046e00000015257365745f64656661756c745f726566657272616c086504620000000825706f6f6c5f69640865046200000005256c705f66076504620000000a257374616b6572735f66046200000006257265665f66000000042566656500000009257365745f6665657304620000000c2573746f705f72616d705f410765046e000000062561646d696e0765046e000000112564656661756c745f726566657272616c07650666036e00000009256d616e6167657273076504620000000c25706f6f6c735f636f756e74076508610362076003620764046e0000000525666131320865046e0000000e25746f6b656e5f6164647265737304620000000925746f6b656e5f696400000004256661320000000725746f6b656e7307650861036903620000000b25706f6f6c5f746f5f6964076508610362076504620000000c25696e697469616c5f415f660765046b0000000f25696e697469616c5f415f74696d65076504620000000b256675747572655f415f660765046b0000000e256675747572655f415f74696d65076508600362076504620000000725726174655f66076504620000001725707265636973696f6e5f6d756c7469706c6965725f660462000000092572657365727665730000000c25746f6b656e735f696e666f07650865046200000005256c705f66076504620000000a257374616b6572735f66046200000006257265665f660000000425666565076508650860036203620000000e25616363756d756c61746f725f6604620000000d25746f74616c5f7374616b656400000013257374616b65725f616363756d756c61746f7204620000000d25746f74616c5f737570706c790000000625706f6f6c73076508610765036e0362036200000007256c6564676572076508610765036e03620566036e0000000b25616c6c6f77616e636573076508610764046e0000000525666131320865046e0000000e25746f6b656e5f6164647265737304620000000925746f6b656e5f6964000000042566613203620000000c256465765f72657761726473076508610765036e0764046e0000000525666131320865046e0000000e25746f6b656e5f6164647265737304620000000925746f6b656e5f6964000000042566613203620000001125726566657272616c5f72657761726473076508610765036e036207650462000000082562616c616e6365086003620765046200000009257265776172645f6604620000000925666f726d65725f6600000009256561726e696e677300000010257374616b6572735f62616c616e636507650865046e0000000e25746f6b656e5f6164647265737304620000000925746f6b656e5f69640000000c2571756970755f746f6b656e0765046e0000001025666163746f72795f6164647265737304590000000825737461727465640765055f036d0765046e000000062561646d696e0765046e000000112564656661756c745f726566657272616c07650666036e00000009256d616e6167657273076504620000000c25706f6f6c735f636f756e74076508610362076003620764046e0000000525666131320865046e0000000e25746f6b656e5f6164647265737304620000000925746f6b656e5f696400000004256661320000000725746f6b656e7307650861036903620000000b25706f6f6c5f746f5f6964076508610362076504620000000c25696e697469616c5f415f660765046b0000000f25696e697469616c5f415f74696d65076504620000000b256675747572655f415f660765046b0000000e256675747572655f415f74696d65076508600362076504620000000725726174655f66076504620000001725707265636973696f6e5f6d756c7469706c6965725f660462000000092572657365727665730000000c25746f6b656e735f696e666f07650865046200000005256c705f66076504620000000a257374616b6572735f66046200000006257265665f660000000425666565076508650860036203620000000e25616363756d756c61746f725f6604620000000d25746f74616c5f7374616b656400000013257374616b65725f616363756d756c61746f7204620000000d25746f74616c5f737570706c790000000625706f6f6c73076508610765036e0362036200000007256c6564676572076508610765036e03620566036e0000000b25616c6c6f77616e636573076508610764046e0000000525666131320865046e0000000e25746f6b656e5f6164647265737304620000000925746f6b656e5f6964000000042566613203620000000c256465765f72657761726473076508610765036e0764046e0000000525666131320865046e0000000e25746f6b656e5f6164647265737304620000000925746f6b656e5f6964000000042566613203620000001125726566657272616c5f72657761726473076508610765036e036207650462000000082562616c616e6365086003620765046200000009257265776172645f6604620000000925666f726d65725f6600000009256561726e696e677300000010257374616b6572735f62616c616e636507650865046e0000000e25746f6b656e5f6164647265737304620000000925746f6b656e5f69640000000c2571756970755f746f6b656e0765046e0000001025666163746f72795f616464726573730459000000082573746172746564072f020000000203270200000004034c03200521000403160570000303420326037a0743036200070570000303190325072c020000018b05700002032105290005072303680369074303690a00000035697066733a2f2f516d5558464a787747456a6b6e5035444b4e774d777357587672593847326a554d62786a64424533656b50384450034607430368010000000c7468756d626e61696c5572690350074303690a00000006736451504c50034607430368010000000673796d626f6c0350074303690a0000000474727565034607430368010000001273686f756c6450726566657253796d626f6c0350074303690a0000001d537461626c652044455820517569707553776170204c5020746f6b656e03460743036801000000046e616d650350074303690a0000002c4c697175696469747920506f6f6c20746f6b656e206f662051756970755377617020537461626c6520444558034607430368010000000b6465736372697074696f6e0350074303690a0000000231380346074303680100000008646563696d616c7303500743036200010521000605290007034b031103420743036200010521000605290007034b0311034c0346034c0350055000050200000004057000020570000205500001034c034202000016bf057000030320072e0200000b820571000203210571000303160529001c072c020000000203200200000002032705700003052100030529000905210003072e0200000044072e020000001c072e0200000008032007430362000502000000080320074303620002020000001c072e02000000080320074303620003020000000803200743036200040200000030072e020000001c072e0200000008032007430362000102000000080320074303620006020000000803200743036200000329072f020000000203270200000004034c032005700003034c050d075e076507640764076408650864046e0000000525666131320865046e0000000e25746f6b656e5f6164647265737304620000000925746f6b656e5f696400000004256661320000000625746f6b656e04620000000725616d6f756e740000000f25636c61696d5f726566657272616c086504620000000825706f6f6c5f6964076508600362036200000010256d696e5f616d6f756e74735f6f75740765046200000007257368617265730765046b0000000925646561646c696e650663036e0000000925726563656976657200000007256469766573740764086504620000000825706f6f6c5f696407650860036203620000000c25616d6f756e74735f6f7574076504620000000b256d61785f7368617265730765046b0000000925646561646c696e6507650663036e000000092572656365697665720663036e0000000925726566657272616c00000012256469766573745f696d62616c616e636564086504620000000825706f6f6c5f6964076504620000000725736861726573076504620000000c25746f6b656e5f696e646578076504620000000f256d696e5f616d6f756e745f6f75740765046b0000000925646561646c696e6507650663036e000000092572656365697665720663036e0000000925726566657272616c00000010256469766573745f6f6e655f636f696e07640764086504620000000825706f6f6c5f696407650462000000072573686172657307650860036203620000000b25696e5f616d6f756e74730765046b0000000925646561646c696e6507650663036e000000092572656365697665720663036e0000000925726566657272616c0000000725696e766573740864086504620000000825706f6f6c5f696404620000000725616d6f756e740000000425616464086504620000000825706f6f6c5f696404620000000725616d6f756e74000000072572656d6f766500000006257374616b65086504620000000825706f6f6c5f69640765046200000009256964785f66726f6d0765046200000007256964785f746f076504620000000725616d6f756e74076504620000000f256d696e5f616d6f756e745f6f75740765046b0000000925646561646c696e6507650663036e000000092572656365697665720663036e0000000925726566657272616c0000000525737761700765046e000000062561646d696e0765046e000000112564656661756c745f726566657272616c07650666036e00000009256d616e6167657273076504620000000c25706f6f6c735f636f756e74076508610362076003620764046e0000000525666131320865046e0000000e25746f6b656e5f6164647265737304620000000925746f6b656e5f696400000004256661320000000725746f6b656e7307650861036903620000000b25706f6f6c5f746f5f6964076508610362076504620000000c25696e697469616c5f415f660765046b0000000f25696e697469616c5f415f74696d65076504620000000b256675747572655f415f660765046b0000000e256675747572655f415f74696d65076508600362076504620000000725726174655f66076504620000001725707265636973696f6e5f6d756c7469706c6965725f660462000000092572657365727665730000000c25746f6b656e735f696e666f07650865046200000005256c705f66076504620000000a257374616b6572735f66046200000006257265665f660000000425666565076508650860036203620000000e25616363756d756c61746f725f6604620000000d25746f74616c5f7374616b656400000013257374616b65725f616363756d756c61746f7204620000000d25746f74616c5f737570706c790000000625706f6f6c73076508610765036e0362036200000007256c6564676572076508610765036e03620566036e0000000b25616c6c6f77616e636573076508610764046e0000000525666131320865046e0000000e25746f6b656e5f6164647265737304620000000925746f6b656e5f6964000000042566613203620000000c256465765f72657761726473076508610765036e0764046e0000000525666131320865046e0000000e25746f6b656e5f6164647265737304620000000925746f6b656e5f6964000000042566613203620000001125726566657272616c5f72657761726473076508610765036e036207650462000000082562616c616e6365086003620765046200000009257265776172645f6604620000000925666f726d65725f6600000009256561726e696e677300000010257374616b6572735f62616c616e636507650865046e0000000e25746f6b656e5f6164647265737304620000000925746f6b656e5f69640000000c2571756970755f746f6b656e0765046e0000001025666163746f72795f6164647265737304590000000825737461727465640765055f036d0765046e000000062561646d696e0765046e000000112564656661756c745f726566657272616c07650666036e00000009256d616e6167657273076504620000000c25706f6f6c735f636f756e74076508610362076003620764046e0000000525666131320865046e0000000e25746f6b656e5f6164647265737304620000000925746f6b656e5f696400000004256661320000000725746f6b656e7307650861036903620000000b25706f6f6c5f746f5f6964076508610362076504620000000c25696e697469616c5f415f660765046b0000000f25696e697469616c5f415f74696d65076504620000000b256675747572655f415f660765046b0000000e256675747572655f415f74696d65076508600362076504620000000725726174655f66076504620000001725707265636973696f6e5f6d756c7469706c6965725f660462000000092572657365727665730000000c25746f6b656e735f696e666f07650865046200000005256c705f66076504620000000a257374616b6572735f66046200000006257265665f660000000425666565076508650860036203620000000e25616363756d756c61746f725f6604620000000d25746f74616c5f7374616b656400000013257374616b65725f616363756d756c61746f7204620000000d25746f74616c5f737570706c790000000625706f6f6c73076508610765036e0362036200000007256c6564676572076508610765036e03620566036e0000000b25616c6c6f77616e636573076508610764046e0000000525666131320865046e0000000e25746f6b656e5f6164647265737304620000000925746f6b656e5f6964000000042566613203620000000c256465765f72657761726473076508610765036e0764046e0000000525666131320865046e0000000e25746f6b656e5f6164647265737304620000000925746f6b656e5f6964000000042566613203620000001125726566657272616c5f72657761726473076508610765036e036207650462000000082562616c616e6365086003620765046200000009257265776172645f6604620000000925666f726d65725f6600000009256561726e696e677300000010257374616b6572735f62616c616e636507650865046e0000000e25746f6b656e5f6164647265737304620000000925746f6b656e5f69640000000c2571756970755f746f6b656e0765046e0000001025666163746f72795f616464726573730459000000082573746172746564072f020000000203270200000004034c03200521000303160570000203420326032105710002031705500001034c031603420200000b2b0571000203210571000303160529001c072c020000000203200200000002032705700003052100030529000a05210003072e0200000044072e020000001c072e0200000008032007430362000102000000080320074303620004020000001c072e0200000008032007430362000002000000080320074303620003020000000803200743036200020329072f020000000203270200000004034c032005700003034c050d075e07650764076407640865065f0765046e00000006256f776e657204620000000925746f6b656e5f696400000009257265717565737473065a055f07650865046e00000006256f776e657204620000000925746f6b656e5f69640000000825726571756573740462000000082562616c616e6365000000092563616c6c6261636b0000000b2562616c616e63655f6f66086504620000000925746f6b656e5f6964065a0362000000092572656365697665720000000d25746f74616c5f737570706c790764065f0765046e000000062566726f6d5f065f0765046e0000000425746f5f076504620000000925746f6b656e5f696404620000000725616d6f756e74000000042574787300000009257472616e73666572086504620000000925746f6b656e5f69640860036803690000000b25746f6b656e5f696e666f00000010257570646174655f6d65746164617461065f07640865046e00000006256f776e65720765046e00000009256f70657261746f7204620000000925746f6b656e5f69640000000d256164645f6f70657261746f720865046e00000006256f776e65720765046e00000009256f70657261746f7204620000000925746f6b656e5f6964000000102572656d6f76655f6f70657261746f7200000011257570646174655f6f70657261746f727307650865046e000000062561646d696e0765046e000000112564656661756c745f726566657272616c07650666036e00000009256d616e6167657273076504620000000c25706f6f6c735f636f756e74076508610362076003620764046e0000000525666131320865046e0000000e25746f6b656e5f6164647265737304620000000925746f6b656e5f696400000004256661320000000725746f6b656e7307650861036903620000000b25706f6f6c5f746f5f6964076508610362076504620000000c25696e697469616c5f415f660765046b0000000f25696e697469616c5f415f74696d65076504620000000b256675747572655f415f660765046b0000000e256675747572655f415f74696d65076508600362076504620000000725726174655f66076504620000001725707265636973696f6e5f6d756c7469706c6965725f660462000000092572657365727665730000000c25746f6b656e735f696e666f07650865046200000005256c705f66076504620000000a257374616b6572735f66046200000006257265665f660000000425666565076508650860036203620000000e25616363756d756c61746f725f6604620000000d25746f74616c5f7374616b656400000013257374616b65725f616363756d756c61746f7204620000000d25746f74616c5f737570706c790000000625706f6f6c73076508610765036e0362036200000007256c6564676572076508610765036e03620566036e0000000b25616c6c6f77616e636573076508610764046e0000000525666131320865046e0000000e25746f6b656e5f6164647265737304620000000925746f6b656e5f6964000000042566613203620000000c256465765f72657761726473076508610765036e0764046e0000000525666131320865046e0000000e25746f6b656e5f6164647265737304620000000925746f6b656e5f6964000000042566613203620000001125726566657272616c5f72657761726473076508610765036e036207650462000000082562616c616e6365086003620765046200000009257265776172645f6604620000000925666f726d65725f6600000009256561726e696e677300000010257374616b6572735f62616c616e636507650865046e0000000e25746f6b656e5f6164647265737304620000000925746f6b656e5f69640000000c2571756970755f746f6b656e0765046e0000001025666163746f72795f616464726573730459000000082573746172746564000000082573746f72616765076508610368036900000009256d65746164617461076508610362076504620000000925746f6b656e5f69640860036803690000000b25746f6b656e5f696e666f0000000f25746f6b656e5f6d6574616461746107650861036203690000000e2561646d696e5f6c616d6264617307650861036203690000000c256465785f6c616d626461730861036203690000000e25746f6b656e5f6c616d626461730765055f036d07650865046e000000062561646d696e0765046e000000112564656661756c745f726566657272616c07650666036e00000009256d616e6167657273076504620000000c25706f6f6c735f636f756e74076508610362076003620764046e0000000525666131320865046e0000000e25746f6b656e5f6164647265737304620000000925746f6b656e5f696400000004256661320000000725746f6b656e7307650861036903620000000b25706f6f6c5f746f5f6964076508610362076504620000000c25696e697469616c5f415f660765046b0000000f25696e697469616c5f415f74696d65076504620000000b256675747572655f415f660765046b0000000e256675747572655f415f74696d65076508600362076504620000000725726174655f66076504620000001725707265636973696f6e5f6d756c7469706c6965725f660462000000092572657365727665730000000c25746f6b656e735f696e666f07650865046200000005256c705f66076504620000000a257374616b6572735f66046200000006257265665f660000000425666565076508650860036203620000000e25616363756d756c61746f725f6604620000000d25746f74616c5f7374616b656400000013257374616b65725f616363756d756c61746f7204620000000d25746f74616c5f737570706c790000000625706f6f6c73076508610765036e0362036200000007256c6564676572076508610765036e03620566036e0000000b25616c6c6f77616e636573076508610764046e0000000525666131320865046e0000000e25746f6b656e5f6164647265737304620000000925746f6b656e5f6964000000042566613203620000000c256465765f72657761726473076508610765036e0764046e0000000525666131320865046e0000000e25746f6b656e5f6164647265737304620000000925746f6b656e5f6964000000042566613203620000001125726566657272616c5f72657761726473076508610765036e036207650462000000082562616c616e6365086003620765046200000009257265776172645f6604620000000925666f726d65725f6600000009256561726e696e677300000010257374616b6572735f62616c616e636507650865046e0000000e25746f6b656e5f6164647265737304620000000925746f6b656e5f69640000000c2571756970755f746f6b656e0765046e0000001025666163746f72795f616464726573730459000000082573746172746564000000082573746f72616765076508610368036900000009256d65746164617461076508610362076504620000000925746f6b656e5f69640860036803690000000b25746f6b656e5f696e666f0000000f25746f6b656e5f6d6574616461746107650861036203690000000e2561646d696e5f6c616d6264617307650861036203690000000c256465785f6c616d626461730861036203690000000e25746f6b656e5f6c616d62646173072f020000000203270200000004034c0320057100020342032609910000015701000000146765745f726566657272616c5f72657761726473055f0765046e0000000525757365720864046e0000000525666131320865046e0000000e25746f6b656e5f6164647265737304620000000925746f6b656e5f696400000004256661320000000625746f6b656e055f07650865046e0000000525757365720864046e0000000525666131320865046e0000000e25746f6b656e5f6164647265737304620000000925746f6b656e5f696400000004256661320000000625746f6b656e000000082572657175657374046200000007257265776172640200000078037a053d07650765036e0764036e0765036e03620362034c053d0765036e0764036e0765036e0362034c05520200000002031b0552020000003a034c0743036200000521000403160529001505210004031705210005031603420329072f02000000000200000004034c0320057000020342031b034c0320000000000991000001cf010000000f6765745f7374616b65725f696e666f055f0765046e00000005257573657204620000000825706f6f6c5f6964055f07650865046e00000005257573657204620000000825706f6f6c5f696400000008257265717565737408650462000000082562616c616e63650860036203620000000825726577617264730000000525696e666f020000014307430368010000000c6e6f742d6c61756e63686564034c037a053d07650765036e036207650362076003620362034c053d0765036e0362034c05520200000002031b055202000000f0052100040521000403160529000d0521000303170329072f020000000203270200000004034c0320052100050521000503160529001705210004031705210005031603420329072f020000000203270200000004034c03200321031705380200000073037a074303620000052100050529000d0316057000020329072f02000000000200000004034c0320052100030316033a07430362008090dfc04a05210003031705700002034b031105700002031603120322072f02000000130743036801000000084449562062792030032702000000000316057000020320034c0316034205700002034c057000020342031b034c05700002052000020000000009910000007c01000000086765745f6665657303620765046200000005256c705f66076504620000000a257374616b6572735f66046200000006257265665f66020000003d037a07430368010000000c6e6f742d6c61756e636865640570000203160529000d057000020329072f020000000203270200000004034c03200529000b000000000991000001780100000006766965775f4103620362020000016407430368010000000f74696d657374616d702d6572726f72034c037a07430368010000000c6e6f742d6c61756e636865640570000203160529000d057000020329072f020000000203270200000004034c03200743036200a401034c0321057100020529000505210003052900070342052100030316057000030529000305700002037a0321034003190337072c020000009905210006052100040340034b0356072f020000000203270200000004034c0320057000060570000405700003034b0356072f020000000203270200000004034c03200521000405210004034b0311034c05710002033a0322072f0200000013074303680100000008444956206279203003270200000000031605210003057000020319032a072c020000000203120200000006034c034b03110200000010057000020570000305700005052000040322072f02000000130743036801000000084449562062792030032702000000000316000000000991000009f201000000066765745f6479076504620000000825706f6f6c5f6964076504620000000225690765046200000002256a046200000003256478036202000009b307430362008080a0f6f4acdbe01b0743036200a40107430368010000000b77726f6e672d696e64657807430368010000000f74696d657374616d702d6572726f7207430368010000001176616c75652d6e6f742d6e61747572616c05700005037a07430368010000000c6e6f742d6c61756e636865640521000303160529000d0521000303160329072f020000000203270200000004034c03200321052100030529000603420521000305290005057000030529000305700002037a034c032105710002052900090538020000003d03170521000c034c03210571000205290004057000020316033a0322072f020000001307430368010000000844495620627920300327020000000003160521000a034c032105710002052100060329072f020000000203270200000004034c03200521000b05210003052100080329072f020000000203270200000004034c03200521000c0521000605290009052100080329072f020000000203270200000004034c03200570000c05210007052900090521000a0329072f020000000203270200000004034c03200521000f05700002031605700006033a0322072f020000001307430368010000000844495620627920300327020000000003160570000303120570000405700004057000020342057000050570000505700002037a0521000505290009034505210005052100050319033c072c0200000000020000001b0743036801000000106661696c656420617373657274696f6e032703210521000603190337072c0200000000020000001b0743036801000000106661696c656420617373657274696f6e032703210521000503190337072c0200000000020000001b0743036801000000106661696c656420617373657274696f6e0327052100060529000505210007052900070342052100070316052100080529000305700002037a0321034003190337072c020000009905210010052100040340034b0356072f020000000203270200000004034c0320057000100570000405700003034b0356072f020000000203270200000004034c03200521000405210004034b0311034c05710002033a0322072f0200000013074303680100000008444956206279203003270200000000031605210003057000020319032a072c020000000203120200000006034c034b0311020000001005700002057000030570000f05200004034c032105710002034c032105710002033a034c05210005074303620000034c0321057100020552020000000403170312034c0321057100020345032105700004033a0743036200000521000403420521001205700005057000050570000505700005054200050931000001c4076507650362076503620765036207650760036203620362076503620362076503590765036203620200000197037a057a000505700005074303620001034c0321057100020317052100030316034b03110319032a072c020000014503210316057000050552020000003d034c05210005057000020317033a05210003031605700002033a0322072f02000000130743036801000000084449562062792030032702000000000316034c0321031605500002032105210003074303620001052100070312033a0521000805210004031607430368010000000f77726f6e672d707265636973696f6e0521000b05210009034b0356072f020000000203270200000004034c0320033a0322072f0200000013074303680100000008444956206279203003270200000000031603120570000203160570000505700004033a057000060570000605700006033a0322072f020000001307430368010000000844495620627920300327020000000003160312033a0322072f020000001307430368010000000844495620627920300327020000000003160550000107430359030a0342020000001e034c0570000205700003057000040570000505200005074303590303034200000000034c0373034c07430359030a0534020000000d051f020000000203210326037a034c03200316074303620000034c032105710002034205700005055202000000ad032105710002031605210009034c0321057100020319033c072c020000007705210008034c03190325072c0200000008034c0320052100050200000004034c0317032105210003031703120570000203160342032103170521000605210004033a05210005057000030316033a0322072f020000001307430368010000000844495620627920300327020000000003160342034c0342020000001005700002052000020743036200000342031705700003057000040570000505700006052000040570000305710002032105710003031703420570000305700003031605700002037a0570000405290009034505210005033a0521000b0521000405700005033a033a0322072f02000000130743036801000000084449562062792030032702000000000316057000030570000905210005033a0322072f0200000013074303680100000008444956206279203003270200000000031605700002031205210003074303620000034205210009057000040570000405700004054200040931000000f6076507650362076503620765036203680765036203620765035907650362036202000000d1037a057a000405700004074303620001034c0321057100020316052100030317034b03110319032a072c020000008303210317034c031703420570000405700004057000030521000403170743035b0002033a0312034b0356072f020000000203270200000004034c0320057100020321057100030317052100040317033a03120322072f02000000130743036801000000084449562062792030032702000000000316034c0316034207430359030a0342020000001a034c05700002057000030570000405200004074303590303034200000000034c0373034c07430359030a0534020000000d051f020000000203210326037a034c03200317057000050743035b00010570000205700004034b034b0356072f020000000203270200000004034c0320034c03160570000405700002033a0322072f0200000013074303680100000008444956206279203003270200000000031607430362008090dfc04a034c03210571000207430368010000000b6e6f2d6665652d766965770570000503160529001b034f079001000000076465765f6665650362072f020000000203270200000004034c0320057000040529000b032105710002052900040521000305290003057000030316031203120312033a0322072f0200000013074303680100000008444956206279203003270200000000031607430368010000000c6665652d6f766572666c6f77034c05700002034b0356072f020000000203270200000004034c032000000000099100000bf4010000001463616c635f6469766573745f6f6e655f636f696e076504620000000825706f6f6c5f6964076504620000000d25746f6b656e5f616d6f756e74046200000002256903620200000ba70743036200a40107430368010000000b77726f6e672d696e64657807430368010000000f74696d657374616d702d6572726f7207430368010000001176616c75652d6e6f742d6e61747572616c052100040521000405210003054200030931000003270765076503680765036803620765076507650362036207650760036203620362076503620765036b076503620765036b076507600362076503620765036203620765076503620765036203620765076507600362036203620362036202000002c6037a057a000305700003037a037a037a05700002037a0521000505290009034505700007034c0321057100020521000703190337072c0200000002032002000000020327032105700004033a0743036200000521000503420570000305520200000083034c052100060521000303160319033c072c0200000059034c0317032105210003031703120570000203160342032103170521000505210004033a05210007057000030316033a0322072f020000001307430368010000000844495620627920300327020000000003160342034c0342020000000c034c032007430362000003420317057000020570000405200002057000030570000305210003031703420570000305700003031605700002037a0570000405290009034505210005033a052100070521000405700005033a033a0322072f02000000130743036801000000084449562062792030032702000000000316057000030570000505210005033a0322072f0200000013074303680100000008444956206279203003270200000000031605700002031205210003074303620000034205710004054200040931000000f6076507650362076503620765036203680765036203620765035907650362036202000000d1037a057a000405700004074303620001034c0321057100020316052100030317034b03110319032a072c020000008303210317034c031703420570000405700004057000030521000403170743035b0002033a0312034b0356072f020000000203270200000004034c0320057100020321057100030317052100040317033a03120322072f02000000130743036801000000084449562062792030032702000000000316034c0316034207430359030a0342020000001a034c05700002057000030570000405200004074303590303034200000000034c0373034c07430359030a0534020000000d051f020000000203210326037a034c0320031700000000034c037305700005037a07430368010000000c6e6f742d6c61756e636865640521000303160529000d0521000303160329072f020000000203270200000004034c0320032105290005034c032105710002052900070342034c0321057100020316052100030529000305700002037a0321034003190337072c02000000990521000a052100040340034b0356072f020000000203270200000004034c03200570000a0570000405700003034b0356072f020000000203270200000004034c03200521000405210004034b0311034c05710002033a0322072f0200000013074303680100000008444956206279203003270200000000031605210003057000020319032a072c020000000203120200000006034c034b0311020000001005700002057000030570000905200004034c07430368010000000b6e6f2d6665652d766965770570000403160529001b034f079001000000076465765f6665650362072f020000000203270200000004034c03200521000405290004034205700003052900030570000305700002037a052100050529000905380200000047031707430362008080a0f6f4acdbe01b034c03210571000205290004057000020316033a0322072f0200000013074303680100000008444956206279203003270200000000031605210004034c032105710002074303620000034c0321057100020552020000000403170312034c0321057100020345032105700004033a0743036200000521000403420570000e05700005057000050570000505700005054200050931000001c4076507650362076503620765036207650760036203620362076503620362076503590765036203620200000197037a057a000505700005074303620001034c0321057100020317052100030316034b03110319032a072c020000014503210316057000050552020000003d034c05210005057000020317033a05210003031605700002033a0322072f02000000130743036801000000084449562062792030032702000000000316034c0321031605500002032105210003074303620001052100070312033a0521000805210004031607430368010000000f77726f6e672d707265636973696f6e0521000b05210009034b0356072f020000000203270200000004034c0320033a0322072f0200000013074303680100000008444956206279203003270200000000031603120570000203160570000505700004033a057000060570000605700006033a0322072f020000001307430368010000000844495620627920300327020000000003160312033a0322072f020000001307430368010000000844495620627920300327020000000003160550000107430359030a0342020000001e034c0570000205700003057000040570000505200005074303590303034200000000034c0373034c07430359030a0534020000000d051f020000000203210326037a034c0320031605210009052100080529000e0521000305210009033a0322072f0200000013074303680100000008444956206279203003270200000000031605210003034b0356072f020000000203270200000004034c032005210008034c03210571000205210005034205210006052100090342034203420521000a034c032605700005052100090529000b0321057100020529000405210003052900030570000303160312031203120521000505380200000181037a05210008034c03190325072c020000004f0521000c05210004052100070521000705210005033a0322072f02000000130743036801000000084449562062792030032702000000000316034b0356072f020000000203270200000004034c0320020000004f0521000c052100060521000605210004033a0322072f0200000013074303680100000008444956206279203003270200000000031605210003034b0356072f020000000203270200000004034c03200521000d07430362008090dfc04a0521000d0529000903450521000607430368010000001277726f6e672d746f6b656e732d636f756e7407430362000105210004034b0356072f020000000203270200000004034c0320074303620004033a05710002033a0322072f0200000013074303680100000008444956206279203003270200000000031605700003033a0322072f0200000013074303680100000008444956206279203003270200000000031605700002034b0356072f020000000203270200000004034c0320034c05700004052000020521000b034c032105710002052100070329072f020000000203270200000004034c03200521000b0521000a05700005057000040342052100070570000803420342034205700008034c032605700002034b0356072f020000000203270200000004034c0320052100080521000705290009052100060329072f020000000203270200000004034c0320052900030570000805700004057000050329072f020000000203270200000004034c0320034c032105710002052100080743035b000105700005034b0356072f020000000203270200000004034c03200322072f0200000013074303680100000008444956206279203003270200000000031605700002052100070570000405700004034b0356072f020000000203270200000004034c03200322072f020000001307430368010000000844495620627920300327020000000003160743036801000000106c6f772d746f74616c2d737570706c7905700003057000040529000e034b0356072f020000000203270200000004034c0320057000030521000405700003034b0356072f020000000203270200000004034c032005700002054200030316000000000991000000b601000000116765745f746f6b5f7065725f736861726503620760036203620200000093037a07430368010000000c6e6f742d6c61756e636865640570000203160529000d057000020329072f020000000203270200000004034c0320032105290009053802000000490317034c0321057100020529000e07430362008080a0f6f4acdbe01b0570000205290004033a0322072f02000000130743036801000000084449562062792030032702000000000316034c032000000000099100000090010000000d6765745f746f6b656e5f6d61700362076003620764046e0000000525666131320865046e0000000e25746f6b656e5f6164647265737304620000000925746f6b656e5f696400000004256661320200000039037a07430368010000000c6e6f742d6c61756e6368656405700002031605290009057000020329072f020000000203270200000004034c032000000000099100000068010000000c6765745f72657365727665730362076003620362020000004a037a07430368010000000c6e6f742d6c61756e636865640570000203160529000d057000020329072f020000000203270200000004034c0320052900090538020000000603170529000400000000034205700002032103210571000403160570000403160529000905210004031705700005034c0346034c03500550000905500001032103210571000203160743036200010570000303160529000703120550000705500001032103160529000d053d036d057000030316031b0342034c03210571000203160529000b052100030316052900050342052100030316052900030521000403160529000e05700002037a05700004037a0570000403480339033f072c02000002ec074303620080897a0570000305210006033a0322072f02000000130743036801000000084449562062792030032702000000000316032105700005034b031105700002052100050544036e0521000303420377034805700002037a034c072e02000000630521000b034c06550765046e000000052566726f6d0765046e0000000325746f0462000000062576616c756500000009257472616e73666572072f020000000203270200000004034c03200743036a000005700002057000040342057000030342034d02000000b50521000b034c03210571000203160655055f0765046e000000062566726f6d5f065f0765046e0000000425746f5f076504620000000925746f6b656e5f696404620000000725616d6f756e74000000042574787300000009257472616e73666572072f020000000203270200000004034c03200743036a0000053d0765036e055f0765036e076503620362053d0765036e076503620362057000050570000503170570000705420003031b057000040342031b034d031b057000040544036e0570000303420743036e0a00000016000098b9732c83017e938ba48cb91cf53e5f919dc844034805700002037a034c072e020000006305700008034c06550765046e000000052566726f6d0765046e0000000325746f0462000000062576616c756500000009257472616e73666572072f020000000203270200000004034c03200743036a000005700002057000040342057000030342034d02000000b505700008034c03210571000203160655055f0765046e000000062566726f6d5f065f0765046e0000000425746f5f076504620000000925746f6b656e5f696404620000000725616d6f756e74000000042574787300000009257472616e73666572072f020000000203270200000004034c03200743036a0000053d0765036e055f0765036e076503620362053d0765036e076503620362057000050570000503170570000705420003031b057000040342031b034d031b034c057000020312034c0342020000001605700002057000030570000405700006052000040342034c032103160521000303170550000d05500001034c03160342"}]},{"int":"40068"}]`),
						BigMapDiffs: []*bigmapdiff.BigMapDiff{
							{
								Ptr:        40064,
								Key:        []byte(`{"bytes":"05070702000000460704000005050a00000016012a0b69e71ece4da314f9904bfb7f8a8d3d373c530007040001050807070a000000160107a2fc7b796ea23ad82e768221a78c86c77c64050000000a00000016000042a7bb84edce2af4cc8ab0bc83ded699efc9300a"}`),
								Value:      []byte(`{"bytes":"011e4e248aea7b65b1941614f481e5801b0ebaacdc00"}`),
								KeyHash:    "exprtjtvPgEDWmVoB1v5jdmX6HGjTZZFYoWxrtr2s4pT8GHr3XJwPi",
								Level:      381735,
								Timestamp:  timestamp,
								Contract:   "KT1Jk8LRDoj6LkopYZwRq5ZEWBhYv8nVc6e6",
								ProtocolID: 6,
							},
						},
					}, {
						Kind: types.OperationKindTransaction,
						Hash: "oozvzXiZmVW9QtYjKmDuYqoHNCEvt32FwM2cUgQee2S1SGWgumA",
						Source: account.Account{
							Address: "KT1Jk8LRDoj6LkopYZwRq5ZEWBhYv8nVc6e6",
							Type:    types.AccountTypeContract,
						},
						Destination: account.Account{
							Address: "KT19H9YbHqsxFTayap7aTEfbcnyPeALKYgt9",
							Type:    types.AccountTypeContract,
						},
						Initiator: account.Account{
							Address: "tz1RiUE3Ao53juAz4uDYx1J3tHJMye6jPfhp",
							Type:    types.AccountTypeTz,
						},
						Status:          types.OperationStatusApplied,
						Nonce:           newInt64Ptr(2),
						Timestamp:       timestamp,
						Level:           381735,
						ProtocolID:      6,
						Internal:        true,
						Counter:         34005,
						Parameters:      []byte(`{"entrypoint":"transfer","value":[{"prim":"Pair","args":[{"bytes":"000042a7bb84edce2af4cc8ab0bc83ded699efc9300a"},[{"prim":"Pair","args":[{"bytes":"000098b9732c83017e938ba48cb91cf53e5f919dc844"},{"prim":"Pair","args":[{"int":"0"},{"int":"0"}]}]}]]}]}`),
						Entrypoint:      types.NullString{Str: "transfer", Valid: true},
						DeffatedStorage: []byte(`[{"int":"34843"},{"int":"34844"},{"int":"34845"},{"int":"34846"},[],{"bytes":"0000dd513ae2c8bb08e7463d04535c28f54be3722286"},{"bytes":"0000dd513ae2c8bb08e7463d04535c28f54be3722286"},{"int":"1"}]`),
						BigMapDiffs: []*bigmapdiff.BigMapDiff{
							{
								Ptr:        34843,
								Key:        []byte(`{"bytes":"000098b9732c83017e938ba48cb91cf53e5f919dc844"}`),
								Value:      []byte(`{"prim":"Pair","args":[[{"prim":"Elt","args":[{"int":"0"},{"int":"0"}]}],{"prim":"Pair","args":[{"int":"1649845410"},[]]}]}`),
								KeyHash:    "exprvHCaW3fmGXHKV22BNVUkR7TLNCca84z3ANPRSaY9ubZUj7QrF6",
								Level:      381735,
								Timestamp:  timestamp,
								Contract:   "KT19H9YbHqsxFTayap7aTEfbcnyPeALKYgt9",
								ProtocolID: 6,
							}, {
								Ptr:        34843,
								Key:        []byte(`{"bytes":"000042a7bb84edce2af4cc8ab0bc83ded699efc9300a"}`),
								Value:      []byte(`{"prim":"Pair","args":[[{"prim":"Elt","args":[{"int":"0"},{"int":"9999999989999000000"}]}],{"prim":"Pair","args":[{"int":"1649845320"},[{"bytes":"01308a4d463c798401eb231fb386ac223e8d44987400"},{"bytes":"016f7656dd6c6df9f8294efae0235ccb7f27025c4900"},{"bytes":"01c14ac0a868ad16d8115a2153bd700cdf0d7898f100"}]]}]}`),
								KeyHash:    "exprtj5G3z2kcdmxvy3y9nFqs57enTtDSyrRttFHm7tL7PdLqR2ek5",
								Level:      381735,
								Timestamp:  timestamp,
								Contract:   "KT19H9YbHqsxFTayap7aTEfbcnyPeALKYgt9",
								ProtocolID: 6,
							},
						},
					}, {
						Kind: types.OperationKindTransaction,
						Hash: "oozvzXiZmVW9QtYjKmDuYqoHNCEvt32FwM2cUgQee2S1SGWgumA",
						Source: account.Account{
							Address: "KT1Jk8LRDoj6LkopYZwRq5ZEWBhYv8nVc6e6",
							Type:    types.AccountTypeContract,
						},
						Destination: account.Account{
							Address: "KT19H9YbHqsxFTayap7aTEfbcnyPeALKYgt9",
							Type:    types.AccountTypeContract,
						},
						Initiator: account.Account{
							Address: "tz1RiUE3Ao53juAz4uDYx1J3tHJMye6jPfhp",
							Type:    types.AccountTypeTz,
						},
						Status:              types.OperationStatusApplied,
						Nonce:               newInt64Ptr(1),
						Timestamp:           timestamp,
						Level:               381735,
						ConsumedGas:         5715,
						Burned:              91000,
						StorageSize:         4113,
						Counter:             34005,
						PaidStorageSizeDiff: 91,
						ProtocolID:          6,
						Internal:            true,
						Parameters:          []byte(`{"entrypoint":"transfer","value":[{"prim":"Pair","args":[{"bytes":"000042a7bb84edce2af4cc8ab0bc83ded699efc9300a"},[{"prim":"Pair","args":[{"bytes":"016f7656dd6c6df9f8294efae0235ccb7f27025c4900"},{"prim":"Pair","args":[{"int":"0"},{"int":"0"}]}]}]]}]}`),
						Entrypoint:          types.NullString{Str: "transfer", Valid: true},
						DeffatedStorage:     []byte(`[{"int":"34843"},{"int":"34844"},{"int":"34845"},{"int":"34846"},[],{"bytes":"0000dd513ae2c8bb08e7463d04535c28f54be3722286"},{"bytes":"0000dd513ae2c8bb08e7463d04535c28f54be3722286"},{"int":"1"}]`),
						BigMapDiffs: []*bigmapdiff.BigMapDiff{
							{
								Ptr:        34843,
								Key:        []byte(`{"bytes":"000042a7bb84edce2af4cc8ab0bc83ded699efc9300a"}`),
								Value:      []byte(`{"prim":"Pair","args":[[{"prim":"Elt","args":[{"int":"0"},{"int":"9999999989999000000"}]}],{"prim":"Pair","args":[{"int":"1649845320"},[{"bytes": "01308a4d463c798401eb231fb386ac223e8d44987400"},{"bytes": "016f7656dd6c6df9f8294efae0235ccb7f27025c4900"},{"bytes": "01c14ac0a868ad16d8115a2153bd700cdf0d7898f100"}]]}]}`),
								KeyHash:    "exprtj5G3z2kcdmxvy3y9nFqs57enTtDSyrRttFHm7tL7PdLqR2ek5",
								Level:      381735,
								Timestamp:  timestamp,
								Contract:   "KT19H9YbHqsxFTayap7aTEfbcnyPeALKYgt9",
								ProtocolID: 6,
							}, {
								Ptr:        34843,
								Key:        []byte(`{"bytes":"016f7656dd6c6df9f8294efae0235ccb7f27025c4900"}`),
								Value:      []byte(`{"prim":"Pair","args":[[{"prim":"Elt","args":[{"int":"0"},{"int":"0"}]}],{"prim":"Pair","args":[{"int":"1649854355"},[]]}]}`),
								KeyHash:    "exprtXp227aTXC9hWQC3H6y5M9rP38UD1qSHDeaEL4hXqxb5ceurBr",
								Level:      381735,
								Timestamp:  timestamp,
								Contract:   "KT19H9YbHqsxFTayap7aTEfbcnyPeALKYgt9",
								ProtocolID: 6,
							},
						},
					}, {
						Kind: types.OperationKindOrigination,
						Hash: "oozvzXiZmVW9QtYjKmDuYqoHNCEvt32FwM2cUgQee2S1SGWgumA",
						Source: account.Account{
							Address: "KT1Jk8LRDoj6LkopYZwRq5ZEWBhYv8nVc6e6",
							Type:    types.AccountTypeContract,
						},
						Destination: account.Account{
							Address: "KT1BM1SyQnTzNU1J8TZv5Mdj4ScuTgNKH5uj",
							Type:    types.AccountTypeContract,
						},
						Initiator: account.Account{
							Address: "tz1RiUE3Ao53juAz4uDYx1J3tHJMye6jPfhp",
							Type:    types.AccountTypeTz,
						},
						Status:                             types.OperationStatusApplied,
						Nonce:                              newInt64Ptr(0),
						Timestamp:                          timestamp,
						Level:                              381735,
						ConsumedGas:                        11494,
						StorageSize:                        26815,
						Counter:                            34005,
						PaidStorageSizeDiff:                26815,
						Burned:                             27072000,
						AllocatedDestinationContractBurned: 257000,
						Tags:                               16640,
						ProtocolID:                         6,
						Internal:                           true,
						BigMapDiffs: []*bigmapdiff.BigMapDiff{
							{
								Ptr:        40078,
								Key:        []byte(`{"int":"0"}`),
								Value:      []byte(`{"prim":"Pair","args":[{"int":"0"},[{"prim":"Elt","args":[{"string":"decimals"},{"bytes":"3138"}]},{"prim":"Elt","args":[{"string":"description"},{"bytes":"4c697175696469747920506f6f6c20746f6b656e206f662051756970755377617020537461626c6520444558"}]},{"prim":"Elt","args":[{"string":"name"},{"bytes":"537461626c652044455820517569707553776170204c5020746f6b656e"}]},{"prim":"Elt","args":[{"string":"shouldPreferSymbol"},{"bytes":"74727565"}]},{"prim":"Elt","args":[{"string":"symbol"},{"bytes":"736451504c50"}]},{"prim":"Elt","args":[{"string":"thumbnailUri"},{"bytes":"697066733a2f2f516d5558464a787747456a6b6e5035444b4e774d777357587672593847326a554d62786a64424533656b50384450"}]}]]}`),
								KeyHash:    "exprtZBwZUeYYYfUs9B9Rg2ywHezVHnCCnmF9WsDQVrs582dSK63dC",
								Level:      381735,
								Timestamp:  timestamp,
								Contract:   "KT1BM1SyQnTzNU1J8TZv5Mdj4ScuTgNKH5uj",
								ProtocolID: 6,
							}, {
								Ptr:        40077,
								Key:        []byte(`{"string":"dex"}`),
								Value:      []byte(`{"bytes":"7b226e616d65223a2251756970755377617020537461626c652044455820706f6f6c222c2276657273696f6e223a2276312e302e30222c226465736372697074696f6e223a22506f6f6c20666f72207377617070696e6720746f6b656e732077697468206c6f7720736c697070616765222c22617574686f7273223a5b224d6164666973682e536f6c7574696f6e73203c68747470733a2f2f7777772e6d6164666973682e736f6c7574696f6e733e225d2c22736f75726365223a7b22746f6f6c73223a5b224c69676f222c22466c657874657361225d2c226c6f636174696f6e223a2268747470733a2f2f6769746875622e636f6d2f6d6164666973682d736f6c7574696f6e732f7175697075737761702d737461626c652d636f72652f626c6f622f6d61696e2f636f6e7472616374732f6d61696e2f6465782e6c69676f227d2c22686f6d6570616765223a2268747470733a2f2f7175697075737761702e636f6d222c22696e7465726661636573223a5b22545a49502d313220676974203137323866636665222c22545a49502d3136225d2c226572726f7273223a5b5d2c227669657773223a5b5d7d"}`),
								KeyHash:    "exprupXFsHdKsx5MFzpvfXijjEPBt18ipyXfJcPbpri7K9zYf1Fb7o",
								Level:      381735,
								Timestamp:  timestamp,
								Contract:   "KT1BM1SyQnTzNU1J8TZv5Mdj4ScuTgNKH5uj",
								ProtocolID: 6,
							}, {
								Ptr:        40077,
								Key:        []byte(`{"string":""}`),
								Value:      []byte(`{"bytes":"74657a6f732d73746f726167653a646578"}`),
								KeyHash:    "expru5X1yxJG6ezR2uHMotwMLNmSzQyh5t1vUnhjx4cS6Pv9qE1Sdo",
								Level:      381735,
								Timestamp:  timestamp,
								Contract:   "KT1BM1SyQnTzNU1J8TZv5Mdj4ScuTgNKH5uj",
								ProtocolID: 6,
							}, {
								Ptr:        40071,
								Key:        []byte(`{"int":"0"}`),
								Value:      []byte(`[{"int":"10000000"},{"int":"1649854355"},{"int":"10000000"},{"int":"1649854355"},[{"prim":"Elt","args":[{"int":"0"},{"prim":"Pair","args":[{"int":"1000000000000000000"},{"prim":"Pair","args":[{"int":"1000000000000000000000000"},{"int":"0"}]}]}]},{"prim":"Elt","args":[{"int":"1"},{"prim":"Pair","args":[{"int":"1000000000000000000"},{"prim":"Pair","args":[{"int":"100000000000000000000000000"},{"int":"0"}]}]}]}],{"prim":"Pair","args":[{"int":"0"},{"prim":"Pair","args":[{"int":"0"},{"int":"0"}]}]},{"prim":"Pair","args":[[],{"int":"0"}]},{"int":"0"}]`),
								KeyHash:    "exprtZBwZUeYYYfUs9B9Rg2ywHezVHnCCnmF9WsDQVrs582dSK63dC",
								Level:      381735,
								Timestamp:  timestamp,
								Contract:   "KT1BM1SyQnTzNU1J8TZv5Mdj4ScuTgNKH5uj",
								ProtocolID: 6,
							}, {
								Ptr:        40070,
								Key:        []byte(`{"bytes":"0502000000460704000005050a00000016012a0b69e71ece4da314f9904bfb7f8a8d3d373c530007040001050807070a000000160107a2fc7b796ea23ad82e768221a78c86c77c6405000000"}`),
								Value:      []byte(`{"int":"0"}`),
								KeyHash:    "expruLxJSi2bGyABA2WpjGBm5dd5zRUy84zWy2dRCHVGJSaNxrLL8t",
								Level:      381735,
								Timestamp:  timestamp,
								Contract:   "KT1BM1SyQnTzNU1J8TZv5Mdj4ScuTgNKH5uj",
								ProtocolID: 6,
							}, {
								Ptr:        40069,
								Key:        []byte(`{"int":"0"}`),
								Value:      []byte(`[{"prim":"Elt","args":[{"int":"0"},{"prim":"Left","args":[{"bytes":"012a0b69e71ece4da314f9904bfb7f8a8d3d373c5300"}]}]},{"prim":"Elt","args":[{"int":"1"},{"prim":"Right","args":[{"prim":"Pair","args":[{"bytes":"0107a2fc7b796ea23ad82e768221a78c86c77c640500"},{"int":"0"}]}]}]}]`),
								KeyHash:    "exprtZBwZUeYYYfUs9B9Rg2ywHezVHnCCnmF9WsDQVrs582dSK63dC",
								Level:      381735,
								Timestamp:  timestamp,
								Contract:   "KT1BM1SyQnTzNU1J8TZv5Mdj4ScuTgNKH5uj",
								ProtocolID: 6,
							},
						},
						BigMapActions: []*bigmapaction.BigMapAction{
							{
								Timestamp:      timestamp,
								Level:          381735,
								SourcePtr:      newInt64Ptr(40067),
								DestinationPtr: newInt64Ptr(40081),
								Action:         types.BigMapActionCopy,
								Address:        "KT1BM1SyQnTzNU1J8TZv5Mdj4ScuTgNKH5uj",
							}, {
								Timestamp:      timestamp,
								Level:          381735,
								DestinationPtr: newInt64Ptr(40080),
								Action:         types.BigMapActionCopy,
								Address:        "KT1BM1SyQnTzNU1J8TZv5Mdj4ScuTgNKH5uj",
							}, {
								Timestamp:      timestamp,
								Level:          381735,
								SourcePtr:      newInt64Ptr(40065),
								DestinationPtr: newInt64Ptr(40079),
								Action:         types.BigMapActionCopy,
								Address:        "KT1BM1SyQnTzNU1J8TZv5Mdj4ScuTgNKH5uj",
							}, {
								Timestamp:      timestamp,
								Level:          381735,
								DestinationPtr: newInt64Ptr(40078),
								Action:         types.BigMapActionCopy,
								Address:        "KT1BM1SyQnTzNU1J8TZv5Mdj4ScuTgNKH5uj",
							}, {
								Timestamp:      timestamp,
								Level:          381735,
								DestinationPtr: newInt64Ptr(40077),
								Action:         types.BigMapActionCopy,
								Address:        "KT1BM1SyQnTzNU1J8TZv5Mdj4ScuTgNKH5uj",
							}, {
								Timestamp:      timestamp,
								Level:          381735,
								DestinationPtr: newInt64Ptr(40076),
								Action:         types.BigMapActionCopy,
								Address:        "KT1BM1SyQnTzNU1J8TZv5Mdj4ScuTgNKH5uj",
							}, {
								Timestamp:      timestamp,
								Level:          381735,
								DestinationPtr: newInt64Ptr(40075),
								Action:         types.BigMapActionCopy,
								Address:        "KT1BM1SyQnTzNU1J8TZv5Mdj4ScuTgNKH5uj",
							}, {
								Timestamp:      timestamp,
								Level:          381735,
								DestinationPtr: newInt64Ptr(40074),
								Action:         types.BigMapActionCopy,
								Address:        "KT1BM1SyQnTzNU1J8TZv5Mdj4ScuTgNKH5uj",
							}, {
								Timestamp:      timestamp,
								Level:          381735,
								DestinationPtr: newInt64Ptr(40073),
								Action:         types.BigMapActionCopy,
								Address:        "KT1BM1SyQnTzNU1J8TZv5Mdj4ScuTgNKH5uj",
							}, {
								Timestamp:      timestamp,
								Level:          381735,
								DestinationPtr: newInt64Ptr(40072),
								Action:         types.BigMapActionCopy,
								Address:        "KT1BM1SyQnTzNU1J8TZv5Mdj4ScuTgNKH5uj",
							}, {
								Timestamp:      timestamp,
								Level:          381735,
								DestinationPtr: newInt64Ptr(40071),
								Action:         types.BigMapActionCopy,
								Address:        "KT1BM1SyQnTzNU1J8TZv5Mdj4ScuTgNKH5uj",
							}, {
								Timestamp:      timestamp,
								Level:          381735,
								DestinationPtr: newInt64Ptr(40070),
								Action:         types.BigMapActionCopy,
								Address:        "KT1BM1SyQnTzNU1J8TZv5Mdj4ScuTgNKH5uj",
							}, {
								Timestamp:      timestamp,
								Level:          381735,
								DestinationPtr: newInt64Ptr(40069),
								Action:         types.BigMapActionCopy,
								Address:        "KT1BM1SyQnTzNU1J8TZv5Mdj4ScuTgNKH5uj",
							},
						},
					},
				},

				BigMapState: []*bigmapdiff.BigMapState{
					{
						LastUpdateTime:  timestamp,
						LastUpdateLevel: 381735,
						Ptr:             40064,
						KeyHash:         "exprtjtvPgEDWmVoB1v5jdmX6HGjTZZFYoWxrtr2s4pT8GHr3XJwPi",
						Contract:        "KT1Jk8LRDoj6LkopYZwRq5ZEWBhYv8nVc6e6",
						Key:             []byte(`{"bytes":"05070702000000460704000005050a00000016012a0b69e71ece4da314f9904bfb7f8a8d3d373c530007040001050807070a000000160107a2fc7b796ea23ad82e768221a78c86c77c64050000000a00000016000042a7bb84edce2af4cc8ab0bc83ded699efc9300a"}`),
						Value:           []byte(`{"bytes":"011e4e248aea7b65b1941614f481e5801b0ebaacdc00"}`),
					},
					{
						LastUpdateTime:  timestamp,
						LastUpdateLevel: 381735,
						Ptr:             34843,
						KeyHash:         "exprvHCaW3fmGXHKV22BNVUkR7TLNCca84z3ANPRSaY9ubZUj7QrF6",
						Contract:        "KT19H9YbHqsxFTayap7aTEfbcnyPeALKYgt9",
						Key:             []byte(`{"bytes":"000098b9732c83017e938ba48cb91cf53e5f919dc844"}`),
						Value:           []byte(`{"prim":"Pair","args":[[{"prim":"Elt","args":[{"int":"0"},{"int":"0"}]}],{"prim":"Pair","args":[{"int":"1649845410"},[]]}]}`),
					},
					{
						LastUpdateTime:  timestamp,
						LastUpdateLevel: 381735,
						Ptr:             34843,
						KeyHash:         "exprtj5G3z2kcdmxvy3y9nFqs57enTtDSyrRttFHm7tL7PdLqR2ek5",
						Contract:        "KT19H9YbHqsxFTayap7aTEfbcnyPeALKYgt9",
						Key:             []byte(`{"bytes":"000042a7bb84edce2af4cc8ab0bc83ded699efc9300a"}`),
						Value:           []byte(`{"prim":"Pair","args":[[{"prim":"Elt","args":[{"int":"0"},{"int":"9999999989999000000"}]}],{"prim":"Pair","args":[{"int":"1649845320"},[{"bytes":"01308a4d463c798401eb231fb386ac223e8d44987400"},{"bytes":"016f7656dd6c6df9f8294efae0235ccb7f27025c4900"},{"bytes":"01c14ac0a868ad16d8115a2153bd700cdf0d7898f100"}]]}]}`),
					},
					{
						LastUpdateTime:  timestamp,
						LastUpdateLevel: 381735,
						Ptr:             34843,
						KeyHash:         "exprtj5G3z2kcdmxvy3y9nFqs57enTtDSyrRttFHm7tL7PdLqR2ek5",
						Contract:        "KT19H9YbHqsxFTayap7aTEfbcnyPeALKYgt9",
						Key:             []byte(`{"bytes":"000042a7bb84edce2af4cc8ab0bc83ded699efc9300a"}`),
						Value:           []byte(`{"prim":"Pair","args":[[{"prim":"Elt","args":[{"int":"0"},{"int":"9999999989999000000"}]}],{"prim":"Pair","args":[{"int":"1649845320"},[{"bytes":"01308a4d463c798401eb231fb386ac223e8d44987400"},{"bytes":"016f7656dd6c6df9f8294efae0235ccb7f27025c4900"},{"bytes":"01c14ac0a868ad16d8115a2153bd700cdf0d7898f100"}]]}]}`),
					},
					{
						LastUpdateTime:  timestamp,
						LastUpdateLevel: 381735,
						Ptr:             34843,
						KeyHash:         "exprtXp227aTXC9hWQC3H6y5M9rP38UD1qSHDeaEL4hXqxb5ceurBr",
						Contract:        "KT19H9YbHqsxFTayap7aTEfbcnyPeALKYgt9",
						Key:             []byte(`{"bytes":"016f7656dd6c6df9f8294efae0235ccb7f27025c4900"}`),
						Value:           []byte(`{"prim":"Pair","args":[[{"prim":"Elt","args":[{"int":"0"},{"int":"0"}]}],{"prim":"Pair","args":[{"int":"1649854355"},[]]}]}`),
					},
					{
						LastUpdateTime:  timestamp,
						LastUpdateLevel: 381735,
						Ptr:             40078,
						KeyHash:         "exprtZBwZUeYYYfUs9B9Rg2ywHezVHnCCnmF9WsDQVrs582dSK63dC",
						Contract:        "KT1BM1SyQnTzNU1J8TZv5Mdj4ScuTgNKH5uj",
						Key:             []byte(`{"int":"0"}`),
						Value:           []byte(`{"prim":"Pair","args":[{"int":"0"},[{"prim":"Elt","args":[{"string":"decimals"},{"bytes":"3138"}]},{"prim":"Elt","args":[{"string":"description"},{"bytes":"4c697175696469747920506f6f6c20746f6b656e206f662051756970755377617020537461626c6520444558"}]},{"prim":"Elt","args":[{"string":"name"},{"bytes":"537461626c652044455820517569707553776170204c5020746f6b656e"}]},{"prim":"Elt","args":[{"string":"shouldPreferSymbol"},{"bytes":"74727565"}]},{"prim":"Elt","args":[{"string":"symbol"},{"bytes":"736451504c50"}]},{"prim":"Elt","args":[{"string":"thumbnailUri"},{"bytes":"697066733a2f2f516d5558464a787747456a6b6e5035444b4e774d777357587672593847326a554d62786a64424533656b50384450"}]}]]}`),
					},
					{
						LastUpdateTime:  timestamp,
						LastUpdateLevel: 381735,
						Ptr:             40077,
						Contract:        "KT1BM1SyQnTzNU1J8TZv5Mdj4ScuTgNKH5uj",
						Key:             []byte(`{"string":"dex"}`),
						Value:           []byte(`{"bytes":"7b226e616d65223a2251756970755377617020537461626c652044455820706f6f6c222c2276657273696f6e223a2276312e302e30222c226465736372697074696f6e223a22506f6f6c20666f72207377617070696e6720746f6b656e732077697468206c6f7720736c697070616765222c22617574686f7273223a5b224d6164666973682e536f6c7574696f6e73203c68747470733a2f2f7777772e6d6164666973682e736f6c7574696f6e733e225d2c22736f75726365223a7b22746f6f6c73223a5b224c69676f222c22466c657874657361225d2c226c6f636174696f6e223a2268747470733a2f2f6769746875622e636f6d2f6d6164666973682d736f6c7574696f6e732f7175697075737761702d737461626c652d636f72652f626c6f622f6d61696e2f636f6e7472616374732f6d61696e2f6465782e6c69676f227d2c22686f6d6570616765223a2268747470733a2f2f7175697075737761702e636f6d222c22696e7465726661636573223a5b22545a49502d313220676974203137323866636665222c22545a49502d3136225d2c226572726f7273223a5b5d2c227669657773223a5b5d7d"}`),
						KeyHash:         "exprupXFsHdKsx5MFzpvfXijjEPBt18ipyXfJcPbpri7K9zYf1Fb7o",
					},
					{
						LastUpdateTime:  timestamp,
						LastUpdateLevel: 381735,
						Ptr:             40077,
						KeyHash:         "expru5X1yxJG6ezR2uHMotwMLNmSzQyh5t1vUnhjx4cS6Pv9qE1Sdo",
						Contract:        "KT1BM1SyQnTzNU1J8TZv5Mdj4ScuTgNKH5uj",
						Key:             []byte(`{"string":""}`),
						Value:           []byte(`{"bytes":"74657a6f732d73746f726167653a646578"}`),
					},
					{
						Ptr:             40071,
						Key:             []byte(`{"int":"0"}`),
						Value:           []byte(`[{"int":"10000000"},{"int":"1649854355"},{"int":"10000000"},{"int":"1649854355"},[{"prim":"Elt","args":[{"int":"0"},{"prim":"Pair","args":[{"int":"1000000000000000000"},{"prim":"Pair","args":[{"int":"1000000000000000000000000"},{"int":"0"}]}]}]},{"prim":"Elt","args":[{"int":"1"},{"prim":"Pair","args":[{"int":"1000000000000000000"},{"prim":"Pair","args":[{"int":"100000000000000000000000000"},{"int":"0"}]}]}]}],{"prim":"Pair","args":[{"int":"0"},{"prim":"Pair","args":[{"int":"0"},{"int":"0"}]}]},{"prim":"Pair","args":[[],{"int":"0"}]},{"int":"0"}]`),
						KeyHash:         "exprtZBwZUeYYYfUs9B9Rg2ywHezVHnCCnmF9WsDQVrs582dSK63dC",
						LastUpdateLevel: 381735,
						LastUpdateTime:  timestamp,
						Contract:        "KT1BM1SyQnTzNU1J8TZv5Mdj4ScuTgNKH5uj",
					},
					{
						Ptr:             40070,
						Key:             []byte(`{"bytes":"0502000000460704000005050a00000016012a0b69e71ece4da314f9904bfb7f8a8d3d373c530007040001050807070a000000160107a2fc7b796ea23ad82e768221a78c86c77c6405000000"}`),
						Value:           []byte(`{"int":"0"}`),
						KeyHash:         "expruLxJSi2bGyABA2WpjGBm5dd5zRUy84zWy2dRCHVGJSaNxrLL8t",
						LastUpdateLevel: 381735,
						LastUpdateTime:  timestamp,
						Contract:        "KT1BM1SyQnTzNU1J8TZv5Mdj4ScuTgNKH5uj",
					},
					{
						Ptr:             40069,
						Key:             []byte(`{"int":"0"}`),
						Value:           []byte(`[{"prim":"Elt","args":[{"int":"0"},{"prim":"Left","args":[{"bytes":"012a0b69e71ece4da314f9904bfb7f8a8d3d373c5300"}]}]},{"prim":"Elt","args":[{"int":"1"},{"prim":"Right","args":[{"prim":"Pair","args":[{"bytes":"0107a2fc7b796ea23ad82e768221a78c86c77c640500"},{"int":"0"}]}]}]}]`),
						KeyHash:         "exprtZBwZUeYYYfUs9B9Rg2ywHezVHnCCnmF9WsDQVrs582dSK63dC",
						LastUpdateLevel: 381735,
						LastUpdateTime:  timestamp,
						Contract:        "KT1BM1SyQnTzNU1J8TZv5Mdj4ScuTgNKH5uj",
					},
				},

				Contracts: []*modelContract.Contract{
					{
						Timestamp: timestamp,
						Level:     381735,
						Account: account.Account{
							Address: "KT1BM1SyQnTzNU1J8TZv5Mdj4ScuTgNKH5uj",
							Type:    types.AccountTypeContract,
						},
						Manager: account.Account{
							Address: "KT1Jk8LRDoj6LkopYZwRq5ZEWBhYv8nVc6e6",
							Type:    types.AccountTypeContract,
						},
						Tags:       16640,
						LastAction: timestamp,
						Babylon: modelContract.Script{
							Hash: "c52584fb0678ae8b5f7e8021899b7c96060bbbe15c26cc52a3fa122f25262105",
							FailStrings: []string{
								"failed assertion",
							},
							Annotations: []string{
								"%candidate",
								"%fa12",
								"%total_supply",
								"%min_amounts_out",
								"%invest",
								"%from_",
								"%token_lambdas",
								"%factory_action",
								"%future_A",
								"%dev_rewards",
								"%started",
								"%shares",
								"%receiver",
								"%min_amount_out",
								"%balance_of",
								"%use_admin",
								"%initial_A_f",
								"%stakers_balance",
								"%divest",
								"%stake",
								"%token_info",
								"%set_default_referral",
								"%staker_accumulator",
								"%balance",
								"%request",
								"%to_",
								"%use_token",
								"%set_admin",
								"%stakers_f",
								"%tokens",
								"%divest_one_coin",
								"%admin",
								"%managers",
								"%precision_multiplier_f",
								"%earnings",
								"%max_shares",
								"%owner",
								"%token_metadata",
								"%token_id",
								"%future_time",
								"%ledger",
								"%referral_rewards",
								"%reward_f",
								"%operator",
								"%add_rem_managers",
								"%ramp_A",
								"%total_staked",
								"%referral",
								"%set_fees",
								"%ref_f",
								"%future_A_time",
								"%quipu_token",
								"%deadline",
								"%amounts_out",
								"%token_index",
								"%idx_to",
								"%add_operator",
								"%admin_lambdas",
								"%token_address",
								"%default_referral",
								"%future_A_f",
								"%former_f",
								"%claim_referral",
								"%in_amounts",
								"%remove",
								"%update_metadata",
								"%freeze",
								"%user_action",
								"%claim_developer",
								"%stop_ramp_A",
								"%pools_count",
								"%reserves",
								"%factory_address",
								"%fee",
								"%lp_f",
								"%tokens_info",
								"%swap",
								"%idx_from",
								"%accumulator_f",
								"%requests",
								"%transfer",
								"%copy_dex_function",
								"%use_dex",
								"%callback",
								"%fa2",
								"%initial_A_time",
								"%allowances",
								"%storage",
								"%add",
								"%token",
								"%amount",
								"%pool_id",
								"%pool_to_id",
								"%pools",
								"%rate_f",
								"%divest_imbalanced",
								"%txs",
								"%update_operators",
								"%remove_operator",
								"%metadata",
								"%dex_lambdas",
							},
							Entrypoints: []string{
								"copy_dex_function",
								"freeze",
								"add_rem_managers",
								"claim_developer",
								"ramp_A",
								"set_admin",
								"set_default_referral",
								"set_fees",
								"stop_ramp_A",
								"claim_referral",
								"divest",
								"divest_imbalanced",
								"divest_one_coin",
								"invest",
								"add",
								"remove",
								"swap",
								"balance_of",
								"total_supply",
								"transfer",
								"update_metadata",
								"update_operators",
							},
							Code: []byte(`[[{"prim":"PUSH","args":[{"prim":"string"},{"string":"function-not-set"}]},{"prim":"PUSH","args":[{"prim":"string"},{"string":"cant-unpack-use-lambda"}]},{"prim":"PUSH","args":[{"prim":"string"},{"string":"not-contract-admin"}]},{"prim":"PUSH","args":[{"prim":"string"},{"string":"not-started"}]},{"prim":"DIG","args":[{"int":"4"}]},{"prim":"UNPAIR"},{"prim":"DUP"},{"prim":"IF_LEFT","args":[[{"prim":"IF_LEFT","args":[[{"prim":"DIG","args":[{"int":"2"}]},{"prim":"DUP"},{"prim":"DUG","args":[{"int":"2"}]},{"prim":"CAR"},{"prim":"GET","args":[{"int":"27"}]},{"prim":"SENDER"},{"prim":"COMPARE"},{"prim":"EQ"},{"prim":"IF","args":[[],[{"prim":"PUSH","args":[{"prim":"string"},{"string":"failed assertion"}]},{"prim":"FAILWITH"}]]},{"prim":"IF_LEFT","args":[[{"prim":"UPDATE","args":[{"int":"9"}]}],[{"prim":"DROP"},{"prim":"DUP"},{"prim":"DUP"},{"prim":"DUG","args":[{"int":"2"}]},{"prim":"CAR"},{"prim":"DIG","args":[{"int":"2"}]},{"prim":"CAR"},{"prim":"GET","args":[{"int":"28"}]},{"prim":"NOT"},{"prim":"UPDATE","args":[{"int":"28"}]},{"prim":"UPDATE","args":[{"int":"1"}]}]]}],[{"prim":"DROP"},{"prim":"SWAP"}]]}],[{"prim":"DROP"},{"prim":"SWAP"}]]},{"prim":"SWAP"},{"prim":"IF_LEFT","args":[[{"prim":"IF_LEFT","args":[[{"prim":"DIG","args":[{"int":"2"}]},{"prim":"DIG","args":[{"int":"3"}]},{"prim":"DIG","args":[{"int":"4"}]},{"prim":"DIG","args":[{"int":"5"}]},{"prim":"DROP","args":[{"int":"5"}]},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}],[{"prim":"DUG","args":[{"int":"2"}]},{"prim":"DUP"},{"prim":"DUG","args":[{"int":"3"}]},{"prim":"CAR"},{"prim":"GET","args":[{"int":"28"}]},{"prim":"IF","args":[[{"prim":"DROP"}],[{"prim":"FAILWITH"}]]},{"prim":"DUP"},{"prim":"IF_LEFT","args":[[{"prim":"IF_LEFT","args":[[{"prim":"IF_LEFT","args":[[{"prim":"DROP"},{"prim":"DUG","args":[{"int":"2"}]},{"prim":"DUP"},{"prim":"DUG","args":[{"int":"3"}]},{"prim":"CAR"},{"prim":"CAR"},{"prim":"SENDER"},{"prim":"COMPARE"},{"prim":"EQ"},{"prim":"IF","args":[[{"prim":"DROP"},{"prim":"UNIT"}],[{"prim":"FAILWITH"}]]}],[{"prim":"DIG","args":[{"int":"3"}]},{"prim":"DROP","args":[{"int":"2"}]},{"prim":"UNIT"}]]}],[{"prim":"IF_LEFT","args":[[{"prim":"DROP"},{"prim":"DUG","args":[{"int":"2"}]},{"prim":"DUP"},{"prim":"DUG","args":[{"int":"3"}]},{"prim":"CAR"},{"prim":"CAR"},{"prim":"SENDER"},{"prim":"COMPARE"},{"prim":"EQ"},{"prim":"IF","args":[[{"prim":"DROP"},{"prim":"UNIT"}],[{"prim":"FAILWITH"}]]}],[{"prim":"DROP"},{"prim":"DUG","args":[{"int":"2"}]},{"prim":"DUP"},{"prim":"DUG","args":[{"int":"3"}]},{"prim":"CAR"},{"prim":"CAR"},{"prim":"SENDER"},{"prim":"COMPARE"},{"prim":"EQ"},{"prim":"IF","args":[[{"prim":"DROP"},{"prim":"UNIT"}],[{"prim":"FAILWITH"}]]}]]}]]}],[{"prim":"IF_LEFT","args":[[{"prim":"IF_LEFT","args":[[{"prim":"DROP"},{"prim":"DUG","args":[{"int":"2"}]},{"prim":"DUP"},{"prim":"DUG","args":[{"int":"3"}]},{"prim":"CAR"},{"prim":"CAR"},{"prim":"SENDER"},{"prim":"COMPARE"},{"prim":"EQ"},{"prim":"IF","args":[[{"prim":"DROP"},{"prim":"UNIT"}],[{"prim":"FAILWITH"}]]}],[{"prim":"DROP"},{"prim":"DUG","args":[{"int":"2"}]},{"prim":"DUP"},{"prim":"DUG","args":[{"int":"3"}]},{"prim":"CAR"},{"prim":"CAR"},{"prim":"SENDER"},{"prim":"COMPARE"},{"prim":"EQ"},{"prim":"IF","args":[[{"prim":"DROP"},{"prim":"UNIT"}],[{"prim":"FAILWITH"}]]}]]}],[{"prim":"DROP"},{"prim":"DUG","args":[{"int":"2"}]},{"prim":"DUP"},{"prim":"DUG","args":[{"int":"3"}]},{"prim":"CAR"},{"prim":"CAR"},{"prim":"SENDER"},{"prim":"COMPARE"},{"prim":"EQ"},{"prim":"IF","args":[[{"prim":"DROP"},{"prim":"UNIT"}],[{"prim":"FAILWITH"}]]}]]}]]},{"prim":"DROP"},{"prim":"DUP"},{"prim":"IF_LEFT","args":[[{"prim":"IF_LEFT","args":[[{"prim":"IF_LEFT","args":[[{"prim":"DROP"},{"prim":"PUSH","args":[{"prim":"nat"},{"int":"0"}]}],[{"prim":"DROP"},{"prim":"PUSH","args":[{"prim":"nat"},{"int":"2"}]}]]}],[{"prim":"IF_LEFT","args":[[{"prim":"DROP"},{"prim":"PUSH","args":[{"prim":"nat"},{"int":"3"}]}],[{"prim":"DROP"},{"prim":"PUSH","args":[{"prim":"nat"},{"int":"1"}]}]]}]]}],[{"prim":"IF_LEFT","args":[[{"prim":"IF_LEFT","args":[[{"prim":"DROP"},{"prim":"PUSH","args":[{"prim":"nat"},{"int":"6"}]}],[{"prim":"DROP"},{"prim":"PUSH","args":[{"prim":"nat"},{"int":"5"}]}]]}],[{"prim":"DROP"},{"prim":"PUSH","args":[{"prim":"nat"},{"int":"4"}]}]]}]]},{"prim":"DIG","args":[{"int":"4"}]},{"prim":"DUP","args":[{"int":"4"}]},{"prim":"GET","args":[{"int":"7"}]},{"prim":"DUP","args":[{"int":"3"}]},{"prim":"GET"},{"prim":"IF_NONE","args":[[{"prim":"FAILWITH"}],[{"prim":"SWAP"},{"prim":"DROP"}]]},{"prim":"DIG","args":[{"int":"4"}]},{"prim":"SWAP"},{"prim":"UNPACK","args":[{"prim":"lambda","args":[{"prim":"pair","args":[{"prim":"or","args":[{"prim":"or","args":[{"prim":"or","args":[{"prim":"pair","args":[{"prim":"bool","annots":["%add"]},{"prim":"address","annots":["%candidate"]}],"annots":["%add_rem_managers"]},{"prim":"pair","args":[{"prim":"or","args":[{"prim":"address","annots":["%fa12"]},{"prim":"pair","args":[{"prim":"address","annots":["%token_address"]},{"prim":"nat","annots":["%token_id"]}],"annots":["%fa2"]}],"annots":["%token"]},{"prim":"nat","annots":["%amount"]}],"annots":["%claim_developer"]}]},{"prim":"or","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%pool_id"]},{"prim":"pair","args":[{"prim":"nat","annots":["%future_A"]},{"prim":"timestamp","annots":["%future_time"]}]}],"annots":["%ramp_A"]},{"prim":"address","annots":["%set_admin"]}]}]},{"prim":"or","args":[{"prim":"or","args":[{"prim":"address","annots":["%set_default_referral"]},{"prim":"pair","args":[{"prim":"nat","annots":["%pool_id"]},{"prim":"pair","args":[{"prim":"nat","annots":["%lp_f"]},{"prim":"pair","args":[{"prim":"nat","annots":["%stakers_f"]},{"prim":"nat","annots":["%ref_f"]}]}],"annots":["%fee"]}],"annots":["%set_fees"]}]},{"prim":"nat","annots":["%stop_ramp_A"]}]}]},{"prim":"pair","args":[{"prim":"address","annots":["%admin"]},{"prim":"pair","args":[{"prim":"address","annots":["%default_referral"]},{"prim":"pair","args":[{"prim":"set","args":[{"prim":"address"}],"annots":["%managers"]},{"prim":"pair","args":[{"prim":"nat","annots":["%pools_count"]},{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"nat"},{"prim":"map","args":[{"prim":"nat"},{"prim":"or","args":[{"prim":"address","annots":["%fa12"]},{"prim":"pair","args":[{"prim":"address","annots":["%token_address"]},{"prim":"nat","annots":["%token_id"]}],"annots":["%fa2"]}]}]}],"annots":["%tokens"]},{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"nat"}],"annots":["%pool_to_id"]},{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"nat"},{"prim":"pair","args":[{"prim":"nat","annots":["%initial_A_f"]},{"prim":"pair","args":[{"prim":"timestamp","annots":["%initial_A_time"]},{"prim":"pair","args":[{"prim":"nat","annots":["%future_A_f"]},{"prim":"pair","args":[{"prim":"timestamp","annots":["%future_A_time"]},{"prim":"pair","args":[{"prim":"map","args":[{"prim":"nat"},{"prim":"pair","args":[{"prim":"nat","annots":["%rate_f"]},{"prim":"pair","args":[{"prim":"nat","annots":["%precision_multiplier_f"]},{"prim":"nat","annots":["%reserves"]}]}]}],"annots":["%tokens_info"]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%lp_f"]},{"prim":"pair","args":[{"prim":"nat","annots":["%stakers_f"]},{"prim":"nat","annots":["%ref_f"]}]}],"annots":["%fee"]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"map","args":[{"prim":"nat"},{"prim":"nat"}],"annots":["%accumulator_f"]},{"prim":"nat","annots":["%total_staked"]}],"annots":["%staker_accumulator"]},{"prim":"nat","annots":["%total_supply"]}]}]}]}]}]}]}]}],"annots":["%pools"]},{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"pair","args":[{"prim":"address"},{"prim":"nat"}]},{"prim":"nat"}],"annots":["%ledger"]},{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"pair","args":[{"prim":"address"},{"prim":"nat"}]},{"prim":"set","args":[{"prim":"address"}]}],"annots":["%allowances"]},{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"or","args":[{"prim":"address","annots":["%fa12"]},{"prim":"pair","args":[{"prim":"address","annots":["%token_address"]},{"prim":"nat","annots":["%token_id"]}],"annots":["%fa2"]}]},{"prim":"nat"}],"annots":["%dev_rewards"]},{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"pair","args":[{"prim":"address"},{"prim":"or","args":[{"prim":"address","annots":["%fa12"]},{"prim":"pair","args":[{"prim":"address","annots":["%token_address"]},{"prim":"nat","annots":["%token_id"]}],"annots":["%fa2"]}]}]},{"prim":"nat"}],"annots":["%referral_rewards"]},{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"pair","args":[{"prim":"address"},{"prim":"nat"}]},{"prim":"pair","args":[{"prim":"nat","annots":["%balance"]},{"prim":"map","args":[{"prim":"nat"},{"prim":"pair","args":[{"prim":"nat","annots":["%reward_f"]},{"prim":"nat","annots":["%former_f"]}]}],"annots":["%earnings"]}]}],"annots":["%stakers_balance"]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address","annots":["%token_address"]},{"prim":"nat","annots":["%token_id"]}],"annots":["%quipu_token"]},{"prim":"pair","args":[{"prim":"address","annots":["%factory_address"]},{"prim":"bool","annots":["%started"]}]}]}]}]}]}]}]}]}]}]}]}]}]}]}]},{"prim":"pair","args":[{"prim":"list","args":[{"prim":"operation"}]},{"prim":"pair","args":[{"prim":"address","annots":["%admin"]},{"prim":"pair","args":[{"prim":"address","annots":["%default_referral"]},{"prim":"pair","args":[{"prim":"set","args":[{"prim":"address"}],"annots":["%managers"]},{"prim":"pair","args":[{"prim":"nat","annots":["%pools_count"]},{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"nat"},{"prim":"map","args":[{"prim":"nat"},{"prim":"or","args":[{"prim":"address","annots":["%fa12"]},{"prim":"pair","args":[{"prim":"address","annots":["%token_address"]},{"prim":"nat","annots":["%token_id"]}],"annots":["%fa2"]}]}]}],"annots":["%tokens"]},{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"nat"}],"annots":["%pool_to_id"]},{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"nat"},{"prim":"pair","args":[{"prim":"nat","annots":["%initial_A_f"]},{"prim":"pair","args":[{"prim":"timestamp","annots":["%initial_A_time"]},{"prim":"pair","args":[{"prim":"nat","annots":["%future_A_f"]},{"prim":"pair","args":[{"prim":"timestamp","annots":["%future_A_time"]},{"prim":"pair","args":[{"prim":"map","args":[{"prim":"nat"},{"prim":"pair","args":[{"prim":"nat","annots":["%rate_f"]},{"prim":"pair","args":[{"prim":"nat","annots":["%precision_multiplier_f"]},{"prim":"nat","annots":["%reserves"]}]}]}],"annots":["%tokens_info"]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%lp_f"]},{"prim":"pair","args":[{"prim":"nat","annots":["%stakers_f"]},{"prim":"nat","annots":["%ref_f"]}]}],"annots":["%fee"]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"map","args":[{"prim":"nat"},{"prim":"nat"}],"annots":["%accumulator_f"]},{"prim":"nat","annots":["%total_staked"]}],"annots":["%staker_accumulator"]},{"prim":"nat","annots":["%total_supply"]}]}]}]}]}]}]}]}],"annots":["%pools"]},{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"pair","args":[{"prim":"address"},{"prim":"nat"}]},{"prim":"nat"}],"annots":["%ledger"]},{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"pair","args":[{"prim":"address"},{"prim":"nat"}]},{"prim":"set","args":[{"prim":"address"}]}],"annots":["%allowances"]},{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"or","args":[{"prim":"address","annots":["%fa12"]},{"prim":"pair","args":[{"prim":"address","annots":["%token_address"]},{"prim":"nat","annots":["%token_id"]}],"annots":["%fa2"]}]},{"prim":"nat"}],"annots":["%dev_rewards"]},{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"pair","args":[{"prim":"address"},{"prim":"or","args":[{"prim":"address","annots":["%fa12"]},{"prim":"pair","args":[{"prim":"address","annots":["%token_address"]},{"prim":"nat","annots":["%token_id"]}],"annots":["%fa2"]}]}]},{"prim":"nat"}],"annots":["%referral_rewards"]},{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"pair","args":[{"prim":"address"},{"prim":"nat"}]},{"prim":"pair","args":[{"prim":"nat","annots":["%balance"]},{"prim":"map","args":[{"prim":"nat"},{"prim":"pair","args":[{"prim":"nat","annots":["%reward_f"]},{"prim":"nat","annots":["%former_f"]}]}],"annots":["%earnings"]}]}],"annots":["%stakers_balance"]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address","annots":["%token_address"]},{"prim":"nat","annots":["%token_id"]}],"annots":["%quipu_token"]},{"prim":"pair","args":[{"prim":"address","annots":["%factory_address"]},{"prim":"bool","annots":["%started"]}]}]}]}]}]}]}]}]}]}]}]}]}]}]}]}]}]},{"prim":"IF_NONE","args":[[{"prim":"FAILWITH"}],[{"prim":"SWAP"},{"prim":"DROP"}]]},{"prim":"DUP","args":[{"int":"4"}]},{"prim":"CAR"},{"prim":"DIG","args":[{"int":"3"}]},{"prim":"PAIR"},{"prim":"EXEC"},{"prim":"UNPAIR"},{"prim":"PUSH","args":[{"prim":"nat"},{"int":"7"}]},{"prim":"DIG","args":[{"int":"3"}]},{"prim":"COMPARE"},{"prim":"EQ"},{"prim":"IF","args":[[{"prim":"DIG","args":[{"int":"2"}]},{"prim":"DUP"},{"prim":"GET","args":[{"int":"5"}]},{"prim":"EMPTY_MAP","args":[{"prim":"string"},{"prim":"bytes"}]},{"prim":"PUSH","args":[{"prim":"bytes"},{"bytes":"697066733a2f2f516d5558464a787747456a6b6e5035444b4e774d777357587672593847326a554d62786a64424533656b50384450"}]},{"prim":"SOME"},{"prim":"PUSH","args":[{"prim":"string"},{"string":"thumbnailUri"}]},{"prim":"UPDATE"},{"prim":"PUSH","args":[{"prim":"bytes"},{"bytes":"736451504c50"}]},{"prim":"SOME"},{"prim":"PUSH","args":[{"prim":"string"},{"string":"symbol"}]},{"prim":"UPDATE"},{"prim":"PUSH","args":[{"prim":"bytes"},{"bytes":"74727565"}]},{"prim":"SOME"},{"prim":"PUSH","args":[{"prim":"string"},{"string":"shouldPreferSymbol"}]},{"prim":"UPDATE"},{"prim":"PUSH","args":[{"prim":"bytes"},{"bytes":"537461626c652044455820517569707553776170204c5020746f6b656e"}]},{"prim":"SOME"},{"prim":"PUSH","args":[{"prim":"string"},{"string":"name"}]},{"prim":"UPDATE"},{"prim":"PUSH","args":[{"prim":"bytes"},{"bytes":"4c697175696469747920506f6f6c20746f6b656e206f662051756970755377617020537461626c6520444558"}]},{"prim":"SOME"},{"prim":"PUSH","args":[{"prim":"string"},{"string":"description"}]},{"prim":"UPDATE"},{"prim":"PUSH","args":[{"prim":"bytes"},{"bytes":"3138"}]},{"prim":"SOME"},{"prim":"PUSH","args":[{"prim":"string"},{"string":"decimals"}]},{"prim":"UPDATE"},{"prim":"PUSH","args":[{"prim":"nat"},{"int":"1"}]},{"prim":"DUP","args":[{"int":"6"}]},{"prim":"GET","args":[{"int":"7"}]},{"prim":"SUB"},{"prim":"ABS"},{"prim":"PAIR"},{"prim":"PUSH","args":[{"prim":"nat"},{"int":"1"}]},{"prim":"DUP","args":[{"int":"6"}]},{"prim":"GET","args":[{"int":"7"}]},{"prim":"SUB"},{"prim":"ABS"},{"prim":"SWAP"},{"prim":"SOME"},{"prim":"SWAP"},{"prim":"UPDATE"},{"prim":"UPDATE","args":[{"int":"5"}]}],[{"prim":"DIG","args":[{"int":"2"}]}]]},{"prim":"DIG","args":[{"int":"2"}]},{"prim":"UPDATE","args":[{"int":"1"}]},{"prim":"SWAP"},{"prim":"PAIR"}]]}],[{"prim":"DIG","args":[{"int":"3"}]},{"prim":"DROP"},{"prim":"IF_LEFT","args":[[{"prim":"DUG","args":[{"int":"2"}]},{"prim":"DUP"},{"prim":"DUG","args":[{"int":"3"}]},{"prim":"CAR"},{"prim":"GET","args":[{"int":"28"}]},{"prim":"IF","args":[[{"prim":"DROP"}],[{"prim":"FAILWITH"}]]},{"prim":"DIG","args":[{"int":"3"}]},{"prim":"DUP","args":[{"int":"3"}]},{"prim":"GET","args":[{"int":"9"}]},{"prim":"DUP","args":[{"int":"3"}]},{"prim":"IF_LEFT","args":[[{"prim":"IF_LEFT","args":[[{"prim":"IF_LEFT","args":[[{"prim":"DROP"},{"prim":"PUSH","args":[{"prim":"nat"},{"int":"5"}]}],[{"prim":"DROP"},{"prim":"PUSH","args":[{"prim":"nat"},{"int":"2"}]}]]}],[{"prim":"IF_LEFT","args":[[{"prim":"DROP"},{"prim":"PUSH","args":[{"prim":"nat"},{"int":"3"}]}],[{"prim":"DROP"},{"prim":"PUSH","args":[{"prim":"nat"},{"int":"4"}]}]]}]]}],[{"prim":"IF_LEFT","args":[[{"prim":"IF_LEFT","args":[[{"prim":"DROP"},{"prim":"PUSH","args":[{"prim":"nat"},{"int":"1"}]}],[{"prim":"DROP"},{"prim":"PUSH","args":[{"prim":"nat"},{"int":"6"}]}]]}],[{"prim":"DROP"},{"prim":"PUSH","args":[{"prim":"nat"},{"int":"0"}]}]]}]]},{"prim":"GET"},{"prim":"IF_NONE","args":[[{"prim":"FAILWITH"}],[{"prim":"SWAP"},{"prim":"DROP"}]]},{"prim":"DIG","args":[{"int":"3"}]},{"prim":"SWAP"},{"prim":"UNPACK","args":[{"prim":"lambda","args":[{"prim":"pair","args":[{"prim":"or","args":[{"prim":"or","args":[{"prim":"or","args":[{"prim":"pair","args":[{"prim":"or","args":[{"prim":"address","annots":["%fa12"]},{"prim":"pair","args":[{"prim":"address","annots":["%token_address"]},{"prim":"nat","annots":["%token_id"]}],"annots":["%fa2"]}],"annots":["%token"]},{"prim":"nat","annots":["%amount"]}],"annots":["%claim_referral"]},{"prim":"pair","args":[{"prim":"nat","annots":["%pool_id"]},{"prim":"pair","args":[{"prim":"map","args":[{"prim":"nat"},{"prim":"nat"}],"annots":["%min_amounts_out"]},{"prim":"pair","args":[{"prim":"nat","annots":["%shares"]},{"prim":"pair","args":[{"prim":"timestamp","annots":["%deadline"]},{"prim":"option","args":[{"prim":"address"}],"annots":["%receiver"]}]}]}]}],"annots":["%divest"]}]},{"prim":"or","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%pool_id"]},{"prim":"pair","args":[{"prim":"map","args":[{"prim":"nat"},{"prim":"nat"}],"annots":["%amounts_out"]},{"prim":"pair","args":[{"prim":"nat","annots":["%max_shares"]},{"prim":"pair","args":[{"prim":"timestamp","annots":["%deadline"]},{"prim":"pair","args":[{"prim":"option","args":[{"prim":"address"}],"annots":["%receiver"]},{"prim":"option","args":[{"prim":"address"}],"annots":["%referral"]}]}]}]}]}],"annots":["%divest_imbalanced"]},{"prim":"pair","args":[{"prim":"nat","annots":["%pool_id"]},{"prim":"pair","args":[{"prim":"nat","annots":["%shares"]},{"prim":"pair","args":[{"prim":"nat","annots":["%token_index"]},{"prim":"pair","args":[{"prim":"nat","annots":["%min_amount_out"]},{"prim":"pair","args":[{"prim":"timestamp","annots":["%deadline"]},{"prim":"pair","args":[{"prim":"option","args":[{"prim":"address"}],"annots":["%receiver"]},{"prim":"option","args":[{"prim":"address"}],"annots":["%referral"]}]}]}]}]}]}],"annots":["%divest_one_coin"]}]}]},{"prim":"or","args":[{"prim":"or","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%pool_id"]},{"prim":"pair","args":[{"prim":"nat","annots":["%shares"]},{"prim":"pair","args":[{"prim":"map","args":[{"prim":"nat"},{"prim":"nat"}],"annots":["%in_amounts"]},{"prim":"pair","args":[{"prim":"timestamp","annots":["%deadline"]},{"prim":"pair","args":[{"prim":"option","args":[{"prim":"address"}],"annots":["%receiver"]},{"prim":"option","args":[{"prim":"address"}],"annots":["%referral"]}]}]}]}]}],"annots":["%invest"]},{"prim":"or","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%pool_id"]},{"prim":"nat","annots":["%amount"]}],"annots":["%add"]},{"prim":"pair","args":[{"prim":"nat","annots":["%pool_id"]},{"prim":"nat","annots":["%amount"]}],"annots":["%remove"]}],"annots":["%stake"]}]},{"prim":"pair","args":[{"prim":"nat","annots":["%pool_id"]},{"prim":"pair","args":[{"prim":"nat","annots":["%idx_from"]},{"prim":"pair","args":[{"prim":"nat","annots":["%idx_to"]},{"prim":"pair","args":[{"prim":"nat","annots":["%amount"]},{"prim":"pair","args":[{"prim":"nat","annots":["%min_amount_out"]},{"prim":"pair","args":[{"prim":"timestamp","annots":["%deadline"]},{"prim":"pair","args":[{"prim":"option","args":[{"prim":"address"}],"annots":["%receiver"]},{"prim":"option","args":[{"prim":"address"}],"annots":["%referral"]}]}]}]}]}]}]}],"annots":["%swap"]}]}]},{"prim":"pair","args":[{"prim":"address","annots":["%admin"]},{"prim":"pair","args":[{"prim":"address","annots":["%default_referral"]},{"prim":"pair","args":[{"prim":"set","args":[{"prim":"address"}],"annots":["%managers"]},{"prim":"pair","args":[{"prim":"nat","annots":["%pools_count"]},{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"nat"},{"prim":"map","args":[{"prim":"nat"},{"prim":"or","args":[{"prim":"address","annots":["%fa12"]},{"prim":"pair","args":[{"prim":"address","annots":["%token_address"]},{"prim":"nat","annots":["%token_id"]}],"annots":["%fa2"]}]}]}],"annots":["%tokens"]},{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"nat"}],"annots":["%pool_to_id"]},{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"nat"},{"prim":"pair","args":[{"prim":"nat","annots":["%initial_A_f"]},{"prim":"pair","args":[{"prim":"timestamp","annots":["%initial_A_time"]},{"prim":"pair","args":[{"prim":"nat","annots":["%future_A_f"]},{"prim":"pair","args":[{"prim":"timestamp","annots":["%future_A_time"]},{"prim":"pair","args":[{"prim":"map","args":[{"prim":"nat"},{"prim":"pair","args":[{"prim":"nat","annots":["%rate_f"]},{"prim":"pair","args":[{"prim":"nat","annots":["%precision_multiplier_f"]},{"prim":"nat","annots":["%reserves"]}]}]}],"annots":["%tokens_info"]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%lp_f"]},{"prim":"pair","args":[{"prim":"nat","annots":["%stakers_f"]},{"prim":"nat","annots":["%ref_f"]}]}],"annots":["%fee"]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"map","args":[{"prim":"nat"},{"prim":"nat"}],"annots":["%accumulator_f"]},{"prim":"nat","annots":["%total_staked"]}],"annots":["%staker_accumulator"]},{"prim":"nat","annots":["%total_supply"]}]}]}]}]}]}]}]}],"annots":["%pools"]},{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"pair","args":[{"prim":"address"},{"prim":"nat"}]},{"prim":"nat"}],"annots":["%ledger"]},{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"pair","args":[{"prim":"address"},{"prim":"nat"}]},{"prim":"set","args":[{"prim":"address"}]}],"annots":["%allowances"]},{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"or","args":[{"prim":"address","annots":["%fa12"]},{"prim":"pair","args":[{"prim":"address","annots":["%token_address"]},{"prim":"nat","annots":["%token_id"]}],"annots":["%fa2"]}]},{"prim":"nat"}],"annots":["%dev_rewards"]},{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"pair","args":[{"prim":"address"},{"prim":"or","args":[{"prim":"address","annots":["%fa12"]},{"prim":"pair","args":[{"prim":"address","annots":["%token_address"]},{"prim":"nat","annots":["%token_id"]}],"annots":["%fa2"]}]}]},{"prim":"nat"}],"annots":["%referral_rewards"]},{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"pair","args":[{"prim":"address"},{"prim":"nat"}]},{"prim":"pair","args":[{"prim":"nat","annots":["%balance"]},{"prim":"map","args":[{"prim":"nat"},{"prim":"pair","args":[{"prim":"nat","annots":["%reward_f"]},{"prim":"nat","annots":["%former_f"]}]}],"annots":["%earnings"]}]}],"annots":["%stakers_balance"]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address","annots":["%token_address"]},{"prim":"nat","annots":["%token_id"]}],"annots":["%quipu_token"]},{"prim":"pair","args":[{"prim":"address","annots":["%factory_address"]},{"prim":"bool","annots":["%started"]}]}]}]}]}]}]}]}]}]}]}]}]}]}]}]},{"prim":"pair","args":[{"prim":"list","args":[{"prim":"operation"}]},{"prim":"pair","args":[{"prim":"address","annots":["%admin"]},{"prim":"pair","args":[{"prim":"address","annots":["%default_referral"]},{"prim":"pair","args":[{"prim":"set","args":[{"prim":"address"}],"annots":["%managers"]},{"prim":"pair","args":[{"prim":"nat","annots":["%pools_count"]},{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"nat"},{"prim":"map","args":[{"prim":"nat"},{"prim":"or","args":[{"prim":"address","annots":["%fa12"]},{"prim":"pair","args":[{"prim":"address","annots":["%token_address"]},{"prim":"nat","annots":["%token_id"]}],"annots":["%fa2"]}]}]}],"annots":["%tokens"]},{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"nat"}],"annots":["%pool_to_id"]},{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"nat"},{"prim":"pair","args":[{"prim":"nat","annots":["%initial_A_f"]},{"prim":"pair","args":[{"prim":"timestamp","annots":["%initial_A_time"]},{"prim":"pair","args":[{"prim":"nat","annots":["%future_A_f"]},{"prim":"pair","args":[{"prim":"timestamp","annots":["%future_A_time"]},{"prim":"pair","args":[{"prim":"map","args":[{"prim":"nat"},{"prim":"pair","args":[{"prim":"nat","annots":["%rate_f"]},{"prim":"pair","args":[{"prim":"nat","annots":["%precision_multiplier_f"]},{"prim":"nat","annots":["%reserves"]}]}]}],"annots":["%tokens_info"]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%lp_f"]},{"prim":"pair","args":[{"prim":"nat","annots":["%stakers_f"]},{"prim":"nat","annots":["%ref_f"]}]}],"annots":["%fee"]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"map","args":[{"prim":"nat"},{"prim":"nat"}],"annots":["%accumulator_f"]},{"prim":"nat","annots":["%total_staked"]}],"annots":["%staker_accumulator"]},{"prim":"nat","annots":["%total_supply"]}]}]}]}]}]}]}]}],"annots":["%pools"]},{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"pair","args":[{"prim":"address"},{"prim":"nat"}]},{"prim":"nat"}],"annots":["%ledger"]},{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"pair","args":[{"prim":"address"},{"prim":"nat"}]},{"prim":"set","args":[{"prim":"address"}]}],"annots":["%allowances"]},{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"or","args":[{"prim":"address","annots":["%fa12"]},{"prim":"pair","args":[{"prim":"address","annots":["%token_address"]},{"prim":"nat","annots":["%token_id"]}],"annots":["%fa2"]}]},{"prim":"nat"}],"annots":["%dev_rewards"]},{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"pair","args":[{"prim":"address"},{"prim":"or","args":[{"prim":"address","annots":["%fa12"]},{"prim":"pair","args":[{"prim":"address","annots":["%token_address"]},{"prim":"nat","annots":["%token_id"]}],"annots":["%fa2"]}]}]},{"prim":"nat"}],"annots":["%referral_rewards"]},{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"pair","args":[{"prim":"address"},{"prim":"nat"}]},{"prim":"pair","args":[{"prim":"nat","annots":["%balance"]},{"prim":"map","args":[{"prim":"nat"},{"prim":"pair","args":[{"prim":"nat","annots":["%reward_f"]},{"prim":"nat","annots":["%former_f"]}]}],"annots":["%earnings"]}]}],"annots":["%stakers_balance"]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address","annots":["%token_address"]},{"prim":"nat","annots":["%token_id"]}],"annots":["%quipu_token"]},{"prim":"pair","args":[{"prim":"address","annots":["%factory_address"]},{"prim":"bool","annots":["%started"]}]}]}]}]}]}]}]}]}]}]}]}]}]}]}]}]}]},{"prim":"IF_NONE","args":[[{"prim":"FAILWITH"}],[{"prim":"SWAP"},{"prim":"DROP"}]]},{"prim":"DUP","args":[{"int":"3"}]},{"prim":"CAR"},{"prim":"DIG","args":[{"int":"2"}]},{"prim":"PAIR"},{"prim":"EXEC"},{"prim":"DUP"},{"prim":"DUG","args":[{"int":"2"}]},{"prim":"CDR"},{"prim":"UPDATE","args":[{"int":"1"}]},{"prim":"SWAP"},{"prim":"CAR"},{"prim":"PAIR"}],[{"prim":"DUG","args":[{"int":"2"}]},{"prim":"DUP"},{"prim":"DUG","args":[{"int":"3"}]},{"prim":"CAR"},{"prim":"GET","args":[{"int":"28"}]},{"prim":"IF","args":[[{"prim":"DROP"}],[{"prim":"FAILWITH"}]]},{"prim":"DIG","args":[{"int":"3"}]},{"prim":"DUP","args":[{"int":"3"}]},{"prim":"GET","args":[{"int":"10"}]},{"prim":"DUP","args":[{"int":"3"}]},{"prim":"IF_LEFT","args":[[{"prim":"IF_LEFT","args":[[{"prim":"IF_LEFT","args":[[{"prim":"DROP"},{"prim":"PUSH","args":[{"prim":"nat"},{"int":"1"}]}],[{"prim":"DROP"},{"prim":"PUSH","args":[{"prim":"nat"},{"int":"4"}]}]]}],[{"prim":"IF_LEFT","args":[[{"prim":"DROP"},{"prim":"PUSH","args":[{"prim":"nat"},{"int":"0"}]}],[{"prim":"DROP"},{"prim":"PUSH","args":[{"prim":"nat"},{"int":"3"}]}]]}]]}],[{"prim":"DROP"},{"prim":"PUSH","args":[{"prim":"nat"},{"int":"2"}]}]]},{"prim":"GET"},{"prim":"IF_NONE","args":[[{"prim":"FAILWITH"}],[{"prim":"SWAP"},{"prim":"DROP"}]]},{"prim":"DIG","args":[{"int":"3"}]},{"prim":"SWAP"},{"prim":"UNPACK","args":[{"prim":"lambda","args":[{"prim":"pair","args":[{"prim":"or","args":[{"prim":"or","args":[{"prim":"or","args":[{"prim":"pair","args":[{"prim":"list","args":[{"prim":"pair","args":[{"prim":"address","annots":["%owner"]},{"prim":"nat","annots":["%token_id"]}]}],"annots":["%requests"]},{"prim":"contract","args":[{"prim":"list","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address","annots":["%owner"]},{"prim":"nat","annots":["%token_id"]}],"annots":["%request"]},{"prim":"nat","annots":["%balance"]}]}]}],"annots":["%callback"]}],"annots":["%balance_of"]},{"prim":"pair","args":[{"prim":"nat","annots":["%token_id"]},{"prim":"contract","args":[{"prim":"nat"}],"annots":["%receiver"]}],"annots":["%total_supply"]}]},{"prim":"or","args":[{"prim":"list","args":[{"prim":"pair","args":[{"prim":"address","annots":["%from_"]},{"prim":"list","args":[{"prim":"pair","args":[{"prim":"address","annots":["%to_"]},{"prim":"pair","args":[{"prim":"nat","annots":["%token_id"]},{"prim":"nat","annots":["%amount"]}]}]}],"annots":["%txs"]}]}],"annots":["%transfer"]},{"prim":"pair","args":[{"prim":"nat","annots":["%token_id"]},{"prim":"map","args":[{"prim":"string"},{"prim":"bytes"}],"annots":["%token_info"]}],"annots":["%update_metadata"]}]}]},{"prim":"list","args":[{"prim":"or","args":[{"prim":"pair","args":[{"prim":"address","annots":["%owner"]},{"prim":"pair","args":[{"prim":"address","annots":["%operator"]},{"prim":"nat","annots":["%token_id"]}]}],"annots":["%add_operator"]},{"prim":"pair","args":[{"prim":"address","annots":["%owner"]},{"prim":"pair","args":[{"prim":"address","annots":["%operator"]},{"prim":"nat","annots":["%token_id"]}]}],"annots":["%remove_operator"]}]}],"annots":["%update_operators"]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address","annots":["%admin"]},{"prim":"pair","args":[{"prim":"address","annots":["%default_referral"]},{"prim":"pair","args":[{"prim":"set","args":[{"prim":"address"}],"annots":["%managers"]},{"prim":"pair","args":[{"prim":"nat","annots":["%pools_count"]},{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"nat"},{"prim":"map","args":[{"prim":"nat"},{"prim":"or","args":[{"prim":"address","annots":["%fa12"]},{"prim":"pair","args":[{"prim":"address","annots":["%token_address"]},{"prim":"nat","annots":["%token_id"]}],"annots":["%fa2"]}]}]}],"annots":["%tokens"]},{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"nat"}],"annots":["%pool_to_id"]},{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"nat"},{"prim":"pair","args":[{"prim":"nat","annots":["%initial_A_f"]},{"prim":"pair","args":[{"prim":"timestamp","annots":["%initial_A_time"]},{"prim":"pair","args":[{"prim":"nat","annots":["%future_A_f"]},{"prim":"pair","args":[{"prim":"timestamp","annots":["%future_A_time"]},{"prim":"pair","args":[{"prim":"map","args":[{"prim":"nat"},{"prim":"pair","args":[{"prim":"nat","annots":["%rate_f"]},{"prim":"pair","args":[{"prim":"nat","annots":["%precision_multiplier_f"]},{"prim":"nat","annots":["%reserves"]}]}]}],"annots":["%tokens_info"]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%lp_f"]},{"prim":"pair","args":[{"prim":"nat","annots":["%stakers_f"]},{"prim":"nat","annots":["%ref_f"]}]}],"annots":["%fee"]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"map","args":[{"prim":"nat"},{"prim":"nat"}],"annots":["%accumulator_f"]},{"prim":"nat","annots":["%total_staked"]}],"annots":["%staker_accumulator"]},{"prim":"nat","annots":["%total_supply"]}]}]}]}]}]}]}]}],"annots":["%pools"]},{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"pair","args":[{"prim":"address"},{"prim":"nat"}]},{"prim":"nat"}],"annots":["%ledger"]},{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"pair","args":[{"prim":"address"},{"prim":"nat"}]},{"prim":"set","args":[{"prim":"address"}]}],"annots":["%allowances"]},{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"or","args":[{"prim":"address","annots":["%fa12"]},{"prim":"pair","args":[{"prim":"address","annots":["%token_address"]},{"prim":"nat","annots":["%token_id"]}],"annots":["%fa2"]}]},{"prim":"nat"}],"annots":["%dev_rewards"]},{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"pair","args":[{"prim":"address"},{"prim":"or","args":[{"prim":"address","annots":["%fa12"]},{"prim":"pair","args":[{"prim":"address","annots":["%token_address"]},{"prim":"nat","annots":["%token_id"]}],"annots":["%fa2"]}]}]},{"prim":"nat"}],"annots":["%referral_rewards"]},{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"pair","args":[{"prim":"address"},{"prim":"nat"}]},{"prim":"pair","args":[{"prim":"nat","annots":["%balance"]},{"prim":"map","args":[{"prim":"nat"},{"prim":"pair","args":[{"prim":"nat","annots":["%reward_f"]},{"prim":"nat","annots":["%former_f"]}]}],"annots":["%earnings"]}]}],"annots":["%stakers_balance"]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address","annots":["%token_address"]},{"prim":"nat","annots":["%token_id"]}],"annots":["%quipu_token"]},{"prim":"pair","args":[{"prim":"address","annots":["%factory_address"]},{"prim":"bool","annots":["%started"]}]}]}]}]}]}]}]}]}]}]}]}]}]}],"annots":["%storage"]},{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"string"},{"prim":"bytes"}],"annots":["%metadata"]},{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"nat"},{"prim":"pair","args":[{"prim":"nat","annots":["%token_id"]},{"prim":"map","args":[{"prim":"string"},{"prim":"bytes"}],"annots":["%token_info"]}]}],"annots":["%token_metadata"]},{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"nat"},{"prim":"bytes"}],"annots":["%admin_lambdas"]},{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"nat"},{"prim":"bytes"}],"annots":["%dex_lambdas"]},{"prim":"big_map","args":[{"prim":"nat"},{"prim":"bytes"}],"annots":["%token_lambdas"]}]}]}]}]}]}]},{"prim":"pair","args":[{"prim":"list","args":[{"prim":"operation"}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address","annots":["%admin"]},{"prim":"pair","args":[{"prim":"address","annots":["%default_referral"]},{"prim":"pair","args":[{"prim":"set","args":[{"prim":"address"}],"annots":["%managers"]},{"prim":"pair","args":[{"prim":"nat","annots":["%pools_count"]},{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"nat"},{"prim":"map","args":[{"prim":"nat"},{"prim":"or","args":[{"prim":"address","annots":["%fa12"]},{"prim":"pair","args":[{"prim":"address","annots":["%token_address"]},{"prim":"nat","annots":["%token_id"]}],"annots":["%fa2"]}]}]}],"annots":["%tokens"]},{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"nat"}],"annots":["%pool_to_id"]},{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"nat"},{"prim":"pair","args":[{"prim":"nat","annots":["%initial_A_f"]},{"prim":"pair","args":[{"prim":"timestamp","annots":["%initial_A_time"]},{"prim":"pair","args":[{"prim":"nat","annots":["%future_A_f"]},{"prim":"pair","args":[{"prim":"timestamp","annots":["%future_A_time"]},{"prim":"pair","args":[{"prim":"map","args":[{"prim":"nat"},{"prim":"pair","args":[{"prim":"nat","annots":["%rate_f"]},{"prim":"pair","args":[{"prim":"nat","annots":["%precision_multiplier_f"]},{"prim":"nat","annots":["%reserves"]}]}]}],"annots":["%tokens_info"]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%lp_f"]},{"prim":"pair","args":[{"prim":"nat","annots":["%stakers_f"]},{"prim":"nat","annots":["%ref_f"]}]}],"annots":["%fee"]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"map","args":[{"prim":"nat"},{"prim":"nat"}],"annots":["%accumulator_f"]},{"prim":"nat","annots":["%total_staked"]}],"annots":["%staker_accumulator"]},{"prim":"nat","annots":["%total_supply"]}]}]}]}]}]}]}]}],"annots":["%pools"]},{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"pair","args":[{"prim":"address"},{"prim":"nat"}]},{"prim":"nat"}],"annots":["%ledger"]},{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"pair","args":[{"prim":"address"},{"prim":"nat"}]},{"prim":"set","args":[{"prim":"address"}]}],"annots":["%allowances"]},{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"or","args":[{"prim":"address","annots":["%fa12"]},{"prim":"pair","args":[{"prim":"address","annots":["%token_address"]},{"prim":"nat","annots":["%token_id"]}],"annots":["%fa2"]}]},{"prim":"nat"}],"annots":["%dev_rewards"]},{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"pair","args":[{"prim":"address"},{"prim":"or","args":[{"prim":"address","annots":["%fa12"]},{"prim":"pair","args":[{"prim":"address","annots":["%token_address"]},{"prim":"nat","annots":["%token_id"]}],"annots":["%fa2"]}]}]},{"prim":"nat"}],"annots":["%referral_rewards"]},{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"pair","args":[{"prim":"address"},{"prim":"nat"}]},{"prim":"pair","args":[{"prim":"nat","annots":["%balance"]},{"prim":"map","args":[{"prim":"nat"},{"prim":"pair","args":[{"prim":"nat","annots":["%reward_f"]},{"prim":"nat","annots":["%former_f"]}]}],"annots":["%earnings"]}]}],"annots":["%stakers_balance"]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address","annots":["%token_address"]},{"prim":"nat","annots":["%token_id"]}],"annots":["%quipu_token"]},{"prim":"pair","args":[{"prim":"address","annots":["%factory_address"]},{"prim":"bool","annots":["%started"]}]}]}]}]}]}]}]}]}]}]}]}]}]}],"annots":["%storage"]},{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"string"},{"prim":"bytes"}],"annots":["%metadata"]},{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"nat"},{"prim":"pair","args":[{"prim":"nat","annots":["%token_id"]},{"prim":"map","args":[{"prim":"string"},{"prim":"bytes"}],"annots":["%token_info"]}]}],"annots":["%token_metadata"]},{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"nat"},{"prim":"bytes"}],"annots":["%admin_lambdas"]},{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"nat"},{"prim":"bytes"}],"annots":["%dex_lambdas"]},{"prim":"big_map","args":[{"prim":"nat"},{"prim":"bytes"}],"annots":["%token_lambdas"]}]}]}]}]}]}]}]}]},{"prim":"IF_NONE","args":[[{"prim":"FAILWITH"}],[{"prim":"SWAP"},{"prim":"DROP"}]]},{"prim":"DUG","args":[{"int":"2"}]},{"prim":"PAIR"},{"prim":"EXEC"}]]}]]}]]`),
						},
					},
				},
			},
		}, {
			name: "opPDkVe1nU5xqLyoWYQ2r6H7PaJM5S4Pe4WtTmEE7UMQAwfnuiJ",
			ctx: &config.Context{
				RPC:             rpc,
				Storage:         generalRepo,
				Contracts:       contractRepo,
				BigMapDiffs:     bmdRepo,
				Blocks:          blockRepo,
				Protocols:       protoRepo,
				Operations:      operaitonsRepo,
				Scripts:         scriptRepo,
				GlobalConstants: globalConstantRepo,
				Cache: cache.NewCache(
					rpc, accountsRepo, contractRepo, protoRepo, bluemonday.UGCPolicy(),
				),
			},
			paramsOpts: []ParseParamsOption{
				WithHead(noderpc.Header{
					Timestamp: timestamp,
					Protocol:  "Psithaca2MLRFYargivpo7YvUr7wUDqyxrdhC5CQq78mRvimz6A",
					Level:     707452,
					ChainID:   "NetXdQprcVkpaWU",
				}),
				WithProtocol(&protocol.Protocol{
					Constants: &protocol.Constants{
						CostPerByte:                  1000,
						HardGasLimitPerOperation:     400000,
						HardStorageLimitPerOperation: 60000,
						TimeBetweenBlocks:            60,
					},
					Hash:    "Psithaca2MLRFYargivpo7YvUr7wUDqyxrdhC5CQq78mRvimz6A",
					ID:      6,
					SymLink: bcd.SymLinkBabylon,
				}),
			},
			storage: map[string]int64{
				"KT1KRzp5hckBwLLswCreweLMdueL3jJhTN1S": 707452,
			},
			filename: "./data/rpc/opg/opPDkVe1nU5xqLyoWYQ2r6H7PaJM5S4Pe4WtTmEE7UMQAwfnuiJ.json",
			want: &parsers.TestStore{
				Operations: []*operation.Operation{
					{
						Kind: types.OperationKindOrigination,
						Source: account.Account{

							Address: "tz1VSUr8wwNhLAzempoch5d6hLRiTh8Cjcjb",
							Type:    types.AccountTypeTz,
						},
						Fee:                                724,
						Counter:                            12837,
						GasLimit:                           3326,
						StorageLimit:                       386,
						Burned:                             386000,
						AllocatedDestinationContractBurned: 257000,
						Destination: account.Account{
							Address: "KT1KRzp5hckBwLLswCreweLMdueL3jJhTN1S",
							Type:    types.AccountTypeContract,
						},
						Status:    types.OperationStatusApplied,
						Level:     707452,
						Hash:      "opPDkVe1nU5xqLyoWYQ2r6H7PaJM5S4Pe4WtTmEE7UMQAwfnuiJ",
						Timestamp: timestamp,
						Initiator: account.Account{
							Address: "tz1VSUr8wwNhLAzempoch5d6hLRiTh8Cjcjb",
							Type:    types.AccountTypeTz,
						},
						Delegate:        account.Account{},
						ProtocolID:      6,
						DeffatedStorage: []byte(`[[{"string":"tz1VSUr8wwNhLAzempoch5d6hLRiTh8Cjcjb"},{"prim":"None"}],{"string":"KT1LnwXLwrH3ejammJ1CJFgezpsehXFDNREU"}]`),
					},
				},
				Contracts: []*modelContract.Contract{
					{
						Account: account.Account{
							Address: "KT1KRzp5hckBwLLswCreweLMdueL3jJhTN1S",
							Type:    types.AccountTypeContract,
						},
						Level:     707452,
						Timestamp: timestamp,
						Manager: account.Account{
							Address: "tz1VSUr8wwNhLAzempoch5d6hLRiTh8Cjcjb",
							Type:    types.AccountTypeTz,
						},
						Delegate: account.Account{},
						Babylon: modelContract.Script{
							Hash:        "32819b8ddb086cf164e0020cfc23bddfbb202e15fd3113d79ae0850f6f594f7f",
							Annotations: []string{"%pending", "%whitelist_contract", "%lambda", "%update_admin", "%admin", "%current"},
							FailStrings: []string{"NOT_PENDING_ADMIN", "ADDRESS_NOT_WHITELISTED", "SENDER_NOT_ADMIN", "CALL_ARE_WHITELISED_VIEW_FAILED", "NO_PENDING_ADMIN"},
							Entrypoints: []string{"lambda", "update_admin"},
							Parameter:   []byte(`[{"prim":"or","args":[{"prim":"lambda","args":[{"prim":"unit"},{"prim":"list","args":[{"prim":"operation"}]}],"annots":["%lambda"]},{"prim":"option","args":[{"prim":"address"}],"annots":["%update_admin"]}]}]`),
							Storage:     []byte(`[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address","annots":["%current"]},{"prim":"option","args":[{"prim":"address"}],"annots":["%pending"]}],"annots":["%admin"]},{"prim":"address","annots":["%whitelist_contract"]}]}]`),
							Code:        []byte(`[[{"prim":"UNPAIR"},{"prim":"IF_LEFT","args":[[{"prim":"SWAP"},{"prim":"DUP"},{"prim":"DUG","args":[{"int":"2"}]},{"prim":"CAR"},{"prim":"CAR"},{"prim":"SENDER"},{"prim":"COMPARE"},{"prim":"NEQ"},{"prim":"IF","args":[[{"prim":"PUSH","args":[{"prim":"string"},{"string":"SENDER_NOT_ADMIN"}]},{"prim":"FAILWITH"}],[]]},{"prim":"SWAP"},{"prim":"UNIT"},{"prim":"DIG","args":[{"int":"2"}]},{"prim":"SWAP"},{"prim":"EXEC"},{"prim":"PAIR"}],[{"prim":"SWAP"},{"prim":"DUP"},{"prim":"DUG","args":[{"int":"2"}]},{"prim":"CDR"},{"prim":"NIL","args":[{"prim":"address"}]},{"prim":"SENDER"},{"prim":"CONS"},{"prim":"VIEW","args":[{"string":"are_whitelisted"},{"prim":"bool"}]},{"prim":"IF_NONE","args":[[{"prim":"PUSH","args":[{"prim":"string"},{"string":"CALL_ARE_WHITELISED_VIEW_FAILED"}]},{"prim":"FAILWITH"}],[]]},{"prim":"IF","args":[[{"prim":"SWAP"},{"prim":"DUP"},{"prim":"DUG","args":[{"int":"2"}]},{"prim":"CAR"},{"prim":"SWAP"},{"prim":"IF_NONE","args":[[{"prim":"CDR"},{"prim":"IF_NONE","args":[[{"prim":"PUSH","args":[{"prim":"string"},{"string":"NO_PENDING_ADMIN"}]},{"prim":"FAILWITH"}],[{"prim":"DUP"},{"prim":"SENDER"},{"prim":"COMPARE"},{"prim":"NEQ"},{"prim":"IF","args":[[{"prim":"DROP"},{"prim":"PUSH","args":[{"prim":"string"},{"string":"NOT_PENDING_ADMIN"}]},{"prim":"FAILWITH"}],[{"prim":"NONE","args":[{"prim":"address"}]},{"prim":"SWAP"},{"prim":"PAIR"}]]}]]}],[{"prim":"SWAP"},{"prim":"DUP"},{"prim":"DUG","args":[{"int":"2"}]},{"prim":"CAR"},{"prim":"SENDER"},{"prim":"COMPARE"},{"prim":"NEQ"},{"prim":"IF","args":[[{"prim":"PUSH","args":[{"prim":"string"},{"string":"SENDER_NOT_ADMIN"}]},{"prim":"FAILWITH"}],[]]},{"prim":"SOME"},{"prim":"UPDATE","args":[{"int":"2"}]}]]},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"DUG","args":[{"int":"2"}]},{"prim":"UPDATE","args":[{"int":"1"}]},{"prim":"SWAP"},{"prim":"PAIR"}],[{"prim":"DROP","args":[{"int":"2"}]},{"prim":"PUSH","args":[{"prim":"string"},{"string":"ADDRESS_NOT_WHITELISTED"}]},{"prim":"FAILWITH"}]]}]]}]]`),
							Constants: []modelContract.GlobalConstant{
								{
									ID:        1,
									Timestamp: timestamp,
									Level:     707452,
									Address:   "exprv5uiw7xXoEgRahR3YBn4iAVwfkNCMsrkneutuBZCGG5sS64kRw",
									Value:     []byte(`[{"prim":"parameter","args":[{"prim":"or","args":[{"prim":"lambda","args":[{"prim":"unit"},{"prim":"list","args":[{"prim":"operation"}]}],"annots":["%lambda"]},{"prim":"option","args":[{"prim":"address"}],"annots":["%update_admin"]}]}]},{"prim":"storage","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address","annots":["%current"]},{"prim":"option","args":[{"prim":"address"}],"annots":["%pending"]}],"annots":["%admin"]},{"prim":"address","annots":["%whitelist_contract"]}]}]},{"prim":"code","args":[[{"prim":"UNPAIR"},{"prim":"IF_LEFT","args":[[{"prim":"SWAP"},{"prim":"DUP"},{"prim":"DUG","args":[{"int":"2"}]},{"prim":"CAR"},{"prim":"CAR"},{"prim":"SENDER"},{"prim":"COMPARE"},{"prim":"NEQ"},{"prim":"IF","args":[[{"prim":"PUSH","args":[{"prim":"string"},{"string":"SENDER_NOT_ADMIN"}]},{"prim":"FAILWITH"}],[]]},{"prim":"SWAP"},{"prim":"UNIT"},{"prim":"DIG","args":[{"int":"2"}]},{"prim":"SWAP"},{"prim":"EXEC"},{"prim":"PAIR"}],[{"prim":"SWAP"},{"prim":"DUP"},{"prim":"DUG","args":[{"int":"2"}]},{"prim":"CDR"},{"prim":"NIL","args":[{"prim":"address"}]},{"prim":"SENDER"},{"prim":"CONS"},{"prim":"VIEW","args":[{"string":"are_whitelisted"},{"prim":"bool"}]},{"prim":"IF_NONE","args":[[{"prim":"PUSH","args":[{"prim":"string"},{"string":"CALL_ARE_WHITELISED_VIEW_FAILED"}]},{"prim":"FAILWITH"}],[]]},{"prim":"IF","args":[[{"prim":"SWAP"},{"prim":"DUP"},{"prim":"DUG","args":[{"int":"2"}]},{"prim":"CAR"},{"prim":"SWAP"},{"prim":"IF_NONE","args":[[{"prim":"CDR"},{"prim":"IF_NONE","args":[[{"prim":"PUSH","args":[{"prim":"string"},{"string":"NO_PENDING_ADMIN"}]},{"prim":"FAILWITH"}],[{"prim":"DUP"},{"prim":"SENDER"},{"prim":"COMPARE"},{"prim":"NEQ"},{"prim":"IF","args":[[{"prim":"DROP"},{"prim":"PUSH","args":[{"prim":"string"},{"string":"NOT_PENDING_ADMIN"}]},{"prim":"FAILWITH"}],[{"prim":"NONE","args":[{"prim":"address"}]},{"prim":"SWAP"},{"prim":"PAIR"}]]}]]}],[{"prim":"SWAP"},{"prim":"DUP"},{"prim":"DUG","args":[{"int":"2"}]},{"prim":"CAR"},{"prim":"SENDER"},{"prim":"COMPARE"},{"prim":"NEQ"},{"prim":"IF","args":[[{"prim":"PUSH","args":[{"prim":"string"},{"string":"SENDER_NOT_ADMIN"}]},{"prim":"FAILWITH"}],[]]},{"prim":"SOME"},{"prim":"UPDATE","args":[{"int":"2"}]}]]},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"DUG","args":[{"int":"2"}]},{"prim":"UPDATE","args":[{"int":"1"}]},{"prim":"SWAP"},{"prim":"PAIR"}],[{"prim":"DROP","args":[{"int":"2"}]},{"prim":"PUSH","args":[{"prim":"string"},{"string":"ADDRESS_NOT_WHITELISTED"}]},{"prim":"FAILWITH"}]]}]]}]]},{"prim":"view","args":[{"string":"admin"},{"prim":"unit"},{"prim":"address"},[{"prim":"CDR"},{"prim":"CAR"},{"prim":"CAR"}]]}]`),
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for address, level := range tt.storage {
				rpc.
					EXPECT().
					GetScriptStorageRaw(context.Background(), address, level).
					DoAndReturn(
						func(_ context.Context, address string, level int64) ([]byte, error) {
							storageFile := fmt.Sprintf("./data/rpc/script/storage/%s_%d.json", address, level)
							return os.ReadFile(storageFile)
						},
					).
					AnyTimes()
			}

			var op noderpc.LightOperationGroup
			if err := readJSONFile(tt.filename, &op); err != nil {
				t.Errorf(`readJSONFile("%s") = error %v`, tt.filename, err)
				return
			}

			parseParams, err := NewParseParams(tt.ctx, tt.paramsOpts...)
			if err != nil {
				t.Errorf(`NewParseParams = error %v`, err)
				return
			}

			store := parsers.NewTestStore()
			if err := NewGroup(parseParams).Parse(op, store); (err != nil) != tt.wantErr {
				t.Errorf("Group.Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !compareParserResponse(t, store, tt.want) {
				t.Errorf("Group.Parse() = %#v, want %#v", store, tt.want)
			}
		})
	}
}
