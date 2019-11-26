package stringutil

import "testing"

func TestFlowN(t *testing.T) {
	have := "`Take some more tea,' the March Hare said to Alice, very earnestly.\r\n\r\n`I've had nothing yet,' Alice replied in an offended tone, `so I can't take more.'\r\n\r\n`You mean you can't take LESS,' said the Hatter: `it's very easy to take MORE than nothing.'\r\n"
	want := "`Take some more tea,' the March Hare said to Alice, very \r\nearnestly.\r\n\r\n`I've had nothing yet,' Alice replied in an offended tone, `so \r\nI can't take more.'\r\n\r\n`You mean you can't take LESS,' said the Hatter: `it's very \r\neasy to take MORE than nothing.'\r\n"
	flowed := FlowN(have, false, 63)
	t.Logf("%q", flowed)
	t.Logf("\n%s", flowed)
	if flowed != want {
		t.Fatal("Flowed text output did not match expected result")
	}
}

//func TestDeflow(t *testing.T) {
//	have := ">>>Take some more tea.\r\n>>T've had nothing yet, so I can't take more.\r\n>You mean you can't take LESS, it's very easy to take MORE than nothing.\r\n"
//	want := "    Take some more tea.\r\n   T've had nothing yet, so I can't take more.\r\n  You mean you can't take LESS, it's very easy to take \r\n  MORE than nothing.\r\n"
//	flowed := FlowN(have, false, 58)
//	t.Logf("%q", flowed)
//	t.Logf("\n%s", flowed)
//	if flowed != want {
//		t.Fatal("Flowed text output did not match expected result")
//	}
//}
