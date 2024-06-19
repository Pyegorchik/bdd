package integrationstests

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strconv"
	"time"

	"github.com/Pyegorchik/bdd/backend/internal/config"
	"github.com/Pyegorchik/bdd/backend/internal/handler"
	"github.com/Pyegorchik/bdd/backend/internal/repository"
	"github.com/Pyegorchik/bdd/backend/internal/repository/postgres"
	"github.com/Pyegorchik/bdd/backend/internal/service"
	"github.com/Pyegorchik/bdd/backend/migrations"
	"github.com/Pyegorchik/bdd/backend/pkg/hash"
	"github.com/Pyegorchik/bdd/backend/pkg/jwtoken"
	"github.com/Pyegorchik/bdd/backend/pkg/logger"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	ModeIntTest = "integrationtest"
)

type TestSuite struct {
	suite.Suite
	postgreSQL *PostgreSQLContainer
	service    service.Service
	repo       repository.Repository
	pgxpool    *pgxpool.Pool
	accounts   map[int64]*Signer
	server     *httptest.Server
	handler    http.Handler

	cfg     *config.Config
	logging logger.Logger
}

func (s *TestSuite) SetupSuite() {
	var err error
	s.cfg, err = config.Init("../configs/local")
	s.Require().NoError(err)
	s.cfg.Service.Mode = ModeIntTest
}

func (s *TestSuite) SetupTest() {
	ctx := context.Background()

	psqlContaiter, err := NewPostgreSQLContainer(ctx)
	s.Require().NoError(err)
	s.postgreSQL = psqlContaiter

	psqlPort, err := strconv.Atoi(psqlContaiter.MappedPort)
	s.Require().NoError(err)

	s.cfg.Postgres.Port = psqlPort
	var pool *pgxpool.Pool
	for i := 0; i < 20; i++ {
		pool, err = postgres.New(ctx, s.cfg.Postgres)
		if err != nil {
			time.Sleep(300 * time.Millisecond)
			continue
		}
		s.Require().NoError(err)
		break
	}
	err = pool.Ping(ctx)
	s.Require().NoError(err)
	s.pgxpool = pool

	err = migrations.Migrate(psqlContaiter.GetDSN())
	s.Require().NoError(err)

	logging, err := logger.NewLogger()
	s.Require().NoError(err)
	defer logging.Sync()

	s.logging = logging

	jwtokenManager := jwtoken.NewTokenManager(s.cfg.TokenManager.SigningKey)

	repo, err := repository.NewRepository(s.cfg, pool)
	s.Require().NoError(err)
	s.repo = *repo

	s.Require().NoError(s.setupBlockChain(ctx))

	s.service, err = service.NewService(repo, jwtokenManager, hash.NewHashManager(), s.cfg.Service, logging)
	s.Require().NoError(err)

	h := handler.NewHandler(s.cfg.Handler, s.service, logging)
	s.handler = h.Init()

	s.server = httptest.NewServer(s.handler)
}

type Signer struct {
	auth *bind.TransactOpts
	pk   *ecdsa.PrivateKey
}

func (s *TestSuite) setupBlockChain(ctx context.Context) error {
	chainId := new(big.Int)
	chainId.SetString("1337", 10)

	accounts := make(map[int64]*Signer)

	pk, err := crypto.GenerateKey()
	if err != nil {
		return err
	}

	auth, err := bind.NewKeyedTransactorWithChainID(pk, chainId)
	if err != nil {
		log.Println("ERROR: NewKeyedTransactorWithChainID: ", err)
		return err
	}
	accounts[0] = &Signer{
		auth: auth,
		pk:   pk,
	}
	pk1, err := crypto.GenerateKey()
	if err != nil {
		log.Println("ERROR: GenerateKey: ", err)
		return err
	}
	auth1, err := bind.NewKeyedTransactorWithChainID(pk1, chainId)
	if err != nil {
		log.Println("ERROR: NewKeyedTransactorWithChainID: ", err)
		return err
	}
	accounts[1] = &Signer{
		auth: auth1,
		pk:   pk1,
	}
	pk2, err := crypto.GenerateKey()
	if err != nil {
		return err
	}
	auth2, err := bind.NewKeyedTransactorWithChainID(pk2, chainId)
	if err != nil {
		log.Println("ERROR: NewKeyedTransactorWithChainID: ", err)
		return err
	}
	accounts[2] = &Signer{
		auth: auth2,
		pk:   pk2,
	}
	pk3, err := crypto.GenerateKey()
	auth3, err := bind.NewKeyedTransactorWithChainID(pk3, chainId)
	if err != nil {
		log.Println("ERROR: NewKeyedTransactorWithChainID: ", err)
		return err
	}
	accounts[3] = &Signer{
		auth: auth3,
		pk:   pk3,
	}

	s.accounts = accounts
	return nil
}

type PostgreSQLContainer struct {
	testcontainers.Container
	MappedPort string
	Host       string
}

func NewPostgreSQLContainer(ctx context.Context) (*PostgreSQLContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "postgres:13",
		Name:         "bdd-postgres-test",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "bdd",
			"POSTGRES_DB":       "bdd",
			"POSTGRES_PASSWORD": "1337",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp"),
	}

	postgresContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	host, err := postgresContainer.Host(ctx)
	if err != nil {
		return nil, err
	}

	mappedPort, err := postgresContainer.MappedPort(ctx, "5432")
	if err != nil {
		return nil, err
	}

	started, err := checkPostgreSQLContainerStarted(postgresContainer)
	if err != nil {
		return nil, err
	}

	if !started {
		return nil, fmt.Errorf("container not started")
	}

	return &PostgreSQLContainer{
		Container:  postgresContainer,
		MappedPort: mappedPort.Port(),
		Host:       host,
	}, nil
}

func checkPostgreSQLContainerStarted(c testcontainers.Container) (bool, error) {
	for i := 0; i < 20; i++ {
		code, _, err := c.Exec(context.Background(), []string{"pg_isready", "-d", "bdd", "-U", "bdd"})
		if err != nil {
			return false, err
		}
		if code == 0 {
			return true, nil
		}
		time.Sleep(200 * time.Millisecond)
	}
	return false, nil
}

func (p *PostgreSQLContainer) GetDSN() string {
	return fmt.Sprintf("postgres://bdd:1337@127.0.0.1:%s/bdd", p.MappedPort)
}

func (s *TestSuite) TearDownTest() {
	ctx := context.Background()
	s.service.Shutdown()
	s.Require().NoError(s.postgreSQL.Terminate(ctx))
}

type TestSuiteUser struct {
	TestSuite
}

func (s *TestSuiteUser) SetupTest() {
	s.TestSuite.SetupTest()
}
