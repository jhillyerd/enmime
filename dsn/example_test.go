package dsn_test

import (
	"fmt"
	"os"

	"github.com/jhillyerd/enmime"
	"github.com/jhillyerd/enmime/dsn"
)

// ExampleParseReport shows how to parse message as Delivery Status Notification (DSN).
func ExampleParseReport() {
	f, err := os.Open("testdata/simple_dsn.raw")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	env, err := enmime.ReadEnvelope(f)
	if err != nil {
		fmt.Print(err)
		return
	}

	rep, err := dsn.ParseReport(env.Root)
	if err != nil {
		fmt.Print(err)
		return
	}

	fmt.Printf("Original message: %s", rep.OriginalMessage)
	fmt.Printf("Failed?: %t\n", dsn.IsFailed(rep.DeliveryStatus.RecipientDSNs[0]))
	fmt.Printf("Why?: %s", rep.Explanation.Text)
	// Output:
	// Original message: [original message goes here]
	// Failed?: true
	// Why?: [human-readable explanation goes here]
}
