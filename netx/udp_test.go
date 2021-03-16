package netx

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/advancevillage/3rd/utils"
	"github.com/stretchr/testify/assert"
)

var udpTestData = map[string]struct {
	host    string
	udpPort int
	pkgLen  int
	f       Handler
}{}

//go test -v -count=1 -timeout 20m -test.run Test_udp_server ./netx/
func Test_udp_server(t *testing.T) {
	var primes = []int{
		2, 3, 5, 7, 11, 13, 17, 19, 23, 29,
		31, 37, 41, 43, 47, 53, 59, 61, 67, 71,
		73, 79, 83, 89, 97, 101, 103, 107, 109, 113,
		127, 131, 137, 139, 149, 151, 157, 163, 167, 173,
		179, 181, 191, 193, 197, 199, 211, 223, 227, 229,
		233, 239, 241, 251, 257, 263, 269, 271, 277, 281,
		283, 293, 307, 311, 313, 317, 331, 337, 347, 349,
		353, 359, 367, 373, 379, 383, 389, 397, 401, 409,
		419, 421, 431, 433, 439, 443, 449, 457, 461, 463,
		467, 479, 487, 491, 499, 503, 509, 521, 523, 541,
		547, 557, 563, 569, 571, 577, 587, 593, 599, 601,
		607, 613, 617, 619, 631, 641, 643, 647, 653, 659,
		661, 673, 677, 683, 691, 701, 709, 719, 727, 733,
		739, 743, 751, 757, 761, 769, 773, 787, 797, 809,
		811, 821, 823, 827, 829, 839, 853, 857, 859, 863,
		877, 881, 883, 887, 907, 911, 919, 929, 937, 941,
		947, 953, 967, 971, 977, 983, 991, 997, 1009, 1013,
		1019, 1021, 1031, 1033, 1039, 1049, 1051, 1061, 1063, 1069,
		1087, 1091, 1093, 1097, 1103, 1109, 1117, 1123, 1129, 1151,
		1153, 1163, 1171, 1181, 1187, 1193, 1201, 1213, 1217, 1223,
		1229, 1231, 1237, 1249, 1259, 1277, 1279, 1283, 1289, 1291,
		1297, 1301, 1303, 1307, 1319, 1321, 1327, 1361, 1367, 1373,
		1381, 1399, 1409, 1423, 1427, 1429, 1433, 1439, 1447, 1451,
		4507, 4513, 4517, 4519, 4523, 4547, 4549, 4561, 4567, 4583,
		4591, 4597, 4603, 4621, 4637, 4639, 4643, 4649, 4651, 4657,
		4663, 4673, 4679, 4691, 4703, 4721, 4723, 4729, 4733, 4751,
		4759, 4783, 4787, 4789, 4793, 4799, 4801, 4813, 4817, 4831,
		4861, 4871, 4877, 4889, 4903, 4909, 4919, 4931, 4933, 4937,
		4943, 4951, 4957, 4967, 4969, 4973, 4987, 4993, 4999, 5003,
		5009, 5011, 5021, 5023, 5039, 5051, 5059, 5077, 5081, 5087,
		5099, 5101, 5107, 5113, 5119, 5147, 5153, 5167, 5171, 5179,
		11159, 11161, 11171, 11173, 11177, 11197, 11213, 11239, 11243, 11251,
		11257, 11261, 11273, 11279, 11287, 11299, 11311, 11317, 11321, 11329,
		11351, 11353, 11369, 11383, 11393, 11399, 11411, 11423, 11437, 11443,
		11447, 11467, 11471, 11483, 11489, 11491, 11497, 11503, 11519, 11527,
		11549, 11551, 11579, 11587, 11593, 11597, 11617, 11621, 11633, 11657,
		11677, 11681, 11689, 11699, 11701, 11717, 11719, 11731, 11743, 11777,
		11779, 11783, 11789, 11801, 11807, 11813, 11821, 11827, 11831, 11833,
		11839, 11863, 11867, 11887, 11897, 11903, 11909, 11923, 11927, 11933,
	}
	//初始化数据
	for k, v := range primes {
		udpTestData[fmt.Sprintf("case%d", k)] = struct {
			host    string
			udpPort int
			pkgLen  int
			f       Handler
		}{
			host:    "127.0.0.1",
			udpPort: 1994,
			pkgLen:  v,
			f: func(ctx context.Context, in []byte) []byte {
				return in
			},
		}
	}

	for n, p := range udpTestData {
		f := func(t *testing.T) {
			var scfg = &ServerOption{}
			scfg.Host = p.host
			scfg.UdpPort = p.udpPort
			scfg.MaxSize = uint32(p.pkgLen * 2)
			var udps, err = NewUDPServer(scfg, p.f)
			if err != nil {
				t.Fatal(err)
				return
			}
			var ccfg = &ClientOption{}
			ccfg.Host = p.host
			ccfg.UdpPort = p.udpPort
			ccfg.MaxSize = uint32(p.pkgLen * 2)
			var udpc IUDPClient
			udpc, err = NewUdpClient(ccfg)
			if err != nil {
				t.Fatal(err)
				return
			}
			//2. 启动服务
			go udps.StartServer()
			time.Sleep(5 * time.Second)
			var im = make(map[string]struct{})
			var rm = make(map[string]struct{})
			go func() {
				for len(im) < p.pkgLen {
					var msg = utils.RandsString(p.pkgLen)
					err := udpc.Send(context.TODO(), []byte(msg))
					if err != nil {
						return
					}
					im[msg] = struct{}{}
				}
			}()
			go func() {
				for len(rm) < p.pkgLen {
					buf, err := udpc.Receive(context.TODO())
					if err != nil {
						return
					}
					rm[string(buf)] = struct{}{}
				}
			}()
			//控制部分
			var timeout = time.NewTicker(time.Second * 5)
			var sig = make(chan os.Signal)
			signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
			select {
			case <-timeout.C:
			case <-sig:
			}
			for k := range im {
				delete(im, k)
				delete(rm, k)
			}
			assert.Equal(t, len(im), len(rm))
		}
		t.Run(n, f)
	}
}
