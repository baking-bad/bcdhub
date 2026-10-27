package forge

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Contract(t *testing.T) {
	tests := []struct {
		name    string
		val     string
		want    string
		wantErr bool
	}{
		{
			name: "tz1eLWfccL46VAUjtyz9kEKgzuKnwyZH4rTA",
			val:  "tz1eLWfccL46VAUjtyz9kEKgzuKnwyZH4rTA",
			want: "0000cd1a410ffd5315ded34337f5f76edff48a13999a",
		}, {
			name: "Case 1",
			val:  "tz1d75oB6T4zUMexzkr5WscGktZ1Nss1JrT7",
			want: "0000bf97f5f1dbfd6ada0cf986d0a812f1bf0a572abc",
		},
		{
			name: "tz1 address",
			val:  `tz1a5fMLLY5WCarCzH7RKTJHX9mJFN8eaaWG`,
			want: "00009e6ac2e529a49aedbcdd0ac9542d5c0f4ce76f77",
		},
		{
			name: "tz1 address",
			val:  "tz1MBqYpcoGU93c1bePp5A6dmwKYjmHXRopS",
			want: "000010fc2282886d9cf8a1eebdc2733e302c7b110f38",
		},
		{
			name: "tz1 address",
			val:  "tz1LFEVYR7YRCxT6Nm3Zfjdnfj77xZqhbR5U",
			want: "000006a868bd80219eb1f6a25108d1bdaa98ae27b2d9",
		},
		{
			name: "tz1 address",
			val:  "tz1RABAzdLWVvxAFf1wpeUALAkp32mVhSGXX",
			want: "00003c8c2fe0f75ce212558df94c7a7306c2eeadd979",
		},
		{
			name: "tz1 address",
			val:  "tz1SZZgtvMVXaBKPcez4gfjKUsDz1gs6vg6X",
			want: "00004bf0acca4cc9e034b1d5f0f783c78e5ed44d866e",
		},
		{
			name: "tz1 address",
			val:  "tz1WkaeRycRr999GrVFepJd9Nqi1FWqGyGqq",
			want: "000079e68d8f0a8d64ec856e193efc0a347ef4adf8ee",
		},
		{
			name: "tz1 address",
			val:  "tz1eq3gqb2iZHjHVHoPJqV84gZdBF2TMQiH4",
			want: "0000d27fcbd31910d2226ba4c8f646d3d4c7b2f3a756",
		},
		{
			name: "tz1 address",
			val:  "tz1M9CMEtsXm3QxA7FmMU2Qh7xzsuGXVbcDr",
			want: "0000107c4009f2bcfcc248d6952998af5b7203b8ff59",
		},
		{
			name: "tz2 address",
			val:  "tz28YZoayJjVz2bRgGeVjxE8NonMiJ3r2Wdu",
			want: "0001028562fb176188114cf437a757cdc75bc4aa8cae",
		},
		{
			name: "tz3 address",
			val:  "tz3agP9LGe2cXmKQyYn6T68BHKjjktDbbSWX",
			want: "00029d6a61cd3510193e257128da8f09a0b173bff695",
		},
		{
			name: "KT address",
			val:  "KT1J8T7U6J1BAo9fJAxvedHsNErnejwvPyUH",
			want: "0168b709e887ddc34c3c9e468b5819b2f012b60ef700",
		},
		{
			name: "KT address",
			val:  "KT1BUKeJTemAaVBfRz6cqxeUBQGQqMxfG19A",
			want: "011fb03e3ff9fedaf3a2200ffc64d27812da734bba00",
		},
		{
			name: "KT address",
			val:  "KT1U1JZaXoG4u1EPnhHL4R4otzkWc1L34q3c",
			want: "01d50e3f6f059dc86f5591455549313ce42d0c50f100",
		},
		{
			name: "KT address",
			val:  "KT1XHAmdRKugP1Q38CxDmpcRSxq143KpEiYx",
			want: "01f8f6c6a0af7c20251bc7df108f2a6e2879a06c9a00",
		}, {
			name: "KT address with entrypoint",
			val:  `KT1Nh9wK8W3j3CXeTVm5DTTaiU5RE8CxLWZ4%receive_bunny_balance`,
			want: "019ac6ee79c4e87a21d094bb8bf00f37fe51717e8700726563656976655f62756e6e795f62616c616e6365",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Contract(tt.val)
			require.Equal(t, tt.wantErr, err != nil)
			if err != nil {
				return
			}
			require.Equal(t, tt.want, got)
		})
	}
}

func TestUnforgeBakerHash(t *testing.T) {
	tests := []struct {
		name    string
		str     string
		want    string
		wantErr bool
	}{
		{
			name: "test 1",
			str:  "94697e9229c88fac7d19d62e139ca6735f9569dd",
			want: "SG1d1wsgMKvSstzZQ8L4WoskCesdWGzVt5k4",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := UnforgeBakerHash(tt.str)
			require.Equal(t, tt.wantErr, err != nil)
			if err != nil {
				return
			}
			require.Equal(t, tt.want, got)
		})
	}
}

func TestUnforgeAddress(t *testing.T) {
	tests := []struct {
		name    string
		str     string
		want    string
		wantErr bool
	}{
		{
			name: "test 1",
			str:  "016e4943f7a23ab9cbe56f48ff72f6c27e8956762400",
			want: "KT1JdufSdfg3WyxWJcCRNsBFV9V3x9TQBkJ2",
		}, {
			name: "test 2",
			str:  "00003a96709901319a5da2968782279dae581b9ba4",
			want: "tz1KfEsrtDaA1sX7vdM4qmEPWuSytuqCDp5j",
		}, {
			name: "tz4",
			str:  "00032fcb1d9307f0b1f94c048ff586c09f46614c7e90",
			want: "tz4DMxfcJJSWuDFSJvRR7YmLXZknFEnZdYs6",
		}, {
			name: "tz5",
			str:  "0004abababababababababababababababababababab",
			want: "tz5cfmjtZ2vdDmkB8gyGQQto6eGjyswGViGK",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := UnforgeAddress(tt.str)
			require.Equal(t, tt.wantErr, err != nil)
			if err != nil {
				return
			}
			require.Equal(t, tt.want, got)
		})
	}
}

// ML-DSA-44 public key from KT1TqiYUk7hSGvjv4JbBBuoCDnP2hhSYxbbM storage (shadownet)
const (
	mldsa44PublicKeyForged = "043fca7463fe261863908f801ca6ee71a6ddc629e1fef51541ec7cefee928d3229d6c13bc5d13edbde64d71b0b8f0f7147361de07d0ae7a08de8fdffcb212e22a77c9933162e8ceda7e5938f8fe0cfef03ca35faf7ad6023341a4fa31a438fb2afd5d331c4f03a5249c6f6acab7eb8712e35e48469c8adb9e9a466df7beed2b9588defbbfa94ac4fe3c0fe0650895b55c92d6e3d1171829b5c0b7fe56fce13706ec995d77c33147c0d9adf5e970479838b56a10439cbdfbd4ac152f625e5878c8fd2833da87a5e556bd0fba5a71f1dd6ff7a4523db5a65767dd243022130986eecf5ce976b36f29bfa00fd8a994e28b700f92b498bacdb78fedc6133eba4f9d3357b8c7ebee5648cbc18821e59b17cc61323af9f9abf86224f19a206ea0c0ac4fa95ea5a972feceec22ce2af610ee79a19f52b8c5ed4adbcdd71718b21f908408fff9b4589165bdfaad8290e0a38a6ada8ff784ad722f6db8ad790db947d9f66d78d9935d45209e03cc540e0c290e5de05e724ebfa565bac0a69a2d81b6fedd5c923fed3a3362b83f1224a93877d85b955610c3631166a40a6d0ff29d69cc5a4e3d0ebedbdc7d9982c0ddab5c42167f11c24ac35bb84a5d6196cb32070bc4c73f0ab3a35cdeb62f1a9536ed4edae452ad58ca8aba996c1c9613fe89fd392ce9374ee748671e6319942c4aafa917c49319632a7b73a288f8a832b0d956b2f49564c6ef18d010ca2e22fb3e8ffbe01911f0254df8b0e3397afdb785d5d497a28e70dd772307fc90977f4c03eeaf7cc56928d88067ebcd9d73e30fa4049608d57bbfd6961f15eff961704a8a5bb43c557b79ed08adbf5000f48981d494404ff6eae1e423bfa722b0ddbf811a6daff06587e4b806446630329e4fe803820b57abd2264c359e93a49d0a080afbab34037a50daefa7b4c8763672fcb4f355ccd045b87b9398d571c377e6706bc2c57d2c38195dbb1d7ca392a79921e5a574befe1fae30920af147aef53700c8f0b6fb13499f1bacc59ff9775188bd2058d405548ba1e121ab3a708b7f8fea53340e0792f2d2bcd1913fbc29f54a74e2f1aecde181255febfbb08b33f0729ff16abeadb8d7cacd2053553615ed501fa2c352e43eec5d0a6556307eb18d206f051d9f0e0bbb7a32e56031749804ff92f7a6d36835d255aba1eae589fca45793206cfebb758362dbc33666eed02a2d81b199c179ee18bb97789a0102f4e7fa72cabb1b4782bf9a1dda220e4a7107405e2d3d8d5598a2cc35e84002301c546b128737b676d2ee853768a326d875de4a017c05b30c3411b50a0330ad0f014cc45e5b7bff40d973e9b08f70b5a65f6962d1a3a49d68ae8d07ea5f002526fd32603b4d3574200ea58bd986785ca8524eeac2ca0fcdc07643a21cb9166994115291eb5b5943b37b2449b1cb2c56c1d249c4f448c9e997afa8821df8195ad0955f72bc49ab7970829d16b27613825ee0f31ccdd0355caffa2317a9440090307138181bf79c7510f65e7633b79bbdff37c7ffa173cd625ab4ac6488f0ac7d8b7fccb91cffa9d30d0e02e094ed218a0c973d2134148803fde50a43e31eca4ad8f22280e93b620a9a5b946a1061c69a7fed6e0d5f94aef6a186176807f7bcc1c0f60b41e4e039ce4913593c110f23b5fc1657901bab418d783c1cb4680bdc50673f18d63c079dc7611c07ec2f13deabd57f7d9a987d82ce0eec06b87ffbdc8a52c69bf6c4e0ee66afe953b2acf057defb95c8c6ae84a8631579b0891362559e10c32c4bb3a223ab7cae664b917ff07bd5e94c4517835d32ed57df6959b2e441555751376500b699879a053cea78d42d3e3d906d56df7229012f34c147d"
	mldsa44PublicKey       = "mdpk2aR8tKbMnAisEqcwuU8sRu2827NxvZipoLrjxn3kD4g8zfpqEaF5vKrq7xncHMu9nrChEcD61VX1unsmHhUXWZuHMkE3JoLPytrQQSeRArq1JUdpuMKsnXF3HAzuEVQceFW8f923d82Vre25GiX1xJRxjrWrs2ijhxWLiiCcjFSzQKEYbDwotd6Y23Eao5Nr1nmwZksZeuippx24gzze4FnceEHbzq7JQ9fppSXtzEqWbNF5M5SS3TYPY2bz2mNPJNoGuiMwrkBHwr55Lp2FzTYT3HUK3Y4PFag8SwLsEy5bciWKovccNhHyEquRhhUqs5fHY3UKftyAoSdVrK89VQpVDTtVgEWDMivgRCm1xPYkLFJgvHuzd5DitoHDDwYnPjHNBQykad6bmm8UVNiHMApaFzwRFmfwv4jf8D1wyxLCxqJ3NeFvdrKM96EK1s8Fz4SwVESMxQpWo4nasYPVUUrFgoVuBUu4rNqakHFEiT6pEahF9n6S7eSwec4dESfNCw2CCaeaLCZt6QANk36fp8Mq87VSwquUJArSFCJmtDXswaHRqHjohyaSq2FP4P229wKdBDZwftoSu5WxzFs2usYPiGEiiH15EhZCVM2y9HxPd4k2wpQFsWxMF1sAerAAeFnaDnZHP318R6BGATgYLdjKsQX9BquSf8TJ6qd772CDxoRZaiKD573rPEcTe7nTiGdeHym1FSWKCFut8H3Wvrwx5waxJRda8u4CTdpd7pb2WCCswrG3e8vsxvNuEhCiNrwyHtQGnme2ZqJX7oCCE7wurVojusW3TDoGB74vn4YA7zfT6GLpL9teuJVHNRvy5h6ik5Jkf1m8PHz42Yt6Wtr82wAHtSkNEVHrwdDjADfz6zvFcZR9mq8Wd4DiVCy7ixhSB52skxhciDAySu9DmfwmJf1u8hEaC4Juh33VJ5CdZAfpFkQctACXxKNM2UsqBfumWSt2JmSYSECjFRcN61SMmXXoSiumNdF4bPQ32aBa75X6Y1KKMEmx2nDTW3uN18qrxcKkbybPawp8CtzS8LHreedk31woG2HVpk92TMv9n2YgWduv8NfQiyT2aF1TLYnf4ks8MdmdouqRQKLNVbDai531R6244oZyeKqmzdYe8MufT8Vo6cArKdk3s4bYxAX7auRD84TdPBy3MsQWZc4QVjR9YWuMvR7AxATyVYf4gq4YqEf3Y2nL31exqNXXz44iYnnAEALLbXDWpfPBv8M1zEMLvFhjviRuecvBPykYeeW9BgxX8Q1FDmKULBqNniwxDq5zFHEyAdmUKiuN5KzKZ6GmQAWy4V1zq2ASdegf5JPzUAhKzQ9rzfdfQ72rZ8asaamqRTW1ukDKXrnwJA8nALeoMzUQYNNbgNUtrQGCthzRipKbZeALE72bxtacs4Ht8KCgSCUWe12uyaBFBGVxo9ZuGBdrruK83y1WDDLuwSgkyXRWpmQ3nT92uPkwDARK9CbeNbimTBs8rxGYaExxxfETzpzfnovfgyhq2CNUky6g5m3BaG1AUpxoqzsR8vcNhCbf4UrzA3LjLpy7wrpUhG7PfXb11kwY8REhqeDUe7nzzYBouFzVW3VNkUdMykKhv3tGTL5k94bJUMYBXkzuppYGYyeKd6uBHrzJkNbW5ZLu7v8njQVSJuFReyz7GuHgivewomeygLepTok3GqMhYo94MnUf7tgYpz6A2bdE1BD2uxEHeM3wZjDBuWLnnudJ2rS6M1u2wvqPndUWKaCv82cXTHDTFnKjhXgBA4qTjJPb9J1YbW9mtKYmMEdLHj49tc"
)

func TestUnforgePublicKey(t *testing.T) {
	tests := []struct {
		name    string
		str     string
		want    string
		wantErr bool
	}{
		{
			name: "test 1",
			str:  "0103682c3aaa998fd9adfe8111cd42cc0daedb5d97647e6020eb629fbc91b613f721",
			want: "sppk7c3Fz7QqhZqY2FZUWWAnDuqTwx4KwDjgFA4VeLPiV8n4tnbsVzG",
		}, {
			name: "test 2",
			str:  "0028fc6875ca69a6f5bde4f377bfcde72fd618bcfa52e7272c7b788d1165449eb4",
			want: "edpktxGsKjnk43ZZ7v6gJe6PFV85peHvoWqVUzDQjTfN8idYwVkBwN",
		}, {
			name: "bls12_381",
			str:  "03a5b8c2d1e0f9a8b7c6d5e4f3a2b1c0d9e8f7a6b5c4d3e2f1a0b9c8d7e6f5a4b3c2d1e0f9a8b7c6d5e4f3a2b1c0d9e8f7",
			want: "BLpk1uhUas78WJUU87PBojYigrVGUSw9nwVgZ7e5MTWHopVmxATJ56oXKMqbkWQRZxvFjSp8rcEW",
		}, {
			name: "mldsa44",
			str:  mldsa44PublicKeyForged,
			want: mldsa44PublicKey,
		}, {
			name: "unknown tag is returned as hex",
			str:  "053fca7463fe261863908f801ca6ee71a6ddc629e1fef51541ec7cefee928d3229d6c13bc5d13edbde64d71b0b8f0f71",
			want: "053fca7463fe261863908f801ca6ee71a6ddc629e1fef51541ec7cefee928d3229d6c13bc5d13edbde64d71b0b8f0f71",
		}, {
			name: "known tag with unexpected length is returned as hex",
			str:  "0028fc6875ca69a6f5bde4f377bfcde72fd618bcfa52e7272c7b788d1165449e",
			want: "0028fc6875ca69a6f5bde4f377bfcde72fd618bcfa52e7272c7b788d1165449e",
		}, {
			name: "empty",
			str:  "",
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := UnforgePublicKey(tt.str)
			require.Equal(t, tt.wantErr, err != nil)
			if err != nil {
				return
			}
			require.Equal(t, tt.want, got)
		})
	}
}

func TestPublicKey(t *testing.T) {
	tests := []struct {
		name    string
		val     string
		want    string
		wantErr bool
	}{
		{
			name: "secp256k1",
			val:  "sppk7c3Fz7QqhZqY2FZUWWAnDuqTwx4KwDjgFA4VeLPiV8n4tnbsVzG",
			want: "0103682c3aaa998fd9adfe8111cd42cc0daedb5d97647e6020eb629fbc91b613f721",
		}, {
			name: "ed25519",
			val:  "edpktxGsKjnk43ZZ7v6gJe6PFV85peHvoWqVUzDQjTfN8idYwVkBwN",
			want: "0028fc6875ca69a6f5bde4f377bfcde72fd618bcfa52e7272c7b788d1165449eb4",
		}, {
			name: "bls12_381",
			val:  "BLpk1uhUas78WJUU87PBojYigrVGUSw9nwVgZ7e5MTWHopVmxATJ56oXKMqbkWQRZxvFjSp8rcEW",
			want: "03a5b8c2d1e0f9a8b7c6d5e4f3a2b1c0d9e8f7a6b5c4d3e2f1a0b9c8d7e6f5a4b3c2d1e0f9a8b7c6d5e4f3a2b1c0d9e8f7",
		}, {
			name: "mldsa44",
			val:  mldsa44PublicKey,
			want: mldsa44PublicKeyForged,
		}, {
			name:    "not a public key",
			val:     "tz1KfEsrtDaA1sX7vdM4qmEPWuSytuqCDp5j",
			wantErr: true,
		}, {
			name:    "too short",
			val:     "ed",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := PublicKey(tt.val)
			require.Equal(t, tt.wantErr, err != nil)
			if err != nil {
				return
			}
			require.Equal(t, tt.want, hex.EncodeToString(got))
		})
	}
}

func TestUnforgeContract(t *testing.T) {
	tests := []struct {
		name    string
		str     string
		want    string
		wantErr bool
	}{
		{
			name: "test 1",
			str:  "019ac6ee79c4e87a21d094bb8bf00f37fe51717e8700726563656976655f62756e6e795f62616c616e6365",
			want: "KT1Nh9wK8W3j3CXeTVm5DTTaiU5RE8CxLWZ4%receive_bunny_balance",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := UnforgeContract(tt.str)
			require.Equal(t, tt.wantErr, err != nil)
			if err != nil {
				return
			}
			require.Equal(t, tt.want, got)
		})
	}
}
