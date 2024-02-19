package awslb

import "testing"

func TestScanAWSLB(t *testing.T) {
	result, err := ScanAWSLB()
	if err != nil {
		t.Errorf("Result was incorrect, got: %s, want: %s.", result, "Foo")
	}

}
