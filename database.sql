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
-- Name: sessions; Type: TABLE; Schema: public; Owner: postgres; Tablespace: 
--

CREATE TABLE sessions (
	session_id text NOT NULL,
	auth_token bytea NOT NULL,
	csrf_token bytea NOT NULL,
	email text NOT NULL,
	school text NOT NULL,
	date bigint NOT NULL
);


ALTER TABLE sessions OWNER TO postgres;

--
-- Name: users; Type: TABLE; Schema: public; Owner: postgres; Tablespace: 
--

CREATE TABLE users (
	email text NOT NULL,
	school text NOT NULL,
	username text NOT NULL,
	password text NOT NULL
);


ALTER TABLE users OWNER TO postgres;

--
-- Name: schools; Type: TABLE; Schema: public; Owner: postgres; Tablespace: 
--

CREATE TABLE schools (
	school text NOT NULL
);


ALTER TABLE schools OWNER TO postgres;

--
-- Name: signups; Type: TABLE; Schema: public; Owner: postgres; Tablespace: 
--

CREATE TABLE signups (
	email text NOT NULL,
	first text NOT NULL,
	last text NOT NULL,
	school text NOT NULL,
	state text NOT NULL
);


ALTER TABLE schools OWNER TO postgres;

--
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY users (email, school, username, password) FROM stdin;
eric@eriscottlagergren@gmail.com	Pierce College	admin	$2a$10$YX6/o2EpdDtfNphzkw7tzO/tbq9sD5c.dEqBYnymwplDx4v7u28G2
\.

--
-- Data for Name: schools; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY schools (school) FROM stdin;
Pierce College
\.

--
-- Name: users_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres; Tablespace: 
--

ALTER TABLE ONLY users
	ADD CONSTRAINT users_pkey PRIMARY KEY (email);

--
-- Name: schools_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres; Tablespace: 
--

ALTER TABLE ONLY schools
	ADD CONSTRAINT schools_pkey PRIMARY KEY (school);

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

