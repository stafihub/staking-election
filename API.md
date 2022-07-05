# API doc

## 0. notice

**status code:**

```go
 codeSuccess               = "80000"
 codeParamParseErr         = "80001"
 codeSymbolErr             = "80002"
 codeStafiAddressErr       = "80003"
 codeTxHashErr             = "80004"
 codePubkeyErr             = "80005"
 codeInternalErr           = "80006"
```

## 1. get annual rate list

### (1) description

* get annual rate list

### (2) path

* /stakingElection/api/v1/annualRateList

### (3) request method

* get

### (4) request payload

* null

### (5) response

* include status、data、message fields
* status、message must be string format,data must be object

| grade 1 | grade 2        | grade 3     | type   | must exist? | encode type | description           |
| :------ | :------------- | :---------- | :----- | :---------- | :---------- | :-------------------- |
| status  | N/A            | N/A         | string | Yes         | null        | status code           |
| message | N/A            | N/A         | string | Yes         | null        | status info           |
| data    | N/A            | N/A         | object | Yes         | null        | data                  |
|         | annualRateList | N/A         | list   | Yes         | null        | list                  |
|         |                | rTokenDenom | string | Yes         | null        | rtoken denom `uratom` |
|         |                | annualRate  | number | Yes         | null        | staking annual rate   |

## 2. get selected validators

### (1) description

* get selected validators

### (2) path

* /stakingElection/api/v1/selectedValidators

### (3) request method

* get

### (4) request param

* null

### (5) response

* include status、data、message fields
* status、message must be string format,data must be object

| grade 1 | grade 2            | grade 3       | grade4           | type   | must exist? | encode type | description           |
| :------ | :----------------- | :------------ | :--------------- | :----- | :---------- | :---------- | :-------------------- |
| status  | N/A                | N/A           | N/A              | string | Yes         | null        | status code           |
| message | N/A                | N/A           | N/A              | string | Yes         | null        | status info           |
| data    | N/A                | N/A           | N/A              | object | Yes         | null        | data                  |
|         | selectedValidators | N/A           | N/A              | list   | Yes         | null        | selected validators   |
|         |                    | rTokenDenom   | N/A              | string | Yes         | null        | rtoken denom `uratom` |
|         |                    | validatorList | N/A              | list   | Yes         | null        | validator list        |
|         |                    |               | validatorAddress | string | Yes         | null        | validator address     |
|         |                    |               | moniker          | string | Yes         | null        | moniker               |
|         |                    |               | logoUrl          | string | Yes         | null        | logo url              |
