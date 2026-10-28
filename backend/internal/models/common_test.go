package models

import "testing"

func TestNewPageParamsNormalizes(t *testing.T) {
	p := NewPageParams(0, 0)
	if p.Page != 1 || p.PerPage != defaultPerPage {
		t.Fatalf("unexpected defaults: %+v", p)
	}
	p = NewPageParams(3, 1000)
	if p.Page != 3 || p.PerPage != maxPerPage {
		t.Fatalf("per_page should be capped: %+v", p)
	}
	if p.Offset != (3-1)*maxPerPage {
		t.Fatalf("unexpected offset: %d", p.Offset)
	}
}

func TestNewPageComputesTotals(t *testing.T) {
	page := NewPage([]int{1, 2, 3}, 7, NewPageParams(1, 3))
	if page.Pagination.Total != 7 {
		t.Fatalf("total = %d", page.Pagination.Total)
	}
	if page.Pagination.TotalPages != 3 {
		t.Fatalf("total pages = %d", page.Pagination.TotalPages)
	}

	empty := NewPage[int](nil, 0, NewPageParams(1, 20))
	if empty.Items == nil {
		t.Fatal("items must be a non-nil empty slice for JSON output")
	}
}

func TestRoleValidity(t *testing.T) {
	if !RoleAdmin.Valid() {
		t.Fatal("admin should be valid")
	}
	if Role("nope").Valid() {
		t.Fatal("unknown role must be invalid")
	}
}

func TestOrgMemberPermissions(t *testing.T) {
	if !OrgRoleOwner.CanWrite() || !OrgRoleEditor.CanWrite() {
		t.Fatal("owner and editor should be able to write")
	}
	if OrgRoleViewer.CanWrite() {
		t.Fatal("viewer must not be able to write")
	}
	if !OrgRoleAdmin.CanManageMembers() || OrgRoleEditor.CanManageMembers() {
		t.Fatal("only owner/admin manage members")
	}
}
