package item

import (
	"reflect"
	"testing"
)

func TestItem_ReleaseItem(t *testing.T) {
	newItem := func() Item {
		return Item{
			Sku:          Sku("TEST"),
			Name:         "Test",
			Price:        1000,
			QtyAvailable: 10,
			QtyReserved:  10,
		}
	}

	type want struct {
		qtyReserved int
	}

	type args struct {
		toReleaseQty int
	}
	tests := []struct {
		name    string
		args    args
		want    want
		wantErr bool
	}{
		{"Release 1", args{1}, want{9}, false},
		{"Release 10", args{10}, want{0}, false},
		{"Release 11 with error", args{11}, want{10}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := newItem()

			if err := i.ReleaseItem(tt.args.toReleaseQty); (err != nil) != tt.wantErr {
				t.Errorf("ReleaseItem() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !reflect.DeepEqual(i.QtyReserved, tt.want.qtyReserved) {
				t.Errorf("ReleaseItem() QtyReserved = %v, want %v", i.QtyReserved, tt.want.qtyReserved)
			}
		})
	}
}

func TestItem_RemoveItem(t *testing.T) {
	newItem := func() Item {
		return Item{
			Sku:          Sku("TEST"),
			Name:         "Test",
			Price:        1000,
			QtyAvailable: 10,
			QtyReserved:  5,
		}
	}
	type want struct {
		qtyReserved  int
		qtyAvailable int
	}
	type args struct {
		toRemoveQty int
	}
	tests := []struct {
		name    string
		args    args
		want    want
		wantErr bool
	}{
		{"Remove 1", args{1}, want{4, 9}, false},
		{"Remove 5", args{5}, want{0, 5}, false},
		{"Remove 6 with error", args{6}, want{5, 10}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := newItem()

			if err := i.RemoveItem(tt.args.toRemoveQty); (err != nil) != tt.wantErr {
				t.Errorf("RemoveItem() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !reflect.DeepEqual(i.QtyReserved, tt.want.qtyReserved) {
				t.Errorf("RemoveItem() QtyReserved = %v, want %v", i.QtyReserved, tt.want.qtyReserved)
			}

			if !reflect.DeepEqual(i.QtyAvailable, tt.want.qtyAvailable) {
				t.Errorf("RemoveItem() QtyAvailable = %v, want %v", i.QtyAvailable, tt.want.qtyAvailable)
			}

		})
	}
}

func TestItem_ReserveItem(t *testing.T) {
	newItem := func() Item {
		return Item{
			Sku:          Sku("TEST"),
			Name:         "Test",
			Price:        1000,
			QtyAvailable: 10,
			QtyReserved:  0,
		}
	}

	type want struct {
		qtyReserved  int

	}
	type args struct {
		toReserveQty int
	}
	tests := []struct {
		name    string
		args    args
		want    want
		wantErr bool
	}{
		{"Reserve 1", args{1}, want{1}, false},
		{"Reserve 10", args{10}, want{10}, false},
		{"Reserve 11 with error", args{11}, want{0}, true},
	}
		for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := newItem()
			if err := i.ReserveItem(tt.args.toReserveQty); (err != nil) != tt.wantErr {
				t.Errorf("ReserveItem() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !reflect.DeepEqual(i.QtyReserved, tt.want.qtyReserved) {
				t.Errorf("ReserveItem() QtyReserved = %v, want %v", i.QtyReserved, tt.want.qtyReserved)
			}
		})
	}
}
