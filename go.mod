module github.com/turnforge/weewar

go 1.24.6

require (
	cloud.google.com/go/datastore v1.21.0
	connectrpc.com/connect v1.19.1
	github.com/alexedwards/scs/v2 v2.9.0
	github.com/aws/aws-sdk-go-v2 v1.40.0
	github.com/aws/aws-sdk-go-v2/config v1.32.1
	github.com/aws/aws-sdk-go-v2/credentials v1.19.1
	github.com/aws/aws-sdk-go-v2/service/s3 v1.92.0
	github.com/chzyer/readline v1.5.1
	github.com/fatih/color v1.18.0
	github.com/felixge/httpsnoop v1.0.4
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.27.3
	github.com/joho/godotenv v1.5.1
	github.com/panyam/goapplib v0.0.4
	github.com/panyam/gocurrent v0.0.9
	github.com/panyam/goutils v0.1.11
	github.com/panyam/oneauth v0.0.15
	github.com/panyam/protoc-gen-dal v0.0.2
	github.com/panyam/protoc-gen-go-wasmjs v0.0.33
	github.com/panyam/servicekit v0.0.4
	github.com/panyam/templar v0.0.27
	github.com/spf13/cobra v1.10.2
	github.com/spf13/viper v1.21.0
	github.com/turnforge/turnengine v0.0.1
	go.opentelemetry.io/contrib/bridges/otelslog v0.13.0
	go.opentelemetry.io/otel v1.38.0
	golang.org/x/image v0.34.0
	golang.org/x/oauth2 v0.33.0
	google.golang.org/genproto/googleapis/api v0.0.0-20251111163417-95abcf5c77ba
	google.golang.org/grpc v1.78.0
	google.golang.org/protobuf v1.36.10
	gorm.io/driver/postgres v1.6.0
	gorm.io/gorm v1.25.10
)

require (
	cloud.google.com/go v0.121.6 // indirect
	cloud.google.com/go/auth v0.16.4 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.8 // indirect
	cloud.google.com/go/compute/metadata v0.9.0 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.7.3 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.18.14 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.4.14 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.7.14 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.4 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.4.14 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.9.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.13.14 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.19.14 // indirect
	github.com/aws/aws-sdk-go-v2/service/signin v1.0.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.30.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.35.9 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.41.1 // indirect
	github.com/aws/smithy-go v1.23.2 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-viper/mapstructure/v2 v2.4.0 // indirect
	github.com/golang-jwt/jwt/v5 v5.2.1 // indirect
	github.com/google/s2a-go v0.1.9 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.6 // indirect
	github.com/googleapis/gax-go/v2 v2.15.0 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.6.0 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/pelletier/go-toml/v2 v2.2.4 // indirect
	github.com/sagikazarmark/locafero v0.11.0 // indirect
	github.com/sourcegraph/conc v0.3.1-0.20240121214520-5f936abd7ae8 // indirect
	github.com/spf13/afero v1.15.0 // indirect
	github.com/spf13/cast v1.10.0 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.61.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.61.0 // indirect
	go.opentelemetry.io/otel/log v0.14.0 // indirect
	go.opentelemetry.io/otel/metric v1.38.0 // indirect
	go.opentelemetry.io/otel/trace v1.38.0 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/crypto v0.44.0 // indirect
	golang.org/x/net v0.47.0 // indirect
	golang.org/x/sync v0.19.0 // indirect
	golang.org/x/sys v0.38.0 // indirect
	golang.org/x/text v0.32.0 // indirect
	golang.org/x/time v0.12.0 // indirect
	google.golang.org/api v0.247.0 // indirect
	google.golang.org/genproto v0.0.0-20250603155806-513f23925822 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251103181224-f26f9409b101 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/turnforge/turnengine v0.0.1 => ./locallinks/turnengine/

replace github.com/panyam/protoc-gen-dal v0.0.2 => ./locallinks/protoc-gen-dal

replace github.com/panyam/gocurrent v0.0.9 => ./locallinks/gocurrent

replace github.com/panyam/goutils v0.1.11 => ./locallinks/goutils

replace github.com/panyam/goapplib v0.0.4 => ./locallinks/goapplib

replace github.com/panyam/templar v0.0.27 => ./locallinks/templar

replace github.com/panyam/servicekit v0.0.4 => ./locallinks/servicekit
