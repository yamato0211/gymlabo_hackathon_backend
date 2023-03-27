create table posts (
    id serial PRIMARY KEY,
    title varchar not null,
    content varchar not null,
    created_at timestamp,
    email varchar not null,

    foreign key (email) references users(email)
);