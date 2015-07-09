--
-- PostgreSQL database dump
--

SET statement_timeout = 0;
SET lock_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SET check_function_bodies = false;
SET client_min_messages = warning;

--
-- Name: plpgsql; Type: EXTENSION; Schema: -; Owner: 
--

CREATE EXTENSION IF NOT EXISTS plpgsql WITH SCHEMA pg_catalog;

--
-- Name: EXTENSION plpgsql; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION plpgsql IS 'PL/pgSQL procedural language';


SET search_path = public, pg_catalog;

SET default_tablespace = '';

SET default_with_oids = false;

--
-- Name: failed_logins; Type: TABLE; Schema: public; Owner: postgres; Tablespace: 
--

CREATE TABLE failed_logins (
	ip inet NOT NULL,
	username text NOT NULL,
	attempts integer NOT NULL,
	last timestamp with time zone NOT NULL
);


ALTER TABLE failed_logins OWNER TO postgres;

--
-- Name: hotp_codes; Type: TABLE; Schema: public; Owner: postgres; Tablespace: 
--

CREATE TABLE hotp_codes (
	email text NOT NULL,
	otp bytea,
	qr bytea
);


ALTER TABLE hotp_codes OWNER TO postgres;

--
-- Name: jobs; Type: TABLE; Schema: public; Owner: postgres; Tablespace: 
--

CREATE TABLE jobs (
	job_id text NOT NULL,
	user_id text NOT NULL,
	job_type integer NOT NULL,
	start_date date NOT NULL,
	end_date date
);


ALTER TABLE jobs OWNER TO postgres;

--
-- Name: sessions; Type: TABLE; Schema: public; Owner: postgres; Tablespace: 
--

CREATE TABLE sessions (
	session_id text NOT NULL,
	auth_token bytea NOT NULL,
	csrf_token bytea NOT NULL,
	email text NOT NULL,
	org text NOT NULL,
	date bigint NOT NULL
);


ALTER TABLE sessions OWNER TO postgres;

--
-- Name: users; Type: TABLE; Schema: public; Owner: postgres; Tablespace: 
--

CREATE TABLE users (
	email text NOT NULL,
	org text NOT NULL,
	username text NOT NULL,
	password text NOT NULL,
	tfa boolean DEFAULT false NOT NULL,
	user_data xml
);


ALTER TABLE users OWNER TO postgres;

--
-- Name: orgs; Type: TABLE; Schema: public; Owner: postgres; Tablespace: 
--

CREATE TABLE orgs (
	org text NOT NULL,
	prefix text NOT NULL
);


ALTER TABLE orgs OWNER TO postgres;

--
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY users (email, org, username, password, tfa, user_data) FROM stdin;
eric@sermodigital.com	SermoDigital	admin	$2a$10$YX6/o2EpdDtfNphzkw7tzO/tbq9sD5c.dEqBYnymwplDx4v7u28G2	 f	\N
\.

--
-- Data for Name: orgs; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY orgs (org, prefix) FROM stdin;
SermoDigital	sd_
\.

--
-- Name: failed_logins_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres; Tablespace: 
--

ALTER TABLE ONLY failed_logins
	ADD CONSTRAINT failed_logins_pkey PRIMARY KEY (ip);


--
-- Name: hotp_codes_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres; Tablespace: 
--

ALTER TABLE ONLY hotp_codes
	ADD CONSTRAINT hotp_codes_pkey PRIMARY KEY (email);


--
-- Name: jobs_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres; Tablespace: 
--

ALTER TABLE ONLY jobs
	ADD CONSTRAINT jobs_pkey PRIMARY KEY (job_id);


--
-- Name: users_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres; Tablespace: 
--

ALTER TABLE ONLY users
	ADD CONSTRAINT users_pkey PRIMARY KEY (email);

--
-- Name: orgs_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres; Tablespace: 
--

ALTER TABLE ONLY orgs
	ADD CONSTRAINT orgs_pkey PRIMARY KEY (org);



--
-- Name: public; Type: ACL; Schema: -; Owner: postgres
--

REVOKE ALL ON SCHEMA public FROM PUBLIC;
REVOKE ALL ON SCHEMA public FROM postgres;
GRANT ALL ON SCHEMA public TO postgres;
GRANT ALL ON SCHEMA public TO PUBLIC;


--
-- PostgreSQL database dump complete
--

