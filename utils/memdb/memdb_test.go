package memdb

import (
	"fmt"
	"github.com/gambarini/flip-shop/utils"
	"reflect"
	"sync"
	"testing"
)

func TestMemoryKVDatabase_Read(t *testing.T) {

	mDb := &MemoryKVDatabase{
		lock: sync.RWMutex{},
		tx: &MemoryKVTx{
			data: map[utils.StoreName]map[string]interface{}{utils.StoreName("TEST"): {"key1": 1, "key2": 2, "key3": 3}},
		},
	}

	type args struct {
		name utils.StoreName
		key  string
	}
	tests := []struct {
		name    string
		args    args
		wantV   interface{}
		wantErr bool
	}{
		{"read key 1", args{utils.StoreName("TEST"), "key1"}, 1, false},
		{"read key 2", args{utils.StoreName("TEST"), "key2"}, 2, false},
		{"read key 3", args{utils.StoreName("TEST"), "key3"}, 3, false},
		{"read key 4 with error", args{utils.StoreName("TEST"), "key4"}, nil, true},
		{"read store with error", args{utils.StoreName("ERROR"), "key1"}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			gotV, err := mDb.Read(tt.args.name, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("Read() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotV, tt.wantV) {
				t.Errorf("Read() gotV = %v, want %v", gotV, tt.wantV)
			}
		})
	}
}

func TestMemoryKVDatabase_WithTx(t *testing.T) {
	mDb := &MemoryKVDatabase{
		lock: sync.RWMutex{},
		tx: &MemoryKVTx{
			data: map[utils.StoreName]map[string]interface{}{utils.StoreName("TEST"): {"key1": 1, "key2": 2, "key3": 3}},
		},
	}

	type args struct {
		txHandler utils.TxHandler
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"tx ok", args{func(tx utils.Tx) error {
			return nil
		}}, false},
		{"tx with error", args{func(tx utils.Tx) error {
			return fmt.Errorf("error")
		}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := mDb.WithTx(tt.args.txHandler); (err != nil) != tt.wantErr {
				t.Errorf("WithTx() error = %v, wantErr %v", err, tt.wantErr)
			}

		})
	}
}

func TestMemoryKVTx_Read(t *testing.T) {
	tx := &MemoryKVTx{
		data: map[utils.StoreName]map[string]interface{}{utils.StoreName("TEST"): {"key1": 1, "key2": 2, "key3": 3}},
	}

	type args struct {
		name utils.StoreName
		key  string
	}

	tests := []struct {
		name    string
		args    args
		wantV   interface{}
		wantErr bool
	}{
		{"read k1", args{utils.StoreName("TEST"), "key1"}, 1, false},
		{"read k2", args{utils.StoreName("TEST"), "key2"}, 2, false},
		{"read k3", args{utils.StoreName("TEST"), "key3"}, 3, false},
		{"read k4 with error", args{utils.StoreName("TEST"), "key4"}, nil, true},
		{"read store with error", args{utils.StoreName("ERROR"), "key1"}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			gotV, err := tx.Read(tt.args.name, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("Read() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotV, tt.wantV) {
				t.Errorf("Read() gotV = %v, want %v", gotV, tt.wantV)
			}
		})
	}
}

func TestMemoryKVTx_Write(t *testing.T) {
	tx := &MemoryKVTx{
		data: map[utils.StoreName]map[string]interface{}{utils.StoreName("TEST"): {}},
	}

	type args struct {
		name utils.StoreName
		key  string
		v    interface{}
	}
	tests := []struct {
		name  string
		args  args
		wantV interface{}
	}{
		{"write k1", args{utils.StoreName("TEST"), "key1", 1}, 1},
		{"write k2", args{utils.StoreName("TEST"), "key2", 2}, 2},
		{"write k3", args{utils.StoreName("TEST"), "key3", 3}, 3},
		{"write new store", args{utils.StoreName("NEW"), "key1", 1}, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			tx.Write(tt.args.name, tt.args.key, tt.args.v)

			if !reflect.DeepEqual(tx.data[tt.args.name][tt.args.key], tt.wantV) {
				t.Errorf("data stored = %v, want %v", tx.data[tt.args.name][tt.args.key], tt.wantV)
			}
		})
	}
}
