module github.com/pingcap/tidb

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/blacktear23/go-proxyprotocol v0.0.0-20180807104634-af7a81e8dd0d
	github.com/codahale/hdrhistogram v0.0.0-20161010025455-3a0bb77429bd // indirect
	github.com/coreos/go-systemd v0.0.0-20181031085051-9002847aa142 // indirect
	github.com/coreos/pkg v0.0.0-20180928190104-399ea9e2e55f // indirect
	github.com/cznic/mathutil v0.0.0-20181122101859-297441e03548
	github.com/cznic/sortutil v0.0.0-20181122101858-f5f958428db8
	github.com/danjacques/gofslock v0.0.0-20191023191349-0a45f885bc37
	github.com/dgraph-io/ristretto v0.0.1
	github.com/dgryski/go-farm v0.0.0-20190423205320-6a90982ecee2
	github.com/dustin/go-humanize v1.0.0 // indirect
	github.com/go-sql-driver/mysql v1.5.0
	github.com/gogo/protobuf v1.3.1
	github.com/golang/groupcache v0.0.0-20190702054246-869f871628b6 // indirect
	github.com/golang/protobuf v1.3.4
	github.com/golang/snappy v0.0.1
	github.com/google/btree v1.0.0
	github.com/google/pprof v0.0.0-20200407044318-7d83b28da2e9
	github.com/google/uuid v1.1.1
	github.com/gorilla/mux v1.7.3
	github.com/gorilla/websocket v1.4.0 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.0.1-0.20190118093823-f849b5445de4
	github.com/klauspost/cpuid v0.0.0-20170728055534-ae7887de9fa5
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/kr/pty v1.1.5 // indirect
	github.com/montanaflynn/stats v0.0.0-20180911141734-db72e6cae808 // indirect
	github.com/ngaut/pools v0.0.0-20180318154953-b7bc8c42aac7
	github.com/ngaut/sync2 v0.0.0-20141008032647-7a24ed77b2ef
	github.com/opentracing/basictracer-go v1.0.0
	github.com/opentracing/opentracing-go v1.0.2
	github.com/pingcap/check v0.0.0-20200212061837-5e12011dc712
	github.com/pingcap/errors v0.11.5-0.20190809092503-95897b64e011
	github.com/pingcap/failpoint v0.0.0-20200210140405-f8f9fb234798
	github.com/pingcap/fn v0.0.0-20191016082858-07623b84a47d
	github.com/pingcap/goleveldb v0.0.0-20191226122134-f82aafb29989
	github.com/pingcap/kvproto v0.0.0-20200518112156-d4aeb467de29
	github.com/pingcap/log v0.0.0-20200511115504-543df19646ad
	github.com/pingcap/parser v0.0.0-20200811093644-e2438712084b
	github.com/pingcap/pd/v4 v4.0.0-rc.1.0.20200422143320-428acd53eba2
	github.com/pingcap/sysutil v0.0.0-20200408114249-ed3bd6f7fdb1
	github.com/pingcap/tidb-tools v4.0.0-beta.1.0.20200306084441-875bd09aa3d5+incompatible
	github.com/pingcap/tipb v0.0.0-20200417094153-7316d94df1ee
	github.com/prometheus/client_golang v1.0.0
	github.com/prometheus/client_model v0.0.0-20190812154241-14fe0d1b01d4
	github.com/prometheus/common v0.4.1
	github.com/shirou/gopsutil v2.19.10+incompatible
	github.com/shurcooL/vfsgen v0.0.0-20200824052919-0d455de96546 // indirect
	github.com/sirupsen/logrus v1.2.0
	github.com/soheilhy/cmux v0.1.4
	github.com/spaolacci/murmur3 v0.0.0-20180118202830-f09979ecbc72
	github.com/stretchr/objx v0.2.0 // indirect
	github.com/tiancaiamao/appdash v0.0.0-20181126055449-889f96f722a2
	github.com/uber-go/atomic v1.3.2
	github.com/uber/jaeger-client-go v2.15.0+incompatible
	github.com/uber/jaeger-lib v1.5.0 // indirect
	go.etcd.io/etcd v0.5.0-alpha.5.0.20191023171146-3cf2f69b5738
	go.uber.org/atomic v1.6.0
	go.uber.org/automaxprocs v1.2.0
	go.uber.org/zap v1.15.0
	golang.org/x/lint v0.0.0-20200302205851-738671d3881b // indirect
	golang.org/x/net v0.0.0-20200520182314-0ba52f642ac2
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e
	golang.org/x/sys v0.0.0-20200323222414-85ca7c5b95cd
	golang.org/x/text v0.3.3
	golang.org/x/tools v0.0.0-20200527183253-8e7acdbce89d
	google.golang.org/grpc v1.25.1
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
	honnef.co/go/tools v0.0.1-2020.1.4 // indirect
	sourcegraph.com/sourcegraph/appdash v0.0.0-20180531100431-4c381bd170b4
	sourcegraph.com/sourcegraph/appdash-data v0.0.0-20151005221446-73f23eafcf67
)

go 1.13

replace github.com/pingcap/pd/v4 => github.com/nolouch/pd/v4 v4.0.0-20210831114947-686590ed34cd
