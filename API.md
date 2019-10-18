# API Documentation

The API is an HTTP REST API that accepts JSON request bodies (where applicable) and responds with JSON.

Errors are returned as JSON in the following format:

```json
{
    "error": "an error occured"
}
```

with an appropriate status code set in the header.

## Endpoints

<!-- MarkdownTOC -->

- [Accounts: List All](#accounts-list-all)
- [Payments: List All](#payments-list-all)
- [Transfer](#transfer)

<!-- /MarkdownTOC -->

### Accounts: List All

```
URI: /v1/accounts
Content-Type: application/json
```

#### Example

```sh
curl 'http://localhost:8888/v1/accounts'
```

#### Request

empty

#### Response

```json
{
    "accounts": [
        {
            "id": "d3f05a8d-1708-47de-8e1c-304e7fb5a93f",
            "currency": "USD",
            "balance": "3.46"
        },
        {
            "id": "5e0281df-cb1e-4b2f-bf61-0286295d07c9",
            "currency": "USD",
            "balance": "196.54"
        },
        {
            "id": "46e0b1dd-5cb2-4b40-b4d9-06b5e3d51059",
            "currency": "SGD",
            "balance": "100.00"
        },
        {
            "id": "92820a1f-4249-44fd-a152-b956fb001274",
            "currency": "EUR",
            "balance": "100.00"
        },
        {
            "id": "a88d1536-73c0-4aef-bf1c-a89e355a00fe",
            "currency": "EUR",
            "balance": "100.00"
        },
        {
            "id": "ab5977f7-cb1a-4619-b76c-25a437d07ea7",
            "currency": "SGD",
            "balance": "100.00"
        }
    ]
}
```

### Payments: List All

```
URI: /v1/payments
Content-Type: application/json
```

#### Example

```sh
curl 'http://localhost:8888/v1/payments'
```

#### Request 

empty

#### Response

```json
{
    "payments": [
        {
            "id": "8c7ecafb-df60-400a-a985-8f260c2fbb2a",
            "to": "5e0281df-cb1e-4b2f-bf61-0286295d07c9",
            "amount": "100.00"
        },
        {
            "id": "2c7bbfa8-e838-4e62-93e3-5915d86e4484",
            "to": "5e0281df-cb1e-4b2f-bf61-0286295d07c9",
            "from": "d3f05a8d-1708-47de-8e1c-304e7fb5a93f",
            "amount": "13.10"
        },
        {
            "id": "0797f6c9-c779-4ef6-adab-12a5b151f20f",
            "to": "5e0281df-cb1e-4b2f-bf61-0286295d07c9",
            "from": "d3f05a8d-1708-47de-8e1c-304e7fb5a93f",
            "amount": "86.90"
        },
        {
            "id": "d9bc38fe-5049-4667-9fcd-48c584caaae9",
            "to": "d3f05a8d-1708-47de-8e1c-304e7fb5a93f",
            "from": "5e0281df-cb1e-4b2f-bf61-0286295d07c9",
            "amount": "1.00"
        },
        {
            "id": "4e1748ce-950a-41be-b896-199e1e3e7d51",
            "to": "d3f05a8d-1708-47de-8e1c-304e7fb5a93f",
            "from": "5e0281df-cb1e-4b2f-bf61-0286295d07c9",
            "amount": "1.23"
        }
    ]
}
```

### Transfer

```
URI: /v1/transfer
Accept: application/json
Content-Type: application/json
```

#### Example

```sh
curl -X POST 'http://localhost:8888/v1/transfer' -d '{"to":"d3f05a8d-1708-47de-8e1c-304e7fb5a93f","from":"5e0281df-cb1e-4b2f-bf61-0286295d07c9","amount":"1.23"}'
```

#### Request body

```json
{
    "to": "d3f05a8d-1708-47de-8e1c-304e7fb5a93f",
    "from": "5e0281df-cb1e-4b2f-bf61-0286295d07c9",
    "amount": "1.23"
}
```

#### Response

```json
{
    "payment": {
        "id": "4e1748ce-950a-41be-b896-199e1e3e7d51",
        "to": "d3f05a8d-1708-47de-8e1c-304e7fb5a93f",
        "from": "5e0281df-cb1e-4b2f-bf61-0286295d07c9",
        "amount": "1.23"
    }
}
```
