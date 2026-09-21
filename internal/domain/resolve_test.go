package domain

import "testing"

func TestResolveIssueAndSpace(t *testing.T) {
	s, err := Resolve(ResolveInput{Issue: "SDO-1"})
	if err != nil || s.Hostname != "sesamidevel.atlassian.net" || s.Alias != "sesamidevel" {
		t.Fatalf("%+v %v", s, err)
	}
	s, err = Resolve(ResolveInput{Issue: "sdp-99"})
	if err != nil || s.Alias != "sesamidevel" {
		t.Fatalf("%+v %v", s, err)
	}
	s, err = Resolve(ResolveInput{Issue: "CAB-1"})
	if err != nil || s.Hostname != "sesami-io.atlassian.net" || s.Alias != "sesami-io" {
		t.Fatalf("%+v %v", s, err)
	}
	s, err = Resolve(ResolveInput{Space: "CCAB"})
	if err != nil || s.Hostname != "sesami-io.atlassian.net" {
		t.Fatalf("%+v %v", s, err)
	}
	s, err = Resolve(ResolveInput{Space: "ccab"})
	if err != nil || s.Alias != "sesami-io" {
		t.Fatalf("%+v %v", s, err)
	}
}

func TestResolveExplicitSite(t *testing.T) {
	s, err := Resolve(ResolveInput{Site: "garda"})
	if err != nil || s.Role != "jsm_customer" || s.Hostname != "gardaworld.atlassian.net" {
		t.Fatalf("%+v %v", s, err)
	}
	s, err = Resolve(ResolveInput{Site: "dd528466-a332-4f63-9cee-0033fe01f875"})
	if err != nil || s.Alias != "sesamidevel" {
		t.Fatalf("%+v %v", s, err)
	}
}

func TestResolveJQLSameCloudAndCrossCloud(t *testing.T) {
	s, err := Resolve(ResolveInput{JQL: "project = CAB"})
	if err != nil || s.Alias != "sesami-io" {
		t.Fatalf("%+v %v", s, err)
	}
	s, err = Resolve(ResolveInput{JQL: `project in (SDO, SDP, SES) AND statusCategory != Done`})
	if err != nil || s.Alias != "sesamidevel" {
		t.Fatalf("%+v %v", s, err)
	}
	_, err = Resolve(ResolveInput{JQL: "project in (SDO, CAB)"})
	if ExitOf(err) != ExitUsage {
		t.Fatal(err)
	}
	_, err = Resolve(ResolveInput{JQL: `project = SDO OR project = CAB`})
	if ExitOf(err) != ExitUsage {
		t.Fatal(err)
	}
}

func TestResolveCannotInfer(t *testing.T) {
	_, err := Resolve(ResolveInput{})
	if ExitOf(err) != ExitUsage {
		t.Fatal(err)
	}
	_, err = Resolve(ResolveInput{Issue: "10001"})
	if ExitOf(err) != ExitUsage {
		t.Fatal(err)
	}
	_, err = Resolve(ResolveInput{JQL: `statusCategory != Done ORDER BY updated DESC`})
	if ExitOf(err) != ExitUsage {
		t.Fatal(err)
	}
	_, err = Resolve(ResolveInput{Issue: "SDO-1", Site: "sesami-io"})
	if ExitOf(err) != ExitUsage {
		t.Fatal(err)
	}
}
