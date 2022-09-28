module github.com/myfintech/ark

go 1.16

replace (
	github.com/containerd/containerd => github.com/containerd/containerd v1.3.1-0.20200227195959-4d242818bf55
	github.com/docker/docker => github.com/docker/docker v1.4.2-0.20200227233006-38f52c9fec82
	github.com/jaguilar/vt100 => github.com/tonistiigi/vt100 v0.0.0-20190402012908-ad4c4a574305
	github.com/spf13/cobra => github.com/spf13/cobra v1.0.1-0.20200909172742-8a63648dd905
	gorm.io/datatypes => github.com/go-gorm/datatypes v1.0.1
	gorm.io/gorm => github.com/go-gorm/gorm v1.21.6
)

require (
	cloud.google.com/go/bigquery v1.8.0
	cloud.google.com/go/firestore v1.1.0 // indirect
	cloud.google.com/go/pubsub v1.5.0
	cloud.google.com/go/storage v1.10.0 // indirect
	contrib.go.opencensus.io/exporter/stackdriver v0.13.4
	contrib.go.opencensus.io/exporter/zipkin v0.1.1
	github.com/AlecAivazis/survey/v2 v2.0.7
	github.com/DataDog/datadog-go v4.8.0+incompatible // indirect
	github.com/StackExchange/wmi v0.0.0-20190523213315-cbe66965904d // indirect
	github.com/ThreeDotsLabs/watermill v1.1.1
	github.com/apex/log v1.1.1 // indirect
	github.com/asaskevich/govalidator v0.0.0-20200108200545-475eaeb16496 // indirect
	github.com/aws/aws-sdk-go v1.36.30 // indirect
	github.com/beevik/etree v1.1.0
	github.com/boltdb/bolt v1.3.1
	github.com/brianvoe/gofakeit/v4 v4.3.0
	github.com/cenkalti/backoff/v4 v4.1.2 // indirect
	github.com/charmbracelet/bubbles v0.8.0
	github.com/charmbracelet/bubbletea v0.14.0
	github.com/charmbracelet/lipgloss v0.2.1
	github.com/cheynewallace/tabby v1.1.0
	github.com/chzyer/readline v0.0.0-20180603132655-2972be24d48e
	github.com/cloudevents/sdk-go/v2 v2.4.1
	github.com/cloudflare/cfssl v1.4.1 // indirect
	github.com/cloudflare/cloudflare-go v0.22.0 // indirect
	github.com/confluentinc/confluent-kafka-go v1.7.0
	github.com/containerd/console v1.0.2
	github.com/denisenkom/go-mssqldb v0.9.0 // indirect
	github.com/docker/cli v20.10.11+incompatible
	github.com/docker/distribution v2.7.1+incompatible
	github.com/docker/docker v20.10.7+incompatible
	github.com/docker/docker-credential-helpers v0.6.3 // indirect
	github.com/docker/go v1.5.1-1 // indirect
	github.com/docker/go-connections v0.4.0
	github.com/docker/go-metrics v0.0.1 // indirect
	github.com/dop251/goja v0.0.0-20210322220816-6fc852574a34
	github.com/elastic/go-elasticsearch/v7 v7.10.0
	github.com/fatih/color v1.10.0 // indirect
	github.com/fsnotify/fsnotify v1.4.9
	github.com/fvbommel/sortorder v1.0.2 // indirect
	github.com/gbrlsnchs/jwt v1.1.0
	github.com/go-chi/chi v4.0.2+incompatible // indirect
	github.com/go-errors/errors v1.0.1
	github.com/go-git/go-git/v5 v5.4.2
	github.com/go-ole/go-ole v1.2.4 // indirect
	github.com/go-ozzo/ozzo-validation v3.6.0+incompatible // indirect
	github.com/go-ozzo/ozzo-validation/v4 v4.3.0
	github.com/go-redis/redis/v8 v8.11.4 // indirect
	github.com/go-sql-driver/mysql v1.6.0 // indirect
	github.com/gobwas/glob v0.2.3
	github.com/gofiber/fiber/v2 v2.25.0
	github.com/gofrs/flock v0.8.0 // indirect
	github.com/golang/protobuf v1.5.2
	github.com/google/certificate-transparency-go v1.1.1 // indirect
	github.com/google/go-github v17.0.0+incompatible // indirect
	github.com/google/go-jsonnet v0.15.0
	github.com/google/uuid v1.3.0
	github.com/google/wire v0.5.0
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/gorilla/schema v1.2.0 // indirect
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0
	github.com/hashicorp/go-hclog v0.16.1 // indirect
	github.com/hashicorp/go-msgpack v1.1.5 // indirect
	github.com/hashicorp/go-multierror v1.1.1
	github.com/hashicorp/go-version v1.3.0
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/hashicorp/hcl/v2 v2.3.0
	github.com/hashicorp/terraform v0.12.21
	github.com/hashicorp/vault v1.3.2
	github.com/hashicorp/vault-plugin-secrets-kv v0.5.3
	github.com/hashicorp/vault/api v1.0.5-0.20200117231345-460d63e36490
	github.com/hashicorp/vault/sdk v0.1.14-0.20200121232954-73f411823aa0
	github.com/heptiolabs/healthcheck v0.0.0-20180807145615-6ff867650f40 // indirect
	github.com/jackc/pgconn v1.12.1 // indirect
	github.com/jackc/pgx v3.6.2+incompatible // indirect
	github.com/jackc/pgx/v4 v4.16.1 // indirect
	github.com/jinzhu/gorm v1.9.14 // indirect
	github.com/jmoiron/sqlx v1.2.1-0.20190826204134-d7d95172beb5 // indirect
	github.com/juju/fslock v0.0.0-20160525022230-4d5c94c67b4b
	github.com/kr/fs v0.1.0 // indirect
	github.com/lib/pq v1.10.2 // indirect
	github.com/ma314smith/signedxml v0.0.0-20191115220055-0d2ff290ff35
	github.com/matoous/go-nanoid v1.5.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/hashstructure/v2 v2.0.1
	github.com/mitchellh/mapstructure v1.4.1
	github.com/moby/buildkit v0.7.1
	github.com/moby/term v0.0.0-20210619224110-3f7ff695adc6 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/muesli/termenv v0.8.1
	github.com/nats-io/nats-server/v2 v2.7.4
	github.com/nats-io/nats.go v1.13.1-0.20220308171302-2f2f6968e98d
	github.com/nbio/st v0.0.0-20140626010706-e9e8d9816f32 // indirect
	github.com/openzipkin/zipkin-go v0.2.2
	github.com/osteele/liquid v1.2.4 // indirect
	github.com/osteele/tuesday v1.0.3 // indirect
	github.com/pkg/errors v0.9.1
	github.com/pkg/profile v1.2.1
	github.com/pkg/sftp v1.13.1 // indirect
	github.com/prometheus/procfs v0.6.0 // indirect
	github.com/reactivex/rxgo/v2 v2.0.0
	github.com/satori/go.uuid v1.2.0
	github.com/shirou/gopsutil v3.21.3+incompatible // indirect
	github.com/sirupsen/logrus v1.8.1
	github.com/slack-go/slack v0.8.1 // indirect
	github.com/spf13/afero v1.6.0
	github.com/spf13/cast v1.3.1
	github.com/spf13/cobra v1.1.3
	github.com/spf13/viper v1.8.1
	github.com/stretchr/testify v1.7.0
	github.com/strongdm/strongdm-sdk-go v0.9.20 // indirect
	github.com/theupdateframework/notary v0.6.1 // indirect
	github.com/tj/assert v0.0.0-20171129193455-018094318fb0
	github.com/tklauser/go-sysconf v0.3.5 // indirect
	github.com/valyala/fasthttp v1.33.0 // indirect
	github.com/vmihailenco/msgpack v4.0.4+incompatible
	github.com/x-cray/logrus-prefixed-formatter v0.5.2
	github.com/zclconf/go-cty v1.2.1
	github.com/zclconf/go-cty-yaml v1.0.1
	go.opencensus.io v0.23.0
	go.temporal.io/sdk v1.14.0
	gocloud.dev v0.19.0
	golang.org/x/crypto v0.0.0-20220112180741-5e0467b6c7ce // indirect
	golang.org/x/net v0.0.0-20220127200216-cd36cc0744dd
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	golang.org/x/sys v0.0.0-20220222200937-f2425489ef4c // indirect
	golang.org/x/term v0.0.0-20210927222741-03fcf44c2211
	google.golang.org/api v0.44.0
	google.golang.org/grpc v1.44.0
	google.golang.org/grpc/examples v0.0.0-20220311002955-722367c4a737 // indirect
	google.golang.org/protobuf v1.27.1
	gopkg.in/DATA-DOG/go-sqlmock.v1 v1.3.0 // indirect
	gopkg.in/auth0.v5 v5.19.2 // indirect
	gopkg.in/dancannon/gorethink.v3 v3.0.5 // indirect
	gopkg.in/fatih/pool.v2 v2.0.0 // indirect
	gopkg.in/gorethink/gorethink.v3 v3.0.5 // indirect
	gopkg.in/h2non/gentleman.v2 v2.0.5
	gopkg.in/osteele/liquid.v1 v1.2.4
	gopkg.in/src-d/go-git.v4 v4.13.1
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
	gorm.io/driver/postgres v1.0.8 // indirect
	gorm.io/driver/sqlite v1.1.4
	gorm.io/gorm v1.21.11
	k8s.io/api v0.19.0
	k8s.io/apiextensions-apiserver v0.19.0
	k8s.io/apimachinery v0.19.0
	k8s.io/cli-runtime v0.19.0
	k8s.io/client-go v0.19.0
	k8s.io/kubectl v0.19.0
	px.dev/pxapi v0.0.0-20210618035933-f07f46b9f09c // indirect
	upper.io/db.v3 v3.6.3+incompatible // indirect
)
