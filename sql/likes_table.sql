create table likes (
    post_id serial,
    email varchar,
    
    foreign key (email) references users(email),
    foreign key (post_id) references posts(id),

    primary key(post_id, email)
)