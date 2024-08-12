# sd-payment-system

## DB

```mermaid
erDiagram
    PAYMENT_EVENT {
        string checkout_id PK
        string buyer_info
        string seller_info
        depends_on_card_provider credit_card_info
        boolean is_payment_done
    }

    PAYMENT_ORDER {
        string payment_order_id PK
        string buyer_account
        string amount
        string currency
        string checkout_id FK
        string payment_order_status
        boolean ledger_updated
        boolean wallet_updated
    }

    PAYMENT_EVENT ||--o{ PAYMENT_ORDER: "creates"
```
