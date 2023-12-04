## windows
```
goos: windows
goarch: amd64
pkg: github.com/246859/river
cpu: 11th Gen Intel(R) Core(TM) i7-11800H @ 2.30GHz
BenchmarkDB_B
BenchmarkDB_B/benchmarkDB_Put_value_16B
BenchmarkDB_B/benchmarkDB_Put_value_16B-4                 112986              9758 ns/op            5548 B/op         38 allocs/op
BenchmarkDB_B/benchmarkDB_Put_value_64B
BenchmarkDB_B/benchmarkDB_Put_value_64B-4                 113967             10474 ns/op            5625 B/op         38 allocs/op
BenchmarkDB_B/benchmarkDB_Put_value_128B
BenchmarkDB_B/benchmarkDB_Put_value_128B-4                104737             11093 ns/op            5690 B/op         38 allocs/op
BenchmarkDB_B/benchmarkDB_Put_value_256B
BenchmarkDB_B/benchmarkDB_Put_value_256B-4                 98302             12175 ns/op            5846 B/op         38 allocs/op
BenchmarkDB_B/benchmarkDB_Put_value_512B
BenchmarkDB_B/benchmarkDB_Put_value_512B-4                 85974             14017 ns/op            6242 B/op         39 allocs/op
BenchmarkDB_B/benchmarkDB_Get_B
BenchmarkDB_B/benchmarkDB_Get_B-4                         452368              2680 ns/op            1817 B/op          5 allocs/op
BenchmarkDB_B/benchmarkDB_Del_B
BenchmarkDB_B/benchmarkDB_Del_B-4                         599091              1740 ns/op             279 B/op          7 allocs/op
BenchmarkDB_KB
BenchmarkDB_KB/benchmarkDB_Put_value_1KB
BenchmarkDB_KB/benchmarkDB_Put_value_1KB-4                 66819             18538 ns/op            6701 B/op         38 allocs/op
BenchmarkDB_KB/benchmarkDB_Put_value_16KB
BenchmarkDB_KB/benchmarkDB_Put_value_16KB-4                10000            124830 ns/op           27682 B/op         40 allocs/op
BenchmarkDB_KB/benchmarkDB_Put_value_64KB
BenchmarkDB_KB/benchmarkDB_Put_value_64KB-4                 2726            440158 ns/op          131245 B/op         43 allocs/op
BenchmarkDB_KB/benchmarkDB_Put_value_256KB
BenchmarkDB_KB/benchmarkDB_Put_value_256KB-4                 660           1773743 ns/op          716746 B/op         49 allocs/op
BenchmarkDB_KB/benchmarkDB_Put_value_512KB
BenchmarkDB_KB/benchmarkDB_Put_value_512KB-4                 386           3314863 ns/op         1583756 B/op         56 allocs/op
BenchmarkDB_KB/benchmarkDB_Get_KB
BenchmarkDB_KB/benchmarkDB_Get_KB-4                        64108             16429 ns/op           25819 B/op          6 allocs/op
BenchmarkDB_KB/benchmarkDB_Del_KB
BenchmarkDB_KB/benchmarkDB_Del_KB-4                      1103665              1060 ns/op             237 B/op          6 allocs/op
BenchmarkDB_MB
BenchmarkDB_MB/benchmarkDB_Put_value_1MB
BenchmarkDB_MB/benchmarkDB_Put_value_1MB-4                   189           6748591 ns/op         1608475 B/op         59 allocs/op
BenchmarkDB_MB/benchmarkDB_Put_value_8MB
BenchmarkDB_MB/benchmarkDB_Put_value_8MB-4                    36          35199903 ns/op        17025122 B/op        173 allocs/op
BenchmarkDB_MB/benchmarkDB_Put_value_16MB
BenchmarkDB_MB/benchmarkDB_Put_value_16MB-4                   28          73703893 ns/op        37587965 B/op        321 allocs/op
BenchmarkDB_MB/benchmarkDB_Put_value_32MB
BenchmarkDB_MB/benchmarkDB_Put_value_32MB-4                    7         148933414 ns/op        122401424 B/op       593 allocs/op
BenchmarkDB_MB/benchmarkDB_Put_value_64MB
BenchmarkDB_MB/benchmarkDB_Put_value_64MB-4                    9         220238289 ns/op        114763863 B/op       863 allocs/op
BenchmarkDB_MB/benchmarkDB_Get_MB
BenchmarkDB_MB/benchmarkDB_Get_MB-4                        16188             66728 ns/op          139405 B/op          6 allocs/op
BenchmarkDB_MB/benchmarkDB_Del_MB
BenchmarkDB_MB/benchmarkDB_Del_MB-4                      1538329               780.8 ns/op           234 B/op          6 allocs/op
BenchmarkDB_Mixed
BenchmarkDB_Mixed/benchmarkDB_Put_Mixed_64B
BenchmarkDB_Mixed/benchmarkDB_Put_Mixed_64B-4             118870             10416 ns/op            5599 B/op         38 allocs/op
BenchmarkDB_Mixed/benchmarkDB_Put_Mixed_128KB
BenchmarkDB_Mixed/benchmarkDB_Put_Mixed_128KB-4             1275            862589 ns/op          292770 B/op         45 allocs/op
BenchmarkDB_Mixed/benchmarkDB_Put_Mixed_MB
BenchmarkDB_Mixed/benchmarkDB_Put_Mixed_MB-4                 190           6180057 ns/op         3617123 B/op         68 allocs/op
BenchmarkDB_Mixed/benchmarkDB_Get_Mixed
BenchmarkDB_Mixed/benchmarkDB_Get_Mixed-4                  51631             20758 ns/op           34963 B/op          6 allocs/op
BenchmarkDB_Mixed/benchmarkDB_Del_Mixed
BenchmarkDB_Mixed/benchmarkDB_Del_Mixed-4                 940024              1262 ns/op             240 B/op          7 allocs/op
PASS
ok      github.com/246859/river 46.353s
```

## linux
```
goos: linux
goarch: amd64
pkg: github.com/246859/river
cpu: 11th Gen Intel(R) Core(TM) i7-11800H @ 2.30GHz
BenchmarkDB_B
BenchmarkDB_B/benchmarkDB_Put_value_16B
BenchmarkDB_B/benchmarkDB_Put_value_16B-4                  83499             15714 ns/op            5557 B/op         38 allocs/op
BenchmarkDB_B/benchmarkDB_Put_value_64B
BenchmarkDB_B/benchmarkDB_Put_value_64B-4                  72183             14642 ns/op            5610 B/op         38 allocs/op
BenchmarkDB_B/benchmarkDB_Put_value_128B
BenchmarkDB_B/benchmarkDB_Put_value_128B-4                 78613             14967 ns/op            5695 B/op         38 allocs/op
BenchmarkDB_B/benchmarkDB_Put_value_256B
BenchmarkDB_B/benchmarkDB_Put_value_256B-4                 76927             16786 ns/op            5844 B/op         38 allocs/op
BenchmarkDB_B/benchmarkDB_Put_value_512B
BenchmarkDB_B/benchmarkDB_Put_value_512B-4                 52198             21617 ns/op            6233 B/op         38 allocs/op
BenchmarkDB_B/benchmarkDB_Get_B
BenchmarkDB_B/benchmarkDB_Get_B-4                         385225              3325 ns/op            1782 B/op          5 allocs/op
BenchmarkDB_B/benchmarkDB_Del_B
BenchmarkDB_B/benchmarkDB_Del_B-4                         401416              2528 ns/op             318 B/op          7 allocs/op
BenchmarkDB_KB
BenchmarkDB_KB/benchmarkDB_Put_value_1KB
BenchmarkDB_KB/benchmarkDB_Put_value_1KB-4                 44911             23509 ns/op            6786 B/op         38 allocs/op
BenchmarkDB_KB/benchmarkDB_Put_value_16KB
BenchmarkDB_KB/benchmarkDB_Put_value_16KB-4                 8968            117828 ns/op           33074 B/op         41 allocs/op
BenchmarkDB_KB/benchmarkDB_Put_value_64KB
BenchmarkDB_KB/benchmarkDB_Put_value_64KB-4                 2878            418228 ns/op          132088 B/op         43 allocs/op
BenchmarkDB_KB/benchmarkDB_Put_value_256KB
BenchmarkDB_KB/benchmarkDB_Put_value_256KB-4                 733           1714247 ns/op          697490 B/op         49 allocs/op
BenchmarkDB_KB/benchmarkDB_Put_value_512KB
BenchmarkDB_KB/benchmarkDB_Put_value_512KB-4                 304           3872659 ns/op         1648409 B/op         56 allocs/op
BenchmarkDB_KB/benchmarkDB_Get_KB
BenchmarkDB_KB/benchmarkDB_Get_KB-4                        28246             38379 ns/op           37384 B/op          6 allocs/op
BenchmarkDB_KB/benchmarkDB_Del_KB
BenchmarkDB_KB/benchmarkDB_Del_KB-4                       778448              1435 ns/op             240 B/op          7 allocs/op
BenchmarkDB_MB
BenchmarkDB_MB/benchmarkDB_Put_value_1MB
BenchmarkDB_MB/benchmarkDB_Put_value_1MB-4                   159           7583711 ns/op         1896272 B/op         60 allocs/op
BenchmarkDB_MB/benchmarkDB_Put_value_8MB
BenchmarkDB_MB/benchmarkDB_Put_value_8MB-4                    58          44785877 ns/op        15314996 B/op        166 allocs/op
BenchmarkDB_MB/benchmarkDB_Put_value_16MB
BenchmarkDB_MB/benchmarkDB_Put_value_16MB-4                   14         128916994 ns/op        49250470 B/op        360 allocs/op
BenchmarkDB_MB/benchmarkDB_Put_value_32MB
BenchmarkDB_MB/benchmarkDB_Put_value_32MB-4                   12         211075902 ns/op        74167508 B/op        546 allocs/op
BenchmarkDB_MB/benchmarkDB_Put_value_64MB
BenchmarkDB_MB/benchmarkDB_Put_value_64MB-4                    4         762692874 ns/op        282556294 B/op      1833 allocs/op
BenchmarkDB_MB/benchmarkDB_Get_MB
BenchmarkDB_MB/benchmarkDB_Get_MB-4                          420           2621809 ns/op          975465 B/op         18 allocs/op
BenchmarkDB_MB/benchmarkDB_Del_MB
BenchmarkDB_MB/benchmarkDB_Del_MB-4                      1000000              1313 ns/op             234 B/op          6 allocs/op
BenchmarkDB_Mixed
BenchmarkDB_Mixed/benchmarkDB_Put_Mixed_64B
BenchmarkDB_Mixed/benchmarkDB_Put_Mixed_64B-4              60070             22250 ns/op            5607 B/op         38 allocs/op
BenchmarkDB_Mixed/benchmarkDB_Put_Mixed_128KB
BenchmarkDB_Mixed/benchmarkDB_Put_Mixed_128KB-4             1387            881623 ns/op          295444 B/op         45 allocs/op
BenchmarkDB_Mixed/benchmarkDB_Put_Mixed_MB
BenchmarkDB_Mixed/benchmarkDB_Put_Mixed_MB-4                 162           7010276 ns/op         3669430 B/op         69 allocs/op
BenchmarkDB_Mixed/benchmarkDB_Get_Mixed
BenchmarkDB_Mixed/benchmarkDB_Get_Mixed-4                  31616             35471 ns/op           44763 B/op          6 allocs/op
BenchmarkDB_Mixed/benchmarkDB_Del_Mixed
BenchmarkDB_Mixed/benchmarkDB_Del_Mixed-4                 740533              1456 ns/op             242 B/op          7 allocs/op
PASS
ok      github.com/246859/river 51.461s
```