CREATE TABLE IF NOT EXISTS account (
    id UUID PRIMARY KEY,
    currency TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS payment (
    id UUID PRIMARY KEY,
    from_account_id UUID REFERENCES account(id),
    to_account_id UUID REFERENCES account(id) NOT NULL,
    amount NUMERIC(20, 2) NOT NULL CHECK (amount > 0.0),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
    -- amount TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS payment_from_idx ON payment(from_account_id);
CREATE INDEX IF NOT EXISTS payment_to_idx ON payment(to_account_id);

CREATE OR REPLACE VIEW account_payment(
    account_id,
    payment_id,
    amount
) AS
    SELECT
        payment.to_account_id,
        payment.id,
        payment.amount
    FROM
        payment
    UNION ALL
    SELECT
        payment.from_account_id,
        payment.id,
        (-1 * payment.amount)
    FROM
        payment
    WHERE
        payment.from_account_id IS NOT NULL;


CREATE OR REPLACE VIEW account_balance(
    id,
    balance,
    currency
) AS
    SELECT
        account.id,
        COALESCE(sum(account_payment.amount), 0.0),
        account.currency
    FROM
        account
        LEFT OUTER JOIN account_payment
        ON account.id = account_payment.account_id
    GROUP BY account.id;
