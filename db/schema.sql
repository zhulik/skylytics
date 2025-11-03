-- Dumped from database version 17.6
-- Dumped by pg_dump version 17.6

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
-- Name: commit_operation; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.commit_operation AS ENUM (
    'create',
    'update',
    'delete'
);


--
-- Name: event_kind; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.event_kind AS ENUM (
    'commit',
    'account',
    'identity'
);


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: events; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.events (
    did text NOT NULL,
    time_us timestamp with time zone NOT NULL,
    kind public.event_kind NOT NULL,
    rev text,
    operation public.commit_operation,
    collection text,
    rkey text,
    record jsonb,
    cid text,
    account_active boolean,
    account_did text,
    account_seq bigint,
    account_status text,
    account_time timestamp with time zone,
    identity_did text,
    identity_handle text,
    identity_seq bigint,
    identity_time timestamp with time zone
);


--
-- Name: schema_migrations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.schema_migrations (
    version character varying NOT NULL
);


--
-- Name: events events_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.events
    ADD CONSTRAINT events_pkey PRIMARY KEY (did, time_us);


--
-- Name: schema_migrations schema_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.schema_migrations
    ADD CONSTRAINT schema_migrations_pkey PRIMARY KEY (version);


--
-- Name: idx_events_account_did; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_events_account_did ON public.events USING btree (account_did);


--
-- Name: idx_events_account_seq; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_events_account_seq ON public.events USING btree (account_seq);


--
-- Name: idx_events_cid; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_events_cid ON public.events USING btree (cid);


--
-- Name: idx_events_collection; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_events_collection ON public.events USING btree (collection);


--
-- Name: idx_events_did; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_events_did ON public.events USING btree (did);


--
-- Name: idx_events_identity_did; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_events_identity_did ON public.events USING btree (identity_did);


--
-- Name: idx_events_identity_handle; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_events_identity_handle ON public.events USING btree (identity_handle);


--
-- Name: idx_events_identity_seq; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_events_identity_seq ON public.events USING btree (identity_seq);


--
-- Name: idx_events_kind; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_events_kind ON public.events USING btree (kind);


--
-- Name: idx_events_operation; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_events_operation ON public.events USING btree (operation);


--
-- Name: idx_events_rev; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_events_rev ON public.events USING btree (rev);


--
-- Name: idx_events_rkey; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_events_rkey ON public.events USING btree (rkey);


--
-- PostgreSQL database dump complete
--

--
-- Dbmate schema migrations
--

INSERT INTO public.schema_migrations (version) VALUES
    ('20251102194428');
