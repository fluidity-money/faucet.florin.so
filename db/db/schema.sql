\restrict nuhwtwzqFCKrAlU14CwHJBsNUGkSlxCGNhr5R9QrVjOd0RfgMovE2F9CbXUYRDM

-- Dumped from database version 16.13 (Ubuntu 16.13-1.pgdg22.04+1)
-- Dumped by pg_dump version 17.11 (Debian 17.11-0+deb13u1)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: timescaledb; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS timescaledb WITH SCHEMA public;


--
-- Name: EXTENSION timescaledb; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION timescaledb IS 'Enables scalable inserts and complex queries for time-series data (Community Edition)';


--
-- Name: timescaledb_toolkit; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS timescaledb_toolkit WITH SCHEMA public;


--
-- Name: EXTENSION timescaledb_toolkit; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION timescaledb_toolkit IS 'Library of analytical hyperfunctions, time-series pipelining, and other SQL utilities';


--
-- Name: address; Type: DOMAIN; Schema: public; Owner: -
--

CREATE DOMAIN public.address AS character(42);


--
-- Name: bytes16; Type: DOMAIN; Schema: public; Owner: -
--

CREATE DOMAIN public.bytes16 AS character(32);


--
-- Name: bytes32; Type: DOMAIN; Schema: public; Owner: -
--

CREATE DOMAIN public.bytes32 AS character(64);


--
-- Name: bytes64; Type: DOMAIN; Schema: public; Owner: -
--

CREATE DOMAIN public.bytes64 AS character(128);


--
-- Name: hash; Type: DOMAIN; Schema: public; Owner: -
--

CREATE DOMAIN public.hash AS character(66);


--
-- Name: hugeint; Type: DOMAIN; Schema: public; Owner: -
--

CREATE DOMAIN public.hugeint AS numeric(78,0);


--
-- Name: accounts_get_private_key_1(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.accounts_get_private_key_1() RETURNS public.bytes32
    LANGUAGE plpgsql
    AS $$
DECLARE
	selected_key BYTES64;
	random_offset INT;
	total_rows INT;
BEGIN
	SELECT COUNT(*) INTO total_rows FROM accounts_sender_keys_1;
	IF total_rows > 0 THEN
		random_offset := floor(random() * LEAST(5, total_rows))::INT;
		UPDATE accounts_sender_keys_1
		SET last_accessed = NOW()
		WHERE id = (
			SELECT id
			FROM accounts_sender_keys_1
			ORDER BY last_accessed ASC
			LIMIT 1
			OFFSET random_offset
		)
		RETURNING private_key INTO selected_key;
	END IF;
	RETURN selected_key;
END;
$$;


--
-- Name: accounts_get_private_key_2(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.accounts_get_private_key_2() RETURNS public.bytes32
    LANGUAGE plpgsql
    AS $$
DECLARE
	selected_key BYTES32;
BEGIN
	UPDATE accounts_sender_keys_1
	SET last_accessed = NOW()
	WHERE id = (
		SELECT id
		FROM accounts_sender_keys_1
		ORDER BY last_accessed ASC NULLS FIRST
		LIMIT 1
	)
	RETURNING private_key INTO selected_key;
	RETURN selected_key;
END;
$$;


--
-- Name: accounts_insert_nonce_1(public.address, integer); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.accounts_insert_nonce_1(p_eoa_addr public.address, p_consumed_nonce integer) RETURNS integer
    LANGUAGE plpgsql
    AS $$
DECLARE
	v_secret_id INTEGER;
	v_nonce_id INTEGER;
BEGIN
	SELECT id INTO v_secret_id
	FROM accounts_secrets_1
	WHERE eoa_addr = p_eoa_addr;
	IF v_secret_id IS NULL THEN
		RAISE EXCEPTION 'address not found';
	END IF;
	INSERT INTO accounts_secrets_nonces_1 (secret_id, consumed_nonce)
	VALUES (v_secret_id, p_consumed_nonce)
	RETURNING id INTO v_nonce_id;
	RETURN v_nonce_id;
END;
$$;


--
-- Name: accounts_insert_nonce_2(public.address, integer); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.accounts_insert_nonce_2(p_eoa_addr public.address, p_consumed_nonce integer) RETURNS integer
    LANGUAGE plpgsql
    AS $$
DECLARE
	v_secret_id INTEGER;
	v_nonce_id INTEGER;
BEGIN
	SELECT id INTO v_secret_id
	FROM accounts_secrets_1
	WHERE eoa_addr = p_eoa_addr AND valid_until > CURRENT_TIMESTAMP;
	IF v_secret_id IS NULL THEN
		RAISE EXCEPTION 'address not found';
	END IF;
	INSERT INTO accounts_secrets_nonces_1 (secret_id, consumed_nonce)
	VALUES (v_secret_id, p_consumed_nonce)
	RETURNING id INTO v_nonce_id;
	RETURN v_nonce_id;
END;
$$;


--
-- Name: accounts_insert_nonce_secret_1(public.address, public.bytes32, integer, public.bytes16); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.accounts_insert_nonce_secret_1(eoa_addr_ public.address, priv_key_ public.bytes32, nonce_ integer, salt_ public.bytes16) RETURNS void
    LANGUAGE plpgsql
    AS $$
BEGIN
	PERFORM accounts_insert_nonce_1(eoa_addr_, nonce_);
	INSERT INTO accounts_secrets_1(eoa_addr, priv_key, salt)
	VALUES (eoa_addr_, priv_key_, salt_);
END;
$$;


--
-- Name: accounts_insert_nonce_secret_2(public.address, public.bytes32, integer, public.bytes16); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.accounts_insert_nonce_secret_2(eoa_addr_ public.address, priv_key_ public.bytes32, nonce_ integer, salt_ public.bytes16) RETURNS void
    LANGUAGE plpgsql
    AS $$
BEGIN
	PERFORM accounts_insert_nonce_2(eoa_addr_, nonce_);
	INSERT INTO accounts_secrets_1(eoa_addr, priv_key, salt)
	VALUES (eoa_addr_, priv_key_, salt_);
END;
$$;


--
-- Name: accounts_insert_nonce_secret_3(public.address, public.bytes32, integer); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.accounts_insert_nonce_secret_3(eoa_addr_ public.address, secret_ public.bytes32, nonce_ integer) RETURNS void
    LANGUAGE plpgsql
    AS $$
BEGIN
	INSERT INTO accounts_secrets_nonces_2(eoa_addr, nonce) VALUES (eoa_addr_, nonce_);
	INSERT INTO accounts_secrets_2(eoa_addr, secret)
	VALUES (eoa_addr_, secret_);
END;
$$;


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: accounts_executed_transactions_1; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.accounts_executed_transactions_1 (
    id integer NOT NULL,
    eoa_addr public.address NOT NULL,
    transaction_hash public.hash NOT NULL
);


--
-- Name: accounts_executed_transactions_1_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.accounts_executed_transactions_1_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: accounts_executed_transactions_1_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.accounts_executed_transactions_1_id_seq OWNED BY public.accounts_executed_transactions_1.id;


--
-- Name: accounts_executed_transactions_2; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.accounts_executed_transactions_2 (
    id integer NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    eoa_addr public.address NOT NULL,
    transaction_hash public.hash NOT NULL,
    gas_limit integer NOT NULL,
    desc_ character varying NOT NULL
);


--
-- Name: accounts_executed_transactions_2_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.accounts_executed_transactions_2_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: accounts_executed_transactions_2_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.accounts_executed_transactions_2_id_seq OWNED BY public.accounts_executed_transactions_2.id;


--
-- Name: accounts_migrations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.accounts_migrations (
    version character varying(255) NOT NULL
);


--
-- Name: accounts_secrets_1; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.accounts_secrets_1 (
    id integer NOT NULL,
    eoa_addr public.address NOT NULL,
    priv_key public.bytes32 NOT NULL,
    salt public.bytes16 NOT NULL,
    valid_until timestamp without time zone DEFAULT (CURRENT_TIMESTAMP + '1 mon'::interval) NOT NULL
);


--
-- Name: accounts_secrets_1_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.accounts_secrets_1_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: accounts_secrets_1_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.accounts_secrets_1_id_seq OWNED BY public.accounts_secrets_1.id;


--
-- Name: accounts_secrets_2; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.accounts_secrets_2 (
    id integer NOT NULL,
    eoa_addr public.address NOT NULL,
    secret public.bytes32,
    valid_until timestamp without time zone DEFAULT (CURRENT_TIMESTAMP + '1 mon'::interval) NOT NULL
);


--
-- Name: accounts_secrets_2_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.accounts_secrets_2_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: accounts_secrets_2_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.accounts_secrets_2_id_seq OWNED BY public.accounts_secrets_2.id;


--
-- Name: accounts_secrets_nonces_1; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.accounts_secrets_nonces_1 (
    id integer NOT NULL,
    secret_id integer NOT NULL,
    consumed_nonce integer NOT NULL
);


--
-- Name: accounts_secrets_nonces_1_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.accounts_secrets_nonces_1_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: accounts_secrets_nonces_1_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.accounts_secrets_nonces_1_id_seq OWNED BY public.accounts_secrets_nonces_1.id;


--
-- Name: accounts_secrets_nonces_2; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.accounts_secrets_nonces_2 (
    id integer NOT NULL,
    eoa_addr public.address NOT NULL,
    nonce integer NOT NULL
);


--
-- Name: accounts_secrets_nonces_2_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.accounts_secrets_nonces_2_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: accounts_secrets_nonces_2_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.accounts_secrets_nonces_2_id_seq OWNED BY public.accounts_secrets_nonces_2.id;


--
-- Name: accounts_sender_keys_1; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.accounts_sender_keys_1 (
    id integer NOT NULL,
    private_key public.bytes32 NOT NULL,
    last_accessed timestamp without time zone
);


--
-- Name: accounts_sender_keys_1_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.accounts_sender_keys_1_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: accounts_sender_keys_1_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.accounts_sender_keys_1_id_seq OWNED BY public.accounts_sender_keys_1.id;


--
-- Name: accounts_transaction_statistics_1; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.accounts_transaction_statistics_1 AS
 SELECT desc_ AS action,
    avg(
        CASE
            WHEN (created_at >= (now() - '24:00:00'::interval)) THEN gas_limit
            ELSE NULL::integer
        END) AS avg_gas_limit_24_hours,
    avg(
        CASE
            WHEN (created_at >= (now() - '7 days'::interval)) THEN gas_limit
            ELSE NULL::integer
        END) AS avg_gas_limit_week,
    avg(gas_limit) AS avg_gas_limit_all_time,
    (count(
        CASE
            WHEN (created_at >= (now() - '24:00:00'::interval)) THEN 1
            ELSE NULL::integer
        END))::integer AS tx_24_hours,
    (count(
        CASE
            WHEN (created_at >= (now() - '7 days'::interval)) THEN 1
            ELSE NULL::integer
        END))::integer AS tx_week,
    (count(*))::integer AS tx_all_time
   FROM public.accounts_executed_transactions_2
  GROUP BY desc_;


--
-- Name: florin_faucet_amounts_sent_1; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.florin_faucet_amounts_sent_1 (
    id integer NOT NULL,
    created_by timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    recipient public.address NOT NULL,
    x_handle character varying,
    transaction_hash public.hash NOT NULL
);


--
-- Name: florin_faucet_amounts_sent_1_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.florin_faucet_amounts_sent_1_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: florin_faucet_amounts_sent_1_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.florin_faucet_amounts_sent_1_id_seq OWNED BY public.florin_faucet_amounts_sent_1.id;


--
-- Name: florin_faucet_migrationst; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.florin_faucet_migrationst (
    version character varying(255) NOT NULL
);


--
-- Name: accounts_executed_transactions_1 id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.accounts_executed_transactions_1 ALTER COLUMN id SET DEFAULT nextval('public.accounts_executed_transactions_1_id_seq'::regclass);


--
-- Name: accounts_executed_transactions_2 id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.accounts_executed_transactions_2 ALTER COLUMN id SET DEFAULT nextval('public.accounts_executed_transactions_2_id_seq'::regclass);


--
-- Name: accounts_secrets_1 id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.accounts_secrets_1 ALTER COLUMN id SET DEFAULT nextval('public.accounts_secrets_1_id_seq'::regclass);


--
-- Name: accounts_secrets_2 id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.accounts_secrets_2 ALTER COLUMN id SET DEFAULT nextval('public.accounts_secrets_2_id_seq'::regclass);


--
-- Name: accounts_secrets_nonces_1 id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.accounts_secrets_nonces_1 ALTER COLUMN id SET DEFAULT nextval('public.accounts_secrets_nonces_1_id_seq'::regclass);


--
-- Name: accounts_secrets_nonces_2 id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.accounts_secrets_nonces_2 ALTER COLUMN id SET DEFAULT nextval('public.accounts_secrets_nonces_2_id_seq'::regclass);


--
-- Name: accounts_sender_keys_1 id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.accounts_sender_keys_1 ALTER COLUMN id SET DEFAULT nextval('public.accounts_sender_keys_1_id_seq'::regclass);


--
-- Name: florin_faucet_amounts_sent_1 id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.florin_faucet_amounts_sent_1 ALTER COLUMN id SET DEFAULT nextval('public.florin_faucet_amounts_sent_1_id_seq'::regclass);


--
-- Name: accounts_executed_transactions_1 accounts_executed_transactions_1_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.accounts_executed_transactions_1
    ADD CONSTRAINT accounts_executed_transactions_1_pkey PRIMARY KEY (id);


--
-- Name: accounts_executed_transactions_1 accounts_executed_transactions_1_transaction_hash_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.accounts_executed_transactions_1
    ADD CONSTRAINT accounts_executed_transactions_1_transaction_hash_key UNIQUE (transaction_hash);


--
-- Name: accounts_executed_transactions_2 accounts_executed_transactions_2_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.accounts_executed_transactions_2
    ADD CONSTRAINT accounts_executed_transactions_2_pkey PRIMARY KEY (id);


--
-- Name: accounts_migrations accounts_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.accounts_migrations
    ADD CONSTRAINT accounts_migrations_pkey PRIMARY KEY (version);


--
-- Name: accounts_secrets_1 accounts_secrets_1_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.accounts_secrets_1
    ADD CONSTRAINT accounts_secrets_1_pkey PRIMARY KEY (id);


--
-- Name: accounts_secrets_2 accounts_secrets_2_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.accounts_secrets_2
    ADD CONSTRAINT accounts_secrets_2_pkey PRIMARY KEY (id);


--
-- Name: accounts_secrets_nonces_1 accounts_secrets_nonces_1_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.accounts_secrets_nonces_1
    ADD CONSTRAINT accounts_secrets_nonces_1_pkey PRIMARY KEY (id);


--
-- Name: accounts_secrets_nonces_2 accounts_secrets_nonces_2_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.accounts_secrets_nonces_2
    ADD CONSTRAINT accounts_secrets_nonces_2_pkey PRIMARY KEY (id);


--
-- Name: accounts_sender_keys_1 accounts_sender_keys_1_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.accounts_sender_keys_1
    ADD CONSTRAINT accounts_sender_keys_1_pkey PRIMARY KEY (id);


--
-- Name: accounts_sender_keys_1 accounts_sender_keys_1_private_key_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.accounts_sender_keys_1
    ADD CONSTRAINT accounts_sender_keys_1_private_key_key UNIQUE (private_key);


--
-- Name: florin_faucet_amounts_sent_1 florin_faucet_amounts_sent_1_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.florin_faucet_amounts_sent_1
    ADD CONSTRAINT florin_faucet_amounts_sent_1_pkey PRIMARY KEY (id);


--
-- Name: florin_faucet_migrationst florin_faucet_migrationst_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.florin_faucet_migrationst
    ADD CONSTRAINT florin_faucet_migrationst_pkey PRIMARY KEY (version);


--
-- Name: accounts_executed_transactions_2_desc__idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX accounts_executed_transactions_2_desc__idx ON public.accounts_executed_transactions_2 USING btree (desc_);


--
-- Name: accounts_executed_transactions_2_eoa_addr_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX accounts_executed_transactions_2_eoa_addr_idx ON public.accounts_executed_transactions_2 USING btree (eoa_addr);


--
-- Name: accounts_executed_transactions_2_transaction_hash_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX accounts_executed_transactions_2_transaction_hash_idx ON public.accounts_executed_transactions_2 USING btree (transaction_hash);


--
-- Name: accounts_secrets_nonces_1_secret_id_consumed_nonce_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX accounts_secrets_nonces_1_secret_id_consumed_nonce_idx ON public.accounts_secrets_nonces_1 USING btree (secret_id, consumed_nonce);


--
-- Name: accounts_secrets_nonces_2_eoa_addr_nonce_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX accounts_secrets_nonces_2_eoa_addr_nonce_idx ON public.accounts_secrets_nonces_2 USING btree (eoa_addr, nonce);


--
-- Name: florin_faucet_amounts_sent_1_recipient_day; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX florin_faucet_amounts_sent_1_recipient_day ON public.florin_faucet_amounts_sent_1 USING btree (recipient, ((created_by)::date));


--
-- Name: florin_faucet_amounts_sent_1_x_handle_day; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX florin_faucet_amounts_sent_1_x_handle_day ON public.florin_faucet_amounts_sent_1 USING btree (x_handle, ((created_by)::date));


--
-- PostgreSQL database dump complete
--

\unrestrict nuhwtwzqFCKrAlU14CwHJBsNUGkSlxCGNhr5R9QrVjOd0RfgMovE2F9CbXUYRDM


--
-- Dbmate schema migrations
--

INSERT INTO public.florin_faucet_migrationst (version) VALUES
    ('1790257287');
