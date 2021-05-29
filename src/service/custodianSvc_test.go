package service

import (
	"context"
	"testing"
)

func TestCustodianSvc_FetchFromCustodian_fetch1(t *testing.T) {

	svc := NewCustodianSvc("http://localhost:9999/custodian/")

	res, err := svc.FetchFromCustodian(context.Background(), 1)
	if err != nil {
		t.Error(err)
	}
	if len(res) != 1 {
		t.Error("len != 1")
	}

	if res[0].ID != 1 {
		t.Errorf("Wrong custodian: %d", res[0].ID)
	}
}

func TestCustodianSvc_FetchFromCustodian_fetch4(t *testing.T) {

	svc := NewCustodianSvc("http://localhost:9999/custodian/")

	custIds := []int32{1, 2, 3, 4}
	res, err := svc.FetchFromCustodian(context.Background(), custIds...)
	if err != nil {
		t.Error(err)
	}
	if len(res) != 4 {
		t.Error("len != 4")
	}
	for i, c := range res {
		if int(c.ID) != i+1 {
			t.Errorf("custodians in wrong order: got %d, expected %d", c.ID, i+1)
		}
	}
}

func TestCustodianSvc_FetchFromCustodian_wrong_host(t *testing.T) {

	svc := NewCustodianSvc("http://this.definitely.doesntexist.com:9999/custodian/")

	custIds := []int32{1, 2, 3, 4}
	_, err := svc.FetchFromCustodian(context.Background(), custIds...)
	if err == nil {
		t.Error("wrong host should return an error")
	}
}

func TestCustodianSvc_FetchFromCustodian_wrong_status_code(t *testing.T) {

	svc := NewCustodianSvc("http://localhost:9999/notfound/")

	custIds := []int32{1, 2, 3, 4}
	_, err := svc.FetchFromCustodian(context.Background(), custIds...)
	if err == nil {
		t.Error("wrong status_code should return an error")
	}
}
