package main

import (
    "gospec"
    "testing"
)


func TestAllSpecs(t *testing.T) {
    r := gospec.NewRunner()
    r.AddSpec(MainSpec)
    FlushDb()
    LoadFixtures()
    gospec.MainGoTest(r, t)
    FlushDb()
}
