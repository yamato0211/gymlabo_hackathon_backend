create table posts (
    id serial PRIMARY KEY,
    title varchar not null,
    content varchar not null,
    created_at timestamp default CURRENT_TIMESTAMP,
    email varchar not null,

    foreign key (email) references users(email)
);