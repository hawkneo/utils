# Utils

## decimal.Decimal
Decimal is a fixed-point decimal number.

> Decimal is immutable, so all methods return a new Decimal

### Create a Decimal

```go
decimal.NewFromBigInt(value *big.Int) Decimal
decimal.NewFromBigIntWithPrec(value *big.Int, precision int) Decimal
decimal.NewFromInt64(value int64, precision int) Decimal
decimal.NewFromUint64(value uint64, precision int) Decimal
decimal.NewFromString(str string) (d Decimal, err error)
decimal.MustFromString(str string) Decimal
```

## bigint.BigInt
BigInt is a wrapper around big.Int that provides some convenience methods.

> BigInt is immutable, so all methods return a new BigInt

### Create a BigInt

```go
bigint.NewFromInt(i int) BigInt
bigint.NewFromInt64(i int64) BigInt
bigint.NewFromUint(i uint) BigInt
bigint.NewFromUint64(i uint64) BigInt
bigint.NewFromBigInt(i *big.Int) BigInt
bigint.NewFromString(s string) (BigInt, bool)
bigint.MustNewFromString(s string) BigInt
```

## Database Migrate

### Install migrate

```bash
$ go install github.com/gridexswap/utils/sql-migrate/cmd/migrate@latest
```

### Create migration config file

```json
{
  "schema_name": "",
  "dialect": "",
  "data_source_name": "",
  "migration_source": "",
  "migrate_out_of_order": false,
  "disable_color_output": false
}
```

> Default config path: `./migration_config.json`, you can change it with `--config` flag

> More details should see: [main.go](migrate/cmd/main.go)

### Create migration schema in database

> More details should run: `migrate command --help`

```bash
$ migrate create
```

### Create a new migration file with description

```bash
$ migrate new version "description"
```

### Create a new migration file with description and version

```bash
$ migrate new version "description" -v <version number>
```

### Create a new config file with custom filename

```bash
$ migrate new config custom_migration_config.json
```

### Create a new config file with default filename

```bash
$ migrate new config
```

### How to write a new migration file

```sql
-- +migrate Up
-- SQL in section 'Up' is executed when this migration is applied

INSERT INTO test_schema (id, version, filename, hash, status) VALUES (3, 0, 'test file3', '0x01', 'applied');
INSERT INTO test_schema (id, version, filename, hash, status) VALUES (4, 0, 'test file4', '0x01', 'applied');

INSERT INTO test_schema (id, version, filename, hash, status) VALUES (5, 0, 'test file5', '0x01', 'applied');
INSERT INTO test_schema (id, version, filename, hash, status) VALUES (6, 0, 'test file6', '0x01', 'applied');

-- +migrate Down
-- SQL section 'Down' is executed when this migration is rolled back

DELETE FROM test_schema WHERE id IN (3, 4, 5, 6);
```

### Perform baseline for existing database

This will cause the database to be marked as having been migrated to the specified version.

```bash
$ migrate baseline
```

### Perform migrations

```bash
$ migrate up
```

### Perform rollback to target version

```bash
$ migrate down [version]
```

### Perform rollback to first version

```bash
$ migrate down --all
```

### Show migration status

```bash
$ migrate status
```

### Integrate with your project

```bash
$ go get github.com/gridexswap/utils/sql-migrate
```

```go
func TestUpMigrator(t *testing.T) {
	db, err := sql.Open("postgres", "host=localhost port=5432 user=postgres password=123 dbname=gridex sslmode=disable TimeZone=Etc/UTC")
	require.NoError(t, err)
	ctx := &Context{
		Context: context.TODO(),
		Conf: &Config{
			Dialect:           PostgresDialect{},
			DB:                db,
			SchemaName:        "migration_schema",
			MigrateOutOfOrder: false,
			MigrationSource:   DirectoryMigrationSource{Directory: "./test_data"},
		},
	}

	migrator, err := NewUpMigrator(ctx)
	require.NoError(t, err)
	require.NoError(t, migrator.Apply())
}
```

## multicall.NewMulticall&multicall.NewMulticall3
`Multicall` is sdk for [AggregateMulticall](https://github.com/GridexProtocol/gridex-facade/blob/main/contracts/AggregateMulticall.sol).
Can assign specific gas limit for call and do a batch of static calls within one http request.

`Multicall3` is sdk for [Multicall3](https://www.multicall3.com) which is deployed on 70+ chains.
Can migrate to various chains easily.

```golang
func TestMulticall3(t *testing.T) {
	mainNet := "https://eth.llamarpc.com"
	client, _ := ethclient.Dial(mainNet)
	muticall3Address := common.HexToAddress("0xcA11bde05977b3631167028862bE2a173976CA11")
	contract, err := NewMulticall3(client, muticall3Address)
	require.NoError(t, err)
	res, err := contract.Call(nil, ViewCalls{
		ViewCall{
			target:    common.HexToAddress("0xdAC17F958D2ee523a2206206994597C13D831ec7"),
			method:    "name()(string)",
			arguments: []interface{}{},
			callback: func(err error, returnValues []interface{}) error {
				require.NoError(t, err)
				require.Len(t, returnValues, 1)
				require.IsType(t, "", returnValues[0])
				require.Equal(t, "Tether USD", returnValues[0].(string))
				return nil
			},
		},
		ViewCall{
			target:    common.HexToAddress("0xdAC17F958D2ee523a2206206994597C13D831ec7"),
			method:    "balanceOf(address)(uint256)",
			arguments: []interface{}{common.HexToAddress("0xdAC17F958D2ee523a2206206994597C13D831ec7")},
			callback: func(err error, returnValues []interface{}) error {
				require.NoError(t, err)
				require.Len(t, returnValues, 1)
				require.IsType(t, big.NewInt(0), returnValues[0])
				require.NotZero(t, returnValues[0].(*big.Int).Uint64())
				return nil
			},
		},
	})
	require.NoError(t, err)
	for _, result := range res.Calls {
		require.NoError(t, result.Error)
	}
}
```

**FAO**:
function with self-defined struct arguments is not Supported.
eg: 
```solidity
balanceOf(address)(uint256) ✓
quoteExactInputSingle(QuoteParameters memory parameters) ❌
```
