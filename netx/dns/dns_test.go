package dns_test

import (
	"fmt"
	"testing"

	"github.com/cocktail828/go-tools/netx/dns"
	"github.com/stretchr/testify/assert"
)

func TestDNS_RRSet(t *testing.T) {
	f := func(r *dns.RRSet) []string {
		r.Normalize(6666)
		result := []string{}
		for idx := range r.Records {
			result = append(result, fmt.Sprintf("%v#%v", r.Records[idx].Target, r.Records[idx].Port))
		}
		return result
	}

	{
		r, err := dns.NewRRSet("10.1.87.70,10.1.87.70:3307")
		assert.Equal(t, nil, err)
		assert.ElementsMatch(t, []string{"10.1.87.70#6666", "10.1.87.70#3307"}, f(r))
	}
	{
		r, err := dns.NewRRSet("dns://www.aisaas.net")
		assert.Equal(t, nil, err)
		assert.ElementsMatch(t, []string{"10.1.87.81#6666", "10.1.87.20#6666"}, f(r))
	}
	{
		r, err := dns.NewRRSet("dns+srv://_mysql-ost._tcp.xx.aisaas.net")
		assert.Equal(t, nil, err)
		assert.ElementsMatch(t, []string{"a.aisaas.net.#3306", "b.aisaas.net.#3306"}, f(r))
	}
	{
		_, err := dns.NewRRSet("dns://sqzdapiaaaa.te.rdg.local.nonexist")
		if err == nil {
			panic("invalid, should err here")
		}
	}
}
