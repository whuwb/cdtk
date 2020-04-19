CREATE TABLE todo
(
    id SERIAL,
    title character varying(100) NOT NULL,
    body character varying(500) NOT NULL,
    Created date
);