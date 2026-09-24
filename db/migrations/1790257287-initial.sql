
-- migrate:up

DO $$
BEGIN
	IF NOT EXISTS (
		SELECT 1 FROM pg_type WHERE typname = 'hugeint'
	) THEN
		CREATE DOMAIN HUGEINT AS NUMERIC(78, 0);
	END IF;

	IF NOT EXISTS (
		SELECT 1 FROM pg_type WHERE typname = 'address'
	) THEN
		CREATE DOMAIN ADDRESS AS CHAR(42);
	END IF;

	IF NOT EXISTS (
		SELECT 1 FROM pg_type WHERE typname = 'hash'
	) THEN
		CREATE DOMAIN HASH AS CHAR(66);
	END IF;
END $$;

CREATE TABLE florin_faucet_amounts_sent_1 (
	id SERIAL PRIMARY KEY,
	created_by TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	recipient ADDRESS NOT NULL,
	x_handle VARCHAR NOT NULL,
	transaction_hash HASH NOT NULL
);

CREATE UNIQUE INDEX florin_faucet_amounts_sent_1_recipient_day
ON florin_faucet_amounts_sent_1 (recipient, (created_by::date));

-- migrate:down
