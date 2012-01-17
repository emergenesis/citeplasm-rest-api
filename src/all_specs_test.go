package main

import (
"gospec"
"testing"
)


func TestAllSpecs(t *testing.T) {
    r := gospec.NewRunner()
    r.AddSpec(MainSpec)
    gospec.MainGoTest(r, t)
}
