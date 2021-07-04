# Flip-shop

Simple online shopping application. Support Items, Promotions and Cart management.


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

#### Cart update from applied promotions

Promotions can change a Cart by:

- Adding and Item quantity to a Cart.
- Adding a discount to an Item purchased.



## REST API

### POST /cart

Create and available Cart

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

Submit the Cart by applying promotions and calculating the total

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

### Unit testes

- Most of the domain models have unit tests.
- Repositories and HTTP handlers are missing unit tests
