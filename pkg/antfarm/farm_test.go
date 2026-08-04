package antfarm

import "testing"

func TestFarmAddLink(t *testing.T) {
	farm := NewFarm()
	farm.AddLink("start", "roomA")

	if len(farm.Links["start"]) != 1 || farm.Links["start"][0] != "roomA" {
		t.Errorf("expected link start -> roomA")
	}
	if len(farm.Links["roomA"]) != 1 || farm.Links["roomA"][0] != "start" {
		t.Errorf("expected link roomA -> start")
	}
}