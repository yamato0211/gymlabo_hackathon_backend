create table users (
    id serial PRIMARY KEY,
    name varchar not null,
    email varchar unique not null,
    image varchar,
    password_hash varchar not null
);