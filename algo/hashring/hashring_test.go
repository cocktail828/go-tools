package hashring

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	node1 = "192.168.1.1"
	node2 = "192.168.1.2"
	node3 = "192.168.1.3"
)

func getNodesCount(nodes nodesArray) (int, int, int) {
	node1Count := 0
	node2Count := 0
	node3Count := 0

	for _, node := range nodes {
		if node.nodeKey == node1 {
			node1Count += 1
		}
		if node.nodeKey == node2 {
			node2Count += 1

		}
		if node.nodeKey == node3 {
			node3Count += 1

		}
	}
	return node1Count, node2Count, node3Count
}

func TestHashRing(t *testing.T) {
	nodeWeight := make(map[string]int)
	nodeWeight[node1] = 2
	nodeWeight[node2] = 2
	nodeWeight[node3] = 3

	hring := New()
	hring.AddNodes(nodeWeight)
	_, _, c3 := getNodesCount(hring.nodes)

	func() {
		if hring.GetNode("1") != node3 {
			t.Fatalf("expetcd %v got %v", node3, hring.GetNode("1"))
		}
		if hring.GetNode("2") != node3 {
			t.Fatalf("expetcd %v got %v", node3, hring.GetNode("2"))
		}
		if hring.GetNode("3") != node2 {
			t.Fatalf("expetcd %v got %v", node2, hring.GetNode("3"))
		}
	}()

	func() {
		hring.RemoveNode(node3)
		if hring.GetNode("1") != node1 {
			t.Fatalf("expetcd %v got %v", node1, hring.GetNode("1"))
		}
		if hring.GetNode("2") != node2 {
			t.Fatalf("expetcd %v got %v", node1, hring.GetNode("2"))
		}
		if hring.GetNode("3") != node2 {
			t.Fatalf("expetcd %v got %v", node2, hring.GetNode("3"))
		}
		_, _, _c3 := getNodesCount(hring.nodes)
		assert.Equal(t, 0, _c3)
	}()

	func() {
		hring.AddNode(node3, 3)
		if hring.GetNode("1") != node3 {
			t.Fatalf("expetcd %v got %v", node3, hring.GetNode("1"))
		}
		if hring.GetNode("2") != node3 {
			t.Fatalf("expetcd %v got %v", node3, hring.GetNode("2"))
		}
		if hring.GetNode("3") != node2 {
			t.Fatalf("expetcd %v got %v", node2, hring.GetNode("3"))
		}
		_, _, _c3 := getNodesCount(hring.nodes)
		assert.Equal(t, c3, _c3)
	}()
}

var (
	patchids = []string{"1866652039436464128", "1866447602558332928", "1866405925902774272", "1866106275048226816", "1866351400563142656", "1866488424989491200", "1866426284274384896", "1866504890283356160", "1866485555053621248", "1866441508553388032", "2434459338076160   ", "1866481704384753664", "1866486266600517632", "1866419898568179712", "1866425953503182848", "1866347844611375104", "1866498944542998528", "1866509022713118720", "1866333031978008576", "1866489952865386496", "1866457023279689728", "1866367485429710848", "1866439333274222592", "1866488636449521664", "1866521268642017280", "1866434897877434368", "1866459718409416704", "1866475433862524928", "1866490058800922624", "1866464241203052544", "1866377282120089600", "1866386485739941888", "1866492806854242304", "1866496622853128192", "1866512539750133760", "2434558723453954   ", "1866407200673394688", "1866488829144100864", "1866455367871922176", "1866495332848660480", "1866490260454670336", "1866516094674567168", "1866514390713266176", "1866458937065242624", "1866449146397884416", "1866476624537350144", "1866505341036818432", "1866483339899072512", "1866483656682270720", "1866498074711322624", "1866393029600903168", "1866366176945274880", "1866436969561620480", "1866320060207755264", "1865421816799850496", "1866130441981816832", "1866145231064166400", "1866383951361245184", "1866519955556564992", "1866359669100802048", "1866433963793350656", "1866473094527217664", "1866394625319530496", "1866481704384753664", "1866430001627299840", "1866465048790986752", "1866483098449768448", "1866435618488221696", "1866505196413157376", "1866396914176524288", "1866489251959304192", "1866459957165977600", "1866430084317868032", "1866471186676740096", "1866408053635444736", "1866342692072488960", "1866452316616425472", "1866357730694819840", "1866510963375636480", "1866336083971702784", "1866385505510260736", "1866480176945201152", "1866345453656768512", "2431323291016192   ", "2434400061912069   ", "2434408704275456   ", "2434506530577408   ", "1866374231275499520", "1866476624537350144", "1866403778452680704", "1866400140007206912", "1866473896012574720", "1866424877706149888", "1866431613431083008", "1865765233685852160", "1866468205164265472", "1866453081179189248", "1866458036782919680", "1866415237156990976", "1866363210213392384", "1866506904388272128", "2434514108798977   ", "1866427276466221056", "1866393269741584384", "1866332493638955008", "1866495887012687872", "1866641970925961216", "1866507403363512320", "1866324070415499264", "1866459363634212864", "1866485373725605888", "1866482284654264320", "1866480065603207168", "2434528676665345   ", "1866381504781778944", "1866434466426159104", "1866511194854944768", "1866101505067614208", "1866497179357577216", "1866472982900011008", "1866391060467449856", "1866495413861777408", "1866344565890383872", "1866350592803241984", "1866362717676273664", "1866401897512857600", "1866358477025079296", "1866373544164618240", "1866432089056903168", "1866494312043147264", "1866515823760142336", "1866458256077910016", "1866380354128867328", "1866112502478041088", "1866425136708616192", "1866478111170789376", "1866467041408348160", "1866476143606005760", "1866499930556624896", "1866367176871542784", "1866493329254940672", "1866466806669930496", "1866405592367661056", "1866343029831401472", "1866407473060020224", "1866491140826361856", "1866444884938747904", "1866318692466327552", "1866406480058413056", "1866306987027890176", "2434355481256960   ", "1866351151715221504", "1866475230833045504", "1866441765609697280", "1866489066680119296", "1866496968967094272", "1866411668286767104", "1866504319153369088", "1866487264232636416", "1866381253056430080", "1865774105083408384", "1866641507547643904", "1866503125769158656", "1866639767884562432", "1866401902315470848", "1866341427133317120", "1866492114433507328", "1866476540223586304", "1866348338553450496", "1866332023604412416", "1866444602976661504", "1866341700929093632", "1866401730604859392", "1866385786625097728", "1866507793991761920", "1866325800519626752", "1866477307609116672", "1866468878098264064", "2434442669912065   ", "2434443580076033   ", "1866336056842809344", "1866491966231969792", "1866450292260765696", "1866487086188625920", "1866447528055046144", "1866349013907832832", "1866417946530574336", "1866415667299778560", "2434437584318467   ", "1866497007382724608", "1866410288792764416", "1866352655645835264", "1866400779688898560", "1866406411817222144", "1866370814863896576", "1866414852052774912", "1866471588050661376", "1866494552171380736", "1866332980153053184", "1866398886061441024", "1866468005553143808", "1866119783722811392", "1866378512586141696", "1866438585975078912", "1866438706016059392", "1866490328360452096", "2434507625453569   ", "1866447260626087936", "1866474918055542784", "1866486940394483712", "1866156490635505664", "1866397044875231232", "1866492652000538624", "1866418232804405248", "1866488266532745216", "1866429637435748352", "1866447583017205760", "1866347442335543296", "1866385097039441920", "1866326766065184768", "1866361250093498368", "1866392433237983232", "1866394421602185216", "1866362288338792448", "1866475018924359680", "1866375698589974528", "1866381554404589568", "1866343303874641920", "1866044248778899456", "1866370500077187072", "1866376359708618752", "1866407140615290880", "1866368829339238400", "1866367495886110720", "1866350094972911616", "1866382330065616896", "1866493225076817920", "1866455151684911104", "1865748728571265024", "1866393111813320704", "1866392466494484480", "1866423798910054400", "1866325279092277248", "1866403150401929216", "1866346485124202496", "2434425767758849   ", "1866367254667493376", "1866326250694275072", "1866440422237040640", "1866377218609803264", "1866391444430815232", "1866369443041275904", "1866378209212268544", "1866363813903761408"}
)

func TestHashRing_Func(t *testing.T) {
	f := func(name string, fn HashFunc) {
		t.Run(name, func(t *testing.T) {
			hring := New(WithHash(fn))
			// 添加节点实例
			hring.AddNodes(map[string]int{
				"xspark13b6k1": 1,
				"xspark13b6k2": 1,
				"xspark13b6k3": 1,
			})
			n1, n2, n3 := 0, 0, 0
			for _, patchid := range patchids {
				switch hring.GetNode(patchid) {
				case "xspark13b6k1":
					n1++
				case "xspark13b6k2":
					n2++
				case "xspark13b6k3":
					n3++
				}
			}
			fmt.Println(name, n1, n2, n3)
		})
	}

	f("sha256", Sha256)
	f("md5", Md5)
	f("crc32", Crc32)
}

func TestHashRingX(t *testing.T) {
	// 新建 hash 实例
	hring := New(WithHash(Sha256), WithSpots(100))
	// 添加节点实例
	hring.AddNodes(map[string]int{
		"xspark13b6k1": 1,
		"xspark13b6k2": 1,
		"xspark13b6k3": 1,
	})

	v := hring.GetNode(patchids[0])
	for i := 0; i < 10000; i++ {
		if v != hring.GetNode(patchids[0]) {
			panic("not equal")
		}
	}

	n1, n2, n3 := 0, 0, 0
	for _, patchid := range patchids {
		switch hring.GetNode(patchid) {
		case "xspark13b6k1":
			n1++
		case "xspark13b6k2":
			n2++
		case "xspark13b6k3":
			n3++
		}
	}
	fmt.Println(n1, n2, n3)
}
