package ast

import (
	"testing"
)

func TestAddressValidator(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{
			name:    "test 1",
			value:   "016e4943f7a23ab9cbe56f48ff72f6c27e8956762400",
			wantErr: false,
		}, {
			name:    "test 2",
			value:   "KT1JdufSdfg3WyxWJcCRNsBFV9V3x9TQBkJ2",
			wantErr: false,
		}, {
			name:    "test 3",
			value:   "tz1KfEsrtDaA1sX7vdM4qmEPWuSytuqCDp5j",
			wantErr: false,
		}, {
			name:    "test 4",
			value:   "tz1KfEsrtDaA1sX7vdM4qmEPWuSytuqCDp5",
			wantErr: true,
		}, {
			name:    "test 5",
			value:   "0x6e4943f7a23ab9cbe56f48ff72f6c27e8956762400",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := AddressValidator(tt.value); (err != nil) != tt.wantErr {
				t.Errorf("AddressValidator() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBakerHashValidator(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{
			name:    "test 1",
			value:   "94697e9229c88fac7d19d62e139ca6735f9569dd",
			wantErr: false,
		}, {
			name:    "test 2",
			value:   "SG1d1wsgMKvSstzZQ8L4WoskCesdWGzVt5k4",
			wantErr: false,
		}, {
			name:    "test 3",
			value:   "SG1d1wsgMKvSstzZQ8L4WoskCesdWGzVt5k",
			wantErr: true,
		}, {
			name:    "test 4",
			value:   "0x697e9229c88fac7d19d62e139ca6735f9569dd",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := BakerHashValidator(tt.value); (err != nil) != tt.wantErr {
				t.Errorf("BakerHashValidator() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPublicKeyValidator(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{
			name:    "test 1",
			value:   "sppk7bMuoa8w2LSKz3XEuPsKx1WavsMLCWgbWG9CZNAsJg9eTmkXRPd",
			wantErr: false,
		}, {
			name:    "test 2",
			value:   "030ed412d33412ab4b71df0aaba07df7ddd2a44eb55c87bf81868ba09a358bc0e0",
			wantErr: false,
		}, {
			name:    "test 3",
			value:   "0x0ed412d33412ab4b71df0aaba07df7ddd2a44eb55c87bf81868ba09a358bc0e0",
			wantErr: true,
		}, {
			name:    "test 4",
			value:   "spsk7bMuoa8w2LSKz3XEuPsKx1WavsMLCWgbWG9CZNAsJg9eTmkXRPd",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := PublicKeyValidator(tt.value); (err != nil) != tt.wantErr {
				t.Errorf("PublicKeyValidator() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBytesValidator(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{
			name:    "test 1",
			value:   "sppk7bMuoa8w2LSKz3XEuPsKx1WavsMLCWgbWG9CZNAsJg9eTmkXRPd",
			wantErr: true,
		}, {
			name:    "test 2",
			value:   "030ed412d33412ab4b71df0aaba07df7ddd2a44eb55c87bf81868ba09a358bc0e0",
			wantErr: false,
		}, {
			name:    "test 3",
			value:   "0x030ed412d33412ab4b71df0aaba07df7ddd2a44eb55c87bf81868ba09a358bc0e0",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := BytesValidator(tt.value); (err != nil) != tt.wantErr {
				t.Errorf("BytesValidator() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestChainIDValidator(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{
			name:    "test 1",
			value:   "7a06a770",
			wantErr: false,
		}, {
			name:    "test 2",
			value:   "NetXdQprcVkpaWU",
			wantErr: false,
		}, {
			name:    "test 3",
			value:   "0x06a770",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ChainIDValidator(tt.value); (err != nil) != tt.wantErr {
				t.Errorf("ChainIDValidator() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSignatureValidator(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{
			name:    "test 1",
			value:   "a04991b4e938cc42d6c01c42be3649a81a9f80d244d9b90e7ec4edf8e0a7b68b6c212da2fef076e48fed66802fa83442b960a36afdb3e60c3cf14d4010f41f03",
			wantErr: false,
		}, {
			name:    "test 2",
			value:   "sigixZejtj1GfDpyiWAQAmvbtnNmCXKyADqVvCaXJH9xHyhSnYYV8696Z3kkns5DNV7oMnMPfNzo3qm84DfEx1XG6saZmHiA",
			wantErr: false,
		}, {
			name:    "test 3",
			value:   "0x4991b4e938cc42d6c01c42be3649a81a9f80d244d9b90e7ec4edf8e0a7b68b6c212da2fef076e48fed66802fa83442b960a36afdb3e60c3cf14d4010f41f03",
			wantErr: true,
		}, {
			name:    "test 4",
			value:   "edsigthTzJ8X7MPmNeEwybRAvdxS1pupqcM5Mk4uCuyZAe7uEk68YpuGDeViW8wSXMrCi5CwoNgqs8V2w8ayB5dMJzrYCHhD8C7",
			wantErr: false,
		}, {
			name:    "test 5",
			value:   "spsig1PPUFZucuAQybs5wsqsNQ68QNgFaBnVKMFaoZZfi1BtNnuCAWnmL9wVy5HfHkR6AeodjVGxpBVVSYcJKyMURn6K1yknYLm",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := SignatureValidator(tt.value); (err != nil) != tt.wantErr {
				t.Errorf("SignatureValidator() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestContractValidator(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{
			name:    "test 1",
			value:   `KT1Nh9wK8W3j3CXeTVm5DTTaiU5RE8CxLWZ4%726563656976655f62756e6e795f62616c616e6365`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ContractValidator(tt.value); (err != nil) != tt.wantErr {
				t.Errorf("ContractValidator() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
