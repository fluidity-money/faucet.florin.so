-- migrate:up

ALTER TABLE florin_faucet_amounts_sent_1
ALTER COLUMN x_handle DROP NOT NULL;

-- migrate:down

DO $$
BEGIN
	RAISE EXCEPTION 'migration is irreversible after claims with NULL x_handle';
END $$;
