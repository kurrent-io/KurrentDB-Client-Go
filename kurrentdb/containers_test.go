package kurrentdb_test

import (
	"context"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/kurrent-io/KurrentDB-Client-Go/v1/kurrentdb"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/davecgh/go-spew/spew"
	"github.com/goombaio/namegenerator"
)

const (
	EVENTSTORE_DOCKER_REGISTRY = "EVENTSTORE_DOCKER_REGISTRY"
	EVENTSTORE_DOCKER_IMAGE    = "EVENTSTORE_DOCKER_IMAGE"
	EVENTSTORE_DOCKER_TAG      = "EVENTSTORE_DOCKER_TAG"
	EVENTSTORE_DOCKER_PORT_ENV = "EVENTSTORE_DOCKER_PORT"
)

var (
	NAME_GENERATOR = namegenerator.NewNameGenerator(0)
)

// Container ...
type Container struct {
	Endpoint  string
	Container testcontainers.Container
}

type EventStoreDockerConfig struct {
	Registry string
	Image    string
	Tag      string
	Port     string
	Insecure bool
}

const (
	DEFAULT_EVENTSTORE_DOCKER_REGISTRY = "docker.eventstore.com/eventstore-utils"
	DEFAULT_EVENTSTORE_DOCKER_IMAGE    = "testdata"
	DEFAULT_EVENTSTORE_DOCKER_TAG      = "latest"
	DEFAULT_EVENTSTORE_DOCKER_PORT     = "2113"
)

var defaultEventStoreDockerConfig = EventStoreDockerConfig{
	Registry: DEFAULT_EVENTSTORE_DOCKER_REGISTRY,
	Image:    DEFAULT_EVENTSTORE_DOCKER_IMAGE,
	Tag:      DEFAULT_EVENTSTORE_DOCKER_TAG,
	Port:     DEFAULT_EVENTSTORE_DOCKER_PORT,
}

func GetEnvOrDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func readEnvironmentVariables(config EventStoreDockerConfig) EventStoreDockerConfig {
	config.Registry = GetEnvOrDefault(EVENTSTORE_DOCKER_REGISTRY, config.Registry)
	config.Image = GetEnvOrDefault(EVENTSTORE_DOCKER_IMAGE, config.Image)
	config.Tag = GetEnvOrDefault(EVENTSTORE_DOCKER_TAG, config.Tag)
	config.Port = GetEnvOrDefault(EVENTSTORE_DOCKER_PORT_ENV, config.Port)

	fmt.Println(spew.Sdump(config))
	return config
}

type ESDBVersion struct {
	Maj   int
	Min   int
	Patch int
}

type VersionPredicateFn = func(ESDBVersion) bool

func IsESDB_Version(predicate VersionPredicateFn) bool {
	value, exists := os.LookupEnv(EVENTSTORE_DOCKER_TAG)
	if !exists || value == "ci" {
		return false
	}

	parts := strings.Split(value, "-")
	versionNumbers := strings.Split(parts[0], ".")

	version := ESDBVersion{
		Maj:   mustConvertToInt(versionNumbers[0]),
		Min:   mustConvertToInt(versionNumbers[1]),
		Patch: mustConvertToInt(versionNumbers[2]),
	}

	return predicate(version)
}

func mustConvertToInt(s string) int {
	val, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}
	return val
}

func IsESDBVersion20() bool {
	return IsESDB_Version(func(version ESDBVersion) bool {
		return version.Maj < 21
	})
}

func getContainerRequest() (*EventStoreDockerConfig, *testcontainers.ContainerRequest, error) {
	config := readEnvironmentVariables(defaultEventStoreDockerConfig)

	env := map[string]string{}
	var files []testcontainers.ContainerFile
	insecure, err := strconv.ParseBool(GetEnvOrDefault("EVENTSTORE_INSECURE", "true"))

	if err != nil {
		insecure = true
	}

	if !insecure {

		err := verifyCertificatesExist()

		if err != nil {
			return nil, nil, err
		}

		certsDir, err := getCertificatesDir()

		if err != nil {
			return nil, nil, err
		}

		env["EVENTSTORE_CERTIFICATE_FILE"] = "/etc/eventstore/certs/node/node.crt"
		env["EVENTSTORE_CERTIFICATE_PRIVATE_KEY_FILE"] = "/etc/eventstore/certs/node/node.key"
		env["EVENTSTORE_TRUSTED_ROOT_CERTIFICATES_PATH"] = "/etc/eventstore/certs/ca"

		files = append(files, testcontainers.ContainerFile{
			HostFilePath:      certsDir,
			ContainerFilePath: "/etc/eventstore/certs",
			FileMode:          int64(0755),
		})
	}

	env["EVENTSTORE_INSECURE"] = strconv.FormatBool(insecure)
	env["EVENTSTORE_RUN_PROJECTIONS"] = "all"
	env["EVENTSTORE_START_STANDARD_PROJECTIONS"] = "true"
	env["EVENTSTORE_ENABLE_ATOM_PUB_OVER_HTTP"] = "true"
	env["EventStore__Plugins__UserCertificates__Enabled"] = "true"

	return &config, &testcontainers.ContainerRequest{
		Image:        fmt.Sprintf("%s/%s:%s", config.Registry, config.Image, config.Tag),
		ExposedPorts: []string{config.Port},
		Env:          env,
		Files:        files,
		WaitingFor: wait.
			ForHTTP("/health/live").
			WithTLS(!insecure).
			WithStartupTimeout(1 * time.Minute).
			WithAllowInsecure(true).
			WithStatusCodeMatcher(func(status int) bool {
				return status >= 200 && status < 300
			}),
	}, nil
}

func (container *Container) Close() {
	timeout := 1 * time.Minute
	err := container.Container.Stop(context.Background(), &timeout)
	if err != nil {
		panic(err)
	}
}

func getDatabase(t *testing.T, config EventStoreDockerConfig, req testcontainers.ContainerRequest) *Container {
	container, err := testcontainers.GenericContainer(context.Background(), testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})

	if err != nil {
		t.Fatalf("error when starting container: %v", err)
	}

	port, err := container.MappedPort(context.Background(), nat.Port(config.Port))

	if err != nil {
		t.Fatalf("error when looking up container mapped port %s: %v", config.Port, err)
	}

	endpoint := fmt.Sprintf("localhost:%s", port.Port())

	if !container.IsRunning() {
		t.Fatalf("failed to get a running container after many attempts")
	}

	return &Container{
		Endpoint:  endpoint,
		Container: container,
	}
}

func verifyCertificatesExist() error {
	certs := []string{
		path.Join("ca", "ca.crt"),
		path.Join("ca", "ca.key"),
		path.Join("node", "node.crt"),
		path.Join("node", "node.key"),
	}

	certsDir, err := getCertificatesDir()
	if err != nil {
		return err
	}

	for _, f := range certs {
		if _, err := os.Stat(path.Join(certsDir, f)); os.IsNotExist(err) {
			return fmt.Errorf("could not locate the certificates needed to run KurrentDB and the tests. Please run 'docker compose up' for generating the certificates")
		}
	}
	return nil
}

func getCertificatesDir() (string, error) {
	rootDir, err := getRootDir()
	if err != nil {
		return "", err
	}
	return path.Join(rootDir, "certs"), nil
}

func getRootDir() (string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	currentDir = strings.Replace(currentDir, "\\", "/", -1)
	return path.Clean(path.Join(currentDir, "../")), nil
}

func GetClient(t *testing.T, container *Container) *kurrentdb.Client {
	isInsecure := GetEnvOrDefault("EVENTSTORE_INSECURE", "true") == "true"
	isCluster := GetEnvOrDefault("CLUSTER", "false") == "true"

	if isCluster {
		return CreateClient("esdb://admin:changeit@localhost:2111,localhost:2112,localhost:2113?nodepreference=leader&tlsverifycert=false&maxDiscoverAttempts=50&defaultDeadline=60000", t)
	} else if isInsecure {
		return createTestClient(fmt.Sprintf("esdb://%s?tls=false", container.Endpoint), container, t)
	}

	return createTestClient(fmt.Sprintf("esdb://admin:changeit@%s?tlsverifycert=false&tls=true", container.Endpoint), container, t)
}

func CreateEmptyDatabase(t *testing.T) (*Container, *kurrentdb.Client) {
	return createDatabase(t, false)
}

func CreatePopulatedDatabase(t *testing.T) (*Container, *kurrentdb.Client) {
	return createDatabase(t, true)
}

func createDatabase(t *testing.T, populated bool) (*Container, *kurrentdb.Client) {
	isInsecure := GetEnvOrDefault("EVENTSTORE_INSECURE", "true") == "true"

	var label string

	if populated {
		label = "populated"
	} else {
		label = "empty"
	}

	var container *Container
	var client *kurrentdb.Client

	if GetEnvOrDefault("CLUSTER", "false") == "true" {
		// When run on the cluster configuration we don't run the pre-populated database, so we have no use for a client
		// either.
		if !populated {
			client = GetClient(t, nil)
		}
	} else {
		if isInsecure {
			t.Logf("[debug] starting %s insecure database container...", label)
		} else {
			t.Logf("[debug] starting %s database container...", label)
		}

		config, req, err := getContainerRequest()

		if err != nil {
			t.Fatalf("error when constructing testcontainer request: %v", err)
		}

		if populated {
			req.Env["EVENTSTORE_DB"] = "/data/integration-tests"
			req.Env["EVENTSTORE_MEM_DB"] = "false"
		}

		container = getDatabase(t, *config, *req)
		client = GetClient(t, container)

	}

	if client != nil {
		WaitForAdminToBeAvailable(t, client)
		WaitForLeaderToBeElected(t, client)
	}

	return container, client
}

func createTestClient(conn string, container *Container, t *testing.T) *kurrentdb.Client {
	config, err := kurrentdb.ParseConnectionString(conn)
	if err != nil {
		t.Fatalf("Unexpected configuration error: %s", err.Error())
	}

	client, err := kurrentdb.NewClient(config)
	if err != nil {
		t.Fatalf("Unexpected failure setting up test connection: %s", err.Error())
	}

	return client
}

func WaitForAdminToBeAvailable(t *testing.T, db *kurrentdb.Client) {
	for count := 0; count < 50; count++ {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		t.Logf("[debug] checking if admin user is available...%v/50", count)

		stream, err := db.ReadStream(ctx, "$users", kurrentdb.ReadStreamOptions{}, 1)

		if ctx.Err() != nil {
			t.Log("[debug] request timed out, retrying...")
			cancel()
			time.Sleep(1 * time.Second)
			continue
		}

		if stream != nil {
			_, err = stream.Recv()
			if err == nil {
				t.Log("[debug] admin is available!")
				cancel()
				stream.Close()
				return
			}
		}

		if err != nil {
			if esdbError, ok := kurrentdb.FromError(err); !ok {
				if esdbError.Code() == kurrentdb.ErrorCodeResourceNotFound ||
					esdbError.Code() == kurrentdb.ErrorCodeUnauthenticated ||
					esdbError.Code() == kurrentdb.ErrorCodeDeadlineExceeded ||
					esdbError.Code() == kurrentdb.ErrorUnavailable {
					time.Sleep(1 * time.Second)
					t.Logf("[debug] not available retrying...")
					cancel()
					continue
				}

				t.Fatalf("unexpected error when waiting the admin account to be available: %+v", esdbError)
			}

			t.Fatalf("unexpected error when waiting the admin account to be available: %+v", err)
		}
	}

	t.Fatalf("failed to access admin account in a timely manner")
}

func WaitForLeaderToBeElected(t *testing.T, db *kurrentdb.Client) {
	for count := 0; count < 50; count++ {
		streamID := NAME_GENERATOR.Generate()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		t.Logf("[debug] checking if a leader has been elected...%v/50", count)

		err := db.CreatePersistentSubscription(ctx, streamID, "group", kurrentdb.PersistentStreamSubscriptionOptions{})

		if ctx.Err() != nil {
			t.Log("[debug] request timed out, retrying...")
			cancel()
			time.Sleep(1 * time.Second)
			continue
		}

		if err == nil {
			cancel()
			t.Log("[debug] a leader has been elected!")
			return
		}

		if err != nil {
			if esdbError, ok := kurrentdb.FromError(err); !ok {
				if esdbError.Code() == kurrentdb.ErrorCodeNotLeader || esdbError.Code() == kurrentdb.ErrorUnavailable || esdbError.Code() == kurrentdb.ErrorCodeUnauthenticated {
					time.Sleep(1 * time.Second)
					t.Logf("[debug] not available retrying...")
					cancel()
					continue
				}

				t.Fatalf("unexpected error when waiting for the cluster to elect a leader: %+v", esdbError)
			}

			t.Fatalf("unexpected error when waiting for the cluster to elect a leader: %+v", err)
		}
	}

	t.Fatalf("cluster failed to elect a leader in a timely manner")
}

func CreateClient(connStr string, t *testing.T) *kurrentdb.Client {
	config, err := kurrentdb.ParseConnectionString(connStr)

	if err != nil {
		t.Fatalf("Error when parsin connection string: %v", err)
	}

	client, err := kurrentdb.NewClient(config)

	if err != nil {
		t.Fatalf("Error when creating an KurrentDB esdb: %v", err)
	}

	return client
}
