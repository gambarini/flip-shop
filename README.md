# Flip-shop

Simple online shopping application. Support Items, Promotions and Cart management.


## Quick start
- Build: go build && ./flip-shop
- Run without building: go run ./...
- Default server port: 8001 (configurable via env, see Configuration)

## Configuration
- FLIPSHOP_PORT or PORT: server port (default 8001)
- FLIPSHOP_VERSION: version string exposed by /health (default "dev")
- FLIPSHOP_INVENTORY_JSON: optional JSON to seed items at startup. Example:
  - [{"sku":"120P90","name":"Google Home","price":4999,"qty":10}]

## Health endpoint
- GET /health → 200 OK
  - Response: {"status":"ok","uptime_seconds":123,"version":"dev"}

## Domain model

### Cart

Describes the Cart by managing Items Purchased and the total price of items.

#### Cart Status

- A new Carts has status Available.
- Carts with status Available can receive Item purchases, or be Submitted.
- Submitted Cart cannot receive Item purchases.
- Submitted a Cart will apply promotions to purchased items and remove purchased Item from being available.

### Item

Describes the items available for purchase.

#### Item availability and reservation

- Only a given quantity of an Item is available for shopping.

- Adding Items to a Cart reserves the Item quantity. Reserved Item quantities are not available for shopping
until removed from a Cart.

### Promotion

Describes the Promotions affecting a Cart, depending on the Items present in the Cart.

#### Available promotions (examples)
- Buy MacBook Pro (43N23P), get a Raspberry Pi B (234234) for free per MacBook purchased.
  - Implemented via FreeItemPromotion; cart will include Raspberry Pi items with a discount equal to their price.
- Buy 3 Google Home (120P90), get 1 free (every third unit is free).
  - Implemented via ItemQtyPriceFreePromotion; discount equals one Google Home price per 3 units.
- Buy more than 3 Alexa Speakers (A304SD), get 10% off on those speakers.
  - Implemented via ItemQtyPriceDiscountPercentagePromotion; applies when qty > 3 (not >=).

#### Cart update from applied promotions

Promotions can change a Cart by:

- Adding an Item quantity to a Cart (e.g., free Raspberry Pi).
- Adding a discount to an Item purchased.



## REST API

OpenAPI specification: docs/openapi.yaml

All responses are JSON with Content-Type: application/json.

### Read endpoints
- GET /items → list all items
- GET /cart/{cartID} → fetch a cart by ID

### Error responses
- 404 Not Found: resource does not exist (e.g., cart not found).
  - {"error":"<message>"}
- 422 Unprocessable Entity: validation or domain error (e.g., invalid qty, item unavailable, item not found).
  - {"error":"<message>"}
- 500 Internal Server Error: unexpected server error.
  - {"error":"<message>"}

### POST /cart

Create an available cart.

Example request (curl):
- curl -s -X POST http://localhost:8001/cart

Response Payload
```json
{
    "CartID": "d6870d29-eb07-4a31-9469-abe898183a1c",
    "Purchases": {},
    "CartStatus": "Available",
    "Total": 0
}
```

### [PUT | DELETE] /cart/{cartID}/purchase

- PUT: Purchase an Item and add it to the Cart
- DELETE: Remove a Purchased Item from the Cart

Example request (PUT):
- curl -s -X PUT http://localhost:8001/cart/{cartID}/purchase -H 'Content-Type: application/json' -d '{"sku":"120P90","qty":3}'

Example request (DELETE):
- curl -s -X DELETE http://localhost:8001/cart/{cartID}/purchase -H 'Content-Type: application/json' -d '{"sku":"120P90","qty":1}'

Request Payload
```json
{
    "sku":"120P90",
    "qty": 3
}
```

Response Payload
```json
{
    "CartID": "d6870d29-eb07-4a31-9469-abe898183a1c",
    "Purchases": {
        "120P90": {
            "Sku": "120P90",
            "Name": "Google Home",
            "Price": 4999,
            "Qty": 3,
            "Discount": 0
        }
    },
    "CartStatus": "Available",
    "Total": 0
}
```



### PUT cart/{cartID}/status/submitted

Submit the Cart by applying promotions and calculating the total.

Example request (curl):
- curl -s -X PUT http://localhost:8001/cart/{cartID}/status/submitted

Response Payload
```json
{
    "CartID": "d6870d29-eb07-4a31-9469-abe898183a1c",
    "Purchases": {
        "120P90": {
            "Sku": "120P90",
            "Name": "Google Home",
            "Price": 4999,
            "Qty": 3,
            "Discount": 4999
        },
        "234234": {
            "Sku": "234234",
            "Name": "Raspberry Pi B",
            "Price": 3000,
            "Qty": 2,
            "Discount": 6000
        },
        "43N23P": {
            "Sku": "43N23P",
            "Name": "MacBook Pro",
            "Price": 539999,
            "Qty": 2,
            "Discount": 0
        },
        "A304SD": {
            "Sku": "A304SD",
            "Name": "Alexa Speaker",
            "Price": 10950,
            "Qty": 4,
            "Discount": 4380
        }
    },
    "CartStatus": "Submitted",
    "Total": 1129416
}
```

## Considerations

### Project organization

- ```/utils```: external utilities
- ```/internal/model```: domain models
- ```/internal/repo```: repositories
- ```/internal/route```: HTTP handlers, and domain service logic (** handlers and service logic could be separated ideally **)

### Memory key/value database

- The application uses a basic memory Key/Value database, in ```/utils/memdb```
- It implements an ACID isolation level of "Serializable" by mutex locking the map stored in memory
- It ** does not implement transaction rollback ** for the sake of demo simplicity (no time for this now)

### Unit tests

- Domain models, repositories, and HTTP handlers have tests.
- Run all tests: go test ./...
- With race: go test -race ./...
- Coverage HTML: go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out
